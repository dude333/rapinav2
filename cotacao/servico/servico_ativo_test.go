package serviço

import (
	"context"
	"errors"
	domínio "github.com/dude333/rapinav2/cotacao/dominio"
	"reflect"
	"testing"
)

var (
	_ativo1 domínio.Ativo
)

func init() {
	const m = "R$"
	d, _ := domínio.NovaData("2021-10-09")

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

// RepositórioAtivoMockOk implementa a interface RepositórioAtivos
type RepositórioAtivoMockOk struct{}

func (r *RepositórioAtivoMockOk) Cotação(ctx context.Context, código string, data domínio.Data) (*domínio.Ativo, error) {
	return &_ativo1, nil
}

// RepositórioAtivoMockFail implementa a interface RepositórioAtivos
type RepositórioAtivoMockFail struct{}

func (r *RepositórioAtivoMockFail) Cotação(ctx context.Context, código string, data domínio.Data) (*domínio.Ativo, error) {
	return &domínio.Ativo{}, errors.New("falha ao buscar dado no repositório MockFail")
}

func TestServiçoAtivos_Cotação(t *testing.T) {
	d1, _ := domínio.NovaData("2021-10-09")

	type fields struct {
		repos []domínio.RepositórioAtivo
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
			name:    "deveria funcionar",
			fields:  fields{[]domínio.RepositórioAtivo{&RepositórioAtivoMockOk{}}},
			args:    args{"TEST3", d1},
			want:    &_ativo1,
			wantErr: false,
		},
		{
			name:    "não deveria funcionar",
			fields:  fields{[]domínio.RepositórioAtivo{&RepositórioAtivoMockFail{}}},
			args:    args{"TEST3", d1},
			want:    &domínio.Ativo{},
			wantErr: true,
		},
		{
			name: "deveria funcionar com um repo bom e um ruim",
			fields: fields{
				repos: []domínio.RepositórioAtivo{
					&RepositórioAtivoMockOk{},
					&RepositórioAtivoMockFail{},
				},
			},
			args:    args{"TEST3", d1},
			want:    &_ativo1,
			wantErr: false,
		},
		{
			name: "deveria funcionar com um repo ruim e um bom",
			fields: fields{
				repos: []domínio.RepositórioAtivo{
					&RepositórioAtivoMockFail{},
					&RepositórioAtivoMockOk{},
				},
			},
			args:    args{"TEST3", d1},
			want:    &_ativo1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ativo{
				repos: tt.fields.repos,
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
