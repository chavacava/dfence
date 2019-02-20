// Package main implements the command line interface of dfence
package main

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/chavacava/dfence/internal"
)

func main() {
	pkgFlag := []string{"."}
	if len(os.Args) > 1 {
		pkgFlag = os.Args[1:2]
	}

	stream := os.Stdin

	policy, err := internal.NewPolicyFromJSON(stream)
	if err != nil {
		log.Fatalf("Unable to load policy : %v", err)
	}

	pkgSelector := pkgFlag[0]
	pkgs, err := retrievePackages(pkgSelector)
	if err != nil {
		log.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
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
	r := []string{}
	cmd := exec.Command("go", "list", pkgSelector)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		return r, errors.New(errStr)
	}

	r = strings.Split(outStr, "\n")

	return r, nil
}
