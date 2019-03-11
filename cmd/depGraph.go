package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/KyleBanks/depth"
	"github.com/chavacava/dfence/internal/infra"
	"github.com/chavacava/dfence/internal/policy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var skip string

var cmdDepsGraph = &cobra.Command{
	Use:   "graph [package selector]",
	Short: "Outputs a graph of dependencies",
	Long:  "Outputs a graph of dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(infra.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger.") // revive:disable-line:deep-exit
		}

		pkgSelector := "./..."

		stream, err := os.Open(policyFile)
		if err != nil {
			logger.Fatalf("Unable to open policy file %s: %+v", policyFile, err)
		}

		policy, err := policy.NewPolicyFromJSON(stream)
		if err != nil {
			logger.Fatalf("Unable to load policy : %v", err) // revive:disable-line:deep-exit
		}

		pkgs, err := retrievePackages(pkgSelector)
		if err != nil {
			logger.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
		}

		toSkip := map[string]bool{}
		for _, s := range strings.Split(skip, ",") {
			toSkip[s] = true
		}

		deps := map[string]struct{}{}
		for _, p := range pkgs {
			t := depth.Tree{}
			if maxDepth > 0 {
				t.MaxDepth = maxDepth
			}

			err := t.Resolve(p)
			if err != nil {
				logger.Warningf("Unable to analyze package '%s': %v", p, err)
			}

			writeDepsGraphRec(*t.Root, deps, policy, toSkip)
		}

		output := os.Stdout
		if len(args) > 0 {
			var err error
			output, err = os.Create(args[0])
			if err != nil {
				logger.Fatalf("Unable to create file '%s': %v", args[0], err)
			}
		}

		fmt.Fprintln(output, "strict digraph deps {")
		for k := range deps {
			fmt.Fprintf(output, "%s\n", k)
		}
		fmt.Fprintln(output, "}")
	},
}

func writeDepsGraphRec(p depth.Pkg, deps map[string]struct{}, policy policy.Policy, toSkip map[string]bool) {
	from := getNodeLabel(p, policy)

	for _, d := range p.Deps {
		to := getNodeLabel(d, policy)

		if mustSkip(from, to, toSkip) {
			continue
		}

		deps[from+" -> "+to] = struct{}{}

		writeDepsGraphRec(d, deps, policy, toSkip)
	}
}

func mustSkip(from, to string, toSkip map[string]bool) bool {
	return from == to || toSkip[from] || toSkip[to]
}

func getNodeLabel(p depth.Pkg, policy policy.Policy) string {
	comps, ok := policy.ComponentsForPackage(p.Name)
	if !ok {
		return "UNDEFINED"
	}

	return comps[0] // return only the first component of a package
}

func init() {
	cmdDeps.AddCommand(cmdDepsGraph)
	cmdDepsGraph.Flags().StringVar(&policyFile, "policy", "", "path to dependencies policy file")
	cmdDepsGraph.MarkFlagRequired("policy")
	cmdDepsGraph.Flags().IntVar(&maxDepth, "maxdepth", 0, "max distance between dependencies")
	cmdDepsGraph.Flags().StringVar(&skip, "skip", "", "nodes to skip")
}
