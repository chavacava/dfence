package deps

import (
	"reflect"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestNewDependenciesContainer(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Test case 1: Successful creation of DependenciesContainer",
			wantErr: false,
		},
		// @TODO Add more test cases as needed
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDependenciesContainer()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDependenciesContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Since the function NewDependenciesContainer() does not return a predictable result,
			// we cannot compare the result with a predefined expected result (tt.want).
			// @TODO Modify the function to return a predictable result to add a comparison here.
		})
	}
}

func TestDependenciesContainer_GetPkgDeps(t *testing.T) {
	type args struct {
		pkg *packages.Package
	}
	tests := []struct {
		name string
		d    *DependenciesContainer
		args args
		want map[string]struct{}
	}{
		{
			name: "Test case 1: Get dependencies of a package",
			d:    &DependenciesContainer{stdPackages: map[string]struct{}{"fmt": {}}},
			args: args{
				pkg: &packages.Package{
					PkgPath: "mypackage",
					Imports: map[string]*packages.Package{
						"fmt": {PkgPath: "fmt"},
					},
				},
			},
			want: map[string]struct{}{"mypackage": {}},
		},
		// @TODO: Add more test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.GetPkgDeps(tt.args.pkg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DependenciesContainer.GetPkgDeps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDependenciesContainer_getPkgDeps(t *testing.T) {
	type args struct {
		pkg *packages.Package
		r   map[string]struct{}
	}
	tests := []struct {
		name string
		d    *DependenciesContainer
		args args
		want map[string]struct{}
	}{
		{
			name: "Test case 1: Get dependencies of a package",
			d:    &DependenciesContainer{stdPackages: map[string]struct{}{"fmt": {}}},
			args: args{
				pkg: &packages.Package{
					PkgPath: "mypackage",
					Imports: map[string]*packages.Package{
						"fmt": {PkgPath: "fmt"},
					},
				},
				r: map[string]struct{}{},
			},
			want: map[string]struct{}{"mypackage": {}},
		},
		// @TODO: Add more test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.getPkgDeps(tt.args.pkg, tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DependenciesContainer.getPkgDeps() = %v, want %v", got, tt.want)
			}
		})
	}
}
