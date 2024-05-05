// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package cotacao

import rapina "github.com/dude333/rapinav2"

type Resultado struct {
	Ativo *rapina.Cotação
	Error error
}
