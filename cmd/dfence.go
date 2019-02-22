// Package main implements the command line interface of dfence
package main

import (
	"bytes"
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/chavacava/dfence/internal"
	"github.com/fatih/color"
)

func main() {
	_ = buildLogger()
	var policyFile string
	args := os.Args[1:]
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&policyFile, "policy", "", "policy file (defaults to stdin)")
	f.Parse(args)

	pkgFlag := []string{"."}
	if len(f.Args()) > 0 {
		pkgFlag = f.Args()[0:1]
	}

	stream := os.Stdin
	if policyFile != "" {
		var err error
		stream, err = os.Open(policyFile)
		if err != nil {
			log.Fatalf("Unable to open policy file %s: %+v", policyFile, err)
		}
	} else {
		log.Println("Policy file not set, reading it from stdin...")
	}

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

	constraints, err := internal.BuildCanonicalConstraints(policy)
	if err != nil {
		log.Fatalf("Unable to aggregate policy constraints: %v", err)
	}
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

func buildLogger() internal.Logger {

	debug := buildLoggerFunc("[DEBUG] ", color.FgCyan)
	info := buildLoggerFunc("[INFO] ", color.FgGreen)
	warn := buildLoggerFunc("[WARN] ", color.FgBlue)
	err := buildLoggerFunc("[ERROR] ", color.FgRed)

	return internal.NewLogger(debug, info, warn, err, os.Stderr, os.Stdout)
}

func buildLoggerFunc(prefix string, c color.Color) internal.LoggerFunc {
	return func(msg string, vars ...interface{}) {
		c.Printf(prefix+msg, vars...)
	}
}
