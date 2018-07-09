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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/jsccast/yaml"
)

func (s *Service) TCPService(ctx context.Context, port string) error {
	log.Printf("TCPService on %s", port)

	l, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	ctl := make(chan bool, 1)

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func() {
			in := bufio.NewReader(conn)

			if err = s.Listener(ctx, in, conn, ctl); err != nil {
				if err != io.EOF {
					log.Printf("TCPService: %s", err)
				}
			}
			conn.Close()

			select {
			case <-ctl:
				l.Close()
			default:
			}

		}()
	}
}

func (s *Service) Listener(ctx context.Context, in *bufio.Reader, out io.Writer, ctl chan bool) error {
	render := "prettyjson"

	sayMutex := sync.Mutex{}

	say := func(x interface{}) bool {
		sayMutex.Lock()
		defer sayMutex.Unlock()

		var js []byte
		var err error
		switch render {
		case "render json":
			js, err = json.Marshal(&x)
		case "render prettyjson":
			js, err = json.MarshalIndent(&x, "  ", "  ")
		case "render yaml":
			js, err = yaml.Marshal(&x)
		default:
			js, err = json.Marshal(&x)
		}
		if err != nil {
			log.Printf("Service.listener warning on rendering: %s on %#v", err, x)
			js = []byte(fmt.Sprintf("error: %s on %#v", err, x))
		}

		js = append(js, '\n')

		if _, err = out.Write(js); err != nil {
			log.Printf("Service.listener warning on Write: %s", err)
			return false
		}

		return true
	}

	complain := func(err error) bool {
		return say(map[string]interface{}{
			"error": err.Error(),
		})
	}

	okay := func() bool {
		return say("okay")
	}

	echo := false

	for {
		line, err := in.ReadBytes('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		{
			sl := strings.TrimSpace(string(line))

			if echo {
				fmt.Fprintf(out, "%s", sl)
			}

			if strings.HasPrefix(sl, "#") || sl == "" {
				continue
			}

			switch sl {
			case "shutdown":
				log.Printf("TCP client says to shutdown")
				ctl <- true
				return nil
			case "prettyjson":
				render = "prettyjson"
				okay()
				continue
			case "yaml":
				render = "yaml"
				okay()
				continue
			case "json":
				render = "json"
				okay()
				continue
			}

			parts := strings.Split(sl, " ")
			switch parts[0] {
			case "sleep":
				if len(parts) != 2 {
					if !complain(fmt.Errorf("sleep DURATION")) {
						return nil
					}
					continue
				}
				d, err := time.ParseDuration(parts[1])
				if err != nil {
					if !complain(err) {
						return nil
					}
					continue
				}
				time.Sleep(d)
				continue
			}

			var op SOp
			js := []byte(sl)
			if err := json.Unmarshal(js, &op); err != nil {
				if !complain(err) {
					return err
				}
				continue
			}
			if err = op.Do(ctx, s); err != nil {
				if !complain(err) {
					return err
				}
				continue
			}

			if !say(&op) {
				return nil
			}
		}
	}

	return nil
}
