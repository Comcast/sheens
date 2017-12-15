package core

import (
	"context"
	"encoding/json"
	"errors"
)

var (
	// TracesInitialCap is the initial capacity for Traces buffers.
	TracesInitialCap = 16
	// ToDo: Provide a configurable limit or implement a rolling buffer.

	// EmittedMessagesInitialCap is the initial capacity for
	// slices of emitted messages.
	EmittedMessagesInitialCap = 16
	// ToDo: Provide a configurable limit.

	// DefaultControl will be used by Spec.Step (and therefore
	// Spec.Walk) if the given control is nil.
	DefaultControl = &Control{
		Limit: 100,
	}

	// Exp_BranchTargetVariables is a switch that enables a branch
	// target to be a reference to a binding.  If a branch target
	// is of the form "@VAR", then the current binding for VAR (if
	// any) is used as the branch target.  If anything goes wrong,
	// the branch target is returned as the literate value of the
	// branch's Target.
	//
	// This feature should be used sparingly if at all.  The
	// motivating use was for a Spec.Boot or a "boot" node/Action,
	// which could be used to state migration when a Spec is
	// updated.  Being able to have a branch target that is passed
	// via bindings would make it much easier to write an Action
	// that can reset a machine based on the machine's previous
	// node.
	Exp_BranchTargetVariables = true
)

type StepProps map[string]interface{}

func (ps StepProps) Copy() StepProps {
	acc := make(StepProps, len(ps))
	for p, v := range ps {
		acc[p] = v
	}
	return acc
}

// StopReason represents the possible reasons for a Walk to terminate.
type StopReason int

//go:generate stringer -type=StopReason
//go:generate jsonenums -type=StopReason

const (
	Done              StopReason = iota // Went as far as the Spec allowed.
	Limited                             // Too many steps.
	InternalError                       // What else to do?
	BreakpointReached                   // During a Walk.
)

// State represents the current state of a machine given a
// specification.
type State struct {
	NodeName string   `json:"node"`
	Bs       Bindings `json:"bs"`
}

func (s *State) String() string {
	if s == nil {
		return "nil"
	}
	js, err := json.Marshal(s.Bs)
	if err != nil {
		return s.NodeName + "/{*}"
	}
	return s.NodeName + "/" + string(js)
}

// Copy makes a deep copy of the State.
func (s *State) Copy() *State {
	return &State{
		NodeName: s.NodeName,
		Bs:       s.Bs.Copy(),
	}
}

// Breakpoint is a *State predicate.
//
// When a Breakpoint returns true for a *State, then processing should
// stop at that point.
type Breakpoint func(context.Context, *State) bool

// Control influences how Walk() operates.
type Control struct {
	// Limit is the maximum number of Steps that a Walk() can take.
	Limit       int
	Breakpoints map[string]Breakpoint
}

func (c *Control) Copy() *Control {
	bs := make(map[string]Breakpoint, len(c.Breakpoints))
	for id, b := range c.Breakpoints {
		bs[id] = b
	}
	return &Control{
		Limit:       c.Limit,
		Breakpoints: bs,
	}
}

// Traces holds trace messages.
type Traces struct {
	Messages []interface{} `json:"messages,omitempty" yaml:",omitempty"`
}

// NewTraces creates an initialized Traces.
//
// The Messages array has TracesSize initial capacity.
func NewTraces() *Traces {
	return &Traces{
		Messages: make([]interface{}, 0, TracesInitialCap),
	}
}

func (ts *Traces) Add(xs ...interface{}) {
	ts.Messages = append(ts.Messages, xs...)
}

// Events contains emitted messages and Traces.
type Events struct {
	Emitted []interface{} `json:"emitted,omitempty" yaml:",omitempty"`
	Traces  *Traces       `json:"traces,omitempty" yaml:",omitempty"`
}

func newEvents() *Events {
	return &Events{
		Emitted: make([]interface{}, 0, EmittedMessagesInitialCap),
		Traces:  NewTraces(),
	}
}

