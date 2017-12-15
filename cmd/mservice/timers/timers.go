// Package timers aspires to provide an efficient facility to
// implement and manage a set of timers.  At any point in time, only
// one time.Timer exists to implement all managed timers.  A Timers
// instance is designed to manage a few hundred timers (and not many
// thousands of timers).
//
// The design is relatively simple.  When a timer is added, it's added
// to a list of pending timers.  That list is order by ascending
// trigger time.  When the first (soonest) timer in that list changes,
// the internal timer is replaced with a new timer that waits until
// the desired time or until a new timer appears at the head of the
// list.  This processing is based on a loop; therefore, this approach
// doesn't scale that much.  Don't expect much quality of service when
// you add a bunch of timers that all fire within a narrow window.
//
// When a timer is triggered, it's work is performed in a new
// goroutine, so it's kinda okay for that work to block.
package timers

import (
	"context"
	"errors"
	"log"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

var (
	NotFound       = errors.New("not found")
	TooMany        = errors.New("too many")
	IdExists       = errors.New("id exists")
	NotRunning     = errors.New("not running")
	AlreadyRunning = errors.New("already running")
)

const (
	notRunning = int64(iota)
	running
)

// Timer represents some work to be done in the future.
type Timer struct {
	// Id is a unique identifier across all timers managed by a
	// given Timers instance.
	Id string `json:"id"`

	// F is the worked to be performed i the future.
	//
	// This timer is passed to this function do make it a little
	// easier -- maybe -- to write more general-purpose work
	// functions.
	F func(context.Context, *Timer) `json:"-"`

	// At is the desired time to execute F.
	At time.Time `json:"time"` // ?

	// Executed, which is the time that F was actually executed,
	// will be written when F is executed.
	Executed time.Time `json:"executed"`
}

// Timers is a managed set of Timer instances.
//
// You need to Run the Timers before calling Add.
type Timers struct {
	Max   int  `json:"max"`
	Debug bool `json:"-"`

	sync.Mutex
	up      chan *Timer
	backlog []*Timer
	running int64
	ready   chan bool
}

// NewTimers makes a new instance with the given maximum number of
// pending timers.
func NewTimers(max int) (*Timers, error) {
	// Let's bring in some magic numbers.
	initial := max / 4
	if initial < 8 {
		initial = 8
	}
	return &Timers{
		Max:     max,
		up:      make(chan *Timer, 32),
		backlog: make([]*Timer, 0, initial),
		ready:   make(chan bool, 1),
	}, nil
}

// Run starts the Timers process in the current goroutine.  This
// method must be running to use the Timers instance.
func (ts *Timers) Run(ctx context.Context) error {
	if ts.IsRunning() {
		return AlreadyRunning
	}

	// timer holds the current timer that we're using.  This timer
	// will be replaced when a new timer becomes the next in line.
	var timer *time.Timer

	atomic.StoreInt64(&ts.running, running)
	ts.ready <- true
LOOP:
	for {
		ts.print("loop")
		select {
		case <-ctx.Done():
			break LOOP
		case t := <-ts.up:
			ts.debugf("timer %s up", t.Id)
			if timer != nil {
				ts.debugf("stopping pending timer")
				if !timer.Stop() {
					// https://github.com/golang/go/issues/14383
					// <-timer.C
				}
				ts.debugf("stopped pending timer")
			}
			d := t.At.Sub(time.Now())
			ts.debugf("timer %s in %s", t.Id, d)
			timer = time.AfterFunc(d, func() {
				ts.debugf("timer %s firing", t.Id)
				ts.Rem(t.Id) // We are optimistic.
				go t.F(ctx, t)
			})
		}
	}

	select {
	case <-ts.ready:
	}
	atomic.StoreInt64(&ts.running, notRunning)

	return nil
}

// IsRunning tries to report whether the Run method is currently
// executing.
func (ts *Timers) IsRunning() bool {
	return atomic.LoadInt64(&ts.running) == running
}

func (ts *Timers) Wait(timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	select {
	case <-timer.C:
		return false
	case <-ts.ready:
		return true
	}
}

// Add adds the given timer to the Timers instance.
func (ts *Timers) Add(t *Timer) error {
	ts.debugf("add %s %s", t.Id, t.At.Sub(time.Now()))

	if !ts.IsRunning() {
		return NotRunning
	}

	ts.Lock()

	var err error
	if len(ts.backlog) == ts.Max {
		err = TooMany
	} else {
		for _, x := range ts.backlog {
			if x.Id == t.Id {
				err = IdExists
				break
			}
		}

		if err == nil {
			n := len(ts.backlog)
			i := sort.Search(len(ts.backlog), func(i int) bool {
				return ts.backlog[i].At.After(t.At)
			})

			ts.debugf("add %s at %d n=%d", t.Id, i, n)
			// Try to avoid leaks ...
			switch i {
			case 0:
				ts.backlog = append(ts.backlog, nil)
				copy(ts.backlog[1:], ts.backlog)
				ts.backlog[0] = t
				ts.reset()
			case n:
				ts.backlog = append(ts.backlog, t)
			default:
				ts.backlog = append(ts.backlog, nil)
				copy(ts.backlog[i+1:], ts.backlog[i:])
				ts.backlog[i] = t
			}
		}
	}

	ts.Unlock()

	ts.print("after add")

	return err
}

// Rem removes the given timer from the Timers instance.
func (ts *Timers) Rem(id string) error {
	ts.debugf("rem %s", id)

	if !ts.IsRunning() {
		return NotRunning
	}

	ts.Lock()
	n := len(ts.backlog)
LOOP:
	for i, t := range ts.backlog {
		if t.Id == id {
			ts.debugf("rem %s at %d", id, i)
			// Try to avoid leaks.
			ts.backlog[i] = nil
			switch i {
			case 0:
				ts.backlog = ts.backlog[1:]
				ts.reset()
				break LOOP
			case n - 1:
				ts.backlog = ts.backlog[0:i]
				break LOOP
			default:
				head := ts.backlog[0:i]
				tail := ts.backlog[i+1:]
				ts.backlog = append(head, tail...)
				break LOOP
			}
		}
	}

	ts.Unlock()

	ts.print("after rem")

	return nil
}

// reset indirectly replaces the existing internal timer with a new
// one that does the right thing.
func (ts *Timers) reset() {
	ts.debugf("reset")
	if 0 < len(ts.backlog) {
		ts.up <- ts.backlog[0]
	}
}

func (ts *Timers) debugf(format string, args ...interface{}) {
	if ts.Debug {
		log.Printf("debug "+format, args...)
	}
}

func (ts *Timers) print(tag string) {
	ts.debugf("backlog %s", tag)

	now := time.Now()
	ts.Lock()
	for i, t := range ts.backlog {
		ts.debugf("  %d %s %s", i, t.Id, t.At.Sub(now))
	}
	ts.Unlock()
}
