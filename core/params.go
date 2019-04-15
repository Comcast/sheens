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


package core

// ParamSpec is a strawman struct to represent data about required and
// optional Machine parameters (which are just initial bindings).
//
// Probably would be better to start with an interface, but then
// deserialization is more trouble.
//
// We're not over-thinking this struct right now.  Really just a
// strawman.  Should be much better.
type ParamSpec struct {

	// Doc describes the spec in English and Markdown.  Audience
	// is developers, not users.
	Doc string `json:"doc,omitempty" yaml:",omitempty"`

	// PrimitiveType is (for now) any string.
	PrimitiveType string `json:"primitiveType" yaml:"primitiveType"`

	// Default is the default value for a parameter used in case a value is not given in the initial bindings
	Default interface{} `json:"default"`

	// Optional means that the parameter is not required!
	Optional bool `json:"optional,omitempty" yaml:",omitempty"`

	// IsArray specifies whether a value must be an array of
	// values of the PrimitiveType.
	IsArray bool `json:"isArray,omitempty" yaml:"isArray,omitempty"`

	// SemanticType is (for now) any string.
	SemanticType string `json:"semanticType,omitempty" yaml:"semanticType,omitempty"`

	// MinCardinality is the minimum number of values.
	MinCardinality int `json:"minCard,omitempty" yaml:"minCard,omitempty"`

	// MaxCardinality is the maximum number of values.
	//
	// A MaxCardinality greater than one implies the param value
	// IsArray.
	MaxCardinality int `json:"maxCard,omitempty" yaml:"maxCard,omitempty"`

	// Predicate isn't implemented.
	//
	// Could be a predicate that could further validate a
	// parameter value.
	Predicate interface{} `json:"predicate,omitempty" yaml:",omitempty"`

	// Advisory indicates that a violation of this spec is a
	// warning, not an error.
	Advisory bool `json:"advisory,omitempty" yaml:",omitempty"`
}

// Valid should return an error if the given spec is bad for some
// reason.
//
// Currently just returns nil.
//
// Probably shouldn't return an error, but we'll just go with that for
// now.
func (s *ParamSpec) Valid() error {
	return nil
}

// ValueCompilesWith checks that the given value complies with the
// spec. Returns an error if not.
//
// Currently just returns nil.
//
// Probably shouldn't return an error, but we'll just go with that for
// now.
func (s *ParamSpec) ValueCompilesWith(x interface{}) error {
	return nil
}