// AddEmitted adds the given thing to the list of emitted messages.
func (es *Events) AddEmitted(x interface{}) {
	es.Emitted = append(es.Emitted, x)
}

// AddTrace adds the given thing to the list of traces.
func (es *Events) AddTrace(x interface{}) {
	es.Traces.Add(x)
}

// AddEvents adds the given Event's emitted messages and traces to the
// receiving Events.
func (es *Events) AddEvents(more *Events) {
	if more == nil {
		return
	}
	for _, x := range more.Emitted {
		es.AddEmitted(x)
	}
	for _, x := range more.Traces.Messages {
		es.AddTrace(x)
	}
}

// Stride represents a step that Walk has taken or attempted.
type Stride struct {
	// Events gather what was emitted during the step.
	*Events `json:"events,omitempty" yaml:",omitempty"`

	// From is the name of the starting node.
	From *State `json:"from,omitempty" yaml:",omitempty"`

	// To is the new State (if any) resulting from the step.
	To *State `json:"to,omitempty" yaml:",omitempty"`

	// Consumed is the message (if any) that was consumed by the step.
	Consumed interface{} `json:"consumed,omitempty" yaml:",omitempty"`
}

func NewStride() *Stride {
	return &Stride{
		Events: newEvents(),
	}
}

// Step is the fundamental operation that attempts to move from the
// given state.
//
// The given pending message (if any) will be consumed by "message"
// type Branches.
func (s *Spec) Step(ctx context.Context, st *State, pending interface{}, c *Control, props StepProps) (*Stride, error) {

	if c == nil {
		c = DefaultControl
	}

	// Each error case should be scrutinized.  It might be
	// possible (and desirable?) to have any error transition to
	// an "error" node, which should have been added during Spec
	// compilation if it didn't already exist.  However, since we
	// really shouldn't modify a Spec outside of Spec.Compile, we
	// can't add an "error" node in this method.
	//
	// Currently there are six places where we return errors.  In theory,

	if !s.compiled {
		return nil, &SpecNotCompiled{s}
	}

	n, have := s.Nodes[st.NodeName]
	if !have {
		// Error (with spec)
		return nil, &UnknownNode{s, st.NodeName}
	}

	if n.Action == nil && n.ActionSource != nil {
		// Error (with spec)
		return nil, &UncompiledAction{s, st.NodeName}
	}

	haveAction := n.Action != nil

	// If we have an action, branch type must be "message".
	//
	// If we insisted that interpreters could not execute code
	// that does IO, then we could remove this limitation.  ToDo:
	// Do that.
	if haveAction && n.Branches != nil && n.Branches.Type == "message" {
		return nil, &BadBranching{s, st.NodeName}
	}

	// If the current node has an action, execute it.
	var (
		err    error
		e      *Execution
		bs     = st.Bs
		stride = NewStride()
	)
	stride.From = st.Copy()

	if haveAction {
		if e, err = n.Action.Exec(ctx, bs, props); err == nil {
			bs = e.Bs
		} else {
			// Bind "actionError" to the error string.  We
			// don't do m.toError since the branches could
			// handle this error themselves.
			bs.Extend("actionError", err.Error())
			bs.Extend("error", err.Error())
			// ToDo: Reconsider reporting this error. Note: it's
			// a user error.
			if !s.ActionErrorBranches {
				// Error (from Action and will not be handled)
				//
				// Don't even give the branches a chance to
				// find (or miss) this error.
				return nil, err
			}
		}
		// Bindings have been updated.

		if e != nil {
			stride.AddEvents(e.Events)
		} // e might be nil if the Exec returned an error.
	}

	// Now evaluate the branches (if any).
	st, ts, consumed, err := n.Branches.consider(ctx, bs, pending, c, props)
	if consumed {
		stride.Consumed = pending
	}

	if ts != nil {
		stride.Traces.Add(ts.Messages...)
	}

	if st != nil {
		stride.To = st.Copy()
	}

	return stride, err
}

