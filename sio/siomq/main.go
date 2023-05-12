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

// Package main is a simple single-crew sheens process that talks to
// an MQTT broker.
//
// The command line args follow those for mosquito_sub.
package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/sio"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {

	var (
		// Follow mosquito_sub command line args.

		broker      = flag.String("h", "tcp://localhost", "Broker hostname")
		clientId    = flag.String("i", "", "Client id")
		port        = flag.Int("p", 1883, "Broker port")
		keepAlive   = flag.Int("k", 600, "Keep-alive in seconds")
		userName    = flag.String("u", "", "Username")
		password    = flag.String("P", "", "Password")
		willTopic   = flag.String("will-topic", "", "Optional will topic")
		willPayload = flag.String("will-payload", "", "Optional will message")
		willQoS     = flag.Int("will-qos", 1, "Optional will QoS")
		willRetain  = flag.Bool("will-retain", false, "Optional will retention")
		reconnect   = flag.Bool("reconnect", false, "Automatically attempt to reconnect")
		clean       = flag.Bool("c", true, "Clean session")
		quiesce     = flag.Int("quiesce", 100, "Disconnection quiescence (in milliseconds)")

		certFilename = flag.String("cert", "", "Optional cert filename")
		keyFilename  = flag.String("key", "", "Optional key filename")
		insecure     = flag.Bool("insecure", false, "Skip broker cert checking")
		caFilename   = flag.String("cafile", "", "Optional CA cert filename")

		tokenKey       = flag.String("token-key-name", "CustAuth", "AWS custom authorizer token key")
		token          = flag.String("token", "", "AWS custom authorizer token")
		tokenSig       = flag.String("token-sig", "", "AWS custom authorizer token signature")
		authorizerName = flag.String("authorizer-name", "", "AWS custom authorizer name")

		subTopics = flag.String("t", "", "subscription topic(s)")
		initFile  = flag.String("init", "", "File containing mqshell 'pub' commands to execute")

		injectTopic          = flag.Bool("inject-topic", true, "put topic in map of incoming messages")
		wrapWithTopic        = flag.Bool("wrap-with-topic", false, "wrap non-maps in a map along with the topic")
		defaultOutboundTopic = flag.String("def-outbound-topic", "misc", "Default out-bound message topic")
		inTimeout            = flag.Duration("in-timeout", 5*time.Second, "timeout for in-bound queuing")
	)

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mqtt.ERROR = log.New(os.Stderr, "mqtt.error", 0)

	opts := mqtt.NewClientOptions()

	if *port != 0 {
		*broker = fmt.Sprintf("%s:%d", *broker, *port)
	}
	log.Printf("broker: %s", *broker)
	opts.AddBroker(*broker)
	opts.SetClientID(*clientId)
	opts.SetKeepAlive(time.Second * time.Duration(*keepAlive))
	opts.SetPingTimeout(10 * time.Second)

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

	io := &Couplings{
		Quiesce:              uint(*quiesce),
		SubTopics:            *subTopics,
		InjectTopic:          *injectTopic,
		WrapWithTopic:        *wrapWithTopic,
		DefaultOutboundTopic: *defaultOutboundTopic,
		InTimeout:            *inTimeout,

		incoming: make(chan interface{}),
		outbound: make(chan *sio.Result),
	}

	opts.DefaultPublishHandler = func(client mqtt.Client, msg mqtt.Message) {
		io.inHandler(ctx, client, msg)
	}

	io.Client = mqtt.NewClient(opts)

	conf := &sio.CrewConf{
		Ctl: core.DefaultControl,
	}

	c, err := sio.NewCrew(ctx, conf, io)
	if err != nil {
		panic(err)
	}
	c.Verbose = true

	if err = io.Start(ctx); err != nil {
		panic(err)
	}

	ms, err := io.Read(ctx)
	if err != nil {
		panic(err)
	}
	for mid, m := range ms {
		if err := c.SetMachine(ctx, mid, m.SpecSource, m.State); err != nil {
			panic(err)
		}
	}

	go io.outLoop(ctx)

	go func() {
		if *initFile != "" {
			in, err := ioutil.ReadFile(*initFile)
			if err != nil {
				panic(err)
			}
			for _, line := range strings.Split(string(in), "\n") {
				line, err = sio.ShellExpand(line) // ToDo: Warn/switch!
				if err != nil {
					panic(fmt.Errorf("shell expansion error %s", err))
				}
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, " ", 3)
				switch strings.TrimSpace(parts[0]) {
				case "pub":
					if len(parts) != 3 {
						log.Printf("bad init line '%s'", line)
						continue
					}
					topic := parts[1]
					msg := parts[2]
					io.consume(ctx, topic, []byte(msg))
				case "echo":
				default:
					log.Printf("ignoring line '%s'", line)
				}
			}
		}
	}()

	if err := c.Loop(ctx); err != nil {
		panic(err)
	}

	if false {

		if err = io.Stop(context.Background()); err != nil {
			panic(err)
		}
	}
}

