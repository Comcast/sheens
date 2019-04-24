/* Copyright 2019 Comcast Cable Communications Management, LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/sio"
)

// HTTPDCouplings implements an sio.Couplings based on a HTTP service.
// The HTTP service accepts in-bound messages and can forward
// out-bound messages.
//
// The HTTP API supports synchronous processing of a message, which
// returns the messages emitted during the processing of a given
// message.
//
// The HTTP API also supports long-polling to obtain messages emitted
// asychronously.
type HTTPDCouplings struct {
	Port string

	sio.JSONStore

	in   chan interface{}
	out  chan *sio.Result
	done chan bool

	sync.RWMutex
	sigs *Signals
	crew *sio.Crew
}

// NewHTTPDCouplings parses the command-line flags to generate an HTTPDCouplings.
//
// To help with command-line usage reportnig, this function also
// returns the flag.FlagSet used to process the command-line args.
func NewHTTPDCouplings(args []string) (*HTTPDCouplings, *flag.FlagSet) {
	c := &HTTPDCouplings{}
	fs := flag.NewFlagSet("httpd", flag.ExitOnError)
	fs.StringVar(&c.Port, "-port", "localhost:8080", "Port (host:port) for HTTP service")
	if args == nil {
		return nil, fs
	}
	fs.Parse(args)
	c.sigs = NewSignals()
	return c, fs
}

// Start creates the HTTP service and starts processing it.
func (c *HTTPDCouplings) Start(ctx context.Context) error {

	c.in = make(chan interface{})
	c.out = make(chan *sio.Result)
	c.done = make(chan bool)
	hist := NewHistory(1024)

	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "\"pong\"\n")
	})

	puntf := func(w http.ResponseWriter, format string, args ...interface{}) {
		s := fmt.Sprintf(format, args...)
		log.Println(s)

		msg := map[string]interface{}{
			"error": s,
		}
		js, err := json.Marshal(&msg)
		if err != nil {
			// Better than nothing?
			js = []byte(s)
		}
		fmt.Fprintf(w, "%s\n", js)
	}

	mux.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
		var since int64
		if n, err := strconv.ParseInt(r.FormValue("since"), 10, 64); err == nil {
			since = n
		}

		timeout, err := time.ParseDuration(r.FormValue("timeout"))
		if err != nil {
			timeout = 10 * time.Second
		}

		msgs := hist.Get(ctx, since, timeout)

		js, err := json.Marshal(&msgs)
		if err != nil {
			puntf(w, "Marshal error %v on %#v", err, msgs)
			return
		}
		fmt.Fprintf(w, "%s\n", js)
	})

	mux.HandleFunc("/in", func(w http.ResponseWriter, r *http.Request) {
		js, err := ioutil.ReadAll(r.Body)
		if err != nil {
			puntf(w, "ReadAll error %v\n", err)
			return
		}

		var msg interface{}
		if err = json.Unmarshal(js, &msg); err != nil {
			puntf(w, "Unmarshal error %v on %s\n", err, js)
			return
		}

		if r.FormValue("sync") == "true" {
			pr, err := c.crew.ProcessMsg(ctx, msg)
			if err != nil {
				puntf(w, "ProcessMsg error %v\n", err)
				return
			}

			resp := map[string]interface{}{
				"msgs": pr.Emitted,
			}

			if r.FormValue("emit") != "true" {
				pr.Emitted = [][]interface{}{}
			}

			js, err = json.Marshal(&resp)
			if err != nil {
				puntf(w, "Marshal error %v on %#v\n", err, resp)
				return
			}

			fmt.Fprintf(w, "%s\n", js)

			// Forward the result (without messages) to support
			// persistence.
			c.out <- pr
		}

		c.in <- msg
		fmt.Fprintf(w, "{}\n")

		return

	})

	s := &http.Server{
		Addr:           c.Port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Printf("Starting HTTP service on %s", c.Port)
		if err := s.ListenAndServe(); err != nil {
			log.Printf("ListenAndServe error %v", err)
			os.Exit(1)
		}
	}()

	// Listen for emitted messages and accumulate them for clients
	// who want to get these messages asynchronously.
	go func() {
		for {
			select {
			case <-ctx.Done():
				break
			case r := <-c.out:
				for _, msgs := range r.Emitted {
					for _, msg := range msgs {
						m, is := msg.(map[string]interface{})
						if is {
							delete(m, "to")
						}

						hist.Add(msg)
					}
				}
				if err := c.Update(r); err != nil {
					E(err, "Update")
					return
				}
			}
		}
	}()

	return nil
}

// IO just returns the channels that Start() initialized.
func (c *HTTPDCouplings) IO(ctx context.Context) (chan interface{}, chan *sio.Result, chan bool, error) {
	return c.in, c.out, c.done, nil
}

// Read reads the JSONStore to return machines' states.
func (c *HTTPDCouplings) Read(ctx context.Context) (map[string]*crew.Machine, error) {
	return c.JSONStore.Read(ctx)
}

// Stop terminates the HTTP service.
func (c *HTTPDCouplings) Stop(ctx context.Context) error {
	log.Printf("Disconnecting")
	close(c.done)
	// ToDo: Ensure no more writes.
	c.JSONStore.WriteState(ctx)
	return nil
}

// Nothings is a channel of nothing.
//
// A Nothings can be used as a semaphore.
type Nothings chan struct{}

// Signals is sort of sequence of semaphors that can be used to report
// when a new message has arrived.
type Signals struct {
	sync.Mutex
	c Nothings
}

func NewSignals() *Signals {
	return &Signals{
		c: make(Nothings),
	}
}

// Signal tells the Signals that something has happened.
func (s *Signals) Signal() {
	s.Lock()
	close(s.c)
	s.c = make(Nothings)
	s.Unlock()
}

// C returns a channel that is closed upon a Signal().
func (s *Signals) C() Nothings {
	s.Lock()
	c := s.c
	s.Unlock()
	return c
}

// History is a message buffer.
//
// Each messages is assigned a sequence number (currently).
type History struct {
	sync.RWMutex
	sigs   *Signals
	last   int64
	limit  int
	buffer []HistoryMsg
}

func NewHistory(size int) *History {
	return &History{
		limit:  size,
		sigs:   NewSignals(),
		buffer: make([]HistoryMsg, 0, size),
	}
}

// HistoryMsg associates a number with a message.
type HistoryMsg struct {
	N   int64       `json:"n"`
	Msg interface{} `json:"msg"`
}

// Wait returns a channel that's closed when the History receives a
// new message.
func (h *History) Wait() Nothings {
	return h.sigs.C()
}

// Add does what you'd expect.
//
// This method also signals the arrival of a new message to the method
// Get().
func (h *History) Add(msg interface{}) {
	h.Lock()
	if h.limit <= len(h.buffer) {
		// ToDo: Verify no leaks.
		copy(h.buffer, h.buffer[1:])
		h.buffer = h.buffer[0 : h.limit-1]
	}
	h.last++
	hm := HistoryMsg{
		N:   h.last,
		Msg: msg,
	}
	h.buffer = append(h.buffer, hm)
	h.Unlock()
	h.sigs.Signal()
}

// get returns messages after the given sequence number.
func (h *History) get(since int64) []HistoryMsg {
	h.RLock()

	var (
		have        = int64(len(h.buffer))
		startSeqNum = h.last - have
	)

	if since < startSeqNum {
		since = startSeqNum
	}
	offset := since - startSeqNum

	msgs := h.buffer[offset:]

	h.RUnlock()

	return msgs
}

// Get obtains messages from the history.
//
// When no messages are available, this method blocks, with the given
// timeout, until a new message arrives.
func (h *History) Get(ctx context.Context, since int64, timeout time.Duration) []HistoryMsg {
	msgs := h.get(since)

	if len(msgs) == 0 {
		timer := time.NewTimer(timeout)
		select {
		case <-ctx.Done():
		case <-timer.C:
		case <-h.Wait():
			msgs = h.get(since)
		}
	}

	return msgs

}
