// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package api

import (
	"database/sql"
	"net/http"
	"sort"
	"strings"
	"time"

	contábil "github.com/dude333/rapinav2/internal/contabil"
	serviço "github.com/dude333/rapinav2/internal/contabil/servico"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type Serviço interface {
	Importar(ano int, trimestral bool) error
	Relatório(cnpj string, ano int) (*contábil.DemonstraçãoFinanceira, error)
	Empresas(nome string) []string
}

type htmlDFP struct {
	svc Serviço
}

func New(e *echo.Echo, db *sqlx.DB, dataDir string) {
	svc, err := serviço.NovoDemonstraçãoFinanceira(db)
	if err != nil {
		panic(err)
	}

	handler := &htmlDFP{svc: svc}

	e.GET("/api/dfp", handler.dfp)
	e.GET("/api/dfp/empresas/:nome", handler.empresas)
}

// dfp retorna um JSON com os DFPs de uma empresa.
//
// Parâmetros:
//   - "cnpj": "string"
//   - "ordem": "asc"|"dsc" // anos em orderm ascendente ou descentente
//                          // default: "dsc"
//
// Retorno:
//	{
//		"nome": "",
//		"cnpj": "",
//		"anos": [],
//		"contas": [
//			{
//				"codigo": "",
//				"descr": "",
//				"totais": [],
//              "subcontas": [],
//			}
//		]
//	}
//
// onde anos[] está em ordem com totais[]:
// anos[ano1, ano2, ...] = totais[total_ano1, total_ano2, ...].
func (h *htmlDFP) dfp(c echo.Context) error {
	cnpj := c.QueryParam("cnpj")
	if cnpj == "" {
		cnpj = c.QueryParam("empresa")
	}
	ordem := c.QueryParam("ordem")

	var ret jsonDFP
	mapContas := make(map[string]jsonConta)
	i := 0
	for _, ano := range listaAnos(ordem) {
		dfp, err := h.svc.Relatório(cnpj, ano)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return err
		}

		ret.Anos = append(ret.Anos, ano)

		if ret.CNPJ == "" {
			ret.Nome = dfp.Nome
			ret.CNPJ = dfp.CNPJ
		}

		for _, c := range dfp.Contas {
			key := c.Código
			m, ok := mapContas[key]
			if !ok {
				m.Código = c.Código
				m.Descr = c.Descr
			}
			for j := len(m.Totais); j < i; j++ {
				m.Totais = append(m.Totais, 0) // fill missing data
			}
			m.Totais = append(m.Totais, c.Total.Valor*float64(c.Total.Escala))
			mapContas[key] = m
		}
		i++
	}

	for _, v := range mapContas {
		c := parent(v.Código)
		p, ok := mapContas[c]
		if ok {
			p.Subcontas = append(p.Subcontas, v)
			mapContas[c] = p
		}
	}

	sorted := keys(mapContas)
	sort.Strings(sorted)
	var retContas []jsonConta

	for _, k := range sorted {
		c := parent(mapContas[k].Código)
		_, ok := mapContas[c]
		if c == "" || !ok { // no parents
			retContas = append(retContas, mapContas[k])
		}
	}

	ret.Contas = sortSubcontas(retContas)

	return c.JSON(http.StatusOK, &ret)
}

type jsonDFP struct {
	Nome   string      `json:"nome"`
	CNPJ   string      `json:"cnpj"`
	Anos   []int       `json:"anos"`
	Contas []jsonConta `json:"contas"`
}

type jsonConta struct {
	Código    string      `json:"codigo"`
	Descr     string      `json:"descr"`
	Totais    []float64   `json:"totais"`
	Subcontas []jsonConta `json:"subcontas"`
}

func keys(m map[string]jsonConta) []string {
	keys := make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

// listaAnos retorna uma lista de anos.
// Se asc == "asc", retorna de 2009 até o ano atual.
// Caso contrário, retorna do ano atual até 2009.
func listaAnos(ordem string) []int {
	atual := time.Now().Year()
	esq := atual
	dir := 2009
	op := -1
	if ordem == "asc" {
		esq = 2009
		dir = atual
		op = 1
	}
	var ret []int
	for i := esq; (i >= esq && i <= dir) || (i >= dir && i <= esq); i = i + op {
		ret = append(ret, i)
	}
	return ret
}

func parent(code string) string {
	c := strings.Split(code, ".")
	if len(c) > 0 {
		c = c[:len(c)-1]
	}
	return strings.Join(c, ".")
}

func sortSubcontas(contas []jsonConta) []jsonConta {
	for k := range contas {
		if len(contas[k].Subcontas) > 0 {
			sortSubcontas(contas[k].Subcontas)
			sort.Slice(contas[k].Subcontas, func(i, j int) bool {
				return contas[k].Subcontas[i].Código < contas[k].Subcontas[j].Código
			})
		}
	}

	return contas
}

// empresas retorna um JSON como nome de empresas similares ao parâmetro 'nome'.
//
// Parâmetros:
//   - "nome": "string"
//
// Retorno:
//	{
//		"empresas": []
//	}
func (h *htmlDFP) empresas(c echo.Context) error {
	nome := c.Param("nome")

	lista := h.svc.Empresas(nome)

	ret := jsonEmpresas{Empresas: lista}

	return c.JSON(http.StatusOK, &ret)
}

type jsonEmpresas struct {
	Empresas []string `json:"empresas"`
}
