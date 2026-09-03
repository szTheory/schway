package session_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestTransferRequiresTake(t *testing.T) {
	problem := ownershipDiagnostic(t, "implicit_noncopy.lang", "ownership.transfer_requires_take")
	assertCauseKinds(t, problem, "declared_here", "missing_ability", "place", "type")
	assertRepairKinds(t, problem, "insert_take")
}

func TestUseAfterMoveDiagnostic(t *testing.T) {
	problem := ownershipDiagnostic(t, "use_after_move.lang", "ownership.use_after_move")
	assertCauseKinds(t, problem, "declared_here", "moved_here", "place", "transfer_target", "type")
	assertRepairKinds(t, problem, "move_use_before_transfer", "use_transfer_target")
}

func TestMoveWhileBorrowedDiagnostic(t *testing.T) {
	problem := ownershipDiagnostic(t, "move_while_borrowed.lang", "ownership.move_while_borrowed")
	assertCauseKinds(t, problem, "borrow_created_here", "borrow_used_later", "loan", "owner", "type")
	assertRepairKinds(t, problem, "move_after_last_borrow_use")

	unused := []byte("module owned.unused_borrow\nexport { fn relay }\nfn relay(buffer: Buffer) -> Buffer {\n  let view = borrow buffer\n  let delivered = take buffer\n  delivered\n}\n")
	checked := session.Check(unused)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unused borrow did not end before move: %+v", checked.Diagnostics)
	}
}

func TestDiagnosticSchemaCompatibility(t *testing.T) {
	legacy := diagnostic.Error("syntax.example", diagnostic.Span{Start: 2, End: 3}, "legacy prose")
	if legacy.Schema != "lang.diagnostic/0" || len(legacy.Repairs) != 0 {
		t.Fatalf("legacy diagnostic changed schema: %+v", legacy)
	}
	for _, name := range []string{"implicit_noncopy.lang", "use_after_move.lang", "move_while_borrowed.lang"} {
		problem := ownershipDiagnostic(t, name, "")
		if problem.Schema != "lang.diagnostic/1" || len(problem.Repairs) == 0 {
			t.Fatalf("ownership diagnostic did not select /1 with repairs: %+v", problem)
		}
	}
}

func TestPhase1DiagnosticGoldenUnchanged(t *testing.T) {
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase1", "non_exhaustive.lang"))
	if err != nil || len(checked.Diagnostics) != 1 {
		t.Fatalf("phase 1 diagnostic setup failed: err=%v diagnostics=%+v", err, checked.Diagnostics)
	}
	encoded, err := json.Marshal(checked.Diagnostics[0])
	if err != nil {
		t.Fatal(err)
	}
	want := `{"schema":"lang.diagnostic/0","id":"diagnostic:f9582fb8c4ad9f90fbe75fa8","code":"match.non_exhaustive","severity":"error","primary_span":{"start":131,"end":158},"message":"match does not cover every alternative","causes":[{"kind":"missing_alternative","detail":"On"}]}`
	if string(encoded) != want {
		t.Fatalf("Phase 1 diagnostic golden changed:\ngot  %s\nwant %s", encoded, want)
	}
}

func ownershipDiagnostic(t *testing.T, fixture, code string) diagnostic.Diagnostic {
	t.Helper()
	path := testsupport.ProjectPath("testdata", "phase2", fixture)
	checked, err := session.CheckFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 1 || (code != "" && checked.Diagnostics[0].Code != code) {
		t.Fatalf("%s: got diagnostics %+v, want one %q", fixture, checked.Diagnostics, code)
	}
	executions, runDiagnostics, err := session.RunInterpreterFile(path)
	if err != nil || len(executions) != 0 || len(runDiagnostics) != 1 || runDiagnostics[0].ID != checked.Diagnostics[0].ID {
		t.Fatalf("invalid source reached execution or changed diagnostic: executions=%+v diagnostics=%+v err=%v", executions, runDiagnostics, err)
	}
	return checked.Diagnostics[0]
}

