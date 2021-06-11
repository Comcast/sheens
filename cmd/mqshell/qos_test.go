package main

import (
	"testing"
)

func TestTestMsg(t *testing.T) {
	if _, err := NewTestMsg(3, 42); err != nil {
		t.Fatal(err)
	}
}
