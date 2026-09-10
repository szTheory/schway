package check

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/core"
)

// ---------------------------------------------------------------------
// Phase 08 Plan 05, Task 1 (D-08-34/D-08-35): the synthetic call-graph
// corpus, the growth-exponent fit, and the <= 1.2 assertion. This file
// builds core.Program/core.Function/core.LinearOperation values DIRECTLY,
// exactly like TestSummaryDerivationTwoHopChainPropagates and its 08-02
// siblings above -- it never generates .lang source text and never touches
// internal/compiler/syntax, because at the sizes this gate needs (hundreds
// of functions) parse time would very plausibly dominate and mask the
// exact curve this gate exists to see (D-08-34). TestCostCorpusIsNotParsed
// asserts that structurally.
//
// Every number this file quotes is buildInterproceduralSummaries' own
// deterministic work counter (one unit per summary derived, plus
// deriveFunctionUsesParam's own per-operation-inspected/per-insertion
// counting) -- never a timing. Wall-clock elapsed time is logged per sweep
// point for an incidental human reading this test's -v output, and is
// NEVER compared to anything or used to gate a verdict: a single
// laptop-class host with no CI fleet confounds battery state, P/E core
// scheduling, and thermal throttling into one noisy sample stream that
// cannot be trusted to see a super-linear curve at any corpus size a CI
// budget can afford (see this plan's own action text and
// wiki/compute-efficiency-constitution.md).
// ---------------------------------------------------------------------

// corpusShapes names exactly the five call-graph shapes this gate sweeps.
// "tree" is deliberately absent (spike S-006 measured it, but it is
// dominated by diamond for sharing and chain for depth, adding sweep cost
// with no distinct argument of its own -- planner discretion recorded in
// this plan's Task 1(b)).
var corpusShapes = []string{"chain", "diamond", "dense", "parser-shaped", "forward"}

// generateCallGraphCorpus builds a synthetic core.Program of the given
// shape with approximately n functions, directly against core's own types
// -- never through the parser (D-08-34).
func generateCallGraphCorpus(shape string, n int) (core.Program, error) {
	switch shape {
	case "chain":
		return corpusChain(n), nil
	case "diamond":
		return corpusLayered("diamond", n, 2), nil
	case "dense":
		return corpusLayered("dense", n, 3), nil
	case "parser-shaped":
		return corpusParserShaped(n), nil
	case "forward":
		return corpusForward(n), nil
	}
	return core.Program{}, fmt.Errorf("unknown corpus shape %q", shape)
}

// corpusRelay builds a function that forwards its own parameter into every
// named callee (an OpCall per callee, SourceID always the function's own
// parameter place -- never chained through a prior call's result) and
// returns the first callee's result (or its own parameter, if it has no
// callees). This is the corpus's one call-bearing template; every shape
// below composes programs entirely out of this, corpusLeafUse, and
// corpusLeafPass.
func corpusRelay(id string, callees ...string) core.Function {
	paramID := id + ":place:0"
	parameter := core.Parameter{ID: paramID, Name: "v", Type: "Byte"}
	operations := make([]core.LinearOperation, 0, len(callees)+1)
	for index, callee := range callees {
		result := fmt.Sprintf("%s:place:r%d", id, index)
		operations = append(operations, core.LinearOperation{
			ID: fmt.Sprintf("%s:op:call%d", id, index), Kind: core.OpCall,
			SourceID: paramID, TargetID: result, CalleeID: callee,
		})
	}
	returnSource := paramID
	if len(callees) > 0 {
		returnSource = fmt.Sprintf("%s:place:r0", id)
	}
	operations = append(operations, core.LinearOperation{ID: id + ":op:ret", Kind: core.OpReturn, SourceID: returnSource})
	return core.Function{
		ID: id, Name: id, Parameter: parameter,
		Linear: &core.LinearBody{ID: id + ":linear", Operations: operations},
	}
}

