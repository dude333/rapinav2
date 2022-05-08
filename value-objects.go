// SPDX-FileCopyrightText: 2022 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package rapina

import (
	"fmt"
	"time"
)

// Empresa ------------------------------------------------
type Empresa struct {
	CNPJ string
	Nome string
}

// Dinheiro -----------------------------------------------
type Dinheiro struct {
	Valor  float64
	Escala int
	Moeda  string
}

func (d Dinheiro) String() string {
	return fmt.Sprintf(`%s %.2f`, d.Moeda, d.Valor*float64(d.Escala))
}

// Data ---------------------------------------------------
type Data time.Time

const layoutISO = "2006-01-02"

func (d Data) String() string { return time.Time(d).Format(layoutISO) }

func NovaData(s string) (Data, error) {
	t, err := time.Parse(layoutISO, s)
	return Data(t), err
}
