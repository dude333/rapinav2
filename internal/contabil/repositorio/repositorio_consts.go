// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

import (
	"errors"
	"fmt"
)

var (
	ErrArquivoInválido = errors.New("arquivo inválido")
	ErrDFPInválida     = errors.New("DFP inválida")
	ErrSemDados        = errors.New("Sem dados")

	ErrAnoInválidoFn = func(ano int) error { return fmt.Errorf("ano inválido: %d", ano) }
)
