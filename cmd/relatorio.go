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

func menuRelatório(cmd *cobra.Command, args []string) {
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
		if err = x.NewSheet("consolidado"); err != nil {
			progress.Fatal(err)
		}
		excelReport(x, rapina.UnificarContasSimilares(itr), !flags.relatorio.crescente)
	}
	progress.RunOK()

	// DADOS INDIVIDUAIS
	progress.Running("Relatório de dados individual")
	itr, err = dfp.RelatórioTrimestal(empresa.CNPJ, false)
	if err != nil {
		progress.Fatal(err)
	}
	if len(itr) > 0 {
		progress.Debug("Dados individuais: %d registros", len(itr))
		if err = x.NewSheet("individual"); err != nil {
			progress.Fatal(err)
		}
		excelReport(x, rapina.UnificarContasSimilares(itr), !flags.relatorio.crescente)
	}
	progress.RunOK()

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

	normalFont, err := x.SetFont(10.0, false)
	if err != nil {
		progress.Fatal(err)
	}

	titleFont, err := x.SetFont(10.0, true)
	if err != nil {
		progress.Fatal(err)
	}

	customerNumFmt := `_(* #,##0_);[RED]_(* (#,##0);_(* "-"_);_(@_)`
	numberNormal, err := x.SetNumber(10.0, false, customerNumFmt)
	if err != nil {
		progress.Fatal(err)
	}

	numberBold, err := x.SetNumber(10.0, true, customerNumFmt)
	if err != nil {
		progress.Fatal(err)
	}

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

	// Delete empty columns
	hasData := rapina.TrimestresComDados(itr)
	if decrescente {
		reverseb(hasData)
	}
	for i := len(hasData) - 1; i >= 0; i-- {
		if !hasData[i] {
			_ = x.RemoveCol(initCol + i)
		}
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
