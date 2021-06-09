/* Copyright 2018-2019 Comcast Cable Communications Management, LLC
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

// Package expect is a tool for testing machine specifications.
//
// You construct a Session, which has inputs and expected outputs.
// Then run the session to see if the expected outputs actually
// appeared.
//
// Specifying what's expect can be simple, as in some literal output,
// or fairly fancy, as in code that computes some property.
//
// This package also has support for delays, timeouts, and other
// time-driven behavior.
//
// See ../../cmd/mexpect for command-line use.
package expect

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/match"
	. "github.com/Comcast/sheens/util/testutil"
)

// Output is a specification for a message that's expected.
type Output struct {
	// Doc is an opaque documentation string.
	Doc string `json:"doc,omitempty" yaml:"doc,omitempty"`

	// Pattern must be matched by an emitted message.
	Pattern interface{} `json:"pattern,omitempty" yaml:"pattern,omitempty"`

	// Guard is an optional guard (as in a Machine spec) that is
	// called to execute procedural code to verify the bindings
	// after a match.
	Guard core.Action `json:"-" yaml:"-"`

	// GuardSource is optional source that will be compiled to the
	// Guard.
	GuardSource *core.ActionSource `json:"guard,omitempty" yaml:"guardSource,omitempty"`

	// Bindings, which is the result of a match (and optional
	// guard) is written during processing.  Just for diagnostics.
	Bindingss []match.Bindings `json:"bs,omitempty" yaml:"bs,omitempty"`

	// Inverted means that matching output isn't desired!
	Inverted bool `json:"inverted,omitempty" yaml:"inverted,omitempty"`
}

// IO is a package of input messages and required output message
// specifications.
//
// This struct includes a list of messages to send and a set of expect
// output messages.
type IO struct {
	// Doc is an opaque documentation string.
	Doc string `json:"doc,omitempty" yaml:"doc,omitempty"`

	// WaitBefore is the time to wait before sending the first message.
	WaitBefore time.Duration `json:"waitBefore,omitempty" yaml:"waitBefore,omitempty"`

	// Waitbefore is the time to wait between sending messages.
	WaitBetween time.Duration `json:"waitBetween,omitempty" yaml:"waitBetween,omitempty"`

	// Inputs are the messages to send.
	Inputs []interface{} `json:"inputs,omitempty" yaml:"inputs,omitempty"`

	// WaitAfter is the time to wait after sending the last
	// message.
	WaitAfter time.Duration `json:"waitAfter,omitempty" yaml:"waitAfter,omitempty"`

	// OutputSet is the set (not a list) of outputs to verify.
	OutputSet []Output `json:"outputSet,omitempty" yaml:"outputSet,omitempty"`

	// Timeout is the optional timeout for this set.
	// Session.DefaultTimeout is the default value.
	Timeout time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// Session is mostly a sequence of IOs.
type Session struct {
	// Doc is an opaque documentation string.
	Doc string `json:"doc,omitempty" yaml:"doc,omitempty"`

	// IOs is sequence of IOs that this session will run.
	IOs []IO `json:"ios" yaml:"ios"`

	// ParsePatterns will parse IO.OutputSet.Patterns as JSON.
	ParsePatterns bool `json:"parsePatterns,omitempty" yaml:"parsePatterns,omitempty"`

	// Interpreters are used (if necessary) to compile any
	// GuardSources.
	Interpreters core.InterpretersMap `json:"-" yaml:"-"`

	// DefaultTimeout is the default timeout for each IO.
	DefaultTimeout time.Duration `json:"defaultTimeout,omitempty" yaml:"defaultTimeout,omitempty"`

	// ShowStderr controls whether the subprocess's stderr is
	// logged.
	ShowStderr bool `json:"showStderr,omitempty" yaml:"showStderr,omitempty"`

	// ShowStdin controls whether the subprocess's stdin is
	// logged.
	ShowStdin bool `json:"showStdin,omitempty" yaml:"showStdin,omitempty"`

	// ShowStdout controls whether the subprocess's stdout is
	// logged.
	ShowStdout bool `json:"showStdout,omitempty" yaml:"showStdout,omitempty"`

	// InputPrefix specifies the prefix of input lines that should
	// be consumed.
	InputPrefix string `json:"inputPrefix,omitempty" yaml:"inputPrefix,omitempty"`

	Verbose bool `json:"verbose,omitempty" yaml:"verbose,omitempty"`
}

// Run processes all the IOs in the Session.
//
// The current directory is changed to 'dir' (and then hopefully
// restored).
//
// The subprocess is given by the args. The first arg is the
// executable.  Example args:
//
//   "siostd", "-spec-file", "specs/double.yaml"
//
func (s *Session) Run(ctx context.Context, dir string, args ...string) error {

	ctx, cancel := context.WithCancel(ctx)

	if dir != "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		if err := os.Chdir(dir); err != nil {
			return err
		}
		// Far from perfect ...
		defer func() {
			if err := os.Chdir(cwd); err != nil {
				log.Printf("error restoring cwd %s", cwd)
			}
		}()
	}

	if len(args) == 0 {
		return fmt.Errorf("need a command (and optional args) (for expect.Session.Run)")
	}

	cmd := exec.Command(args[0], args[1:]...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()
	out := bufio.NewReader(stdout)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		return err
	}

	newline := []byte{'\n'}

	// Log subprocess's stderr.
	go func() {
		out := bufio.NewReader(stderr)
		for {
			line, err := out.ReadBytes('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				if strings.Index(err.Error(), "already closed") < 0 {
					log.Printf("stderr error %s", err)
				}
				break
			}
			if s.ShowStderr {
				log.Printf("stderr %s", line)
			}
		}
	}()

	for _, iop := range s.IOs {

		if iop.Timeout == 0 {
			iop.Timeout = s.DefaultTimeout
		}

		var (
			errs = make(chan error, 4)

			happy    = errors.New("happy")
			timeout  = errors.New("timeout")
			canceled = errors.New("canceled")
		)

		if 0 < iop.Timeout {
			time.AfterFunc(iop.Timeout, func() {
				errs <- timeout
				errs <- timeout
			})
		}

		// Consume stdout.
		go func() {
			f := func() error {

				need := 0
				for _, o := range iop.OutputSet {
					if !o.Inverted {
						need++
					}
				}

				for 0 < need {
					line, err := out.ReadBytes('\n')
					if err != nil {
						return err
					}

					if s.ShowStdout {
						log.Printf("out %s", line)
					}

					if bytes.HasPrefix(line, []byte(s.InputPrefix)) {
						line = bytes.TrimSpace(line[len(s.InputPrefix):])
					}

					var message interface{}
					if err = json.Unmarshal(line, &message); err != nil {
						log.Printf("ignoring '%s'", line)
						continue
					} else {
						for _, output := range iop.OutputSet {
							if output.Bindingss != nil {
								continue
							}
							var pattern = output.Pattern
							var js []byte
							if s.ParsePatterns {
								js = []byte(pattern.(string))
								if err = json.Unmarshal(js, &pattern); err != nil {
									return fmt.Errorf("Unmarshal error %v for %s", err, js)
								}
							} else {
								if js, err = json.Marshal(&pattern); err != nil {
									return err
								}
							}

							bss, err := match.Match(pattern, message, match.NewBindings())
							if err != nil {
								return err
							}
							if bss != nil {
								if 1 < len(bss) {
									log.Printf("warning: multiple Bindingss")
								}

								if output.GuardSource != nil {
									if output.Guard, err = output.GuardSource.Compile(ctx, s.Interpreters); err != nil {
										return err
									}
								}

								if output.Guard != nil {
									exe, err := output.Guard.Exec(ctx, bss[0], nil)
									if err != nil {
										return err
									}
									bss = []match.Bindings{exe.Bs}
								}
							}
							if bss != nil {
								output.Bindingss = bss
								if output.Inverted {
									return fmt.Errorf("undesired output %s", JS(output))
								}
								need--
							}
						}
					}
				}

				return nil
			}

			if err := f(); err == nil {
				errs <- happy
			} else {
				errs <- err
			}
		}()

		// Send messages to stdin.
		go func() {

			f := func() error {
				s.pause("waitBefore", iop.WaitBefore)

				for i, input := range iop.Inputs {
					if 0 < i {
						s.pause("waitBetween", iop.WaitBetween)
					}
					js := []byte(input.(string))

					if s.ShowStdin {
						log.Printf("in %s\n", js)
					}

					if _, err := stdin.Write(js); err != nil {
						return err
					}

					if _, err := stdin.Write(newline); err != nil {
						return err
					}
				}

				s.pause("waitAfter", iop.WaitAfter)
				return nil
			}

			if err := f(); err == nil {
				errs <- happy
			} else {
				errs <- err
			}
		}()

		// Wait until we are done.

		happies := 0
		want := 2

	LOOP:
		for happies < want {
			select {
			case <-ctx.Done():
				return canceled
			case err = <-errs:
				switch err {
				case happy:
					happies++
				default:
					break LOOP
				}
			}
		}

		if happies < want {
			return err
		}
	}

	cancel()

	if err := stdin.Close(); err != nil {
		log.Printf("stdin.Close() error %s", err)
	}

	if err := stdout.Close(); err != nil {
		log.Printf("stdout.Close() error %s", err)
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *Session) pause(why string, d time.Duration) {
	if 0 < d {
		if s.Verbose {
			log.Printf("pause %s %s", why, d)
		}
		time.Sleep(d)
	}
}
