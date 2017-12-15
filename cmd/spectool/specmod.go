package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/tools"

	"github.com/jsccast/yaml"
)

var Mods = map[string]Mod{
	"addMessageBranches":    &AddMessageBranchesMod{},
	"addGenericCancelNode":  &AddGenericCancelNodeMod{},
	"addOrderedOutMessages": &AddOrderedOutMessagesMod{},
	"analyze":               &Analyzer{},
	"graph":                 &Grapher{},
}

var (
	NoTargetNode = errors.New("no target node")
	NodeExists   = errors.New("node exists")
)

type Mod interface {
	F(*core.Spec) error
	Doc() string
	Flags() *flag.FlagSet
}

// AddMessageBranches adds a branch to each message node in the
// Spec. The added branches match the given pattern and target the
// given target.
//
// The Spec's Doc is updated to note that this processing has occurred.
func AddMessageBranches(s *core.Spec, pattern interface{}, target string) error {
	if _, have := s.Nodes[target]; !have {
		return NoTargetNode
	}

	for name, node := range s.Nodes {
		if node == nil {
			node = &core.Node{}
			s.Nodes[name] = node
		}
		if node.Branches == nil {
			continue
		}
		if node.Branches.Type != "message" {
			continue
		}
		branch := &core.Branch{
			Pattern: pattern,
			Target:  target,
		}
		node.Branches.Branches = append(node.Branches.Branches, branch)
	}

	s.Doc = s.Doc + fmt.Sprintf(`

This spec has processed by AddMessageBranches with target "%s".
`, target)

	return nil
}

type AddMessageBranchesMod struct {
	PatternJS    string
	Target       string
	ParsePattern bool
}

func (c *AddMessageBranchesMod) Doc() string {
	return `
Adds an additional branch to every message node. The branch has the specified 
pattern and target.
`
}

func (c *AddMessageBranchesMod) Flags() *flag.FlagSet {
	flags := flag.NewFlagSet("addMessageBranches", flag.PanicOnError)

	flags.StringVar(&c.PatternJS, "p", `{"ctl":"cancel"}`, "pattern")
	flags.StringVar(&c.Target, "t", "cancel", "target")
	flags.BoolVar(&c.ParsePattern, "P", false, "parse the pattern as JSON")

	return flags
}

func (c *AddMessageBranchesMod) F(s *core.Spec) error {
	var pattern interface{}

	if c.ParsePattern {
		if err := json.Unmarshal([]byte(c.PatternJS), &pattern); err != nil {
			return err
		}
	} else {
		pattern = c.PatternJS
	}

	return AddMessageBranches(s, pattern, c.Target)
}

// CancelNodeYAML is a node that iterates through any
// _.bindings.cleanupMessages and emits them.  Returns the bindings
// with empty 'cleanupMessages'.
var CancelNodeYAML = `
action:
  doc: Emit any clean-up messages.
  interpreter: goja
  source: |-
    for (var i = 0; i < _.bindings.cleanupMessages.length; i++) {
       _.out(_.bindings.cleanupMessages[i]);
    }
    _.bindings.cleanupMessages = [];
    _.bindings;
branching:
  branches:
  - target: listenForEnable
`

// AddGenericCancelNode adds a 'cancel' node as given by
// CancelNodeYAML.
func AddGenericCancelNode(s *core.Spec) error {
	if _, have := s.Nodes["cancel"]; have {
		return NodeExists
	}

	var n core.Node
	if err := yaml.Unmarshal([]byte(CancelNodeYAML), &n); err != nil {
		return err
	}

	if s.Nodes == nil {
		s.Nodes = make(map[string]*core.Node, 32)
	}

	s.Nodes["cancel"] = &n

	return nil
}

type AddGenericCancelNodeMod struct {
}

func (m *AddGenericCancelNodeMod) Doc() string {
	return `
Adds a "cancel" node.  The node iterates over '_.bindings.cleanupMessages' 
and emits each message.  The node's only branch has the target 
'listenForEnable'.
`
}

func (m *AddGenericCancelNodeMod) Flags() *flag.FlagSet {
	return flag.NewFlagSet("addGenericCancelNode", flag.PanicOnError)
}

