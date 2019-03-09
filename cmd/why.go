package cmd

import (
	"log"

	"github.com/KyleBanks/depth"
	"github.com/chavacava/dfence/internal/deps"
	"github.com/chavacava/dfence/internal/infra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cmdWhy = &cobra.Command{
	Use:   "why [package] [package]",
	Short: "explains why a package depends on the other",
	Long:  "explains why a package depends on the other",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(infra.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger.") // revive:disable-line:deep-exit
		}

		pkgSource := args[0]
		pkgTarget := args[1]
		var t depth.Tree
		err := t.Resolve(pkgSource)
		if err != nil {
			logger.Errorf("Unable to analyze package '%s': %v", pkgSource, err.Error())
		}

		explanations := explainDep(*t.Root, pkgTarget)

		for _, e := range explanations {
			logger.Infof(e.String())
		}
	},
}

func init() {
	cmdDeps.AddCommand(cmdWhy)
}

// explainDep yields a list of dependency chains going from -> ... -> to
func explainDep(from depth.Pkg, to string) []deps.DepChain {
	explanations := []deps.DepChain{}

	recExplainDep(from, to, deps.NewDepChain(), &explanations)
	return explanations
}

func recExplainDep(pkg depth.Pkg, explain string, chain deps.DepChain, explanations *[]deps.DepChain) {
	chain.Append(deps.NewRawChainItem(pkg.Name))

	if pkg.Name == explain {
		*explanations = append(*explanations, chain.Clone())
		return
	}

	for _, pkg := range pkg.Deps {
		recExplainDep(pkg, explain, chain, explanations)
	}
}
