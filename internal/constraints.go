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
	Components           map[string]interface{}
	canonicalComponents  map[string][]pattern
	Classes              map[string]interface{}
	classIds             []string // **sorted** list of class ids
	Constraints          []Constraint
	canonicalConstraints []CanonicalConstraint
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

	err = policy.buildCanonicalConstraints()
	if err != nil {
		return Policy{}, fmt.Errorf("unable to aggregate policy constraints: %v", err)
	}

	return policy, nil
}

type pattern interface {
	match(string) bool
	String() string
}

type rawPattern struct {
	pattern string
}

func (p rawPattern) match(s string) bool {
	return strings.Contains(s, p.pattern)
}

func (p rawPattern) String() string {
	return p.pattern
}

type rePatter struct {
	pattern *regexp.Regexp
}

func (p rePatter) match(s string) bool {
	return p.pattern.MatchString(s)
}

func (p rePatter) String() string {
	return p.pattern.String()
}

func newREPattern(re string) (rePatter, error) {
	r, err := regexp.Compile(re)
	if err != nil {
		return rePatter{}, fmt.Errorf("unable to compile regexp '%s': %v", re, err)
	}

	return rePatter{pattern: r}, nil
}

// CanonicalConstraint is a plain raw (ie without references to components) dependency constraint
type CanonicalConstraint struct {
	scope             string // scope from which this constraint was built
	componentPatterns []pattern
	kind              constraintKind
	depPatterns       []pattern
	onBreak           errorLevel
}

func (c CanonicalConstraint) String() string {
	return fmt.Sprintf("scope\t%s\ncomps\t%v\nkind\t%v\ndeps\t%v\nlevel\t%v", c.scope, c.componentPatterns, c.kind, c.depPatterns, c.onBreak)
}

// ComponentsForPackage yields all components matching the given package
// Ideally, a package should match a single component but it will be not always the case
func (p *Policy) ComponentsForPackage(pkg string) []string {
	r := []string{}
	for k, patterns := range p.canonicalComponents {
		for _, pat := range patterns {
			if pat.match(pkg) {
				r = append(r, k)
			}
		}
	}

	return r
}

// buildCanonicalConstraints populates canonical constraints of a dependency policy
func (p *Policy) buildCanonicalConstraints() error {
	if len(p.canonicalConstraints) > 0 {
		return nil // avoid building twice
	}

	r := []CanonicalConstraint{}

	var err error
	p.canonicalComponents, err = p.extractComponentsPatterns()
	if err != nil {
		return err
	}

	classesPatterns, err := p.extractClassesPatterns(p.canonicalComponents)
	if err != nil {
		return err
	}

	for _, c := range p.Constraints {
		newConstraint := CanonicalConstraint{}
		newConstraint.scope = c.Scope
		for _, id := range strings.Split(c.Scope, patternSeparator) {
			if id == "" {
				continue
			}

			patterns, ok := resolveID(id, p.canonicalComponents, classesPatterns)
			if !ok {
				return fmt.Errorf("undefined id '%s' in constraint scope '%s' ", id, c.Scope)
			}

			newConstraint.componentPatterns = append(newConstraint.componentPatterns, patterns...)
		}
		newConstraint.kind = c.Kind
		for _, id := range strings.Split(c.Deps, patternSeparator) {
			if id == "" {
				continue
			}

			patterns, ok := resolveID(id, p.canonicalComponents, classesPatterns)
			if !ok {
				return fmt.Errorf("undefined id '%s' in constraint deps '%s'", id, c.Deps)
			}

			newConstraint.depPatterns = append(newConstraint.depPatterns, patterns...)
		}
		newConstraint.onBreak = c.OnBreak

		r = append(r, newConstraint)
	}

	p.canonicalConstraints = r

	return nil
}

// prefixREPattern prefix of regular expression patterns
const prefixREPattern = "#"

func buildPatterns(sp []string) ([]pattern, error) {
	result := []pattern{}
	for _, s := range sp {
		if strings.HasPrefix(s, prefixREPattern) {
			r, err := newREPattern(strings.TrimLeft(s, prefixREPattern))
			if err != nil {
				return nil, err
			}

			result = append(result, r)
			continue
		}

		result = append(result, rawPattern{pattern: s})
	}

	return result, nil
}

// GetApplicableConstraints yields constraints applicable to a given package
func (p Policy) GetApplicableConstraints(pkg string) (constraints []CanonicalConstraint) {
	constraints = []CanonicalConstraint{}
	for _, c := range p.canonicalConstraints {
		for _, p := range c.componentPatterns {
			if p.match(pkg) {
				constraints = append(constraints, c)
				break
			}
		}
	}

	return constraints
}

func (p Policy) extractComponentsPatterns() (map[string][]pattern, error) {
	r := map[string][]pattern{}
	for k, v := range p.Components {
		patterns, _ := v.(string) // TODO check type
		var err error
		r[k], err = buildPatterns(strings.Split(patterns, patternSeparator))
		if err != nil {
			return r, err
		}
	}

	return r, nil
}

func (p Policy) extractClassesPatterns(compPatterns map[string][]pattern) (map[string][]pattern, error) {
	r := map[string][]pattern{}

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

func resolveID(id string, comps, cls map[string][]pattern) ([]pattern, bool) {
	p, ok := cls[id]
	if !ok {
		p, ok = comps[id]
	}

	return p, ok
}

func getSortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func buildRegExprs(from string) []*regexp.Regexp {
	rExprs := []*regexp.Regexp{}

	for _, p := range strings.Split(from, patternSeparator) {
		re := regexp.MustCompile(p)
		rExprs = append(rExprs, re)
	}

	return rExprs
}
