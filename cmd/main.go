// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

// import "github.com/pkg/profile"

func main() {
	// defer profile.Start(profile.MemProfileRate(2048)).Stop()
	defer func() {
		if _db != nil {
			_db.Close()
		}
	}()

	Execute()
}
