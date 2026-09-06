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
