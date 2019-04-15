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
	"fmt"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/match"
)

// TimerMsg is a command that the timers machine can execute.
type TimerMsg struct {

	// Add the given timer.
	Add struct {
		Id  string      `json:"id"`
		Msg interface{} `json:"msg"`
		In  string      `json:"in"`
		To  string      `json:"to"` // ToDo: Support array
	} `json:"makeTimer"`

	// Cancel the given timer.
	Cancel struct {
		Id string
	} `json:"cancelTimer"`
}

// NewTimersSpec creates a new spec that can process a TimerMsg.
func (c *Crew) NewTimersSpec() *core.Spec {

	onlyTimers := func(bs match.Bindings) match.Bindings {
		acc := match.NewBindings()
		acc["timers"] = bs["timers"]
		return acc
	}

	spec := &core.Spec{
		Name: "timers",
		Doc:  "A machine that makes in-memory timers that send messages.",
		Nodes: map[string]*core.Node{
			"start": {
				Doc: "Wait to hear a request to create or delete a timer.",
				Branches: &core.Branches{
					Type: "message",
					Branches: []*core.Branch{
						{
							Pattern: mustParse(`{"makeTimer":{"in":"?in", "msg":"?msg", "id":"?id"}}`),
							Target:  "make",
						},
						{
							Pattern: mustParse(`{"cancelTimer":"?id"}`),
							Target:  "cancel",
						},
					},
				},
			},
			"make": {
				Doc: "Try to make the timer.",
				Action: &core.FuncAction{
					F: func(ctx context.Context, bs match.Bindings, props core.StepProps) (*core.Execution, error) {
						x, have := bs["?in"]
						if !have {
							return core.NewExecution(bs.Extend("error", "no in")), nil
						}
						in, is := x.(string)
						if !is {
							return core.NewExecution(bs.Extend("error", fmt.Sprintf("non-string in: %T %#v", x, x))), nil
						}

						d, err := time.ParseDuration(in)
						if err != nil {
							msg := fmt.Sprintf("bad in '%s': %v", in, err)
							return core.NewExecution(bs.Extend("error", msg)), nil
						}

						x, have = bs["?id"]
						if !have {
							return core.NewExecution(bs.Extend("error", "no id")), nil
						}
						id, is := x.(string)
						if !is {
							return core.NewExecution(bs.Extend("error", fmt.Sprintf("non-string id: %T %#v", x, x))), nil
						}

						msg, have := bs["?msg"]
						if !have {
							return core.NewExecution(bs.Extend("error", "no message")), nil
						}

						if err = c.timers.Add(ctx, id, msg, d); err != nil {
							return core.NewExecution(bs.Extend("error", err.Error())), nil
						}

						c.timers.changed()

						return core.NewExecution(onlyTimers(bs)), nil
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
			"cancel": {
				Doc: "Try to delete the timer.",
				Action: &core.FuncAction{
					F: func(ctx context.Context, bs match.Bindings, props core.StepProps) (*core.Execution, error) {
						x, have := bs["?id"]
						if !have {
							return core.NewExecution(bs.Extend("error", "no id")), nil
						}
						id, is := x.(string)
						if !is {
							return core.NewExecution(bs.Extend("error", fmt.Sprintf("non-string id: %T %#v", x, x))), nil
						}

						if err := c.timers.Cancel(ctx, id); err != nil {
							return core.NewExecution(bs.Extend("error", err.Error())), nil
						}

						c.timers.changed()

						return core.NewExecution(onlyTimers(bs)), nil
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

func mustParse(x interface{}) interface{} {
	switch vv := x.(type) {
	case []byte:
		var y interface{}
		err := json.Unmarshal(vv, &y)
		if err != nil {
			panic(err)
		}
		return y
	case string:
		return mustParse([]byte(vv))
	default:
		return x
	}
}
