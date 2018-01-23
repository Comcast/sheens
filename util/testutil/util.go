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
