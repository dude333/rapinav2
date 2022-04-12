// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

import (
	"context"
	"database/sql"
	"strings"
	"unicode"

	rapina "github.com/dude333/rapinav2/internal"
	contábil "github.com/dude333/rapinav2/internal/contabil"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	// "github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/mattn/go-sqlite3"

	// _ "github.com/mattn/go-sqlite3"
	"strconv"
)

// Sqlite implementa RepositórioLeituraEscrita
type Sqlite struct {
	db *sqlx.DB

	// limpo serve para sinalizar se os dados de um determinado CNPJ+ANO
	// foi limpo ao rodar a primeira vez (para evitar duplicação de dados
	// ao rodar a coleta mais de uma vez). Portanto, o armazenamento do
	// de *todas* as empresas em um determinado ano (CNPJ+ANO) deve ser
	// feito de uma única vez.
	limpo map[string]bool

	cache []string
	cfg
}

func NovoSqlite(db *sqlx.DB, configs ...ConfigFn) (*Sqlite, error) {
	var s Sqlite
	for _, cfg := range configs {
		cfg(&s.cfg)
	}

	s.db = db

	err := criarTabelas(s.db)
	if err != nil {
		return nil, err
	}

	s.limpo = make(map[string]bool)
	s.cache = make([]string, 0, 500)

	return &s, nil
}

