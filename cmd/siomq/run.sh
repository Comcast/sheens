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

# This file might have changed after the fork.

set -e

go install && (cd ../mqclient && go install)

siomq -t foo &

sleep 1

cat<<EOF | mqclient
sub misc
pub foo {"to":"captain","update":{"c":{"spec":{"inline":<<cat ../../specs/collatz.yaml | yaml2json>>}}}}
pub foo {"to":"captain","update":{"dc":{"spec":{"inline":<<cat ../../specs/doublecount.yaml | yaml2json>>}}}}
pub foo {"to":"captain","update":{"d":{"spec":{"inline":<<cat ../../specs/double.yaml | yaml2json>>}}}}
pub foo {"double":10}
pub foo {"double":100}
pub foo {"double":1000}
pub foo {"collatz":17}
pub foo {"collatz":5}
pub foo {"to":"timers","makeTimer":{"in":"2s","msg":{"double":10000},"id":"t0"}}
pub foo {"to":"timers","makeTimer":{"in":"4s","msg":{"collatz":21},"id":"t1"}}
pub foo {"to":"d","double":3}
pub foo {"to":"dc","double":4}
sleep 5s
EOF

kill %1
