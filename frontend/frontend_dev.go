// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package frontend

import (
	"io/fs"
	"log"
	"os"
)

var ContentFS fs.FS

func init() {
	var err error
	ContentFS = os.DirFS("./frontend/")
	if err != nil {
		log.Fatal(err)
	}
}
