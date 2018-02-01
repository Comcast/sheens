package goja

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Comcast/sheens/core"

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
	core.DefaultInterpreters["goja"] = NewInterpreter()
}

// Interpreter implements core.Intepreter using Goja, which is a
// Go implementation of ECMAScript 5.1+.
//
// See https://github.com/dop251/goja.
type Interpreter struct {

	// Testing is used to expose or hide some runtime
	// capabilities.
	Testing bool

	// Provider is a pluggable library provider, which can be used
	// instead of (or in addition to) the standard Provide method,
	// which will just use DefaultProvider if this Provider is
	// nil.
	//
	// A problem: For a multitenant service, we need some access
	// control. If a single LibraryProvider will provide all the
	// libraries for all tenants, we need a mechanism to provide
	// access control.  We could add another parameter that
	// carries the required data (something related to tenant
	// name), but it's hard to provide something generic.  With
	// trepidation, perhaps just use a Value in the ctx?
	LibraryProvider func(ctx context.Context, i *Interpreter, libraryName string) (string, error)
}

// NewInterpreter makes a new Interpreter.
func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

// CompileLibraries checks any libraries at LibrarySources.
//
// This method originally precompiled these libraries, but Goja can't
// current support combining ast.Programs.  So we won't actually use
// anything we precompile!  Perhaps in the future.  But we can at
// least check that the libraries do in fact compile.
func (i *Interpreter) CompileLibrary(ctx context.Context, name, src string) (interface{}, error) {
	return goja.Compile(name, src, true)
}

// ProvideLibrary resolves the library name into a library.
//
// We experimented with other approaches including returning parsed
// code and a struct representing a library.  Probably will want to
// move back in that direction.
func (i *Interpreter) ProvideLibrary(ctx context.Context, name string) (string, error) {
	if i.LibraryProvider != nil {
		return i.LibraryProvider(ctx, i, name)
	}
	return DefaultLibraryProvider(ctx, i, name)
}

var DefaultLibraryProvider = MakeFileLibraryProvider(".")

// DefaultProvider is a method that Provide will use if the
// interpreter's Provider is nil.
//
// This method supports (barely) names that are URLs with protocols of
// "file", "http", and "https". There currently is no additional
// control when using HTTP/HTTPS.
func MakeFileLibraryProvider(dir string) func(context.Context, *Interpreter, string) (string, error) {
	return func(ctx context.Context, i *Interpreter, name string) (string, error) {
		parts := strings.SplitN(name, "://", 2)
		if 2 != len(parts) {
			return "", fmt.Errorf("bad link '%s'", name)
		}
		switch parts[0] {
		case "file":
			// ToDo: Maybe protest any ".."?
			filename := parts[1]
			bs, err := ioutil.ReadFile(dir + "/" + filename)
			if err != nil {
				return "", err
			}
			return string(bs), nil
		case "http", "https":
			req, err := http.NewRequest("GET", name, nil)
			if err != nil {
				return "", err
			}
			req = req.WithContext(ctx)
			client := http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return "", err
			}
			switch resp.StatusCode {
			case http.StatusOK:
				bs, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return "", err
				}
				return string(bs), nil
			default:
				return "", fmt.Errorf("library fetch status %s %d",
					resp.Status, resp.StatusCode)
			}
		default:
			return "", fmt.Errorf("unknown protocol '%s'", parts[0])
		}
	}
}

func MakeMapLibraryProvider(srcs map[string]string) func(context.Context, *Interpreter, string) (string, error) {
	return func(ctx context.Context, i *Interpreter, name string) (string, error) {
		src, have := srcs[name]
		if !have {
			return "", fmt.Errorf("undefined library '%s'", name)
		}
		return src, nil
	}
}

func wrapSrc(src string) string {
	return fmt.Sprintf("(function() {\n%s\n}());\n", src)
}

// parseSource looks into the given map to try to find "requires" and
// "code" properties.
//
// Background: The YAML parser https://github.com/go-yaml/yaml will
// return map[interface{}]interface{}, which is correct but
// inconvenient.  So this repo uses a fork at
// https://github.com/jsccast/yaml, which will return
// map[string]interface{}.  However, this parseSource function
// supports map[interface{}]interface{} so that others don't need to
// use that fork.
func parseSource(vv map[string]interface{}) (code string, libs []string, err error) {
	x, have := vv["code"]
	if !have {
		code = ""
	}
	if s, is := x.(string); is {
		code = s
	} else {
		err = errors.New("bad Goja action code")
		return
	}

	x, have = vv["requires"]
	switch vv := x.(type) {
	case string:
		libs = []string{vv}
	case []string:
		libs = vv
	case []interface{}:
		libs = make([]string, 0, len(vv))
		for _, x := range vv {
			switch vv := x.(type) {
			case string:
				libs = append(libs, vv)
			default:
				err = errors.New("bad library")
				return
			}
		}
	}

	return
}

