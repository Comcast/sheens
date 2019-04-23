# Single crew I/O

This program runs a single crew that can talk via

1. `stdin`/`stdout` with `-io std` or as the default
2. MQTT with `-io mq`
3. WebSockets with `-io ws`

Use `-h` to see command-line arguments.

## Example

```Shell
cat input.json | (cd ../.. && sio -wait 10s)
```

Also see [`dewo.sh`](demo.sh).

