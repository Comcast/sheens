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

go install

cat<<EOF | siostd -sh -ts -echo -pad -wait 7s -state-out state.json 
{"to":"captain","update":{"c":{"spec":{"inline":<<cat ../../specs/collatz.yaml | yaml2json>>}}}}
{"to":"captain","update":{"dc":{"spec":{"inline":<<cat ../../specs/doublecount.yaml | yaml2json>>}}}}
{"to":"captain","update":{"d":{"spec":{"inline":<<cat ../../specs/double.yaml | yaml2json>>}}}}
{"double":10}
{"double":100}
{"double":1000}
{"collatz":17}
{"collatz":5}
{"to":"timers","makeTimer":{"in":"2s","msg":{"double":10000},"id":"t0"}}
{"to":"timers","makeTimer":{"in":"4s","msg":{"collatz":21},"id":"t1"}}
{"to":"d","double":3}
{"to":"dc","double":4}
EOF

echo

cat state.json | jq -c 'to_entries[]|([.key,.value.state])'
