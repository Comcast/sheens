package main

// ToDo: Timers.Suspend, Timers.Resune

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	. "github.com/Comcast/sheens/util/testutil"
)

type Emitter func(ctx context.Context, message interface{}) error

var (
	Exists   = errors.New("id exists")
	NotFound = errors.New("not found")
)

type TimerEntry struct {
	Id      string      `json:"id"`
	Message interface{} `json:"message"`
	At      time.Time   `json:"at"`

	ctl chan bool
}

type Timers struct {
	Errors chan interface{} `json:"-" yaml:"-"`

	sync.Mutex

	timers map[string]*TimerEntry `json:"map"`
	ctl    chan bool
	emit   Emitter
}

func NewTimers(emitter Emitter) *Timers {
	return &Timers{
		timers: make(map[string]*TimerEntry, 32),
		emit:   emitter,
		ctl:    make(chan bool),
	}
}

func (ts *Timers) MarshalJSON() ([]byte, error) {
	ts.Lock()
	m := map[string]interface{}{
		"map": ts.timers,
	}
	bs, err := json.Marshal(&m)
	ts.Unlock()
	return bs, err
}

func (ts *Timers) MarshalYAML() (interface{}, error) {
	ts.Lock()
	cp := Copy(map[string]interface{}{
		"map": ts.timers,
	})
	ts.Unlock()
	return cp, nil
}

func (ts *Timers) Add(ctx context.Context, id string, message interface{}, in time.Duration) error {
	ts.Lock()
	defer ts.Unlock()

	if _, have := ts.timers[id]; have {
		return Exists
	}

	te := &TimerEntry{
		Id:      id,
		Message: message,
		At:      time.Now().UTC().Add(in),
		ctl:     make(chan bool),
	}

	ts.timers[id] = te

	stop := func() {
		if err := ts.Rem(ctx, id); err != nil {
			ts.err(fmt.Errorf("Timers rem error %v id=%s", err, id))

		}
	}

	go func() {
		timer := time.NewTimer(te.At.Sub(time.Now()))
		select {
		case <-ctx.Done():
			stop()
		case <-te.ctl:
			// We only get here via a Rem() call.
		case <-ts.ctl:
			stop()

			// Not exactly what we want ...
		case <-timer.C:
			Logf("Timers firing %s", JS(ts))
			if err := ts.emit(ctx, te.Message); err != nil {
				ts.err(fmt.Errorf("Timers emit error %v id=%s", err, id))
			}

			// See https://github.com/Comcast/sheens/issues/19
			ts.Lock()
			delete(ts.timers, id)
			ts.Unlock()
		}
	}()

	return nil
}

func (ts *Timers) Shutdown() error {
	close(ts.ctl)
	return nil
}

func (ts *Timers) Rem(ctx context.Context, id string) error {
	ts.Lock()
	defer ts.Unlock()

	te, have := ts.timers[id]
	if !have {
		return NotFound
	}

	delete(ts.timers, id)

	close(te.ctl)

	return nil
}

func (ts *Timers) err(err error) {
	if ts.Errors != nil {
		ts.Errors <- err
	} else {
		log.Println(err)
	}
}
