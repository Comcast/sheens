# Crew that talks to an MQTT broker

Doesn't provide any persistence.

Also see [`mqshell`](../mqshell).

## Usage

Command-line arguments follow `mosquitto_sub`'s. Run `siomq -h` for
details.

For an example session, see `run.sh`, which expects an insecure MQTT
broker on 1883.  (See the _insecure_
[`../siost/mosquitto.conf`](mosquitto.conf) to use with `mosquitto -c
mosquitto.conf`.)


