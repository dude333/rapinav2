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

// Títulos das abas dos relatórios (máximo 31 caracteres)
const (
	tituloRelatCompletoAnual   = "completo anual (%s)" // %s = "consolidado" ou "individual"
	tituloRelatCompletoTrim    = "completo trim. (%s)"
	tituloRelatResumoAnual     = "resumo anual (%s)"
	tituloRelatResumoTrim      = "resumo trim. (%s)"
	tituloRelatResumoAnualVert = "resumo anual vert. (%s)"
	tituloRelatResumoTrimVert  = "resumo trim. vert. (%s)"
)

const (
	_customerNumFmt  = `_(* #,##0_);[RED]_(* (#,##0);_(* "-"_);_(@_)`
	_customerPercFmt = `0.0%;[RED]0.0%;_(* "-"_);_(@_)`
	_customerFracFmt = `_(0.00_);[RED]_((0.00);_(* "-"_);_(@_)`
)

type Excel interface {
	NewSheet(sheetName string) error
	SetZoom(zoomScale float64) error
	SetColWidth(widths []float64)
	FreezePane(cell string) error
	SetFont(size float64, bold, wrap bool) (int, error)
	SetNumber(size float64, bold bool, format string) (int, error)
	PrintCell(row, col, style int, value interface{})
	RemoveRow(row int) error
	RemoveCol(col int) error
	SaveAs(name string) error
	Close() error
}

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

type reportOpts struct {
	anual       bool
	vertical    bool
	decrescente bool
}

func init() {
	relatorioCmd.Flags().StringVarP(&flags.relatorio.outputDir, "dir", "d", ".", "Diretório do relatório")
	relatorioCmd.Flags().BoolVarP(&flags.relatorio.crescente, "crescente", "c", false, "Mostrar trimestres em ordem crescente")

	rootCmd.AddCommand(relatorioCmd)
}

func menuRelatório(_ *cobra.Command, _ []string) {
	dfp, err := contabil.NewService(db(), flags.tempDir)
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

		criarRelatórios(empresa, dfp)
	}
}

// criarRelatórios gera e salva relatórios da empresa em planilhas Excel.
// Os relatórios podem ser consolidados ou, caso não existam, individuais.
func criarRelatórios(empresa rapina.Empresa, dfp *contabil.ContabilServices) {
	filename, err := prepareFilename(flags.relatorio.outputDir, empresa.Nome)
	if err != nil {
		progress.Fatal(err)
	}

	var x Excel = excel.New()
	defer func() {
		if err := x.Close(); err != nil {
			progress.Error(err)
		}
	}()

	// DADOS CONSOLIDADOS
	hasConsolidated := criarPlanilhas(x, empresa, dfp, true)

	// DADOS INDIVIDUAIS
	hasIndividual := criarPlanilhas(x, empresa, dfp, false)

	if !hasConsolidated && !hasIndividual {
		progress.Warning(fmt.Sprintf("Nenhum dado disponível para %s", empresa.Nome))
		return
	}

	// Salva planilha
	if err := x.SaveAs(filename); err != nil {
		progress.Error(fmt.Errorf("erro ao salvar relatório: %w", err))
		return
	}

	status := fmt.Sprintf("Relatório salvo como: %s", filename)
	line := strings.Repeat("-", min(len(status), 80))
	progress.Status(line)
	progress.Status(status)
	progress.Status(line + "\n\n")
}

