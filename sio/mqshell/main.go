/* Copyright 2021 Comcast Cable Communications Management, LLC
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

// Package main is a (command-line) MQTT shell.
//
// See README.md for documentation.
package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {

	var (
		// Try to follow mosquito_sub command line args?

		broker = flag.String("b", "tcp://localhost", "MQTT broker (like 'tcps://hostname')")
		port   = flag.Int("p", 443, "MQTT broker port")

		keepAlive   = flag.Int("k", 60, "MQTT Keep-alive in seconds")
		willTopic   = flag.String("will-topic", "", "MQTT will topic (optional)")
		willPayload = flag.String("will-payload", "", "MQTT will message (optional)")
		willQoS     = flag.Int("will-qos", 0, "MQTT will QoS (optional)")
		willRetain  = flag.Bool("will-retain", false, "MQTT will retention (optional)")
		reconnect   = flag.Bool("reconnect", false, "Automatically attempt to reconnect to broker")
		clean       = flag.Bool("c", true, "MQTT clean session ")
		quiesce     = flag.Int("quiesce", 100, "MQTT disconnection quiescence (in milliseconds)")
		alpn        = flag.String("alpn", "x-amzn-mqtt-ca", "ALPB next protocol (optional, maybe 'x-amzn-mqtt-ca')")

		clientId     = flag.String("i", "", "MQTT client id (optional)")
		userName     = flag.String("u", "", "MQTT username (optional)")
		password     = flag.String("P", "", "MQTT password (optional)")
		certFilename = flag.String("cert", "cred/aws-mtls-client-cert", "MQTT cert filename (optional)")
		keyFilename  = flag.String("key", "cred/aws-mtls-private-key", "MQTT key filename (optional)")
		insecure     = flag.Bool("insecure", false, "Skip MQTT broker cert checking")
		caFilename   = flag.String("cafile", "", "MQTT CA cert filename (optional)")

		tokenKey       = flag.String("token-key-name", "CustAuth", "AWS custom authorizer token key")
		token          = flag.String("token", "", "AWS custom authorizer token")
		tokenSig       = flag.String("token-sig", "", "AWS custom authorizer token signature")
		authorizerName = flag.String("authorizer-name", "", "AWS custom authorizer name")

		connectRetryInterval = flag.Duration("connect-retry-interval", time.Second, "Connection retry interval")
		connectRetry         = flag.Bool("connect-retry", true, "Connection retry")
		connectTimeout       = flag.Duration("connect-timeout", time.Second, "Connection timeout")

		shellExpand = flag.Bool("sh", true, "Enable shell expansion (<<...>>)")
	)

	flag.Parse()

	mqtt.ERROR = log.New(os.Stderr, "mqtt.error", 0)

	opts := mqtt.NewClientOptions()

	*broker = fmt.Sprintf("%s:%d", *broker, *port)
	opts.AddBroker(*broker)
	opts.SetClientID(*clientId)
	opts.SetKeepAlive(time.Second * time.Duration(*keepAlive))
	opts.SetPingTimeout(10 * time.Second)

	opts.ConnectRetry = *connectRetry
	opts.ConnectRetryInterval = *connectRetryInterval
	opts.ConnectTimeout = *connectTimeout

	opts.Username = *userName
	opts.Password = *password
	opts.AutoReconnect = *reconnect
	opts.CleanSession = *clean

	if *token != "" {
		var (
			bs     = make([]byte, 16)
			_, err = rand.Read(bs)
			key    = hex.EncodeToString(bs)
		)
		if err != nil {
			panic(err)
		}

		opts.HTTPHeaders = http.Header{
			"x-amz-customauthorizer-name":      []string{*authorizerName},
			"x-amz-customauthorizer-signature": []string{*tokenSig},
			*tokenKey:                          []string{*token},
			"sec-WebSocket-Key":                []string{key},
			"sec-websocket-protocol":           []string{"mqtt"},
			"sec-WebSocket-Version":            []string{"13"},
		}

	}
	if *willTopic != "" {
		if *willPayload == "" {
			log.Fatal("will topic without payload")
		}
		log.Printf("configuring will")
		opts.WillEnabled = true
		opts.WillTopic = *willTopic
		opts.WillPayload = []byte(*willPayload)
		opts.WillRetained = *willRetain
		opts.WillQos = byte(*willQoS)
	}

	var rootCAs *x509.CertPool
	if rootCAs, _ = x509.SystemCertPool(); rootCAs == nil {
		rootCAs = x509.NewCertPool()
		log.Printf("Including system CA certs")
	}
	if *caFilename != "" {
		certs, err := ioutil.ReadFile(*caFilename)
		if err != nil {
			log.Fatalf("couldn't read '%s': %s", *caFilename, err)
		}

		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			log.Println("No certs appended, using system certs only")
		}
	}

	var certs []tls.Certificate
	if *keyFilename != "" {
		cert, err := tls.LoadX509KeyPair(*certFilename, *keyFilename)
		if err != nil {
			log.Fatal(err)
		}
		certs = []tls.Certificate{cert}
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: *insecure,
	}

	if *alpn != "" {
		// https://docs.aws.amazon.com/iot/latest/developerguide/protocols.html
		tlsConf.NextProtos = []string{
			*alpn,
		}
	}
	if rootCAs != nil {
		tlsConf.RootCAs = rootCAs
	}

	if certs != nil {
		tlsConf.Certificates = certs
	}

	opts.SetTLSConfig(tlsConf)

	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("MQTT connection lost")
	}

	opts.DefaultPublishHandler = func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("> %s %s\n", msg.Topic(), msg.Payload())
	}

	var (
		c      = mqtt.NewClient(opts)
		in     = bufio.NewReader(os.Stdin)
		qos    byte
		retain bool

		space = regexp.MustCompile(" +")
	)

	if t := c.Connect(); t.Wait() && t.Error() != nil {
		log.Fatal(t.Error())
	}

	log.Printf("Connected")

	// Set up our embedded Javascript interpreter!  Yes, this
	// capability is pretty weird.  Why not just use Nodejs?
	// That's a good question, and the answer is beyond the scope
	// of this comment.  That's all I'm going to say about that
	// topic.

	var (
		// js is our Javascript interpreter.
		js = goja.New()

		// jsMutex will allow us to serialize access to the
		// Javascript interpreter.
		jsMutex = sync.Mutex{}

		// exec executes the given Javascript (after getting
		// the lock).
		exec = func(src string) {
			log.Printf("exec\n%s\n", src)
			jsMutex.Lock()
			v, err := js.RunString(src)
			if err != nil {
				fmt.Printf("error: %s\n", err)
			}
			s := JS(v)
			jsMutex.Unlock()
			fmt.Printf("%s\n", s)
		}
	)

	{

		// Define some globals in the Javascript environment.

		// publish to an MQTT topic.
		js.Set("publish", func(topic string, qos int, retain bool, msg goja.Value) {
			log.Printf("publishing to %s", topic)
			s := msg.String()
			t := c.Publish(topic, byte(qos), retain, []byte(s))
			if t.Wait() && t.Error() != nil {
				fmt.Printf("publish error: %s", t.Error())
			}
		})

		defaultHandler := func(c mqtt.Client, m mqtt.Message) {
			fmt.Printf("heard %s %s\n", m.Topic(), m.Payload())
		}

		// subscribe to an MQTT topic pattern.
		js.Set("subscribe", func(topic string, qos int, h goja.Value) {
			log.Printf("subscribing to %s", topic)
			var f goja.Callable
			if h != nil {
				var ok bool
				if f, ok = goja.AssertFunction(h); !ok {
					fmt.Printf("error: not callable: %T\n", h)
				}
			}
			t := c.Subscribe(topic, byte(qos), func(_ mqtt.Client, m mqtt.Message) {
				if h == nil {
					defaultHandler(c, m)
					return
				}
				f(nil, js.ToValue(m.Topic()), js.ToValue(string(m.Payload())))
			})
			if t.Wait() && t.Error() != nil {
				fmt.Printf("subscribe error: %s", t.Error())
			}
		})

		// unsubscribe from an MQTT topic pattern.
		js.Set("unsubscribe", func(topic string, qos int, msg interface{}) {
			log.Printf("unsubscribing from %s", topic)
			t := c.Unsubscribe(topic)
			if t.Wait() && t.Error() != nil {
				fmt.Printf("unsubscribe error: %s", t.Error())
			}
		})

		// A bad version of Javascript's setTimeout().
		//
		// ToDo: Support cancelation.
		js.Set("setTimeout", func(ms int, f goja.Value) {
			if callable, ok := goja.AssertFunction(f); !ok {
				log.Printf("error: setTimeout: %T not callable", f)
			} else {
				go func() {
					time.Sleep(time.Duration(ms) * time.Millisecond)
					callable(nil)
				}()
			}
		})

		// A bad version of Javascript's setInterval().
		//
		// ToDo: Support cancelation.
		js.Set("setInterval", func(ms int, f goja.Value) {
			if callable, ok := goja.AssertFunction(f); !ok {
				log.Printf("error: setTimeout: %T not callable", f)
			} else {
				// Can't cancel.  Good luck!
				go func() {
					for {
						time.Sleep(time.Duration(ms) * time.Millisecond)
						callable(nil)
					}
				}()
			}
		})

		// print()
		js.Set("print", func(args ...interface{}) {
			var acc string
			for i, x := range args {
				if 0 < i {
					acc += " "
				}
				acc += fmt.Sprintf("%s", JS(x))
			}
			fmt.Printf("%s\n", acc)
		})

	}

LOOP: // REPL
	for {
		line, err := in.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		s := string(line)
		if *shellExpand {
			if s, err = ShellExpand(string(line)); err != nil {
				fmt.Printf("shell expansion error: %s", err)
				continue
			}
		}
		parts := space.Split(strings.TrimSpace(s), 3)
		switch parts[0] {

		case "testsub": // topic, number
			var (
				topic    = "test"
				count    = 10
				history  = make(map[int]*TestMsg, count)
				previous *TestMsg
			)

			if 1 < len(parts) {
				topic = parts[1]
			}

			if 2 < len(parts) {
				if count, err = strconv.Atoi(parts[2]); err != nil {
					fmt.Printf("error: bad integer %s", parts[2])
					continue
				}
			}

			h := func(_ mqtt.Client, m mqtt.Message) {
				js := m.Payload()
				var msg TestMsg
				if err = json.Unmarshal(js, &msg); err != nil {
					log.Printf("testsub message error: %s on %s", err, js)
					return
				}
				q := msg.QoS(previous, history)
				log.Printf("latency: %f ms, order delta: %d", float64(q.Latency)/1000/1000, q.Delta)
				previous = &msg
				if len(history) == count {
					log.Printf("testsub terminating")
					go func() {
						if t := c.Unsubscribe(topic); t.Wait() && t.Error() != nil {
							fmt.Printf("unsubscribe error: %s", t.Error())
						}
					}()
				}
			}

			if t := c.Subscribe(topic, qos, h); t.Wait() && t.Error() != nil {
				fmt.Printf("subscribe error: %s", t.Error())
			}

		case "testpub": // topic, number, interval
			var (
				topic    = "test"
				count    = 10
				interval = time.Second
			)

			if 1 < len(parts) {
				topic = parts[1]
			}

			if 2 < len(parts) {
				if count, err = strconv.Atoi(parts[2]); err != nil {
					fmt.Printf("error: bad integer %s", parts[2])
					continue
				}
			}

			if 3 < len(parts) {
				if interval, err = time.ParseDuration(parts[3]); err != nil {
					fmt.Printf("error: bad duration %s", parts[3])
					continue
				}
			}

			for i := 0; i < count; i++ {
				msg, err := NewTestMsg(i, 64)
				js, err := json.Marshal(&msg)
				if err != nil {
					fmt.Printf("serialization error %s", err)
					break
				}
				log.Printf("publishing test message %d to %s", i, topic)
				if t := c.Publish(topic, qos, retain, js); t.Wait() && t.Error() != nil {
					fmt.Printf("publish error: %s", t.Error())
				}
				time.Sleep(interval)
			}

		case "jsfile": // Read and execute a Javascript file.
			if len(parts) != 2 {
				fmt.Printf("error: jsfile FILENAME\n")
				continue
			}
			bs, err := ioutil.ReadFile(parts[1])
			if err != nil {
				fmt.Printf("error: %s\n", err)
				continue
			}
			exec(string(bs))

		case "js": // Execute some Javascript.
			if len(parts) < 2 {
				fmt.Printf("error: js CODE\n")
				continue
			}
			src := strings.Join(parts[1:], " ")

			exec(src)

		case "echo": // Print the input line.
			fmt.Printf("%s\n", space.Split(strings.TrimSpace(s), 2)[1])

		case "sleep": // Sleep for the given duration.
			if len(parts) != 2 {
				fmt.Printf("error: sleep DURATION\n")
				continue
			}
			d, err := time.ParseDuration(parts[1])
			if err != nil {
				fmt.Printf("error: sleep DURATION: %s\n", err)
				continue
			}
			time.Sleep(d)

		case "qos":
			// Set the MQTT QoS for all subsequent
			// subscribes/publishes.
			if len(parts) != 2 {
				fmt.Printf("error: qos [0-9]\n")
				continue
			}
			n, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Printf("error: qos [0-9]: %s\n", err)
				continue
			}
			qos = byte(n)

		case "retain":
			// Set message retention for all subsequent
			// MQTT publishes.
			if len(parts) != 2 {
				fmt.Printf("error: retain (true|false)")
				continue
			}
			switch parts[1] {
			case "true":
				retain = true
			case "false":
				retain = false
			default:
				fmt.Printf("error: retain (true|false)")
				continue
			}

		case "sub", "subscribe": // MQTT subscribe
			if len(parts) != 2 {
				fmt.Printf("error: sub TOPIC")
				continue
			}
			if t := c.Subscribe(parts[1], qos, nil); t.Wait() && t.Error() != nil {
				fmt.Printf("subscribe error: %s", t.Error())
			}

		case "unsub", "unsubscribe": // MQTT unsubscribe
			if len(parts) != 2 {
				fmt.Printf("error: unsub TOPIC")
				continue
			}
			if t := c.Unsubscribe(parts[1]); t.Wait() && t.Error() != nil {
				fmt.Printf("unsubscribe error: %s", t.Error())
			}

		case "pub", "publish": // MQTT publish
			if len(parts) < 3 {
				fmt.Printf("error: pub TOPIC MSG")
				continue
			}
			if t := c.Publish(parts[1], qos, retain, parts[2]); t.Wait() && t.Error() != nil {
				fmt.Printf("subscribe error: %s", t.Error())
			}

		case "quit":
			break LOOP
		}

	}

	log.Printf("Disconnecting")

	c.Disconnect(uint(*quiesce))
}

// shell is a regexp for the notation for finding shell expansions.
//
// The first group is eventually passed to ShellExpand() if the flag
// 'shellExpand' is true.
var shell = regexp.MustCompile(`<<(.*?)>>`)

// ShellExpand expands shell commands delimited by '<<' and '>>'.  Use
// at your wown risk, of course!
//
// Only called if the flag 'shellExpand' is true.
func ShellExpand(msg string) (string, error) {
	literals := shell.Split(msg, -1)
	ss := shell.FindAllStringSubmatch(msg, -1)
	acc := literals[0]
	for i, s := range ss {
		var sh = s[1]
		cmd := exec.Command("bash", "-c", sh)
		// cmd.Stdin = strings.NewReader("")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			return "", fmt.Errorf("shell error %s on %s", err, sh)
		}
		got := out.String()
		acc += got
		acc += literals[i+1]
	}
	return acc, nil
}

// JS tries to return a one-line JSON representation.  Failing that,
// returns some JSON representing the marshalling error.
func JS(x interface{}) string {
	bs, err := json.Marshal(&x)
	if err != nil {
		bs, _ = json.Marshal(map[string]interface{}{
			"error": err.Error(),
			"on":    fmt.Sprintf("%#v", x),
		})
	}
	return string(bs)
}
