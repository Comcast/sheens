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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Comcast/sheens/crew"
)

// Stdio is a fairly simple Couplings that uses stdin for input and
// stdout for output.
//
// State is optionally written as JSON to a file.
type Stdio struct {
	// In is coupled to crew input.
	In io.Reader

	// Out is coupled to crew output.
	Out io.Writer

	// ShellExpand enables input to include inline shell commands
	// delimited by '<<' and '>>'.  Use at your wown risk, of
	// course!
	ShellExpand bool

	// Timestamps prepends a timestamp to each output line.
	Timestamps bool

	// EchoInput writes input lines (prepended with "input") to
	// the output.
	EchoInput bool

	// PadTags adds some padding to tags ("input", "emit",
	// "update") used in output.
	PadTags bool

	JSONStore

	// WriteStatePerMsg will write out all state after every input
	// message is processed.
	//
	// Inefficient!
	WriteStatePerMsg bool

	// InputEOF will be closed on EOF from stdin.
	InputEOF chan bool
}

// NewStdio creates a new Stdio.
//
// ShellExpand enables input to include inline shell commands
// delimited by '<<' and '>>'.  Use at your wown risk, of course!
//
// In and Out are initialized with os.Stdin and os.Stdout
// respectively.
func NewStdio(shellExpand bool) *Stdio {
	return &Stdio{
		In:          os.Stdin,
		Out:         os.Stdout,
		ShellExpand: shellExpand,
		InputEOF:    make(chan bool),
	}
}

// Start does nothing.
func (s *Stdio) Start(ctx context.Context) error {
	return nil
}

// Stop writes out the state if requested by StateInputFilename.
//
// This function waits until IO is complete or was terminated via its
// context.
func (s *Stdio) Stop(ctx context.Context) error {
	s.WG.Wait()
	return s.writeState(ctx)
}

// Read reads s.StateInputFilename, which should contain a JSON
// representation of the crew's state.
func (s *Stdio) Read(ctx context.Context) (map[string]*crew.Machine, error) {
	if s.StateInputFilename != "" {
		js, err := ioutil.ReadFile(s.StateInputFilename)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(js, &s.state); err != nil {
			return nil, err
		}
		return s.state, nil

	}
	return make(map[string]*crew.Machine), nil
}

// IO returns channels for reading from stdin and writing to stdout.
func (s *Stdio) IO(ctx context.Context) (chan interface{}, chan *Result, error) {
	in := make(chan interface{})

	if s.StateOutputFilename != "" {
		s.state = make(map[string]*crew.Machine)
	}

	printf := func(tag, format string, args ...interface{}) {
		if s.PadTags {
			tag = fmt.Sprintf("% 10s", tag)
		}
		format = tag + " " + format
		if s.Timestamps {
			ts := fmt.Sprintf("%-31s", time.Now().UTC().Format(time.RFC3339Nano))
			format = ts + " " + format
		}

		fmt.Fprintf(s.Out, format, args...)
	}

	s.WG.Add(1)
	go func() {
		defer s.WG.Done()
		stdin := bufio.NewReader(s.In)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, err := stdin.ReadString('\n')
				if err == io.EOF || strings.TrimSpace(line) == "quit" {
					close(s.InputEOF)
					return
				}
				if err != nil {
					log.Printf("stdin error %s", err)
					return
				}
				if strings.HasPrefix(line, "#") {
					continue
				}
				if s.EchoInput {
					printf("input", "%s", line)
				}
				if s.ShellExpand {
					line, err = ShellExpand(line)
					if err != nil {
						log.Printf("stdin error %s", err)
						return
					}
				}

				var msg interface{}
				if err := json.Unmarshal([]byte(line), &msg); err != nil {
					fmt.Fprintf(os.Stderr, "bad input: %s\n", err)
					continue
				}
				in <- msg
			}
		}
	}()

	out := make(chan *Result)

	s.WG.Add(1)
	go func() {
		defer s.WG.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case r := <-out:
				if r == nil {
					return
				}
				for i, emitted := range r.Emitted {
					for j, msg := range emitted {
						printf("emit", "%d,%d %s\n", i, j, JS(msg))
					}
				}
				for mid, m := range r.Changed {
					printf("update", "%s %s\n", mid, JShort(m))
					if s.state != nil {
						if m.Deleted {
							delete(s.state, mid)
						} else {
							n, have := s.state[mid]
							if !have {
								n = &crew.Machine{}
								s.state[mid] = n
							}
							if m.State != nil {
								n.State = m.State.Copy()
							}
							if m.SpecSrc != nil {
								n.SpecSource = m.SpecSrc.Copy()
							}
						}
					}
				}

				if s.WriteStatePerMsg {
					if err := s.writeState(ctx); err != nil {
						panic(err)
					}
				}
			}
		}

	}()

	return in, out, nil
}

// writeState writes the entire crew as JSON.
func (s *Stdio) writeState(ctx context.Context) error {
	if s.state != nil {
		js, err := json.MarshalIndent(&s.state, "", "  ")
		if err != nil {
			return err
		}
		if err = ioutil.WriteFile(s.StateOutputFilename, js, 0644); err != nil {
			return err
		}
	}
	return nil
}
