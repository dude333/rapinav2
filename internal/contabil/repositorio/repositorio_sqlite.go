// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

import (
	"context"
	"database/sql"
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"strings"
	"unicode"

	// "github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/mattn/go-sqlite3"

	// _ "github.com/mattn/go-sqlite3"
	"strconv"
)

// sqlite implementa RepositórioLeituraEscritaDFP
type sqlite struct {
	db *sqlx.DB

	// limpo serve para sinalizar se os dados DFP de um determinado CNPJ+ANO
	// foi limpo ao rodar a primeira vez (para evitar duplicação de dados ao
	// rodar a coleta mais de uma vez). Portanto, o armazenamento do DFP
	// de *todas* as empresas em um determinado ano (CNPJ+ANO) deve ser feito
	// de uma única vez.
	limpo map[string]bool

	cache []string
}

func NovoSqlite(db *sqlx.DB) (contábil.RepositórioLeituraEscritaDFP, error) {
	err := criarTabelas(db)
	if err != nil {
		return nil, err
	}

	limpo := make(map[string]bool)
	cache := make([]string, 0, 500)

	return &sqlite{db: db, limpo: limpo, cache: cache}, nil
}

func (s *sqlite) Ler(ctx context.Context, cnpj string, ano int) (*contábil.DFP, error) {
	var sd sqliteDFP
	err := s.db.GetContext(ctx, &sd, `SELECT * FROM dfp WHERE cnpj=? AND ano=?`, &cnpj, &ano)
	if err == sql.ErrNoRows {
		err = s.db.GetContext(ctx, &sd, `SELECT * FROM dfp WHERE nome=? AND ano=?`, &cnpj, &ano)
	}
	if err != nil {
		progress.Error(err)
		return nil, err
	}

	dfp := contábil.DFP{
		CNPJ:   sd.CNPJ,
		Nome:   sd.Nome,
		Ano:    sd.Ano,
		Contas: nil,
	}

	contas := make([]contábil.Conta, 0, 100)
	rows, err := s.db.QueryxContext(ctx,
		`SELECT * FROM contas WHERE dfp_id=? ORDER BY codigo`, &sd.ID)
	if err != nil {
		progress.Error(err)
		return nil, err
	}
	for rows.Next() {
		var sc sqliteConta
		err := rows.StructScan(&sc)
		if err != nil {
			progress.Error(err)
			return nil, err
		}
		conta := contábil.Conta{
			Código:       sc.Código,
			Descr:        sc.Descr,
			Consolidado:  sc.Consolidado != 0,
			GrupoDFP:     sc.GrupoDFP,
			DataFimExerc: sc.DataFimExerc,
			OrdemExerc:   "",
			Total: contábil.Dinheiro{
				Valor:  sc.Valor,
				Escala: sc.Escala,
				Moeda:  sc.Moeda,
			},
		}
		contas = append(contas, conta)
	}

	dfp.Contas = contas

	return &dfp, err
}

func (s *sqlite) Empresas(ctx context.Context, nome string) []string {
	if len(s.cache) == 0 {
		err := s.db.SelectContext(ctx, &s.cache,
			`SELECT DISTINCT(nome) FROM dfp ORDER BY nome`)
		if err != nil {
			progress.Error(err)
			return nil
		}
	}

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	nome, _, err := transform.String(t, nome)
	nome = strings.ToLower(nome)
	if err != nil {
		return nil
	}

	var ret []string
	for _, n := range s.cache {
		x, _, err := transform.String(t, n)
		if err != nil {
			return nil
		}
		x = strings.ToLower(x)
		if strings.HasPrefix(x, nome) {
			ret = append(ret, n)
		}
	}
	return ret
}

func (s *sqlite) Salvar(ctx context.Context, empresa *contábil.DFP) error {
	// progress.Status("%-60s %4d\n", empresa.Nome, len(empresa.Contas))

	return s.inserirOuAtualizarDFP(ctx, empresa)
}

type sqliteDFP struct {
	ID   int    `db:"id"`
	CNPJ string `db:"cnpj"`
	Nome string `db:"nome"`
	Ano  int    `db:"ano"`
}

type sqliteConta struct {
	ID           int     `db:"dfp_id"`
	Código       string  `db:"codigo"`
	Descr        string  `db:"descr"`
	GrupoDFP     string  `db:"grupo_dfp"`
	Consolidado  int     `db:"consolidado"`
	DataFimExerc string  `db:"data_fim_exerc"`
	Valor        float64 `db:"valor"`
	Escala       int     `db:"escala"`
	Moeda        string  `db:"moeda"`
}

func (s *sqlite) inserirOuAtualizarDFP(ctx context.Context, dfp *contábil.DFP) error {
	d := sqliteDFP{
		CNPJ: dfp.CNPJ,
		Nome: dfp.Nome,
		Ano:  dfp.Ano,
	}

	idRegistro := func() (int, error) {
		var id int
		err := s.db.GetContext(ctx, &id, `SELECT id FROM dfp WHERE cnpj=? AND ano=?`, d.CNPJ, d.Ano)
		return id, err
	}

	k := d.CNPJ + strconv.Itoa(d.Ano)
	if _, ok := s.limpo[k]; !ok {
		// Verificar o id do registro e apagá-lo caso exista
		id, err := idRegistro()
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if err != sql.ErrNoRows {
			if err := removerDFPeContas(ctx, s.db, id); err != nil {
				return err
			}
		}
		s.limpo[k] = true
		// Criar novo registro
		query := `INSERT INTO dfp (cnpj, nome, ano) VALUES (:cnpj, :nome, :ano)`
		_, err = s.db.NamedExecContext(ctx, query, &d)
		if err != nil {
			return err
		}
	}

	id, err := idRegistro()
	if err != nil {
		return err
	}

	err = inserirContas(ctx, s.db, id, dfp.Contas)

	return err
}

