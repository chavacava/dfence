[![Build Status](https://travis-ci.com/chavacava/dfence.svg?branch=master)](https://travis-ci.com/chavacava/dfence)
[![Go Report Card](https://goreportcard.com/badge/github.com/chavacava/dfence)](https://goreportcard.com/report/github.com/chavacava/dfence)
 [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

![logo](./doc/dfence-logo.png)

# dFence

**dFence** (for _dependencies fence_) helps maintaining dependencies under 
control by enforcing dependencies policies on your packages.

A _dependencies policy_ defines dependencies constraints among the
components of your application. **dFence** will check if these constraints
are respected.

## Installation

```
go get github.com/chavacava/dfence
```

Requirements:

* GO >= 1.11 installed

### Building from sources

1. clone the repo: `git clone https://github.com/chavacava/dfence.git`
2. set `GO111MODULE=on`
3. `make build` will generate an executable under `./bin`

## Usage

```
  dfence [flags]

  -log string
        log level: none, error, warn, info, debug (default "error")
  -mode string
        run mode (check or info) (default "check")
  -policy string
        the policy file to enforce (default "dfence.json")
```

**dFence** will perform the check on packages defined in the current 
directory and its subdirectories.

Examples:

```
dfence -log debug -policy policy.revive.json
```

The above command runs `dfence` to enforce constraints described in the file
`policy.revive.json` and over all the packages in the current directory and its
subdirectories.

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
a dependency will be forbidden (or accepted) if it matches with one of the package 
patterns in `deps`.
The `onbreak` field can take one of two values, `warn` or `error`; it 
indicates the error level to produce when the constraint is not respected.

The previous example can be read: _Rise an error if a package with a path 
containing `dfence/internal` depends on a package with a path containing 
`dfence/cmd`._

Please read the [more detailed documentation on how to write policies](./doc/policy.md).
Also, a JSON Schema for the policy file is [available here](./doc/policy.schema.json)

