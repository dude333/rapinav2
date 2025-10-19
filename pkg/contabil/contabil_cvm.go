// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package contabil

import (
	"context"
	"net/url"
	"os"
	"path"

	"github.com/dude333/rapinav2/pkg/contabil/dominio"
	ext "github.com/dude333/rapinav2/pkg/infra"
	"github.com/dude333/rapinav2/pkg/progress"
)

// infra define uma interface para que este respositório não fique amarrado
// na implementação de uma única biblioteca externa.
type infra interface {
	DownloadAndUnzip(url string, filtros []string) ([]Arquivo, string, error)
	Cleanup(files []Arquivo) []string
}

type Arquivo struct {
	path string
	hash string
}

type LocalInfra struct {
	dirDados string // diretório de dados
}

// DownloadAndUnzip baixa e descompacta o arquivo e retorna a lista de arquivos descompactados, com
// o hash de cada arquivo; retorna o hash do .zip e o erro, se houver.
func (l LocalInfra) DownloadAndUnzip(urlString string, filtros []string) ([]Arquivo, string, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return []Arquivo{}, "", err
	}
	arquivo := path.Base(u.Path)
	zip := path.Join(l.dirDados, arquivo)
	arqs, err := ext.DownloadAndUnzip(urlString, zip, filtros)
	if err != nil {
		return []Arquivo{}, "", err
	}

	arquivos := make([]Arquivo, len(arqs))
	for i := range arqs {
		h, _ := ext.FileHash(arqs[i])
		arquivos[i] = Arquivo{
			path: arqs[i],
			hash: h,
		}
	}

	zipHash, err := ext.FileHash(zip)
	os.Remove(zip)

	return arquivos, zipHash, err
}

func (l LocalInfra) Cleanup(arqs []Arquivo) []string {
	files := make([]string, len(arqs))
	for i := range arqs {
		files[i] = arqs[i].path
	}
	return ext.Cleanup(files)
}

// CVMImporter é um importador genérico de dados da CVM.
type CVMImporter struct {
	nome          string
	cfg           *cfg
	infra         infra
	url           func(ano int, trimestral bool) string
	filtros       func() []string
	processar     func(ctx context.Context, arq Arquivo, results chan<- dominio.ImportResult) error
	reportarAviso bool
}

// Import baixa o arquivo de dados de todas as empresas de um determinado
// ano do site da CVM.
func (c *CVMImporter) Import(ctx context.Context, ano int, trimestral bool) <-chan dominio.ImportResult {
	results := make(chan dominio.ImportResult)

	go func() {
		defer close(results)

		url := c.url(ano, trimestral)

		arquivos, zipHash, err := c.infra.DownloadAndUnzip(url, c.filtros())
		if err != nil {
			results <- dominio.ImportResult{Error: err}
			return
		}

		defer c.infra.Cleanup(arquivos)

		if c.existe(zipHash) {
			if c.reportarAviso {
				progress.Warning("Este arquivo '%s' já foi processado anteriormente", c.nome)
			}
			return
		}

		for _, arquivo := range arquivos {
			progress.Running(arquivo.path)

			// Processa o arquivo e envia o resultado para o canal 'results'
			err = c.processar(ctx, arquivo, results)
			if err != nil {
				results <- dominio.ImportResult{Hash: arquivo.hash}
			}

			progress.RunOK()
		}

		// Grava o hash do zip no banco de dados
		results <- dominio.ImportResult{Hash: zipHash}
	}()

	return results
}

func (c CVMImporter) existe(hash string) bool {
	if len(hash) == 0 || c.cfg.force {
		return false
	}
	for i := range c.cfg.processedHashes {
		if c.cfg.processedHashes[i] == hash {
			return true
		}
	}
	return false
}
