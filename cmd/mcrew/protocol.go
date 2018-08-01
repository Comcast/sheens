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
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/match"
	. "github.com/Comcast/sheens/util/testutil"
)

// SOp is a Service Operation.
//
// Only one of GetSpec, GetCrewOp, or COp should have value.
type SOp struct {
	// GetSpec is a utility that invokes the service's SpecProvider.
	GetSpec *GetSpecOp `json:"getSpec,omitempty" yaml:",omitempty"`

	// ToDo: UpdateSpec.

	// GetCrewOp that gets (a copy of) a Crew.
	GetCrewOp *GetCrewOp `json:"getCrew,omitempty" yaml:",omitempty"`

	// ToDo: UpdateCrewOp -- or something like that.

	// Error will hold an error (if any) that results from
	// processing this operation.
	Error error `json:"-" yaml:"-"`

	// Err will hold a string representation of an error (if any)
	// that results from processing this operation.
	Err string `json:"err,omitempty" yaml:",omitempty"`

	// COp gives a Crew operation.
	COp *COp `json:"cop,omitempty" yaml:"cop,omitempty"`
}

// erred is a utility function to return values to assign to operation
// Error and Err fields.
func erred(err error) (error, string) {
	if err == nil {
		return nil, ""
	}
	return err, err.Error()
}

func (o *SOp) Do(ctx context.Context, s *Service) error {

	s.op(ctx, map[string]interface{}{
		"do": o,
	})

	var err error
	if o.GetSpec != nil {
		err = o.GetSpec.Do(ctx, s)
	} else if o.GetCrewOp != nil {
		err = o.GetCrewOp.Do(ctx, s)
	} else if o.COp != nil {
		err = o.COp.Do(ctx, s)
	} else {
		err = fmt.Errorf("not implemented: %s", JS(o))
	}

	if err != nil && o.Error == nil {
		o.Error, o.Err = erred(err)
	}

	s.op(ctx, map[string]interface{}{
		"did": o,
	})

	return o.Error
}

type GetSpecOp struct {
	Source *crew.SpecSource `json:"source,omitempty" yaml:",omitempty"`
	Spec   *core.Spec       `json:"spec,omitempty" yaml:",spec"`
}

func (o *GetSpecOp) Do(ctx context.Context, s *Service) error {
	var spec core.Specter
	spec, err := s.GetSpec(ctx, o.Source)
	if err == nil {
		o.Spec = spec.Spec()
	}
	return err
}

type GetCrewOp struct {
	Crew *crew.Crew `json:"crew,omitempty" yaml:",omitempty"`
}

func (o *GetCrewOp) Do(ctx context.Context, s *Service) error {
	o.Crew = s.crew.Copy()
	return nil
}

// COp is a Crew Operation.
//
// In normal use, only one field should be given.
type COp struct {
	// Add a machine to the Crew.
	Add *OpAdd `json:"add,omitempty" yaml:",omitempty"`

	// Rem removes a machine from the Crew.
	Rem *OpRem `json:"rem,omitempty" yaml:",omitempty"`

	// Process sends messages to the Crew.
	Process *OpProcess `json:"process,omitempty" yaml:",omitempty"`

	Exercise *OpExercise `json:"exercise,omitempty" yaml:",omitempty"`
}

func (o *COp) Do(ctx context.Context, s *Service) error {
	if o.Add != nil {
		return o.Add.Do(ctx, s)
	}
	if o.Rem != nil {
		return o.Rem.Do(ctx, s)
	}
	if o.Process != nil {
		return o.Process.Do(ctx, s)
	}
	if o.Exercise != nil {
		return o.Exercise.Do(ctx)
	}
	panic("not implemented")
}

type OpAdd struct {
	// Oid is the optional operation id.  A "transaction" id.
	Oid string `json:"oid,omitempty" yaml:",omitempty"`

	// Machine represents the Machine to create and add.
	Machine *crew.Machine `json:"m"`

	// Error will hold an error (if any) that results from
	// processing this operation.
	Error error `json:"-" yaml:"-"`

	// Err will hold a string representation of an error (if any)
	// that results from processing this operation.
	Err string `json:"err,omitempty" yaml:",omitempty"`
}

