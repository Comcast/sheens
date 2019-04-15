# Single-crew core

This single-crew approach is designed for running individual crews in
containers or [micro
VMs](https://github.com/firecracker-microvm/firecracker).

Can be coupled to MQTT, Redis, SQS, SNS, Kafka, `stdin`/`stdout`, etc.

See command-line programs [`siostd`](siostd) and [`siomq`](siomq).
[`mqclient`](mqclient) is a simple MQTT command-line client that's
convenient for talking to `siomq` (or to MQTT brokers in general).

