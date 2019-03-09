package cmd

import (
	"log"

	"github.com/spf13/viper"

	"github.com/KyleBanks/depth"
	dfence "github.com/chavacava/dfence/internal"
	"github.com/spf13/cobra"
)

var cmdWhy = &cobra.Command{
	Use:   "why [package] [package]",
	Short: "explains why a package depends on the other",
	Long:  "explains why a package depends on the other",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(dfence.Logger)
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
func explainDep(from depth.Pkg, to string) []dfence.DepChain {
	explanations := []dfence.DepChain{}

	recExplainDep(from, to, dfence.NewDepChain(), &explanations)
	return explanations
}

func recExplainDep(pkg depth.Pkg, explain string, chain dfence.DepChain, explanations *[]dfence.DepChain) {
	chain.Append(dfence.NewRawChainItem(pkg.Name))

	if pkg.Name == explain {
		*explanations = append(*explanations, chain.Clone())
		return
	}

	for _, pkg := range pkg.Deps {
		recExplainDep(pkg, explain, chain, explanations)
	}
}
