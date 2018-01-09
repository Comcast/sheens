package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSockets adds Websockets support to the existing HTTP server.
//
// Warning: This is demo code, and it does not scale.  In particular,
// this code turns on a firehose of operations for the entire service.
// This firehose reports all ops (in and out) as well as all routed
// messages to ALL websocket clients.
func (s *Service) WebSockets(ctx context.Context, port string) error {
	s.firehose = make(chan interface{}, 1024)

	var upgrader = websocket.Upgrader{} // use default options

	// We aren't proud.
	conns := sync.Map{}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case x := <-s.firehose:
				conns.Range(func(k, v interface{}) bool {
					c := v.(chan interface{})
					select {
					case c <- x:
					default:
						log.Printf("%v firehose blocked", k)
					}
					return true
				})
			}
		}

	}()

	api := func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade error", err)
			return
		}
		defer c.Close()

		ctl := make(chan bool)
		defer close(ctl)

		firehose := make(chan interface{}, 32)
		defer close(firehose)

		id := c.LocalAddr().String()
		conns.Store(id, firehose)
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
				case x := <-firehose:
					if x == nil {
						break LOOP
					}
					js, err := json.Marshal(&x)
					if err != nil {
						log.Printf("s.firehose Marshal error %v on %#v", err, x)
						continue
					}
					if err = c.WriteMessage(mt, js); err != nil {
						log.Println("s.firehose write:", err)
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

	var uiTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.insertBefore(d, output.firstChild);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("ws://{{.}}/ws/api");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
<style>
body { margin: 2em }
</style>
</head>
<body>
<form>
<button id="open">Open connection</button>
<button id="close">Close connection</button>
<br><input id="input" size="100" type="text" value='{"getSpec":{"source":{"name":"double"}}}'>
<br><button id="send">Send</button>
<hr>
<div id="output"></div>
</body>
</html>
`))

	ui := func(w http.ResponseWriter, r *http.Request) {
		uiTemplate.Execute(w, "localhost"+port)
	}

	http.HandleFunc("/ws/api", api)
	http.HandleFunc("/ws/ui", ui)

	log.Printf("Service.HTTPServer (%s) has Websockets", port)

	return nil
}
