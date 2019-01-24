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

// Package main is a little command-line MQTT client.
//
// Commands:
//
//   qos QOS               Set the QoS for subsequent operations.
//   sub TOPIC             Subscribe to the given topic.
//   unsub TOPIC           Unsubscribe from the given topic.
//   retain (true|false)   Set retain flag for subsequent pubs.
//   pub TOPIC MSG         Publish MSG to the given TOPIC.
//   sleep DURATION        Sleep for DURATION (Go syntax).
//
package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Comcast/sheens/sio"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {

	var (
		// Follow mosquitto_sub command line args.

		broker      = flag.String("h", "tcp://localhost", "Broker hostname")
		clientId    = flag.String("i", "", "Client id")
		port        = flag.Int("p", 1883, "Broker port")
		keepAlive   = flag.Int("k", 10, "Keep-alive in seconds")
		userName    = flag.String("u", "", "Username")
		password    = flag.String("P", "", "Password")
		willTopic   = flag.String("will-topic", "", "Optional will topic")
		willPayload = flag.String("will-payload", "", "Optional will message")
		willQoS     = flag.Int("will-qos", 0, "Optional will QoS")
		willRetain  = flag.Bool("will-retain", false, "Optional will retention")
		reconnect   = flag.Bool("reconnect", false, "Automatically attempt to reconnect")
		clean       = flag.Bool("c", true, "Clean session")
		quiesce     = flag.Int("quiesce", 100, "Disconnection quiescence (in milliseconds)")

		certFilename = flag.String("cert", "", "Optional cert filename")
		keyFilename  = flag.String("key", "", "Optional key filename")
		insecure     = flag.Bool("insecure", false, "Skip broker cert checking")
		caFilename   = flag.String("cafile", "", "Optional CA cert filename")
		caPath       = flag.String("capath", "", "Optional path to CA cert filename") // Why separate?
		shellExpand  = flag.Bool("sh", true, "Enable shell expansion (<<...>>)")
	)

	flag.Parse()

	mqtt.ERROR = log.New(os.Stderr, "mqtt.error", 0)

	opts := mqtt.NewClientOptions()

	*broker = fmt.Sprintf("%s:%d", *broker, *port)
	opts.AddBroker(*broker)
	opts.SetClientID(*clientId)
	opts.SetKeepAlive(time.Second * time.Duration(*keepAlive))
	opts.SetPingTimeout(10 * time.Second)

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

LOOP:
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
			if s, err = sio.ShellExpand(string(line)); err != nil {
				fmt.Printf("shell expansion error: %s", err)
				continue
			}
		}
		parts := space.Split(strings.TrimSpace(s), 3)
		switch parts[0] {
		case "sleep":
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
		case "sub":
			if len(parts) != 2 {
				fmt.Printf("error: sub TOPIC")
				continue
			}
			if t := c.Subscribe(parts[1], qos, nil); t.Wait() && t.Error() != nil {
				fmt.Printf("subscribe error: %s", t.Error())
			}
		case "unsub":
			if len(parts) != 2 {
				fmt.Printf("error: unsub TOPIC")
				continue
			}
			if t := c.Unsubscribe(parts[1]); t.Wait() && t.Error() != nil {
				fmt.Printf("unsubscribe error: %s", t.Error())
			}
		case "retain":
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
		case "pub":
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

func parseTopic(s string) (string, byte) {
	var topic string
	var qos byte
	if _, err := fmt.Sscanf(strings.Replace(s, ":", " ", 1), "%s %d", &topic, &qos); err != nil {
		return topic, qos
	}
	return s, 0
}
