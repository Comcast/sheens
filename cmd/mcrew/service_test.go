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
	"log"
	"os"
	"time"

	"testing"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	. "github.com/Comcast/sheens/util/testutil"
)

func TestServiceBasic(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := testServiceBasic(ctx, t)
	s.store.Close(ctx) // ToDo: Check error.
}

func testServiceBasic(ctx context.Context, t *testing.T) *Service {

	dbFile := "test.db"

	removeDBFile := func() {
		if _, err := os.Stat(dbFile); err == nil {
			log.Printf("removing dbFile %s", dbFile)
			if err := os.Remove(dbFile); err != nil {
				t.Fatal(err)
			}
		}
	}

	removeDBFile()

	defer removeDBFile()

	s, err := NewService(ctx, "../../specs", dbFile, "lib")
	if err != nil {
		t.Fatal(err)
	}

	s.Emitted = make(chan interface{}, 8)
	s.Processing = make(chan interface{}, 8)

	op := COp{
		Add: &OpAdd{
			Machine: &crew.Machine{
				Id: "double",
				SpecSource: &crew.SpecSource{
					Name: "double",
				},
				State: &core.State{
					NodeName: "start",
					Bs:       nil,
				},
			},
		},
	}

	if err = op.Do(ctx, s); err != nil {
		t.Fatal(err)
	}

	op = COp{
		Process: &OpProcess{
			// Render:  true,
			Message: Dwimjs(`{"double":2}`),
		},
	}

	if err = op.Do(ctx, s); err != nil {
		t.Fatal(err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-s.Processing:
				Logf("processing %s", JS(m))
			}
		}
	}()

	m := <-s.Emitted
	Logf("emitted %s", JS(m))

	// s.store.Close(ctx) // ToDo: Check error.

	return s
}

func TestServiceRemMachine(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := testServiceBasic(ctx, t)

	op := COp{
		Rem: &OpRem{
			Id: "double",
		},
	}

	if err := op.Do(ctx, s); err != nil {
		t.Fatal(err)
	}

	op = COp{
		Process: &OpProcess{
			// Render:  true,
			Message: Dwimjs(`{"double":2}`),
		},
	}

	if err := op.Do(ctx, s); err != nil {
		t.Fatal(err)
	}

	select {
	case <-time.NewTimer(time.Second).C:
	case x := <-s.Emitted:
		t.Fatal("didn't want %#v", x)
	}

	s.store.Close(ctx) // ToDo: Check error.
}
