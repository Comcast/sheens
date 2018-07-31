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

package core

import (
	"context"
	"testing"

	. "github.com/Comcast/sheens/match"
)

func TestPermanentBindings(t *testing.T) {
	if !Exp_PermanentBindings {
		return
	}

	ctx := context.Background()

	action := &FuncAction{
		F: func(ctx context.Context, bs Bindings, props StepProps) (*Execution, error) {
			return &Execution{
				Bs:     bs.Remove("ephemeral", "permament!"),
				Events: newEvents(),
			}, nil
		},
	}

	bs := NewBindings()
	bs["ephemeral"] = "queso"
	bs["permament!"] = "tacos"
	exe, err := action.Exec(ctx, bs.Copy(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, have := exe.Bs["ephemeral"]; have {
		t.Fatal("ephemeral wasn't")
	}
	if _, have := exe.Bs["permament!"]; !have {
		t.Fatal("permament wasn't")
	}
}
