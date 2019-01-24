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
	"context"
	"encoding/json"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/match"
)

var (
	// CaptainMachine is the id of the captain.
	CaptainMachine = "captain"
)

// CrewOp is a crude structure for crew-level operations (such as
// adding a machine).
type CrewOp struct {
	Update map[string]*crew.Machine `json:"update,omitempty"`
	Delete []string                 `json:"delete,omitempty"`
}

// AsCrewOp attempts to interpret the given message (hopefully a map)
// as a CrewOp.
func AsCrewOp(msg interface{}) (*CrewOp, error) {
	js, err := json.Marshal(&msg)
	if err != nil {
		return nil, err
	}
	var op CrewOp
	if err = json.Unmarshal(js, &op); err != nil {
		return nil, err
	}
	if op.Update == nil && op.Delete == nil {
		// Not much of a a CrewOp.
		return nil, nil
	}
	return &op, nil
}

// DoOp executes the given CrewOp.
func (c *Crew) DoOp(ctx context.Context, op *CrewOp) error {
	for mid, m := range op.Update {
		c.Logf("Crew.Do Update %s", mid)
		if err := c.SetMachine(ctx, mid, m.SpecSource, m.State); err != nil {
			return err
		}
	}

	for _, mid := range op.Delete {
		c.Logf("Crew.Do Delete %s", mid)
		if err := c.DeleteMachine(ctx, mid); err != nil {
			return err
		}
	}

	return nil
}

// NewCaptainSpec creates a machine Spec for a "captain" who can
// execute CrewOps.
func (c *Crew) NewCaptainSpec() *core.Spec {
	spec := &core.Spec{
		Nodes: map[string]*core.Node{
			"start": {
				Branches: &core.Branches{
					Type: "message",
					Branches: []*core.Branch{
						{
							Pattern: "?op",
							Target:  "do",
						},
					},
				},
			},
			"do": {
				Action: &core.FuncAction{
					F: func(ctx context.Context, bs match.Bindings, props core.StepProps) (*core.Execution, error) {
						x, have := bs["?op"]
						if !have {
							return core.NewExecution(bs.Extend("error", "no op")), nil
						}
						op, err := AsCrewOp(x)
						if err != nil {
							return core.NewExecution(bs.Extend("error", "bad crew op: "+err.Error())), nil
						}
						if op == nil {
							return core.NewExecution(bs), nil
						}

						err = c.DoOp(ctx, op)
						if err != nil {
							return core.NewExecution(bs.Extend("error", "crew op error: "+err.Error())), nil
						}

						return core.NewExecution(match.NewBindings()), nil
					},
				},
				Branches: &core.Branches{
					Type: "bindings",
					Branches: []*core.Branch{
						{
							Target: "start",
						},
					},
				},
			},
		},
	}

	return spec
}
