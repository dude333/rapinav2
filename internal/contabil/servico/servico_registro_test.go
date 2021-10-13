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
	_cache   map[uint32]*contábil.Registro
	_exemplo = contábil.Registro{
		CNPJ:         "1234",
		Empresa:      "Empresa1",
		Ano:          2021,
		DataFimExerc: "2021-12-31",
		Versão:       1,
		Total: contábil.Dinheiro{
			Valor:  123.45,
			Escala: 1000,
			Moeda:  "R$",
		},
	}
)

func init() {
	_cache = make(map[uint32]*contábil.Registro)
}

type repoBD struct{}

func (r *repoBD) Ler(ctx context.Context, cnpj string, ano int) (*contábil.Registro, error) {
	x := fmt.Sprintf("%s%d", cnpj, ano)
	y, _ := strconv.Atoi(x)
	return _cache[uint32(y)], nil
}

func (r *repoBD) Salvar(ctx context.Context, e *contábil.Registro) error {
	x := fmt.Sprintf("%s%d", e.CNPJ, e.Ano)
	y, _ := strconv.Atoi(x)
	_cache[uint32(y)] = e

	return nil
}

type repoAPI struct {
	bd contábil.RepositórioEscritaRegistro
}

func (r *repoAPI) Importar(ctx context.Context, ano int) error {
	return r.Salvar(ctx, &_exemplo)
}

func (r *repoAPI) Salvar(ctx context.Context, e *contábil.Registro) error {
	if r.bd == nil {
		return fmt.Errorf("bd não definido")
	}
	return r.bd.Salvar(ctx, e)
}

func Test_registro_Importar(t *testing.T) {
	type fields struct {
		api contábil.RepositórioImportaçãoRegistro
		bd  contábil.RepositórioLeituraEscritaRegistro
	}
	type args struct {
		cnpj string
		ano  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *contábil.Registro
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
			r := &registro{
				api: tt.fields.api,
				bd:  tt.fields.bd,
			}
			if err := r.Importar(tt.args.ano); (err != nil) != tt.wantErr {
				t.Errorf("empresa.Importar() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			got, err := r.Relatório(tt.args.cnpj, tt.args.ano)
			if (err != nil) != tt.wantErr {
				t.Errorf("empresa.Relatório() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServiçoRegistro.Relatório() = %v, want %v", got, tt.want)
			}
		})
	}
}
