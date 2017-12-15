package timers

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestTimersBasic(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ts, err := NewTimers(10)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		ts.Run(ctx)
	}()

	if !ts.Wait(time.Second) {
		t.Fatal("timers didn't start running")
	}

	firings := make(chan string, 1024)
	f := func(_ context.Context, t *Timer) {
		t.Executed = time.Now().UTC()
		log.Printf("firing %s (late: %s)", t.Id, t.Executed.Sub(t.At))
		firings <- t.Id
	}

	heard := make([]string, 0, 1024)
	stop := make(chan bool)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-stop:
				return
			case fired := <-firings:
				heard = append(heard, fired)
			}
		}
	}()

	ft := func(id string, d time.Duration) {
		err := ts.Add(&Timer{
			Id: id,
			At: time.Now().Add(d),
			F:  f,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	ft("3", 1000*time.Millisecond)
	ft("2", 500*time.Millisecond)
	ft("1", 100*time.Millisecond)
	ts.Rem("2")
	ft("5", 1500*time.Millisecond)
	ft("4", 1200*time.Millisecond)
	ts.Rem("5")
	ft("6", 2500*time.Millisecond)

	want := []string{"1", "3", "4", "6"}

	time.Sleep(5 * time.Second)

	stop <- true

	for i, s := range heard {
		expect := want[i]
		log.Printf("heard %d '%s'", i, s)
		if expect != s {
			log.Fatalf("expected '%s' but got '%s' at %d", expect, s, i)
		}
	}

}

func TestTimersLag001(t *testing.T) {
	testTimersLag(t, time.Millisecond, 100)
}

func TestTimersLag010(t *testing.T) {
	testTimersLag(t, 10*time.Millisecond, 100)
}

func TestTimersLag020(t *testing.T) {
	testTimersLag(t, 20*time.Millisecond, 100)
}

func TestTimersLag100(t *testing.T) {
	testTimersLag(t, 100*time.Millisecond, 100)
}

func TestTimersLag200(t *testing.T) {
	testTimersLag(t, 200*time.Millisecond, 100)
}

func testTimersLag(t *testing.T, dMax time.Duration, n int) {
	timeout := 10 * dMax * time.Duration(n)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ts, err := NewTimers(n)
	if err != nil {
		t.Fatal(err)
	}
	// ts.Debug = true

	go func() {
		ts.Run(ctx)
	}()

	if !ts.Wait(time.Second) {
		t.Fatal("timers didn't start running")
	}

	var totalLag time.Duration
	var totalFired int

	wg := sync.WaitGroup{}

	wanted := make(map[string]bool, n)
	wantLock := sync.Mutex{}
	want := func(id string) {
		wantLock.Lock()
		wanted[id] = true
		wantLock.Unlock()
	}
	got := func(id string) {
		wantLock.Lock()
		delete(wanted, id)
		wantLock.Unlock()
	}

	f := func(_ context.Context, t *Timer) {
		// log.Printf("firing %s", t.Id)
		got(t.Id)
		totalLag += time.Now().Sub(t.At)
		totalFired++
		wg.Done()
	}

	for i := 0; i < n; i++ {
		time.Sleep(time.Millisecond)

		wg.Add(1)
		d := time.Duration(rand.Intn(int(dMax/time.Millisecond))) * time.Millisecond
		// log.Printf("scheduling %d at %s", i, d)
		id := strconv.Itoa(i)
		err := ts.Add(&Timer{
			Id: id,
			At: time.Now().Add(d),
			F:  f,
		})
		want(id)
		if err != nil {
			t.Fatal(err)
		}
	}

	waited := make(chan bool)
	go func() {
		wg.Wait()
		close(waited)
	}()

	select {
	case <-time.NewTimer(timeout).C:
		for id := range wanted {
			log.Printf("waiting on %s", id)
		}
		t.Fatal("timeout")
	case <-waited:
	}

	if 0 < totalFired {
		log.Printf("dMax: %v fired: %d mean lag: %v", dMax, totalFired, totalLag/time.Duration(totalFired))
	}
}
