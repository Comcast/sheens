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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Comcast/sheens/match"
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
			bss, err := match.Match(Dwimjs(`{"body":"hello"}`), x, match.NewBindings())
			if err != nil {
				t.Fatal(err)
			}
			if 0 < len(bss) {
				return
			}
		}
	}

}
