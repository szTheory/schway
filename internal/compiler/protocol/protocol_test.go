package protocol_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cache"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/protocol"
)

func TestOwnershipProjectionIdentityParity(t *testing.T) {
	result := protocol.New("run", protocol.StatusInvalid)
	result.Diagnostics = []diagnostic.Diagnostic{diagnostic.ErrorWithRepairs("ownership.use_after_move", diagnostic.Span{Start: 1, End: 2}, "moved value", nil, diagnostic.Repair{Kind: "use_transfer_target"})}
	result.Executions = []execution.Execution{{Schema: execution.Schema1, Outcome: execution.Outcome{Kind: "returned", Value: "01020304"}, Events: []execution.Event{{Schema: execution.Schema1, ID: "owned:event:0", Kind: "value.transferred", FunctionID: "owned:fn", SourcePlace: "owned:p0", TargetPlace: "owned:p1", TypeID: "owned:t0"}}, LiveResources: []string{}}}
	machine := result.Finalize()
	human, err := protocol.Human(result)
	if err != nil {
		t.Fatal(err)
	}
	if machine.Schema != "lang.command/1" || !strings.Contains(human, machine.ID) || !strings.Contains(human, machine.Diagnostics[0].ID) || !strings.Contains(human, machine.Executions[0].Events[0].ID) {
		t.Fatalf("human/JSON ownership identities diverged: result=%+v human=%q", machine, human)
	}
	for _, want := range []string{"source_place=owned:p0", "target_place=owned:p1", "type_id=owned:t0"} {
		if !strings.Contains(human, want) {
			t.Fatalf("human ownership event omitted %q: %q", want, human)
		}
	}
}

func TestResultIdentitySemanticSensitivity(t *testing.T) {
	base := protocol.New("run", protocol.StatusPass)
	base.Executions = []interp.Execution{{
		Schema:        interp.Schema,
		Outcome:       interp.Outcome{Kind: "returned", Value: "On"},
		Events:        []interp.Event{{Schema: interp.Schema, ID: "event:one", Kind: "function.returned", Input: "Off", Output: "On"}},
		LiveResources: []string{},
	}}
	first := base.Finalize()

	metricsChanged := base
	metricsChanged.Metrics.ElapsedNS = 999
	if got := metricsChanged.Finalize().ID; got != first.ID {
		t.Fatalf("operational metrics changed semantic result ID: first=%s got=%s", first.ID, got)
	}

	outcomeChanged := base
	outcomeChanged.Executions = append([]interp.Execution(nil), base.Executions...)
	outcomeChanged.Executions[0].Outcome.Value = "Off"
	if got := outcomeChanged.Finalize().ID; got == first.ID {
		t.Fatalf("semantic outcome change retained result ID %s", got)
	}
}

// TestHumanJSONProjectionParity is D-02-08's falsifier: Human and JSON share
// the identical self-consistency guarantee. Both converge their own
// output_bytes metric against the actual rendered length before returning,
// and both fail the same way (an error, not a silently unconverged output)
// when that convergence loop cannot stabilize. Before this fix, Human's
// loop fell through to an unconditional return on its fourth attempt even
// if length had not stabilized, while JSON's identical loop returned an
// error in that same case -- a caller comparing the two projections'
// self-reported sizes against reality could see JSON refuse to serve a
// broken result while Human silently served one.
func TestHumanJSONProjectionParity(t *testing.T) {
	for _, count := range []int{0, 1, 5, 20, 200} {
		result := protocol.New("check", protocol.StatusInvalid)
		for index := 0; index < count; index++ {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("ownership.use_after_move", diagnostic.Span{Start: index, End: index + 1}, "moved value"))
		}
		human, err := protocol.Human(result)
		if err != nil {
			t.Fatalf("count=%d: Human failed to converge: %v", count, err)
		}
		if !strings.Contains(human, fmt.Sprintf("output_bytes=%d", len(human))) {
			t.Fatalf("count=%d: Human's self-reported output_bytes does not match its own rendered length: %q", count, human)
		}

		encoded, err := protocol.JSON(result)
		if err != nil {
			t.Fatalf("count=%d: JSON failed to converge: %v", count, err)
		}
		if !strings.Contains(string(encoded), fmt.Sprintf(`"output_bytes":%d`, len(encoded))) {
			t.Fatalf("count=%d: JSON's self-reported output_bytes does not match its own rendered length: %s", count, encoded)
		}
	}
}

