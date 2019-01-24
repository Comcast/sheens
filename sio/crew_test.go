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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
{"to":"dc","double":4}
quit
`

	input = fmt.Sprintf(input,
		yaml2json(specPath+"/collatz.yaml"),
		yaml2json(specPath+"/doublecount.yaml"),
		yaml2json(specPath+"/double.yaml"))

	io := NewStdio(true)
	io.In = strings.NewReader(input)

	var buf bytes.Buffer
	io.Out = &buf

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf := &CrewConf{
		Ctl: core.DefaultControl,
	}

	c, err := NewCrew(ctx, conf, io)
	if err != nil {
		t.Fatal(err)
	}

	if err = io.Start(ctx); err != nil {
		t.Fatal(err)
	}

	ms, err := io.Read(ctx)
	if err != nil {
		panic(err)
	}
	for mid, m := range ms {
		if err := c.SetMachine(ctx, mid, m.SpecSource, m.State); err != nil {
			panic(err)
		}
	}

	go func() {
		<-io.InputEOF
		time.Sleep(5 * time.Second)
		cancel()
	}()

	if err := c.Loop(ctx); err != nil {
		t.Fatal(err)
	}

	if err = io.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	fmt.Printf("%s", output)

	{
		want := `{"doubled":20000}`
		if strings.Index(output, want) < 0 {
			t.Fatalf("Didn't see %s", want)
		}
	}

	{
		want := `{"collatz":40}`
		if strings.Index(output, want) < 0 {
			t.Fatalf("Didn't see %s", want)
		}
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
