package core

import (
	"context"
)

func TurnstileSpec(ctx context.Context) (*Spec, error) {
	// https://en.wikipedia.org/wiki/Finite-state_machine#Example:_coin-operated_turnstile

	makePattern := func(input string) interface{} {
		return map[string]interface{}{
			"input": input,
		}
	}

	spec := &Spec{
		Name: "turnstile",
		Nodes: map[string]*Node{
			"locked": &Node{
				Branches: &Branches{
					Type: "message",
					Branches: []*Branch{
						&Branch{
							Pattern: makePattern("coin"),
							Target:  "unlocked",
						},
						&Branch{
							Pattern: makePattern("push"),
							Target:  "locked",
						},
					},
				},
			},
			"unlocked": &Node{
				Branches: &Branches{
					Type: "message",
					Branches: []*Branch{
						&Branch{
							Pattern: makePattern("coin"),
							Target:  "unlocked",
						},
						&Branch{
							Pattern: makePattern("push"),
							Target:  "locked",
						},
					},
				},
			},
		},
	}

	if err := spec.Compile(ctx, nil, true); err != nil {
		return nil, err
	}

	return spec, nil
}
