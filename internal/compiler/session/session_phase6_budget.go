package session

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"github.com/codename-lang/lang/internal/compiler/measure"
	"github.com/codename-lang/lang/internal/compiler/protocol"

	_ "embed"
)

//go:embed qlt02_budget_manifest.json
var qlt02BudgetManifestBytes []byte

// QLT02BudgetRow is one checked-in budget-manifest row (D-06-16), mirroring
// qlt01.go's registry-row shape exactly: a flat JSON registry is the review
// surface, never Go constants and never a bespoke DSL. JSON tag order
// matches the plan's declared field order verbatim.
type QLT02BudgetRow struct {
	MachineID        string `json:"machine_id"`
	Metric           string `json:"metric"`
	GateType         string `json:"gate_type"`
	ValueOrBound     int64  `json:"value_or_bound"`
	Unit             string `json:"unit"`
	RatifiedAt       string `json:"ratified_at"`
	RatifiedByCommit string `json:"ratified_by_commit"`
}

// QLT02GateTypeHard and QLT02GateTypeObserved are the two values a row's
// GateType may hold (D-06-16). Any third value is an audit failure.
const (
	QLT02GateTypeHard     = "hard"
	QLT02GateTypeObserved = "observed"
)

// QLT02MetricVocabulary is the closed set of metrics a budget row may
// declare (D-06-14, D-06-15, D-06-21): the one deterministic hard-gate
// candidate plus the two loose, never-blocking-on-their-own observations.
// A row naming a metric outside this set is malformed.
func QLT02MetricVocabulary() []string {
	return []string{"recomputed_work", "elapsed_ns", "output_bytes", "recomputed_work_growth_exponent"}
}

// QLT02GateEligibleMetrics is the closed set of metrics that may ever
// carry gate_type "hard" (D-06-14, D-06-22, widened to two elements by
// Phase 08 Plan 05's D-08-31/D-08-32): recomputed_work and
// recomputed_work_growth_exponent are both deterministic and
// machine-independent, so each requires zero statistics to gate on.
// Supplied by the caller (not hardcoded inside the audit) so the audit
// itself carries no opinion of its own.
//
// This is deliberately a SECOND, independent declaration of the same two
// names measure.GateEligibleMetrics() returns -- session imports measure,
// so it COULD read that slice directly, but the two chokepoints are kept
// independent on purpose (T-08-17): TestGateEligibleMetricSetsAgreeAcrossChokepoints
// proves they agree as sets, so a future one-sided widening of only one of
// them is a test failure, not a silently decorative manifest row.
func QLT02GateEligibleMetrics() []string {
	return []string{"recomputed_work", "recomputed_work_growth_exponent"}
}

// qlt02GateEligibleMetricSet is QLT02GateEligibleMetrics() as a membership
// set, consulted by EvaluateBudget (D-08-32's Rule 1 fix, see its own doc
// comment) so the strict value-vs-bound comparison applies to EVERY
// gate-eligible metric, not merely a hardcoded literal.
func qlt02GateEligibleMetricSet() map[string]bool {
	set := make(map[string]bool, len(QLT02GateEligibleMetrics()))
	for _, metric := range QLT02GateEligibleMetrics() {
		set[metric] = true
	}
	return set
}

// LoadQLT02BudgetManifest parses the embedded QLT-02 budget manifest.
func LoadQLT02BudgetManifest() ([]QLT02BudgetRow, error) {
	var rows []QLT02BudgetRow
	if err := json.Unmarshal(qlt02BudgetManifestBytes, &rows); err != nil {
		return nil, fmt.Errorf("qlt02: failed to parse embedded manifest: %w", err)
	}
	return rows, nil
}

// BudgetFor looks up the one row declaring machineID and metric, mirroring
// the lookup shape callers need before ratifying a lane's gate verdict.
func BudgetFor(rows []QLT02BudgetRow, machineID, metric string) (QLT02BudgetRow, bool) {
	for _, row := range rows {
		if row.MachineID == machineID && row.Metric == metric {
			return row, true
		}
	}
	return QLT02BudgetRow{}, false
}

