package cmd

import (
	"log"
	"strings"

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

		explanations := []string{}
		explainDep(*t.Root, pkgTarget, []string{}, &explanations)

		for _, e := range explanations {
			logger.Infof(e)
		}
	},
}

func init() {
	cmdDeps.AddCommand(cmdWhy)
}

func explainDep(pkg depth.Pkg, explain string, stack []string, explanations *[]string) {
	stack = append(stack, pkg.Name)

	if pkg.Name == explain {
		*explanations = append(*explanations, strings.Join(stack, " -> "))
		return
	}

	for _, pkg := range pkg.Deps {
		explainDep(pkg, explain, stack, explanations)
	}
}
