package cmd

import (
	"log"
	"strings"

	"github.com/spf13/viper"

	"github.com/KyleBanks/depth"
	dfence "github.com/chavacava/dfence/internal"
	"github.com/spf13/cobra"
)

var cmdFindCycles = &cobra.Command{
	Use:   "find-cycles [package] [package]",
	Short: "Searches for dependency cycles among the given packages",
	Long:  "Searches for dependency cycles among the given packages",
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(dfence.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger.") // revive:disable-line:deep-exit
		}

		pkgSelector := strings.Join(args, " ")
		logger.Infof("Retrieving packages...")
		pkgs, err := retrievePackages(pkgSelector)
		if err != nil {
			logger.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
		}
		logger.Infof("Will work with %d package(s).", len(pkgs))

		allDeps := map[string]*depth.Pkg{}
		for _, pkg := range pkgs {
			t := depth.Tree{
				ResolveInternal: true,
				ResolveTest:     true,
			}
			err := t.Resolve(pkg)
			if err != nil {
				logger.Warningf("Skipping package '%s', unable to analyze: %v", pkg, err.Error())
			}
			allDeps[t.Root.Name] = t.Root
		}
		cycles := [][]string{}
		for _, pkg := range pkgs {
			cycles = append(cycles, findCycles(pkg, allDeps, logger)...)
		}
		if len(cycles) == 0 {
			logger.Infof("No cycles found")
			return
		}

		for _, c := range cycles {
			logger.Errorf("Cycle: %s", strings.Join(c, " -> "))
		}
	},
}

func init() {
	rootCmd.AddCommand(cmdFindCycles)
}

func findCycles(pkg string, allDeps map[string]*depth.Pkg, logger dfence.Logger) [][]string {
	cycles := [][]string{}

	logger.Debugf("Searching cycles for: %s", pkg)
	for _, dep := range allDeps[pkg].Deps {
		rFindCycles(pkg, &dep, allDeps, []string{}, cycles, logger)
	}

	return cycles
}

func rFindCycles(target string, pkg *depth.Pkg, allDeps map[string]*depth.Pkg, path []string, cycles [][]string, logger dfence.Logger) {
	if pkg == nil {
		return
	}

	if pkg.Name == target {
		logger.Debugf("Find cycle %v\n%s %s", path, pkg.Name)
		path = append(path, pkg.Name)
		cycles = append(cycles, path)
		return
	}

	path = append(path, pkg.Name)

	_, ok := allDeps[pkg.Name]
	if !ok {
		logger.Debugf("No deps info of %s", pkg.Name)
		return
	}

	for _, dep := range allDeps[pkg.Name].Deps {
		rFindCycles(target, &dep, allDeps, path, cycles, logger)
	}
}