// Control identifiers for the QLT-02 budget-manifest audit (D-06-16).
const (
	// ControlQLT02ManifestEmpty fires when the manifest has no rows at all,
	// or when a row is otherwise structurally malformed (an out-of-
	// vocabulary gate_type or metric) -- the general malformed-manifest
	// bucket, mirroring qlt01.go's ControlQLT01RegistryIncomplete overload.
	ControlQLT02ManifestEmpty = "control:qlt02.manifest_empty"
	// ControlQLT02UnknownMachine fires when a row declares a machine_id
	// that the live probe running this audit did not produce -- the "edit
	// the number until it is green" defense (D-06-16).
	ControlQLT02UnknownMachine = "control:qlt02.unknown_machine"
	// ControlQLT02DuplicateRow fires when two rows share the same
	// (machine_id, metric) pair -- a duplicate must never silently win.
	ControlQLT02DuplicateRow = "control:qlt02.duplicate_row"
	// ControlQLT02IneligibleHardGate fires when a row's gate_type is "hard"
	// for a metric other than recomputed_work -- only the deterministic
	// counter is gate-eligible (D-06-14, D-06-22).
	ControlQLT02IneligibleHardGate = "control:qlt02.ineligible_hard_gate"
)

// QLT02AuditFailure is one budget-manifest audit violation.
type QLT02AuditFailure struct {
	MachineID string
	Metric    string
	Reason    string
	Control   string
}

func (f QLT02AuditFailure) Error() string {
	return fmt.Sprintf("qlt02 budget manifest row machine_id=%s metric=%s failed %s: %s", f.MachineID, f.Metric, f.Control, f.Reason)
}

func qlt02MetricVocabularySet() map[string]bool {
	set := make(map[string]bool, len(QLT02MetricVocabulary()))
	for _, metric := range QLT02MetricVocabulary() {
		set[metric] = true
	}
	return set
}

// AuditQLT02BudgetManifest implements D-06-16's executable audit: an empty
// manifest is a hard failure (FND-04's empty-input edge); a row whose
// machine_id has no live-probe counterpart this invocation is refused (the
// "edit the number until it is green" defense); two rows sharing a
// (machine_id, metric) pair is a refused duplicate (FND-04's adjacency
// edge); a row naming a metric outside the closed vocabulary is refused;
// and a "hard" row for anything but a gate-eligible metric is refused.
func AuditQLT02BudgetManifest(rows []QLT02BudgetRow, liveMachineID string, gateEligibleMetrics []string) []QLT02AuditFailure {
	var failures []QLT02AuditFailure

	if len(rows) == 0 {
		return append(failures, QLT02AuditFailure{
			Reason: "manifest is empty", Control: ControlQLT02ManifestEmpty,
		})
	}

	eligible := make(map[string]bool, len(gateEligibleMetrics))
	for _, metric := range gateEligibleMetrics {
		eligible[metric] = true
	}
	metricVocabulary := qlt02MetricVocabularySet()

	seenPairs := make(map[string]bool)
	unknownMachines := make(map[string]bool)

	for _, row := range rows {
		if row.GateType != QLT02GateTypeHard && row.GateType != QLT02GateTypeObserved {
			failures = append(failures, QLT02AuditFailure{
				MachineID: row.MachineID, Metric: row.Metric,
				Reason:  fmt.Sprintf("gate_type %q is neither %q nor %q", row.GateType, QLT02GateTypeHard, QLT02GateTypeObserved),
				Control: ControlQLT02ManifestEmpty,
			})
		}

		if !metricVocabulary[row.Metric] {
			failures = append(failures, QLT02AuditFailure{
				MachineID: row.MachineID, Metric: row.Metric,
				Reason:  fmt.Sprintf("metric %q is not in the closed metric vocabulary %v", row.Metric, QLT02MetricVocabulary()),
				Control: ControlQLT02ManifestEmpty,
			})
		}

		if row.GateType == QLT02GateTypeHard && !eligible[row.Metric] {
			failures = append(failures, QLT02AuditFailure{
				MachineID: row.MachineID, Metric: row.Metric,
				Reason:  fmt.Sprintf("metric %q is gate_type %q but is not gate-eligible (only %v may be)", row.Metric, QLT02GateTypeHard, gateEligibleMetrics),
				Control: ControlQLT02IneligibleHardGate,
			})
		}

		pairKey := row.MachineID + "\x00" + row.Metric
		if seenPairs[pairKey] {
			failures = append(failures, QLT02AuditFailure{
				MachineID: row.MachineID, Metric: row.Metric,
				Reason:  "duplicate (machine_id, metric) pair -- exactly one row per pair is required",
				Control: ControlQLT02DuplicateRow,
			})
		}
		seenPairs[pairKey] = true

		if row.MachineID != liveMachineID && !unknownMachines[row.MachineID] {
			unknownMachines[row.MachineID] = true
			failures = append(failures, QLT02AuditFailure{
				MachineID: row.MachineID,
				Reason:    fmt.Sprintf("machine_id %q has no live probe counterpart this invocation (live machine_id is %q)", row.MachineID, liveMachineID),
				Control:   ControlQLT02UnknownMachine,
			})
		}
	}

	return failures
}

