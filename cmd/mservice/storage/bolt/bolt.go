package bolt

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/Comcast/sheens/cmd/mservice/storage"

	"github.com/boltdb/bolt"
)

func JS(x interface{}) string {
	js, err := json.Marshal(&x)
	if err != nil {
		panic(err)
	}
	return string(js)
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

func (s *Storage) Open() error {
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

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) logf(format string, args ...interface{}) {
	if s.Debug {
		log.Printf("BoltDB Storage."+format, args...)
	}
}

func (s *Storage) MakeCrew(ctx context.Context, pid string) error {
	s.logf("MakeCrew %s", pid)
	return s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(pid))
		return err
	})
}

func (s *Storage) RemCrew(ctx context.Context, pid string) error {
	s.logf("RemCrew %s", pid)
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(pid))
	})
}

func (s *Storage) GetCrew(ctx context.Context, pid string) ([]*storage.MachineState, error) {
	s.logf("GetCrew %s", pid)
	mss := make([]*storage.MachineState, 0, 32)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pid))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for id, bs := c.First(); id != nil; id, bs = c.Next() {
			var ms storage.MachineState
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

func (s *Storage) WriteState(ctx context.Context, pid string, mss []*storage.MachineState) error {
	s.logf("WriteState %s %s", pid, JS(mss))

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
			ms = &storage.MachineState{
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
