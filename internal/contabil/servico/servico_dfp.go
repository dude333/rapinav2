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

// dfp é um serviço que implementa RepositórioImportaçãoDFP e
// busca os relatórios contábeis de uma empresa em vários repositórios (API e BD).
type dfp struct {
	api contábil.RepositórioImportaçãoDFP
	bd  contábil.RepositórioLeituraEscritaDFP
}

func NovoDFP(
	api contábil.RepositórioImportaçãoDFP,
	bd contábil.RepositórioLeituraEscritaDFP) contábil.ServiçoDFP {
	return &dfp{api: api, bd: bd}
}

// Importar importa os relatórios contábeis no ano especificado e os salva
// no banco de dados.
func (r *dfp) Importar(ano int) error {
	if r.api == nil {
		return ErrRepositórioInválido
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// result retorna o registro após a leitura de cada linha
	// do arquivo importado
	for result := range r.api.Importar(ctx, ano) {
		if result.Error != nil {
			return result.Error
		}
		if r.bd != nil {
			_ = r.bd.Salvar(context.Background(), result.DFP)
		}
	}

	return nil
}

func (r *dfp) Relatório(cnpj string, ano int) (*contábil.DFP, error) {
	if r.bd == nil {
		return &contábil.DFP{}, ErrRepositórioInválido
	}
	progress.Debug("Ler(%s, %d)", cnpj, ano)
	dfp, err := r.bd.Ler(context.Background(), cnpj, ano)
	return dfp, err
}
