package core

import (
	"context"
	"testing"
)

func TestPermanentBindings(t *testing.T) {
	if !Exp_PermanentBindings {
		return
	}

	ctx := context.Background()

	action := &FuncAction{
		F: func(ctx context.Context, bs Bindings, params Params) (*Execution, error) {
			return &Execution{
				Bs:     bs.Remove("ephemeral", "permament!"),
				Events: newEvents(),
			}, nil
		},
	}

	bs := NewBindings()
	bs["ephemeral"] = "queso"
	bs["permament!"] = "tacos"
	exe, err := action.Exec(ctx, bs.Copy(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, have := exe.Bs["ephemeral"]; have {
		t.Fatal("ephemeral wasn't")
	}
	if _, have := exe.Bs["permament!"]; !have {
		t.Fatal("permament wasn't")
	}
}
