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
	ext "github.com/dude333/rapinav2/pkg/infra"
	"github.com/dude333/rapinav2/pkg/progress"
)

var ErrRepositórioInválido = errors.New("repositório inválido")

type CVM interface {
	Importar(ctx context.Context, ano int, trimestral bool) <-chan dominio.Resultado
}

// DadosContábeis é um serviço que busca os relatórios contábeis de uma empresa
// em vários repositórios (API e BD).
type DadosContábeis struct {
	cvm CVM
	db  *Sqlite
}

func NovoServiço(db *sqlx.DB, tempDir string, force ...bool) (*DadosContábeis, error) {
	dfp := DadosContábeis{}

	repoSqlite, err := NovoSqlite(db)
	if err != nil {
		return &dfp, err
	}

	f := false
	if len(force) > 0 {
		f = force[0]
	}

	hashes, _ := ext.Hashes(db)

	cvmDFP, err := NovoDFP(
		CfgDirDados(tempDir),
		CfgArquivosJáProcessados(hashes),
		CfgForce(f),
	)
	if err != nil {
		return &dfp, err
	}

	return &DadosContábeis{cvm: cvmDFP, db: repoSqlite}, nil
}

// Importar importa os relatórios contábeis no ano especificado e os salva
// no banco de dados.
func (c *DadosContábeis) Importar(ano int, trimestral bool) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	// result retorna o registro após a leitura de cada linha
	// do arquivo importado
	for result := range c.cvm.Importar(ctx, ano, trimestral) {
		if result.Error != nil {
			progress.Error(result.Error)
			continue
		}
		if result.DFP != nil {
			err := c.db.Salvar(ctx, result.DFP)
			if err != nil {
				return err
			}
		}
		if len(result.Hash) > 0 {
			err := ext.SaveHash(c.db.db, result.Hash)
			if err != nil {
				progress.ErrorMsg("erro salvando hash: %v", err)
			}
		}
	}

	return nil
}

/*
func (df *DadosContábeis) Relatório(cnpj string, ano int) (*dominio.DemonstraçãoFinanceira, error) {
	if df.bd == nil {
		return &dominio.DemonstraçãoFinanceira{}, ErrRepositórioInválido
	}
	progress.Debug("Ler(%s, %d)", cnpj, ano)
	dfp, err := df.bd.Ler(context.Background(), cnpj, ano)
	return dfp, err
}
*/

func (c *DadosContábeis) DadosTrimestrais(cnpj string, consolidado bool) ([]rapina.InformeTrimestral, error) {
	if c.db == nil {
		return nil, ErrRepositórioInválido
	}
	return c.db.Trimestral(context.Background(), cnpj, consolidado)
}

func (c *DadosContábeis) Empresas() ([]rapina.Empresa, error) {
	if c.db == nil {
		return []rapina.Empresa{}, ErrRepositórioInválido
	}
	return c.db.Empresas(context.Background())
}

func (c *DadosContábeis) BuscaEmpresas(nome string) ([]rapina.Empresa, error) {
	if c.db == nil {
		return []rapina.Empresa{}, ErrRepositórioInválido
	}
	progress.Debug("Empresas(%s)", nome)
	return c.db.BuscaEmpresas(context.Background(), nome)
}
