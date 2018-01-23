package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/interpreters/goja"
	. "github.com/Comcast/sheens/util/testutil"

	"github.com/jsccast/yaml"
)

type Service struct {
	ProcessCtl *core.Control
	Emitted    chan interface{}
	Processing chan interface{}
	Errors     chan interface{} // Should be error
	Tracing    bool

	ops chan interface{}

	interpreters map[string]core.Interpreter
	crewName     string
	crew         crew.Crew
	specDir      string
	store        *Storage
	timers       *Timers

	wsClientC chan interface{}
}

func (s *Service) trf(format string, args ...interface{}) {
	if !s.Tracing {
		return
	}
	log.Printf("trace "+format, args...)
}

func NewService(ctx context.Context, specDir, dbFile, libDir string) (*Service, error) {

	crewName := "home"

	var store *Storage
	if dbFile != "" {
		var err error

		if store, err = NewStorage(dbFile); err != nil {
			return nil, err
		}

		if err = store.Open(ctx); err != nil {
			return nil, err
		}

		go func() {
			<-ctx.Done()
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			if err := store.Close(ctx); err != nil {
				log.Printf("Service.store.Close error %s", err)
				// Race if we try to use s.Errors.
			}
		}()
	}

	s := Service{
		ProcessCtl: core.DefaultControl,
		crewName:   crewName,
		specDir:    specDir,
		crew: crew.Crew{
			Id:       crewName,
			Machines: make(map[string]*crew.Machine, 32),
		},
		store: store,
	}

	if store != nil {
		if err := store.EnsureCrew(ctx, crewName); err != nil {
			return nil, err
		}
	}

	emitter := func(ctx context.Context, msg interface{}) error {
		// ToDo: Consider going through a COp.Process.
		_, err := s.Process(ctx, msg, s.ProcessCtl)
		return err
	}
	s.timers = NewTimers(emitter)
	s.timers.Errors = s.Errors

	gi := goja.NewInterpreter()
	gi.LibraryProvider = goja.MakeFileLibraryProvider(libDir)
	s.interpreters = map[string]core.Interpreter{
		"goja": gi,
	}

	return &s, nil
}

func (s *Service) op(ctx context.Context, x interface{}) {
	if s.ops != nil {
		select {
		case s.ops <- Copy(x):
		default:
			log.Printf("Service ops chan blocked")
		}
	}
}

func (s *Service) GetSpec(ctx context.Context, src *crew.SpecSource) (core.Specter, error) {

	if src.Name == "" {
		return nil, fmt.Errorf("Unsupported SpecSource %s: needs name", JS(src))
	}
	specSrc, err := ioutil.ReadFile(s.specDir + "/" + src.Name + ".yaml")
	if err != nil {
		return nil, err
	}
	var spec core.Spec
	if err = yaml.Unmarshal(specSrc, &spec); err != nil {
		return nil, err
	}

	if err = spec.Compile(ctx, s.interpreters, true); err != nil {
		return nil, nil
	}

	return &spec, nil
}

func (s *Service) Process(ctx context.Context, msg interface{}, ctl *core.Control) (map[string]*core.Walked, error) {
	s.trf("Service.Process %s", JS(msg))

	if s.Processing != nil {
		select {
		case s.Processing <- msg:
		default:
			log.Printf("Service.Process Processing chan blocked")
		}
	}

	if ctl == nil {
		ctl = core.DefaultControl
	}

	c := &s.crew

	s.trf("Service.Process routing %s", JS(msg))
	mids, all, err := s.Route(ctx, msg)
	if err != nil {
		return nil, err
	}

	s.trf("Service.Process routed %s: mids=%s all=%s (err=%v)", JS(msg), JS(mids), JS(all), err)

	c.Lock()
	defer c.Unlock()

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

		// ToDo: Cache
		spec, err := s.GetSpec(ctx, m.SpecSource)
		if err != nil {
			return nil, err
		}
		specs[mid] = spec.Spec()
	}

	var (
		// The batch of msgs we're submitting.
		msgs []interface{}

		// Processed will accumulate each Machine's Walked.
		processed = make(map[string]*core.Walked, len(c.Machines))

		// States will accumulate each Machine's end state.
		// We'll probably want to write these all out.
		states = make(map[string]*core.State, len(c.Machines))
	)

	if msg == nil {
		msgs = []interface{}{}
	} else {
		msgs = []interface{}{msg}
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

		walked, err := spec.Walk(ctx, m.State, msgs, ctl, props)
		if err != nil {
			if walked.Error != nil {
				walked.Error = NewWrappedError(err, walked.Error)
			} else {
				walked.Error = err
			}
		}
		processed[mid] = walked

		if to := walked.To(); to != nil {
			states[mid] = to.Copy()
		}
	}

	// Gather and write out machine changes.
	mss := AsMachinesStates(states)
	for _, ms := range mss {
		ms.SpecSource = c.Machines[ms.Mid].SpecSource
	}

	if err = s.store.WriteState(ctx, s.crewName, mss); err != nil {
		log.Printf("Service.Process warning for '%s' failed WriteState: %s", s.crewName, err)
	} else {
		for mid, state := range states {
			c.Machines[mid].State = state
		}
	}

	Render(os.Stderr, "processed", processed)

	// Recursively (and asynchronously) process the emitted
	// msgs.
	for _, walked := range processed {
		for _, stride := range walked.Strides {
			for _, msg := range stride.Emitted {
				if s.Emitted != nil {
					select {
					case s.Emitted <- msg:
					default:
						log.Printf("Service.Process Emitted chan blocked")
					}
				}
				go s.Process(ctx, msg, ctl)
			}
		}
	}

	return processed, err
}

func (s *Service) AddMachine(ctx context.Context, specName, id, nodeName string, bs core.Bindings) error {
	if nodeName == "" {
		nodeName = "start"
	}
	if bs == nil {
		bs = core.NewBindings()
	}

	c := &s.crew

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

	c.Lock()
	_, have := c.Machines[id]
	if !have {
		c.Machines[id] = &m
	}
	c.Unlock()

	if have {
		return Exists
	}

	ms := MachineState{
		Mid:        m.Id,
		SpecSource: m.SpecSource,
		NodeName:   m.State.NodeName,
		Bs:         m.State.Bs,
	}

	return s.store.WriteState(ctx, s.crewName, []*MachineState{&ms})
}

func (s *Service) RemMachine(ctx context.Context, mid string) error {
	ms := MachineState{
		Mid:     mid,
		Deleted: true,
	}

	// ToDo: Remove timers?

	return s.store.WriteState(ctx, s.crewName, []*MachineState{&ms})
}

func (s *Service) Route(ctx context.Context, msg interface{}) ([]string, bool, error) {
	m, is := msg.(map[string]interface{})
	if !is {
		return nil, true, nil
	}
	x, have := m["to"]
	if !have {
		return nil, true, nil
	}
	mid, is := x.(string)
	if !is {
		// Not a machine id, so ignore it?
		return nil, true, nil
	}
	switch mid {
	case "ws":
		s.wsClientC <- msg
		return nil, false, nil
	case "http":
		if err := s.toHTTP(ctx, msg); err != nil {
			// Not a "Route" problem.
			s.err(err)
		}
		return nil, false, nil
	case "timers":
		if err := s.toTimers(ctx, msg); err != nil {
			// Not a "Route" problem.
			s.err(err)
		}
		return nil, false, nil
	default:
		return []string{mid}, false, nil
	}
}

func (s *Service) err(err error) {
	// ToDo: Possibly send errors back to the service as messagess.

	if s.Errors != nil {
		s.Errors <- err
	} else {
		log.Println(err)
	}
}
