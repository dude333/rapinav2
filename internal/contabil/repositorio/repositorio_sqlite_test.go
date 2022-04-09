// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

import (
	"context"
	"fmt"
	"os"
	"testing"

	rapina "github.com/dude333/rapinav2/internal"
	_ "github.com/mattn/go-sqlite3"
)

func Test_inserirDFP(t *testing.T) {
	var c ConfigFn
	if testing.Short() {
		c = RodarBDNaMemória()
	} else {
		c = DirBD(os.TempDir())
	}

	s, err := NovoSqlite(c)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("criar e inserir dados", func(t *testing.T) {

		var contas []rapina.Conta
		for n := 1; n <= 10; n++ {
			c := rapina.Conta{
				Código:       fmt.Sprintf("C%03d", n),
				Descr:        fmt.Sprintf("D%03d", n),
				Grupo:        fmt.Sprintf("G%03d", n),
				DataFimExerc: fmt.Sprintf("D%03d", n),
				OrdemExerc:   fmt.Sprintf("O%03d", n),
				Total: rapina.Dinheiro{
					Valor:  1234567.89,
					Escala: 1000,
					Moeda:  "R$",
				},
			}
			contas = append(contas, c)
		}

		empresa := rapina.Empresa{
			CNPJ:   "123",
			Nome:   "N1",
			Ano:    2020,
			Contas: contas,
		}

		err := s.Salvar(context.Background(), &empresa)
		if err != nil {
			t.Logf("%v", err)
		}
	})
}
