package tools

import (
	"log"
	"strings"
	"testing"
)

func TestInline(t *testing.T) {
	input := `
I like %inline("tacos"), and
I also like %inline("queso").
Both are delicious.
`
	want := `
I like TACOS, and
I also like QUESO.
Both are delicious.
`

	find := func(name string) ([]byte, error) {
		return []byte(strings.ToUpper(name)), nil
	}

	got, err := Inline([]byte(input), find)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		log.Fatalf("got %s", got)
	}
}
