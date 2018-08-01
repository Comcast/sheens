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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/tools"

	"github.com/jsccast/yaml"
)

func main() {

	if len(os.Args) < 2 {
		Usage()
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	switch os.Args[1] {
	case "inline", "inlines":

		fs := flag.NewFlagSet("inlines", flag.PanicOnError)
		var dir string
		fs.StringVar(&dir, "d", ".", "directory for referenced files")

		if err := fs.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		bs, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}

		f := func(name string) ([]byte, error) {
			return ioutil.ReadFile(dir + string(os.PathSeparator) + name)
		}

		if bs, err = tools.Inline(bs, f); err != nil {
			panic(err)
		}

		fmt.Printf("%s\n", bs)

	case "macroexpand", "expand":
		bs, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		var x interface{}
		if err = yaml.Unmarshal(bs, &x); err != nil {
			panic(err)
		}

		if x, err = MacroExpand(x); err != nil {
			panic(err)
		}

		if bs, err = yaml.Marshal(&x); err != nil {
			panic(err)
		}

		fmt.Printf("%s\n", bs)

	case "yamltojson":
		pretty := false

		switch len(os.Args) {
		case 2:
		case 3:
			if 2 < len(os.Args) {
				switch os.Args[2] {
				case "-p":
					pretty = true
				default:
					panic(fmt.Sprintf("unsupported args: %v", os.Args[1:]))
				}
			}
		default:
			panic(fmt.Sprintf("unsupported args: %v", os.Args[1:]))
		}

		bs, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if len(bs) == 0 {
			bs = []byte(DefaultSpecYAML)
		}

		var s *core.Spec

		if err = yaml.Unmarshal(bs, &s); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if pretty {
			bs, err = json.MarshalIndent(&s, "  ", "  ")
		} else {
			bs, err = json.Marshal(&s)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if _, err = os.Stdout.Write(bs); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "jsontoyaml":

		bs, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		var s *core.Spec

		if err = json.Unmarshal(bs, &s); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if bs, err = yaml.Marshal(&s); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if _, err = os.Stdout.Write(bs); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	default:

		mod, have := Mods[os.Args[1]]
		if !have {
			fmt.Printf("Unknow subcommand \"%s\"\n", os.Args[1])
			Usage()
			os.Exit(1)
		}

		if err := mod.Flags().Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		bs, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if len(bs) == 0 {
			bs = []byte(DefaultSpecYAML)
		}

		var s *core.Spec

		if err = yaml.Unmarshal(bs, &s); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if err = s.ParsePatterns(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if err := mod.F(s); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if bs, err = yaml.Marshal(&s); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if _, err = os.Stdout.Write(bs); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
}

func Usage() {
	fmt.Printf("Subcommands:\n\n")
	for _, mod := range Mods {
		mod.Flags().Usage()
		fmt.Println("  " + mod.Doc())
		fmt.Println()
	}
	fmt.Println("Usage of yamltojson:")
	// go vet says "Println call ends with newline"!
	fmt.Printf("  -p    pretty-print\n\n")
	fmt.Printf("Usage of jsontoyaml: (no arguments)\n\n")
	fmt.Println("Usage of inlines:")
	fmt.Printf("  -d    directory for source files\n\n")
}

var DefaultSpecYAML = `nodes:
`
