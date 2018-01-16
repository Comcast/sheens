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

func (s *Service) TCPListener(ctx context.Context, port string) error {
	log.Printf("Starting TCP listener on %s", port)

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
					log.Printf("TCPListener: %s", err)
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
	log.Printf("Service listener %p", in)
	defer log.Printf("Service listener closed %p", in)

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

	outHook := func(x interface{}) {
		say(map[string]interface{}{
			"outbound": x,
		})
	}

	inHook := func(x interface{}) {
		say(map[string]interface{}{
			"inbound": x,
		})
	}

	defer func() {
		s.InSubs.RemAll(inHook)
		s.OutSubs.RemAll(outHook)
	}()

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
			case "insub", "inunsub", "outsub", "outunsub":
				if len(parts) != 2 {
					if !complain(fmt.Errorf("*sub CREW_ID")) {
						return nil
					}
					continue
				}
				cid := parts[1]
				switch parts[0] {
				case "insub":
					s.InSubs.Add(cid, inHook)
				case "inunsub":
					s.InSubs.Rem(cid, inHook)
				case "outsub":
					s.OutSubs.Add(cid, outHook)
				case "outunsub":
					s.OutSubs.Rem(cid, outHook)
				}
				continue
			case "echo":
				fmt.Println(strings.Join(parts[1:], " "))
				continue
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
