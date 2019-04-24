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

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/interpreters/ecmascript"
	"github.com/Comcast/sheens/match"

	"github.com/jsccast/yaml"
)

var (
	interpreter = ecmascript.NewInterpreter()

	// Interpreters are the standard action interpreters.
	Interpreters = core.InterpretersMap{
		"goja":           interpreter,
		"ecmascript":     interpreter,
		"ecmascript-5.1": interpreter,
	}
)

// Changed represents changes to a machine after message processing.
type Changed struct {
	State   *core.State      `json:",omitempty"`
	SpecSrc *crew.SpecSource `json:",omitempty"`
	Deleted bool             `json:",omitempty"`

	// PreviousState is optional data that can be used to decide
	// if the new state is really different from the old state.
	//
	// In this implementation, PreviousState is a JSON
	// representation of this struct.
	PreviousState []byte `json:"-"`
}

// Result represents all visible output from processing a message.
type Result struct {
	// Changed represents all machine changes.
	Changed map[string]*Changed

	// Emitted is list of message batches emitted by machines
	// during processing.
	//
	// A message batch is ordered: A machine (usually) emits
	// messages in a specified, deterministic order.
	//
	// The collection of batches is a partial order given by
	// recursive message processing calls.  When processing a
	// message results in emitted messages that are directed back
	// to the crew, the results of those recursive processings
	// give a determinstic order their emitted batches.  However,
	// with respect to a processig a single message, multiple
	// batches are NOT orders (because the order that machines are
	// presented with an in-bound message is not specified).
	Emitted [][]interface{}

	// Diag includes internal processing data.
	Diag []*Stroll
}

// Stroll is a internal processing data for the given message.
//
// Result.Diag gathers this information.
type Stroll struct {
	Msg     interface{} `json:"msg"`
	Walkeds interface{} `json:"walks"`
	Err     string      `json:"err,omitempty"`
}

// Crew represents a collection of machines and associated gear to
// support message processing, with I/O coupled via two channels (in
// and out).
type Crew struct {
	// Machines represents this's Crews current machines.
	Machines map[string]*crew.Machine

	// Conf provides some basic Crew parameters.
	Conf *CrewConf `json:"conf"`

	// Verbose turns on logging.
	Verbose bool

	// changed is a cache of machine state changes that are
	// accumulated during message processing.
	changed map[string]*Changed

	// previous is a cache of machine states prior to processing a
	// message.  Used to compute net changes.
	previous map[string]string

	// timers holds the local, internal, native Timers system.
	timers *Timers

	// in receives all in-bound messages.
	in chan interface{}

	// out receives all out-bound messages.
	out chan *Result

	// done is closed by Couplings when its input is closed.
	done chan bool

	// Mutex can probably be removed once code is cleaned up to
	// perform all state changes, including timers state changes,
	// the Crew loop.  ToDo.
	sync.Mutex
}

// NewCrew makes a crew with the given configuration and couplings.
//
// The coupling's IO() method is called to obtain the crew's in/out
// channels.
func NewCrew(ctx context.Context, conf *CrewConf, couplings Couplings) (*Crew, error) {
	in, out, done, err := couplings.IO(ctx)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		conf = &CrewConf{
			Ctl: core.DefaultControl,
		}
	}
	c := &Crew{
		Conf: conf,
		in:   in,
		out:  out,
		done: done,
	}

	return c, c.init(ctx)
}

// init creates a timers machine and a captain.
func (c *Crew) init(ctx context.Context) error {
	c.Machines = make(map[string]*crew.Machine, 32)
	c.changed = make(map[string]*Changed, 8)
	c.previous = make(map[string]string, 8)

	f := func(ctx context.Context, te *TimerEntry) {
		select {
		case <-ctx.Done():
		case c.in <- te.Msg:
		}
	}
	c.timers = NewTimers(f)
	c.timers.c = c

	// c.UpdateHook = func(m map[string]*Changed) error {
	// 	log.Printf("changes: %s", JS(m))
	// 	return nil
	// }

	if err := c.SetMachine(ctx, CaptainMachine, nil, nil); err != nil {
		return err
	}

	if err := c.SetMachine(ctx, TimersMachine, nil, nil); err != nil {
		return err
	}

	if c.Conf.EnableHTTP {
		if err := c.SetMachine(ctx, HTTPMachine, nil, nil); err != nil {
			return err
		}
	}

	return nil
}