// corpusLeafUse builds a function whose body genuinely reads through its
// own parameter (an OpCopy), so its own usesParam bit is true and any
// caller reaching it transitively inherits that bit.
func corpusLeafUse(id string) core.Function {
	paramID := id + ":place:0"
	return core.Function{
		ID: id, Name: id, Parameter: core.Parameter{ID: paramID, Name: "v", Type: "Byte"},
		Linear: &core.LinearBody{ID: id + ":linear", Operations: []core.LinearOperation{
			{ID: id + ":op:0", Kind: core.OpCopy, SourceID: paramID, TargetID: id + ":place:1"},
			{ID: id + ":op:1", Kind: core.OpReturn, SourceID: id + ":place:1"},
		}},
	}
}

// corpusLeafPass builds a function that merely hands its own parameter
// back on its terminating OpReturn -- D-08-01's leaf-forwards-to-its-own-
// return case, which is NOT a use, so its own usesParam bit is false.
func corpusLeafPass(id string) core.Function {
	paramID := id + ":place:0"
	return core.Function{
		ID: id, Name: id, Parameter: core.Parameter{ID: paramID, Name: "v", Type: "Byte"},
		Linear: &core.LinearBody{ID: id + ":linear", Operations: []core.LinearOperation{
			{ID: id + ":op:0", Kind: core.OpReturn, SourceID: paramID},
		}},
	}
}

// corpusChain builds the pure-repetition, zero-sharing shape (D-08-35):
// f0 -> f1 -> ... -> f(n-1), a leaf. Required because the naive charitable
// arm in spike S-006 died on a chain at 32 functions with NO sharing at
// all -- repetition alone compounds per level. Without this shape a corpus
// cannot separate "exponential because of repetition" from "quadratic
// because of sharing", which is exactly what the diamond/dense shapes
// below test for instead.
func corpusChain(n int) core.Program {
	if n < 1 {
		n = 1
	}
	functions := make([]core.Function, 0, n)
	for i := 0; i < n-1; i++ {
		functions = append(functions, corpusRelay(fmt.Sprintf("chain:f%d", i), fmt.Sprintf("chain:f%d", i+1)))
	}
	functions = append(functions, corpusLeafPass(fmt.Sprintf("chain:f%d", n-1)))
	return core.Program{Functions: functions}
}

// corpusLayered builds the sharing-heavy shapes (D-08-35): width-`width`
// layers where every node calls `width` nodes of the next layer. The
// number of distinct call-graph PATHS is width^layers while the number of
// functions is only width*layers -- the exact gap a memoized derivation
// (one summary per function, reused at every call site) collapses and an
// unmemoized recomputation would pay in full. width=2 is "diamond", the
// minimal shared-callee shape: without it a gate cannot distinguish an
// analysis that works when every callee has exactly one caller from one
// that genuinely reuses a per-function summary across callers, which IS
// the load-bearing claim this whole phase rests on. width=3 is "dense":
// spike S-006 measured it at 1.50 against function count and 1.11 against
// operation count, the gap being entirely the corpus's leaf-to-relay ratio
// drifting with size -- exactly what makes fitting against operation count
// (rather than function count) demonstrably necessary, not merely
// "the honest axis" in the abstract.
func corpusLayered(prefix string, n, width int) core.Program {
	layers := n / width
	if layers < 2 {
		layers = 2
	}
	id := func(layer, index int) string { return fmt.Sprintf("%s:l%dn%d", prefix, layer, index) }

	functions := []core.Function{corpusRelay(prefix+":f0", id(0, 0))}
	for layer := 0; layer < layers; layer++ {
		for index := 0; index < width; index++ {
			nodeID := id(layer, index)
			if layer == layers-1 {
				if index%2 == 0 {
					functions = append(functions, corpusLeafPass(nodeID))
				} else {
					functions = append(functions, corpusLeafUse(nodeID))
				}
				continue
			}
			callees := make([]string, 0, width)
			for offset := 0; offset < width; offset++ {
				callees = append(callees, id(layer+1, (index+offset)%width))
			}
			functions = append(functions, corpusRelay(nodeID, callees...))
		}
	}
	return core.Program{Functions: functions}
}

