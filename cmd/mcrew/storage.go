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

package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/match"
	. "github.com/Comcast/sheens/util/testutil"

	"github.com/boltdb/bolt"
)

// MachineState is pretty cool type if you ask me. It's the basic building block of what
// a "machine" is. When the idea of a machine is just an inkling in your eye, the next
// step is to create a MachineState
type MachineState struct {
	// Mid is the id for the machine.
	Mid string `json:"id,omitempty"`

	SpecSource *crew.SpecSource `json:"spec,omitempty" yaml:"spec,omitempty"`
	NodeName   string           `json:"node"`
	Bs         match.Bindings   `json:"bs"`

	// Deleted indicated that this machine has been deleted.
	//
	// Yes, this flag is a hack.
	Deleted bool `json:"-" yaml:"-"`
}

// AsMachinesStates is a function, naturally, and it takes in changes
// to a MachineState and record them each as a new state. In return it will
// give you back those states to do with as you please
func AsMachinesStates(changes map[string]*core.State) []*MachineState {
	acc := make([]*MachineState, 0, len(changes))
	for mid, s := range changes {
		ms := &MachineState{
			Mid:      mid,
			NodeName: s.NodeName,
			Bs:       s.Bs,
		}
		acc = append(acc, ms)
	}
	return acc
}


// AsMachines takes in a mss, which I believe might be a set of machines
// also known as a crew. A crew of machines. This function takes in Machine
// states, and builds a crew of out them. You will get a crew back. I think.
func AsMachines(mss []*MachineState) map[string]*crew.Machine {
	acc := make(map[string]*crew.Machine, len(mss))
	for _, ms := range mss {
		m := &crew.Machine{
			Id: ms.Mid,
			State: &core.State{
				NodeName: ms.NodeName,
				Bs:       ms.Bs,
			},
			SpecSource: ms.SpecSource,
		}
		acc[ms.Mid] = m
	}
	return acc
}

// Storage is a type of persistence
type Storage struct {
	Debug    bool
	filename string
	db       *bolt.DB
}

// NewStorage takes in a filename and returns a Storage object
func NewStorage(filename string) (*Storage, error) {
	return &Storage{
		filename: filename,
	}, nil
}

// Open is a function which uses a specific persistence layer,
// bolt, and calls its Open() function on the set Storage objects
// filename
func (s *Storage) Open(ctx context.Context) error {
	opts := &bolt.Options{
		Timeout: time.Second,
	}

	db, err := bolt.Open(s.filename, 0644, opts)
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

// Close is a function which uses a specific persistance layer,
// bolt, and call its Close() function on the set Storage object 
func (s *Storage) Close(ctx context.Context) error {
	if s == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Storage) logf(format string, args ...interface{}) {
	if s == nil {
		return
	}
	if s.Debug {
		log.Printf("BoltDB "+format, args...)
	}
}

// EnsureCrew is a function which uses a specific persistence layer,
// bolt, and does some very specifically scoped calling of the
// CreateBucketIfNotExists function on the Storage object
func (s *Storage) EnsureCrew(ctx context.Context, pid string) error {
	if s == nil {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(pid))
		return err
	})
}

// RemCrew also uses a specific persistence layer known as bolt,
// it has some specificlly scoped calls to DeleteBucket
func (s *Storage) RemCrew(ctx context.Context, pid string) error {
	if s == nil {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(pid))
	})
}

// GetCrew looks like a function that is a record "retriever", it takes in a pid
// which is a string and not an int, yet a string might also be a int "thing". The universe
// is mystical. Hopefully if all works, this function will give you back an array of machine states.
// which might be crew. Yet, it might just be a collection of MachineStates.
func (s *Storage) GetCrew(ctx context.Context, pid string) ([]*MachineState, error) {
	if s == nil {
		return []*MachineState{}, nil
	}
	mss := make([]*MachineState, 0, 32)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pid))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for id, bs := c.First(); id != nil; id, bs = c.Next() {
			var ms MachineState
			if err := json.Unmarshal(bs, &ms); err != nil {
				return err
			}
			ms.Mid = string(id)
			s.logf("GetCrew %s machine %s", pid, JS(ms))
			mss = append(mss, &ms)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logf("GetCrew %s found %d machines", pid, len(mss))

	if len(mss) == 0 {
		return nil, nil
	}

	return mss, nil
}

// Huh... ?
var NotImplemented = errors.New("not implemented")

// WriteState is a critical function. Generally as you change a machines state, you will
// need to write it back to the persistence layer. This function will help you do that.
// Possibly bolt related. Just fyi.
func (s *Storage) WriteState(ctx context.Context, pid string, mss []*MachineState) error {
	if s == nil {
		return nil
	}

	if 0 == len(mss) {
		return nil
	}

	vals := make(map[string][]byte, len(mss))

	for _, ms := range mss {
		id := ms.Mid
		if ms.Deleted {
			vals[id] = nil
		} else {
			// To save some space, remove id.
			ms = &MachineState{
				SpecSource: ms.SpecSource,
				NodeName:   ms.NodeName,
				Bs:         ms.Bs,
			}
			js, err := json.Marshal(&ms)
			if err != nil {
				return err
			}
			vals[id] = js
		}
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(pid))
		if err != nil {
			return err
		}
		for id, bs := range vals {
			var (
				key = []byte(id)
				err error
			)
			if bs == nil {
				err = b.Delete(key)
			} else {
				err = b.Put(key, bs)
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
}
