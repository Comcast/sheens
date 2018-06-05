# ECMAScript

This `core.Interpreter` is based on
[Goja](https://github.com/dop251/goja), which executes Ecmascript
5.1.

Unlike the demo `goja` interpreter (`interpreters/goja`), this
interpreter doesn't support libraries.  If you want to use this
interpreter but also want to use libraries, then you should implement
a preprocessing step that generates source that has the libraries
embedded in that source.

Eventually deprecate `interpreters/goja` in favor of this interpreter.

## Environment

The base runtime environment includes a binding for the variable `_`:

1. `_.bindings`: The current set of machine bindings.

1.  `_.params`: The current parameters for the machine execution.
    These parameters are provided by the application that uses the
    Machines `core`.  In the example `mservice` process, these
    parameters include
   
    1. `mid`: The id of the current machine
	1. `cid`: The id of the machine's crew

1. `_.props`: `core.StepProps` (if any).

1. `_.out(X)`: "Emits" the given message.


If the interpreter's `Extended` flag is `true`, then `_` has these
additional properties:

1. `_.genstr()`: Generates a random 32-char string

1.  `_.cronNext(CRONEXPR)` → `TIMESTAMP`: `cronNext` attempts to parse
    its argument as a
    [cron expression](https://github.com/gorhill/cronexpr). If
    successful, returns the next time in
    [Go RFC3339Nano](https://golang.org/pkg/time/#pkg-constants)
    format.
   
    Example:
	
	```Javascript
	({next: _.cronNext("* 0 * * *")});
	```

1. `_.match(PATTERN, MESSAGE)`: Invokes pattern matching.

If the interpreter's `Test` flag is `true`, then `_` has these
additional properties:

1. `_.sleep(MS)`: Sleeps for the given number of milliseconds.

1. `_.exit(N)`: Terminates the process (!) with the given exit code.