// corpusParserShaped builds a seeded, deterministic mixed DAG with depth
// ~8 and fan-out 2-4, roughly a third of it leaves toward the tail --
// criterion 3's named "realistic fan-out" shape (EFF-02), and also
// D-08-35's dense/parser-shaped bullet: its leaf-to-relay ratio drifts with
// size exactly like the layered "dense" shape, which is what makes fitting
// against operation count (not function count) demonstrably necessary --
// spike S-006 found fitting only against function count "would have
// produced a wrong finding" on this exact shape family. The generator uses
// a fixed-seed linear congruential sequence (never math/rand's global
// state, never time-seeded) so the corpus -- and therefore the fitted
// exponent -- is byte-for-byte reproducible across runs.
func corpusParserShaped(n int) core.Program {
	if n < 1 {
		n = 1
	}
	functions := make([]core.Function, 0, n)
	seed := uint64(0x5EED_5006)
	next := func(mod int) int {
		seed = seed*6364136223846793005 + 1442695040888963407
		return int((seed >> 33) % uint64(mod))
	}
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("parser:f%d", i)
		remaining := n - i - 1
		if remaining == 0 || (i > n/3 && next(10) == 0) {
			if i%2 == 0 {
				functions = append(functions, corpusLeafPass(id))
			} else {
				functions = append(functions, corpusLeafUse(id))
			}
			continue
		}
		window := remaining
		if window > 6 {
			window = 6
		}
		fanout := 2 + next(3)
		seen := map[string]bool{}
		var callees []string
		for k := 0; k < fanout; k++ {
			target := fmt.Sprintf("parser:f%d", i+1+next(window))
			if seen[target] {
				continue
			}
			seen[target] = true
			callees = append(callees, target)
		}
		if len(callees) == 0 {
			callees = append(callees, fmt.Sprintf("parser:f%d", i+1))
		}
		functions = append(functions, corpusRelay(id, callees...))
	}
	return core.Program{Functions: functions}
}

// corpusForward builds the shape that stresses the DERIVATION itself
// rather than the call graph (D-08-35): one caller whose body threads a
// single borrow through k successive calls, each call's SourceID being the
// PREVIOUS call's own result -- a star-shaped call graph of depth 1, whose
// every cost belongs to deriving ONE summary from one body. This isolates
// deriveFunctionUsesParam's own program-order invariant
// (check.go:650-667): without this shape a future pass that batches or
// reorders summary derivation could reintroduce a quadratic regression
// every other shape here is blind to, since every other shape's cost is
// dominated by call-graph traversal, not by any one function's body
// length. Spike S-006 measured this exact shape at a flat 4.0 work units
// per operation from k=8 to k=512 in program order, against 12.4 -> 767.0
// per operation for the identical chain listed in reverse -- a 192x
// penalty at k=512.
func corpusForward(n int) core.Program {
	callees := n - 1
	if callees < 1 {
		callees = 1
	}
	paramID := "f0:place:0"
	parameter := core.Parameter{ID: paramID, Name: "v", Type: "Byte"}
	operations := make([]core.LinearOperation, 0, callees+1)
	functions := make([]core.Function, 0, callees+1)
	previous := paramID
	for i := 0; i < callees; i++ {
		callee := fmt.Sprintf("f0:g%d", i)
		result := fmt.Sprintf("f0:place:r%d", i)
		operations = append(operations, core.LinearOperation{
			ID: fmt.Sprintf("f0:op:call%d", i), Kind: core.OpCall,
			SourceID: previous, TargetID: result, CalleeID: callee,
		})
		previous = result
		functions = append(functions, corpusLeafPass(callee))
	}
	operations = append(operations, core.LinearOperation{ID: "f0:op:ret", Kind: core.OpReturn, SourceID: previous})
	caller := core.Function{
		ID: "f0", Name: "f0", Parameter: parameter,
		Linear: &core.LinearBody{ID: "f0:linear", Operations: operations},
	}
	functions = append(functions, caller)
	return core.Program{Functions: functions}
}

