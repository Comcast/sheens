package main

import (
	"context"
	"testing"
	"time"

	"github.com/Comcast/sheens/crew"
)

func TestProtocol(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	routed := make(chan interface{}, 1024)
	s, err := makeDemoService(ctx, routed, "", "", "")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := s.Storage.Close(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	cid := "simpsons"

	s.RemCrew(ctx, cid)

	sop := &SOp{
		Make: cid,
	}

	if err = sop.Do(ctx, s); err != nil {
		t.Fatal(err)
	}

	cop := &COp{
		Cid: cid,
		Add: &OpAdd{
			Machine: &crew.Machine{
				Id: "doubler",
				SpecSource: &crew.SpecSource{
					Name: "double",
				},
			},
		},
	}

	if err = cop.Do(ctx, s); err != nil {
		t.Fatal(err)
	}

	cop = &COp{
		Cid: cid,
		Add: &OpAdd{
			Machine: &crew.Machine{
				Id: "timers",
				SpecSource: &crew.SpecSource{
					Name: "timers",
				},
			},
		},
	}

	if err = cop.Do(ctx, s); err != nil {
		t.Fatal(err)
	}

	cop = &COp{
		Cid: cid,
		Process: &OpProcess{
			Message: Dwimjs(`{"double":1000}`),
		},
	}

	if err = cop.Do(ctx, s); err != nil {
		t.Fatal(err)
	}

	Render("test", cop.Process.Walked)

	cop = &COp{
		Cid: cid,
		Process: &OpProcess{
			Message: Dwimjs(`{"makeTimer":{"in":"1s","id":"chips","message":{"double":3000}}}`),
		},
	}

	if err = cop.Do(ctx, s); err != nil {
		t.Fatal(err)
	}

	Render("test", cop.Process.Walked)

	time.Sleep(2 * time.Second)

	cop = &COp{
		Cid: cid,
		Rem: &OpRem{
			Id: "doubler",
		},
	}

	if err = cop.Do(ctx, s); err != nil {
		t.Fatal(err)
	}

	cop = &COp{
		Cid: cid,
		Process: &OpProcess{
			Message: Dwimjs(`{"double":7000}`),
		},
	}

	if err = cop.Do(ctx, s); err != nil {
		t.Fatal(err)
	}

	sop = &SOp{
		Rem: cid,
	}

	if err = sop.Do(ctx, s); err != nil {
		t.Fatal(err)
	}

	drain(routed)
}
