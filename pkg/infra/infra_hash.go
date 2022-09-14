// SPDX-FileCopyrightText: 2022 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package infra

import (
	"fmt"
	"hash/fnv"
	"io"
	"os"
)

// fnvHash is a global var set to speed up Hash
var fnvHash = fnv.New64a()

func FileHash(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	fnvHash.Reset()
	if _, err = io.Copy(fnvHash, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", fnvHash.Sum64()), nil
}
