package core

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	. "github.com/Comcast/sheens/util/testutil"
)

func TestStepSimple(t *testing.T) {
	count := 0

	spec := &Spec{
		Name:          "test",
		PatternSyntax: "json",
		Nodes: map[string]*Node{
			"start": {
				Branches: &Branches{
					Type: "message",
					Branches: []*Branch{
						{
							Pattern: `{"trigger":"?triggered"}`,
							Target:  "do",
						},
					},
				},
			},
			"do": {
				Action: &FuncAction{
					F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
						count++
						e := NewExecution(bs)
						e.Events.AddEmitted("tacos")
						e.Events.AddTrace("queso")
						return e, nil
					},
				},
				Branches: &Branches{
					Branches: []*Branch{
						{
							Target: "start",
						},
					},
				},
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := spec.Compile(ctx, nil, true); err != nil {
		t.Fatal(err)
	}

	st := &State{
		NodeName: "start",
		Bs:       make(Bindings),
	}

	c := &Control{
		Limit: 10,
	}

	stride, err := spec.Step(ctx, st, Dwimjs(`{"trigger":"do"}`), c, nil)
	if err != nil {
		t.Fatal(err)
	}

	st = stride.To

	if stride, err = spec.Step(ctx, st, nil, c, nil); err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("count == %d", count)
	}

}

func TestActionErrors(t *testing.T) {
	spec := &Spec{
		Name:          "test",
		PatternSyntax: "json",
		Nodes: map[string]*Node{
			"start": {
				Action: &FuncAction{
					F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
						return nil, fmt.Errorf("something terrible happened")
					},
				},
				Branches: &Branches{
					Branches: []*Branch{
						{
							Pattern: Dwimjs(`{"actionError":"?err"}`),
							Target:  "handle",
						},
						{
							Target: "start",
						},
					},
				},
			},
			"handle": {
				Action: &FuncAction{
					F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
						return NewExecution(make(Bindings)), nil
					},
				},
				Branches: &Branches{
					Branches: []*Branch{
						{
							Target: "recovered",
						},
					},
				},
			},
			"recovered": {},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := spec.Compile(ctx, nil, true); err != nil {
		t.Fatal(err)
	}

	f := func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		st := &State{
			NodeName: "start",
			Bs:       make(Bindings),
		}

		c := &Control{
			Limit: 10,
		}

		walked, err := spec.Walk(ctx, st, nil, c, nil)
		if err != nil {
			t.Fatal(err)
		}

		to := walked.To()
		if to == nil {
			to = &State{}
		}
		if spec.ActionErrorBranches {
			if to.NodeName != "recovered" {
				t.Fatalf("went to '%s' instead of 'handle'", to.NodeName)
			}
			if _, have := to.Bs["actionError"]; have {
				t.Fatal("should have eliminated actionError")
			}
		} else {
			if to.NodeName != "error" {
				t.Fatalf("went to '%s' instead of 'error'", to.NodeName)
			}
			if _, have := to.Bs["actionError"]; !have {
				t.Fatal("no actionError: " + JS(to.Bs))
			}
		}
	}

	spec.ActionErrorNode = "error"
	t.Run("not handling", f)
	spec.ActionErrorBranches = true
	t.Run("handling", f)
}

func TestTurnstile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	spec, err := TurnstileSpec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	c := &Control{
		Limit: 10,
	}

	mes := []struct {
		Message  string
		Expected string
	}{
		{
			Message:  `{"input":"coin"}`,
			Expected: "unlocked",
		},
		{
			Message:  `{"input":"push"}`,
			Expected: "locked",
		},
		{
			Message:  `{"input":"push"}`,
			Expected: "locked",
		},
		{
			Message:  `{"input":"coin"}`,
			Expected: "unlocked",
		},
		{
			Message:  `{"input":"coin"}`,
			Expected: "unlocked",
		},
		{
			Message:  `{"input":"push"}`,
			Expected: "locked",
		},
	}

	st := &State{
		NodeName: "locked",
		Bs:       make(Bindings),
	}

	for i, me := range mes {
		pending := []interface{}{
			Dwimjs(me.Message),
		}
		walked, err := spec.Walk(ctx, st, pending, c, nil)
		if err != nil {
			t.Fatal(err)
		}

		st := walked.To()
		if st.NodeName != me.Expected {
			t.Fatalf(`%d expected "%s" but found "%s"`, i, me.Expected, st.NodeName)
		}
	}
}

