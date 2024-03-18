// Yo, check this out! We're setting up the building blocks for our cool tool.
package tools

import (
	"context"       // Gotta have this for managing goroutine lifetimes.
	"encoding/json" // JSON is everywhere, so we'll need this to read and write it.
	"fmt"           // Basic input/output - can't live without it.
	"io"            // Working with input/output streams. Think of it like a data river!
	"os"            // New import for reading files in Go 1.21.

	"github.com/Comcast/sheens/core"              // This is where the magic happens in Sheens.
	"github.com/Comcast/sheens/interpreters/noop" // A noop interpreter for when we want to do... nothing?
	. "github.com/Comcast/sheens/util/testutil"   // Test utils, but using the dot import for direct access.
	"gopkg.in/yaml.v2"                            // YAML is JSON's cool cousin. We're using it for configuration.

	md "github.com/russross/blackfriday/v2" // Markdown processor because text is boring without formatting.
)

// RenderSpecHTML takes a Sheens Spec and turns it into HTML. Think of it as making a plain text look fancy on the web.
func RenderSpecHTML(s *core.Spec, out io.Writer) error {
	f := func(format string, args ...interface{}) {
		fmt.Fprintf(out, format+"\n", args...) // Writing formatted output to our writer, adding a newline for readability.
	}

	// Using Blackfriday to turn Markdown into HTML. It's like magic for text!
	f(`<div class="specDoc doc">%s</div>`, md.Run([]byte(s.Doc)))

	{ // Nodes section. Here we're gonna layout all the nodes in a neat table.
		f(`<div class="nodes"><table>`)

		// This inner function is a neat way to process each node. Encapsulation, baby!
		fn := func(id string, node *core.Node) {
			// For each node, we're creating a row in our HTML table.
			f(`<tr class="node"><td><span id="%s" class="nodeName">%s</span></td><td>`, id, id)

			// If a node has documentation, let's include that too. More info is always better!
			if node.Doc != "" {
				f(`<div class="nodeDoc doc">%s</div>`, md.Run([]byte(node.Doc)))
			}
			// If there's action source code, we'll show that in a <pre> tag for formatting.
			if node.ActionSource != nil {
				f(`<div class="code"><pre>%s</pre></div>`, node.ActionSource.Source)
			}
			// Branches are tricky. They decide where to go next based on messages. It's like a choose-your-own-adventure book!
			if node.Branches != nil {
				if node.Branches.Type == "message" {
					f(`<div>type: <span class="branchingType">message</span><div>`)
				}
				f(`<div class="branches">`)
				f(`<table>`)
				for i, b := range node.Branches.Branches {
					// Each branch gets its own row. It's like laying out options on a table.
					f(`<tr><td><div class="branchNum">%d</div></td><td>`, i)
					f(`<table>`)

					// Showing the pattern and guard of each branch. It's like setting rules for the adventure paths!
					if b.Pattern != nil {
						f(`<tr><td></td><td>pattern</td>`)
						f(`<td><code>%s</code></td></tr>`, JS(b.Pattern))
					}
					if b.GuardSource != nil {
						f(`<tr><td></td><td>guard</td>`)
						f(`<td><div class="code"><pre>%s</pre></div></td></tr>`, b.GuardSource.Source)
					}
					// If the branch has a target, we make it a clickable link. Fancy navigation!
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
		// Making sure we always show the start node first. It's like starting at the beginning of the book.
		if n, has := s.Nodes["start"]; has {
			fn("start", n)
		}
		// Looping through all nodes except "start" since we already handled it.
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

// RenderSpecPage takes a spec and makes a whole web page out of it. This is where things get real.
func RenderSpecPage(s *core.Spec, out io.Writer, cssFiles []string, includeGraph bool) error {

	if cssFiles == nil {
		cssFiles = []string{"/static/spec-html.css"} // Default styling if none provided. Gotta make it look good!
	}

	js, err := json.Marshal(s) // Turning our spec into JSON because JavaScript can understand it.
	if err != nil {
		return err
	}

	// Basic HTML structure. This is like the skeleton of our web page.
	fmt.Fprintf(out, `<!DOCTYPE html>
<meta charset="utf-8">
<html>
  <head>
  <title>%s</title>
`, s.Name)

	// If we want to include a graph, we load up some extra JavaScript libraries for drawing.
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

	// Linking to CSS files for styling. Everyone wants their page to be the prettiest, right?
	for _, cssFile := range cssFiles {
		fmt.Fprintf(out, "  <link href=\"%s\" rel=\"stylesheet\">\n", cssFile)
	}

	fmt.Fprintf(out, `
  </head>
  <body>
    <h1>%s</h1>
`, s.Name)

	// If we decided to include a graph, here's where it will show up. Like a map of our adventure!
	if includeGraph {
		fmt.Fprintf(out, `<div id="graph"></div>`)
	}

	// Here's where we render the spec into HTML. It's like filling the skeleton with muscles and skin.
	if err = RenderSpecHTML(s, out); err != nil {
		return err
	}

	fmt.Fprintf(out, `
  </body>
</html>
`)

	return nil
}

// ReadAndRenderSpecPage reads a spec from a file, does some processing, and turns it into a web page.
func ReadAndRenderSpecPage(filename string, cssFiles []string, out io.Writer, includeGraph bool) error {
	// Reading the spec from a file. It's like opening a treasure chest!
	specSrc, err := os.ReadFile(filename) // Updated to use os.ReadFile which is the recommended way since Go 1.16.
	if err != nil {
		return err
	}
	var spec core.Spec
	// Unmarshalling YAML. It's like translating ancient scrolls into modern language.
	if err = yaml.Unmarshal(specSrc, &spec); err != nil {
		return err
	}

	// Setting up a noop interpreter. Sometimes, doing nothing is an important step.
	interpreters := noop.NewInterpreters()
	interpreters.I.Silent = true

	// Contexts are great for managing go routines, like directing traffic in your code.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Always clean up after yourself. Cancel the context when you're done.

	// Compiling the spec with our context and interpreter. It's like putting puzzle pieces together.
	if err = spec.Compile(ctx, interpreters, true); err != nil {
		return err
	}

	// Finally, we render our spec into a beautiful web page. Showtime!
	return RenderSpecPage(&spec, out, cssFiles, includeGraph)
}
