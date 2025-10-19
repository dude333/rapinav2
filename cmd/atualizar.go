// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"time"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil"
	"github.com/dude333/rapinav2/pkg/cotacao"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/spf13/cobra"
)

type flagsAtualizar struct {
	ano   int
	tudo  bool
	force bool
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
	atualizarCmd.Flags().BoolVar(&flags.atualizar.tudo, "all", false, "Atualiza todos os anos, desde 2009")
	atualizarCmd.Flags().BoolVar(&flags.atualizar.force, "force", false, "Forçar atualização, mesmo que o dado já exista")

	rootCmd.AddCommand(atualizarCmd)
}

func atualizar(_ *cobra.Command, _ []string) {
	var anoi, anof int

	if flags.atualizar.tudo {
		anoi = 2010
		anof = time.Now().Year()
	} else if flags.atualizar.ano >= 2009 {
		progress.Status("{%d}", flags.atualizar.ano)
		anoi = flags.atualizar.ano
		anof = anoi
	} else {
		progress.Error(fmt.Errorf("escolher o ano com --ano <ano> ou --all"))
		return
	}

	svcContabil, err := contabil.NewService(db(), flags.tempDir, flags.atualizar.force)
	if err != nil {
		progress.Fatal(err)
	}

	importar := func(trimestral bool) {
		for ano := anof; ano >= anoi; ano-- {
			err := svcContabil.Import(ano, trimestral)
			if err != nil {
				progress.Error(err)
				continue
			}
		}
	}

	importar(false)
	importar(true)

	if err := cotacao.AtualizarTickers(db()); err != nil {
		progress.Error(err)
	}
}

func atualizar2(cmd *cobra.Command, _ []string) {
	cvm, err := rapina.NewCVM(db(), flags.tempDir, flags.atualizar.force)
	if err != nil {
		progress.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Minute)
	err = cvm.Importar(ctx, flags.atualizar.ano)
	cancel()
	if err != nil {
		progress.Fatal(err)
	}
}
