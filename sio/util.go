/* Copyright 2019 Comcast Cable Communications Management, LLC
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

package sio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
)

// JS renders its argument as JSON or as '%#v'.
func JS(x interface{}) string {
	if x == nil {
		return "null"
	}
	js, err := json.Marshal(&x)
	if err != nil {
		return fmt.Sprintf("%#v", x)
	}
	return string(js)
}

// JSON renders its argument as pretty JSON or as '%#v".
func JSON(x interface{}) string {
	if x == nil {
		return "null"
	}
	js, err := json.MarshalIndent(&x, "", "  ")
	if err != nil {
		return fmt.Sprintf("%#v", x)
	}
	return string(js)
}

// JShort renders its argument as JS() but only up to 73 characters.
func JShort(x interface{}) string {
	js := []byte(JS(x))
	if 70 < len(js) {
		js = js[0:70]
		js = append(js, []byte("...")...)
	}
	return string(js)
}

var shell = regexp.MustCompile(`<<(.*?)>>`)

// ShellExpand expands shell commands delimited by '<<' and '>>'.  Use
// at your wown risk, of course!
func ShellExpand(msg string) (string, error) {
	literals := shell.Split(msg, -1)
	ss := shell.FindAllStringSubmatch(msg, -1)
	acc := literals[0]
	for i, s := range ss {
		var sh = s[1]
		cmd := exec.Command("bash", "-c", sh)
		// cmd.Stdin = strings.NewReader("")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			return "", fmt.Errorf("shell error %s on %s", err, sh)
		}
		got := out.String()
		acc += got
		acc += literals[i+1]
	}
	return acc, nil
}
