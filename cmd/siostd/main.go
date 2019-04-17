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

// Package main is a simple single-crew sheens process that reads from
// stdin and writes to stdout.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/sio"
)

func main() {
	io := sio.NewStdio(true)

	flag.BoolVar(&io.EchoInput, "echo", false, "echo input")
	flag.BoolVar(&io.Timestamps, "ts", false, "print timestamps")
	flag.BoolVar(&io.ShellExpand, "sh", false, "shell-expand input")
	flag.BoolVar(&io.PadTags, "pad", false, "pad tags")
	flag.BoolVar(&io.Tags, "tags", true, "tags")
	flag.StringVar(&io.StateOutputFilename, "state-out", "", "state output filename")
	flag.BoolVar(&io.WriteStatePerMsg, "write-state-msg", false, "write state after each msg")
	flag.BoolVar(&io.PrintDiag, "diag", false, "print diagnostic data")

	var (
		specFile  = flag.String("spec-file", "", "optional spec filename")
		mid       = flag.String("mid", "m", "Machine id for -spec-file (if given)")
		stateJS   = flag.String("state", "{}", "State for -spec-file (if given)")
		wait      = flag.Duration("wait", time.Second, "wait this long before shutting down couplings")
		haltOnEOF = flag.Bool("halt-on-eof", false, "stop on input EOF")
		verbose   = flag.Bool("v", false, "verbose")

		specSource *crew.SpecSource
		state      *core.State
	)

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *specFile != "" {
		bs, err := ioutil.ReadFile(*specFile)
		if err != nil {
			panic(err)
		}
		specSource = &crew.SpecSource{
			Source: string(bs),
		}

		if specSource, _, err = sio.ResolveSpecSource(ctx, specSource); err != nil {
			panic(err)
		}
	}

	if *stateJS != "" {
		if err := json.Unmarshal([]byte(*stateJS), &state); err != nil {
			panic(err)
		}
	}

	conf := &sio.CrewConf{
		Ctl:            core.DefaultControl,
		EnableHTTP:     true,
		HaltOnInputEOF: *haltOnEOF,
	}

	c, err := sio.NewCrew(ctx, conf, io)
	if err != nil {
		panic(err)
	}
	c.Verbose = *verbose

	if err = io.Start(ctx); err != nil {
		panic(err)
	}

	ms, err := io.Read(ctx)
	if err != nil {
		panic(err)
	}

	if specSource != nil {
		if err := c.SetMachine(ctx, *mid, specSource, state); err != nil {
			panic(err)
		}
	}

	for mid, m := range ms {
		if err := c.SetMachine(ctx, mid, m.SpecSource, m.State); err != nil {
			panic(err)
		}
	}

	go func() {
		<-io.InputEOF
		log.Printf("input EOF (%v)", *wait)
		time.Sleep(*wait)
		cancel()
	}()

	if err := c.Loop(ctx); err != nil {
		panic(err)
	}

	if err = io.Stop(context.Background()); err != nil {
		log.Printf("error from io.Stop: %v", err)
	}
}
