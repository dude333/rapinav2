// SPDX-FileCopyrightText: 2022 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package contábil

import (
	"strings"

	rapina "github.com/dude333/rapinav2"
)

// DemonstraçãoFinanceira contém a demonstração financeira de uma empresa
// num dado ano (contém dados acumulados desde DataIniExerc).
type DemonstraçãoFinanceira struct {
	rapina.Empresa
	Ano          int
	DataIniExerc string
	Contas       []Conta
}

func (d DemonstraçãoFinanceira) Válida() bool {
	return len(d.CNPJ) == len("17.836.901/0001-10") &&
		len(d.Nome) > 0 &&
		d.Ano >= 2000 && d.Ano < 2221 && // 2 séculos de rapina :)
		len(d.Contas) > 0
}

// Conta com os dados das Demonstrações Financeiras Padronizadas (DFP) ou
// com as Informações Trimestrais (ITR).
type Conta struct {
	Código       string // 1, 1.01, 1.02...
	Descr        string
	Consolidado  bool   // Individual ou Consolidado
	Grupo        string // BPA, BPP, DRE, DFC...
	DataIniExerc string // AAAA-MM-DD
	DataFimExerc string // AAAA-MM-DD
	Meses        int    // Meses acumulados desde o início do período
	OrdemExerc   string // ÚLTIMO ou PENÚLTIMO
	Total        rapina.Dinheiro
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

type ConfigConta struct {
	AtivoTotal        []string
	AtivoCirc         []string
	AtivoNCirc        []string
	Caixa             []string
	AplicFinanceiras  []string
	Estoque           []string
	ContasARecebCirc  []string
	ContasARecebNCirc []string
	PassivoTotal      []string
	PassivoCirc       []string
	PassivoNCirc      []string
	Equity            []string
	DividaCirc        []string
	DividaNCirc       []string
	DividendosJCP     []string
	DividendosMin     []string
	Vendas            []string
	CustoVendas       []string
	DespesasOp        []string
	EBIT              []string
	ResulFinanc       []string
	ResulOpDescont    []string
	LucLiq            []string
	FCO               []string
	FCI               []string
	FCF               []string
	Deprec            []string
	JurosCapProp      []string
	Dividendos        []string
}

// -- REPOSITÓRIO & SERVIÇO --

type Resultado struct {
	Empresa *DemonstraçãoFinanceira
	Hash    string
	Error   error
}

type Serviço interface {
	// Importar(ano int, trimestral bool) error
	Relatório(cnpj string, ano int) (*DemonstraçãoFinanceira, error)
	Empresas(nome string) []rapina.Empresa
}
