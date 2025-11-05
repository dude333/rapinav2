// SPDX-FileCopyrightText: 2023 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package rapina

import (
	"math"
	"strings"

	"github.com/dude333/rapinav2/pkg/progress"
)

// Dados Financeiros e Contábeis ----------------------------------------------

type InformeTrimestral struct {
	Codigo  string
	Descr   string
	Valores []ValoresTrimestrais
}

// ValoresTrimestrais armazena os valores ValoresTrimestrais de um determinado ano.
// Considera-se nulo o valor do trimestre se o valor for NaN.
type ValoresTrimestrais struct {
	Ano int
	T1  float64
	T2  float64
	T3  float64
	T4  float64
}

// T retorna o valor do trimestre pelo índice (1 <= n <= 4)
func (v *ValoresTrimestrais) T(n int) float64 {
	switch n {
	case 1:
		return v.T1
	case 2:
		return v.T2
	case 3:
		return v.T3
	case 4:
		return v.T4
	}
	return math.NaN()
}

// SetT salva o valor do trimestre pelo índice (1 <= n <= 4)
func (v *ValoresTrimestrais) SetT(n int, val float64) {
	switch n {
	case 1:
		v.T1 = val
	case 2:
		v.T2 = val
	case 3:
		v.T3 = val
	case 4:
		v.T4 = val
	}
}

// add soma dois números de ponto flutuante tratando NaN como zero. Se ambos
// forem NaN, o resultado será NaN.
func add(a, b float64) float64 {
	if math.IsNaN(a) {
		if math.IsNaN(b) {
			return math.NaN()
		}
		return b
	}
	if math.IsNaN(b) {
		return a
	}
	return a + b
}

// sub subtrai dois números de ponto flutuante tratando NaN como zero. Se ambos
// forem NaN, o resultado será NaN.
func sub(a, b float64) float64 {
	if math.IsNaN(a) {
		if math.IsNaN(b) {
			return math.NaN()
		}
		return -b
	}
	if math.IsNaN(b) {
		return a
	}
	return a - b
}

func (v ValoresTrimestrais) Add(other ValoresTrimestrais) ValoresTrimestrais {
	if v.Ano != other.Ano {
		return v
	}
	return ValoresTrimestrais{
		Ano: v.Ano,
		T1:  add(v.T1, other.T1),
		T2:  add(v.T2, other.T2),
		T3:  add(v.T3, other.T3),
		T4:  add(v.T4, other.T4),
	}
}

func (v ValoresTrimestrais) Sub(other ValoresTrimestrais) ValoresTrimestrais {
	if v.Ano != other.Ano {
		return v
	}
	return ValoresTrimestrais{
		Ano: v.Ano,
		T1:  sub(v.T1, other.T1),
		T2:  sub(v.T2, other.T2),
		T3:  sub(v.T3, other.T3),
		T4:  sub(v.T4, other.T4),
	}
}

func (v ValoresTrimestrais) Mult(other ValoresTrimestrais) ValoresTrimestrais {
	if v.Ano != other.Ano {
		return v
	}
	return ValoresTrimestrais{
		Ano: v.Ano,
		T1:  v.T1 * other.T1,
		T2:  v.T2 * other.T2,
		T3:  v.T3 * other.T3,
		T4:  v.T4 * other.T4,
	}
}

func (v ValoresTrimestrais) Div(other ValoresTrimestrais) ValoresTrimestrais {
	safeDiv := func(num, divisor float64) float64 {
		if divisor == 0 || math.IsNaN(divisor) || math.IsNaN(num) {
			return math.NaN()
		}
		return num / divisor
	}
	return ValoresTrimestrais{
		Ano: v.Ano,
		T1:  safeDiv(v.T1, other.T1),
		T2:  safeDiv(v.T2, other.T2),
		T3:  safeDiv(v.T3, other.T3),
		T4:  safeDiv(v.T4, other.T4),
	}
}

func (v ValoresTrimestrais) MultNum(factor float64) ValoresTrimestrais {
	return ValoresTrimestrais{
		Ano: v.Ano,
		T1:  v.T1 * factor,
		T2:  v.T2 * factor,
		T3:  v.T3 * factor,
		T4:  v.T4 * factor,
	}
}

func (v ValoresTrimestrais) DivNum(divisor float64) ValoresTrimestrais {
	if divisor == 0 {
		return ValoresTrimestrais{v.Ano, 0.0, 0.0, 0.0, 0.0}
	}
	return ValoresTrimestrais{
		Ano: v.Ano,
		T1:  v.T1 / divisor,
		T2:  v.T2 / divisor,
		T3:  v.T3 / divisor,
		T4:  v.T4 / divisor,
	}
}

