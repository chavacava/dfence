package cmd

import (
	"log"

	dfence "github.com/chavacava/dfence/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cmdDeps = &cobra.Command{
	Use:   "deps [command]",
	Short: "executes dependencies related commands",
	Long:  "executes dependencies related commands",
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(dfence.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger. Please use a subcommand") // revive:disable-line:deep-exit
		}

		logger.Infof("Please use a subcommand")
	},
}

func init() {
	rootCmd.AddCommand(cmdDeps)
}