// TestExplainFoldsIntoResultIdentity is Task 1's own identity-fold
// falsifier for the net-new lang.explain/0 schema (D-06-04): a nil Explain
// omits the "explain" key entirely, and a non-nil Explain both appears in
// JSON and changes Result.Finalize().ID versus an otherwise-identical
// Result with Explain == nil.
func TestExplainFoldsIntoResultIdentity(t *testing.T) {
	base := protocol.New("explain", protocol.StatusPass)
	baseID := base.Finalize().ID

	encoded, err := protocol.JSON(base)
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if strings.Contains(string(encoded), `"explain":`) {
		t.Fatalf("nil Explain leaked an \"explain\" key: %s", encoded)
	}

	withExplain := base
	withExplain.Explain = &protocol.ExplainSummary{
		Schema: protocol.ExplainSchema, RootID: "diagnostic:x",
		Nodes: []protocol.ExplainNode{{ID: "diagnostic:x", Kind: "test.code", Availability: "available"}},
	}
	explainID := withExplain.Finalize().ID
	if explainID == baseID {
		t.Fatalf("Explain did not perturb Result.ID: %s", explainID)
	}

	encodedWithExplain, err := protocol.JSON(withExplain)
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if !strings.Contains(string(encodedWithExplain), `"schema":"`+protocol.ExplainSchema+`"`) {
		t.Fatalf("Explain summary schema missing from JSON: %s", encodedWithExplain)
	}
}

// TestCommandSchemaIsVersionOne pins 06-06's coordinated bump (D-06-31):
// protocol.New() now stamps every new Result with Schema1
// ("lang.command/1"), not the frozen Schema ("lang.command/0").
func TestCommandSchemaIsVersionOne(t *testing.T) {
	if protocol.Schema1 != "lang.command/1" {
		t.Fatalf("protocol.Schema1 = %q, want %q", protocol.Schema1, "lang.command/1")
	}
	if protocol.LaneSchema1 != "lang.verify-lane/1" {
		t.Fatalf("protocol.LaneSchema1 = %q, want %q", protocol.LaneSchema1, "lang.verify-lane/1")
	}
	result := protocol.New("verify", protocol.StatusPass)
	if result.Schema != protocol.Schema1 {
		t.Fatalf("protocol.New().Schema = %q, want protocol.Schema1 (%q)", result.Schema, protocol.Schema1)
	}
}

// TestSchemaZeroConstantsStillExist pins D-06-31's frozen-record half of the
// coordinated bump: protocol.Schema and protocol.LaneSchema remain declared
// and bound to their original /0 strings after the bump lands. A /0
// document's bytes are frozen once shipped and may only be superseded
// additively by a /1 -- deleting either constant as "dead code" would erase
// the ability to reproduce those already-published bytes, and this test
// exists to catch exactly that mistake.
func TestSchemaZeroConstantsStillExist(t *testing.T) {
	if protocol.Schema != "lang.command/0" {
		t.Fatalf("protocol.Schema = %q, want frozen %q", protocol.Schema, "lang.command/0")
	}
	if protocol.LaneSchema != "lang.verify-lane/0" {
		t.Fatalf("protocol.LaneSchema = %q, want frozen %q", protocol.LaneSchema, "lang.verify-lane/0")
	}
}

