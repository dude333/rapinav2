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

var tabelasInfra = []Tabela{
	{
		Nome:   "hashes",
		Versão: 1,
		Up: `CREATE TABLE IF NOT EXISTS hashes (
        hash VARCHAR PRIMARY KEY
      )`,
		Down: "DROP TABLE IF EXISTS hashes",
	},
}

// CriarTabelas cria as tabelas em tabelasInfra, além das tabelas passadas
// como parâmetro.
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

	for _, t := range append(tabelas, tabelasInfra...) {
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

func Hashes(db *sqlx.DB) ([]string, error) {
	var hashes []string
	err := db.Select(&hashes, `SELECT hash FROM hashes`)
	return hashes, err
}

func HasHash(db *sqlx.DB, hash string) (bool, error) {
	var count int
	err := db.Get(&count, `SELECT COUNT(*) FROM hashes WHERE hash = ?`, hash)
	return count > 0, err
}

func SaveHash(db *sqlx.DB, hash string) error {
	_, err := db.Exec(`INSERT OR REPLACE INTO hashes (hash) VALUES (?)`, hash)
	return err
}
