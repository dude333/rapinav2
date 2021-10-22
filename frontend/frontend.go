//go:build final

package frontend

import "embed"

//go:embed public/*
var ContentFS embed.FS