func assertCauseKinds(t *testing.T, problem diagnostic.Diagnostic, want ...string) {
	t.Helper()
	got := make([]string, len(problem.Causes))
	for index, cause := range problem.Causes {
		got[index] = cause.Kind
		if (strings.HasSuffix(cause.Kind, "_here") || cause.Kind == "borrow_used_later") && cause.Span == nil {
			t.Fatalf("cause %q omitted its source span: %+v", cause.Kind, problem)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cause kinds=%v want=%v in %+v", got, want, problem)
	}
}

func assertRepairKinds(t *testing.T, problem diagnostic.Diagnostic, want ...string) {
	t.Helper()
	got := make([]string, len(problem.Repairs))
	for index, repair := range problem.Repairs {
		got[index] = repair.Kind
	}
	if !sort.StringsAreSorted(got) || !reflect.DeepEqual(got, want) {
		t.Fatalf("repair kinds=%v want sorted %v in %+v", got, want, problem)
	}
}

func TestTogglePipeline(t *testing.T) {
	result, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase1", "toggle.lang"))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
	}
	if result.Program.Schema != "lang.core/0" || len(result.Program.Functions) != 1 {
		t.Fatalf("unexpected core: %+v", result.Program)
	}
	function := result.Program.Functions[0]
	if function.EntryPointID == "" || function.ReturnPointID == "" || function.Match.PointID == "" || len(function.Match.Arms) == 0 || function.Match.Arms[0].EdgeID == "" {
		t.Fatalf("typed core omitted stable point/edge identities: %+v", function)
	}
}

func TestOwnedTransferInterpreter(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
	checked, err := session.CheckFile(path)
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("owned check failed: err=%v diagnostics=%+v", err, checked.Diagnostics)
	}
	if checked.Program.Schema != "lang.core/1" || len(checked.Program.Functions) != 1 {
		t.Fatalf("unexpected owned core: %+v", checked.Program)
	}
	linear := checked.Program.Functions[0].Linear
	if linear == nil || checked.Program.Functions[0].Match != nil {
		t.Fatalf("owned function did not select exactly one linear body: %+v", checked.Program.Functions[0])
	}
	moves := 0
	for _, operation := range linear.Operations {
		if operation.Kind == "move" {
			moves++
		}
	}
	if moves != 1 {
		t.Fatalf("expected one explicit move, got %d in %+v", moves, linear.Operations)
	}
	executions, diagnostics, err := session.RunInterpreterFile(path)
	if err != nil || len(diagnostics) != 0 || len(executions) != 1 {
		t.Fatalf("owned interpreter failed: err=%v diagnostics=%+v executions=%+v", err, diagnostics, executions)
	}
	if executions[0].Schema != "lang.execution/1" || len(executions[0].Events) != 2 || executions[0].Events[0].Kind != "value.transferred" || executions[0].Events[1].Kind != "function.returned" {
		t.Fatalf("unexpected owned execution: %+v", executions[0])
	}
}

func TestImplicitByteCopy(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "implicit_copy.lang")
	checked, err := session.CheckFile(path)
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("copy check failed: err=%v diagnostics=%+v", err, checked.Diagnostics)
	}
	linear := checked.Program.Functions[0].Linear
	if linear == nil || len(linear.Operations) != 2 || linear.Operations[0].Kind != "copy" {
		t.Fatalf("bare Byte binding did not lower to copy: %+v", linear)
	}
	if linear.Operations[1].SourceID != checked.Program.Functions[0].Parameter.ID {
		t.Fatalf("implicit copy consumed its source: operations=%+v parameter=%+v", linear.Operations, checked.Program.Functions[0].Parameter)
	}
	executions, diagnostics, err := session.RunInterpreterFile(path)
	if err != nil || len(diagnostics) != 0 || len(executions) != 1 {
		t.Fatalf("copy interpreter failed: err=%v diagnostics=%+v executions=%+v", err, diagnostics, executions)
	}
	if executions[0].Schema != "lang.execution/1" || len(executions[0].Events) != 2 || executions[0].Events[0].Kind != "value.copied" {
		t.Fatalf("unexpected copy execution: %+v", executions[0])
	}
}

