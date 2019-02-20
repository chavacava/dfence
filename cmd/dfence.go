// Package main implements the command line interface of dfence
package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

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
	log.Println("Retrieving packages to check...")
	pkgs, err := retrievePackages(pkgSelector)
	if err != nil {
		log.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
	}

	fmt.Printf("Pacages %+v ?\n", pkgs)

	constraints := internal.BuildCanonicalConstraints(policy)
	checker := internal.NewChecker(constraints)

	pkgCount := len(pkgs)
	log.Printf("Will check dependencies of %d package(s).", pkgCount)
	status := 0
	var wg sync.WaitGroup
	out := make(chan internal.CheckResult, pkgCount)
	for _, pkg := range pkgs {
		wg.Add(1)
		go checker.CheckPkg(pkg, out, &wg)
	}

	wg.Wait()
	log.Printf("Check done.")

	for i := 0; i < pkgCount; i++ {
		result := <-out
		for _, w := range result.Warns {
			log.Println(w)
		}
		for _, e := range result.Errs {
			log.Println(e)
		}
		if len(result.Errs) > 0 {
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

	return r[:len(r)-1], nil
}
