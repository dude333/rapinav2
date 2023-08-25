package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

type flagsRelatorio struct {
	ano int
}

// relatorioCmd represents the relatorio command
var relatorioCmd = &cobra.Command{
	Use:     "relatorio",
	Aliases: []string{"report"},
	Short:   "imprimir relatório",
	Long:    `relatorio das informações financeiras de uma empresa`,
	Run:     imprimirRelatório,
}

func init() {
	relatorioCmd.Flags().IntVarP(&flags.relatorio.ano, "ano", "a", 0, "Ano do relatório")

	rootCmd.AddCommand(relatorioCmd)
}

func imprimirRelatório(cmd *cobra.Command, args []string) {
	progress.Status("{%d}", flags.relatorio.ano)

	if len(args) < 1 && len(args[0]) != len("60.872.504/0001-23") {
		progress.ErrorMsg("CNPJ inválido")
		os.Exit(1)
	}

	dfp, err := contabil.NovaDemonstraçãoFinanceira(db(), flags.tempDir)
	if err != nil {
		progress.Error(err)
		os.Exit(1)
	}

	cnpj := args[0]
	// df, err := dfp.Relatório(cnpj, flags.relatorio.ano)
	itr, err := dfp.RelatórioTrimestal(cnpj)

	if err != nil {
		progress.Error(err)
		os.Exit(1)
	}

	excel(itr)

	// fmt.Println(r)

	// Print Conta in tabular format
	// fmt.Println("Contas:")
	// w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug|tabwriter.TabIndent)
	// fmt.Fprintf(w, "Código\tDescr\tConsolidado\tGrupo\tDataIniExerc\tDataFimExerc\tMeses\tOrdemExerc\tTotal\n")
	// for _, conta := range df.Contas {
	// 	fmt.Fprintf(w, "%s\t%s\t%t\t%s\t%s\t%s\t%d\t%s\t%.2f\n",
	// 		conta.Código, conta.Descr, conta.Consolidado, conta.Grupo, conta.DataIniExerc, conta.DataFimExerc,
	// 		conta.Meses, conta.OrdemExerc, conta.Total.Valor)
	// }
	// w.Flush()
}