// setSentinelValue sets a single settable reflect.Value to a non-zero
// sentinel, recursing into slice-of-struct elements (e.g.
// protocol.StageTiming) so newly added structured fields do not require a
// bespoke case here. Used by setSentinelFields.
func setSentinelValue(t *testing.T, field reflect.Value) {
	t.Helper()
	if !field.CanSet() {
		return
	}
	switch field.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32:
		field.SetInt(7)
	case reflect.String:
		field.SetString("sentinel")
	case reflect.Bool:
		field.SetBool(true)
	case reflect.Slice:
		elemType := field.Type().Elem()
		switch elemType.Kind() {
		case reflect.String:
			field.Set(reflect.ValueOf([]string{"sentinel"}))
		case reflect.Struct:
			element := reflect.New(elemType).Elem()
			for index := 0; index < element.NumField(); index++ {
				setSentinelValue(t, element.Field(index))
			}
			slice := reflect.MakeSlice(field.Type(), 1, 1)
			slice.Index(0).Set(element)
			field.Set(slice)
		default:
			t.Fatalf("setSentinelValue: unhandled slice element kind: %s", elemType.Kind())
		}
	default:
		t.Fatalf("setSentinelValue: unhandled field kind: %s", field.Kind())
	}
}

// setSentinelFields sets every settable field of value (a pointer to a
// struct) to a non-zero sentinel via setSentinelValue. Fields whose name is
// in skip are left untouched. It is used by
// TestMetricsAndLaneFieldsExcludedFromIdentity to prove that non-identity
// fields cannot perturb Result.Finalize()'s ID no matter what value they
// hold (D-06-32).
func setSentinelFields(t *testing.T, value interface{}, skip map[string]bool) {
	t.Helper()
	elem := reflect.ValueOf(value).Elem()
	structType := elem.Type()
	for index := 0; index < elem.NumField(); index++ {
		field := elem.Field(index)
		name := structType.Field(index).Name
		if skip[name] {
			continue
		}
		setSentinelValue(t, field)
	}
}

// TestMetricsAndLaneFieldsExcludedFromIdentity pins D-06-32: no field of
// protocol.Metrics, and no field of protocol.Lane other than ID and Status,
// may ever reach Result.Finalize()'s identity struct. Two positive
// assertions (changing Lane.ID and changing Lane.Status each change
// Result.ID) prove the test is not inert -- without them it could pass
// vacuously if Finalize ignored lanes entirely.
func TestMetricsAndLaneFieldsExcludedFromIdentity(t *testing.T) {
	base := protocol.New("verify", protocol.StatusPass)
	baseID := base.Finalize().ID

	t.Run("Metrics", func(t *testing.T) {
		mutated := base
		setSentinelFields(t, &mutated.Metrics, nil)
		if got := mutated.Finalize().ID; got != baseID {
			t.Fatalf("a Metrics field reached Result.Finalize()'s identity: base=%s got=%s metrics=%+v", baseID, got, mutated.Metrics)
		}
	})

	t.Run("Lane non-identity fields", func(t *testing.T) {
		withLane := base
		withLane.Lanes = []protocol.Lane{{ID: "lane:one", Status: "pass"}}
		laneBaselineID := withLane.Finalize().ID

		mutated := withLane
		mutated.Lanes = append([]protocol.Lane(nil), withLane.Lanes...)
		setSentinelFields(t, &mutated.Lanes[0], map[string]bool{"ID": true, "Status": true})
		if got := mutated.Finalize().ID; got != laneBaselineID {
			t.Fatalf("a Lane field other than ID/Status reached Result.Finalize()'s identity: base=%s got=%s lane=%+v", laneBaselineID, got, mutated.Lanes[0])
		}

		t.Run("Lane.ID changes Result.ID", func(t *testing.T) {
			changedID := withLane
			changedID.Lanes = append([]protocol.Lane(nil), withLane.Lanes...)
			changedID.Lanes[0].ID = "lane:two"
			if got := changedID.Finalize().ID; got == laneBaselineID {
				t.Fatalf("changing Lane.ID did not change Result.ID: still %s", got)
			}
		})

		t.Run("Lane.Status changes Result.ID", func(t *testing.T) {
			changedStatus := withLane
			changedStatus.Lanes = append([]protocol.Lane(nil), withLane.Lanes...)
			changedStatus.Lanes[0].Status = "fail"
			if got := changedStatus.Finalize().ID; got == laneBaselineID {
				t.Fatalf("changing Lane.Status did not change Result.ID: still %s", got)
			}
		})
	})
}

