# Code Notes

## Thin core

We're trying to keep `core/` light:

```Shell
go get github.com/KyleBanks/depth/cmd/depth
(cd core && depth . )
```

```
github.com/Comcast/sheens/core
  ├ context
  ├ encoding/json
  ├ errors
  ├ fmt
  ├ log
  ├ math/rand
  ├ strings
  ├ sync/atomic
  ├ time
  └ unsafe
10 dependencies (10 internal, 0 external, 0 testing).
```

The top-level `fmt` depedency is due to `go:generate stringer`, which
we of course don't really need.  ToDo: remove this `fmt` dependency.

The `encoding/json` dependency probably isn't critical, and it pulls
in a lot of code.  (See `depth -interal .`.)  ToDo: Remove that
dependency.

Oh, but `context` includes `fmt`, which drags in a ton of stuff.  (See
`depth -interal context`.)  Looks like `context` should maybe be freed
from its `fmt` dependency.  Alternately -- and much easier -- we can
just define our own `Ctx`, which has the exact same signature as
`context.Context`.  Then users of `core` can use a real
`context.Context` if they want, and `core` will be
`context.Context`-free.  ToDo: `Ctx` instead of `context.Context`.


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


