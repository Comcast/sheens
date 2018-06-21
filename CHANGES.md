# Ch-ch-changes

## Goja interpreter is deprecated

In favor of `interpreters/ecmascript`.  This interpreter doesn't
support libraries, which should instead be provided at a higher level
and compiled into specs.  That way a spec is a stand-alone entity,
which is an important characteristic for robustess and determinism.
You can of course always implement your own interpreter that supports
libraries (or whatever else).

The `interpreters/ecmascript` has a `randstr()` function (instead of
the old, misnamed `gensym`), which is available only if the
interpreter's `Extended` property is set (to `true`).

See [`interpreters`](interpreters) for the function `Standard()` that
provides a standard set of interpreters.

## `cmd/sheensio`

See the [README](cmd/sheensio/README.md).


## `Params` renamed to `StepProps`

`core.Params` was a confusing name since a `Spec` can have
`ParamSpecs`, which are related.

`core.Params` has been renamed to `core.StepProps`.  With the standard
Goja interpreter, that data is at `_.props` (instead of at
`_.params`).


## Explicit `return` required by standard Goja actions/guards

Now you must explicitly `return` your bindings in Goja actions and
guards.


## Experimental permanent bindings

A binding key ending in "!" will survive any Action's attempt to
remove it.

See the documentation for `core.Exp_PermanentBindings` in
[`core/actions.go`](core/actions.go).

This experiment and the two changes below are motivated by a desire to
have better support for migrating a Machine's state when its Spec is
updated.


## Spec Ids

A Spec can have an Id, which should be a globally unique identifier
for that spec and its canonical equivalents.

See the documentation for `core.Spec.Id` in
[`core/spec.go`](core/spec.go).


## Experimental branch target variables

If a branch target has the form @VAR, the the real target is the value
of VAR in the current bindings (if any).

See the documentation for `core.Exp_BranchTargetVariables` in
[`core/step.go`](core/step.go).
