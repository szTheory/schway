package iplive

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

const oracleBudget = 400000

// expectedConflicts states, by hand, what each fixture must refuse. Neither
// mechanism nor the oracle produced these strings.
var expectedConflicts = map[string][]string{
	"returns_borrow_conflicts":   {"caller|caller:move|l"},
	"owned_return_legal":         nil,
	"uses_param_conflicts":       {"caller|caller:move|l"},
	"unused_param_legal":         nil,
	"transitive_relay_conflicts": {"caller|caller:move|l"},
	"branch_arm_conflicts":       {"caller|caller:then:move|l"},
}

func conflictKeys(conflicts []Conflict) []string {
	keys := make([]string, 0, len(conflicts))
	for _, conflict := range conflicts {
		keys = append(keys, conflict.key())
	}
	sort.Strings(keys)
	return keys
}

func TestFixturesMatchHandStatedRefusals(t *testing.T) {
	for name, program := range Fixtures() {
		t.Run(name, func(t *testing.T) {
			analysis, err := Analyze(program, NewMemoProvider(program, 0))
			if err != nil {
				t.Fatalf("analyze: %v", err)
			}
			got := conflictKeys(analysis.Conflicts)
			want := expectedConflicts[name]
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Fatalf("conflicts: got %v want %v", got, want)
			}
		})
	}
}

func TestFixturesAgreeWithExpansionOracle(t *testing.T) {
	for name, program := range Fixtures() {
		t.Run(name, func(t *testing.T) {
			for _, provider := range []Provider{NewMemoProvider(program, 0), NewRecomputeProvider(program, 0)} {
				agreement, err := CheckAgainstOracle(program, provider, oracleBudget)
				if err != nil {
					t.Fatalf("%s: %v", provider.Name(), err)
				}
				if !agreement.Agreed {
					t.Fatalf("%s disagreed with oracle: conflicts=%v liveness=%v",
						provider.Name(), agreement.ConflictDiffs, agreement.LivenessDiffs)
				}
			}
		})
	}
}

func TestGeneratedShapesAgreeWithExpansionOracle(t *testing.T) {
	cases := []struct {
		shape Shape
		size  int
	}{
		{ShapeChain, 12},
		{ShapeTree, 15},
		{ShapeDiamond, 10},
		{ShapeDense, 9},
		{ShapeParser, 14},
	}
	for _, testCase := range cases {
		t.Run(string(testCase.shape), func(t *testing.T) {
			program, err := Generate(testCase.shape, testCase.size)
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			for _, provider := range []Provider{NewMemoProvider(program, 0), NewRecomputeProvider(program, 0)} {
				agreement, err := CheckAgainstOracle(program, provider, oracleBudget)
				if err != nil {
					t.Fatalf("%s: %v", provider.Name(), err)
				}
				if !agreement.Agreed {
					t.Fatalf("%s disagreed with oracle on %s: conflicts=%v liveness=%v",
						provider.Name(), testCase.shape, agreement.ConflictDiffs, agreement.LivenessDiffs)
				}
				if agreement.OracleConflict == 0 {
					t.Fatalf("%s: shape %s produced no refusals at all -- the corpus would not detect a summary defect",
						provider.Name(), testCase.shape)
				}
			}
		})
	}
}

// --- fault seams -----------------------------------------------------------

func TestSummaryDropInjectionIsCaught(t *testing.T) {
	program := Fixtures()["transitive_relay_conflicts"]
	provider := NewMemoProvider(program, 0)
	provider.InjectSummaryDrop()
	agreement, err := CheckAgainstOracle(program, provider, oracleBudget)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if agreement.Agreed {
		t.Fatal("injected summary drop went undetected: the harness cannot prove anything")
	}
	found := false
	for _, diff := range agreement.ConflictDiffs {
		if strings.Contains(diff, "only-in-second: caller|caller:move|l") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected the dropped returned-borrow refusal to be missing, got %v", agreement.ConflictDiffs)
	}
}