func AsSource(src interface{}) (code string, libs []string, err error) {
	switch vv := src.(type) {
	case string:
		code = vv
		return
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, v := range vv {
			str, ok := k.(string)
			if !ok {
				err = errors.New(fmt.Sprintf("bad src key (%T)", k))
				return
			}
			m[str] = v
		}
		return parseSource(m)
	case map[string]interface{}:
		return parseSource(vv)
	default:
		err = errors.New(fmt.Sprintf("bad Goja source (%T)", src))
		return
	}
}

// Compile calls goja.Compile after calling InlineRequires.
//
// This method can block if the interpreter's library Provider blocks
// in order to obtain external libraries.
func (i *Interpreter) Compile(ctx context.Context, src interface{}) (interface{}, error) {
	code, libs, err := AsSource(src)
	if err != nil {
		return nil, err
	}

	code = wrapSrc(code)

	// We no longer do InlineRequires.  Instead, we use an
	// explicit "requires".
	//
	// Background: Since we now want an explicit `return` of
	// bindings, we're in a block context, and in-lining code in a
	// block context would -- I guess -- require that the inlined
	// code (the libraries) also be blocks, which they might not
	// be.  Maybe document, enforce, and support later.
	//
	// if code, err = InlineRequires(ctx, code, i.ProvideLibrary); err != nil {
	//     return nil, err
	// }

	var libsSrc string
	for _, lib := range libs {
		libSrc, err := i.ProvideLibrary(ctx, lib)
		if err != nil {
			return nil, err
		}
		libsSrc += libSrc + "\n"
	}

	code = libsSrc + code

	obj, err := goja.Compile("", code, true)
	if err != nil {
		return nil, errors.New(err.Error() + ": " + code)
	}

	return obj, nil
}

func protest(o *goja.Runtime, x interface{}) {
	panic(o.ToValue(x))
}

// Exec implements the Interpreter method of the same name.
//
// The following properties are available from the runtime at _.
//
// These two things are most important:
//
//    bindings: the map of the current bindings.
//    out(obj): Add the given object as a message to emit.
//
// Some useful utilities:
//
//    gensym(): generate a random string.
//    esc(s): URL query-escape the given string.
//    match(pat, obj): Execute the pattern matcher.
//
// For testing only:
//
//    sleep(ms): sleep for the given number of milliseconds.  For testing.
//    exit(msg): Terminate the process after printing the given message.
//      For testing.
//
// The Testing flag must be set to see sleep().
func (i *Interpreter) Exec(ctx context.Context, bs core.Bindings, props core.StepProps, src interface{}, compiled interface{}) (*core.Execution, error) {
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
		return exe, fmt.Errorf("Goja bad compilation: %T %#v", compiled, compiled)
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
		env["bindings"] = map[string]interface{}(bs.Copy())
	}

	o := goja.New()

	o.Set("_", env)

	if i.Testing {
		o.Set("sleep", func(ms int) {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		})
	}

	env["gensym"] = func() interface{} {
		return core.Gensym(32)
	}

	env["genstr"] = func() interface{} {
		return core.Gensym(32)
	}

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

	env["esc"] = func(x interface{}) interface{} {
		switch vv := x.(type) {
		case goja.Value:
			x = vv.Export()
		}
		s, is := x.(string)
		if !is {
			panic("not a string")
		}
		return url.QueryEscape(s)
	}

	if i.Testing {
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

	// match is a utility that invokes the pattern matcher.
	env["match"] = func(pat, mess, bs goja.Value) interface{} {
		var bindings core.Bindings

		if bs == nil {
			bindings = core.NewBindings()
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
			bindings = core.Bindings(m)
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

		bss, err := core.Match(nil, p, m, bindings)
		if err != nil {
			panic(err)
		}

		var x interface{}
		if x, err = canonicalize(bss); err != nil {
			panic(err)
		}

		return x
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

	v, err := o.RunProgram(p)
	cancel()

	if err != nil {
		if _, is := err.(*goja.InterruptedError); is {
			return nil, Interrupted
		}
		return nil, err
	}

	x := v.Export()

	var result core.Bindings
	switch vv := x.(type) {
	case *goja.InterruptedError:
		return nil, vv
	case map[string]interface{}:
		result = core.Bindings(vv)
	case core.Bindings:
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
