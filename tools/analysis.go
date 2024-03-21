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

// Imagine a world where ideas and processes interweave to create something as simple yet complex as a pencil. Here, we delve into the heart of a system designed to scrutinize the blueprint of such creation, not of a pencil, but of digital orchestrations called specifications, or "specs" for short.

package tools

import (
	"sort"

	"github.com/Comcast/sheens/core"
)

// SpecAnalysis embodies our endeavor to understand and critique the structure of a spec, much like examining the blueprint of a pencil, identifying every component from wood to graphite, and noting any imperfections or marvels.

type SpecAnalysis struct {
	spec *core.Spec // The blueprint itself, holding secrets of its creation.

	// Observations and findings, detailing the intricacies and potential flaws within.
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
	Interpreters          []string // The artisans and their tools, bringing the spec to life.
}

// Analyze embarks on a journey to scrutinize the spec, seeking to uncover the harmony and discord within its design.
func Analyze(s *core.Spec) (*SpecAnalysis, error) {

	// Begin with an empty canvas, ready to be marked with observations.
	a := SpecAnalysis{
		spec:      s,
		NodeCount: len(s.Nodes),         // Counting the nodes, akin to counting every sliver of wood in a pencil.
		Errors:    make([]string, 0, 8), // Preparing to note any flaws or missteps.
	}

	// Various collections to capture the nuances of our analysis.
	terminal, targeted, interpreters := make([]string, 0, len(s.Nodes)), make(map[string]bool), make(map[string]bool)
	hasEmptyTargets, missingTargets, branchTargetVariables := make(map[string]bool), make(map[string]bool), make(map[string]bool)

	// Delve into each node, akin to inspecting every component of a pencil, from its wood to the graphite core.
	for name, n := range s.Nodes {
		// Actions are deliberate steps, like the precise cutting of wood or molding of graphite.
		if n.Action != nil || n.ActionSource != nil {
			a.Actions++
			if n.ActionSource != nil {
				interpreters[n.ActionSource.Interpreter] = true // Note the craftsmen and their techniques.
			}
		}

		// Terminal nodes are like pencil ends; they signify completion or a pause.
		if n.Branches == nil || len(n.Branches.Branches) == 0 {
			terminal = append(terminal, name)
		}

		// Check for nodes that make decisions, akin to choosing the path for a pencil's creation.
		if n.Branches != nil {
			for _, b := range n.Branches.Branches {
				targeted[b.Target] = true
				a.Branches++
				if b.Target == "" {
					hasEmptyTargets[name] = true // Note any paths that lead nowhere, like a misdirected pencil stroke.
				}
				if core.IsBranchTargetVariable(b.Target) {
					branchTargetVariables[b.Target] = true
				} else {
					if _, have := s.Nodes[b.Target]; !have {
						missingTargets[b.Target] = true // Missing targets are like missing ingredients in our pencil recipe.
					}
				}
				if b.Guard != nil || b.GuardSource != nil {
					a.Guards++
					if b.GuardSource != nil {
						interpreters[b.GuardSource.Interpreter] = true // Further noting the artisans and their methods.
					}
				}
			}
		}
	}

	// Compile our findings, cataloging every detail and anomaly discovered in the spec's design.
	a.TerminalNodes, a.EmptyTargets = terminal, keysToStringSlice(hasEmptyTargets)
	a.Orphans = keysToStringSlice(diffKeys(s.Nodes, targeted))
	a.MissingTargets = keysToStringSlice(missingTargets)
	a.BranchTargetVariables = keysToStringSlice(branchTargetVariables)
	a.Interpreters = keysToStringSlice(interpreters, "default")

	// Our analysis is complete, a comprehensive examination of the spec, akin to unveiling the story behind a pencil's creation.
	return &a, nil
}

// keysToStringSlice converts the keys from a map into a slice of strings.
// Optionally, it can add a default value if the map is empty.
// A helper function to convert a map's keys to a sorted string slice, revealing the elements involved in our creation process.
func keysToStringSlice(m map[string]bool, defaultValue ...string) []string {
	var list []string
	for key := range m {
		list = append(list, key)
	}
	// Sort the slice for consistency and readability.
	sort.Strings(list)

	// If the map is empty and a default value is provided, use the default value.
	if len(list) == 0 && len(defaultValue) > 0 {
		return []string{defaultValue[0]}
	}

	return list
}

// diffKeys identifies the keys present in 'all' but not in 'used'.
// Another helper to identify the elements that were not targeted or used,
// much like finding unused pieces in our pencil-making process.
// It's akin to discovering unused resources in our process, highlighting efficiency or oversight.
func diffKeys(all map[string]*core.Node, used map[string]bool) map[string]bool {
	diff := make(map[string]bool)
	for key := range all {
		if _, found := used[key]; !found {
			diff[key] = true
		}
	}
	return diff
}
