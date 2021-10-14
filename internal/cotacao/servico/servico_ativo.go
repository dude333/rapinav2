// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

// serviço define os casos de uso de um "ativo", como leitura da sua cotação em
// um repositório (banco de dados ou API).
package serviço

import (
	"context"
	cotação "github.com/dude333/rapinav2/internal/cotacao/dominio"
	"time"
)

// ativo é um serviço de busca de informações de um ativo em vários
// repositórios, como banco de dados ou via API.
type ativo struct {
	bd  cotação.RepositórioLeituraEscritaAtivo
	api []cotação.RepositórioImportaçãoAtivo
}

func NovoAtivo(
	bd cotação.RepositórioLeituraEscritaAtivo,
	api []cotação.RepositórioImportaçãoAtivo) cotação.ServiçoAtivo {
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
func (a *ativo) Cotação(código string, dia cotação.Data) (*cotação.Ativo, error) {
	atv, err := a.cotaçãoBD(código, dia)
	if err != nil {
		return a.cotaçãoAPI(código, dia)
	}
	return atv, err
}

func (a *ativo) cotaçãoBD(código string, dia cotação.Data) (*cotação.Ativo, error) {
	if a.bd == nil {
		return nil, ErrRepositórioInválido
	}

	atv, err := a.bd.Cotação(context.Background(), código, dia)
	if err == nil {
		return atv, nil
	}

	return nil, ErrCotaçãoNãoEncontrada
}

func (a *ativo) cotaçãoAPI(código string, dia cotação.Data) (*cotação.Ativo, error) {
	if len(a.api) < 1 {
		return nil, ErrRepositórioInválido
	}

	var atv *cotação.Ativo
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
