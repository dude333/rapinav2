// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package serviço

import (
	"context"
	"fmt"
	cotação "github.com/dude333/rapinav2/internal/cotacao/dominio"
	"reflect"
	"testing"
)

var (
	_cache  map[string]cotação.Ativo
	_ativo1 cotação.Ativo
)

func init() {
	const m = "R$"
	d, _ := cotação.NovaData("2021-10-09")

	_cache = make(map[string]cotação.Ativo)

	_ativo1 = cotação.Ativo{
		Código:       "TEST3",
		Data:         d,
		Abertura:     cotação.Dinheiro{Valor: 1, Moeda: m},
		Máxima:       cotação.Dinheiro{Valor: 2, Moeda: m},
		Mínima:       cotação.Dinheiro{Valor: 0.8, Moeda: m},
		Encerramento: cotação.Dinheiro{Valor: 1.6, Moeda: m},
		Volume:       1000000,
	}
}

// apiMockOk implementa domínio.RepositórioLeituraAtivo
type apiMockOk struct {
	db cotação.RepositórioLeituraEscritaAtivo
}

func (r *apiMockOk) Cotação(ctx context.Context, código string, data cotação.Data) (*cotação.Ativo, error) {
	_cache[_ativo1.Código+_ativo1.Data.String()] = _ativo1
	return &_ativo1, nil
}

func (r *apiMockOk) Salvar(ctx context.Context, ativo *cotação.Ativo) error {
	if r.db == nil {
		return fmt.Errorf("bd inválido")
	}
	return r.db.Salvar(ctx, ativo)
}

// apiMockFail implementa domínio.RepositórioLeituraAtivo
type apiMockFail struct {
	db cotação.RepositórioLeituraEscritaAtivo
}

func (r *apiMockFail) Cotação(ctx context.Context, código string, data cotação.Data) (*cotação.Ativo, error) {
	return &cotação.Ativo{}, ErrCotaçãoNãoEncontrada
}

func (r *apiMockFail) Salvar(ctx context.Context, ativo *cotação.Ativo) error {
	if r.db == nil {
		return fmt.Errorf("bd inválido")
	}
	return r.db.Salvar(ctx, ativo)
}

// bdMock implementa domínio.RepositórioLeituraEscritaAtivo
type bdMock struct{}

func (r *bdMock) Cotação(ctx context.Context, código string, data cotação.Data) (*cotação.Ativo, error) {
	ativo, ok := _cache[código+data.String()]
	if !ok {
		return &cotação.Ativo{}, ErrCotaçãoNãoEncontrada
	}
	return &ativo, nil
}

func (r *bdMock) Salvar(ctx context.Context, ativo *cotação.Ativo) error {
	_cache[ativo.Código+ativo.Data.String()] = *ativo
	return nil
}

// TESTES -------------------------------------------------

func TestServiçoAtivo_Cotação(t *testing.T) {
	d1, _ := cotação.NovaData("2021-10-09")

	type fields struct {
		api []cotação.RepositórioLeituraEscritaAtivo
		bd  cotação.RepositórioLeituraEscritaAtivo
	}
	type args struct {
		código string
		data   cotação.Data
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *cotação.Ativo
		wantErr bool
	}{
		{
			name: "não deveria funcionar",
			fields: fields{
				api: []cotação.RepositórioLeituraEscritaAtivo{&apiMockFail{}},
				bd:  nil, // testa se o serviço.Cotação ignora o bd caso seja 'nil'
			},
			args:    args{"TEST3", d1},
			want:    &cotação.Ativo{},
			wantErr: true,
		},
		{
			name: "não deveria funcionar também",
			fields: fields{
				api: nil, // testa se o serviço.Cotação ignora a api caso seja 'nil'
				bd:  nil, // testa se o serviço.Cotação ignora o bd caso seja 'nil'
			},
			args:    args{"TEST3", d1},
			want:    &cotação.Ativo{},
			wantErr: true,
		},
		{
			name: "deveria funcionar",
			fields: fields{
				api: []cotação.RepositórioLeituraEscritaAtivo{&apiMockOk{}},
				bd:  &bdMock{},
			},
			args:    args{"TEST3", d1},
			want:    &_ativo1,
			wantErr: false,
		},
		{
			name: "deveria funcionar com dados do cache",
			fields: fields{
				api: []cotação.RepositórioLeituraEscritaAtivo{&apiMockFail{}},
				bd:  &bdMock{},
			},
			args:    args{"TEST3", d1},
			want:    &_ativo1,
			wantErr: false,
		},
		{
			name: "deveria funcionar com um repo bom e um ruim",
			fields: fields{
				api: []cotação.RepositórioLeituraEscritaAtivo{
					&apiMockOk{},
					&apiMockFail{},
				},
				bd: &bdMock{},
			},
			args:    args{"TEST3", d1},
			want:    &_ativo1,
			wantErr: false,
		},
		{
			name: "deveria funcionar com um repo ruim e um bom",
			fields: fields{
				api: []cotação.RepositórioLeituraEscritaAtivo{
					&apiMockFail{},
					&apiMockOk{},
				},
				bd: &bdMock{},
			},
			args:    args{"TEST3", d1},
			want:    &_ativo1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ativo{
				api: tt.fields.api,
				bd:  tt.fields.bd,
			}
			got, err := s.Cotação(tt.args.código, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServiçoAtivos.Cotação() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServiçoAtivos.Cotação() = %v, want %v", got, tt.want)
			}
		})
	}
}
