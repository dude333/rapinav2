// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	rapina "github.com/dude333/rapinav2"
	contábil "github.com/dude333/rapinav2/pkg/contabil"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func Test_inserirDFP(t *testing.T) {
	var db *sqlx.DB
	if testing.Short() {
		db = sqlx.MustConnect("sqlite3", ":memory:")
		db.SetMaxOpenConns(1)
	} else {
		connStr := "file:/tmp/rapina.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
		db = sqlx.MustConnect("sqlite3", connStr)
	}

	s, err := NovoSqlite(db)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("criar e inserir dados", func(t *testing.T) {

		var contas []contábil.Conta
		for n := 1; n <= 10; n++ {
			c := contábil.Conta{
				Código:       fmt.Sprintf("C%03d", n),
				Descr:        fmt.Sprintf("D%03d", n),
				Grupo:        fmt.Sprintf("G%03d", n),
				DataFimExerc: fmt.Sprintf("D%03d", n),
				OrdemExerc:   fmt.Sprintf("O%03d", n),
				Total: rapina.Dinheiro{
					Valor:  1234567.89,
					Escala: 1000,
					Moeda:  "R$",
				},
			}
			contas = append(contas, c)
		}

		dfp := contábil.DemonstraçãoFinanceira{
			Empresa: rapina.Empresa{
				CNPJ: "123",
				Nome: "N1",
			},
			Ano:    2020,
			Contas: contas,
		}

		err := s.Salvar(context.Background(), &dfp)
		if err != nil {
			t.Logf("%v", err)
		}
	})
}

func Test_ordenar(t *testing.T) {
	type args struct {
		orig   []string
		transf []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "deveria funcionar",
			args: args{
				orig:   []string{"AAAaA", "BbB", "AaaaA", "AA", "bônus", "açaí", "aliás", "caçapa"},
				transf: []string{"aaaaa", "bbb", "aaaaa", "aa", "bonus", "acai", "alias", "cacapa"},
			},
			want: []string{"AA", "AAAaA", "AaaaA", "açaí", "aliás", "BbB", "bônus", "caçapa"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ordenar(tt.args.orig, tt.args.transf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ordenar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSqlite_Empresas(t *testing.T) {
	type fields struct {
		db    *sqlx.DB
		limpo map[string]bool
		cache []string
		cfg   cfg
	}
	type args struct {
		ctx  context.Context
		nome string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "deveria funcionar",
			fields: fields{
				db:    &sqlx.DB{}, // vai funcionar desde que o db só seja lido caso cache esteja vazio
				limpo: map[string]bool{},
				cache: []string{"Ótimo", "zinco", "Base", "Zircônio", "azul", "capaz", "Exceção", "óculos", "Também"},
				cfg:   cfg{},
			},
			args: args{
				ctx:  nil,
				nome: "",
			},
			want: []string{"azul", "Base", "capaz", "Exceção", "óculos", "Ótimo", "Também", "zinco", "Zircônio"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sqlite{
				db:    tt.fields.db,
				limpo: tt.fields.limpo,
				cache: tt.fields.cache,
				cfg:   tt.fields.cfg,
			}
			if got := s.Empresas(tt.args.ctx, tt.args.nome); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sqlite.Empresas() = %#v, want %v", got, tt.want)
			}
		})
	}
}
