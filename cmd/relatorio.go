// SPDX-FileCopyrightText: 2023 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/spf13/cobra"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil"
	"github.com/dude333/rapinav2/pkg/excel"
	"github.com/dude333/rapinav2/pkg/progress"
)

const (
	_customerNumFmt  = `_(* #,##0_);[RED]_(* (#,##0);_(* "-"_);_(@_)`
	_customerPercFmt = `0.0%;[RED]0.0%;_(* "-"_);_(@_)`
	_customerFracFmt = `_(0.00_);[RED]_((0.00);_(* "-"_);_(@_)`
)

type flagsRelatorio struct {
	outputDir string
	crescente bool
}

// relatorioCmd represents the relatorio command
var relatorioCmd = &cobra.Command{
	Use:     "relatorio",
	Aliases: []string{"relat", "report"},
	Short:   "imprimir relatório",
	Long:    `relatorio das informações financeiras de uma empresa`,
	Run:     menuRelatório,
}

func init() {
	relatorioCmd.Flags().StringVarP(&flags.relatorio.outputDir, "dir", "d", ".", "Diretório do relatório")
	relatorioCmd.Flags().BoolVarP(&flags.relatorio.crescente, "crescente", "c", false, "Mostrar trimestres em ordem crescente")

	rootCmd.AddCommand(relatorioCmd)
}

func menuRelatório(_ *cobra.Command, _ []string) {
	dfp, err := contabil.NovaDemonstraçãoFinanceira(db(), flags.tempDir)
	if err != nil {
		progress.Fatal(err)
	}

	empresas, err := dfp.Empresas()
	if err != nil {
		progress.Fatal(err)
	}

	for {
		empresa, ok := escolherEmpresa(empresas)
		if !ok {
			progress.Warning("Até logo!")
			os.Exit(0)
		}

		criarRelatório(empresa, dfp)
	}
}

func criarRelatório(empresa rapina.Empresa, dfp *contabil.DemonstraçãoFinanceira) {
	filename, err := prepareFilename(flags.relatorio.outputDir, empresa.Nome)
	if err != nil {
		progress.Fatal(err)
	}

	x := excel.New()
	defer func() {
		if err := x.Close(); err != nil {
			progress.Error(err)
		}
	}()

	// DADOS CONSOLIDADOS
	progress.Running("Relatório de dados consolidados")
	itr, err := dfp.RelatórioTrimestal(empresa.CNPJ, true)
	if err != nil {
		progress.Fatal(err)
	}
	if len(itr) > 0 {
		progress.Debug("Dados consolidados: %d registros", len(itr))
		itrUnificado := rapina.UnificarContasSimilares(itr)

		if err = x.NewSheet("consolidado"); err != nil {
			progress.Fatal(err)
		}
		excelReport(x, itrUnificado, !flags.relatorio.crescente)

		if err = x.NewSheet("resumo - consolidado"); err != nil {
			progress.Fatal(err)
		}
		excelSummaryReport(x, itrUnificado, false, !flags.relatorio.crescente)

		if err = x.NewSheet("resumo - consolidado vert"); err != nil {
			progress.Fatal(err)
		}
		excelSummaryReport(x, itrUnificado, true, !flags.relatorio.crescente)
	}
	progress.RunOK()

	// DADOS INDIVIDUAIS
	if len(itr) == 0 {
		progress.Running("Relatório de dados individual")
		itr, err = dfp.RelatórioTrimestal(empresa.CNPJ, false)
		if err != nil {
			progress.Fatal(err)
		}
		if len(itr) > 0 {
			progress.Debug("Dados individuais: %d registros", len(itr))
			itrUnificado := rapina.UnificarContasSimilares(itr)

			if err = x.NewSheet("individual"); err != nil {
				progress.Fatal(err)
			}
			excelReport(x, itrUnificado, !flags.relatorio.crescente)

			if err = x.NewSheet("resumo - individual"); err != nil {
				progress.Fatal(err)
			}
			excelSummaryReport(x, itrUnificado, false, !flags.relatorio.crescente)

			if err = x.NewSheet("resumo - individual vert"); err != nil {
				progress.Fatal(err)
			}
			excelSummaryReport(x, itrUnificado, true, !flags.relatorio.crescente)
		}
		progress.RunOK()
	}

	// Salva planilha
	if err := x.SaveAs(filename); err != nil {
		progress.Fatal(err)
		os.Exit(1)
	}

	status := fmt.Sprintf("Relatório salvo como: %s", filename)
	line := strings.Repeat("-", min(len(status), 80))
	progress.Status(line)
	progress.Status(status)
	progress.Status(line + "\n\n")
}

