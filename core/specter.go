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
	"sync/atomic"
	"unsafe"
)

// Specter enables other things to manifest themselves as Specs.
//
// Specters can be spooky.
//
// A Spec is itself a Specter.  An UpdatableSpec is also a Specter,
// but it's not itself Spec.
//
// Specter is not used anywhere in this package. It's defined here for
// convenience and encouragement.
type Specter interface {
	Spec() *Spec

	// Might want to bring back GetName() string
}

// Spec makes any Spec a Specter.
func (s *Spec) Spec() *Spec {
	return s
}

// UpdatableSpec is a scary yet handy Specter with an underlying Spec
// that can be changed at any time.
//
// This capability motivated Specters.
type UpdatableSpec struct {
	spec unsafe.Pointer // *Spec
}

// NewUpdatableSpec makes one with the given initial spec, which can
// be changed later via SetSpec.
func NewUpdatableSpec(spec *Spec) *UpdatableSpec {
	return &UpdatableSpec{
		spec: unsafe.Pointer(spec),
	}
}

// SetSpec atomically changes the underlying spec.
func (s *UpdatableSpec) SetSpec(spec *Spec) error {
	atomic.StorePointer(&s.spec, unsafe.Pointer(spec))
	return nil
}

// Spec implements the Specter interface.
func (s *UpdatableSpec) Spec() *Spec {
	return (*Spec)(atomic.LoadPointer(&s.spec))
}