// LaneQLT02BudgetAudit is the lane identifier this plan wires the budget
// audit under, mirroring qlt01.go's LaneQLT01RegistryAudit precedent.
const LaneQLT02BudgetAudit = "lane:qlt02-budget-audit"

// QLT02LaneFromRows is VerifyQLT02BudgetManifest's core, exported so tests
// can exercise the empty-manifest and fabricated-live-machine-id shapes
// directly without needing to swap out the embedded manifest (mirroring
// QLT01LaneFromRows's own test seam). One work unit is counted per row
// inspected plus one per machine-id cross-check performed, matching the
// QLT-01 lane's counted-work discipline -- an empty manifest therefore
// reports zero work rather than a fabricated nonzero count.
func QLT02LaneFromRows(rows []QLT02BudgetRow, liveMachineID string, gateEligibleMetrics []string) LaneResult {
	failures := AuditQLT02BudgetManifest(rows, liveMachineID, gateEligibleMetrics)

	work := 0
	for range rows {
		work++ // one unit per row inspected
		work++ // one unit for the machine-id cross-check performed
	}

	fired := map[string]bool{
		ControlQLT02ManifestEmpty:      false,
		ControlQLT02UnknownMachine:     false,
		ControlQLT02DuplicateRow:       false,
		ControlQLT02IneligibleHardGate: false,
	}
	for _, failure := range failures {
		fired[failure.Control] = true
	}

	status := "pass"
	var firedControls []string
	if len(failures) != 0 {
		status = "invalid"
	} else {
		firedControls = []string{
			ControlQLT02ManifestEmpty, ControlQLT02UnknownMachine,
			ControlQLT02DuplicateRow, ControlQLT02IneligibleHardGate,
		}
	}

	return LaneResult{
		ID:             LaneQLT02BudgetAudit,
		Status:         status,
		Controls:       firedControls,
		RecomputedWork: work,
		Fired:          fired,
	}
}

// VerifyQLT02BudgetManifest runs the QLT-02 budget-manifest audit as a
// counted-work lane, cross-checking the checked-in manifest against a LIVE
// probe of the machine actually running this invocation (D-06-16) -- a
// silent "edit the number until it is green" cannot pass unreviewed.
func VerifyQLT02BudgetManifest(ctx context.Context) (LaneResult, error) {
	rows, err := LoadQLT02BudgetManifest()
	if err != nil {
		return LaneResult{}, err
	}
	facts, err := measure.ProbeMachine(ctx)
	if err != nil {
		return LaneResult{}, err
	}
	liveMachineID := measure.MachineID(facts)
	return QLT02LaneFromRows(rows, liveMachineID, QLT02GateEligibleMetrics()), nil
}

// Mode is RatificationMode's report: whether the given machineID is a
// declared machine in the manifest (D-06-18). A machine with no rows at
// all is NOT ratified -- fail-closed on ratification, never on running.
type Mode struct {
	MachineID string
	Ratified  bool
}

