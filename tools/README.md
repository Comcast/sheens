# Some machine tools

## "Expect"

This tool facilitates Machine Spec testing.

See the godoc for `Session`, and see [../cmd/mexpect](`cmd/mexpect`).

## Generate a Graphviz dot file for a spec

See `dot.go`.

## Analyze a spec

See `analyze.go` and `analyze_test.go`.

## Render a spec as HTML

See `spec-html*.*`.

### Usage

Copy `spec-html.js` and `spec-html.css` to `../cmd/mcrew/httpfies`.
Then run `mcrew`:

```Shell
mcrew -h :9050 -f cmd/mcrew/httpfiles/
```

Then try [`http://localhost:9050/specs/double.html`](http://localhost:9050/specs/double.html).