func TestSummaryDropIsCaughtOnBothArms(t *testing.T) {
	program := Fixtures()["returns_borrow_conflicts"]
	provider := NewRecomputeProvider(program, 0)
	provider.InjectSummaryDrop()
	agreement, err := CheckAgainstOracle(program, provider, oracleBudget)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if agreement.Agreed {
		t.Fatal("injected summary drop went undetected on the unmemoized arm")
	}
}

func TestStaleCacheInjectionIsCaught(t *testing.T) {
	program := Fixtures()["transitive_relay_conflicts"]
	mutated := MutateLeafToUse(program, "leaf")

	stale := NewMemoProvider(program, 0)
	stale.InjectStaleKeys()
	if _, err := Analyze(program, stale); err != nil {
		t.Fatalf("warm: %v", err)
	}
	stale.Rebind(mutated)
	stale.Invalidate(mutated, "leaf")
	staleAgreement, err := CheckAgainstOracle(mutated, stale, oracleBudget)
	if err != nil {
		t.Fatalf("stale: %v", err)
	}
	if staleAgreement.Agreed {
		t.Fatal("an ID-keyed cache served a stale caller summary and nothing noticed")
	}

	sound := NewMemoProvider(program, 0)
	if _, err := Analyze(program, sound); err != nil {
		t.Fatalf("warm: %v", err)
	}
	sound.Rebind(mutated)
	evicted := sound.Invalidate(mutated, "leaf")
	if evicted != 4 {
		t.Fatalf("transitive invalidation evicted %d summaries, want 4 (leaf + three callers)", evicted)
	}
	soundAgreement, err := CheckAgainstOracle(mutated, sound, oracleBudget)
	if err != nil {
		t.Fatalf("sound: %v", err)
	}
	if !soundAgreement.Agreed {
		t.Fatalf("caller-invalidating cache disagreed with oracle: %v %v",
			soundAgreement.ConflictDiffs, soundAgreement.LivenessDiffs)
	}
}

// TestInertSummaryDefectEscapesByDesign is this spike's explicitly expected
// escape. A summary defect in a function no caller ever probes with a move is
// semantically inert on this corpus: nothing observable changes, so no oracle
// can catch it. The claim this workbench makes is therefore bounded -- it
// prices mechanisms and catches OBSERVABLE drift, and says nothing about
// summary facts that no admission decision consumes.
func TestInertSummaryDefectEscapesByDesign(t *testing.T) {
	program := &Program{
		Roots: []string{"caller"},
		Functions: []Function{
			relay("caller", "", "callee"),
			leafPass("callee"),
		},
	}
	provider := NewMemoProvider(program, 0)
	provider.InjectSummaryDrop()
	agreement, err := CheckAgainstOracle(program, provider, oracleBudget)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if !agreement.Agreed {
		t.Fatal("expected the inert defect to escape; if this now fails, the corpus grew a probe and the README's escape note must be revised")
	}
}

// --- structural refusals ---------------------------------------------------

func TestCallGraphCycleRefusedByEveryConsumer(t *testing.T) {
	program := &Program{
		Roots: []string{"a"},
		Functions: []Function{
			relay("a", "", "b"),
			relay("b", "", "a"),
		},
	}
	if _, err := program.ReversePostorder(); err == nil {
		t.Fatal("expected the call graph order to refuse a cycle")
	}
	if _, err := Analyze(program, NewMemoProvider(program, 0)); err == nil {
		t.Fatal("expected the memoized arm to refuse a cycle")
	}
	if _, err := Analyze(program, NewRecomputeProvider(program, 5000000)); err == nil {
		t.Fatal("expected the unmemoized arm to refuse a cycle")
	}
	if _, err := Oracle(program, oracleBudget); err == nil {
		t.Fatal("expected the oracle to refuse a cycle")
	}
}

func TestDeepChainDoesNotExhaustTheStack(t *testing.T) {
	program := chain(20000)
	if _, err := program.ReversePostorder(); err != nil {
		t.Fatalf("deep chain order: %v", err)
	}
	if _, err := Analyze(program, NewMemoProvider(program, 0)); err != nil {
		t.Fatalf("deep chain analysis: %v", err)
	}
}

