// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositório_test

import (
	"context"
	"fmt"
	"testing"

	domínio "github.com/dude333/rapinav2/internal/cotacao/dominio"
	repositório "github.com/dude333/rapinav2/internal/cotacao/repositorio"
)

type repo struct {
	contador int
}

func (r *repo) Salvar(ctx context.Context, ativo *domínio.Ativo) error {
	fmt.Printf("%6d) %+v\n", r.contador, *ativo)
	r.contador++
	return nil
}

func Test_b3_Cotação(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	b3 := repositório.B3(&repo{1}, "/tmp")

	d, _ := domínio.NovaData("2021-10-08")
	atv, err := b3.Cotação(context.Background(), "WEGE3", d)
	fmt.Println("=>", atv, err)
}
