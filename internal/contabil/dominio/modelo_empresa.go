// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package domínio

import "context"

type Hash uint32

type Empresa struct {
	CNPJ   string
	Ano    int
	Contas []Conta
}

type Conta map[Hash]Dinheiro

type Dinheiro struct {
	Valor  float64
	Escala int
	Moeda  string
}

type RepositórioImportaçãoEmpresa interface {
	Importar(ctx context.Context, ano int) error
	RepositórioEscritaEmpresa
}

type RepositórioLeituraEmpresa interface {
	Ler(ctx context.Context, cnpj string, ano int) (*Empresa, error)
}

type RepositórioEscritaEmpresa interface {
	Salvar(ctx context.Context, empresa *Empresa) error
}

type RepositórioLeituraEscritaEmpresa interface {
	RepositórioLeituraEmpresa
	RepositórioEscritaEmpresa
}

type ServiçoEmpresa interface {
	Importar(ano int) error
	Relatório(cnpj string, ano int) (*Empresa, error)
}