// corpusOpCount sums every declared function's own operation count -- the
// honest axis (D-08-31/EFF-02) this plan fits the growth exponent against,
// rather than function count (spike S-006's own finding: a corpus whose
// leaf-to-relay ratio drifts with size moves the function-count fit
// without any mechanism changing).
func corpusOpCount(program core.Program) int {
	count := 0
	for _, function := range program.Functions {
		if function.Linear != nil {
			count += len(function.Linear.Operations)
		}
	}
	return count
}

// fitGrowthExponentInOps is Growth.ExponentInOps ported from
// .planning/spikes/006-interprocedural-liveness-cost-scaling/iplive/scaling.go,
// generalized from a two-point slope to an ordinary least-squares fit of
// log(work) against log(x) over every usable (positive, non-zero) point --
// more robust than a first/last-point slope across a six-point ladder.
// Returns NaN when fewer than two usable points exist (never a
// silently-fabricated 0, matching this project's honest-unavailability
// house style, measure/statistics.go's own precedent).
func fitGrowthExponentInOps(xs, ys []int) float64 {
	if len(xs) != len(ys) {
		return math.NaN()
	}
	var sumX, sumY, sumXY, sumXX float64
	usable := 0
	for i := range xs {
		x, y := xs[i], ys[i]
		if x <= 0 || y <= 0 {
			continue
		}
		logX, logY := math.Log(float64(x)), math.Log(float64(y))
		sumX += logX
		sumY += logY
		sumXY += logX * logY
		sumXX += logX * logX
		usable++
	}
	if usable < 2 {
		return math.NaN()
	}
	n := float64(usable)
	denominator := n*sumXX - sumX*sumX
	if denominator == 0 {
		return math.NaN()
	}
	return (n*sumXY - sumX*sumY) / denominator
}

// growthExponentMilliByShape records every fitted milli-exponent observed
// for a shape across repeated invocations of
// TestInterproceduralSummaryGrowthExponentInOps within ONE test binary
// (e.g. `go test -count=2`, which re-runs the test function in the same
// process): if a later invocation disagrees with the first, the
// measurement is not deterministic and the test fails immediately, rather
// than silently reporting only the last run's number.
var growthExponentMilliByShape = map[string][]int64{}

// interproceduralSizeLadder is the sweep this task's own action text
// requires be recorded with its rationale: it starts at the spike's own
// ladder (up to 512 functions) rather than a smaller one, because the
// ratio-stability tripwire below is the empirical signal for whether a
// smaller ladder would already have stabilized -- and per-shape work at
// 512 functions (well under a few thousand operations for every shape
// here) completes in low-single-digit milliseconds even for `go test
// -count=2 -shuffle=on`, so there is no latency pressure to shrink it. The
// ratio-stability tripwire compares work/operations at S=128 and 4S=512,
// both already present in this ladder -- no extra corpus size is
// generated solely for that comparison.
var interproceduralSizeLadder = []int{16, 32, 64, 128, 256, 512}

const (
	ratioStabilitySmall = 128
	ratioStabilityLarge = 512
)

