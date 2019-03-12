package deps

import (
	"github.com/KyleBanks/depth"
)

// ExplainDep yields a list of dependency chains going from -> ... -> to
func ExplainDep(from depth.Pkg, to string) []DepChain {
	explanations := []DepChain{}

	recExplainDep(from, to, NewDepChain(), &explanations)

	return explanations
}

func recExplainDep(pkg depth.Pkg, explain string, chain DepChain, explanations *[]DepChain) {
	chain.Append(NewRawChainItem(pkg.Name))

	if pkg.Name == explain {
		*explanations = append(*explanations, chain)
		return
	}

	for _, dep := range pkg.Deps {
		recExplainDep(dep, explain, chain.Clone(), explanations)
	}
}

// ResolvePkgDeps recursively finds all dependencies for the root pkg name provided,
// and the packages it depends on.
func ResolvePkgDeps(pkg string, maxDepth int) (*depth.Pkg, error) {
	t := depth.Tree{ResolveTest: true}
	if maxDepth > 0 {
		t.MaxDepth = maxDepth
	}

	err := t.Resolve(pkg)

	return t.Root, err
}
