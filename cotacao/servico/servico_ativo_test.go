package serviço

import (
	"context"
	domínio "github.com/dude333/rapinav2/cotacao/dominio"
	"reflect"
	"testing"
)

var (
	_cache  map[string]domínio.Ativo
	_ativo1 domínio.Ativo
)

func init() {
	const m = "R$"
	d, _ := domínio.NovaData("2021-10-09")

	_cache = make(map[string]domínio.Ativo)

	_ativo1 = domínio.Ativo{
		Código:       "TEST3",
		Data:         d,
		Abertura:     domínio.Dinheiro{Valor: 1, Moeda: m},
		Máxima:       domínio.Dinheiro{Valor: 2, Moeda: m},
		Mínima:       domínio.Dinheiro{Valor: 0.8, Moeda: m},
		Encerramento: domínio.Dinheiro{Valor: 1.6, Moeda: m},
		Volume:       domínio.Dinheiro{Valor: 1000000, Moeda: m},
	}
}

// apiMockOk implementa domínio.RepositórioLeituraAtivo
type apiMockOk struct{}

func (r *apiMockOk) Cotação(ctx context.Context, código string, data domínio.Data) (*domínio.Ativo, error) {
	return &_ativo1, nil
}

// apiMockFail implementa domínio.RepositórioLeituraAtivo
type apiMockFail struct{}

func (r *apiMockFail) Cotação(ctx context.Context, código string, data domínio.Data) (*domínio.Ativo, error) {
	return &domínio.Ativo{}, ErrCotaçãoNãoEncontrada
}

// bdMock implementa domínio.RepositórioLeituraEscritaAtivo
type bdMock struct{}

func (r *bdMock) Cotação(ctx context.Context, código string, data domínio.Data) (*domínio.Ativo, error) {
	ativo, ok := _cache[código+data.String()]
	if !ok {
		return &domínio.Ativo{}, ErrCotaçãoNãoEncontrada
	}
	return &ativo, nil
}

func (r *bdMock) Salvar(ctx context.Context, ativo *domínio.Ativo) error {
	_cache[ativo.Código+ativo.Data.String()] = *ativo
	return nil
}

// TESTES -------------------------------------------------

func TestServiçoAtivo_Cotação(t *testing.T) {
	d1, _ := domínio.NovaData("2021-10-09")

	type fields struct {
		api []domínio.RepositórioLeituraAtivo
		bd  domínio.RepositórioLeituraEscritaAtivo
	}
	type args struct {
		código string
		data   domínio.Data
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *domínio.Ativo
		wantErr bool
	}{
		{
			name: "não deveria funcionar",
			fields: fields{
				api: []domínio.RepositórioLeituraAtivo{&apiMockFail{}},
				bd:  nil, // testa se o serviço.Cotação ignora o bd caso seja 'nil'
			},
			args:    args{"TEST3", d1},
			want:    &domínio.Ativo{},
			wantErr: true,
		},
		{
			name: "não deveria funcionar também",
			fields: fields{
				api: nil, // testa se o serviço.Cotação ignora a api caso seja 'nil'
				bd:  nil, // testa se o serviço.Cotação ignora o bd caso seja 'nil'
			},
			args:    args{"TEST3", d1},
			want:    &domínio.Ativo{},
			wantErr: true,
		},
		{
			name: "deveria funcionar",
			fields: fields{
				api: []domínio.RepositórioLeituraAtivo{&apiMockOk{}},
				bd:  &bdMock{},
			},
			args:    args{"TEST3", d1},
			want:    &_ativo1,
			wantErr: false,
		},
		{
			name: "deveria funcionar com dados do cache",
			fields: fields{
				api: []domínio.RepositórioLeituraAtivo{&apiMockFail{}},
				bd:  &bdMock{},
			},
			args:    args{"TEST3", d1},
			want:    &_ativo1,
			wantErr: false,
		},
		{
			name: "deveria funcionar com um repo bom e um ruim",
			fields: fields{
				api: []domínio.RepositórioLeituraAtivo{
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
				api: []domínio.RepositórioLeituraAtivo{
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
