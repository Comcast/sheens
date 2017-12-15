package tools

import (
	"context"
	"os"
	"testing"

	. "github.com/Comcast/sheens/core"
)

func TestDot(t *testing.T) {
	filename := "g.dot"

	out, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Remove(filename); err != nil {
			t.Fatal(err)
		}
	}()

	spec, err := TurnstileSpec(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if err := Dot(spec, out, "", ""); err != nil {
		t.Fatal(err)
	}

}
