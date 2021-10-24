// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package dominio

import "testing"

func benchmarkVálida(c Conta, b *testing.B) {
	for n := 0; n < b.N; n++ {
		c.Válida()
	}
}

var cc = []Conta{
	{
		Código:       "C1",
		Descr:        "D1",
		Consolidado:  true,
		GrupoDFP:     "DF Individual - Balanço Patrimonial Passivo",
		DataFimExerc: "2020-12-31",
		OrdemExerc:   "ÚLTIMO",
		Total: Dinheiro{
			Valor:  123.45,
			Escala: 1000,
			Moeda:  "R$",
		},
	},
	{
		Código:       "C2",
		Descr:        "D2",
		Consolidado:  false,
		GrupoDFP:     "DF Consolidado - Demonstração do Fluxo de Caixa (Método Direto)",
		DataFimExerc: "2020-12-31",
		OrdemExerc:   "ÚLTIMO",
		Total: Dinheiro{
			Valor:  123.45,
			Escala: 1000,
			Moeda:  "R$",
		},
	},
	{
		Código:       "C3",
		Descr:        "D3",
		GrupoDFP:     "DF Consolidado - Demonstração de Valor Adicionado",
		DataFimExerc: "2020-12-31",
		OrdemExerc:   "ÚLTIMO",
		Total: Dinheiro{
			Valor:  123.45,
			Escala: 1000,
			Moeda:  "R$",
		},
	},
}

func BenchmarkVálida0(b *testing.B) { benchmarkVálida(cc[0], b) }
func BenchmarkVálida1(b *testing.B) { benchmarkVálida(cc[1], b) }
func BenchmarkVálida2(b *testing.B) { benchmarkVálida(cc[2], b) }