// RatificationMode reports whether machineID has any declared row in rows.
// Callers MUST structure the branch so manifest-consulting logic (BudgetFor
// lookups, EvaluateBudget calls) lives strictly inside the Ratified==true
// arm; the Ratified==false arm reports the full distribution and the
// computed machine_id, sets every lane's GateVerdict to not_ratified, and
// takes NO further manifest-consulting or manifest-writing branch at all
// (D-06-18). This function's own single membership scan over the
// already-loaded rows is not itself the forbidden "consult" -- it is what
// determines which branch runs; the forbidden consulting is BudgetFor/
// EvaluateBudget running for an undeclared machine.
func RatificationMode(rows []QLT02BudgetRow, machineID string) Mode {
	for _, row := range rows {
		if row.MachineID == machineID {
			return Mode{MachineID: machineID, Ratified: true}
		}
	}
	return Mode{MachineID: machineID, Ratified: false}
}

// Observation is a flagged, non-blocking observation for a metric whose
// loose bound was exceeded (D-06-22). It is NEVER blocking on its own; it
// must cite the responsible stage's work-count delta (Citable) before it
// can ever become a gate candidate in a future ratification. When no stage
// delta is available, Citable is false and the observation reports that
// honestly rather than being dropped or silently promoted.
type Observation struct {
	Metric     string `json:"metric"`
	Bound      int64  `json:"bound"`
	Value      int64  `json:"value"`
	StageDelta int64  `json:"stage_delta"`
	Citable    bool   `json:"citable"`
}

// stageWorkDelta sums the elapsed-time deltas from the stage breakdown into
// one citable delta, or reports uncitable when no stage timing is
// available -- an uncitable observation must say so honestly rather than
// being dropped or silently promoted (D-06-22).
func stageWorkDelta(stageDeltas []protocol.StageTiming) (int64, bool) {
	if len(stageDeltas) == 0 {
		return 0, false
	}
	var total int64
	for _, delta := range stageDeltas {
		total += delta.ElapsedNS
	}
	return total, true
}

// EvaluateBudget implements D-06-22's blocking rule exactly: a regression is
// blocking if and only if a GATE-ELIGIBLE metric (QLT02GateEligibleMetrics(),
// widened by D-08-31/D-08-32 to also admit recomputed_work_growth_exponent)
// strictly exceeds its ratified ceiling -- exact, deterministic, and
// independent of measure.Demote's CoV quarantine, which exists for noisy
// metrics that a gate-eligible metric is not. Exactly equal to the ceiling
// is NOT a regression (FND-04's adjacency edge).
//
// [Rule 1 deviation, Phase 08 Plan 05 Task 2]: this branch used to test the
// single literal string "recomputed_work" rather than membership in
// QLT02GateEligibleMetrics(). That was equivalent while exactly one metric
// was ever gate-eligible, but widening Demote's own eligible set (this
// plan's own change) exposed the latent bug TestEvaluateBudgetAgreesWithDemote
// already existed to catch: a second gate-eligible metric NOT named
// "recomputed_work" would fall through to the else branch below, where
// requestedVerdict is computed from measure.Demote alone -- which knows
// nothing about row.ValueOrBound -- so EvaluateBudget would report
// VerdictBlocking for every low-CoV observation of that metric regardless
// of whether value was actually under or over its bound. Testing set
// membership instead of a literal closes that gap and keeps this function
// consulting the SAME chokepoint (QLT02GateEligibleMetrics()) Task 2
// widened, rather than inventing a third, independent one.
//
// Every other (non-gate-eligible) metric is routed through measure.Demote,
// which already refuses to return blocking for it -- the two mechanisms
// agree by construction on metric eligibility (TestEvaluateBudgetAgreesWithDemote)
// -- and an observation exceeding its loose bound is flagged, never
// blocking on its own.
func EvaluateBudget(row QLT02BudgetRow, observed measure.Summary, stageDeltas []protocol.StageTiming) (verdict string, flagged *Observation) {
	value := observed.P50

	if qlt02GateEligibleMetricSet()[row.Metric] {
		if value > row.ValueOrBound {
			return measure.VerdictBlocking, nil
		}
		return measure.VerdictObserved, nil
	}

	// Non-gate-eligible metrics can never be blocking on their own
	// (D-06-14/D-06-22): route through Demote, which refuses blocking for
	// any metric outside QLT02GateEligibleMetrics() regardless of what
	// verdict is requested.
	requestedVerdict := measure.Demote(measure.VerdictBlocking, row.Metric, observed, nil)

	if value <= row.ValueOrBound {
		return requestedVerdict, nil
	}

	delta, citable := stageWorkDelta(stageDeltas)
	return requestedVerdict, &Observation{
		Metric: row.Metric, Bound: row.ValueOrBound, Value: value,
		StageDelta: delta, Citable: citable,
	}
}

