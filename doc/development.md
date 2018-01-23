# Development Notes

Audience: Those who want to contribute to this repo.

## Directory/package organization

In each case, see the actual directory for more documentation.

1. [`core`](../core): The only directory that an application really
   needs.
1. [`cmd`](../cmd): Demos and utilities programs.
1. [`crew`](../crew): Basic container for a set of related machines.
1. [`interpreters`](../interpreters): Some example action/guard interpreters.
1. [`specs`](../specs): Example machine specifications.
1. [`tools`](../tools): Miscellaneous tools.
1. [`util`](../util): Generally useful utilties (hopefully frugal in their dependencies)
1. [`util/testutil`](../util/testutil): Utilities for demo and test code
1. [`doc`](../doc): Document (this directory)

## Small core

We're trying to keep `core/` light:

```Shell
go get github.com/KyleBanks/depth/cmd/depth
(cd core && depth -internal . )
```

See [Issue 13](https://github.com/Comcast/sheens/issues/13).

## YAML deserialization

Some code deserializes YAML.  We were using
[`github.com/go-yaml`](https://github.com/go-yaml/yaml), but that code
-- correctly -- deserializes maps as `map[interface{}]interface{}` and
not as `map[string]interface{}`.  (See
[issue 139](https://github.com/go-yaml/yaml/issues/139).) We also work
with JSON representations, and
[`encoding/json`](https://golang.org/pkg/encoding/json/) doesn't
support serialization of `map[interface{}]interface{}`.

Rather than either do lots of additional work or pull in reflection
code, we're just using a fork of
[`github.com/go-yaml`](https://github.com/go-yaml/yaml):
[`github.com/jsccast/yaml`](https://github.com/jsccast/yaml).  Yes,
we'll need to check to see that that fork remains current w.r.t. its
upstream.


