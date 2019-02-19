package internal

import (
	"github.com/KyleBanks/depth"
	"errors"
	"fmt"
	"strings"
)

type checker struct {
	constraints []plainConstraint
}

func NewChecker(c []plainConstraint) checker {
	return checker{constraints: c}
}

func (c checker) String() string {
	return fmt.Sprintf("Checker constraints: %+v", c.constraints)
}

func (c checker) CheckPkg(pkg string) (warns []error, errs []error) {
	errs = []error{}
	warns = []error{}

	applicableConstraints := c.getApplicableConstraints(pkg)
	fmt.Printf(">>>> applicable constraints:%+v\n", applicableConstraints)
	if len(applicableConstraints) == 0 {
		return warns, errs
	}

	t := depth.Tree {
		ResolveInternal: true,
		ResolveTest: true,
	}

	err := t.Resolve(pkg)
	if err != nil {
		return warns, append(errs,errors.New(fmt.Sprintf("[%s] Unable to get dependencies of %s: %v",Error,pkg,err)))
	}

	pkgDeps := t.Root.Deps

	for _, constr := range applicableConstraints {
		switch kind:=constr.kind; kind {
		case Allow, Forbid:
			w, e := checkConstraint(constr,pkg,pkgDeps) 
			warns = append(warns, w...)
			errs = append(errs, e...)
		default:
			errs = append(errs,errors.New(fmt.Sprintf("[%s] Unable to check constraints of kind `%s`",Error,kind)))
		}
	}

	return warns, errs
}

func checkConstraint(c plainConstraint, pkg string, pkgDeps []depth.Pkg) (warns []error,errs []error) {
	errs = []error{}
	warns = []error{}
	
	for _, d := range pkgDeps {
		ok := false
		for _,t:= range c.targetPatterns {
			matches := strings.Contains(d.Name,t)
			if matches && c.kind == Allow {
				ok = true
				break
			}
			if matches && c.kind == Forbid {
				break
			}
		}

		if !ok {
			warns, errs = appendByLevel(warns, errs, c.onBreak, fmt.Sprintf("[%s] %s depends on %s",c.onBreak,pkg,d.Name))
		}
	}

	return warns, errs
}

func appendByLevel(w, e []error, level errorLevel, msg string) (warns []error,errs []error) {
	newErr := errors.New(msg)

	if level == Warn {
		return append(w, newErr), e
	}

	return w, append(e, newErr)
}

func (c checker) getApplicableConstraints(pkg string) (constraints []plainConstraint) {
	constraints = []plainConstraint{}
	for _, constr := range c.constraints {
		for _, mp := range constr.modulePatterns {
			println(">>> check if",pkg,"contains", mp)
			if strings.Contains(pkg,mp) {
				println(">>> yes")
				constraints = append(constraints, constr)
				break
			}
		}
	}

	return constraints
}