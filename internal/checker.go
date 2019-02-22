package internal

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/KyleBanks/depth"
)

// Checker models a dependencies constraints checker
type Checker struct {
	constraints []CanonicalConstraint
}

// NewChecker yields a dependencies constraint checker
func NewChecker(c []CanonicalConstraint) Checker {
	return Checker{constraints: c}
}

// String yields a string representation of this checker
func (c Checker) String() string {
	return fmt.Sprintf("Checker constraints: %+v", c.constraints)
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

	applicableConstraints := c.getApplicableConstraints(pkg)
	log.Printf("Checking (%d constraints) package %s", len(applicableConstraints), pkg)
	if len(applicableConstraints) == 0 {
		out <- buildCheckResult(warns, errs)
		return
	}

	t := depth.Tree{
		ResolveInternal: true,
		ResolveTest:     true,
	}

	err := t.Resolve(pkg)
	if err != nil {
		out <- buildCheckResult(warns, append(errs, fmt.Errorf("[%s] Unable to get dependencies of %s: %v", Error, pkg, err)))
		return
	}

	pkgDeps := t.Root.Deps

	for _, constr := range applicableConstraints {
		switch kind := constr.kind; kind {
		case Allow:
			w, e := checkAllowConstraint(constr, pkg, pkgDeps)
			warns = append(warns, w...)
			errs = append(errs, e...)
		case Forbid:
			w, e := checkForbidConstraint(constr, pkg, pkgDeps)
			warns = append(warns, w...)
			errs = append(errs, e...)
		default:
			errs = append(errs, fmt.Errorf("[%s] Unable to check constraints of kind `%s`", Error, kind))
		}
	}

	out <- buildCheckResult(warns, errs)
}

func checkAllowConstraint(c CanonicalConstraint, pkg string, pkgDeps []depth.Pkg) (warns []error, errs []error) {
	errs = []error{}
	warns = []error{}

	for _, d := range pkgDeps {
		if d.Internal {
			continue // skip stdlib packages
		}

		ok := false
		for _, t := range c.targetPatterns {
			if t == "" {
				continue // TODO check why this happens
			}
			log.Printf("allow: %s contains %s ?", d.Name, t)
			matches := strings.Contains(d.Name, t)
			if matches {
				ok = true
				break
			}
		}

		if !ok {
			warns, errs = appendByLevel(warns, errs, c.onBreak, fmt.Sprintf("[%s] %s depends on %s", c.onBreak, pkg, d.Name))
		}
	}

	return warns, errs
}

func checkForbidConstraint(c CanonicalConstraint, pkg string, pkgDeps []depth.Pkg) (warns []error, errs []error) {
	errs = []error{}
	warns = []error{}

	for _, d := range pkgDeps {
		if d.Internal {
			continue // skip stdlib packages
		}

		ok := true
		for _, t := range c.targetPatterns {
			matches := strings.Contains(d.Name, t)
			if matches {
				ok = false
				break
			}
		}

		if !ok {
			warns, errs = appendByLevel(warns, errs, c.onBreak, fmt.Sprintf("[%s] %s depends on %s", c.onBreak, pkg, d.Name))
		}
	}

	return warns, errs
}

func appendByLevel(w, e []error, level errorLevel, msg string) (warns []error, errs []error) {
	newErr := errors.New(msg)

	if level == Warn {
		return append(w, newErr), e
	}

	return w, append(e, newErr)
}

func (c Checker) getApplicableConstraints(pkg string) (constraints []CanonicalConstraint) {
	constraints = []CanonicalConstraint{}
	for _, constr := range c.constraints {
		for _, mp := range constr.componentPatterns {
			log.Printf("[DEBUG] %s contains %s", pkg, mp)
			if strings.Contains(pkg, mp) {
				constraints = append(constraints, constr)
				break
			}
		}
	}

	return constraints
}
