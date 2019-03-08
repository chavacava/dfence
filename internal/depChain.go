package internal

import (
	"fmt"
)

// DepChain represents the chaining of pkg dependencies
type DepChain struct {
	items    []DepChainItem
	lastComp string // component of the last item in the chain
	isCyclic bool
}

// NewDepChain yields a fresh new (empty) chain of dependencies
func NewDepChain() DepChain { return DepChain{items: []DepChainItem{}} }

// Append adds the item to the end of the chain of dependencies
func (c *DepChain) Append(pkg, component string) {
	c.items = append(c.items, DepChainItem{pkg: pkg, component: component})

	c.isCyclic = (len(c.items) > 1 && c.lastComp != component && c.items[0].component == component) || c.isCyclic
	c.lastComp = component
}

// IsCyclic returns true if the first and last chain items are of the same component.
// If the chain is shorter than 2 elements, the function will return false
func (c *DepChain) IsCyclic() bool {
	return c.isCyclic
}

// AsDotEdges yields this chain in a representation compatible with DOT
func (c *DepChain) AsDotEdges() string {
	result := ""

	if len(c.items) < 2 {
		return result // not enough items to have an edge
	}

	from := c.items[0].component
	for i := 1; i < len(c.items); i++ {
		to := c.items[i].component
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

// String representation of this dependencies chain
func (c *DepChain) String() string {
	result := ""

	for _, item := range c.items {
		result += fmt.Sprintf("\t%s\n", item.String())
	}
	return result
}

// DepChainItem is the representation of a package in the chain of dependencies
type DepChainItem struct {
	pkg       string
	component string
}

// String yields the string representation
func (i DepChainItem) String() string {
	return fmt.Sprintf("%s (%s)", i.pkg, i.component)
}
