package session

import (
	"context"
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/measure"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
)

func liveMachineIDForTest(t *testing.T) string {
	t.Helper()
	facts, err := measure.ProbeMachine(context.Background())
	if err != nil {
		t.Fatalf("ProbeMachine: %v", err)
	}
	return measure.MachineID(facts)
}

// --- Task 1: manifest load / tracer -----------------------------------

func TestBudgetManifestLoads(t *testing.T) {
	rows, err := LoadQLT02BudgetManifest()
	if err != nil {
		t.Fatalf("LoadQLT02BudgetManifest: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("LoadQLT02BudgetManifest returned no rows")
	}
	for _, row := range rows {
		if row.GateType != QLT02GateTypeHard && row.GateType != QLT02GateTypeObserved {
			t.Errorf("row %+v has out-of-vocabulary gate_type %q", row, row.GateType)
		}
		if row.RatifiedAt == "" || row.RatifiedAt == "TODO" {
			t.Errorf("row %+v has placeholder or empty ratified_at", row)
		}
		if row.RatifiedByCommit == "" || row.RatifiedByCommit == "TODO" {
			t.Errorf("row %+v has placeholder or empty ratified_by_commit", row)
		}
	}
}

func TestBudgetLaneCarriesMachineIDAndVerdict(t *testing.T) {
	corpus := nat03CorpusPath("testdata/phase1")
	phase6TestRoots(t)
	runner := native.DefaultRunner()

	result, err := VerifyPhase6ChangedRisk(context.Background(), corpus, runner)
	if err != nil {
		t.Fatalf("VerifyPhase6ChangedRisk error: %v", err)
	}

	lane := phase6FindLane(result, phase6NativeDifferentialLane)
	if lane == nil {
		t.Fatal("no lane:native-differential in result")
	}

	live := liveMachineIDForTest(t)
	if lane.MachineID != live {
		t.Errorf("lane.MachineID = %q, want %q", lane.MachineID, live)
	}

	found := false
	for _, verdict := range protocol.LaneGateVerdicts() {
		if lane.GateVerdict == verdict {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("lane.GateVerdict = %q, not in %v", lane.GateVerdict, protocol.LaneGateVerdicts())
	}
}

func TestMeasureVerdictsMatchProtocolLaneGateVerdicts(t *testing.T) {
	measureSet := measure.Verdicts()
	protocolSet := protocol.LaneGateVerdicts()
	if len(measureSet) != len(protocolSet) {
		t.Fatalf("measure.Verdicts() = %v, protocol.LaneGateVerdicts() = %v -- not set-equal", measureSet, protocolSet)
	}
	seen := make(map[string]bool, len(protocolSet))
	for _, v := range protocolSet {
		seen[v] = true
	}
	for _, v := range measureSet {
		if !seen[v] {
			t.Errorf("measure.Verdicts() contains %q, not in protocol.LaneGateVerdicts() %v", v, protocolSet)
		}
	}
}

// --- Task 2: audit and observation-only mode --------------------------

func TestBudgetAuditRefusesUndeclaredMachine(t *testing.T) {
	rows, err := LoadQLT02BudgetManifest()
	if err != nil {
		t.Fatalf("LoadQLT02BudgetManifest: %v", err)
	}
	failures := AuditQLT02BudgetManifest(rows, "machine:fabricated000000", QLT02GateEligibleMetrics())
	found := false
	for _, f := range failures {
		if f.Control == ControlQLT02UnknownMachine {
			found = true
		}
	}
	if !found {
		t.Errorf("expected control:qlt02.unknown_machine to fire against a fabricated live machine ID, got failures: %+v", failures)
	}

	// The checked-in manifest against THIS host's real live machine ID must
	// pass with zero fired controls.
	live := liveMachineIDForTest(t)
	clean := AuditQLT02BudgetManifest(rows, live, QLT02GateEligibleMetrics())
	if len(clean) != 0 {
		t.Errorf("AuditQLT02BudgetManifest against real live machine_id %q returned failures: %+v", live, clean)
	}
}

func TestBudgetAuditRefusesEmptyManifest(t *testing.T) {
	failures := AuditQLT02BudgetManifest(nil, "machine:anything", QLT02GateEligibleMetrics())
	if len(failures) != 1 || failures[0].Control != ControlQLT02ManifestEmpty {
		t.Fatalf("AuditQLT02BudgetManifest(nil, ...) = %+v, want exactly one manifest_empty failure", failures)
	}

	lane := QLT02LaneFromRows(nil, "machine:anything", QLT02GateEligibleMetrics())
	if lane.Status != "invalid" {
		t.Errorf("QLT02LaneFromRows(nil, ...) status = %q, want invalid", lane.Status)
	}
	if lane.RecomputedWork != 0 {
		t.Errorf("QLT02LaneFromRows(nil, ...) RecomputedWork = %d, want 0 (empty manifest reports zero work)", lane.RecomputedWork)
	}
}

func TestBudgetAuditRefusesDuplicateRow(t *testing.T) {
	rows := []QLT02BudgetRow{
		{MachineID: "machine:aaa", Metric: "recomputed_work", GateType: "hard", ValueOrBound: 4, Unit: "count", RatifiedAt: "2026-01-01T00:00:00Z", RatifiedByCommit: "deadbeef"},
		{MachineID: "machine:aaa", Metric: "recomputed_work", GateType: "hard", ValueOrBound: 5, Unit: "count", RatifiedAt: "2026-01-01T00:00:00Z", RatifiedByCommit: "deadbeef"},
	}
	failures := AuditQLT02BudgetManifest(rows, "machine:aaa", []string{"recomputed_work"})
	found := false
	for _, f := range failures {
		if f.Control == ControlQLT02DuplicateRow {
			found = true
		}
	}
	if !found {
		t.Errorf("expected control:qlt02.duplicate_row to fire on a duplicated (machine_id, metric) pair, got: %+v", failures)
	}
}

func TestBudgetAuditRefusesIneligibleHardGate(t *testing.T) {
	rows := []QLT02BudgetRow{
		{MachineID: "machine:aaa", Metric: "elapsed_ns", GateType: "hard", ValueOrBound: 100, Unit: "nanoseconds", RatifiedAt: "2026-01-01T00:00:00Z", RatifiedByCommit: "deadbeef"},
	}
	failures := AuditQLT02BudgetManifest(rows, "machine:aaa", []string{"recomputed_work"})
	found := false
	for _, f := range failures {
		if f.Control == ControlQLT02IneligibleHardGate {
			found = true
		}
	}
	if !found {
		t.Errorf("expected control:qlt02.ineligible_hard_gate to fire on a hard elapsed_ns row, got: %+v", failures)
	}
}

func TestBudgetAuditRefusesUnknownMetricVocabulary(t *testing.T) {
	rows := []QLT02BudgetRow{
		{MachineID: "machine:aaa", Metric: "peak_rss_bytes", GateType: "observed", ValueOrBound: 100, Unit: "bytes", RatifiedAt: "2026-01-01T00:00:00Z", RatifiedByCommit: "deadbeef"},
	}
	failures := AuditQLT02BudgetManifest(rows, "machine:aaa", []string{"recomputed_work"})
	if len(failures) == 0 {
		t.Error("expected a failure for a metric outside the closed vocabulary, got none")
	}
}

// TestGateEligibleMetricSetsAgreeAcrossChokepoints is Task 2(d)'s own
// point (D-08-32/T-08-17): session.QLT02GateEligibleMetrics() and
// measure.GateEligibleMetrics() are two INDEPENDENT declarations of the
// same set (measure cannot import session, so they cannot share a
// constant) -- this test is what makes a future one-sided widening of only
// one chokepoint a test failure instead of a silently decorative manifest
// row that reads gate_type: hard but can never actually block.
func TestGateEligibleMetricSetsAgreeAcrossChokepoints(t *testing.T) {
	sessionSet := QLT02GateEligibleMetrics()
	measureSet := measure.GateEligibleMetrics()
	if len(sessionSet) != len(measureSet) {
		t.Fatalf("QLT02GateEligibleMetrics() = %v, measure.GateEligibleMetrics() = %v -- not set-equal", sessionSet, measureSet)
	}
	seen := make(map[string]bool, len(measureSet))
	for _, metric := range measureSet {
		seen[metric] = true
	}
	for _, metric := range sessionSet {
		if !seen[metric] {
			t.Errorf("QLT02GateEligibleMetrics() contains %q, not in measure.GateEligibleMetrics() %v", metric, measureSet)
		}
	}
}

func TestUndeclaredMachineRunsObservationOnly(t *testing.T) {
	rows, err := LoadQLT02BudgetManifest()
	if err != nil {
		t.Fatalf("LoadQLT02BudgetManifest: %v", err)
	}
	mode := RatificationMode(rows, "machine:fabricated000000")
	if mode.Ratified {
		t.Fatal("RatificationMode reported Ratified=true for a machine_id absent from the manifest")
	}
	if mode.MachineID != "machine:fabricated000000" {
		t.Errorf("Mode.MachineID = %q, want the computed machine_id echoed back", mode.MachineID)
	}
}

func TestUndeclaredMachineNeverWritesManifest(t *testing.T) {
	rows, err := LoadQLT02BudgetManifest()
	if err != nil {
		t.Fatalf("LoadQLT02BudgetManifest: %v", err)
	}
	manifestPath := qlt02RepoPath("internal/compiler/session/qlt02_budget_manifest.json")

	fabricated := "machine:fabricated000000"
	err = qlt02ManifestNeverWritten(manifestPath, func() error {
		mode := RatificationMode(rows, fabricated)
		if mode.Ratified {
			t.Fatal("fabricated machine unexpectedly ratified")
		}
		// Observation-only: never call BudgetFor/EvaluateBudget for an
		// undeclared machine. The lane's GateVerdict must be not_ratified.
		lane := protocol.Lane{GateVerdict: protocol.LaneGateNotRatified, MachineID: fabricated}
		if lane.GateVerdict != protocol.LaneGateNotRatified {
			t.Errorf("lane.GateVerdict = %q, want not_ratified", lane.GateVerdict)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("qlt02ManifestNeverWritten: %v", err)
	}
}

func TestNoFunctionWritesTheBudgetManifest(t *testing.T) {
	if err := AssertNoManifestWriteCalls(); err != nil {
		t.Fatal(err)
	}
}

// --- Task 3: the blocking rule -----------------------------------------

func TestRecomputedWorkIsTheOnlyHardGate(t *testing.T) {
	row := QLT02BudgetRow{MachineID: "machine:aaa", Metric: "recomputed_work", GateType: "hard", ValueOrBound: 10, Unit: "count"}
	below := deterministicSummary(9)
	exact := deterministicSummary(10)
	above := deterministicSummary(11)

	if v, obs := EvaluateBudget(row, below, nil); v != measure.VerdictObserved || obs != nil {
		t.Errorf("below ceiling: verdict=%q obs=%+v, want observed/nil", v, obs)
	}
	if v, obs := EvaluateBudget(row, exact, nil); v != measure.VerdictObserved || obs != nil {
		t.Errorf("exactly at ceiling: verdict=%q obs=%+v, want observed/nil (exactly-equal is not a regression)", v, obs)
	}
	if v, obs := EvaluateBudget(row, above, nil); v != measure.VerdictBlocking || obs != nil {
		t.Errorf("above ceiling: verdict=%q obs=%+v, want blocking/nil", v, obs)
	}
}

func TestBudgetCeilingIsStrictlyExceeded(t *testing.T) {
	row := QLT02BudgetRow{MachineID: "machine:aaa", Metric: "recomputed_work", GateType: "hard", ValueOrBound: 4}
	verdict, _ := EvaluateBudget(row, deterministicSummary(4), nil)
	if verdict != measure.VerdictObserved {
		t.Errorf("exactly-equal verdict = %q, want observed (strictly-exceeds rule)", verdict)
	}
	verdict, _ = EvaluateBudget(row, deterministicSummary(5), nil)
	if verdict != measure.VerdictBlocking {
		t.Errorf("one-over verdict = %q, want blocking", verdict)
	}
}

func TestWallClockObservationIsNeverBlocking(t *testing.T) {
	row := QLT02BudgetRow{MachineID: "machine:aaa", Metric: "elapsed_ns", GateType: "observed", ValueOrBound: 100}
	summary := measure.Summary{P50: 1_000_000, P95: 1_000_000, Mean: 1_000_000, StdDev: 0, CoV: 0, Count: 20}
	verdict, obs := EvaluateBudget(row, summary, nil)
	if verdict == measure.VerdictBlocking {
		t.Errorf("a wall-clock overrun produced verdict=blocking, want never-blocking")
	}
	if obs == nil {
		t.Fatal("expected a flagged Observation for the wall-clock overrun, got nil")
	}
	if obs.Citable {
		t.Error("Observation.Citable = true with no stage deltas supplied, want false")
	}
}

func TestOutputBytesObservationBehavesLikeWallClock(t *testing.T) {
	row := QLT02BudgetRow{MachineID: "machine:aaa", Metric: "output_bytes", GateType: "observed", ValueOrBound: 10}
	summary := measure.Summary{P50: 20, P95: 20, Mean: 20, StdDev: 0, CoV: 0, Count: 20}
	verdict, obs := EvaluateBudget(row, summary, nil)
	if verdict == measure.VerdictBlocking {
		t.Errorf("output_bytes overrun produced verdict=blocking, want never-blocking")
	}
	if obs == nil {
		t.Fatal("expected a flagged Observation, got nil")
	}
}

func TestUncitableObservationReportsHonestly(t *testing.T) {
	row := QLT02BudgetRow{MachineID: "machine:aaa", Metric: "elapsed_ns", GateType: "observed", ValueOrBound: 10}
	summary := measure.Summary{P50: 50, P95: 50, Mean: 50, StdDev: 0, CoV: 0, Count: 20}
	_, obs := EvaluateBudget(row, summary, nil)
	if obs == nil {
		t.Fatal("expected an Observation")
	}
	if obs.Citable {
		t.Error("Observation.Citable = true with no stage deltas, want false (uncitable, not dropped)")
	}
}

func TestBlockingRegressionCitesStageWorkDelta(t *testing.T) {
	row := QLT02BudgetRow{MachineID: "machine:aaa", Metric: "recomputed_work", GateType: "hard", ValueOrBound: 4}
	diagnostic := BudgetRegressionDiagnostic(phase6NativeDifferentialLane, row, 5)
	for _, want := range []string{phase6NativeDifferentialLane, "4", "5"} {
		if !stringsContains(diagnostic, want) {
			t.Errorf("diagnostic %q does not name %q", diagnostic, want)
		}
	}
}

// TestEvaluateBudgetAgreesWithDemote, widened by Phase 08 Plan 05 (D-08-32):
// over every metric OUTSIDE QLT02GateEligibleMetrics() (now a two-element
// set, not merely "everything but recomputed_work"), EvaluateBudget's
// verdict must equal measure.Demote's own verdict, and neither may ever be
// blocking. Gate-eligible metrics are deliberately excluded from this loop
// -- they go through EvaluateBudget's OWN strict value-vs-bound comparison
// (see TestGrowthExponentIsAlsoAHardGate below), not through Demote, so
// asserting parity with Demote for them would be asserting the wrong
// contract.
func TestEvaluateBudgetAgreesWithDemote(t *testing.T) {
	summary := measure.Summary{P50: 10, P95: 10, Mean: 10, StdDev: 0, CoV: 0, Count: 20}
	eligible := make(map[string]bool, len(QLT02GateEligibleMetrics()))
	for _, metric := range QLT02GateEligibleMetrics() {
		eligible[metric] = true
	}
	for _, metric := range QLT02MetricVocabulary() {
		if eligible[metric] {
			continue
		}
		row := QLT02BudgetRow{MachineID: "machine:aaa", Metric: metric, GateType: "observed", ValueOrBound: 1}
		verdict, _ := EvaluateBudget(row, summary, nil)
		demoted := measure.Demote(measure.VerdictBlocking, metric, summary, nil)
		if verdict != demoted {
			t.Errorf("metric %q: EvaluateBudget verdict %q != measure.Demote verdict %q", metric, verdict, demoted)
		}
		if verdict == measure.VerdictBlocking {
			t.Errorf("metric %q: verdict blocking, but only QLT02GateEligibleMetrics() are gate-eligible", metric)
		}
	}
}

// TestGrowthExponentIsAlsoAHardGate is D-08-32's Rule 1 fix, made
// affirmative: recomputed_work_growth_exponent -- the SECOND gate-eligible
// metric -- gets the exact same strict value-vs-bound treatment
// TestRecomputedWorkIsTheOnlyHardGate already pins for recomputed_work
// (below/exact/above the ceiling), and high CoV never masks a genuine
// regression for it either, mirroring TestRecomputedWorkBlockingIgnoresHighCoV.
// Before the Rule 1 fix in EvaluateBudget, this metric fell through to the
// Demote-only branch and reported VerdictBlocking unconditionally whenever
// requested (regardless of value vs bound) -- this test is what would have
// caught that.
func TestGrowthExponentIsAlsoAHardGate(t *testing.T) {
	row := QLT02BudgetRow{MachineID: "machine:aaa", Metric: "recomputed_work_growth_exponent", GateType: "hard", ValueOrBound: 1200}
	below := deterministicSummary(1199)
	exact := deterministicSummary(1200)
	above := deterministicSummary(1201)

	if v, obs := EvaluateBudget(row, below, nil); v != measure.VerdictObserved || obs != nil {
		t.Errorf("below ceiling: verdict=%q obs=%+v, want observed/nil", v, obs)
	}
	if v, obs := EvaluateBudget(row, exact, nil); v != measure.VerdictObserved || obs != nil {
		t.Errorf("exactly at ceiling: verdict=%q obs=%+v, want observed/nil (exactly-equal is not a regression)", v, obs)
	}
	if v, obs := EvaluateBudget(row, above, nil); v != measure.VerdictBlocking || obs != nil {
		t.Errorf("above ceiling: verdict=%q obs=%+v, want blocking/nil", v, obs)
	}

	noisy := measure.Summary{P50: 1201, P95: 1201, Mean: 1201, StdDev: 400, CoV: 0.9, Count: 20}
	if v, _ := EvaluateBudget(row, noisy, nil); v != measure.VerdictBlocking {
		t.Errorf("growth-exponent regression with high CoV verdict = %q, want blocking (deterministic, no noise floor)", v)
	}
}

func TestRecomputedWorkBlockingIgnoresHighCoV(t *testing.T) {
	row := QLT02BudgetRow{MachineID: "machine:aaa", Metric: "recomputed_work", GateType: "hard", ValueOrBound: 4}
	noisy := measure.Summary{P50: 5, P95: 5, Mean: 5, StdDev: 2, CoV: 0.9, Count: 20}
	verdict, _ := EvaluateBudget(row, noisy, nil)
	if verdict != measure.VerdictBlocking {
		t.Errorf("recomputed_work regression with high CoV verdict = %q, want blocking (deterministic, no noise floor)", verdict)
	}
}

// --- housekeeping --------------------------------------------------------

// TestQLT02BudgetManifestFileUnchangedDuringAudit is a defensive
// double-check that the qlt02ManifestNeverWritten test seam itself works
// against the real checked-in file, independent of TestUndeclaredMachineNeverWritesManifest's
// own use of it.
func TestQLT02BudgetManifestFileUnchangedDuringAudit(t *testing.T) {
	manifestPath := qlt02RepoPath("internal/compiler/session/qlt02_budget_manifest.json")
	if _, err := os.Stat(manifestPath); err != nil {
		t.Fatalf("checked-in manifest not found at %s: %v", manifestPath, err)
	}
	if err := qlt02ManifestNeverWritten(manifestPath, func() error { return nil }); err != nil {
		t.Fatalf("qlt02ManifestNeverWritten reported unexpected change: %v", err)
	}
}
