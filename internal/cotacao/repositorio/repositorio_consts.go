// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositório

import (
	"errors"
	"fmt"
)

var (
	ErrFalhaDownload      = errors.New("falha no download")
	ErrAtivoNãoEncontrado = errors.New("ativo não encontrado")

	ErrDataInválida = func(dia string) error { return fmt.Errorf("data com formato inválido: %s", dia) }
)
