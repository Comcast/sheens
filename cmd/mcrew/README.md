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
(cd ../.. && mcrew -v -t :9000) &
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

## Demo: Home Assistant

This code can be used for [Home Assistant](https://home-assistant.io/)
automation via its [WebSockets
API](https://home-assistant.io/developers/websocket_api/).

See

1. [`ha-boot.json`](ha-boot.json)
1. [`specs/homeassistant.yaml`](../../specs/homeassistant.yaml)
1. [`specs/ha-switches.yaml`](../../specs/ha-switches.yaml)

The file [`ha-boot.json`](ha-boot.json) expects switches `switch.lamp`
and `switch.shed_light`. Edit accordingly.  When you use Home
Assistant to turn on `switch.lamp`, `switch.shed_light` should turn
on.

```Shell
export SHEEN_HA_PASSWORD=mypassword
mcrew -c ws://HAHOST:8123/api/websocket -t :9000 -s "../../specs" -b ha-boot.json
```


## Features

### Using environment variables for WebSocket output

This demo code will use environment variables of the form `SHEEN_*`
that are referenced, with a `$` prefix, in messages sent to the
WebSocket client connection via `"to":"ws"`.  See
[`specs/homeassistant.yaml`](../../specs/homeassistant.yaml).  A
message given to the WebSocket client has strings of the form
`$SHEEN_*` expanded to their corresponding environment variable values
(`SHEEN_*`).  This crude mechanism is used in the demo Home Assistant
machine to pass a password from the process's environment to Home
Assistant.

### Timers service

Messages like

```JSON
{"to":"timers","makeTimer":{"in":"1s","id":"1","message":{"to":"doubler","double":100}}}}}}
```

will create a timer.  The duration (`"in":`) is expressed in Go
[`time.Duration` syntax](https://golang.org/pkg/time/#ParseDuration).

At the appointed time, the process will send the given message to the
crew.

The given `id` can be used to try to cancel the timer (assuming it
hasn't fired):

```JSON
{"to":"timers","deleteTimer":"1"}
```

### HTTP service

Messages like

```JSON
{"to":"http","request":{"url":"%s", ...},"replyTo":"machine42"}
```

will result in an asychronous HTTP request.  For now, see
`HTTPRequest` in [`http.go`](http.go) for the supported request
structure.  The response is submitted as a message to the crew or to
the `replyTo` machine given in the request.  For now, see
`HTTPResponse` in [`http.go`](http.go) for the supported response
structure.
