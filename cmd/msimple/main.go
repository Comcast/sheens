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

// Package main is a simple program that runs a single machine.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Comcast/sheens/core"
	ints "github.com/Comcast/sheens/interpreters"
	"github.com/Comcast/sheens/match"
	"github.com/Comcast/sheens/tools"

	"github.com/jsccast/yaml"
)

func main() {

	var (
		specFilename     = flag.String("s", "", "specs filename (YAML)")
		startingNode     = flag.String("n", "start", "starting node")
		startingBindings = flag.String("b", "{}", "starting bindings (in JSON)")

		recycle = flag.Bool("r", true, "ingest emitted messages")
		diag    = flag.Bool("d", false, "print diagnostics")
		echo    = flag.Bool("e", false, "echo input messages")
	)

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interpreters := ints.Standard()

	// Parse the given initial bindings (as JSON).
	var bs match.Bindings
	if err := json.Unmarshal([]byte(*startingBindings), &bs); err != nil {
		panic(err)
	}

	// Read and compile the spec from the given filename.
	specSrc, err := tools.ReadFileWithInlines(*specFilename)
	if err != nil {
		panic(err)
	}
	var spec core.Spec
	if err = yaml.Unmarshal(specSrc, &spec); err != nil {
		panic(err)
	}
	if err = spec.Compile(ctx, interpreters, true); err != nil {
		panic(err)
	}

	// Set up our execution environment.
	var (
		// The machine's state that we'll update as we go.
		st = &core.State{
			NodeName: *startingNode,
			Bs:       bs,
		}

		// Some static properties that are exposed to actions
		// (and guards) via '_.params'
		props = map[string]interface{}{
			"mid": "default",
			"cid": "default",
		}

		// Our standard Walk control.
		ctl = core.DefaultControl
	)

	// Utility functions for processing (and ingesting emitted)
	// messages.  These functions call themselves mututally
	// recursively, so we define them this clumsy way.
	var (
		// process sends a message to the machine.
		process func(message interface{}) error

		// reprocess takes an emitted message, prints it, and
		// optionally sends the message back to the machine as
		// an in-bound message (via process()).
		reprocess func(message interface{}) error
	)

	{
		process = func(message interface{}) error {

			walked, err := spec.Walk(ctx, st, []interface{}{message}, ctl, props)
			if err != nil {
				return err
			}

			if *diag {
				fmt.Printf("# walked\n")
				fmt.Printf("#   message    %s\n", JS(message))
				if walked.Error != nil {
					fmt.Printf("#   error    %v\n", walked.Error)
				}
				for i, stride := range walked.Strides {
					fmt.Printf("#   %02d from     %s\n", i, JS(stride.From))
					fmt.Printf("#      to       %s\n", JS(stride.To))
					if stride.Consumed != nil {
						fmt.Printf("#      consumed %s\n", JS(stride.Consumed))
					}
					if 0 < len(stride.Events.Emitted) {
						fmt.Printf("#      emitted\n")
					}
					for _, emitted := range stride.Events.Emitted {
						fmt.Printf("#         %s\n", JS(emitted))
					}
				}
			}

			if walked.Error != nil {
				return err
			}

			if next := walked.To(); next != nil {
				st = next
			}

			if *diag {
				// If we had persistence, we'd
				// probably write out the new state
				// here.
				fmt.Printf("# next %s\n", JS(st))
			}

			// For each "emitted" message, reprocess it.
			if err = walked.DoEmitted(reprocess); err != nil {
				return err
			}

			return nil
		}

		reprocess = func(message interface{}) error {
			js, err := json.Marshal(message)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", js)

			if err = handle(ctx, message, process); err != nil {
				return err
			}

			if *recycle {
				return process(message)
			}

			return nil
		}
	}

	// We can accept input like "sleep 1s" to pause for 1 second.
	// We'll check for that kind of input with this regexp.
	sleeper := regexp.MustCompile("sleep +(.*)")

	in := bufio.NewReader(os.Stdin)
	for {
		line, err := in.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		{
			// We can accept some non-message input.

			s := strings.TrimSpace(string(line))
			if strings.HasPrefix(s, "#") {
				// Comment input.
				continue
			}
			if s == "timers" {
				// Show pending timers.
				timers.Range(func(k, v interface{}) bool {
					t := v.(*timer)
					fmt.Printf("# %s %s %s\n", t.id, t.at.Format(time.RFC3339), JS(t.message))
					return true
				})
				continue
			}
			if ss := sleeper.FindStringSubmatch(s); ss != nil {
				// A request to sleep.
				d, err := time.ParseDuration(ss[1])
				if err != nil {
					warn(err)
					continue
				}
				time.Sleep(d)
				continue
			}
		}

		// Parse the input line as message in JSON.
		var message interface{}
		if err = json.Unmarshal(line, &message); err != nil {
			warn(err)
			continue
		}

		if *echo {
			fmt.Printf("in: %s\n", JS(message))
		}

		// Allow input to make and cancel timers.
		if err = handle(ctx, message, process); err != nil {
			warn(err)
		}

		if err = process(message); err != nil {
			warn(err)
		}
	}
}
