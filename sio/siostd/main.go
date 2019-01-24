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
	"flag"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/sio"
)

func main() {
	io := sio.NewStdio(true)

	flag.BoolVar(&io.EchoInput, "echo", false, "echo input")
	flag.BoolVar(&io.Timestamps, "ts", false, "print timestamps")
	flag.BoolVar(&io.ShellExpand, "sh", false, "shell-expand input")
	flag.BoolVar(&io.PadTags, "pad", false, "pad tags")
	flag.StringVar(&io.StateOutputFilename, "state-out", "", "state output filename")
	flag.BoolVar(&io.WriteStatePerMsg, "write-state-msg", false, "write state after each msg")

	wait := flag.Duration("wait", 0, "wait this long before shutting down couplings")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf := &sio.CrewConf{
		Ctl: core.DefaultControl,
	}

	c, err := sio.NewCrew(ctx, conf, io)
	if err != nil {
		panic(err)
	}

	if err = io.Start(ctx); err != nil {
		panic(err)
	}

	ms, err := io.Read(ctx)
	if err != nil {
		panic(err)
	}
	for mid, m := range ms {
		if err := c.SetMachine(ctx, mid, m.SpecSource, m.State); err != nil {
			panic(err)
		}
	}

	go func() {
		<-io.InputEOF
		time.Sleep(*wait)
		cancel()
	}()

	if err := c.Loop(ctx); err != nil {
		panic(err)
	}

	if err = io.Stop(context.Background()); err != nil {
		panic(err)
	}

}
