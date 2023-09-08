// SPDX-FileCopyrightText: 2023 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"

	rapina "github.com/dude333/rapinav2"
)

type noBellStdout struct{}

func (n *noBellStdout) Write(p []byte) (int, error) {
	if len(p) == 1 && p[0] == readline.CharBell {
		return 0, nil
	}
	return readline.Stdout.Write(p)
}

func (n *noBellStdout) Close() error {
	return readline.Stdout.Close()
}

var NoBellStdout = &noBellStdout{}

func escolherEmpresa(empresas []rapina.Empresa) (rapina.Empresa, bool) {
	// The Active and Selected templates set a small pepper icon next to the name colored and the heat unit for the
	// active template. The details template is show at the bottom of the select's list and displays the full info
	// for that pepper in a multi-line template.
	templates := &promptui.SelectTemplates{
		Help: `{{ "Para navegar:" | faint }} [{{ .NextKey | faint }} ` +
			`{{ .PrevKey | faint }} {{ .PageUpKey | faint }} {{ .PageDownKey | faint }}]` +
			`{{ if .Search }} {{ "Para procurar:" | faint }} [{{ .SearchKey | faint }}]{{ end }}` +
			` Para sair: [Ctrl-c]`,
		Label:    "{{ . }}:",
		Active:   " > {{ .Nome | red }}",
		Inactive: "  {{ .Nome | blue }}",
		Selected: " > {{ .Nome | red | cyan }}",
		Details: `
--------------------------------------
| {{ "Name:" | bold }}	{{ .Nome }}
| {{ "CNPJ:" | faint }}	{{ .CNPJ }}
------------------------------------------`,
	}

	// A searcher function is implemented which enabled the search mode for the select. The function follows
	// the required searcher signature and finds any pepper whose name contains the searched string.
	searcher := func(input string, index int) bool {
		empresa := empresas[index]
		nome := rapina.NormalizeString(empresa.Nome)
		ninput := rapina.NormalizeString(input)

		return strings.Contains(nome, ninput)
	}

	prompt := promptui.Select{
		Label:     "Empresas",
		Items:     empresas,
		Templates: templates,
		Size:      15,
		Searcher:  searcher,
		Stdout:    NoBellStdout,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return rapina.Empresa{}, false
	}

	return empresas[i], true
}

// prepareFilename cleans up the filename and returns the path/filename
func prepareFilename(path, name string) (fpath string, err error) {
	clean := func(r rune) rune {
		switch r {
		case ' ', ',', '/', '\\':
			return '_'
		}
		return r
	}
	path = strings.TrimSuffix(path, "/")
	name = strings.TrimSuffix(name, ".")
	name = strings.Map(clean, name)
	fpath = filepath.FromSlash(path + "/" + name + ".xlsx")

	const max = 50
	var x int
	for x = 1; x <= max; x++ {
		_, err = os.Stat(fpath)
		if err == nil {
			// File exists, try again with another name
			fpath = fmt.Sprintf("%s/%s(%d).xlsx", path, name, x)
		} else if os.IsNotExist(err) {
			err = nil // reset error
			break
		} else {
			err = fmt.Errorf("file %s stat error: %v", fpath, err)
			return
		}
	}

	if x > max {
		err = fmt.Errorf("remova o arquivo %s/%s.xlsx antes de continuar", path, name)
		return
	}

	// Create directory
	_ = os.Mkdir(path, os.ModePerm)

	// Check if the directory was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.Wrap(err, "diretório não pode ser criado")
	}

	return
}
