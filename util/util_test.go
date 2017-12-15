package util

import "testing"

func TestLogf(t *testing.T) {
	Logf("I'd like a couple of %s, please.", "tacos")
}
