// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositório

import (
	"context"
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
)

type cvm struct{}

func NovoCVM() contábil.RepositórioImportaçãoDFP {
	return &cvm{}
}

func (r *cvm) Importar(ctx context.Context, ano int) <-chan contábil.ResultadoImportação {
	results := make(chan contábil.ResultadoImportação)
	go func() {
		defer close(results)

		// for _, r := range _exemplos {
		// 	result := contábil.ResultadoImportação{
		// 		Error:    nil,
		// 		Registro: r,
		// 	}
		// 	select {
		// 	case <-ctx.Done():
		// 		return
		// 	case results <- result:
		// 		time.Sleep(1 * time.Millisecond)
		// 	}
		// }

	}()
	return results
}
