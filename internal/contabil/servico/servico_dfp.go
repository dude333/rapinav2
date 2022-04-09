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

	"github.com/dude333/rapinav2/internal/contabil/dominio"
	"github.com/dude333/rapinav2/pkg/progress"
)

type Importação interface {
	Importar(ctx context.Context, ano int, trimestral bool) <-chan dominio.Resultado
}

type Leitura interface {
	Ler(ctx context.Context, cnpj string, ano int) (*dominio.Empresa, error)
	Empresas(ctx context.Context, nome string) []string
}

type Escrita interface {
	Salvar(ctx context.Context, empresa *dominio.Empresa) error
}

type LeituraEscrita interface {
	Leitura
	Escrita
}

// serviço é um serviço que implementa RepositórioImportação e
// busca os relatórios contábeis de uma serviço em vários repositórios (API e BD).
type Serviço struct {
	api Importação
	bd  LeituraEscrita
}

func NovoServiço(
	api Importação,
	bd LeituraEscrita) *Serviço {

	return &Serviço{api: api, bd: bd}
}

// Importar importa os relatórios contábeis no ano especificado e os salva
// no banco de dados.
func (s *Serviço) Importar(ano int, trimestral bool) error {
	if s.api == nil {
		return ErrRepositórioInválido
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// result retorna o registro após a leitura de cada linha
	// do arquivo importado
	for result := range s.api.Importar(ctx, ano, trimestral) {
		if result.Error != nil {
			return result.Error
		}
		if s.bd != nil {
			_ = s.bd.Salvar(context.Background(), result.Empresa)
		}
	}

	return nil
}

func (s *Serviço) Relatório(cnpj string, ano int) (*dominio.Empresa, error) {
	if s.bd == nil {
		return &dominio.Empresa{}, ErrRepositórioInválido
	}
	progress.Debug("Ler(%s, %d)", cnpj, ano)
	dfp, err := s.bd.Ler(context.Background(), cnpj, ano)
	return dfp, err
}

func (s *Serviço) Empresas(nome string) []string {
	if s.bd == nil {
		return []string{}
	}
	progress.Debug("Empresas(%s)", nome)
	lista := s.bd.Empresas(context.Background(), nome)
	return lista
}
