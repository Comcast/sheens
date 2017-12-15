# Example Crew service

This package contains a demo Machines service.

This service is set up mostly in `service.go`, which implements two
APIs: a "control plane" API via HTTP and a "data plane" API via plain
TCP.  The APIs are actually the same; however, the control plane is
intended for managing things (crews, machines), and the data plane is
intended for submitting and getting messages.

Specs are stored on the file system.

## Demo 1: Basics

For convenience, we'll just use the TCP API for all our calls.

```Shell
go install && mservice &
cat input.txt | nc localhost 8081
```

## Demo 2: Making an HTTP request

Don't take this demo too seriously.  Better to have a real HTTP
request service.

To see a demo of an HTTP request, try

```Shell
cat http.txt | nc localhost 8081
```

## Demo 3: Silly anomaly detection

```Shell
cat anomalytoy.txt | nc localhost 8081
```

## ToDo

1. Don't write out state that isn't really any different.
