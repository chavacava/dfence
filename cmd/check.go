package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/viper"

	dfence "github.com/chavacava/dfence/internal"
	"github.com/spf13/cobra"
)

var policyFile string

var cmdCheck = &cobra.Command{
	Use:   "check [flags] [packages]",
	Short: "Check policy on given packages",
	Long:  `check if the packages respect the dependencies policy.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(dfence.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger.") // revive:disable-line:deep-exit
		}

		var err error
		stream, err := os.Open(policyFile)
		if err != nil {
			logger.Fatalf("Unable to open policy file %s: %+v", policyFile, err)
		}

		policy, err := dfence.NewPolicyFromJSON(stream)
		if err != nil {
			logger.Fatalf("Unable to load policy : %v", err) // revive:disable-line:deep-exit
		}

		pkgSelector := strings.Join(args, " ")
		logger.Infof("Retrieving packages...")
		pkgs, err := retrievePackages(pkgSelector)
		if err != nil {
			logger.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
		}
		logger.Infof("Will work with %d package(s).", len(pkgs))

		err = check(policy, pkgs, logger)
		if err != nil {
			logger.Errorf(err.Error())
			os.Exit(1) // revive:disable-line:deep-exit
		}
	},
}

func init() {
	rootCmd.AddCommand(cmdCheck)
	cmdCheck.Flags().StringVar(&policyFile, "policy", "", "path to dependencies policy file ")
	cmdCheck.MarkFlagRequired("policy")
}

func check(p dfence.Policy, pkgs []string, logger dfence.Logger) error {
	checker, err := dfence.NewChecker(p, logger)
	if err != nil {
		logger.Fatalf("Unable to run the checker: %v", err)
	}

	pkgCount := len(pkgs)
	errCount := 0
	out := make(chan dfence.CheckResult, pkgCount)
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
