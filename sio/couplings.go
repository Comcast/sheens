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


package sio

import (
	"context"

	"github.com/Comcast/sheens/crew"
)

// Couplings provide channels for message input, results output, and
// persistence.
//
// For example, an implementation could couple a crew to an MQTT
// broker (for IO).  For persistence, an implementation could use
// https://github.com/etcd-io/bbolt, DynamoDB, SQLite, etc.
type Couplings interface {
	// Start initializes the Couplings.
	Start(context.Context) error

	// IO returns the input and result channels.
	//
	// Consumer can see all emitted messages and state updates via
	// the Result(s).
	IO(context.Context) (chan interface{}, chan *Result, chan bool, error)

	// Read (optionally) returns an initial set of machines.
	//
	// An implementation that supports persistence would read
	// machine state and pass it to this method.
	Read(context.Context) (map[string]*crew.Machine, error)

	// Stop shuts down the Couplings.
	Stop(context.Context) error
}