func TestSourceBoxPairAbilityFacts(t *testing.T) {
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase2", "ability_shapes.lang"))
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("ability shape check failed: err=%v diagnostics=%+v", err, checked.Diagnostics)
	}
	if checked.Program.Schema != core.Schema1 || len(checked.Program.Functions) != 3 {
		t.Fatalf("unexpected ability core: %+v", checked.Program)
	}

	type expectedFact struct {
		shape     core.TypeRef
		granted   []core.Ability
		witnesses []core.AbilityWitness
	}
	want := map[string]expectedFact{
		"keep_boxed_byte": {
			shape:     core.TypeRef{Constructor: "Box", Arguments: []core.TypeRef{{Constructor: "Byte", Arguments: []core.TypeRef{}}}},
			granted:   []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
			witnesses: []core.AbilityWitness{},
		},
		"keep_boxed_buffer": {
			shape:   core.TypeRef{Constructor: "Box", Arguments: []core.TypeRef{{Constructor: "Buffer", Arguments: []core.TypeRef{}}}},
			granted: []core.Ability{core.AbilityDrop, core.AbilitySend, core.AbilityEscape},
			witnesses: []core.AbilityWitness{
				{Ability: core.AbilityCopy, Path: []string{"Box.value", "Buffer"}},
				{Ability: core.AbilityShare, Path: []string{"Box.value", "Buffer"}},
			},
		},
		"keep_pair": {
			shape: core.TypeRef{Constructor: "Pair", Arguments: []core.TypeRef{
				{Constructor: "Byte", Arguments: []core.TypeRef{}},
				{Constructor: "Buffer", Arguments: []core.TypeRef{}},
			}},
			granted: []core.Ability{core.AbilityDrop, core.AbilitySend, core.AbilityEscape},
			witnesses: []core.AbilityWitness{
				{Ability: core.AbilityCopy, Path: []string{"Pair.right", "Buffer"}},
				{Ability: core.AbilityShare, Path: []string{"Pair.right", "Buffer"}},
			},
		},
	}
	for _, function := range checked.Program.Functions {
		expected, ok := want[function.Name]
		if !ok {
			t.Fatalf("unexpected function %q", function.Name)
		}
		if function.Linear == nil || len(function.Linear.Types) != 1 {
			t.Fatalf("%s omitted its materialized type fact: %+v", function.Name, function.Linear)
		}
		fact := function.Linear.Types[0]
		if !reflect.DeepEqual(fact.Shape, expected.shape) || !reflect.DeepEqual(fact.Abilities, expected.granted) || !reflect.DeepEqual(fact.NegativeWitnesses, expected.witnesses) {
			t.Fatalf("%s ability fact mismatch:\ngot=%+v\nwant shape=%+v abilities=%+v witnesses=%+v", function.Name, fact, expected.shape, expected.granted, expected.witnesses)
		}
		encoded, err := json.Marshal(fact)
		if err != nil {
			t.Fatal(err)
		}
		for _, abilityName := range []string{"copy", "drop", "share", "send", "escape"} {
			if bytes.Count(encoded, []byte(`"`+abilityName+`"`)) > 1 {
				t.Fatalf("%s serialized duplicate %s fact: %s", function.Name, abilityName, encoded)
			}
		}
	}
}

func TestSourceCannotGrantAbilityRoots(t *testing.T) {
	source := []byte("module forged.abilities\nexport { fn forge }\nfn forge(value: Buffer abilities { copy share }) -> Buffer { value }\n")
	checked := session.Check(source)
	if len(checked.Diagnostics) == 0 {
		t.Fatalf("source-written positive abilities reached checking: %+v", checked.Program)
	}
	if len(checked.Program.Functions) != 0 {
		t.Fatalf("malformed ability declaration manufactured core authority: %+v", checked.Program.Functions)
	}
	encoded, err := json.Marshal(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte(`"copy"`)) || bytes.Contains(encoded, []byte(`"share"`)) {
		t.Fatalf("source-written abilities entered serialized core: %s", encoded)
	}
}

