# Javascript Sheens implementation

Ported from
[`github.com/Comcast/littlesheens`](https://github.com/Comcast/littlesheens),
which wraps a C API around this Javascript implementation.

This code works with [Duktape](https://duktape.org/), which can give
you a full Sheens implementation in 400KB (along with a decent sandbox
for your actions).

## Demo usage

```Shell
make test
```

You'll need
[`minify`](https://github.com/tdewolff/minify/tree/master/cmd/minify)
to use the `Makefile` as it's currently written.

```
./demo double.js demo.js
reading 'double.js'
read 1153 bytes from 'double.js'
true
reading 'demo.js'
read 925 bytes from 'demo.js'
state 0 {"bs":{"count":0},"node":"start"}
stepping 0 {"double":1}
stepped 0 {"to":{"node":"listen","bs":{"count":1}},"consumed":true,"emitted":[{"doubled":2}]}
0 0 emitted {"doubled":2}
state 1 {"bs":{"count":1},"node":"listen"}
stepping 1 {"double":10}
stepped 1 {"to":{"node":"listen","bs":{"count":2}},"consumed":true,"emitted":[{"doubled":20}]}
1 0 emitted {"doubled":20}
state 2 {"bs":{"count":2},"node":"listen"}
stepping 2 {"double":100}
stepped 2 {"to":{"node":"listen","bs":{"count":3}},"consumed":true,"emitted":[{"doubled":200}]}
2 0 emitted {"doubled":200}
true
```
