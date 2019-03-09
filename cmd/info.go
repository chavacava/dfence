package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/chavacava/dfence/internal/infra"
	"github.com/chavacava/dfence/internal/policy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cmdInfo = &cobra.Command{
	Use:   "info [package selector]",
	Short: "Provides information about policy on the given packages",
	Long:  "Provides information about policy on the given packages",
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(infra.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger.") // revive:disable-line:deep-exit
		}

		var err error
		stream, err := os.Open(policyFile)
		if err != nil {
			logger.Fatalf("Unable to open policy file %s: %+v", policyFile, err)
		}

		policy, err := policy.NewPolicyFromJSON(stream)
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

		err = info(policy, pkgs, logger)
		if err != nil {
			logger.Errorf(err.Error())
			os.Exit(1) // revive:disable-line:deep-exit
		}
	},
}

func init() {
	cmdPolicy.AddCommand(cmdInfo)
	cmdInfo.Flags().StringVar(&policyFile, "policy", "", "path to dependencies policy file ")
	cmdInfo.MarkFlagRequired("policy")
}

func info(p policy.Policy, pkgs []string, logger infra.Logger) error {
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
