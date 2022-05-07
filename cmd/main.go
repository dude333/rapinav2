// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/dude333/rapinav2/pkg/progress"
)

func main() {
	// defer profile.Start().Stop()
	progress.SetDebug(true)
	Execute()
}
