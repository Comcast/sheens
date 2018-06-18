package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// handle considers messages for special treatment.
//
// This handler can deal with requests to make a timer and to cancel a
// timer.
func handle(ctx context.Context, message interface{}, ingest func(message interface{}) error) error {

	type MakeTimerRequest struct {
		MakeTimer *struct {
			Id      string      `json:"id"`
			In      string      `json:"in"`
			Message interface{} `json:"message"`
		} `json:"makeTimer"`
	}

	type DeleteTimerRequest struct {
		Id string `json:"deleteTimer"`
	}

	// Parse the message as a MakeTimerRequest or a
	// DeleteTimerRequest.  Sorry!
	js := []byte(JS(message))

	var makeRequest MakeTimerRequest
	if err := json.Unmarshal(js, &makeRequest); err != nil {
		return nil
	} else if makeRequest.MakeTimer != nil {
		in, err := time.ParseDuration(makeRequest.MakeTimer.In)
		if err != nil {
			return err
		}
		t := &timer{
			id:      makeRequest.MakeTimer.Id,
			in:      in,
			message: makeRequest.MakeTimer.Message,
			ingest:  ingest,
		}
		return start(ctx, t)
	}

	var deleteRequest DeleteTimerRequest
	if err := json.Unmarshal(js, &deleteRequest); err != nil {
		return err
	} else if deleteRequest.Id != "" {
		return stop(ctx, deleteRequest.Id)
	}

	return nil
}

// timer represents a message to be ingested in the future.
type timer struct {
	id      string
	in      time.Duration
	at      time.Time
	message interface{}
	ctl     chan bool
	ingest  func(interface{}) error
}

// timers is a map from timer ids to timers.
var timers = &sync.Map{}

// start starts a timer.
func start(ctx context.Context, t *timer) error {
	log.Printf("starting timer %s in %v", t.id, t.in)
	t.ctl = make(chan bool)
	t.at = time.Now().UTC().Add(t.in) // Just for debugging.

	if _, have := timers.LoadOrStore(t.id, t); have {
		return fmt.Errorf("timer '%s' exists", t.id)
	}

	go func() {
		tick := time.NewTimer(t.in)
		select {
		case <-tick.C:
			log.Printf("firing timer %s", t.id)
			t.ingest(t.message)
		case <-t.ctl:
			log.Printf("canceling timer %s", t.id)
		}
	}()

	return nil
}

// stop cancels a timer.
func stop(ctx context.Context, id string) error {
	x, have := timers.Load(id)
	if !have {
		return nil
	}
	t := x.(*timer)
	close(t.ctl)
	timers.Delete(id)
	return nil
}
