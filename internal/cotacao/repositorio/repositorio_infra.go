// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositório

import (
	"path"

	ext "github.com/dude333/rapinav2/pkg/infra"
)

// infra define uma interface para que este respositório não fique amarrado
// na implementação de uma única biblioteca externa.
type infra interface {
	DownloadAndUnzip(url, zip string, filtros []string) ([]string, error)
	Cleanup(files []string) []string
}

type localInfra struct {
	dirDados string // diretório de dados
}

func (l localInfra) DownloadAndUnzip(url, arquivo string, filtros []string) ([]string, error) {
	zip := path.Join(l.dirDados, arquivo)
	return ext.DownloadAndUnzip(url, zip, filtros)
}

func (l localInfra) Cleanup(files []string) []string {
	return ext.Cleanup(files)
}
