// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package contabil

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil/dominio"
	ext "github.com/dude333/rapinav2/pkg/infra"
	"github.com/dude333/rapinav2/pkg/progress"
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
	cfg           *cfg
}

func NewSqlite(db *sqlx.DB, configs ...Option) (*Sqlite, error) {
	var s Sqlite
	s.cfg = &cfg{}
	if err := s.cfg.apply(configs...); err != nil {
		return nil, err
	}

	s.db = db

	err := ext.CriarTabelas(s.db, tabelas)
	if err != nil {
		return nil, err
	}

	s.limpo = make(map[string]bool)
	s.cacheEmpresas = make([]rapina.Empresa, 0, 500)

	return &s, nil
}

/*
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
*/

func (s *Sqlite) Trimestral(ctx context.Context, cnpj string, consolidado bool) ([]rapina.InformeTrimestral, error) {
	var ids []int
	err := s.db.SelectContext(ctx, &ids, `SELECT id FROM empresas WHERE cnpj=? ORDER BY ano`, &cnpj)
	if err == sql.ErrNoRows {
		err = s.db.SelectContext(ctx, &ids, `SELECT id FROM empresas WHERE nome=? ORDER BY ano`, &cnpj)
	}
	if err != nil {
		return nil, err
	}

	progress.Trace("[]sqliteEmpresa => %+v", ids)

	var resultados []resultadoTrimestral
	err = s.db.SelectContext(ctx, &resultados, sqlTrimestral(ids, consolidado))
	if err != nil {
		return nil, err
	}

	return converterResultadosTrimestrais(resultados)
}

func (s *Sqlite) Empresas(ctx context.Context) ([]rapina.Empresa, error) {
	var empresas []rapina.Empresa
	err := s.db.SelectContext(ctx, &empresas,
		`SELECT DISTINCT(cnpj), nome FROM empresas ORDER BY nome`)
	if err != nil {
		progress.Error(err)
		return nil, err
	}
	return empresas, nil
}

