// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package contabil

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil/dominio"
	"github.com/dude333/rapinav2/pkg/contabil/repositorio"
	"github.com/dude333/rapinav2/pkg/progress"
)

var ErrRepositórioInválido = errors.New("repositório inválido")

type Importação interface {
	Importar(ctx context.Context, ano int, trimestral bool) <-chan dominio.Resultado
}

type Leitura interface {
	Ler(ctx context.Context, cnpj string, ano int) (*dominio.DemonstraçãoFinanceira, error)
	Empresas(ctx context.Context, nome string) []rapina.Empresa
	Hashes() []string
}

type Escrita interface {
	Salvar(ctx context.Context, empresa *dominio.DemonstraçãoFinanceira) error
	SalvarHash(ctx context.Context, hash string) error
}

type LeituraEscrita interface {
	Leitura
	Escrita
}

// DemonstraçãoFinanceira é um serviço que busca os relatórios contábeis de uma empresa
// em vários repositórios (API e BD).
type DemonstraçãoFinanceira struct {
	api *repositorio.CVM
	bd  *repositorio.Sqlite
}

func NovaDemonstraçãoFinanceira(db *sqlx.DB, tempDir string) (*DemonstraçãoFinanceira, error) {
	dfp := DemonstraçãoFinanceira{}

	repoSqlite, err := repositorio.NovoSqlite(db)
	if err != nil {
		return &dfp, err
	}

	repoCVM, err := repositorio.NovoCVM(
		repositorio.CfgDirDados(tempDir),
		repositorio.CfgArquivosJáProcessados(repoSqlite.Hashes()),
	)
	if err != nil {
		return &dfp, err
	}

	return &DemonstraçãoFinanceira{api: repoCVM, bd: repoSqlite}, nil
}

// Importar importa os relatórios contábeis no ano especificado e os salva
// no banco de dados.
func (df *DemonstraçãoFinanceira) Importar(ano int, trimestral bool) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	// result retorna o registro após a leitura de cada linha
	// do arquivo importado
	for result := range df.api.Importar(ctx, ano, trimestral) {
		if result.Error != nil {
			progress.Error(result.Error)
			continue
		}
		if result.Empresa != nil {
			err := df.bd.Salvar(ctx, result.Empresa)
			if err != nil {
				return err
			}
		}
		if len(result.Hash) > 0 {
			err := df.bd.SalvarHash(ctx, result.Hash)
			if err != nil {
				progress.ErrorMsg("erro salvando hash: %v", err)
			}
		}
	}

	return nil
}

func (df *DemonstraçãoFinanceira) Relatório(cnpj string, ano int) (*dominio.DemonstraçãoFinanceira, error) {
	if df.bd == nil {
		return &dominio.DemonstraçãoFinanceira{}, ErrRepositórioInválido
	}
	progress.Debug("Ler(%s, %d)", cnpj, ano)
	dfp, err := df.bd.Ler(context.Background(), cnpj, ano)
	return dfp, err
}

func (df *DemonstraçãoFinanceira) RelatórioTrimestal(cnpj string, consolidado bool) ([]rapina.InformeTrimestral, error) {
	if df.bd == nil {
		return nil, ErrRepositórioInválido
	}
	return df.bd.Trimestral(context.Background(), cnpj, consolidado)
}

func (df *DemonstraçãoFinanceira) Empresas() ([]rapina.Empresa, error) {
	if df.bd == nil {
		return []rapina.Empresa{}, ErrRepositórioInválido
	}
	return df.bd.Empresas(context.Background())
}

func (df *DemonstraçãoFinanceira) BuscaEmpresas(nome string) ([]rapina.Empresa, error) {
	if df.bd == nil {
		return []rapina.Empresa{}, ErrRepositórioInválido
	}
	progress.Debug("Empresas(%s)", nome)
	return df.bd.BuscaEmpresas(context.Background(), nome)
}
