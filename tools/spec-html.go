package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/interpreters/noop"
	. "github.com/Comcast/sheens/util/testutil"
	"github.com/jsccast/yaml"

	md "gopkg.in/russross/blackfriday.v2"
)

func RenderSpecHTML(s *core.Spec, out io.Writer) error {
	f := func(format string, args ...interface{}) {
		fmt.Fprintf(out, format+"\n", args...)
	}

	f(`<div class="specDoc doc">%s</div>`, md.Run([]byte(s.Doc)))

	{ // Nodes
		f(`<div class="nodes"><table>`)
		// Need to try to order these nodes sensibly.
		fn := func(id string, node *core.Node) {
			f(`<tr class="node"><td><span id="%s" class="nodeName">%s</span></td><td>`, id, id)

			if node.Doc != "" {
				f(`<div class="nodeDoc doc">%s</div>`, md.Run([]byte(node.Doc)))
			}
			if node.ActionSource != nil {
				f(`<div class="code"><pre>%s</pre></div>`, node.ActionSource.Source)
			}
			if node.Branches != nil {
				if node.Branches.Type == "message" {
					f(`<div>type: <span class="branchingType">message</span><div>`)
				}
				f(`<div class="branches">`)
				f(`<table>`)
				for i, b := range node.Branches.Branches {
					f(`<tr><td><div class="branchNum">%d</div></td><td>`, i)
					f(`<table>`)
					// if b.Doc != "" {
					// 	f(`<tr><td></td><td>pattern</td>`)
					// 	f(`<td><div class="branchDoc doc">%s</div></td></tr>`, JS(b.Pattern))
					// }
					if b.Pattern != nil {
						f(`<tr><td></td><td>pattern</td>`)
						f(`<td><code>%s</code></td></tr>`, JS(b.Pattern))
					}
					if b.GuardSource != nil {
						f(`<tr><td></td><td>guard</td>`)
						f(`<td><div class="code"><pre>%s</pre></div></td></tr>`, b.GuardSource.Source)
					}
					if b.Target != "" {
						f(`<tr><td></td><td>target</td>`)
						f(`<td><a href="#%s"><code>%s</code></a></td></tr>`, b.Target, b.Target)
					}
					f(`</table>`)
					f(`</td</tr>`)
				}
				f(`</table>`)
				f(`</div>`)
			}
			f(`</td></tr>`)
		}
		if n, has := s.Nodes["start"]; has {
			fn("start", n)
		}
		// ToDo: Order.
		for id, node := range s.Nodes {
			if id == "start" {
				continue
			}
			fn(id, node)
		}
		f(`</div></table>`)
	}

	return nil
}

func RenderSpecPage(s *core.Spec, out io.Writer, cssFiles []string, includeGraph bool) error {

	if cssFiles == nil {
		cssFiles = []string{"/static/spec-html.css"}
	}

	js, err := json.Marshal(s)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, `<!DOCTYPE html>
<meta charset="utf-8">
<html>
  <head>
  <title>%s</title>
`, s.Name)

	if includeGraph {
		fmt.Fprintf(out, `
  <script src="https://cdnjs.cloudflare.com/ajax/libs/d3/4.12.2/d3.min.js"></script>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/cytoscape/3.2.8/cytoscape.min.js"></script>
  <script src="/static/spec-html.js"></script>
  <script>
  var thisSpec = %s;
  </script>
`, js)
	}

	for _, cssFile := range cssFiles {
		fmt.Fprintf(out, "  <link href=\"%s\" rel=\"stylesheet\">\n", cssFile)
	}

	fmt.Fprintf(out, `
  </head>
  <body>
    <h1>%s</h1>
`, s.Name)

	if includeGraph {
		fmt.Fprintf(out, `<div id="graph"></div>`)
	}

	if err = RenderSpecHTML(s, out); err != nil {
		return err
	}

	fmt.Fprintf(out, `
  </body>
</html>
`)

	return nil
}

func ReadAndRenderSpecPage(filename string, cssFiles []string, out io.Writer, includeGraph bool) error {
	specSrc, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var spec core.Spec
	if err = yaml.Unmarshal(specSrc, &spec); err != nil {
		return err
	}

	interpreters := noop.NewInterpreters()
	interpreters.I.Silent = true

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err = spec.Compile(ctx, interpreters, true); err != nil {
		return err
	}

	return RenderSpecPage(&spec, out, cssFiles, includeGraph)

}
