package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/spf13/viper"

	"github.com/KyleBanks/depth"
	dfence "github.com/chavacava/dfence/internal"
	"github.com/spf13/cobra"
)

var maxDepth int
var format string

var cmdDepsList = &cobra.Command{
	Use:   "list [package selector]",
	Short: "list dependencies of the given packages",
	Long:  "list dependencies of the given packages",
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(dfence.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger.") // revive:disable-line:deep-exit
		}

		pkgSelector := "."
		if len(args) > 0 {
			pkgSelector = strings.Join(args, " ")
		}

		pkgs, err := retrievePackages(pkgSelector)
		if err != nil {
			logger.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
		}

		for _, p := range pkgs {
			t := depth.Tree{}
			if maxDepth > 0 {
				t.MaxDepth = maxDepth
			}

			err := t.Resolve(p)
			if err != nil {
				logger.Warningf("Unable to analyze package '%s': %v", p, err)
			}

			buf := new(bytes.Buffer)
			switch format {
			case "plain":
				writeDeps(buf, *t.Root)
			case "tree":
				writeDepsTree(buf, *t.Root, []bool{}, false)
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

func writeDeps(w io.Writer, p depth.Pkg) {
	fmt.Fprintf(w, "%s\n", p.String())
	for _, d := range p.Deps {
		writeDeps(w, d)
	}
}

// writeDepsTree borrowed from  https://github.com/KyleBanks/depth/blob/master/cmd/depth/depth.go
func writeDepsTree(w io.Writer, p depth.Pkg, status []bool, isLast bool) {
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

	for idx, d := range p.Deps {
		writeDepsTree(w, d, status, idx == len(p.Deps)-1)
	}
}

func init() {
	rootCmd.AddCommand(cmdDepsList)
	cmdDepsList.Flags().IntVar(&maxDepth, "maxdepth", 0, "generate a graph")
	cmdDepsList.Flags().StringVar(&format, "format", "plain", "output format: plain, tree")
}
