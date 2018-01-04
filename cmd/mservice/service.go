package main

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/Comcast/sheens/cmd/mservice/storage"
	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
)

// Router is a function that determines what machines (within a crew)
// should receive the given message.
//
// If a Router returns "all", then all machines in the crew will see
// the message.  Otherwise, the router can enumerate the machines (by
// machine id) that should receive the message.
type Router func(ctx context.Context, c *crew.Crew, message interface{}) (targets []string, all bool, err error)

// SpecProvider resolves a SpecSource to a Spec(ter).
type SpecProvider func(ctx context.Context, s *crew.SpecSource) (core.Specter, error)

// Service is an example multi-crew service.
type Service struct {
	sync.Mutex

	Interpreters map[string]core.Interpreter
	Router       Router
	SpecProvider SpecProvider
	Storage      storage.Storage

	timers    map[string]*timers
	crewCache *CrewCache

	InSubs  *Subs
	OutSubs *Subs
}

// NewService makes a new, empty Service.
//
// You'll need to populate Interpreters, Router, etc.
func NewService() (*Service, error) {
	return &Service{
		timers:  make(map[string]*timers, 32),
		InSubs:  NewSubs(),
		OutSubs: NewSubs(),
	}, nil
}

// MakeCrew is a service-level API to create a Crew.
func (s *Service) MakeCrew(ctx context.Context, cid string) error {
	return s.Storage.MakeCrew(ctx, cid)
}

// RemCrew is a service-level API to remove a Crew.
func (s *Service) RemCrew(ctx context.Context, cid string) error {
	c, err := s.findCrew(ctx, cid)
	if err != nil {
		return err
	}

	for _, m := range c.Machines {
		if m.SpecSource != nil && m.SpecSource.Name == "timers" {
			t, have := s.timers[m.Id]
			if !have {
				// ToDo: warn?
				break
			}
			t.stop()
			delete(s.timers, m.Id)
			break
		}
	}

	if s.crewCache != nil {
		s.crewCache.Rem(cid)
	}

	return s.Storage.RemCrew(ctx, c.Id)
}

// findCrew reads Crew state from Storage.
//
// This method also configures any native 'timers' machine.
//
// This method should use a cache or otherwise be much more clever.
func (s *Service) findCrew(ctx context.Context, cid string) (*crew.Crew, error) {
	s.Lock()
	defer s.Unlock()

	log.Printf("Service.findCrew %s", cid)

	if s.crewCache != nil {
		if c := s.crewCache.Get(cid); c != nil {
			return c, nil
		}
	}

	t := NewTimer("findCrew")

	mss, err := s.Storage.GetCrew(ctx, cid)
	if err != nil {
		return nil, err
	}

	ms := storage.AsMachines(mss)

	// A hack: make sure we have an internal timers instance when
	// we see a "timers" machine.
	for _, m := range ms {
		if m.SpecSource != nil && m.SpecSource.Name == "timers" {
			if err = s.ensureTimersMachine(ctx, cid, m); err != nil {
				return nil, err
			}
		}
	}

	c := &crew.Crew{
		Id:       cid,
		Machines: ms,
	}

	t.StopLog()

	if s.crewCache != nil {
		s.crewCache.Put(cid, c)
	}

	return c, nil
}

// AddMachine is a service-level API to add a Machine to a Crew.
func (s *Service) AddMachine(ctx context.Context, cid, specName, id, nodeName string, bs core.Bindings) error {
	if nodeName == "" {
		nodeName = "start"
	}
	if bs == nil {
		bs = core.NewBindings()
	}

	c, err := s.findCrew(ctx, cid)
	if err != nil {
		return err
	}

	m := crew.Machine{
		Id: id,
		State: &core.State{
			NodeName: nodeName,
			Bs:       bs,
		},
		SpecSource: &crew.SpecSource{
			Name: specName,
		},
	}

	if m.SpecSource != nil && m.SpecSource.Name == "timers" {
		if err = s.ensureTimersMachine(ctx, cid, &m); err != nil {
			return err
		}
	}

	c.Lock()
	c.Machines[id] = &m
	c.Unlock()

	ms := storage.MachineState{
		Mid:        m.Id,
		SpecSource: m.SpecSource,
		NodeName:   m.State.NodeName,
		Bs:         m.State.Bs,
	}

	return s.Storage.WriteState(ctx, cid, []*storage.MachineState{&ms})
}

