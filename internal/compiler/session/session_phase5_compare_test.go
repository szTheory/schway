package session_test

import (
	"context"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase5CompareBaseline is a hand-constructed, self-consistent execution
// document used to seed exactly one axis of divergence at a time.
func phase5CompareBaseline() execution.Execution {
	return execution.Execution{
		Schema: execution.Schema1,
		Outcome: execution.Outcome{
			Kind:  execution.OutcomeReturned,
			Value: "42",
		},
		Events: []execution.Event{
			{Schema: execution.Schema1, ID: "e0", Kind: "value.borrowed", FunctionID: "f0", SourcePlace: "p0", TargetPlace: "p1", TypeID: "Buffer"},
			{Schema: execution.Schema1, ID: "e1", Kind: "resource.released", FunctionID: "f0", SourcePlace: "p1", TypeID: "Buffer"},
			{Schema: execution.Schema1, ID: "e2", Kind: "function.returned", FunctionID: "f0", SourcePlace: "p1", TypeID: "Buffer"},
		},
		LiveResources: []string{},
	}
}

// TestPhase5ComparatorNamesTheDivergingAxis seeds one divergence per axis
// (D-05-20) and asserts each produces a *Phase5EngineDisagreement naming
// that exact axis.
func TestPhase5ComparatorNamesTheDivergingAxis(t *testing.T) {
	assertAxis := func(t *testing.T, left, right execution.Execution, wantAxis string) {
		t.Helper()
		err := session.Phase5CompareEngines("seeded-fixture", map[string]execution.Execution{"interpreter": left, "O0": right})
		if err == nil {
			t.Fatal("expected a disagreement, got nil")
		}
		disagreement, ok := err.(*session.Phase5EngineDisagreement)
		if !ok {
			t.Fatalf("expected *session.Phase5EngineDisagreement, got %T (%v)", err, err)
		}
		if disagreement.Axis != wantAxis {
			t.Fatalf("expected axis %q, got %q (%v)", wantAxis, disagreement.Axis, disagreement)
		}
	}

	t.Run("terminal-outcome", func(t *testing.T) {
		left := phase5CompareBaseline()
		right := phase5CompareBaseline()
		right.Outcome.Value = "43"
		assertAxis(t, left, right, session.AxisTerminalOutcome)
	})

	t.Run("event-order", func(t *testing.T) {
		left := phase5CompareBaseline()
		right := phase5CompareBaseline()
		right.Events[0].TargetPlace = "p2"
		assertAxis(t, left, right, session.AxisEventOrder)
	})

	t.Run("resource-ledger", func(t *testing.T) {
		left := phase5CompareBaseline()
		right := phase5CompareBaseline()
		right.LiveResources = []string{"p1"}
		assertAxis(t, left, right, session.AxisResourceLedger)
	})

	t.Run("exit-status-signal", func(t *testing.T) {
		left := phase5CompareBaseline()
		right := phase5CompareBaseline()
		// Deliberately diverge ExitSignaled/ExitSignal while leaving
		// Outcome.Kind identical on both sides -- proving this axis is
		// compared independently of axis:terminal-outcome, not merely
		// derived from it.
		right.ExitSignaled = true
		right.ExitSignal = "SIGABRT"
		assertAxis(t, left, right, session.AxisExitStatusSignal)
	})

	t.Run("diagnostic-id", func(t *testing.T) {
		left := diagnostic.Error("check.reject_probe", diagnostic.Span{Start: 0, End: 1}, "rejected")
		right := diagnostic.Error("check.reject_probe", diagnostic.Span{Start: 2, End: 3}, "rejected")
		err := session.Phase5CompareDiagnosticIDs("seeded-fixture", map[string]diagnostic.Diagnostic{"check": left, "corevalidate": right})
		if err == nil {
			t.Fatal("expected a disagreement, got nil")
		}
		disagreement, ok := err.(*session.Phase5EngineDisagreement)
		if !ok {
			t.Fatalf("expected *session.Phase5EngineDisagreement, got %T (%v)", err, err)
		}
		if disagreement.Axis != session.AxisDiagnosticID {
			t.Fatalf("expected axis %q, got %q (%v)", session.AxisDiagnosticID, disagreement.Axis, disagreement)
		}
	})
}

// TestPhase5ComparatorComparesOkPayloadAndErrAlternative proves
// axis:terminal-outcome makes the ok payload and the err-edge ADT
// alternative explicit comparison targets -- SC1's wording left the payload
// implicit; D-05-20 makes it explicit.
func TestPhase5ComparatorComparesOkPayloadAndErrAlternative(t *testing.T) {
	t.Run("same-kind-different-ok-payload", func(t *testing.T) {
		left := phase5CompareBaseline()
		right := phase5CompareBaseline()
		right.Outcome.Value = "99"
		err := session.Phase5CompareEngines("payload-fixture", map[string]execution.Execution{"interpreter": left, "O0": right})
		var disagreement *session.Phase5EngineDisagreement
		if err == nil {
			t.Fatal("expected a disagreement for differing ok payloads, got nil")
		}
		disagreement, _ = err.(*session.Phase5EngineDisagreement)
		if disagreement == nil || disagreement.Axis != session.AxisTerminalOutcome {
			t.Fatalf("expected axis:terminal-outcome, got %v", err)
		}
	})

	t.Run("same-kind-different-err-alternative", func(t *testing.T) {
		left := phase5CompareBaseline()
		left.Outcome.Kind = execution.OutcomeTypedFailure
		left.Outcome.Value = "AcquireError.OpenFailed"
		right := phase5CompareBaseline()
		right.Outcome.Kind = execution.OutcomeTypedFailure
		right.Outcome.Value = "AcquireError.OtherFailed"
		err := session.Phase5CompareEngines("err-fixture", map[string]execution.Execution{"interpreter": left, "O0": right})
		disagreement, _ := err.(*session.Phase5EngineDisagreement)
		if disagreement == nil || disagreement.Axis != session.AxisTerminalOutcome {
			t.Fatalf("expected axis:terminal-outcome for differing err-edge alternatives, got %v", err)
		}
	})

	t.Run("identical-payloads-agree", func(t *testing.T) {
		left := phase5CompareBaseline()
		right := phase5CompareBaseline()
		if err := session.Phase5CompareEngines("agree-fixture", map[string]execution.Execution{"interpreter": left, "O0": right}); err != nil {
			t.Fatalf("expected identical documents to agree, got %v", err)
		}
	})
}

// TestRejectProgramDiagnosticIDsAgree proves a reject-program's diagnostic
// ID is identical across independent derivations, under
// control:diagnostic.reject_program_id_equivalence -- distinct from
// control:interpreter-o0-o3(-lto) since reject-programs never execute.
func TestRejectProgramDiagnosticIDsAgree(t *testing.T) {
	// Two independent derivations of the SAME reject-program's diagnostic:
	// session.Check run twice via session.Check on identical source bytes.
	// check.Program is source-position-keyed and deterministic, so its own
	// diagnostic ID must be stable across independent invocations -- the
	// exact equivalence D-05-20 requires be asserted under its own control
	// ID rather than silently assumed.
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_call_target_not_foreign.schway"))
	if err != nil {
		t.Fatal(err)
	}
	firstChecked := session.Check(source)
	secondChecked := session.Check(source)
	if len(firstChecked.Diagnostics) == 0 || len(secondChecked.Diagnostics) == 0 {
		t.Fatalf("expected reject-program diagnostics, got first=%v second=%v", firstChecked.Diagnostics, secondChecked.Diagnostics)
	}
	diagnostics := map[string]diagnostic.Diagnostic{
		"check-run-1": firstChecked.Diagnostics[0],
		"check-run-2": secondChecked.Diagnostics[0],
	}
	if err := session.Phase5CompareDiagnosticIDs("foreign_call_target_not_foreign.schway", diagnostics); err != nil {
		t.Fatalf("expected identical diagnostic IDs across independent derivations, got %v", err)
	}
	if firstChecked.Diagnostics[0].ID == "" {
		t.Fatal("expected a non-empty diagnostic ID")
	}
}

// TestPhase4ComparatorUnchanged proves Phase4CompareThreeEngines's own
// behaviour is unchanged by this plan's axis-widening, driven through its
// existing corpus (T-05-22: peer-not-fork, session.go stays untouched).
func TestPhase4ComparatorUnchanged(t *testing.T) {
	corpus := testsupport.ProjectPath("testdata", "phase4")
	runner := native.DefaultRunner()
	program, functionName, err := session.Phase4CheckedProgram(corpus, "acquire_three_success.schway")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.Phase4ThreeEngineDifferential(context.Background(), "acquire_three_success.schway", program, functionName, "7", runner, native.ExpectValue); err != nil {
		requirePhase16M004Refusal(t, err, "foreign")
		return
	}
	t.Fatal("expected terminal Phase 16 M004 foreign refusal")
}

// mirrorExecutionWithExtraField is TestUnroutedFieldFailsTheRoutingTest's
// local, test-only mirror of execution.Execution plus one extra field --
// proving the routing check's failure mode without requiring a real schema
// change to demonstrate it (D-05-21).
type mirrorExecutionWithExtraField struct {
	Schema           string
	Outcome          execution.Outcome
	Events           []execution.Event
	LiveResources    []string
	AllocatorAddress string
	WallClockNanos   int64
	ExitSignaled     bool
	ExitSignal       string
	ExtraField       string
}

// reachableFieldPathsForTest mirrors session.go's own unexported
// reachableFieldPaths walk for the test-only mirror struct above, since the
// production walker is unexported and this specific struct is deliberately
// never part of production code.
func reachableFieldPathsForTest(t reflect.Type, prefix string) []string {
	var paths []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		path := prefix + "." + field.Name
		fieldType := field.Type
		for fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Array {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct {
			paths = append(paths, reachableFieldPathsForTest(fieldType, path)...)
			continue
		}
		paths = append(paths, path)
	}
	return paths
}

func setDiff(a, b []string) []string {
	inB := make(map[string]bool, len(b))
	for _, v := range b {
		inB[v] = true
	}
	var diff []string
	for _, v := range a {
		if !inB[v] {
			diff = append(diff, v)
		}
	}
	sort.Strings(diff)
	return diff
}

// TestComparisonFieldRoutingIsExhaustive walks, by reflection, every
// exported field transitively reachable from execution.Execution and
// asserts the set is EXACTLY the union of Phase5ComparedComparisonFields
// and Phase5ExcludedComparisonFields -- no field in neither, no listed
// field that no longer exists. This is D-05-21's fail-closed guarantee: a
// NEW field defaults to unrouted, and unrouted fails the build.
func TestComparisonFieldRoutingIsExhaustive(t *testing.T) {
	actual := reachableFieldPathsForTest(reflect.TypeOf(execution.Execution{}), "Execution")
	union := append(append([]string(nil), session.Phase5ComparedComparisonFields...), session.Phase5ExcludedComparisonFields...)

	if unrouted := setDiff(actual, union); len(unrouted) > 0 {
		t.Fatalf("field(s) reachable from execution.Execution are UNROUTED -- route each to Phase5ComparedComparisonFields or Phase5ExcludedComparisonFields: %v", unrouted)
	}
	if stale := setDiff(union, actual); len(stale) > 0 {
		t.Fatalf("routed field path(s) no longer exist on execution.Execution -- remove from Phase5ComparedComparisonFields/Phase5ExcludedComparisonFields: %v", stale)
	}
}

// TestInvocationFieldsAreCompared proves the /2 occurrence and call-edge
// identity facts are part of whole-event equality, not merely listed in the
// routing table.
func TestInvocationFieldsAreCompared(t *testing.T) {
	baseline := phase5CompareBaseline()
	baseline.Schema = execution.Schema2
	baseline.Events[0].Schema = execution.Schema2
	baseline.Events[0].Invocation = "inv:entry:f0"
	baseline.Events[0].CalleeFunctionID = "callee:one"

	for _, mutate := range []func(*execution.Event){
		func(event *execution.Event) { event.Invocation = "inv:entry:other" },
		func(event *execution.Event) { event.CalleeFunctionID = "callee:other" },
		func(event *execution.Event) { event.Occurrence = 1 },
	} {
		left, right := baseline, baseline
		left.Events = append([]execution.Event(nil), baseline.Events...)
		right.Events = append([]execution.Event(nil), baseline.Events...)
		mutate(&right.Events[0])
		err := session.Phase5CompareEngines("invocation-fields", map[string]execution.Execution{"interpreter": left, "O0": right})
		disagreement, ok := err.(*session.Phase5EngineDisagreement)
		if !ok || disagreement.Axis != session.AxisEventOrder {
			t.Fatalf("mutating /2 identity field = %v, want axis:event-order disagreement", err)
		}
	}
}

// TestSchema2ComparisonRequiresPeerVerdict proves that pair equality cannot
// bless matching forged /2 evidence: the peer is consulted before comparison.
func TestSchema2ComparisonRequiresPeerVerdict(t *testing.T) {
	program, entryName := phase11CheckedFixture(t, "multi_function_diamond_call.schway")
	entry := phase11EntryFunction(t, program, entryName)
	engines := phase11RunFourTiers(t, context.Background(), program, entryName, phase11EntryInput(t, entry.Parameter.Type))
	for name, document := range engines {
		document.Events = append([]execution.Event(nil), document.Events...)
		document.Events[0].Invocation = "not-an-invocation"
		engines[name] = document
	}
	err := session.Phase5CompareProgramEngines("peer-verdict", program, engines)
	if err == nil || !strings.Contains(err.Error(), "executionpeer.malformed_invocation") {
		t.Fatalf("matching forged documents passed comparison without a named peer refusal: %v", err)
	}
}

// TestPhase5CompareProgramEnginesPreservesLegacySchemas proves that the
// program-aware wrapper delegates untouched /0 and /1 documents to the
// historical comparator without consulting the Schema 2 peer.
func TestPhase5CompareProgramEnginesPreservesLegacySchemas(t *testing.T) {
	for _, schema := range []string{execution.Schema0, execution.Schema1} {
		t.Run(schema, func(t *testing.T) {
			peerCalled := false
			restore := session.SetPhase5Schema2PeerValidatorForTest(func(core.Program, execution.Execution) error {
				peerCalled = true
				t.Error("legacy document unexpectedly consulted the Schema 2 peer")
				return nil
			})
			t.Cleanup(restore)

			baseline := phase5CompareBaseline()
			baseline.Schema = schema
			for index := range baseline.Events {
				baseline.Events[index].Schema = schema
			}

			matching := map[string]execution.Execution{"interpreter": baseline, "O0": baseline}
			if want, got := session.Phase5CompareEngines("legacy-match", matching), session.Phase5CompareProgramEngines("legacy-match", core.Program{}, matching); want != nil || got != nil {
				t.Fatalf("matching %s documents disagree: legacy=%v wrapper=%v", schema, want, got)
			}

			divergent := baseline
			divergent.Outcome.Value = "legacy-divergence"
			engines := map[string]execution.Execution{"interpreter": baseline, "O0": divergent}
			want := session.Phase5CompareEngines("legacy-divergence", engines)
			got := session.Phase5CompareProgramEngines("legacy-divergence", core.Program{}, engines)
			wantDisagreement, wantOK := want.(*session.Phase5EngineDisagreement)
			gotDisagreement, gotOK := got.(*session.Phase5EngineDisagreement)
			if !wantOK || !gotOK || wantDisagreement.Axis != gotDisagreement.Axis {
				t.Fatalf("divergent %s documents disagree on comparison axis: legacy=%v wrapper=%v", schema, want, got)
			}
			if peerCalled {
				t.Fatal("legacy comparison consulted the Schema 2 peer")
			}
		})
	}
}

// TestUnroutedFieldFailsTheRoutingTest demonstrates the routing check's red
// path using a local mirror struct with one extra field, so the failure
// mode is proven without requiring a real schema change (D-05-21).
func TestUnroutedFieldFailsTheRoutingTest(t *testing.T) {
	actual := reachableFieldPathsForTest(reflect.TypeOf(mirrorExecutionWithExtraField{}), "Execution")
	union := append(append([]string(nil), session.Phase5ComparedComparisonFields...), session.Phase5ExcludedComparisonFields...)
	unrouted := setDiff(actual, union)
	if len(unrouted) != 1 || unrouted[0] != "Execution.ExtraField" {
		t.Fatalf("expected exactly [Execution.ExtraField] reported as unrouted, got %v", unrouted)
	}
}

// TestExcludedFieldsAreNeverRead asserts Phase5CompareEngines produces NO
// disagreement when two executions differ ONLY in fields listed in
// Phase5ExcludedComparisonFields, varying an address-shaped
// (AllocatorAddress) and a timestamp-shaped (WallClockNanos) field.
func TestExcludedFieldsAreNeverRead(t *testing.T) {
	left := phase5CompareBaseline()
	left.AllocatorAddress = "0xdeadbeef"
	left.WallClockNanos = 1
	right := phase5CompareBaseline()
	right.AllocatorAddress = "0xfeedface"
	right.WallClockNanos = 999999
	if err := session.Phase5CompareEngines("excluded-fixture", map[string]execution.Execution{"interpreter": left, "O0": right}); err != nil {
		t.Fatalf("expected excluded fields to never produce a disagreement, got %v", err)
	}
}

// duplicateUnion is a compile-time sanity check that
// Phase5ComparedComparisonFields and Phase5ExcludedComparisonFields share no
// entry.
func TestComparedAndExcludedFieldsHaveNoDuplicate(t *testing.T) {
	seen := make(map[string]bool)
	for _, field := range session.Phase5ComparedComparisonFields {
		if seen[field] {
			t.Fatalf("duplicate entry in Phase5ComparedComparisonFields: %s", field)
		}
		seen[field] = true
	}
	for _, field := range session.Phase5ExcludedComparisonFields {
		if seen[field] {
			t.Fatalf("field %s appears in both Phase5ComparedComparisonFields and Phase5ExcludedComparisonFields", field)
		}
		seen[field] = true
	}
}