// TestIdentityFieldEnumerationIsExhaustive reflects over protocol.Metrics
// and protocol.Lane and compares each struct's sorted field-name set against
// a literal expected slice declared here. It is the self-invalidating half
// of D-06-32: a field added to either struct without updating the expected
// slice below fails this test, forcing the reader back to
// TestMetricsAndLaneFieldsExcludedFromIdentity to reconfirm the new field is
// identity-excluded before 06-06 adds cache_status, selection_reason,
// machine_id, gate_verdict, cold_or_warm, stage_breakdown, and
// cache_inputs_reused_count.
func TestIdentityFieldEnumerationIsExhaustive(t *testing.T) {
	assertFieldSet := func(t *testing.T, label string, value interface{}, expected []string) {
		t.Helper()
		structType := reflect.TypeOf(value)
		var got []string
		for index := 0; index < structType.NumField(); index++ {
			got = append(got, structType.Field(index).Name)
		}
		sort.Strings(got)
		wanted := append([]string(nil), expected...)
		sort.Strings(wanted)
		if !reflect.DeepEqual(got, wanted) {
			t.Fatalf("%s field set changed: got %v, want %v -- if you added a field, add it to the sentinel loop in TestMetricsAndLaneFieldsExcludedFromIdentity and confirm it is identity-excluded before updating this expected set (D-06-32)", label, got, wanted)
		}
	}

	assertFieldSet(t, "protocol.Metrics", protocol.Metrics{}, []string{
		"ElapsedNS", "PeakRSSStatus", "PeakRSSBytes", "OutputBytes", "RecomputedWork",
		"CacheInputsReusedCount",
	})

	assertFieldSet(t, "protocol.Lane", protocol.Lane{}, []string{
		"Schema", "ID", "Status", "Controls", "RecomputedWork", "ElapsedNS", "PeakRSSStatus", "PeakRSSBytes", "OutputBytes",
		"CacheStatus", "SelectionReason", "MachineID", "GateVerdict", "ColdOrWarm", "StageBreakdown",
	})
}

// TestNewLaneAndMetricsFieldsAreAdditive is Behavior Test 1 (D-06-06's Task
// 2): a zero-valued Lane's seven /1 fields never appear in marshalled JSON,
// so /1 output for an unaffected lane is a superset-by-absence of /0.
func TestNewLaneAndMetricsFieldsAreAdditive(t *testing.T) {
	lane := protocol.Lane{ID: "lane:one", Status: "pass"}
	encoded, err := json.Marshal(lane)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		`"cache_status"`, `"selection_reason"`, `"machine_id"`,
		`"gate_verdict"`, `"cold_or_warm"`, `"stage_breakdown"`,
	} {
		if strings.Contains(string(encoded), key) {
			t.Fatalf("zero-valued Lane leaked %s: %s", key, encoded)
		}
	}

	metrics := protocol.Metrics{}
	encodedMetrics, err := json.Marshal(metrics)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encodedMetrics), `"cache_inputs_reused_count"`) {
		t.Fatalf("zero-valued Metrics leaked cache_inputs_reused_count: %s", encodedMetrics)
	}
}

