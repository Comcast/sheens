package main

type WrappedError struct {
	Outer error `json:"outer"`
	Inner error `json:"inner"`
}

func (e *WrappedError) Error() string {
	return e.Outer.Error() + " after " + e.Inner.Error()
}

func NewWrappedError(outer, inner error) error {
	if inner == nil {
		return outer
	}
	return &WrappedError{
		Outer: outer,
		Inner: inner,
	}
}
