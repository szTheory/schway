package iplive

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// Arm is one mechanism's cost at one measured size. Work is the deterministic,
// machine-independent counter (transfer evaluations, worklist reinsertions,
// summary consultations, per-op derivation visits). Nanos is measurement and
// is kept deliberately separate from the semantic counter.
type Arm struct {
	Work        int   `json:"work"`
	Derivations int   `json:"derivations"`
	Nanos       int64 `json:"nanos"`
	BudgetHit   bool  `json:"budget_hit,omitempty"`
}

// Point is one measured program size for one shape.
type Point struct {
	Shape     Shape `json:"shape"`
	Functions int   `json:"functions"`
	Ops       int   `json:"ops"`
	CallEdges int   `json:"call_edges"`
	Memo      Arm   `json:"memoized"`
	Scratch   Arm   `json:"recompute_scratch"`
	Recompute Arm   `json:"recompute"`
	Agreed    bool  `json:"arms_agreed"`
}

func (p Point) arm(name string) Arm {
	switch name {
	case "memoized":
		return p.Memo
	case "recompute+scratch":
		return p.Scratch
	}
	return p.Recompute
}

// Growth is one arm's empirical growth, fitted against both the call-graph
// size (function count) and the raw program size (operation count). The
// operation-count fit is the honest one for a claim about the ANALYZER: a
// corpus whose leaf-to-relay ratio drifts with size moves the function-count
// fit without any mechanism changing.
type Growth struct {
	Arm             string `json:"arm"`
	ExponentInFuncs string `json:"exponent_in_functions"`
	ExponentInOps   string `json:"exponent_in_ops"`
	Class           string `json:"class"`
	WorkPerOpFirst  string `json:"work_per_op_smallest"`
	WorkPerOpLast   string `json:"work_per_op_largest"`
}

// Series is one shape measured across sizes.
type Series struct {
	Shape  Shape    `json:"shape"`
	Points []Point  `json:"points"`
	Growth []Growth `json:"growth"`
}

// Measure runs all three arms over one program and records a point.
func Measure(shape Shape, p *Program, budget int) (Point, error) {
	point := Point{
		Shape: shape, Functions: len(p.Functions),
		Ops: p.OpCount(), CallEdges: p.CallEdgeCount(),
	}

	memo := NewMemoProvider(p, 0)
	memoAnalysis, arm, err := timedAnalyze(p, memo)
	if err != nil {
		return point, fmt.Errorf("memoized: %w", err)
	}
	point.Memo = arm

	scratch := NewRecomputeProvider(p, budget)
	scratch.EnableScratch()
	scratchAnalysis, arm, err := timedAnalyze(p, scratch)
	if err != nil && err != ErrBudget {
		return point, fmt.Errorf("recompute+scratch: %w", err)
	}
	point.Scratch = arm

	recompute := NewRecomputeProvider(p, budget)
	_, arm, err = timedAnalyze(p, recompute)
	if err != nil && err != ErrBudget {
		return point, fmt.Errorf("recompute: %w", err)
	}
	point.Recompute = arm

	point.Agreed = point.Scratch.BudgetHit ||
		len(DiffConflicts(memoAnalysis.Conflicts, scratchAnalysis.Conflicts)) == 0
	return point, nil
}

func timedAnalyze(p *Program, provider Provider) (Analysis, Arm, error) {
	start := time.Now()
	analysis, err := Analyze(p, provider)
	arm := Arm{
		Work: provider.Work(), Derivations: provider.Derivations(),
		Nanos: time.Since(start).Nanoseconds(),
	}
	if err == ErrBudget {
		arm.BudgetHit = true
		return Analysis{}, arm, ErrBudget
	}
	if err != nil {
		return Analysis{}, arm, err
	}
	arm.Work = analysis.Work
	return analysis, arm, nil
}

// Sweep measures one shape across the given sizes.
func Sweep(shape Shape, sizes []int, budget int) (Series, error) {
	series := Series{Shape: shape}
	for _, size := range sizes {
		program, err := Generate(shape, size)
		if err != nil {
			return series, err
		}
		point, err := Measure(shape, program, budget)
		if err != nil {
			return series, err
		}
		series.Points = append(series.Points, point)
	}
	for _, name := range []string{"memoized", "recompute+scratch", "recompute"} {
		series.Growth = append(series.Growth, fit(name, series.Points))
	}
	return series, nil
}

// fit computes one arm's growth over the largest prefix of measured sizes that
// the arm actually completed. An arm that exhausted its budget is reported as
// unbounded rather than extrapolated.
func fit(name string, points []Point) Growth {
	growth := Growth{Arm: name, ExponentInFuncs: "n/a", ExponentInOps: "n/a", Class: "not measurable"}
	var usable []Point
	budgetHit := false
	for _, point := range points {
		if point.arm(name).BudgetHit {
			budgetHit = true
			break
		}
		usable = append(usable, point)
	}
	if len(usable) >= 2 {
		first, last := usable[0], usable[len(usable)-1]
		firstArm, lastArm := first.arm(name), last.arm(name)
		inFuncs := slope(float64(first.Functions), float64(last.Functions), float64(firstArm.Work), float64(lastArm.Work))
		inOps := slope(float64(first.Ops), float64(last.Ops), float64(firstArm.Work), float64(lastArm.Work))
		growth.ExponentInFuncs = formatExponent(inFuncs)
		growth.ExponentInOps = formatExponent(inOps)
		growth.WorkPerOpFirst = fmt.Sprintf("%.1f", float64(firstArm.Work)/float64(first.Ops))
		growth.WorkPerOpLast = fmt.Sprintf("%.1f", float64(lastArm.Work)/float64(last.Ops))
		growth.Class = classify(inOps, false)
	}
	if budgetHit {
		growth.Class = "unbounded (work budget exhausted before the largest size)"
	}
	return growth
}

