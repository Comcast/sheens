package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	. "github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
)

type emitter func(ctx context.Context, message interface{}) error

type timerEntry struct {
	Id      string      `json:"id"`
	Message interface{} `json:"message"`
	At      time.Time   `json:"at"`

	// Ctx: Probably shouldn't have this field!
	Ctx context.Context `json:"-" yaml:"-"`

	Cancel func() `json:"-" yaml:"-"`
}

var (
	IdExists = errors.New("id exists")
	NotFound = errors.New("not found")
)

type timers struct {
	Debug bool `json:"-"`

	// See MarshalJSON, which is defined in order to obtain the
	// lock before permitting marshalling.

	sync.Mutex

	// Id will also be the id of the machine.
	//
	// If used, we're assuming a timers instance isn't shared.
	//
	// This data is used to emit a message directly to this
	// machine in order to change states in order to write out
	// updated data in this instance!
	Id string `json:"id,omitempty"`

	Timers map[string]*timerEntry `json:"map"`

	ctl chan bool

	emit emitter
}

func newTimers(emit emitter, id string) *timers {
	return &timers{
		Timers: make(map[string]*timerEntry, 32),
		emit:   emit,
		Id:     id,
		ctl:    make(chan bool),
	}
}

func (ts *timers) Logf(format string, args ...interface{}) {
	if ts.Debug {
		log.Printf(format, args...)
	}
}

func (ts *timers) MarshalJSON() ([]byte, error) {
	ts.Lock()
	m := map[string]interface{}{
		"map": ts.Timers,
	}
	bs, err := json.Marshal(&m)
	ts.Unlock()
	return bs, err
}

func (ts *timers) MarshalYAML() (interface{}, error) {
	ts.Lock()
	cp := Copy(map[string]interface{}{
		"map": ts.Timers,
	})
	ts.Unlock()
	return cp, nil
}

func (ts *timers) add(id string, message interface{}, in time.Duration) error {
	ts.Logf("timers.add %s %v", id, in)

	ts.Lock()
	defer ts.Unlock()

	if _, have := ts.Timers[id]; have {
		return IdExists
	}

	ctx, cancel := context.WithCancel(context.Background())
	te := &timerEntry{
		Ctx:     ctx,
		Cancel:  cancel,
		Id:      id,
		Message: message,
		At:      time.Now().UTC().Add(in),
	}

	ts.Timers[id] = te

	ts.start(te)

	return nil
}

func (ts *timers) rem(id string) error {
	ts.Logf("timers.rem %s", id)

	ts.Lock()
	defer ts.Unlock()

	te, have := ts.Timers[id]
	if !have {
		return NotFound
	}
	te.Cancel()
	return nil
}

func (ts *timers) stop() {
	close(ts.ctl)
}

func (ts *timers) start(te *timerEntry) {
	trigger := make(chan bool)
	go func() {
		cleanup := func() {
			ts.Lock()
			delete(ts.Timers, te.Id)
			ts.Unlock()
		}
		select {
		case <-ts.ctl:
			cleanup()
			ts.Logf("timers halted %s", ts.Id)
		case <-te.Ctx.Done():
			cleanup()
			ts.Logf("timer cancelled %s", te.Id)
			// Probably canceled.
		case <-trigger:
			ts.Logf("timer triggered %s", te.Id)
			// ToDo: Expose an emit timeout via the timer request/struct.
			ctx := context.Background()
			if err := ts.emit(ctx, te.Message); err != nil {
				log.Printf("timers emit error: %s", err)
			}
			// Send a message to ourselves to allow us to clean up.
			//
			// ToDo: Get our machine id.
			{
				cleanup() // Before emitting this message.
				mid := ts.Id
				msg := map[string]interface{}{
					"to": map[string]interface{}{
						"mid": mid,
					},
					"emitted": te.Id,
				}
				if err := ts.emit(ctx, msg); err != nil {
					// ToDo: Probably bring back Machine.Hooks.Errors?
					panic(err)
				}
			}
		}
	}()
	go func() {
		d := te.At.Sub(time.Now())
		ts.Logf("timers.start %s %v", te.Id, d)
		time.Sleep(d)
		select {
		case <-te.Ctx.Done():
		default:
			ts.Logf("timers.start trigger %s", te.Id)
			close(trigger)
		}
	}()
}

