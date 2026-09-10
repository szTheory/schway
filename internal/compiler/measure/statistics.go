package measure

import (
	"math"
	"sort"
)

// WarmSampleCount is D-06-19's 20-warm-sample discipline, matching
// scripts/verify-phase2.sh's observe() loop (`while [ "$index" -lt 20 ]`)
// exactly. It is exported so scripts/verify-phase6.sh can pin it verbatim,
// mirroring the TestPhase5CorpusBoundMatchesScript precedent.
const WarmSampleCount = 20

// ColdSampleCount is a small fixed n for the cold path -- cold state
// cannot be re-established in-process, so a large cold sample count is not
// meaningful. Claude's Discretion per D-06-19: 5 is small enough to be
// cheap per invocation while still giving a non-degenerate cold
// distribution to report alongside the warm one.
const ColdSampleCount = 5

// Samples is a raw collection of nanosecond elapsed-time (or other counted
// metric) observations, in insertion order. Summary sorts a private copy;
// the exported slice itself is never mutated.
type Samples []int64

// Summary is the sorted-index percentile summary and noise statistics for
// one Samples set: p50/p95 selected by the same sorted-index rule
// scripts/verify-phase2.sh's `sort -n` + `sed -n '10p'`/`sed -n '19p'`
// uses, plus Mean/StdDev/CoV computed in float64 for the D-06-19 auto-
// demotion rule.
type Summary struct {
	P50    int64
	P95    int64
	Mean   float64
	StdDev float64
	CoV    float64
	Count  int
}

