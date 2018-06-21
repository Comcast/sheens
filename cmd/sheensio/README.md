# Simple Sheens I/O

This program aspires to be compatible with [Little
Sheens](https://github.com/Comcast/littlesheens)'s
[`sheensio`](https://github.com/Comcast/littlesheens/blob/master/sheensio.c).

Current the `steps` output is very different, but the `updated` and
`out` lines should be compatible (with the possible exception of JSON
property rendering order from Little Sheens).

## Usage

This program expects JSON specs, so you'll need to convert a spec in
YAML to a spec in JSON.
[`yaml2json`](https://github.com/bronze1man/yaml2json) can do that.

```Shell
go get github.com/bronze1man/yaml2json

cat specs/double.yaml | yaml2json | jq . > specs/double.js

cat<<EOF > crew.json
{"id":"simpsons",
 "machines":{
     "m1":{"spec":"double","node":"start","bs":{}}}}
EOF

echo '{"double":3}' | sheensio
```

produces

```JSON
out      {"doubled":6}
```

The JSON on the `out` lines is canonical (in the sense that properties
are sorted lexicographically).

Note that the order of processing among machines within a crew is
unspecified; therefore, the order of messages emitted by multiple
machines is unspecified.  If your crew has only one machine, then the
order of emitted messages is specified completely by that machine.