func excelReport(x *excel.Excel, itr []rapina.InformeTrimestral, decrescente bool) {
	if err := x.SetZoom(90.0); err != nil {
		progress.Fatal(err)
	}

	normalFont, _ := x.SetFont(10.0, false, false)
	titleFont, _ := x.SetFont(10.0, true, false)
	numberNormal, _ := x.SetNumber(10.0, false, _customerNumFmt)
	numberBold, _ := x.SetNumber(10.0, true, _customerNumFmt)

	// ===== Relatório - início =====

	seq := []int{0, 1, 2, 3}
	anos := rapina.RangeAnos(itr)

	if decrescente {
		reverse(seq)
		reverse(anos)
	}

	const initCol = 3

	cabeçalho := func(row, col int) {
		x.PrintCell(row, 1, titleFont, "Código")
		x.PrintCell(row, 2, titleFont, "Descrição")
		for _, ano := range anos {
			x.PrintCell(row, col+seq[0], titleFont, fmt.Sprintf("1T%d", ano))
			x.PrintCell(row, col+seq[1], titleFont, fmt.Sprintf("2T%d", ano))
			x.PrintCell(row, col+seq[2], titleFont, fmt.Sprintf("3T%d", ano))
			x.PrintCell(row, col+seq[3], titleFont, fmt.Sprintf("4T%d", ano))
			col += 4
		}
	}

	row := 1
	col := initCol
	cabeçalho(row, col)

	row++
	for i, informe := range itr {
		if rapina.Zerado(informe.Valores) {
			continue
		}
		if i > 1 && (itr[i-1].Codigo[0] != itr[i].Codigo[0]) {
			x.PrintCell(row, 1, normalFont, "______________")
			row++
		}
		font := normalFont
		number := numberNormal
		if strings.Count(informe.Codigo, ".") <= 1 {
			font = titleFont
			number = numberBold
		}
		spc := space(informe.Codigo)
		x.PrintCell(row, 1, font, spc+informe.Codigo)
		x.PrintCell(row, 2, font, spc+informe.Descr)
		col = initCol
		for _, ano := range anos {
			for _, valor := range informe.Valores {
				if valor.Ano != ano {
					continue
				}
				x.PrintCell(row, col+seq[0], number, valor.T1)
				x.PrintCell(row, col+seq[1], number, valor.T2)
				x.PrintCell(row, col+seq[2], number, valor.T3)
				x.PrintCell(row, col+seq[3], number, valor.T4)
			}
			col += 4
		}
		row++
	}

	// Auto-resize columns
	widths := make([]float64, col+3)
	widths[0], widths[1] = colWidths(itr)
	for i := 2; i < col+3; i++ {
		widths[i] = 12
	}
	x.SetColWidth(widths)

	// Freeze panes
	_ = x.FreezePane("C2")

	// Trim empty columns
	hasData := rapina.TrimestresComDados(itr)
	if decrescente {
		reverseb(hasData)
	}
	for i := len(hasData) - 1; i >= 0; i-- {
		if hasData[i] {
			break
		}
		_ = x.RemoveCol(initCol + i)
	}
	for i := 0; i < len(hasData); i++ {
		if hasData[i] {
			break
		}
		_ = x.RemoveCol(initCol)

	}
} // excelReport =====

func colWidths(itr []rapina.InformeTrimestral) (float64, float64) {
	var codWidth, descrWidth float64
	for i := range itr {
		spc := space(itr[i].Codigo)
		codWidth = math.Max(codWidth, excel.StringWidth(spc+itr[i].Codigo))
		descrWidth = math.Max(descrWidth, excel.StringWidth(spc+itr[i].Descr))
	}

	return codWidth, descrWidth
}

