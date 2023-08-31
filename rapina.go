package rapina

import (
	"fmt"
	"strings"

	"github.com/dude333/rapinav2/pkg/progress"
)

type InformeTrimestral struct {
	Codigo  string
	Descr   string
	Valores []ValoresTrimestrais
}

type ValoresTrimestrais struct {
	Ano   int
	T1    float64
	T2    float64
	T3    float64
	T4    float64
	Anual float64
}

func UnificarContasSimilares(itr []InformeTrimestral) []InformeTrimestral {
	itr2 := make([]InformeTrimestral, 1, len(itr))
	anos := rangeAnos(itr)
	ultimaLinha := len(itr) - 1
	for linha := 1; linha <= ultimaLinha; linha++ {
		unir := false
		if Similar(itr[linha-1].Descr, itr[linha].Descr) {
			unir = true
			var novosValores []ValoresTrimestrais
			for _, ano := range anos {
				v1, existe1 := valorAno(ano, itr[linha-1].Valores)
				v2, existe2 := valorAno(ano, itr[linha].Valores)

				if existe1 && !existe2 {
					novosValores = append(novosValores, v1)
				}
				if !existe1 && existe2 {
					novosValores = append(novosValores, v2)
				}
				if existe1 && existe2 {
					var v ValoresTrimestrais
					v.Ano = ano
					if (v1.T1 != 0.0 && v2.T1 != 0.0) ||
						(v1.T2 != 0.0 && v2.T2 != 0.0) ||
						(v1.T3 != 0.0 && v2.T3 != 0.0) ||
						(v1.T4 != 0.0 && v2.T4 != 0.0) ||
						(v1.Anual != 0.0 && v2.Anual != 0.0) {
						itr2 = append(itr2, itr[linha-1])
						unir = false
						break
					}

					if v1.T1 != 0.0 && v2.T1 == 0.0 {
						v.T1 = v1.T1
					} else {
						v.T1 = v2.T1
					}
					if v1.T2 != 0.0 && v2.T2 == 0.0 {
						v.T2 = v1.T2
					} else {
						v.T2 = v2.T2
					}
					if v1.T3 != 0.0 && v2.T3 == 0.0 {
						v.T3 = v1.T3
					} else {
						v.T3 = v2.T3
					}
					if v1.T4 != 0.0 && v2.T4 == 0.0 {
						v.T4 = v1.T4
					} else {
						v.T4 = v2.T4
					}
					if v1.Anual != 0.0 && v2.Anual == 0.0 {
						v.Anual = v1.Anual
					} else {
						v.Anual = v2.Anual
					}
					novosValores = append(novosValores, v)
				}
			} // next ano

			if linha == ultimaLinha {
				itr2 = append(itr2, itr[linha])
			}
			if unir {
				informe := InformeTrimestral{
					Codigo:  itr[linha-1].Codigo,
					Descr:   itr[linha-1].Descr,
					Valores: novosValores,
				}
				itr[linha].Valores = novosValores
				itr2 = append(itr2, informe)
				progress.Trace("Joining:\n\t+ %v\n\t+ %v\n\t= %v\n\n", itr[linha-1], itr[linha], informe)
				linha++
			}
		} else {
			itr2 = append(itr2, itr[linha-1])
			if linha == ultimaLinha {
				itr2 = append(itr2, itr[linha])
			}
		}
	}
	return itr2
}

// func existeAno(ano int, valores []ValoresTrimestrais) bool {
// 	for _, v := range valores {
// 		if v.Ano == ano {
// 			return true
// 		}
// 	}
// 	return false
// }

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
		if v.T1 != 0 || v.T2 != 0 || v.T3 != 0 || v.T4 != 0 || v.Anual != 0 {
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
					fmt.Printf("- Ano=%d, col=%d, %.2f\n", ano, i, v.T1)
				}
				if v.T2 != 0.0 {
					colunas[i+1] = true
				}
				if v.T3 != 0.0 {
					colunas[i+2] = true
				}
				if strings.HasPrefix(informe.Codigo, "1") || strings.HasPrefix(informe.Codigo, "2") {
					if v.Anual != 0.0 {
						colunas[i+3] = true
					}
				} else {
					if v.T4 != 0.0 {
						colunas[i+3] = true
					}
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

func rangeAnos(itr []InformeTrimestral) []int {
	min, max := MinMax(itr)
	seq := make([]int, max-min+1)
	for i := min; i <= max; i++ {
		seq[i-min] = i
	}
	return seq
}