// criarPlanilhas gera e salva relatório consolidado/individual em Excel.
func criarPlanilhas(x Excel, empresa rapina.Empresa, dfp *contabil.ContabilServices, consolidado bool) bool {
	titulo := "consolid"
	if !consolidado {
		titulo = "individ"
	}

	progress.Running("Relatório de dados " + titulo)
	itr, err := dfp.DadosTrimestrais(empresa.CNPJ, consolidado)
	if err != nil {
		progress.Fatal(err)
	}
	if len(itr) == 0 {
		progress.RunFail()
		return false
	}

	progress.Debug("Dados %s: %d registros", titulo, len(itr))
	itrUnificado := rapina.UnificarContasSimilares(itr)

	// Relatório completo, anual
	newSheet(x, fmt.Sprintf(tituloRelatCompletoAnual, titulo))
	excelReport(x, itrUnificado, reportOpts{anual: true, vertical: false, decrescente: !flags.relatorio.crescente})
	// Relatório completo, trimestral
	newSheet(x, fmt.Sprintf(tituloRelatCompletoTrim, titulo))
	excelReport(x, itrUnificado, reportOpts{anual: false, vertical: false, decrescente: !flags.relatorio.crescente})

	// Relatório resumo, anual
	newSheet(x, fmt.Sprintf(tituloRelatResumoAnual, titulo))
	excelSummaryReport(x, itrUnificado, reportOpts{anual: true, vertical: false, decrescente: !flags.relatorio.crescente})
	// Relatório resumo, trimestral
	newSheet(x, fmt.Sprintf(tituloRelatResumoTrim, titulo))
	excelSummaryReport(x, itrUnificado, reportOpts{anual: false, vertical: false, decrescente: !flags.relatorio.crescente})

	// Relatório resumo, vertical, anual
	newSheet(x, fmt.Sprintf(tituloRelatResumoAnualVert, titulo))
	excelSummaryReport(x, itrUnificado, reportOpts{anual: true, vertical: true, decrescente: !flags.relatorio.crescente})
	// Relatório resumo, vertical, trimestral
	newSheet(x, fmt.Sprintf(tituloRelatResumoTrimVert, titulo))
	excelSummaryReport(x, itrUnificado, reportOpts{anual: false, vertical: true, decrescente: !flags.relatorio.crescente})

	progress.RunOK()
	return true
}

func newSheet(x Excel, name string) {
	if err := x.NewSheet(name); err != nil {
		progress.Fatal(err)
	}
}

