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
	"testing"
	"time"

	. "github.com/Comcast/sheens/util/testutil"
)

func TestTimersGlue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := NewService(ctx, ".", "", ".")
	if err != nil {
		t.Fatal(err)
	}
	s.Emitted = make(chan interface{}, 8)
	s.Processing = make(chan interface{}, 8)
	s.Errors = make(chan interface{}, 8)

	{

		msg := Dwimjs(`{"to":"timers", "makeTimer":{"in":"1s","id":"1","message":"hello"}}`)
		op := COp{
			Process: &OpProcess{
				Message: msg,
			},
		}

		if err = op.Do(ctx, s); err != nil {
			t.Fatal(err)
		}

		timeout := time.NewTimer(2 * time.Second)

	LOOP1:
		for {
			select {
			case <-timeout.C:
				t.Fatal("timeout")
			case err := <-s.Errors:
				t.Fatal(err)
			case op := <-s.Processing:
				if op == "hello" {
					break LOOP1 // Happy
				}
			case <-s.Emitted:
			case <-ctx.Done():
				return
			}
		}
	}

	{
		msg := Dwimjs(`{"to":"timers", "makeTimer":{"in":"1s","id":"2","message":"hello"}}`)
		op := COp{
			Process: &OpProcess{
				Message: msg,
			},
		}

		if err = op.Do(ctx, s); err != nil {
			t.Fatal(err)
		}

		msg = Dwimjs(`{"to":"timers", "deleteTimer":"2"}`)
		op = COp{
			Process: &OpProcess{
				Message: msg,
			},
		}

		if err = op.Do(ctx, s); err != nil {
			t.Fatal(err)
		}
		timeout := time.NewTimer(2 * time.Second)

	LOOP2:
		for {
			select {
			case <-timeout.C:
				break LOOP2 // Happy
			case err := <-s.Errors:
				t.Fatal(err)
			case op := <-s.Processing:
				if op == "hello" {
					t.Fatal(op)
				}
			case <-s.Emitted:
			case <-ctx.Done():
				return
			}
		}
	}
}