func excel(itr []rapina.InformeTrimestral) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Create a new sheet.
	sheetName := "Informe Trimestral"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		fmt.Println(err)
		return
	}
	f.SetActiveSheet(index)

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// customerNumFmt := `_-* #.##0_-;[RED]* (#.##0)_-;_-* "-"_-;_-@_-`
	customerNumFmt := `_(* #,##0_);[RED]_(* (#,##0);_(* "-"_);_(@_)`
	numberStyle, err := f.NewStyle(&excelize.Style{
		CustomNumFmt: &customerNumFmt,
	})
	if err != nil {
		log.Fatal(err)
	}

	minAno, maxAno := minmax(itr)

	_ = f.SetCellValue(sheetName, "A1", "Código")
	_ = f.SetCellValue(sheetName, "B1", "Descrição")

	col, _ := excelize.ColumnNameToNumber("C")
	for ano := minAno; ano <= maxAno; ano++ {
		_ = f.SetCellValue(sheetName, cell(col, 1), fmt.Sprintf("1T%d", ano))
		_ = f.SetCellValue(sheetName, cell(col+1, 1), fmt.Sprintf("2T%d", ano))
		_ = f.SetCellValue(sheetName, cell(col+2, 1), fmt.Sprintf("3T%d", ano))
		_ = f.SetCellValue(sheetName, cell(col+3, 1), fmt.Sprintf("4T%d", ano))
		col += 4
	}

	codAtual := byte('1')
	row := 2
	for _, informe := range itr {
		if zerado(informe.Valores) {
			continue
		}
		if informe.Codigo[0] != codAtual {
			codAtual = informe.Codigo[0]
			row++
		}
		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), informe.Codigo)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), informe.Descr)
		col, _ = excelize.ColumnNameToNumber("C")
		for ano := minAno; ano <= maxAno; ano++ {
			for _, valor := range informe.Valores {
				if valor.Ano != ano {
					continue
				}
				_ = f.SetCellValue(sheetName, cell(col, row), valor.T1)
				_ = f.SetCellValue(sheetName, cell(col+1, row), valor.T2)
				_ = f.SetCellValue(sheetName, cell(col+2, row), valor.T3)
				if strings.HasPrefix(informe.Codigo, "1") || strings.HasPrefix(informe.Codigo, "2") {
					_ = f.SetCellValue(sheetName, cell(col+3, row), valor.Anual)
				} else {
					_ = f.SetCellValue(sheetName, cell(col+3, row), valor.T4)
				}
				_ = f.SetCellStyle(sheetName, cell(col, row), cell(col+3, row), numberStyle)
			}
			col += 4
		}
		row++
	}

	// Auto-resize columns
	codWidth, descrWidth := colWidths(itr)
	_ = f.SetColWidth(sheetName, "A", "A", codWidth)
	_ = f.SetColWidth(sheetName, "B", "B", descrWidth)
	_ = f.SetColWidth(sheetName, num2name(3), num2name(col+3), 12)

	_ = f.SetCellStyle(sheetName, num2name(1)+"1", num2name(col)+"1", titleStyle)

	// Freeze panes
	_ = f.SetPanes(sheetName, &excelize.Panes{
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

	if err := f.SaveAs("InformeTrimestral.xlsx"); err != nil {
		log.Fatal(err)
	}
}

func zerado(valores []rapina.ValoresTrimestrais) bool {
	for _, v := range valores {
		if v.T1 != 0 || v.T2 != 0 || v.T3 != 0 || v.T4 != 0 || v.Anual != 0 {
			return false
		}
	}
	return true
}

func num2name(col int) string {
	n, _ := excelize.ColumnNumberToName(col)
	return n
}

func minmax(itr []rapina.InformeTrimestral) (int, int) {
	minAno := 99999
	maxAno := 0
	for i := range itr {
		for _, valores := range itr[i].Valores {
			if valores.Ano < minAno {
				minAno = valores.Ano
			}
			if valores.Ano > maxAno {
				maxAno = valores.Ano
			}
		}
	}

	return minAno, maxAno
}

func colWidths(itr []rapina.InformeTrimestral) (float64, float64) {
	var codWidth, descrWidth float64
	for i := range itr {
		if codWidth < stringWidth(itr[i].Codigo) {
			fmt.Printf("--- [%.2f] %s\n", stringWidth(itr[i].Codigo), itr[i].Codigo)
		}
		if descrWidth < stringWidth(itr[i].Descr) {
			fmt.Printf("--- [%.2f] %s\n", stringWidth(itr[i].Descr), itr[i].Descr)
		}
		codWidth = math.Max(codWidth, stringWidth(itr[i].Codigo))
		descrWidth = math.Max(descrWidth, stringWidth(itr[i].Descr))
	}

	return codWidth, descrWidth
}

func cell(col, row int) string {
	return fmt.Sprintf("%s%d", num2name(col), row)
}

func stringWidth(str string) float64 {
	var width float64 = 0.0
	for _, ch := range str {
		width += charWidth(ch)
	}
	return width
}

func charWidth(ch rune) float64 {
	keys := map[rune]float64{
		'0': 6.0,
		'1': 6.0,
		'2': 6.0,
		'3': 6.0,
		'4': 6.0,
		'5': 6.0,
		'6': 6.0,
		'7': 6.0,
		'8': 6.0,
		'9': 6.0,
		'A': 6.0,
		'B': 6.0,
		'C': 6.0,
		'D': 7.0,
		'E': 5.0,
		'F': 5.0,
		'G': 7.0,
		'H': 7.0,
		'I': 3.0,
		'J': 4.0,
		'K': 6.0,
		'L': 5.0,
		'M': 9.0,
		'N': 7.0,
		'O': 7.0,
		'P': 6.0,
		'Q': 7.0,
		'R': 6.0,
		'S': 5.0,
		'T': 5.0,
		'U': 7.0,
		'V': 6.0,
		'W': 10.0,
		'X': 6.0,
		'Y': 5.0,
		'Z': 5.0,
		'a': 5.0,
		'b': 6.0,
		'c': 5.0,
		'd': 6.0,
		'e': 5.0,
		'f': 3.0,
		'g': 5.0,
		'h': 6.0,
		'i': 3.0,
		'j': 3.0,
		'k': 5.0,
		'l': 3.0,
		'm': 9.0,
		'n': 6.0,
		'o': 6.0,
		'p': 6.0,
		'q': 6.0,
		'r': 4.0,
		's': 4.0,
		't': 4.0,
		'u': 6.0,
		'v': 5.0,
		'w': 8.0,
		'x': 5.0,
		'y': 5.0,
		'z': 4.0,
		'À': 6.0,
		'Á': 6.0,
		'Â': 6.0,
		'Ã': 6.0,
		'Ç': 6.0,
		'É': 5.0,
		'Ê': 5.0,
		'Í': 3.0,
		'Ó': 7.0,
		'Ô': 7.0,
		'Õ': 7.0,
		'Ú': 7.0,
		'à': 5.0,
		'á': 5.0,
		'â': 5.0,
		'ã': 5.0,
		'ç': 5.0,
		'é': 5.0,
		'ê': 5.0,
		'í': 3.0,
		'ó': 6.0,
		'ô': 6.0,
		'õ': 6.0,
		'ú': 6.0,
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
		fmt.Printf("---> %c\n", ch)
		return 1.4
	}
	return width / 5.2
}
