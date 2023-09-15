// SPDX-FileCopyrightText: 2023 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil"
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

	x := &Excel{
		file:       excelize.NewFile(),
		sheetName:  "Informe Trimestral",
		sheetIndex: 0,
	}
	defer func() {
		if err := x.file.Close(); err != nil {
			progress.Error(err)
		}
	}()

	// DADOS CONSOLIDADOS
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

	// DADOS INDIVIDUAIS
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

	// Salva planilha
	if err := x.file.SaveAs(filename); err != nil {
		progress.Fatal(err)
		os.Exit(1)
	}

	status := fmt.Sprintf("Relatório salvo como: %s", filename)
	line := strings.Repeat("-", min(len(status), 80))
	progress.Status(line)
	progress.Status(status)
	progress.Status(line + "\n\n")
}

type Excel struct {
	file       *excelize.File
	sheetName  string
	sheetIndex int
}

func (x *Excel) NewSheet(sheetName string) error {
	n := len(x.file.GetSheetList())
	if n == 1 && x.sheetIndex == 0 {
		err := x.file.SetSheetName(x.file.GetSheetList()[x.sheetIndex], sheetName)
		if err != nil {
			return err
		}
	} else {
		_, err := x.file.NewSheet(sheetName)
		if err != nil {
			return err
		}
	}
	x.sheetIndex++
	x.sheetName = sheetName
	return nil
}

func (x *Excel) printCell(row, col int, style int, value interface{}) {
	_ = x.file.SetCellValue(x.sheetName, cell(row, col), value)
	_ = x.file.SetCellStyle(x.sheetName, cell(row, col), cell(row, col), style)
}

func cell(row, col int) string {
	return fmt.Sprintf("%s%d", num2name(col), row)
}

