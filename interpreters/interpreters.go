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


// Package interpreters is an example set of action interpreters that
// are available in this repo.
package interpreters

import (
	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/interpreters/ecmascript"
	"github.com/Comcast/sheens/interpreters/noop"
)

// Standard returns a map of interpreters that includes ECMAScript,
// ECMAScript with some extensions, and a no-op interpreter.
//
// See the code and subdirectories for details.
func Standard() core.InterpretersMap {
	is := core.NewInterpretersMap()

	es := ecmascript.NewInterpreter()
	is["ecmascript"] = es
	is["ecmascript-5.1"] = es
	is[""] = es // Default

	ext := ecmascript.NewInterpreter()
	ext.Extended = true
	is["ecmascript-ext"] = ext
	is["ecmascript-5.1-ext"] = ext

	is["noop"] = noop.NewInterpreter()

	// For backwards compatibility
	is["goja"] = ext

	return is
}
