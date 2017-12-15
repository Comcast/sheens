package core

// These errors are user errors, not internal errors.
//
// Probably should have a type just for user errors.

import (
	"errors"
)

type SpecNotCompiled struct {
	Spec *Spec
}

func (e *SpecNotCompiled) Error() string {
	return `spec "` + e.Spec.Name + `" not compiled`
}

type UnknownNode struct {
	Spec     *Spec
	NodeName string
}

func (e *UnknownNode) Error() string {
	return `node "` + e.NodeName + `" not found in spec "` + e.Spec.Name + `"`
}

type UncompiledAction struct {
	Spec     *Spec
	NodeName string
}

func (e *UncompiledAction) Error() string {
	return `uncompiled action at node "` + e.NodeName + `" in spec "` + e.Spec.Name + `"`
}

type BadBranching struct {
	Spec     *Spec
	NodeName string
}

func (e *BadBranching) Error() string {
	return `branching at node "` + e.NodeName + `" in spec "` + e.Spec.Name + `" ` +
		`has "message" branching and an action`
}

var TooManyBindingss = errors.New("too many bindingss")
