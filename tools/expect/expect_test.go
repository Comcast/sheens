/* Copyright 2018-2019 Comcast Cable Communications Management, LLC
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


package expect

import (
	"context"
	"io/ioutil"
	"os/exec"
	"testing"

	"github.com/Comcast/sheens/interpreters"

	"github.com/jsccast/yaml"
)

// TestExpectBasic runs specs/test/double.test.yaml with a real siostd
// process.
//
// Requires a current siostd in the path.  Sorry.
func TestExpectBasic(t *testing.T) {

	root := "../.."

	// This test requires `cmd/siosstd` in the PATH!  That's not good.
	if _, err := exec.LookPath("siostd"); err != nil {
		t.Skip(err)
	}

	bs, err := ioutil.ReadFile(root + "/specs/tests/double.test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var s *Session
	if err := yaml.Unmarshal(bs, &s); err != nil {
		t.Fatal(err)
	}
	s.Interpreters = interpreters.Standard()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if false {
		s.ShowStderr = true
		s.ShowStdout = true
		s.ShowStdin = true
	}

	if err := s.Run(ctx, root, "siostd", "-tags=false"); err != nil {
		panic(err)
	}
}
