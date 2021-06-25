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
	"errors"
	"testing"

	. "github.com/Comcast/sheens/match"
)

var (
	// ensures that the InterpretersMap is an Interpreters.
	_ Interpreters = InterpretersMap{}

	goodAS = &ActionSource{
		Interpreter: "test interpreter",
		Source:      "test source",
		Binds: []Bindings{
			{"a": "b", "c": 5, "d": false, "e": 1.25},
			{"test": true, "binding": true, "two": false},
		},
	}

	errTestCompile = errors.New("test interpreter failed to compile")
)

type testInterpreter struct {
	errOnCompile bool
}

func (i *testInterpreter) Compile(_ context.Context, _ interface{}) (interface{}, error) {
	if i.errOnCompile {
		return nil, errTestCompile
	}
	return nil, nil
}

func (*testInterpreter) Exec(_ context.Context, _ Bindings, _ StepProps, _ interface{}, _ interface{}) (*Execution, error) {
	return nil, nil
}

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

func TestInterpretersMap(t *testing.T) {
	goodInterpreter := &testInterpreter{}
	iMap := NewInterpretersMap()
	if iMap == nil {
		t.Fatalf("NewInterpretersMap should return non-nil value")
	}
	iMap["good"] = goodInterpreter
	iMap["bad"] = nil
	tests := []struct {
		description    string
		name           string
		expectedResult Interpreter
	}{
		{
			description:    "Success",
			name:           "good",
			expectedResult: goodInterpreter,
		},
		{
			description:    "Nil key",
			name:           "bad",
			expectedResult: nil,
		},
		{
			description:    "Missing key",
			name:           "missing",
			expectedResult: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			r := iMap.Find(tc.name)
			if r != tc.expectedResult {
				t.Fatalf("expected %v interpreter but received %v",
					tc.expectedResult, r)
			}
		})
	}
}

func TestActionSourceCopy(t *testing.T) {
	var (
		nilAS *ActionSource = nil
	)
	tests := []struct {
		description  string
		actionsource *ActionSource
		expectedCopy *ActionSource
	}{
		{
			description:  "Success",
			actionsource: goodAS,
			expectedCopy: goodAS,
		},
		{
			description:  "Empty values",
			actionsource: &ActionSource{},
			expectedCopy: &ActionSource{},
		},
		{
			description:  "Nil action source",
			actionsource: nilAS,
			expectedCopy: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			copy := tc.actionsource.Copy()

			// deal with anything nil first.
			if tc.expectedCopy == nil && copy != nil ||
				tc.expectedCopy != nil && copy == nil {
				t.Fatalf("expected %v copy but received %v", tc.expectedCopy,
					copy)
			}
			if tc.expectedCopy == nil || copy == nil {
				return
			}

			if &tc.expectedCopy == &copy {
				t.Fatalf("expected new address for copy")
			}
			if tc.expectedCopy.Interpreter != copy.Interpreter ||
				tc.expectedCopy.Source != copy.Source ||
				len(tc.expectedCopy.Binds) != len(copy.Binds) {
				t.Fatalf("copies don't match; expected %v but received %v",
					tc.expectedCopy, copy)
			}

			// is this too much? we can remove it.
			for i, b := range tc.expectedCopy.Binds {
				if len(b) != len(copy.Binds[i]) {
					t.Fatalf("copies don't match; expected %v but received %v",
						tc.expectedCopy, copy)
				}
				for k, v := range b {
					if v != copy.Binds[i][k] {
						t.Fatalf("copies don't match; expected %v but received %v",
							tc.expectedCopy, copy)
					}
				}
			}
		})
	}
}

func TestActionSourceCompile(t *testing.T) {
	interpreters := InterpretersMap{
		"good":           &testInterpreter{},
		"compile issues": &testInterpreter{errOnCompile: true},
	}
	tests := []struct {
		description  string
		actionsource *ActionSource
		interpreters Interpreters
		expectedErr  error
	}{
		{
			description: "Success",
			actionsource: &ActionSource{
				Interpreter: "good",
				Binds:       goodAS.Binds,
			},
			interpreters: interpreters,
		},
		{
			description:  "Not found error",
			actionsource: &ActionSource{Interpreter: "nope"},
			interpreters: interpreters,
			expectedErr:  InterpreterNotFound,
		},
		{
			description:  "Not found error with default interpreters",
			actionsource: &ActionSource{Interpreter: "good"},
			expectedErr:  InterpreterNotFound,
		},
		{
			description:  "Compile error",
			actionsource: &ActionSource{Interpreter: "compile issues"},
			interpreters: interpreters,
			expectedErr:  errTestCompile,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			action, err := tc.actionsource.Compile(context.Background(), tc.interpreters)
			if tc.expectedErr != err {
				t.Fatalf("expected error %v but received %v", tc.expectedErr, err)
			}
			if tc.expectedErr != nil {
				if action != nil {
					t.Fatalf("expected nil action but received %v", action)
				}
				return
			}
			if action == nil {
				t.Fatalf("expected non-nil action")
			}
			if len(tc.actionsource.Binds) != len(action.Binds()) {
				t.Errorf("expected binds %v but received %v", tc.actionsource.Binds, action.Binds())
			}
		})
	}
}
