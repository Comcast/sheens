# Single-crew core

This single-crew approach is designed for running individual crews in
containers or [micro
VMs](https://github.com/firecracker-microvm/firecracker).

Can be coupled to MQTT, Redis, SQS, SNS, Kafka, `stdin`/`stdout`, etc.

See command-line programs [`siostd`](siostd) and [`siomq`](siomq).
[`mqshell`](mqshell) is a simple MQTT command-line, shell-like client
that's convenient for talking to `siomq` (or to MQTT brokers in
general).

## Timers

This code supports timers with a native machine `timers`.

Create a timer with a message like

```JSON
{"to":"timer","makeTimer":{"in":"IN","msg":MSG,"id":"ID"}}
```

where `IN` is a duration in [Go
syntax](https://golang.org/pkg/time/#ParseDuration).  For example,
`"10s"` means "10 seconds from now".

At the appointed time, the Sheens will receive `MSG`.

You can cancel a timer with a message like

```JSON
{"to":"timers","cancelTimer":"ID"}
```

## HTTP

If `Crew.EnableHTTP` is `true`, this crew can make HTTP request.
These requests are asynchronous.  The responses are returned as
in-bound messages.

```JSON
{"to":"http","httpRequest":{"url":"http://worldclockapi.com/api/json/est/now"},"replyTo":"thisMachine"}
```

For now, see [`http.go`](http.go) for the structure of the HTTP request.
