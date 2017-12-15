package main

import (
	"encoding/json"
	"fmt"

	"github.com/Comcast/sheens/core"
)

func Copy(x interface{}) interface{} {
	js, err := json.Marshal(&x)
	if err != nil {
		panic(err)
	}
	var y interface{}
	if err = json.Unmarshal(js, &y); err != nil {
		panic(err)
	}
	return y
}

func JS(x interface{}) string {
	js, err := json.Marshal(&x)
	if err != nil {
		panic(err)
	}
	return string(js)
}

func Dwimjs(x interface{}) interface{} {
	switch vv := x.(type) {
	case []byte:
		var y interface{}
		if err := json.Unmarshal(vv, &y); err != nil {
			panic(err)
		}
		return y
	case string:
		return Dwimjs([]byte(vv))
	default:
		return x
	}
}

func Render(tag string, m map[string]*core.Walked) {
	fmt.Printf("March %s\n", tag)
	for mid, walked := range m {
		fmt.Printf("%s\n", mid)
		for i, stride := range walked.Strides {
			fmt.Printf("  %02d from     %s\n", i, JS(stride.From))
			fmt.Printf("     to       %s\n", JS(stride.To))
			if stride.Consumed != nil {
				fmt.Printf("     consumed %s\n", JS(stride.Consumed))
			}
			if 0 < len(stride.Events.Emitted) {
				fmt.Printf("     emitted\n")
			}
			for _, emitted := range stride.Events.Emitted {
				fmt.Printf("        %s\n", JS(emitted))
			}
		}
		if walked.Error != nil {
			fmt.Printf("  error    %v\n", walked.Error)
		}
		fmt.Printf("  stopped     %v\n", walked.StoppedBecause)
	}
}
