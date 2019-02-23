// Package main implements the command line interface of dfence
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
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
	var mode string
	args := os.Args[1:]
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&policyFile, "policy", "", "path to dependencies policy file ")
	f.StringVar(&loggerLevel, "log", "info", "log level: none, error, warn, info, debug")
	f.StringVar(&mode, "mode", "check", "running mode: check, info")
	f.Parse(args)

	logger := buildlogger(loggerLevel)

	pkgFlag := []string{"."}
	if len(f.Args()) > 0 {
		pkgFlag = f.Args()[0:1]
	}

	if policyFile == "" {
		logger.Fatalf("Policy file not set.")
	}

	var err error
	stream, err := os.Open(policyFile)
	if err != nil {
		logger.Fatalf("Unable to open policy file %s: %+v", policyFile, err)
	}

	policy, err := dfence.NewPolicyFromJSON(stream)
	if err != nil {
		logger.Fatalf("Unable to load policy : %v", err)
	}

	pkgSelector := pkgFlag[0]
	logger.Infof("Retrieving packages...")
	pkgs, err := retrievePackages(pkgSelector)
	if err != nil {
		logger.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
	}
	logger.Infof("Will work with %d package(s).", len(pkgs))

	switch mode {
	case "check":
		err = check(policy, pkgs, logger)
	case "info":
		err = info(policy, pkgs, logger)
	default:
		logger.Fatalf("Unknown running mode '%s'", mode)
	}

	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
}

func check(p dfence.Policy, pkgs []string, logger dfence.Logger) error {
	checker, err := dfence.NewChecker(p, logger)
	if err != nil {
		logger.Fatalf("Unable to run the checker: %v", err)
	}

	pkgCount := len(pkgs)
	errCount := 0
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

		errCount += len(result.Errs)
	}

	if errCount > 0 {
		return fmt.Errorf("found %d error(s)", errCount)
	}

	return nil
}

func info(p dfence.Policy, pkgs []string, logger dfence.Logger) error {
	for _, pkg := range pkgs {
		cs := p.GetApplicableConstraints(pkg)
		if len(cs) == 0 {
			logger.Warningf("No constraints for %s", pkg)
			continue
		}
		logger.Infof("Constraints for %s:", pkg)
		for _, c := range cs {
			for _, l := range strings.Split(c.String(), "\n") {
				logger.Infof("\t%+v", l)
			}
		}
	}

	return nil
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
		debug = buildLoggerFunc("[DEBUG] ", color.New(color.FgCyan))
		fallthrough
	case "info":
		info = buildLoggerFunc("[INFO] ", color.New(color.FgGreen))
		fallthrough
	case "warn":
		warn = buildLoggerFunc("[WARN] ", color.New(color.FgHiYellow))
		fallthrough
	default:
		err = buildLoggerFunc("[ERROR] ", color.New(color.BgHiRed))
	}

	fatal := buildLoggerFunc("[FATAL] ", color.New(color.BgRed))
	return dfence.NewLogger(debug, info, warn, err, fatal)
}

func buildLoggerFunc(prefix string, c *color.Color) dfence.LoggerFunc {
	return func(msg string, vars ...interface{}) {
		log.Println(c.Sprintf(prefix+msg, vars...))
	}
}
