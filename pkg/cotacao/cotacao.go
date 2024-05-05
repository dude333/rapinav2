// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package cotacao

import (
	"context"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/progress"
)

// Erros
var (
	ErrCotaçãoNãoEncontrada = errors.New("cotação não encontrada")
	ErrRepositórioInválido  = errors.New("repositório inválido")
	ErrProvedorInválido     = errors.New("provedor inválido")
)

type Importar interface {
	Importar(ctx context.Context, dia rapina.Data) <-chan Resultado
}

// Serviço é um serviço que implementa Importação e busca
// cotações de um Ativo em vários repositórios (API e BD).
type Serviço struct {
	bd  *Sqlite
	api []Importar
}

func NovoServiço(db *sqlx.DB, dirDados string) (*Serviço, error) {
	sqlite, err := NovoSqlite(db)
	if err != nil {
		return nil, err
	}

	return &Serviço{
		bd:  sqlite,
		api: []Importar{NovoB3(dirDados)},
	}, nil
}

// Cotação busca a cotação de uma empresa em vários repositórios num determinado
// "dia", retornando o primeiro valor encontado ou o erro de todos os repositórios.
// Caso a cotação seja encontrada via API, ela será armazenada no bando de dados
// para agilizar a próxima leitura do mesmo código, na mesma data.
func (svc *Serviço) Cotação(empresa rapina.Empresa, dia rapina.Data) ([]*rapina.Cotação, error) {
	código, err := GetTickerPrefix(svc.bd.db, empresa.CNPJ)
	if err != nil {
		return nil, err
	}

	progress.Debug("Cotação(%s, %s)", código, dia)
	atv, err := svc.cotaçãoBD(código, dia)
	if err != nil {
		return svc.cotaçãoAPI(código, dia)
	}
	return atv, err
}

func (svc *Serviço) cotaçãoBD(código string, dia rapina.Data) ([]*rapina.Cotação, error) {
	progress.Debug("Lendo cotação de %s, em %s, do bd", código, dia)
	if svc.bd == nil {
		return nil, ErrRepositórioInválido
	}

	return svc.bd.LerCotações(context.Background(), código, dia)
}

func (svc *Serviço) cotaçãoAPI(código string, dia rapina.Data) ([]*rapina.Cotação, error) {
	progress.Debug("Lendo cotação de %s, em %s, via API", código, dia)
	if len(svc.api) < 1 {
		return nil, ErrProvedorInválido
	}

	var alvo []*rapina.Cotação
	var todos []*rapina.Cotação
	ctx := context.Background()

	// Tentativa de coletar a cotação usando vários servidores de API
	for i := range svc.api {
		// result retorna o registro após a leitura de cada linha
		// do arquivo importado
		for result := range svc.api[i].Importar(ctx, dia) {
			if result.Error != nil {
				progress.Error(result.Error)
				return nil, result.Error
			}
			if strings.HasPrefix(strings.ToLower(result.Ativo.Código), strings.ToLower(código)) {
				alvo = append(alvo, result.Ativo)
			}
			todos = append(todos, result.Ativo)
		}
		// Finaliza se ativo alvo (código) já tiver sido encontrado
		if len(alvo) > 0 {
			break
		}
	}

	if svc.bd != nil {
		_ = svc.bd.SalvarCotações(ctx, todos)
	}

	if len(alvo) == 0 {
		return nil, ErrCotaçãoNãoEncontrada
	}

	return alvo, nil
}
