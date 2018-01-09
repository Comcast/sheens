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
