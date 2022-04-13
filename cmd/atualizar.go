// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package cmd

import (
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	serviço "github.com/dude333/rapinav2/internal/contabil/servico"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
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
	dirDB := os.TempDir()
	dirDB = strings.ReplaceAll(dirDB, "\\", "/")
	err := os.MkdirAll(dirDB, os.ModePerm)
	if err != nil {
		panic(err)
	}

	filename := "rapina.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
	connStr := "file:" + path.Join(dirDB, filename)

	db, err := sqlx.Connect("sqlite3", connStr)
	if err != nil {
		panic(err)
	}

	svc, err := serviço.NovoDemonstraçãoFinanceira(db)
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