// change just updates the cache of what machines have changed.
func (c *Crew) change(mid string) *Changed {
	ch, have := c.changed[mid]
	if !have {
		ch = &Changed{}
		c.changed[mid] = ch
	}
	return ch
}

// Logf logs if c.Verbose.
func (c *Crew) Logf(format string, args ...interface{}) {
	if !c.Verbose {
		return
	}
	log.Printf(format, args...)
}

// Errorf emits an error message and writes a log line with "ERROR"
// prepended.
func (c *Crew) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Println("ERROR " + msg)
	c.out <- &Result{
		Emitted: [][]interface{}{
			[]interface{}{
				map[string]interface{}{
					"error": msg,
				},
			},
		},
	}
}

// SetMachine creates or updates a machine.
//
// When the mid is either (the variable) TimersMachine and the given
// state is nil, the timers machine's state is reset.
func (c *Crew) SetMachine(ctx context.Context, mid string, src *crew.SpecSource, state *core.State) error {
	m, have := c.Machines[mid]

	if !have {
		m = &crew.Machine{
			Id:    mid,
			State: DefaultState(state),
		}

		c.Machines[mid] = m
	}

	if src != nil {
		c.change(mid).SpecSrc = src
	}

	if state != nil {
		c.change(mid).State = state
	}

	switch mid {
	case HTTPMachine:
		if m.Specter == nil {
			spec := c.NewHTTPSpec()
			if err := spec.Compile(ctx, Interpreters, true); err != nil {
				return err
			}
			m.Specter = spec
		}
	case TimersMachine:
		if m.Specter == nil {
			spec := c.NewTimersSpec()
			if err := spec.Compile(ctx, Interpreters, true); err != nil {
				return err
			}
			m.Specter = spec
		}

		if state == nil {
			state = DefaultState(nil)
			state.Bs["timers"] = c.timers.Map

		}
		if ts, have := state.Bs["timers"]; have {
			if err := c.timers.withMap(ts); err != nil {
				return err
			}
			m.State.Bs["timers"] = c.timers.Map
			if err := c.timers.Start(ctx); err != nil {
				return err
			}
		}
	case CaptainMachine:
		spec := c.NewCaptainSpec()
		if err := spec.Compile(ctx, Interpreters, true); err != nil {
			return err
		}
		m.Specter = spec
	default:
		if src != nil {
			ss, spec, err := ResolveSpecSource(ctx, src)
			if err != nil {
				delete(c.Machines, mid)
				return err
			}
			m.SpecSource = ss
			m.Specter = spec
		}
	}

	return nil
}

// DeleteMachine removes a machine from the crew.
//
// No error is returned if the machine doesn't exist.
func (c *Crew) DeleteMachine(ctx context.Context, mid string) error {
	delete(c.Machines, mid)
	c.change(mid).Deleted = true
	return nil
}

// ProcessMsg processes the given message and returns the results,
// which can then be processed by the crew's Result coupling.
func (c *Crew) ProcessMsg(ctx context.Context, msg interface{}) (*Result, error) {
	c.Logf("ProcessMsg %s", JS(msg))

	c.Lock()
	defer c.Unlock()

	// Some emitted messages are routed back to sheens.  Rather
	// than call ProcessMsg recursively, we take a breadth-first
	// approach.  That approach is the correct one since an
	// emitted message shouldn't be processed until all machines
	// have processed the current message.
	pending := make([]interface{}, 0, 32)
	pending = append(pending, msg)

	r := &Result{
		Emitted: make([][]interface{}, 0, 8),
		Diag:    make([]*Stroll, 0, 8),
	}

	for 0 < len(pending) {
		// ToDo: Enforce a limit.
		msg := pending[0]
		pending = pending[1:] // ToDo: Consider leak.
		c.Logf("ProcessMsg at %s (%d)", JS(msg), len(pending))

		if f, is := msg.(func(*Crew) interface{}); is {
			msg = f(c)
		}

		walkeds, err := c.RunMachines(ctx, msg)
		stroll := &Stroll{
			Msg:     msg,
			Walkeds: walkeds,
		}
		r.Diag = append(r.Diag, stroll)

		if err != nil {
			stroll.Err = err.Error()
			return r, err
		}

		for _, walked := range walkeds {
			if walked.Error != nil {
				c.Errorf("ProcessMsg %s", walked.Error)
			}
			emitted := make([]interface{}, 0, 8)
			walked.DoEmitted(func(msg interface{}) error {
				if m, is := msg.(map[string]interface{}); is {
					if _, has := m["emit"]; has {
						// ToDo: Do not reprocess.
					}
				}
				pending = append(pending, msg)
				emitted = append(emitted, msg)
				return nil
			})
			if 0 < len(emitted) {
				r.Emitted = append(r.Emitted, emitted)
			}
		}
	}

	changed, err := c.GetChanged(ctx)
	if err != nil {
		return nil, err
	}

	r.Changed = changed

	return r, nil
}