// BudgetRegressionDiagnostic renders the human-readable diagnostic a
// blocking recomputed_work regression must carry: the lane it was observed
// on, the ratified ceiling it exceeded, and the observed count that
// exceeded it (D-06-22, TestBlockingRegressionCitesStageWorkDelta).
func BudgetRegressionDiagnostic(laneID string, row QLT02BudgetRow, value int64) string {
	return fmt.Sprintf("lane %s: %s observed %d exceeds ratified ceiling %d on %s", laneID, row.Metric, value, row.ValueOrBound, row.MachineID)
}

// deterministicSummary builds a zero-noise measure.Summary around a single
// exact count -- the shape recomputed_work's own deterministic value takes
// when reported through EvaluateBudget's uniform Summary parameter, since
// an exact counter has no distribution to sample (D-06-14).
func deterministicSummary(value int64) measure.Summary {
	return measure.Summary{P50: value, P95: value, Mean: float64(value), StdDev: 0, CoV: 0, Count: 1}
}

// qlt02BudgetManifestFilePathForAST names the embedded manifest file this
// package's go/ast write-scan (TestNoFunctionWritesTheBudgetManifest, in
// the sibling _test.go file) checks for the absence of any write call
// against -- declared here, once, rather than duplicated as a string
// literal in the test.
const qlt02BudgetManifestFilePathForAST = "qlt02_budget_manifest.json"

// qlt02BudgetSourceFileForAST resolves this file's own path for the go/ast
// scan below, mirroring phase6SelfSourcePath's runtime.Caller(0) shape.
func qlt02BudgetSourceFileForAST() (string, error) {
	return qlt02RepoPath("internal/compiler/session/session_phase6_budget.go"), nil
}

// qlt02RepoPath resolves a repo-relative path the same way qlt01RepoPath
// does, so this scan works regardless of the test binary's working
// directory.
func qlt02RepoPath(relative string) string {
	return nat03CorpusPath(relative)
}

// AssertNoManifestWriteCalls parses session_phase6_budget.go and fails if
// any function in it calls os.WriteFile, os.Create, ioutil.WriteFile, or
// (*os.File).Write/WriteString against anything -- a structural proof that
// nothing in this file ever writes the budget manifest programmatically
// (D-06-18's "neither consults nor writes"; ratifying a new machine is a
// reviewed human act by design, never automated here). Exported so
// session_phase6_budget_test.go's TestNoFunctionWritesTheBudgetManifest can
// call it directly.
func AssertNoManifestWriteCalls() error {
	path, err := qlt02BudgetSourceFileForAST()
	if err != nil {
		return err
	}
	fileSet := token.NewFileSet()
	tree, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		return err
	}

	var found []string
	ast.Inspect(tree, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch selector.Sel.Name {
		case "WriteFile", "Create", "OpenFile", "Write", "WriteString":
			found = append(found, selector.Sel.Name)
		}
		return true
	})

	if len(found) != 0 {
		return fmt.Errorf("qlt02: session_phase6_budget.go contains forbidden manifest-write call(s): %v", found)
	}
	return nil
}

// qlt02ManifestNeverWritten is a defensive runtime check mirroring the
// structural AST guard above: it stats the checked-in manifest file before
// and after calling fn, refusing if its mtime or size changed. Exported so
// TestUndeclaredMachineNeverWritesManifest can wrap the observation-only
// verify path with it.
func qlt02ManifestNeverWritten(manifestPath string, fn func() error) error {
	before, statErr := os.Stat(manifestPath)
	if statErr != nil {
		return statErr
	}
	if err := fn(); err != nil {
		return err
	}
	after, statErr := os.Stat(manifestPath)
	if statErr != nil {
		return statErr
	}
	if before.ModTime() != after.ModTime() || before.Size() != after.Size() {
		return fmt.Errorf("qlt02: manifest at %s was modified", manifestPath)
	}
	return nil
}