// Couplings is an sio.Couplings.
type Couplings struct {
	Client               mqtt.Client
	Quiesce              uint
	SubTopics            string
	InjectTopic          bool
	WrapWithTopic        bool
	DefaultOutboundTopic string

	InTimeout time.Duration

	c        *sio.Crew
	incoming chan interface{}
	outbound chan *sio.Result
}

func (c *Couplings) consume(ctx context.Context, topic string, payload []byte) {
	var x interface{}
	if err := json.Unmarshal(payload, &x); err != nil {
		log.Printf("Couldn't JSON-parse payload: %s", payload)
		x = string(payload)
	} else {
		if m, is := x.(map[string]interface{}); is {
			if c.InjectTopic {
				m["topic"] = topic
			}
		} else {
			if c.WrapWithTopic {
				x = map[string]interface{}{
					"topic":   topic,
					"payload": string(payload),
				}
			}
		}
	}

	to := time.NewTimer(c.InTimeout)

	select {
	case <-ctx.Done():
		log.Printf("Publisher not publishing due to ctx.Done()")
	case c.incoming <- x:
		log.Printf("Couplings forwarded incoming %s", payload)
	case <-to.C:
		log.Printf("Publisher not publishing due to stall ('%s','%s')", topic, payload)
	}
}

// inHandler is a Paho publish handler, which is used to handle
// messages send to us from the MQTT broker due to our subscriptions.
func (c *Couplings) inHandler(ctx context.Context, client mqtt.Client, msg mqtt.Message) {
	log.Printf("incoming: %s %s\n", msg.Topic(), msg.Payload())
	c.consume(ctx, msg.Topic(), msg.Payload())
}

// Start creates the MQTT session.
func (c *Couplings) Start(ctx context.Context) error {
	log.Printf("Attempting to connected to broker")
	if token := c.Client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	log.Printf("Connected to broker")

	for _, topic := range strings.Split(c.SubTopics, ",") {
		topic, qos := parseTopic(topic)
		if topic == "" {
			continue
		}
		log.Printf("Subscribing to %s (%d)", topic, qos)
		if t := c.Client.Subscribe(topic, qos, nil); t.Wait() && t.Error() != nil {
			return t.Error()
		}
		log.Printf("Subscribed to %s (%d)", topic, qos)
	}
	log.Printf("Couplings started")

	return nil
}

// IO starts a loop to publish out-bound Results and forward incoming
// messages.
func (c *Couplings) IO(ctx context.Context) (chan interface{}, chan *sio.Result, error) {
	return c.incoming, c.outbound, nil
}

// outLoop forwards messages outbound from the Crew to the MQTT
// broker.
func (c *Couplings) outLoop(ctx context.Context) error {
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case r := <-c.outbound:
			for _, xs := range r.Emitted {
				for _, x := range xs {
					topic, qos := parseTopic(c.DefaultOutboundTopic)
					if m, is := x.(map[string]interface{}); is {
						if t, have := m["topic"]; have {
							if s, is := t.(string); is {
								topic = s
							}
						}
						if n, have := m["qos"]; have {
							if f, is := n.(float64); is {
								qos = byte(f)
							} else {
								log.Printf("Warning: ignoring qos %#v %T", n, n)
							}
						}
					}
					js, err := json.Marshal(x)
					if err != nil {
						log.Printf("Failed to marshal %#v", x)
						continue
					}
					log.Printf("Publishing %s %s", topic, js)
					token := c.Client.Publish(topic, qos, false, js)
					token.Wait()
					if token.Error() != nil {
						log.Fatalf("Publish error: %s", token.Error())
					}
					log.Printf("Published to %s", topic)
				}

				// Where we could store state changes.
				for mid, m := range r.Changed {
					if false {
						log.Printf("Update %s %s\n", mid, sio.JShort(m))
					}
				}
			}
		}
	}
	return nil
}

// Read currently does nothing.
//
// No persistence yet.
func (c *Couplings) Read(context.Context) (map[string]*crew.Machine, error) {
	return nil, nil
}

// Stop terminates the MQTT session.
func (c *Couplings) Stop(context.Context) error {
	log.Printf("Disconnecting")
	c.Client.Disconnect(c.Quiesce)
	return nil
}

// parseTopic can extract QoS from a topic name of the form TOPIC:QOS.
func parseTopic(s string) (string, byte) {
	var topic string
	var qos byte
	if _, err := fmt.Sscanf(strings.Replace(s, ":", " ", 1), "%s %d", &topic, &qos); err != nil {
		return topic, qos
	}
	return s, 0
}
