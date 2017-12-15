package core

import (
	"context"
	"errors"
	"strings"
)

var (
	// InterpreterNotFound occurs when you try to Compile an
	// ActionSource, and the required interpreter isn't in the
	// given map of interpreters.
	InterpreterNotFound = errors.New("interpreter not found")

	// DefaultInterpreters will be used in ActionSource.Compile if
	// the given nil interpreters.
	DefaultInterpreters = make(map[string]Interpreter)

	// Exp_PermanentBindings is a switch to enable an experiment
	// that makes a binding key ending in "!" a permament binding
	// in the sense that an Action cannot remove that binding.
	//
	// The implementation has overhead for gathering the permanent
	// bindings before the Action execution and restoring those
	// bindings after the execution.
	//
	// One motivation for this feature is enabling an updated Spec
	// to be asked to do something along the lines of migration.
	// The old Spec version can be remembered as a permanent
	// binding.  When that Spec is updated, we can take some
	// action.
	//
	// Another motivation is simply the persistence (so to speak)
	// of configuration-like bindings, which we do not want to
	// remove accidentally.  With this experiment enabled, an
	// Action can call for removing all bindings, but the result
	// will still include the permanent bindings.  Can make
	// writing Actions easier and safer.
	Exp_PermanentBindings = true
)

type Execution struct {
	Bs Bindings
	*Events
}

func NewExecution(bs Bindings) *Execution {
	return &Execution{
		Bs:     bs,
		Events: newEvents(),
	}
}

// Interpreter can optionally compile and execute code for Actions and guards.
type Interpreter interface {
	// Compile can make something that helps when Exec()ing the
	// code later.
	Compile(ctx context.Context, code interface{}) (interface{}, error)

	// Exec executes the code.  The result of previous Compile()
	// might be provided.
	Exec(ctx context.Context, bs Bindings, props StepProps, code interface{}, compiled interface{}) (*Execution, error)
}

// Action returns Bindings based on the given (current) Bindings.
type Action interface {
	// Exec executes this action.
	//
	// Third argument is for parameters (which can be exposed in
	// the Action's dynamic environment).
	//
	// ToDo: Generalize to return []Bindings?
	Exec(context.Context, Bindings, StepProps) (*Execution, error)

	// Binds optionally gives the set of patterns that match
	// bindings that could be bound during execution.
	//
	// If not nil, the returned set can help with static and
	// dynamic analysis of the machine.
	Binds() []Bindings

	// Emits optionally gives the set of patterns that match
	// messages that could be emitted by this action.
	Emits() []interface{}
}

// FuncAction is currently a wrapper around a Go function, but an Action
// will eventually be a specification for generating an outbound
// message.
type FuncAction struct {
	F func(context.Context, Bindings, StepProps) (*Execution, error) `json:"-" yaml:"-"`

	// Binds is an optional declaration that specifies what new
	// bindings this action might create.
	binds []Bindings

	// emits is an optional declaration of patterns that emitted
	// messages match.
	emits []interface{}
}

func (a *FuncAction) Binds() []Bindings {
	return a.binds
}

func (a *FuncAction) Emits() []interface{} {
	return a.emits
}

func isPermanent(p string) bool {
	return strings.HasSuffix(p, "!")
}

// Exec runs the given action.
func (a *FuncAction) Exec(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
	if a == nil {
		return NewExecution(bs), nil
	}

	var permanent map[string]interface{}
	if Exp_PermanentBindings {
		permanent = make(map[string]interface{}, len(bs))
		for p, v := range bs {
			if isPermanent(p) {
				permanent[p] = v
			}
		}
	}

	exe, err := a.F(ctx, bs, props)

	if Exp_PermanentBindings {
		for p, v := range permanent {
			exe.Bs[p] = v
		}
	}

	{ // This block just generates tracing data.
		if exe == nil {
			exe = NewExecution(nil)
		}
		t := map[string]interface{}{
			"action":  "executed",
			"emitted": len(exe.Events.Emitted),
			"bs":      exe.Bs,
		}
		if err != nil {
			t["error"] = err.Error()
		}
		exe.AddTrace(t)
	}

	return exe, err
}

// ActionSource can be compiled to an Action.
type ActionSource struct {
	Interpreter string      `json:"interpreter,omitempty" yaml:",omitempty"`
	Source      interface{} `json:"source"`
	Binds       []Bindings  `json:"binds,omitempty" yaml:",omitempty"`
}

// Copy makes a shallow copy.
//
// Needed for Specification.Copy().
func (a *ActionSource) Copy() *ActionSource {
	if a == nil {
		return nil
	}
	binds := make([]Bindings, len(a.Binds))
	for i, b := range a.Binds {
		binds[i] = b.Copy()
	}
	return &ActionSource{
		Interpreter: a.Interpreter,
		Source:      a.Source,
		Binds:       binds,
	}
}

// Compile attempts to compile the ActionSource into an Action using
// the given interpreters, which defaults to DefaultInterpreters.
func (a *ActionSource) Compile(ctx context.Context, interpreters map[string]Interpreter) (Action, error) {
	if interpreters == nil {
		interpreters = DefaultInterpreters
	}

	interpreter, have := interpreters[a.Interpreter]
	if !have {
		return nil, InterpreterNotFound
	}

	x, err := interpreter.Compile(ctx, a.Source)
	if err != nil {
		return nil, err
	}

	return &FuncAction{
		F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
			return interpreter.Exec(ctx, bs, props, a.Source, x)
		},
		binds: a.Binds,
	}, nil
}
