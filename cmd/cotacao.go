/*
Copyright © 2024 Adriano Prado <prado@gmx.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"github.com/dude333/rapinav2/pkg/cotacao"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/spf13/cobra"
)

type flagsCotação struct{}

// cotacaoCmd represents the cotacao command
var cotacaoCmd = &cobra.Command{
	Use:   "cotacao",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: cotação,
}

func init() {
	rootCmd.AddCommand(cotacaoCmd)
}

func cotação(_ *cobra.Command, args []string) {
	c, err := cotacao.NovoServiço(db(), flags.tempDir)
	if err != nil {
		progress.Fatal(err)
	}
	ativos, err := c.Cotação(args[0], args[1])
	if err != nil {
		progress.Error(err)
		return
	}
	for _, ativo := range ativos {
		progress.Status("%+v", ativo)
	}
}
