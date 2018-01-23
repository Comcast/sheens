package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Comcast/sheens/core"
	. "github.com/Comcast/sheens/util/testutil"
)

func Copy(x interface{}) interface{} { // Sorry
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

func Render(w io.Writer, tag string, m map[string]*core.Walked) {
	fmt.Fprintf(w, "Walkeds %s (%d machines)\n", tag, len(m))
	for mid, walked := range m {
		fmt.Fprintf(w, "%s\n", mid)
		for i, stride := range walked.Strides {
			fmt.Fprintf(w, "  %02d from     %s\n", i, JS(stride.From))
			fmt.Fprintf(w, "     to       %s\n", JS(stride.To))
			if stride.Consumed != nil {
				fmt.Fprintf(w, "     consumed %s\n", JS(stride.Consumed))
			}
			if 0 < len(stride.Events.Emitted) {
				fmt.Fprintf(w, "     emitted\n")
			}
			for _, emitted := range stride.Events.Emitted {
				fmt.Fprintf(w, "        %s\n", JS(emitted))
			}
		}
		if walked.Error != nil {
			fmt.Fprintf(w, "  error    %v\n", walked.Error)
		}
		fmt.Fprintf(w, "  stopped     %v\n", walked.StoppedBecause)
	}
}

type WrappedError struct {
	Outer error `json:"outer"`
	Inner error `json:"inner"`
}

func (e *WrappedError) Error() string {
	return e.Outer.Error() + " after " + e.Inner.Error()
}

func NewWrappedError(outer, inner error) error {
	if inner == nil {
		return outer
	}
	return &WrappedError{
		Outer: outer,
		Inner: inner,
	}
}
