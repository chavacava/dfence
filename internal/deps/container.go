package deps

import (
	"fmt"

	"golang.org/x/tools/go/packages"
)

// DependenciesContainer provides utility functions to retrieve deps of packages
type DependenciesContainer struct {
	stdPackages map[string]struct{}
}

// NewDependenciesContainer creates a new container
// Returns an error if it is unable to retrieve the listing of the GO standard packages
func NewDependenciesContainer() (*DependenciesContainer, error) {
	stdPkgs, err := packages.Load(nil, "std")
	if err != nil {
		return nil, fmt.Errorf("unable to load standard GO packages, got: %w", err)
	}

	stdPackages := make(map[string]struct{}, len(stdPkgs))
	for _, p := range stdPkgs {
		stdPackages[p.PkgPath] = struct{}{}
	}

	return &DependenciesContainer{stdPackages: stdPackages}, nil
}

// GetPkgDeps retrieves the dependencies of the given package
func (d *DependenciesContainer) GetPkgDeps(pkg *packages.Package) map[string]struct{} {
	r := map[string]struct{}{}

	return d.getPkgDeps(pkg, r)
}

// getPkgDeps recursivelly retrieves dependencies of a package by inspecting its imports
func (d *DependenciesContainer) getPkgDeps(pkg *packages.Package, r map[string]struct{}) map[string]struct{} {
	if _, ok := r[pkg.PkgPath]; ok {
		return r // already seen this package
	}

	r[pkg.PkgPath] = struct{}{}

	for _, iDep := range pkg.Imports {
		if _, isStdPkg := d.stdPackages[iDep.PkgPath]; isStdPkg {
			continue // skip dependencies on standard packages
		}

		d.getPkgDeps(iDep, r)
	}

	return r
}
