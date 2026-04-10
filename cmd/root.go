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
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dude333/rapinav2/pkg/progress"
)

var flags = struct {
	cfgFile   string
	dataSrc   string // banco de dados sqlite (ex.: "file:/var/local/rapina.db")
	tempDir   string // arquivos temporários
	servidor  flagsServidor
	relatorio flagsRelatorio
	atualizar flagsAtualizar
	cotação   flagsCotação
	debug     bool
	trace     bool
}{}

const (
	sep                    = string(os.PathSeparator)
	configFileName         = "rapina.yaml"
	dataSrcDefault         = ".dados" + sep + "rapina.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
	tempDirDefault         = ".dados" + sep + "temp"
	reportDirDefault       = ".dados" + sep + "reports"
	googleSheetsDirDefault = "/rapina"
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

	// Setup viper configuration
	if flags.cfgFile != "" {
		viper.SetConfigFile(flags.cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName(configFileName)
	}

	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			progress.Warning("Erro ao ler arquivo de configuração: %v", err)
		}
	} else {
		progress.Debug("Arquivo de configuração: %v", viper.ConfigFileUsed())
	}

	// Initialize all config directories
	initConfigPath(&flags.dataSrc, "dataSrc", dataSrcDefault)
	initConfigPath(&flags.tempDir, "tempDir", tempDirDefault)
	initConfigPath(&flags.relatorio.outputDir, "relatorio.outputDir", reportDirDefault)

	// Initialize flags from config file
	initFlag(&flags.relatorio.googlesheetsDir, "relatorio.googleSheetsDir")
	initFlag(&flags.relatorio.crescente, "relatorio.crescente")
	initFlag(&flags.relatorio.tokenport, "relatorio.tokenport")
	initFlag(&flags.relatorio.oauthurl, "relatorio.oauthurl")

	fmt.Printf("\n\n")
}

// initConfigPath loads a config path from viper with a default value, then creates the directory.
func initConfigPath(ptr *string, key, defaultValue string) {
	*ptr = defaultValue
	if viper.IsSet(key) {
		*ptr = viper.GetString(key)
	}
	progress.Debug("initConfigPath: %s = %s", key, *ptr)
	if err := createDir(*ptr); err != nil {
		progress.Fatal(err)
	}
}

// initFlag loads a config value from viper into the provided pointer if the key is set.
func initFlag[T any](ptr *T, key string) {
	if !viper.IsSet(key) {
		return
	}
	if viper.IsSet(key) {
		switch v := any(ptr).(type) {
		case *string:
			*v = viper.GetString(key)
		case *int:
			*v = viper.GetInt(key)
		case *bool:
			*v = viper.GetBool(key)
		default:
			progress.FatalMsg("Tipo de flag não suportado: %s", key)
		}
		progress.Debug("initFlag: %s = %v", key, *ptr)
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
		progress.FatalMsg("Erro ao abrir/criar o banco de dados, verificar se o diretório existe: %s", flags.dataSrc)
	}
	_db.SetMaxOpenConns(1)

	return _db
}

func createDir(filePath string) error {
	dirPath := filepath.Dir(filePath)
	return os.MkdirAll(dirPath, os.ModePerm)
}
