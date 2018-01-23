package main

import (
	"context"
	"log"
	"os"

	"testing"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	. "github.com/Comcast/sheens/util/testutil"
)

func TestService(t *testing.T) {

	dbFile := "test.db"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
				log.Printf("processing %s", JS(m))
			}
		}
	}()

	m := <-s.Emitted
	log.Printf("emitted %s", JS(m))

	defer s.store.Close(ctx) // ToDo: Check error.
}
