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

// Package match implements the core pattern matcher.
package match

import (
	"errors"
	"strings"
)

type Matcher struct {
	// AllowPropertyVariables enables the experimental support for a
	// property variable in a pattern that contains only one property.
	AllowPropertyVariables bool

	// CheckForBadPropertyVariables runs a test to verify that a pattern
	// does not contain a property variable along with other properties.
	//
	// This check might not be necessary because the other code will
	// report an error if a bad property variable is actually enountered
	// during matching.  The interesting twist is that if a match fails
	// before encountering the bad property variable, then that code will
	// not report the problem.  In order to report the problem always,
	// turn on this switch.  Performance will suffer, but any bad property
	// variable will at least be caught.
	CheckForBadPropertyVariables bool

	// Inequalities is a switch to turn on experimental binding
	// inequality support.
	//
	// With this feature, pattern matching supports numeric
	// inequalities in addition to the standard equality
	// predicate.  The input bindings should include a binding for
	// a variable with a name that contains either "<", ">", "<=",
	// ">=", or "!=" immediately after the leading "?".  The input
	// pattern can then use that variable.  When matching, a value
	// X will match that variable only if the binding Y for that
	// variable satisfies the inequality with X and Y (in that
	// order). In this case, the output bindings will include a
	// new binding for a variable with the same name as the
	// inequality variable but without the actual inequality.
	//
	// For example, given input bindings {"?<n":10}, pattern
	// {"n":"?<n"}, and message {"n":3}, the match will succeed
	// with bindings {"?<n":10,"?n":3}.
	//
	// See match_test.js for several examples. (Search for
	// "inequality".)
	//
	// For now at least, the inequalities only work for numeric
	// values.  We might support strings later.
	//
	// Yes, such a feature makes us stare down a slippery slope.
	// However, we are brave, and we do not shy away from even
	// grave danger.
	//
	//   "[In Sheens] we live exactly as we please, and yet are
	//   just as ready to encounter every legitimate danger."
	//
	//   --Pericles
	//
	// The immediate motivation for this feature was to support
	// timer fallbacks.  For example, using the Goja interpreter,
	// an action could establish a binding with a value that's a
	// number representing a future time in UNIX milliseconds.
	// Then a branch pattern can check for a message containing
	// the current time which is greater than that number.  When a
	// machine is loaded, a system could send the machine a
	// "current time" message, which could advance the machine
	// according to such a branch.  Using this technique, a
	// machine that creates a timer that somehow gets lots could
	// regain at least some sense of what's going on.  Similarly,
	// a system that doesn't have asynchronous, timer-driven
	// messaging support could simply send messages every second
	// (or at whatever internal) to provide timer-like
	// functionality (albeit with inefficiencies and without the
	// message-oriented timer protocol that's been offered
	// elsewhere).
	Inequalities bool
}

var DefaultMatcher = &Matcher{
	AllowPropertyVariables:       true,
	CheckForBadPropertyVariables: true,
	Inequalities:                 true,
}

func (m *Matcher) checkForBadPropertyVariables(pattern map[string]interface{}) error {
	if !m.CheckForBadPropertyVariables {
		return nil
	}
	if len(pattern) <= 1 {
		return nil
	}
	for k := range pattern {
		if m.IsVariable(k) {
			return errors.New(`can't have a variable as a key ("` + k + `") with other keys`)
		}
	}
	return nil
}

// Bindings is a map from variables (strings starting with a '?') to
// their values.
type Bindings map[string]interface{}

func NewBindings() Bindings {
	return make(Bindings, 8)
}

// Extends adds the property; modifies and returns the Bindings.
//
// The Bindings are modified.
func (bs Bindings) Extend(p string, v interface{}) Bindings {
	bs[p] = v
	return bs
}

// Extends adds the properties; modifies and returns the Bindings.
//
// The Bindings are modified.
func (bs Bindings) Extendm(pairs ...interface{}) (Bindings, error) {
	for i := 0; i < len(pairs); i += 2 {
		x := pairs[i]
		p, is := x.(string)
		if !is {
			return nil, errors.New("Bindings.Entendm given a non-string key")
		}
		if len(pairs) <= i+1 {
			return nil, errors.New("odd args to Bindings.Extendm")
		}
		bs[p] = pairs[i+1]
	}
	return bs, nil
}

// Remove removes the given keys.
//
// The Bindings are modified.
func (bs Bindings) Remove(ps ...string) Bindings {
	for _, p := range ps {
		delete(bs, p)
	}
	return bs
}

