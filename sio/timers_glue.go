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

	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/match"
)

// GetTimers gets the Timers for the crew.
func (c *Crew) GetTimers(ctx context.Context) (*Timers, error) {

	emitter := func(ctx context.Context, te *TimerEntry) {
		c.Logf("queuing timed message: %s", JS(te.Msg))
		c.in <- func(c *Crew) interface{} {
			timers, err := c.GetTimers(ctx)
			if err != nil {
				// ToDo
				c.Errorf("emitter GetTimers error %s", err)
			} else {
				timers.Cancel(ctx, te.Id)
			}
			return te.Msg
		}
	}

	ms := c.Machines
	tm, have := ms[TimersMachine]
	if !have {
		tm = &crew.Machine{
			Id: TimersMachine,
		}
		ms[TimersMachine] = tm
	}
	if tm.State == nil {
		tm.State = DefaultState(nil)
	}
	if tm.State.Bs == nil {
		tm.State.Bs = match.NewBindings()
	}
	x, have := tm.State.Bs[TimersMachine]
	if !have {
		c.Logf("no state for timers")
		timers := c.timers
		timers.c = c
		timers.Emitter = emitter
		tm.State.Bs[TimersMachine] = timers
		return timers, nil
	}

	switch vv := x.(type) {
	case *Timers:
		return vv, nil
	case interface{}:
		c.Logf("raw state for timers: %s", JS(x))
		js, err := json.Marshal(&x)
		if err != nil {
			return nil, fmt.Errorf("couldn't serialize %v: %v", x, err)
		}
		var timers Timers
		if err = json.Unmarshal(js, &timers); err != nil {
			return nil, fmt.Errorf("couldn't deserialize %s: %v", js, err)
		}
		timers.Emitter = emitter

		// "go vet" said "assignment copies lock value to tm.State.Bs[TimersMachine]: sio.Timers"!
		tm.State.Bs[TimersMachine] = &timers
		return &timers, nil
	default:
		return nil, fmt.Errorf("Bad timers: %T", x)
	}
}
