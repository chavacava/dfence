package cmd

import (
	"fmt"
	"os"

	dfence "github.com/chavacava/dfence/internal"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logLevel string

var rootCmd = &cobra.Command{
	Use:   "dfence",
	Short: "Dependency fences",
	Long: `
         ________                   
    ____/ / ____/__  ____  ________ 
   / __  / /_  / _ \/ __ \/ ___/ _ \
  / /_/ / __/ /  __/ / / / /__/  __/
  \__,_/_/    \___/_/ /_/\___/\___/ 
																		 
	
  dFence helps you tame your dependencies`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		viper.Set("logger", buildlogger(logLevel))
	},
}

// Execute executes this command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
	rootCmd.PersistentFlags().StringVar(&logLevel, "log", "info", "log level: none, error, warn, info, debug")
}

func buildlogger(level string) dfence.Logger {
	nop := func(string, ...interface{}) {}
	debug, info, warn, err := nop, nop, nop, nop
	switch level {
	case "none":
		// do nothing
	case "debug":
		debug = buildLoggerFunc("[DEBUG] ", color.New(color.FgCyan))
		fallthrough
	case "info":
		info = buildLoggerFunc("", color.New(color.FgGreen))
		fallthrough
	case "warn":
		warn = buildLoggerFunc("", color.New(color.FgHiYellow))
		fallthrough
	default:
		err = buildLoggerFunc("", color.New(color.BgHiRed))
	}

	fatal := buildLoggerFunc("", color.New(color.BgRed))
	return dfence.NewLogger(debug, info, warn, err, fatal)
}

func buildLoggerFunc(prefix string, c *color.Color) dfence.LoggerFunc {
	return func(msg string, vars ...interface{}) {
		fmt.Println(c.Sprintf(prefix+msg, vars...))
	}
}
