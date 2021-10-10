package domínio

import (
	"context"
	"fmt"
	"time"
)

// Ativo --------------------------------------------------
type Ativo struct {
	Código       string
	Data         Data
	Abertura     Dinheiro
	Máxima       Dinheiro
	Mínima       Dinheiro
	Encerramento Dinheiro
	Volume       Dinheiro
}

// Ativos -------------------------------------------------
type Ativos []Ativo

// RepositórioLeituraAtivo --------------------------------------
type RepositórioLeituraAtivo interface {
	Cotação(ctx context.Context, código string, data Data) (*Ativo, error)
}

type RepositórioEscritaAtivo interface {
	Salvar(ctx context.Context, ativo *Ativo) error
}

type RepositórioLeituraEscritaAtivo interface {
	RepositórioLeituraAtivo
	RepositórioEscritaAtivo
}

// ServiçoAtivo --------------------------------------
type ServiçoAtivo interface {
	Cotação(código string, data Data) (*Ativo, error)
}

// ========================================================

// Dinheiro -----------------------------------------------
type Dinheiro struct {
	Valor float64
	Moeda string // R$, $
}

func (d Dinheiro) String() string {
	return fmt.Sprintf(`%s %.2f`, d.Moeda, d.Valor)
}

// Data ---------------------------------------------------
type Data time.Time

const layoutISO = "2006-01-02"

func (d Data) String() string { return time.Time(d).Format(layoutISO) }

func NovaData(s string) (Data, error) {
	t, err := time.Parse(layoutISO, s)
	return Data(t), err
}
