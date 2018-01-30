package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime/pprof"

	"github.com/Comcast/sheens/tools"
	. "github.com/Comcast/sheens/util/testutil"
)

func main() {

	var (
		dbFile   = flag.String("d", "home.db", "storage filename")
		specsDir = flag.String("s", "specs", "specs directory")
		libDir   = flag.String("l", "libs", "libraries directory")
		bootFile = flag.String("b", "", "file to read for initial ops")

		httpPort  = flag.String("h", "", "HTTP port for our service")
		wsService = flag.Bool("w", true, "WebSockets service")
		httpDir   = flag.String("f", "", "directory to serve via HTTP")
		tcpPort   = flag.String("t", ":9000", "port for out TCP listener")

		wsClient = flag.String("c", "", "web socket service for client") // ws://localhost:8123/api/websocket

		listenOnStdin = flag.Bool("I", false, "listen for ops on stdin")
		emitToStdout  = flag.Bool("O", false, "emit messages to stdout")
	)

	flag.BoolVar(&Verbose, "v", false, "log lots of wonderful things")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	s, err := NewService(ctx, *specsDir, *dbFile, *libDir)
	if err != nil {
		panic(err)
	}
	s.Tracing = true
	defer s.store.Close(ctx) // ToDo: Check error.

	s.Emitted = make(chan interface{}, 8)
	s.Processing = make(chan interface{}, 8)
	s.Errors = make(chan interface{}, 8)

	if Verbose {
		monitor(ctx, s.Processing, "processing", false)
		monitor(ctx, s.Emitted, "emitted", *emitToStdout)
	}
	monitor(ctx, s.Errors, "errors", false)

	// We need to boot before starting the WebSocketClient, which
	// will send us a message we need to be ready to process.
	if *bootFile != "" {
		if err := s.Boot(ctx, *bootFile); err != nil {
			panic(err)
		}
	}

	if *wsClient != "" {
		go func() {
			if err := s.WebSocketClient(ctx, *wsClient); err != nil {
				panic(err)
			}
		}()
	}

	if *listenOnStdin {
		go func() {
			if err = s.Listener(ctx, bufio.NewReader(os.Stdin), os.Stdout, nil); err != nil {
				log.Printf("Service.Listener os.Stdin os.Stdout error %s", err)
			}
			Logf("stdin listener done")
			cancel()
		}()
	}

	if *tcpPort != "" {
		go func() {
			if err := s.TCPService(ctx, *tcpPort); err != nil {
				panic(fmt.Errorf("Service.Listener TCP error %s", err))
			}
		}()
	}

	if *httpPort != "" {

		go func() {
			if *wsService {
				log.Printf("WebSockets service starting")
				if err := s.WebSocketService(ctx); err != nil {
					panic(err)
				}
			}

			if *httpDir != "" {
				log.Printf("HTTP serving files in %s", *httpDir)
				fs := http.FileServer(http.Dir(*httpDir))
				http.Handle("/static/", http.StripPrefix("/static", fs))
			}

			p := regexp.MustCompile("/specs/([-a-zA-Z0-9_]+)\\.html")

			http.HandleFunc("/specs/", func(w http.ResponseWriter, r *http.Request) {
				ss := p.FindStringSubmatch(r.RequestURI)
				if ss == nil {
					fmt.Fprintf(w, "No spec name in %s", r.RequestURI)
					fmt.Fprintf(w, "try /specs/double.html")
					return
				}
				err := tools.ReadAndRenderSpecPage("specs/"+ss[1]+".yaml", nil, w, true)
				if err != nil {
					fmt.Fprintf(w, "ReadAndRenderSpecPage error: %s", err)
				}
			})

			log.Printf("HTTP service on %s", *httpPort)
			if err = s.HTTPServer(ctx, *httpPort); err != nil {
				panic(err)
			}
		}()
	}

	<-ctx.Done()
}

func monitor(ctx context.Context, c chan interface{}, tag string, toStdout bool) {
	go func() {
		log.Printf("monitoring %s", tag)
	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			case x := <-c:
				js := JS(x)
				log.Printf("%s %s", tag, js)
				if toStdout {
					fmt.Println(js)
				}
			}
		}
		log.Printf("halting monitoring of %s", tag)
	}()
}

func (s *Service) HTTPServer(ctx context.Context, port string) error {
	complain := func(w http.ResponseWriter, x interface{}, status int) {
		w.WriteHeader(status)
		fmt.Fprintf(w, `{"error":"%s"}`+"\n", x)
	}

	http.Handle("/goroutines", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pprof.Lookup("goroutine").WriteTo(w, 1)
	}))

	http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		js, err := ioutil.ReadAll(r.Body)
		if err != nil {
			complain(w, err, http.StatusBadRequest)
			return
		}
		if err := r.Body.Close(); err != nil {
			log.Printf("Service.HTTPServer warning on Body.Close(): %v", err)
		}

		var op SOp
		if err := json.Unmarshal(js, &op); err != nil {
			complain(w, err, http.StatusBadRequest)
			return
		}
		if err = op.Do(ctx, s); err != nil {
			complain(w, err, http.StatusInternalServerError)
			return
		}
		js, err = json.Marshal(&op)
		if err != nil {
			complain(w, err, http.StatusInternalServerError)
		}
		if _, err = w.Write(js); err != nil {
			log.Printf("Service.HTTPServer warning on Write(): %v", err)
		}
	}))

	return http.ListenAndServe(port, nil)

}

func (s *Service) Boot(ctx context.Context, filename string) error {
	in, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer in.Close()

	r := bufio.NewReader(in)
	for {
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("#")) || bytes.HasPrefix(line, []byte("//")) {
			continue
		}
		var op SOp
		if err = json.Unmarshal(line, &op); err != nil {
			return err
		}
		if err := op.Do(ctx, s); err != nil {
			return err
		}
	}

	return nil
}
