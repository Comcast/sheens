/* Copyright 2018-2019 Comcast Cable Communications Management, LLC
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


// Package main is a little command-line utility to invoke pattern matching.
//
//   patmatch -p '{"likes":"?liked"}' -m '{"likes":["tacos","chips"]}' -w '[{"?liked":["tacos","chipss"]}]'
//
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"time"

	"github.com/Comcast/sheens/match"
)

func main() {
	var (
		messageJS  = flag.String("m", "", "message in JSON")
		patternJS  = flag.String("p", "", "pattern in JSON")
		bindingsJS = flag.String("b", "{}", "bindings in JSON")
		wantJS     = flag.String("w", "", "wanted bindings in JSON")

		bench = flag.Int("bench", 0, "number of times to run (and report time)")

		verbose = flag.Bool("v", false, "verbosity")

		message  interface{}
		pattern  interface{}
		want     []match.Bindings
		wanted   bool
		bindings match.Bindings
	)

	flag.Parse()

	if *messageJS != "" {
		if err := json.Unmarshal([]byte(*messageJS), &message); err != nil {
			panic(err)
		}
	}

	if *patternJS != "" {
		if err := json.Unmarshal([]byte(*patternJS), &pattern); err != nil {
			panic(err)
		}
	}

	if *bindingsJS != "" {
		if err := json.Unmarshal([]byte(*bindingsJS), &bindings); err != nil {
			panic(err)
		}
	}

	if *wantJS != "" {
		if err := json.Unmarshal([]byte(*wantJS), &want); err != nil {
			panic(err)
		}
		wanted = true
	}

	if 0 < *bench {
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		allocs := stats.TotalAlloc
		then := time.Now()
		for i := 0; i < *bench; i++ {
			if _, err := match.Match(pattern, message, bindings); err != nil {
				panic(err)
			}
		}
		elapsed := time.Now().Sub(then)
		meanNanos := elapsed.Nanoseconds() / int64(*bench)

		runtime.ReadMemStats(&stats)
		allocated := (stats.TotalAlloc - allocs) / uint64(*bench)

		log.Printf("%d iterations, %d mean ns/Match, %d mean bytes allocated per Match", *bench, meanNanos, allocated)
	}

	bss, err := match.Match(pattern, message, bindings)
	if err != nil {
		panic(err)
	}

	if wanted {
		// Loop over all wanted Bindings and check that each
		// appeared in what we got.
	WANTED:
		for _, wantedBs := range want {
			for _, haveBs := range bss {
				eq, err := Subset(wantedBs, haveBs, *verbose)
				if err != nil {
					panic(err)
				}
				if !eq {
					continue
				}
				eq, err = Subset(haveBs, wantedBs, *verbose)
				if eq {
					continue WANTED
				}
			}
			fmt.Printf("false\n")
			return
		}
		fmt.Printf("true\n")
		return
	}

	bssJS, err := json.Marshal(&bss)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", bssJS)
}

// Subset tries to check that Bindings x is a subset of Bindings y.
//
// Uses reflect.DeepEqual to do the hard work.
func Subset(x, y match.Bindings, verbose bool) (bool, error) {
	for p, bx := range x {
		by, have := y[p]
		if !have {
			return false, nil
		}
		if !reflect.DeepEqual(bx, by) {
			if verbose {
				xjs, err := json.Marshal(&bx)
				if err != nil {
					return false, err
				}
				yjs, err := json.Marshal(&by)
				if err != nil {
					return false, err
				}

				fmt.Printf("disagreement at %s: %s != %s", p, xjs, yjs)
			}
			return false, nil
		}
	}
	return true, nil
}
