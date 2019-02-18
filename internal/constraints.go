package internal

type ruleKind string

const (
	// Allow is a rule kind that enumerates accepted dependencies
	Allow ruleKind = "allow"
	// Forbid is a rule kind that enumerates forbidden dependencies
	Forbid ruleKind = "forbid"
)

// Constraint represents the set of dependency constraints to enforce on a module
type Constraint struct {
	Module   string   `json:"module"`
	Kind     ruleKind `json:"kind"`
	Patterns string   `json:"patterns"`
}
