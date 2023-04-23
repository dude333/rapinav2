// SPDX-FileCopyrightText: 2022 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil"
	serviço "github.com/dude333/rapinav2/pkg/contabil/servico"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
	"golang.org/x/text/encoding/charmap"
)

type CSV struct {
	svc contabil.Serviço
}

func NovoCSV(db *sqlx.DB) (*CSV, error) {
	svc, err := serviço.NovoDemonstraçãoFinanceira(db, "")
	if err != nil {
		return &CSV{}, err
	}

	return &CSV{svc: svc}, nil
}

func ImprimirCSV(db *sqlx.DB) {
	progress.Status("Imprimindo csv")
}

func Empresas(db *sqlx.DB, nome string) ([]rapina.Empresa, error) {
	c, err := NovoCSV(db)
	if err != nil {
		return []rapina.Empresa{}, err
	}

	return c.svc.Empresas(nome), nil
}

func Relatório(db *sqlx.DB, cnpj string, ano int) error {
	c, err := NovoCSV(db)
	if err != nil {
		return err
	}

	var dfps []*contabil.DemonstraçãoFinanceira

	for a := ano; ; a-- {
		dfp, err := c.svc.Relatório(cnpj, a)
		if err != nil {
			break
		}
		dfps = append(dfps, dfp)
	}

	if len(dfps) == 0 {
		return errors.New("vazio")
	}

	fmt.Println("SEP=;")
	fmt.Printf("Empresa: %s\n", dfps[0].Empresa.Nome)
	fmt.Printf("CNPJ:    %s\n\n", dfps[0].Empresa.CNPJ)

	writer := csv.NewWriter(os.Stdout)
	writer.Comma = rune(';')
	for i := range dfps {
		for _, c := range dfps[i].Contas {
			enc := charmap.Windows1252.NewEncoder()
			out, err := enc.String(c.Descr)
			if err != nil {
				progress.Debug(err.Error())
			}

			row := []string{
				c.Código,
				out,
				c.DataIniExerc,
				c.DataFimExerc,
				c.Total.String(),
			}
			err = writer.Write(row)
			if err != nil {
				log.Println("Cannot write to CSV file:", err)
			}
		}
	}
	writer.Flush()

	return nil
}
