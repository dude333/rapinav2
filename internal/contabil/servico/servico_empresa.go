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
package servico

import (
	"context"
	"time"

	rapina "github.com/dude333/rapinav2/internal"
	"github.com/dude333/rapinav2/pkg/progress"
)

type Importação interface {
	Importar(ctx context.Context, ano int, trimestral bool) <-chan rapina.Resultado
}

type Leitura interface {
	Ler(ctx context.Context, cnpj string, ano int) (*rapina.Empresa, error)
	Empresas(ctx context.Context, nome string) []string
}

type Escrita interface {
	Salvar(ctx context.Context, empresa *rapina.Empresa) error
}

type LeituraEscrita interface {
	Leitura
	Escrita
}

// Empresa é um serviço que busca os relatórios contábeis de uma empresa
// em vários repositórios (API e BD).
type Empresa struct {
	api Importação
	bd  LeituraEscrita
}

func NovoSvcEmpresa(api Importação, bd LeituraEscrita) *Empresa {
	return &Empresa{api: api, bd: bd}
}

// Importar importa os relatórios contábeis no ano especificado e os salva
// no banco de dados.
func (e *Empresa) Importar(ano int, trimestral bool) error {
	if e.api == nil {
		return ErrRepositórioInválido
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// result retorna o registro após a leitura de cada linha
	// do arquivo importado
	for result := range e.api.Importar(ctx, ano, trimestral) {
		if result.Error != nil {
			return result.Error
		}
		if e.bd != nil {
			_ = e.bd.Salvar(context.Background(), result.Empresa)
		}
	}

	return nil
}

func (e *Empresa) Relatório(cnpj string, ano int) (*rapina.Empresa, error) {
	if e.bd == nil {
		return &rapina.Empresa{}, ErrRepositórioInválido
	}
	progress.Debug("Ler(%s, %d)", cnpj, ano)
	dfp, err := e.bd.Ler(context.Background(), cnpj, ano)
	return dfp, err
}

func (e *Empresa) Empresas(nome string) []string {
	if e.bd == nil {
		return []string{}
	}
	progress.Debug("Empresas(%s)", nome)
	lista := e.bd.Empresas(context.Background(), nome)
	return lista
}
