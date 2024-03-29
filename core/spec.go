/* Copyright 2021 Comcast Cable Communications Management, LLC
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

import (
	"context"
	"encoding/json" // ToDo: remove
	"errors"
)

var (
	// DefaultBranchType is used for Branches.Type when
	// Branches.Type is zero.
	DefaultBranchType = "bindings"

	// DefaultErrorNodeName is the name of the node state
	// switched to in the event of an internal error.
	DefaultErrorNodeName = "error"

	defaultErrorNode = &Node{}
)

// DefaultPatternParser is used during Spec.Compile if the
// given Spec has no PatternParser.
//
// This function is useful to allow a Spec to provide branch
// patterns in whatever syntax is convenient.  For example, if
// a Spec is authored in YAML, patterns in JSON might be more
// convenient (or easier to read) that patterns in YAML.
var DefaultPatternParser = func(syntax string, p interface{}) (interface{}, error) {
	switch syntax {
	case "none", "":
		return p, nil
	case "json":
		if js, is := p.(string); is {
			var x interface{}
			if err := json.Unmarshal([]byte(js), &x); err != nil {
				return nil, err
			}
			return x, nil
		}
		return p, nil
	default:
		return nil, errors.New("unsupposed pattern syntax: " + syntax)
	}
}

// Spec is a specification used to build a machine.
//
// A specification gives the structure of the machine.  This data does
// not include any state (such as the name of the current Node or a
// Machine's Bindings).
//
// If a specification includes Nodes with ActionSources, then the
// specification should be Compiled before use.
type Spec struct {
	// Name is the generic name for this machine.  Something like
	// "door-open-notification".  Cf. Id.
	Name string `json:"name,omitempty" yaml:",omitempty"`

	// Version is the version of this generic machine.  Something
	// like "1.2".
	Version string `json:"version,omitempty" yaml:",omitempty"`

	// Id should be a globally unique identifier (such as a hash
	// of a canonical representation of the Spec).
	//
	// This value could be used to determine when a Spec has
	// changed.
	//
	// This package does not read or write this value.
	Id string `json:"id,omitempty" yaml:",omitempty"`

	// Doc is general documentation about how this specification works.
	Doc string `json:"doc,omitempty" yaml:",omitempty"`

	// ParamSpecs is an optional name from a parameter name to a
	// specification for that parameter.
	//
	// A parameter is really just an initial binding that's
	// provided when a machine is created.
	//
	// ToDo: Implement actual check of parameters when machine is
	// created.
	ParamSpecs map[string]ParamSpec `json:"paramSpecs,omitempty" yaml:",omitempty"`

	// Uses is a set of feature tags.
	Uses []string `json:"uses,omitempty" yaml:",omitempty"`

	// Nodes is the structure of the machine.  This value could be
	// a reference that points into a library or whatever.
	Nodes map[string]*Node `json:"nodes,omitempty" yaml:",omitempty"`

	// ErrorNode is an optional name of a node for the machine in
	// the event of an internal error.
	//
	// Probably should just always assume the convention that a
	// node named 'error' is the error node.  ToDo: Consider.
	ErrorNode string `json:"errorNode,omitempty" yaml:",omitempty"`

	// NoAutoErrorNode will instruct the spec compiler not to add
	// an error node if one does not already exist.
	NoAutoErrorNode bool `json:"noErrorNode,omitempty" yaml:",omitempty"`

	// ActionErrorBranches (when true) means that this spec uses
	// branches to handle action errors.  (A branch can match an
	// action error using a "actionError" property with a variable
	// value.)
	//
	// If this switch is off, then any action error will result in
	// a transition to the error state, which is probably not what
	// you want.
	ActionErrorBranches bool `json:"actionErrorBranches,omitempty" yaml:",omitempty"`

	// ActionErrorNode is the name of the target node when an
	// action produces an error.
	//
	// If no value is given, then Step() will return an error
	// rather than a stride ending at a node given by this value.
	ActionErrorNode string `json:"actionErrorNode,omitempty" yaml:",omitempty"`

	// Boot is an optional Action that should be (?) executed when
	// the machine is loaded.  Not implemented yet.
	Boot Action `json:"-" yaml:"-"`

	// BootSource, if given, can be compiled to a Boot Action.
	// See Spec.Compile.
	BootSource *ActionSource `json:"boot,omitempty" yaml:"boot,omitempty"`

	// Toob is of course Boot in reverse.  It's also an optional
	// Action that can/should be executed when a Machine is
	// unloaded, suspended, or whatever.  Not currently connected
	// to anything.
	Toob Action `json:"-" yaml:"-"`

	// ToobSource, if given, can be compiled to a Toob Action.
	// See Spec.Compile.
	ToobSource *ActionSource `json:"toob,omitempty" yaml:"toob,omitempty"`

	// PatternSyntax indicates the syntax (if any) for branch patterns.
	PatternSyntax string `json:"patternSyntax,omitempty" yaml:",omitempty"`

	PatternParser func(string, interface{}) (interface{}, error) `json:"-" yaml:"-"`

	// NoNewMachines will make Step return an error if a pattern
	// match returns more than one set of bindings.
	//
	// ToDo: Implement.
	NoNewMachines bool `json:"noNewMachined,omitempty" yaml:",omitempty"`

	compiled bool
}

// Copy makes a deep copy of the Spec.
func (spec *Spec) Copy(version string) *Spec {
	if version == "" {
		version = spec.Version
	}
	ns := make(map[string]*Node, len(spec.Nodes))
	for name, n := range spec.Nodes {
		ns[name] = n.Copy()
	}

	return &Spec{
		Name:    spec.Name,
		Version: version,
		Doc:     spec.Doc,
		Nodes:   ns,
	}
}

// ParsePatterns parses branch patterns.
//
// The method Compile calls this method.  ParsePatterns is exposed to
// tools that might need to parse patterns without wanted to Compile
// them.
func (spec *Spec) ParsePatterns(ctx context.Context) error {
	if spec.PatternParser == nil {
		spec.PatternParser = DefaultPatternParser
	}

	if spec.Nodes == nil {
		return nil
	}

	for _, n := range spec.Nodes {
		if n == nil || n.Branches == nil {
			continue
		}

		for _, b := range n.Branches.Branches {
			if b == nil {
				continue
			}
			x, err := spec.PatternParser(spec.PatternSyntax, b.Pattern)
			if err != nil {
				return err
			}
			// ToDo: Remove
			if x, err = Canonicalize(x); err != nil {
				return err
			}
			b.Pattern = x
		}
	}
	return nil
}

// Compile compiles all action-like sources into actions. Might also do some
// other things. When force is true, everything is built from source even if
// it's been built before.
//
// Action-like sources include Actions, Boot, Toob, and Guards.
func (spec *Spec) Compile(ctx context.Context, interpreters Interpreters, force bool) error {

	if err := spec.ParsePatterns(ctx); err != nil {
		return err
	}

	if spec.BootSource != nil && (force || spec.Boot == nil) {
		action, err := spec.BootSource.Compile(ctx, interpreters)
		if err != nil {
			return err
		}
		spec.Boot = action
	}

	if spec.ToobSource != nil && (force || spec.Toob == nil) {
		action, err := spec.ToobSource.Compile(ctx, interpreters)
		if err != nil {
			return err
		}
		spec.Toob = action
	}

	if spec.ErrorNode == "" {
		spec.ErrorNode = DefaultErrorNodeName
	}

	if spec.Nodes == nil {
		spec.Nodes = make(map[string]*Node)
	}

	if _, have := spec.Nodes[spec.ErrorNode]; !have && !spec.NoAutoErrorNode {
		spec.Nodes[spec.ErrorNode] = defaultErrorNode
	}

	for name, n := range spec.Nodes {

		if n == nil {
			n = &Node{}
			spec.Nodes[name] = n
		}

		if n.ActionSource != nil && (force || n.Action == nil) {
			action, err := n.ActionSource.Compile(ctx, interpreters)
			if err != nil {
				src := "<opaque>"
				if s, is := n.ActionSource.Source.(string); is {
					src = s
				}
				return errors.New(err.Error() + ": node: " + name + " source:\n" + src)
			}
			n.Action = action
		}

		if n.Branches == nil {
			continue
		}

		switch n.Branches.Type {
		case "":
			n.Branches.Type = DefaultBranchType
		case "message", "bindings":
		default:
			return errors.New("unknown branching type '" + n.Branches.Type + "'")
		}

		for _, b := range n.Branches.Branches {
			x, err := spec.PatternParser(spec.PatternSyntax, b.Pattern)
			if err != nil {
				return err
			}
			// ToDo: Remove
			if x, err = Canonicalize(x); err != nil {
				return err
			}
			b.Pattern = x
			if b.GuardSource != nil && (force || b.Guard == nil) {
				guard, err := b.GuardSource.Compile(ctx, interpreters)
				if err != nil {
					return err
				}
				b.Guard = guard
			}
		}
	}

	spec.compiled = true

	return nil
}

// Node represents the structure of something like a state in a state machine.
//
// In our machines, the state is really given by (1) the name of the
// current node (At) and (2) the current Bindings.  A Node given a
// optional Action and possible state transitions.
type Node struct {
	// Doc is optional document in a format of your choosing.
	Doc string `json:"doc,omitempty" yaml:",omitempty"`

	// Action is an optional action that will be executed upon
	// transition to this node.
	//
	// Note that a node with "message"-based branching cannot have
	// an Action.
	Action Action `json:"-" yaml:"-"`

	// ActionSource, if given, is Compile()ed to an Action.
	ActionSource *ActionSource `json:"action,omitempty" yaml:"action,omitempty"`

	// Branches contains the transitions out of this node.
	Branches *Branches `json:"branching,omitempty" yaml:"branching,omitempty"`
}

// Copy makes a deep copy of the Node.
func (n *Node) Copy() *Node {
	return &Node{
		Doc:          n.Doc,
		Action:       n.Action,
		ActionSource: n.ActionSource.Copy(),
		Branches:     n.Branches.Copy(),
	}
}

// Terminal determines if a node has no branches.
func (n *Node) Terminal() bool {
	return n.Branches == nil || 0 == len(n.Branches.Branches)
}

// Branches represents the possible transitions to next states.
type Branches struct {
	// Type is either "message", "bindings", or nil.
	//
	// Type "message" means that an message is required and will be
	// consumed when branches are considered.  Branch Patterns are
	// matched against that message.
	//
	// Type "bindings" means that Branch Patterns are matched
	// against the current Bindings.
	//
	// A nil Type should imply only one Branch with no Pattern.
	Type string `json:"type,omitempty" yaml:",omitempty"`

	// Modes is a set of flags that can inform Branch processing
	// and analysis.  Currently no modes are considered.
	//
	// Example: "exclusive" might declare that the Branch patterns
	// should be mututally exclusive.
	//
	// ToDo: Use a real type instead of string.
	Modes []string `json:"modes,omitempty" yaml:",omitempty"`

	// Branches is the list (ordered) of possible transitions to
	// the next state (if any).
	//
	// No Branches means that this node is terminal.
	Branches []*Branch `json:"branches,omitempty" yaml:",omitempty"`
}

// Copy makes a deep copy of the Branches.
func (b *Branches) Copy() *Branches {
	if b == nil {
		return nil
	}
	modes := make([]string, len(b.Modes))
	for i, mode := range b.Modes {
		modes[i] = mode
	}
	bs := make([]*Branch, len(b.Branches))
	for i, br := range b.Branches {
		bs[i] = br.Copy()
	}
	return &Branches{
		Type:     b.Type,
		Modes:    modes,
		Branches: bs,
	}
}

// Branch is a possible transition to the next state.
type Branch struct {
	// Pattern is matched against either a pending message or
	// bindings -- depending on the Branches.Type.
	Pattern interface{} `json:"pattern,omitempty" yaml:",omitempty"`

	// Guard is an optional procedure that will prevent the
	// transition if the procedure returns nil Bindings.
	Guard Action `json:"-" yaml:"-"`

	// GuardSource is an ActionSource that serves as a guard to
	// following this branch.  If the guard returns nil bindings,
	// the branch isn't followed.  Otherwise, the returns bindings
	// are used and the branch is followed.
	GuardSource *ActionSource `json:"guard,omitempty" yaml:"guard,omitempty"`

	// Target is the name of the next state for this transition.
	Target string `json:"target,omitempty" yaml:",omitempty"`
}

// Copy makes a shallow copy of the Branch.
func (b *Branch) Copy() *Branch {
	if b == nil {
		return nil
	}
	return &Branch{
		Pattern:     b.Pattern,
		Guard:       b.Guard,
		GuardSource: b.GuardSource,
		Target:      b.Target,
	}
}
