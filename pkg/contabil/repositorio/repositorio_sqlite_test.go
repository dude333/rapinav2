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
	"github.com/dude333/rapinav2/pkg/contabil"
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

		var contas []contabil.Conta
		for n := 1; n <= 10; n++ {
			c := contabil.Conta{
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

		dfp := contabil.DemonstraçãoFinanceira{
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
		orig   []rapina.Empresa
		transf []string
	}
	tests := []struct {
		name string
		args args
		want []rapina.Empresa
	}{
		{
			name: "deveria funcionar",
			args: args{
				orig: []rapina.Empresa{
					{CNPJ: "1", Nome: "AAAaA"},
					{CNPJ: "2", Nome: "BbB"},
					{CNPJ: "3", Nome: "AaaaA"},
					{CNPJ: "4", Nome: "AA"},
					{CNPJ: "5", Nome: "bônus"},
					{CNPJ: "6", Nome: "açaí"},
					{CNPJ: "7", Nome: "aliás"},
					{CNPJ: "8", Nome: "caçapa"},
				},
				transf: []string{"aaaaa", "bbb", "aaaaa", "aa", "bonus", "acai", "alias", "cacapa"},
			},
			want: []rapina.Empresa{
				{CNPJ: "4", Nome: "AA"},
				{CNPJ: "1", Nome: "AAAaA"},
				{CNPJ: "3", Nome: "AaaaA"},
				{CNPJ: "6", Nome: "açaí"},
				{CNPJ: "7", Nome: "aliás"},
				{CNPJ: "2", Nome: "BbB"},
				{CNPJ: "5", Nome: "bônus"},
				{CNPJ: "8", Nome: "caçapa"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ordenar(tt.args.orig, tt.args.transf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ordenar() = %#v, want %v", got, tt.want)
			}
		})
	}
}

func TestSqlite_Empresas(t *testing.T) {
	type fields struct {
		db    *sqlx.DB
		limpo map[string]bool
		cache []rapina.Empresa
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
		want   []rapina.Empresa
	}{
		{
			name: "deveria funcionar",
			fields: fields{
				db:    &sqlx.DB{}, // vai funcionar desde que o db só seja lido caso cache esteja vazio
				limpo: map[string]bool{},
				cache: []rapina.Empresa{
					{CNPJ: "", Nome: "Ótimo"},
					{CNPJ: "", Nome: "zinco"},
					{CNPJ: "", Nome: "Base"},
					{CNPJ: "", Nome: "Zircônio"},
					{CNPJ: "", Nome: "azul"},
					{CNPJ: "", Nome: "capaz"},
					{CNPJ: "", Nome: "Exceção"},
					{CNPJ: "", Nome: "óculos"},
					{CNPJ: "", Nome: "Também"},
				},
				cfg: cfg{},
			},
			args: args{
				ctx:  nil,
				nome: "",
			},
			want: []rapina.Empresa{
				{CNPJ: "", Nome: "azul"},
				{CNPJ: "", Nome: "Base"},
				{CNPJ: "", Nome: "capaz"},
				{CNPJ: "", Nome: "Exceção"},
				{CNPJ: "", Nome: "óculos"},
				{CNPJ: "", Nome: "Ótimo"},
				{CNPJ: "", Nome: "Também"},
				{CNPJ: "", Nome: "zinco"},
				{CNPJ: "", Nome: "Zircônio"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sqlite{
				db:            tt.fields.db,
				limpo:         tt.fields.limpo,
				cacheEmpresas: tt.fields.cache,
				cfg:           tt.fields.cfg,
			}
			if got := s.Empresas(tt.args.ctx, tt.args.nome); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sqlite.Empresas() = %#v, want %v", got, tt.want)
			}
		})
	}
}
