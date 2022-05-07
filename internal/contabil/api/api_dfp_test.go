// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	rapina "github.com/dude333/rapinav2"
	contábil "github.com/dude333/rapinav2/internal/contabil"
	"github.com/labstack/echo/v4"
)

var dfp = contábil.DemonstraçãoFinanceira{
	Empresa: rapina.Empresa{
		CNPJ: "123",
		Nome: "Web",
	},
	Ano: 2021,
	Contas: []contábil.Conta{{
		Código:       "c1",
		Descr:        "d1",
		Consolidado:  false,
		Grupo:        "g1",
		DataFimExerc: "dt1",
		OrdemExerc:   "x",
		Total: rapina.Dinheiro{
			Valor:  123,
			Escala: 1,
			Moeda:  "R$",
		},
	}},
}

type mockService struct{}

func (m *mockService) Importar(ano int, trim bool) error {
	return nil
}
func (m *mockService) Relatório(cnpj string, ano int) (*contábil.DemonstraçãoFinanceira, error) {
	d := dfp
	d.CNPJ = cnpj
	d.Ano = ano
	return &d, nil
}
func (m *mockService) Empresas(nome string) []string {
	return nil
}

func Test_htmlDFP_dfp(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/dfp")
	c.SetParamNames("cnpj")
	c.SetParamValues("5555555555")

	h := &htmlDFP{svc: &mockService{}}

	err := h.dfp(c)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("dfp() = %#v", rec.Body.String())
}
