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


package tools

import (
	"context"
	"os"
	"testing"

	"github.com/Comcast/sheens/core"
)

func TestDot(t *testing.T) {
	filename := "g.dot"

	out, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Remove(filename); err != nil {
			t.Fatal(err)
		}
	}()

	spec, err := core.TurnstileSpec(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if err := Dot(spec, out, "", ""); err != nil {
		t.Fatal(err)
	}

}
