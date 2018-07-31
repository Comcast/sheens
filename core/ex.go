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
