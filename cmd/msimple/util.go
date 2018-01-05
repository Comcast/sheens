package main

import (
	"encoding/json"
	"fmt"
)

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
