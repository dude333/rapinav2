package cotacao

import (
	"context"

	rapina "github.com/dude333/rapinav2"
	ext "github.com/dude333/rapinav2/pkg/infra"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
)

const _ver_ = 2

var tabelas []ext.Tabela

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

func (s *Sqlite) LerCotações(ctx context.Context, código string, dia rapina.Data) ([]*rapina.Cotação, error) {
	query := "SELECT codigo, data, moeda, abertura, maxima, minima, encerramento, volume FROM cotacoes WHERE codigo LIKE ? AND data=?"
	var rows []tabelaCotação
	err := s.db.SelectContext(ctx, &rows, query, código+"%", dia.String())
	if err != nil {
		progress.ErrorMsg("Erro ao ler cotação do bd: %v", err)
		return nil, err
	}
	if len(rows) == 0 {
		progress.Status("Cotação não encontrada no bd")
		return nil, ErrCotaçãoNãoEncontrada
	}

	ativos := make([]*rapina.Cotação, 0, len(rows))
	for _, row := range rows {
		ativo, err := converteParaAtivo(&row)
		if err != nil {
			return nil, err
		}
		ativos = append(ativos, ativo)
	}
	return ativos, nil
}

func (s *Sqlite) SalvarCotações(ctx context.Context, ativos []*rapina.Cotação) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for _, ativo := range ativos {
		row := converteParaTabela(ativo)
		insert := "INSERT OR IGNORE INTO cotacoes (codigo, data, moeda, abertura, maxima, minima, encerramento, volume) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
		_, err := tx.Exec(insert, row.Codigo, row.Data, row.Moeda, row.Abertura, row.Maxima, row.Minima, row.Encerramento, row.Volume)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

type tabelaCotação struct {
	Codigo       string
	Data         string
	Moeda        string
	Abertura     float64
	Maxima       float64
	Minima       float64
	Encerramento float64
	Volume       float64
}

func converteParaTabela(ativo *rapina.Cotação) *tabelaCotação {
	m := "R$"
	if len(ativo.Encerramento.Moeda) != 0 {
		m = ativo.Encerramento.Moeda
	}
	return &tabelaCotação{
		Codigo:       ativo.Código,
		Data:         ativo.Data.String(),
		Moeda:        m,
		Abertura:     ativo.Abertura.Total(),
		Maxima:       ativo.Máxima.Total(),
		Minima:       ativo.Mínima.Total(),
		Encerramento: ativo.Encerramento.Total(),
		Volume:       ativo.Volume,
	}
}

func converteParaAtivo(row *tabelaCotação) (*rapina.Cotação, error) {
	d, err := rapina.NovaData(row.Data)
	if err != nil {
		return nil, err
	}
	return &rapina.Cotação{
		Código:       row.Codigo,
		Data:         d,
		Abertura:     rapina.NovoDinheiro(row.Moeda, row.Abertura, 1),
		Máxima:       rapina.NovoDinheiro(row.Moeda, row.Maxima, 1),
		Mínima:       rapina.NovoDinheiro(row.Moeda, row.Minima, 1),
		Encerramento: rapina.NovoDinheiro(row.Moeda, row.Encerramento, 1),
		Volume:       row.Volume,
	}, nil
}

func init() {
	tabelas = append(tabelas, []ext.Tabela{
		{
			Nome:   "cotacoes",
			Versão: _ver_,
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
	}...)
}
