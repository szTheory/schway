package testsupport

import (
	"fmt"

	"github.com/szTheory/schway/internal/compiler/core"
)

// ---------------------------------------------------------------------
// Phase 08 Plan 05, Task 1 (D-08-34/D-08-35): the synthetic call-graph
// corpus, the growth-exponent fit, and the <= 1.2 assertion. This file
// builds core.Program/core.Function/core.LinearOperation values DIRECTLY,
// exactly like TestSummaryDerivationTwoHopChainPropagates and its 08-02
// siblings above -- it never generates .schway source text and never touches
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
// Relocated here in Phase 09 (D-09-46): the generator originally lived
// unexported inside the "check" package's own cost-corpus test file, which
// made it structurally uncallable from session's or corevalidate's own test
// binaries. It is relocated verbatim (never duplicated) so every consumer
// that needs the same five-shape corpus can reach exactly one copy, without
// granting session or corevalidate a production import of the checker --
// testsupport's only non-stdlib import is internal/compiler/core.
// ---------------------------------------------------------------------

// CallGraphCorpusShapes names exactly the five call-graph shapes this gate
// sweeps. "tree" is deliberately absent (spike S-006 measured it, but it is
// dominated by diamond for sharing and chain for depth, adding sweep cost
// with no distinct argument of its own -- planner discretion recorded in
// 08-05-PLAN.md's Task 1(b)). Returns a fresh copy on every call so a
// consumer cannot mutate the package-level slice out from under another.
func CallGraphCorpusShapes() []string {
	shapes := []string{"chain", "diamond", "dense", "parser-shaped", "forward"}
	out := make([]string, len(shapes))
	copy(out, shapes)
	return out
}

// GenerateCallGraphCorpus builds a synthetic core.Program of the given
// shape with approximately n functions, directly against core's own types
// -- never through the parser (D-08-34).
func GenerateCallGraphCorpus(shape string, n int) (core.Program, error) {
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
