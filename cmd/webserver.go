// SPDX-FileCopyrightText: 2024 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil"
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

	rootCmd.AddCommand(servidorCmd)
}

func webserver(_ *cobra.Command, _ []string) {
	http.HandleFunc("/empresas", displayEmpresas)
	http.HandleFunc("/select", handleSelection)
	http.HandleFunc("/files", handleFiles)

	fs := http.FileServer(http.Dir("./cmd/assets/"))
	http.Handle("/scripts/", logHandler(fs))
	http.Handle("/styles/", logHandler(fs))

	fsRelat := http.FileServer(http.Dir(flags.relatorio.outputDir))
	http.Handle("/relatorios/", logHandler(stripPrefixHandler("/relatorios", fsRelat)))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

//go:embed assets/templates
var indexHTML embed.FS

func displayEmpresas(w http.ResponseWriter, _ *http.Request) {
	dfp, err := contabil.NovaDemonstraçãoFinanceira(db(), flags.tempDir)
	if err != nil {
		progress.Fatal(err)
	}

	empresas, err := dfp.Empresas()
	if err != nil {
		progress.Fatal(err)
	}

	t, err := template.ParseFS(indexHTML, "assets/templates/empresas.html")
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

	// w.WriteHeader(http.StatusOK)
	// for _, opt := range selectedEmpresas {
	// 	fmt.Fprintf(w, "CNPJ: %s, Nome: %s <br />\n", opt.CNPJ, opt.Nome)
	// }

	progress.SetOutput(w)
	dfp, err := contabil.NovaDemonstraçãoFinanceira(db(), flags.tempDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		progress.Fatal(err)
	}

	for _, empresa := range selectedEmpresas {
		criarRelatório(empresa, dfp)
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
}

func handleFiles(w http.ResponseWriter, _ *http.Request) {
	path := flags.relatorio.outputDir
	files, err := os.ReadDir(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(files) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var filesList []File
	for _, f := range files {
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
		})
	}

	t, err := template.ParseFS(indexHTML, "assets/templates/files.html")
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
