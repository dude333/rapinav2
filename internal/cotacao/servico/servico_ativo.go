// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

// serviço define os casos de uso de um "ativo", como leitura da sua cotação em
// um repositório (banco de dados ou API).
package serviço

import (
	"context"
	domínio "github.com/dude333/rapinav2/internal/cotacao/dominio"
)

// ativo é um serviço de busca de informações de um ativo em vários
// repositórios, como banco de dados ou via API.
type ativo struct {
	bd  domínio.RepositórioLeituraEscritaAtivo
	api []domínio.RepositórioLeituraAtivo
}

func Novo(
	bd domínio.RepositórioLeituraEscritaAtivo,
	api []domínio.RepositórioLeituraAtivo) domínio.ServiçoAtivo {
	return &ativo{
		bd:  bd,
		api: api,
	}
}

// Cotação busca a cotação de um ativo em vários repositórios com base
// no "código" de um ativo de um determinado "dia", retornando o primeiro
// valor encontado ou o erro de todos os repositórios. Caso a cotação seja
// encontrada via API, ela será armazenada no bando de dados para agilizar a
// próxima leitura do mesmo código, na mesma data.
func (a *ativo) Cotação(código string, dia domínio.Data) (*domínio.Ativo, error) {
	atv, err := a.cotação(código, dia)
	if err != nil {
		return &domínio.Ativo{}, err
	}

	return atv, nil
}

func (a *ativo) cotação(código string, dia domínio.Data) (*domínio.Ativo, error) {
	if a.bd != nil {
		atv, err := a.bd.Cotação(context.Background(), código, dia)
		if err == nil {
			return atv, nil
		}
	}

	for i := range a.api {
		atv, err := a.api[i].Cotação(context.Background(), código, dia)
		if err == nil {
			return atv, nil
		}
	}

	return &domínio.Ativo{}, ErrCotaçãoNãoEncontrada
}
