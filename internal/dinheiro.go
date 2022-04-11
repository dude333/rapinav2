// SPDX-FileCopyrightText: 2022 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package rapina

import "fmt"

type Dinheiro struct {
	Valor  float64
	Escala int
	Moeda  string
}

func (d Dinheiro) String() string {
	return fmt.Sprintf(`%s %.2f`, d.Moeda, d.Valor*float64(d.Escala))
}
