// Package policy provides definitions and functionality related to dependency policies
package policy

import (
	"errors"
	"fmt"

	"github.com/chavacava/dfence/internal/deps"
	"golang.org/x/tools/go/packages"
)

type logger interface {
	Warningf(string, ...interface{})
	Debugf(string, ...interface{})
}

// Checker models a dependencies constraints checker
type Checker struct {
	policy        Policy
	depsContainer *deps.DependenciesContainer
	logger        logger
}

// NewChecker yields a dependencies constraint checker
func NewChecker(p Policy, pkgs []*packages.Package, l logger) (*Checker, error) {
	depsContainer, err := deps.NewDependenciesContainer()
	if err != nil {
		return nil, err
	}

	return &Checker{policy: p, depsContainer: depsContainer, logger: l}, nil
}

// String yields a string representation of this checker
func (c Checker) String() string {
	return fmt.Sprintf("Checker constraints: %+v", c.policy.canonicalConstraints)
}

// CheckResult models the result of a dependency checking
type CheckResult struct {
	Warns []error
	Errs  []error
}

func buildCheckResult(warns, errs []error) CheckResult {
	return CheckResult{Warns: warns, Errs: errs}
}

// CheckPkg checks if the given package respects the dependency constraints of this checker
func (c Checker) CheckPkg(pkg *packages.Package, out chan<- CheckResult) {
	errs := []error{}
	warns := []error{}

	applicableConstraints := c.policy.GetApplicableConstraints(pkg.PkgPath)

	c.logger.Debugf("Checking (%d constraints) package %s", len(applicableConstraints), pkg)
	if len(applicableConstraints) == 0 {
		c.logger.Warningf("%s does not have constraints.", pkg)
		out <- buildCheckResult(warns, errs)
		return
	}

	pkgDeps := c.depsContainer.GetPkgDeps(pkg)
	for _, constr := range applicableConstraints {
		switch kind := constr.kind; kind {
		case Allow:
			w, e := c.checkAllowConstraint(constr, pkg.PkgPath, pkgDeps)
			warns = append(warns, w...)
			errs = append(errs, e...)
		case Forbid:
			w, e := c.checkForbidConstraint(constr, pkg.PkgPath, pkgDeps)
			warns = append(warns, w...)
			errs = append(errs, e...)
		default:
			errs = append(errs, fmt.Errorf("unable to check constraints of kind `%s`", kind))
		}
	}

	out <- buildCheckResult(warns, errs)
}

func (c Checker) checkAllowConstraint(constraint CanonicalConstraint, pkg string, pkgDeps map[string]struct{}) (warns []error, errs []error) {
	errs = []error{}
	warns = []error{}

	for depName := range pkgDeps {
		matchedAtLeastOne := false
		for _, t := range constraint.depPatterns {
			c.logger.Debugf("allow: %s matches %s ?", depName, t.String())
			if t.match(depName) {
				c.logger.Debugf("allow: YES")
				matchedAtLeastOne = true
				break
			}
			c.logger.Debugf("allow: NO")
		}

		if !matchedAtLeastOne {
			warns, errs = c.appendByLevel(warns, errs, constraint.onBreak, fmt.Sprintf("%s depends on %s so it breaks %q", pkg, depName, constraint.name))
		}
	}

	return warns, errs
}

func (c Checker) checkForbidConstraint(constraint CanonicalConstraint, pkg string, pkgDeps map[string]struct{}) (warns []error, errs []error) {
	errs = []error{}
	warns = []error{}

	for depName := range pkgDeps {
		for _, t := range constraint.depPatterns {
			c.logger.Debugf("forbid: %s matches %s ?", depName, t.String())
			if t.match(depName) {
				c.logger.Debugf("forbid: YES (ERROR)")
				warns, errs = c.appendByLevel(warns, errs, constraint.onBreak, fmt.Sprintf("%s depends on %s so it breaks %s", pkg, depName, constraint.name))
				break
			}
			c.logger.Debugf("forbid: NO")
		}
	}

	return warns, errs
}

func (Checker) appendByLevel(w, e []error, level errorLevel, msg string) (warns []error, errs []error) {
	newErr := errors.New(msg)

	if level == Warn {
		return append(w, newErr), e
	}

	return w, append(e, newErr)
}