// TestInterproceduralSummaryGrowthExponentInOps is this plan's Task 1
// falsifier (EFF-02/D-08-30/D-08-31): for every corpus shape, sweep
// interproceduralSizeLadder, drive buildInterproceduralSummaries' own
// deterministic work counter directly (never wall clock), fit the growth
// exponent against operation count, and assert it stays at or under the
// milli-exponent bound of 1200 (<= 1.2). The SAME corpus fitted against
// function count is computed too, purely for contrast in a failure
// message -- it is never itself a gated value (D-08-30's must-have:
// applying measure.Samples/Summary's existing protocol to work-count
// samples is the metric this gate is layered on top of; the fitted
// exponent is an ADDITIONAL hard control, not a replacement).
func TestInterproceduralSummaryGrowthExponentInOps(t *testing.T) {
	for _, shape := range corpusShapes {
		shape := shape
		t.Run(shape, func(t *testing.T) {
			operationCounts := make([]int, len(interproceduralSizeLadder))
			functionCounts := make([]int, len(interproceduralSizeLadder))
			workCounts := make([]int, len(interproceduralSizeLadder))

			for index, n := range interproceduralSizeLadder {
				program, err := generateCallGraphCorpus(shape, n)
				if err != nil {
					t.Fatalf("generateCallGraphCorpus(%q, %d): %v", shape, n, err)
				}
				start := time.Now()
				_, work := buildInterproceduralSummaries(program, callSignatureTable{})
				elapsed := time.Since(start)
				ops := corpusOpCount(program)

				// Wall clock is OBSERVED ONLY here -- logged for an
				// incidental human reader, never compared to anything and
				// never used to gate a verdict (D-08-29/T-08-20).
				t.Logf("shape=%s n=%d functions=%d ops=%d work=%d elapsed=%s (elapsed is observed-only, never gated)",
					shape, n, len(program.Functions), ops, work, elapsed)

				if work <= 0 {
					t.Fatalf("shape=%s n=%d: zero counted work from buildInterproceduralSummaries", shape, n)
				}
				operationCounts[index] = ops
				functionCounts[index] = len(program.Functions)
				workCounts[index] = work
			}

			exponentInOps := fitGrowthExponentInOps(operationCounts, workCounts)
			exponentInFuncs := fitGrowthExponentInOps(functionCounts, workCounts)
			if math.IsNaN(exponentInOps) {
				t.Fatalf("shape=%s: fitted exponent in ops is NaN (fewer than two usable sweep points)", shape)
			}
			milli := int64(math.Round(exponentInOps * 1000))

			growthExponentMilliByShape[shape] = append(growthExponentMilliByShape[shape], milli)
			if first := growthExponentMilliByShape[shape][0]; milli != first {
				t.Fatalf("shape=%s: fitted milli-exponent is not deterministic across repeated invocations: first=%d this=%d",
					shape, first, milli)
			}

			if milli > 1200 {
				t.Fatalf("shape=%s: fitted milli-exponent (against OPERATION count) = %d (exponent=%.4f) exceeds the bound 1200 (<= 1.2); "+
					"for contrast, the same corpus fitted against FUNCTION count = %.4f (this contrast value is never gated)",
					shape, milli, exponentInOps, exponentInFuncs)
			}

			var ratioAtSmall, ratioAtLarge float64
			var sawSmall, sawLarge bool
			for index, n := range interproceduralSizeLadder {
				switch n {
				case ratioStabilitySmall:
					ratioAtSmall = float64(workCounts[index]) / float64(operationCounts[index])
					sawSmall = true
				case ratioStabilityLarge:
					ratioAtLarge = float64(workCounts[index]) / float64(operationCounts[index])
					sawLarge = true
				}
			}
			if !sawSmall || !sawLarge {
				t.Fatalf("shape=%s: ratio-stability tripwire requires both S=%d and 4S=%d in interproceduralSizeLadder",
					shape, ratioStabilitySmall, ratioStabilityLarge)
			}
			delta := math.Abs(ratioAtLarge-ratioAtSmall) / ratioAtSmall
			if delta > 0.15 {
				t.Fatalf("shape=%s: work/operations ratio at S=%d (%.4f) vs 4S=%d (%.4f) differ by %.1f%%, exceeds the 15%% ratio-stability bound (D-08-31)",
					shape, ratioStabilitySmall, ratioAtSmall, ratioStabilityLarge, ratioAtLarge, delta*100)
			}
		})
	}
}

