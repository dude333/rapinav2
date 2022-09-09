// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/dude333/rapinav2/pkg/contabil/apresentacao/csv"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/spf13/cobra"
)

type flagsRelatorio struct {
	ano       int
	trimestre string
}

// relatorioCmd represents the relatorio command
var relatorioCmd = &cobra.Command{
	Use: "relatorio [-a <ano>|-t <trimestre>] empresa",
	Long: `relatorio [-a <ano>|-t <trimestre>] empresa
Exemplo:
	relatorio -a 2021 WEG
	relatorio -t 2T2020 AMBEV`,
	Aliases: []string{"r", "relat", "report"},
	Short:   "Relatorio",
	Run:     relatorio,
	Args:    cobra.MinimumNArgs(1),
}

func init() {
	relatorioCmd.Flags().IntVarP(&flags.relatorio.ano, "ano", "a", 0, "Ano do relatório")
	relatorioCmd.Flags().StringVarP(&flags.relatorio.trimestre, "trimestre", "t", "", "Trimestre do relatório")

	rootCmd.AddCommand(relatorioCmd)
}

func relatorio(cmd *cobra.Command, args []string) {
	if flags.relatorio.ano > 2000 && len(args) > 0 && len(args[0]) > 1 {
		err := csv.Relatório(db(), args[0], flags.relatorio.ano)
		if err != nil {
			progress.Error(err)
		}
		return
	}

	empresas, err := csv.Empresas(db(), args[0])
	if err != nil {
		progress.Error(err)
	}
	lista := make([]string, len(empresas))
	for i := range empresas {
		lista[i] = empresas[i].Nome
	}
	progress.Status(promptUser(empresas))
}
