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
)

func TestTimersBasic(t *testing.T) {
	c := make(chan interface{})

	emitter := func(ctx context.Context, m interface{}) error {
		c <- m
		return nil
	}

	ts := NewTimers(emitter)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	then := time.Now()

	if err := ts.Add(ctx, "1", 1, time.Second); err != nil {
		t.Fatal(err)
	}

	if err := ts.Add(ctx, "1", 1, time.Second); err != Exists {
		t.Fatal(err)
	}

	if x := <-c; x != 1 {
		t.Fatal(x)
	}
	elapsed := time.Now().Sub(then)

	if 2*time.Second < elapsed {
		t.Fatal(elapsed)
	} else if elapsed < 990*time.Millisecond {
		t.Fatal(elapsed)
	}

	if err := ts.Add(ctx, "2", 2, time.Second); err != nil {
		t.Fatal(err)
	}

	if err := ts.Rem(ctx, "2"); err != nil {
		t.Fatal(err)
	}

	if err := ts.Rem(ctx, "2"); err != NotFound {
		t.Fatal(err)
	}

	timeout := time.NewTimer(1200 * time.Millisecond)
	select {
	case x := <-c:
		t.Fatal(x)
	case <-timeout.C:
	}

}

func TestTimersIdReuse(t *testing.T) {
	c := make(chan interface{})

	emitter := func(ctx context.Context, m interface{}) error {
		c <- m
		return nil
	}

	ts := NewTimers(emitter)

	d := 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ts.Add(ctx, "1", 1, d); err != nil {
		t.Fatal(err)
	}

	<-c

	if err := ts.Add(ctx, "1", 1, d); err != nil {
		t.Fatal(err)
	}

	<-c
}
