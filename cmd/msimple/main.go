// A simple, single-machine process that reads from stdin and writes
// to stdout.

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

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/interpreters/goja"

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

		libDir = flag.String("i", ".", "directory containing 'interpreters'")
	)

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Our specs all use the Goja-based interpreter (and only that
	// one).
	gi := goja.NewInterpreter()
	gi.LibraryProvider = goja.MakeFileLibraryProvider(*libDir)
	interpreters := map[string]core.Interpreter{
		"goja": gi,
	}

	// Parse the initial bindings (as JSON).
	var bs core.Bindings
	if err := json.Unmarshal([]byte(*startingBindings), &bs); err != nil {
		panic(err)
	}

	// Read and compile the spec from the given filename.
	specSrc, err := ioutil.ReadFile(*specFilename)
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
		// The state that we'll update as we go.
		st = &core.State{
			NodeName: *startingNode,
			Bs:       bs,
		}

		// Static properties that are exposed via '_.params'
		// to the actions (and guards).
		props = map[string]interface{}{
			"mid": "default",
			"cid": "default",
		}

		// Our standard Walk control.
		ctl = core.DefaultControl
	)

	// Utility functions for processing (and ingesting emitted)
	// messages.
	var (
		process   func(message interface{}) error
		reprocess func(message interface{}) error
	)

	// This function calls itself.
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
			fmt.Printf("# next %s\n", JS(st))
		}

		if err = walked.DoEmitted(reprocess); err != nil {
			return err
		}

		return nil
	}

	// For dealing with messages that were emitted.
	reprocess = func(message interface{}) error {
		js, err := json.Marshal(message)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", js)
		if *recycle {
			return process(message)
		}
		return nil
	}

	in := bufio.NewReader(os.Stdin)
	for {
		line, err := in.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		var message interface{}
		if err = json.Unmarshal(line, &message); err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}

		if *echo {
			fmt.Printf("in: %s\n", JS(message))
		}

		if err = process(message); err != nil {
			fmt.Printf("error: %s\n", err)
		}
	}
}

func JS(x interface{}) string {
	js, err := json.Marshal(&x)
	if err != nil {
		panic(err)
	}
	return string(js)
}
