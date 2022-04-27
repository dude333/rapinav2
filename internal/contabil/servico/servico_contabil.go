// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

// serviço de relatórios contábeis que contém a lógica de como os dados são
// capturados, armazenados (pelos repositórios) e disponibilizados para o domínio.
//
//     .---------------------.
//     |   Domínio: Contábil |
//     '---------------------'
//                 ↓
//     .---------------------.
//     |      *Serviço*      |
//     '---------------------'
//          ↓             ↓
//     .---------    .-------.
//     |   API  |    |  BD   |   <= Repositórios
//     '--------'    '-------'
//
package serviço

import (
	"context"
	"errors"
	"time"

	contábil "github.com/dude333/rapinav2/internal/contabil"
	"github.com/dude333/rapinav2/internal/contabil/repositorio"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
)

var (
	ErrRepositórioInválido = errors.New("repositório inválido")
)

type Importação interface {
	Importar(ctx context.Context, ano int, trimestral bool) <-chan contábil.Resultado
}

type Leitura interface {
	Ler(ctx context.Context, cnpj string, ano int) (*contábil.DemonstraçãoFinanceira, error)
	Empresas(ctx context.Context, nome string) []string
}

type Escrita interface {
	Salvar(ctx context.Context, empresa *contábil.DemonstraçãoFinanceira) error
}

type LeituraEscrita interface {
	Leitura
	Escrita
}

// DemonstraçãoFinanceira é um serviço que busca os relatórios contábeis de uma empresa
// em vários repositórios (API e BD).
type DemonstraçãoFinanceira struct {
	api Importação
	bd  LeituraEscrita
}

func NovoDemonstraçãoFinanceira(db *sqlx.DB, tempDir string) (*DemonstraçãoFinanceira, error) {
	dfp := DemonstraçãoFinanceira{}

	repoCVM, err := repositorio.NovoCVM(repositorio.CfgDirDados(tempDir))
	if err != nil {
		return &dfp, err
	}

	repoSqlite, err := repositorio.NovoSqlite(db)
	if err != nil {
		return &dfp, err
	}

	return novoDemonstraçãoFinanceira(repoCVM, repoSqlite)
}

func novoDemonstraçãoFinanceira(api Importação, bd LeituraEscrita) (*DemonstraçãoFinanceira, error) {
	if api == nil || bd == nil {
		return &DemonstraçãoFinanceira{}, ErrRepositórioInválido
	}

	return &DemonstraçãoFinanceira{api: api, bd: bd}, nil
}

// Importar importa os relatórios contábeis no ano especificado e os salva
// no banco de dados.
func (e *DemonstraçãoFinanceira) Importar(ano int, trimestral bool) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	// result retorna o registro após a leitura de cada linha
	// do arquivo importado
	for result := range e.api.Importar(ctx, ano, trimestral) {
		if result.Error != nil {
			progress.Error(result.Error)
			continue
		}
		err := e.bd.Salvar(ctx, result.Empresa)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *DemonstraçãoFinanceira) Relatório(cnpj string, ano int) (*contábil.DemonstraçãoFinanceira, error) {
	if e.bd == nil {
		return &contábil.DemonstraçãoFinanceira{}, ErrRepositórioInválido
	}
	progress.Debug("Ler(%s, %d)", cnpj, ano)
	dfp, err := e.bd.Ler(context.Background(), cnpj, ano)
	return dfp, err
}

func (e *DemonstraçãoFinanceira) Empresas(nome string) []string {
	if e.bd == nil {
		return []string{}
	}
	progress.Debug("Empresas(%s)", nome)
	lista := e.bd.Empresas(context.Background(), nome)
	return lista
}