func space(str string) string {
	n := strings.Count(str, ".")
	if n > 0 && len(str) > 0 && str[0] != byte('1') && str[0] != byte('2') {
		n--
	}
	return strings.Repeat("  ", n)
}

func reverse(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func reverseb(s []bool) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type accountType int

const (
	UNDEF accountType = iota

	// Balanço Patrimonial
	Caixa
	AplicFinanceiras
	Estoque
	Equity
	ContasARecebCirc
	ContasARecebNCirc
	AtivoCirc
	AtivoNCirc
	AtivoTotal
	PassivoCirc
	PassivoNCirc
	PassivoTotal
	DividaCirc
	DividaNCirc
	DividendosJCP
	DividendosMin

	// DRE
	Vendas
	CustoVendas
	DespesasOp
	EBIT
	ResulFinanc
	ResulOpDescont
	LucLiq

	// DFC
	FCO
	FCI
	FCF

	// DVA
	Deprec
	JurosCapProp
	Dividendos
)

// conta code, description and bookkeeping code
type conta struct {
	cod   string
	descr string
}

var _tabelaContas = map[accountType][]conta{
	// BPA
	AtivoTotal:        {{"1", "Ativo Total"}},
	AtivoCirc:         {{"1.01", "Ativo Circulante"}},
	AtivoNCirc:        {{"1.02", "Ativo Não Circulante"}},
	Caixa:             {{"1.01.01", "Caixa e Equivalentes de Caixa"}},
	AplicFinanceiras:  {{"1.01.02", "Aplicações Financeiras"}},
	Estoque:           {{"1.01.04", "Estoques"}},
	ContasARecebCirc:  {{"1.01.03", "Contas a Receber"}},
	ContasARecebNCirc: {{"1.02.01.03", "Contas a Receber"}, {"1.02.01.04", "Contas a Receber"}},

	// BPP
	PassivoTotal:  {{"2", "Passivo Total"}},
	PassivoCirc:   {{"2.01", "Passivo Circulante"}},
	PassivoNCirc:  {{"2.02", "Passivo Não Circulante"}},
	Equity:        {{"2.*", "Patrimônio Líquido Consolidado"}, {"2.*", "Patrimônio Líquido"}},
	DividaCirc:    {{"2.01.04", "Empréstimos e Financiamentos"}},
	DividaNCirc:   {{"2.02.01", "Empréstimos e Financiamentos"}},
	DividendosJCP: {{"2.01.05.02.01", "Dividendos e JCP a Pagar"}},
	DividendosMin: {{"2.01.05.02.02", "Dividendo Mínimo Obrigatório a Pagar"}},

	// DRE
	Vendas:      {{"3.01", ""}},
	CustoVendas: {{"3.02", ""}},
	DespesasOp:  {{"3.04", ""}},
	EBIT:        {{"3.*", "Resultado Antes do Resultado Financeiro e dos Tributos"}},
	ResulFinanc: {
		{"3.06", "Resultado Financeiro"},
		{"3.07", "Resultado Financeiro"},
		{"3.08", "Resultado Financeiro"},
	},
	ResulOpDescont: {
		{"3.10", "Resultado Líquido de Operações Descontinuadas"},
		{"3.11", "Resultado Líquido de Operações Descontinuadas"},
		{"3.12", "Resultado Líquido de Operações Descontinuadas"},
	},
	LucLiq: {
		{"3.*", "Lucro/Prejuízo Consolidado do Período"},
		{"3.*", "Lucro/Prejuízo do Período"},
	},

	// DFC
	FCO: {{"6.01", ""}},
	FCI: {{"6.02", ""}},
	FCF: {{"6.03", ""}},

	// DVA
	Deprec:       {{"7.*", "Depreciação, Amortização e Exaustão"}},
	JurosCapProp: {{"7.*", "Juros sobre o Capital Próprio"}},
	Dividendos:   {{"7.*", "Dividendos"}},
}

func acctCode(cod, descr string) accountType {
	for key, v := range _tabelaContas {
		for _, acc := range v {
			l := len(acc.cod)
			if cod == acc.cod || (l > 1 && acc.cod[l-1] == '*' && strings.HasPrefix(cod, acc.cod[:l-1])) {
				if acc.descr == "" {
					return key
				}
				if rapina.NormalizeString(descr) == rapina.NormalizeString(acc.descr) {
					return key
				}
			}
		}
	}
	return UNDEF
}

func ifElse[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func excelSummaryReport(x *excel.Excel, itr []rapina.InformeTrimestral, vert, decrescente bool) {
	if err := x.SetZoom(90.0); err != nil {
		progress.Fatal(err)
	}

	number, _ := x.SetNumber(10.0, false, _customerNumFmt)
	percent, _ := x.SetNumber(10.0, false, _customerPercFmt)
	frac, _ := x.SetNumber(10.0, false, _customerFracFmt)
	titleFont, _ := x.SetFont(10.0, true, vert)

	seq := []int{0, 1, 2, 3}
	anos := rapina.RangeAnos(itr)
	if decrescente {
		reverse(seq)
		reverse(anos)
	}

	cabeçalho := func(row, col int) {
		x.PrintCell(row, col, titleFont, ifElse(vert, "Trimestre", "Descrição"))
		row += ifElse(vert, 1, 0)
		col += ifElse(vert, 0, 1)
		for _, ano := range anos {
			if !vert {
				x.PrintCell(row, col+seq[0], titleFont, fmt.Sprintf("1T%d", ano))
				x.PrintCell(row, col+seq[1], titleFont, fmt.Sprintf("2T%d", ano))
				x.PrintCell(row, col+seq[2], titleFont, fmt.Sprintf("3T%d", ano))
				x.PrintCell(row, col+seq[3], titleFont, fmt.Sprintf("4T%d", ano))
				col += 4
			} else {
				x.PrintCell(row+seq[0], col, titleFont, fmt.Sprintf("1T%d", ano))
				x.PrintCell(row+seq[1], col, titleFont, fmt.Sprintf("2T%d", ano))
				x.PrintCell(row+seq[2], col, titleFont, fmt.Sprintf("3T%d", ano))
				x.PrintCell(row+seq[3], col, titleFont, fmt.Sprintf("4T%d", ano))
				row += 4
			}
		}
	}

	c := map[accountType][]rapina.ValoresTrimestrais{}
	for _, informe := range itr {
		c[acctCode(informe.Codigo, informe.Descr)] = informe.Valores
	}

	const row2 = 2
	const colB = 2
	sumRows := make([]float64, len(anos)*4)
	sumCols := make([]float64, len(anos)*4)
	imprimirTrimestres := func(row, col int, estilo int, valores []rapina.ValoresTrimestrais) {
		for _, ano := range anos {
			for _, valor := range valores {
				if valor.Ano != ano {
					continue
				}
				if !vert {
					x.PrintCell(row, col+seq[0], estilo, valor.T1)
					x.PrintCell(row, col+seq[1], estilo, valor.T2)
					x.PrintCell(row, col+seq[2], estilo, valor.T3)
					x.PrintCell(row, col+seq[3], estilo, valor.T4)

					sumCols[col+seq[0]-colB] += valor.T1
					sumCols[col+seq[1]-colB] += valor.T2
					sumCols[col+seq[2]-colB] += valor.T3
					sumCols[col+seq[3]-colB] += valor.T4
				} else {
					x.PrintCell(row+seq[0], col, estilo, valor.T1)
					x.PrintCell(row+seq[1], col, estilo, valor.T2)
					x.PrintCell(row+seq[2], col, estilo, valor.T3)
					x.PrintCell(row+seq[3], col, estilo, valor.T4)

					sumRows[row+seq[0]-row2] += valor.T1
					sumRows[row+seq[1]-row2] += valor.T2
					sumRows[row+seq[2]-row2] += valor.T3
					sumRows[row+seq[3]-row2] += valor.T4
				}
			}
			if !vert {
				col += 4
			} else {
				row += 4
			}
		}
	}

	// ------------------[ Relatório ]------------------
	cabeçalho(1, 1)
	row := ifElse(vert, 1, 2)
	col := ifElse(vert, 2, 1)
	p := func(descr string, estilo int, valores []rapina.ValoresTrimestrais) {
		if !vert {
			x.PrintCell(row, 1, titleFont, descr)
			imprimirTrimestres(row, 2, estilo, valores)
			row++
		} else {
			x.PrintCell(1, col, titleFont, descr)
			imprimirTrimestres(2, col, estilo, valores)
			col++
		}
	}
	p("Patrimônio Líquido", number, c[Equity])
	row += ifElse(vert, 0, 1)
	p("Receita Líquida", number, c[Vendas])
	ebitda := rapina.SubVTs(c[EBIT], c[Deprec])
	p("EBITDA", number, ebitda)
	p("EBIT", number, c[EBIT])
	p("Resultado Financeiro", number, c[ResulFinanc])
	p("Operações Descont.", number, c[ResulOpDescont])
	p("Lucro Líquido", number, c[LucLiq])
	row += ifElse(vert, 0, 1)
	p("Marg. EBITDA", percent, rapina.DivVTs(ebitda, c[Vendas]))
	p("Marg. EBIT", percent, rapina.DivVTs(c[EBIT], c[Vendas]))
	p("Marg. Líq.", percent, rapina.DivVTs(c[LucLiq], c[Vendas]))
	p("ROA", percent, rapina.DivVTs(c[LucLiq], c[AtivoTotal]))
	p("ROE", percent, rapina.DivVTs(c[LucLiq], c[Equity]))
	row += ifElse(vert, 0, 1)
	caixa := rapina.AddVTs(c[Caixa], c[AplicFinanceiras])
	dividaBruta := rapina.AddVTs(c[DividaCirc], c[DividaNCirc])
	dividaLiquida := rapina.SubVTs(dividaBruta, caixa)
	p("Caixa", number, caixa)
	p("Dívida Bruta", number, dividaBruta)
	p("Dívida Líq.", number, dividaLiquida)
	p("Dív. Bru./PL", frac, rapina.DivVTs(dividaBruta, c[Equity]))
	p("Dív.Líq./ EBITDA", frac, rapina.DivVTs(dividaLiquida, ebitda))
	row += ifElse(vert, 0, 1)
	p("FCO", number, c[FCO])
	p("FCI", number, c[FCI])
	p("FCF", number, c[FCF])
	p("FCT", number, rapina.AddVTs(rapina.AddVTs(c[FCO], c[FCI]), c[FCF]))
	p("FCL (FCO+FCI)", number, rapina.AddVTs(c[FCO], c[FCI]))
	row += ifElse(vert, 0, 1)
	proventos := rapina.AddVTs(c[Dividendos], c[JurosCapProp])
	p("Proventos", number, proventos)
	p("Payout", frac, rapina.DivVTs(proventos, c[LucLiq]))
	// -------------------------------------------------

	// Auto-resize columns
	cols := ifElse(vert, col, colB+len(anos)*4)
	widths := make([]float64, cols)
	widths[0] = ifElse(vert, 8.5, 18.0)
	for i := 1; i < cols; i++ {
		widths[i] = 12.0
	}
	x.SetColWidth(widths)

	// Freeze panes
	_ = x.FreezePane("B2")

	if !vert {
		// Trim empty columns
		for i := len(sumCols) - 1; i >= 0; i-- {
			if sumCols[i] != 0.0 {
				break
			}
			_ = x.RemoveCol(colB + i)
		}
		for i := 0; i < len(sumCols); i++ {
			if sumCols[i] != 0.0 {
				break
			}
			_ = x.RemoveCol(colB)
		}
	} else {
		// Trim empty rows
		for i := len(sumRows) - 1; i >= 0; i-- {
			if sumRows[i] != 0.0 {
				break
			}
			_ = x.RemoveRow(row2 + i)
		}
		for i := 0; i < len(sumRows); i++ {
			if sumRows[i] != 0.0 {
				break
			}
			_ = x.RemoveRow(row2)
		}
	}
}
