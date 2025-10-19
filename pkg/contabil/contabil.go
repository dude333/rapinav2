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
	Import(ctx context.Context, ano int, trimestral bool) <-chan dominio.ImportResult
}

// ContabilServices contém os serviços que implementam a interface CVM de
// importação de dados da CVM.
type ContabilServices struct {
	cvmDFP CVM
	cvmFRE CVM
	db     *Sqlite
}

func NewService(db *sqlx.DB, tempDir string, force ...bool) (*ContabilServices, error) {
	dfp := ContabilServices{}

	repoSqlite, err := NewSqlite(db)
	if err != nil {
		return &dfp, err
	}

	f := false
	if len(force) > 0 {
		f = force[0]
	}

	hashes, _ := ext.Hashes(db)

	cvmDFP, err := NewDFPImporter(
		WithDataDir(tempDir),
		WithProcessedHashes(hashes),
		WithForce(f),
	)
	if err != nil {
		return &dfp, err
	}

	cvmFRE, err := NovaFRE(
		WithDataDir(tempDir),
		WithProcessedHashes(hashes),
		WithForce(f),
	)
	if err != nil {
		return &dfp, err
	}

	return &ContabilServices{cvmDFP: cvmDFP, cvmFRE: cvmFRE, db: repoSqlite}, nil
}

// Import importa os relatórios contábeis no ano especificado e os salva
// no banco de dados.
func (c *ContabilServices) Import(ano int, trimestral bool) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	// importa dfp/itr
	for result := range c.cvmDFP.Import(ctx, ano, trimestral) {
		if result.Error != nil {
			progress.Error(result.Error)
			continue
		}
		if result.DFP != nil {
			err := c.db.Save(ctx, result.DFP)
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

	// importa fre
	for result := range c.cvmFRE.Import(ctx, ano, false) {
		if result.Error != nil {
			progress.Error(result.Error)
			continue
		}
		if result.FRE != nil {
			err := c.db.SaveFRE(ctx, result.FRE)
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

func (c *ContabilServices) DadosTrimestrais(cnpj string, consolidado bool) ([]rapina.InformeTrimestral, error) {
	if c.db == nil {
		return nil, ErrRepositórioInválido
	}
	return c.db.Trimestral(context.Background(), cnpj, consolidado)
}

func (c *ContabilServices) Empresas() ([]rapina.Empresa, error) {
	if c.db == nil {
		return []rapina.Empresa{}, ErrRepositórioInválido
	}
	return c.db.Empresas(context.Background())
}

func (c *ContabilServices) BuscaEmpresas(nome string) ([]rapina.Empresa, error) {
	if c.db == nil {
		return []rapina.Empresa{}, ErrRepositórioInválido
	}
	progress.Debug("Empresas(%s)", nome)
	return c.db.BuscaEmpresas(context.Background(), nome)
}
