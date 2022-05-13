// SPDX-FileCopyrightText: 2022 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package csv

import (
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
)

func ImprimirCSV(db *sqlx.DB) {
	progress.Status("Imprimindo csv")
}
