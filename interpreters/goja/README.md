# The Goja(-based) Intepreter

This `core.Interpreter` is based on
[Goja](https://github.com/dop251/goja), which interprets Ecmascript
5.1.

An action should always `return` updated bindings, and guard should
always `return` either updated bindings or `null`.


Feel free to make your own `core.Interpreter` that does what you want,
of course

## Libraries

If the "source" is given as map with keys `requires` and `code`, as in

```YAML
interpreter: goja
source:
  requires:
  - 'file://interpreters/goja/libs/time.js'
  code: |-
    return isCurrent(_.bindings.during) ? _.bindings : null;
```

then the given libraries will be inlined in the order given.


## Environment

The runtime environment includes a binding for the variable `_`:

1. `_.bindings`: The current set of machine bindings.

1.  `_.params`: The current parameters for the machine execution.
    These parameters are provided by the application that uses the
    Machines `core`.  In the example `mservice` process, these
    parameters include
   
    1. `mid`: The id of the current machine
	1. `cid`: The id of the machine's crew

1. `_.genstr()`: Generates a random 32-char string

1. `_.esc(STRING)`: Calls [url.QueryEscape](https://golang.org/pkg/net/url/#QueryEscape).

1. `_.out(X)`: "Emits" the given message.

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

1. `_.sleep(MS)`: Sleeps for the given number of milliseconds.  Only
   available if the interpreter's `Testing` property is true.

1. `_.exit(N)`: Terminates the process (!) with the given exit code.
   Only available if the interpreter's `Testing` property is true.

