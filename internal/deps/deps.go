package deps

import (
	"fmt"

	"github.com/KyleBanks/depth"
)

type Pkg interface {
	Name() string
	Deps() []Pkg
	fmt.Stringer
}

type depthPkg struct {
	pkg depth.Pkg
	deps []Pkg
}

func (p depthPkg) String() string {
	return p.pkg.String()
}
func (p depthPkg) Name() string {
	return p.pkg.Name
}

func (p depthPkg) Deps() []Pkg {
	if len(p.deps)>0 {
		return p.deps
	}

	p.deps = []Pkg{}
	for _, pkg := range p.pkg.Deps {
		p.deps = append(p.deps, &depthPkg{pkg:pkg})
	}

	return p.deps 
}

// ExplainDep yields a list of dependency chains going from -> ... -> to
func ExplainDep(from Pkg, to string) []DepChain {
	explanations := []DepChain{}

	recExplainDep(from, to, NewDepChain(), &explanations)

	return explanations
}

func recExplainDep(pkg Pkg, explain string, chain DepChain, explanations *[]DepChain) {
	chain.Append(NewRawChainItem(pkg.Name()))

	if pkg.Name() == explain {
		*explanations = append(*explanations, chain)
		return
	}

	for _, dep := range pkg.Deps() {
		recExplainDep(dep, explain, chain.Clone(), explanations)
	}
}

// ResolvePkgDeps recursively finds all dependencies for the root pkg name provided,
// and the packages it depends on.
func ResolvePkgDeps(pkg string, maxDepth int) (depthPkg, error) {
	t := depth.Tree{ResolveTest: true}
	if maxDepth > 0 {
		t.MaxDepth = maxDepth
	}

	err := t.Resolve(pkg)

	return depthPkg{pkg:*t.Root}, err
}
