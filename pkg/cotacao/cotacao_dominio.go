// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package cotacao

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

// PreçoTípico é a média aritmética entre o preço máximo, o preço mínimo e o
// preço de fechamento
func (a *Ativo) PreçoTípico() rapina.Dinheiro {
	return rapina.Dinheiro{
		Moeda:  a.Encerramento.Moeda,
		Valor:  (a.Máxima.Valor + a.Mínima.Valor + a.Encerramento.Valor) / 3,
		Escala: a.Encerramento.Escala,
	}
}

// VWAP é o preço ponderado médio SUM(PRECO_TÍPICO * VOLUME) / SUM(VOLUME)
func VWAP(ativos []*Ativo) rapina.Dinheiro {
	if len(ativos) == 0 {
		return rapina.Dinheiro{}
	}
	var totalPreçoTípico float64
	var totalVolume float64
	for _, ativo := range ativos {
		pt := ativo.PreçoTípico()
		totalPreçoTípico += pt.Valor * ativo.Volume
		totalVolume += ativo.Volume
	}
	var v float64
	if totalVolume > 0 {
		v = totalPreçoTípico / totalVolume
	}
	return rapina.Dinheiro{
		Moeda:  ativos[0].Encerramento.Moeda,
		Valor:  v,
		Escala: ativos[0].Encerramento.Escala,
	}
}

// Ativos -------------------------------------------------
type Ativos []Ativo

// Repositório --------------------------------------------

type Resultado struct {
	Ativo *Ativo
	Error error
}
