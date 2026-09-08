package session_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/interp/interptestdirect"
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

// TestLoanLivenessIsTransitive pins the source-level law that one hop of
// indirection does not expire a loan. Both shapes previously checked clean and
// ran to completion, so the shipped move_while_borrowed control was weaker
// than the gate implied.
func TestLoanLivenessIsTransitive(t *testing.T) {
	problem := ownershipDiagnostic(t, "reborrow_while_moved.lang", "ownership.move_while_borrowed")
	assertCauseKinds(t, problem, "borrow_created_here", "borrow_used_later", "loan", "owner", "type")
	assertRepairKinds(t, problem, "move_after_last_borrow_use")

	copyOfLoan := []byte("module owned.copy_of_loan\nexport { fn relay }\nfn relay(code: Byte) -> Byte {\n  let view = borrow code\n  let alias = view\n  let delivered = take code\n  let observed = alias\n  delivered\n}\n")
	checked := session.Check(copyOfLoan)
	if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "ownership.move_while_borrowed" {
		t.Fatalf("copy of a loan expired the loan early: %+v", checked.Diagnostics)
	}

	// The dual: a transitively derived loan with no use after the move must
	// still end, so transitivity does not degrade into blocking every move.
	expired := []byte("module owned.expired_reborrow\nexport { fn relay }\nfn relay(code: Byte) -> Byte {\n  let view = borrow code\n  let review = borrow view\n  let observed = review\n  let delivered = take code\n  delivered\n}\n")
	if checked := session.Check(expired); len(checked.Diagnostics) != 0 {
		t.Fatalf("transitive loan over-blocked a move after its last use: %+v", checked.Diagnostics)
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

// TestSourceBoxPairAbilityFacts closes D-02-09/D-07 (03-02-03): Box/Pair
// shapes type-check but have no native execution lowering this phase, so
// `check` now refuses them at check time with a causal, repair-bearing
// diagnostic per function rather than admitting them into the checked core
// (where every downstream engine used to die spanless at exit 3). This test
// used to assert the OPPOSITE — that these three functions checked clean and
// their derived ability facts landed in the returned core.Program — because
// closing this gap is exactly what 03-02-03 changes. See
// TestAbilityFactsSurviveExecutionRejection for proof that the underlying
// ability derivation these functions exercise is unaffected by the new
// execution-admission gate.
func TestSourceBoxPairAbilityFacts(t *testing.T) {
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase2", "ability_shapes.lang"))
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 3 {
		t.Fatalf("want 3 unexecutable-shape diagnostics (one per function), got %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) != 0 {
		t.Fatalf("a rejected function must not be admitted into the checked core: %+v", checked.Program.Functions)
	}
	wantConstructors := map[string]bool{"Box": false, "Pair": false}
	for _, problem := range checked.Diagnostics {
		if problem.Code != "check.unexecutable_shape" {
			t.Fatalf("unexpected diagnostic code %q: %+v", problem.Code, problem)
		}
		if problem.Primary == (diagnostic.Span{}) {
			t.Fatalf("unexecutable-shape diagnostic carries no span: %+v", problem)
		}
		if len(problem.Repairs) == 0 {
			t.Fatalf("unexecutable-shape diagnostic carries no repair: %+v", problem)
		}
		for _, cause := range problem.Causes {
			if cause.Kind == "constructor" {
				if _, known := wantConstructors[cause.Detail]; known {
					wantConstructors[cause.Detail] = true
				}
			}
		}
	}
	for constructor, seen := range wantConstructors {
		if !seen {
			t.Fatalf("expected a diagnostic naming constructor %q among %+v", constructor, checked.Diagnostics)
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

// TestForeignCallInterpreterNative proves the Phase 4 tracer fixture's
// terminal outcome, ordered events, and live-resource state agree across
// the interpreter and Clang-built native code at both -O0 and -O3
// (D-04-01/D-04-04/D-04-05/D-04-10). RunNative already asserts this
// agreement internally (an EngineMismatch is a returned error); this test
// additionally inspects the shape directly so a future regression that
// weakens RunNative's own comparison is still caught here.
func TestForeignCallInterpreterNative(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.lang")
	result, diagnostics, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("native run failed: err=%v diagnostics=%+v", err, diagnostics)
	}
	if len(result.Interpreter) != 1 || len(result.O0.Pairs) != 1 || len(result.O3.Pairs) != 1 {
		t.Fatalf("unexpected pair counts: interpreter=%d O0=%d O3=%d", len(result.Interpreter), len(result.O0.Pairs), len(result.O3.Pairs))
	}
	interpreted := result.Interpreter[0]
	if interpreted.Outcome.Kind != "returned" || interpreted.Outcome.Value != "7" {
		t.Fatalf("interpreter outcome = %+v", interpreted.Outcome)
	}
	if len(interpreted.Events) != 2 || interpreted.Events[0].Kind != "foreign.called" || interpreted.Events[1].Kind != "function.returned" {
		t.Fatalf("interpreter events = %+v", interpreted.Events)
	}
	if len(interpreted.LiveResources) != 0 {
		t.Fatalf("interpreter live resources = %+v, want empty", interpreted.LiveResources)
	}
	for _, engineResult := range []struct {
		name      string
		execution execution.Execution
	}{{"O0", result.O0.Pairs[0].Execution}, {"O3", result.O3.Pairs[0].Execution}} {
		if !execution.Equal(interpreted, engineResult.execution) {
			t.Fatalf("%s execution disagrees with interpreter:\ninterpreter: %+v\nnative:      %+v", engineResult.name, interpreted, engineResult.execution)
		}
	}
	if !strings.Contains(result.CSource, "extern _LANG_lang_res_open_result _LANG_lang_res_open") {
		t.Fatalf("generated C does not declare the frozen foreign symbol:\n%s", result.CSource)
	}
	if strings.Contains(result.CSource, "lang_foreign_resource_private") {
		t.Fatalf("generated C must never reference the frozen TU's private header:\n%s", result.CSource)
	}
}

// TestForeignPolicyValueInjectionRefusedFromSource is 04-13's end-to-end
// tracer (04-VERIFICATION.md gap 2b, FFI-01): an ordinary, honest .lang
// source file whose sole `allocator:` policy value carries a
// comment-terminator payload is refused at SOURCE ADMISSION by
// session.Check, before any core artifact is ever produced -- so the
// downstream emitters (cgen.EmitForeignHeader/EmitForeignConformance) are
// unreachable for this source, not merely guarded once reached.
func TestForeignPolicyValueInjectionRefusedFromSource(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_policy_value_injection.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) == 0 {
		t.Fatalf("expected at least one diagnostic, got none (program: %+v)", checked.Program)
	}
	first := checked.Diagnostics[0]
	if first.Code != "check.foreign_policy_value_unsafe" {
		t.Fatalf("expected check.foreign_policy_value_unsafe, got %+v", checked.Diagnostics)
	}
	if first.Primary.Start == 0 && first.Primary.End == 0 {
		t.Fatalf("expected a non-zero span locating the offending policy, got %+v", first.Primary)
	}
	if first.Primary.Start < 0 || first.Primary.End > len(source) {
		t.Fatalf("expected the diagnostic span to lie within the fixture's byte length %d, got %+v", len(source), first.Primary)
	}
	if len(checked.Program.Functions) != 0 {
		t.Fatalf("expected no checked function to be produced, got %+v", checked.Program.Functions)
	}
}

// TestReleaseInterpreterNative proves RES-01/D-04-07's three-acquisition
// success fixture agrees between the interpreter and Clang-built native code
// at both -O0 and -O3 on terminal outcome, ordered events (including the
// three reverse-order resource.released events), and live-resource state
// (empty -- every acquired resource was released).
func TestReleaseInterpreterNative(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.lang")
	result, diagnostics, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("native run failed: err=%v diagnostics=%+v", err, diagnostics)
	}
	interpreted := result.Interpreter[0]
	if interpreted.Outcome.Kind != "returned" || interpreted.Outcome.Value != "7" {
		t.Fatalf("interpreter outcome = %+v", interpreted.Outcome)
	}
	if len(interpreted.LiveResources) != 0 {
		t.Fatalf("interpreter live resources = %+v, want empty", interpreted.LiveResources)
	}
	var released []string
	for _, event := range interpreted.Events {
		if event.Kind == "resource.released" {
			released = append(released, event.SourcePlace)
		}
	}
	if len(released) != 3 {
		t.Fatalf("expected 3 resource.released events, got %+v", released)
	}
	for _, engineResult := range []struct {
		name      string
		execution execution.Execution
	}{{"O0", result.O0.Pairs[0].Execution}, {"O3", result.O3.Pairs[0].Execution}} {
		if !execution.Equal(interpreted, engineResult.execution) {
			t.Fatalf("%s execution disagrees with interpreter:\ninterpreter: %+v\nnative:      %+v", engineResult.name, interpreted, engineResult.execution)
		}
		if len(engineResult.execution.LiveResources) != 0 {
			t.Fatalf("%s live resources = %+v, want empty", engineResult.name, engineResult.execution.LiveResources)
		}
	}
}

// TestReleaseOmissionMutationIsMismatch proves control:resource.release_omitted
// (D-04-07/Pitfall 2): deleting one generated line bearing lang:release-site
// (the event AND the runtime ledger decrement together, since both are
// emitted on the same line) leaves a resource observably live at
// termination, surfaced through the same pre-existing "returned outcome
// requires empty live_resources" hard-reject Pitfall 3 protects
// (native.invalid_execution) -- attacking the EMITTER's own generated C, a
// different artifact than the transposition mutation below attacks.
func TestReleaseOmissionMutationIsMismatch(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.lang")
	// RunNative's special ForeignSources wiring only fires for a bare
	// native.Runner value (a mutation-runner wrapper opts out of it by
	// design, per its own doc comment), so the frozen foreign TU must be
	// linked in here explicitly, before wrapping.
	inner := native.DefaultRunner()
	inner.ForeignSources = []string{native.ForeignResourceSourcePath()}
	runner := session.NewReleaseOmissionMutationRunner(inner)
	_, _, err := session.RunNativeFile(context.Background(), path, runner)
	if err == nil {
		t.Fatal("expected the release omission mutation to be detected, got no error")
	}
	var toolError *native.ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.invalid_execution" {
		t.Fatalf("expected native.invalid_execution (a live resource on a returned outcome), got %v", err)
	}
	if len(runner.Optimizations()) == 0 {
		t.Fatal("expected the omission runner to have recorded at least one optimization attempt")
	}
}

// TestReleaseTranspositionMutationIsMismatch proves
// control:resource.release_order_transposed (D-04-07/Pitfall 1): exchanging
// two emitted OpRelease operations in the three-acquisition fixture's core
// artifact is rejected by corevalidate's independent rederivation, naming
// the release_order_mismatch code -- attacking the CHECKER's materialized
// order, a different artifact than the omission mutation above attacks.
func TestReleaseTranspositionMutationIsMismatch(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	if valid := corevalidate.Validate(checked.Program); !valid.Valid {
		t.Fatalf("valid core rejected before mutation: %+v", valid)
	}
	mutated, err := session.TransposeReleaseOrder(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	result := corevalidate.Validate(mutated)
	if result.Valid || result.Problems[0].Code != "core.release_order_mismatch" {
		t.Fatalf("expected core.release_order_mismatch, got %+v", result)
	}
}

// TestTwoAcquisitionTranspositionIsIndistinguishable converts CONTEXT.md's
// sharpest research finding into a standing invariant: transposing two
// elements never changes the SET they belong to, at any count N -- so an
// oracle that only compares release SETS (not order) cannot ever falsify a
// transposition, for two acquisitions or three. This project's OWN
// corevalidate differential is deliberately STRONGER than that naive oracle
// -- it compares full ordered sequences (TestReleaseOrderMutationMatrix's
// "moved" case already proves this catches a transposition at N=3), and this
// test proves the SAME is true at N=2. The set-equality blind spot below is
// therefore demonstrated against an illustrative naive oracle, not against
// production code -- production code never has this blind spot, which is
// exactly why a three-acquisition fixture (not two) is this project's own
// minimum falsifying witness for RES-01 (Pitfall 1): a fixture too small to
// need this project's own stronger, order-sensitive check would not prove
// the check does anything a weaker one couldn't.
func TestTwoAcquisitionTranspositionIsIndistinguishable(t *testing.T) {
	releaseSetEqual := func(a, b []string) bool {
		setA, setB := map[string]bool{}, map[string]bool{}
		for _, id := range a {
			setA[id] = true
		}
		for _, id := range b {
			setB[id] = true
		}
		return reflect.DeepEqual(setA, setB)
	}
	original := []string{"B", "A"}
	transposed := []string{"A", "B"}
	if !releaseSetEqual(original, transposed) {
		t.Fatal("expected a two-element transposition to be set-equal -- this IS the finding being demonstrated")
	}

	twoAcquisitionSource := `module phase4.two_acquisition_probe

export {
  fn main
}

foreign C {

  fn lang_res_open(request: Byte) -> Byte {
    unwind: forbidden
    nonlocal_exit: forbidden
    allocator: "libc_malloc"
    fails: AcquireError
  }
}

data AcquireError =
  | OpenFailed

fn main(request: Byte) -> Byte {
  let a = try lang_res_open(request)
  let b = try lang_res_open(request)
  request
}
`
	checked := session.Check([]byte(twoAcquisitionSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	mutated, err := session.TransposeReleaseOrder(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	result := corevalidate.Validate(mutated)
	if result.Valid {
		t.Fatal("expected corevalidate to reject the transposed two-acquisition core; this project's own oracle is order-sensitive even at N=2")
	}
	if result.Problems[0].Code != "core.release_order_mismatch" {
		t.Fatalf("expected core.release_order_mismatch, got %+v", result)
	}
}

// TestReleaseMutationsAttackDifferentArtifacts asserts the omission runner's
// input is generated C (never the frozen foreign translation unit) and that
// no release mutation runner ever opens a path under native/.
func TestReleaseMutationsAttackDifferentArtifacts(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.lang")
	result, diagnostics, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("native run failed: err=%v diagnostics=%+v", err, diagnostics)
	}
	if !strings.Contains(result.CSource, "lang:release-site") {
		t.Fatal("generated C carries no release-site marker for the omission runner to locate")
	}
	if strings.Contains(result.CSource, "lang_foreign_resource_private") {
		t.Fatal("generated C must never reference the frozen TU's private header")
	}
	// The omission runner's Run signature takes a cSource string (generated
	// C), never a file path -- structurally, it cannot open a path under
	// native/ at all. TransposeReleaseOrder's signature takes a core.Program,
	// not a file path either. Both are asserted here by the type system: if
	// either runner's signature ever grows a path parameter, this file fails
	// to compile against the call sites above, which pass no path.
	var _ session.NativeRunner = session.NewReleaseOmissionMutationRunner(native.DefaultRunner())
}

// TestVerifyPhase4ReleaseControls proves the two release controls
// (control:resource.release_omitted, control:resource.release_order_transposed)
// are visible to the Phase 4 verify gate, each with nonzero recomputed work.
func TestVerifyPhase4ReleaseControls(t *testing.T) {
	corpus := testsupport.ProjectPath("testdata", "phase4")
	result := session.VerifyCorpus(context.Background(), corpus, native.DefaultRunner(), session.VerifyOptions{})
	if result.Status != protocol.StatusPass {
		t.Fatalf("Phase 4 release verify failed: status=%s diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	for _, required := range []string{"control:resource.release_order_transposed", "control:resource.release_omitted"} {
		found := false
		for _, lane := range result.Lanes {
			for _, control := range lane.Controls {
				if control == required {
					found = true
					if lane.RecomputedWork == 0 {
						t.Fatalf("%s lane carries zero recomputed work", required)
					}
				}
			}
		}
		if !found {
			t.Fatalf("required control %s is missing from the verify result", required)
		}
	}
}

func TestNativeIdentifiersRemainCollisionFree(t *testing.T) {
	tests := []string{
		"module collision.locals\nexport { fn keep }\nfn keep(code: Byte) -> Byte {\n  let α = code\n  let β = code\n  let __1 = code\n  __1\n}\n",
		"module collision.shadow\nexport { fn keep }\nfn keep(code: Byte) -> Byte {\n  let value = code\n  let value = code\n  value\n}\n",
		"module collision.variants\nexport { type Thing fn thing }\ndata Thing = | a | A | A_1\nfn thing(value: Thing) -> Thing {\n  match value {\n    a => A\n    A => A_1\n    A_1 => a\n  }\n}\n",
		"module collision.cross_category\nexport { type Thing fn thing_LANG_THING }\ndata Thing = | thing | other\nfn thing_LANG_THING(value: Thing) -> Thing {\n  match value {\n    thing => other\n    other => thing\n  }\n}\n",
	}
	for _, source := range tests {
		result, diagnostics, err := session.RunNative(context.Background(), []byte(source), native.DefaultRunner())
		if err != nil || len(diagnostics) != 0 || len(result.O0.Pairs) == 0 || len(result.O3.Pairs) == 0 {
			t.Fatalf("collision-safe native path failed: err=%v diagnostics=%+v result=%+v", err, diagnostics, result)
		}
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

func TestOwnedBackendMutationRunnerConcurrentUse(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	runner := session.NewOwnedBackendMutationRunner(native.DefaultRunner())
	var group sync.WaitGroup
	for index := 0; index < 4; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			_, _ = runner.Run(context.Background(), generated, "-O0", []string{"01020304"})
			_ = runner.Optimizations()
		}()
	}
	group.Wait()
	if len(runner.Optimizations()) != 4 {
		t.Fatalf("recorded optimizations=%v", runner.Optimizations())
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
	fixture := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	before, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	beforeDigest := sha256.Sum256(before)

	binary := testsupport.BuildCLI(t)

	type commandResult struct {
		output []byte
		stderr []byte
		exit   int
		err    error
	}
	// RunCLIErr rather than RunCLI: t.Fatalf may only be called from the
	// goroutine running the test, so the spawns report typed errors that the
	// test goroutine inspects after the WaitGroup drains.
	run := func(target *commandResult, arguments ...string) {
		result, err := testsupport.RunCLIErr(context.Background(), binary, nil, arguments...)
		target.output, target.stderr, target.exit, target.err = result.Stdout, result.Stderr, result.Exit, err
	}
	formatResults := make([]commandResult, 2)
	checkResults := make([]commandResult, 2)
	var group sync.WaitGroup
	for index := 0; index < 2; index++ {
		index := index
		group.Add(2)
		go func() {
			defer group.Done()
			run(&formatResults[index], "format", fixture)
		}()
		go func() {
			defer group.Done()
			run(&checkResults[index], "format", "--check", fixture)
		}()
	}
	group.Wait()

	for index, result := range append(append([]commandResult(nil), formatResults...), checkResults...) {
		if result.err != nil {
			t.Fatalf("concurrent command %d failed: %v", index, result.err)
		}
		if result.exit != 0 {
			t.Fatalf("concurrent command %d exited %d\n%s%s", index, result.exit, result.output, result.stderr)
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
		"control:ownership.move_while_reborrowed",
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

const branchLoanControlSource = `module owned.branch_loan_control

export {
  type Switch
  fn choose
}

data Switch =
  | On
  | Off

fn choose(flag: Switch) -> Switch {
  match flag {
    On => {
      let view = borrow flag
      let noted = view
      let moved = take flag
      moved
    }
    Off => {
      let held = take flag
      held
    }
  }
}
`

const branchNoLoanControlSource = `module owned.branch_no_loan_control

export {
  type Switch
  fn choose
}

data Switch =
  | On
  | Off

fn choose(flag: Switch) -> Switch {
  match flag {
    On => {
      let held = take flag
      held
    }
    Off => {
      let held = take flag
      held
    }
  }
}
`

func checkedBranchProgram(t *testing.T, source string) core.Program {
	t.Helper()
	checked := session.Check([]byte(source))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

// TestLoanEndpointMutationMatrix (T-03-06/T-03-14) is the endpoint mutation
// matrix the plan's own must-have requires: a validated branch program's
// declared loan endpoint moved to a different block, dropped entirely, or
// invented for a loan that does not exist, is rejected in all three
// variants by exactly core.loan_endpoint_mismatch. This is the SAME
// matrix corevalidate_endpoint_test.go's TestLoanEndpointMismatchRejected
// proves at the corevalidate package boundary; asserted again here, against
// the session-level Check(...) entry point, as the control this package's
// own verify lane (BorrowedLoanEndpointControlLane) relies on.
func TestLoanEndpointMutationMatrix(t *testing.T) {
	baseline := checkedBranchProgram(t, branchLoanControlSource)
	if result := corevalidate.Validate(baseline); !result.Valid {
		t.Fatalf("baseline branch-loan program rejected: %+v", result.Problems)
	}
	linear := baseline.Functions[0].Linear
	if len(linear.LoanEndpoints) == 0 {
		t.Fatalf("expected the control fixture to declare at least one loan endpoint: %+v", linear)
	}

	tests := []struct {
		name string
		edit func(*core.Program)
	}{
		{name: "moved to the join block", edit: func(program *core.Program) {
			endpoints := program.Functions[0].Linear.LoanEndpoints
			endpoints[0].BlockID = program.Functions[0].Linear.Blocks[len(program.Functions[0].Linear.Blocks)-1].ID
		}},
		{name: "dropped entirely", edit: func(program *core.Program) {
			program.Functions[0].Linear.LoanEndpoints = program.Functions[0].Linear.LoanEndpoints[1:]
		}},
		{name: "invented for a loan that does not exist", edit: func(program *core.Program) {
			linear := program.Functions[0].Linear
			linear.LoanEndpoints = append(append([]core.LoanEndpoint(nil), linear.LoanEndpoints...), core.LoanEndpoint{
				ID: linear.LoanEndpoints[0].ID + ":invented", LoanID: program.Functions[0].ID + ":loan:invented",
				Kind: "point", BlockID: linear.LoanEndpoints[0].BlockID, AfterOperationID: linear.LoanEndpoints[0].AfterOperationID,
			})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(baseline)
			if err != nil {
				t.Fatal(err)
			}
			var mutated core.Program
			if err := json.Unmarshal(encoded, &mutated); err != nil {
				t.Fatal(err)
			}
			test.edit(&mutated)
			result := corevalidate.Validate(mutated)
			if result.Valid {
				t.Fatalf("mutation %q was accepted", test.name)
			}
			if len(result.Problems) == 0 || result.Problems[0].Code != "core.loan_endpoint_mismatch" {
				t.Fatalf("mutation %q code=%v want=core.loan_endpoint_mismatch", test.name, result.Problems)
			}
		})
	}
}

// TestBorrowedLaneRecordsFailureStatus proves BorrowedLoanEndpointControlLane
// uses the Phase 1 addLane shape (PATTERNS I-1): a control that cannot fire
// (here, a program with no loan endpoint to mutate against) still returns a
// real lane with status "fail" and a nonzero work count, rather than
// vanishing the way Phase 2's verifyOwnedCorpus addLane would on any
// failure path.
func TestBorrowedLaneRecordsFailureStatus(t *testing.T) {
	noLoan := checkedBranchProgram(t, branchNoLoanControlSource)
	lane := session.BorrowedLoanEndpointControlLane(noLoan)
	if lane.Status != "fail" {
		t.Fatalf("expected a failing lane for a program with no loan endpoint, got status=%q lane=%+v", lane.Status, lane)
	}
	if lane.ID == "" {
		t.Fatal("failing control vanished instead of being recorded as a lane")
	}
	if lane.RecomputedWork == 0 {
		t.Fatalf("failing lane recorded zero work, losing partial-work evidence: %+v", lane)
	}
	if len(lane.Controls) != 0 {
		t.Fatalf("a failing lane must not claim a control it never observed: %+v", lane)
	}

	honest := checkedBranchProgram(t, branchLoanControlSource)
	passLane := session.BorrowedLoanEndpointControlLane(honest)
	if passLane.Status != "pass" {
		t.Fatalf("expected the honest, loan-bearing program to drive the control to pass: %+v", passLane)
	}
	if passLane.RecomputedWork == 0 {
		t.Fatalf("passing lane recorded zero work: %+v", passLane)
	}
	if len(passLane.Controls) != 1 || passLane.Controls[0] != "control:core.loan_endpoint_mismatch" {
		t.Fatalf("passing lane did not claim control:core.loan_endpoint_mismatch: %+v", passLane)
	}
}

// TestPathOracleLaneRecordsControl proves session.PathOracleDisagreementLane
// (03-05, ROADMAP criterion 1's third mechanism) records nonzero-work lanes
// on both paths: a program with no branch loan fails without claiming a
// control it never observed, and the honest, loan-bearing fixture passes
// and claims exactly control:cfg.path_oracle_disagreement.
func TestPathOracleLaneRecordsControl(t *testing.T) {
	noLoan := checkedBranchProgram(t, branchNoLoanControlSource)
	lane := session.PathOracleDisagreementLane(noLoan)
	if lane.Status != "fail" {
		t.Fatalf("expected a failing lane for a program with no branch loan, got status=%q lane=%+v", lane.Status, lane)
	}
	if lane.ID == "" {
		t.Fatal("failing control vanished instead of being recorded as a lane")
	}
	if len(lane.Controls) != 0 {
		t.Fatalf("a failing lane must not claim a control it never observed: %+v", lane)
	}

	honest := checkedBranchProgram(t, branchLoanControlSource)
	passLane := session.PathOracleDisagreementLane(honest)
	if passLane.Status != "pass" {
		t.Fatalf("expected the honest, loan-bearing program to drive the control to pass: %+v", passLane)
	}
	if passLane.RecomputedWork == 0 {
		t.Fatalf("passing lane recorded zero work: %+v", passLane)
	}
	if len(passLane.Controls) != 1 || passLane.Controls[0] != "control:cfg.path_oracle_disagreement" {
		t.Fatalf("passing lane did not claim control:cfg.path_oracle_disagreement: %+v", passLane)
	}
}

// TestVerifyPhase3ControlsAndWork is Task 03-07-02's falsifier for the Phase
// 3 verify path: dispatching on borrowed_view.lang, verifyBorrowedCorpus
// requires every named Phase 3 control fail-closed, with nonzero work on
// every lane and both expected escapes (corevalidate's and originvalidate's)
// surfaced, never reported as detected controls.
func TestVerifyPhase3ControlsAndWork(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase3"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("Phase 3 verify failed: status=%s diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	required := []string{
		"control:ownership.exclusive_conflict",
		"control:ownership.exclusive_move",
		"control:core.loan_endpoint_mismatch",
		"control:cfg.path_oracle_disagreement",
		"control:origin.understated_summary",
		"control:origin.impossible_summary",
		"control:origin.stale_summary",
		"control:origin.omitted_summary",
		"control:origin.mixed_access_chain",
		"control:origin.multi_arm_omitted",
		"control:origin.multi_arm_access_conflict",
	}
	found := make(map[string]bool)
	for _, lane := range result.Lanes {
		if lane.Status != "pass" || lane.RecomputedWork == 0 {
			t.Fatalf("incomplete lane: %+v", lane)
		}
		for _, control := range lane.Controls {
			found[control] = true
		}
	}
	for _, control := range required {
		if !found[control] {
			t.Fatalf("missing required Phase 3 control %s: lanes=%+v", control, result.Lanes)
		}
	}
	expected := map[string]bool{}
	for _, escape := range result.ExpectedEscapes {
		expected[escape] = true
	}
	if !expected["escape:coordinated-source-core-lie"] || !expected["escape:coordinated-frontend-summary-lie"] {
		t.Fatalf("Phase 3 verify omitted an expected escape: %+v", result.ExpectedEscapes)
	}
}

// TestVerifyPhase4ForeignControls proves both Phase 4 admission refusals
// (D-04-16/D-04-02) are visible to the gate as required negative controls
// with nonzero recomputed work, and that the tracer fixture itself still
// admits cleanly.
// TestLayoutMutationIsCompileTimeRefusal proves control:foreign.layout_mismatch
// (D-04-11/D-11): compiling the generated conformance unit against the
// transposed frozen fixture is refused at compile time under the existing
// -Werror flag set, reporting the distinct "native.conformance_failed" code.
func TestLayoutMutationIsCompileTimeRefusal(t *testing.T) {
	runner := session.LayoutMutationRunner{
		Runner:      native.DefaultRunner(),
		Contract:    session.LayoutProbeContract(),
		FixturePath: testsupport.ProjectPath("testdata", "phase4", "foreign_layout_mismatch.golden.c"),
	}
	err := runner.Run(context.Background())
	if err == nil {
		t.Fatal("expected the transposed fixture to be refused at compile time")
	}
	var toolErr *native.ToolError
	if !errors.As(err, &toolErr) || toolErr.Code != "native.conformance_failed" {
		t.Fatalf("expected native.conformance_failed, got %v", err)
	}
}

// TestLayoutMutationAttacksFrozenFixtureOnly proves the layout mutation
// runner's own input path is the fixture, and never a generated source:
// pointing FixturePath at a correctly-ordered (untransposed) private header
// compiles cleanly, proving the runner's verdict depends solely on
// FixturePath's own content. The runner type itself carries no field of a
// generated-source shape (Runner, Contract, FixturePath only), so "never
// opens a generated source" is a structural property, not merely a runtime
// behavior demonstrated here.
func TestLayoutMutationAttacksFrozenFixtureOnly(t *testing.T) {
	correctPath := filepath.Join(t.TempDir(), "lang_foreign_layout_probe_correct.h")
	correctHeader := []byte(`#ifndef LANG_FOREIGN_LAYOUT_PROBE_PRIVATE_H
#define LANG_FOREIGN_LAYOUT_PROBE_PRIVATE_H
typedef struct lang_foreign_layout_probe_block {
  unsigned char first;
  unsigned char second;
} lang_foreign_layout_probe_block;
#endif
`)
	if err := os.WriteFile(correctPath, correctHeader, 0o600); err != nil {
		t.Fatal(err)
	}
	runner := session.LayoutMutationRunner{Runner: native.DefaultRunner(), Contract: session.LayoutProbeContract(), FixturePath: correctPath}
	if err := runner.Run(context.Background()); err != nil {
		t.Fatalf("expected the untransposed fixture to conform, got %v", err)
	}
	value := reflect.ValueOf(runner)
	if value.NumField() != 3 {
		t.Fatalf("LayoutMutationRunner grew an unexpected field: %+v", runner)
	}
}

// TestNoUnprovenAttributesEmitted proves control:foreign.no_unproven_attributes
// (D-04-13): every emitted C artifact for the tracer and release-lifecycle
// fixtures -- the compiled program, the generated header, and the generated
// conformance unit (D-04-12's three named inspectable layers) -- plus their
// lang.foreign/0 sidecar manifests, carries no banned optimizer-visible
// attribute token, and each manifest's emitted_attributes field is present
// and empty (not omitted). EmitForeignHeader/EmitForeignConformance are
// scanned here too (WR-01): the header is the artifact a human reviewer is
// most likely to actually read, and skipping it would let a banned token
// injected there go completely undetected.
func TestNoUnprovenAttributesEmitted(t *testing.T) {
	for _, fixture := range []string{"foreign_acquire_one.lang", "acquire_three_success.lang"} {
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", fixture))
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("%s: unexpected diagnostics: %+v", fixture, checked.Diagnostics)
		}
		cSource, err := cgen.Emit(checked.Program)
		if err != nil {
			t.Fatalf("%s: %v", fixture, err)
		}
		manifest, err := cgen.EmitForeignManifest(checked.Program)
		if err != nil {
			t.Fatalf("%s: %v", fixture, err)
		}
		header, err := cgen.EmitForeignHeader(checked.Program)
		if err != nil {
			t.Fatalf("%s: %v", fixture, err)
		}
		conformance, err := cgen.EmitForeignConformance(checked.Program, native.ForeignResourcePrivateHeaderPath())
		if err != nil {
			t.Fatalf("%s: %v", fixture, err)
		}
		if found := cgen.ScanForBannedAttributes(cSource, manifest, header, conformance); len(found) != 0 {
			t.Fatalf("%s: found banned attribute tokens %v", fixture, found)
		}
		if !strings.Contains(manifest, `"emitted_attributes":[]`) {
			t.Fatalf("%s: emitted_attributes is not a present, empty array:\n%s", fixture, manifest)
		}
	}
}

// TestAttributeInjectionIntoHeaderOnlyMakesControlFail is WR-01's dedicated
// mutation-kill test: it injects a banned token into EmitForeignHeader's
// output ALONE, leaving emitLinearForeign's own extern declaration (the
// compiled-program C that TestAttributeInjectionMakesControlFail already
// covers) untouched, proving the header artifact is independently scanned
// rather than only incidentally covered because the two extern-declaration
// format strings happen to collide.
func TestAttributeInjectionIntoHeaderOnlyMakesControlFail(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	cSource, err := cgen.Emit(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	header, err := cgen.EmitForeignHeader(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if found := cgen.ScanForBannedAttributes(cSource, header); len(found) != 0 {
		t.Fatalf("real emitted C/header already contains a banned token: %v", found)
	}
	injectedHeader := strings.Replace(header, "extern", "extern __attribute__((malloc)) restrict ", 1)
	if injectedHeader == header {
		t.Fatal("injection site not found in emitted header")
	}
	// The compiled program's own extern declaration is untouched.
	if found := cgen.ScanForBannedAttributes(cSource); len(found) != 0 {
		t.Fatalf("untouched compiled-program C unexpectedly flagged: %v", found)
	}
	// But scanning header output alone (as the production lane now does)
	// must catch the injected tokens.
	if found := cgen.ScanForBannedAttributes(injectedHeader); len(found) == 0 {
		t.Fatal("expected the header-only injected banned tokens to be detected")
	}
}

// TestAttributeInjectionIntoConformanceOnlyMakesControlFail is 04-09's
// dedicated mutation-kill test for the third D-04-12 inspectable layer.
// EmitForeignConformance embeds EmitForeignHeader's full output verbatim
// before writing its own added text (the private-header #include line and
// the per-field _Static_assert triples), so injecting anywhere in the
// shared prefix would prove only what
// TestAttributeInjectionIntoHeaderOnlyMakesControlFail already proves. This
// test isolates the conformance unit's OWN added text -- everything after
// the embedded header -- and injects there alone, proving the conformance
// layer is independently scanned rather than only incidentally covered by
// the header's coverage. This is the same coincidental-overlap trap
// 04-VERIFICATION.md gap 2 named, one layer further in.
func TestAttributeInjectionIntoConformanceOnlyMakesControlFail(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	cSource, err := cgen.Emit(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	header, err := cgen.EmitForeignHeader(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	conformance, err := cgen.EmitForeignConformance(checked.Program, native.ForeignResourcePrivateHeaderPath())
	if err != nil {
		t.Fatal(err)
	}
	// The real emitted artifacts are all clean first -- a fixture that
	// already contains a banned token would make the whole test vacuous.
	if found := cgen.ScanForBannedAttributes(cSource, header, conformance); len(found) != 0 {
		t.Fatalf("real emitted C/header/conformance already contains a banned token: %v", found)
	}
	// Isolate the conformance unit's own added text: everything after the
	// embedded header. Fail loudly if the header is not found inside the
	// conformance output -- that would mean the embedding assumption this
	// test is built on no longer holds and the test must be re-derived,
	// not silently weakened.
	headerIndex := strings.Index(conformance, header)
	if headerIndex < 0 {
		t.Fatal("expected the conformance unit to embed the header's full output verbatim")
	}
	conformanceOnly := conformance[headerIndex+len(header):]
	// Consume cgen.BannedOptimizerAttributes rather than a hard-coded token
	// literal, matching the existing tests' stated reason: widening the
	// banned set must not silently bypass the mutation-kill.
	if len(cgen.BannedOptimizerAttributes) == 0 {
		t.Fatal("cgen.BannedOptimizerAttributes is empty; cannot construct an injection")
	}
	injectionToken := " " + strings.Join(cgen.BannedOptimizerAttributes, " ") + " "
	injectedConformanceOnly := strings.Replace(conformanceOnly, "_Static_assert", "_Static_assert"+injectionToken, 1)
	if injectedConformanceOnly == conformanceOnly {
		t.Fatal("injection site not found in the conformance unit's own added text")
	}
	injectedConformance := conformance[:headerIndex+len(header)] + injectedConformanceOnly
	// The untouched header and untouched compiled-program C still scan
	// clean -- proving the injection did not leak into either.
	if found := cgen.ScanForBannedAttributes(cSource); len(found) != 0 {
		t.Fatalf("untouched compiled-program C unexpectedly flagged: %v", found)
	}
	if found := cgen.ScanForBannedAttributes(header); len(found) != 0 {
		t.Fatalf("untouched header unexpectedly flagged: %v", found)
	}
	// But scanning the reassembled injected conformance unit (as the
	// production lane now does) must catch the injected tokens.
	if found := cgen.ScanForBannedAttributes(injectedConformance); len(found) == 0 {
		t.Fatal("expected the conformance-only injected banned tokens to be detected")
	}
}

// TestAttributeInjectionMakesControlFail demonstrates D-04-13's
// mutation-kill by injection (rather than by isolated assertion): a banned
// token injected into a COPY of the REAL corpus-emitted C (as if the emitter
// had produced it) makes the scan report it, proving the control is not
// vacuously green.
func TestAttributeInjectionMakesControlFail(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	cSource, err := cgen.Emit(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if found := cgen.ScanForBannedAttributes(cSource); len(found) != 0 {
		t.Fatalf("real emitted C already contains a banned token: %v", found)
	}
	injected := strings.Replace(cSource, "extern", "extern __attribute__((malloc)) restrict ", 1)
	if injected == cSource {
		t.Fatal("injection site not found in emitted C")
	}
	found := cgen.ScanForBannedAttributes(injected)
	if len(found) == 0 {
		t.Fatal("expected the injected banned tokens to be detected")
	}
}

// TestNoreturnExemptionIsNamed proves D-04-14's one named exemption: the
// _Noreturn marker is explicitly named as an exemption and is not itself a
// member of the banned-attribute set the zero-attribute scan enforces.
func TestNoreturnExemptionIsNamed(t *testing.T) {
	if cgen.NoreturnExemption != "_Noreturn" {
		t.Fatalf("NoreturnExemption = %q, want _Noreturn", cgen.NoreturnExemption)
	}
	for _, token := range cgen.BannedOptimizerAttributes {
		if token == cgen.NoreturnExemption {
			t.Fatalf("NoreturnExemption must not be a member of BannedOptimizerAttributes: %v", cgen.BannedOptimizerAttributes)
		}
	}
}

// TestVerifyPhase4ForeignLayoutControls proves the Phase 4 gate observes
// both new task-03 required controls with nonzero recomputed work.
func TestVerifyPhase4ForeignLayoutControls(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase4"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("Phase 4 foreign verify failed: status=%s diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	required := []string{"control:foreign.layout_mismatch", "control:foreign.no_unproven_attributes"}
	found := make(map[string]bool)
	for _, lane := range result.Lanes {
		if lane.Status != "pass" || lane.RecomputedWork == 0 {
			t.Fatalf("incomplete lane: %+v", lane)
		}
		for _, control := range lane.Controls {
			found[control] = true
		}
	}
	for _, control := range required {
		if !found[control] {
			t.Fatalf("missing required Phase 4 control %s: lanes=%+v", control, result.Lanes)
		}
	}
}

func TestVerifyPhase4ForeignControls(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase4"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("Phase 4 foreign verify failed: status=%s diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	required := []string{
		"control:foreign.unwind_policy_undeclared",
		"control:foreign.call_target_not_foreign",
	}
	found := make(map[string]bool)
	for _, lane := range result.Lanes {
		if lane.Status != "pass" || lane.RecomputedWork == 0 {
			t.Fatalf("incomplete lane: %+v", lane)
		}
		for _, control := range lane.Controls {
			found[control] = true
		}
	}
	for _, control := range required {
		if !found[control] {
			t.Fatalf("missing required Phase 4 control %s: lanes=%+v", control, result.Lanes)
		}
	}
}

// TestVerifyPhase4UnwindControl proves control:foreign.unwind_forbidden
// (D-04-19, task 04-05-02) is a required control the Phase 4 gate observes
// with nonzero recomputed work, and mutation-kills it: temporarily emptying
// the checked-in allowlist makes every real undefined symbol in the tracer
// binary unlisted, which must turn the gate red.
func TestVerifyPhase4UnwindControl(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase4"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("Phase 4 foreign verify failed: status=%s diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	found := false
	for _, lane := range result.Lanes {
		if lane.Status != "pass" || lane.RecomputedWork == 0 {
			t.Fatalf("incomplete lane: %+v", lane)
		}
		for _, control := range lane.Controls {
			if control == "control:foreign.unwind_forbidden" {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("missing required Phase 4 control control:foreign.unwind_forbidden: lanes=%+v", result.Lanes)
	}

	original := native.AllowedUndefinedSymbols
	native.AllowedUndefinedSymbols = nil
	defer func() { native.AllowedUndefinedSymbols = original }()
	mutated := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase4"), native.DefaultRunner())
	if mutated.Status == protocol.StatusPass {
		t.Fatal("emptying the undefined-symbol allowlist must turn the gate red, but it stayed green")
	}
	if len(mutated.Diagnostics) == 0 || mutated.Diagnostics[0].Code != "verify.control_missing" {
		t.Fatalf("expected verify.control_missing, got %+v", mutated.Diagnostics)
	}
}

// TestVerifyPhase4NonlocalExitControl proves control:foreign.
// nonlocal_exit_undetected (D-04-17/D-04-21, task 04-05-03) is a required
// control the Phase 4 gate observes with nonzero recomputed work, backing
// the two mutation-kill demonstrations already exercised inline inside
// verifyForeignCorpus (pad omission, and ledger-population omission
// surfacing specifically as a leak-count disagreement).
func TestVerifyPhase4NonlocalExitControl(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase4"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("Phase 4 foreign verify failed: status=%s diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	found := false
	for _, lane := range result.Lanes {
		if lane.Status != "pass" || lane.RecomputedWork == 0 {
			t.Fatalf("incomplete lane: %+v", lane)
		}
		for _, control := range lane.Controls {
			if control == "control:foreign.nonlocal_exit_undetected" {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("missing required Phase 4 control control:foreign.nonlocal_exit_undetected: lanes=%+v", result.Lanes)
	}
}

// TestVerifyPhase4DefectControls proves the Phase 4 gate observes both new
// task-04 required controls with nonzero recomputed work.
func TestVerifyPhase4DefectControls(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase4"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("Phase 4 defect verify failed: status=%s diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	required := []string{"control:defect.no_release_on_defect", "control:defect.signal_adjudicated"}
	found := make(map[string]bool)
	for _, lane := range result.Lanes {
		if lane.Status != "pass" || lane.RecomputedWork == 0 {
			t.Fatalf("incomplete lane: %+v", lane)
		}
		for _, control := range lane.Controls {
			found[control] = true
		}
	}
	for _, control := range required {
		if !found[control] {
			t.Fatalf("missing required Phase 4 control %s: lanes=%+v", control, result.Lanes)
		}
	}
}

// TestVerifyPhase4TerminatorControl proves control:terminator.walk_incomplete
// (D-04-29, task 04-06-01) is a required Phase 4 control with nonzero
// recomputed work -- the single highest-risk item in the phase, asserted
// directly rather than inferred from a passing differential.
func TestVerifyPhase4TerminatorControl(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase4"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("Phase 4 verify failed: status=%s diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	found := false
	for _, lane := range result.Lanes {
		if lane.Status != "pass" || lane.RecomputedWork == 0 {
			t.Fatalf("incomplete lane: %+v", lane)
		}
		for _, control := range lane.Controls {
			if control == "control:terminator.walk_incomplete" {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("missing required Phase 4 control control:terminator.walk_incomplete: lanes=%+v", result.Lanes)
	}
}

// TestVerifyPhase4ForeignOriginControl proves control:origin.foreign_origin_omitted
// (D-04-28, task 04-06-02) is a required Phase 4 control with nonzero
// recomputed work.
func TestVerifyPhase4ForeignOriginControl(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase4"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("Phase 4 verify failed: status=%s diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	found := false
	for _, lane := range result.Lanes {
		if lane.Status != "pass" || lane.RecomputedWork == 0 {
			t.Fatalf("incomplete lane: %+v", lane)
		}
		for _, control := range lane.Controls {
			if control == "control:origin.foreign_origin_omitted" {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("missing required Phase 4 control control:origin.foreign_origin_omitted: lanes=%+v", result.Lanes)
	}
}

// TestPublishedOriginValidatedOnCheckAndRun is D-04-27/WR-01's falsifier:
// originvalidate.ValidatePublished now runs on the `check` and `run`
// command-file paths, not only `interface export`. testdata/phase3's
// public_view_omitted.lang (the exclusive_borrow_clean/relay shape D-04-03
// keeps checking clean on purpose) previously PASSED session.CheckCommandFile
// unchanged; it must now be refused with core.origin_omitted there too. The
// new foreign_origin_omitted.lang fixture demonstrates the same wiring
// through session.RunInterpreterCommandFile/RunNativeCommandFile.
func TestPublishedOriginValidatedOnCheckAndRun(t *testing.T) {
	omittedPath := testsupport.ProjectPath("testdata", "phase3", "public_view_omitted.lang")
	checkResult, err := session.CheckCommandFile(omittedPath)
	if err != nil {
		t.Fatal(err)
	}
	if checkResult.Status != protocol.StatusInvalid || len(checkResult.Diagnostics) != 1 || checkResult.Diagnostics[0].Code != "core.origin_omitted" {
		t.Fatalf("expected lang check to refuse public_view_omitted.lang with core.origin_omitted, got %+v", checkResult)
	}

	foreignOmittedPath := testsupport.ProjectPath("testdata", "phase4", "foreign_origin_omitted.lang")
	interpResult, err := session.RunInterpreterCommandFile(foreignOmittedPath)
	if err != nil {
		t.Fatal(err)
	}
	if interpResult.Status != protocol.StatusInvalid || len(interpResult.Diagnostics) != 1 || interpResult.Diagnostics[0].Code != "core.foreign_origin_omitted" {
		t.Fatalf("expected lang run --engine=interpreter to refuse foreign_origin_omitted.lang with core.foreign_origin_omitted, got %+v", interpResult)
	}

	nativeResult, err := session.RunNativeCommandFile(context.Background(), foreignOmittedPath, native.DefaultRunner())
	if err != nil {
		t.Fatal(err)
	}
	if nativeResult.Status != protocol.StatusInvalid || len(nativeResult.Diagnostics) != 1 || nativeResult.Diagnostics[0].Code != "core.foreign_origin_omitted" {
		t.Fatalf("expected lang run --engine=native to refuse foreign_origin_omitted.lang with core.foreign_origin_omitted, got %+v", nativeResult)
	}
}

// TestNoReleaseAfterDefect is control:defect.no_release_on_defect's own
// falsifier: an honest defect execution (no release event anywhere) passes,
// and a hand-constructed one with a release emitted on the defect path --
// the document-level mutation-kill shape, in the same family as
// LayoutMutationRunner attacking an artifact rather than the interpreter's
// own source -- turns the control red.
func TestNoReleaseAfterDefect(t *testing.T) {
	honest := execution.Execution{
		Outcome: execution.Outcome{Kind: "defect"},
		Events:  []execution.Event{{Kind: "function.defected", Output: "halt requested"}},
	}
	if !session.DefectHasNoReleaseAfter(honest) {
		t.Fatal("honest defect execution flagged as violating no-release-on-defect")
	}
	mutated := execution.Execution{
		Outcome: execution.Outcome{Kind: "defect"},
		Events: []execution.Event{
			{Kind: "function.defected", Output: "halt requested"},
			{Kind: "resource.released"},
		},
	}
	if session.DefectHasNoReleaseAfter(mutated) {
		t.Fatal("control did not catch a release emitted on the defect path")
	}
	nonDefect := execution.Execution{
		Outcome: execution.Outcome{Kind: "returned"},
		Events:  []execution.Event{{Kind: "resource.released"}},
	}
	if !session.DefectHasNoReleaseAfter(nonDefect) {
		t.Fatal("a non-defect execution must be vacuously true")
	}
}

// nonlocalProbeChecked checks the nonlocal-exit landing-pad fixture, reused
// by every Task 04-05-01 test below.
func nonlocalProbeChecked(t *testing.T) session.CheckResult {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "nonlocal_exit_probe.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	return checked
}

// TestNonlocalExitEmitsLeakPerLiveAcquisition proves D-04-17 end to end
// through the real generated pipeline: compiled and linked against the
// second frozen foreign translation unit (native/lang_foreign_nonlocal.c),
// the probe's second call performs a genuine longjmp back into the
// process-root landing pad, which reports exactly the one still-live
// acquisition as leaked, then terminates as a defect via a real SIGABRT --
// never a hardcoded exit code (D-04-24, reused here).
func TestNonlocalExitEmitsLeakPerLiveAcquisition(t *testing.T) {
	checked := nonlocalProbeChecked(t)
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	runner := native.DefaultRunner()
	runner.Expect = native.ExpectDefect
	runner.ForeignSources = []string{native.ForeignNonlocalSourcePath()}
	result, err := runner.Run(context.Background(), generated, "-O0", []string{"7"})
	if err != nil {
		t.Fatalf("native run failed: %v", err)
	}
	if len(result.Pairs) != 1 {
		t.Fatalf("expected one execution pair, got %d", len(result.Pairs))
	}
	got := result.Pairs[0].Execution
	if got.Outcome.Kind != "defect" || got.Outcome.Value != "" {
		t.Fatalf("expected a clean defect outcome with no value, got %+v", got.Outcome)
	}
	wantKinds := []string{"foreign.called", "foreign.nonlocal_exit", "resource.leaked", "function.defected"}
	if len(got.Events) != len(wantKinds) {
		t.Fatalf("event count = %d, want %d: %+v", len(got.Events), len(wantKinds), got.Events)
	}
	for i, kind := range wantKinds {
		if got.Events[i].Kind != kind {
			t.Fatalf("event %d kind = %q, want %q: %+v", i, got.Events[i].Kind, kind, got.Events)
		}
	}
	if len(got.LiveResources) != 1 {
		t.Fatalf("expected exactly one leaked resource recorded in live_resources, got %+v", got.LiveResources)
	}
}

// TestPadRunsNoRelease proves D-04-18: the generated pad's own body (the
// span between the setjmp-installation marker and its matching end marker)
// never calls a release -- it emits no resource.released event and never
// clears a ledger slot -- so an honest leak can never be converted into a
// use-after-free by a stray release running from indeterminate state.
func TestPadRunsNoRelease(t *testing.T) {
	checked := nonlocalProbeChecked(t)
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(generated, "/* lang:nonlocal-pad-site */")
	end := strings.Index(generated, "/* lang:nonlocal-pad-end */")
	if start < 0 || end < 0 || end < start {
		t.Fatalf("expected a well-formed pad span in generated C:\n%s", generated)
	}
	padBody := generated[start:end]
	if strings.Contains(padBody, "resource.released") {
		t.Fatalf("pad body must never emit resource.released:\n%s", padBody)
	}
	if strings.Contains(padBody, "= 0;") {
		t.Fatalf("pad body must never clear a ledger slot (a release-shaped mutation):\n%s", padBody)
	}
}

// TestNonlocalExitProbeInterpreterNative proves SC4 on exactly the path SC3
// is about: the interpreter (which cannot actually call C, so it models the
// same shared nonlocal-exit-on-second-call convention documented in both
// cgen.go and interp.go) and the real compiled-and-linked native binary
// agree, byte-for-byte, on the ordered event sequence and the defect
// terminal record for this probe.
func TestNonlocalExitProbeInterpreterNative(t *testing.T) {
	checked := nonlocalProbeChecked(t)
	interpreted, err := interp.Run(checked.Program, "main", "7")
	if err != nil {
		t.Fatal(err)
	}
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	runner := native.DefaultRunner()
	runner.Expect = native.ExpectDefect
	runner.ForeignSources = []string{native.ForeignNonlocalSourcePath()}
	result, err := runner.Run(context.Background(), generated, "-O0", []string{"7"})
	if err != nil {
		t.Fatalf("native run failed: %v", err)
	}
	if len(result.Pairs) != 1 {
		t.Fatalf("expected one execution pair, got %d", len(result.Pairs))
	}
	if !execution.Equal(interpreted, result.Pairs[0].Execution) {
		interpretedBytes, _ := execution.CanonicalBytes(interpreted)
		nativeBytes, _ := execution.CanonicalBytes(result.Pairs[0].Execution)
		t.Fatalf("interpreter and native disagree on the nonlocal-exit probe:\ninterpreter: %s\nnative:      %s", interpretedBytes, nativeBytes)
	}
}

// TestNonlocalExitDetectionIsMutationKilled is control:foreign.
// nonlocal_exit_undetected's FIRST mutation-kill demonstration (D-04-21/
// D-09/D-10), standing alone from the session-level lane: removing the
// process-root pad's entire generated span (setjmp installation through its
// matching end marker) from the probe's own generated C must make the
// probe's own defect terminal record and its foreign.nonlocal_exit/
// resource.leaked events disappear -- proving detection by mutation rather
// than by assertion alone.
func TestNonlocalExitDetectionIsMutationKilled(t *testing.T) {
	checked := nonlocalProbeChecked(t)
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	runner := native.DefaultRunner()
	runner.Expect = native.ExpectDefect
	runner.ForeignSources = []string{native.ForeignNonlocalSourcePath()}
	mutationRunner := session.NewNonlocalPadOmissionMutationRunner(runner)
	result, runErr := mutationRunner.Run(context.Background(), generated, "-O0", []string{"7"})
	if runErr == nil && len(result.Pairs) == 1 && result.Pairs[0].Execution.Outcome.Kind == "defect" {
		t.Fatalf("removing the pad installation must make detection disappear, but the mutated run still produced a clean defect outcome: %+v", result.Pairs[0].Execution)
	}
	if len(mutationRunner.Optimizations()) == 0 {
		t.Fatal("expected the pad-omission mutation runner to record at least one optimization pass")
	}
}

// TestLeakCountMatchesLiveAcquisitions proves the golden probe's leak count
// equals its true live-acquisition count (the positive case), then
// mutation-kills control:foreign.nonlocal_exit_undetected a SECOND, DIFFERENT
// way (D-09/D-10/D-21): dropping one ledger-population site must make the
// mutated run's own leak count disagree with the golden run's -- checked
// SPECIFICALLY as a leak-count disagreement, not merely "some difference."
func TestLeakCountMatchesLiveAcquisitions(t *testing.T) {
	checked := nonlocalProbeChecked(t)
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	runner := native.DefaultRunner()
	runner.Expect = native.ExpectDefect
	runner.ForeignSources = []string{native.ForeignNonlocalSourcePath()}
	goldenResult, err := runner.Run(context.Background(), generated, "-O0", []string{"7"})
	if err != nil || len(goldenResult.Pairs) != 1 {
		t.Fatalf("golden run failed: err=%v result=%+v", err, goldenResult)
	}
	golden := goldenResult.Pairs[0].Execution
	goldenLeaks := 0
	for _, event := range golden.Events {
		if event.Kind == "resource.leaked" {
			goldenLeaks++
		}
	}
	if goldenLeaks == 0 || goldenLeaks != len(golden.LiveResources) {
		t.Fatalf("golden run's leak count (%d) must equal its live-acquisition count (%+v)", goldenLeaks, golden.LiveResources)
	}

	mutationRunner := session.NewNonlocalLedgerOmissionMutationRunner(runner)
	mutatedResult, mutatedErr := mutationRunner.Run(context.Background(), generated, "-O0", []string{"7"})
	mutatedLeaks := -1
	if mutatedErr == nil && len(mutatedResult.Pairs) == 1 {
		mutatedLeaks = 0
		for _, event := range mutatedResult.Pairs[0].Execution.Events {
			if event.Kind == "resource.leaked" {
				mutatedLeaks++
			}
		}
	}
	if mutatedLeaks == goldenLeaks {
		t.Fatalf("dropping a ledger-population site must make the leak count disagree with the golden run's (%d), but the mutated run reported the SAME count", goldenLeaks)
	}
}

// TestVerifyPhase4ControlsAndWork is task 04-07-01's own control-and-work
// pin, matching TestVerifyPhase2ControlsAndWork/TestVerifyPhase3ControlsAndWork's
// established shape: every one of session.Phase4RequiredControls()'s twelve
// identifiers must be observed by the in-process VerifyCorpus path with
// nonzero recomputed work on every lane, and the phase's three new expected
// escapes must be declared without ever appearing as a detected control.
func TestVerifyPhase4ControlsAndWork(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase4"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("Phase 4 verify failed: status=%s diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	found := make(map[string]bool)
	for _, lane := range result.Lanes {
		if lane.Status != "pass" || lane.RecomputedWork == 0 {
			t.Fatalf("incomplete lane: %+v", lane)
		}
		for _, control := range lane.Controls {
			found[control] = true
		}
	}
	for _, control := range session.Phase4RequiredControls() {
		if !found[control] {
			t.Fatalf("missing required Phase 4 control %s: lanes=%+v", control, result.Lanes)
		}
	}
	expectedEscapes := map[string]bool{}
	for _, escape := range result.ExpectedEscapes {
		expectedEscapes[escape] = true
	}
	for _, escape := range []string{session.EscapeCoordinatedForeignBoundaryLie, session.EscapeNonlocalExitBelowThePad, session.EscapeForeignProcessExit} {
		if !expectedEscapes[escape] {
			t.Fatalf("Phase 4 verify omitted an expected escape: %s (got %+v)", escape, result.ExpectedEscapes)
		}
		if found[escape] {
			t.Fatalf("expected escape %s must never appear as a detected control: lanes=%+v", escape, result.Lanes)
		}
	}
}

// TestAttributeScanLaneCoversEveryInspectableLayer is 04-09's pin on
// lane:foreign-no-unproven-attributes's own scanned-artifact count. The
// number 8 is two fixture programs (tracer, release) times four artifacts
// each (compiled program, sidecar manifest, generated header, generated
// conformance unit). This assertion is what makes dropping an argument from
// session.go's ScanForBannedAttributes call a red test rather than a silent
// coverage regression -- the exact state 04-VERIFICATION.md gap 2 recorded:
// nothing asserted how many artifacts the production lane scanned, so
// deleting an argument turned no test red even though the lane still
// reported "pass". The lane must also be present at all: an absent lane is
// the strongest form of this regression, not a case to pass over.
func TestAttributeScanLaneCoversEveryInspectableLayer(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase4"), native.DefaultRunner())
	var lane *protocol.Lane
	for i := range result.Lanes {
		if result.Lanes[i].ID == "lane:foreign-no-unproven-attributes" {
			lane = &result.Lanes[i]
			break
		}
	}
	if lane == nil {
		t.Fatalf("lane:foreign-no-unproven-attributes not found in result.Lanes: %+v", result.Lanes)
	}
	if lane.Status != "pass" {
		t.Fatalf("lane:foreign-no-unproven-attributes status = %q, want \"pass\": %+v", lane.Status, lane)
	}
	if lane.RecomputedWork != 8 {
		t.Fatalf("lane:foreign-no-unproven-attributes work = %d, want 8 (two fixture programs x four artifacts each); a value below 8 means a scanned artifact was dropped from session.go's ScanForBannedAttributes argument list: %+v", lane.RecomputedWork, lane)
	}
	if lane.RecomputedWork <= 0 {
		t.Fatalf("lane:foreign-no-unproven-attributes RecomputedWork must be strictly greater than zero, got %d", lane.RecomputedWork)
	}
}

// TestVerifyPhase4CLI proves TestVerifyPhase4ControlsAndWork's claim holds
// through the shipped binary, not only the in-process session layer
// (D-04-21's "the gate only ever sees what ships with it"): `lang --json
// verify testdata/phase4` must report every required Phase 4 control and
// every expected escape.
func TestVerifyPhase4CLI(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	corpus := testsupport.ProjectPath("testdata", "phase4")
	run := testsupport.RunCLI(t, binary, nil, "--json", "verify", corpus)
	if run.Exit != 0 || len(run.Stderr) != 0 {
		t.Fatalf("verify testdata/phase4: %+v", run)
	}
	var decoded protocol.Result
	if err := json.Unmarshal(run.Stdout, &decoded); err != nil {
		t.Fatalf("decode verify output: %v (stdout=%s)", err, run.Stdout)
	}
	if decoded.Status != protocol.StatusPass {
		t.Fatalf("verify status=%s stdout=%s", decoded.Status, run.Stdout)
	}
	found := make(map[string]bool)
	for _, lane := range decoded.Lanes {
		for _, control := range lane.Controls {
			found[control] = true
		}
	}
	for _, control := range session.Phase4RequiredControls() {
		if !found[control] {
			t.Fatalf("shipped binary omitted required Phase 4 control %s: stdout=%s", control, run.Stdout)
		}
	}
	escapes := make(map[string]bool)
	for _, escape := range decoded.ExpectedEscapes {
		escapes[escape] = true
	}
	for _, escape := range []string{session.EscapeCoordinatedForeignBoundaryLie, session.EscapeNonlocalExitBelowThePad, session.EscapeForeignProcessExit} {
		if !escapes[escape] {
			t.Fatalf("shipped binary omitted expected escape %s: stdout=%s", escape, run.Stdout)
		}
	}
}

// phase4VerifierScriptText and phase4OwnControlBlock are shared helpers for
// TestPhase4VerifierScriptContract and TestPhase4RequiredControlsMatchScript
// below: both tests read scripts/verify-phase4.sh's own text, so the second
// pulls its extraction into a helper rather than re-reading the file.
func phase4VerifierScriptText(t testing.TB) string {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("scripts", "verify-phase4.sh"))
	if err != nil {
		t.Fatalf("read scripts/verify-phase4.sh: %v", err)
	}
	return string(data)
}

// phase4OwnControlIdentifiers extracts the exact set of `control:` tokens
// from the script's OWN required-control block (the `for control in ...`
// loop following the "Phase 4's own required-control set" comment), as
// distinct from the earlier block re-asserting Phase 3's non-regression
// controls against phase3.json -- conflating the two blocks would let a
// Phase 3 control identifier masquerade as a Phase 4 one.
func phase4OwnControlIdentifiers(t testing.TB, text string) []string {
	t.Helper()
	anchor := "Phase 4's own required-control set"
	anchorIndex := strings.Index(text, anchor)
	if anchorIndex == -1 {
		t.Fatalf("scripts/verify-phase4.sh is missing its own required-control-set comment anchor")
	}
	rest := text[anchorIndex:]
	doneIndex := strings.Index(rest, "\ndone")
	if doneIndex == -1 {
		t.Fatalf("scripts/verify-phase4.sh's own required-control block has no closing done")
	}
	block := rest[:doneIndex]
	var identifiers []string
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), "\\"))
		if strings.HasPrefix(trimmed, "control:") {
			identifiers = append(identifiers, trimmed)
		}
	}
	return identifiers
}

// TestPhase4VerifierScriptContract is task 04-07-01's contract test over the
// gate script's own text (T-04-41): it asserts the script never invokes a
// previous-phase gate script (T-04-42), never touches scripts/verify-phase3.sh,
// and names every one of the twelve required Phase 4 control identifiers.
func TestPhase4VerifierScriptContract(t *testing.T) {
	text := phase4VerifierScriptText(t)
	for _, forbidden := range []string{"verify-phase1.sh", "verify-phase2.sh", "verify-phase3.sh"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("scripts/verify-phase4.sh must never invoke a previous-phase gate script, but its text contains %q", forbidden)
		}
	}
	for _, control := range session.Phase4RequiredControls() {
		if !strings.Contains(text, control) {
			t.Fatalf("scripts/verify-phase4.sh's text is missing required control %s", control)
		}
	}
	phase3Path := testsupport.ProjectPath("scripts", "verify-phase3.sh")
	if _, err := os.Stat(phase3Path); err != nil {
		t.Fatalf("scripts/verify-phase3.sh must still exist unchanged: %v", err)
	}
}

// TestPhase4RequiredControlsMatchScript asserts set equality between
// session.Phase4RequiredControls() and the identifiers named in the
// script's own required-control block, so a control added to one and
// forgotten in the other fails here rather than silently drifting apart.
func TestPhase4RequiredControlsMatchScript(t *testing.T) {
	text := phase4VerifierScriptText(t)
	fromScript := phase4OwnControlIdentifiers(t, text)
	scriptSet := make(map[string]bool, len(fromScript))
	for _, control := range fromScript {
		scriptSet[control] = true
	}
	sessionSet := make(map[string]bool, len(session.Phase4RequiredControls()))
	for _, control := range session.Phase4RequiredControls() {
		sessionSet[control] = true
	}
	for control := range sessionSet {
		if !scriptSet[control] {
			t.Fatalf("session.Phase4RequiredControls() names %s, which the script's own required-control block does not", control)
		}
	}
	for control := range scriptSet {
		if !sessionSet[control] {
			t.Fatalf("the script's required-control block names %s, which session.Phase4RequiredControls() does not", control)
		}
	}
	if len(scriptSet) != len(sessionSet) {
		t.Fatalf("script control set (%d) and session control set (%d) differ in size: script=%v session=%v", len(scriptSet), len(sessionSet), fromScript, session.Phase4RequiredControls())
	}
}

// TestExpectedEscapesAreVisibleNotSolved is task 04-07-01's expected-escape
// visibility control: the three Phase 4 expected escapes must be declared
// under the verify result's expected escapes, and none of the three may
// ever appear as a detected lane control -- an accepted residual claimed
// solved would be a false positive of exactly the kind this project has
// never tolerated (D-04-31).
func TestExpectedEscapesAreVisibleNotSolved(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase4"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("Phase 4 verify failed: status=%s diagnostics=%+v", result.Status, result.Diagnostics)
	}
	declared := make(map[string]bool, len(result.ExpectedEscapes))
	for _, escape := range result.ExpectedEscapes {
		declared[escape] = true
	}
	detected := make(map[string]bool)
	for _, lane := range result.Lanes {
		for _, control := range lane.Controls {
			detected[control] = true
		}
	}
	for _, escape := range []string{session.EscapeCoordinatedForeignBoundaryLie, session.EscapeNonlocalExitBelowThePad, session.EscapeForeignProcessExit} {
		if !declared[escape] {
			t.Fatalf("expected escape %s is not declared: %+v", escape, result.ExpectedEscapes)
		}
		if detected[escape] {
			t.Fatalf("expected escape %s must never be presented as a detected control: lanes=%+v", escape, result.Lanes)
		}
	}
}

// writeFailingForeignDouble writes a TEST-ONLY, throwaway foreign
// translation unit implementing the SAME symbol and ABI shape as the
// frozen native/lang_foreign_resource.c (_LANG_lang_res_open_result
// _LANG_lang_res_open(unsigned char)) but that genuinely returns ok=0 on
// its failOnCall'th invocation within one process, using a static call
// counter -- the exact "Nth call" convention native/lang_foreign_nonlocal.c
// already establishes for the nonlocal-exit probe (D-04-17). This is NOT a
// change to any frozen, byte-committed file: it is written fresh to
// t.TempDir() for this test alone, never touching
// native/lang_foreign_resource.c, so D-04-10's freeze is untouched. It
// exists because the shipped foreign TU always succeeds at real runtime (a
// genuine allocation failure is unreachable in practice), so proving the
// interpreter and BOTH native optimization levels genuinely agree on a
// typed-failure path's terminal outcome, events, and live resources
// requires a real ok=0 SOMEWHERE in the toolchain -- this double supplies
// it without touching production code cgen.go emits or the frozen TU it
// links against for every other Phase 4 fixture.
func writeFailingForeignDouble(t *testing.T, failOnCall int) string {
	t.Helper()
	source := fmt.Sprintf(`#include <stdint.h>

typedef struct _LANG_lang_res_open_result {
  unsigned char ok;
  unsigned char value;
} _LANG_lang_res_open_result;

static int lang_test_double_call_count = 0;

_LANG_lang_res_open_result _LANG_lang_res_open(unsigned char argument) {
  _LANG_lang_res_open_result result;
  lang_test_double_call_count++;
  if (lang_test_double_call_count == %d) {
    result.ok = 0;
    result.value = 0;
    return result;
  }
  result.ok = 1;
  result.value = argument;
  return result;
}
`, failOnCall)
	path := filepath.Join(t.TempDir(), "lang_test_double_resource.c")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("write test-double foreign TU: %v", err)
	}
	return path
}

// assertTypedFailurePathAgrees is task 04-07-02's shared engine-agreement
// assertion for the second-stage and third-stage typed-failure path
// shapes. Both are genuinely EXECUTED, not hand-constructed: the
// interpreter runs the exact err block directly via
// interptestdirect.RunLinearBlockDirect (Run's own public entry point cannot reach
// it, since the interpreter's documented discretionary stub always
// simulates success for every OpForeignCall, D-04-04/04-PATTERNS Pattern
// 3), and both native optimization levels compile and run the REAL,
// UNMUTATED program's own generated C (already containing a compilable,
// merely dead, "if (!result.ok) { <release*, fail> }" branch for every
// call site) linked against writeFailingForeignDouble's test-only object
// instead of the frozen production TU -- genuinely returning ok=0 on the
// failing call, making that dead branch live. No engine is stubbed or fed
// a constructed execution document; all three consume the SAME real
// program through their ordinary machinery.
func assertTypedFailurePathAgrees(t *testing.T, fixture string, failOnCall int, liveOpIndexes []int, errBlockSuffix string) {
	t.Helper()
	corpus := testsupport.ProjectPath("testdata", "phase4")
	program, functionName, err := session.Phase4CheckedProgram(corpus, fixture)
	if err != nil {
		t.Fatal(err)
	}
	functionID := program.Functions[0].ID
	precedingCallIDs := make([]string, 0, len(liveOpIndexes))
	for _, index := range liveOpIndexes {
		precedingCallIDs = append(precedingCallIDs, fmt.Sprintf("%s:op:%d", functionID, index))
	}
	// The failing call is always the next OpForeignCall in declaration
	// order after the preceding successful ones (0-indexed: A=0, B=1, C=2).
	failingCallID := fmt.Sprintf("%s:op:%d", functionID, len(liveOpIndexes))
	blockID := functionID + errBlockSuffix

	interpreted, err := interptestdirect.RunLinearBlockDirect(program, functionName, blockID, precedingCallIDs, failingCallID)
	if err != nil {
		t.Fatalf("interptestdirect.RunLinearBlockDirect: %v", err)
	}

	cSource, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("cgen.EmitNative: %v", err)
	}
	doublePath := writeFailingForeignDouble(t, failOnCall)
	runner := native.DefaultRunner()
	runner.Expect = native.ExpectTypedFailure
	runner.ForeignSources = []string{doublePath}

	o0, err := runner.Run(context.Background(), cSource, "-O0", []string{"7"})
	if err != nil || len(o0.Pairs) != 1 {
		t.Fatalf("-O0 run failed: err=%v result=%+v", err, o0)
	}
	o3, err := runner.Run(context.Background(), cSource, "-O3", []string{"7"})
	if err != nil || len(o3.Pairs) != 1 {
		t.Fatalf("-O3 run failed: err=%v result=%+v", err, o3)
	}

	if compareErr := session.Phase4CompareThreeEngines(fixture, interpreted, o0.Pairs[0].Execution, o3.Pairs[0].Execution); compareErr != nil {
		t.Fatalf("typed-failure path disagreement: %v", compareErr)
	}
	if interpreted.Outcome.Kind != "typed_failure" {
		t.Fatalf("expected typed_failure outcome, got %+v", interpreted.Outcome)
	}
}

// TestPhase4CorpusThreeEngineAgreement is task 04-07-02's own differential
// (ROADMAP SC4): the interpreter, the unoptimized native build, and the
// optimized native build must agree on terminal outcome, ordered event
// sequence, and live-resource state across all five Phase 4 path shapes --
// success, second-stage typed failure, third-stage typed failure, defect,
// and nonlocal exit -- named here explicitly so the coverage cannot
// silently shrink to return-only paths.
func TestPhase4CorpusThreeEngineAgreement(t *testing.T) {
	corpus := testsupport.ProjectPath("testdata", "phase4")
	runner := native.DefaultRunner()

	t.Run("success", func(t *testing.T) {
		program, functionName, err := session.Phase4CheckedProgram(corpus, "acquire_three_success.lang")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := session.Phase4ThreeEngineDifferential(context.Background(), "acquire_three_success.lang", program, functionName, "7", runner, native.ExpectValue); err != nil {
			t.Fatalf("success path disagreement: %v", err)
		}
	})

	t.Run("second-stage-typed-failure", func(t *testing.T) {
		// Second call (index 1, B) fails; only A (op:0) is live on entry to
		// block:err:1, which releases A and never B (B never completed).
		assertTypedFailurePathAgrees(t, "acquire_three_fail_second.lang", 2, []int{0}, ":block:err:1")
	})

	t.Run("third-stage-typed-failure", func(t *testing.T) {
		// Third call (index 2, C) fails; A and B (op:0, op:1) are live on
		// entry to block:err:2, which releases B then A in reverse order.
		assertTypedFailurePathAgrees(t, "acquire_three_fail_third.lang", 3, []int{0, 1}, ":block:err:2")
	})

	t.Run("defect", func(t *testing.T) {
		program, functionName, err := session.Phase4CheckedProgram(corpus, "defect_terminal.lang")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := session.Phase4ThreeEngineDifferential(context.Background(), "defect_terminal.lang", program, functionName, "Halt", runner, native.ExpectDefect); err != nil {
			t.Fatalf("defect path disagreement: %v", err)
		}
	})

	t.Run("nonlocal-exit", func(t *testing.T) {
		program, functionName, err := session.Phase4CheckedProgram(corpus, "nonlocal_exit_probe.lang")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := session.Phase4ThreeEngineDifferential(context.Background(), "nonlocal_exit_probe.lang", program, functionName, "7", runner, native.ExpectDefect); err != nil {
			t.Fatalf("nonlocal-exit path disagreement: %v", err)
		}
	})
}

// TestPhase4DifferentialNamesFirstDisagreement proves
// Phase4CompareThreeEngines's own required-behavior (T-04-43): a
// disagreement is reported with the fixture identifier, the specific
// engine pair, and the first differing field, never merely "a mismatch
// occurred." It hand-constructs three documents where the interpreter and
// -O3 agree on "returned" but -O0 disagrees on outcome.kind alone, proving
// the report names -O0 specifically (interpreter-vs-O0) rather than every
// pair, and names outcome.kind as the first field the comparison checks --
// matching the established hand-constructed-document technique this file
// already uses for TestOwnedExecutionFieldMutationMatrix, so this
// comparison is directly testable without a native toolchain invocation.
func TestPhase4DifferentialNamesFirstDisagreement(t *testing.T) {
	honest := execution.Execution{
		Schema:  execution.Schema1,
		Outcome: execution.Outcome{Kind: "returned", Value: "7"},
		Events: []execution.Event{
			{Schema: execution.Schema1, ID: "event:0", Kind: "function.returned", FunctionID: "fn:main"},
		},
		LiveResources: []string{},
	}
	disagreeing := honest
	disagreeing.Outcome = execution.Outcome{Kind: "typed_failure", Value: "OpenFailed"}

	diffErr := session.Phase4CompareThreeEngines("acquire_three_success.lang", honest, disagreeing, honest)
	if diffErr == nil {
		t.Fatal("expected a disagreement between the interpreter and -O0 documents, got none")
	}
	var disagreement *session.Phase4EngineDisagreement
	if !errors.As(diffErr, &disagreement) {
		t.Fatalf("expected a *session.Phase4EngineDisagreement, got %v (%T)", diffErr, diffErr)
	}
	if disagreement.Fixture != "acquire_three_success.lang" {
		t.Fatalf("disagreement did not name the fixture: %+v", disagreement)
	}
	if disagreement.EnginePair != "interpreter-vs-O0" {
		t.Fatalf("disagreement named the wrong engine pair, want interpreter-vs-O0: %+v", disagreement)
	}
	if !strings.Contains(disagreement.Detail, "outcome.kind") {
		t.Fatalf("disagreement did not name outcome.kind as the first differing field: %+v", disagreement)
	}

	// The honest triple (no mutation) must be reported as agreement.
	if err := session.Phase4CompareThreeEngines("acquire_three_success.lang", honest, honest, honest); err != nil {
		t.Fatalf("an honest, identical triple must not be reported as a disagreement: %v", err)
	}
}

// TestNoCoverageClaimedForNamedResiduals is task 04-07-03's own scan
// (D-04-31): none of Phase 4's declared expected escapes may be presented,
// anywhere in a Go comment or a fixture's own comment, as covered, closed,
// resolved, or otherwise no longer a limitation. It scans every .go and
// .lang file under internal/compiler and testdata/phase4 for a line naming
// one of the three declared Phase 4 escapes, and fails if that same line
// also contains a claim-of-coverage phrase.
func TestNoCoverageClaimedForNamedResiduals(t *testing.T) {
	namedResiduals := []string{
		session.EscapeCoordinatedForeignBoundaryLie,
		session.EscapeNonlocalExitBelowThePad,
		session.EscapeForeignProcessExit,
	}
	forbiddenPhrases := []string{
		"is covered", "fully covered", "now covered", "closes this residual",
		"resolves this residual", "no longer a limitation", "proven to close",
		"is solved", "is resolved", "is closed",
	}
	roots := []string{
		testsupport.ProjectPath("internal", "compiler"),
		testsupport.ProjectPath("testdata", "phase4"),
	}
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, ".lang") {
				return nil
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			for lineNumber, line := range strings.Split(string(data), "\n") {
				lower := strings.ToLower(line)
				for _, residual := range namedResiduals {
					if !strings.Contains(line, residual) {
						continue
					}
					for _, phrase := range forbiddenPhrases {
						if strings.Contains(lower, phrase) {
							t.Fatalf("%s:%d claims coverage of named residual %s: %q", path, lineNumber+1, residual, line)
						}
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// phaseArtifactGlob resolves a phase-artifact pattern against both the live
// phase tree (.planning/phases/) and the milestone archives
// (.planning/milestones/<milestone>-phases/), which is where a completed
// milestone's phase directories move. A test that only knew the live tree
// would silently start failing -- or, worse for a completeness check,
// silently start passing on an empty set -- the moment a milestone closed.
func phaseArtifactGlob(parts ...string) ([]string, error) {
	live, err := filepath.Glob(testsupport.ProjectPath(append([]string{".planning", "phases"}, parts...)...))
	if err != nil {
		return nil, err
	}
	archived, err := filepath.Glob(testsupport.ProjectPath(append([]string{".planning", "milestones", "*-phases"}, parts...)...))
	if err != nil {
		return nil, err
	}
	return append(live, archived...), nil
}

// TestPhase4ReachabilityRecordIsComplete is task 04-07-03's own closure of
// the Generator and Probe Reachability Register (D-04-21): every row of
// that register in 04-VALIDATION.md must be completed from an in-code
// comment, each naming at least one shape the generator or probe does not
// reach -- coverage is not the question, per D-04-21; reachability is.
func TestPhase4ReachabilityRecordIsComplete(t *testing.T) {
	matches, err := phaseArtifactGlob("04-fallible-resources-and-c-boundary", "04-VALIDATION.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly one 04-VALIDATION.md, found %d: %v", len(matches), matches)
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	anchor := "## Generator and Probe Reachability Register"
	anchorIndex := strings.Index(text, anchor)
	if anchorIndex == -1 {
		t.Fatalf("04-VALIDATION.md is missing its Generator and Probe Reachability Register")
	}
	rest := text[anchorIndex:]
	nextSectionIndex := strings.Index(rest[len(anchor):], "\n## ")
	var section string
	if nextSectionIndex == -1 {
		section = rest
	} else {
		section = rest[:len(anchor)+nextSectionIndex]
	}
	rows := 0
	for _, line := range strings.Split(section, "\n") {
		if !strings.HasPrefix(line, "|") {
			continue
		}
		columns := strings.Split(strings.Trim(line, "|"), "|")
		if len(columns) < 4 {
			continue
		}
		header := strings.TrimSpace(columns[0])
		if header == "Generator / probe" || strings.HasPrefix(header, "---") {
			continue
		}
		rows++
		mustReach := strings.TrimSpace(columns[1])
		knownNotToReach := strings.TrimSpace(columns[2])
		if mustReach == "" {
			t.Fatalf("register row %q has no 'Must reach' entry", header)
		}
		if knownNotToReach == "" || knownNotToReach == "-" {
			t.Fatalf("register row %q names no unreached shape in 'Known not to reach'", header)
		}
	}
	if rows == 0 {
		t.Fatal("Generator and Probe Reachability Register has no data rows")
	}
}

// debtRegisterLandingPhaseExemptions names the registers written before the
// "landing phase" column became part of the standing debt-register shape.
// Each entry is an explicit, dated waiver for a frozen historical document,
// never a licence for a new register to omit the column: any register not
// listed here must carry it. Adding an entry is a review failure unless the
// register is genuinely frozen prior art.
var debtRegisterLandingPhaseExemptions = map[string]string{
	"02-DEBT.md": "written 2026-09-04, before the landing-phase column existed; frozen prior art",
	"03-DEBT.md": "written 2026-09-04, before the landing-phase column existed; frozen prior art",
}

// debtRegisterSeverities is the closed severity vocabulary. A register that
// invents a severity outside this set is drifting rather than recording.
var debtRegisterSeverities = map[string]bool{"blocker": true, "warning": true, "info": true}

// TestDebtRegistersAreWellFormed is the mechanical half of the debt-register
// checkpoint that phase 04-06 recorded as human judgment ("each dated with
// identifier/severity/source/landing phase"). Register *honesty* -- whether
// a deferral is truthfully described -- stays a human reading and is
// deliberately not claimed here. Register *completeness and shape* is
// checkable, recurs in every phase, and is what silently rots, so it is
// asserted on every commit instead:
//
//   - the frontmatter's declared `items:` count matches the Items table,
//   - every row names an identifier, a source, a threat/requirement, a
//     severity from the closed vocabulary, and the item itself,
//   - post-legacy registers additionally name a landing phase, and
//   - every identifier in the table has a matching `### <ID>` detail
//     section, and every detail section has a matching table row.
//
// The register is scanned for every phase, not only Phase 4: a register that
// stops being maintained is exactly the failure this catches.
func TestDebtRegistersAreWellFormed(t *testing.T) {
	registers, err := phaseArtifactGlob("*", "*-DEBT.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(registers) == 0 {
		t.Fatal("no *-DEBT.md register found under .planning/phases or .planning/milestones/*-phases")
	}
	for _, path := range registers {
		t.Run(filepath.Base(path), func(t *testing.T) {
			checkDebtRegister(t, path)
		})
	}
}

func checkDebtRegister(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Base(path)
	text := string(data)

	if !strings.HasPrefix(text, "---\n") {
		t.Fatalf("%s: has no frontmatter block", name)
	}
	frontmatterEnd := strings.Index(text[4:], "\n---")
	if frontmatterEnd == -1 {
		t.Fatalf("%s: frontmatter block is unterminated", name)
	}
	declared := -1
	for _, line := range strings.Split(text[4:4+frontmatterEnd], "\n") {
		if !strings.HasPrefix(line, "items:") {
			continue
		}
		if _, scanErr := fmt.Sscanf(strings.TrimSpace(line), "items: %d", &declared); scanErr != nil {
			t.Fatalf("%s: unreadable `items:` frontmatter line %q", name, line)
		}
		break
	}
	if declared < 0 {
		t.Fatalf("%s: frontmatter declares no `items:` count", name)
	}

	columns, rows := debtRegisterTable(t, name, text)
	if len(rows) != declared {
		t.Fatalf("%s: frontmatter declares items: %d but the Items table holds %d rows", name, declared, len(rows))
	}

	required := []string{"ID", "Source", "Threat/Req", "Severity", "Item"}
	if _, exempt := debtRegisterLandingPhaseExemptions[name]; !exempt {
		required = append(required, "Landing phase")
	}
	for _, column := range required {
		if _, present := columns[column]; !present {
			t.Fatalf("%s: Items table has no %q column (columns: %v)", name, column, columns)
		}
	}

	identifiers := make(map[string]bool, len(rows))
	for _, row := range rows {
		identifier := row[columns["ID"]]
		if !strings.HasPrefix(identifier, "D-") || len(identifier) < len("D-00-0") {
			t.Fatalf("%s: row %q does not name a D-XX-NN identifier", name, identifier)
		}
		if identifiers[identifier] {
			t.Fatalf("%s: identifier %s appears in two rows", name, identifier)
		}
		identifiers[identifier] = true
		for _, column := range required {
			value := row[columns[column]]
			if value == "" || value == "-" {
				t.Fatalf("%s: row %s has an empty %q cell", name, identifier, column)
			}
		}
		if severity := row[columns["Severity"]]; !debtRegisterSeverities[severity] {
			t.Fatalf("%s: row %s has severity %q outside the closed vocabulary (blocker, warning, info)", name, identifier, severity)
		}
	}

	details := make(map[string]bool, len(rows))
	for _, line := range strings.Split(text, "\n") {
		if !strings.HasPrefix(line, "### D-") {
			continue
		}
		heading := strings.TrimSpace(strings.TrimPrefix(line, "###"))
		identifier := heading
		if cut := strings.IndexAny(heading, " \t"); cut != -1 {
			identifier = heading[:cut]
		}
		identifier = strings.Trim(identifier, "`")
		if !identifiers[identifier] {
			t.Fatalf("%s: detail section %q has no row in the Items table", name, identifier)
		}
		details[identifier] = true
	}
	for identifier := range identifiers {
		if !details[identifier] {
			t.Fatalf("%s: row %s has no `### %s` detail section", name, identifier, identifier)
		}
	}
}

// debtRegisterTable returns the Items table's column index by header name and
// its data rows, each row indexed the same way. Only the `## Items` section
// is read: later sections (closures, process debt) hold their own tables and
// are not the register.
func debtRegisterTable(t *testing.T, name, text string) (map[string]int, [][]string) {
	t.Helper()
	anchor := "\n## Items\n"
	start := strings.Index(text, anchor)
	if start == -1 {
		t.Fatalf("%s: has no `## Items` section", name)
	}
	section := text[start+len(anchor):]
	if end := strings.Index(section, "\n## "); end != -1 {
		section = section[:end]
	}
	var columns map[string]int
	var rows [][]string
	for _, line := range strings.Split(section, "\n") {
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cells := strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|")
		for index, cell := range cells {
			cells[index] = strings.TrimSpace(cell)
		}
		if columns == nil {
			columns = make(map[string]int, len(cells))
			for index, cell := range cells {
				columns[cell] = index
			}
			continue
		}
		if strings.HasPrefix(cells[0], "---") {
			continue
		}
		if len(cells) != len(columns) {
			t.Fatalf("%s: row %q has %d cells, want %d", name, cells[0], len(cells), len(columns))
		}
		rows = append(rows, cells)
	}
	if columns == nil {
		t.Fatalf("%s: `## Items` section holds no table", name)
	}
	if len(rows) == 0 {
		t.Fatalf("%s: `## Items` table holds no data row", name)
	}
	return columns, rows
}