// DeleteExcept removes all but the given properties.
//
// Does not copy.
func (bs Bindings) DeleteExcept(keeps ...string) Bindings {
REM:
	for p := range bs {
		for _, keep := range keeps {
			if keep == p {
				continue REM
			}
		}
		delete(bs, p)
	}

	return bs
}

// Copy makes a shallow copy of the Bindings.
func (bs Bindings) Copy() Bindings {
	acc := make(Bindings, len(bs))
	for k, v := range bs {
		acc[k] = v
	}
	return acc
}

// IsVariable reports if the string represents a pattern variable.
//
// All pattern variables start with a '?".
func (m *Matcher) IsVariable(s string) bool {
	return strings.HasPrefix(s, "?")
}

func (m *Matcher) IsOptionalVariable(x interface{}) bool {
	if s, is := x.(string); is {
		return strings.HasPrefix(s, "??")
	}
	return false
}

// IsAnonymousVariable detects a variable of the form '?'.  An binding
// for an anonymous variable shouldn't ever make it into bindins.
func (m *Matcher) IsAnonymousVariable(s string) bool {
	return s == "?"
}

// IsConstant reports if the string represents a constant (and not a
// pattern variable).
func (m *Matcher) IsConstant(s string) bool {
	return !m.IsVariable(s)
}

// mapcatMatch attempts to extend the given bindingss 'bss' based on
// pair-wise matching of the pattern to the fact.
func (m *Matcher) mapcatMatch(bss []Bindings, pattern map[string]interface{}, fact map[string]interface{}) ([]Bindings, error) {
	if err := m.checkForBadPropertyVariables(pattern); err != nil {
		return nil, err
	}

	for k, v := range pattern {
		if m.IsVariable(k) {
			if m.AllowPropertyVariables {
				if len(pattern) == 1 {
					// Iterate over the fact keys and collect match results.
					gather := make([]Bindings, 0, 0)
					for fk, fv := range fact {
						ext := copyBindingss(bss)

						// Try to match keys.
						ext, err := m.matchWithBindingss(ext, k, fk)
						if err != nil {
							return nil, err
						}
						if 0 == len(ext) {
							// Didn't match keys.
							continue
						}
						// Matched keys.  Now check values.
						ext, err = m.matchWithBindingss(ext, v, fv)
						if err != nil {
							return nil, err
						}
						if 0 == len(ext) {
							// Didn't match values.
							continue
						}
						gather = append(gather, ext...)
					}
					bss = gather
					// Since we know we have only one key, the outer loop will terminate.
					// Probably should make that termination explicit.
					return bss, nil
				} else {
					return nil, errors.New(`can't have a variable as a key ("` + k + `") with other keys`)
				}
			} else {
				return nil, errors.New(`can't have a variable as a key ("` + k + `")`)
			}
		} else {
			fv, found := fact[k]
			if !found {
				if m.IsOptionalVariable(v) {
					continue
				}

				return nil, nil
			}

			acc, err := m.matchWithBindingss(bss, v, fv)
			if nil != err {
				return nil, err
			}

			if 0 == len(acc) {
				return nil, nil
			}
			bss = acc
		}
	}
	return bss, nil
}

// arraycatMatch attempts to extend the given list of bindingss 'bsss'
// based on element-wise matching.
//
// An array represents a set; therefore, this function can backtrack,
// which can be scary.
func (m *Matcher) arraycatMatch(bsss [][]Bindings, pattern interface{}, fxas []map[int]interface{}) ([][]Bindings, []map[int]interface{}, error) {
	var nbsss [][]Bindings
	var nfxas []map[int]interface{}
	for i, bss := range bsss {
		mm := fxas[i]
		for j, fact := range mm {
			acc, err := m.matchWithBindingss(copyBindingss(bss), pattern, fact)
			if nil != err {
				return nil, nil, err
			}
			if 0 != len(acc) {
				nbsss = append(nbsss, acc)
				copy := copyMap(mm)
				delete(copy, j)
				nfxas = append(nfxas, copy)
			}
		}
	}
	return nbsss, nfxas, nil
}

func copyMap(source map[int]interface{}) map[int]interface{} {
	target := make(map[int]interface{})
	for p, v := range source {
		target[p] = v
	}
	return target
}

