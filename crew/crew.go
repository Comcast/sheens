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

package crew

import (
	"sync"
)

type Crew struct {
	sync.RWMutex

	Id       string              `json:"id"`
	Machines map[string]*Machine `json:"machines"`
}

// Copy gets a read lock and returns a copy of the crew.
func (c *Crew) Copy() *Crew {
	c.RLock()
	ms := make(map[string]*Machine, len(c.Machines))
	for id, m := range c.Machines {
		ms[id] = m.Copy()
	}
	acc := &Crew{
		Id:       c.Id,
		Machines: ms,
	}
	c.RUnlock()
	return acc
}
