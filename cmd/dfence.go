// Package main implements the command line interface of dfence
package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	dfence "github.com/chavacava/dfence/internal"
	"github.com/fatih/color"
)

func main() {
	var policyFile string
	var loggerLevel string
	args := os.Args[1:]
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&policyFile, "policy", "", "policy file (defaults to stdin)")
	f.StringVar(&loggerLevel, "log", "info", "log level: none, error, warn, info (default), debug)")
	f.Parse(args)

	logger := buildlogger(loggerLevel)

	pkgFlag := []string{"."}
	if len(f.Args()) > 0 {
		pkgFlag = f.Args()[0:1]
	}

	stream := os.Stdin
	if policyFile != "" {
		var err error
		stream, err = os.Open(policyFile)
		if err != nil {
			logger.Fatalf("Unable to open policy file %s: %+v", policyFile, err)
		}
	} else {
		logger.Infof("Policy file not set, reading it from stdin...")
	}

	policy, err := dfence.NewPolicyFromJSON(stream)
	if err != nil {
		logger.Fatalf("Unable to load policy : %v", err)
	}

	pkgSelector := pkgFlag[0]
	logger.Infof("Retrieving packages to check...")
	pkgs, err := retrievePackages(pkgSelector)
	if err != nil {
		logger.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
	}

	constraints, err := dfence.BuildCanonicalConstraints(policy)
	if err != nil {
		logger.Fatalf("Unable to aggregate policy constraints: %v", err)
	}
	checker := dfence.NewChecker(constraints, logger)

	pkgCount := len(pkgs)
	logger.Infof("Will check dependencies of %d package(s).", pkgCount)
	status := 0
	var wg sync.WaitGroup
	out := make(chan dfence.CheckResult, pkgCount)
	for _, pkg := range pkgs {
		wg.Add(1)
		go checker.CheckPkg(pkg, out, &wg)
	}

	wg.Wait()
	logger.Infof("Check done.")

	for i := 0; i < pkgCount; i++ {
		result := <-out
		for _, w := range result.Warns {
			logger.Warningf(w.Error())
		}
		for _, e := range result.Errs {
			logger.Errorf(e.Error())
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

func buildlogger(level string) dfence.Logger {
	nop := func(string, ...interface{}) {}
	debug, info, warn, err := nop, nop, nop, nop
	switch level {
	case "none":
		// do nothing
	case "debug":
		debug = buildLoggerFunc(os.Stdout, "[DEBUG] ", color.New(color.FgCyan))
		fallthrough
	case "info":
		info = buildLoggerFunc(os.Stdout, "[INFO] ", color.New(color.FgGreen))
		fallthrough
	case "warn":
		warn = buildLoggerFunc(os.Stdout, "[WARN] ", color.New(color.FgBlue))
		fallthrough
	default:
		err = buildLoggerFunc(os.Stderr, "[ERROR] ", color.New(color.FgRed))
	}

	fatal := buildLoggerFunc(os.Stderr, "[FATAL] ", color.New(color.FgRed))
	return dfence.NewLogger(debug, info, warn, err, fatal)
}

func buildLoggerFunc(w io.Writer, prefix string, c *color.Color) dfence.LoggerFunc {
	return func(msg string, vars ...interface{}) {
		log.Println(c.Sprintf(prefix+msg, vars...))
	}
}
