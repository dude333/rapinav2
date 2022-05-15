// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/dude333/rapinav2/frontend"
	"github.com/dude333/rapinav2/pkg/contabil/apresentacao/api"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

// servidorCmd represents the servidor command
var servidorCmd = &cobra.Command{
	Use:     "servidor",
	Aliases: []string{"server", "web"},
	Short:   "Servidor web para apresetanção dos relatórios",
	Long:    `Servidor web para apresetanção dos relatórios`,
	Run:     servidor,
}

func init() {
	rootCmd.AddCommand(servidorCmd)
}

func servidor(cmd *cobra.Command, args []string) {
	e := echo.New()

	api.NewAPI(e, openDatabase(), flags.tempDir)

	contentFS, err := fs.Sub(frontend.ContentFS, "public")
	if err != nil {
		panic(err)
	}
	fs := http.FileServer(http.FS(contentFS))
	e.GET("/*", echo.WrapHandler(fs))

	log.Println("Listening on port", 3000)
	if err := e.Start(":3000"); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
