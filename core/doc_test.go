package core

import (
	"context"
	"fmt"

	. "github.com/Comcast/sheens/util/testutil"
)

// Example demonstrates Walk()ing.
func Example() {

	type note string

	spec := &Spec{
		Name:          "test",
		PatternSyntax: "json",
		Nodes: map[string]*Node{
			"start": {
				Branches: &Branches{
					Type: "message",
					Branches: []*Branch{
						{
							Pattern: `{"request":"?something"}`,
							Target:  "obey",
						},
						{
							Pattern: `{"gimme":"?something"}`,
							Target:  "ignore",
						},
					},
				},
			},
			"obey": {
				Action: &FuncAction{
					F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
						e := NewExecution(make(Bindings)) // Forget current bindings.
						e.Events.AddEmitted(bs["?something"])
						e.Events.AddTrace(note("polite"))
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
			"ignore": {
				Action: &FuncAction{
					F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
						e := NewExecution(make(Bindings)) // Forget current bindings.
						e.Events.AddTrace(note("rude"))
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
		panic(err)
	}

	st := &State{
		NodeName: "start",
		Bs:       make(Bindings),
	}

	ctl := &Control{
		Limit: 10,
	}

	messages := []interface{}{
		Dwimjs(`{"gimme":"queso"}`),
		Dwimjs(`{"request":"chips"}`),
	}

	walked, _ := spec.Walk(ctx, st, messages, ctl, nil)
	for i, stride := range walked.Strides {
		if stride.To != nil {
			fmt.Printf("%02d stride % -32s → % -32s consumed: %s\n",
				i, stride.From, stride.To, JS(stride.Consumed))
		} else {
			fmt.Printf("%02d stride % -32s (no movement)\n",
				i, stride.From)
		}
		for _, m := range stride.Events.Emitted {
			fmt.Printf("   emit   %s\n", JS(m))
		}
		for _, m := range stride.Events.Traces.Messages {
			switch m.(type) {
			case note:
				fmt.Printf("   note   %s\n", JS(m))
			}
		}
	}
	// Output:
	// 00 stride start/{}                         → ignore/{"?something":"queso"}    consumed: {"gimme":"queso"}
	// 01 stride ignore/{"?something":"queso"}    → start/{}                         consumed: null
	//    note   "rude"
	// 02 stride start/{}                         → obey/{"?something":"chips"}      consumed: {"request":"chips"}
	// 03 stride obey/{"?something":"chips"}      → start/{}                         consumed: null
	//    emit   "chips"
	//    note   "polite"
	// 04 stride start/{}                         (no movement)
}
