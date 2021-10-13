// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package domínio

import (
	"context"
	"fmt"
)

type Hash uint32

type Registro struct {
	CNPJ         string
	Empresa      string
	Ano          int
	DataFimExerc string // AAAA-MM-DD
	Versão       int
	Total        Dinheiro
}

type Dinheiro struct {
	Valor  float64
	Escala int
	Moeda  string
}

func (d Dinheiro) String() string {
	return fmt.Sprintf(`%s %.2f`, d.Moeda, d.Valor*float64(d.Escala))
}

type ResultadoImportação struct {
	Registro *Registro
	Error    error
}

type RepositórioImportaçãoRegistro interface {
	Importar(ctx context.Context, ano int) <-chan ResultadoImportação
}

type RepositórioLeituraRegistro interface {
	Ler(ctx context.Context, cnpj string, ano int) (*Registro, error)
}

type RepositórioEscritaRegistro interface {
	Salvar(ctx context.Context, empresa *Registro) error
}

type RepositórioLeituraEscritaRegistro interface {
	RepositórioLeituraRegistro
	RepositórioEscritaRegistro
}

type ServiçoRegistro interface {
	Importar(ano int) error
	Relatório(cnpj string, ano int) (*Registro, error)
}
