package cotacao

import (
	"context"

	rapina "github.com/dude333/rapinav2"
	ext "github.com/dude333/rapinav2/pkg/infra"
	"github.com/jmoiron/sqlx"
)

type Sqlite struct {
	db *sqlx.DB
}

func NovoSqlite(db *sqlx.DB) (*Sqlite, error) {
	if db == nil {
		return nil, ErrRepositórioInválido
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := ext.CriarTabelas(db, tabelas); err != nil {
		return nil, err
	}
	return &Sqlite{db: db}, nil
}

func (s *Sqlite) LerCotações(ctx context.Context, código string, dia rapina.Data) ([]*Ativo, error) {
	query := "SELECT codigo, data, moeda, encerramento, volume FROM cotacao WHERE codigo LIKE ? AND data=?"
	var rows []tabelaCotação
	err := s.db.SelectContext(ctx, &rows, query, código+"%", dia.String())
	if len(rows) == 0 {
		return nil, ErrCotaçãoNãoEncontrada
	}
	if err != nil {
		return nil, err
	}

	ativos := make([]*Ativo, 0, len(rows))
	for _, row := range rows {
		ativo, err := converteParaAtivo(&row)
		if err != nil {
			return nil, err
		}
		ativos = append(ativos, ativo)
	}
	return ativos, nil
}

func (s *Sqlite) SalvarCotações(ctx context.Context, ativos []*Ativo) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for _, ativo := range ativos {
		row := converteParaTabela(ativo)
		insert := "INSERT OR IGNORE INTO cotacoes (codigo, data, moeda, abertura, maxima, minima, encerramento, volume) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
		_, err := tx.Exec(insert, row.codigo, row.data, row.moeda, row.abertura, row.maxima, row.minima, row.encerramento, row.volume)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

type tabelaCotação struct {
	codigo       string
	data         string
	moeda        string
	abertura     float64
	maxima       float64
	minima       float64
	encerramento float64
	volume       float64
}

func converteParaTabela(ativo *Ativo) *tabelaCotação {
	m := "R$"
	if len(ativo.Encerramento.Moeda) != 0 {
		m = ativo.Encerramento.Moeda
	}
	return &tabelaCotação{
		codigo:       ativo.Código,
		data:         ativo.Data.String(),
		moeda:        m,
		abertura:     ativo.Abertura.Total(),
		maxima:       ativo.Máxima.Total(),
		minima:       ativo.Mínima.Total(),
		encerramento: ativo.Encerramento.Total(),
		volume:       ativo.Volume,
	}
}

func converteParaAtivo(row *tabelaCotação) (*Ativo, error) {
	d, err := rapina.NovaData(row.data)
	if err != nil {
		return nil, err
	}
	return &Ativo{
		Código:       row.codigo,
		Data:         d,
		Encerramento: rapina.NovoDinheiro(row.moeda, row.encerramento, 1),
		Volume:       row.volume,
	}, nil
}

const __ver__ = 17

var tabelas = []ext.Tabela{
	{
		Nome:   "cotacoes",
		Versão: __ver__,
		Up: `CREATE TABLE IF NOT EXISTS cotacoes (
      codigo         TEXT NOT NULL,
      data           TEXT NOT NULL,
      moeda          TEXT DEFAULT 'R$' NOT NULL,
      abertura       REAL,
      maxima         REAL,
      minima         REAL,
      encerramento   REAL NOT NULL,
      volume         REAL NOT NULL,
      PRIMARY KEY (codigo, data)
    );`,
		Down: `DROP TABLE IF EXISTS cotacoes`,
	},
}
