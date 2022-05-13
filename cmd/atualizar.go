// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"strconv"
	"time"

	serviço "github.com/dude333/rapinav2/pkg/contabil/servico"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/spf13/cobra"
)

type flagsAtualizar struct {
	ano int
}

// atualizarCmd represents the atualizar command
var atualizarCmd = &cobra.Command{
	Use:     "atualizar",
	Aliases: []string{"update"},
	Short:   "Atualizar os dados do banco de dados",
	Long:    `Atualizar o banco de dados com as informações coletadas dos arquivos da CVM e B3`,
	Run:     atualizar,
}

func init() {
	atualizarCmd.Flags().IntVarP(&flags.atualizar.ano, "ano", "a", 0, "Ano do relatório")

	rootCmd.AddCommand(atualizarCmd)
}

func atualizar(cmd *cobra.Command, args []string) {
	svc, err := serviço.NovoDemonstraçãoFinanceira(db, flags.tempDir)
	if err != nil {
		panic(err)
	}

	progress.Status("{%d}", flags.atualizar.ano)

	anoi := 2010
	anof, err := strconv.Atoi(time.Now().Format("2006"))
	if err != nil {
		progress.Error(err)
		return
	}

	if flags.atualizar.ano >= 2000 {
		anoi = flags.atualizar.ano
		anof = anoi
	}

	trimestral := false
	for passo := 1; passo <= 2; passo++ {
		for ano := anof; ano >= anoi; ano-- {

			err := svc.Importar(ano, trimestral)
			if err != nil {
				progress.Error(err)
				continue
			}

		} // próx. ano
		trimestral = true
	} // próx. passo

}
