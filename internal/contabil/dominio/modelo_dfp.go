// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package dominio

import (
	"context"
)

type Hash uint32

// DFP = Demonstrações Financeiras Padronizadas de uma Empresa
type DFP struct {
	CNPJ   string
	Nome   string // Nome da empresa
	Ano    int
	Contas []Conta
}

func (d DFP) Válida() bool {
	return len(d.CNPJ) == len("17.836.901/0001-10") &&
		len(d.Nome) > 0 &&
		d.Ano >= 2000 && d.Ano < 2221 && // 2 séculos de rapina :)
		len(d.Contas) > 0
}

func init() {
	Config.GruposDFP = []string{
		"DF Individual - Balanço Patrimonial Ativo",
		"DF Consolidado - Balanço Patrimonial Ativo",
		"DF Individual - Balanço Patrimonial Passivo",
		"DF Consolidado - Balanço Patrimonial Passivo",
		"DF Individual - Demonstração do Fluxo de Caixa (Método Direto)",
		"DF Consolidado - Demonstração do Fluxo de Caixa (Método Direto)",
		"DF Individual - Demonstração do Fluxo de Caixa (Método Indireto)",
		"DF Consolidado - Demonstração do Fluxo de Caixa (Método Indireto)",
		"DF Individual - Demonstração do Resultado",
		"DF Consolidado - Demonstração do Resultado",
		"DF Individual - Demonstração de Valor Adicionado",
		"DF Consolidado - Demonstração de Valor Adicionado",
	}
}

// -- REPOSITÓRIO & SERVIÇO --

type ResultadoImportaçãoDFP struct {
	DFP   *DFP
	Error error
}

type RepositórioImportaçãoDFP interface {
	Importar(ctx context.Context, ano int) <-chan ResultadoImportaçãoDFP
}

type RepositórioLeituraDFP interface {
	Ler(ctx context.Context, cnpj string, ano int) (*DFP, error)
	Empresas(ctx context.Context, nome string) []string
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
	Empresas(nome string) []string
}
