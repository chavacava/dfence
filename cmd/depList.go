package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/chavacava/dfence/internal/deps"
	"github.com/chavacava/dfence/internal/infra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var maxDepth int
var format string

var cmdDepsList = &cobra.Command{
	Use:   "list [package selector]",
	Short: "List dependencies of the given packages",
	Long: `List dependencies of the given packages.
	By default it will list dependencies of the package defined in the current dir`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(infra.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger.") // revive:disable-line:deep-exit
		}

		pkgSelector := "."
		if len(args) > 0 {
			pkgSelector = args[0]
		}

		pkgs, err := retrievePackages(pkgSelector)
		if err != nil {
			logger.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
		}

		for _, pkg := range pkgs {
			depsRoot, err := deps.ResolvePkgDeps(pkg, maxDepth)
			if err != nil {
				logger.Warningf("Unable to analyze package '%s': %v", pkg, err)
				continue
			}

			buf := new(bytes.Buffer)
			switch format {
			case "plain":
				writeDeps(buf, depsRoot)
			case "tree":
				writeDepsTree(buf, depsRoot, []bool{}, false)
			}

			out := buf.String()
			for _, l := range strings.Split(out, "\n") {
				logger.Infof("%s", l)
			}
		}
	},
}

const (
	outputPadding    = "│ "
	outputPrefix     = "├ "
	outputPrefixLast = "└ "
)

func writeDeps(w io.Writer, p deps.Pkg) {
	deps := map[string]struct{}{}
	for _, d := range p.Deps() {
		writeDepsRec(w, d, deps)
	}
	for k := range deps {
		fmt.Fprintf(w, "%s\n", k)
	}
}

func writeDepsRec(w io.Writer, p deps.Pkg, deps map[string]struct{}) {
	deps[p.Name()] = struct{}{}
	for _, d := range p.Deps() {
		writeDepsRec(w, d, deps)
	}
}

// writeDepsTree borrowed from  https://github.com/KyleBanks/depth/blob/master/cmd/depth/depth.go
func writeDepsTree(w io.Writer, p deps.Pkg, status []bool, isLast bool) {
	for i, isOpen := range status {
		if i == 0 {
			fmt.Fprintf(w, " ")
			continue
		}

		if isOpen {
			fmt.Fprintf(w, outputPadding)
			continue
		}

		fmt.Fprintf(w, " ")
	}

	status = append(status, true)

	var prefix string
	indent := len(status)
	if indent > 1 {
		prefix = outputPrefix

		if isLast {
			prefix = outputPrefixLast
			status[len(status)-1] = false
		}
	}

	fmt.Fprintf(w, "%v%v\n", prefix, p.String())

	kDeps := len(p.Deps())
	for idx, d := range p.Deps() {
		writeDepsTree(w, d, status, idx == kDeps-1)
	}
}

func init() {
	cmdDeps.AddCommand(cmdDepsList)
	cmdDepsList.Flags().IntVar(&maxDepth, "maxdepth", 0, "max distance between dependencies")
	cmdDepsList.Flags().StringVar(&format, "format", "plain", "output format: plain, tree")
}