// --- cost -----------------------------------------------------------------

const sweepBudget = 4000000

// TestMemoizedArmIsLinearOnEveryShape is the gate's first half: one derivation
// per function, and total work linear in program size on every call-graph
// shape, including the sharing-heavy ones.
func TestMemoizedArmIsLinearOnEveryShape(t *testing.T) {
	cases := []struct {
		shape Shape
		sizes []int
	}{
		{ShapeChain, []int{8, 64, 512}},
		{ShapeTree, []int{15, 63, 255}},
		{ShapeDiamond, []int{9, 17, 29}},
		{ShapeDense, []int{10, 16, 22}},
		{ShapeParser, []int{8, 64, 512}},
	}
	for _, testCase := range cases {
		t.Run(string(testCase.shape), func(t *testing.T) {
			series, err := Sweep(testCase.shape, testCase.sizes, sweepBudget)
			if err != nil {
				t.Fatalf("sweep: %v", err)
			}
			memo := series.Growth[0]
			if memo.Class != "linear" {
				t.Fatalf("memoized arm is %q (exp_ops=%s), not linear", memo.Class, memo.ExponentInOps)
			}
			for _, point := range series.Points {
				if point.Memo.Derivations > point.Functions {
					t.Fatalf("n=%d: %d derivations for %d functions -- a summary was derived twice",
						point.Functions, point.Memo.Derivations, point.Functions)
				}
			}
		})
	}
}

// TestUnmemoizedArmsDoNotScale is the gate's second half. The naive arm blows a
// four-million-unit budget on shapes of a few dozen functions; even the
// charitable arm (a scratch cache that survives one caller's analysis) is
// quadratic in program size.
func TestUnmemoizedArmsDoNotScale(t *testing.T) {
	series, err := Sweep(ShapeParser, []int{8, 16, 32, 64, 128, 256, 512}, sweepBudget)
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	scratch, recompute := series.Growth[1], series.Growth[2]
	if scratch.Class == "linear" {
		t.Fatalf("charitable unmemoized arm measured linear (exp_ops=%s) -- the cost case for a cache is gone", scratch.ExponentInOps)
	}
	if recompute.Class != "unbounded (work budget exhausted before the largest size)" {
		t.Fatalf("naive unmemoized arm measured %q, expected it to exhaust its budget", recompute.Class)
	}
	largest := series.Points[len(series.Points)-1]
	ratio := float64(largest.Scratch.Work) / float64(largest.Memo.Work)
	if ratio < 10 {
		t.Fatalf("at %d functions the charitable arm costs only %.1fx the memoized one", largest.Functions, ratio)
	}
	t.Logf("at %d functions: memo=%d scratch=%d (%.0fx)", largest.Functions, largest.Memo.Work, largest.Scratch.Work, ratio)
}

// TestInvalidationFalloutIsShapeDependent measures what a cache COSTS once it
// exists: how much of it one callee edit must throw away.
func TestInvalidationFalloutIsShapeDependent(t *testing.T) {
	for _, shape := range []Shape{ShapeChain, ShapeTree, ShapeParser} {
		program, err := Generate(shape, 128)
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		probe, err := ProbeInvalidation(shape, program)
		if err != nil {
			t.Fatalf("probe: %v", err)
		}
		if probe.MaxEvicted < 1 {
			t.Fatalf("%s: an edit evicted nothing at all", shape)
		}
		t.Logf("%s: worst edit evicts %d/%d (%.0f%%), mean %.1f",
			shape, probe.MaxEvicted, probe.Functions, probe.MaxEvictedFrac*100, probe.MeanEvicted)
	}
	parser, _ := Generate(ShapeParser, 128)
	probe, _ := ProbeInvalidation(ShapeParser, parser)
	if probe.MaxEvictedFrac < 0.5 {
		t.Fatalf("parser-shaped worst-case eviction fell to %.2f; the cross-run caching caveat in the README must be revised", probe.MaxEvictedFrac)
	}
}