// TestGrowthExponentRoundingBoundary pins this plan's own tie-breaking
// contract (a planner decision, not a CONTEXT lock -- see 08-05-PLAN.md's
// "Planner assumptions" section): the gate compares
// int64(math.Round(exponent*1000)) > 1200, so 1.2 exactly and 1.2004 both
// pass (round to 1200), and 1.2006 fails (rounds to 1201). math.Round
// rounds half away from zero.
func TestGrowthExponentRoundingBoundary(t *testing.T) {
	cases := []struct {
		exponent float64
		wantPass bool
	}{
		{1.2, true},
		{1.2004, true},
		{1.2006, false},
	}
	for _, tc := range cases {
		milli := int64(math.Round(tc.exponent * 1000))
		pass := milli <= 1200
		if pass != tc.wantPass {
			t.Fatalf("exponent=%.4f -> milli=%d -> pass=%v, want pass=%v", tc.exponent, milli, pass, tc.wantPass)
		}
	}
}

// TestCostCorpusLeafTemplatesDifferOnlyInTheCallee proves the corpus
// generator's own templates are genuine discriminators, not decoration:
// two otherwise-identical relay functions whose sole difference is which
// leaf template their shared callee uses (corpusLeafUse vs corpusLeafPass)
// must derive DIFFERENT usesParam bits through buildInterproceduralSummaries
// -- exactly the twin-pair discipline 08-03's fixture corpus already
// established for the checker's own admission decisions (see that plan's
// own, differently-named TestInterproceduralTwinPairsDifferOnlyInTheCallee
// over real .lang fixtures), applied here to the cost corpus's own
// synthetic templates.
func TestCostCorpusLeafTemplatesDifferOnlyInTheCallee(t *testing.T) {
	usingLeaf := corpusLeafUse("twin:leaf")
	passingLeaf := corpusLeafPass("twin:leaf")

	for _, tc := range []struct {
		name string
		leaf core.Function
		want bool
	}{
		{"callee uses its parameter", usingLeaf, true},
		{"callee merely forwards its parameter", passingLeaf, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			caller := corpusRelay("twin:caller", "twin:leaf")
			program := core.Program{Functions: []core.Function{caller, tc.leaf}}
			summaries, work := buildInterproceduralSummaries(program, callSignatureTable{})
			if work <= 0 {
				t.Fatalf("expected positive counted work, got %d", work)
			}
			callerSummary, ok := summaries.lookup("twin:caller")
			if !ok {
				t.Fatal("expected a summary entry for twin:caller")
			}
			if callerSummary.usesParam != tc.want {
				t.Fatalf("twin:caller usesParam = %v, want %v (the pair must differ ONLY in the callee's own template)", callerSummary.usesParam, tc.want)
			}
		})
	}
}

// TestCostCorpusIsNotParsed is D-08-34's own structural guard: a go/ast
// import scan of this file asserting it imports neither
// internal/compiler/syntax nor os. At the sizes this gate needs (hundreds
// of functions), generating .lang source and running it through the real
// parser would very plausibly let parse time dominate and mask the exact
// curve this gate exists to see -- this test makes that prohibition
// mechanically enforced, not merely a doc comment.
func TestCostCorpusIsNotParsed(t *testing.T) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "costcorpus_test.go", nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parsing costcorpus_test.go for its own import scan: %v", err)
	}

	forbidden := map[string]bool{
		"github.com/codename-lang/lang/internal/compiler/syntax": true,
		"os": true,
	}

	var found []string
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.ImportSpec)
		if !ok {
			return true
		}
		path, unquoteErr := strconv.Unquote(spec.Path.Value)
		if unquoteErr != nil {
			t.Fatalf("could not unquote import path %s: %v", spec.Path.Value, unquoteErr)
		}
		if forbidden[path] {
			found = append(found, path)
		}
		return true
	})
	if len(found) != 0 {
		t.Fatalf("costcorpus_test.go imports forbidden package(s) %v -- the cost corpus must never be generated as .lang source through the real parser (D-08-34)", found)
	}
}
