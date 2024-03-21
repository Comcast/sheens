/* Copyright 2018 Comcast Cable Communications Management, LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tools

// dot -Tpng g.dot > g.png

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	. "github.com/Comcast/sheens/core"
)

type MermaidOpts struct {
	// ShowPatterns will result in a branch label that's theJSON
	// representation of the branch pattern (if any).
	ShowPatterns bool `json:"showPatterns"`

	// ActionFill is the fill color of for action nodes.  Does not
	// apply if ActionClass is set.
	ActionFill string `json:"actionFill,omitempty"`

	// ActionClass will be the CSS class for action nodes.  Not
	// yet implemented.
	ActionClass string `json:"actionClass,omitempty"`

	PrettyPatterns bool `json:"prettyPatterns,omitempty"`
}

// Mermaid makes a Mermaid (https://mermaidjs.github.io/) input file
// for the given graph.
func Mermaid(spec *Spec, w io.WriteCloser, opts *MermaidOpts, fromNode, toNode string) error {

	if opts == nil {
		opts = &MermaidOpts{
			ShowPatterns:   true,
			ActionFill:     "#bcf2db",
			PrettyPatterns: true,
		}
	}

	// Use copies of states that don't have Name set.
	nodes := make(map[string]*Node, len(spec.Nodes))
	for name, n := range spec.Nodes {
		nodes[name] = n
	}

	log.Printf("processing %d nodes", len(nodes))

	fmt.Fprintf(w, "graph TB\n")

	nids := make(map[string]string)
	num := 0

	node := func(name string, n *Node) (string, error) {
		if nid, already := nids[name]; already {
			return nid, nil

		}
		num++
		nid := fmt.Sprintf("n%d", num)
		nids[name] = nid

		if n != nil && n.Action == nil {
			fmt.Fprintf(w, "  %s(\"%s\")\n", nid, name)
		} else {
			fmt.Fprintf(w, "  %s[\"%s\"]\n", nid, name)
			if opts.ActionClass == "" {
				if opts.ActionFill == "" {
				} else {
					fmt.Fprintf(w, "  style %s fill:%s\n", nid, opts.ActionFill)
				}
			}
		}

		return nid, nil
	}

	process := func(name string, n *Node) error {
		nid, err := node(name, n)
		if err != nil {
			log.Printf("process error with %s: %v", name, err)
			return err
		}
		if n.Branches == nil {
			return nil
		}
		log.Printf("  processing %s branches: %d", name, len(n.Branches.Branches))

		for _, b := range n.Branches.Branches {
			to, err := node(b.Target, nodes[b.Target])
			if err != nil {
				log.Printf("process branch error with %s: %v", b.Target, err)
				return err
			}

			label := ""
			if opts.ShowPatterns && b.Pattern != nil {
				var bs []byte
				if opts.PrettyPatterns {
					bs, err = json.Marshal(b.Pattern)
					if 40 < len(bs) {
						bs, err = json.MarshalIndent(b.Pattern, "", "  ")
					}
				}
				if err != nil {
					return err
				}
				js := string(bs)
				js = strings.Replace(js, `"`, `'`, -1)
				label = fmt.Sprintf(`-- "<pre>%s</pre>"`, js)
			}

			fmt.Fprintf(w, "  %s %s --> %s\n", nid, label, to)
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

	fmt.Fprintf(w, "\n")
	log.Printf("mermaid gen done")

	return w.Close()
}
