/* Copyright 2018 Comcast Cable Communications Management, LLC
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

package tools

import (
	"bytes"
	"testing"
)

func TestRenderSpecHTML(t *testing.T) {

	t.Run("withoutGraph", func(t *testing.T) {
		out := bytes.NewBuffer(make([]byte, 0, 1024*128))

		err := ReadAndRenderSpecPage("../specs/double.yaml", []string{"spec.css"}, out, false)

		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("withGraph", func(t *testing.T) {
		out := bytes.NewBuffer(make([]byte, 0, 1024*128))

		err := ReadAndRenderSpecPage("../specs/double.yaml", []string{"spec.css"}, out, true)

		if err != nil {
			t.Fatal(err)
		}
	})

}
