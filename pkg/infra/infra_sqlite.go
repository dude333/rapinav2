// SPDX-FileCopyrightText: 2024 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package infra

import (
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
)

type Tabela struct {
	Nome   string
	Up     string
	Down   string
	Versão int
}

func CriarTabelas(db *sqlx.DB, tabelas []Tabela) (err error) {
	ins := func(n string, v int) error {
		query := `INSERT OR REPLACE INTO tabelas (nome, versao) VALUES (?, ?)`
		_, err := db.Exec(query, n, v)
		return err
	}

	ver := func(tabela string) int {
		var versão int
		err := db.Get(&versão, `SELECT versao FROM tabelas WHERE nome=?`, tabela)
		if err != nil {
			progress.Debug("Erro ao buscar versão da tabela %s: %v", tabela, err)
		}
		return versão
	}

	const sqlCreateTableTabelas = `CREATE TABLE IF NOT EXISTS tabelas (
    nome   VARCHAR PRIMARY KEY,
    versao INTEGER NOT NULL
  )`
	_, _ = db.Exec(sqlCreateTableTabelas)

	for _, t := range tabelas {
		v := ver(t.Nome)
		if v == t.Versão {
			continue
		}
		progress.Status(`Apagando tabela "%s", versão %d, e recriando nova versão (v%d)`,
			t.Nome, v, t.Versão)

		_, _ = db.Exec(t.Down)
		_, err := db.Exec(t.Up)
		if err != nil {
			return err
		}
		err = ins(t.Nome, t.Versão)
		if err != nil {
			return err
		}
	}

	return nil
}
