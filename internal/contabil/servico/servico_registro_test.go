// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package serviço_test

import (
	"context"
	"fmt"
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
	serviço "github.com/dude333/rapinav2/internal/contabil/servico"
	"strconv"
	"testing"
	"time"
)

var (
	_cache    map[uint32]*contábil.Registro
	_exemplos = []*contábil.Registro{}
)

func init() {
	_cache = make(map[uint32]*contábil.Registro)

	for i := 1; i <= 10; i++ {
		r := contábil.Registro{
			CNPJ:         fmt.Sprintf("%010d", i),
			Empresa:      fmt.Sprintf("Empresa %02d", i),
			Ano:          2021,
			DataFimExerc: "2021-12-31",
			Versão:       1,
			Total: contábil.Dinheiro{
				Valor:  float64(i),
				Escala: 1000,
				Moeda:  "R$",
			},
		}
		_exemplos = append(_exemplos, &r)
	}
}

// Implementação de repositórios de teste ---

type repoBD struct{}

func (r *repoBD) Ler(ctx context.Context, cnpj string, ano int) (*contábil.Registro, error) {
	x := fmt.Sprintf("%s%d", cnpj, ano)
	y, _ := strconv.Atoi(x)
	return _cache[uint32(y)], nil
}

func (r *repoBD) Salvar(ctx context.Context, e *contábil.Registro) error {
	x := fmt.Sprintf("%s%d", e.CNPJ, e.Ano)
	y, _ := strconv.Atoi(x)
	_cache[uint32(y)] = e

	return nil
}

type repoAPI struct{}

func (r *repoAPI) Importar(ctx context.Context, ano int) <-chan contábil.ResultadoImportação {
	results := make(chan contábil.ResultadoImportação)
	go func() {
		defer close(results)

		for _, r := range _exemplos {
			result := contábil.ResultadoImportação{
				Error:    nil,
				Registro: r,
			}
			select {
			case <-ctx.Done():
				return
			case results <- result:
				time.Sleep(1 * time.Millisecond)
			}
		}

	}()
	return results

}

// Testes ---

func Test_registro_Importar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	reg := serviço.NovoRegistro(&repoAPI{}, &repoBD{})

	err := reg.Importar(2021)
	if err != nil {
		t.Fatalf("ServiçoRegistro.Importar(): %v", err)
	}

	for _, r := range _exemplos {
		x := fmt.Sprintf("%s%d", r.CNPJ, r.Ano)
		y, _ := strconv.Atoi(x)

		c := _cache[uint32(y)]
		if c.Empresa != r.Empresa {
			t.Fatalf("Valor salvo esperado: %v, recebido: %v", r.Empresa, c.Empresa)
		}
	}
}