// GetChanged computes the net machine changes since this method was
// previously called.
//
// ToDo: Make private.
func (c *Crew) GetChanged(ctx context.Context) (map[string]*Changed, error) {

	changed := make(map[string]*Changed, 32)

	for mid, change := range c.changed {
		delete(c.changed, mid)

		if mid == CaptainMachine {
			continue
		}

		if change.Deleted {
			changed[mid] = &Changed{
				Deleted: true,
			}
			continue
		}

		ched, have := changed[mid]
		if !have {
			ched = &Changed{}
			changed[mid] = ched
		}

		if change.State != nil {
			ched.State = change.State.Copy()
		}

		if change.SpecSrc != nil {
			ched.SpecSrc = change.SpecSrc
		}

	}

	for mid, ch := range changed {
		if ch.Deleted {
			delete(c.previous, mid)
			continue
		}
		js, err := json.Marshal(ch)
		if err != nil {
			return nil, err
		}
		current := string(js)
		if previous, have := c.previous[mid]; have {
			if current == previous {
				delete(changed, mid)
				continue
			}
		}
		c.previous[mid] = current
	}

	return changed, nil
}

// Loop starts the input processing loop in the current goroutine.
//
// This loop calls ProcessMsg on each message that arrives via the
// input coupling, and the loop halts when ctx.Done().
func (c *Crew) Loop(ctx context.Context) error {
	c.Logf("Crew.Loop starting")
LOOP:
	for {
		select {
		case <-c.done:
			if c.Conf.HaltOnInputEOF {
				c.Logf("Crew.Loop shutting down (c.done)")
				break LOOP
			}
		case <-ctx.Done():
			c.Logf("Crew.Loop shutting down (ctx.Done)")
			break LOOP
		case msg := <-c.in:
			if msg == nil {
				break LOOP
			}
			r, err := c.ProcessMsg(ctx, msg)
			if err != nil {
				c.Errorf("Crew.Loop ProcessMsg %s", err)
				// ToDo: Consider reprocessing msg?
				continue
			}
			select {
			case <-ctx.Done():
			case c.out <- r:
			}
		}
	}

	c.Logf("Crew.Loop done")
	return nil
}

// allMachines returns the set of all machines except for the
// TimersMachine and the CaptainMachine.
//
// We want to present a message to the TimersMachine or the
// CaptainMachine only if specifically directed.
func (c *Crew) allMachines() []string {
	acc := make([]string, 0, len(c.Machines))
	for mid, _ := range c.Machines {
		switch mid {
		case HTTPMachine:
		case TimersMachine:
		case CaptainMachine:
		default:
			acc = append(acc, mid)
		}
	}
	return acc
}

// toMachines determines the set of machines that should see this
// message.
//
// Calls allMachines if the message doesn't have a "to" property.
func (c *Crew) toMachines(ctx context.Context, msg interface{}) ([]string, error) {
	m, is := msg.(map[string]interface{})
	if !is {
		return c.allMachines(), nil
	}
	if x, have := m["to"]; have {
		switch vv := x.(type) {
		case string:
			if vv == "*" {
				return c.allMachines(), nil
			}
			return []string{vv}, nil
		case []string:
			return vv, nil
		case []interface{}:
			mids := make([]string, len(vv))
			for i, x := range vv {
				switch vv := x.(type) {
				case string:
					mids[i] = vv
				}
			}
			return mids, nil
		}
	}
	return c.allMachines(), nil
}