func OpVTs(op rune, v1, v2 []ValoresTrimestrais) []ValoresTrimestrais {
	type par struct {
		// ano   int
		p1Idx int
		p2Idx int
		p1    bool
		p2    bool
	}
	anos := RangeAnosVTs(v1, v2)
	pares := make([]par, len(anos))
	for i := range v1 {
		pares[v1[i].Ano-anos[0]].p1 = true
		pares[v1[i].Ano-anos[0]].p1Idx = i
	}
	for j := range v2 {
		pares[v2[j].Ano-anos[0]].p2 = true
		pares[v2[j].Ano-anos[0]].p2Idx = j
	}
	v := make([]ValoresTrimestrais, 0, len(anos))
	for k := range anos {
		var p1Ptr, p2Ptr ValoresTrimestrais
		i := pares[k].p1Idx
		j := pares[k].p2Idx
		if pares[k].p1 && pares[k].p2 {
			p1Ptr = v1[i]
			p2Ptr = v2[j]
		} else if pares[k].p1 && !pares[k].p2 {
			p1Ptr = v1[i]
			p2Ptr = ValoresTrimestrais{v1[i].Ano, 0.0, 0.0, 0.0, 0.0}
		} else if !pares[k].p1 && pares[k].p2 {
			p1Ptr = ValoresTrimestrais{v2[j].Ano, 0.0, 0.0, 0.0, 0.0}
			p2Ptr = v2[j]
		} else {
			continue
		}

		r := ValoresTrimestrais{}
		switch op {
		case '+':
			r = p1Ptr.Add(p2Ptr)
		case '-':
			r = p1Ptr.Sub(p2Ptr)
		case '*':
			r = p1Ptr.Mult(p2Ptr)
		case '/':
			r = p1Ptr.Div(p2Ptr)
		}
		v = append(v, r)
	}
	return v
}

func AddVTs(v1, v2 []ValoresTrimestrais) []ValoresTrimestrais {
	return OpVTs('+', v1, v2)
}

func SubVTs(v1, v2 []ValoresTrimestrais) []ValoresTrimestrais {
	return OpVTs('-', v1, v2)
}

func MultVTs(v1, v2 []ValoresTrimestrais) []ValoresTrimestrais {
	return OpVTs('*', v1, v2)
}

func DivVTs(v1, v2 []ValoresTrimestrais) []ValoresTrimestrais {
	return OpVTs('/', v1, v2)
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
	anos := RangeAnos(itr, false)
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
			if cod1 == cod2 && Similar(cod1+itr[linha1].Descr, cod2+itr[linha2].Descr) {
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
		if !ok {
			return math.NaN(), false
		}

		v1TemValor := !math.IsNaN(v1Tn) && v1Tn != 0.0
		v2TemValor := !math.IsNaN(v2Tn) && v2Tn != 0.0

		if v1TemValor && v2TemValor {
			return math.NaN(), false // Conflito: ambos têm valores
		}

		if v1TemValor {
			return v1Tn, true
		}
		if v2TemValor {
			return v2Tn, true
		}

		// Se um for NaN e o outro não, o resultado será NaN (ausente)
		if math.IsNaN(v1Tn) || math.IsNaN(v2Tn) {
			return math.NaN(), true
		}

		return 0.0, true // Ambos são 0
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
		if (!math.IsNaN(v.T1) && v.T1 != 0) ||
			(!math.IsNaN(v.T2) && v.T2 != 0) ||
			(!math.IsNaN(v.T3) && v.T3 != 0) ||
			(!math.IsNaN(v.T4) && v.T4 != 0) {
			return false
		}
	}
	return true
}

