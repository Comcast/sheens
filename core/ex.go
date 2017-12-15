package core

import (
	"context"
)

// TurnstileSpec makes an example Spec that's useful to have around.
//
// See https://en.wikipedia.org/wiki/Finite-state_machine#Example:_coin-operated_turnstile.
func TurnstileSpec(ctx context.Context) (*Spec, error) {

	makePattern := func(input string) interface{} {
		return map[string]interface{}{
			"input": input,
		}
	}

	spec := &Spec{
		Name: "turnstile",
		Nodes: map[string]*Node{
			"locked": {
				Branches: &Branches{
					Type: "message",
					Branches: []*Branch{
						{
							Pattern: makePattern("coin"),
							Target:  "unlocked",
						},
						{
							Pattern: makePattern("push"),
							Target:  "locked",
						},
					},
				},
			},
			"unlocked": {
				Branches: &Branches{
					Type: "message",
					Branches: []*Branch{
						{
							Pattern: makePattern("coin"),
							Target:  "unlocked",
						},
						{
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
