#!/bin/bash

# Copyright 2019 Comcast Cable Communications Management, LLC
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

cat<<EOF | siostd -sh -ts -echo -pad -wait 8s -state-out state.json
{"to":"captain","update":{"c":{"spec":{"inline":<<cat ../../specs/collatz.yaml | yaml2json>>}}}}
{"to":"captain","update":{"dc":{"spec":{"inline":<<cat ../../specs/doublecount.yaml | yaml2json>>}}}}
{"to":"captain","update":{"d":{"spec":{"inline":<<cat ../../specs/double.yaml | yaml2json>>}}}}
{"to":"echo","like":"queso"}
{"double":10}
{"double":100}
{"double":1000}
{"collatz":17}
{"collatz":5}
{"to":"timers","makeTimer":{"in":"2s","id":"t0","msg":{"double":10000}}}
{"to":"timers","makeTimer":{"in":"3s","id":"t1","msg":{"collatz":21}}}
{"to":"timers","makeTimer":{"in":"4s","id":"t2","msg":{"to":"http","replyTo":"echo","httpRequest":{"url":"http://worldclockapi.com/api/json/est/now"}}}}
{"to":"d","double":3}
{"to":"dc","double":4}
{"to":"captain","update":{"echo":{"spec":{"inline":<<cat ../../specs/echo.yaml | yaml2json>>}}}}
EOF

echo

cat state.json | jq -c 'to_entries[]|([.key,.value.state])'
