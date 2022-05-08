// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package cotação

import rapina "github.com/dude333/rapinav2"

// Ativo --------------------------------------------------
type Ativo struct {
	Código       string
	Data         rapina.Data
	Abertura     rapina.Dinheiro
	Máxima       rapina.Dinheiro
	Mínima       rapina.Dinheiro
	Encerramento rapina.Dinheiro
	Volume       float64
}

// Ativos -------------------------------------------------
type Ativos []Ativo

// Repositório --------------------------------------------

type Resultado struct {
	Ativo *Ativo
	Error error
}
