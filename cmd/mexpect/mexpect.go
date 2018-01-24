package main

import (
	"context"
	"flag"
	"io/ioutil"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/interpreters/goja"
	"github.com/Comcast/sheens/tools"

	"github.com/jsccast/yaml"
)

func main() {

	var (
		inputFilename = flag.String("f", "specs/tests/double.test.yaml", "filename for test session")
		dir           = flag.String("d", ".", "working directory")
		showStderr    = flag.Bool("e", true, "show subprocess stderr")
		timeout       = flag.Duration("t", 10*time.Second, "main timeout")

		specDir = flag.String("s", "specs", "specs directory")
		libDir  = flag.String("i", ".", "directory containing 'interpreters'")
	)

	flag.Parse()

	bs, err := ioutil.ReadFile(*inputFilename)
	if err != nil {
		panic(err)
	}

	var s tools.Session
	if err = yaml.Unmarshal(bs, &s); err != nil {
		panic(err)
	}

	s.Interpreters = map[string]core.Interpreter{
		"goja": goja.NewInterpreter(),
	}
	s.ShowStderr = *showStderr

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if err = s.Run(ctx, *dir, "mcrew", "-v", "-s", *specDir, "-l", *libDir, "-d", "", "-I", "-O", "-h", ""); err != nil {
		panic(err)
	}
}
