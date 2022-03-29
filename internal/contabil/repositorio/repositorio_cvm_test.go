// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

import (
	"context"
	"fmt"
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
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
			c, err := NovoCVM(ComLimiteAnual(5))
			if err != nil {
				t.Fatal(err)
			}

			var cfg ConfigFn
			if testing.Short() {
				cfg = RodarBDNaMemória()
			} else {
				cfg = DirBD(os.TempDir())
			}
			s, err := NovoSqlite(cfg)
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
				err = s.Salvar(tt.args.ctx, result.Empresa)
				if (err != nil) != tt.wantErr {
					t.Errorf("RepositórioEscritaDFP.Salvar() error = %v, wantErr %v, para Empresa = %s | %s | %d", err, tt.wantErr,
						result.Empresa.CNPJ, result.Empresa.Nome, result.Empresa.Ano)
				}
			}
		})
	}
}

func Test_meses(t *testing.T) {
	type args struct {
		ini string
		fim string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "deveria retornar erro",
			args:    args{ini: "2021-01-01", fim: "2021-12"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "deveria retornar data inválida",
			args:    args{ini: "2021-01-01", fim: "aaaa-mm-dd"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "deveria retornar 1 meses",
			args:    args{ini: "2020-01-01", fim: "2020-01-31"},
			want:    1,
			wantErr: false,
		},
		{
			name:    "deveria retornar 3 meses",
			args:    args{ini: "2020-01-01", fim: "2020-03-31"},
			want:    3,
			wantErr: false,
		},
		{
			name:    "deveria retornar 6 meses",
			args:    args{ini: "2020-09-01", fim: "2021-03-31"},
			want:    6,
			wantErr: false,
		},
		{
			name:    "deveria retornar 6 meses",
			args:    args{ini: "2021-01-01", fim: "2021-06-30"},
			want:    6,
			wantErr: false,
		},
		{
			name:    "deveria retornar 9 meses",
			args:    args{ini: "2021-01-01", fim: "2021-09-30"},
			want:    9,
			wantErr: false,
		},
		{
			name:    "deveria retornar 12 meses",
			args:    args{ini: "2021-01-01", fim: "2021-12-31"},
			want:    12,
			wantErr: false,
		},
		{
			name:    "deveria retornar 12 meses",
			args:    args{ini: "", fim: "2021-12-31"},
			want:    12,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := meses(tt.args.ini, tt.args.fim)
			if (err != nil) != tt.wantErr {
				t.Errorf("meses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("meses() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ==== BENCHMARKS ====

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

func benchmarkMeses(dataI, dataF string, b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = meses(dataI, dataF)
	}
}

func BenchmarkMeses0(b *testing.B) { benchmarkMeses("2020-01-01", "2020-09-30", b) }
func BenchmarkMeses1(b *testing.B) { benchmarkMeses("2019-09-01", "2020-09-30", b) }
