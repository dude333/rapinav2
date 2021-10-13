// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package serviço

import (
	"context"
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
)

type empresa struct {
	api contábil.RepositórioImportaçãoEmpresa
	bd  contábil.RepositórioLeituraEscritaEmpresa
}

func Empresa() contábil.ServiçoEmpresa {
	return &empresa{}
}

// Importar importa os relatórios contábeis no ano especificado e os salva
// no banco de dados.
func (e *empresa) Importar(ano int) error {
	if e.api == nil {
		return ErrRepositórioInválido
	}

	err := e.api.Importar(context.Background(), ano)

	return err
}

func (e *empresa) Relatório(cnpj string, ano int) (*contábil.Empresa, error) {
	if e.bd == nil {
		return &contábil.Empresa{}, ErrRepositórioInválido
	}
	empresa, err := e.bd.Ler(context.Background(), cnpj, ano)
	return empresa, err
}