// excelReport cria e formata o relatório anual ou trimestral completo em
// planilha Excel com base nos dados fornecidos. O parâmetro 'decrescente'
// indica se o relatório deve ser criado em ordem crescente (false) ou
// decrescente (true) de ano.
func excelReport(x Excel, itr []rapina.InformeTrimestral, opts reportOpts) {
	if err := x.SetZoom(90.0); err != nil {
		progress.Fatal(err)
	}

	normalFont, _ := x.SetFont(10.0, false, false)
	titleFont, _ := x.SetFont(10.0, true, false)
	numberNormal, _ := x.SetNumber(10.0, false, _customerNumFmt)
	numberBold, _ := x.SetNumber(10.0, true, _customerNumFmt)

	// ===== Relatório - início =====
	anos := rapina.RangeAnos(itr, opts.decrescente)
	seq4 := func(n int) int {
		return seq(4, n, opts.decrescente)
	}
	const initCol = 3

	cabeçalho := func(row, col int) {
		x.PrintCell(row, 1, titleFont, "Código")
		x.PrintCell(row, 2, titleFont, "Descrição")
		for _, ano := range anos {
			if opts.anual {
				x.PrintCell(row, col, titleFont, fmt.Sprintf("%d", ano))
				col++
				continue
			}
			x.PrintCell(row, col+seq4(0), titleFont, fmt.Sprintf("1T%d", ano))
			x.PrintCell(row, col+seq4(1), titleFont, fmt.Sprintf("2T%d", ano))
			x.PrintCell(row, col+seq4(2), titleFont, fmt.Sprintf("3T%d", ano))
			x.PrintCell(row, col+seq4(3), titleFont, fmt.Sprintf("4T%d", ano))
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
				if opts.anual {
					total := valor.T1 + valor.T2 + valor.T3 + valor.T4
					if strings.HasPrefix(informe.Codigo, "1") || strings.HasPrefix(informe.Codigo, "2") {
						progress.Trace("informe.Valores[0]: %+v", informe.Valores[0])
						total = valor.T(4) // TODO: ajustar período para TTM para o último ano
					}
					x.PrintCell(row, col, number, total)
					continue
				}
				x.PrintCell(row, col+seq4(0), number, valor.T1)
				x.PrintCell(row, col+seq4(1), number, valor.T2)
				x.PrintCell(row, col+seq4(2), number, valor.T3)
				x.PrintCell(row, col+seq4(3), number, valor.T4)
			}
			col += ifElse(opts.anual, 1, 4)
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
	if opts.decrescente {
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

// seq retorna o enésimo valor da sequência de valores de 0 a max (exclusivo),
// ou de max (exclusivo) a 0 se reverse for true: 0 <= n < max.
func seq(max, n int, reverse bool) int {
	if n >= max || n < 0 || max < 0 {
		return 0
	}
	if reverse {
		return max - n - 1
	}
	return n
}

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

// excelSummaryReport cria e formata o relatório resumido em planilha Excel com
// base nos dados fornecidos. O parâmetro 'decrescente' indica se o relatório
// deve ser criado em ordem crescente (false) ou decrescente (true) de ano.
func excelSummaryReport(x Excel, itr []rapina.InformeTrimestral, opts reportOpts) {
	if err := x.SetZoom(90.0); err != nil {
		progress.Fatal(err)
	}

	number, _ := x.SetNumber(10.0, false, _customerNumFmt)
	percent, _ := x.SetNumber(10.0, false, _customerPercFmt)
	frac, _ := x.SetNumber(10.0, false, _customerFracFmt)
	titleFont, _ := x.SetFont(10.0, true, opts.vertical)

	anos := rapina.RangeAnos(itr, opts.decrescente)
	seq4 := func(n int) int {
		return seq(4, n, opts.decrescente)
	}

	cabeçalho := func(row, col int) {
		x.PrintCell(row, col, titleFont, ifElse(opts.vertical, "Trimestre", "Descrição"))
		row += ifElse(opts.vertical, 1, 0)
		col += ifElse(opts.vertical, 0, 1)
		for _, ano := range anos {
			if !opts.vertical {
				if opts.anual {
					x.PrintCell(row, col, titleFont, fmt.Sprintf("%d", ano))
					col++
					continue
				}
				x.PrintCell(row, col+seq4(0), titleFont, fmt.Sprintf("1T%d", ano))
				x.PrintCell(row, col+seq4(1), titleFont, fmt.Sprintf("2T%d", ano))
				x.PrintCell(row, col+seq4(2), titleFont, fmt.Sprintf("3T%d", ano))
				x.PrintCell(row, col+seq4(3), titleFont, fmt.Sprintf("4T%d", ano))
				col += 4
			} else {
				if opts.anual {
					x.PrintCell(row, col, titleFont, fmt.Sprintf("%d", ano))
					row++
					continue
				}
				x.PrintCell(row+seq4(0), col, titleFont, fmt.Sprintf("1T%d", ano))
				x.PrintCell(row+seq4(1), col, titleFont, fmt.Sprintf("2T%d", ano))
				x.PrintCell(row+seq4(2), col, titleFont, fmt.Sprintf("3T%d", ano))
				x.PrintCell(row+seq4(3), col, titleFont, fmt.Sprintf("4T%d", ano))
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
				if !opts.vertical {
					if opts.anual {
						x.PrintCell(row, col, estilo, valor.T1+valor.T2+valor.T3+valor.T4)
						sumCols[col-colB] += valor.T1 + valor.T2 + valor.T3 + valor.T4
						continue
					}
					x.PrintCell(row, col+seq4(0), estilo, valor.T1)
					x.PrintCell(row, col+seq4(1), estilo, valor.T2)
					x.PrintCell(row, col+seq4(2), estilo, valor.T3)
					x.PrintCell(row, col+seq4(3), estilo, valor.T4)

					sumCols[col+seq4(0)-colB] += valor.T1
					sumCols[col+seq4(1)-colB] += valor.T2
					sumCols[col+seq4(2)-colB] += valor.T3
					sumCols[col+seq4(3)-colB] += valor.T4
				} else {
					if opts.anual {
						x.PrintCell(row, col, estilo, valor.T1+valor.T2+valor.T3+valor.T4)
						sumRows[row-row2] += valor.T1 + valor.T2 + valor.T3 + valor.T4
						continue
					}
					x.PrintCell(row+seq4(0), col, estilo, valor.T1)
					x.PrintCell(row+seq4(1), col, estilo, valor.T2)
					x.PrintCell(row+seq4(2), col, estilo, valor.T3)
					x.PrintCell(row+seq4(3), col, estilo, valor.T4)

					sumRows[row+seq4(0)-row2] += valor.T1
					sumRows[row+seq4(1)-row2] += valor.T2
					sumRows[row+seq4(2)-row2] += valor.T3
					sumRows[row+seq4(3)-row2] += valor.T4
				}
			}
			if !opts.vertical {
				col += ifElse(opts.anual, 1, 4)
			} else {
				row += ifElse(opts.anual, 1, 4)
			}
		}
	}

	// ------------------[ Relatório ]------------------
	cabeçalho(1, 1)
	row := ifElse(opts.vertical, 1, 2)
	col := ifElse(opts.vertical, 2, 1)
	p := func(descr string, estilo int, valores []rapina.ValoresTrimestrais) {
		if !opts.vertical {
			x.PrintCell(row, 1, titleFont, descr)
			imprimirTrimestres(row, 2, estilo, valores)
			row++
		} else {
			x.PrintCell(1, col, titleFont, descr)
			imprimirTrimestres(2, col, estilo, valores)
			col++
		}
	}
	// ttm calcula o trailing 12-month para os relatórios trimestrais. Para ao anual,
	// mantém apenas o último trimestre com valor maior que zero.
	ttm := func(vts []rapina.ValoresTrimestrais) []rapina.ValoresTrimestrais {
		if !opts.anual {
			return rapina.TTM(vts)
		}
		return rapina.ManterÚltimoTrimestre(rapina.TTM(vts))
	}
	// ajusteBalanço ajusta os VTs do balanço patrimonial para o relatório anual
	// (retém apenas o último valor). Aplicar em todos os items do balanço patrimonial.
	ajusteBalanço := func(vts []rapina.ValoresTrimestrais) []rapina.ValoresTrimestrais {
		if !opts.anual {
			return vts
		}
		return rapina.ManterÚltimoTrimestre(vts)
	}
	// divVTs divide os VTs trimestrais normalmente, usa o último trimestre não
	// nulo do ttm para relatório anual.
	divVTs := func(v1, v2 []rapina.ValoresTrimestrais) []rapina.ValoresTrimestrais {
		if !opts.anual {
			return rapina.DivVTs(v1, v2)
		}
		w1 := rapina.ManterÚltimoTrimestre(rapina.TTM(v1))
		w2 := rapina.ManterÚltimoTrimestre(rapina.TTM(v2))
		return rapina.DivVTs(w1, w2)
	}
	p("Ativo Total", number, ajusteBalanço(c[AtivoTotal]))
	p("Patrimônio Líquido", number, ajusteBalanço(c[Equity]))
	row += ifElse(opts.vertical, 0, 1)
	p("Receita Líquida", number, c[Vendas])
	lucroBruto := rapina.AddVTs(c[Vendas], c[CustoVendas])
	p("Lucro Bruto", number, lucroBruto)
	p("Marg. Bruta", percent, divVTs(lucroBruto, c[Vendas]))
	ebitda := rapina.SubVTs(c[EBIT], c[Deprec])
	p("EBITDA", number, ebitda)
	p("Marg. EBITDA", percent, divVTs(ebitda, c[Vendas]))
	p("EBIT", number, c[EBIT])
	p("Marg. EBIT", percent, divVTs(c[EBIT], c[Vendas]))
	p("Resultado Financeiro", number, c[ResulFinanc])
	if !rapina.Zerado(c[ResulOpDescont]) {
		p("Operações Descont.", number, c[ResulOpDescont])
	}
	p("Lucro Líquido", number, c[LucLiq])
	p("Marg. Líq.", percent, divVTs(c[LucLiq], c[Vendas]))
	row += ifElse(opts.vertical, 0, 1)
	p("ROA", percent, divVTs(ttm(c[LucLiq]), ajusteBalanço(c[AtivoTotal])))
	p("ROE", percent, divVTs(ttm(c[LucLiq]), ajusteBalanço(c[Equity])))
	row += ifElse(opts.vertical, 0, 1)
	caixa := ajusteBalanço(rapina.AddVTs(c[Caixa], c[AplicFinanceiras]))
	dividaBruta := ajusteBalanço(rapina.AddVTs(c[DividaCirc], c[DividaNCirc]))
	dividaLiquida := ajusteBalanço(rapina.SubVTs(dividaBruta, caixa))
	p("Caixa", number, caixa)
	p("Dívida Bruta", number, dividaBruta)
	p("Dívida Líq.", number, dividaLiquida)
	p("Dív. Bru./PL", frac, divVTs(dividaBruta, ajusteBalanço(c[Equity])))
	ebitdattm := ajusteBalanço(rapina.SubVTs(ttm(c[EBIT]), ttm(c[Deprec])))
	p("Dív.Líq./ EBITDA TTM", frac, divVTs(dividaLiquida, ebitdattm))
	row += ifElse(opts.vertical, 0, 1)
	p("FCO", number, c[FCO])
	p("FCI", number, c[FCI])
	p("FCF", number, c[FCF])
	p("FCT", number, rapina.AddVTs(rapina.AddVTs(c[FCO], c[FCI]), c[FCF]))
	p("FCL (FCO+FCI)", number, rapina.AddVTs(c[FCO], c[FCI]))
	row += ifElse(opts.vertical, 0, 1)
	proventos := rapina.AddVTs(c[Dividendos], c[JurosCapProp])
	p("Proventos", number, proventos)
	p("Payout", frac, divVTs(proventos, c[LucLiq]))
	// -------------------------------------------------

	// Auto-resize columns, trim empty rows/cols, and freeze pane
	cols := ifElse(opts.vertical, col, colB+len(anos)*4)
	widths := make([]float64, cols)
	widths[0] = ifElse(opts.vertical, 8.5, 18.0)
	for i := 1; i < cols; i++ {
		widths[i] = 12.0
	}
	x.SetColWidth(widths)
	trimEmpty(x, row2, colB, sumRows, sumCols, opts.vertical)
	_ = x.FreezePane("B2")
}

// trimEmpty remove linhas e colunas vazias.
func trimEmpty(x Excel, row, col int, sumRows, sumCols []float64, vert bool) {
	if !vert {
		// Trim empty columns
		for i := len(sumCols) - 1; i >= 0; i-- {
			if sumCols[i] != 0.0 {
				break
			}
			_ = x.RemoveCol(col + i)
		}
		for i := 0; i < len(sumCols); i++ {
			if sumCols[i] != 0.0 {
				break
			}
			_ = x.RemoveCol(col)
		}
	} else {
		// Trim empty rows
		for i := len(sumRows) - 1; i >= 0; i-- {
			if sumRows[i] != 0.0 {
				break
			}
			_ = x.RemoveRow(row + i)
		}
		for i := 0; i < len(sumRows); i++ {
			if sumRows[i] != 0.0 {
				break
			}
			_ = x.RemoveRow(row)
		}
	}
}
