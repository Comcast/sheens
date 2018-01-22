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
