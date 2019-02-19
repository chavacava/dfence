package internal

import "io"

// constraintKind is the type of kind of constraints
type constraintKind string

const (
	// Allow is a rule kind that enumerates accepted dependencies
	Allow constraintKind = "allow"
	// Forbid is a rule kind that enumerates forbidden dependencies
	Forbid constraintKind = "forbid"
)

// errorLevel is the type of error level for broken dependency constraints
type errorLevel string

const (
	// Error error level
	Error errorLevel = "error"
	// Warn error level
	Warn errorLevel = "warn"
)

// Constraint represents the set of dependency constraints to enforce on a module
type Constraint struct {
	Module  string
	Kind    constraintKind
	Targets string
	OnBreak errorLevel
}

// Policy represents the set of dependency constraints to enforce
type Policy struct {
	Modules     map[string]interface{}
	Constraints []Constraint
}

// NewPolicyFromJSON builds a Policy from a JSON
func NewPolicyFromJSON(json io.Reader) (Policy, error) {
	// TODO
	return Policy{}, nil
}

type plainConstraint struct {
	modulePattern  string
	kind           constraintKind
	targetPatterns []string
	onBreak        errorLevel
}

func buildPlainConstraints(p Policy) []plainConstraint {
	// TODO
	return []plainConstraint{}
}
