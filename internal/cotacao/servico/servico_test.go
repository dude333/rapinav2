// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package serviço

import (
	"context"
	"fmt"
	cotação "github.com/dude333/rapinav2/internal/cotacao"
	"reflect"
	"testing"
	"time"
)

var (
	_cache    map[string]cotação.Ativo
	_ativo1   cotação.Ativo
	_exemplos = []*cotação.Ativo{}
)

func init() {
	const m = "R$"
	d, _ := cotação.NovaData("2021-10-14")

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

	for i := 1; i <= 10; i++ {
		r := cotação.Ativo{
			Código:       fmt.Sprintf("TEST%d", i),
			Data:         d,
			Abertura:     cotação.Dinheiro{Valor: 1, Moeda: m},
			Máxima:       cotação.Dinheiro{Valor: 2, Moeda: m},
			Mínima:       cotação.Dinheiro{Valor: 0.8, Moeda: m},
			Encerramento: cotação.Dinheiro{Valor: 1.6, Moeda: m},
			Volume:       1000000,
		}
		_exemplos = append(_exemplos, &r)
	}

}

// apiMockOk implementa domínio.RepositórioLeituraAtivo
type apiMockOk struct{}

func (r *apiMockOk) Importar(ctx context.Context, data cotação.Data) <-chan cotação.Resultado {
	// _cache[_ativo1.Código+_ativo1.Data.String()] = _ativo1
	results := make(chan cotação.Resultado)
	go func() {
		defer close(results)

		for _, ex := range _exemplos {
			result := cotação.Resultado{
				Error: nil,
				Ativo: ex,
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

// apiMockFail implementa domínio.RepositórioLeituraAtivo
type apiMockFail struct{}

func (r *apiMockFail) Importar(ctx context.Context, data cotação.Data) <-chan cotação.Resultado {
	// _cache[_ativo1.Código+_ativo1.Data.String()] = _ativo1
	results := make(chan cotação.Resultado)
	go func() {
		defer close(results)

		for _, ex := range _exemplos {
			result := cotação.Resultado{
				Error: nil,
				Ativo: ex,
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
		api []Importação
		bd  LeituraEscrita
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
			name: "deveria funcionar sem bd",
			fields: fields{
				api: []Importação{&apiMockFail{}},
				bd:  nil, // testa se o serviço.Cotação ignora o bd caso seja 'nil'
			},
			args:    args{"TEST3", d1},
			want:    &_ativo1,
			wantErr: false,
		},
		{
			name: "não deveria funcionar",
			fields: fields{
				api: nil,
				bd:  nil,
			},
			args:    args{"TEST3", d1},
			want:    nil,
			wantErr: true,
		},
		{
			name: "deveria funcionar",
			fields: fields{
				api: []Importação{&apiMockOk{}},
				bd:  &bdMock{},
			},
			args:    args{"TEST3", d1},
			want:    &_ativo1,
			wantErr: false,
		},
		{
			name: "deveria funcionar com dados do cache",
			fields: fields{
				api: []Importação{&apiMockFail{}},
				bd:  &bdMock{},
			},
			args:    args{"TEST3", d1},
			want:    &_ativo1,
			wantErr: false,
		},
		{
			name: "deveria funcionar com um repo bom e um ruim",
			fields: fields{
				api: []Importação{
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
				api: []Importação{
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
			s := NovoServiço(
				tt.fields.api,
				tt.fields.bd,
			)
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
