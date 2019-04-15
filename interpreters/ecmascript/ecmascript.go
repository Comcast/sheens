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


// Package ecmascript provides an ECMAScript-compatible action
// interpreter.
package ecmascript

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/match"

	"github.com/dop251/goja"
	"github.com/gorhill/cronexpr"
)

var (
	// InterruptedMessage is the string value of Interrupted.
	InterruptedMessage = "RuntimeError: timeout"

	// Interrupted is returned by Exec if the execution is
	// interrupted.
	Interrupted = errors.New(InterruptedMessage)

	// IgnoreExit will prevent the Goja function "exit" from
	// terminating the process. Being able to halt the process
	// from Goja is useful for some tests and utilities.  Maybe.
	IgnoreExit = false
)

// init adds a Interpreter as one of the DefaultInterpreters
func init() {
	core.DefaultInterpreters["ecmascript"] = NewInterpreter()
}

// Interpreter implements core.Intepreter using Goja, which is a
// Go implementation of ECMAScript 5.1+.
//
// See https://github.com/dop251/goja.
type Interpreter struct {

	// Testing is used to expose or hide some runtime
	// capabilities.
	Test bool

	// Extended adds some additional properties.
	Extended bool
}

// NewInterpreter makes a new Interpreter.
func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

func wrapSrc(src string) string {
	return fmt.Sprintf("(function() {\n%s\n}());\n", src)
}

func AsSource(src interface{}) (code string, err error) {
	switch vv := src.(type) {
	case string:
		code = vv
		return
	default:
		err = errors.New(fmt.Sprintf("bad ECMAScript source (%T)", src))
		return
	}
}

// Compile calls goja.Compile.  This step is optional.
//
// See BenchmarkPrecompile and BenchmarkNoPrecompile for a comparison
// of what compilation can do for you.
func (i *Interpreter) Compile(ctx context.Context, src interface{}) (interface{}, error) {
	code, err := AsSource(src)
	if err != nil {
		return nil, err
	}

	code = wrapSrc(code)

	obj, err := goja.Compile("", code, true)
	if err != nil {
		return nil, errors.New(err.Error() + ": " + code)
	}

	return obj, nil
}

func protest(o *goja.Runtime, x interface{}) {
	panic(o.ToValue(x))
}

func deepCopy(x interface{}) (interface{}, error) {
	return core.Canonicalize(x)
}

