/* Copyright 2018-2019 Comcast Cable Communications Management, LLC
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


// Package core provides the core gear for specification-driven
// message processing.  These specifications are structured as state
// transitions rules based on pattern matching against either a
// pending message or the state's current Bindings.
//
// The primary type is Spec(ification), and the primary method is
// Walk().  A Spec specifies how to transition from one State to
// another State.  A State is a node name (a string) and a set of
// Bindings (a map[string]interface{}).
//
// A Spec can use arbitrary code for actions (and guards).  When a
// Spec is Compiled, the compiler looks for ActionSources, each of
// which should specify an Interpreter. An Interpreter should know how
// to Compile and Exec an ActionSource.  Alternately, a native Spec
// can provide an ActionFunc implemented in Go.
//
// Ideally an Action does not block or perform any IO.  Instead, an
// Action returns a structure that includes updated Bindings and zero
// or more messages to emit.  (This structure can also include tracing
// and diagnostic messages.)  Therefore, an action should have no side
// effects.
//
// In order for an action to influence the world in some way, the
// package user must do something with messages that actions emit.
// For example, an application could provide a mechanism for certain
// messages to result in HTTP requests.  (Such an HTTP request would
// then result in the subsequent response (or error) to be forwarded
// as a message for further processing.)
//
// To use this package, make a Spec. Then Compile() it.  You might
// also want to Analyze() it.  Then, given an initial State an a
// sequence of messages, you can Walk() to the next State.
//
// See https://github.com/Comcast/sheens for an overview.
package core
