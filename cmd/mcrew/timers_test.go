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
