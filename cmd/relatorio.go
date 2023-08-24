package main

import (
	"fmt"
	"log"
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

func excel(data []rapina.InformeTrimestral) {
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

	_ = f.SetCellValue(sheetName, "A1", "Código")
	_ = f.SetCellValue(sheetName, "B1", "Descrição")
	_ = f.SetCellValue(sheetName, "C1", "Ano")
	_ = f.SetCellValue(sheetName, "D1", "T1")
	_ = f.SetCellValue(sheetName, "E1", "T2")
	_ = f.SetCellValue(sheetName, "F1", "T3")
	_ = f.SetCellValue(sheetName, "G1", "T4")

	for i, informe := range data {
		row := i + 2
		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), informe.Codigo)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), informe.Descr)
		col, _ := excelize.ColumnNameToNumber("C")
		for _, valor := range informe.Valores {
			_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", num2name(col), row), valor.Ano)
			col++
			_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", num2name(col), row), valor.T1)
			col++
			_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", num2name(col), row), valor.T2)
			col++
			_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", num2name(col), row), valor.T3)
			col++
			if strings.HasPrefix(informe.Codigo, "1") || strings.HasPrefix(informe.Codigo, "2") {
				_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", num2name(col), row), valor.Anual)
			} else {
				_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", num2name(col), row), valor.T4)
			}
			col++
		}
	}

	// Auto-resize columns
	// for i := 'A'; i <= 'H'; i++ {
	// 	colName := string(i)
	// 	f.SetColWidth(sheetName, colName, colName, 15) // Adjust width as needed
	// }

	if err := f.SetCellStyle(sheetName, "A1", "H1", titleStyle); err != nil {
		log.Fatal(err)
	}

	if err := f.SaveAs("InformeTrimestral.xlsx"); err != nil {
		log.Fatal(err)
	}
}

func num2name(col int) string {
	n, _ := excelize.ColumnNumberToName(col)
	return n
}
