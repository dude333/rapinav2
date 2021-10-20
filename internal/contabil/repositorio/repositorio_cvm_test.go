// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositório

import (
	"context"
	"fmt"
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
	"github.com/jmoiron/sqlx"
	"os"
	"testing"
)

func Test_cvm_Importar(t *testing.T) {
	type args struct {
		ctx context.Context
		ano int
	}
	tests := []struct {
		name    string
		args    args
		want    <-chan contábil.ResultadoImportação
		wantErr bool
	}{
		{
			name: "deveria funcionar",
			args: args{
				ctx: context.Background(),
				ano: 2020,
			},
			want:    make(<-chan contábil.ResultadoImportação),
			wantErr: false,
		},
		{
			name: "deveria funcionar",
			args: args{
				ctx: context.Background(),
				ano: 2019,
			},
			want:    make(<-chan contábil.ResultadoImportação),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var db *sqlx.DB
			if testing.Short() {
				db = sqlx.MustConnect("sqlite3", ":memory:")
				db.SetMaxOpenConns(1)
			} else {
				connStr := "file:/tmp/rapina.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
				db = sqlx.MustConnect("sqlite3", connStr)
			}

			c := NovoCVM(os.TempDir())
			s, err := NovoSqlite(db)
			if err != nil {
				t.Fatal(err)
			}

			for result := range c.Importar(tt.args.ctx, tt.args.ano) {
				if (result.Error != nil) != tt.wantErr {
					t.Errorf("RepositórioImportaçãoDFP.Importar() error = %v, wantErr %v", result.Error, tt.wantErr)
					return
				}
				if result.Error != nil {
					fmt.Printf("=> %+v\n", result.Error)
				}
				err = s.Salvar(tt.args.ctx, result.DFP)
				if (err != nil) != tt.wantErr {
					t.Errorf("RepositórioEscritaDFP.Salvar() error = %v, wantErr %v, para DFP = %s | %s | %d", err, tt.wantErr,
						result.DFP.CNPJ, result.DFP.Nome, result.DFP.Ano)
				}
			}
		})
	}
}

func benchmarkconverteConta(c *cvmDFP, b *testing.B) {
	for n := 0; n < b.N; n++ {
		c.converteConta()
	}
}

var cc = []cvmDFP{
	{
		CNPJ:         "C1",
		Nome:         "N1",
		Ano:          "2020",
		Consolidado:  false,
		Versão:       "1",
		Código:       "1.1",
		Descr:        "D1",
		GrupoDFP:     "Balanço Patrimonial Passivo",
		DataFimExerc: "2020-12-30",
		OrdemExerc:   "ÚLTIMO",
		Valor:        12.34,
		Escala:       1,
		Moeda:        "R$",
	},
	{
		CNPJ:         "C2",
		Nome:         "N2",
		Ano:          "2020",
		Consolidado:  false,
		Versão:       "1",
		Código:       "1.1",
		Descr:        "D1",
		GrupoDFP:     "Demonstração de Valor Adicionado",
		DataFimExerc: "2020-12-30",
		OrdemExerc:   "ÚLTIMO",
		Valor:        12.34,
		Escala:       1,
		Moeda:        "R$",
	},
}

func BenchmarkConverteConta0(b *testing.B) { benchmarkconverteConta(&cc[0], b) }
func BenchmarkConverteConta1(b *testing.B) { benchmarkconverteConta(&cc[1], b) }
