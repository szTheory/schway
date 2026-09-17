// Command iplive is S-006's decision harness.
//
// Ordinary run: it checks both mechanisms against the expansion oracle on the
// hand-written fixtures and on small generated graphs, sweeps cost across
// call-graph shapes and sizes, probes cache-invalidation fallout, and prints a
// compact JSON decision summary. Full per-fixture detail is retained for a
// mismatch or for an explicit -detail request rather than charging every
// healthy run for it.
//
// Injected-fault runs (-inject-summary-drop, -inject-stale-cache) are expected
// to exit unsuccessfully after reporting the first disagreement.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"ai-lang/interprocedural-liveness-cost/iplive"
)

type report struct {
	Schema        string                     `json:"schema"`
	Gate          string                     `json:"gate"`
	Agreement     []iplive.Agreement         `json:"agreement"`
	Disagreements []string                   `json:"disagreements,omitempty"`
	Scaling       []iplive.Series            `json:"scaling"`
	Invalidation  []iplive.InvalidationProbe `json:"invalidation"`
	BodyOrder     []iplive.BodyOrder         `json:"body_order"`
	Verdict       string                     `json:"verdict"`
}

func main() {
	pretty := flag.Bool("pretty", true, "indent the JSON report")
	detail := flag.Bool("detail", false, "retain per-case agreement detail in the report")
	dropSummary := flag.Bool("inject-summary-drop", false, "arm fault seam 1: summaries forget a returned borrow of the parameter")
	staleCache := flag.Bool("inject-stale-cache", false, "arm fault seam 2: the summary cache is keyed by function ID alone")
	budget := flag.Int("budget", 4000000, "work budget for the unmemoized arm")
	oracleBudget := flag.Int("oracle-budget", 400000, "expanded-operation budget for the context-sensitive oracle")
	flag.Parse()

	result := report{
		Schema: "s006-iplive/1",
		Gate:   "Does summary-based interprocedural loan liveness stay linear/sub-quadratic in call-graph size, or must a memoized summary cache be designed in from the start?",
	}

	cases := map[string]*iplive.Program{}
	for name, program := range iplive.Fixtures() {
		cases["fixture:"+name] = program
	}
	for _, shape := range []iplive.Shape{iplive.ShapeChain, iplive.ShapeTree, iplive.ShapeDiamond, iplive.ShapeDense, iplive.ShapeParser} {
		program, err := iplive.Generate(shape, 12)
		if err != nil {
			fail(err)
		}
		cases["generated:"+string(shape)] = program
	}

	names := make([]string, 0, len(cases))
	for name := range cases {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		program := cases[name]
		providers := []iplive.Provider{}

		memo := iplive.NewMemoProvider(program, 0)
		if *dropSummary {
			memo.InjectSummaryDrop()
		}
		if *staleCache {
			memo.InjectStaleKeys()
		}
		providers = append(providers, memo)

		recompute := iplive.NewRecomputeProvider(program, *budget)
		if *dropSummary {
			recompute.InjectSummaryDrop()
		}
		providers = append(providers, recompute)

		for _, provider := range providers {
			agreement, err := iplive.CheckAgainstOracle(program, provider, *oracleBudget)
			if err != nil {
				fail(fmt.Errorf("%s/%s: %w", name, provider.Name(), err))
			}
			if !agreement.Agreed {
				for _, diff := range append(agreement.ConflictDiffs, agreement.LivenessDiffs...) {
					result.Disagreements = append(result.Disagreements, name+"/"+provider.Name()+": "+diff)
				}
			}
			if *detail || !agreement.Agreed {
				agreement.Mechanism = name + "/" + agreement.Mechanism
				result.Agreement = append(result.Agreement, agreement)
			}
		}
	}

	// A stale ID-keyed cache only misbehaves after an edit, so the harness
	// replays the edit explicitly rather than hoping a fresh run exposes it.
	if *staleCache {
		result.Disagreements = append(result.Disagreements, staleEditProbe(*oracleBudget)...)
	}

	sweeps := []struct {
		shape iplive.Shape
		sizes []int
	}{
		{iplive.ShapeChain, []int{8, 16, 32, 64, 128, 256, 512}},
		{iplive.ShapeTree, []int{15, 31, 63, 127, 255}},
		{iplive.ShapeDiamond, []int{8, 12, 16, 20, 24, 28}},
		{iplive.ShapeDense, []int{9, 12, 15, 18, 21}},
		{iplive.ShapeParser, []int{8, 16, 32, 64, 128, 256, 512}},
		{iplive.ShapeForward, []int{8, 16, 32, 64, 128, 256, 512}},
	}
	for _, sweep := range sweeps {
		series, err := iplive.Sweep(sweep.shape, sweep.sizes, *budget)
		if err != nil {
			fail(fmt.Errorf("sweep %s: %w", sweep.shape, err))
		}
		result.Scaling = append(result.Scaling, series)
	}

	for _, shape := range []iplive.Shape{iplive.ShapeChain, iplive.ShapeTree, iplive.ShapeDense, iplive.ShapeParser} {
		program, err := iplive.Generate(shape, 128)
		if err != nil {
			fail(err)
		}
		probe, err := iplive.ProbeInvalidation(shape, program)
		if err != nil {
			fail(err)
		}
		result.Invalidation = append(result.Invalidation, probe)
	}

	for _, k := range []int{8, 128, 512} {
		probe, err := iplive.ProbeBodyOrder(k)
		if err != nil {
			fail(err)
		}
		result.BodyOrder = append(result.BodyOrder, probe)
	}

	result.Verdict = "agreed"
	if len(result.Disagreements) > 0 {
		result.Verdict = "DISAGREEMENT"
	}

	encoder := json.NewEncoder(os.Stdout)
	if *pretty {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(result); err != nil {
		fail(err)
	}
	if len(result.Disagreements) > 0 {
		os.Exit(1)
	}
}

// staleEditProbe warms a cache, edits one leaf, invalidates under the armed
// policy, and reports whether the answer survived the edit.
func staleEditProbe(oracleBudget int) []string {
	program := iplive.Fixtures()["transitive_relay_conflicts"]
	mutated := iplive.MutateLeafToUse(program, "leaf")
	provider := iplive.NewMemoProvider(program, 0)
	provider.InjectStaleKeys()
	if _, err := iplive.Analyze(program, provider); err != nil {
		fail(err)
	}
	provider.Rebind(mutated)
	provider.Invalidate(mutated, "leaf")
	agreement, err := iplive.CheckAgainstOracle(mutated, provider, oracleBudget)
	if err != nil {
		fail(err)
	}
	var out []string
	for _, diff := range append(agreement.ConflictDiffs, agreement.LivenessDiffs...) {
		out = append(out, "stale-edit-probe/memoized: "+diff)
	}
	return out
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "iplive:", err)
	os.Exit(2)
}
