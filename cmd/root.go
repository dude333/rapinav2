// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var flags = struct {
	verbose   bool
	atualizar flagsAtualizar
	dataSrc   string // banco de dados sqlite (ex.: "file:/var/local/rapina.db")
	tempDir   string // arquivos temporários
}{}

const (
	configFileName = "rapina.yaml"
	dataSrcDefault = ".dados" + string(os.PathSeparator) + "rapina.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
	tempDirDefault = ".dados" + string(os.PathSeparator)
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rapinav2",
	Short: "Dados de empresas e fundos imobiliários",
	Long: `Este programa coleta dados financeiros de empresas e fundos imobiliários
(FIIs) da B3 e CVM e os armazena num banco de dados local.`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", `arquivo de configuração (default = ./`+configFileName+`)`)

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
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
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
		progress.Error(err)
		os.Exit(1)
	}

	flags.tempDir = tempDirDefault
	if viper.IsSet("tempDir") {
		flags.tempDir = viper.GetString("tempDir")
	}
	progress.Debug("tempDir = %s", flags.tempDir)
	if err := createDir(flags.tempDir); err != nil {
		progress.Error(err)
		os.Exit(1)
	}
}

var _db *sqlx.DB

func db() *sqlx.DB {
	if _db != nil {
		return _db // abre o banco de dados apenas umas vez
	}
	var err error
	_db, err := sqlx.Open("sqlite3", flags.dataSrc)
	if err != nil {
		progress.ErrorMsg("Erro ao abrir/criar o banco de dados, verificar se o diretório existe: %s", flags.dataSrc)
		os.Exit(1)
	}
	_db.SetMaxOpenConns(1)

	return _db
}

func createDir(filePath string) error {
	dirPath := filepath.Dir(filePath)
	progress.Debug("dirPath: %s", dirPath)
	return os.MkdirAll(dirPath, os.ModePerm)
}
