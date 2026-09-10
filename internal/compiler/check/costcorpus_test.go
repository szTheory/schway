package check

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
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
//
// Relocated in Phase 09 (D-09-46): the generator itself
// (generateCallGraphCorpus, corpusShapes, and every corpusXxx builder) now
// lives in internal/compiler/testsupport/callgraphcorpus.go as
// testsupport.GenerateCallGraphCorpus / testsupport.CallGraphCorpusShapes,
// so session's and corevalidate's own test binaries can reach the same
// generator without granting either a production import of check. The
// sweep tests below are unchanged and delegate to that relocated
// generator.
// ---------------------------------------------------------------------

// corpusShapes names exactly the five call-graph shapes this gate sweeps,
// delegating to testsupport.CallGraphCorpusShapes() (relocated, D-09-46).
var corpusShapes = testsupport.CallGraphCorpusShapes()

// generateCallGraphCorpus delegates to testsupport.GenerateCallGraphCorpus
// (relocated, D-09-46).
func generateCallGraphCorpus(shape string, n int) (core.Program, error) {
	return testsupport.GenerateCallGraphCorpus(shape, n)
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
	// corpusLeafUse/corpusLeafPass/corpusRelay are unexported inside
	// testsupport after the D-09-46 relocation, so this test constructs the
	// exact same two shapes directly against core's own types -- byte-for-
	// byte identical to what testsupport's private builders would have
	// produced for these same IDs, never delegating to the generator's own
	// unexported internals from outside its package.
	usingLeaf := core.Function{
		ID: "twin:leaf", Name: "twin:leaf", Parameter: core.Parameter{ID: "twin:leaf:place:0", Name: "v", Type: "Byte"},
		Linear: &core.LinearBody{ID: "twin:leaf:linear", Operations: []core.LinearOperation{
			{ID: "twin:leaf:op:0", Kind: core.OpCopy, SourceID: "twin:leaf:place:0", TargetID: "twin:leaf:place:1"},
			{ID: "twin:leaf:op:1", Kind: core.OpReturn, SourceID: "twin:leaf:place:1"},
		}},
	}
	passingLeaf := core.Function{
		ID: "twin:leaf", Name: "twin:leaf", Parameter: core.Parameter{ID: "twin:leaf:place:0", Name: "v", Type: "Byte"},
		Linear: &core.LinearBody{ID: "twin:leaf:linear", Operations: []core.LinearOperation{
			{ID: "twin:leaf:op:0", Kind: core.OpReturn, SourceID: "twin:leaf:place:0"},
		}},
	}

	for _, tc := range []struct {
		name string
		leaf core.Function
		want bool
	}{
		{"callee uses its parameter", usingLeaf, true},
		{"callee merely forwards its parameter", passingLeaf, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			caller := core.Function{
				ID: "twin:caller", Name: "twin:caller", Parameter: core.Parameter{ID: "twin:caller:place:0", Name: "v", Type: "Byte"},
				Linear: &core.LinearBody{ID: "twin:caller:linear", Operations: []core.LinearOperation{
					{ID: "twin:caller:op:call0", Kind: core.OpCall, SourceID: "twin:caller:place:0", TargetID: "twin:caller:place:r0", CalleeID: "twin:leaf"},
					{ID: "twin:caller:op:ret", Kind: core.OpReturn, SourceID: "twin:caller:place:r0"},
				}},
			}
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
