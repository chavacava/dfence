# Writing dFence policies

**dFence** enforces _dependency policies_.
Policies are a set of dependency _constraints_ that defines what dependencies are allowed between packages. If a package has a forbidden dependency, **dFence** will spot it.

A constraint has the form:

_packages_ are _allowed_ to depend on _packages_

or

_packages_ are _forbidden_ to depend on _packages_

A not allowed dependency is forbidden. A not forbidden dependency is allowed.

Writing constraints for each package of an application is cumbersome therefore, for writing constraints, **dFence** use the concept of _component_: a set of packages.

To define the packages of a component we use _patterns_:
all packages matching a pattern are part of the component.

Patterns are of one of two kinds: 

* _plain_ patterns are a string (not starting with #) that will match if a package name contains it. 

* _regexp_ is a string (starting with #) that encodes 
a regular expression. The pattern matches if the regular expression matches the package name.

Examples of patterns:

Candidate packages

```
github.com/chavacava/dbc4go
github.com/chavacava/dfence
github.com/chavacava/dfence/cmd
github.com/chavacava/dfence/internal
github.com/pkg/errors
```


| Pattern | Type | Example of matches |
|---------|------|--------------------|
| `github.com` | plain | github.com/chavacava/dbc4go <br>github.com/chavacava/dfence<br>github.com/chavacava/dfence/cmd<br>github.com/chavacava/dfence/internal<br>github.com/pkg/errors |
| `dfence` | plain | github.com/chavacava/dfence/cmd <br> github.com/chavacava/dfence <br> github.com/chavacava/dfence/internal |
| `chavacava` | plain | github.com/chavacava/dfence <br> github.com/chavacava/dfence/cmd <br> github.com/chavacava/dfence/internal <br> github.com/chavacava/dbc4go |
| `#github.com/chavacava/dfence$` | regexp | github.com/chavacava/dfence |
| `#github\.com\/(?!pkg).*` | regexp | github.com/chavacava/dbc4go <br>github.com/chavacava/dfence<br>github.com/chavacava/dfence/cmd<br>github.com/chavacava/dfence/internal<br>github.com/pkg/errors |
