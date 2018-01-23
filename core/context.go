package core

import "time"

// Context is a copy of context.Context.
//
// This type is defined here in order to avoid importing
// context.Context, which pulls with it a ton of other stuff
// (e.g. fmt).  See https://github.com/Comcast/sheens/issues/13 and
// 14.
type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}
