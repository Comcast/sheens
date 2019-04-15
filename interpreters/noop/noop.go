/* Copyright 2018-2019 Comcast Cable Communications Management, LLC
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


// Package noop provides a no-op interpreter that can be handy for
// some tests.
package noop

import (
	"context"
	"log"

	"github.com/Comcast/sheens/core"
	. "github.com/Comcast/sheens/match"
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

// Exec returns the given bindings and emits no messages.
func (i *Interpreter) Exec(ctx context.Context, bs Bindings, props core.StepProps, code interface{}, compiled interface{}) (*core.Execution, error) {
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
