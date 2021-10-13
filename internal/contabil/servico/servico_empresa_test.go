// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package serviço

import (
	"context"
	"fmt"
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
	"reflect"
	"strconv"
	"testing"
)

var (
	_cache   map[uint32]*contábil.Empresa
	_exemplo = contábil.Empresa{
		CNPJ:   "1234",
		Ano:    2021,
		Contas: []contábil.Conta{},
	}
)

func init() {
	_cache = make(map[uint32]*contábil.Empresa)
	conta := make(contábil.Conta)
	conta[12342021] = contábil.Dinheiro{
		Valor:  1234.56,
		Escala: 1000,
		Moeda:  "R$",
	}
	_exemplo.Contas = []contábil.Conta{conta}
}

type repoBD struct{}

func (r *repoBD) Ler(ctx context.Context, cnpj string, ano int) (*contábil.Empresa, error) {
	x := fmt.Sprintf("%s%d", cnpj, ano)
	y, _ := strconv.Atoi(x)
	return _cache[uint32(y)], nil
}

func (r *repoBD) Salvar(ctx context.Context, e *contábil.Empresa) error {
	x := fmt.Sprintf("%s%d", e.CNPJ, e.Ano)
	y, _ := strconv.Atoi(x)
	_cache[uint32(y)] = e

	return nil
}

type repoAPI struct {
	bd contábil.RepositórioEscritaEmpresa
}

func (r *repoAPI) Importar(ctx context.Context, ano int) error {
	return r.Salvar(ctx, &_exemplo)
}

func (r *repoAPI) Salvar(ctx context.Context, e *contábil.Empresa) error {
	if r.bd == nil {
		return fmt.Errorf("bd não definido")
	}
	return r.bd.Salvar(ctx, e)
}

func Test_empresa_Importar(t *testing.T) {
	type fields struct {
		api contábil.RepositórioImportaçãoEmpresa
		bd  contábil.RepositórioLeituraEscritaEmpresa
	}
	type args struct {
		cnpj string
		ano  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *contábil.Empresa
		wantErr bool
	}{
		{
			name: "não deveria funcionar",
			fields: fields{
				api: &repoAPI{},
				bd:  &repoBD{},
			},
			args:    args{"1234", 2021},
			want:    nil,
			wantErr: true,
		},
		{
			name: "deveria funcionar",
			fields: fields{
				api: &repoAPI{bd: &repoBD{}},
				bd:  &repoBD{},
			},
			args:    args{"1234", 2021},
			want:    &_exemplo,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &empresa{
				api: tt.fields.api,
				bd:  tt.fields.bd,
			}
			if err := e.Importar(tt.args.ano); (err != nil) != tt.wantErr {
				t.Errorf("empresa.Importar() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			got, err := e.Relatório(tt.args.cnpj, tt.args.ano)
			if (err != nil) != tt.wantErr {
				t.Errorf("empresa.Relatório() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServiçoEmpresa.Relatório() = %v, want %v", got, tt.want)
			}
		})
	}
}
