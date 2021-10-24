// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio_test

import (
	"context"
	"os"
	"testing"
	"time"

	domínio "github.com/dude333/rapinav2/internal/cotacao/dominio"
	repositório "github.com/dude333/rapinav2/internal/cotacao/repositorio"
	"github.com/dude333/rapinav2/pkg/progress"
)

func Test_b3_Importar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	b3 := repositório.B3(os.TempDir())

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	dia, _ := domínio.NovaData("2021-10-08")
	count := 10
	for result := range b3.Importar(ctx, dia) {
		if result.Error != nil {
			t.Logf(result.Error.Error())
			return
		}
		progress.Status("%v", result.Ativo)
		time.Sleep(10 * time.Millisecond)
		count--
		if count < 0 {
			cancel()
			progress.RunFail()
			break
		}
	}
}
