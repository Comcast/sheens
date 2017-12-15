# Ch-ch-changes

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
