package measure

import (
	"errors"
	"math"
	"testing"
)

func TestSampleCountConstantsAreExported(t *testing.T) {
	if WarmSampleCount != 20 {
		t.Fatalf("WarmSampleCount = %d, want 20", WarmSampleCount)
	}
	if ColdSampleCount <= 0 {
		t.Fatalf("ColdSampleCount = %d, want a positive exported constant", ColdSampleCount)
	}
}

// TestWarmSamplePercentilesMatchShellPrecedent is table-driven over at
// least three 20-element sample sets and asserts the exact element
// scripts/verify-phase2.sh's observe() would select via `sort -n` then
// `sed -n '10p'` / `sed -n '19p'` (0-indexed 9 and 18) -- the expected
// values below are written as literals hand-derived from each sorted set.
func TestWarmSamplePercentilesMatchShellPrecedent(t *testing.T) {
	cases := []struct {
		name        string
		samples     Samples
		wantP50     int64
		wantP95     int64
	}{
		{
			name:    "ascending 1..20",
			samples: Samples{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			// sorted[9] (0-indexed) = 10 (the 10th line, sed -n '10p')
			wantP50: 10,
			// sorted[18] (0-indexed) = 19 (the 19th line, sed -n '19p')
			wantP95: 19,
		},
		{
			name:    "descending 20..1",
			samples: Samples{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			wantP50: 10,
			wantP95: 19,
		},
		{
			name: "multiples of 3",
			samples: Samples{
				3, 6, 9, 12, 15, 18, 21, 24, 27, 30,
				33, 36, 39, 42, 45, 48, 51, 54, 57, 60,
			},
			// sorted[9] = 3*10 = 30
			wantP50: 30,
			// sorted[18] = 3*19 = 57
			wantP95: 57,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			summary, err := tc.samples.Summary()
			if err != nil {
				t.Fatalf("Summary() error = %v", err)
			}
			if summary.P50 != tc.wantP50 {
				t.Fatalf("P50 = %d, want %d", summary.P50, tc.wantP50)
			}
			if summary.P95 != tc.wantP95 {
				t.Fatalf("P95 = %d, want %d", summary.P95, tc.wantP95)
			}
			if summary.Count != 20 {
				t.Fatalf("Count = %d, want 20", summary.Count)
			}
		})
	}
}

// TestPercentileSelectionIsStableOnTies asserts the same index rule applies
// when several values are equal, and the result does not depend on input
// order.
func TestPercentileSelectionIsStableOnTies(t *testing.T) {
	sortedGroups := Samples{
		5, 5, 5, 5, 5,
		10, 10, 10, 10, 10,
		15, 15, 15, 15, 15,
		20, 20, 20, 20, 20,
	}
	shuffled := Samples{
		20, 10, 5, 15, 20, 10, 5, 15, 20, 10,
		5, 15, 20, 10, 5, 15, 20, 10, 5, 15,
	}

	sortedSummary, err := sortedGroups.Summary()
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	shuffledSummary, err := shuffled.Summary()
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if sortedSummary.P50 != shuffledSummary.P50 {
		t.Fatalf("P50 depends on input order: sorted=%d shuffled=%d", sortedSummary.P50, shuffledSummary.P50)
	}
	if sortedSummary.P95 != shuffledSummary.P95 {
		t.Fatalf("P95 depends on input order: sorted=%d shuffled=%d", sortedSummary.P95, shuffledSummary.P95)
	}
	// index 9 (0-indexed) of the sorted group set falls in the "10" group.
	if sortedSummary.P50 != 10 {
		t.Fatalf("P50 = %d, want 10", sortedSummary.P50)
	}
	// index 18 (0-indexed) falls in the "20" group.
	if sortedSummary.P95 != 20 {
		t.Fatalf("P95 = %d, want 20", sortedSummary.P95)
	}
}

func TestSampleStatisticsRefuseShortSampleSets(t *testing.T) {
	nineteen := make(Samples, 19)
	for i := range nineteen {
		nineteen[i] = int64(i + 1)
	}
	_, err := nineteen.Summary()
	var typed *Error
	if !errors.As(err, &typed) {
		t.Fatalf("Summary() error = %v, want *Error", err)
	}
	if typed.Got != 19 || typed.Want != 20 {
		t.Fatalf("Summary() error = %+v, want Got=19 Want=20", typed)
	}
}

func TestSampleStatisticsRefuseEmptyAndNonPositive(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		_, err := Samples{}.Summary()
		var typed *Error
		if !errors.As(err, &typed) {
			t.Fatalf("Summary() error = %v, want *Error", err)
		}
		if typed.Code != "measure.stats_empty" {
			t.Fatalf("Summary() error code = %q, want measure.stats_empty", typed.Code)
		}
	})

	t.Run("nonpositive", func(t *testing.T) {
		withZero := make(Samples, WarmSampleCount)
		for i := range withZero {
			withZero[i] = int64(i + 1)
		}
		withZero[5] = 0
		_, err := withZero.Summary()
		var typed *Error
		if !errors.As(err, &typed) {
			t.Fatalf("Summary() error = %v, want *Error", err)
		}
		if typed.Code != "measure.stats_nonpositive_sample" {
			t.Fatalf("Summary() error code = %q, want measure.stats_nonpositive_sample", typed.Code)
		}
	})

	t.Run("negative", func(t *testing.T) {
		withNegative := make(Samples, WarmSampleCount)
		for i := range withNegative {
			withNegative[i] = int64(i + 1)
		}
		withNegative[5] = -3
		_, err := withNegative.Summary()
		var typed *Error
		if !errors.As(err, &typed) {
			t.Fatalf("Summary() error = %v, want *Error", err)
		}
		if typed.Code != "measure.stats_nonpositive_sample" {
			t.Fatalf("Summary() error code = %q, want measure.stats_nonpositive_sample", typed.Code)
		}
	})
}

// TestSummaryComputesCoVFromStdDevOverMean is a supporting behavioural test
// (not in the plan's named list, but required to exercise CoV/StdDev/Mean
// before Task 3's Demote rule consumes them).
func TestSummaryComputesCoVFromStdDevOverMean(t *testing.T) {
	constant := make(Samples, WarmSampleCount)
	for i := range constant {
		constant[i] = 100
	}
	summary, err := constant.Summary()
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.Mean != 100 {
		t.Fatalf("Mean = %v, want 100", summary.Mean)
	}
	if summary.StdDev != 0 {
		t.Fatalf("StdDev = %v, want 0 for a constant sample set", summary.StdDev)
	}
	if summary.CoV != 0 {
		t.Fatalf("CoV = %v, want 0 for a constant sample set", summary.CoV)
	}

	varied := Samples{
		90, 95, 100, 105, 110, 90, 95, 100, 105, 110,
		90, 95, 100, 105, 110, 90, 95, 100, 105, 110,
	}
	variedSummary, err := varied.Summary()
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if variedSummary.CoV <= 0 {
		t.Fatalf("CoV = %v, want > 0 for a varied sample set", variedSummary.CoV)
	}
	if math.IsInf(variedSummary.CoV, 0) || math.IsNaN(variedSummary.CoV) {
		t.Fatalf("CoV = %v, want a finite value", variedSummary.CoV)
	}
}
