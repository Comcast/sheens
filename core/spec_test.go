/* Copyright 2021 Comcast Cable Communications Management, LLC
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
	"strings"
	"testing"
)

func mockPatternParser(val interface{}, err error) func(string, interface{}) (interface{}, error) {
	return func(_ string, _ interface{}) (interface{}, error) {
		return val, err
	}
}

func TestParsePatterns(t *testing.T) {
	goodNodes := map[string]*Node{
		"yay": {
			Branches: &Branches{
				Branches: []*Branch{
					{Pattern: "something nice"},
					{Pattern: "something else"},
					{Pattern: "something??"},
				},
			},
		},
		"next thing": {
			Branches: &Branches{
				Branches: []*Branch{{Pattern: 5}},
			},
		},
	}
	emptyNodes := map[string]*Node{
		"nowhere": {},
		"a little further": {
			Branches: &Branches{},
		},
		"getting there": {
			Branches: &Branches{
				Branches: []*Branch{
					nil,
					nil,
					{Pattern: nil},
				},
			},
		},
	}
	testErr := errors.New("test parser error")
	tests := []struct {
		description   string
		parser        func(string, interface{}) (interface{}, error)
		nodes         map[string]*Node
		expectedNodes map[string]*Node
		expectedErr   error
	}{
		{
			description: "Success",
			parser:      mockPatternParser("a", nil),
			nodes:       goodNodes,
			expectedNodes: map[string]*Node{
				"yay": {
					Branches: &Branches{
						Branches: []*Branch{{Pattern: "a"},
							{Pattern: "a"}, {Pattern: "a"},
						},
					},
				},
				"next thing": {
					Branches: &Branches{
						Branches: []*Branch{{Pattern: "a"}},
					},
				},
			},
		},
		{
			description:   "Success with default parser",
			nodes:         goodNodes,
			expectedNodes: goodNodes,
		},
		{
			description: "Success with no nodes",
			parser:      mockPatternParser("c", nil),
		},
		{
			description:   "Success with no branches",
			parser:        mockPatternParser("d", nil),
			nodes:         emptyNodes,
			expectedNodes: emptyNodes,
		},
		{
			description: "Parse failure",
			parser:      mockPatternParser("e", testErr),
			nodes:       goodNodes,
			expectedErr: testErr,
		},
		{
			description: "Canonicalize failure",
			parser:      mockPatternParser([]interface{}{func() string { return ":(" }}, nil),
			nodes:       goodNodes,
			expectedErr: errors.New("unsupported type"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			s := &Spec{
				PatternParser: tc.parser,
				Nodes:         tc.nodes,
			}
			err := s.ParsePatterns(context.Background())
			if s.PatternParser == nil {
				t.Error("pattern parser shouldn't be nil")
			}

			if err == nil || tc.expectedErr == nil {
				if err != tc.expectedErr {
					t.Errorf("expected %v error but received %v",
						tc.expectedErr, err)
				}
			} else if !strings.Contains(err.Error(), tc.expectedErr.Error()) {
				t.Errorf("error %s doesn't include expected string %s",
					err, tc.expectedErr)
			}

			// if we expected an error, leave before the nightmare begins.
			if tc.expectedErr != nil {
				return
			}

			// start checking nodes' branches' patterns...
			if len(tc.expectedNodes) != len(s.Nodes) {
				t.Fatalf("nodes don't match; expected %v but received %v",
					tc.expectedNodes, s.Nodes)
			}
			for k, n := range tc.expectedNodes {
				if n == nil {
					if s.Nodes[k] != nil {
						t.Fatalf("nodes don't match; expected %v but received %v",
							tc.expectedNodes, s.Nodes)
					}
					continue
				}
				if n.Branches == nil {
					if s.Nodes[k].Branches != nil {
						t.Fatalf("nodes don't match; expected %v but received %v",
							tc.expectedNodes, s.Nodes)
					}
					continue
				}
				if s.Nodes[k] == nil || s.Nodes[k].Branches == nil {
					t.Fatalf("nodes don't match; expected %v but received %v",
						tc.expectedNodes, s.Nodes)
				}
				expectedBranches := n.Branches.Branches
				sBranches := s.Nodes[k].Branches.Branches
				if len(expectedBranches) != len(sBranches) {
					t.Fatalf("nodes don't match; expected %v but received %v",
						tc.expectedNodes, s.Nodes)
				}
				for j, b := range expectedBranches {
					if b == nil {
						if sBranches[j] != nil {
							t.Fatalf("nodes don't match; expected %v but received %v",
								tc.expectedNodes, s.Nodes)
						}
						continue
					}
					if sBranches[j] == nil {
						t.Fatalf("nodes don't match; expected %v but received %v",
							tc.expectedNodes, s.Nodes)
					}
					if b.Pattern != sBranches[j].Pattern {
						t.Fatalf("nodes don't match; expected %v but received %v",
							tc.expectedNodes, s.Nodes)
					}
				}
			}
		})
	}
}

func TestCompile(t *testing.T) {
	s := &Spec{}
	err := s.Compile(context.Background(), nil, false)
	if err != nil {
		t.Fatalf("expected no error but received %v", err)
	}
}

func TestDefaultPatternParser(t *testing.T) {
	tests := []struct {
		description    string
		syntax         string
		val            interface{}
		expectedResult interface{}
		expectedErr    error
	}{
		{
			description: "Default error",
			syntax:      "clearly invalid",
			val:         "testing123",
			expectedErr: errors.New("unsupposed pattern syntax"),
		},
		{
			description:    "None syntax success",
			syntax:         "none",
			val:            struct{}{},
			expectedResult: struct{}{},
		},
		{
			description: "Empty syntax success",
		},
		{
			description:    "JSON syntax success",
			syntax:         "json",
			val:            `"test"`,
			expectedResult: "test",
		},
		{
			description: "JSON syntax non-string success",
			syntax:      "json",
		},
		{
			description: "JSON unmarshal error",
			syntax:      "json",
			val:         "",
			expectedErr: errors.New("unexpected end of JSON input"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			r, err := DefaultPatternParser(tc.syntax, tc.val)
			if tc.expectedResult != r {
				t.Errorf("expected %v pattern but received %v",
					tc.expectedResult, r)
			}
			if err == nil || tc.expectedErr == nil {
				if err != tc.expectedErr {
					t.Errorf("expected %v error but received %v",
						tc.expectedErr, err)
				}
				return
			}
			if !strings.Contains(err.Error(), tc.expectedErr.Error()) {
				t.Errorf("error %s doesn't include expected string %s",
					err, tc.expectedErr)
			}
		})
	}
}
