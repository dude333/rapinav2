// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package contabil

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil/dominio"
	"github.com/dude333/rapinav2/pkg/progress"
)

var (
	_cache    map[uint32]*dominio.DemonstraçãoFinanceira
	_exemplos = []*dominio.DemonstraçãoFinanceira{}
)

func init() {
	_cache = make(map[uint32]*dominio.DemonstraçãoFinanceira)

	for i := 1; i <= 10; i++ {
		r := dominio.DemonstraçãoFinanceira{
			Empresa: rapina.Empresa{
				CNPJ: fmt.Sprintf("%010d", i),
				Nome: fmt.Sprintf("Empresa %02d", i),
			},
			Ano: 2021,
			Contas: []dominio.Conta{
				{
					Código:       fmt.Sprintf("%d.%d", i, i),
					Descr:        fmt.Sprintf("Descrição %d", i),
					Grupo:        "Grupo DFP",
					DataFimExerc: "2021-12-31",
					Total: rapina.Dinheiro{
						Valor:  float64(i),
						Escala: 1000,
						Moeda:  "R$",
					},
				},
			},
		}
		_exemplos = append(_exemplos, &r)
	}
}

// Implementação de repositórios de teste ---

type repoBD struct{}

func (r repoBD) Ler(ctx context.Context, cnpj string, ano int) (*dominio.DemonstraçãoFinanceira, error) {
	x := fmt.Sprintf("%s%d", cnpj, ano)
	y, _ := strconv.Atoi(x)
	return _cache[uint32(y)], nil
}

func (r repoBD) Hashes() []string {
	var hashes []string
	return hashes
}

func (r repoBD) SalvarHash(ctx context.Context, hash string) error {
	return nil
}

func (r *repoBD) Salvar(_ context.Context, e *dominio.DemonstraçãoFinanceira) error {
	x := fmt.Sprintf("%s%d", e.CNPJ, e.Ano)
	y, _ := strconv.Atoi(x)
	_cache[uint32(y)] = e
	progress.Debug("Saving %#v", e)

	return nil
}

func (r repoBD) Empresas(ctx context.Context, nome string) []rapina.Empresa {
	return []rapina.Empresa{
		{Nome: "a", CNPJ: "cnpjA"},
		{Nome: "b", CNPJ: "cnpjB"},
		{Nome: "c", CNPJ: "cnpjC"},
	}
}

type repoAPI struct{}

func (r *repoAPI) Importar(ctx context.Context, ano int, trim bool) <-chan dominio.Resultado {
	results := make(chan dominio.Resultado)
	go func() {
		defer close(results)

		for _, ex := range _exemplos {
			result := dominio.Resultado{
				Empresa: ex,
				Error:   nil,
			}
			select {
			case <-ctx.Done():
				return
			case results <- result:
				time.Sleep(1 * time.Millisecond)
			}
		}

	}()
	return results

}

// Testes ---

func Test_registro_Importar(t *testing.T) {
	type fields struct {
		api Importação
		db  LeituraEscrita
	}
	type args struct {
		ano        int
		trimestral bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "não deveria funcionar sem bd",
			fields: fields{
				api: &repoAPI{},
				db:  nil,
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "não deveria funcionar sem api e bd",
			fields: fields{
				api: nil,
				db:  nil,
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "deveria funcionar",
			fields: fields{
				api: &repoAPI{},
				db:  &repoBD{},
			},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := novoSvcDemonstraçãoFinanceira(
				tt.fields.api,
				tt.fields.db,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("NovoServiçoDemonstraçãoFinanceira() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			err = r.Importar(tt.args.ano, tt.args.trimestral)
			if (err != nil) != tt.wantErr {
				t.Errorf("registro.Importar() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil || tt.fields.db == nil {
				return
			}
			// Verifica se os dados foram salvos no "banco de dados"
			for _, r := range _exemplos {
				x := fmt.Sprintf("%s%d", r.CNPJ, r.Ano)
				y, _ := strconv.Atoi(x)
				c, ok := _cache[uint32(y)]
				if !ok {
					t.Fatal("item não encontrado no cache")
				}
				if c.Nome != r.Nome {
					t.Fatalf("valor salvo esperado: %v, recebido: %v", r.Nome, c.Nome)
				}
			}
		})
	}
}

func Test_dfp_Empresas(t *testing.T) {
	type fields struct {
		api Importação
		bd  LeituraEscrita
	}
	type args struct {
		nome string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []rapina.Empresa
		wantErr bool
	}{
		{
			name: "deveria funcionar",
			fields: fields{
				api: &repoAPI{},
				bd:  &repoBD{},
			},
			args: args{nome: "a"},
			want: []rapina.Empresa{
				{Nome: "a", CNPJ: "cnpjA"},
				{Nome: "b", CNPJ: "cnpjB"},
				{Nome: "c", CNPJ: "cnpjC"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := novoSvcDemonstraçãoFinanceira(
				tt.fields.api,
				tt.fields.bd,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("NovoServiçoDemonstraçãoFinanceira() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got := r.Empresas(tt.args.nome); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dfp.Empresas() = %v, want %v", got, tt.want)
			}
		})
	}
}
