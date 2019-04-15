package match

// Fuzz patterns and messages.  Match and then verify non-error
// results.

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// Fuzz has parameters used to generate random patterns and messages.
type Fuzz struct {
	MapWidth    int
	ArrayWidth  int
	Alphabet    string
	VarAlphabet string
	VarWidth    int
	PropWidth   int
	StringWidth int
	MaxNumber   float64

	Nils     float64
	Strings  float64
	Vars     float64
	VarProps float64
	Bools    float64
	Numbers  float64
	Arrays   float64
	Maps     float64
	Aliens   float64

	// ToDo: Add inequalities and optional vars.

	// generate counts the number of atomic values generated.
	generated int64
}

// NoVars sets Vars and VarProps to zero so that no variables will be
// generated.
func (f *Fuzz) NoVars() {
	f.Vars = 0
	f.VarProps = 0
}

// NewFuzz returns a reasonable, general-purpose Fuzz.
func NewFuzz() *Fuzz {
	return &Fuzz{
		MapWidth:    5,
		ArrayWidth:  5,
		Alphabet:    "abcde",
		VarAlphabet: "UVWXYZ",
		VarWidth:    2,
		StringWidth: 4,
		PropWidth:   2,
		MaxNumber:   10,

		VarProps: 0.2,
		Nils:     1,
		Strings:  3,
		Vars:     2,
		Bools:    1,
		Numbers:  4,
		Arrays:   3,
		Maps:     3,
		Aliens:   0.5,
	}
}

// Gen generates a random pattern or message.
//
// If Gen.Vars and Gen.VarProps are both zero, then the generated
// value will contain no variables (and can be interpreted as a
// message).
func (f *Fuzz) Gen(r *rand.Rand, d int) interface{} {
	f.generated++

	m := f.Strings + f.Bools + f.Numbers + f.Aliens + f.Nils + f.Vars

	if 0 < d {
		m += f.Arrays + f.Maps
	}

	t := rand.Float64() * m
	if t < f.Strings {
		return f.genString(r)
	} else if t < f.Strings+f.Bools {
		return f.genBool(r)
	} else if t < f.Strings+f.Bools+f.Numbers {
		return f.genNumber(r)
	} else if t < f.Strings+f.Bools+f.Numbers+f.Aliens {
		return struct{}{}
	} else if t < f.Strings+f.Bools+f.Numbers+f.Aliens+f.Nils {
		return nil
	} else if t < f.Strings+f.Bools+f.Numbers+f.Aliens+f.Nils+f.Vars {
		return f.genVar(r)
	} else if t < f.Strings+f.Bools+f.Numbers+f.Aliens+f.Nils+f.Vars+f.Arrays {
		return f.genArray(r, d)
	} else {
		return f.genMap(r, d)
	}
}

func (f *Fuzz) genProp(r *rand.Rand) string {
	if r.Float64() < f.VarProps {
		return f.genVar(r)
	}
	return f.genString(r)
}

func (f *Fuzz) genString(r *rand.Rand) string {
	n := r.Intn(f.StringWidth-1) + 1
	s := make([]byte, n)
	for i, _ := range s {
		s[i] = f.Alphabet[r.Intn(len(f.Alphabet))]
	}
	return string(s)
}

func (f *Fuzz) genVar(r *rand.Rand) string {
	n := r.Intn(f.VarWidth-1) + 1
	s := make([]byte, n)
	for i, _ := range s {
		s[i] = f.VarAlphabet[r.Intn(len(f.VarAlphabet))]
	}
	return "?" + string(s)
}

func (f *Fuzz) genBool(r *rand.Rand) interface{} {
	return r.Intn(1024) % 2
}

func (f *Fuzz) genNumber(r *rand.Rand) interface{} {
	return float64(r.Intn(int(f.MaxNumber)))
}

func (f *Fuzz) genArray(r *rand.Rand, d int) interface{} {
	xs := make([]interface{}, r.Intn(f.ArrayWidth))
	for i, _ := range xs {
		xs[i] = f.Gen(r, d)
	}
	return xs
}

func (f *Fuzz) genMap(r *rand.Rand, d int) interface{} {
	n := r.Intn(f.MapWidth)
	m := make(map[string]interface{}, n)
	for i := 0; i < n; i++ {
		m[f.genProp(r)] = f.Gen(r, d)
	}
	return m
}

// TestMatchFuzz matches a bunch of patterns against a bunch of messages messages.
//
// Verifies some of the results.
func TestMatchFuzz(t *testing.T) {
	var (
		pats       = 2000
		msgsPerPat = 2000
		// Doing more doesn't seem to increase coverage.

		d = 4
		r = rand.New(rand.NewSource(42))
		p = NewFuzz()
		m = NewFuzz()

		matched           = 0
		attempted         = 0
		errs              = 0
		nontrivialMatches = 0
		maxBindings       = 0
	)
	m.NoVars()

	then := time.Now()
	for i := 0; i < pats; i++ {
		pat := p.Gen(r, d)
		for j := 0; j < msgsPerPat; j++ {
			msg := m.Gen(r, d)
			bss, err := Match(pat, msg, NewBindings())
			attempted++
			if err != nil {
				errs++
			}
			if 0 < len(bss) {
				matched++
				if s, is := pat.(string); is && !DefaultMatcher.IsVariable(s) {
					nontrivialMatches++
					// Verify the matches.
					for _, bs := range bss {
						check, err := Match(pat, msg, bs)
						if err != nil {
							t.Fatal(err)
						}
						if len(check) != 1 {
							t.Fatal(check)
						}
						if len(check[0]) != 0 {
							t.Fatal(check[0])
						}
					}
				}
				if maxBindings < len(bss) {
					maxBindings = len(bss)
				}
			}
		}
	}
	elapsed := time.Now().Sub(then)

	fmt.Printf(`fuzzed      %d
matched     %f%%
nontrivial  %f%% (%d)
errors      %f%% (%d)
elapsed     %fms
maxBindings %d
generated   %d
`,
		attempted,
		100*float64(matched)/float64(attempted),
		100*float64(nontrivialMatches)/float64(attempted), nontrivialMatches,
		100*float64(errs)/float64(attempted), errs,
		elapsed.Seconds()*100,
		maxBindings,
		p.generated+m.generated)
}
