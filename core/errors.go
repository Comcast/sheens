/* Copyright 2018 Comcast Cable Communications Management, LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

// These errors are user errors, not internal errors.
//
// Probably should have a type just for user errors.

import (
	"errors"
)

// SpecNotCompiled occurs when a Spec is used (say via Step()) before
// it has been Compile()ed.
type SpecNotCompiled struct {
	Spec *Spec
}

func (e *SpecNotCompiled) Error() string {
	return `spec "` + e.Spec.Name + `" not compiled`
}

// UnknownNode occurs when a branch is followed and its target node is
// not in the Spec.
type UnknownNode struct {
	Spec     *Spec
	NodeName string
}

func (e *UnknownNode) Error() string {
	return `node "` + e.NodeName + `" not found in spec "` + e.Spec.Name + `"`
}

// UncompiledAction occurs when an ActionSource execution is attempted
// but that ActionSource hasn't been Compile()ed.  Usually, this
// compilation happens as part of Spec.Compile().
type UncompiledAction struct {
	Spec     *Spec
	NodeName string
}

func (e *UncompiledAction) Error() string {
	return `uncompiled action at node "` + e.NodeName + `" in spec "` + e.Spec.Name + `"`
}

// BadBranching occurs when somebody the a Spec.Branches isn't right.
//
// For example, a Branch with an action must have braching type
// "message".  If not, you'll get an BadBranching error.
type BadBranching struct {
	Spec     *Spec
	NodeName string
}

func (e *BadBranching) Error() string {
	return `branching at node "` + e.NodeName + `" in spec "` + e.Spec.Name + `" ` +
		`has "message" branching and an action`
}

// TooManyBindingss occurs when a guard returns more than one set of
// bindings.
var TooManyBindingss = errors.New("too many bindingss")
