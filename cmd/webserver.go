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

	fs := http.FileServer(http.Dir("./cmd/templates/"))
	http.Handle("/assets/", logHandler(fs))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//go:embed templates
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

	t, err := template.ParseFS(indexHTML, "templates/empresas.html")
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
	progress.Debug("obj: %v", objs)
	if err := json.Unmarshal([]byte(objs[0]), &selectedEmpresas); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	for _, opt := range selectedEmpresas {
		fmt.Fprintf(w, "CNPJ: %s, Name: %s <br />\n", opt.CNPJ, opt.Nome)
	}

	// fmt.Fprintf(w, "Empresa(s) selecionada(s): %#v", selectedEmpresas)
	progress.Debug("Empresa(s) selecionada(s): %+v", selectedEmpresas)

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
