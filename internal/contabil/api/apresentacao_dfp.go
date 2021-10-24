// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package api

import (
	"net/http"
	"strconv"

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