func TestClosedBodyUnion(t *testing.T) {
	phase1, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase1", "toggle.lang"))
	if err != nil || len(phase1.Diagnostics) != 0 {
		t.Fatalf("phase 1 setup failed: err=%v diagnostics=%+v", err, phase1.Diagnostics)
	}
	zero := phase1.Program
	zero.Functions[0].Match = nil
	if _, err := interp.Run(zero, "toggle", "Off"); err == nil {
		t.Fatal("zero body variant reached execution")
	}
	dual := phase1.Program
	dual.Functions[0].Linear = &core.LinearBody{ID: dual.Functions[0].ID + ":linear"}
	if _, err := interp.Run(dual, "toggle", "Off"); err == nil {
		t.Fatal("dual body variants reached execution")
	}
}

func TestLinearIdentityStability(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	first := session.Check(source)
	second := session.Check(append([]byte("// unrelated offset-changing comment\n"), source...))
	if len(first.Diagnostics) != 0 || len(second.Diagnostics) != 0 {
		t.Fatalf("identity setup failed: first=%+v second=%+v", first.Diagnostics, second.Diagnostics)
	}
	left := first.Program.Functions[0]
	right := second.Program.Functions[0]
	if !reflect.DeepEqual(left.Linear, right.Linear) {
		t.Fatalf("linear identities depend on byte offsets or trivia:\nleft=%+v\nright=%+v", left.Linear, right.Linear)
	}
	wantPlaces := []string{left.ID + ":place:0", left.ID + ":place:1"}
	for index, want := range wantPlaces {
		if left.Linear.Places[index].ID != want {
			t.Fatalf("place %d uses incidental identity %q, want %q", index, left.Linear.Places[index].ID, want)
		}
	}
	if left.Linear.Types[0].ID != left.ID+":type:0" {
		t.Fatalf("type uses incidental identity %q", left.Linear.Types[0].ID)
	}
	for index, operation := range left.Linear.Operations {
		if operation.ID != fmt.Sprintf("%s:op:%d", left.ID, index) || operation.PointID != fmt.Sprintf("%s:point:linear:%d", left.ID, index) {
			t.Fatalf("operation %d lacks ordinal identities: %+v", index, operation)
		}
	}
}

func TestFeatureSpecificCoreExecutionSchemas(t *testing.T) {
	phase1Path := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	phase2Path := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
	phase1, err := session.CheckFile(phase1Path)
	if err != nil || len(phase1.Diagnostics) != 0 || phase1.Program.Schema != "lang.core/0" {
		t.Fatalf("phase 1 core schema changed: err=%v result=%+v", err, phase1)
	}
	phase2, err := session.CheckFile(phase2Path)
	if err != nil || len(phase2.Diagnostics) != 0 || phase2.Program.Schema != "lang.core/1" {
		t.Fatalf("phase 2 core schema missing: err=%v result=%+v", err, phase2)
	}
	oldExecution, diagnostics, err := session.RunInterpreterFile(phase1Path)
	if err != nil || len(diagnostics) != 0 || len(oldExecution) == 0 || oldExecution[0].Schema != "lang.execution/0" {
		t.Fatalf("phase 1 execution schema changed: err=%v diagnostics=%+v executions=%+v", err, diagnostics, oldExecution)
	}
	ownedExecution, diagnostics, err := session.RunInterpreterFile(phase2Path)
	if err != nil || len(diagnostics) != 0 || len(ownedExecution) != 1 || ownedExecution[0].Schema != "lang.execution/1" {
		t.Fatalf("phase 2 execution schema missing: err=%v diagnostics=%+v executions=%+v", err, diagnostics, ownedExecution)
	}
}

