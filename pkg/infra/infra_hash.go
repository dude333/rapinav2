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

func FileHash(filename string) (h string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	fnvHash.Reset()
	if _, err = io.Copy(fnvHash, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", fnvHash.Sum64()), nil
}

func Hash(data *[]byte) string {
	fnvHash.Reset()
	fnvHash.Write([]byte(*data))
	return fmt.Sprintf("%x", fnvHash.Sum64())
}
