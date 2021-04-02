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

package main

import (
	"encoding/json"
	"fmt"
)

// JS is provided here for DEMO purposes ONLY, use testutil package JS func
// for all other cases
func JS(x interface{}) string {
	js, err := json.Marshal(&x)
	if err != nil {
		panic(err)
	}
	return string(js)
}

func warn(err error) {
	fmt.Printf("warning: %s\n", err)
}
