package storage

import (
	"context"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
)

// MachineState is a presentation of a machine's state as stored in a
// Storage system.
type MachineState struct {
	// Mid is the id for the machine.
	Mid string `json:"id,omitempty"`

	SpecSource *crew.SpecSource `json:"spec,omitempty" yaml:"spec,omitempty"`
	NodeName   string           `json:"node"`
	Bs         core.Bindings    `json:"bs"`

	// Deleted indicated that this machine has been deleted.
	//
	// Yes, this flag is a hack.
	Deleted bool `json:"-" yaml:"-"`
}

// Storage is a persistence interface that's suitable for Crews.
type Storage interface {
	MakeCrew(ctx context.Context, pid string) error

	RemCrew(ctx context.Context, pid string) error

	GetCrew(ctx context.Context, pid string) ([]*MachineState, error)

	WriteState(ctx context.Context, pid string, ss []*MachineState) error

	// Iterate over Crews?
}

func AsMachinesStates(changes map[string]*core.State) []*MachineState {
	acc := make([]*MachineState, 0, len(changes))
	for mid, s := range changes {
		ms := &MachineState{
			Mid:      mid,
			NodeName: s.NodeName,
			Bs:       s.Bs,
		}
		acc = append(acc, ms)
	}
	return acc
}

func AsMachines(mss []*MachineState) map[string]*crew.Machine {
	acc := make(map[string]*crew.Machine, len(mss))
	for _, ms := range mss {
		m := &crew.Machine{
			Id: ms.Mid,
			State: &core.State{
				NodeName: ms.NodeName,
				Bs:       ms.Bs,
			},
			SpecSource: ms.SpecSource,
		}
		acc[ms.Mid] = m
	}
	return acc
}
