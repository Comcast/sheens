/* Copyright 2018 Comcast Cable Communications Management, LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package testutil

import (
	"encoding/json"
	"fmt"
	"log"
)

// JS renders its argument as JSON or as a string indicating an error.
func JS(x interface{}) string {
	bs, err := json.Marshal(&x)
	if err != nil {
		log.Printf("warning: testutil.JS error %s for %#v", err, x)
		return fmt.Sprintf("%#v", x)
	}
	return string(bs)
}

// Dwimjs, when given a string or bytes, parses that data as JSON.
// When given anything else, just returns what's given.
//
// See https://en.wikipedia.org/wiki/DWIM.
func Dwimjs(x interface{}) interface{} {
	switch vv := x.(type) {
	case []byte:
		return Dwimjs(string(vv))
	case string:
		var v interface{}
		if err := json.Unmarshal([]byte(vv), &v); err != nil {
			panic(err)
		}
		return v
	default:
		return x
	}
}
