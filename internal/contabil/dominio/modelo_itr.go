// SPDX-FileCopyrightText: 2022 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package dominio

import (
	"context"
)

// ITR = Informações Trimestrais de uma Empresa
type ITR struct {
	CNPJ   string
	Nome   string // Nome da empresa
	Ano    int
	Contas []Conta
}

func (d ITR) Válida() bool {
	return len(d.CNPJ) == len("17.836.901/0001-10") &&
		len(d.Nome) > 0 &&
		d.Ano >= 2000 && d.Ano < 2221 && // 2 séculos de rapina :)
		len(d.Contas) > 0
}

func init() {
	Config.GruposITR = []string{
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

type ResultadoImportaçãoITR struct {
	ITR   *ITR
	Error error
}

type RepositórioImportaçãoITR interface {
	Importar(ctx context.Context, ano int) <-chan ResultadoImportaçãoITR
}

type RepositórioLeituraITR interface {
	Ler(ctx context.Context, cnpj string, ano int) (*ITR, error)
	Empresas(ctx context.Context, nome string) []string
}

type RepositórioEscritaITR interface {
	Salvar(ctx context.Context, empresa *ITR) error
}

type RepositórioLeituraEscritaITR interface {
	RepositórioLeituraITR
	RepositórioEscritaITR
}

type ServiçoITR interface {
	Importar(ano int) error
	Relatório(cnpj string, ano int) (*ITR, error)
	Empresas(nome string) []string
}
