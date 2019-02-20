// Package main implements the command line interface of dfence
package main

import (
	"log"
	"os"

	"github.com/chavacava/dfence/internal"
	"github.com/mgechev/dots"
)

func main() {
	pkgFlag := os.Args[1:1]
	stream := os.Stdin

	policy, err := internal.NewPolicyFromJSON(stream)
	if err != nil {
		log.Fatalf("Unable to load policy : %v", err)
	}

	pkgs, err := retrievePackages(pkgFlag[0])
	if err != nil {
		log.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgFlag, err)
	}

	constraints := internal.BuildPlainConstraints(policy)
	checker := internal.NewChecker(constraints)

	status := 0
	for _, pkg := range pkgs {
		warns, errs := checker.CheckPkg(pkg)
		for _, w := range warns {
			log.Println(w)
		}
		for _, e := range errs {
			log.Println(e)
		}
		if len(errs) > 0 {
			status = 1
		}
	}

	os.Exit(status)
}

func retrievePackages(pkgSelector string) ([]string, error) {
	return dots.Resolve([]string{pkgSelector}, []string{"vendor/..."})
}
