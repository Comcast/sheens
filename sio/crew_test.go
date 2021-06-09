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

package sio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Comcast/sheens/core"

	"github.com/jsccast/yaml"
)

func TestCrew(t *testing.T) {

	specPath := "../specs"

	for _, spec := range []string{"collatz", "doublecount", "double"} {
		filename := fmt.Sprintf("%s/%s.yaml", specPath, spec)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Logf("%s isn't available", filename)
			t.Skip(t)
		}
	}

	log.Printf("This test takes a few seconds.")

	input := `{"to":"captain","update":{"c":{"spec":{"inline":%s}}}}
{"to":"captain","update":{"dc":{"spec":{"inline":%s}}}}
{"to":"captain","update":{"d":{"spec":{"inline":%s}}}}
{"double":10}
{"double":100}
{"double":1000}
{"collatz":17}
{"collatz":5}
{"to":"timers","makeTimer":{"in":"2s","msg":{"double":10000,"from":"timer"},"id":"t0"}}
{"to":"timers","makeTimer":{"in":"4s","msg":{"collatz":13, "from":"timer"},"id":"t1"}}
{"to":"d","double":3}
{"to":"dc","double":4}`

	input = fmt.Sprintf(input,
		yaml2json(specPath+"/collatz.yaml"),
		yaml2json(specPath+"/doublecount.yaml"),
		yaml2json(specPath+"/double.yaml"))

	sio := NewStdio(true)
	ri, wi := io.Pipe()
	sio.In = ri
	ro, wo := io.Pipe()
	sio.Out = wo

	ctx, cancel := context.WithCancel(context.Background())

	if err := sio.Start(ctx); err != nil {
		t.Fatal(err)
	}

	conf := &CrewConf{
		Ctl: core.DefaultControl,
	}

	c, err := NewCrew(ctx, conf, sio)
	if err != nil {
		t.Fatal(err)
	}

	ms, err := sio.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for mid, m := range ms {
		if err := c.SetMachine(ctx, mid, m.SpecSource, m.State); err != nil {
			t.Fatal(err)
		}
	}

	// Send input to crew.
	go func() {
		for _, line := range strings.Split(input, "\n") {
			fmt.Fprintf(wi, "%s\n", line)
		}
	}()

	// Read output from crew.
	need1 := true
	need2 := true
	go func() {
		out := bufio.NewReader(ro)
		for {
			line, err := out.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			log.Printf("heard %s", line)

			if 0 <= strings.Index(line, `{"doubled":20000}`) {
				need1 = false
			}

			if 0 <= strings.Index(line, `{"collatz":40}`) {
				need2 = false
			}
		}
	}()

	// In a few seconds, shut us down.
	go func() {
		time.Sleep(6 * time.Second)
		cancel()
	}()

	if err := c.Loop(ctx); err != nil {
		t.Fatal(err)
	}

	fmt.Fprintf(wi, "quit\n")

	if err = sio.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}

	if need1 {
		t.Fatal(1)
	}
	if need2 {
		t.Fatal(2)
	}

}

// yaml2json reads the file with the given name, parses the contents
// as YAML, and returns a string of a JSON presentation of the object.
//
// When something goes wrong, this function just panics.
func yaml2json(filename string) string {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var x interface{}
	if err = yaml.Unmarshal(bs, &x); err != nil {
		panic(err)
	}
	js, err := json.Marshal(&x)
	if err != nil {
		panic(err)
	}
	return string(js)
}
