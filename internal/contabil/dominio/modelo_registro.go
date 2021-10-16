// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package domínio

import (
	"context"
	"fmt"
)

type Hash uint32

const (
	Individual int = iota
	Consolidado
)

// DFP = Demonstrações Financeiras Padronizadas de uma Empresa
type DFP struct {
	CNPJ   string
	Nome   string // Nome da empresa
	Ano    int
	Contas []Conta
}

type Conta struct {
	Código       string
	Descr        string
	GrupoDFP     string
	DataFimExerc string // AAAA-MM-DD
	Total        Dinheiro
}

type Dinheiro struct {
	Valor  float64
	Escala int
	Moeda  string
}

func (d Dinheiro) String() string {
	return fmt.Sprintf(`%s %.2f`, d.Moeda, d.Valor*float64(d.Escala))
}

type ResultadoImportação struct {
	DFP   *DFP
	Error error
}

type RepositórioImportaçãoDFP interface {
	Importar(ctx context.Context, ano int) <-chan ResultadoImportação
}

type RepositórioLeituraDFP interface {
	Ler(ctx context.Context, cnpj string, ano int) (*DFP, error)
}

type RepositórioEscritaDFP interface {
	Salvar(ctx context.Context, empresa *DFP) error
}

type RepositórioLeituraEscritaDFP interface {
	RepositórioLeituraDFP
	RepositórioEscritaDFP
}

type ServiçoDFP interface {
	Importar(ano int) error
	Relatório(cnpj string, ano int) (*DFP, error)
}
