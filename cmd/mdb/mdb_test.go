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

package main

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	. "github.com/Comcast/sheens/util/testutil"
)

func TestMain(t *testing.T) {
	h, err := NewHost("../../specs", "libs")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spec, err := h.GetSpec(ctx, &crew.SpecSource{
		Name: "double.yaml",
	})
	if err != nil {
		t.Fatal(err)
	}

	nodeName := "start"
	bs := core.NewBindings()
	mid := "doubler"

	m := crew.Machine{
		Id: mid,
		State: &core.State{
			NodeName: nodeName,
			Bs:       bs,
		},
		Specter: spec,
	}

	c := h.crew

	c.Lock()
	_, have := c.Machines[mid]
	if !have {
		c.Machines[mid] = &m
	}
	c.Unlock()

	if have {
		t.Fatal(fmt.Errorf(`machine "%s" exists`, mid))
	}

	msg := Dwimjs(`{"double":3}`)
	walkeds, err := h.Process(ctx, msg, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, stride := range walkeds["doubler"].Strides {
		if 0 < len(stride.Emitted) {
			log.Println(JS(stride.Emitted))
		}
	}
}
