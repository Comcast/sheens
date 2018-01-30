package tools

import (
	"bytes"
	"testing"
)

func TestRenderSpecHTML(t *testing.T) {

	t.Run("withoutGraph", func(t *testing.T) {
		out := bytes.NewBuffer(make([]byte, 0, 1024*128))

		err := ReadAndRenderSpecPage("../specs/double.yaml", []string{"spec.css"}, out, false)

		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("withGraph", func(t *testing.T) {
		out := bytes.NewBuffer(make([]byte, 0, 1024*128))

		err := ReadAndRenderSpecPage("../specs/double.yaml", []string{"spec.css"}, out, true)

		if err != nil {
			t.Fatal(err)
		}
	})

}
