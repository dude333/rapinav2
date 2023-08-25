// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"unicode"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil/dominio"
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
	// ao rodar a coleta mais de uma vez). Portanto, o armazenamento de
	// *todas* as empresas em um determinado ano (CNPJ+ANO) deve ser feito
	// uma única vez.
	limpo map[string]bool

	cacheEmpresas []rapina.Empresa
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
	s.cacheEmpresas = make([]rapina.Empresa, 0, 500)

	return &s, nil
}

func (s *Sqlite) Ler(ctx context.Context, cnpj string, ano int) (*dominio.DemonstraçãoFinanceira, error) {
	var sd sqliteEmpresa
	err := s.db.GetContext(ctx, &sd, `SELECT * FROM empresas WHERE cnpj=? AND ano=?`, &cnpj, &ano)
	if err == sql.ErrNoRows {
		err = s.db.GetContext(ctx, &sd, `SELECT * FROM empresas WHERE nome=? AND ano=?`, &cnpj, &ano)
	}
	if err != nil {
		progress.Error(err)
		return nil, err
	}

	dfp := dominio.DemonstraçãoFinanceira{
		Empresa: rapina.Empresa{
			CNPJ: sd.CNPJ,
			Nome: sd.Nome,
		},
		Ano:    sd.Ano,
		Contas: nil,
	}

	contas := make([]dominio.Conta, 0, 100)
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
		conta := dominio.Conta{
			Código:       sc.Código,
			Descr:        sc.Descr,
			Consolidado:  sc.Consolidado != 0,
			Grupo:        sc.Grupo,
			DataIniExerc: sc.DataIniExerc,
			DataFimExerc: sc.DataFimExerc,
			Meses:        sc.Meses,
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

func (s *Sqlite) Trimestral(ctx context.Context, cnpj string) ([]rapina.InformeTrimestral, error) {
	var ids []int
	err := s.db.SelectContext(ctx, &ids, `SELECT id FROM empresas WHERE cnpj=? ORDER BY ano`, &cnpj)
	if err == sql.ErrNoRows {
		err = s.db.SelectContext(ctx, &ids, `SELECT id FROM empresas WHERE nome=? ORDER BY ano`, &cnpj)
	}
	if err != nil {
		return nil, err
	}

	progress.Debug("[]sqliteEmpresa => %+v", ids)

	var resultados []resultadoTrimestral
	err = s.db.SelectContext(ctx, &resultados, sqlTrimestral(ids, true))
	if err != nil {
		return nil, err
	}
	if len(resultados) == 0 {
		// Buscar por dados individuais caso a empresa não tenha dados consolidados
		progress.Trace("Dados consolidados não encontrados; procurando por dados individuais...")
		err = s.db.SelectContext(ctx, &resultados, sqlTrimestral(ids, false))
		if err != nil {
			return nil, err
		}
	}

	return converterResultadosTrimestrais(resultados)
}

func (s *Sqlite) Empresas(ctx context.Context, nome string) []rapina.Empresa {
	if len(s.cacheEmpresas) == 0 {
		err := s.db.SelectContext(ctx, &s.cacheEmpresas,
			`SELECT DISTINCT(cnpj), nome FROM empresas ORDER BY nome`)
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

	var ret []rapina.Empresa
	var toSort []string
	for _, empr := range s.cacheEmpresas {
		x, _, err := transform.String(t, empr.Nome)
		if err != nil {
			return nil
		}
		x = strings.ToLower(x)
		if strings.HasPrefix(x, nome) {
			ret = append(ret, empr)
			toSort = append(toSort, x)
		}
	}
	return ordenar(ret, toSort)
}

func (s *Sqlite) Hashes() []string {
	var hashes []string
	_ = s.db.Select(&hashes, `SELECT DISTINCT(hash) FROM hashes`)
	return hashes
}

func (s *Sqlite) SalvarHash(ctx context.Context, hash string) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO hashes (hash) VALUES ($1)`, hash)
	return err
}

// ordenar ordenada a []string "orig" com base na []string
// transformada "transf". Serve para ordenar []string com acentos
// ou outros sinais diacríticos.
func ordenar(orig []rapina.Empresa, transf []string) []rapina.Empresa {
	s := NewSlice(transf)
	sort.Sort(s)

	ord := make([]rapina.Empresa, len(transf))
	for i, j := range s.idx {
		ord[i] = orig[j]
	}

	return ord
}

func NewSlice(str []string) *Slice {
	s := &Slice{StringSlice: str, idx: make([]int, len(str))}
	for i := range s.idx {
		s.idx[i] = i
	}
	return s
}

type Slice struct {
	sort.StringSlice
	idx []int
}

func (s Slice) Swap(i, j int) {
	s.StringSlice.Swap(i, j)
	s.idx[i], s.idx[j] = s.idx[j], s.idx[i]
}

type sqliteEmpresa struct {
	ID   int    `db:"id"`
	CNPJ string `db:"cnpj"`
	Nome string `db:"nome"`
	Ano  int    `db:"ano"`
}

type sqliteConta struct {
	ID           int     `db:"id_empresa"`
	Código       string  `db:"codigo"`
	Descr        string  `db:"descr"`
	Grupo        string  `db:"grupo"`
	Consolidado  int     `db:"consolidado"`
	DataIniExerc string  `db:"data_ini_exerc"`
	DataFimExerc string  `db:"data_fim_exerc"`
	Meses        int     `db:"meses"` // diferença entre data_ini_exerc e data_fim_exerc
	Valor        float64 `db:"valor"`
	Escala       int     `db:"escala"`
	Moeda        string  `db:"moeda"`
}

func (s *Sqlite) Salvar(ctx context.Context, dfp *dominio.DemonstraçãoFinanceira) error {
	progress.Trace("%-60s %4d\n", dfp.Nome, len(dfp.Contas))

	d := sqliteEmpresa{
		CNPJ: dfp.CNPJ,
		Nome: dfp.Nome,
		Ano:  dfp.Ano,
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
			progress.Debug("Apagando empresa %s, %d (%d): ", d.Nome, d.Ano, id)
			if err := removerEmpresa(ctx, s.db, id); err != nil {
				return err
			}
		}
		s.limpo[k] = true
		// Criar novo registro
		query := `INSERT INTO empresas (cnpj, nome, ano) VALUES (:cnpj, :nome, :ano)`
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

	return inserirContas(ctx, s.db, id, dfp.Contas, dfp.Nome)
}

// inserirContas insere os registro das contas, sendo que deve ter sido garantido
// previamente que não exista nenhum registro com o id_empresa das contas a serem
// inseridas.
func inserirContas(ctx context.Context, db *sqlx.DB, id int, contas []dominio.Conta, nome string) error {
	if len(contas) == 0 {
		return nil
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareNamedContext(ctx, `INSERT or IGNORE INTO contas
		(id_empresa, codigo, descr, grupo, consolidado, data_ini_exerc, data_fim_exerc, meses, valor, escala, moeda)
		VALUES
		(:id_empresa, :codigo, :descr, :grupo, :consolidado, :data_ini_exerc, :data_fim_exerc, :meses, :valor, :escala, :moeda)`)
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
		c := sqliteConta{
			ID:           id,
			Código:       contas[i].Código,
			Descr:        contas[i].Descr,
			Grupo:        contas[i].Grupo,
			Consolidado:  boolToInt(contas[i].Consolidado),
			DataIniExerc: contas[i].DataIniExerc,
			DataFimExerc: contas[i].DataFimExerc,
			Meses:        contas[i].Meses,
			Valor:        contas[i].Total.Valor,
			Escala:       contas[i].Total.Escala,
			Moeda:        contas[i].Total.Moeda,
		}

		_, err = stmt.ExecContext(ctx, c)
		// Erros no banco de dados estão sendo ignorados ("INSERT or IGNORE INTO").
		// Verificar PRIMARY KEY da tabela 'contas'.
		if err != nil {
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.Code != sqlite3.ErrConstraint {
				_ = tx.Rollback()
				return err
			} else {
				progress.ErrorMsg("%s: %d, %s, %#v", err, id, nome, contas[i])
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
			UNIQUE (cnpj, ano)
		)`,
		down: `DROP TABLE IF EXISTS empresas`,
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
			PRIMARY KEY (id_empresa, codigo, data_ini_exerc, data_fim_exerc)
		)`,
		down: `DROP TABLE IF EXISTS contas`,
	},
	{
		nome:   "hashes",
		versão: _ver_,
		up: `CREATE TABLE IF NOT EXISTS hashes (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			hash           VARCHAR NOT NULL,
			UNIQUE (hash)
		)`,
		down: "DROP TABLE IF EXISTS hashes",
	},
}

const (
	_ver_                 = 17
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
		err := db.Get(&versão, `SELECT versao FROM tabelas WHERE nome=?`, tabela)
		if err != nil {
			progress.Debug("Erro ao buscar versão da tabela %s: %v", tabela, err)
		}
		return versão
	}

	_, _ = db.Exec(sqlCreateTableTabelas)

	for _, t := range tabelas {
		v := ver(t.nome)
		if v == t.versão {
			continue
		}
		progress.Status(`Apagando tabela "%s", versão %d, e recriando nova versão (v%d)`,
			t.nome, v, t.versão)

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
