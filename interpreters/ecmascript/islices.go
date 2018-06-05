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
