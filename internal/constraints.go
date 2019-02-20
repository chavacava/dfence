package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const patternSeparator = " "

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

// Constraint represents the set of dependency constraints to enforce on a set of modules
type Constraint struct {
	Scope   string
	Kind    constraintKind
	Deps    string
	OnBreak errorLevel
}

// Policy represents the set of dependency constraints to enforce
type Policy struct {
	Components  map[string]interface{}
	Constraints []Constraint
}

// NewPolicyFromJSON builds a Policy from a JSON
func NewPolicyFromJSON(stream io.Reader) (Policy, error) {
	var policy Policy

	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)

	err := json.Unmarshal(buf.Bytes(), &policy)
	if err != nil {
		return Policy{}, errors.New(fmt.Sprintf("Unable to read policy from JSON file: %v", err))
	}

	return policy, nil
}

// CanonicalConstraint is a plain raw (ie without references to components) dependency constraint
type CanonicalConstraint struct {
	componentPatterns []string
	kind              constraintKind
	targetPatterns    []string
	onBreak           errorLevel
}

// BuildCanonicalConstraints yields canonical constraints from a dependency policy
func BuildCanonicalConstraints(p Policy) ([]CanonicalConstraint, error) {
	r := []CanonicalConstraint{}

	componentPatterns := extractComponentsPatterns(p.Components)

	for _, c := range p.Constraints {
		newConstraint := CanonicalConstraint{}
		for _, m := range strings.Split(c.Scope, patternSeparator) {
			p, ok := componentPatterns[m]
			if !ok {
				return r, fmt.Errorf("component '%s' undefined", m)
			}

			newConstraint.componentPatterns = append(newConstraint.componentPatterns, p...)
		}
		newConstraint.kind = c.Kind
		for _, d := range strings.Split(c.Deps, patternSeparator) {
			p, ok := componentPatterns[d]
			if !ok {
				return r, fmt.Errorf("component '%s' undefined", d)
			}

			newConstraint.targetPatterns = append(newConstraint.targetPatterns, p...)
		}
		newConstraint.onBreak = c.OnBreak

		r = append(r, newConstraint)
	}

	return r, nil
}

func extractComponentsPatterns(mods map[string]interface{}) map[string][]string {
	r := map[string][]string{}
	for k, v := range mods {
		patterns, _ := v.(string) // TODO check type
		r[k] = strings.Split(patterns, patternSeparator)
	}

	return r
}

func buildRegExprs(from string) []*regexp.Regexp {
	rExprs := []*regexp.Regexp{}

	for _, p := range strings.Split(from, patternSeparator) {
		re := regexp.MustCompile(p)
		rExprs = append(rExprs, re)
	}

	return rExprs
}
