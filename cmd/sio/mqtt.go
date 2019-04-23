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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/sio"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTCouplings is an sio.Couplings for an MQTT client.
type MQTTCouplings struct {
	Client               mqtt.Client
	Quiesce              uint
	SubTopics            string
	InjectTopic          bool
	WrapWithTopic        bool
	DefaultOutboundTopic string

	InTimeout time.Duration

	*sio.JSONStore

	c        *sio.Crew
	incoming chan interface{}
	outbound chan *sio.Result
	done     chan bool
}

func NewMQTTCouplings(args []string) (*MQTTCouplings, *flag.FlagSet) {
	var (
		// Follow mosquitto_sub command line args.

		fs = flag.NewFlagSet("mq", flag.ExitOnError)

		broker      = fs.String("h", "tcp://localhost", "Broker hostname")
		clientId    = fs.String("i", "", "Client id")
		port        = fs.Int("p", 1883, "Broker port")
		keepAlive   = fs.Int("k", 10, "Keep-alive in seconds")
		userName    = fs.String("u", "", "Username")
		password    = fs.String("P", "", "Password")
		willTopic   = fs.String("will-topic", "", "Optional will topic")
		willPayload = fs.String("will-payload", "", "Optional will message")
		willQoS     = fs.Int("will-qos", 0, "Optional will QoS")
		willRetain  = fs.Bool("will-retain", false, "Optional will retention")
		reconnect   = fs.Bool("reconnect", false, "Automatically attempt to reconnect")
		clean       = fs.Bool("c", true, "Clean session")
		quiesce     = fs.Int("quiesce", 100, "Disconnection quiescence (in milliseconds)")

		certFilename = fs.String("cert", "", "Optional cert filename")
		keyFilename  = fs.String("key", "", "Optional key filename")
		insecure     = fs.Bool("insecure", false, "Skip broker cert checking")
		caFilename   = fs.String("cafile", "", "Optional CA cert filename")
		caPath       = fs.String("capath", "", "Optional path to CA cert filename") // Why separate?

		subTopics = fs.String("t", "", "subscription topic(s)")

		injectTopic          = fs.Bool("inject-topic", true, "put topic in map of incoming messages")
		wrapWithTopic        = fs.Bool("wrap-with-topic", false, "wrap non-maps in a map along with the topic")
		defaultOutboundTopic = fs.String("def-outbound-topic", "misc", "Default out-bound message topic")
		inTimeout            = fs.Duration("in-timeout", time.Second, "timeout for in-bound queuing")
	)

	if args == nil {
		return nil, fs
	}

	fs.Parse(args)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mqtt.ERROR = log.New(os.Stderr, "mqtt.error", 0)

	opts := mqtt.NewClientOptions()

	*broker = fmt.Sprintf("%s:%d", *broker, *port)
	opts.AddBroker(*broker)
	opts.SetClientID(*clientId)
	opts.SetKeepAlive(time.Second * time.Duration(*keepAlive))
	// opts.SetPingTimeout(10 * time.Second)

	opts.Username = *userName
	opts.Password = *password
	opts.AutoReconnect = *reconnect
	opts.CleanSession = *clean

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
	{
		if *caPath != "" {
			if rootCAs, _ = x509.SystemCertPool(); rootCAs == nil {
				rootCAs = x509.NewCertPool()
				log.Printf("Including system CA certs")
			}

			if !strings.HasSuffix(*caPath, "/") {
				*caPath += "/"
			}
			filename := *caPath + *caFilename
			certs, err := ioutil.ReadFile(filename)
			log.Fatalf("couldn't read '%s': %s", filename, err)

			if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
				log.Println("No certs appended, using system certs only")
			}
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

	io := &MQTTCouplings{
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

	return io, fs
}

// inHandler is a Paho publish handler, which is used to handle
// messages send to us from the MQTT broker due to our subscriptions.
func (c *MQTTCouplings) inHandler(ctx context.Context, client mqtt.Client, msg mqtt.Message) {
	log.Printf("incoming: %s %s\n", msg.Topic(), msg.Payload())
	var (
		x       interface{}
		payload = msg.Payload()
		topic   = msg.Topic()
	)

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
		log.Printf("Publisher not publishing due to stall")
	}

}

// Start creates the MQTT session.
func (c *MQTTCouplings) Start(ctx context.Context) error {
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
	}
	log.Printf("Couplings started")

	return nil
}

// IO starts a loop to publish out-bound Results and forward incoming
// messages.
func (c *MQTTCouplings) IO(ctx context.Context) (chan interface{}, chan *sio.Result, chan bool, error) {
	return c.incoming, c.outbound, c.done, nil
}

// outLoop forwards messages outbound from the Crew to the MQTT
// broker.
func (c *MQTTCouplings) outLoop(ctx context.Context) error {
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
								log.Printf("warning: ignoring qos %#v %T", n, n)
							}
						}
					}
					js, err := json.Marshal(x)
					if err != nil {
						log.Printf("Failed to marshal %#v", x)
						continue
					}
					token := c.Client.Publish(topic, qos, false, js)
					token.Wait()
					if token.Error() != nil {
						log.Fatalf("Publish error: %s", token.Error())
					}
				}
			}
			if err := c.Update(r); err != nil {
				E(err, "Update")
				return err
			}
		}
	}
	return nil
}

func (c *MQTTCouplings) Read(ctx context.Context) (map[string]*crew.Machine, error) {
	return c.JSONStore.Read(ctx)
}

// Stop terminates the MQTT session.
func (c *MQTTCouplings) Stop(ctx context.Context) error {
	log.Printf("Disconnecting")
	c.Client.Disconnect(c.Quiesce)
	close(c.done)
	// ToDo: Ensure no more writes.
	c.JSONStore.WriteState(ctx)
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
