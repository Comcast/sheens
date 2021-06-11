# Simple (?) command-line MQTT client

Command-line args follow those for `mosquitto_sub`.  Run `mqclient -h`
for help.

## Commands

1. `qos QOS`: Set the QoS for subsequent operations.
1. `sub TOPIC`: Subscribe to the given topic.
1. `unsub TOPIC`: Unsubscribe from the given topic.
1. `retain (true|false)`:  Set retain flag for subsequent pubs.
1. `pub TOPIC MSG`: Publish MSG to the given TOPIC.
1. `kill`: Ungracefully close the connection (without the disconnect protocol).
1. `sleep DURATION`: Sleep for the given [duration](https://golang.org/pkg/time/#ParseDuration)(e.g, '1s').
1. `js CODE`: Run the given Javascript!  See below.
1. `jsfile FILENAME`: Run the Javascript code in the given file.

By default (controlled by `-sh`), input lines are shell-expanded: Each
substring of the form `<<SHELL_COMMAND>>` is replaced by the `stdout`
of executing `SHELL_COMMAND`.

## Javascript

Yes, there's an embedded Javascript interpreter.

Some available functions:

1. `print`
1. `subscribe(topic, qos, handler)`, where `hander` is a `function(topic,payload)`
1. `unsubscribe(topic)`
1. `publish(topic,qos,retained,msg)`
1. `setInterval`
1. `setTimeout`
