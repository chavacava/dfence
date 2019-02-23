package internal

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/KyleBanks/depth"
)

// Checker models a dependencies constraints checker
type Checker struct {
	policy Policy
	logger Logger
}

// NewChecker yields a dependencies constraint checker
func NewChecker(p Policy, l Logger) (Checker, error) {
	return Checker{policy: p, logger: l}, nil
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
func (c Checker) CheckPkg(pkg string, out chan<- CheckResult, wg *sync.WaitGroup) {
	defer wg.Done()

	errs := []error{}
	warns := []error{}

	applicableConstraints := c.policy.GetApplicableConstraints(pkg)

	c.logger.Debugf("Checking (%d constraints) package %s", len(applicableConstraints), pkg)
	if len(applicableConstraints) == 0 {
		c.logger.Warningf("%s does not have constraints.", pkg)
		out <- buildCheckResult(warns, errs)
		return
	}

	t := depth.Tree{
		ResolveInternal: true,
		ResolveTest:     true,
	}

	err := t.Resolve(pkg)
	if err != nil {
		out <- buildCheckResult(warns, append(errs, fmt.Errorf("unable to get dependencies of %s: %v", pkg, err)))
		return
	}

	pkgDeps := t.Root.Deps

	for _, constr := range applicableConstraints {
		switch kind := constr.kind; kind {
		case Allow:
			w, e := c.checkAllowConstraint(constr, pkg, pkgDeps)
			warns = append(warns, w...)
			errs = append(errs, e...)
		case Forbid:
			w, e := c.checkForbidConstraint(constr, pkg, pkgDeps)
			warns = append(warns, w...)
			errs = append(errs, e...)
		default:
			errs = append(errs, fmt.Errorf("unable to check constraints of kind `%s`", kind))
		}
	}

	out <- buildCheckResult(warns, errs)
}

func (c Checker) checkAllowConstraint(constraint CanonicalConstraint, pkg string, pkgDeps []depth.Pkg) (warns []error, errs []error) {
	errs = []error{}
	warns = []error{}

	for _, d := range pkgDeps {
		if d.Internal {
			continue // skip stdlib packages
		}

		ok := false
		for _, t := range constraint.depPatterns {
			if t == "" {
				continue // TODO check why this happens
			}
			c.logger.Debugf("allow: %s contains %s ?", d.Name, t)
			matches := strings.Contains(d.Name, t)
			if matches {
				ok = true
				break
			}
		}

		if !ok {
			warns, errs = c.appendByLevel(warns, errs, constraint.onBreak, fmt.Sprintf("%s depends on %s", pkg, d.Name))
		}
	}

	return warns, errs
}

func (c Checker) checkForbidConstraint(constraint CanonicalConstraint, pkg string, pkgDeps []depth.Pkg) (warns []error, errs []error) {
	errs = []error{}
	warns = []error{}

	for _, d := range pkgDeps {
		if d.Internal {
			continue // skip stdlib packages
		}

		ok := true
		for _, t := range constraint.depPatterns {
			matches := strings.Contains(d.Name, t)
			if matches {
				ok = false
				break
			}
		}

		if !ok {
			warns, errs = c.appendByLevel(warns, errs, constraint.onBreak, fmt.Sprintf("%s depends on %s", pkg, d.Name))
		}
	}

	return warns, errs
}

func (c Checker) appendByLevel(w, e []error, level errorLevel, msg string) (warns []error, errs []error) {
	newErr := errors.New(msg)

	if level == Warn {
		return append(w, newErr), e
	}

	return w, append(e, newErr)
}