func (o *OpAdd) Do(ctx context.Context, s *Service) error {
	if o.Machine == nil {
		return fmt.Errorf("no machine given")
	}
	if o.Machine.State == nil {
		o.Machine.State = &core.State{
			NodeName: "start",
			Bs:       match.NewBindings(),
		}
	}
	// get spec and set default values if they are not provided by
	// initial bindings
	specter, err := s.GetSpec(ctx, o.Machine.SpecSource)
	if err != nil {
		return err
	}
	spec := specter.Spec()
	for key, param := range spec.ParamSpecs {
		if param.Default != nil {
			_, ok := o.Machine.State.Bs[key]
			if !ok {
				o.Machine.State.Bs[key] = param.Default
			}
		}
	}
	//
	o.Error, o.Err = erred(s.AddMachine(ctx,
		o.Machine.SpecSource.Name,
		o.Machine.Id,
		o.Machine.State.NodeName,
		o.Machine.State.Bs))

	return nil
}

type OpRem struct {
	// Oid is the optional operation id.  A "transaction" id.
	Oid string `json:"oid,omitempty" yaml:",omitempty"`

	// Id is the id of the Machine to remove.
	Id string `json:"id"`

	// Error will hold an error (if any) that results from
	// processing this operation.
	Error error `json:"-" yaml:"-"`

	// Err will hold a string representation of an error (if any)
	// that results from processing this operation.
	Err string `json:"err,omitempty" yaml:",omitempty"`
}

func (o *OpRem) Do(ctx context.Context, s *Service) error {
	o.Error, o.Err = erred(s.RemMachine(ctx, o.Id))
	return nil
}

type OpProcess struct {
	// Oid is the optional operation id.  A "transaction" id.
	Oid string `json:"oid,omitempty" yaml:",omitempty"`

	// Ctl specifies how the processing behaves.
	Ctl *core.Control `json:"ctl,omitempty" yaml:",omitempty"`

	// Message is the message to process.
	Message interface{} `json:"message,omitempty" yaml:",omitempty"`

	Walked map[string]*core.Walked `json:"walked,omitempty" yaml:",omitempty"`

	Render bool `json:"render,omitempty" yaml:",omitempty"`

	// Error will hold an error (if any) that results from
	// processing this operation.
	Error error `json:"-" yaml:"-"`

	// Err will hold a string representation of an error (if any)
	// that results from processing this operation.
	Err string `json:"err,omitempty" yaml:",omitempty"`
}

func (o *OpProcess) Do(ctx context.Context, s *Service) error {
	var err error
	if o.Ctl == nil {
		o.Ctl = core.DefaultControl
	}
	o.Walked, err = s.Process(ctx, o.Message, o.Ctl)
	o.Error, o.Err = erred(err)

	if o.Render && o.Walked != nil {
		Render(os.Stderr, "op", o.Walked)
	}
	return err
}

type OpExercise struct {
	Count      int    `json:"count,omitempty" yaml:",omitempty"`
	Port       string `json:"port,omitempty" yaml:",omitempty"`
	Error      error  `json:"-" yaml:"-"`
	Err        string `json:"err,omitempty" yaml:",omitempty"`
	Background bool   `json:"background,omitempty" yaml:",omitempty"`
}

func (o *OpExercise) Do(ctx context.Context) error {
	addr := o.Port
	port, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		o.Error, o.Err = erred(err)
		return err
	}

	c, err := net.DialTCP("tcp", nil, port)
	if err != nil {
		o.Error, o.Err = erred(err)
		return err
	}

	f := func(n int) {
		in := bufio.NewReader(c)
		out := bufio.NewWriter(c)

		for i := 0; i < n; i++ {
			msg := fmt.Sprintf(`{"cop":{"process":{"message":{"to":"doubler","double":%d}}}}`+"\n", i)
			if _, err := out.Write([]byte(msg)); err != nil {
				log.Printf("OpExercise Writer error %v", err)
				break
			}
			if err = out.Flush(); err != nil {
				log.Printf("OpExercise Writer flush error %v", err)
				break
			}
			_, err := in.ReadBytes('\n')
			if err != nil {
				log.Printf("OpExercise read error %v at %d", err, i)
				break
			}
		}

		log.Printf("OpExercise wrote, read %d", n)
		c.Close()
	}

	if o.Background {
		log.Printf("OpExercise %d background", o.Count)
		go f(o.Count)
	} else {
		f(o.Count)
	}

	o.Error, o.Err = erred(err)
	return err
}
