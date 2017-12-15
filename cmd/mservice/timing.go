package main

import (
	"log"
	"time"
)

var TimerOutput = false

type Timer struct {
	Tag  string
	Then time.Time
}

func NewTimer(tag string) *Timer {
	return &Timer{
		Tag:  tag,
		Then: time.Now(),
	}
}

func (t *Timer) Stop() time.Duration {
	now := time.Now()
	d := now.Sub(t.Then)
	t.Then = now
	return d
}

func (t *Timer) Sub(d time.Duration) {
	t.Then = t.Then.Add(d)
}

func (t *Timer) StopLog() time.Duration {
	d := t.Stop()
	if TimerOutput {
		log.Printf("timer %s %fμ", t.Tag, d.Seconds()*1000*1000)
	}
	return d
}