// matchWithBindings attempts to extend the given bindingss 'bss' from
// matches of the fact against the pattern.
//
// Ths function mostly just calls 'Match()'.
func (m *Matcher) matchWithBindingss(bss []Bindings, pattern interface{}, fact interface{}) ([]Bindings, error) {
	acc := make([]Bindings, 0, len(bss))
	for _, bs := range bss {
		matches, err := m.Match(pattern, fact, bs)
		if nil != err {
			return nil, err
		}
		if nil != matches {
			acc = append(acc, matches...)
		}
	}

	return acc, nil
}

// getVariable finds the first variable and the other non-variables.
// For now, we look for at most one variable.
//
// ToDo: Improve.
func (m *Matcher) getVariable(xs []interface{}) (string, []interface{}, error) {
	var v string
	acc := make([]interface{}, 0, len(xs))
	for _, x := range xs {
		switch x.(type) {
		case string:
			s := x.(string)
			if m.IsVariable(s) {
				if v == "" {
					v = s
					continue
				}
				if v == s {
					return "", nil, errors.New("repeated variables not supported")
				}

				return "", nil, errors.New("multiple variables not supported here")
			}
		}
		acc = append(acc, x)
	}
	return v, acc, nil
}

// Matches attempts to match the given fact with the given pattern.
// Returns an array of 'Bindings'.  Each Bindings is just a map from
// variables to their values.
//
// Note that this function returns multiple (sets of) bindings.  This
// ambiguity is introduced when a pattern contains an array that
// contains a variable.
func (m *Matcher) Matches(pattern interface{}, fact interface{}) ([]Bindings, error) {
	return m.Match(pattern, fact, make(Bindings))
}

// fudge is a hack to cast numbers to float64s.
func fudge(x interface{}) interface{} {
	switch vv := x.(type) {
	case float64:
		return vv
	case float32:
		return float64(vv)
	case int64:
		return float64(vv)
	case int32:
		return float64(vv)
	case int:
		return float64(vv)
	default:
		return x
	}
}

// Match is a verion of 'Matches' that takes initial bindings.
//
// Those initial bindings are not modified.
func (m *Matcher) Match(pattern interface{}, fact interface{}, bindings Bindings) ([]Bindings, error) {
	return m.match(pattern, fact, bindings.Copy())
}

// match is a verion of 'Matches' that takes initial bindings (which
// can be modified).
func (m *Matcher) match(pattern interface{}, fact interface{}, bindings Bindings) ([]Bindings, error) {

	pattern = fudge(pattern)
	fact = fudge(fact)

	// ToDo: Review for garbage reduction.

	if bindings == nil {
		return nil, nil
	}

	p := pattern
	f := fact
	bs := bindings

	switch vv := p.(type) {
	case nil:
		switch f.(type) {
		case nil:
			return []Bindings{bindings}, nil
		default:
			return nil, nil
		}

	// case reflect.Value:
	// 	return match(vv.Interface(), fact, bindings)

	case bool:
		switch f.(type) {
		case bool:
			y := f.(bool)
			if vv == y {
				return []Bindings{bindings}, nil
			}
			return nil, nil
		default:
			return nil, nil
		}

	case float64:
		switch f.(type) {
		case float64:
			y := f.(float64)
			if vv == y {
				return []Bindings{bindings}, nil
			} else {
				return nil, nil
			}
		default:
			return nil, nil
		}

	case string:
		if m.IsConstant(vv) {
			switch f.(type) {
			case string:
				fs := f.(string)
				if vv == fs {
					return []Bindings{bindings}, nil
				} else {
					return nil, nil
				}
			default:
				return nil, nil
			}
		} else { // IsVariable
			if m.IsAnonymousVariable(vv) {
				return []Bindings{bs}, nil
			}
			if using, bss, err := m.inequal(fact, bindings, vv); err != nil {
				return nil, err
			} else if using {
				return bss, err
			}
			binding, found := bs[vv]
			if found {
				return m.match(binding, fact, bindings)
			} else {
				// add new binding
				bs[vv] = fact
				return []Bindings{bs}, nil
			}
		}

	case map[string]interface{}:
		mm, ok := p.(map[string]interface{})
		if !ok {
			mm = map[string]interface{}(p.(mmap))
		}
		switch f.(type) {
		case map[string]interface{}:
			fm, ok := f.(map[string]interface{})
			if !ok {
				fm = map[string]interface{}(f.(mmap))
			}
			if 0 == len(mm) {
				// Empty map pattern matched any given map.
				return []Bindings{bs}, nil
			}
			return m.mapcatMatch([]Bindings{bs}, mm, fm)
		default:
			return nil, nil
		}

	case []interface{}:
		//separate variable and constants
		v, xs, err := m.getVariable(vv)
		if nil != err {
			return nil, err
		}
		switch f.(type) {
		case []interface{}:

			// index fact array, separate array/map from the string/float fact
			fa := f.([]interface{})
			fxs := make(map[interface{}]bool)
			fxa := make(map[int]interface{})
			for i, y := range fa {
				switch y.(type) {
				case float64, string, bool, nil:
					fxs[y] = true
				default:
					fxa[i] = y
				}
			}

			bsss := [][]Bindings{{bindings}}
			fxas := []map[int]interface{}{fxa}

			// iterate pattern values and match with fact values
			for _, x := range xs {
				switch x.(type) {
				case float64, string, bool, nil:
					_, found := fxs[x]
					if found {
						delete(fxs, x)
					} else {
						return nil, nil
					}
				default:
					if 0 == len(fxa) {
						return nil, nil
					} else {
						bsss, fxas, err = m.arraycatMatch(bsss, x, fxas)
						if nil != err {
							return nil, err
						}
						if nil == bsss {
							return nil, nil
						}
					}
				}
			}

			// merge left-over facts
			for _, fxa := range fxas {
				i := len(fa)
				for fact := range fxs {
					fxa[i] = fact
					i++
				}
			}

			// bind pattern variable and match again
			if "" == v {
				return combine(bsss), nil
			} else {
				previous := bsss
				bsss, fxas, err = m.arraycatMatch(bsss, v, fxas)
				if nil != err {
					return nil, err
				}
				if len(bsss) == 0 && m.IsOptionalVariable(v) {
					bsss = previous
				}
				return combine(bsss), nil
			}

		default:
			return nil, nil
		}

	default:
		return nil, &UnknownPatternType{p}
	}
}

