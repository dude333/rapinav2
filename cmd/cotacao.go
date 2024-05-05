// SPDX-FileCopyrightText: 2024 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/cotacao"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/spf13/cobra"
)

type flagsCotação struct{}

// cotacaoCmd represents the cotacao command
var cotacaoCmd = &cobra.Command{
	Use:   "cotacao",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: cotação,
}

func init() {
	rootCmd.AddCommand(cotacaoCmd)
}

// report --> need quote --> get quote and VWAP
// on report: quote.FromCNPJ("84.429.695/0001-11", "2019-01-01", -10)
// 1. CNPJ -> código
// 2. código -> cotações
// 3. VWAP
func cotação(_ *cobra.Command, args []string) {
	for empresa := range menuEmpresas() {
		// empresa := rapina.Empresa{
		// 	CNPJ: "84.429.695/0001-11",
		// 	Nome: "WEG S.A.",
		// }
		if empresa.CNPJ == "" {
			continue
			// return
		}
		progress.Status("empresa: %+v", empresa)

		c, err := cotacao.NovoServiço(db(), flags.tempDir)
		if err != nil {
			progress.Fatal(err)
		}

		d, err := rapina.NovaData(args[0])
		if err != nil {
			d = rapina.DiaUtilAnterior(rapina.Hoje())
		}

		var all []*rapina.Cotação
		for n := 0; n < 10; n++ {
			d = rapina.DiaUtilAnterior(d)
			ativos, err := c.Cotação(empresa, d)
			if err != nil {
				progress.Error(err)
				continue
			}
			all = append(all, ativos...)
		}
		for _, ativo := range all {
			progress.Status("%+v", ativo)
		}
		progress.Status("VWAP: %+v", rapina.VWAP(all))
	}
}
