// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package api

import (
	"database/sql"
	"net/http"
	"sort"
	"time"

	"github.com/dude333/rapinav2/internal/contabil/dominio"
	"github.com/dude333/rapinav2/internal/contabil/repositorio"
	"github.com/dude333/rapinav2/internal/contabil/servico"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type htmlDFP struct {
	svc dominio.ServiçoDFP
}

func New(e *echo.Echo, db *sqlx.DB, dataDir string) {

	sqlite, _ := repositorio.NovoSqlite(db)
	api := repositorio.NovoCVM(dataDir)
	svc := servico.NovoDFP(api, sqlite)
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
			key := c.Código + " " + c.Descr
			m, ok := mapContas[key]
			if !ok {
				m.Código = c.Código
				m.Descr = c.Descr
			}
			m.Totais = append(m.Totais, c.Total.Valor*float64(c.Total.Escala))
			mapContas[key] = m
		}
	}

	sorted := keys(mapContas)
	sort.Strings(sorted)
	retContas := make([]jsonConta, len(sorted))
	for i, k := range sorted {
		retContas[i] = mapContas[k]
	}

	ret.Contas = retContas

	return c.JSON(http.StatusOK, &ret)
}

type jsonDFP struct {
	Nome   string      `json:"nome"`
	CNPJ   string      `json:"cnpj"`
	Anos   []int       `json:"anos"`
	Contas []jsonConta `json:"contas"`
}

type jsonConta struct {
	Código string    `json:"codigo"`
	Descr  string    `json:"descr"`
	Totais []float64 `json:"totais"`
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
