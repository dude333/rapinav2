// SPDX-FileCopyrightText: 2024 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

// cvm.go:
//  1. Preparação de arquivos:
//     input: zip
//     output: csv
//  2. Parsing:
//     input: csv
//     output: struct
//  3. Storage:
//     input: struct
//     output: tabela
//  4. Load:
//     input: tabela
//     output: struct
//
// Uso:
// - Importar (preparação de arquivos, parsing, storage)
//   . importar por tipo, anual ou trimestral
//   . tipos diferentes podem ser armazenados na mesma tabela
// - Carregar (load)

package rapina

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
)

type CVM struct {
	db       *sqlx.DB
	dirDados string   // Diretório de dados temporários
	hashes   []string // Hashes dos arquivos já processados
	force    bool     // Forçar atualização, mesmo que o dado já exista
}

func NovaCVM(db *sqlx.DB, tempDir string, force bool) (*CVM, error) {
	c := &CVM{
		db:       db,
		dirDados: tempDir,
		force:    force,
	}
	if tempDir == "" {
		c.dirDados = os.TempDir()
	} else {
		err := os.MkdirAll(c.dirDados, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	return c, nil
}

type CvmType int

const (
	CvmDfp CvmType = iota
	CvmFre
)

type CvmDataSource interface {
	Salvar(ctx context.Context, db *sqlx.DB) error
}

// ImportarCVMimporta dados da CVM para a base de dados local
func (c *CVM) Importar(ctx context.Context, ano int, trimestral bool) error {
	for _, tipo := range []CvmType{CvmDfp, CvmFre} {
		chDados, err := c.importarTipo(ctx, tipo, ano, trimestral)
		if err != nil {
			return err
		}
		for dado := range chDados {
			err := dado.(CvmDataSource).Salvar(ctx, c.db)
			if err != nil {
				progress.Error(err)
			}
		}
	}
	return nil
}

// importarTipo prepara os dados e os carrega na struct correspondente
func (c *CVM) importarTipo(ctx context.Context, tipo CvmType, ano int, trimestral bool) (<-chan interface{}, error) {
	switch tipo {
	case CvmDfp:
		return c.importarDFP(ctx, ano, trimestral)
	case CvmFre:
		return c.importarFRE(ctx, ano, trimestral)
	}
	return nil, fmt.Errorf("tipo %d inválido", tipo)
}

// DFP ------------------------------------------------------------------------

func (c *CVM) importarDFP(ctx context.Context, ano int, trimestre bool) (<-chan interface{}, error) {
	ch := make(chan interface{})
	go func() {
		defer close(ch)
		for i := 1; i <= 10; i++ {
			ch <- &DFP{DataDFP: fmt.Sprintf("DFP %d", i)}
			time.Sleep(time.Second)
		}
	}()
	return ch, nil
}

type DFP struct {
	err error

	DataDFP string
}

func (dfp *DFP) Salvar(ctx context.Context, db *sqlx.DB) error {
	progress.Status("DFP: %s", dfp.DataDFP)
	return nil
}

// FRE ------------------------------------------------------------------------

func (c *CVM) importarFRE(ctx context.Context, ano int, trimestre bool) (<-chan interface{}, error) {
	ch := make(chan interface{})
	go func() {
		defer close(ch)
		for i := 1; i <= 10; i++ {
			ch <- &FRE{}
			time.Sleep(time.Second)
		}
	}()
	return ch, nil
}

type FRE struct{}

func (fre *FRE) Salvar(ctx context.Context, db *sqlx.DB) error {
	// Salvar FRE no bando de dados
	progress.Status("FRE")
	return nil
}
