![logo](./doc/dfence-logo.png)

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

```
Usage of ./bin/dfence:
  -log string
        log level: none, error, warn, info, debug (default "info")
  -policy string
        path to dependencies policy file
```

`dfence` takes another parameter, the name of the packages to analyze. Here you can use `.` and `./...`

Examples:

```
dfence -log debug -policy policy.revive.json ./...
```

The above command runs `dfence` tu enforce constraints described in the file
`policy.revive.json` and over all the packages in the current directory and its
subdirectories.

```
dfence -policy policy.revive.json .
```

The above command runs `dfence` tu enforce constraints described in the file
`policy.revive.json` and over all the packages in the current directory.
