// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package dominio

import (
	"context"
	"fmt"
	"strings"
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

type Conta struct {
	Código       string
	Descr        string
	Consolidado  bool // Individual ou Consolidado
	GrupoDFP     string
	DataFimExerc string // AAAA-MM-DD
	OrdemExerc   string
	Total        Dinheiro
}

// Válida retorna verdadeiro se os dados da conta são válidos. Ignora os registros
// do penúltimo ano, com exceção de 2009, uma vez que a CVM só disponibliza (pelo
// menos em 2021) dados até 2010.
func (c Conta) Válida() bool {
	return len(c.Código) > 0 &&
		len(c.Descr) > 0 &&
		len(c.DataFimExerc) == len("AAAA-MM-DD") &&
		(c.OrdemExerc == "ÚLTIMO" ||
			(c.OrdemExerc == "PENÚLTIMO" && strings.HasPrefix(c.DataFimExerc, "2009")))
}

type Dinheiro struct {
	Valor  float64
	Escala int
	Moeda  string
}

func (d Dinheiro) String() string {
	return fmt.Sprintf(`%s %.2f`, d.Moeda, d.Valor*float64(d.Escala))
}

type config struct {
	GruposDFP []string
}

var Config = config{}

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
