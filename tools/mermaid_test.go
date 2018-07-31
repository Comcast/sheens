package tools

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/interpreters/goja"

	"github.com/jsccast/yaml"
)

func TestMermaid(t *testing.T) {
	var (
		leaveFile = false
		filename  = "g.mermaid"
		// specFilename = "../specs/homeassistant.yaml"
		specFilename = ""
	)

	out, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}

	if !leaveFile {
		defer func() {
			log.Printf("removing %s", filename)
			if err := os.Remove(filename); err != nil {
				t.Fatal(err)
			}
		}()
	}

	var spec *core.Spec

	if specFilename == "" {
		if spec, err = core.TurnstileSpec(context.Background()); err != nil {
			t.Fatal(err)
		}
	} else {
		interpreters := core.NewInterpretersMap()
		i := goja.NewInterpreter()
		i.LibraryProvider = goja.MakeFileLibraryProvider("")
		interpreters["goja"] = i

		specSrc, err := ioutil.ReadFile(specFilename)
		if err != nil {
			t.Fatal(err)
		}
		if err = yaml.Unmarshal(specSrc, &spec); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		if err = spec.Compile(ctx, interpreters, true); err != nil {
			t.Fatal(err)
		}
	}

	if err := Mermaid(spec, out, nil, "", ""); err != nil {
		t.Fatal(err)
	}

}
