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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/sio"
)

func main() {

	var (
		coupling            = flag.String("io", "std", `IO protocol: "std", "mq", or "ws"`)
		stateInputFilename  = flag.String("state-input-filename", "", "Optional name for input JSON state file")
		stateOutputFilename = flag.String("state-output-filename", "state.json", "Optional name for output JSON state file")

		specFile = flag.String("spec-file", "", "Optional spec filename")
		mid      = flag.String("mid", "m", "Machine id for -spec-file (if given)")
		stateJS  = flag.String("state", "{}", "State (JSON) for -spec-file (if given)")

		wait      = flag.Duration("wait", time.Second, "Wait this long before shutting down couplings")
		haltOnEOF = flag.Bool("halt-on-eof", false, "Stop on input EOF")
		verbose   = flag.Bool("v", false, "Verbose")
		help      = flag.Bool("h", false, "Get usage")

		specSource *crew.SpecSource
		state      *core.State
	)

	flag.Parse()

	if *help {
		flag.PrintDefaults()

		{
			fmt.Fprintf(os.Stderr, "\n-io std (default):\n\n")
			_, fs := NewStdCouplings(nil)
			fs.PrintDefaults()
		}

		{
			fmt.Fprintf(os.Stderr, "\n-io mq:\n\n")
			_, fs := NewMQTTCouplings(nil)
			fs.PrintDefaults()
		}

		{
			fmt.Fprintf(os.Stderr, "\n-io ws:\n\n")
			_, fs := NewWebSocketCouplings(nil)
			fs.PrintDefaults()
		}

		{
			fmt.Fprintf(os.Stderr, "\n-io httpds:\n\n")
			_, fs := NewHTTPDCouplings(nil)
			fs.PrintDefaults()
		}

		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cio sio.Couplings
	var store *sio.JSONStore
	switch *coupling {
	case "std":
		c, _ := NewStdCouplings(flag.Args())
		store = &c.JSONStore
		cio = c
	case "mq", "mqtt":
		c, _ := NewMQTTCouplings(flag.Args())
		store = c.JSONStore
		cio = c
	case "ws":
		c, _ := NewWebSocketCouplings(flag.Args())
		store = &c.JSONStore
		cio = c
	case "httpd", "http":
		c, _ := NewHTTPDCouplings(flag.Args())
		store = &c.JSONStore
		// But see hack below to set the crew.
		cio = c
	default:
		panic(fmt.Errorf("unknown io: '%s'", *coupling))
	}

	if store != nil {
		if *stateInputFilename != "" {
			store.StateInputFilename = *stateInputFilename
		}
		if *stateOutputFilename != "" {
			store.StateOutputFilename = *stateOutputFilename
		}
	}

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

	if err := cio.Start(ctx); err != nil {
		panic(err)
	}

	c, err := sio.NewCrew(ctx, conf, cio)
	if err != nil {
		panic(err)
	}
	c.Verbose = *verbose

	// Hack to set crew.
	if h, is := cio.(*HTTPDCouplings); is {
		h.crew = c
	}

	ms, err := cio.Read(ctx)
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
		if std, is := cio.(*sio.Stdio); is {
			<-std.InputEOF // ToDo!
			log.Printf("input EOF (waiting %v)", *wait)
			time.Sleep(*wait)
			cancel()
		}
	}()

	if err := c.Loop(ctx); err != nil {
		panic(err)
	}

	if err = cio.Stop(context.Background()); err != nil {
		log.Printf("error from io.Stop: %v", err)
	}
}

func E(err error, args ...interface{}) error {
	log.Printf("error %s: %v", err, args)
	return err
}