// RemMachine is a service-level API to remove a Machine from a Crew.
func (s *Service) RemMachine(ctx context.Context, cid string, mid string) error {
	ms := storage.MachineState{
		Mid:     mid,
		Deleted: true,
	}

	// ToDo: Remove timers?

	return s.Storage.WriteState(ctx, cid, []*storage.MachineState{&ms})
}

// Process is a service-level API to send a message to a Crew.
func (s *Service) Process(ctx context.Context, cid string, message interface{}, ctl *core.Control) (map[string]*core.Walked, error) {

	log.Printf("Service.Process %s %s", cid, JS(message))

	s.InSubs.Do(cid, message)

	t := NewTimer("process")

	if ctl == nil {
		ctl = core.DefaultControl
	}

	c, err := s.findCrew(ctx, cid)
	if err != nil {
		return nil, err
	}

	c.Lock()
	defer c.Unlock()

	mids, all, err := s.Router(ctx, c, message)
	if err != nil {
		return nil, err
	}

	if all {
		mids = make([]string, 0, len(c.Machines))
		for mid := range c.Machines {
			mids = append(mids, mid)
		}
	}

	specs := make(map[string]*core.Spec, len(mids))

	for _, mid := range mids {
		m, have := c.Machines[mid]
		if !have {
			// Warn?
			continue
		}
		spec, err := s.SpecProvider(ctx, m.SpecSource)
		if err != nil {
			return nil, err
		}
		specs[mid] = spec.Spec()
	}

	var (
		// The batch of messages we're submitting.
		messages []interface{}

		// Processed will accumulate each Machine's Walked.
		processed = make(map[string]*core.Walked, len(c.Machines))

		// States will accumulate each Machine's end state.
		// We'll probably want to write these all out.
		states = make(map[string]*core.State, len(c.Machines))
	)

	if message == nil {
		messages = []interface{}{}
	} else {
		messages = []interface{}{message}
	}

	for _, mid := range mids {
		m, have := c.Machines[mid]
		if !have {
			// Warn?
			continue
		}
		spec, have := specs[mid]
		if !have {
			return nil, errors.New("internal error: lost spec for " + mid)
		}
		props := core.StepProps{
			"mid": mid,
			"cid": c.Id,
		}
		walked, err := spec.Walk(ctx, m.State, messages, ctl, props)
		if err != nil {
			if walked.Error != nil {
				walked.Error = NewWrappedError(err, walked.Error)
			} else {
				walked.Error = err
			}
		}
		processed[mid] = walked

		if to := walked.To(); to != nil {
			to = to.Copy()
			states[mid] = to
		}
	}

	// Gather and write out machine changes.
	mss := storage.AsMachinesStates(states)
	for _, ms := range mss {
		ms.SpecSource = c.Machines[ms.Mid].SpecSource
	}

	wt := NewTimer("writestate")
	err = s.Storage.WriteState(ctx, cid, mss)
	t.Sub(wt.StopLog())

	if err != nil {
		log.Printf("Service.Process warning for '%s' failed WriteState: %s", cid, err)
	} else {
		// Update our in-memory states.
		for mid, state := range states {
			c.Machines[mid].State = state
		}
	}

	// Recursively (and asynchronously) process the emitted
	// messages.
	for _, walked := range processed {
		for _, stride := range walked.Strides {
			for _, message := range stride.Emitted {
				s.OutSubs.Do(cid, message)
				go s.Process(ctx, cid, message, ctl)
			}
		}
	}

	t.StopLog()

	return processed, err
}
