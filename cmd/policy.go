package cmd

import (
	"log"

	dfence "github.com/chavacava/dfence/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cmdPolicy = &cobra.Command{
	Use:   "policy [command]",
	Short: "executes policy-related commands",
	Long:  "executes policy-related commands",
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(dfence.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger. Please use a subcommand") // revive:disable-line:deep-exit
		}

		logger.Infof("Please use a subcommand")
	},
}

func init() {
	rootCmd.AddCommand(cmdPolicy)
}