func (m *AddGenericCancelNodeMod) F(s *core.Spec) error {
	return AddGenericCancelNode(s)
}

type AddOrderedOutMessagesMod struct {
	Prefix string

	ParseJSON bool

	Timeout time.Duration

	TimeoutNodeName string

	StartNodeName string

	OutAndIns []struct {
		Out interface{} `json:"e"` // "emit"
		In  interface{} `json:"r"` // "receive"
	}

	OutAndInsJS string

	EndNodeName string
}

// specmod addOrderedOutMessages -p 'lunch_' -s '00' -e done -m '[{"e":{"order":"beer"},"r":{"deliver":"beer"}},{"e":{"order":"queso"},"r":{"deliver":"queso"}},{"e":{"order":"tacos"},"r":{"deliver":"tacos"}}]'

func (m *AddOrderedOutMessagesMod) F(s *core.Spec) error {

	if m.ParseJSON {
		if err := json.Unmarshal([]byte(m.OutAndInsJS), &m.OutAndIns); err != nil {
			return err
		}
	}

	if m.EndNodeName == "" {
		return fmt.Errorf("need an EndNodeName (-e)")
	}

	if s.Nodes == nil {
		s.Nodes = make(map[string]*core.Node, 32)
	}

	{
		// Initial node

		nodeName := fmt.Sprintf("%s%s", m.Prefix, m.StartNodeName)
		if _, have := s.Nodes[nodeName]; have {
			return fmt.Errorf(`node "%s" exists`, nodeName)
		}

		targetNodeName := fmt.Sprintf("%s%02d_emit", m.Prefix, 0)

		node := &core.Node{
			Branches: &core.Branches{
				Branches: []*core.Branch{
					{
						Target: targetNodeName,
					},
				},
			},
		}

		s.Nodes[nodeName] = node
	}

	last := len(m.OutAndIns) - 1
	for i, oi := range m.OutAndIns {

		var timeoutMsg interface{}

		{
			// Node to emit the message.

			nodeName := fmt.Sprintf("%s%02d_emit", m.Prefix, i)
			if _, have := s.Nodes[nodeName]; have {
				return fmt.Errorf(`node "%s" exists`, nodeName)
			}

			targetNodeName := fmt.Sprintf("%s%02d_recv", m.Prefix, i)

			out, err := json.Marshal(&oi.Out)
			if err != nil {
				return err
			}
			src := fmt.Sprintf("_.out(%s);", out)
			if 0 < m.Timeout {
				timer := core.Gensym(32) // Better seed the RNG!
				msg := map[string]interface{}{
					"timeout": timer,
					"from":    nodeName,
					"at":      targetNodeName,
				}

				js, err := json.Marshal(&msg)
				if err != nil {
					return err
				}
				// Canonicalize
				if err = json.Unmarshal(js, &timeoutMsg); err != nil {
					return err
				}

				src += "\n" +
					fmt.Sprintf("_.out({makeTimer:{id:'%s',in:'%s',message:%s}});\n",
						timer,
						m.Timeout,
						js)

				src += fmt.Sprintf("var bs = _.bindings.Extend('timer', '%s');\n", timer)
				src += "bs.cleanupMessages = [{cancelTimer: _.bindings.timer}];\n"
				src += "bs;\n"

			} else {
				src += "\n_.bindings;"
			}

			node := &core.Node{
				ActionSource: &core.ActionSource{
					Interpreter: "goja",
					Source:      src,
				},
				Branches: &core.Branches{
					Branches: []*core.Branch{
						{
							Target: targetNodeName,
						},
					},
				},
			}

			s.Nodes[nodeName] = node
		}

		{
			// Node to wait for ack.

			nodeName := fmt.Sprintf("%s%02d_recv", m.Prefix, i)
			if _, have := s.Nodes[nodeName]; have {
				return fmt.Errorf(`node "%s" exists`, nodeName)
			}

			targetNodeName := fmt.Sprintf("%s%02d_remt", m.Prefix, i)
			if _, have := s.Nodes[targetNodeName]; have {
				return fmt.Errorf(`node "%s" exists`, targetNodeName)
			}

			if timeoutMsg == nil {
				targetNodeName = fmt.Sprintf("%s%02d_emit", m.Prefix, i+1)
				if i == last {
					targetNodeName = m.EndNodeName
				} else {
					if _, have := s.Nodes[targetNodeName]; have {
						return fmt.Errorf(`node "%s" exists`, targetNodeName)
					}
				}
			}

			node := &core.Node{
				Branches: &core.Branches{
					Type: "message",
					Branches: []*core.Branch{
						{
							Pattern: oi.In,
							Target:  targetNodeName,
						},
					},
				},
			}

			if timeoutMsg != nil {
				b := &core.Branch{
					Pattern: timeoutMsg,
					Target:  m.TimeoutNodeName,
				}
				node.Branches.Branches = append(node.Branches.Branches, b)
			}

			s.Nodes[nodeName] = node
		}

		if timeoutMsg != nil { // Sorry
			// Node to cancel timer

			nodeName := fmt.Sprintf("%s%02d_remt", m.Prefix, i)
			if _, have := s.Nodes[nodeName]; have {
				return fmt.Errorf(`node "%s" exists`, nodeName)
			}

			targetNodeName := fmt.Sprintf("%s%02d_emit", m.Prefix, i+1)
			if i == last {
				targetNodeName = m.EndNodeName
			} else {
				if _, have := s.Nodes[targetNodeName]; have {
					return fmt.Errorf(`node "%s" exists`, targetNodeName)
				}
			}

			src := "_.out({cancelTimer: _.bindings.timer});\n"
			src += "var bs = _.bindings.Remove('timer');\n"
			src += "bs.cleanupMessages = [];\n"
			src += "bs;\n"

			node := &core.Node{
				ActionSource: &core.ActionSource{
					Interpreter: "goja",
					Source:      src,
				},
				Branches: &core.Branches{
					Branches: []*core.Branch{
						{
							Pattern: oi.In,
							Target:  targetNodeName,
						},
					},
				},
			}

			s.Nodes[nodeName] = node
		}
	}

	return nil
}

