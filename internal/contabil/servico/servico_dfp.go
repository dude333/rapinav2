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
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
	"github.com/dude333/rapinav2/pkg/progress"
	"time"
)

// serviço é um serviço que implementa RepositórioImportação e
// busca os relatórios contábeis de uma serviço em vários repositórios (API e BD).
type serviço struct {
	api contábil.RepositórioImportação
	bd  contábil.RepositórioLeituraEscrita
}

func NovoServiço(
	api contábil.RepositórioImportação,
	bd contábil.RepositórioLeituraEscrita) contábil.Serviço {

	return &serviço{api: api, bd: bd}
}

// Importar importa os relatórios contábeis no ano especificado e os salva
// no banco de dados.
func (s *serviço) Importar(ano int) error {
	if s.api == nil {
		return ErrRepositórioInválido
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// result retorna o registro após a leitura de cada linha
	// do arquivo importado
	for result := range s.api.Importar(ctx, ano) {
		if result.Error != nil {
			return result.Error
		}
		if s.bd != nil {
			_ = s.bd.Salvar(context.Background(), result.Empresa)
		}
	}

	return nil
}

func (s *serviço) Relatório(cnpj string, ano int) (*contábil.Empresa, error) {
	if s.bd == nil {
		return &contábil.Empresa{}, ErrRepositórioInválido
	}
	progress.Debug("Ler(%s, %d)", cnpj, ano)
	dfp, err := s.bd.Ler(context.Background(), cnpj, ano)
	return dfp, err
}

func (s *serviço) Empresas(nome string) []string {
	if s.bd == nil {
		return []string{}
	}
	progress.Debug("Empresas(%s)", nome)
	lista := s.bd.Empresas(context.Background(), nome)
	return lista
}
