// Package deps provides definitions and functions to represent and handle dependencies
package deps

import (
	"fmt"

	"github.com/KyleBanks/depth"
)

// DepChainItem defines the type of items composing dependency chains
type DepChainItem interface {
	// name of the dependency item (typically a package or a component)
	name() string
	// String representation of the item
	String() string
}

// DepChain represents the chaining of pkg dependencies
type DepChain struct {
	items    []DepChainItem
	lastName string // component of the last item in the chain
	isCyclic bool
}

// NewDepChain yields a fresh new (empty) chain of dependencies
func NewDepChain() DepChain { return DepChain{items: []DepChainItem{}} }

const firstItemIndex = 0
const secondItemIndex = 1

// Append adds the item to the end of the chain of dependencies
func (c *DepChain) Append(item DepChainItem) {
	const minCyclicChainLength = 1

	c.items = append(c.items, item)

	c.isCyclic = (len(c.items) > minCyclicChainLength && c.lastName != item.name() && c.items[firstItemIndex].name() == item.name()) || c.isCyclic
	c.lastName = item.name()
}

// IsCyclic returns true if the chain has a cycle (calculated using names of items)
// If the chain is shorter than 2 elements, the function will return false
func (c *DepChain) IsCyclic() bool {
	return c.isCyclic
}

// AsDotEdges yields this chain in a representation compatible with DOT
func (c *DepChain) AsDotEdges() string {
	const minChainLength = 2
	result := ""

	if len(c.items) < minChainLength {
		return result // not enough items to have an edge
	}

	from := c.items[firstItemIndex].name()
	for i := secondItemIndex; i < len(c.items); i++ {
		to := c.items[i].name()
		if to == from {
			continue // skip self edges
		}
		result += fmt.Sprintf("%s -> %s \n", from, to)
		from = to
	}

	return result
}

// Items yields an ordered list of items of this chain.
func (c *DepChain) Items() []DepChainItem {
	return c.items
}

// Clone yields a shallow copy of this chain
func (c *DepChain) Clone() (result DepChain) {
	result = NewDepChain()
	for _, item := range c.items {
		result.Append(item) // Append also updates all chain fields (lastName, isCyclic...)
	}

	return
}

// String representation of this dependencies chain
func (c *DepChain) String() string {
	result := ""

	for _, item := range c.items {
		result += fmt.Sprintf("%s\n", item.String())
	}
	return result
}

// RawChainItem is the representation of an element in the chain of dependencies
type RawChainItem struct {
	aName string // name of the dependency item (typically a package or a component)
}

// NewRawChainItem yields a fresh new RawChainItem
func NewRawChainItem(name string) RawChainItem {
	return RawChainItem{aName: name}
}

func (i RawChainItem) name() string {
	return i.aName
}

// String representation of this item
func (i RawChainItem) String() string {
	return i.aName
}

// CompoundChainItem is the representation of an element in the chain of dependencies
// A compound item has two parts: a name and an attribute
type CompoundChainItem struct {
	aName string
	attr  string
}

// NewCompoundChainItem yields a fresh new CompoundChainItem
func NewCompoundChainItem(name, attr string) CompoundChainItem {
	return CompoundChainItem{aName: name, attr: attr}
}

func (i CompoundChainItem) name() string {
	return i.aName
}

// String representation of this item
func (i CompoundChainItem) String() string {
	return fmt.Sprintf("%s [%s]", i.aName, i.attr)
}

// ExplainDep yields a list of dependency chains going from -> ... -> to
func ExplainDep(from depth.Pkg, to string) []DepChain {
	explanations := []DepChain{}

	recExplainDep(from, to, NewDepChain(), &explanations)
	return explanations
}

func recExplainDep(pkg depth.Pkg, explain string, chain DepChain, explanations *[]DepChain) {
	chain.Append(NewRawChainItem(pkg.Name))

	if pkg.Name == explain {
		*explanations = append(*explanations, chain.Clone())
		return
	}

	for _, pkg := range pkg.Deps {
		recExplainDep(pkg, explain, chain, explanations)
	}
}
