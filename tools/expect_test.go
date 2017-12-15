package tools

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/interpreters/goja"

	"github.com/jsccast/yaml"
)

func TestExpectBasic(t *testing.T) {
	// This test requires `cmd/mserivce`!  That's not good.

	s := &Session{
		Interpreters: map[string]core.Interpreter{
			"goja": goja.NewInterpreter(),
		},
		Doc:           "A test session",
		ParsePatterns: true,
		IOs: []IO{
			{
				Doc:        "Create a crew and wait to hear that that worked",
				WaitBefore: 100 * time.Millisecond,
				Inputs: []interface{}{
					`{"make":"simpsons"}`,
				},
				OutputSet: []Output{
					{
						Pattern: `{"make":"simpsons"}`,
					},
				},
			},
			{
				Doc:         "Create a machine, send it a message, and verify the result",
				WaitBetween: 100 * time.Millisecond,
				Inputs: []interface{}{
					`{"cop":{"cid":"simpsons","add":{"m":{"id":"doubler","spec":{"name":"double"}}}}}`,
					`{"cop":{"cid":"simpsons","process":{"message":{"to":{"mid":"doubler"},"double":1}}}}`,
				},
				OutputSet: []Output{
					{
						Pattern: `{"doubled":2}`,
					},
					{
						Doc:     "Just an example of using a guard.",
						Pattern: `{"doubled":"?n"}`,
						GuardSource: &core.ActionSource{
							Interpreter: "goja",
							Source:      "var bs = _.bindings; if (bs.n != 2) { bs = null; } bs;",
						},
					},
				},
			},
		},
	}

	{
		bs, err := yaml.Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		if err = ioutil.WriteFile("../double.test.yaml", bs, 0644); err != nil {
			t.Fatal(err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.ShowStderr = true
	if err := s.Run(ctx, "..", "mservice", "-r", "-s", "specs", "-i", "."); err != nil {
		t.Fatal(err)
	}
}
