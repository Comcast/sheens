package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Comcast/sheens/core"
	. "github.com/Comcast/sheens/util/testutil"
)

func TestHTTPGlue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello")
	}))
	defer ts.Close()

	s, err := NewService(ctx, ".", "", ".")
	if err != nil {
		t.Fatal(err)
	}
	s.Emitted = make(chan interface{}, 8)
	s.Processing = make(chan interface{}, 8)
	s.Errors = make(chan interface{}, 8)

	msg := Dwimjs(fmt.Sprintf(`{"to":"http", "request":{"url":"%s"}}`, ts.URL))

	op := COp{
		Process: &OpProcess{
			Message: msg,
		},
	}

	if err = op.Do(ctx, s); err != nil {
		t.Fatal(err)
	}

	timeout := time.NewTimer(time.Second)

	for {
		select {
		case <-timeout.C:
			t.Fatal("timeout")
		case x := <-s.Processing:
			bss, err := core.Match(nil, Dwimjs(`{"body":"hello"}`), x, core.NewBindings())
			if err != nil {
				t.Fatal(err)
			}
			if 0 < len(bss) {
				return
			}
		}
	}

}
