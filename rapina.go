// SPDX-FileCopyrightText: 2023 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package rapina

import (
	"strings"

	"github.com/dude333/rapinav2/pkg/progress"
)

type InformeTrimestral struct {
	Codigo  string
	Descr   string
	Valores []ValoresTrimestrais
}

type ValoresTrimestrais struct {
	Ano int
	T1  float64
	T2  float64
	T3  float64
	T4  float64
}

func codPai(codigo string) string {
	if len(codigo) < 1 {
		return codigo
	}
	lvl := strings.Count(codigo, ".") + 1
	if lvl <= 3 {
		return codigo
	}
	idx := strings.LastIndex(codigo, ".")
	if idx <= 0 {
		return codigo
	}
	return codigo[:idx]
}

// UnificarContasSimilares unifica as linhas similares do InformeTrimestral
// comparando o código, sem o último grupo (ex.: 1.02.05.01 => 1.02.05),
// com as próximas linhas.
// Cada linha (InformeTrimestral) possui o seguinte formato:
// Linha n => [Ano:ano Valor trimestre 1 | Valor T2 | Valor T3 | Valor T4]
// Exemplo:
// "Tributo a recuperar"  => [2019 1|0|5|3; 2021 5|2|0|0]
// "Tributos a recuperar" => [2019 0|2|0|0; 2020 1|4|2|2; 2021 0|0|1|2]
// Resultado:
// "Tributo a recuperar"  => [2019 1|2|5|3; 2020 1|4|2|2; 2021 5|2|1|2]
func UnificarContasSimilares(itr []InformeTrimestral) []InformeTrimestral {
	itrUnificado := make([]InformeTrimestral, 1, len(itr))
	unida := make([]bool, len(itr))
	anos := RangeAnos(itr)
	ultimaLinha := len(itr) - 1
	for linha1 := 0; linha1 <= ultimaLinha; linha1++ {
		if unida[linha1] {
			continue
		}
		valoresUnificados := itr[linha1].Valores
		for linha2 := linha1 + 1; linha2 <= ultimaLinha; linha2++ {
			if unida[linha2] {
				continue
			}
			cod1 := codPai(itr[linha1].Codigo)
			cod2 := codPai(itr[linha2].Codigo)
			if Similar(cod1+itr[linha1].Descr, cod2+itr[linha2].Descr) {
				unida[linha2] = true
				for _, ano := range anos {
					v1, existe1 := valorAno(ano, valoresUnificados)
					v2, existe2 := valorAno(ano, itr[linha2].Valores)

					if !existe1 && existe2 {
						valoresUnificados = append(valoresUnificados, v2)
					}
					if existe1 && existe2 {
						v, ok := equalizarValores(ano, v1, v2)
						unida[linha2] = ok
						if ok {
							valoresUnificados = append(valoresUnificados, v)
						} else {
							break
						}
					}
				} // next ano
				if unida[linha2] {
					progress.Trace("Joining:\n\t+ %v\n\t+ %v\n\t", itr[linha1], itr[linha2])
				}
			}
		} // next linha2
		informe := InformeTrimestral{
			Codigo:  itr[linha1].Codigo,
			Descr:   itr[linha1].Descr,
			Valores: valoresUnificados,
		}
		itrUnificado = append(itrUnificado, informe)
	} // next linha1
	return itrUnificado
}

func equalizarValores(ano int, v1, v2 ValoresTrimestrais) (ValoresTrimestrais, bool) {
	var v ValoresTrimestrais
	v.Ano = ano
	ok := true

	check := func(v1Tn, v2Tn float64) (float64, bool) {
		if !ok || (v1Tn != 0.0 && v2Tn != 0.0) {
			return 0.0, false
		}
		if v1Tn != 0.0 && v2Tn == 0.0 {
			return v1Tn, true
		} else {
			return v2Tn, true
		}
	}

	v.T1, ok = check(v1.T1, v2.T1)
	v.T2, ok = check(v1.T2, v2.T2)
	v.T3, ok = check(v1.T3, v2.T3)
	v.T4, ok = check(v1.T4, v2.T4)

	return v, ok
}

func valorAno(ano int, valores []ValoresTrimestrais) (ValoresTrimestrais, bool) {
	for _, v := range valores {
		if v.Ano == ano {
			return v, true
		}
	}
	return ValoresTrimestrais{}, false
}

func Zerado(valores []ValoresTrimestrais) bool {
	for _, v := range valores {
		if v.T1 != 0 || v.T2 != 0 || v.T3 != 0 || v.T4 != 0 {
			return false
		}
	}
	return true
}

func TrimestresComDados(itr []InformeTrimestral) []bool {
	minAno, maxAno := MinMax(itr)
	colunas := make([]bool, 4*(1+maxAno-minAno))

	for _, informe := range itr {
		for ano := minAno; ano <= maxAno; ano++ {
			for _, v := range informe.Valores {
				if v.Ano != ano {
					continue
				}
				i := (v.Ano - minAno) * 4
				if !colunas[i+0] && v.T1 != 0.0 {
					colunas[i+0] = true
				}
				if v.T2 != 0.0 {
					colunas[i+1] = true
				}
				if v.T3 != 0.0 {
					colunas[i+2] = true
				}
				if v.T4 != 0.0 {
					colunas[i+3] = true
				}
			}
		}
	}
	return colunas
}

func MinMax(itr []InformeTrimestral) (int, int) {
	minAno := 99999
	maxAno := 0
	for i := range itr {
		for _, valores := range itr[i].Valores {
			if valores.Ano < minAno {
				minAno = valores.Ano
			}
			if valores.Ano > maxAno {
				maxAno = valores.Ano
			}
		}
	}

	return minAno, maxAno
}

func RangeAnos(itr []InformeTrimestral) []int {
	min, max := MinMax(itr)
	seq := make([]int, max-min+1)
	for i := min; i <= max; i++ {
		seq[i-min] = i
	}
	return seq
}
