package cmd

import (
	"log"

	"github.com/chavacava/dfence/internal/deps"
	dependencies "github.com/chavacava/dfence/internal/deps"
	"github.com/chavacava/dfence/internal/infra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cmdWhy = &cobra.Command{
	Use:   "why [package] [package]",
	Short: "Explains why a package depends on the other",
	Long:  "Explains why a package depends on the other",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(infra.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger.") // revive:disable-line:deep-exit
		}

		pkgSource := args[0]
		pkgTarget := args[1]
		depsRoot, err := dependencies.ResolvePkgDeps(pkgSource, maxDepth)
		if err != nil {
			logger.Warningf("Unable to analyze package '%s': %v", pkgSource, err)
			return
		}

		explanations := deps.ExplainDep(depsRoot, pkgTarget)

		for _, e := range explanations {
			logger.Infof(e.String())
		}
	},
}

func init() {
	cmdDeps.AddCommand(cmdWhy)
}
