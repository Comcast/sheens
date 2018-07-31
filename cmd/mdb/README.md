# A very simple crew debugger

## Usage

Run `mdb`.  By default, specs are loaded from the current directory.

Example session:

```Shell
cat input.txt | mdb -s ../../specs -e
```

gives

```
set doubler spec doublecount.yaml
# crew now has 1 machines
print
# machine doubler:
#   node:     start
#   bindings: {}
#   spec:     doublecount.yaml
run {"double":3}
# Walkeds  (1 machines)
# Machine doubler
#   00 from     {"node":"start","bs":{}}
#      to       {"node":"listen","bs":{"count":0}}
#   01 from     {"node":"listen","bs":{"count":0}}
#      to       {"node":"process","bs":{"?n":3,"count":0}}
#      consumed {"double":3}
#   02 from     {"node":"process","bs":{"?n":3,"count":0}}
#      to       {"node":"listen","bs":{"count":1}}
#      emitted
#         {"doubled":6}
#   03 from     {"node":"listen","bs":{"count":1}}
#      to       null
#   stopped     Done
# queue has 1 messages
print doubler
#   node:     listen
#   bindings: {"count":1}
#   spec:     doublecount.yaml
printqueue
# 0. {"doubled":6}
pop
# processing {"doubled":6}
# Walkeds  (1 machines)
# Machine doubler
#   00 from     {"node":"listen","bs":{"count":1}}
#      to       null
#      consumed {"doubled":6}
#   stopped     Done
# queue has 0 messages
printqueue
# queue is empty
print doubler
#   node:     listen
#   bindings: {"count":1}
#   spec:     doublecount.yaml
printqueue
# queue is empty
run {"double":4}
# Walkeds  (1 machines)
# Machine doubler
#   00 from     {"node":"listen","bs":{"count":1}}
#      to       {"node":"process","bs":{"?n":4,"count":1}}
#      consumed {"double":4}
#   01 from     {"node":"process","bs":{"?n":4,"count":1}}
#      to       {"node":"listen","bs":{"count":2}}
#      emitted
#         {"doubled":8}
#   02 from     {"node":"listen","bs":{"count":2}}
#      to       null
#   stopped     Done
# queue has 1 messages
print doubler
#   node:     listen
#   bindings: {"count":2}
#   spec:     doublecount.yaml
set doubler bs {"count":0}
print doubler
#   node:     listen
#   bindings: {"count":0}
#   spec:     doublecount.yaml
run {"double":5}
# Walkeds  (1 machines)
# Machine doubler
#   00 from     {"node":"listen","bs":{"count":0}}
#      to       {"node":"process","bs":{"?n":5,"count":0}}
#      consumed {"double":5}
#   01 from     {"node":"process","bs":{"?n":5,"count":0}}
#      to       {"node":"listen","bs":{"count":1}}
#      emitted
#         {"doubled":10}
#   02 from     {"node":"listen","bs":{"count":1}}
#      to       null
#   stopped     Done
# queue has 2 messages
help
# 
#   set ID spec FILENAME       Set the spec for the machine with that ID
#   set ID node NODENAME       Set the node for the machine with that ID
#   set ID bindings BINDINGS   Set the bindings (JSON) for the machine with that ID
#   rem ID                     Remove the machine with that ID
#   print [ID]                 Print the state of the machine with that ID
#   run [MSG]                  Run the crew.
#   printqueue                 Show the queue of emitted messages
#   pop                        Send first message in the queue to the crew
#   drop                       Drop the first message in the queue
#   save FILENAME              Save the crew machines to this file
#   load FILENAME              Load the crew machines from this file
#   help                       Show this documentation
# 
```

## Action intepreters

Just the demo ECMAscript (Goja-based) interpreter (via `ecmascript` or
`goja`).


## Routing

A message containing a property `"to":"ID"` is send directly the the
machine with that id (if any).


## Protocol environment

None.  In particular, doesn't support timers or HTTP requests or
anything.


## ToDo

1. User-supplied custom routing function defined by a Javascript
   function.
1. `core.Control` control.

