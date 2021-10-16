// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositório_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	domínio "github.com/dude333/rapinav2/internal/cotacao/dominio"
	repositório "github.com/dude333/rapinav2/internal/cotacao/repositorio"
)

func Test_b3_Importar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	b3 := repositório.B3("/tmp")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	dia, _ := domínio.NovaData("2021-10-08")
	for result := range b3.Importar(ctx, dia) {
		if result.Error != nil {
			t.Logf(result.Error.Error())
			return
		}
		fmt.Printf("=> %+v\n", result.Ativo)
	}
}
