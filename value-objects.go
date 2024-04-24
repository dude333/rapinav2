// SPDX-FileCopyrightText: 2022 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package rapina

import (
	"errors"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Empresa ------------------------------------------------
type Empresa struct {
	CNPJ string `json:"cnpj"`
	Nome string `json:"nome"`
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
	e := float64(d.Escala)
	if e == 0 {
		e = 1
	}
	return p.Sprintf(`%s %.2f`, d.Moeda, e*d.Valor)
}

// Data ---------------------------------------------------
type Data time.Time

const layoutISO = "2006-01-02"

var ErrDataInválida = errors.New("data inválida")

func (d Data) String() string { return time.Time(d).Format(layoutISO) }

// NovaData converte uma string no formato "AAAA-MM-DD" em Data.
func NovaData(s string) (Data, error) {
	// Verificar se a string está no formato AAAA-MM-DD
	if len(s) != len("AAAA-MM-DD") && (s[4] != '-' || s[7] != '-') {
		return Data(time.Time{}), ErrDataInválida
	}
	t, err := time.Parse(layoutISO, s)
	return Data(t), err
}
