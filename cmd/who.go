package cmd

import (
	"log"
	"sync"

	"github.com/KyleBanks/depth"
	"github.com/chavacava/dfence/internal/deps"
	"github.com/chavacava/dfence/internal/infra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cmdWho = &cobra.Command{
	Use:   "who [package]",
	Short: "Explains what packages, from a package list, depend on a package",
	Long:  "Explains what packages, from a package list, depend on a package",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(infra.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger.") // revive:disable-line:deep-exit
		}

		pkgTarget := args[0]
		pkgSelector := "./..."
		pkgs, err := retrievePackages(pkgSelector)
		if err != nil {
			logger.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
		}

		const concLevel = 10 // concurrency level
		tokens := make(chan struct{}, concLevel)
		var wg sync.WaitGroup
		wg.Add(len(pkgs))

		for _, pkg := range pkgs {
			go func(pkg string) {
				tokens <- struct{}{}
				var t depth.Tree
				if maxDepth > 0 {
					t.MaxDepth = maxDepth
				}

				err := t.Resolve(pkg)
				if err != nil {
					logger.Warningf("Unable to analyze package '%s': %v", pkg, err)
				}

				explanations := deps.ExplainDep(*t.Root, pkgTarget)

				for _, e := range explanations {
					logger.Infof(e.String())
				}
				<-tokens
				wg.Done()
			}(pkg)
		}

		wg.Wait()
	},
}

func init() {
	const unlimitedDepth = 0
	cmdDeps.AddCommand(cmdWho)
	cmdWho.Flags().IntVar(&maxDepth, "maxdepth", unlimitedDepth, "max dependence distance")
}
