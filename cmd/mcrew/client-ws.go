package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"

	"github.com/gorilla/websocket"
)

func (s *Service) WebSocketClient(ctx context.Context, urls string) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	u, err := url.Parse(urls)
	if err != nil {
		return err
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	log.Printf("Service.WebSocketClient starting: %s", urls)

	s.wsClientC = make(chan interface{}, 10) // ?

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("WebSocketClient reader closing per ctx")
				return
			default:
			}

			log.Printf("wsclient listening")
			_, message, err := c.ReadMessage()
			if err != nil {
				s.err(err)
				continue
			}
			log.Printf("wsclient heard %s", message)

			var x interface{}
			if err = json.Unmarshal(message, &x); err != nil {
				err = fmt.Errorf("Service WebSocket client in-bound Unmarshal error %s on %s", err, message)
				s.err(err)
				continue
			}

			op := SOp{
				COp: &COp{
					Process: &OpProcess{
						// Render:  true,
						Message: x,
					},
				},
			}

			if err = op.Do(ctx, s); err != nil {
				s.Errors <- err
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("WebSocketClient writer closing per ctx")
			break
		case x := <-s.wsClientC:
			log.Printf("WebSocketClient writer heard %s", JS(x))
			m, is := x.(map[string]interface{})
			if !is {
				err := fmt.Errorf(`%s (%T) isn't a %T`, JS(x), x, m)
				s.Errors <- err
				continue
			}

			// Remove the "to"
			delete(m, "to")

			js, err := json.Marshal(&m)
			if err != nil {
				s.Errors <- err
				continue
			}

			js = withSheenEnvVars(js)

			log.Printf("WebSocketClient writer writing %s", js)

			if err = c.WriteMessage(websocket.TextMessage, js); err != nil {
				s.Errors <- err
				continue
			}
		}
	}

	return nil
}

// withSheenEnvVars replaces all substrings matching sheenEnvVars with
// their corresponding values of environment variables.
func withSheenEnvVars(msg []byte) []byte {
	// ToDo: Make more efficient!
	return sheenEnvVars.ReplaceAllFunc(msg, func(bs []byte) []byte {
		if val := os.Getenv(string(bs[1:])); val != "" {
			return []byte(val)
		}
		return bs
	})
}

// sheenEnvVars matches strings that get expanded based on the
// environment.  See withSheenEnvVars.
var sheenEnvVars = regexp.MustCompile(`\$SHEEN_\w+`)
