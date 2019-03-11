package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/KyleBanks/depth"
	"github.com/chavacava/dfence/internal/deps"
	"github.com/chavacava/dfence/internal/infra"
	"github.com/chavacava/dfence/internal/policy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var graphFile string

var cmdFindCycles = &cobra.Command{
	Use:   "find-cycles [package selector]",
	Short: "Spots dependency cycles among the given packages",
	Long:  "Spot dependency cycles among the given packages",
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(infra.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger.") // revive:disable-line:deep-exit
		}

		var err error
		stream, err := os.Open(policyFile)
		if err != nil {
			logger.Fatalf("Unable to open policy file %s: %+v", policyFile, err)
		}

		policy, err := policy.NewPolicyFromJSON(stream)
		if err != nil {
			logger.Fatalf("Unable to load policy : %v", err) // revive:disable-line:deep-exit
		}

		pkgSelector := strings.Join(args, " ")
		logger.Infof("Retrieving packages...")
		pkgs, err := retrievePackages(pkgSelector)
		if err != nil {
			logger.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
		}
		logger.Infof("Will work with %d package(s).", len(pkgs))

		pkg2comps, errs := getCompsForPkgs(pkgs, policy)
		if len(errs) > 0 {
			for _, e := range errs {
				logger.Errorf(e.Error())
			}
			return
		}

		allDeps := getAllDeps(pkgs, logger)

		cycles := []deps.DepChain{}
		for _, pkg := range pkgs {
			cycles = append(cycles, findCycles(pkg, allDeps, pkg2comps, logger)...)
		}

		if len(cycles) == 0 {
			logger.Infof("No cycles found")
			return
		}

		for _, depChain := range cycles {
			logger.Errorf("Cycle: %s", depChain.String())
		}

		if graphFile != "" {
			err := generateCyclicGraph(graphFile, cycles)
			if err != nil {
				logger.Errorf("Unable to generate graph: %v", err)
			}
		}
	},
}

func init() {
	cmdDeps.AddCommand(cmdFindCycles)
	cmdFindCycles.Flags().StringVar(&policyFile, "policy", "", "path to dependencies policy file")
	cmdFindCycles.MarkFlagRequired("policy")
	cmdFindCycles.Flags().StringVar(&graphFile, "graph", "", "path of the graph of cyclic dependencies to be generated")
}

func generateCyclicGraph(file string, cycles []deps.DepChain) error {
	outFile, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("unable to create file %s: %+v", file, err)
	}

	fmt.Fprintf(outFile, "strict digraph deps {\n")
	for _, cycle := range cycles {
		fmt.Fprint(outFile, cycle.AsDotEdges())
	}
	fmt.Fprintf(outFile, "}")

	return nil
}

func getCompsForPkgs(pkgs []string, p policy.Policy) (map[string]string, []error) {
	r := make(map[string]string, len(pkgs))
	errs := []error{}

	for _, pkg := range pkgs {
		comps, ok := p.ComponentsForPackage(pkg)
		if !ok {
			errs = append(errs, fmt.Errorf("unable to check cycles, package %s is not in a component", pkg))
			continue
		}
		if len(comps) > 1 {
			errs = append(errs, fmt.Errorf("package %s belongs multiple components: %v", pkg, comps))
			continue
		}

		r[pkg] = comps[0]
	}

	return r, errs
}

func getAllDeps(pkgs []string, logger infra.Logger) map[string]*depth.Pkg {
	logger.Debugf("Retrieving dependencies...")

	r := map[string]*depth.Pkg{}

	for _, pkg := range pkgs {
		t := depth.Tree{
			ResolveTest: true,
		}
		err := t.Resolve(pkg)
		if err != nil {
			logger.Warningf("Skipping package '%s', unable to analyze: %v", pkg, err.Error())
			continue
		}

		r[pkg] = t.Root
	}

	logger.Debugf("Retrieving dependencies... done")

	return r
}

func findCycles(pkg string, allDeps map[string]*depth.Pkg, pkg2comp map[string]string, logger infra.Logger) []deps.DepChain {
	cycles := []deps.DepChain{}

	comp, ok := pkg2comp[pkg]
	if !ok {
		logger.Warningf("Skipping package %s: isn't part of any component.", pkg)
		return cycles
	}

	logger.Debugf("Searching cycles for: %s of component %s", pkg, comp)

	depChain := deps.NewDepChain()
	depChain.Append(deps.NewCompoundChainItem(comp, pkg))
	for _, dep := range allDeps[pkg].Deps {
		rFindCycles(&dep, allDeps, depChain, &cycles, pkg2comp, logger)
	}

	return cycles
}

func rFindCycles(pkg *depth.Pkg, allDeps map[string]*depth.Pkg, depChain deps.DepChain, cycles *[]deps.DepChain, pkg2comp map[string]string, logger infra.Logger) {
	if pkg == nil {
		return // skip if pkg is nil. It happens in some cases TODO(chavacava)
	}

	comp, ok := pkg2comp[pkg.Name]
	if !ok {
		return // skip the package because it does not belong to any known component
	}

	depChain.Append(deps.NewCompoundChainItem(comp, pkg.Name))

	if depChain.IsCyclic() {
		*cycles = append(*cycles, depChain.Clone())
		return
	}

	_, ok = allDeps[pkg.Name]
	if !ok {
		logger.Debugf("No deps info of %s", pkg.Name)
		return
	}

	for _, dep := range allDeps[pkg.Name].Deps {
		rFindCycles(&dep, allDeps, depChain, cycles, pkg2comp, logger)
	}
}
