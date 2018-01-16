package main

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/Comcast/sheens/core"
)

func TestService(t *testing.T) {
	if err := run(); err != nil {
		t.Fatal(err)
	}
}

func run() error {
	verbose := false

	render := func(tag string, m map[string]*core.Walked) {
		if !verbose {
			return
		}
		Render(tag, m)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	routed := make(chan interface{}, 1024)
	s, err := makeDemoService(ctx, routed, "", "", "")
	if err != nil {
		return err
	}

	defer func() {
		if err := s.Storage.Close(ctx); err != nil {
			log.Printf("warning storage.Close() error %v", err)
		}
	}()

	cid := "simpsons"
	mid := "doubler"

	if err = s.MakeCrew(ctx, cid); err != nil {
		// Maybe we just need to delete the Crew first!
		if err = s.RemCrew(ctx, cid); err != nil {
			return err
		}
		if err = s.MakeCrew(ctx, cid); err != nil {
			return err
		}
	}

	if err = s.AddMachine(ctx, cid, "double", mid, "", nil); err != nil {
		return err
	}

	if err = s.AddMachine(ctx, cid, "timers", "timers", "", nil); err != nil {
		return err
	}

	var p map[string]*core.Walked

	if p, err = s.Process(ctx, cid, Dwimjs(`{"double":1}`), nil); err != nil {
		return err
	}

	render("test", p)

	if p, err = s.Process(ctx, cid, Dwimjs(`{"makeTimer":{"in":"1s","id":"chips","message":{"double":10}}}`), nil); err != nil {
		return err
	}

	render("test", p)

	time.Sleep(2 * time.Second)

	if err = s.RemMachine(ctx, cid, mid); err != nil {
		return err
	}

	if p, err = s.Process(ctx, cid, Dwimjs(`{"double":1}`), nil); err != nil {
		return err
	}

	if err = s.RemCrew(ctx, cid); err != nil {
		return err
	}

	render("test", p)

	drain(routed)

	return nil
}

func drain(routed chan interface{}) {
	i := 0
LOOP:
	for {
		select {
		case message := <-routed:
			i++
			log.Printf("history: %d routed %s", i, JS(message))
		default:
			break LOOP
		}
	}
}
