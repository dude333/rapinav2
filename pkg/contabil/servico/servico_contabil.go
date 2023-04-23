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

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil"
	"github.com/dude333/rapinav2/pkg/contabil/repositorio"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
)

var (
	ErrRepositórioInválido = errors.New("repositório inválido")
)

type Importação interface {
	Importar(ctx context.Context, ano int, trimestral bool) <-chan contabil.Resultado
}

type Leitura interface {
	Ler(ctx context.Context, cnpj string, ano int) (*contabil.DemonstraçãoFinanceira, error)
	Empresas(ctx context.Context, nome string) []rapina.Empresa
	Hashes() []string
}

type Escrita interface {
	Salvar(ctx context.Context, empresa *contabil.DemonstraçãoFinanceira) error
	SalvarHash(ctx context.Context, hash string) error
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
		if len(result.Hash) > 0 {
			err := e.bd.SalvarHash(ctx, result.Hash)
			if err != nil {
				progress.ErrorMsg("erro salvando hash: %v", err)
			}
		}
		if result.Empresa != nil {
			err := e.bd.Salvar(ctx, result.Empresa)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *DemonstraçãoFinanceira) Relatório(cnpj string, ano int) (*contabil.DemonstraçãoFinanceira, error) {
	if e.bd == nil {
		return &contabil.DemonstraçãoFinanceira{}, ErrRepositórioInválido
	}
	progress.Debug("Ler(%s, %d)", cnpj, ano)
	dfp, err := e.bd.Ler(context.Background(), cnpj, ano)
	return dfp, err
}

func (e *DemonstraçãoFinanceira) Empresas(nome string) []rapina.Empresa {
	if e.bd == nil {
		return []rapina.Empresa{}
	}
	progress.Debug("Empresas(%s)", nome)
	lista := e.bd.Empresas(context.Background(), nome)
	return lista
}
