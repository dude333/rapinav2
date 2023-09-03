// SPDX-FileCopyrightText: 2022 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package rapina

import (
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Empresa ------------------------------------------------
type Empresa struct {
	CNPJ string
	Nome string
}

func (e Empresa) String() string {
	return e.CNPJ + " - " + e.Nome
}

// Dinheiro -----------------------------------------------
type Dinheiro struct {
	Moeda  string
	Valor  float64
	Escala int
}

func (d Dinheiro) String() string {
	p := message.NewPrinter(language.BrazilianPortuguese)
	return p.Sprintf(`%s %.2f`, d.Moeda, d.Valor*float64(d.Escala))
}

// Data ---------------------------------------------------
type Data time.Time

const layoutISO = "2006-01-02"

func (d Data) String() string { return time.Time(d).Format(layoutISO) }

func NovaData(s string) (Data, error) {
	t, err := time.Parse(layoutISO, s)
	return Data(t), err
}
