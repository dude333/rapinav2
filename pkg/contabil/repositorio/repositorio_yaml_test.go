// SPDX-FileCopyrightText: 2022 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

import "testing"

func TestUnmarshal(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "should work",
			args: args{
				data: fileContent,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Unmarshal(tt.args.data)
		})
	}
}

const fileContent = `
dataSrc: "/home/adr/var/rapina/rapina.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
tempDir: "/home/adr/var/rapina"

modelos:
  global:
    # BPA
    AtivoTotal: 
    - ["1", "Ativo Total"]
    AtivoCirc: 
    - ["1.01", "Ativo Circulante"]
    AtivoNCirc: 
    - ["1.02", "Ativo Não Circulante"]
    Caixa: 
    - ["1.01.01", "Caixa e Equivalentes de Caixa"]
    AplicFinanceiras: 
    - ["1.01.02", "Aplicações Financeiras"]
    Estoque: 
    - ["1.01.04", "Estoques"]
    ContasARecebCirc: 
    - ["1.01.03", "Contas a Receber"]
    ContasARecebNCirc: 
    - ["1.02.01.03", "Contas a Receber"]
    - ["1.02.01.04", "Contas a Receber"]

    # BPP
    PassivoTotal:
    - ["2", "Passivo Total"]
    PassivoCirc:
    - ["2.01", "Passivo Circulante"]
    PassivoNCirc:
    - ["2.02", "Passivo Não Circulante"]
    Equity:
    - ["2.*", "Patrimônio Líquido Consolidado"]
    DividaCirc:
    - ["2.01.04", "Empréstimos e Financiamentos"]
    DividaNCirc:
    - ["2.02.01", "Empréstimos e Financiamentos"]
    DividendosJCP:
    - ["2.01.05.02.01", "Dividendos e JCP a Pagar"]
    DividendosMin:
    - ["2.01.05.02.02", "Dividendo Mínimo Obrigatório a Pagar"]

    # DRE
    Vendas:
    - ["3.01", ""]
    CustoVendas:
    - ["3.02", ""]
    DespesasOp:
    - ["3.04", ""]
    EBIT:
    - ["3.*", "Resultado Antes do Resultado Financeiro e dos Tributos"]
    ResulFinanc: 
    - ["3.06", "Resultado Financeiro"]
    - ["3.07", "Resultado Financeiro"]
    - ["3.08", "Resultado Financeiro"]
    ResulOpDescont: 
    - ["3.10", "Resultado Líquido de Operações Descontinuadas"]
    - ["3.11", "Resultado Líquido de Operações Descontinuadas"]
    - ["3.12", "Resultado Líquido de Operações Descontinuadas"]
    LucLiq: 
    - ["3.*", "Lucro/Prejuízo Consolidado do Período"]
    - ["3.*", "Lucro/Prejuízo do Período"]

    # DFC
    FCO: [["6.01", ""]]
    FCI: [["6.02", ""]]
    FCF: [["6.03", ""]]

    # DVA
    Deprec:
    - ["7.*", "Depreciação, Amortização e Exaustão"]
    JurosCapProp:
    - ["7.*", "Juros sobre o Capital Próprio"]
    Dividendos:
    - ["7.*", "Dividendos"]

  empresaA:
    # BPA
    Estoque: 
    - ["1.01.05", "Estoques"]

`
