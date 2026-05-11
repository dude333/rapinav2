// SPDX-FileCopyrightText: 2024 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil"
	"github.com/dude333/rapinav2/pkg/googlesheets"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// Comando "servidor"
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// SSE writer (compatível com progress.SetOutput — usado em handleUpdate)
// ---------------------------------------------------------------------------

/*
type sseWriter struct {
	w  http.ResponseWriter
	w2 *os.File
}

func (sw sseWriter) Write(p []byte) (n int, err error) {
	if sw.w2 != nil {
		_, _ = sw.w2.Write(p)
	}
	n, err = fmt.Fprintf(sw.w, "data: %s\n\n", string(p))
	if f, ok := sw.w.(http.Flusher); ok {
		f.Flush()
	}
	return n, err
}
*/

// ---------------------------------------------------------------------------
// SSE hub — distribui mensagens para clientes em /api/stream
// ---------------------------------------------------------------------------

type sseHub struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

func newSSEHub() *sseHub {
	return &sseHub{clients: make(map[chan string]struct{})}
}

func (h *sseHub) subscribe() chan string {
	ch := make(chan string, 64)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *sseHub) unsubscribe(ch chan string) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
}

func (h *sseHub) broadcast(msg string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

// hubWriter implementa io.Writer e encaminha linhas para o hub SSE
type hubWriter struct {
	hub *sseHub
	buf []byte
}

func (hw *hubWriter) Write(p []byte) (int, error) {
	hw.buf = append(hw.buf, p...)
	for {
		idx := -1
		for i, b := range hw.buf {
			if b == '\n' {
				idx = i
				break
			}
		}
		if idx == -1 {
			break
		}
		line := string(hw.buf[:idx])
		hw.buf = hw.buf[idx+1:]
		if line != "" {
			hw.hub.broadcast(line)
			_, _ = os.Stdout.WriteString(line + "\n")
		}
	}
	return len(p), nil
}

// ---------------------------------------------------------------------------
// Servidor web principal
// ---------------------------------------------------------------------------

func webserver(_ *cobra.Command, _ []string) {
	hub := newSSEHub()
	hw := &hubWriter{hub: hub}

	// Redireciona progress para o hub SSE
	progress.SetOutput(hw)

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/empresas", handleEmpresas)
	http.HandleFunc("/api/relatorios", makeHandleRelatorios(hub))
	http.HandleFunc("/api/files", makeHandleFiles())
	http.HandleFunc("/api/update-db", makeHandleUpdateDB(hub))
	http.HandleFunc("/api/stream", makeHandleStream(hub))

	// Serve arquivos locais gerados em /relatorios/
	fsRelat := http.FileServer(http.Dir(flags.relatorio.outputDir))
	http.Handle("/relatorios/", logHandler(stripPrefixHandler("/relatorios", fsRelat)))

	// Serve assets embutidos (CSS, JS)
	fsAssets := http.FileServer(http.Dir(flags.assetsDir))
	http.Handle("/assets/", logHandler(stripPrefixHandler("/assets", fsAssets)))

	addr := ":" + flags.servidor.porta
	progress.Status("Iniciando servidor em http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// GET /
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	//w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, flags.assetsDir+"/pages/index.html")
}

// GET /api/empresas
func handleEmpresas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	dfp, err := contabil.NewService(db(), flags.tempDir)
	if err != nil {
		jsonError(w, "erro ao criar serviço contábil: "+err.Error(), http.StatusInternalServerError)
		return
	}

	empresas, err := dfp.Empresas()
	if err != nil {
		jsonError(w, "erro ao listar empresas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type empresaJSON struct {
		CNPJ string `json:"cnpj"`
		Nome string `json:"nome"`
	}

	result := make([]empresaJSON, 0, len(empresas))
	for _, e := range empresas {
		result = append(result, empresaJSON{CNPJ: e.CNPJ, Nome: e.Nome})
	}

	writeJSON(w, result)
}

// POST /api/relatorios
func makeHandleRelatorios(hub *sseHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Empresas []rapina.Empresa `json:"empresas"`
			Output   string           `json:"output"` // "local" | "googledrive"
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "corpo inválido: "+err.Error(), http.StatusBadRequest)
			return
		}

		go func() {
			if req.Output == "googledrive" {
				flags.relatorio.googlesheets = true
			}

			dfp, err := contabil.NewService(db(), flags.tempDir)
			if err != nil {
				progress.Error(err)
				hub.broadcast("[done]")
				return
			}

			for _, empresa := range req.Empresas {
				criarRelatórios(empresa, dfp)
			}

			hub.broadcast("[done]")
		}()

		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"iniciado"}`))
	}
}

// GET /api/files
func makeHandleFiles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
			return
		}

		type fileInfo struct {
			Name    string `json:"name"`
			Size    string `json:"size"`
			ModTime string `json:"modTime"`
			IsDir   bool   `json:"isDir"`
			Link    string `json:"link,omitempty"`
		}

		var files []fileInfo

		// Arquivos locais
		if entries, err := os.ReadDir(flags.relatorio.outputDir); err == nil {
			for _, entry := range entries {
				info, err := entry.Info()
				if err != nil {
					continue
				}
				files = append(files, fileInfo{
					Name:    entry.Name(),
					Size:    humanize(info.Size()),
					ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
					IsDir:   entry.IsDir(),
				})
			}
		}

		// Arquivos no Google Drive
		if flags.relatorio.googlesheets {
			cfg := googlesheets.Config{
				CredentialsFile: "credentials.json",
				TokenFile:       "token.json",
				TokenPort:       flags.relatorio.tokenport,
				OAuthURL:        flags.relatorio.oauthurl,
			}
			driveFiles, err := googlesheets.ListFilesInFolder(context.Background(), cfg, flags.relatorio.googlesheetsDir)
			if err != nil {
				progress.ErrorMsg("Erro ao listar arquivos do Google Drive: %v", err)
			} else {
				for _, f := range driveFiles {
					modTime := ""
					if f.ModifiedTime != "" {
						if t, err := time.Parse(time.RFC3339, f.ModifiedTime); err == nil {
							modTime = t.Format("2006-01-02 15:04:05")
						}
					}
					files = append(files, fileInfo{
						Name:    f.Name,
						Size:    humanize(f.Size),
						ModTime: modTime,
						Link:    fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s", f.Id),
					})
				}
			}
		}

		// Ordena por data de modificação (mais recente primeiro)
		sort.Slice(files, func(i, j int) bool {
			ti, err1 := time.Parse("2006-01-02 15:04:05", files[i].ModTime)
			tj, err2 := time.Parse("2006-01-02 15:04:05", files[j].ModTime)
			if err1 != nil || err2 != nil {
				return false
			}
			return tj.Before(ti)
		})

		if len(files) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		writeJSON(w, files)
	}
}

// POST /api/update-db
var alreadyIn = false

func makeHandleUpdateDB(hub *sseHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
			return
		}

		if alreadyIn {
			log.Println("Cancelando update — já em execução")
			_, cancel := context.WithCancel(r.Context())
			cancel()
			return
		}
		alreadyIn = true

		go func() {
			defer func() { alreadyIn = false }()

			dfp, err := contabil.NewService(db(), flags.tempDir)
			if err != nil {
				progress.Error(err)
				hub.broadcast("[done]")
				return
			}

			anof := time.Now().Year()
			anoi := anof - 1

			importar := func(trimestral bool) {
				for ano := anof; ano >= anoi; ano-- {
					if err := dfp.Import(ano, trimestral); err != nil {
						progress.Error(err)
						continue
					}
				}
			}

			importar(false)
			importar(true)

			hub.broadcast("[done]")
		}()

		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"iniciado"}`))
	}
}

// GET /api/stream (SSE)
func makeHandleStream(hub *sseHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		ch := hub.subscribe()
		defer hub.unsubscribe(ch)

		f, ok := w.(http.Flusher)
		flush := func() {
			if ok {
				f.Flush()
			}
		}

		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if msg == "[done]" {
					_, _ = fmt.Fprintf(w, "event: close\ndata: [>] Concluído\n\n")
					flush()
					return
				}
				_, _ = fmt.Fprintf(w, "data: %s\n\n", msg)
				flush()
			case <-ticker.C:
				_, _ = fmt.Fprintf(w, ": keep-alive\n\n")
				flush()
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func logHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func stripPrefixHandler(prefix string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = r.URL.Path[len(prefix):]
		handler.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func humanize(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
