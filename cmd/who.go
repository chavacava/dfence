package cmd

import (
	"log"

	"github.com/spf13/viper"

	"github.com/KyleBanks/depth"
	dfence "github.com/chavacava/dfence/internal"
	"github.com/spf13/cobra"
)

var cmdWho = &cobra.Command{
	Use:   "who [package] [package selector]",
	Short: "explains what packages, from a package list, depend on a package",
	Long:  "explains what packages, from a package list, depend on a package",
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

		for _, pkg := range pkgs {
			var t depth.Tree
			err := t.Resolve(pkg)
			if err != nil {
				logger.Warningf("Unable to analyze package '%s': %v", pkg, err)
			}

			explanations := []string{}
			explainDep(*t.Root, pkgTarget, []string{}, &explanations)

			for _, e := range explanations {
				logger.Infof(e)
			}
		}
	},
}

func init() {
	cmdDeps.AddCommand(cmdWho)
}
