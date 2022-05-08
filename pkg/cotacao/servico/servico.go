// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

// serviço de cotações que contém a lógica de como os dados são capturados,
// armazenados (pelos repositórios) e disponibilizados para o domínio.
//
//     .---------------------.
//     |   Domínio: Cotação  |
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
	"time"

	rapina "github.com/dude333/rapinav2"
	domínio "github.com/dude333/rapinav2/pkg/cotacao"
)

type Importação interface {
	Importar(ctx context.Context, dia rapina.Data) <-chan domínio.Resultado
}

type Leitura interface {
	Cotação(ctx context.Context, código string, data rapina.Data) (*domínio.Ativo, error)
}

type Escrita interface {
	Salvar(ctx context.Context, ativo *domínio.Ativo) error
}

type LeituraEscrita interface {
	Leitura
	Escrita
}

// Serviço é um serviço que implementa Importação e busca
// cotações de um Ativo em vários repositórios (API e BD).
type Serviço struct {
	api []Importação
	bd  LeituraEscrita
}

func NovoServiço(
	api []Importação,
	bd LeituraEscrita) *Serviço {
	return &Serviço{
		api: api,
		bd:  bd,
	}
}

// Cotação busca a cotação de um ativo em vários repositórios com base
// no "código" de um ativo de um determinado "dia", retornando o primeiro
// valor encontado ou o erro de todos os repositórios. Caso a cotação seja
// encontrada via API, ela será armazenada no bando de dados para agilizar a
// próxima leitura do mesmo código, na mesma data.
func (a *Serviço) Cotação(código string, dia rapina.Data) (*domínio.Ativo, error) {
	atv, err := a.cotaçãoBD(código, dia)
	if err != nil {
		return a.cotaçãoAPI(código, dia)
	}
	return atv, err
}

func (a *Serviço) cotaçãoBD(código string, dia rapina.Data) (*domínio.Ativo, error) {
	if a.bd == nil {
		return nil, ErrRepositórioInválido
	}

	atv, err := a.bd.Cotação(context.Background(), código, dia)
	if err == nil {
		return atv, nil
	}

	return nil, ErrCotaçãoNãoEncontrada
}

func (a *Serviço) cotaçãoAPI(código string, dia rapina.Data) (*domínio.Ativo, error) {
	if len(a.api) < 1 {
		return nil, ErrRepositórioInválido
	}

	var atv *domínio.Ativo
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Tentativa de coletar a cotação usando vários servidores de API
	for i := range a.api {
		// result retorna o registro após a leitura de cada linha
		// do arquivo importado
		for result := range a.api[i].Importar(ctx, dia) {
			if result.Error != nil {
				return nil, result.Error
			}
			if result.Ativo.Código == código {
				atv = result.Ativo
			}
			if a.bd != nil {
				_ = a.bd.Salvar(ctx, result.Ativo)
			}
		}
		// Finaliza se ativo já tiver sido encontrado
		if atv != nil {
			break
		}
	}

	if atv == nil {
		return nil, ErrCotaçãoNãoEncontrada
	}

	return atv, nil
}
