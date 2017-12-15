package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Comcast/sheens/core"

	"github.com/jsccast/yaml"
)

func main() {

	if len(os.Args) < 2 {
		Usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "expand":
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
}

var DefaultSpecYAML = `nodes:
`
