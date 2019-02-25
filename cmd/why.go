package cmd

import (
	"log"
	"strings"
	"sync"

	"github.com/spf13/viper"

	"github.com/KyleBanks/depth"
	dfence "github.com/chavacava/dfence/internal"
	"github.com/spf13/cobra"
)

var cmdWhy = &cobra.Command{
	Use:   "why [package] [package]",
	Short: "explains why a package depends on other",
	Long:  "explains why a package depends on other",
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
		writeExplain(logger, *t.Root, []string{}, pkgTarget)
	},
}

func init() {
	rootCmd.AddCommand(cmdWhy)
}

func writeExplain(logger dfence.Logger, pkg depth.Pkg, stack []string, explain string) {
	stack = append(stack, pkg.Name)

	if pkg.Name == explain {
		logger.Infof(strings.Join(stack, " -> "))
	}

	var wg sync.WaitGroup
	for _, p := range pkg.Deps {
		wg.Add(1)
		go func(pkg depth.Pkg) {
			writeExplain(logger, pkg, stack, explain)
			wg.Done()
		}(p)
	}

	wg.Wait()
}
