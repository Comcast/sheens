/* Copyright 2019 Comcast Cable Communications Management, LLC
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

package sio

// ToDo: Timers.Suspend, Timers.Resune

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/match"
)

var (
	// TimersMachine is the id of the timers machine.
	TimersMachine = "timers"
)

// TimerEntry represents a pending timer.
type TimerEntry struct {
	Id  string
	Msg interface{}
	At  time.Time
	Ctl chan bool `json:"-"`

	timers *Timers
}

// Timers represents pending timers.
type Timers struct {
	Map     map[string]*TimerEntry
	Emitter func(context.Context, *TimerEntry) `json:"-"`

	sync.Mutex

	c *Crew
}

// NewTimers creates a Timers with the given function that the
// TimerEntries will use to emit their messages.
func NewTimers(emitter func(context.Context, *TimerEntry)) *Timers {
	return &Timers{
		Map:     make(map[string]*TimerEntry, 8),
		Emitter: emitter,
	}
}

// withMap populates the Timers based on the given raw map.
//
// This method is used to initialize the Timers from a timers
// machine's state, which is generated via Timers.State().
func (ts *Timers) withMap(x interface{}) error {
	js, err := json.Marshal(&x)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(js, &ts.Map); err != nil {
		return err
	}
	for _, te := range ts.Map {
		te.timers = ts
		te.Ctl = make(chan bool)
	}

	return nil
}

// State creates a machine state that Timers.withMap can use.
func (ts *Timers) State() *core.State {
	return &core.State{
		Bs: match.NewBindings().Extend("timers", ts.Map),
	}
}

// Start starts all known timers.
//
// Call this method when your have just created a Timers with existing
// data.
func (ts *Timers) Start(ctx context.Context) error {
	ts.c.Logf("Timers.Start")
	for _, t := range ts.Map {
		go t.run(ctx)
	}
	return nil
}

func (ts *Timers) add(ctx context.Context, e *TimerEntry) error {
	if _, have := ts.Map[e.Id]; have {
		return ts.cancel(ctx, e.Id)
	}

	ts.Map[e.Id] = e
	e.timers = ts
	ts.changed()

	go e.run(ctx)

	return nil
}

// Add creates a new Timer that will emit the given message later (if
// the timer isn't cancelled first).
func (ts *Timers) Add(ctx context.Context, id string, msg interface{}, d time.Duration) error {
	ts.c.Logf("Timers.Add %s", id)

	ts.Lock()

	e := &TimerEntry{
		Id:     id,
		At:     time.Now().UTC().Add(d),
		Msg:    msg,
		Ctl:    make(chan bool),
		timers: ts,
	}

	ts.add(ctx, e)

	ts.Unlock()

	return nil
}

// run starts a timer that will execute the TimerEntry at the
// appointed time if the TimerEntry isn't cancelled first.
func (te *TimerEntry) run(ctx context.Context) error {
	te.timers.c.Logf("TimerEntry %s run", te.Id)

	t := time.NewTimer(te.At.Sub(time.Now()))
	select {
	case <-t.C:
		te.timers.c.Logf("Firing timer '%s'", te.Id)
		te.timers.Emitter(ctx, te)
		te.timers.Lock()
		delete(te.timers.Map, te.Id)
		te.timers.Unlock()
		te.timers.c.Lock()
		te.timers.changed()
		te.timers.c.Unlock()
	case <-te.Ctl:
		te.timers.c.Logf("Canceling timer '%s'", te.Id)
	case <-ctx.Done():
	}
	return nil
}

func (ts *Timers) changed() {
	ts.c.change(TimersMachine).State = ts.State()
}

func (ts *Timers) cancel(ctx context.Context, id string) error {
	ts.c.Logf("Timers.cancel %s", id)

	t, have := ts.Map[id]
	if !have {
		return fmt.Errorf("timer '%s' doesn't exist", id)
	}
	delete(ts.Map, id)
	ts.changed()

	close(t.Ctl)

	return nil
}

// Cancel attepts to cancel the timer with the given id.
func (ts *Timers) Cancel(ctx context.Context, id string) error {
	ts.Lock()
	err := ts.cancel(ctx, id)
	ts.Unlock()
	return err
}
