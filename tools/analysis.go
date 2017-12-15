package tools

import (
	"fmt"

	"github.com/Comcast/sheens/core"
)

type SpecAnalysis struct {
	spec *core.Spec

	Errors                []string
	NodeCount             int
	Branches              int
	Actions               int
	Guards                int
	TerminalNodes         []string
	Orphans               []string
	EmptyTargets          []string
	MissingTargets        []string
	BranchTargetVariables []string

	Interpreters []string
}

func Analyze(s *core.Spec) (*SpecAnalysis, error) {

	// Check for timeout branches in nodes with actions.

	// Check for mutually exclusive branches (which would not
	// necessarily be errors).

	// Check for obvious infinite loops.  For example, a node with
	// no action and a default branch that targets the same node
	// -- or similar or indirectly or ...

	a := SpecAnalysis{
		spec:      s,
		NodeCount: len(s.Nodes),
		Errors:    make([]string, 0, 8),
	}

	var (
		terminal              = make([]string, 0, len(s.Nodes))
		targeted              = make(map[string]bool)
		interpreters          = make(map[string]bool)
		hasEmptyTargets       = make(map[string]bool)
		missingTargets        = make(map[string]bool)
		branchTargetVariables = make(map[string]bool)
	)

	// ToDo: Check that ErrorNode exists.

	for name, n := range s.Nodes {
		haveAction := false
		if n.Action != nil || n.ActionSource != nil {
			a.Actions++
			haveAction = true
			if n.ActionSource != nil {
				interpreters[n.ActionSource.Interpreter] = true
			}
		}
		if n.Branches == nil || len(n.Branches.Branches) == 0 {
			terminal = append(terminal, name)
		}
		if haveAction && n.Branches != nil && n.Branches.Type == "message" {
			a.Errors = append(a.Errors,
				fmt.Sprintf(`node "%s" has an action with "%s" branching"`,
					name, n.Branches.Type))
		}
		if n.Branches != nil {
			for _, b := range n.Branches.Branches {
				targeted[b.Target] = true
				a.Branches++
				if b.Target == "" {
					hasEmptyTargets[name] = true
				}
				if core.IsBranchTargetVariable(b.Target) {
					branchTargetVariables[b.Target] = true
				} else {
					if _, have := s.Nodes[b.Target]; !have {
						missingTargets[b.Target] = true
					}
				}
				if b.Guard != nil || b.GuardSource != nil {
					a.Guards++
					if b.GuardSource != nil {
						interpreters[b.GuardSource.Interpreter] = true
					}
				}
			}
		}

		// ToDo: If ActionErrorBranches, then see if any
		// branches explicitly binding a variable to
		// "actionError". If not, warn?  Would be nice if we
		// had OCaml-style type inference!

	}
	a.TerminalNodes = terminal

	emptyTargets := make([]string, 0, len(hasEmptyTargets))
	for name := range hasEmptyTargets {
		emptyTargets = append(emptyTargets, name)
	}
	a.EmptyTargets = emptyTargets

	all := make(map[string]bool, len(s.Nodes))
	for name := range s.Nodes {
		all[name] = true
	}
	for name := range targeted {
		delete(all, name)
	}
	orphans := make([]string, 0, len(all))
	for name := range all {
		orphans = append(orphans, name)
	}
	a.Orphans = orphans

	missing := make([]string, 0, len(missingTargets))
	for name := range missingTargets {
		missing = append(missing, name)
	}
	a.MissingTargets = missing

	vars := make([]string, 0, len(branchTargetVariables))
	for name := range branchTargetVariables {
		vars = append(vars, name)
	}
	a.BranchTargetVariables = vars

	interps := make([]string, 0, len(interpreters))
	for name := range interpreters {
		if name == "" {
			name = "default"
		}
		interps = append(interps, name)
	}
	a.Interpreters = interps

	return &a, nil
}
