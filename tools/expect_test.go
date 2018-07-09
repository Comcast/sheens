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

package tools

import (
	"context"
	"io/ioutil"
	"os/exec"
	"testing"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/interpreters/goja"

	"github.com/jsccast/yaml"
)

// TestExpectBasic runs a real "expect" test on a real mservice
// process, so another mservice can't be running at the same time.
//
// Requires a current mservice in the path.
//
// If this test hangs, check to see if there's a (unclosed) storage.db
// file.  If there is, remove it.
//
// ToDo: Don't do any of that.
func TestExpectBasic(t *testing.T) {

	// This test requires `cmd/mcrew` in the PATH!  That's not good.
	if _, err := exec.LookPath("mcrew"); err != nil {
		t.Skip(err)
	}

	s := &Session{
		Interpreters: map[string]core.Interpreter{
			"goja": goja.NewInterpreter(),
		},
		Doc:           "A test session",
		ParsePatterns: true,
		IOs: []IO{
			{
				Doc:         "Create a machine, send it a message, and verify the result",
				WaitBetween: 100 * time.Millisecond,
				Inputs: []interface{}{
					`{"cop":{"add":{"m":{"id":"doubler","spec":{"name":"double"}}}}}`,
					`{"cop":{"process":{"message":{"to":{"mid":"doubler"},"double":1}}}}`,
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

	if err := s.Run(ctx, "..", "mcrew", "-v", "-s", "specs", "-l", ".", "-d", "", "-I", "-O", "-h", ""); err != nil {
		panic(err)
	}
}
