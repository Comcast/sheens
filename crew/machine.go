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

package crew

import (
	"context"

	"github.com/Comcast/sheens/core"
)

// <machines,message> → <walks> → <machines,messages>
//
// Side-effecting operations can occur during the second arrow.  When
// there are side-effecting operations, we write new state.

// Machine is a triple: id, core.Spec, and core.State.
type Machine struct {
	Id      string       `json:"id,omitempty"`
	Specter core.Specter `json:"-" yaml:"-"`
	State   *core.State  `json:"state"`

	// SpecSource_ is here only to facilitate serialization and
	// deserialization.  This field is not used anywhere in this
	// package.
	SpecSource *SpecSource `json:"spec,omitempty"`
}

// Update overlays the given machine data on the target machine.
//
//
// State and SpecSource (if any) are copied.
//
// Not thread-safe.
func (m *Machine) Update(overlay *Machine) {
	if overlay.Id != "" {
		m.Id = overlay.Id
	}
	if overlay.Specter != nil {
		m.Specter = overlay.Specter
	}
	if overlay.State != nil {
		m.State = overlay.State.Copy()
	}
	if overlay.SpecSource != nil {
		m.SpecSource = overlay.SpecSource.Copy()
	}
}

// Copy returns a new Machine with the same id, same spec, and a copy
// of the machine's state.
func (m *Machine) Copy() *Machine {
	return &Machine{
		Id:         m.Id,
		Specter:    m.Specter,    // Not copied!  ToDo?
		SpecSource: m.SpecSource, // Not copied!  ToDo?
		State:      m.State.Copy(),
	}
}

// SpecSource aspires to hold the origin of a specification.
//
// Currently a source for a Spec can either be a name, a URL, or maybe
// given explicitly as a string in an unspecified syntax.
//
// Just how a SpecSource is used is up to the application.
type SpecSource struct {
	// Name is an optional string that could be used by a resolve
	// to obtain some spec.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// URL is an optional pointer to a spec.
	URL string `json:"url,omitempty" yaml:"url,omitempty"`

	// Source is a optional string representing a spec (in a
	// representation determined by the application).
	Source string `json:"source,omitempty" yaml:"source,omitempty"`

	// Inline is an optional actual spec right here.
	Inline *core.Spec `json:"inline,omitempty" yaml:",omitempty"`
}

// NewSpecSource creates a SpecSource with the given name.
func NewSpecSource(name string) *SpecSource {
	return &SpecSource{
		Name: name,
	}
}

// Copy makes a (deep?) copy of the given SpecSource.
func (s *SpecSource) Copy() *SpecSource {
	return &SpecSource{
		Name:   s.Name,
		URL:    s.URL,
		Source: s.Source,
		Inline: s.Inline,
	}
}

// SpecProvider can FindSpec given a SpecSource.
type SpecProvider interface {
	FindSpec(ctx context.Context, s *SpecSource) (*core.Spec, error)
}
