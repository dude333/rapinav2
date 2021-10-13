// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package serviço

import (
	"context"
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
	"time"
)

type registro struct {
	api contábil.RepositórioImportaçãoRegistro
	bd  contábil.RepositórioLeituraEscritaRegistro
}

func NovoRegistro(
	api contábil.RepositórioImportaçãoRegistro,
	bd contábil.RepositórioLeituraEscritaRegistro) contábil.ServiçoRegistro {
	return &registro{api: api, bd: bd}
}

// Importar importa os relatórios contábeis no ano especificado e os salva
// no banco de dados.
func (r *registro) Importar(ano int) error {
	if r.api == nil {
		return ErrRepositórioInválido
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// result retorna o registro após a leitura de cada linha
	// do arquivo importado
	for result := range r.api.Importar(ctx, ano) {
		if result.Error != nil {
			return result.Error
		}
		err := r.bd.Salvar(context.Background(), result.Registro)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *registro) Relatório(cnpj string, ano int) (*contábil.Registro, error) {
	if r.bd == nil {
		return &contábil.Registro{}, ErrRepositórioInválido
	}
	registro, err := r.bd.Ler(context.Background(), cnpj, ano)
	return registro, err
}
