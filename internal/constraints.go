package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
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
	Classes     map[string]interface{}
	classIds    []string // **sorted** list of class ids
	Constraints []Constraint
}

// NewPolicyFromJSON builds a Policy from a JSON
func NewPolicyFromJSON(stream io.Reader) (Policy, error) {
	var policy Policy

	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)

	err := json.Unmarshal(buf.Bytes(), &policy)
	if err != nil {
		return Policy{}, fmt.Errorf("unable to read policy from JSON file: %v", err)
	}

	policy.classIds = getSortedKeys(policy.Classes)

	return policy, nil
}

func getSortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

// CanonicalConstraint is a plain raw (ie without references to components) dependency constraint
type CanonicalConstraint struct {
	componentPatterns []string
	kind              constraintKind
	depPatterns       []string
	onBreak           errorLevel
}

// buildCanonicalConstraints yields canonical constraints from a dependency policy
func (p Policy) buildCanonicalConstraints() ([]CanonicalConstraint, error) {
	r := []CanonicalConstraint{}

	componentPatterns := p.extractComponentsPatterns()
	classesPatterns, err := p.extractClassesPatterns(componentPatterns)
	if err != nil {
		return r, err
	}

	for _, c := range p.Constraints {
		newConstraint := CanonicalConstraint{}
		for _, id := range strings.Split(c.Scope, patternSeparator) {
			if id == "" {
				continue
			}

			patterns, ok := resolveId(id, componentPatterns, classesPatterns)
			if !ok {
				return r, fmt.Errorf("undefined id '%s' in constraint scope '%s' ", id, c.Scope)
			}

			newConstraint.componentPatterns = append(newConstraint.componentPatterns, patterns...)
		}
		newConstraint.kind = c.Kind
		for _, id := range strings.Split(c.Deps, patternSeparator) {
			if id == "" {
				continue
			}

			patterns, ok := resolveId(id, componentPatterns, classesPatterns)
			if !ok {
				return r, fmt.Errorf("undefined id '%s' in constraint deps '%s'", id, c.Deps)
			}

			newConstraint.depPatterns = append(newConstraint.depPatterns, patterns...)
		}
		newConstraint.onBreak = c.OnBreak

		r = append(r, newConstraint)
	}

	return r, nil
}

func (p Policy) extractComponentsPatterns() map[string][]string {
	r := map[string][]string{}
	for k, v := range p.Components {
		patterns, _ := v.(string) // TODO check type
		r[k] = strings.Split(patterns, patternSeparator)
	}

	return r
}

func (p Policy) extractClassesPatterns(compPatterns map[string][]string) (map[string][]string, error) {
	r := map[string][]string{}

	for _, k := range p.classIds {
		v, _ := p.Classes[k]
		classDef, _ := v.(string) // TODO check type
		compRefs := strings.Split(classDef, patternSeparator)

		for _, cr := range compRefs {

			patterns, ok := compPatterns[cr]
			if !ok {
				patterns, ok = r[cr]
				if !ok {
					return r, fmt.Errorf("class %s refers unknown id '%s'", k, cr)
				}
			}

			r[k] = append(r[k], patterns...)
		}
	}

	return r, nil
}

func resolveId(id string, comps, cls map[string][]string) ([]string, bool) {
	p, ok := cls[id]
	if !ok {
		p, ok = comps[id]
	}

	return p, ok
}

func buildRegExprs(from string) []*regexp.Regexp {
	rExprs := []*regexp.Regexp{}

	for _, p := range strings.Split(from, patternSeparator) {
		re := regexp.MustCompile(p)
		rExprs = append(rExprs, re)
	}

	return rExprs
}
