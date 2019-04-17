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
	"log"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/match"
)

// NewHTTPSpec creates a new spec that can process a TimerMsg.
func (c *Crew) NewHTTPSpec() *core.Spec {

	spec := &core.Spec{
		Name: "http",
		Doc:  "A machine that makes HTTP requests.",
		Nodes: map[string]*core.Node{
			"start": {
				Doc: "Wait to hear an HTTP request.",
				Branches: &core.Branches{
					Type: "message",
					Branches: []*core.Branch{
						{
							Pattern: mustParse(`{"httpRequest":"?r","replyTo":"??replyTo"}`),
							Target:  "make",
						},
					},
				},
			},
			"make": {
				Doc: "Try to make the HTTP request.",
				Action: &core.FuncAction{
					F: func(ctx context.Context, bs match.Bindings, props core.StepProps) (*core.Execution, error) {
						msg, have := bs["?r"]
						if !have {
							return core.NewExecution(bs.Extend("error", "no request (r)")), nil
						}
						replyTo := bs["??replyTo"]

						var r HTTPRequest
						// Sorry.
						js, err := json.Marshal(&msg)
						if err != nil {
							return core.NewExecution(bs.Extend("error", "bad HTTP request: "+err.Error())), nil
						}
						if err = json.Unmarshal(js, &r); err != nil {
							return core.NewExecution(bs.Extend("error", "bad HTTP request: "+err.Error())), nil
						}

						go func() {
							err := r.Do(ctx, func(ctx context.Context, resp *HTTPResponse) error {
								resp.From = "http" // me

								// Again: sorry.
								js, err := json.Marshal(&resp)
								if err != nil {
									return fmt.Errorf("Service toHTTP result Marshal error %s", err)
								}
								var msg map[string]interface{}
								if err = json.Unmarshal(js, &msg); err != nil {
									return fmt.Errorf("Service toHTTP result Unmarshal error %s", err)
								}
								if replyTo != nil {
									msg["to"] = replyTo
								}

								if resp.Body != "" {
									var x interface{}
									if err := json.Unmarshal([]byte(resp.Body), &x); err == nil {
										msg["parsedBody"] = x
									}
								}

								// ToDo: timeout?
								c.in <- msg

								return nil
							})
							if err != nil {
								log.Printf("ToDo: error %v", err)
							}
						}()

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
