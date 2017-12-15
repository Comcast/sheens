package tools

// dot -Tpng g.dot > g.png

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	. "github.com/Comcast/sheens/core"

	"gopkg.in/yaml.v2"
)

// Dot makes a Graphviz dot file for the given machine.  A really ugly
// dot file.
//
// The optional fromNode and toNode can be names of nodes during a
// transition.  If non-zero, then the fromNode will be black and the
// toNode will be red.  Maybe.
func Dot(spec *Spec, w io.WriteCloser, fromNode, toNode string) error {

	yamlPatterns := true

	// Use copies of states that don't have Name set.
	nodes := make(map[string]*Node, len(spec.Nodes))
	for name, n := range spec.Nodes {
		nodes[name] = n
	}

	log.Printf("processing %d nodes", len(nodes))

	fmt.Fprintf(w, "digraph G {\n")
	// nodesep=0.3,ranksep=0.3,
	fmt.Fprintf(w, `  graph [ordering=out,rankdir=TB,nodesep=0.3,ranksep=0.6]
  node [shape="record" style="rounded,filled"]
  edge [fontsize = "12"]
`)

	seen := make(map[string]bool)
	node := func(name string, n *Node) error {
		if n == nil {
			return fmt.Errorf("Unknown node '%s'", name)
		}

		if _, already := seen[name]; already {
			return nil
		}
		seen[name] = true
		label := name
		if n.Doc != "" {
			doc := n.Doc
			if 40 < len(doc) {
				period := strings.Index(doc, ". ")
				if 0 < period {
					doc = doc[0 : period+1]
				}
			}
			label += "<BR/><FONT POINT-SIZE='8'>" + doc + "</FONT>"
		}
		fillcolor := "#99ddc8"
		if n.Branches != nil {
			switch n.Branches.Type {
			case "message":
				fillcolor = "#2d93ad"
			case "":
				fillcolor = "#52aa5e"
			}
		}
		color := "black"
		shape := "record"
		style := "filled"
		if n.Action != nil || n.ActionSource != nil {
			shape = "note"
			var src string
			x := n.ActionSource.Source
			if s, is := x.(string); is {
				src = s
			} else {
				src = fmt.Sprintf("%#v", x)
			}
			src = strings.Replace(src, "<", `&lt;`, -1)
			src = strings.Replace(src, ">", `&gt;`, -1)
			label += `<FONT POINT-SIZE="6">` +
				`<BR/>` + strings.Replace(string(src)+"\n", "\n", `<BR ALIGN="LEFT"/>`, -1) + `<BR/>` +
				`</FONT>`
		}
		if toNode == name {
			color = "red"
			fillcolor = "#f98b8b"
		}
		if name == "start" {
			style += ",bold"
		}
		if n.Branches == nil || len(n.Branches.Branches) == 0 {
			style += ",dashed"
		}
		fmt.Fprintf(w, "  %s [shape=\"%s\", style=\"%s\", color=\"%s\", fillcolor=\"%s\", label=<%s> ]\n",
			name, shape, style, color, fillcolor, label)

		return nil
	}

	process := func(name string, n *Node) error {
		if err := node(name, n); err != nil {
			log.Printf("process error with %s: %v", name, err)
			return err
		}
		if n.Branches == nil {
			return nil
		}
		log.Printf("  processing %s branches: %d", name, len(n.Branches.Branches))
		for i, b := range n.Branches.Branches {
			if err := node(b.Target, nodes[b.Target]); err != nil {
				log.Printf("process branch error with %s: %v", b.Target, err)
				return err
			}
			var label = "{}"

			if b.Pattern != nil {
				var js []byte
				var err error
				if yamlPatterns {
					js, err = yaml.Marshal(b.Pattern)
				} else {
					js, err = json.MarshalIndent(b.Pattern, " ", " ")
				}
				if err != nil {
					js = []byte(err.Error())
				} else {
				}
				label = string(js)
				label = strings.Replace(label, "\n", `<BR ALIGN="LEFT"/>`, -1)
			}
			label += `<BR ALIGN="LEFT"/>`
			if n.Branches.Type != "" {
				var typeLabel string
				var color = "orange"
				switch n.Branches.Type {
				case "message":
					color = "#2d93ad"
					typeLabel = "message"
				case "bindings":
					color = "#52aa5e"
					typeLabel = "bindings"
				}
				// label = "<FONT COLOR=\"" + color + "\">" + n.Branches.Type + "</FONT>: " + label
				// label = "<FONT POINT-SIZE=\"6\" COLOR=\"" + color + "\">" + typeLabel + "</FONT><BR ALIGN=\"LEFT\"/>" + label
				label = `<FONT COLOR="` + color + `">` + typeLabel + `</FONT>` +
					`<FONT POINT-SIZE="8"><BR ALIGN="LEFT"/>` + label + `</FONT>`
			}
			if b.Guard != nil {
				if b.GuardSource != nil {
					var src string
					x := b.GuardSource.Source
					if s, is := x.(string); is {
						src = s
					} else {
						src = fmt.Sprintf("%#v", x)
					}
					src = strings.Replace(src, "<", `&lt;`, -1)
					src = strings.Replace(src, ">", `&gt;`, -1)
					label += `<FONT POINT-SIZE="6">` +
						`<BR/>` + strings.Replace(src+"\n", "\n", `<BR ALIGN="LEFT"/>`, -1) + `<BR/>` +
						`</FONT>`
				} else {
					label += `<BR ALIGN="LEFT"/>guarded<BR ALIGN="LEFT"/>`
				}
			}
			label = strings.Replace(label, "\n", "", -1)

			color := "black"
			if fromNode == name && toNode == b.Target {
				color = "red"
			}

			// label = fmt.Sprintf("[%d/%d] %s", i+1, len(n.Branches.Branches), label)
			label = fmt.Sprintf("%d/%d %s", i+1, len(n.Branches.Branches), label)
			fmt.Fprintf(w, "  %s -> %s [ color=\"%s\" label = <%s> ]\n",
				name, b.Target, color, label)
		}

		return nil
	}

	start, have := nodes["start"]
	if have {
		process("start", start)
	}

	for name, n := range nodes {
		if name == "start" {
			continue
		}
		process(name, n)
	}

	fmt.Fprintf(w, "}\n")
	return w.Close()
}

// PNG generates a PNG image based on output from Dot.
//
// This function with write two files: basename.dot and basename.png,
// where the basename is the given string.
func PNG(spec *Spec, basename string, fromNode, toNode string) (string, error) {
	dotname := basename + ".dot"
	pngname := basename + ".png"

	// ToDo: Use mktemp
	dotfile, err := os.Create(dotname)
	if err != nil {
		return pngname, err
	}
	if err := Dot(spec, dotfile, fromNode, toNode); err != nil {
		return pngname, err
	}
	cmd := "dot -Tpng -Gstart=1 " + dotname + " > " + pngname
	if err := exec.Command("bash", "-c", cmd).Run(); err != nil {
		return pngname, err
	}
	return pngname, nil
}

func escape(s string) string {
	return strings.Replace(s, `"`, `\"`, -1)
}

func escbraces(s string) string {
	s = strings.Replace(s, "{", "\\{", -1)
	s = strings.Replace(s, "}", "\\}", -1)
	return s
}
