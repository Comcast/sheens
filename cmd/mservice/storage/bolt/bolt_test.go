package bolt

import (
	"context"
	"os"
	"testing"

	"github.com/Comcast/sheens/cmd/mservice/storage"
	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
)

func TestImpl(t *testing.T) {
	// Just confirm that this code compiles.
	var _ storage.Storage = &Storage{}

}

func TestBasics(t *testing.T) {
	var (
		filename = "storage.db"
		pid      = "simpsons"
	)

	s, err := NewStorage(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return
		}
		if err := os.Remove(filename); err != nil {
			t.Fatal(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s.Open(ctx); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := s.Close(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	if err := s.MakeCrew(ctx, pid); err != nil {
		t.Fatal(err)
	}

	{

		mss := []*storage.MachineState{
			{
				Mid:        "a",
				SpecSource: crew.NewSpecSource("aspec"),
				NodeName:   "anode",
				Bs: core.Bindings{
					"likes": "tacos",
				},
			},
			{
				Mid:        "b",
				SpecSource: crew.NewSpecSource("bspec"),
				NodeName:   "bnode",
				Bs: core.Bindings{
					"likes": "queso",
				},
			},
		}

		if err := s.WriteState(ctx, pid, mss); err != nil {
			t.Fatal(err)
		}

	}

	check := func(who, what string) {
		got, err := s.GetCrew(ctx, pid)
		if err != nil {
			t.Fatal(err)
		}

		found := false

		for _, m := range got {
			if m.Mid == who {
				found = true
				if likes, have := m.Bs["likes"]; have {
					if likes != what {
						t.Fatalf(`"%s" != "%s"`, likes, what)
					}
				} else {
					t.Fatal("lost likes")
				}
			}
		}

		if what != "" && !found {
			t.Fatalf(`didn't find "%s"`, who)
		}
	}

	check("a", "tacos")
	check("b", "queso")

	{

		mss := []*storage.MachineState{
			{
				Mid:        "a",
				SpecSource: crew.NewSpecSource("aspec1"),
				NodeName:   "anode1",
				Bs: core.Bindings{
					"likes": "chips",
				},
			},
		}

		if err := s.WriteState(ctx, pid, mss); err != nil {
			t.Fatal(err)
		}

	}

	check("a", "chips")
	check("b", "queso")

	{

		mss := []*storage.MachineState{
			{
				Mid:     "a",
				Deleted: true,
			},
		}

		if err := s.WriteState(ctx, pid, mss); err != nil {
			t.Fatal(err)
		}

	}

	check("a", "")
	check("b", "queso")

}

// BenchmarkBolt is just for fun.  Bolt is slow.
func BenchmarkBolt(b *testing.B) {

	var (
		filename = "storage.db"
		pid      = "simpsons"
	)

	s, err := NewStorage(filename)
	if err != nil {
		b.Fatal(err)
	}

	defer func() {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return
		}
		if err := os.Remove(filename); err != nil {
			b.Fatal(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s.Open(ctx); err != nil {
		b.Fatal(err)
	}

	defer func() {
		if err := s.Close(ctx); err != nil {
			b.Fatal(err)
		}
	}()

	if err := s.MakeCrew(ctx, pid); err != nil {
		b.Fatal(err)
	}

	mss := []*storage.MachineState{
		{
			Mid:        "a",
			SpecSource: crew.NewSpecSource("aspec"),
			NodeName:   "anode",
			Bs: core.Bindings{
				"likes": "tacos",
			},
		},
		{
			Mid:        "b",
			SpecSource: crew.NewSpecSource("bspec"),
			NodeName:   "bnode",
			Bs: core.Bindings{
				"likes": "queso",
			},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var err error
		if i%2 == 0 {
			err = s.WriteState(ctx, pid, mss)
		} else {
			_, err = s.GetCrew(ctx, pid)
		}
		if err != nil {
			b.Fatal(err)
		}

	}

}
