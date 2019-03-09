package cmd

import (
	"log"

	"github.com/chavacava/dfence/internal/infra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cmdDeps = &cobra.Command{
	Use:   "deps [command]",
	Short: "Executes dependencies-related commands",
	Long:  "Executes dependencies-related commands",
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(infra.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger. Please use a subcommand") // revive:disable-line:deep-exit
		}

		logger.Infof("Please use a subcommand")
	},
}

func init() {
	rootCmd.AddCommand(cmdDeps)
}