func TestMixedBodyVersionsFailAtCheckInEitherOrder(t *testing.T) {
	match := "fn toggle(state: Switch) -> Switch {\n  match state {\n    Off => On\n    On => Off\n  }\n}\n"
	linear := "fn retain(code: Byte) -> Byte {\n  let kept = code\n  kept\n}\n"
	prefix := "module mixed.bodies\nexport { type Switch fn toggle fn retain }\ndata Switch = | Off | On\n"
	for _, source := range []string{prefix + match + linear, prefix + linear + match} {
		checked := session.Check([]byte(source))
		if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "core.mixed_body_versions" {
			t.Fatalf("mixed module diagnostics=%+v", checked.Diagnostics)
		}
		if len(checked.Program.Functions) != 0 {
			t.Fatalf("mixed module emitted downstream core: %+v", checked.Program.Functions)
		}
	}
}

func TestNativeToggleO0O3(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	result, diagnostics, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("native run failed: err=%v diagnostics=%+v", err, diagnostics)
	}
	if len(result.O0.Pairs) != 2 || len(result.O3.Pairs) != 2 {
		t.Fatalf("unexpected native results: O0=%+v O3=%+v", result.O0, result.O3)
	}
	if !strings.Contains(result.CSource, "typedef enum LANG_SWITCH") || !strings.Contains(result.CSource, "switch (LANG_STATE)") {
		t.Fatalf("generated C is not reviewable S1 lowering:\n%s", result.CSource)
	}
}

func TestOwnedTransferInterpreterNative(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
	result, diagnostics, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("owned native run failed: err=%v diagnostics=%+v", err, diagnostics)
	}
	if len(result.Interpreter) != 1 || len(result.O0.Pairs) != 1 || len(result.O3.Pairs) != 1 {
		t.Fatalf("unexpected owned executions: %+v", result)
	}
	want := result.Interpreter[0]
	for _, actual := range []execution.Execution{result.O0.Pairs[0].Execution, result.O3.Pairs[0].Execution} {
		if !execution.Equal(want, actual) {
			t.Fatalf("owned semantic execution mismatch:\nwant=%+v\ngot =%+v", want, actual)
		}
	}
	golden, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase2", "owned_transfer.golden.c"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte(result.CSource), golden) {
		t.Fatalf("owned C golden changed:\n%s", result.CSource)
	}
	for _, required := range []string{"lang_record_event(\"value.transferred\"", "lang_record_event(\"function.returned\"", "lang_write_buffer_hex(&lang_value_delivered)"} {
		if !strings.Contains(result.CSource, required) {
			t.Fatalf("owned C does not derive execution from runtime state at %q:\n%s", required, result.CSource)
		}
	}
	if strings.Contains(result.CSource, `puts("{\"schema\":\"lang.execution/1\"`) {
		t.Fatalf("owned C still embeds a precomputed execution document:\n%s", result.CSource)
	}
	for _, forbidden := range []string{"restrict", "noalias", "malloc", "free("} {
		if strings.Contains(result.CSource, forbidden) {
			t.Fatalf("owned C makes forbidden %q claim:\n%s", forbidden, result.CSource)
		}
	}
}

type fakeNativeRunner struct {
	results []native.Result
	err     error
	calls   int
}

func (runner *fakeNativeRunner) Run(_ context.Context, _ string, _ string, _ []string) (native.Result, error) {
	if runner.err != nil {
		return native.Result{}, runner.err
	}
	result := runner.results[runner.calls]
	runner.calls++
	return result, nil
}

func TestOwnedEventReorderIsMismatch(t *testing.T) {
	expected := ownedInterpreterExecution(t)
	mutated := cloneExecution(t, expected)
	mutated.Events[0], mutated.Events[1] = mutated.Events[1], mutated.Events[0]
	assertOwnedMismatch(t, expected, mutated)
}