func slope(x1, x2, y1, y2 float64) float64 {
	if x1 <= 0 || x2 <= 0 || y1 <= 0 || y2 <= 0 || x1 == x2 {
		return math.NaN()
	}
	return math.Log(y2/y1) / math.Log(x2/x1)
}

func formatExponent(exp float64) string {
	if math.IsNaN(exp) || math.IsInf(exp, 0) {
		return "n/a"
	}
	return fmt.Sprintf("%.2f", exp)
}

func classify(exp float64, budgetHit bool) string {
	if budgetHit {
		return "unbounded (work budget exhausted before the largest size)"
	}
	switch {
	case math.IsNaN(exp):
		return "not measurable"
	case exp <= 1.25:
		return "linear"
	case exp <= 2.0:
		return "sub-quadratic"
	case exp <= 3.0:
		return "super-quadratic"
	}
	return "worse than cubic"
}

// DiffConflicts reports the symmetric difference of two refusal sets.
func DiffConflicts(a, b []Conflict) []string {
	inA := map[string]bool{}
	for _, conflict := range a {
		inA[conflict.key()] = true
	}
	inB := map[string]bool{}
	for _, conflict := range b {
		inB[conflict.key()] = true
	}
	var diffs []string
	for key := range inA {
		if !inB[key] {
			diffs = append(diffs, "only-in-first: "+key)
		}
	}
	for key := range inB {
		if !inA[key] {
			diffs = append(diffs, "only-in-second: "+key)
		}
	}
	sort.Strings(diffs)
	return diffs
}

// DiffLocal reports where two local-liveness maps disagree.
func DiffLocal(mechanism, oracle LocalLiveness, covered map[string]bool) []string {
	var diffs []string
	for key, oracleLoans := range oracle {
		mechanismLoans := mechanism[key]
		if fmt.Sprint(mechanismLoans) != fmt.Sprint(oracleLoans) {
			diffs = append(diffs, fmt.Sprintf("%s: mechanism=%v oracle=%v", key, mechanismLoans, oracleLoans))
		}
	}
	sort.Strings(diffs)
	return diffs
}

// InvalidationProbe answers the second half of the cache question: if a cache
// IS designed in, what does one callee edit cost? It reports how many cached
// summaries a single function's change must evict (itself plus every
// transitive caller) -- the work a ClosureDigest chain forces on the next run.
type InvalidationProbe struct {
	Shape          Shape   `json:"shape"`
	Functions      int     `json:"functions"`
	MaxEvicted     int     `json:"max_evicted"`
	MeanEvicted    float64 `json:"mean_evicted"`
	MaxEvictedFrac float64 `json:"max_evicted_fraction"`
	WorstFunction  string  `json:"worst_function"`
}

// ProbeInvalidation edits each function in turn and measures cache fallout.
func ProbeInvalidation(shape Shape, p *Program) (InvalidationProbe, error) {
	probe := InvalidationProbe{Shape: shape, Functions: len(p.Functions)}
	total := 0
	ids := make([]string, 0, len(p.Functions))
	for i := range p.Functions {
		ids = append(ids, p.Functions[i].ID)
	}
	sort.Strings(ids)
	for _, id := range ids {
		memo := NewMemoProvider(p, 0)
		if _, err := Analyze(p, memo); err != nil {
			return probe, err
		}
		evicted := memo.Invalidate(p, id)
		total += evicted
		if evicted > probe.MaxEvicted {
			probe.MaxEvicted, probe.WorstFunction = evicted, id
		}
	}
	if len(ids) > 0 {
		probe.MeanEvicted = float64(total) / float64(len(ids))
		probe.MaxEvictedFrac = float64(probe.MaxEvicted) / float64(len(ids))
	}
	return probe, nil
}

// BodyOrder reports what the program-order assumption is worth: the memoized
// arm's work per operation on a k-link forwarding chain listed in program
// order versus the same chain listed in reverse.
type BodyOrder struct {
	Links           int     `json:"links"`
	OrderedPerOp    float64 `json:"ordered_work_per_op"`
	ReversedPerOp   float64 `json:"reversed_work_per_op"`
	ReversedPenalty float64 `json:"reversed_penalty"`
}

// ProbeBodyOrder measures both orders at one chain length.
func ProbeBodyOrder(k int) (BodyOrder, error) {
	probe := BodyOrder{Links: k}
	ordered, err := Generate(ShapeForward, k+1)
	if err != nil {
		return probe, err
	}
	orderedAnalysis, err := Analyze(ordered, NewMemoProvider(ordered, 0))
	if err != nil {
		return probe, err
	}
	probe.OrderedPerOp = float64(orderedAnalysis.Work) / float64(ordered.OpCount())

	reversed := ReversedForward(k)
	reversedAnalysis, err := Analyze(reversed, NewMemoProvider(reversed, 0))
	if err != nil {
		return probe, err
	}
	probe.ReversedPerOp = float64(reversedAnalysis.Work) / float64(reversed.OpCount())
	if probe.OrderedPerOp > 0 {
		probe.ReversedPenalty = probe.ReversedPerOp / probe.OrderedPerOp
	}
	return probe, nil
}
