// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

//go:build final

package frontend

import "embed"

//go:embed public/*
var ContentFS embed.FS
