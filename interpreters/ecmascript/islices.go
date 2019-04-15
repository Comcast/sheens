/* Copyright 2018-2019 Comcast Cable Communications Management, LLC
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


package ecmascript

import (
	"reflect"
)

// iSlice will convert reflect.Slices to actual slices.
//
// Sometimes Match is given a reflect.Slice instead of a plain old
// slice.
func iSlice(xs interface{}) (interface{}, bool) {
	v := reflect.ValueOf(xs)
	switch v.Kind() {
	case reflect.Slice:
		acc := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			acc[i] = v.Index(i).Interface()
		}
		return acc, true
	}
	return v, false
}
