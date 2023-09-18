// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dude333/rapinav2/pkg/progress"
)

var flags = struct {
	debug     bool
	trace     bool
	cfgFile   string
	atualizar flagsAtualizar
	relatorio flagsRelatorio
	dataSrc   string // banco de dados sqlite (ex.: "file:/var/local/rapina.db")
	tempDir   string // arquivos temporários
}{}

const (
	configFileName = "rapina.yaml"
	dataSrcDefault = ".dados" + string(os.PathSeparator) + "rapina.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
	tempDirDefault = ".dados" + string(os.PathSeparator)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rapinav2",
	Short: "Dados de empresas e fundos imobiliários",
	Long: `Este programa coleta dados financeiros de empresas e fundos imobiliários
(FIIs) da B3 e CVM e os armazena num banco de dados local.`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	// Run: func(cmd *cobra.Command, args []string) {
	// 	progress.SetDebug(flags.debug)
	// 	progress.SetTrace(flags.trace)
	// },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		progress.Error(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&flags.cfgFile, "config", "", `arquivo de configuração (default = ./`+configFileName+`)`)
	rootCmd.PersistentFlags().BoolVarP(&flags.debug, "debug", "g", false, "Mostrar logs de depuração")
	rootCmd.PersistentFlags().BoolVarP(&flags.trace, "trace", "t", false, "Mostrar logs de rastreamento")

	str := `Uso:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Alternativas:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Exemplos:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Comandos Disponíveis:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Flags Globais:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Tópicos de ajuda opcionais:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [comando] --help" para mais informações sobre um comando.{{end}}
`
	rootCmd.SetUsageTemplate(str)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	progress.SetDebug(flags.debug)
	progress.SetTrace(flags.trace)

	if flags.cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(flags.cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName(configFileName)
	}

	viper.SetConfigType("yaml")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		progress.Debug("Arquivo de configuração: %v", viper.ConfigFileUsed())
	}

	flags.dataSrc = dataSrcDefault
	if viper.IsSet("dataSrc") {
		flags.dataSrc = viper.GetString("dataSrc")
	}
	progress.Debug("dataSrc = %s", flags.dataSrc)
	if err := createDir(flags.dataSrc); err != nil {
		progress.Fatal(err)
	}

	flags.tempDir = tempDirDefault
	if viper.IsSet("tempDir") {
		flags.tempDir = viper.GetString("tempDir")
	}
	progress.Debug("tempDir = %s", flags.tempDir)
	if err := createDir(flags.tempDir); err != nil {
		progress.Fatal(err)
	}

	if viper.IsSet("reportDir") {
		flags.relatorio.outputDir = viper.GetString("reportDir")
	}
	progress.Debug("reportDir = %s", flags.relatorio.outputDir)
	if err := createDir(flags.relatorio.outputDir); err != nil {
		progress.Fatal(err)
	}

	fmt.Printf("\n\n")
}

var _db *sqlx.DB

func db() *sqlx.DB {
	if _db != nil {
		return _db // abre o banco de dados apenas umas vez
	}
	var err error
	_db, err := sqlx.Open("sqlite3", flags.dataSrc)
	if err != nil {
		progress.FatalMsg("Erro ao abrir/criar o banco de dados, verificar se o diretório existe: %s", flags.dataSrc)
	}
	_db.SetMaxOpenConns(1)

	return _db
}

func createDir(filePath string) error {
	dirPath := filepath.Dir(filePath)
	progress.Debug("dirPath: %s", dirPath)
	return os.MkdirAll(dirPath, os.ModePerm)
}