// consider considers the Branches to determine the next state.
//
// If this method returns more than one set of Bindings, new machines
// will be created!  ToDo: A switch to warn or prevent.
func (b *Branches) consider(ctx context.Context, bs Bindings, pending interface{}, c *Control, props StepProps) (*State, *Traces, bool, error) {

	// This method will return an error only if its call to try()
	// returns an error.

	ts := NewTraces()

	ts.Add(map[string]interface{}{
		"consider": b,
		"bs":       bs,
		"pending":  pending,
	})

	if b == nil {
		return nil, ts, false, nil
	}

	var (
		against  interface{}
		consumer = b.Type == "message"
	)

	if consumer {
		if pending == nil {
			return nil, ts, consumer, nil
		}
		against = pending
	} else {
		against = map[string]interface{}(bs)
	}

	for _, br := range b.Branches {
		to, more, err := br.try(ctx, bs, against, props)

		ts.Add(more.Messages...)
		if err != nil {
			// Error (forwarded)
			return nil, ts, consumer, err
		}
		if to != nil {
			return to, ts, consumer, nil
		}
	}

	// No branch was traversed.

	return nil, ts, consumer, nil
}

func IsBranchTargetVariable(s string) bool {
	if len(s) == 0 {
		return false
	}
	return s[0] == '@'
}

func (b *Branch) target(bs Bindings) string {
	if 0 < len(bs) && IsBranchTargetVariable(b.Target) {
		if x, have := bs[b.Target[1:]]; have {
			if s, is := x.(string); is {
				return s
			} // else warn?
		}
	}
	return b.Target
}

// try evaluates this Branch to see if it applies.
func (b *Branch) try(ctx context.Context, bs Bindings, against interface{}, props StepProps) (*State, *Traces, error) {

	ts := NewTraces()

	ts.Add(map[string]interface{}{
		"try":     b,
		"bs":      bs,
		"against": against,
	})

	var bss []Bindings

	if b.Pattern != nil {
		var err error
		if bss, err = Match(nil, b.Pattern, against, bs); err != nil {
			ts.Add(map[string]interface{}{
				"error":   err.Error(),
				"pattern": b.Pattern,
			})
			return nil, ts, err
		}
	} else {
		bss = []Bindings{bs}
	}

	ts.Add(map[string]interface{}{
		"bss": bss,
	})

	if b.Guard == nil {
		switch len(bss) {
		case 0:
			// No match
			return nil, ts, nil
		case 1:
			bs = bss[0]
		default:
			return nil, ts, TooManyBindingss
		}
	} else {
		bs = nil
		for _, candidate := range bss {
			ts.Add(map[string]interface{}{
				"guarding": bs,
			})

			exe, err := b.Guard.Exec(ctx, candidate, props)

			if exe != nil {
				ts.Add(exe.Events.Traces.Messages...)
			}

			if err != nil {
				ts.Add(map[string]interface{}{
					"error": err.Error(),
				})

				return nil, ts, err
			}

			if exe.Bs != nil {
				ts.Add(map[string]interface{}{
					"guarded": exe.Bs,
				})
				bs = exe.Bs
				break
			}
		}
	}

	ts.Add(map[string]interface{}{
		"bs":     bs,
		"target": b.target(bs),
	})

	if bs == nil {
		return nil, ts, nil
	}

	st := &State{
		NodeName: b.target(bs),
		Bs:       bs,
	}

	return st, ts, nil
}

// Walked represents a sequence of strides taken by a Walk().
type Walked struct {
	// Strides contains each Stride taken and the last one
	// attempted.
	Strides []*Stride `json:"strides" yaml:",omitempty"`

	// Remaining stores the messages that Walk failed to consume.
	Remaining []interface{} `json:"remaining,omitempty" yaml:",omitempty"`

	// StoppedBecause reports the reason why the Walk stopped.
	StoppedBecause StopReason `json:"stoppedBecause,omitempty" yaml:",omitempty"`

	// Error stores an internal error that occured (if any).
	Error error `json:"error,omitempty" yaml:",omitempty"`

	// BreakpointId is the id of the breakpoint, if any, that
	// caused this Walk to stop.
	BreakpointId string `json:"breakpoint,omitempty" yaml:",omitempty"`
}

