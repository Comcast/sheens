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
	"flag"

	"github.com/Comcast/sheens/sio"
)

func NewStdCouplings(args []string) (*sio.Stdio, *flag.FlagSet) {

	var (
		std = sio.NewStdio(true)
		fs  = flag.NewFlagSet("std", flag.ExitOnError)
	)

	fs.BoolVar(&std.EchoInput, "echo", false, "echo input")
	fs.BoolVar(&std.Timestamps, "ts", false, "print timestamps")
	fs.BoolVar(&std.ShellExpand, "sh", false, "shell-expand input")
	fs.BoolVar(&std.PadTags, "pad", false, "pad tags")
	fs.BoolVar(&std.Tags, "tags", true, "tags")
	fs.StringVar(&std.StateOutputFilename, "state-out", "", "state output filename")
	fs.BoolVar(&std.WriteStatePerMsg, "write-state-msg", false, "write state after each msg")
	fs.BoolVar(&std.PrintDiag, "diag", false, "print diagnostic data")

	if args != nil {
		fs.Parse(args)
	}

	return std, fs
}
