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
	"io/ioutil"
	"sync"

	"github.com/Comcast/sheens/crew"
)

// JSONStore is a primitive facility to store crew state as JSON in a
// file.
//
// Not glamorous or efficient.
type JSONStore struct {
	// StateOutputFilename, if not empty, will be the filename
	// writing state as JSON.
	StateOutputFilename string

	// StateInputFilename optionally gives a filename that
	// contains state to return when Read is called.
	StateInputFilename string

	State map[string]*crew.Machine

	WG sync.WaitGroup
}

func NewJSONStore() *JSONStore {
	return &JSONStore{
		StateOutputFilename: "state.json",
	}
}

// Start does nothing.
func (s *JSONStore) Start(ctx context.Context) error {
	return nil
}

// Stop writes out the state if requested by StateInputFilename.
//
// This function first waits for s.WG if told to.
func (s *JSONStore) Stop(ctx context.Context, wait bool) error {
	if wait {
		s.WG.Wait()
	}
	return s.WriteState(ctx)
}

// Read reads s.StateInputFilename, which should contain a JSON
// representation of the crew's state.
func (s *JSONStore) Read(ctx context.Context) (map[string]*crew.Machine, error) {
	if s.StateInputFilename != "" {
		js, err := ioutil.ReadFile(s.StateInputFilename)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(js, &s.State); err != nil {
			return nil, err
		}
		return s.State, nil

	}
	return make(map[string]*crew.Machine), nil
}

// writeState writes the entire crew as JSON.
func (s *JSONStore) WriteState(ctx context.Context) error {
	if s.State != nil {
		js, err := json.MarshalIndent(&s.State, "", "  ")
		if err != nil {
			return err
		}
		if err = ioutil.WriteFile(s.StateOutputFilename, js, 0644); err != nil {
			return err
		}
	}
	return nil
}

func (s *JSONStore) Update(r *Result) error {
	// We could spend CPU now to determine any net changes, which
	// would allow others to avoid unnecessary writes.  But we
	// don't.
	state := s.State
	for mid, m := range r.Changed {
		if state != nil {
			if m.Deleted {
				delete(state, mid)
			} else {
				n, have := state[mid]
				if !have {
					n = &crew.Machine{}
					state[mid] = n
				}
				if m.State != nil {
					n.State = m.State.Copy()
				}
				if m.SpecSrc != nil {
					n.SpecSource = m.SpecSrc.Copy()
				}
			}
		}
	}
	return nil
}
