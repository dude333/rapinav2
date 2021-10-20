package repositório

import (
	"context"
	"database/sql"
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
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
}

func NovoSqlite(db *sqlx.DB) (contábil.RepositórioLeituraEscritaDFP, error) {
	err := criarTabelas(db)
	if err != nil {
		return nil, err
	}

	limpo := make(map[string]bool)

	return &sqlite{db: db, limpo: limpo}, nil
}

func (s *sqlite) Ler(ctx context.Context, cnpj string, ano int) (*contábil.DFP, error) {

	return nil, nil
}

func (s *sqlite) Salvar(ctx context.Context, empresa *contábil.DFP) error {
	// progress.Status("%-60s %4d\n", empresa.Nome, len(empresa.Contas))

	return s.inserirOuAtualizarDFP(ctx, empresa)
}

type sqliteDFP struct {
	CNPJ string `db:"cnpj"`
	Nome string `db:"nome"`
	Ano  int    `db:"ano"`
}

type sqliteConta struct {
	ID           int     `db:"dfp_id"`
	Código       string  `db:"codigo"`
	Descr        string  `db:"descr"`
	GrupoDFP     string  `db:"grupo_dfp"`
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

	for i := range contas {

		c := sqliteConta{
			ID:           id,
			Código:       contas[i].Código,
			Descr:        contas[i].Descr,
			GrupoDFP:     contas[i].GrupoDFP,
			DataFimExerc: contas[i].DataFimExerc,
			Valor:        contas[i].Total.Valor,
			Escala:       contas[i].Total.Escala,
			Moeda:        contas[i].Total.Moeda,
		}

		query := `INSERT INTO contas 
			( dfp_id,  codigo,  descr,  grupo_dfp,  data_fim_exerc,  valor,  escala,  moeda)
			VALUES 
			(:dfp_id, :codigo, :descr, :grupo_dfp, :data_fim_exerc, :valor, :escala, :moeda)`
		_, err := db.NamedExecContext(ctx, query, c)
		if err != nil {
			// Ignora erro em caso de registro duplicado (dfp_id + codigo), pois se
			// trata de erro no arquivo da CVM (raramente acontece)
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.Code != sqlite3.ErrConstraint {
				return err
			}
		}
		progress.Spinner()
	}

	return nil
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
		versão: 3,
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
		versão: 3,
		up: `CREATE TABLE IF NOT EXISTS contas (
			dfp_id         INTEGER,
			codigo         VARCHAR NOT NULL,
			descr          VARCHAR NOT NULL,
			grupo_dfp      VARCHAR NOT NULL,
			data_fim_exerc VARCHAR NOT NULL,
			valor          REAL NOT NULL,
			escala         INTEGER NOT NULL,
			moeda          VARCHAR,
			PRIMARY KEY (dfp_id, codigo)
		)`,
		down: `DROP TABLE contas`,
	},
}

const sqlCreateTableTabelas = `CREATE TABLE IF NOT EXISTS tabelas (
		nome   VARCHAR PRIMARY KEY,
		versao INTEGER NOT NULL
	)`

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
