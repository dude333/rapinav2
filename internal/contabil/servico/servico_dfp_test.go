// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package servico_test

import (
	"context"
	"fmt"
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
	serviço "github.com/dude333/rapinav2/internal/contabil/servico"
	"reflect"
	"strconv"
	"testing"
	"time"
)

var (
	_cache    map[uint32]*contábil.DFP
	_exemplos = []*contábil.DFP{}
)

func init() {
	_cache = make(map[uint32]*contábil.DFP)

	for i := 1; i <= 10; i++ {
		r := contábil.DFP{
			CNPJ: fmt.Sprintf("%010d", i),
			Nome: fmt.Sprintf("Empresa %02d", i),
			Ano:  2021,
			Contas: []contábil.Conta{
				{
					Código:       fmt.Sprintf("%d.%d", i, i),
					Descr:        fmt.Sprintf("Descrição %d", i),
					Grupo:        "Grupo DFP",
					DataFimExerc: "2021-12-31",
					Total: contábil.Dinheiro{
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

func (r repoBD) Ler(ctx context.Context, cnpj string, ano int) (*contábil.DFP, error) {
	x := fmt.Sprintf("%s%d", cnpj, ano)
	y, _ := strconv.Atoi(x)
	return _cache[uint32(y)], nil
}

func (r *repoBD) Salvar(ctx context.Context, e *contábil.DFP) error {
	x := fmt.Sprintf("%s%d", e.CNPJ, e.Ano)
	y, _ := strconv.Atoi(x)
	_cache[uint32(y)] = e

	return nil
}

func (r repoBD) Empresas(ctx context.Context, nome string) []string {
	return []string{"a", "b", "c"}
}

type repoAPI struct{}

func (r *repoAPI) Importar(ctx context.Context, ano int) <-chan contábil.ResultadoImportaçãoDFP {
	results := make(chan contábil.ResultadoImportaçãoDFP)
	go func() {
		defer close(results)

		for _, ex := range _exemplos {
			result := contábil.ResultadoImportaçãoDFP{
				DFP:   ex,
				Error: nil,
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
		api contábil.RepositórioImportaçãoDFP
		bd  contábil.RepositórioLeituraEscritaDFP
	}
	type args struct {
		ano int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "deveria funcionar sem bd",
			fields: fields{
				api: &repoAPI{},
				bd:  nil,
			},
			args:    args{},
			wantErr: false,
		},
		{
			name: "não deveria funcionar sem api e bd",
			fields: fields{
				api: nil,
				bd:  nil,
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "deveria funcionar",
			fields: fields{
				api: &repoAPI{},
				bd:  &repoBD{},
			},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := serviço.NovoDFP(
				tt.fields.api,
				tt.fields.bd,
			)
			err := r.Importar(tt.args.ano)
			if (err != nil) != tt.wantErr {
				t.Errorf("registro.Importar() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil || tt.fields.bd == nil {
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
		api contábil.RepositórioImportaçãoDFP
		bd  contábil.RepositórioLeituraEscritaDFP
	}
	type args struct {
		nome string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "deveria funcionar",
			fields: fields{
				api: nil,
				bd:  &repoBD{},
			},
			args:    args{nome: "a"},
			want:    []string{"a", "b", "c"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := serviço.NovoDFP(
				tt.fields.api,
				tt.fields.bd,
			)
			if got := r.Empresas(tt.args.nome); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dfp.Empresas() = %v, want %v", got, tt.want)
			}
		})
	}
}