// TestLaneVocabulariesAreClosed is Behavior Tests 2 and 3 (plus GateVerdict/
// ColdOrWarm): cache_status, selection_reason, gate_verdict, and
// cold_or_warm each accept only their declared closed vocabulary, and
// ValidateLaneVocabularies refuses any out-of-vocabulary value.
func TestLaneVocabulariesAreClosed(t *testing.T) {
	gotCacheStatuses := append([]string(nil), protocol.LaneCacheStatuses()...)
	sort.Strings(gotCacheStatuses)
	wantCacheStatuses := append([]string(nil), cache.CacheStatuses()...)
	sort.Strings(wantCacheStatuses)
	if !reflect.DeepEqual(gotCacheStatuses, wantCacheStatuses) {
		t.Fatalf("protocol.LaneCacheStatuses() = %v, want set-equal to cache.CacheStatuses() = %v", gotCacheStatuses, wantCacheStatuses)
	}
	for _, want := range []string{"artifact_reused", "artifact_recomputed", "not_cacheable", "unavailable"} {
		if !containsString(gotCacheStatuses, want) {
			t.Fatalf("protocol.LaneCacheStatuses() missing %q: %v", want, gotCacheStatuses)
		}
	}

	if got := protocol.LaneGateVerdicts(); !reflect.DeepEqual(got, []string{"blocking", "observed", "not_ratified"}) {
		t.Fatalf("protocol.LaneGateVerdicts() = %v, want [blocking observed not_ratified]", got)
	}

	if got := protocol.LaneSelectionReasons(); len(got) != 3 || !containsString(got, "selected") || !containsString(got, "deferred") || !containsString(got, "widened") {
		t.Fatalf("protocol.LaneSelectionReasons() = %v, want exactly the three declared forms", got)
	}

	if got := protocol.LaneColdOrWarmValues(); !reflect.DeepEqual(got, []string{"cold", "warm"}) {
		t.Fatalf("protocol.LaneColdOrWarmValues() = %v, want [cold warm]", got)
	}

	valid := protocol.Lane{ID: "lane:one", CacheStatus: "artifact_reused", SelectionReason: "selected", GateVerdict: "blocking", ColdOrWarm: "cold"}
	if err := protocol.ValidateLaneVocabularies(valid); err != nil {
		t.Fatalf("ValidateLaneVocabularies rejected an in-vocabulary lane: %v", err)
	}

	for _, invalid := range []protocol.Lane{
		{ID: "lane:x", CacheStatus: "hit"},
		{ID: "lane:x", CacheStatus: "miss"},
		{ID: "lane:x", SelectionReason: "maybe"},
		{ID: "lane:x", GateVerdict: "unknown"},
		{ID: "lane:x", ColdOrWarm: "lukewarm"},
	} {
		if err := protocol.ValidateLaneVocabularies(invalid); err == nil {
			t.Fatalf("ValidateLaneVocabularies accepted out-of-vocabulary lane: %+v", invalid)
		}
	}
}

func containsString(haystack []string, needle string) bool {
	for _, candidate := range haystack {
		if candidate == needle {
			return true
		}
	}
	return false
}

// TestCacheInputsReusedCountSharesUnitWithRecomputedWork is Behavior Test 4:
// CacheInputsReusedCount and RecomputedWork are both plain int counters in
// the same counted-unit currency, with no unit-conversion factor between
// them -- setting one to a value and reading it back must not scale it.
func TestCacheInputsReusedCountSharesUnitWithRecomputedWork(t *testing.T) {
	metrics := protocol.Metrics{RecomputedWork: 5, CacheInputsReusedCount: 5}
	if metrics.RecomputedWork != metrics.CacheInputsReusedCount {
		t.Fatalf("RecomputedWork=%d and CacheInputsReusedCount=%d diverged despite equal assignment -- both must be plain int counters", metrics.RecomputedWork, metrics.CacheInputsReusedCount)
	}
	metricsType := reflect.TypeOf(protocol.Metrics{})
	recomputed, _ := metricsType.FieldByName("RecomputedWork")
	reused, _ := metricsType.FieldByName("CacheInputsReusedCount")
	if recomputed.Type.Kind() != reflect.Int || reused.Type.Kind() != reflect.Int {
		t.Fatalf("RecomputedWork (%s) and CacheInputsReusedCount (%s) must both be plain int", recomputed.Type.Kind(), reused.Type.Kind())
	}
}
