![logo](./doc/dfence-logo.png)

# dfence

**dFence** (for _dependencies fence_) helps maintaining dependencies under 
control by enforcing dependencies policies on your packages.

A _dependencies policy_ defines dependencies constraints among the
components of your application. **dFence** will check that constraints
are respected.

## Describing dependencies policies

`dFence` will enforce dependencies policies described through a JSON file like the 
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
a dependency will be forbidden (accepted) if it matches with one of the package 
patterns in `deps`.
The `onbreak` field can take one of two values, `warn` or `error`; it 
indicates the error level to produce when the constraint is not respected.

The previous example can be read: _Rise an error if a package with a path 
containing `dfence/internal` depends on a package with a path containing 
`dfence/cmd`._

The JSON Schema for the policy file is available [here](./doc/policy.schema.json)

## Usage

Constraint checking is the main functionality of **dFence**, its usage is the
following:

```
Usage:
  dfence policy check [package selector] [flags]

Flags:
  -h, --help            help for check
      --policy string   path to dependencies policy file

Global Flags:
      --log string   log level: none, error, warn, info, debug (default "info")
```

**dFence** will perform the check on the package set defined by the 
_package selector_. Typically you will use `.` or `./...` to refer to the 
package defined in the current directory or to packages defined in the current 
directory and its subdirectories respectively.

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

### Other functionalities

**dFence** also provides commands for analyzing dependencies and to facilitate
the definition of policies.

#### `find-cycles`

This command will look for dependencies cycles.

```
Usage:
  dfence deps find-cycles [package selector] [flags]

Flags:
      --graph string    path of the graph of cyclic dependencies to be generated
  -h, --help            help for find-cycles
      --policy string   path to dependencies policy file

Global Flags:
      --log string   log level: none, error, warn, info, debug (default "info")
```

Example:

```
dfence deps find-cycles -policy policy.revive.json ./...
```

#### `list` 

Lists dependencies of packages

```
Usage:
  dfence deps list [package selector] [flags]

Flags:
      --format string   output format: plain, tree (default "plain")
  -h, --help            help for list
      --maxdepth int    maximum level of dependency nesting

Global Flags:
      --log string   log level: none, error, warn, info, debug (default "info")
```

Example:

```
dfence deps list ./... --format tree
```

#### `who`

List all packages that depend on a given one

```
Usage:
  dfence deps who [package] [package selector] [flags]

Flags:
      --graph   generate a graph
  -h, --help    help for who

Global Flags:
      --log string   log level: none, error, warn, info, debug (default "info")

```

Example:

```
dfence deps who github.com/spf13/jwalterweatherman ./...
```

Output:

```
github.com/chavacava/dfence/cmd -> github.com/spf13/viper -> github.com/spf13/jwalterweatherman
```

#### `why`

Explains why a package depends on the other

```
Usage:
  dfence deps why [package] [package] [flags]

Flags:
  -h, --help   help for why

Global Flags:
      --log string   log level: none, error, warn, info, debug (default "info")
```

Example:

```
dfence deps why github.com/chavacava/dfence  github.com/pelletier/go-toml
```
