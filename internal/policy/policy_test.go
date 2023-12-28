package policy

import (
	"os"
	"reflect"
	"regexp"
	"testing"
)

func TestNewPolicyFromJSON(t *testing.T) {
	// Open the dfence.json file
	file, err := os.Open("../../dfence.json")
	if err != nil {
		t.Fatalf("Failed to open dfence.json: %v", err)
	}
	defer file.Close()

	// Call the function with the file as input
	policy, err := NewPolicyFromJSON(file)
	if err != nil {
		t.Fatalf("Failed to create policy from JSON: %v", err)
	}

	// Check if the components are correctly loaded
	if _, exists := policy.Components["cli"]; !exists {
		t.Errorf("Expected 'cli' component to exist")
	}

	// Check if the constraints are correctly loaded
	if len(policy.Constraints) != 3 {
		t.Errorf("Expected 3 constraints, got %d", len(policy.Constraints))
	}

	// Check if the first constraint is correctly loaded
	firstConstraint := policy.Constraints[0]
	if firstConstraint.Name != "internal -x-> cli" {
		t.Errorf("Expected first constraint name to be 'internal -x-> cli', got '%s'", firstConstraint.Name)
	}

	secondConstraint := policy.Constraints[1]
	if secondConstraint.Name != "cli depends only with internal, vendored or golang" {
		t.Errorf("Expected second constraint name to be 'cli depends only with internal, vendored or golang', got '%s'", secondConstraint.Name)
	}

	thirdConstraint := policy.Constraints[2]
	if thirdConstraint.Name != "main only depends on cli" {
		t.Errorf("Expected third constraint name to be 'main only depends on cli', got '%s'", thirdConstraint.Name)
	}

}

func Test_rawPattern_match(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		p    rawPattern
		args args
		want bool
	}{
		{
			name: "Test case 1: Match found",
			p:    rawPattern{pattern: "test"},
			args: args{s: "this is a test string"},
			want: true,
		},
		{
			name: "Test case 2: Match not found",
			p:    rawPattern{pattern: "test"},
			args: args{s: "this is a string"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.match(tt.args.s); got != tt.want {
				t.Errorf("rawPattern.match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawPattern_String(t *testing.T) {
	tests := []struct {
		name string
		p    rawPattern
		want string
	}{
		{
			name: "Test case 1: String conversion",
			p:    rawPattern{pattern: "test"},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("rawPattern.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rePatter_match(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		p       rePatter
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Test case 1: Regex match found",
			p: func() rePatter {
				r, _ := newREPattern("test")
				return r
			}(),
			args: args{s: "this is a test string"},
			want: true,
		},
		{
			name: "Test case 2: Regex match not found",
			p: func() rePatter {
				r, _ := newREPattern("test")
				return r
			}(),
			args: args{s: "this is a string"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.match(tt.args.s); got != tt.want {
				t.Errorf("rePatter.match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rePatter_String(t *testing.T) {
	tests := []struct {
		name string
		p    rePatter
		want string
	}{
		{
			name: "Test case 1: Regex string conversion",
			p: func() rePatter {
				r, _ := newREPattern("test")
				return r
			}(),
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("rePatter.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newREPattern(t *testing.T) {
	type args struct {
		re string
	}
	tests := []struct {
		name    string
		args    args
		want    rePatter
		wantErr bool
	}{
		{
			name: "Valid regular expression",
			args: args{
				re: "[a-z]+",
			},
			want: rePatter{
				pattern: regexp.MustCompile("[a-z]+"),
			},
			wantErr: false,
		},
		{
			name: "Invalid regular expression",
			args: args{
				re: "[a-z",
			},
			want:    rePatter{},
			wantErr: true,
		},
		{
			name: "Empty regular expression",
			args: args{
				re: "",
			},
			want: rePatter{
				pattern: regexp.MustCompile(""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newREPattern(tt.args.re)
			if (err != nil) != tt.wantErr {
				t.Errorf("newREPattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newREPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanonicalConstraint_String(t *testing.T) {
	tests := []struct {
		name       string
		constraint CanonicalConstraint
		want       string
	}{
		{
			name: "Test Case 1",
			constraint: CanonicalConstraint{
				name:              "constraint1",
				scope:             "scope1",
				componentPatterns: []pattern{rawPattern{pattern: "pattern1"}},
				kind:              Allow,
				depPatterns:       []pattern{rawPattern{pattern: "pattern2"}},
				onBreak:           Error,
			},
			want: "name:\tconstraint1\nscope:\tscope1\ncomps:\t[pattern1]\nkind:\tallow\ndeps:\t[pattern2]\nlevel:\terror",
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.constraint.String(); got != tt.want {
				t.Errorf("CanonicalConstraint.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveID(t *testing.T) {
	comps := map[string][]pattern{
		"comp1": {rawPattern{"comp1"}},
		"comp2": {rawPattern{"comp2"}},
	}
	cls := map[string][]pattern{
		"class1": {rawPattern{"class1"}},
		"class2": {rawPattern{"class2"}},
	}

	patterns, ok := resolveID("comp1", comps, cls)
	if !ok || len(patterns) != 1 || patterns[0].String() != "comp1" {
		t.Errorf("Expected to resolve comp1, got %v", patterns)
	}

	patterns, ok = resolveID("class1", comps, cls)
	if !ok || len(patterns) != 1 || patterns[0].String() != "class1" {
		t.Errorf("Expected to resolve class1, got %v", patterns)
	}

	patterns, ok = resolveID("unknown", comps, cls)
	if ok {
		t.Errorf("Expected to not resolve unknown, got %v", patterns)
	}
}

func TestGetSortedKeys(t *testing.T) {
	m := map[string]interface{}{
		"b": 2,
		"a": 1,
		"c": 3,
	}

	keys := getSortedKeys(m)
	if len(keys) != 3 || keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("Expected sorted keys, got %v", keys)
	}
}

func TestBuildRegExprs(t *testing.T) {
	from := "abc def ghi"

	exprs := buildRegExprs(from)
	if len(exprs) != 3 || exprs[0].String() != "abc" || exprs[1].String() != "def" || exprs[2].String() != "ghi" {
		t.Errorf("Expected regular expressions, got %v", exprs)
	}

	from = "abc"

	exprs = buildRegExprs(from)
	if len(exprs) != 1 || exprs[0].String() != "abc" {
		t.Errorf("Expected regular expressions, got %v", exprs)
	}
}
