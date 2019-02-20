// Package main implements the command line interface of dfence
package main

import (
	"log"
	"os"
	"fmt"
	"github.com/chavacava/dfence/internal"
)

func main() {
	pkgFlag := os.Args[1:]

	stream := os.Stdin
	
	policy, err := internal.NewPolicyFromJSON(stream)
	if err!= nil {
		log.Fatalf("Unable to load policy : %v", err)
	}

	constraints := internal.BuildPlainConstraints(policy)
	fmt.Printf(">>>> plain constraints:%+v\n", constraints)
	
	checker := internal.NewChecker(constraints)

	status := 0
	for _, pkg := range pkgFlag {
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