func TestWalkLimit(t *testing.T) {
	count := 0
	spec := &Spec{
		Name:          "test",
		PatternSyntax: "json",
		Nodes: map[string]*Node{
			"start": {
				Branches: &Branches{
					Type: "message",
					Branches: []*Branch{
						{
							Pattern: `{"trigger":"?triggered"}`,
							Target:  "loop",
						},
					},
				},
			},
			"loop": {
				Action: &FuncAction{
					F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
						count++
						return NewExecution(make(Bindings)), nil
					},
				},
				Branches: &Branches{
					Branches: []*Branch{
						{
							Target: "loop",
						},
					},
				},
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := spec.Compile(ctx, nil, true); err != nil {
		t.Fatal(err)
	}

	st := &State{
		NodeName: "start",
		Bs:       make(Bindings),
	}

	c := &Control{
		Limit: 10,
	}

	pending := []interface{}{
		Dwimjs(`{"trigger":"do"}`),
	}

	walked, err := spec.Walk(ctx, st, pending, c, nil)
	if err != nil {
		t.Fatal(err)
	}

	if walked.StoppedBecause != Limited {
		t.Fatalf("bad reason: %v", walked.StoppedBecause)
	}
}

func TestWalkBreakpoint(t *testing.T) {
	spec := &Spec{
		Name:          "test",
		PatternSyntax: "json",
		Nodes: map[string]*Node{
			"start": {
				Branches: &Branches{
					Type: "message",
					Branches: []*Branch{
						{
							Pattern: `{"trigger":"?triggered"}`,
							Target:  "loop",
						},
					},
				},
			},
			"loop": {
				Action: &FuncAction{
					F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
						x, have := bs["n"]
						if !have {
							x = 0
						}
						bs = Bindings{
							"n": x.(int) + 1,
						}
						return NewExecution(bs), nil
					},
				},
				Branches: &Branches{
					Branches: []*Branch{
						{
							Target: "loop",
						},
					},
				},
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := spec.Compile(ctx, nil, true); err != nil {
		t.Fatal(err)
	}

	st := &State{
		NodeName: "start",
		Bs:       make(Bindings),
	}

	to := 4
	c := &Control{
		Limit: 10,
		Breakpoints: map[string]Breakpoint{
			"1": func(ctx context.Context, st *State) bool {
				if n, have := st.Bs["n"]; have {
					return to == n.(int)
				}
				return false
			},
		},
	}

	pending := []interface{}{
		Dwimjs(`{"trigger":"do"}`),
	}

	walked, err := spec.Walk(ctx, st, pending, c, nil)
	if err != nil {
		t.Fatal(err)
	}

	if walked.StoppedBecause != BreakpointReached {
		t.Fatalf("bad reason: %v", walked.StoppedBecause)
	}

	toState := walked.To()
	if n, have := toState.Bs["n"]; !have {
		t.Fatal("lost n")
	} else if to != n.(int) {
		t.Fatalf("%v != %v", n.(int), toState)
	}
}

func BenchmarkTurnstile(b *testing.B) {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	spec, err := TurnstileSpec(ctx)
	if err != nil {
		b.Fatal(err)
	}

	c := &Control{
		Limit: 100,
	}

	ss := []string{
		`{"input":"coin"}`,
		`{"input":"push"}`,
		`{"input":"push"}`,
		`{"input":"coin"}`,
		`{"input":"coin"}`,
		`{"input":"push"}`,
		`{"input":"coin"}`,
		`{"input":"push"}`,
		`{"input":"push"}`,
		`{"input":"coin"}`,
	}

	pending := make([]interface{}, 0, len(ss))
	for _, s := range ss {
		pending = append(pending, Dwimjs(s))
	}

	st := &State{
		NodeName: "locked",
		Bs:       make(Bindings),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := spec.Walk(ctx, st, pending, c, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func TestNoMatchGuardLeak(t *testing.T) {
	spec := &Spec{
		Name:          "test",
		PatternSyntax: "json",
		Nodes: map[string]*Node{
			"start": {
				Branches: &Branches{
					Type: "message",
					Branches: []*Branch{
						{
							Pattern: `{"x":"?x"}`,
							Guard: &FuncAction{
								F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
									e := NewExecution(nil)
									return e, nil
								},
							},
							Target: "other",
						},
					},
				},
			},
			"other": {},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := spec.Compile(ctx, nil, true); err != nil {
		t.Fatal(err)
	}

	st := &State{
		NodeName: "start",
		Bs:       make(Bindings),
	}

	c := DefaultControl

	_, err := spec.Step(ctx, st, Dwimjs(`{"x":"a"}`), c, nil)
	if err != nil {
		t.Fatal(err)
	}

	if x, have := st.Bs["?x"]; have {
		log.Fatal(x)
	}
}

func TestTerminalBindingsNode(t *testing.T) {
	spec := &Spec{
		Name:          "test",
		PatternSyntax: "json",
		Nodes: map[string]*Node{
			"start": {
				Branches: &Branches{
					Branches: []*Branch{
						{
							Target: "there",
						},
					},
				},
			},
			"there": {},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := spec.Compile(ctx, nil, true); err != nil {
		t.Fatal(err)
	}

	st := &State{
		NodeName: "start",
		Bs:       make(Bindings),
	}

	c := &Control{
		Limit: 10,
	}

	pending := []interface{}{Dwimjs(`{"trigger":"fire"}`)}

	walked, err := spec.Walk(ctx, st, pending, c, nil)
	if err != nil {
		t.Fatal(err)
	}

	if walked.StoppedBecause != Done {
		t.Fatalf("stopped because %s", walked.StoppedBecause)
	}
}
