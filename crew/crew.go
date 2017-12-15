package crew

import (
	"sync"
)

type Crew struct {
	sync.RWMutex

	Id       string              `json:"id"`
	Machines map[string]*Machine `json:"machines"`
}
