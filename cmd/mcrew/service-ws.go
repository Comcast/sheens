/* Copyright 2018 Comcast Cable Communications Management, LLC
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
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

func (s *Service) WebSocketService(ctx context.Context) error {

	s.ops = make(chan interface{}, 1024)

	var upgrader = websocket.Upgrader{} // use default options

	conns := sync.Map{}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case x := <-s.ops:
				conns.Range(func(k, v interface{}) bool {
					c := v.(chan interface{})
					select {
					case c <- x:
					default:
						log.Printf("%v ops blocked", k)
					}
					return true
				})
			}
		}

	}()

	api := func(w http.ResponseWriter, r *http.Request) {
		Logf("Service.WebSocketService connection")

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade error", err)
			return
		}
		defer c.Close()

		ctl := make(chan bool)
		defer close(ctl)

		in := make(chan interface{}, 32)
		defer close(in)

		id := c.LocalAddr().String()
		conns.Store(id, in)
		defer conns.Delete(id)

		go func() {
			mt := websocket.TextMessage

		LOOP:
			for {
				select {
				case <-ctl:
					break LOOP
				case <-ctx.Done():
					break LOOP
				case x := <-in:
					if x == nil {
						break LOOP
					}
					js, err := json.Marshal(&x)
					if err != nil {
						log.Printf("s.firehose Marshal error %v on %#v", err, x)
						continue
					}
					if err = c.WriteMessage(mt, js); err != nil {
						log.Println("s.firehose write error:", err)
					}
				}
			}
		}()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read error", err)
				break
			}

			var op SOp
			if err := json.Unmarshal(message, &op); err != nil {
				msg := fmt.Sprintf("can't parse: %v", err)
				err = c.WriteMessage(mt, []byte(msg))
				if err != nil {
					log.Println("write (err)", err)
					continue
				}
			}
			if err = op.Do(ctx, s); err != nil {
				log.Println("op.Do error", err)
				// Should be conveyed via op.Do.
			}
		}
	}

	http.HandleFunc("/ws/api", api)

	return nil
}
