# dfence

**dfence** (from _dependencies fence_) helps maintaining dependencies under 
control by enforcing dependencies policies on your packages.

## Describing dependencies policies

`dfence` will enforce dependencies policies described through a JSON file like the 
following:

```json
{
    "components":{
        "cli":"dfence/cmd",
        "internal":"dfence/internal"
    },
    "constraints":[
        {
            "scope":"internal",
            "kind":"forbid",
            "deps":"cli",
            "onbreak":"error"
        }
    ]
}
```

In the `components` section of the document we define logical names for sets of
package name patterns. All packages with names containing one of the given 
patterns will referred with the corresponding logical name.

The `constraints` section contains the dependency constraints to enforce.
`scope` lists all the components (thus packages patterns) to which the 
constraint applies.

A constraint can be of one of two `kind`s: `forbid` or `allows`. Meaning that a
a dependency will be forbidden(accepted) if it matches with one of the package 
patterns in `deps`.
The `onbreak` field that can take one of two values, `warn` or `error`, 
indicates the error level to produce when the constraint is not respected.

The previous example can be read: _Rise an error if a package with a path 
containing `dfence/internal` depends on a package with a path containing 
`dfence/cmd`._

The JSON Schema for the policy file is available [here](./doc/policy.schema.json)

## Usage

`dfence` takes two parameters:

* the JSON file describing the dependencies policies
* the set of packages to check

The first parameter is read from `stdin`, the second is optional and defaults to `.` (the package in the current dir)

Examples:

```
cat myConstraints.json | dfence ./...
```

The above command runs `dfence` tu enforce constraints described in the file
`myConstraints.json` and over all the packages in the current directory and its
subdirectories.

```
cat myConstraints.json | dfence .
```

The above command runs `dfence` tu enforce constraints described in the file
`myConstraints.json` and over all the packages in the current directory.
