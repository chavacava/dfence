package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/chavacava/dfence/internal/policy"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"golang.org/x/tools/go/packages"
)

func main() {
	policyFile := flag.String("policy", "dfence.json", "the policy file to enforce")
	logLevel := flag.String("log", "error", "log level: none, trace, debug, info, warn, error")
	mode := flag.String("mode", "check", "run mode (check or info)")
	flag.Parse()

	// Create a new instance of the logger. You can have any number of instances.
	logger := buildlogger(*logLevel)

	var err error
	stream, err := os.Open(*policyFile)
	if err != nil {
		logger.Fatalf("Unable to open policy file %s: %+v", *policyFile, err)
	}

	policy, err := policy.NewPolicyFromJSON(stream)
	if err != nil {
		logger.Fatalf("Unable to load policy : %v", err) // revive:disable-line:deep-exit
	}

	const pkgSelector = "./..."
	logger.Infof("Retrieving packages...")
	pkgs, err := scanPackages([]string{pkgSelector}, logger)

	if err != nil {
		logger.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
	}

	logger.Infof("Will work with %d package(s).", len(pkgs))

	var execErr error
	switch *mode {
	case "check":
		execErr = check(policy, pkgs, logger)
	case "info":
		execErr = info(policy, pkgs, logger)
	default:
		logger.Fatalf("Unknown mode %q, valid modes are check and info", *mode)
	}

	if execErr != nil {
		logger.Errorf(execErr.Error())
		os.Exit(1)
	}
}

// buildlogger is a function that creates and configures a new instance of a logrus.Logger.
// It takes a string parameter 'level' which determines the logging level.
func buildlogger(level string) *logrus.Logger {
	// Create a new instance of logrus.Logger
	logger := logrus.New()

	// Set the logger's formatter to logrus.TextFormatter with specific settings
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,         // Force colored output
		FullTimestamp:   true,         // Enable full timestamp in logs
		TimestampFormat: time.RFC3339, // Set the format of the timestamp
	})

	// Set the logger's output destination to the standard output, with support for color encoding
	logrus.SetOutput(colorable.NewColorableStdout())

	// Set the logger's level based on the 'level' parameter
	switch level {
	case "none":
		// Discard all log messages
		logger.Out = io.Discard
	case "trace":
		// Log all messages
		logger.SetLevel(logrus.TraceLevel)
	case "debug":
		// Log debug, info, warning, error, fatal and panic level messages
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		// Log info, warning, error, fatal and panic level messages
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		// Log warning, error, fatal and panic level messages
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		// Log error, fatal and panic level messages
		logger.SetLevel(logrus.ErrorLevel)
	}

	// Return the configured logger
	return logger
}

func check(p policy.Policy, pkgs []*packages.Package, logger *logrus.Logger) error {
	checker, err := policy.NewChecker(p, pkgs, logger)
	if err != nil {
		logger.Fatalf("Unable to run the checker: %v", err)
	}

	pkgCount := len(pkgs)
	errCount := 0
	out := make(chan policy.CheckResult, pkgCount)
	for _, pkg := range pkgs {
		go checker.CheckPkg(pkg, out)
	}

	logger.Infof("Checking...")

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

	logger.Infof("Check done")

	if errCount > 0 {
		return fmt.Errorf("found %d error(s)", errCount)
	}

	return nil
}

func info(p policy.Policy, pkgs []*packages.Package, logger *logrus.Logger) error {
	for _, pkg := range pkgs {
		pkgName := pkg.PkgPath
		cs := p.GetApplicableConstraints(pkgName)
		if len(cs) == 0 {
			logger.Warningf("No constraints for %s", pkgName)
			continue
		}
		logger.Infof("Constraints for %s:", pkgName)
		for _, c := range cs {
			for _, l := range strings.Split(c.String(), "\n") {
				logger.Infof("\t%+v", l)
			}
			logger.Infof("")
		}
	}

	return nil
}

func scanPackages(args []string, logger *logrus.Logger) ([]*packages.Package, error) {
	var emptyResult = []*packages.Package{}

	// Load packages and their dependencies.
	config := &packages.Config{
		Mode: packages.NeedName | packages.NeedImports | packages.NeedDeps | packages.NeedModule | packages.NeedFiles,
	}

	initial, err := packages.Load(config, args...)
	if err != nil {
		return emptyResult, fmt.Errorf("error loading packages: %w", err)
	}

	nerrs := 0
	for _, p := range initial {
		for _, err := range p.Errors {
			logger.Errorf(err.Msg)
			nerrs++
		}
	}

	if nerrs > 0 {
		return emptyResult, fmt.Errorf("failed to load initial packages. Ensure this command works first:\n\t$ go list %s", strings.Join(args, " "))
	}

	return initial, nil
}