func (s *Sqlite) Ler(ctx context.Context, cnpj string, ano int) (*contábil.DemonstraçãoFinanceira, error) {
	var sd sqliteEmpresa
	err := s.db.GetContext(ctx, &sd, `SELECT * FROM empresas WHERE cnpj=? AND ano=?`, &cnpj, &ano)
	if err == sql.ErrNoRows {
		err = s.db.GetContext(ctx, &sd, `SELECT * FROM empresas WHERE nome=? AND ano=?`, &cnpj, &ano)
	}
	if err != nil {
		progress.Error(err)
		return nil, err
	}

	dfp := contábil.DemonstraçãoFinanceira{
		Empresa: rapina.Empresa{
			CNPJ: sd.CNPJ,
			Nome: sd.Nome,
		},
		Ano:    sd.Ano,
		Contas: nil,
	}

	contas := make([]contábil.Conta, 0, 100)
	rows, err := s.db.QueryxContext(ctx,
		`SELECT * FROM contas WHERE id_empresa=? ORDER BY codigo`, &sd.ID)
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
			Grupo:        sc.Grupo,
			DataIniExerc: sc.DataIniExerc,
			DataFimExerc: sc.DataFimExerc,
			Meses:        12,
			OrdemExerc:   "",
			Total: rapina.Dinheiro{
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

func (s *Sqlite) Empresas(ctx context.Context, nome string) []string {
	if len(s.cache) == 0 {
		err := s.db.SelectContext(ctx, &s.cache,
			`SELECT DISTINCT(nome) FROM empresas ORDER BY nome`)
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

func (s *Sqlite) Salvar(ctx context.Context, dfp *contábil.DemonstraçãoFinanceira) error {
	// progress.Status("%-60s %4d\n", empresa.Nome, len(empresa.Contas))

	return s.inserirOuAtualizarEmpresa(ctx, dfp)
}

type sqliteEmpresa struct {
	ID           int    `db:"id"`
	CNPJ         string `db:"cnpj"`
	Nome         string `db:"nome"`
	Ano          int    `db:"ano"`
	DataIniExerc string `db:"data_ini_exerc"`
}

type sqliteConta struct {
	ID           int     `db:"id_empresa"`
	Código       string  `db:"codigo"`
	Descr        string  `db:"descr"`
	Grupo        string  `db:"grupo"`
	Consolidado  int     `db:"consolidado"`
	DataIniExerc string  `db:"data_ini_exerc"`
	DataFimExerc string  `db:"data_fim_exerc"`
	Meses        int     `db:"meses"` // Meses acumulados desde o início do exercício
	Valor        float64 `db:"valor"`
	Escala       int     `db:"escala"`
	Moeda        string  `db:"moeda"`
}

func (s *Sqlite) inserirOuAtualizarEmpresa(ctx context.Context, e *contábil.DemonstraçãoFinanceira) error {
	d := sqliteEmpresa{
		CNPJ:         e.CNPJ,
		Nome:         e.Nome,
		Ano:          e.Ano,
		DataIniExerc: e.DataIniExerc,
	}

	idRegistro := func() (int, error) {
		var id int
		err := s.db.GetContext(ctx, &id, `SELECT id FROM empresas WHERE cnpj=? AND ano=?`, d.CNPJ, d.Ano)
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
			if err := removerEmpresa(ctx, s.db, id); err != nil {
				return err
			}
		}
		s.limpo[k] = true
		// Criar novo registro
		query := `INSERT INTO empresas (cnpj, nome, ano, data_ini_exerc) VALUES (:cnpj, :nome, :ano, :data_ini_exerc)`
		_, err = s.db.NamedExecContext(ctx, query, &d)
		if err != nil {
			progress.Debug("Falha ao inserir %v", d)
			return err
		}
	}

	id, err := idRegistro()
	if err != nil {
		return err
	}

	return inserirContas(ctx, s.db, id, e.Contas)
}

// inserirContas insere os registro das contas, sendo que deve ter sido garantido
// previamente que não exista nenhum registro com o id_empresa das contas a serem
// inseridas.
func inserirContas(ctx context.Context, db *sqlx.DB, id int, contas []contábil.Conta) error {
	if len(contas) == 0 {
		return nil
	}

	var dataIniExerc string
	err := db.GetContext(ctx, &dataIniExerc, `SELECT data_ini_exerc FROM empresas WHERE id=?`, id)
	if err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT or IGNORE INTO contas 
		(id_empresa, codigo, descr, grupo, consolidado, data_ini_exerc, data_fim_exerc, meses, valor, escala, moeda) 
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`)
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

		// pular registros com início do exercício diferente do início do
		// exercício anual, ou seja, ignora registros trimestrais e armazena
		// apenas os dados acumulados desde o início do exercício anual:
		// 3, 6, 9 e 12 meses.
		if dataIniExerc != "" && dataIniExerc != contas[i].DataIniExerc {
			progress.Debug("Ignorando registro trimestral não acumulado %d:%v", id, contas[i])
			continue
		}
		if contas[i].Meses%3 != 0 || contas[i].Meses > 12 {
			progress.Debug("Ignorando registro com meses != 3,6,9,12 %d:%v", id, contas[i])
			continue
		}

		var args []interface{}
		args = append(args, id)
		args = append(args, contas[i].Código)
		args = append(args, contas[i].Descr)
		args = append(args, contas[i].Grupo)
		args = append(args, boolToInt(contas[i].Consolidado))
		args = append(args, contas[i].DataIniExerc)
		args = append(args, contas[i].DataFimExerc)
		args = append(args, contas[i].Meses)
		args = append(args, contas[i].Total.Valor)
		args = append(args, contas[i].Total.Escala)
		args = append(args, contas[i].Total.Moeda)

		_, err = stmt.ExecContext(ctx, args...)
		// Erros no banco de dados estão sendo ignorados ("INSERT or IGNORE INTO")
		if err != nil {
			// Ignora erro em caso de registro duplicado (id_empresa + codigo),
			// pois se trata de erro no arquivo da CVM (raramente acontece)
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.Code != sqlite3.ErrConstraint {
				_ = tx.Rollback()
				return err
			} else {
				progress.Error(sqliteErr)
			}
		}
	}

	progress.Spinner()

	return tx.Commit()
}

func removerEmpresa(ctx context.Context, db *sqlx.DB, id int) error {
	query := `DELETE FROM contas WHERE id_empresa=?`
	_, err := db.ExecContext(ctx, query, &id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	query = `DELETE FROM empresas WHERE id=?`
	_, err = db.ExecContext(ctx, query, &id)

	return err
}

// tabelas
//
//   +------------+      +------------+
//   | empresas   |      | contas     |
//   +------------+      +------------+
//   | id*        |-----<| id_empresa*|
//   | cnpj       |      | codigo*    |
//   | nome       |      | descr      |
//   | ano        |      | ...        |
//   +------------+      +------------+
//
// Passos oo inserir um registro empresa:
//
// 1. Verificar e remover se o registro já existe:
//    a. SELECT id FROM empresas WHERE cnpj = ? AND ano = ?;
//    b. DELETE FROM contas WHERE id_empresa = ?;
//    c. DELETE FROM empresas WHERE id = ?;
// 2. Inserir os novos registro:
//    a. INSERT INTO empresas (cnpj, nome, ano) VALUES (?,?,?);
//    b. SELECT id FROM empresas WHERE cnpj = ? AND ano = ?;
//    b. for range contas => INSERT INTO contas (id_empresa, ...) VALUES (?, ...)
//
var tabelas = []struct {
	nome   string
	versão int
	up     string
	down   string
}{
	{
		nome:   "empresas",
		versão: _ver_,
		up: `CREATE TABLE IF NOT EXISTS empresas (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			cnpj           VARCHAR NOT NULL,
			nome           VARCHAR NOT NULL,
			ano            INT NOT NULL,
			data_ini_exerc VARCHAR,
			UNIQUE (cnpj, ano)
		)`,
		down: `DROP TABLE empresas`,
	},
	{nome: "contas",
		versão: _ver_,
		up: `CREATE TABLE IF NOT EXISTS contas (
			id_empresa     INTEGER,
			codigo         VARCHAR NOT NULL,
			descr          VARCHAR NOT NULL,
			grupo          VARCHAR NOT NULL,
			consolidado    INTEGER NOT NULL,
			data_ini_exerc VARCHAR,
			data_fim_exerc VARCHAR NOT NULL,
			meses          INTEGER NOT NULL,
			valor          REAL NOT NULL,
			escala         INTEGER NOT NULL,
			moeda          VARCHAR,
			PRIMARY KEY (id_empresa, codigo, meses)
		)`,
		down: `DROP TABLE contas`,
	},
}

const (
	_ver_                 = 13
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
