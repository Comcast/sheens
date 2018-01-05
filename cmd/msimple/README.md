# Simple single-machine process

This demo runs a single machine that listens for input on `stdin`.
Outbound messages are written to `stdout`.


## Usage

### Basic

Using the spec [`double.yaml`](../../specs/double.yaml):

```Shell
echo '{"double":10}' | msimple -s ../../specs/double.yaml
```

Output:

```
{"doubled":20}
```


### Fancier

Using the spec [`test.yaml`](../../specs/test.yaml):

```Shell
cat<<EOF | msimple -s ../../specs/test.yaml -b '{"has":"beer","n":2}'
{"message":{"wants":"tacos"}}
{"message":{"wants":"tacos"}}
{"message":{"wants":"tacos"}}
EOF
```

Output:

```
{"remaining":1,"serve":"tacos","with":"beer"}
{"remaining":0,"serve":"tacos","with":"beer"}
{"remaining":0,"serve":"tacos","with":"water"}
```

With the `-d` (diagnostics flag) and `-e` (echo input flag), you get

```
in: {"message":{"wants":"tacos"}}
# walked
#   message    {"message":{"wants":"tacos"}}
#   00 from     {"node":"start","bs":{"has":"beer","n":2}}
#      to       {"node":"deliver","bs":{"?wants":"tacos","has":"beer","n":2}}
#      consumed {"message":{"wants":"tacos"}}
#   01 from     {"node":"deliver","bs":{"?wants":"tacos","has":"beer","n":2}}
#      to       {"node":"start","bs":{"has":"beer","n":1}}
#      emitted
#         {"remaining":1,"serve":"tacos","with":"beer"}
#   02 from     {"node":"start","bs":{"has":"beer","n":1}}
#      to       null
# next {"node":"start","bs":{"has":"beer","n":1}}
{"remaining":1,"serve":"tacos","with":"beer"}
# walked
#   message    {"remaining":1,"serve":"tacos","with":"beer"}
#   00 from     {"node":"start","bs":{"has":"beer","n":1}}
#      to       null
#      consumed {"remaining":1,"serve":"tacos","with":"beer"}
# next {"node":"start","bs":{"has":"beer","n":1}}
in: {"message":{"wants":"tacos"}}
# walked
#   message    {"message":{"wants":"tacos"}}
#   00 from     {"node":"start","bs":{"has":"beer","n":1}}
#      to       {"node":"deliver","bs":{"?wants":"tacos","has":"beer","n":1}}
#      consumed {"message":{"wants":"tacos"}}
#   01 from     {"node":"deliver","bs":{"?wants":"tacos","has":"beer","n":1}}
#      to       {"node":"change","bs":{"has":"beer","n":0}}
#      emitted
#         {"remaining":0,"serve":"tacos","with":"beer"}
#   02 from     {"node":"change","bs":{"has":"beer","n":0}}
#      to       {"node":"start","bs":{"has":"water","n":1}}
#   03 from     {"node":"start","bs":{"has":"water","n":1}}
#      to       null
# next {"node":"start","bs":{"has":"water","n":1}}
{"remaining":0,"serve":"tacos","with":"beer"}
# walked
#   message    {"remaining":0,"serve":"tacos","with":"beer"}
#   00 from     {"node":"start","bs":{"has":"water","n":1}}
#      to       null
#      consumed {"remaining":0,"serve":"tacos","with":"beer"}
# next {"node":"start","bs":{"has":"water","n":1}}
in: {"message":{"wants":"tacos"}}
# walked
#   message    {"message":{"wants":"tacos"}}
#   00 from     {"node":"start","bs":{"has":"water","n":1}}
#      to       {"node":"deliver","bs":{"?wants":"tacos","has":"water","n":1}}
#      consumed {"message":{"wants":"tacos"}}
#   01 from     {"node":"deliver","bs":{"?wants":"tacos","has":"water","n":1}}
#      to       {"node":"change","bs":{"has":"water","n":0}}
#      emitted
#         {"remaining":0,"serve":"tacos","with":"water"}
#   02 from     {"node":"change","bs":{"has":"water","n":0}}
#      to       {"node":"start","bs":{"has":"water","n":1}}
#   03 from     {"node":"start","bs":{"has":"water","n":1}}
#      to       null
# next {"node":"start","bs":{"has":"water","n":1}}
{"remaining":0,"serve":"tacos","with":"water"}
# walked
#   message    {"remaining":0,"serve":"tacos","with":"water"}
#   00 from     {"node":"start","bs":{"has":"water","n":1}}
#      to       null
#      consumed {"remaining":0,"serve":"tacos","with":"water"}
# next {"node":"start","bs":{"has":"water","n":1}}
```
