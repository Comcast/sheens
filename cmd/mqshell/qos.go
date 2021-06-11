package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

type TestMsg struct {
	State struct {
		Reported struct {
			N int
			S string
			T time.Time
		}
	}

	received int
}

func NewTestMsg(sequence, size int) (*TestMsg, error) {
	t := &TestMsg{}
	t.State.Reported.N = sequence

	if err := t.GenS(size); err != nil {
		return nil, err
	}
	t.State.Reported.T = time.Now().UTC()
	return t, nil
}

type QoS struct {
	Latency   time.Duration
	Delta     int
	Duplicate bool
}

func (t *TestMsg) QoS(previous *TestMsg, history map[int]*TestMsg) *QoS {
	n := -1
	if previous != nil {
		n = previous.State.Reported.N
	}
	q := &QoS{
		Latency: time.Now().Sub(t.State.Reported.T),
		Delta:   t.State.Reported.N - n - 1,
	}
	if history != nil {
		if u, have := history[t.State.Reported.N]; have {
			q.Duplicate = 0 < u.received
			u.received++
		} else {
			history[t.State.Reported.N] = t
			t.received = 1
		}
	}

	return q
}

func (t *TestMsg) GenS(n int) error {
	var (
		buf    = make([]byte, n)
		r, err = rand.Read(buf)
	)
	if err != nil {
		return err
	}
	if r != n {
		return fmt.Errorf("bad rand.Read: %d != %d", n, r)
	}

	t.State.Reported.S = hex.EncodeToString(buf)

	return nil
}
