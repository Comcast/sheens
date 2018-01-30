package noop

import (
	"context"
	"log"

	"github.com/Comcast/sheens/core"
)

// NoopInterpreter is an interpreter which just returns the bindings
// without modification.
type NoopInterpreter struct {
	// Silent, if false, will suppress warning log messages.
	Silent bool
}

func (i *NoopInterpreter) Compile(ctx context.Context, code interface{}) (interface{}, error) {
	if !i.Silent {
		log.Printf("warning: Using NoopInterpreter for compilation")
	}
	return nil, nil
}

func (i *NoopInterpreter) Exec(ctx context.Context, bs core.Bindings, props core.StepProps, code interface{}, compiled interface{}) (*core.Execution, error) {
	if !i.Silent {
		log.Printf("warning: Using NoopInterpreter for execution")
	}
	return core.NewExecution(bs), nil
}

type NoopInterpreters struct {
	I *NoopInterpreter
}

func NewNoopInterpreters() *NoopInterpreters {
	return &NoopInterpreters{
		I: &NoopInterpreter{},
	}
}

func (i *NoopInterpreters) Find(name string) core.Interpreter {
	return i.I
}
