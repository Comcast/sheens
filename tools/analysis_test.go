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
	"testing"

	"github.com/Comcast/sheens/core"
)

func TestAnalysis(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spec, err := core.TurnstileSpec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Analyze(spec); err != nil {
		t.Fatal(err)
	}
}