// Summary refuses to report a percentile over a sample set that was never
// fully collected -- reporting one would be exactly the fabrication the
// honest-unavailability house style forbids (T-06-MEASURE-03). It refuses,
// in order: an empty set, a set shorter than WarmSampleCount, and any
// non-positive sample (the shell loop already refuses a zero timing; this
// port must not be more permissive).
func (s Samples) Summary() (Summary, error) {
	if len(s) == 0 {
		return Summary{}, &Error{Code: "measure.stats_empty"}
	}
	if len(s) < WarmSampleCount {
		return Summary{}, &Error{Code: "measure.stats_short_sample_set", Got: len(s), Want: WarmSampleCount}
	}
	sorted := make([]int64, len(s))
	copy(sorted, s)
	var sum float64
	for _, value := range sorted {
		if value <= 0 {
			return Summary{}, &Error{Code: "measure.stats_nonpositive_sample"}
		}
		sum += float64(value)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	n := len(sorted)
	p50 := sorted[percentileIndex(n, 50)]
	p95 := sorted[percentileIndex(n, 95)]

	mean := sum / float64(n)
	if mean == 0 {
		// Unreachable given the non-positive-sample refusal above (every
		// sample is > 0, so the mean of a non-empty set is > 0 too), but
		// guarded explicitly per D-06-19's "never fabricate 0" house
		// style rather than relying on that invariant implicitly.
		return Summary{}, &Error{Code: "measure.stats_zero_mean"}
	}

	var varianceSum float64
	for _, value := range sorted {
		delta := float64(value) - mean
		varianceSum += delta * delta
	}
	variance := varianceSum / float64(n)
	stdDev := math.Sqrt(variance)
	cov := stdDev / mean

	return Summary{P50: p50, P95: p95, Mean: mean, StdDev: stdDev, CoV: cov, Count: n}, nil
}

// CoVDemotionThreshold is the fixed coefficient-of-variation threshold
// D-06-19's auto-demotion rule fires on. Claude's Discretion: this project
// runs on exactly one laptop-class Apple-silicon host with no CI fleet
// (D-06-15) -- battery state, P/E core scheduling, and thermal throttling
// all confound into a single sample stream, and there is no cross-machine
// averaging to smooth them out. 0.15 (15%) is a conservative threshold for
// that single-host, no-fleet situation: tight enough to catch genuinely
// unstable measurements, loose enough not to demote every warm run on a
// machine with no dedicated benchmarking isolation. Exported so
// scripts/verify-phase6.sh and a pin test can both cite it.
const CoVDemotionThreshold = 0.15

// The verdict vocabulary is closed to exactly these three values (D-06-19,
// D-06-22): any other string is a bug, never a fourth state quietly
// introduced later.
const (
	VerdictBlocking    = "blocking"
	VerdictObserved    = "observed"
	VerdictNotRatified = "not_ratified"
)

// Verdicts returns the closed three-value verdict set, in stable order.
func Verdicts() []string {
	return []string{VerdictBlocking, VerdictObserved, VerdictNotRatified}
}

// GateEligibleMetrics is the closed set of metric names Demote's rule 2
// treats as gate-eligible at all (D-06-14/D-06-22, widened by Phase 08
// Plan 05's D-08-31/D-08-32 to a two-element set): "recomputed_work" is
// deterministic and machine-independent, and so is
// "recomputed_work_growth_exponent" -- both require zero statistics to
// gate on. Every OTHER metric (wall clock, output bytes, ...) is never
// blocking on its own.
//
// measure must not import session (session imports measure), so this is
// deliberately a SECOND, independent declaration of the same two names
// session.QLT02GateEligibleMetrics() returns -- never a shared constant.
// TestGateEligibleMetricSetsAgreeAcrossChokepoints (package session) proves
// the two chokepoints agree as sets; a future one-sided widening of only
// one of them fails that test rather than silently shipping a manifest row
// that can never block (T-08-17).
func GateEligibleMetrics() []string {
	return []string{"recomputed_work", "recomputed_work_growth_exponent"}
}

func gateEligibleMetricSet() map[string]bool {
	set := make(map[string]bool, len(GateEligibleMetrics()))
	for _, metric := range GateEligibleMetrics() {
		set[metric] = true
	}
	return set
}

// Demote is D-06-19's mechanical, one-directional quarantine rule: it only
// ever demotes a requested gate type toward observed/not_ratified, never
// promotes one toward blocking. The rule, in this order, with no other
// branches:
//
//  1. A refused sample set (err != nil) is undeterminable -- not_ratified,
//     never a silent pass and never blocking.
//  2. Only a metric in GateEligibleMetrics() is gate-eligible at all
//     (D-06-14/D-06-22, widened to a SET by D-08-31/D-08-32): wall clock
//     and output bytes are never blocking on their own, and neither is any
//     future metric this function has not been told about.
//  3. A CoV above the fixed threshold demotes to observed, regardless of
//     the caller's requested gate type.
//  4. Otherwise, requested passes through unchanged -- and the ONLY
//     literal `return VerdictBlocking` in this function is that
//     pass-through, asserted structurally by
//     TestDemoteHasExactlyOnePromotionPassthrough so a future edit cannot
//     add a second promotion path without failing that test.
func Demote(requested string, metric string, summary Summary, err error) string {
	if err != nil {
		return VerdictNotRatified
	}
	if !gateEligibleMetricSet()[metric] {
		return VerdictObserved
	}
	if summary.CoV > CoVDemotionThreshold {
		return VerdictObserved
	}
	if requested == VerdictBlocking {
		return VerdictBlocking
	}
	return requested
}

// percentileIndex reproduces scripts/verify-phase2.sh's sorted-index
// percentile rule exactly rather than inventing a new interpolation: at
// n=20, `sed -n '10p'` on a 1-indexed sorted stream is 0-indexed position
// 9 for p50, and `sed -n '19p'` is 0-indexed position 18 for p95. The
// general form floor(percentile/100*n) - 1 reproduces both: floor(0.50*20)
// -1 = 9, floor(0.95*20)-1 = 18. This generalizes the same fractional-
// index rule for other n, clamped to the valid index range so it never
// under- or over-runs the sorted slice.
func percentileIndex(n int, percentile float64) int {
	index := int(percentile/100*float64(n)) - 1
	if index < 0 {
		index = 0
	}
	if index > n-1 {
		index = n - 1
	}
	return index
}