// RunMachines presents the message to the machines returned by
// toMachines.
func (c *Crew) RunMachines(ctx context.Context, msg interface{}) (map[string]*core.Walked, error) {
	mids, err := c.toMachines(ctx, msg)
	if err != nil {
		return nil, err
	}
	c.Logf("RunMachines routing to %#v", mids)

	acc := make(map[string]*core.Walked, len(mids))

	for _, mid := range mids {
		if m, have := c.Machines[mid]; have {
			walked, err := c.RunMachine(ctx, msg, m)
			if err != nil {
				c.Errorf("RunMachines %s", err)
			} else {
				acc[mid] = walked
			}
		}
	}

	return acc, nil
}

// RunMachines presents the message to the given machine.
func (c *Crew) RunMachine(ctx context.Context, msg interface{}, m *crew.Machine) (*core.Walked, error) {
	if m.Specter == nil {
		return nil, fmt.Errorf("no Spectre for %s in %s", m.Id, c.Conf.Id)

	}
	spec := m.Specter.Spec()
	if spec == nil {
		return nil, fmt.Errorf("no Spectre.Spec for %s in %s", m.Id, c.Conf.Id)
	}

	props := core.StepProps{
		"mid": m.Id,
		"ctx": ctx,
	}

	// Only the captain can mess with the entire crew.  But our
	// current demo captain doesn't prevent messing with the
	// captain via itself!
	if m.Id == "captain" {
		props["crew"] = c
	}

	// if UnsafeCmd {
	// 	props["exec"] = ecmascript.UnsafeCmd
	// }

	msgs := []interface{}{msg}

	walked, err := spec.Walk(ctx, m.State, msgs, c.Conf.Ctl, props)
	if err != nil {
		return nil, err
	}

	if to := walked.To(); to != nil {
		m.State = to.Copy()
		c.change(m.Id).State = to.Copy()
	}

	return walked, err
}

// ResolveSpecSource attempts to find and compile a spec based on a
// crew.SpecSource (or something that looks like one).
//
// Attempts to obtain a spec by examining .Inline, .URL, and .Source
// in that order.  The URL can be a 'file://' (with support for
// relative paths).  The Source can be JSON or YAML.
func ResolveSpecSource(ctx context.Context, specSource interface{}) (*crew.SpecSource, *core.Spec, error) {
	js, err := json.Marshal(&specSource)
	if err != nil {
		return nil, nil, err
	}
	var src crew.SpecSource
	if err = json.Unmarshal(js, &src); err != nil {
		return nil, nil, err
	}

	if src.Inline != nil {
		if err = src.Inline.Compile(ctx, Interpreters, true); err != nil {
			log.Printf("spec.Compile error: %v", err)
			return nil, nil, err
		}
		return &src, src.Inline, nil
	}

	var body []byte

	if src.URL != "" {
		// Yikes.  We detest IO.

		if strings.HasPrefix(src.URL, "file://") {
			filename := src.URL[7:]
			body, err = ioutil.ReadFile(filename)
		} else {
			resp, err := http.Get(src.URL)
			if err != nil {
				return nil, nil, err
			}
			body, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}

	if src.Source != "" {
		body = []byte(src.Source)
	}

	if len(body) == 0 {
		return nil, nil, fmt.Errorf("spec source is empty")
	}

	var spec core.Spec
	switch body[0] {
	case '{':
		err = json.Unmarshal(body, &spec)
	default:
		err = yaml.Unmarshal(body, &spec)
	}
	if err != nil {
		return nil, nil, err
	}
	if err = spec.Compile(ctx, Interpreters, true); err != nil {
		log.Printf("spec.Compile error: %v", err)
		return nil, nil, err
	}

	return &src, &spec, nil
}

// DefaultState returns a state at "state" with empty bindings.
func DefaultState(s *core.State) *core.State {
	if s == nil {
		s = &core.State{}
	}
	if s.NodeName == "" {
		s.NodeName = "start"
	}
	if s.Bs == nil {
		s.Bs = match.NewBindings()
	}
	return s
}
