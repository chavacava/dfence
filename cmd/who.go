package cmd

import (
	"log"

	"github.com/spf13/viper"

	"github.com/KyleBanks/depth"
	dfence "github.com/chavacava/dfence/internal"
	"github.com/spf13/cobra"
)

var graph bool

var cmdWho = &cobra.Command{
	Use:   "who [package] [package]",
	Short: "explains who depends on a package",
	Long:  "explains who depends on a package",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(dfence.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger.") // revive:disable-line:deep-exit
		}

		pkgTarget := args[0]
		pkgSelector := args[1]
		pkgs, err := retrievePackages(pkgSelector)
		if err != nil {
			logger.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
		}
		for _, p := range pkgs {
			var t depth.Tree
			err := t.Resolve(p)
			if err != nil {
				logger.Warningf("Unable to analyze package '%s': %v", p, err)
			}

			writeExplain(logger, *t.Root, []string{}, pkgTarget)
		}
	},
}

func init() {
	rootCmd.AddCommand(cmdWho)
	cmdWho.Flags().BoolVar(&graph, "graph", false, "generate a graph")
}