func TrimestresComDados(itr []InformeTrimestral) []bool {
	minAno, maxAno := MinMax(itr)
	colunas := make([]bool, 4*(1+maxAno-minAno))

	for _, informe := range itr {
		for _, v := range informe.Valores {
			i := (v.Ano - minAno) * 4
			if !colunas[i+0] && !math.IsNaN(v.T1) && v.T1 != 0.0 {
				colunas[i+0] = true
			}
			if !math.IsNaN(v.T2) && v.T2 != 0.0 {
				colunas[i+1] = true
			}
			if !math.IsNaN(v.T3) && v.T3 != 0.0 {
				colunas[i+2] = true
			}
			if !math.IsNaN(v.T4) && v.T4 != 0.0 {
				colunas[i+3] = true
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
	if minAno > maxAno {
		return 0, 0
	}
	return minAno, maxAno
}

// RangeAnos retorna a sequência de anos entre o mínimo e o máximo de anos
// presentes nos InformeTrimestral. Crescente se reverse for false, senão
// decrescente.
func RangeAnos(itr []InformeTrimestral, reverse bool) []int {
	min, max := MinMax(itr)
	seq := make([]int, max-min+1)

	if reverse {
		for i := max; i >= min; i-- {
			seq[max-i] = i
		}
		return seq
	}

	for i := min; i <= max; i++ {
		seq[i-min] = i
	}
	return seq
}

func RangeAnosVTs(v1, v2 []ValoresTrimestrais) []int {
	itr := []InformeTrimestral{
		{"", "", v1},
		{"", "", v2},
	}
	return RangeAnos(itr, false)
}

// ÚltimoTrimestreReal retorna o último trimestre com valor não nulo
// usando o valor do Ativo Total como base.
func ÚltimoTrimestreReal(itrs []InformeTrimestral) int {
	for _, itr := range itrs {
		if itr.Codigo == "1" {
			max := 0
			for _, valor := range itr.Valores {
				if valor.Ano > max {
					max = valor.Ano
				}
			}
			return ÚltimoTrimestre(max, itr.Valores)
		}
	}
	return 1
}

// ÚltimoTrimestre retorna o último trimestre com valor não nulo
func ÚltimoTrimestre(ano int, valores []ValoresTrimestrais) int {
	for _, valor := range valores {
		if valor.Ano != ano {
			continue
		}
		if !math.IsNaN(valor.T4) {
			return 4
		} else if !math.IsNaN(valor.T3) {
			return 3
		} else if !math.IsNaN(valor.T2) {
			return 2
		}
	}
	return 1
}

// TTM armazena a soma dos últimos 4 trimestres em cada um dos trimestres;
// usado em métricas que comparam como valores do balanço patrimonial.
// Exemplo: ROE = Lucro Líq. dos últimos 12 meses / Patrim.Líq.
func TTM(acct []ValoresTrimestrais) []ValoresTrimestrais {
	if len(acct) == 0 {
		return []ValoresTrimestrais{}
	}

	min, max := MinMax([]InformeTrimestral{{Codigo: "", Descr: "", Valores: acct}})
	if min == 0 && max == 0 {
		return []ValoresTrimestrais{}
	}

	// Create a flat array of all quarterly values for easier access
	valores := make([]float64, (max-min+1)*4)

	for _, valor := range acct {
		if valor.Ano < min || valor.Ano > max {
			continue
		}
		idx := 4 * (valor.Ano - min)
		valores[idx+0] = nanToZero(valor.T1)
		valores[idx+1] = nanToZero(valor.T2)
		valores[idx+2] = nanToZero(valor.T3)
		valores[idx+3] = nanToZero(valor.T4)
	}

	// Funcão auxiliar para somar valores do índice 'from' até 'to' (inclusivo).
	somaValores := func(from, to int) float64 {
		if from < 0 || to >= len(valores) || from > to {
			return 0.0
		}

		total := 0.0
		for i := from; i <= to; i++ {
			total += valores[i]
		}
		return total
	}

	// Calcula o TTM para cada trimestre.
	valoresAcum := make([]ValoresTrimestrais, 0, max-min+1)

	for ano := min; ano <= max; ano++ {
		idx := 4 * (ano - min)

		// Para cada trimestre, calcula o somatório dos últimos 4 trimestres.
		vt := ValoresTrimestrais{
			Ano: ano,
			T1:  somaValores(idx-3, idx+0),
			T2:  somaValores(idx-2, idx+1),
			T3:  somaValores(idx-1, idx+2),
			T4:  somaValores(idx+0, idx+3),
		}

		valoresAcum = append(valoresAcum, vt)
	}

	return valoresAcum
}

func nanToZero(val float64) float64 {
	if math.IsNaN(val) {
		return 0.0
	}
	return val
}

// ManterÚltimoTrimestre mantém apenas o último trimestre não nulo de cada ano
func ManterÚltimoTrimestre(vts []ValoresTrimestrais) []ValoresTrimestrais {
	w := make([]ValoresTrimestrais, len(vts))
	for i, v := range vts {
		t := ÚltimoTrimestre(v.Ano, vts)
		w[i].Ano = v.Ano
		w[i].SetT(4, v.T(t))
	}
	return w
}

// Cotação --------------------------------------------------

type Cotação struct {
	Código       string
	Data         Data
	Abertura     Dinheiro
	Máxima       Dinheiro
	Mínima       Dinheiro
	Encerramento Dinheiro
	Volume       float64
}

// PreçoTípico é a média aritmética entre o preço máximo, o preço mínimo e o
// preço de fechamento
func (a *Cotação) PreçoTípico() Dinheiro {
	return Dinheiro{
		Moeda:  a.Encerramento.Moeda,
		Valor:  (a.Máxima.Valor + a.Mínima.Valor + a.Encerramento.Valor) / 3,
		Escala: a.Encerramento.Escala,
	}
}

// VWAP é o preço ponderado médio SUM(PRECO_TÍPICO * VOLUME) / SUM(VOLUME)
func VWAP(ativos []*Cotação) Dinheiro {
	if len(ativos) == 0 {
		return Dinheiro{}
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
	return Dinheiro{
		Moeda:  ativos[0].Encerramento.Moeda,
		Valor:  v,
		Escala: ativos[0].Encerramento.Escala,
	}
}
