/* Copyright 2018 Comcast Cable Communications Management, LLC
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

// Package main is a command-line machine debugger in the spirit of gdb.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/interpreters"
	"github.com/Comcast/sheens/match"
	. "github.com/Comcast/sheens/util/testutil"

	"github.com/jsccast/yaml"
)

type Opts struct {
	specDir string
	libDir  string
	echo    bool
}

func main() {

	opts := &Opts{}
	flag.StringVar(&opts.specDir, "s", "specs", "spec directory")
	flag.StringVar(&opts.libDir, "l", "libs", "libraries directory")
	flag.BoolVar(&opts.echo, "e", false, "echo input")
	flag.Parse()

	if err := opts.run(); err != nil {
		panic(err)
	}
}

func (opts *Opts) run() error {

	in := os.Stdin
	w := os.Stdout

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h, err := NewHost(opts.specDir, opts.libDir)
	if err != nil {
		return err
	}

	var (
		setNode = regexp.MustCompile("^set +([-a-zA-Z0-9_]+) +node +([-a-zA-Z0-9_]+)")

		setBindings = regexp.MustCompile("^set +([-a-zA-Z0-9_]+) +(bs|bindings) +(.*)")

		setSpec = regexp.MustCompile("^set +([-a-zA-Z0-9_]+) +spec +(.*)")

		reloadSpec = regexp.MustCompile("^reload +([-a-zA-Z0-9_]+)")

		rem = regexp.MustCompile("^(rem|del|remove|delete) +([-a-zA-Z0-9_]+)")

		print = regexp.MustCompile("^print( +([-a-zA-Z0-9_]+))?")

		printqueue = regexp.MustCompile("^printqueue")

		send = regexp.MustCompile("^(run|run +(.*))$")

		pop = regexp.MustCompile("^pop")

		drop = regexp.MustCompile("^drop")

		help = regexp.MustCompile("^(help|h|\\?)")

		save = regexp.MustCompile("^save +(.*)")

		load = regexp.MustCompile("^load +(.*)")

		debug = regexp.MustCompile("^debug(ging)? (on|off)")

		outputPrefix = "# "

		debugging = false

		say = func(format string, args ...interface{}) {

			fmt.Fprintf(w, outputPrefix+format+"\n", args...)
		}

		protest = func(format string, args ...interface{}) {
			say("error: "+format, args...)
		}

		queue = make([]interface{}, 0, 128)

		setSpecHistory = make(map[string]string)
	)

	r := bufio.NewReader(in)
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)

		if opts.echo {
			fmt.Println(line)
		}

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		var ss []string

		if ss = help.FindStringSubmatch(line); 0 < len(ss) {
			for _, s := range strings.Split(doc(), "\n") {
				say("%s", s)
			}
			continue
		}
		if ss = reloadSpec.FindStringSubmatch(line); 0 < len(ss) {
			mid := ss[1]
			filename, have := setSpecHistory[mid]
			if !have {
				protest("no spec filename history for '%s'", mid)
			}
			line = fmt.Sprintf("set %s spec %s", mid, filename)
			say("reloading spec for '%s' from %s", mid, filename)
			// Fall through!
		}
		if ss = setSpec.FindStringSubmatch(line); 0 < len(ss) {
			mid := ss[1]
			specFilename := ss[2]
			setSpecHistory[mid] = specFilename
			m, have := h.crew.Machines[mid]
			if !have {
				m = &crew.Machine{
					Id: mid,
				}
				m.State = &core.State{
					NodeName: "start",
					Bs:       match.NewBindings(),
				}
			}
			m.SpecSource = &crew.SpecSource{
				Name: specFilename,
			}
			spec, err := h.GetSpec(ctx, m.SpecSource)
			if err != nil {
				protest("couldn't load spec %s: %s", specFilename, err)
				continue
			}
			m.Specter = spec
			if !have {
				h.crew.Machines[mid] = m
				say("crew now has %d machines", len(h.crew.Machines))
			}
			continue
		}
		if ss = rem.FindStringSubmatch(line); 0 < len(ss) {
			mid := ss[1]
			if _, have := h.crew.Machines[mid]; !have {
				say("error: machine '%s' not found", mid)
				continue
			}
			delete(h.crew.Machines, mid)
			say("crew now has %d machines", len(h.crew.Machines))
			continue
		}
		if ss = setNode.FindStringSubmatch(line); 0 < len(ss) {
			mid := ss[1]
			node := ss[2]
			m, have := h.crew.Machines[mid]
			if !have {
				say("error: machine '%s' not found", mid)
				continue
			}
			m.State.NodeName = node
			continue
		}
		if ss = setBindings.FindStringSubmatch(line); 0 < len(ss) {
			mid := ss[1]
			js := ss[3]
			var bs match.Bindings
			if err = json.Unmarshal([]byte(js), &bs); err != nil {
				protest("couldn't parse bindings %s", js)
				continue
			}
			m, have := h.crew.Machines[mid]
			if !have {
				protest("machine '%s' not found", mid)
				continue
			}
			m.State.Bs = bs
			continue
		}
		if ss = drop.FindStringSubmatch(line); 0 < len(ss) {
			if len(queue) == 0 {
				protest("queue is empty")
				continue
			}
			queue = queue[1:]
			say("queue now has %d messages", len(queue))
		}

		if ss = pop.FindStringSubmatch(line); 0 < len(ss) {
			if len(queue) == 0 {
				protest("queue is empty")
				continue
			}
			msg := queue[0]
			queue = queue[1:]
			js, err := json.Marshal(&msg)
			if err != nil {
				return err // Internal error
			}
			say("processing %s", js)
			line = fmt.Sprintf("run %s", js)
			// Fall through!
		}

		if ss = send.FindStringSubmatch(line); 0 < len(ss) {
			js := ss[2]
			var msg interface{}
			if js == "" {
				msg = NA
			} else {
				if err = json.Unmarshal([]byte(js), &msg); err != nil {
					protest("couldn't parse message %s", js)
					continue
				}
			}

			walkeds, err := h.Process(ctx, msg, h.ctl)
			if err != nil {
				protest("processing failed: %s", err)
				continue
			}

			Render(w, outputPrefix, "", walkeds)

			if debugging {
				js, _ := json.MarshalIndent(walkeds, "  ", "  ")
				fmt.Println(string(js))
			}

			for _, walked := range walkeds {
				for _, stride := range walked.Strides {
					for _, msg := range stride.Emitted {
						queue = append(queue, msg)
					}
				}
			}
			say("queue has %d messages", len(queue))

			continue
		}

		if ss = printqueue.FindStringSubmatch(line); 0 < len(ss) {
			if len(queue) == 0 {
				say("queue is empty")
				continue
			}
			for i, msg := range queue {
				js, err := json.Marshal(&msg)
				if err != nil {
					return err // Internal error
				}
				say("%d. %s", i, js)
			}
			continue
		}

		if ss = debug.FindStringSubmatch(line); 0 < len(ss) {
			switch ss[2] {
			case "on":
				debugging = true
				say("debugging")
			case "off":
				debugging = true
				say("not debugging")
			}
			continue
		}

		if ss = print.FindStringSubmatch(line); 0 < len(ss) {
			mid := ss[2]
			printer := func(id string) error {
				m, have := h.crew.Machines[id]
				if !have {
					return fmt.Errorf("machine '%s' not found", id)
				}
				say("  node:     %s", m.State.NodeName)
				js, err := json.Marshal(m.State.Bs)
				if err != nil {
					return err // Internal error
				}
				say("  bindings: %s", js)
				say("  spec:     %s", m.SpecSource.Name)
				return nil
			}
			if mid == "" {
				for mid, _ := range h.crew.Machines {
					say("machine %s:", mid)
					if err := printer(mid); err != nil {
						protest("%s", err)
					}
				}
				continue
			}

			if err = printer(mid); err != nil {
				protest("%s", err)
			}
			continue

		}

		if ss = save.FindStringSubmatch(line); 0 < len(ss) {
			filename := ss[1]
			js, err := json.MarshalIndent(&h.crew.Machines, "  ", "  ")
			if err != nil {
				return err // Internal error
			}
			if err = ioutil.WriteFile(filename, js, 0644); err != nil {
				protest("writing file: %s", err)
				continue
			}
			continue
		}

		if ss = load.FindStringSubmatch(line); 0 < len(ss) {
			filename := ss[1]
			js, err := ioutil.ReadFile(filename)
			if err != nil {
				protest("reading file '%s': %s", filename, err)
				continue
			}
			var c crew.Crew
			if err = json.Unmarshal(js, &c.Machines); err != nil {
				protest("loading data: %s", filename, err)
				continue
			}
			for _, m := range c.Machines {
				spec, err := h.GetSpec(ctx, m.SpecSource)
				if err != nil {
					protest("couldn't load spec source %s: %s", JS(m.SpecSource), err)
					continue
				}
				m.Specter = spec
			}
			h.crew.Machines = c.Machines
			say("crew now has %d machines", len(h.crew.Machines))
			continue
		}

		protest("unsupported command: %s", line)
	}

	return nil
}

type Host struct {
	ctl          *core.Control
	interpreters core.InterpretersMap
	crew         crew.Crew
	specDir      string
}

func NewHost(specDir, libDir string) (*Host, error) {
	return &Host{
		ctl:     core.DefaultControl,
		specDir: specDir,
		crew: crew.Crew{
			Machines: make(map[string]*crew.Machine, 32),
		},
		interpreters: interpreters.Standard(),
	}, nil
}

func (h *Host) GetSpec(ctx context.Context, src *crew.SpecSource) (core.Specter, error) {

	if src.Name == "" {
		return nil, fmt.Errorf("Unsupported SpecSource %s: needs name", JS(src))
	}
	specSrc, err := ioutil.ReadFile(h.specDir + "/" + src.Name)
	if err != nil {
		return nil, err
	}
	if len(specSrc) == 0 {
		return nil, fmt.Errorf("empty spec")
	}
	var spec core.Spec
	switch specSrc[0] {
	case '{':
		err = json.Unmarshal(specSrc, &spec)
	default:
		err = yaml.Unmarshal(specSrc, &spec)
	}

	if err = yaml.Unmarshal(specSrc, &spec); err != nil {
		return nil, err
	}
	if err = spec.Compile(ctx, h.interpreters, true); err != nil {
		return nil, err
	}

	return &spec, nil
}

func (h *Host) Process(ctx context.Context, msg interface{}, ctl *core.Control) (map[string]*core.Walked, error) {

	if ctl == nil {
		ctl = h.ctl
	}

	c := &h.crew

	mids, all, err := h.Route(ctx, msg)
	if err != nil {
		return nil, err
	}

	c.Lock()
	defer c.Unlock()

	if all {
		mids = make([]string, 0, len(c.Machines))
		for mid := range c.Machines {
			mids = append(mids, mid)
		}
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

	if msg == NA {
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
		spec := m.Specter.Spec()

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

	for mid, state := range states {
		c.Machines[mid].State = state
	}

	return processed, err
}

func (h *Host) Route(ctx context.Context, msg interface{}) ([]string, bool, error) {
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
	default:
		return []string{mid}, false, nil
	}
}

type WrappedError struct {
	Outer error `json:"outer"`
	Inner error `json:"inner"`
}

func (e *WrappedError) Error() string {
	return e.Outer.Error() + " after " + e.Inner.Error()
}

func NewWrappedError(outer, inner error) error {
	if inner == nil {
		return outer
	}
	return &WrappedError{
		Outer: outer,
		Inner: inner,
	}
}

func doc() string {
	return `
  set ID spec FILENAME       Set the spec for the machine with that ID
  reload ID                  Reload the last spec for the machine with that ID
  set ID node NODENAME       Set the node for the machine with that ID
  set ID bindings BINDINGS   Set the bindings (JSON) for the machine with that ID
  rem ID                     Remove the machine with that ID
  print [ID]                 Print the state of the machine with that ID
  run [MSG]                  Run the crew.
  printqueue                 Show the queue of emitted messages
  pop                        Send first message in the queue to the crew
  drop                       Drop the first message in the queue
  save FILENAME              Save the crew machines to this file
  load FILENAME              Load the crew machines from this file
  debug on/off               When debugging, show walking details
  help                       Show this documentation
`
}

var NA = struct{}{}

func Render(w io.Writer, prefix, tag string, m map[string]*core.Walked) {
	fmt.Fprintf(w, "%sWalkeds %s (%d machines)\n", prefix, tag, len(m))
	for mid, walked := range m {
		fmt.Fprintf(w, "%sMachine %s\n", prefix, mid)
		for i, stride := range walked.Strides {
			fmt.Fprintf(w, "%s  %02d from     %s\n", prefix, i, JS(stride.From))
			fmt.Fprintf(w, "%s     to       %s\n", prefix, JS(stride.To))
			if stride.Consumed != nil {
				fmt.Fprintf(w, "%s     consumed %s\n", prefix, JS(stride.Consumed))
			}
			if 0 < len(stride.Events.Emitted) {
				fmt.Fprintf(w, "%s     emitted\n", prefix)
			}
			for _, emitted := range stride.Events.Emitted {
				fmt.Fprintf(w, "%s        %s\n", prefix, JS(emitted))
			}
		}
		if walked.Error != nil {
			fmt.Fprintf(w, "%s  error    %v\n", prefix, walked.Error)
		}
		fmt.Fprintf(w, "%s  stopped     %v\n", prefix, walked.StoppedBecause)
	}
}
