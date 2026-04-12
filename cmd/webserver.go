// SPDX-FileCopyrightText: 2024 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil"
	"github.com/dude333/rapinav2/pkg/googlesheets"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/spf13/cobra"
)

type flagsServidor struct {
	porta string
}

var servidorCmd = &cobra.Command{
	Use:     "servidor",
	Aliases: []string{"server"},
	Short:   "ativar servidor web",
	Long:    `ativar servidor web`,
	Run:     webserver,
}

func init() {
	servidorCmd.Flags().StringVarP(&flags.servidor.porta, "porta", "p", "8005", "Porta tcp do servidor")
	// Usando variável do relatório (flags.relatorio)
	servidorCmd.Flags().BoolVarP(&flags.relatorio.googlesheets, "googlesheets", "s", false, "Usar Google Sheets")
	servidorCmd.Flags().IntVarP(&flags.relatorio.tokenport, "tokenport", "k", 0, "Porta para autenticação OAuth (0 = porta automática)")
	servidorCmd.Flags().StringVarP(&flags.relatorio.oauthurl, "oauthurl", "u", "", "URL externa para autenticação OAuth (ex: https://seu-dominio.com)")

	rootCmd.AddCommand(servidorCmd)
}

//go:embed assets
var assets embed.FS

func webserver(_ *cobra.Command, _ []string) {
	http.HandleFunc("/", displayEmpresas)
	http.HandleFunc("/select", handleSelection)
	http.HandleFunc("/update", handleUpdate)
	http.HandleFunc("/files", handleFiles)

	fs := http.FileServer(http.FS(assets))
	http.Handle("/assets/", logHandler(fs))

	fsRelat := http.FileServer(http.Dir(flags.relatorio.outputDir))
	http.Handle("/relatorios/", logHandler(stripPrefixHandler("/relatorios", fsRelat)))

	addr := ":8080"
	if flags.servidor.porta != "" {
		addr = ":" + flags.servidor.porta
	}
	progress.Status("Iniciando servidor em %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func displayEmpresas(w http.ResponseWriter, _ *http.Request) {
	dfp, err := contabil.NewService(db(), flags.tempDir)
	if err != nil {
		progress.Fatal(err)
	}

	empresas, err := dfp.Empresas()
	if err != nil {
		progress.Fatal(err)
	}

	t, err := template.ParseFS(assets, "assets/templates/empresas.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, empresas); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleSelection(w http.ResponseWriter, r *http.Request) {
	var selectedEmpresas []rapina.Empresa

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	objs := r.Form["allOptions"]
	if err := json.Unmarshal([]byte(objs[0]), &selectedEmpresas); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	progress.SetOutput(w)
	dfp, err := contabil.NewService(db(), flags.tempDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		progress.Fatal(err)
	}

	for _, empresa := range selectedEmpresas {
		criarRelatórios(empresa, dfp)
	}
}

type sseWriter struct {
	w  io.Writer
	w2 io.Writer
}

func (sw sseWriter) Write(p []byte) (n int, err error) {
	if sw.w2 != nil {
		_, _ = sw.w2.Write(p)
	}
	return sw.w.Write([]byte("data: " + string(p) + "\n\n"))
}

var alreadyIn = false

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	log.Println("***************************************************")
	if alreadyIn {
		log.Println("Cancelando update")
		_, cancel := context.WithCancel(r.Context())
		cancel()
		return
	}
	alreadyIn = true
	w.Header().Set("Content-Type", "text/event-stream;  charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	progress.SetOutput(sseWriter{w: w, w2: os.Stdout})

	dfp, err := contabil.NewService(db(), flags.tempDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		panic(err)
	}

	f, ok := w.(http.Flusher)
	if !ok {
		progress.Warning("Servidor não suporta flush")
	}
	flush := func() {
		if ok {
			f.Flush()
		}
	}

	done := make(chan bool, 1)
	defer close(done)

	anof := time.Now().Year()
	anoi := anof - 1

	importar := func(trimestral bool) {
		for ano := anof; ano >= anoi; ano-- {
			err := dfp.Import(ano, trimestral)
			if err != nil {
				progress.Error(err)
				continue
			}
		}
	}

	go func() {
		importar(false)
		importar(true)
		progress.SetOutput(os.Stdout)
		done <- true
	}()

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			_, _ = fmt.Fprintf(w, "event: close\ndata: [>] Importação concluída\n\n")
			flush()
			progress.Status("Importação concluída")
			alreadyIn = false
			return
		case <-ticker.C:
			_, _ = fmt.Fprintf(w, ": keep-alive")
			flush()
		}
	}
}

func logHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

type File struct {
	Name    string
	ModTime string
	Size    string
	Mode    os.FileMode
	IsDir   bool
	Link    string // URL for Google Drive files
}

func handleFiles(w http.ResponseWriter, _ *http.Request) {
	var filesList []File

	// Add local files
	path := flags.relatorio.outputDir
	if localFiles, err := os.ReadDir(path); err == nil {
		for _, f := range localFiles {
			info, err := f.Info()
			if err != nil {
				continue
			}
			filesList = append(filesList, File{
				Name:    f.Name(),
				Size:    humanize(info.Size()),
				Mode:    info.Mode(),
				ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
				IsDir:   f.IsDir(),
				Link:    "", // Local files don't have external links
			})
		}
	}

	// Add Google Drive files if enabled
	if flags.relatorio.googlesheets {
		cfg := googlesheets.Config{
			CredentialsFile: "credentials.json",
			TokenFile:       "token.json",
			TokenPort:       flags.relatorio.tokenport,
			OAuthURL:        flags.relatorio.oauthurl,
		}
		driveFiles, err := googlesheets.ListFilesInFolder(context.Background(), cfg, flags.relatorio.googlesheetsDir)
		if err != nil {
			progress.ErrorMsg("Error listing Google Drive files: %v", err)
		} else {
			for _, f := range driveFiles {
				size := f.Size
				modTime := ""
				if f.ModifiedTime != "" {
					if t, err := time.Parse(time.RFC3339, f.ModifiedTime); err == nil {
						modTime = t.Format("2006-01-02 15:04:05")
					}
				}
				link := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s", f.Id)
				filesList = append(filesList, File{
					Name:    f.Name,
					Size:    humanize(size),
					Mode:    0, // Not applicable for Drive files
					ModTime: modTime,
					IsDir:   false, // Assume files, not folders
					Link:    link,
				})
			}
		}
	}

	if len(filesList) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	t, err := template.ParseFS(assets, "assets/templates/files.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, filesList); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func stripPrefixHandler(prefix string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		trimmedPath := strings.TrimPrefix(r.URL.Path, prefix)
		r.URL.Path = trimmedPath
		handler.ServeHTTP(w, r)
	})
}

func humanize(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d  B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
