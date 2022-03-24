// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

import (
	"context"
	"fmt"
	"testing"

	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func Test_inserirDFP(t *testing.T) {
	var db *sqlx.DB
	if testing.Short() {
		db = sqlx.MustConnect("sqlite3", ":memory:")
		db.SetMaxOpenConns(1)
	} else {
		connStr := "file:/tmp/rapina.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
		db = sqlx.MustConnect("sqlite3", connStr)
	}

	s, err := NovoSqlite(db)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("criar e inserir dados", func(t *testing.T) {

		var contas []contábil.Conta
		for n := 1; n <= 10; n++ {
			c := contábil.Conta{
				Código:       fmt.Sprintf("C%03d", n),
				Descr:        fmt.Sprintf("D%03d", n),
				Grupo:        fmt.Sprintf("G%03d", n),
				DataFimExerc: fmt.Sprintf("D%03d", n),
				OrdemExerc:   fmt.Sprintf("O%03d", n),
				Total: contábil.Dinheiro{
					Valor:  1234567.89,
					Escala: 1000,
					Moeda:  "R$",
				},
			}
			contas = append(contas, c)
		}

		empresa := contábil.Empresa{
			CNPJ:         "123",
			Nome:         "N1",
			Ano:          2020,
			ContasAnuais: contas,
		}

		err := s.Salvar(context.Background(), &empresa)
		if err != nil {
			t.Logf("%v", err)
		}
	})
}
