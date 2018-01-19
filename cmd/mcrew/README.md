# A demo process that hosts a "crew" of machines

This demo process is serves a single "crew" of machines.

See [`protocol.go`](protocol.go) for the protocol, and see
[`input.txt`](input.txt) for a simple example.


## Usage

### Via stdin/out

From this directory:

```Shell
cat input.txt | (cd ../.. && mcrew -I -O)
```

### Via TCP

From this directory:

```Shell
(cd ../.. && | -t :9000) &
cat input.txt | nc localhost 9000
```

### Via WebSockets

From this directory:

```Shell
(cd ../.. && mcrew -h :8080 -w -f cmd/mcrew/httpfiles)
```

Then try [`localhost:8080/static/demo.html`](http://localhost:8080/static/demo.html).


### Via HTTP

ToDo.
