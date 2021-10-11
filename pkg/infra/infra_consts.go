// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package infra

import (
	"errors"
	"time"
)

const (
	_http_timeout = 30 * time.Second
)

var (
	ErrFileNotFound = errors.New("arquivo não encontrado")
)
