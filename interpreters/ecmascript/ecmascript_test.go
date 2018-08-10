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

package ecmascript

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/match"
	. "github.com/Comcast/sheens/util/testutil"
)

func TestActionsSimple(t *testing.T) {
	code := `return {likes:"chips"};`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	i := NewInterpreter()
	i.Test = true
	compiled, err := i.Compile(ctx, code)
	if err != nil {
		t.Fatal(err)
	}

	exe, err := i.Exec(ctx, nil, nil, code, compiled)
	if err != nil {
		t.Fatal(err)
	}
	x, have := exe.Bs["likes"]
	if !have {
		t.Fatalf("nothing liked in %#v", exe.Bs)
	}
	s, is := x.(string)
	if !is {
		t.Fatalf("liked %#v is a %T, not a %T", x, x, s)
	}
	if s != "chips" {
		t.Fatalf("didn't want \"%s\"", s)
	}
}

func TestActionsParam(t *testing.T) {
	code := `return {machineId:_.props.mid};`
	props := map[string]interface{}{
		"mid": "simpsons",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	i := NewInterpreter()
	i.Test = true
	compiled, err := i.Compile(ctx, code)
	if err != nil {
		t.Fatal(err)
	}

	exe, err := i.Exec(ctx, nil, props, code, compiled)
	if err != nil {
		t.Fatal(err)
	}
	x, have := exe.Bs["machineId"]
	if !have {
		t.Fatalf("no machineId in %#v", exe.Bs)
	}
	s, is := x.(string)
	if !is {
		t.Fatalf("machineId %#v is a %T, not a %T", x, x, s)
	}
	if s != "simpsons" {
		t.Fatalf("didn't want \"%s\"", s)
	}
}

func TestActionsTimeout(t *testing.T) {
	code := `for (;;) { _.sleep(10); } null;`

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	i := NewInterpreter()
	i.Test = true
	i.Extended = true

	compiled, err := i.Compile(ctx, code)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = i.Exec(ctx, nil, nil, code, compiled); err == nil {
		t.Fatal("didn't timeout")
	}
	msg := err.Error()
	if msg != InterruptedMessage {
		t.Fatalf("surprised by \"%s\"", msg)
	}
}

func TestActionsError(t *testing.T) {
	code := `likes + tacos; null;`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	i := NewInterpreter()
	compiled, err := i.Compile(ctx, code)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = i.Exec(ctx, nil, nil, code, compiled); err == nil {
		t.Fatal("didn't protest")
	}
}

func TestActionsCronNextGood(t *testing.T) {
	cronExpr := "* 0 * * *"
	code := fmt.Sprintf(`({next: _.cronNext("%s")});`, cronExpr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	i := NewInterpreter()
	i.Extended = true
	compiled, err := i.Compile(ctx, code)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = i.Exec(ctx, nil, nil, code, compiled); err != nil {
		t.Fatal(err)
	}
	// ToDo: Parse the result.
}

func TestActionsCronNextBad(t *testing.T) {
	cronExpr := "bad"
	code := fmt.Sprintf(`({next: _.cronNext("%s")});`, cronExpr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	i := NewInterpreter()
	compiled, err := i.Compile(ctx, code)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := i.Exec(ctx, nil, nil, code, compiled); err == nil {
		t.Fatal("didn't protest")
	}
}

func TestActionsMachinePrimitive(t *testing.T) {
	as := core.ActionSource{
		Interpreter: "ecmascript",
		Source:      `var bs = _.bindings; bs.want = "tacos"; return bs;`,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	action, err := as.Compile(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	spec := &core.Spec{
		Name: "test",
		Nodes: map[string]*core.Node{
			"start": {
				Branches: &core.Branches{
					Type: "message",
					Branches: []*core.Branch{
						{
							Pattern: Dwimjs(`{"trigger":"?triggered"}`),
							Target:  "do",
						},
					},
				},
			},
			"do": {
				Action: action,
				Branches: &core.Branches{
					Branches: []*core.Branch{
						{
							Pattern: Dwimjs(`{"want":"tacos"}`),
							Target:  "happy",
						},
						{
							Target: "sad",
						},
					},
				},
			},
			"happy": {},
			"sad":   {},
		},
	}

	if err := spec.Compile(ctx, nil, true); err != nil {
		t.Fatal(err)
	}

	st := &core.State{
		NodeName: "start",
		Bs:       make(match.Bindings),
	}
	ctl := &core.Control{
		Limit: 10,
	}

	walked, err := spec.Walk(ctx, st, []interface{}{Dwimjs(`{"trigger":"do"}`)}, ctl, nil)
	if err != nil {
		t.Fatal(err)
	}

	if walked.To().NodeName != "happy" {
		t.Fatal(walked.To().NodeName)
	}
}

func TestActionsMachineFancy(t *testing.T) {
	spec := &core.Spec{
		Name: "test",
		Nodes: map[string]*core.Node{
			"start": {
				Branches: &core.Branches{
					Type: "message",
					Branches: []*core.Branch{
						{
							Pattern: Dwimjs(`{"trigger":"?triggered"}`),
							Target:  "do",
						},
					},
				},
			},
			"do": {
				ActionSource: &core.ActionSource{
					Interpreter: "ecmascript",
					Source:      `var bs = _.bindings; bs.want = "tacos"; return bs;`,
				},
				Branches: &core.Branches{
					Branches: []*core.Branch{
						{
							Pattern: Dwimjs(`{"want":"tacos"}`),
							Target:  "happy",
						},
						{
							Target: "sad",
						},
					},
				},
			},
			"happy": {},
			"sad":   {},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := spec.Compile(ctx, nil, true); err != nil {
		t.Fatal(err)
	}

	st := &core.State{
		NodeName: "start",
		Bs:       match.NewBindings(),
	}
	ctl := &core.Control{
		Limit: 10,
	}

	walked, err := spec.Walk(ctx, st, []interface{}{Dwimjs(`{"trigger":"do"}`)}, ctl, nil)
	if err != nil {
		t.Fatal(err)
	}

	if walked.To().NodeName != "happy" {
		t.Fatal(walked.To().NodeName)
	}
}

func benchmarkCompiling(b *testing.B, compiling bool) {

	// We have a lot of code, but we only use a little of it.

	code := `

function radians (num) {
  return num * Math.PI / 180;
}

function haversine (lon1,lat1,lon2,lat2) {
  var R = 6371;
  var dLat = radians(lat2-lat1);
  var dLon = radians(lon2-lon1);
  var lat1 = radians(lat1);
  var lat2 = radians(lat2);
  var a = Math.sin(dLat/2) * Math.sin(dLat/2) + Math.sin(dLon/2) * Math.sin(dLon/2) * Math.cos(lat1) * Math.cos(lat2);
  var c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));
  var d = R * c;
  return d;
}

// https://rosettacode.org/wiki/ABC_Problem#ES5
function abc(strWords) {
 
    var strBlocks =
        'BO XK DQ CP NA GT RE TG QD FS JW HU VI AN OB ER FS LY PC ZM',
        blocks = strBlocks.split(' ');
 
    function abc(lstBlocks, strWord) {
        var lngChars = strWord.length;
 
        if (!lngChars) return [];
 
        var b = lstBlocks[0],
            c = strWord[0];
 
        return chain(lstBlocks, function (b) {
            return (b.indexOf(c.toUpperCase()) !== -1) ? [
                (b + ' ').concat(
                    abc(removed(b, lstBlocks), strWord.slice(1)))
            ] : [];
        })
    }
 
    // Monadic bind (chain) for lists
    function chain(xs, f) {
        return [].concat.apply([], xs.map(f));
    }
 
    // a -> [a] -> [a]
    function removed(x, xs) {
        var h = xs.length ? xs[0] : null,
            t = h ? xs.slice(1) : [];
 
        return h ? (
            h === x ? t : [h].concat(removed(x, t))
        ) : [];
    }
 
    function solution(strWord) {
        var strAttempt = abc(blocks, strWord)[0].split(',')[0];
 
        // two chars per block plus one space -> 3
        return strWord + ((strAttempt.length === strWord.length * 3) ?
            ' -> ' + strAttempt : ': [no solution]');
    }
 
    return strWords.split(' ').map(solution).join('\n');
 
}

function bar() { return "chips"; }

({likes:bar()});
`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	i := NewInterpreter()
	i.Test = true

	var compiled interface{}
	if compiling {
		var err error
		if compiled, err = i.Compile(ctx, code); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		if _, err := i.Exec(context.Background(), nil, nil, code, compiled); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPrecompile(b *testing.B) {
	benchmarkCompiling(b, true)
}

func BenchmarkNoPrecompile(b *testing.B) {
	benchmarkCompiling(b, false)
}

func TestActionsOutNaN(t *testing.T) {
	code := `_.out(NaN); return {};`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	i := NewInterpreter()
	compiled, err := i.Compile(ctx, code)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = i.Exec(ctx, nil, nil, code, compiled); err == nil {
		t.Fatal("expected an error")
	}
}

func TestActionsModifyBindingValue(t *testing.T) {
	bs := match.NewBindings()
	bs["likes"] = map[string]interface{}{
		"weekdays": "tacos",
		"weekends": "chips",
	}

	code := `_.bindings.likes.weekends = "queso"; throw "a fit";`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	i := NewInterpreter()
	i.Test = true
	compiled, err := i.Compile(ctx, code)
	if err != nil {
		t.Fatal(err)
	}

	exe, _ := i.Exec(ctx, bs, nil, code, compiled)
	// Ignore the error.  We want to see if the action had a side
	// effect.

	x, have := bs["likes"]
	if !have {
		t.Fatalf("nothing liked in %#v", exe.Bs)
	}
	m, is := x.(map[string]interface{})
	if !is {
		t.Fatalf("liked %#v is a %T, not a %T", x, x, m)
	}
	y, have := m["weekends"]
	if !have {
		t.Fatalf("nothing liked on weekends in  %#v", exe.Bs)
	}
	s, is := y.(string)
	if !is {
		t.Fatalf("liked %#v is a %T, not a %T", y, y, s)
	}
	if s != "chips" {
		t.Fatalf("didn't want \"%s\"", s)
	}
}