// Exec implements the Interpreter method of the same name.
//
// The following properties are available from the runtime at _.
//
// These two things are most important:
//
//    bindings: the map of the current bindings.
//    props: core.StepProps
//    out(obj): Add the given object as a message to emit.
//
// Extended properties (enabled by interpreter's Extended property):
//
//    randstr(): generate a random string.
//    cronNext(s): Return a string representing (RFC3999Nano) the
//      next time for the given crontab expression.
//    esc(s): URL query-escape the given string.
//    match(pat, obj): Execute the pattern matcher.
//
// Testing properties (enabled by the interpreter's Test property):
//
//    sleep(ms): sleep for the given number of milliseconds.  For testing.
//    exit(msg): Terminate the process after printing the given message.
//      For testing.
//
func (i *Interpreter) Exec(ctx context.Context, bs match.Bindings, props core.StepProps, src interface{}, compiled interface{}) (*core.Execution, error) {
	exe := core.NewExecution(nil)

	var p *goja.Program
	if compiled == nil {
		var err error
		if compiled, err = i.Compile(ctx, src); err != nil {
			return exe, err
		}
	}
	var is bool
	if p, is = compiled.(*goja.Program); !is {
		return exe, fmt.Errorf("ECMAScript bad compilation: %T %#v", compiled, compiled)
	}

	env := map[string]interface{}{
		"ctx": ctx,
	}
	if props == nil {
		env["props"] = map[string]interface{}{}
	} else {
		env["props"] = map[string]interface{}(props.Copy())
	}

	if bs != nil {
		// This particular action interpreter allows code to
		// modify values, and we don't want any side effects.
		// So:
		x, err := deepCopy(bs)
		if err != nil {
			return nil, err
		}
		bsCopy, is := x.(map[string]interface{})
		if !is {
			return nil, fmt.Errorf("internal error: %#v copy failed; %s", bs, err)
		}
		env["bindings"] = bsCopy
	}

	o := goja.New()

	o.Set("_", env)

	// "output" adds the given message to the list of messages to
	// emit.
	env["out"] = func(x interface{}) interface{} {
		var err error

		switch vv := x.(type) {
		case goja.Value:
			x = vv.Export()
		}

		if x, err = core.Canonicalize(x); err != nil {
			// Will end up as a Javascript exception.
			panic(err)
		}

		exe.AddEmitted(x)

		return x
	}

	if i.Extended {
		env["randstr"] = func() interface{} {
			return core.Gensym(32)
		}

		// cronNext parses the given string as a crontab expression
		// using github.com/gorhill/cronexpr.  Returns the next time
		// as a string formatted in time.RFC3339Nano (UTC).
		env["cronNext"] = func(x interface{}) interface{} {
			switch vv := x.(type) {
			case goja.Value:
				x = vv.Export()
			}
			cronExpr, is := x.(string)
			if !is {
				protest(o, "not a string")
			}

			c, err := cronexpr.Parse(cronExpr)
			if err != nil {
				protest(o, err.Error())
			}
			return c.Next(time.Now()).UTC().Format(time.RFC3339Nano)
		}

		// match is a utility that invokes the pattern matcher.
		env["match"] = func(pat, mess, bs goja.Value) interface{} {
			var bindings match.Bindings

			if bs == nil {
				bindings = match.NewBindings()
			} else {

				// Having some trouble here.  Please don't
				// tell anyone I'm resorting to
				x, err := canonicalize(bs.Export())
				if err != nil {
					panic(err)
				}
				var is bool
				m, is := x.(map[string]interface{})
				if !is {
					panic("bad bindings")
				}
				bindings = match.Bindings(m)
			}

			var (
				p   interface{}
				m   interface{}
				err error
			)

			if p, err = canonicalize(pat.Export()); err != nil {
				panic(err)
			}

			if m, err = canonicalize(mess.Export()); err != nil {
				panic(err)
			}

			bss, err := match.Match(p, m, bindings)
			if err != nil {
				panic(err)
			}

			var x interface{}
			if x, err = canonicalize(bss); err != nil {
				panic(err)
			}

			return x
		}
	}

	if i.Test {

		env["sleep"] = func(n interface{}) interface{} {
			switch vv := n.(type) {
			case goja.Value:
				n = vv.Export()
			}
			ms, is := n.(int64)
			if !is {
				panic(fmt.Sprintf("a %T is not an %T", n, ms))
			}
			time.Sleep(time.Duration(ms) * time.Millisecond)
			return nil
		}

		env["log"] = func(x interface{}) interface{} {
			switch vv := x.(type) {
			case goja.Value:
				x = vv.Export()
			}
			js, err := json.Marshal(&x)
			if err != nil {
				log.Println("goja.log (can't marshal: " + err.Error() + ")")
			} else {
				log.Println(string(js))
			}

			return x
		}
		env["exit"] = func(n interface{}, msg interface{}) interface{} {
			switch vv := msg.(type) {
			case goja.Value:
				msg = vv.Export()
			}
			s, is := msg.(string)
			if !is {
				panic("not a string")
			}
			switch vv := n.(type) {
			case goja.Value:
				n = vv.Export()
			}
			ec, is := n.(int64)
			if !is {
				panic(fmt.Sprintf("a %T is not an %T", n, ec))
			}
			log.Println(s)
			if !IgnoreExit {
				os.Exit(int(ec))
			}
			return msg
		}
	}

	// We want to make sure that the following goroutine is
	// terminated as soon as possible.
	ictx, cancel := context.WithCancel(ctx)
	go func() {
		<-ictx.Done()
		// If this Exec method calls cancel() after RunProgram
		// returns, then we'll never see this
		// InterruptedMessage, which is actually the behavior
		// we want.  In this case, we weren't actually interrupted.
		o.Interrupt(InterruptedMessage)
	}()

	v, err := RunProgram(o, p)
	cancel()

	if err != nil {
		if _, is := err.(*goja.InterruptedError); is {
			return nil, Interrupted
		}
		return nil, err
	}

	x := v.Export()

	var result match.Bindings
	switch vv := x.(type) {
	case *goja.InterruptedError:
		return nil, vv
	case map[string]interface{}:
		result = match.Bindings(vv)
	case match.Bindings:
		result = vv
	case nil:
	default:
		return nil, fmt.Errorf("%#v (%T) isn't Bindings", x, x)
	}
	exe.Bs = result

	return exe, nil
}

// canonicalize is an abomination
func canonicalize(x interface{}) (interface{}, error) {
	js, err := json.Marshal(&x)
	if err != nil {
		return nil, err
	}
	var y interface{}
	if err = json.Unmarshal(js, &y); err != nil {
		return nil, err
	}
	return y, nil
}

func RunProgram(o *goja.Runtime, p *goja.Program) (v goja.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	return o.RunProgram(p)
}
