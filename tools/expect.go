package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Comcast/sheens/core"
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
	Bindingss []core.Bindings `json:"-" yaml:"-"`
}

// IO is a package of input messages and required output message
// specifications.
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
	Interpreters map[string]core.Interpreter `json:"-" yaml:"-"`

	// DefaultTimeout is the default timeout for each IO.
	DefaultTimeout time.Duration `json:"defaultTimeout,omitempty" yaml:"defaultTimeout,omitempty"`

	// ShowStderr controls whether the subprocess's stderr is
	// logged.
	ShowStderr bool `json:"showStderr,omitempty" yaml:"showStderr,omitempty"`

	ShowStdin bool `json:"showStdin,omitempty" yaml:"showStdin,omitempty"`

	ShowStdout bool `json:"showStdout,omitempty" yaml:"showStdout,omitempty"`

	Verbose bool `json:"verbose,omitempty" yaml:"verbose,omitempty"`
}

// Run processes all the IOs in the Sesson.
//
// The current directory is changed to 'dir'.
//
// The subprocess is given by the args. The first arg is the
// executable.
func (s *Session) Run(ctx context.Context, dir string, args ...string) error {

	if dir != "" {
		if err := os.Chdir(dir); err != nil {
			return err
		}
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
				log.Printf("stderr error %s", err)
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
			timer    *time.Timer
			happy    = errors.New("happy")
			timeout  = errors.New("timeout")
			canceled = errors.New("canceled")
			errs     = make(chan error, 3) // At least three
		)

		if 0 < iop.Timeout {
			timer = time.AfterFunc(iop.Timeout, func() {
				errs <- timeout
			})
		}

		// Process stdout.
		go func() {
			f := func() error {
				need := len(iop.OutputSet)

				for {
					line, err := out.ReadBytes('\n')
					if err != nil {
						return err
					}

					if s.ShowStdout {
						log.Printf("out %s", line)
					}

					var message interface{}
					if err = json.Unmarshal(line, &message); err != nil {
						log.Printf("ignoring %s", line)
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
									return err
								}
							} else {
								if js, err = json.Marshal(&pattern); err != nil {
									return err
								}
							}

							bss, err := core.Match(nil, pattern, message, core.NewBindings())
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
									bss = []core.Bindings{exe.Bs}
								}
							}
							if bss != nil {
								need--
								output.Bindingss = bss
							}
						}
						if need == 0 {
							return nil
						}
					}
				}
			}

			err := f()
			if timer != nil {
				timer.Stop()
			}
			if err == nil {
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

		happies := 0
		want := 2

	LOOP:
		for {

			select {
			case <-ctx.Done():
				return canceled
			case err = <-errs:
				switch err {
				case happy:
					happies++
					if want <= happies {
						break LOOP
					}
				default:
					break LOOP
				}
			}
		}

		if happies < want {
			return err
		}
	}

	if err := stdin.Close(); err != nil {
		log.Printf("stdin.Close() error %s", err)
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