func TestOwnedExecutionFieldMutationMatrix(t *testing.T) {
	expected := ownedInterpreterExecution(t)
	operationalRunner := &fakeNativeRunner{err: &native.ToolError{Code: "native.invalid_execution", Err: errors.New("malformed child stdout")}}
	operational, err := session.RunNativeCommandFile(context.Background(), testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang"), operationalRunner)
	if err != nil || operational.Status != protocol.StatusOperational || protocol.ExitCode(operational.Status) != 3 || operational.Diagnostics[0].Code != "native.invalid_execution" {
		t.Fatalf("malformed native output classification changed: result=%+v err=%v", operational, err)
	}
	tests := []struct {
		name   string
		mutate func(*execution.Execution)
	}{
		{name: "schema", mutate: func(value *execution.Execution) { value.Schema = execution.Schema0 }},
		{name: "outcome kind", mutate: func(value *execution.Execution) { value.Outcome.Kind = "other" }},
		{name: "outcome value", mutate: func(value *execution.Execution) { value.Outcome.Value = "other" }},
		{name: "event schema", mutate: func(value *execution.Execution) { value.Events[0].Schema = execution.Schema0 }},
		{name: "event id", mutate: func(value *execution.Execution) { value.Events[0].ID = "other" }},
		{name: "event kind", mutate: func(value *execution.Execution) { value.Events[0].Kind = "other" }},
		{name: "function id", mutate: func(value *execution.Execution) { value.Events[0].FunctionID = "other" }},
		{name: "input", mutate: func(value *execution.Execution) { value.Events[0].Input = "physical-input" }},
		{name: "output", mutate: func(value *execution.Execution) { value.Events[0].Output = "physical-output" }},
		{name: "source place", mutate: func(value *execution.Execution) { value.Events[0].SourcePlace = "other" }},
		{name: "target place", mutate: func(value *execution.Execution) { value.Events[0].TargetPlace = "other" }},
		{name: "type id", mutate: func(value *execution.Execution) { value.Events[0].TypeID = "other" }},
		{name: "live resources", mutate: func(value *execution.Execution) { value.LiveResources = []string{"resource:other"} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := cloneExecution(t, expected)
			test.mutate(&mutated)
			assertOwnedMismatch(t, expected, mutated)
		})
	}
}

func TestOwnedBackendMutationIsMismatch(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
	runner := session.NewOwnedBackendMutationRunner(native.DefaultRunner())
	result, err := session.RunNativeCommandFile(context.Background(), path, runner)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != protocol.StatusMismatch || protocol.ExitCode(result.Status) != 4 {
		t.Fatalf("causal backend mutation status=%s exit=%d result=%+v", result.Status, protocol.ExitCode(result.Status), result)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "native.engine_mismatch" {
		t.Fatalf("causal backend mutation diagnostic=%+v", result.Diagnostics)
	}
	if got, want := runner.Optimizations(), []string{"-O0", "-O3"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("causal backend mutation optimizations=%v want=%v", got, want)
	}
}

func ownedInterpreterExecution(t *testing.T) execution.Execution {
	t.Helper()
	values, diagnostics, err := session.RunInterpreterFile(testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang"))
	if err != nil || len(diagnostics) != 0 || len(values) != 1 {
		t.Fatalf("owned interpreter setup failed: values=%+v diagnostics=%+v err=%v", values, diagnostics, err)
	}
	return values[0]
}

func cloneExecution(t *testing.T, value execution.Execution) execution.Execution {
	t.Helper()
	encoded, err := execution.CanonicalBytes(value)
	if err != nil {
		t.Fatal(err)
	}
	var cloned execution.Execution
	if err := json.Unmarshal(encoded, &cloned); err != nil {
		t.Fatal(err)
	}
	return cloned
}

func assertOwnedMismatch(t *testing.T, expected, mutated execution.Execution) {
	t.Helper()
	runner := &fakeNativeRunner{results: []native.Result{
		{Optimization: "-O0", Pairs: []native.Pair{{Input: "01020304", Execution: expected}}},
		{Optimization: "-O3", Pairs: []native.Pair{{Input: "01020304", Execution: mutated}}},
	}}
	path := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
	_, diagnostics, err := session.RunNativeFile(context.Background(), path, runner)
	if len(diagnostics) != 0 {
		t.Fatalf("semantic drift became source diagnostics: %+v", diagnostics)
	}
	var mismatch *session.EngineMismatch
	if !errors.As(err, &mismatch) {
		t.Fatalf("semantic drift was not an engine mismatch: %T %v", err, err)
	}
	commandRunner := &fakeNativeRunner{results: []native.Result{
		{Optimization: "-O0", Pairs: []native.Pair{{Input: "01020304", Execution: expected}}},
		{Optimization: "-O3", Pairs: []native.Pair{{Input: "01020304", Execution: mutated}}},
	}}
	result, commandErr := session.RunNativeCommandFile(context.Background(), path, commandRunner)
	if commandErr != nil || result.Status != protocol.StatusMismatch || protocol.ExitCode(result.Status) != 4 || result.Diagnostics[0].Code != "native.engine_mismatch" {
		t.Fatalf("semantic mismatch classification changed: result=%+v err=%v", result, commandErr)
	}
}

func TestNativeToolFailureIsOperational(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	_, diagnostics, err := session.RunNativeFile(context.Background(), path, native.Runner{ClangPath: testsupport.ProjectPath("missing-clang")})
	if len(diagnostics) != 0 {
		t.Fatalf("tool absence became source diagnostics: %+v", diagnostics)
	}
	var toolError *native.ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.tool_missing" {
		t.Fatalf("expected native.tool_missing, got %T %v", err, err)
	}
}

func TestNonExhaustiveMatch(t *testing.T) {
	result, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase1", "non_exhaustive.lang"))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "match.non_exhaustive" {
		t.Fatalf("expected match.non_exhaustive, got %+v", result.Diagnostics)
	}
	if got := result.Diagnostics[0].Causes[0].Detail; got != "On" {
		t.Fatalf("expected missing On, got %q", got)
	}
}

func TestCLIToggleTracer(t *testing.T) {
	result, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase1", "toggle.lang"))
	if err != nil || len(result.Diagnostics) != 0 {
		t.Fatalf("check tracer failed: err=%v diagnostics=%+v", err, result.Diagnostics)
	}
}

func TestInterpreterDeterministic(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	first, diagnostics, err := session.RunInterpreterFile(path)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("first run failed: err=%v diagnostics=%+v", err, diagnostics)
	}
	second, diagnostics, err := session.RunInterpreterFile(path)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("second run failed: err=%v diagnostics=%+v", err, diagnostics)
	}
	firstBytes, _ := json.Marshal(first)
	secondBytes, _ := json.Marshal(second)
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("nondeterministic executions:\n%s\n%s", firstBytes, secondBytes)
	}
	if len(first) != 2 || first[0].Outcome.Value != "On" || first[1].Outcome.Value != "Off" {
		t.Fatalf("unexpected toggle executions: %+v", first)
	}
}

