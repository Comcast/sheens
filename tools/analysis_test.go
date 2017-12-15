package tools

import (
	"context"
	"testing"

	"github.com/Comcast/sheens/core"
)

func TestAnalysis(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spec, err := core.TurnstileSpec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Analyze(spec); err != nil {
		t.Fatal(err)
	}
}
