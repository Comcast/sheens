package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	. "github.com/Comcast/sheens/util/testutil"

	"github.com/boltdb/bolt"
)

type MachineState struct {
	// Mid is the id for the machine.
	Mid string `json:"id,omitempty"`

	SpecSource *crew.SpecSource `json:"spec,omitempty" yaml:"spec,omitempty"`
	NodeName   string           `json:"node"`
	Bs         core.Bindings    `json:"bs"`

	// Deleted indicated that this machine has been deleted.
	//
	// Yes, this flag is a hack.
	Deleted bool `json:"-" yaml:"-"`
}

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

type Storage struct {
	Debug    bool
	filename string
	db       *bolt.DB
}

func NewStorage(filename string) (*Storage, error) {
	return &Storage{
		filename: filename,
	}, nil
}

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

func (s *Storage) EnsureCrew(ctx context.Context, pid string) error {
	if s == nil {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(pid))
		return err
	})
}

func (s *Storage) RemCrew(ctx context.Context, pid string) error {
	if s == nil {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(pid))
	})
}

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

var NotImplemented = errors.New("not implemented")

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
