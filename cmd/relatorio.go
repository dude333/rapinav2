package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil"
	"github.com/dude333/rapinav2/pkg/progress"
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

	// if len(args) < 1 && len(args[0]) != len("60.872.504/0001-23") {
	// 	progress.ErrorMsg("CNPJ inválido")
	// 	os.Exit(1)
	// }

	dfp, err := contabil.NovaDemonstraçãoFinanceira(db(), flags.tempDir)
	if err != nil {
		progress.Error(err)
		os.Exit(1)
	}

	empresas, err := dfp.Empresas()
	if err != nil {
		log.Fatal(err)
	}
	cnpj := escolherEmpresa(empresas)
	if len(cnpj) != len("60.872.504/0001-23") {
		log.Fatal("Nenhuma empresa foi escolhida")
	}

	// cnpj := args[0]
	// df, err := dfp.Relatório(cnpj, flags.relatorio.ano)
	itr, err := dfp.RelatórioTrimestal(cnpj)
	if err != nil {
		progress.Error(err)
		os.Exit(1)
	}

	excel(rapina.UnificarContasSimilares(itr), true)
}

type noBellStdout struct{}

func (n *noBellStdout) Write(p []byte) (int, error) {
	if len(p) == 1 && p[0] == readline.CharBell {
		return 0, nil
	}
	return readline.Stdout.Write(p)
}

func (n *noBellStdout) Close() error {
	return readline.Stdout.Close()
}

var NoBellStdout = &noBellStdout{}

func escolherEmpresa(empresas []rapina.Empresa) string {
	// The Active and Selected templates set a small pepper icon next to the name colored and the heat unit for the
	// active template. The details template is show at the bottom of the select's list and displays the full info
	// for that pepper in a multi-line template.
	templates := &promptui.SelectTemplates{
		Help: `{{ "Para navegar:" | faint }} [{{ .NextKey | faint }} ` +
			`{{ .PrevKey | faint }} {{ .PageUpKey | faint }} {{ .PageDownKey | faint }}]` +
			`{{ if .Search }} {{ "Para procurar:" | faint }} [{{ .SearchKey | faint }}]{{ end }}`,
		Label:    "{{ . }}:",
		Active:   " > {{ .Nome | red }}",
		Inactive: "  {{ .Nome | blue }}",
		Selected: " > {{ .Nome | red | cyan }}",
		Details: `
--------------------------------------
| {{ "Name:" | bold }}	{{ .Nome }}
| {{ "CNPJ:" | faint }}	{{ .CNPJ }}
------------------------------------------`,
	}

	// A searcher function is implemented which enabled the search mode for the select. The function follows
	// the required searcher signature and finds any pepper whose name contains the searched string.
	searcher := func(input string, index int) bool {
		empresa := empresas[index]
		nome := rapina.NormalizeString(empresa.Nome)
		ninput := rapina.NormalizeString(input)

		return strings.Contains(nome, ninput)
	}

	prompt := promptui.Select{
		Label:     "Empresas",
		Items:     empresas,
		Templates: templates,
		Size:      15,
		Searcher:  searcher,
		Stdout:    NoBellStdout,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return ""
	}

	return empresas[i].CNPJ
}

func excel(itr []rapina.InformeTrimestral, decrescente bool) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Create a new sheet.
	sheet := "Informe Trimestral"
	err := f.SetSheetName(f.GetSheetList()[0], sheet)
	if err != nil {
		log.Fatal(err)
	}

	x := Excel{
		f:         f,
		sheetName: sheet,
	}

	ZoomScale := 90.0
	if err := f.SetSheetView(sheet, 0, &excelize.ViewOptions{ZoomScale: &ZoomScale}); err != nil {
		log.Fatal(err)
	}

	fontStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 10.0,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

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
		Font: &excelize.Font{
			Size: 10.0,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	boldNumberStyle, err := f.NewStyle(&excelize.Style{
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

	x.printCell(1, 1, titleStyle, "Código")
	x.printCell(1, 2, titleStyle, "Descrição")

	seq := []int{0, 1, 2, 3}
	anos := rapina.RangeAnos(itr)

	if decrescente {
		reverse(seq)
		reverse(anos)
	}

	const initCol = 3
	col := initCol
	for _, ano := range anos {
		x.printCell(1, col+seq[0], titleStyle, fmt.Sprintf("1T%d", ano))
		x.printCell(1, col+seq[1], titleStyle, fmt.Sprintf("2T%d", ano))
		x.printCell(1, col+seq[2], titleStyle, fmt.Sprintf("3T%d", ano))
		x.printCell(1, col+seq[3], titleStyle, fmt.Sprintf("4T%d", ano))
		col += 4
	}

	row := 2
	for _, informe := range itr {
		if rapina.Zerado(informe.Valores) {
			continue
		}
		spc := space(informe.Codigo)
		x.printCell(row, 1, fontStyle, spc+informe.Codigo)
		x.printCell(row, 2, fontStyle, spc+informe.Descr)
		if strings.Count(informe.Codigo, ".") <= 1 {
			_ = f.SetCellStyle(sheet, cell(row, 1), cell(row, 2), titleStyle)
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
				if strings.HasPrefix(informe.Codigo, "1") || strings.HasPrefix(informe.Codigo, "2") {
					x.printCell(row, col+seq[3], numberStyle, valor.Anual)
				} else {
					x.printCell(row, col+seq[3], numberStyle, valor.T4)
				}

				if strings.Count(informe.Codigo, ".") <= 1 {
					_ = f.SetCellStyle(sheet, cell(row, col), cell(row, col+3), boldNumberStyle)
				}
			}
			col += 4
		}
		row++
	}

	// Auto-resize columns
	codWidth, descrWidth := colWidths(itr)
	_ = f.SetColWidth(sheet, "A", "A", codWidth)
	_ = f.SetColWidth(sheet, "B", "B", descrWidth)
	_ = f.SetColWidth(sheet, num2name(3), num2name(col+3), 12)

	// Freeze panes
	_ = f.SetPanes(sheet, &excelize.Panes{
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
			err := f.RemoveCol(sheet, num2name(initCol+i))
			fmt.Printf("RemoveCol(%s) [%v]\n", num2name(initCol+i), err)
		}
	}

	// Save spreadsheet
	if err := f.SaveAs("InformeTrimestral.xlsx"); err != nil {
		log.Fatal(err)
	}
}

type Excel struct {
	f         *excelize.File
	sheetName string
}

func (x *Excel) printCell(row, col int, style int, value interface{}) {
	_ = x.f.SetCellValue(x.sheetName, cell(row, col), value)
	_ = x.f.SetCellStyle(x.sheetName, cell(row, col), cell(row, col), style)
}

func cell(row, col int) string {
	return fmt.Sprintf("%s%d", num2name(col), row)
}

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
	return strings.Repeat("  ", strings.Count(str, "."))
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
		fmt.Printf("---> %c\n", ch)
		return 1.4
	}
	return width / 5.2
}
