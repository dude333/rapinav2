// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

import (
	"path"

	ext "github.com/dude333/rapinav2/pkg/infra"
)

// infra define uma interface para que este respositório não fique amarrado
// na implementação de uma única biblioteca externa.
type infra interface {
	DownloadAndUnzip(url, zip string, filtros []string) ([]Arquivo, error)
	Cleanup(files []Arquivo) []string
}

type Arquivo struct {
	path string
	hash string
}

type localInfra struct {
	dirDados string // diretório de dados
}

func (l localInfra) DownloadAndUnzip(url, arquivo string, filtros []string) ([]Arquivo, error) {
	zip := path.Join(l.dirDados, arquivo)
	arqs, err := ext.DownloadAndUnzip(url, zip, filtros)
	if err != nil {
		return []Arquivo{}, err
	}

	arquivos := make([]Arquivo, len(arqs))
	for i := range arqs {
		h, _ := ext.FileHash(arqs[i])
		arquivos[i] = Arquivo{
			path: arqs[i],
			hash: h,
		}
	}

	return arquivos, nil
}

func (l localInfra) Cleanup(arqs []Arquivo) []string {
	files := make([]string, len(arqs))
	for i := range arqs {
		files[i] = arqs[i].path
	}
	return ext.Cleanup(files)
}
