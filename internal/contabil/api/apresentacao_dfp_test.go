// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package api

import (
	domínio "github.com/dude333/rapinav2/internal/contabil/dominio"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
	"testing"
)

var dfp = domínio.DFP{
	CNPJ: "123",
	Nome: "Web",
	Ano:  2021,
	Contas: []domínio.Conta{{
		Código:       "c1",
		Descr:        "d1",
		Consolidado:  false,
		GrupoDFP:     "g1",
		DataFimExerc: "dt1",
		OrdemExerc:   "x",
		Total: domínio.Dinheiro{
			Valor:  123,
			Escala: 1,
			Moeda:  "R$",
		},
	}},
}

type mockService struct{}

func (m *mockService) Importar(ano int) error {
	return nil
}
func (m *mockService) Relatório(cnpj string, ano int) (*domínio.DFP, error) {
	// dfp.CNPJ = cnpj
	// dfp.Ano = ano
	return &dfp, nil
}

func Test_htmlDFP_lucros(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/lucros")
	c.SetParamNames("cnpj", "ano")
	c.SetParamValues("5555555555", "2010")

	h := &htmlDFP{svc: &mockService{}}

	err := h.lucros(c)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("lucros() = %#v", rec.Body.String())
}
