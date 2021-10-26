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
//				"totais": []
//			}
//		]
//	}
//
// onde anos[] está em ordem com totais[]:
// anos[ano1, ano2, ...] = totais[total_ano1, total_ano2, ...].
func (h *htmlDFP) dfp(c echo.Context) error {
	cnpj := c.QueryParam("cnpj")
	ordem := c.QueryParam("ordem")

	atual := time.Now().Year()
	esq := atual
	dir := 2009
	op := -1
	if ordem == "asc" {
		esq = 2009
		dir = atual
		op = 1
	}
	inRange := func(a, b int) bool {
		if op == -1 {
			return a >= b
		}
		return a <= b
	}

	var ret jsonDFP
	mapContas := make(map[string]jsonConta)
	n := 0
	for ano := esq; inRange(ano, dir); ano = ano + op {
		dfp, err := h.svc.Relatório(cnpj, ano)
		if err == sql.ErrNoRows {
			n++
			continue
		}
		if err != nil {
			return err
		}

		if ret.CNPJ == "" {
			ret.Nome = dfp.Nome
			ret.CNPJ = dfp.CNPJ
		}

		ret.Anos = append(ret.Anos, ano)

		for _, c := range dfp.Contas {
			key := c.Código + c.Descr
			m, ok := mapContas[key]
			if !ok {
				m.Código = c.Código
				m.Descr = c.Descr
				m.Totais = make([]string, abs(esq-dir)+1)
			}
			m.Totais[n] = c.Total.String()
			mapContas[key] = m
		}
		n++
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
	Código string   `json:"codigo"`
	Descr  string   `json:"descr"`
	Totais []string `json:"totais"`
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
