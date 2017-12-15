package main

import (
	"time"

	"github.com/Comcast/sheens/crew"
)

type CrewCacheEntry struct {
	Crew    *crew.Crew
	Expires time.Time
}

func (e *CrewCacheEntry) Get() *crew.Crew {
	if time.Now().After(e.Expires) {
		return nil
	}
	return e.Crew
}

type CrewCache struct {
	// Only expires entries when they are fetched.
	TTL     time.Duration
	Entries map[string]*CrewCacheEntry
}

func NewCrewCache(ttl time.Duration, size int) *CrewCache {
	return &CrewCache{
		TTL:     ttl,
		Entries: make(map[string]*CrewCacheEntry, size),
	}
}

func (c *CrewCache) Put(cid string, crew *crew.Crew) {
	c.Entries[cid] = &CrewCacheEntry{
		Crew:    crew,
		Expires: time.Now().Add(c.TTL),
	}
}

func (c *CrewCache) Rem(cid string) {
	delete(c.Entries, cid)
}

func (c *CrewCache) Get(cid string) *crew.Crew {
	e, have := c.Entries[cid]
	if !have {
		return nil
	}
	if c := e.Get(); c != nil {
		return c
	}
	delete(c.Entries, cid)
	return nil
}