func (s *Sqlite) BuscaEmpresas(ctx context.Context, nome string) ([]rapina.Empresa, error) {
	if len(s.cacheEmpresas) == 0 {
		err := s.db.SelectContext(ctx, &s.cacheEmpresas,
			`SELECT DISTINCT(cnpj), nome FROM empresas ORDER BY nome`)
		if err != nil {
			progress.Error(err)
			return nil, err
		}
	}

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	nome, _, err := transform.String(t, nome)
	nome = strings.ToLower(nome)
	if err != nil {
		return nil, err
	}

	var ret []rapina.Empresa
	for _, empr := range s.cacheEmpresas {
		x, _, err := transform.String(t, empr.Nome)
		if err != nil {
			return nil, err
		}
		x = strings.ToLower(x)
		if strings.HasPrefix(x, nome) {
			ret = append(ret, empr)
		}
	}

	cl := collate.New(language.BrazilianPortuguese, collate.Loose)
	customSort := func(i, j int) bool {
		return cl.CompareString(ret[i].Nome, ret[j].Nome) < 0
	}
	sort.Slice(ret, customSort)

	return ret, nil
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

func (s *Sqlite) Save(ctx context.Context, dfp *dominio.DemonstracaoFinanceira) error {
	progress.Trace("%-60s %4d\n", dfp.Nome, len(dfp.Contas))

	d := sqliteEmpresa{
		CNPJ: dfp.CNPJ,
		Nome: dfp.Nome,
		Ano:  dfp.Ano,
	}

	// Garante que a empresa existe no banco de dados
	query := `INSERT OR IGNORE INTO empresas (cnpj, nome, ano) VALUES (:cnpj, :nome, :ano)`
	_, err := s.db.NamedExecContext(ctx, query, &d)
	if err != nil {
		progress.Debug("Falha ao inserir %v", d)
		return err
	}

	// Pega o ID da empresa
	var id int
	err = s.db.GetContext(ctx, &id, `SELECT id FROM empresas WHERE cnpj=? AND ano=?`, d.CNPJ, d.Ano)
	if err != nil {
		return err
	}

	// Limpa os dados antigos apenas uma vez por execução
	k := d.CNPJ + strconv.Itoa(d.Ano)
	if _, ok := s.limpo[k]; !ok {
		progress.Debug("Apagando contas de %s, %d (%d)", d.Nome, d.Ano, id)
		query := `DELETE FROM contas WHERE id_empresa=?`
		_, err := s.db.ExecContext(ctx, query, id)
		if err != nil {
			return err
		}
		s.limpo[k] = true
	}

	progress.Debug("Salvando empresa %s, %d (%d): %d contas", d.Nome, d.Ano, id, len(dfp.Contas))
	return insertContas(ctx, s.db, id, dfp)
}

// insertContas insere os registro das contas, sendo que deve ter sido garantido
// previamente que não exista nenhum registro com o id_empresa das contas a serem
// inseridas.
func insertContas(ctx context.Context, db *sqlx.DB, id int, dfp *dominio.DemonstracaoFinanceira) error {
	if len(dfp.Contas) == 0 {
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

	for _, conta := range dfp.Contas {
		c := sqliteConta{
			ID:           id,
			Código:       conta.Código,
			Descr:        conta.Descr,
			Grupo:        conta.Grupo,
			Consolidado:  boolToInt(conta.Consolidado),
			DataIniExerc: conta.DataIniExerc,
			DataFimExerc: conta.DataFimExerc,
			Meses:        conta.Meses,
			Valor:        conta.Total.Valor,
			Escala:       conta.Total.Escala,
			Moeda:        conta.Total.Moeda,
		}

		_, err = stmt.ExecContext(ctx, c)
		// Erros no banco de dados estão sendo ignorados ("INSERT or IGNORE INTO").
		// Verificar PRIMARY KEY da tabela 'contas'.
		if err != nil {
			var sqliteErr sqlite3.Error
			if errors.As(err, &sqliteErr) {
				if sqliteErr.Code != sqlite3.ErrConstraint {
					_ = tx.Rollback()
					return err
				}
				progress.ErrorMsg("%s: %d, %s, %#v", err, id, dfp.Nome, conta)
			}
		}
	}

	progress.Spinner()

	return tx.Commit()
}

// func deleteEmpresa(ctx context.Context, db *sqlx.DB, id int) error {
// 	query := `DELETE FROM contas WHERE id_empresa=?`
// 	_, err := db.ExecContext(ctx, query, &id)
// 	if err != nil && err != sql.ErrNoRows {
// 		return err
// 	}
//
// 	query = `DELETE FROM empresas WHERE id=?`
// 	_, err = db.ExecContext(ctx, query, &id)
//
// 	return err
// }

const _ver_ = 17

// tabelas
//
//	+------------+      +------------+
//	| empresas   |      | contas     |
//	+------------+      +------------+
//	| id*        |-----<| id_empresa*|
//	| cnpj       |      | codigo*    |
//	| nome       |      | descr      |
//	| ano        |      | ...        |
//	+------------+      +------------+
//
// Passos oo inserir um registro empresa:
//
//  1. Verificar e remover se o registro já existe:
//     a. SELECT id FROM empresas WHERE cnpj = ? AND ano = ?;
//     b. DELETE FROM contas WHERE id_empresa = ?;
//     c. DELETE FROM empresas WHERE id = ?;
//  2. Inserir os novos registro:
//     a. INSERT INTO empresas (cnpj, nome, ano) VALUES (?,?,?);
//     b. SELECT id FROM empresas WHERE cnpj = ? AND ano = ?;
//     b. for range contas => INSERT INTO contas (id_empresa, ...) VALUES (?, ...)
var tabelas = []ext.Tabela{
	{
		Nome:   "empresas",
		Versão: _ver_,
		Up: `CREATE TABLE IF NOT EXISTS empresas (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			cnpj           VARCHAR NOT NULL,
			nome           VARCHAR NOT NULL,
			ano            INT NOT NULL,
			UNIQUE (cnpj, ano)
		)`,
		Down: `DROP TABLE IF EXISTS empresas`,
	},
	{
		Nome:   "contas",
		Versão: _ver_,
		Up: `CREATE TABLE IF NOT EXISTS contas (
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
		Down: `DROP TABLE IF EXISTS contas`,
	},
	{
		Nome:   "fre",
		Versão: _ver_,
		Up: `CREATE TABLE IF NOT EXISTS fre (
			id_empresa                   INTEGER,
			data_ref                     VARCHAR NOT NULL,
			data_ultima_assembleia       VARCHAR,
			id_doc                       INTEGER,
			pct_acoes_ord_circulacao     REAL,
			pct_acoes_pref_circulacao    REAL,
			pct_total_acoes_circulacao   REAL,
			qtd_acionistas_inst          INTEGER,
			qtd_acionistas_pf            INTEGER,
			qtd_acionistas_pj            INTEGER,
			qtd_acoes_ord_circulacao     INTEGER,
			qtd_acoes_pref_circulacao    INTEGER,
			qtd_total_acoes_circulacao   INTEGER,
			PRIMARY KEY (id_empresa, data_ref, id_doc)
		)`,
		Down: `DROP TABLE IF EXISTS fre`,
	},
}

func (s *Sqlite) SaveFRE(ctx context.Context, fre *dominio.FreDistribCapital) error {
	if fre == nil {
		return nil
	}

	// Busca ou cria empresa
	d := sqliteEmpresa{
		CNPJ: fre.Empresa.CNPJ,
		Nome: fre.Empresa.Nome,
		Ano:  0, // FRE não tem ano explícito, pode ser extraído de DataRef se necessário
	}

	idEmpresa := 0
	err := s.db.GetContext(ctx, &idEmpresa, `SELECT id FROM empresas WHERE cnpj=?`, d.CNPJ)
	if err == sql.ErrNoRows {
		// Cria empresa se não existir
		query := `INSERT INTO empresas (cnpj, nome, ano) VALUES (?, ?, ?)`
		_, err = s.db.ExecContext(ctx, query, d.CNPJ, d.Nome, d.Ano)
		if err != nil {
			return err
		}
		err = s.db.GetContext(ctx, &idEmpresa, `SELECT id FROM empresas WHERE cnpj=?`, d.CNPJ)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Insere FRE
	query := `INSERT OR REPLACE INTO fre (
		id_empresa, data_ref, data_ultima_assembleia, id_doc,
		pct_acoes_ord_circulacao, pct_acoes_pref_circulacao, pct_total_acoes_circulacao,
		qtd_acionistas_inst, qtd_acionistas_pf, qtd_acionistas_pj,
		qtd_acoes_ord_circulacao, qtd_acoes_pref_circulacao, qtd_total_acoes_circulacao
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = s.db.ExecContext(ctx, query,
		idEmpresa,
		fre.DataRef,
		fre.DataUltimaAssembleia,
		fre.IDDoc,
		fre.PctAcoesOrdCirculacao,
		fre.PctAcoesPrefCirculacao,
		fre.PctTotalAcoesCirculacao,
		fre.QtdAcionistasInst,
		fre.QtdAcionistasPF,
		fre.QtdAcionistasPJ,
		fre.QtdAcoesOrdCirculacao,
		fre.QtdAcoesPrefCirculacao,
		fre.QtdTotalAcoesCirculacao,
	)
	if err != nil {
		return err
	}

	return nil
}
