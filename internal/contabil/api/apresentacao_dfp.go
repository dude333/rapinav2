// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package api

import (
	domínio "github.com/dude333/rapinav2/internal/contabil/dominio"
	repositório "github.com/dude333/rapinav2/internal/contabil/repositorio"
	serviço "github.com/dude333/rapinav2/internal/contabil/servico"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type htmlDFP struct {
	svc domínio.ServiçoDFP
}

func New(e *echo.Echo, db *sqlx.DB, dataDir string) {

	sqlite, _ := repositório.NovoSqlite(db)
	api := repositório.NovoCVM(dataDir)
	svc := serviço.NovoDFP(api, sqlite)
	handler := &htmlDFP{svc: svc}

	e.GET("/api/lucros", handler.lucros)
}

func (h *htmlDFP) lucros(c echo.Context) error {
	cnpj := c.QueryParam("cnpj")
	ano, _ := strconv.Atoi(c.QueryParam("ano"))

	dfp, err := h.svc.Relatório(cnpj, ano)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &dfp)
}
