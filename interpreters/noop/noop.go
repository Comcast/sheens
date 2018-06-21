package noop

import (
	"context"
	"log"

	"github.com/Comcast/sheens/core"
)

// Interpreter is an core.Interpreter which just returns the bindings
// without modification.
type Interpreter struct {
	// Silent, if false, will suppress warning log messages.
	Silent bool
}

func (i *Interpreter) Compile(ctx context.Context, code interface{}) (interface{}, error) {
	if !i.Silent {
		log.Printf("warning: Using Interpreter for compilation")
	}
	return nil, nil
}

func (i *Interpreter) Exec(ctx context.Context, bs core.Bindings, props core.StepProps, code interface{}, compiled interface{}) (*core.Execution, error) {
	if !i.Silent {
		log.Printf("warning: Using Interpreter for execution")
	}
	return core.NewExecution(bs), nil
}

type Interpreters struct {
	I *Interpreter
}

func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

func NewInterpreters() *Interpreters {
	return &Interpreters{
		I: &Interpreter{},
	}
}

func (i *Interpreters) Find(name string) core.Interpreter {
	return i.I
}
