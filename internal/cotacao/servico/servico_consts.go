// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package serviço

import "errors"

var (
	ErrCotaçãoNãoEncontrada = errors.New("cotação não encontrada")
	ErrRepositórioInválido  = errors.New("repositório inválido")
)