func TestConcurrentReadOnlyCommands(t *testing.T) {
	projectRoot := testsupport.ProjectPath()
	fixture := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	before, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	beforeDigest := sha256.Sum256(before)

	binary := filepath.Join(t.TempDir(), "lang")
	build := exec.Command("go", "build", "-o", binary, "./cmd/lang")
	build.Dir = projectRoot
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}

	type commandResult struct {
		output []byte
		err    error
	}
	formatResults := make([]commandResult, 2)
	checkResults := make([]commandResult, 2)
	var group sync.WaitGroup
	for index := 0; index < 2; index++ {
		index := index
		group.Add(2)
		go func() {
			defer group.Done()
			formatResults[index].output, formatResults[index].err = exec.Command(binary, "format", fixture).CombinedOutput()
		}()
		go func() {
			defer group.Done()
			checkResults[index].output, checkResults[index].err = exec.Command(binary, "format", "--check", fixture).CombinedOutput()
		}()
	}
	group.Wait()

	for index, result := range append(formatResults, checkResults...) {
		if result.err != nil {
			t.Fatalf("concurrent command %d failed: %v\n%s", index, result.err, result.output)
		}
	}
	if !bytes.Equal(formatResults[0].output, formatResults[1].output) || !bytes.Equal(formatResults[0].output, before) {
		t.Fatalf("concurrent format output drifted:\nfirst=%q\nsecond=%q", formatResults[0].output, formatResults[1].output)
	}
	if !bytes.Equal(checkResults[0].output, checkResults[1].output) {
		t.Fatalf("concurrent check output drifted:\nfirst=%q\nsecond=%q", checkResults[0].output, checkResults[1].output)
	}
	after, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if afterDigest := sha256.Sum256(after); afterDigest != beforeDigest {
		t.Fatalf("read-only commands changed fixture digest: before=%x after=%x", beforeDigest, afterDigest)
	}
}