func (w *Walked) From() *State {
	if 0 == len(w.Strides) {
		return nil
	}
	return w.Strides[0].From.Copy()
}

func (w *Walked) To() *State {
	for i := len(w.Strides) - 1; 0 <= i; i-- {
		if s := w.Strides[i]; s.To != nil {
			return s.To.Copy()
		}
	}
	return nil
}

// DoEmitted is a convenience method to iterate over messages emitted
// by the Walked.
func (w Walked) DoEmitted(f func(x interface{}) error) error {
	for _, stride := range w.Strides {
		for _, x := range stride.Emitted {
			if err := f(x); err != nil {
				return err
			}
		}
	}
	return nil
}

func newWalked(siz int) *Walked {
	max := 1024
	if max < siz {
		siz = max
	}
	return &Walked{
		Strides: make([]*Stride, 0, siz),
	}
}

func (w *Walked) add(s *Stride) {
	w.Strides = append(w.Strides, s)
}

// Walk takes as many steps as it can.
//
// Any returned error is an internal error.  Almost all errors
// encountered during processing should transition to the "error" node
// with a binding for "error".
//
// Any unprocessed messages are returned. This method should only
// returned some unprocessed messages if the method encountered an
// internal error.
func (s *Spec) Walk(ctx context.Context, st *State, pendings []interface{}, c *Control, props StepProps) (*Walked, error) {

	// This method should (probably) neven return an error.  When
	// an error of some sort occurs during the processing, the
	// Walked.To state is returned with a NodeName of "error" and
	// Bindings that include a binding for "error".  In other
	// words, errors are normal.  (ToDo: Develop a dictionary for
	// errors.)
	//
	// If a Spec doesn't have an "error" node, then that problem
	// will surface later. Hopefully Spec.Compile will (1) have
	// been called and (2) either verified the existence of an
	// "error" node or addeded one.

	walked := newWalked(c.Limit)

	for i := 0; i < c.Limit; i++ {
		for id, breakpoint := range c.Breakpoints {
			if breakpoint(ctx, st) {
				walked.StoppedBecause = BreakpointReached
				walked.BreakpointId = id
				walked.Remaining = pendings
				return walked, nil
			}
		}

		var pending interface{}
		if 0 < len(pendings) {
			pending = pendings[0]
		}
		stride, err := s.Step(ctx, st, pending, c, props)

		if stride == nil {
			// We hope we never get here. ToDo: Warn?
			if err == nil {
				err = errors.New("nil stride")
			}
			stride = NewStride()
			stride.From = st.Copy()
			// Since the returned stride is nil,
			// we conclude that no message was
			// consumed.
		}

		// Currently, a spec can have a branch with a guard
		// even when branching type is "message"!

		if err != nil {
			if st.NodeName == "error" {
				// We're already at an error.
			} else {
				errorBs, _ := st.Bs.Extendm("error", err.Error(),
					"lastNode", st.NodeName,
					"lastBindings", st.Bs.Copy())
				stride.To = &State{
					NodeName: "error",
					Bs:       errorBs,
				}
			}
		}

		walked.add(stride)

		if stride.Consumed != nil {
			// We consumed a message, so get the next
			// message ready.
			pendings = pendings[1:]
		}

		if stride.To == nil {
			// We went nowhere.
			if 0 == len(pendings) {
				// We have no messages to offer, so we're done.
				walked.StoppedBecause = Done
				walked.Remaining = nil
				return walked, nil
			}
			// Leave st unchanged.
		} else {
			st = stride.To.Copy()
		}
	}

	// We hit the c.Limit. That's a problem.
	walked.StoppedBecause = Limited
	walked.Remaining = pendings

	return walked, nil
}