func combine(bsss [][]Bindings) []Bindings {
	switch len(bsss) {
	case 0:
		return nil
	case 1:
		return bsss[0]
	default:
		var nbss []Bindings
		for _, bss := range bsss {
			nbss = append(nbss, bss...)
		}
		return nbss
	}
}

func copyBindingss(bss []Bindings) []Bindings {
	acc := make([]Bindings, 0, len(bss))
	for _, bs := range bss {
		acc = append(acc, bs.Copy())
	}

	return acc
}

// UnknownPatternType is an error that includes the thing that's
// causing the trouble.
type UnknownPatternType struct {
	Pattern interface{}
}

func (e *UnknownPatternType) Error() string {
	return "unknown pattern type"
}

// mmap is now a mystery to me.
type mmap map[string]interface{}

func (m *Matcher) inequal(fact interface{}, bs Bindings, v string) (bool, []Bindings, error) {
	if !m.Inequalities {
		return false, nil, nil
	}

	if v[0] != '?' {
		return false, nil, nil
	}

	x, have := bs[v]
	if !have {
		return false, nil, nil
	}
	x = fudge(x)
	b, is := x.(float64)
	if !is {
		return false, nil, nil
	}

	x = fudge(fact)
	a, is := x.(float64)
	if !is {
		return false, nil, nil
	}

	var ineq, vv string
	switch len(v) {
	case 1, 2:
		return false, nil, nil
	default:
		ineqv := v[1:]
		for _, ie := range []string{"<=", ">=", "!=", ">", "<"} {
			if strings.HasPrefix(ineqv, ie) {
				ineq = ie
				vv = "?" + ineqv[len(ie):]
				break
			}
		}
	}
	if vv == "" {
		return false, nil, nil
	}

	satisfied := false
	switch ineq {
	case "<":
		if a < b {
			satisfied = true
		}
	case "<=":
		if a <= b {
			satisfied = true
		}
	case ">":
		if a > b {
			satisfied = true
		}
	case ">=":
		if a >= b {
			satisfied = true
		}
	case "!=":
		if a != b {
			satisfied = true
		}
	}

	if !satisfied {
		return true, nil, nil
	}

	x, given := bs[vv]
	if given {
		c, is := fudge(x).(float64)
		if !is {
			return false, nil, nil
		}
		if c != a {
			return true, nil, nil
		}
		// Don't need to update the bindings.
		return true, []Bindings{bs}, nil
	}

	bs[vv] = a
	return true, []Bindings{bs}, nil
}

func Match(pattern interface{}, fact interface{}, bindings Bindings) ([]Bindings, error) {
	return DefaultMatcher.Match(pattern, fact, bindings)
}
