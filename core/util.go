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

package core

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"
)

// alphabet is used by Gensym.
var alphabet = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Gensym makes a random string of the given length.
//
// Since we're returning a string and not (somehow a symbol), should
// be named something else.  Using this name just brings back good
// memories.
func Gensym(n int) string {
	bs := make([]byte, n)
	for i := 0; i < len(bs); i++ {
		bs[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(bs)
}

// // StringMaps recursively converts map[interface{}]interface{}] to
// // map[string]interface{}.
// //
// // Recursively processes values.
// //
// // Warning: Doesn't work through structs.
// //
// // Had to go to this trouble because the YAML deserializer likes to
// // make map[interface{}] instead of map[string].
// func StringMaps(x interface{}) (interface{}, error) {
// 	switch vv := x.(type) {
// 	case map[interface{}]interface{}:
// 		m := make(map[string]interface{}, len(vv))
// 		for thing, val := range vv {
// 			s, is := thing.(string)
// 			if !is {
// 				// return nil, fmt.Errorf("%#v (%T) isn't a %T", thing, thing, s)
// 				return nil, errors.New("stringMaps encountered a non-string key")
// 			}
// 			val, err := StringMaps(val)
// 			if err != nil {
// 				return nil, err
// 			}
// 			m[s] = val
// 		}
// 		return m, nil
// 	case map[string]interface{}:
// 		for s, val := range vv {
// 			val, err := StringMaps(val)
// 			if err != nil {
// 				return nil, err
// 			}
// 			vv[s] = val
// 		}
// 		return vv, nil
// 	case []interface{}:
// 		for i, x := range vv {
// 			y, err := StringMaps(x)
// 			if err != nil {
// 				return nil, err
// 			}
// 			vv[i] = y
// 		}
// 		return vv, nil
// 	default:
// 		return x, nil
// 	}
// }

// Canonicalize is ... hey, look over there!
func Canonicalize(x interface{}) (interface{}, error) {
	var err error

	js, err := json.Marshal(&x)
	if err != nil {
		return nil, err
	}
	var y interface{}
	if err = json.Unmarshal(js, &y); err != nil {
		return nil, err
	}

	return y, nil
}

// Timestamp returns a string representing the current time in
// RFC3339Nano.
func Timestamp() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// Unquestion removes (so to speak) a leading question mark (if any).
func Unquestion(p string) string {
	if strings.HasPrefix(p, "?") {
		return p[1:]
	}
	return p
}