func TestVerifyCorpus(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase1"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("verify failed: status=%s diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	if len(result.Lanes) != 5 || result.Metrics.RecomputedWork == 0 || result.Metrics.ElapsedNS <= 0 {
		t.Fatalf("verify omitted work or observations: lanes=%+v metrics=%+v", result.Lanes, result.Metrics)
	}
	for _, lane := range result.Lanes {
		if lane.Status != "pass" || lane.RecomputedWork == 0 || lane.PeakRSSStatus == "" {
			t.Fatalf("incomplete lane: %+v", lane)
		}
	}
}

func TestVerifyMutationControls(t *testing.T) {
	corpus := testsupport.ProjectPath("testdata", "phase1")
	engine := session.VerifyCorpus(context.Background(), corpus, native.DefaultRunner(), session.VerifyOptions{ForceEngineMismatch: true})
	if engine.Status != protocol.StatusMismatch || protocol.ExitCode(engine.Status) != 4 || len(engine.Diagnostics) == 0 || engine.Diagnostics[0].Code != "native.engine_mismatch" {
		t.Fatalf("forced engine mismatch escaped: status=%s diagnostics=%+v", engine.Status, engine.Diagnostics)
	}
	stale := session.VerifyCorpus(context.Background(), corpus, native.DefaultRunner(), session.VerifyOptions{ForceStaleManifest: true})
	if stale.Status == protocol.StatusPass || protocol.ExitCode(stale.Status) == 0 || len(stale.Diagnostics) == 0 || stale.Diagnostics[0].Code != "evidence.source_mismatch" {
		t.Fatalf("forced stale manifest escaped: status=%s diagnostics=%+v", stale.Status, stale.Diagnostics)
	}
}

func TestVerifyPhase2ControlsAndWork(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase2"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("Phase 2 verify failed: status=%s diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	required := []string{
		"control:ownership.use_after_move",
		"control:ownership.move_while_borrowed",
		"control:ownership.transfer_requires_take",
		"control:ability.forged_copy",
		"control:core.duplicate_operation_id",
		"control:interpreter-o0-o3-owned",
		"control:evidence.core_mismatch",
		"control:backend.runtime_causality",
	}
	seen := map[string]bool{}
	for _, lane := range result.Lanes {
		if lane.Status != "pass" || lane.RecomputedWork == 0 || lane.PeakRSSStatus != "unavailable" {
			t.Fatalf("incomplete Phase 2 lane: %+v", lane)
		}
		for _, control := range lane.Controls {
			seen[control] = true
		}
	}
	for _, control := range required {
		if !seen[control] {
			t.Fatalf("Phase 2 verify omitted %s: %+v", control, result.Lanes)
		}
	}
	if !reflect.DeepEqual(result.ExpectedEscapes, []string{"escape:coordinated-source-core-lie"}) {
		t.Fatalf("expected escape was counted or omitted: %+v", result.ExpectedEscapes)
	}
	if seen[result.ExpectedEscapes[0]] {
		t.Fatalf("expected escape was counted as detected: %+v", result.Lanes)
	}
}