// inserirContas insere os registro das contas, sendo que deve ter sido garantido
// previamente que não exista nenhum registro com o dfp_id das contas a serem
// inseridas.
func inserirContas(ctx context.Context, db *sqlx.DB, id int, contas []contábil.Conta) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO contas 
		(dfp_id, codigo, descr, grupo_dfp, consolidado, data_fim_exerc, valor, escala, moeda) 
		VALUES (?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}

	boolToInt := func(is bool) int {
		if is {
			return 1
		}
		return 0
	}

	for i := range contas {

		var args []interface{}
		args = append(args, id)
		args = append(args, contas[i].Código)
		args = append(args, contas[i].Descr)
		args = append(args, contas[i].GrupoDFP)
		args = append(args, boolToInt(contas[i].Consolidado))
		args = append(args, contas[i].DataFimExerc)
		args = append(args, contas[i].Total.Valor)
		args = append(args, contas[i].Total.Escala)
		args = append(args, contas[i].Total.Moeda)

		_, err = stmt.ExecContext(ctx, args...)
		if err != nil {
			// Ignora erro em caso de registro duplicado (dfp_id + codigo), pois se
			// trata de erro no arquivo da CVM (raramente acontece)
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.Code != sqlite3.ErrConstraint {
				_ = tx.Rollback()
				return err
			}
		}
	}

	progress.Spinner()

	return tx.Commit()
}

func removerDFPeContas(ctx context.Context, db *sqlx.DB, id int) error {
	query := `DELETE FROM contas WHERE dfp_id=?`
	_, err := db.ExecContext(ctx, query, &id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	query = `DELETE FROM dfp WHERE id=?`
	_, err = db.ExecContext(ctx, query, &id)

	return err
}

// tabelas
//
//   +------------+      +------------+
//   | dfp        |      | contas     |
//   +------------+      +------------+
//   | id*        |-----<| dfp_id*    |
//   | cnpj       |      | codigo*    |
//   | nome       |      | descr      |
//   | ano        |      | ...        |
//   +------------+      +------------+
//
// Passos oo inserir um registro DFP:
//
// 1. Verificar e remover se o registro já existe:
//    a. SELECT id FROM dfp WHERE cnpj = ? AND ano = ?;
//    b. DELETE FROM contas WHERE dfp_id = ?;
//    c. DELETE FROM dfp WHERE id = ?;
// 2. Inserir os novos registro:
//    a. INSERT INTO dfp (cnpj, nome, ano) VALUES (?,?,?);
//    b. SELECT id FROM dfp WHERE cnpj = ? AND ano = ?;
//    b. for range contas => INSERT INTO contas (dfp_id, ...) VALUES (?, ...)
//
var tabelas = []struct {
	nome   string
	versão int
	up     string
	down   string
}{
	{
		nome:   "dfp",
		versão: _ver_,
		up: `CREATE TABLE IF NOT EXISTS dfp (
			id     INTEGER PRIMARY KEY AUTOINCREMENT,
			cnpj   VARCHAR NOT NULL,
			nome   VARCHAR NOT NULL,
			ano    INT NOT NULL,
			UNIQUE (cnpj, ano)
		)`,
		down: `DROP TABLE dfp`,
	},
	{nome: "contas",
		versão: _ver_,
		up: `CREATE TABLE IF NOT EXISTS contas (
			dfp_id         INTEGER,
			codigo         VARCHAR NOT NULL,
			descr          VARCHAR NOT NULL,
			grupo_dfp      VARCHAR NOT NULL,
			consolidado    INTEGER NOT NULL,
			data_fim_exerc VARCHAR NOT NULL,
			valor          REAL NOT NULL,
			escala         INTEGER NOT NULL,
			moeda          VARCHAR,
			PRIMARY KEY (dfp_id, codigo)
		)`,
		down: `DROP TABLE contas`,
	},
}

const (
	_ver_                 = 6
	sqlCreateTableTabelas = `CREATE TABLE IF NOT EXISTS tabelas (
		nome   VARCHAR PRIMARY KEY,
		versao INTEGER NOT NULL
	)`
)

func criarTabelas(db *sqlx.DB) (err error) {
	ins := func(n string, v int) error {
		query := `INSERT OR REPLACE INTO tabelas (nome, versao) VALUES (?, ?)`
		_, err := db.Exec(query, n, v)
		return err
	}

	ver := func(tabela string) int {
		var versão int
		_ = db.Get(&versão, `SELECT versao FROM tabelas WHERE nome=?`, tabela)
		return versão
	}

	_, _ = db.Exec(sqlCreateTableTabelas)

	for _, t := range tabelas {
		if ver(t.nome) == t.versão {
			continue
		}
		progress.Status(`Apagando tabela "%s" e recriando nova versão (v%d)`,
			t.nome, t.versão)

		_, _ = db.Exec(t.down)
		_, err := db.Exec(t.up)
		if err != nil {
			return err
		}
		err = ins(t.nome, t.versão)
		if err != nil {
			return err
		}
	}

	return nil
}