func NewTimersSpec() *Spec {
	getTimers := func(bs Bindings) (*timers, string) {
		x, have := bs["timers"]
		if !have {
			return nil, "no timers"
		}
		ts, is := x.(*timers)
		if !is {
			return nil, fmt.Sprintf("bad timers: %T %#v", x, x)
		}
		return ts, ""
	}

	spec := &Spec{
		Name: "timers",
		Doc:  "A machine that makes in-memory timers that send messages.",
		ParamSpecs: map[string]ParamSpec{
			"timers": {
				PrimitiveType: "timers",
			},
		},
		Nodes: map[string]*Node{
			"start": {
				Doc: "Wait to hear a request to create or delete a timer.",
				Branches: &Branches{
					Type: "message",
					Branches: []*Branch{
						{
							Pattern: Dwimjs(`{"makeTimer":{"in":"?in", "message":"?m", "id":"?id"}}`),
							Target:  "make",
						},
						{
							Pattern: Dwimjs(`{"deleteTimer":"?id"}`),
							Target:  "delete",
						},
						{
							Pattern: Dwimjs(`{"emitted":"?id"}`),
							Target:  "emitted",
						},
					},
				},
			},
			"make": {
				Doc: "Try to make the timer.",
				Action: &FuncAction{
					F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
						x, have := bs["?in"]
						if !have {
							return NewExecution(bs.Extend("error", "no in")), nil
						}
						in, is := x.(string)
						if !is {
							return NewExecution(bs.Extend("error", fmt.Sprintf("non-string in: %T %#v", x, x))), nil
						}

						d, err := time.ParseDuration(in)
						if err != nil {
							msg := fmt.Sprintf("bad in '%s': %v", in, err)
							return NewExecution(bs.Extend("error", msg)), nil
						}

						x, have = bs["?id"]
						if !have {
							return NewExecution(bs.Extend("error", "no id")), nil
						}
						id, is := x.(string)
						if !is {
							return NewExecution(bs.Extend("error", fmt.Sprintf("non-string id: %T %#v", x, x))), nil
						}

						message, have := bs["?m"]
						if !have {
							return NewExecution(bs.Extend("error", "no message")), nil
						}

						ts, oops := getTimers(bs)
						if oops != "" {
							return NewExecution(bs.Extend("error", oops)), nil
						}

						if err = ts.add(id, message, d); err != nil {
							return NewExecution(bs.Extend("error", err.Error())), nil
						}

						return NewExecution(bs), nil
					},
				},
				Branches: &Branches{
					Type: "bindings",
					Branches: []*Branch{
						{
							Pattern: Dwimjs(`{"error":"?oops"}`),
							Target:  "problem",
						},
						{
							Target: "success",
						},
					},
				},
			},
			"delete": {
				Doc: "Try to delete the timer.",
				Action: &FuncAction{
					F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
						x, have := bs["?id"]
						if !have {
							return NewExecution(bs.Extend("error", "no id")), nil
						}
						id, is := x.(string)
						if !is {
							return NewExecution(bs.Extend("error", fmt.Sprintf("non-string id: %T %#v", x, x))), nil
						}

						ts, oops := getTimers(bs)
						if oops != "" {
							return NewExecution(bs.Extend("error", oops)), nil
						}

						if err := ts.rem(id); err != nil {
							return NewExecution(bs.Extend("error", err.Error())), nil
						}

						return NewExecution(bs), nil
					},
				},
				Branches: &Branches{
					Type: "bindings",
					Branches: []*Branch{
						{
							Pattern: Dwimjs(`{"error":"?oops"}`),
							Target:  "problem",
						},
						{
							Target: "success",
						},
					},
				},
			},
			"emitted": {
				Doc: "State change to force a write.",
				Action: &FuncAction{
					F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
						return NewExecution(bs.DeleteExcept("timers")), nil
					},
				},
				Branches: &Branches{
					Branches: []*Branch{
						{
							Target: "start",
						},
					},
				},
			},
			"problem": {
				Doc: "Report the problem.",
				Action: &FuncAction{
					F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
						id, have := bs["?id"]
						if !have {
							id = "NA"
						}

						problem, have := bs["error"]
						if !have {
							problem = "unknown"
						}

						message := map[string]interface{}{
							"id":    id,
							"error": problem,
						}

						bs = bs.DeleteExcept("timers")

						e := NewExecution(bs)
						e.AddEmitted(message)
						return e, nil
					},
				},
				Branches: &Branches{
					Branches: []*Branch{
						{
							Target: "start",
						},
					},
				},
			},
			"success": {
				Doc: "Report happiness.",
				Action: &FuncAction{
					F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
						id, have := bs["?id"]
						if !have {
							id = "NA"
						}
						message := map[string]interface{}{
							"changed": id,
						}
						bs = bs.DeleteExcept("timers")
						e := NewExecution(bs)
						e.AddEmitted(message)
						return e, nil
					},
				},
				Branches: &Branches{
					Branches: []*Branch{
						{
							Target: "start",
						},
					},
				},
			},
		},
	}

	return spec
}

func (s *Service) ensureTimersMachine(ctx context.Context, cid string, m *crew.Machine) error {
	// Assumes the lock.

	t, have := s.timers[m.Id]
	if !have {
		log.Printf("Service creating timers for %s", m.Id)
		emitter := func(ctx context.Context, message interface{}) error {
			_, err := s.Process(ctx, cid, message, nil)
			return err
		}
		t = newTimers(emitter, m.Id)
		s.timers[m.Id] = t
	}
	m.State.Bs.Extend("timers", t)

	return nil
}