func (m *AddOrderedOutMessagesMod) Doc() string {
	return "ToDo"
}

func (m *AddOrderedOutMessagesMod) Flags() *flag.FlagSet {
	flags := flag.NewFlagSet("addOrderedOutMessages", flag.PanicOnError)

	flags.StringVar(&m.Prefix, "p", "oi_", "prefix for node names")
	flags.BoolVar(&m.ParseJSON, "P", true, "parse the messages as JSON")
	flags.StringVar(&m.StartNodeName, "s", "start", "name suffix for starting node")
	flags.DurationVar(&m.Timeout, "d", 0, "timeout duration (if any)")
	flags.StringVar(&m.TimeoutNodeName, "t", "timedout", "node name for node for timeouts")
	flags.StringVar(&m.OutAndInsJS, "m", "[]", "name for starting node")
	flags.StringVar(&m.EndNodeName, "e", "", "name for following node")

	return flags
}

type Analyzer struct {
}

func (m *Analyzer) F(s *core.Spec) error {
	a, err := tools.Analyze(s)
	if err != nil {
		return err
	}
	bs, err := yaml.Marshal(&a)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s\n", bs)

	return nil
}

func (m *Analyzer) Doc() string {
	return "ToDo"
}

func (m *Analyzer) Flags() *flag.FlagSet {
	return flag.NewFlagSet("analyze", flag.PanicOnError)
}

type Grapher struct {
	OutputFilename string
}

func (m *Grapher) F(s *core.Spec) error {
	f, err := os.Create(m.OutputFilename)
	if err != nil {
		return err
	}

	return tools.Dot(s, f, "", "") // Will Close f.
}

func (m *Grapher) Doc() string {
	return "ToDo"
}

func (m *Grapher) Flags() *flag.FlagSet {
	fs := flag.NewFlagSet("graph", flag.PanicOnError)
	flag.StringVar(&m.OutputFilename, "o", "spec.dot", "output filename")
	return fs
}
