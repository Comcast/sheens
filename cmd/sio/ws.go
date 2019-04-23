/* Copyright 2018-2019 Comcast Cable Communications Management, LLC
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
	"log"
	"net/url"

	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/sio"

	"github.com/gorilla/websocket"
)

type WebSocketCouplings struct {
	URL string
	sio.JSONStore

	in   chan interface{}
	out  chan *sio.Result
	done chan bool
	conn *websocket.Conn
}

func NewWebSocketCouplings(args []string) (*WebSocketCouplings, *flag.FlagSet) {
	c := &WebSocketCouplings{}
	fs := flag.NewFlagSet("ws", flag.ExitOnError)
	fs.StringVar(&c.URL, "-url", "ws://localhost:8080", "Target URL for WebSocket server")
	if args == nil {
		return nil, fs
	}
	fs.Parse(args)
	return c, fs
}

// Start creates the WebSocket session and starts processing it.
func (c *WebSocketCouplings) Start(ctx context.Context) error {

	u, err := url.Parse(c.URL)
	if err != nil {
		return err
	}

	c.in = make(chan interface{})
	c.out = make(chan *sio.Result)
	c.done = make(chan bool)

	log.Println("wsconnect", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	c.conn = conn

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			_, bs, err := conn.ReadMessage()
			if err != nil {
				E(err, "ReadMessage")
				return
			}
			if len(bs) == 0 {
				continue
			}
			log.Println("heard", string(bs))

			var msg interface{}
			if err = json.Unmarshal(bs, &msg); err != nil {
				E(err, "Unmarshal", string(bs))
				continue
			}

			select {
			case <-ctx.Done():
				return
			case c.in <- msg:
				log.Println("processing", string(bs))
			}
		}
	}()

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
							// Remove the "to"
							delete(m, "to")
						}

						js, err := json.Marshal(&msg)
						if err != nil {
							E(err, "Marshal")
							continue
						}

						if err = conn.WriteMessage(websocket.TextMessage, js); err != nil {
							E(err, "WriteMessage")
							return
						}
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
func (c *WebSocketCouplings) IO(ctx context.Context) (chan interface{}, chan *sio.Result, chan bool, error) {
	return c.in, c.out, c.done, nil
}

func (c *WebSocketCouplings) Read(ctx context.Context) (map[string]*crew.Machine, error) {
	return c.JSONStore.Read(ctx)
}

// Stop terminates the WebSocket connection.
func (c *WebSocketCouplings) Stop(ctx context.Context) error {
	log.Printf("Disconnecting")
	c.conn.Close()
	close(c.done)
	// ToDo: Ensure no more writes.
	c.JSONStore.WriteState(ctx)
	return nil
}
