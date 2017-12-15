package main

import (
	"context"
	"testing"

	"github.com/Comcast/sheens/core"
)

func TestSetSpecId(t *testing.T) {
	ctx := context.Background()

	s, err := core.TurnstileSpec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	id, err := SetSpecId(s)
	if err != nil {
		t.Fatal(err)
	}

	if len(id) < 16 {
		t.Fatalf(`id "%s" too short`, id)
	}

	s.Version = s.Version + " with queso"

	id2, err := SetSpecId(s)
	if err != nil {
		t.Fatal(err)
	}

	if len(id2) < 16 {
		t.Fatalf(`id "%s" too short`, id)
	}

	if id == id2 {
		t.Fatalf(`second id "%s" should be different`, id2)
	}
}
