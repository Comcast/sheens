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
