// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package serviço

import (
	"context"
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
)

type registro struct {
	api contábil.RepositórioImportaçãoRegistro
	bd  contábil.RepositórioLeituraEscritaRegistro
}

func Registro() contábil.ServiçoRegistro {
	return &registro{}
}

// Importar importa os relatórios contábeis no ano especificado e os salva
// no banco de dados.
func (r *registro) Importar(ano int) error {
	if r.api == nil {
		return ErrRepositórioInválido
	}

	err := r.api.Importar(context.Background(), ano)

	return err
}

func (r *registro) Relatório(cnpj string, ano int) (*contábil.Registro, error) {
	if r.bd == nil {
		return &contábil.Registro{}, ErrRepositórioInválido
	}
	registro, err := r.bd.Ler(context.Background(), cnpj, ano)
	return registro, err
}
