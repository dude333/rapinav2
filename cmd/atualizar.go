// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	repositório "github.com/dude333/rapinav2/internal/contabil/repositorio"
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
	connStr := "file:/tmp/rapina.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
	db := sqlx.MustConnect("sqlite3", connStr)

	c := repositório.NovoCVM(os.TempDir())
	s, err := repositório.NovoSqlite(db)
	if err != nil {
		panic(err)
	}

	progress.Status("{%d}", flags.atualizar.ano)

	ctx := cmd.Context()

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

	for ano := anof; ano >= anoi; ano-- {

		for result := range c.Importar(ctx, ano) {
			if result.Error != nil {
				panic(err)
			}
			err = s.Salvar(ctx, result.DFP)
			if err != nil {
				fmt.Println("*", err)
			}
		}

	} // next ano

}
