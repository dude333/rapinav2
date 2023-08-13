// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"time"

	"github.com/dude333/rapinav2/pkg/contabil"
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
	progress.Status("{%d}", flags.atualizar.ano)

	anoi := 2010
	anof := time.Now().Year()

	if flags.atualizar.ano >= 2000 {
		anoi = flags.atualizar.ano
		anof = anoi
	}

	dfp, err := contabil.NovaDemonstraçãoFinanceira(db(), flags.tempDir)
	if err != nil {
		panic(err)
	}

	importar := func(trimestral bool) {
		for ano := anof; ano >= anoi; ano-- {
			err := dfp.Importar(ano, trimestral)
			if err != nil {
				progress.Error(err)
				continue
			}
		}
	}

	importar(false)
	importar(true)

}
