package main

import (
	"sync"
)

type Hook func(interface{})

type Subs struct {
	sync.Mutex
	hooks map[string][]Hook
}

func NewSubs() *Subs {
	return &Subs{
		hooks: make(map[string][]Hook, 32),
	}
}

func (s *Subs) Add(cid string, h Hook) {
	s.Lock()
	hooks, have := s.hooks[cid]
	if !have {
		hooks = []Hook{}
	}
	hooks = append(hooks, h)
	s.hooks[cid] = hooks
	s.Unlock()
}

func (s *Subs) Rem(cid string, h Hook) {
	s.Lock()
	s.rem(cid, h)
	s.Unlock()
}

func (s *Subs) rem(cid string, h Hook) {
	if hooks, have := s.hooks[cid]; have {
		acc := make([]Hook, 0, len(hooks))
		for _, f := range hooks {
			if &f != &h {
				acc = append(acc, h)
			}
		}
		s.hooks[cid] = acc
	}
}

func (s *Subs) RemAll(h Hook) {
	s.Lock()
	for cid, _ := range s.hooks {
		s.rem(cid, h)
	}
	s.Unlock()
}

func (s *Subs) Do(cid string, x interface{}) {
	var acc []Hook
	s.Lock()
	if hooks, have := s.hooks[cid]; have {
		acc = make([]Hook, len(hooks))
		for i, h := range hooks {
			acc[i] = h
		}
	}
	s.Unlock()
	for _, h := range acc {
		h(x)
	}
}