func excelReport(x *Excel, itr []rapina.InformeTrimestral, decrescente bool) {
	ZoomScale := 90.0
	if err := x.file.SetSheetView(x.sheetName, 0, &excelize.ViewOptions{ZoomScale: &ZoomScale}); err != nil {
		log.Fatal(err)
	}

	fontStyle, err := x.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 10.0,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	titleStyle, err := x.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	customerNumFmt := `_(* #,##0_);[RED]_(* (#,##0);_(* "-"_);_(@_)`
	numberStyle, err := x.file.NewStyle(&excelize.Style{
		CustomNumFmt: &customerNumFmt,
		Font: &excelize.Font{
			Size: 10.0,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	boldNumberStyle, err := x.file.NewStyle(&excelize.Style{
		CustomNumFmt: &customerNumFmt,
		Font: &excelize.Font{
			Bold: true,
			Size: 10.0,
		},
	})
	if err != nil {
		log.Fatal(err)
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
		x.printCell(row, 1, titleStyle, "Código")
		x.printCell(row, 2, titleStyle, "Descrição")
		for _, ano := range anos {
			x.printCell(row, col+seq[0], titleStyle, fmt.Sprintf("1T%d", ano))
			x.printCell(row, col+seq[1], titleStyle, fmt.Sprintf("2T%d", ano))
			x.printCell(row, col+seq[2], titleStyle, fmt.Sprintf("3T%d", ano))
			x.printCell(row, col+seq[3], titleStyle, fmt.Sprintf("4T%d", ano))
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
			x.printCell(row, 1, fontStyle, "______________")
			row++
		}
		spc := space(informe.Codigo)
		x.printCell(row, 1, fontStyle, spc+informe.Codigo)
		x.printCell(row, 2, fontStyle, spc+informe.Descr)
		if strings.Count(informe.Codigo, ".") <= 1 {
			_ = x.file.SetCellStyle(x.sheetName, cell(row, 1), cell(row, 2), titleStyle)
		}
		col = initCol
		for _, ano := range anos {
			for _, valor := range informe.Valores {
				if valor.Ano != ano {
					continue
				}
				x.printCell(row, col+seq[0], numberStyle, valor.T1)
				x.printCell(row, col+seq[1], numberStyle, valor.T2)
				x.printCell(row, col+seq[2], numberStyle, valor.T3)
				x.printCell(row, col+seq[3], numberStyle, valor.T4)

				if strings.Count(informe.Codigo, ".") <= 1 {
					_ = x.file.SetCellStyle(x.sheetName, cell(row, col), cell(row, col+3), boldNumberStyle)
				}
			}
			col += 4
		}
		row++
	}

	// Auto-resize columns
	codWidth, descrWidth := colWidths(itr)
	_ = x.file.SetColWidth(x.sheetName, "A", "A", codWidth)
	_ = x.file.SetColWidth(x.sheetName, "B", "B", descrWidth)
	_ = x.file.SetColWidth(x.sheetName, num2name(3), num2name(col+3), 12)

	// Freeze panes
	_ = x.file.SetPanes(x.sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      2,
		YSplit:      1,
		TopLeftCell: "C2",
		ActivePane:  "bottomRight",
		Panes: []excelize.PaneOptions{
			{SQRef: "C2", ActiveCell: "C2", Pane: "bottomRight"},
		},
	})

	// Delete empty columns
	hasData := rapina.TrimestresComDados(itr)
	if decrescente {
		reverseb(hasData)
	}
	for i := len(hasData) - 1; i >= 0; i-- {
		if !hasData[i] {
			_ = x.file.RemoveCol(x.sheetName, num2name(initCol+i))
		}
	}
} // excelReport =====

func num2name(col int) string {
	n, _ := excelize.ColumnNumberToName(col)
	return n
}

func colWidths(itr []rapina.InformeTrimestral) (float64, float64) {
	var codWidth, descrWidth float64
	for i := range itr {
		spc := space(itr[i].Codigo)
		codWidth = math.Max(codWidth, stringWidth(spc+itr[i].Codigo))
		descrWidth = math.Max(descrWidth, stringWidth(spc+itr[i].Descr))
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

func stringWidth(str string) float64 {
	var width float64 = 0.0
	for _, ch := range str {
		width += charWidth(ch)
	}
	return width
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// charWidth foi criado com este script em Python:
// from PIL import ImageFont
// font = ImageFont.truetype("calibri.ttf", 11)
// for char in "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzÀÁÂÃÇÉÊÍÓÔÕÚàáâãçéêíóôõú.-_ /(),":
//
//	width = font.getlength(char)
//	print(f"\'{char}\': {width},")
func charWidth(ch rune) float64 {
	keys := map[rune]float64{
		'0': 5.0,
		'1': 5.0,
		'2': 5.0,
		'3': 5.0,
		'4': 5.0,
		'5': 5.0,
		'6': 5.0,
		'7': 5.0,
		'8': 5.0,
		'9': 5.0,
		'A': 6.0,
		'B': 5.0,
		'C': 5.0,
		'D': 6.0,
		'E': 5.0,
		'F': 5.0,
		'G': 6.0,
		'H': 6.0,
		'I': 3.0,
		'J': 3.0,
		'K': 5.0,
		'L': 4.0,
		'M': 9.0,
		'N': 6.0,
		'O': 7.0,
		'P': 5.0,
		'Q': 7.0,
		'R': 5.0,
		'S': 5.0,
		'T': 5.0,
		'U': 6.0,
		'V': 6.0,
		'W': 9.0,
		'X': 5.0,
		'Y': 5.0,
		'Z': 5.0,
		'a': 5.0,
		'b': 5.0,
		'c': 4.0,
		'd': 5.0,
		'e': 5.0,
		'f': 3.0,
		'g': 5.0,
		'h': 5.0,
		'i': 2.0,
		'j': 2.0,
		'k': 5.0,
		'l': 2.0,
		'm': 8.0,
		'n': 5.0,
		'o': 5.0,
		'p': 5.0,
		'q': 5.0,
		'r': 3.0,
		's': 4.0,
		't': 3.0,
		'u': 5.0,
		'v': 5.0,
		'w': 7.0,
		'x': 4.0,
		'y': 5.0,
		'z': 4.0,
		'À': 6.0,
		'Á': 6.0,
		'Â': 6.0,
		'Ã': 6.0,
		'Ç': 5.0,
		'É': 5.0,
		'Ê': 5.0,
		'Í': 3.0,
		'Ó': 7.0,
		'Ô': 7.0,
		'Õ': 7.0,
		'Ú': 6.0,
		'à': 5.0,
		'á': 5.0,
		'â': 5.0,
		'ã': 5.0,
		'ç': 4.0,
		'é': 5.0,
		'ê': 5.0,
		'í': 2.0,
		'ó': 5.0,
		'ô': 5.0,
		'õ': 5.0,
		'ú': 5.0,
		'.': 3.0,
		'-': 3.0,
		'_': 5.0,
		' ': 2.0,
		'/': 4.0,
		'(': 3.0,
		')': 3.0,
		',': 3.0,
	}
	width, ok := keys[ch]
	if !ok {
		progress.Trace("charWidth: Caractere '%c' não encontrado na tabela", ch)
		return 1.4
	}
	return width / 5.2
}
