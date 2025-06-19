/*
Copyright © 2024 Timur Kulakov totusfloreodev@proton.me
*/
package cmd

import (
	"github.com/Totus-Floreo/h3-go-polyfill-extension/internal/app"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
)

var (
	targetTable  string
	outlineTable string
	dbConnStr    string
	configPath   string
)

// Config структура для чтения параметров из YAML-конфига
type Config struct {
	Target  string `yaml:"target"`
	Outline string `yaml:"outline"`
	DBConn  string `yaml:"db_conn"`
}

// readConfig читает конфигурацию из файла
func readConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// rootCmd представляет собой базовую команду, когда она вызывается без каких-либо подкоманд
var rootCmd = &cobra.Command{
	Use:   "h3-go-polyfill-extension",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Раскомментируйте следующую строку, если ваше основное приложение
	// имеет действие, связанное с ним:
	Run: func(cmd *cobra.Command, args []string) {
		if configPath != "" {
			cfg, err := readConfig(configPath)
			if err != nil {
				cmd.PrintErrln("Ошибка чтения конфига:", err)
				os.Exit(1)
			}
			if targetTable == "h3" && cfg.Target != "" {
				targetTable = cfg.Target
			}
			if outlineTable == "outline" && cfg.Outline != "" {
				outlineTable = cfg.Outline
			}
			if dbConnStr == "postgres://postgres:qwerty@localhost:5432/postgres" && cfg.DBConn != "" {
				dbConnStr = cfg.DBConn
			}
		}
		app.Polyfill(targetTable, outlineTable, dbConnStr)
	},
}

// Execute добавляет все дочерние команды к корневой команде и устанавливает флаги соответствующим образом.
// Это вызывается main.main(). Это нужно сделать только один раз для rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Здесь вы определите свои флаги и настройки конфигурации.
	// Cobra поддерживает постоянные флаги, которые, если они определены здесь,
	// будут глобальными для вашего приложения.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.h3-go-polyfill-extension.yaml)")

	// Cobra также поддерживает локальные флаги, которые будут выполняться только
	// когда это действие вызывается непосредственно.
	rootCmd.Flags().StringVarP(&targetTable, "target", "d", "h3", "Target table for insert")
	rootCmd.Flags().StringVarP(&outlineTable, "outline", "o", "outline", "Outline table name")
	rootCmd.Flags().StringVarP(&dbConnStr, "conn", "c", "postgres://postgres:qwerty@localhost:5432/postgres", "Database connection string")
	rootCmd.Flags().StringVarP(&configPath, "config", "f", "", "Путь к yaml конфигу (опционально)")
}
