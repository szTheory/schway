package session_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

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
	if !strings.Contains(result.CSource, "typedef enum LANG_SWITCH") || !strings.Contains(result.CSource, "switch (value)") {
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
	result, _, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	if err != nil {
		requirePhase16M004Refusal(t, err, "foreign")
		return
	}
	t.Fatal("expected terminal Phase 16 M004 foreign refusal")
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
	result, _, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	if err != nil {
		requirePhase16M004Refusal(t, err, "foreign")
		return
	}
	t.Fatal("expected terminal Phase 16 M004 foreign refusal")
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
	if err != nil {
		requirePhase16M004Refusal(t, err, "foreign")
		return
	}
	t.Fatal("expected terminal Phase 16 M004 foreign refusal")
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
	result, _, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	if err != nil {
		requirePhase16M004Refusal(t, err, "foreign")
		return
	}
	t.Fatal("expected terminal Phase 16 M004 foreign refusal")
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
		requirePhase4VerifierTerminalM004(t)
		return
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
	for _, required := range []string{"lang_record_event(\"value.transferred\"", "lang_record_event(\"function.returned\"", "lang_write_buffer_hex(&lang_entry_output)"} {
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
	assertOwnedSemanticDriftRejected(t, expected, mutated)
}

func TestOwnedExecutionFieldMutationMatrix(t *testing.T) {
	expected := ownedInterpreterExecution(t)
	operationalRunner := &fakeNativeRunner{err: &native.ToolError{Code: "native.invalid_execution", Err: errors.New("malformed child stdout")}}
	operational, err := session.RunNativeCommandFile(context.Background(), testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang"), operationalRunner)
	if err != nil || operational.Status != protocol.StatusOperational || protocol.ExitCode(operational.Status) != 3 || operational.Diagnostics[0].Code != "native.invalid_execution" {
		t.Fatalf("malformed native output classification changed: result=%+v err=%v", operational, err)
	}
	tests := []struct {
		name       string
		normalized bool
		mutate     func(*execution.Execution)
	}{
		{name: "schema", normalized: true, mutate: func(value *execution.Execution) { value.Schema = execution.Schema0 }},
		{name: "outcome kind", mutate: func(value *execution.Execution) { value.Outcome.Kind = "other" }},
		{name: "outcome value", mutate: func(value *execution.Execution) { value.Outcome.Value = "other" }},
		{name: "event schema", normalized: true, mutate: func(value *execution.Execution) { value.Events[0].Schema = execution.Schema0 }},
		{name: "event id", mutate: func(value *execution.Execution) { value.Events[0].ID = "other" }},
		{name: "event kind", mutate: func(value *execution.Execution) { value.Events[0].Kind = "other" }},
		{name: "function id", mutate: func(value *execution.Execution) { value.Events[0].FunctionID = "other" }},
		{name: "input", normalized: true, mutate: func(value *execution.Execution) { value.Events[0].Input = "physical-input" }},
		{name: "output", normalized: true, mutate: func(value *execution.Execution) { value.Events[0].Output = "physical-output" }},
		{name: "source place", mutate: func(value *execution.Execution) { value.Events[0].SourcePlace = "other" }},
		{name: "target place", mutate: func(value *execution.Execution) { value.Events[0].TargetPlace = "other" }},
		{name: "type id", mutate: func(value *execution.Execution) { value.Events[0].TypeID = "other" }},
		{name: "live resources", mutate: func(value *execution.Execution) { value.LiveResources = []string{"resource:other"} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := cloneExecution(t, expected)
			test.mutate(&mutated)
			if test.normalized {
				source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang"))
				if err != nil {
					t.Fatal(err)
				}
				checked := session.Check(source)
				want, err := session.ProjectExecutionSchema2(checked.Program, expected)
				if err != nil {
					t.Fatal(err)
				}
				projected, err := session.ProjectExecutionSchema2(checked.Program, mutated)
				if err != nil {
					t.Fatal(err)
				}
				if !execution.Equal(want, projected) {
					t.Fatalf("legacy-only field %q survived schema-2 projection:\nwant=%+v\ngot =%+v", test.name, want, projected)
				}
				return
			}
			assertOwnedSemanticDriftRejected(t, expected, mutated)
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

func assertOwnedSemanticDriftRejected(t *testing.T, expected, mutated execution.Execution) {
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
	if err == nil {
		t.Fatal("semantic drift passed both schema-2 peer validation and engine comparison")
	}
	commandRunner := &fakeNativeRunner{results: []native.Result{
		{Optimization: "-O0", Pairs: []native.Pair{{Input: "01020304", Execution: expected}}},
		{Optimization: "-O3", Pairs: []native.Pair{{Input: "01020304", Execution: mutated}}},
	}}
	result, commandErr := session.RunNativeCommandFile(context.Background(), path, commandRunner)
	if commandErr != nil || result.Status == protocol.StatusPass || len(result.Diagnostics) != 1 {
		t.Fatalf("semantic drift was not rejected at the command boundary: result=%+v err=%v", result, commandErr)
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
		cSource, err := phase16FileFrozenEvidenceC(t, checked.Program, "testdata/phase4/"+fixture)
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
// output alone, leaving the digest-bound historical program C artifact
// untouched. This proves the header artifact is independently scanned; the
// archived program artifact is not emitted by today's public emitter.
func TestAttributeInjectionIntoHeaderOnlyMakesControlFail(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	cSource, err := phase16FileFrozenEvidenceC(t, checked.Program, "testdata/phase4/foreign_acquire_one.lang")
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
	cSource, err := phase16FileFrozenEvidenceC(t, checked.Program, "testdata/phase4/foreign_acquire_one.lang")
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
// mutation-kill by injecting a banned token into a copy of the digest-bound
// historical C artifact. It proves the scanner is not vacuously green; it is
// not evidence that the current emitter accepts foreign bodies.
func TestAttributeInjectionMakesControlFail(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	cSource, err := phase16FileFrozenEvidenceC(t, checked.Program, "testdata/phase4/foreign_acquire_one.lang")
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
func requirePhase4VerifierTerminalM004(t testing.TB) {
	t.Helper()
	path := testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.lang")
	_, _, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	requirePhase16M004Refusal(t, err, "foreign")
}

func TestVerifyPhase4ForeignLayoutControls(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase4"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		requirePhase4VerifierTerminalM004(t)
		return
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
		requirePhase4VerifierTerminalM004(t)
		return
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
		requirePhase4VerifierTerminalM004(t)
		return
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
		requirePhase4VerifierTerminalM004(t)
		return
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
		requirePhase4VerifierTerminalM004(t)
		return
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
		requirePhase4VerifierTerminalM004(t)
		return
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
		requirePhase4VerifierTerminalM004(t)
		return
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
	if nativeResult.Status != protocol.StatusOperational || len(nativeResult.Diagnostics) != 1 || nativeResult.Diagnostics[0].Code != "native.tool_failure" {
		t.Fatalf("expected lang run --engine=native to stop at the terminal Phase 16 foreign M004 refusal, got %+v", nativeResult)
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

// TestNonlocalExitEmitsLeakPerLiveAcquisition replays Phase 16's
// digest-bound historical C artifact against the frozen foreign translation
// unit. It preserves the old D-04-17 receipt and does not claim the retired
// emitter is available through current Emit/EmitNative.
func TestNonlocalExitEmitsLeakPerLiveAcquisition(t *testing.T) {
	checked := nonlocalProbeChecked(t)
	generated, err := phase16FileFrozenEvidenceC(t, checked.Program, "testdata/phase4/nonlocal_exit_probe.lang")
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

// TestPadRunsNoRelease inspects Phase 16's digest-bound historical C
// artifact for the D-04-18 property. It is archival evidence for the removed
// foreign lowering, not a current native-cleanup witness.
func TestPadRunsNoRelease(t *testing.T) {
	checked := nonlocalProbeChecked(t)
	generated, err := phase16FileFrozenEvidenceC(t, checked.Program, "testdata/phase4/nonlocal_exit_probe.lang")
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

// TestNonlocalExitProbeInterpreterNative replays the archived Phase 16 C
// artifact and compares it with the interpreter's historical probe model.
// Current whole-program emission refuses foreign-call bodies, so this is
// retained as a bounded archival comparison only.
func TestNonlocalExitProbeInterpreterNative(t *testing.T) {
	checked := nonlocalProbeChecked(t)
	interpreted, err := interp.Run(checked.Program, "main", "7")
	if err != nil {
		t.Fatal(err)
	}
	generated, err := phase16FileFrozenEvidenceC(t, checked.Program, "testdata/phase4/nonlocal_exit_probe.lang")
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

// TestNonlocalExitDetectionIsMutationKilled is the archival D-04-21
// mutation-kill receipt for control:foreign.nonlocal_exit_undetected. It
// mutates the digest-bound Phase 16 C artifact; it does not exercise a
// current production emitter path.
func TestNonlocalExitDetectionIsMutationKilled(t *testing.T) {
	checked := nonlocalProbeChecked(t)
	generated, err := phase16FileFrozenEvidenceC(t, checked.Program, "testdata/phase4/nonlocal_exit_probe.lang")
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

// TestLeakCountMatchesLiveAcquisitions preserves the archival D-09/D-10
// mutation receipt for control:foreign.nonlocal_exit_undetected. Its golden
// and mutated inputs are historical C artifacts; this does not demonstrate
// current emitter cleanup behavior.
func TestLeakCountMatchesLiveAcquisitions(t *testing.T) {
	checked := nonlocalProbeChecked(t)
	generated, err := phase16FileFrozenEvidenceC(t, checked.Program, "testdata/phase4/nonlocal_exit_probe.lang")
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
		requirePhase4VerifierTerminalM004(t)
		return
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
	if result.Status != protocol.StatusPass {
		requirePhase4VerifierTerminalM004(t)
		return
	}
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
		if run.Exit != 3 || len(run.Stderr) != 0 {
			t.Fatalf("verify testdata/phase4: %+v", run)
		}
	}
	var decoded protocol.Result
	if err := json.Unmarshal(run.Stdout, &decoded); err != nil {
		t.Fatalf("decode verify output: %v (stdout=%s)", err, run.Stdout)
	}
	if decoded.Status != protocol.StatusPass {
		requirePhase4VerifierTerminalM004(t)
		return
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
		requirePhase4VerifierTerminalM004(t)
		return
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
// typed-failure path's historical terminal outcome, events, and live
// resources requires a real ok=0 SOMEWHERE in the archived toolchain -- this double supplies
// it without touching production source or the frozen TU. Current
// whole-program native emission refuses foreign-call bodies; later code in
// this file uses the digest-bound historical C receipt.
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
// assertion for the historical second-stage and third-stage typed-failure
// path shapes. The interpreter runs the exact err block directly via
// interptestdirect.RunLinearBlockDirect (Run's own public entry point cannot reach
// it, since the interpreter's documented discretionary stub always
// simulates success for every OpForeignCall, D-04-04/04-PATTERNS Pattern
// 3), while native optimization levels use the digest-bound historical C
// artifact with writeFailingForeignDouble's test-only object. This retains
// prior evidence but is not current emitter admission or cleanup proof.
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

	cSource, err := phase16FileFrozenEvidenceC(t, program, "testdata/phase4/"+fixture)
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
			requirePhase16M004Refusal(t, err, "foreign")
			return
		}
		t.Fatal("expected terminal Phase 16 M004 foreign refusal")
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
			requirePhase16M004Refusal(t, err, "foreign")
			return
		}
		t.Fatal("expected terminal Phase 16 M004 foreign refusal")
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

// debtRegisterNamingPattern is the naming convention (D-14-24) required of
// any register created from M003 on: PHASE-<NN>-DEBT.md.
var debtRegisterNamingPattern = regexp.MustCompile(`^PHASE-\d{2}-DEBT\.md$`)

// debtRegisterNamingExemptions lists every register written before M003's
// PHASE-<NN>-DEBT.md naming convention existed. Frozen prior art only -- any
// register created from M003 on must satisfy debtRegisterNamingPattern
// without an entry here; adding one is a review failure unless the register
// genuinely predates the convention.
var debtRegisterNamingExemptions = map[string]string{
	"02-DEBT.md": "written 2026-09-04, M001, before the PHASE-<NN>-DEBT.md naming convention existed; frozen prior art",
	"03-DEBT.md": "written 2026-09-04, M001, before the PHASE-<NN>-DEBT.md naming convention existed; frozen prior art",
	"04-DEBT.md": "written 2026-09-05, M001, before the PHASE-<NN>-DEBT.md naming convention existed; frozen prior art",
	"05-DEBT.md": "written 2026-09-06, M001, before the PHASE-<NN>-DEBT.md naming convention existed; frozen prior art",
	"06-DEBT.md": "written 2026-09-07, M001, before the PHASE-<NN>-DEBT.md naming convention existed; frozen prior art",
}

// The closed three-form owning-phase vocabulary PRC-01 requires (D-14-24):
// a phase identifier (P<NN>), a closed-with-commit form naming a commit
// sha, or an unowned-with-witness form naming a witness identifier. Free
// prose no longer satisfies "non-empty" -- a cell must resolve to exactly
// one of these three syntactic shapes.
//
// The witness identifier inside UNOWNED(...) is NOT required to resolve to
// an executed probe here -- that resolution check, and the register's own
// Witness column, land in plan 14-07 (D-14-23/D-14-24's own stated
// hand-off). This plan mechanizes the SHAPE of the vocabulary only.
var (
	debtRegisterPhaseIDPattern = regexp.MustCompile(`^P\d{2}$`)
	debtRegisterClosedPattern  = regexp.MustCompile(`^CLOSED\([0-9a-fA-F]{7,40}\)$`)
	debtRegisterUnownedPattern = regexp.MustCompile(`^UNOWNED\([A-Za-z0-9][A-Za-z0-9:_./-]*\)$`)
)

// debtRegisterOwningPhaseForm reports whether cell -- already trimmed by the
// caller -- is one of the three closed forms. Trimming happens at the call
// site so that a whitespace-only cell reduces to "" and is rejected
// identically to a genuinely absent cell: neither is a distinct valid
// owner.
func debtRegisterOwningPhaseForm(cell string) bool {
	return debtRegisterPhaseIDPattern.MatchString(cell) ||
		debtRegisterClosedPattern.MatchString(cell) ||
		debtRegisterUnownedPattern.MatchString(cell)
}

// debtRegisterFirstRecordedPattern matches the D-14-31 two-milestone-carry
// guard's per-row metadata line, required in every row's detail section:
// `first-recorded: M0NN`.
var debtRegisterFirstRecordedPattern = regexp.MustCompile(`(?m)^first-recorded:\s*(M\d{3})\s*$`)

// ---------------------------------------------------------------------
// Plan 14-07: the Grade and Witness columns (EVD-03/EVD-04, D-14-21
// through D-14-30). Extends checkDebtRegister in place -- same function,
// same file, no new parser -- with a closed four-kind executed witness
// grammar and the two D-14-29 closure assertions. A phase identifier is
// NEVER a witness (D-14-23): it may appear only in the Landing phase cell
// plan 14-04 closed, because whether a phase lands is a human editing
// ROADMAP.md, an artifact read, not an executed claim.
// ---------------------------------------------------------------------

// debtRegisterGradeWitnessExemptions names every register written before
// the Grade/Witness columns existed AND not otherwise migrated by this
// plan. Frozen prior art only: mechanically deriving and witnessing the
// ~68 remaining pre-existing rows across 02-DEBT.md through
// PHASE-10-DEBT.md is out of this plan's budget (recorded as debt in this
// plan's own SUMMARY, not silently deferred). Any register created or
// extended to carry Grade/Witness cells must NOT be listed here --
// PHASE-13-DEBT.md and PHASE-14-DEBT.md are migrated in full by this
// plan, and PHASE-11-DEBT.md/PHASE-12-DEBT.md are migrated in full too
// (every row DEFINED/n/a except the one claim each file's day-one table
// names, D-11-02 and D-12-43, which carry real probes).
var debtRegisterGradeWitnessExemptions = map[string]string{
	"02-DEBT.md":       "written 2026-09-04, before Grade/Witness existed; frozen prior art",
	"03-DEBT.md":       "written 2026-09-04, before Grade/Witness existed; frozen prior art",
	"04-DEBT.md":       "written 2026-09-05, before Grade/Witness existed; frozen prior art",
	"05-DEBT.md":       "written 2026-09-06, before Grade/Witness existed; frozen prior art",
	"06-DEBT.md":       "written 2026-09-07, before Grade/Witness existed; frozen prior art",
	"PHASE-07-DEBT.md": "written before Grade/Witness existed (plan 14-07); frozen prior art",
	"PHASE-08-DEBT.md": "written before Grade/Witness existed (plan 14-07); frozen prior art",
	"PHASE-09-DEBT.md": "written before Grade/Witness existed (plan 14-07); frozen prior art",
	"PHASE-10-DEBT.md": "written before Grade/Witness existed (plan 14-07); frozen prior art",
}

// debtRegisterGrades is the closed grade vocabulary. This plan (EVD-03/04)
// declares Grade rather than mechanically deriving a ceiling from evidence
// (EVD-02's D-14-01 cap, a separate law) -- the same vocabulary, reused as
// a closed-membership check the way Severity already is.
var debtRegisterGrades = map[string]bool{
	"DEFINED": true, "WIRED": true, "REACHABLE": true, "EXERCISED": true, "MUTATION-KILLED": true,
}

// debtRegisterGradeOrdinal ranks the closed grade vocabulary so
// "below the satisfying bar" (D-14-29) is a numeric comparison, not a
// string special-case.
var debtRegisterGradeOrdinal = map[string]int{
	"DEFINED": 0, "WIRED": 1, "REACHABLE": 2, "EXERCISED": 3, "MUTATION-KILLED": 4,
}

// debtRegisterSatisfyingGradeOrdinal is EVD-02's own satisfying bar
// (>= EXERCISED); a row graded below this must carry a witness or be
// explicitly withdrawn (D-14-29).
var debtRegisterSatisfyingGradeOrdinal = debtRegisterGradeOrdinal["EXERCISED"]

// debtRegisterGradeCellPattern accepts a bare closed-vocabulary grade, or
// the same grade suffixed " (withdrawn)" -- the explicit-withdrawal escape
// hatch D-14-29's closure assertion names for a below-bar row that carries
// no witness because the claim itself has been retracted, not merely
// deferred.
var debtRegisterGradeCellPattern = regexp.MustCompile(`^(DEFINED|WIRED|REACHABLE|EXERCISED|MUTATION-KILLED)( \(withdrawn\))?$`)

// The closed four-kind witness grammar (D-14-23). Free text and a bare
// phase identifier (P\d\d) both fail by construction: neither matches any
// of these four patterns.
var (
	debtRegisterProbeTokenPattern    = regexp.MustCompile(`^probe:([A-Za-z][A-Za-z0-9_]*)$`)
	debtRegisterCallsiteTokenPattern = regexp.MustCompile(`^callsite:([A-Za-z0-9_./]+)\.([A-Za-z][A-Za-z0-9_]*)=(\d+)$`)
	debtRegisterEscapeTokenPattern   = regexp.MustCompile(`^escape:([a-z][a-z0-9-]*)$`)
	debtRegisterEnvTokenPattern      = regexp.MustCompile(`^env:([a-z][a-z0-9_-]*)$`)
)

// debtRegisterEscapeRegistry is the closed set of declared escape
// identifiers an escape: witness token may name (D-14-23/D-14-27). Each
// entry's own probe: token is independently resolved through
// debtRegisterWitnessTokenProblem -- an escape is a witnessed admission,
// never a silent skip.
var debtRegisterEscapeRegistry = map[string]string{
	"callback-invocation-unsubjected": "probe:TestRetainedPointerEscapeIsStillUnsubjected",
}

// debtRegisterEnvironmentalSet is the closed set of env: witness tokens
// (D-14-23): host/toolchain preconditions that never close and are
// explicitly not debt, but must still be declared -- never a
// citation-free skip laundered through a new invented env value.
var debtRegisterEnvironmentalSet = map[string]bool{
	"clang": true,
	"otool": true,
}

// debtRegisterModuleTestNames scans every *_test.go file in the module
// (honoring build constraints via go/build.Context.MatchFile, exactly as
// verification_groundedness_test.go's buildTestIndex does) and returns the
// flat set of top-level Test/Fuzz/Benchmark/Example identifiers declared
// anywhere -- a probe: witness token's existence check. Rebuilt on every
// call rather than cached: the module is small (~7ms per
// verification_groundedness_test.go's own measurement) and a stale cache
// across seeded-fixture subtests is a worse failure mode than a rebuild.
func debtRegisterModuleTestNames() (map[string]bool, error) {
	names := make(map[string]bool)
	root := testsupport.ProjectPath()
	fset := token.NewFileSet()
	buildCtx := build.Default
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			base := d.Name()
			if base == ".git" || base == "testdata" || (path != root && strings.HasPrefix(base, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		dir := filepath.Dir(path)
		match, matchErr := buildCtx.MatchFile(dir, filepath.Base(path))
		if matchErr != nil {
			return matchErr
		}
		if !match {
			return nil
		}
		file, parseErr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if parseErr != nil {
			return parseErr
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			if isTestLikeFunc(fn) {
				names[fn.Name.Name] = true
			}
		}
		return nil
	})
	return names, err
}

// debtRegisterProbeExists reports whether name is a top-level Test-shaped
// function declared anywhere in the module.
func debtRegisterProbeExists(name string) (bool, error) {
	names, err := debtRegisterModuleTestNames()
	if err != nil {
		return false, err
	}
	return names[name], nil
}

// debtRegisterCallSiteWitnessMatches resolves a callsite: token
// (pkgPath.Symbol=count) via a stdlib go/ast scan of pkgPath's PRODUCTION
// (non-test) .go files only, counting exact CALL sites of Symbol -- the
// anti-source-search rule (D-14-23): a Symbol that is not declared at all
// in the package is a FAIL, never a count of zero, so a rename is loud
// instead of silently reading as "zero call sites." Scanning production
// files only (excluding _test.go) is deliberate: a call site inside a unit
// test exercising the function directly does not make the function
// reachable from the production pipeline, which is exactly the claim a
// row like D-13-02b's `callsite:internal/compiler/check.resolveBlame=0`
// makes (resolveBlame is unit-tested directly but never called from
// production code -- see this plan's SUMMARY).
func debtRegisterCallSiteWitnessMatches(pkgPath, symbol string, want int) (bool, string) {
	dir := testsupport.ProjectPath(strings.Split(pkgPath, "/")...)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, fmt.Sprintf("callsite package %q does not exist: %v", pkgPath, err)
	}
	fset := token.NewFileSet()
	declared := false
	count := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		file, parseErr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if parseErr != nil {
			return false, fmt.Sprintf("callsite scan: parse %s: %v", path, parseErr)
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == symbol {
				declared = true
			}
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				if fn.Name == symbol {
					count++
				}
			case *ast.SelectorExpr:
				if fn.Sel.Name == symbol {
					count++
				}
			}
			return true
		})
	}
	if !declared {
		return false, fmt.Sprintf("callsite symbol %q does not exist in package %q -- absence is never a pass", symbol, pkgPath)
	}
	if count != want {
		return false, fmt.Sprintf("callsite %s.%s: expected %d call site(s), found %d", pkgPath, symbol, want, count)
	}
	return true, ""
}

// debtRegisterWitnessTokenProblem resolves ONE witness token against the
// closed four-kind grammar (D-14-23) and returns a non-empty problem
// string when it fails to resolve, "" when it holds. Free text and a bare
// phase identifier both fall through to the default case and fail --
// there is no fifth kind.
func debtRegisterWitnessTokenProblem(tok string) string {
	switch {
	case debtRegisterProbeTokenPattern.MatchString(tok):
		name := debtRegisterProbeTokenPattern.FindStringSubmatch(tok)[1]
		ok, err := debtRegisterProbeExists(name)
		if err != nil {
			return fmt.Sprintf("witness token %q: %v", tok, err)
		}
		if !ok {
			return fmt.Sprintf("witness token %q names a test that does not exist", tok)
		}
		return ""
	case debtRegisterCallsiteTokenPattern.MatchString(tok):
		m := debtRegisterCallsiteTokenPattern.FindStringSubmatch(tok)
		pkgPath, symbol, wantStr := m[1], m[2], m[3]
		want, convErr := strconv.Atoi(wantStr)
		if convErr != nil {
			return fmt.Sprintf("witness token %q: unreadable call count: %v", tok, convErr)
		}
		if ok, reason := debtRegisterCallSiteWitnessMatches(pkgPath, symbol, want); !ok {
			return fmt.Sprintf("witness token %q: %s", tok, reason)
		}
		return ""
	case debtRegisterEscapeTokenPattern.MatchString(tok):
		id := debtRegisterEscapeTokenPattern.FindStringSubmatch(tok)[1]
		probeToken, known := debtRegisterEscapeRegistry[id]
		if !known {
			return fmt.Sprintf("witness token %q names an escape identifier absent from the closed escape registry", tok)
		}
		if problem := debtRegisterWitnessTokenProblem(probeToken); problem != "" {
			return fmt.Sprintf("witness token %q: its registered probe does not resolve: %s", tok, problem)
		}
		return ""
	case debtRegisterEnvTokenPattern.MatchString(tok):
		id := debtRegisterEnvTokenPattern.FindStringSubmatch(tok)[1]
		if !debtRegisterEnvironmentalSet[id] {
			return fmt.Sprintf("witness token %q names an environmental value outside the closed set", tok)
		}
		return ""
	default:
		return fmt.Sprintf("witness token %q is outside the closed four-kind grammar (probe:/callsite:/escape:/env:) -- free text and bare phase identifiers are never witnesses", tok)
	}
}

// debtRegisterWitnessCellProblems splits cell on commas -- several
// witnesses separated by commas pass only when ALL of them hold (D-14-23)
// -- and returns every unresolved token's problem, in cell order.
func debtRegisterWitnessCellProblems(cell string) []string {
	var problems []string
	for _, tok := range strings.Split(cell, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			problems = append(problems, "empty witness token between commas")
			continue
		}
		if problem := debtRegisterWitnessTokenProblem(tok); problem != "" {
			problems = append(problems, problem)
		}
	}
	return problems
}

// debtRegisterGradeWitnessRowProblems checks one row's Grade and Witness
// cells once both columns are known to be present, applying D-14-29's
// closure assertion: every row graded below the satisfying bar carries a
// witness (every comma-separated token resolves) or is explicitly
// withdrawn. "n/a" is the literal, explicit spelling for "no witness
// declared" -- distinct from an empty cell, which is always a problem.
//
// DEFINED is exempted from the witness-or-withdrawn requirement: D-14-03's
// own ladder treats DEFINED as "the floor, always permitted" -- a bare
// declaration that a claim exists, asserting nothing a witness could
// substantiate. WIRED and REACHABLE, by contrast, are affirmative
// below-bar claims ("this exists," "this reaches") and DO require a
// witness or an explicit withdrawal. Without this exemption, a genuinely
// open, not-yet-actionable debt row (e.g. PHASE-14-DEBT.md's D-14-46/47,
// prospective triggers with no executed probe to cite yet) would have no
// honest way to be recorded: marking it "(withdrawn)" would misstate that
// the claim has been retracted, when it has not.
func debtRegisterGradeWitnessRowProblems(name, identifier, gradeCell, witnessCell string) []string {
	var problems []string
	m := debtRegisterGradeCellPattern.FindStringSubmatch(strings.TrimSpace(gradeCell))
	if m == nil {
		return []string{fmt.Sprintf("%s: row %s has Grade %q outside the closed vocabulary (%v)", name, identifier, gradeCell, debtRegisterGrades)}
	}
	grade := m[1]
	withdrawn := m[2] != ""
	if withdrawn && grade != "DEFINED" {
		problems = append(problems, fmt.Sprintf("%s: row %s: only DEFINED may be marked (withdrawn)", name, identifier))
	}
	ordinal := debtRegisterGradeOrdinal[grade]
	belowBar := ordinal > debtRegisterGradeOrdinal["DEFINED"] && ordinal < debtRegisterSatisfyingGradeOrdinal
	witnessCell = strings.TrimSpace(witnessCell)
	switch {
	case witnessCell == "":
		problems = append(problems, fmt.Sprintf("%s: row %s has an empty Witness cell", name, identifier))
	case witnessCell == "n/a":
		if belowBar && !withdrawn {
			problems = append(problems, fmt.Sprintf("%s: row %s is graded %s (below the satisfying bar) with Witness \"n/a\" and is not explicitly withdrawn -- every below-bar claim must carry a witness or be withdrawn (D-14-29)", name, identifier, grade))
		}
	default:
		for _, problem := range debtRegisterWitnessCellProblems(witnessCell) {
			problems = append(problems, fmt.Sprintf("%s: row %s: %s", name, identifier, problem))
		}
	}
	return problems
}

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
//   - post-legacy registers additionally name a landing phase drawn from
//     the closed three-form owning-phase vocabulary (PRC-01, D-14-24),
//   - any register created from M003 on is named PHASE-<NN>-DEBT.md,
//   - every row's detail section carries a first-recorded: milestone, and
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

// phase20DebtAuditCohort is the M002 audit's explicitly enumerated ten-item
// carry-forward. Keep this exact set pinned so edits to the source audit or
// current registers cannot silently redefine PRC-02's starting population.
var phase20DebtAuditCohort = []string{
	"D-10-C04", "D-11-02", "D-11-27", "D-11-51", "D-12-21",
	"D-12-36", "D-12-43", "D-13-02b", "D-13-10a", "D-13-34",
}

// phase20DebtNewM003IDs are the seven distinct rows introduced by the live
// M003 PHASE-14-DEBT register. The eight-row UNREACHABLE-CLAIMS view is a
// probe-backed subset and is not denominator authority.
var phase20DebtNewM003IDs = []string{
	"D-14-45", "D-14-46", "D-14-47", "D-14-50", "D-14-51", "D-14-52", "D-14-54",
}

var phase20DebtStartingUnownedIDs = []string{
	"D-10-C04", "D-11-02", "D-11-27", "D-12-36", "D-12-43", "D-13-34",
	"D-14-45", "D-14-46", "D-14-47", "D-14-50", "D-14-51", "D-14-52", "D-14-54",
}

var phase20DebtCurrentUnownedIDs = []string{
	"D-10-C04", "D-12-43", "D-14-46", "D-14-47",
}

type phase20DebtDisposition struct {
	ID, Landing, Register, Witness string
}

// phase20QualifiedDebtDispositions resolves only the historical M002 audit
// cohort plus the new live M003 rows. It reads each current Landing phase
// cell from the source register, rejecting missing/duplicate provenance and
// unknown disposition syntax instead of counting raw UNOWNED prose.
func phase20QualifiedDebtDispositions() (map[string]phase20DebtDisposition, error) {
	root := testsupport.ProjectPath(".planning")
	auditBytes, err := os.ReadFile(filepath.Join(root, "milestones", "M002-MILESTONE-AUDIT.md"))
	if err != nil {
		return nil, err
	}
	audit := string(auditBytes)
	anchor := "### The 10 items that are OPEN and UNOWNED"
	start := strings.Index(audit, anchor)
	if start < 0 {
		return nil, fmt.Errorf("M002 milestone audit is missing the ten-item carry-forward table")
	}
	section := audit[start+len(anchor):]
	if end := strings.Index(section, "\n## "); end >= 0 {
		section = section[:end]
	}
	var auditIDs []string
	for _, line := range strings.Split(section, "\n") {
		if !strings.HasPrefix(line, "| D-") {
			continue
		}
		cells := strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|")
		if len(cells) < 2 {
			return nil, fmt.Errorf("malformed M002 audit carry-forward row %q", line)
		}
		auditIDs = append(auditIDs, strings.TrimSpace(cells[0]))
	}
	if !equalStringSets(auditIDs, phase20DebtAuditCohort) || len(auditIDs) != len(phase20DebtAuditCohort) {
		return nil, fmt.Errorf("M002 audit cohort drift: got %v, want %v", auditIDs, phase20DebtAuditCohort)
	}

	qualified := make(map[string]bool, len(phase20DebtAuditCohort)+len(phase20DebtNewM003IDs))
	for _, id := range phase20DebtAuditCohort {
		qualified[id] = true
	}
	for _, id := range phase20DebtNewM003IDs {
		if qualified[id] {
			return nil, fmt.Errorf("duplicate qualified debt identifier %s", id)
		}
		qualified[id] = true
	}
	paths, err := phaseArtifactGlob("*", "*-DEBT.md")
	if err != nil {
		return nil, err
	}
	found := make(map[string]phase20DebtDisposition, len(qualified))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		columns, rows, err := parseDebtRegisterTable(filepath.Base(path), string(data))
		if err != nil {
			continue // unrelated legacy register shape is checked by PRC-01 separately
		}
		idColumn, hasID := columns["ID"]
		landingColumn, hasLanding := columns["Landing phase"]
		if !hasID || !hasLanding {
			continue
		}
		witnessColumn, hasWitness := columns["Witness"]
		for _, row := range rows {
			id := row[idColumn]
			landing := strings.TrimSpace(row[landingColumn])
			if filepath.Base(path) == "PHASE-14-DEBT.md" && strings.HasPrefix(landing, "UNOWNED(") && !containsString(phase20DebtNewM003IDs, id) {
				return nil, fmt.Errorf("unknown current M003 debt provenance %s in %s", id, path)
			}
			if !qualified[id] {
				continue
			}
			if _, duplicate := found[id]; duplicate {
				return nil, fmt.Errorf("qualified debt identifier %s has duplicate current provenance", id)
			}
			if !debtRegisterOwningPhaseForm(landing) {
				return nil, fmt.Errorf("qualified debt identifier %s has unknown current provenance %q", id, landing)
			}
			witness := ""
			if hasWitness && witnessColumn < len(row) {
				witness = row[witnessColumn]
			}
			found[id] = phase20DebtDisposition{ID: id, Landing: landing, Register: filepath.ToSlash(strings.TrimPrefix(path, root+string(filepath.Separator))), Witness: witness}
		}
	}
	for id := range qualified {
		if _, ok := found[id]; !ok {
			return nil, fmt.Errorf("qualified debt identifier %s has no current source register row", id)
		}
	}
	return found, nil
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// equalStringSets compares exact membership while allowing source table order.
func equalStringSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]bool, len(a))
	for _, value := range a {
		if seen[value] {
			return false
		}
		seen[value] = true
	}
	for _, value := range b {
		if !seen[value] {
			return false
		}
	}
	return true
}

func phase20OpenUnownedIDs(dispositions map[string]phase20DebtDisposition) []string {
	var ids []string
	for id, disposition := range dispositions {
		if strings.HasPrefix(disposition.Landing, "UNOWNED(") {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

// phase20WithinUnownedDebtCap is intentionally pure so the seeded six-item
// control proves the threshold independently from today's register contents.
func phase20WithinUnownedDebtCap(ids []string) bool { return len(ids) <= 5 }

func TestPhase20UnownedDebtPopulation(t *testing.T) {
	dispositions, err := phase20QualifiedDebtDispositions()
	if err != nil {
		t.Fatal(err)
	}
	if len(phase20DebtAuditCohort) != 10 {
		t.Fatalf("authoritative M002 historical cohort has %d IDs, want 10", len(phase20DebtAuditCohort))
	}
	baselines, err := phaseArtifactGlob("20-nyquist-d-13-34-and-the-frontier-fixture", "20-DEBT-BASELINE.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(baselines) != 1 {
		t.Fatalf("expected exactly one Phase 20 debt baseline, found %d: %v", len(baselines), baselines)
	}
	baselineBytes, err := os.ReadFile(baselines[0])
	if err != nil {
		t.Fatal(err)
	}
	var baselineIDs []string
	for _, id := range phase20DebtStartingUnownedIDs {
		if !strings.Contains(string(baselineBytes), "`"+id+"`") {
			t.Fatalf("baseline does not pin starting identifier %s", id)
		}
		baselineIDs = append(baselineIDs, "`"+id+"`")
	}
	if !strings.Contains(string(baselineBytes), strings.Join(baselineIDs, ", ")) {
		t.Fatal("baseline does not pin the exact 13-ID phase-start population")
	}
	got := phase20OpenUnownedIDs(dispositions)
	if !equalStringSets(got, phase20DebtCurrentUnownedIDs) {
		t.Fatalf("current source-derived open-unowned IDs changed: got %v, want %v", got, phase20DebtCurrentUnownedIDs)
	}
	viewBytes, err := os.ReadFile(testsupport.ProjectPath(".planning", "UNREACHABLE-CLAIMS.md"))
	if err != nil {
		t.Fatal(err)
	}
	view := string(viewBytes)
	for _, id := range got {
		if strings.Contains(dispositions[id].Witness, "probe:") && !strings.Contains(view, "| "+id+" | ") {
			t.Errorf("probe-backed open row %s is absent from the generated claims view", id)
		}
	}
}

func TestPhase20UnownedDebtCapRule(t *testing.T) {
	if !phase20WithinUnownedDebtCap([]string{"a", "b", "c", "d", "e"}) {
		t.Fatal("five open-unowned rows should satisfy the cap")
	}
	if phase20WithinUnownedDebtCap([]string{"a", "b", "c", "d", "e", "f"}) {
		t.Fatal("seeded sixth open-unowned row should exceed the five-item cap")
	}
	dispositions := map[string]phase20DebtDisposition{
		"closed": {ID: "closed", Landing: "CLOSED(abc1234)"},
		"owned":  {ID: "owned", Landing: "P21"},
	}
	if got := phase20OpenUnownedIDs(dispositions); len(got) != 0 {
		t.Fatalf("closed/owned historical rows must not count as open-unowned: %v", got)
	}
}

// debtRegisterGlobalIdentifierProblems is D-14-29's closure assertion (ii):
// "every row of the generated view appears in exactly one register." The
// generated view (Task 4, .planning/UNREACHABLE-CLAIMS.md) is derived
// one-to-one from register rows, so this reduces to a checkable invariant
// over the registers themselves, ahead of the view existing: no D-XX-NN
// identifier may appear as a table row in more than one *-DEBT.md
// register. A duplicate would mean the eventual view has two candidate
// sources for the same row -- an ambiguity the "derived, not authored"
// topology (D-14-13) forbids by construction.
func debtRegisterGlobalIdentifierProblems(paths []string) ([]string, error) {
	owner := make(map[string]string)
	var problems []string
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		columns, rows, tableErr := parseDebtRegisterTable(filepath.Base(path), string(data))
		if tableErr != nil {
			// A malformed table is already reported by
			// debtRegisterProblems/checkDebtRegister -- do not duplicate
			// that failure here, but do not silently skip the file either.
			continue
		}
		idIndex, ok := columns["ID"]
		if !ok {
			continue
		}
		for _, row := range rows {
			if idIndex >= len(row) {
				continue
			}
			identifier := row[idIndex]
			if existing, seen := owner[identifier]; seen {
				if existing != path {
					problems = append(problems, fmt.Sprintf("identifier %s appears in two registers: %s and %s", identifier, existing, path))
				}
				continue
			}
			owner[identifier] = path
		}
	}
	return problems, nil
}

// TestDebtRegisterIdentifiersAreGloballyUnique is D-14-29's closure
// assertion (ii): "every row of the generated view appears in exactly one
// register." The generated view (Task 4) is sourced only from registers
// that carry Grade/Witness cells -- a register on
// debtRegisterGradeWitnessExemptions has no graded rows to contribute to
// the view at all -- so this check is scoped to the non-exempt registers.
// Scoping it to the whole corpus would misfire on the project's own
// existing, intentional carry-forward pattern (D-03-02 is deliberately
// re-cited from 03-DEBT.md into PHASE-07-DEBT.md, predating and distinct
// from this plan's Grade/Witness law), which is not the ambiguity this
// assertion exists to catch.
func TestDebtRegisterIdentifiersAreGloballyUnique(t *testing.T) {
	registers, err := phaseArtifactGlob("*", "*-DEBT.md")
	if err != nil {
		t.Fatal(err)
	}
	var graded []string
	for _, path := range registers {
		if _, exempt := debtRegisterGradeWitnessExemptions[filepath.Base(path)]; !exempt {
			graded = append(graded, path)
		}
	}
	problems, err := debtRegisterGlobalIdentifierProblems(graded)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) > 0 {
		t.Fatalf("%s", strings.Join(problems, "\n"))
	}
}

// TestDebtRegisterIdentifierUniquenessGuardIsNotInert is D-14-29's closure
// assertion (ii) proven non-inert: two temp copies of the same real
// register, sharing every identifier, must be reported as duplicates; the
// unmodified single-register case must not.
func TestDebtRegisterIdentifierUniquenessGuardIsNotInert(t *testing.T) {
	t.Run("single register has no duplicates", func(t *testing.T) {
		src := debtRegisterOwnershipFixture(t)
		problems, err := debtRegisterGlobalIdentifierProblems([]string{src})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) != 0 {
			t.Fatalf("a single register should never self-collide; got: %v", problems)
		}
	})

	t.Run("two registers sharing an identifier are flagged", func(t *testing.T) {
		first := debtRegisterOwnershipFixture(t)
		data, err := os.ReadFile(first)
		if err != nil {
			t.Fatal(err)
		}
		second := filepath.Join(filepath.Dir(first), "PHASE-98-DEBT.md")
		if err := os.WriteFile(second, data, 0o644); err != nil {
			t.Fatal(err)
		}
		problems, err := debtRegisterGlobalIdentifierProblems([]string{first, second})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("two registers sharing D-14-45 should be flagged as duplicate owners, but nothing was reported")
		}
		if !containsSubstring(problems, "D-14-45") {
			t.Fatalf("refusal does not name the duplicated identifier: %v", problems)
		}
	})
}

// checkDebtRegister is TestDebtRegistersAreWellFormed's *testing.T wrapper:
// it delegates every check to debtRegisterProblems (a pure, non-fataling
// function) and fails the test with every problem found, joined together,
// rather than stopping at the first -- satisfying D-14-24's "report every
// ownerless row, not just the first, deterministically" requirement in one
// pass. TestDebtRegisterOwnershipGuardIsNotInert calls debtRegisterProblems
// directly (the "variant that reports rather than fatals" Task 3 asks for)
// so a seeded fault's refusal text is inspectable without a nested subtest.
func checkDebtRegister(t *testing.T, path string) {
	t.Helper()
	problems, err := debtRegisterProblems(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) > 0 {
		t.Fatalf("%s", strings.Join(problems, "\n"))
	}
}

// debtRegisterProblems is the pure, table-order-deterministic core of the
// debt-register well-formedness law. It returns every problem found (nil
// means well-formed) instead of failing on the first, so a register with
// three ownerless rows reports all three, and a caller can inspect the
// exact refusal text without a *testing.T. A register whose `## Items`
// table holds zero rows is handled as a vacuous pass, never an index panic,
// provided the frontmatter's declared items: count also reads 0.
func debtRegisterProblems(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(path)
	text := string(data)

	if !strings.HasPrefix(text, "---\n") {
		return []string{fmt.Sprintf("%s: has no frontmatter block", name)}, nil
	}
	frontmatterEnd := strings.Index(text[4:], "\n---")
	if frontmatterEnd == -1 {
		return []string{fmt.Sprintf("%s: frontmatter block is unterminated", name)}, nil
	}
	declared := -1
	for _, line := range strings.Split(text[4:4+frontmatterEnd], "\n") {
		if !strings.HasPrefix(line, "items:") {
			continue
		}
		if _, scanErr := fmt.Sscanf(strings.TrimSpace(line), "items: %d", &declared); scanErr != nil {
			return []string{fmt.Sprintf("%s: unreadable `items:` frontmatter line %q", name, line)}, nil
		}
		break
	}
	if declared < 0 {
		return []string{fmt.Sprintf("%s: frontmatter declares no `items:` count", name)}, nil
	}

	columns, rows, tableErr := parseDebtRegisterTable(name, text)
	if tableErr != nil {
		return []string{tableErr.Error()}, nil
	}
	var problems []string
	if len(rows) != declared {
		problems = append(problems, fmt.Sprintf("%s: frontmatter declares items: %d but the Items table holds %d rows", name, declared, len(rows)))
	}

	_, landingExempt := debtRegisterLandingPhaseExemptions[name]
	_, gradeWitnessExempt := debtRegisterGradeWitnessExemptions[name]
	required := []string{"ID", "Source", "Threat/Req", "Severity", "Item"}
	if !landingExempt {
		required = append(required, "Landing phase")
	}
	if !gradeWitnessExempt {
		required = append(required, "Grade", "Witness")
	}
	for _, column := range required {
		if _, present := columns[column]; !present {
			problems = append(problems, fmt.Sprintf("%s: Items table has no %q column (columns: %v)", name, column, columns))
		}
	}
	if len(problems) > 0 {
		// A missing required column makes every row-level check below
		// meaningless (columns[...] would silently index 0) -- report the
		// shape defect now rather than cascade into confusing row errors.
		return problems, nil
	}

	if !debtRegisterNamingPattern.MatchString(name) {
		if _, exempt := debtRegisterNamingExemptions[name]; !exempt {
			problems = append(problems, fmt.Sprintf("%s: register filename does not match the PHASE-<NN>-DEBT.md convention required for any register created from M003 on, and is not on the frozen-prior-art exemption list", name))
		}
	}

	identifiers := make(map[string]bool, len(rows))
	var ownerless []string
	for _, row := range rows {
		identifier := row[columns["ID"]]
		if !strings.HasPrefix(identifier, "D-") || len(identifier) < len("D-00-0") {
			problems = append(problems, fmt.Sprintf("%s: row %q does not name a D-XX-NN identifier", name, identifier))
			continue
		}
		if identifiers[identifier] {
			problems = append(problems, fmt.Sprintf("%s: identifier %s appears in two rows", name, identifier))
			continue
		}
		identifiers[identifier] = true
		for _, column := range required {
			value := row[columns[column]]
			if value == "" || value == "-" {
				problems = append(problems, fmt.Sprintf("%s: row %s has an empty %q cell", name, identifier, column))
			}
		}
		if severity := row[columns["Severity"]]; !debtRegisterSeverities[severity] {
			problems = append(problems, fmt.Sprintf("%s: row %s has severity %q outside the closed vocabulary (blocker, warning, info)", name, identifier, severity))
		}
		if !landingExempt {
			raw := row[columns["Landing phase"]]
			cell := strings.TrimSpace(raw)
			if !debtRegisterOwningPhaseForm(cell) {
				// Every ownerless row is collected and reported together,
				// in table order, rather than failing on the first
				// (D-14-24) -- the join below is what makes the report
				// deterministic across runs.
				ownerless = append(ownerless, fmt.Sprintf("row %s has owning-phase cell %q outside the closed P<NN>|CLOSED(<sha>)|UNOWNED(<witness>) vocabulary", identifier, raw))
			}
		}
		if !gradeWitnessExempt {
			problems = append(problems, debtRegisterGradeWitnessRowProblems(name, identifier, row[columns["Grade"]], row[columns["Witness"]])...)
		}
	}
	if len(ownerless) > 0 {
		problems = append(problems, fmt.Sprintf("%s: %d row(s) failed the owning-phase vocabulary: %s", name, len(ownerless), strings.Join(ownerless, "; ")))
	}

	sections := debtRegisterDetailSections(text)
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
			problems = append(problems, fmt.Sprintf("%s: detail section %q has no row in the Items table", name, identifier))
			continue
		}
		details[identifier] = true
	}
	for identifier := range identifiers {
		if !details[identifier] {
			problems = append(problems, fmt.Sprintf("%s: row %s has no `### %s` detail section", name, identifier, identifier))
		}
	}

	// D-14-31: each row's detail section carries a first-recorded:
	// milestone. The two-milestone-carry ratification requirement is
	// scoped to the LIVE phase tree only (.planning/phases/**, never
	// .planning/milestones/**): an archived register from a shipped
	// milestone is frozen historical record, not an active re-deferral --
	// requiring a ratification line on it would mean fabricating one that
	// never happened. A row that is genuinely still being carried forward
	// reappears, unowned, in a LIVE register instead (the existing
	// "resolves"/"REVERSES"/"CARRIED" cross-reference convention this
	// migration followed), which is exactly where this guard is live.
	archived := strings.Contains(filepath.ToSlash(path), "/milestones/")
	for identifier := range identifiers {
		section, ok := sections[identifier]
		if !ok {
			continue // already reported above as a missing detail section
		}
		m := debtRegisterFirstRecordedPattern.FindStringSubmatch(section)
		if m == nil {
			problems = append(problems, fmt.Sprintf("%s: row %s's detail section has no `first-recorded: M0NN` line", name, identifier))
			continue
		}
		if archived {
			continue
		}
		// Live-tree carry check: only an UNOWNED row can be "carried" at
		// all (a P<NN> or CLOSED row is owned or resolved, not deferred).
		rowIdx := -1
		for i, row := range rows {
			if row[columns["ID"]] == identifier {
				rowIdx = i
				break
			}
		}
		if rowIdx == -1 || landingExempt {
			continue
		}
		cell := strings.TrimSpace(rows[rowIdx][columns["Landing phase"]])
		if !strings.HasPrefix(cell, "UNOWNED(") {
			continue
		}
		firstMilestone := 0
		fmt.Sscanf(m[1], "M%d", &firstMilestone)
		currentMilestone := 3 // M003, this milestone -- bump when M004 opens
		if firstMilestone > 0 && currentMilestone-firstMilestone >= 2 {
			ratified := strings.Contains(section, "blocking-human") && strings.Contains(strings.ToLower(section), "ratif")
			if !ratified {
				problems = append(problems, fmt.Sprintf("%s: row %s was first recorded in %s and is still UNOWNED in the live tree, carried past a second milestone boundary with no recorded blocking-human ratification (D-14-31)", name, identifier, m[1]))
			}
		}
	}

	return problems, nil
}

// debtRegisterDetailSections splits text's `## Detail` (or any) section into
// per-identifier substrings, keyed by the bare `D-XX-NN` identifier, each
// running from its `### D-XX-NN` heading to the next `### ` or `## `
// heading (or EOF). Used to scope the first-recorded: search to the correct
// row's own section rather than matching anywhere in the file.
func debtRegisterDetailSections(text string) map[string]string {
	sections := make(map[string]string)
	lines := strings.Split(text, "\n")
	var currentID string
	var buf []string
	flush := func() {
		if currentID != "" {
			sections[currentID] = strings.Join(buf, "\n")
		}
		buf = nil
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "### D-") {
			flush()
			heading := strings.TrimSpace(strings.TrimPrefix(line, "###"))
			identifier := heading
			if cut := strings.IndexAny(heading, " \t"); cut != -1 {
				identifier = heading[:cut]
			}
			currentID = strings.Trim(identifier, "`")
			continue
		}
		if strings.HasPrefix(line, "## ") {
			flush()
			currentID = ""
			continue
		}
		if currentID != "" {
			buf = append(buf, line)
		}
	}
	flush()
	return sections
}

// debtRegisterTable returns the Items table's column index by header name and
// its data rows, each row indexed the same way. Only the `## Items` section
// is read: later sections (closures, process debt) hold their own tables and
// are not the register. Delegates to parseDebtRegisterTable so
// debtRegisterProblems shares exactly one parser with this *testing.T-based
// caller (session_peer_gate_test.go's peerDivergenceDebtIDResolvesToOpenRow).
func debtRegisterTable(t *testing.T, name, text string) (map[string]int, [][]string) {
	t.Helper()
	columns, rows, err := parseDebtRegisterTable(name, text)
	if err != nil {
		t.Fatal(err)
	}
	return columns, rows
}

// parseDebtRegisterTable is debtRegisterTable's pure core. A zero-row Items
// table is a valid, non-error result (D-14-24's vacuous-pass requirement) --
// callers that need at least one row check len(rows) themselves.
func parseDebtRegisterTable(name, text string) (map[string]int, [][]string, error) {
	anchor := "\n## Items\n"
	start := strings.Index(text, anchor)
	if start == -1 {
		return nil, nil, fmt.Errorf("%s: has no `## Items` section", name)
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
			return nil, nil, fmt.Errorf("%s: row %q has %d cells, want %d", name, cells[0], len(cells), len(columns))
		}
		rows = append(rows, cells)
	}
	if columns == nil {
		return nil, nil, fmt.Errorf("%s: `## Items` section holds no table", name)
	}
	return columns, rows, nil
}

// debtRegisterOwnershipFixture copies a real, currently well-formed debt
// register (PHASE-14-DEBT.md) into t.TempDir(), so each seeded fault below
// is attributable to exactly one deliberate mutation rather than to any
// pre-existing defect in the source file. git status --porcelain over
// .planning never sees these copies: they live entirely under t.TempDir().
func debtRegisterOwnershipFixture(t *testing.T) string {
	t.Helper()
	registers, err := phaseArtifactGlob("14-evidence-instrument-and-honest-scoping", "PHASE-14-DEBT.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(registers) != 1 {
		t.Fatalf("expected exactly one PHASE-14-DEBT.md, found %d: %v", len(registers), registers)
	}
	data, err := os.ReadFile(registers[0])
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), "PHASE-14-DEBT.md")
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return dst
}

// debtRegisterFixtureLandingCell is the exact Landing phase cell value
// debtRegisterOwnershipFixture's copy carries for row D-14-46 today. Seeded
// faults below rewrite exactly this cell and nothing else, so a fault is
// attributable to that one row.
const debtRegisterFixtureLandingCell = "UNOWNED(conditional-surface-lands-in-language)"

// debtRegisterSeedLandingPhaseFault rewrites row D-14-46's Landing phase
// cell in the Items table (and only the Items table cell -- not the Detail
// section's own prose mention of the same string) to newCell, following
// TestInjectorMarkerCountGuardIsNotInert's temp-copy-and-seed-one-fault
// shape (session_phase6_injectors_test.go).
func debtRegisterSeedLandingPhaseFault(t *testing.T, path, newCell string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	needle := "| " + debtRegisterFixtureLandingCell + " |"
	if !strings.Contains(text, needle) {
		t.Fatalf("fixture does not contain the expected D-14-46 Items-table cell %q -- fixture drifted from the seam this test seeds", needle)
	}
	text = strings.Replace(text, needle, "| "+newCell+" |", 1)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// debtRegisterEmptyTableFixture builds a well-formed, zero-row variant of
// the real register: the frontmatter items: count is adjusted to 0, every
// Items table data row is removed (header and separator kept), and the
// Detail section is removed entirely (a zero-row table has no identifiers,
// so no `### D-ID` section could match one). This is the "the gate looked
// at nothing and found nothing" control D-14-24's vacuous-pass requirement
// names -- it must be indistinguishable, in result, from a genuinely empty
// register, never merely a register whose rows happen to all pass.
func debtRegisterEmptyTableFixture(t *testing.T) string {
	t.Helper()
	src := debtRegisterOwnershipFixture(t)
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)

	// Derive the live "items: N" line rather than hardcoding N: this
	// fixture copies the real, currently-live PHASE-14-DEBT.md (via
	// debtRegisterOwnershipFixture), and that register's own item count
	// grows over time (D-14-01..D-14-54 and beyond) -- a hardcoded prior
	// count silently stops matching and this replace becomes a no-op,
	// which is exactly the kind of drift no test in this file should be
	// able to hide.
	itemsLinePattern := regexp.MustCompile(`(?m)^items:\s*\d+\s*$`)
	if !itemsLinePattern.MatchString(text) {
		t.Fatal("fixture has no `items: N` frontmatter line to zero out")
	}
	text = itemsLinePattern.ReplaceAllString(text, "items: 0")

	itemsAnchor := "\n## Items\n"
	itemsStart := strings.Index(text, itemsAnchor)
	if itemsStart == -1 {
		t.Fatal("fixture has no `## Items` section")
	}
	afterItems := text[itemsStart+len(itemsAnchor):]
	detailAnchor := "\n## Detail\n"
	detailStart := strings.Index(afterItems, detailAnchor)
	if detailStart == -1 {
		t.Fatal("fixture has no `## Detail` section")
	}
	itemsSection := afterItems[:detailStart]
	// Keep only the header and separator lines (the first two `|`-prefixed
	// lines); drop every data row.
	var kept []string
	tableLines := 0
	for _, line := range strings.Split(itemsSection, "\n") {
		if strings.HasPrefix(line, "|") {
			tableLines++
			if tableLines > 2 {
				continue
			}
		}
		kept = append(kept, line)
	}
	newItemsSection := strings.Join(kept, "\n")

	rest := afterItems[detailStart:]
	// Drop everything from `## Detail` up to (not including) the next
	// top-level `## ` heading, or EOF if none -- zero rows means zero
	// identifiers, so zero detail sections are required or permitted.
	afterDetailAnchor := rest[len(detailAnchor):]
	if next := strings.Index(afterDetailAnchor, "\n## "); next != -1 {
		rest = "\n## Detail\n\n" + afterDetailAnchor[next+1:]
	} else {
		rest = "\n## Detail\n"
	}

	text = text[:itemsStart+len(itemsAnchor)] + newItemsSection + rest
	dst := filepath.Join(filepath.Dir(src), "PHASE-99-DEBT.md")
	if err := os.WriteFile(dst, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return dst
}

// debtRegisterWitnessGrammarFixture builds a synthetic, otherwise
// well-formed register with exactly one row (D-99-01), Grade cell
// gradeCell and Witness cell witnessCell, in t.TempDir(). The filename
// (PHASE-97-DEBT.md) matches the naming convention and is not on either
// exemption list, so Grade/Witness enforcement is live against it.
func debtRegisterWitnessGrammarFixture(t *testing.T, gradeCell, witnessCell string) string {
	t.Helper()
	text := "---\n" +
		"phase: 97-synthetic\n" +
		"recorded: 2026-09-18\n" +
		"status: accepted\n" +
		"disposition: synthetic-fixture\n" +
		"items: 1\n" +
		"blocking: 0\n" +
		"---\n\n" +
		"# Synthetic Witness Grammar Fixture\n\n" +
		"## Items\n\n" +
		"| ID | Source | Threat/Req | Severity | Landing phase | Grade | Witness | Item |\n" +
		"|---|---|---|---|---|---|---|---|\n" +
		"| D-99-01 | synthetic | SYN-01 | warning | P17 | " + gradeCell + " | " + witnessCell + " | synthetic claim for TestDebtRegisterWitnessGrammarIsClosed |\n\n" +
		"## Detail\n\n" +
		"### D-99-01 -- synthetic claim\n\n" +
		"first-recorded: M003\n\n" +
		"Synthetic fixture body.\n"
	dst := filepath.Join(t.TempDir(), "PHASE-97-DEBT.md")
	if err := os.WriteFile(dst, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return dst
}

// TestDebtRegisterWitnessGrammarIsClosed is Task 1's own <behavior> cases,
// each run over synthetic register text in t.TempDir() rather than the
// live corpus, so a fixture drift in the real registers can never mask a
// grammar regression. Written failing-first against the unextended law,
// then made to pass by the grammar this task adds (RED then GREEN, per
// this task's tdd="true").
func TestDebtRegisterWitnessGrammarIsClosed(t *testing.T) {
	t.Run("probe token naming an existing test passes", func(t *testing.T) {
		path := debtRegisterWitnessGrammarFixture(t, "EXERCISED", "probe:TestDebtRegistersAreWellFormed")
		if problems, err := debtRegisterProblems(path); err != nil || len(problems) != 0 {
			t.Fatalf("expected a clean pass; got err=%v problems=%v", err, problems)
		}
	})

	t.Run("probe token naming a nonexistent test fails", func(t *testing.T) {
		path := debtRegisterWitnessGrammarFixture(t, "EXERCISED", "probe:TestThisTestNameDoesNotExistAnywhereInTheModule")
		problems, err := debtRegisterProblems(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("a probe: token naming a nonexistent test should fail, but nothing was reported")
		}
	})

	t.Run("callsite token with correct symbol and count passes", func(t *testing.T) {
		path := debtRegisterWitnessGrammarFixture(t, "WIRED", "callsite:internal/compiler/check.resolveBlame=0")
		if problems, err := debtRegisterProblems(path); err != nil || len(problems) != 0 {
			t.Fatalf("expected a clean pass; got err=%v problems=%v", err, problems)
		}
	})

	t.Run("callsite token with wrong count fails", func(t *testing.T) {
		path := debtRegisterWitnessGrammarFixture(t, "WIRED", "callsite:internal/compiler/check.resolveBlame=99")
		problems, err := debtRegisterProblems(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("a callsite: token with the wrong count should fail, but nothing was reported")
		}
	})

	t.Run("callsite token naming a nonexistent symbol fails as missing, not as a count of zero", func(t *testing.T) {
		path := debtRegisterWitnessGrammarFixture(t, "WIRED", "callsite:internal/compiler/check.thisSymbolDoesNotExistAnywhere=0")
		problems, err := debtRegisterProblems(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("a callsite: token naming a nonexistent symbol should fail, but nothing was reported")
		}
		if !containsSubstring(problems, "does not exist in package") {
			t.Fatalf("refusal did not name the symbol as missing (absence must never read as a pass): %v", problems)
		}
	})

	t.Run("escape token resolving in the closed registry with a resolving probe passes", func(t *testing.T) {
		path := debtRegisterWitnessGrammarFixture(t, "DEFINED", "escape:callback-invocation-unsubjected")
		if problems, err := debtRegisterProblems(path); err != nil || len(problems) != 0 {
			t.Fatalf("expected a clean pass; got err=%v problems=%v", err, problems)
		}
	})

	t.Run("escape token absent from the closed registry fails", func(t *testing.T) {
		path := debtRegisterWitnessGrammarFixture(t, "DEFINED", "escape:this-escape-id-is-not-registered")
		problems, err := debtRegisterProblems(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("an unregistered escape: token should fail, but nothing was reported")
		}
	})

	t.Run("env token in the closed set is declared, never closes, and is explicitly not debt", func(t *testing.T) {
		path := debtRegisterWitnessGrammarFixture(t, "DEFINED (withdrawn)", "env:clang")
		if problems, err := debtRegisterProblems(path); err != nil || len(problems) != 0 {
			t.Fatalf("expected a clean pass; got err=%v problems=%v", err, problems)
		}
	})

	t.Run("env token outside the closed set fails", func(t *testing.T) {
		path := debtRegisterWitnessGrammarFixture(t, "DEFINED (withdrawn)", "env:some-invented-toolchain")
		problems, err := debtRegisterProblems(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("an env: token outside the closed set should fail, but nothing was reported")
		}
	})

	t.Run("free text fails even though non-empty", func(t *testing.T) {
		path := debtRegisterWitnessGrammarFixture(t, "WIRED", "OPEN and UNOWNED -- reopens only when the language grows")
		problems, err := debtRegisterProblems(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("free text must never satisfy the witness grammar merely by being non-empty")
		}
	})

	t.Run("a bare phase identifier is never a witness", func(t *testing.T) {
		path := debtRegisterWitnessGrammarFixture(t, "WIRED", "P17")
		problems, err := debtRegisterProblems(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("a bare phase identifier must never satisfy the witness grammar")
		}
	})

	t.Run("several comma-separated witnesses pass only when all hold", func(t *testing.T) {
		allHold := debtRegisterWitnessGrammarFixture(t, "WIRED", "probe:TestDebtRegistersAreWellFormed, callsite:internal/compiler/check.resolveBlame=0")
		if problems, err := debtRegisterProblems(allHold); err != nil || len(problems) != 0 {
			t.Fatalf("expected a clean pass when all comma-separated witnesses hold; got err=%v problems=%v", err, problems)
		}

		oneFails := debtRegisterWitnessGrammarFixture(t, "WIRED", "probe:TestDebtRegistersAreWellFormed, callsite:internal/compiler/check.resolveBlame=99")
		problems, err := debtRegisterProblems(oneFails)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("one failing token among several comma-separated witnesses must fail the whole cell")
		}
	})

	t.Run("a row graded below the satisfying bar with an empty witness cell fails unless withdrawn", func(t *testing.T) {
		emptyWitness := debtRegisterWitnessGrammarFixture(t, "WIRED", "n/a")
		problems, err := debtRegisterProblems(emptyWitness)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("a below-bar row with Witness \"n/a\" and no withdrawal should fail")
		}

		withdrawn := debtRegisterWitnessGrammarFixture(t, "DEFINED (withdrawn)", "n/a")
		if problems, err := debtRegisterProblems(withdrawn); err != nil || len(problems) != 0 {
			t.Fatalf("an explicitly withdrawn row should pass with no witness; got err=%v problems=%v", err, problems)
		}
	})

	t.Run("a row graded at or above the satisfying bar may declare Witness n/a", func(t *testing.T) {
		path := debtRegisterWitnessGrammarFixture(t, "MUTATION-KILLED", "n/a")
		if problems, err := debtRegisterProblems(path); err != nil || len(problems) != 0 {
			t.Fatalf("expected a clean pass; got err=%v problems=%v", err, problems)
		}
	})

	t.Run("DEFINED is the no-claim floor and needs no witness even when not withdrawn", func(t *testing.T) {
		path := debtRegisterWitnessGrammarFixture(t, "DEFINED", "n/a")
		if problems, err := debtRegisterProblems(path); err != nil || len(problems) != 0 {
			t.Fatalf("expected a clean pass; got err=%v problems=%v", err, problems)
		}
	})
}

// TestDebtRegisterOwnershipGuardIsNotInert demonstrates concretely that the
// closed owning-phase vocabulary (D-14-24) is not merely wired but actually
// exercised: it seeds one fault per mechanizable kind (emptied cell,
// whitespace-only cell, free-prose cell) into a temp copy of a real
// register and asserts debtRegisterProblems -- the reports-rather-than-
// fatals variant checkDebtRegister wraps -- refuses each one, names the
// offending row, and reports nothing for the unmodified control. The
// emptied-Items-table control passes vacuously, distinguishing "the gate
// found nothing" from "the gate looked at nothing" (T-14-23). A guard
// proven red on only one of the three fault kinds would be inert for the
// other two, so all three seeds ship together.
func TestDebtRegisterOwnershipGuardIsNotInert(t *testing.T) {
	t.Run("unmodified copy passes", func(t *testing.T) {
		path := debtRegisterOwnershipFixture(t)
		problems, err := debtRegisterProblems(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) != 0 {
			t.Fatalf("unmodified copy should pass; got problems: %v", problems)
		}
	})

	t.Run("emptied cell fails", func(t *testing.T) {
		path := debtRegisterOwnershipFixture(t)
		debtRegisterSeedLandingPhaseFault(t, path, "")
		problems, err := debtRegisterProblems(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("emptied Landing phase cell should refuse, but debtRegisterProblems reported no problems")
		}
		if !containsSubstring(problems, "D-14-46") {
			t.Fatalf("refusal does not name the offending row D-14-46: %v", problems)
		}
	})

	t.Run("whitespace-only cell fails identically", func(t *testing.T) {
		path := debtRegisterOwnershipFixture(t)
		debtRegisterSeedLandingPhaseFault(t, path, "   ")
		problems, err := debtRegisterProblems(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("whitespace-only Landing phase cell should refuse, but debtRegisterProblems reported no problems")
		}
		if !containsSubstring(problems, "D-14-46") {
			t.Fatalf("refusal does not name the offending row D-14-46: %v", problems)
		}
	})

	t.Run("free-prose cell fails as out-of-vocabulary", func(t *testing.T) {
		path := debtRegisterOwnershipFixture(t)
		debtRegisterSeedLandingPhaseFault(t, path, "Reopen when the language gains conditionals")
		problems, err := debtRegisterProblems(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("free-prose Landing phase cell should refuse, but debtRegisterProblems reported no problems")
		}
		if !containsSubstring(problems, "D-14-46") {
			t.Fatalf("refusal does not name the offending row D-14-46: %v", problems)
		}
	})

	t.Run("emptied Items table passes vacuously", func(t *testing.T) {
		path := debtRegisterEmptyTableFixture(t)
		problems, err := debtRegisterProblems(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) != 0 {
			t.Fatalf("a zero-row Items table with a matching items: 0 frontmatter should pass vacuously; got problems: %v", problems)
		}
	})

	t.Run("every seeded copy lived under t.TempDir()", func(t *testing.T) {
		// Every helper above writes exclusively through filepath.Join with
		// t.TempDir() (directly, or via debtRegisterOwnershipFixture) --
		// asserted here as a structural invariant rather than re-deriving
		// it via a subprocess `git status` call (which this project's own
		// TestSourceNeverSpawnsUnboundedProcesses guard, D-02-01, watches).
		// No test above ever constructs a path under testsupport.ProjectPath's
		// live .planning tree, so git status --porcelain .planning is empty
		// by construction, not by a post-hoc check.
		planningRoot := testsupport.ProjectPath(".planning")
		for _, p := range []string{
			debtRegisterOwnershipFixture(t),
			debtRegisterEmptyTableFixture(t),
		} {
			if strings.HasPrefix(p, planningRoot) {
				t.Fatalf("fixture path %q leaked into the live .planning tree", p)
			}
		}
	})
}

// containsSubstring reports whether any element of problems contains sub.
func containsSubstring(problems []string, sub string) bool {
	for _, p := range problems {
		if strings.Contains(p, sub) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------
// Plan 14-10 Task 1: the reconciliation verdict vocabulary (D-14-12).
//
// A dead command in an archived evidence document is never rewritten in
// place -- that would be exactly the falsification this phase exists to
// prevent. The correction lives OUTSIDE the archive, in this phase's own
// PHASE-14-DEBT.md, as a debt row with a witness: a "```reconciliation"
// fenced block inside the row's own ### Detail section, keyed by
// (file, line, verbatim original command). This is deliberately NOT a
// third authored register (D-14-13) -- the ledger's content is right, its
// authorship moves into the existing register law.
//
// There is NO inline suppression syntax anywhere in this mechanism: no
// comment-shaped directive, no per-file ignore, no count-keyed ceiling.
// Every entry is satisfied only by a claim that can itself fail. Refusing
// a count-keyed ceiling is deliberate: a count-neutral swap (one dead
// pattern traded for another, leaving the total unchanged) would land
// unnoticed under a ceiling -- the documented ESLint-bulk-suppressions
// failure mode (D-14-12).
// ---------------------------------------------------------------------

// reconciliationVerdict is the closed four-member vocabulary. A verdict
// outside this set fails.
type reconciliationVerdict string

const (
	reconciliationRenamed          reconciliationVerdict = "renamed"
	reconciliationSuperseded       reconciliationVerdict = "superseded"
	reconciliationObsoleteByDesign reconciliationVerdict = "obsolete-by-design"
	reconciliationUnderScoped      reconciliationVerdict = "under-scoped"
)

// reconciliationVerdicts is the closed vocabulary, mirroring
// debtRegisterSeverities' own shape.
var reconciliationVerdicts = map[reconciliationVerdict]bool{
	reconciliationRenamed:          true,
	reconciliationSuperseded:       true,
	reconciliationObsoleteByDesign: true,
	reconciliationUnderScoped:      true,
}

// reconciliationEntry is one parsed ```reconciliation block. Every finding
// carries File/Line/Command/Verdict/Classification; the remaining fields
// are populated per-verdict only (D-14-12's four distinct obligations).
type reconciliationEntry struct {
	ID string // the enclosing "### D-14-NN" heading

	File           string
	Line           int
	Command        string
	Classification classification
	Verdict        reconciliationVerdict

	// renamed
	Replacement string

	// superseded
	SupersedingPhase  string
	SupersedingCommit string
	CoveringCommand   string

	// obsolete-by-design
	DeletedPackage string
	DeletedSymbol  string
	DeletingPhase  string
	DeletingCommit string

	// under-scoped (permitted only for R2b findings, D-14-12)
	MissingClause string
	LandingPhase  string
}

var (
	reconciliationBlockPattern   = regexp.MustCompile("(?s)```reconciliation\n(.*?)\n```")
	reconciliationHeadingPattern = regexp.MustCompile(`(?m)^### (D-14-\d+)`)
	reconciliationFieldPattern   = regexp.MustCompile(`^([a-z][a-z-]*):\s*(.*)$`)
)

// parseReconciliationEntries scans a *-DEBT.md register's Detail section
// for "```reconciliation" fenced blocks (never the Items table itself,
// which cannot host raw command text -- many verification commands
// contain an unescaped "|", and parseDebtRegisterTable's row splitter has
// no escaping discipline). Each block's ID is the nearest preceding
// "### D-14-NN" heading.
func parseReconciliationEntries(path string) ([]reconciliationEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := string(data)

	headings := reconciliationHeadingPattern.FindAllStringSubmatchIndex(text, -1)

	var entries []reconciliationEntry
	for _, loc := range reconciliationBlockPattern.FindAllStringSubmatchIndex(text, -1) {
		blockStart, blockContentStart, blockContentEnd := loc[0], loc[2], loc[3]
		block := text[blockContentStart:blockContentEnd]

		id := ""
		for _, h := range headings {
			if h[0] < blockStart {
				id = text[h[2]:h[3]]
			} else {
				break
			}
		}

		fields := make(map[string]string)
		for _, line := range strings.Split(block, "\n") {
			m := reconciliationFieldPattern.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			fields[m[1]] = m[2]
		}

		lineNum := 0
		if raw, ok := fields["line"]; ok {
			n, convErr := strconv.Atoi(strings.TrimSpace(raw))
			if convErr != nil {
				return nil, fmt.Errorf("%s: entry %s has non-numeric line %q", path, id, raw)
			}
			lineNum = n
		}

		entries = append(entries, reconciliationEntry{
			ID: id,

			File:           fields["file"],
			Line:           lineNum,
			Command:        fields["command"],
			Classification: classification(fields["classification"]),
			Verdict:        reconciliationVerdict(fields["verdict"]),

			Replacement: fields["replacement"],

			SupersedingPhase:  fields["superseding-phase"],
			SupersedingCommit: fields["superseding-commit"],
			CoveringCommand:   fields["covering-command"],

			DeletedPackage: fields["deleted-package"],
			DeletedSymbol:  fields["deleted-symbol"],
			DeletingPhase:  fields["deleting-phase"],
			DeletingCommit: fields["deleting-commit"],

			MissingClause: fields["missing-clause"],
			LandingPhase:  fields["landing-phase"],
		})
	}
	return entries, nil
}

// reconciliationGitTimeout bounds every git subprocess this file spawns
// (TestSourceNeverSpawnsUnboundedProcesses, D-02-01).
const reconciliationGitTimeout = 30 * time.Second

// reconciliationCommitExists confirms a commit SHA genuinely resolves in
// this repository's history -- a superseding/deleting commit that does not
// exist is a fabricated citation, not evidence.
func reconciliationCommitExists(sha string) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), reconciliationGitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "cat-file", "-e", sha+"^{commit}")
	cmd.Dir = testsupport.ProjectPath()
	var stderr groundednessBoundedWriter
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, fmt.Sprintf("commit %q not found in this repository's history: %v (%s)", sha, err, stderr.bytes())
	}
	return true, ""
}

// reconciliationSymbolAbsentFromTree is the obsolete-by-design obligation's
// absence assertion (D-14-12, key_links): the SAME anti-source-search rule
// as the callsite: witness grammar (D-14-23) -- a claim that a symbol is
// gone is falsifiable via an AST declaration scan, never via a textual
// grep returning zero matches (a claim that a search found nothing is not
// falsifiable). Reuses debtRegisterCallSiteWitnessMatches directly rather
// than writing a second AST walker.
func reconciliationSymbolAbsentFromTree(pkgPath, symbol string) (bool, string) {
	ok, msg := debtRegisterCallSiteWitnessMatches(pkgPath, symbol, 0)
	if ok {
		// declared==true, count==0: the symbol still exists (with zero
		// call sites), which is NOT the same claim as "deleted".
		return false, fmt.Sprintf("obsolete-by-design symbol %q still exists (declared) in package %q -- a deleted symbol must be entirely absent, not merely uncalled", symbol, pkgPath)
	}
	if strings.Contains(msg, "does not exist in package") {
		return true, ""
	}
	return false, msg
}

// reconciliationEntryProblem checks ONE entry's verdict obligation
// (D-14-12) and returns a non-empty problem description on failure. Every
// "resolves" obligation reuses classifyCommand -- the SAME primitive the
// groundedness lint's own classifier uses -- so "resolves" means one thing
// project-wide: the command classifies classOK (never R1/R2/R2b/R3/
// unparseable).
func reconciliationEntryProblem(index *testIndex, e reconciliationEntry) string {
	if e.ID == "" {
		return "reconciliation block has no enclosing \"### D-14-NN\" heading"
	}
	if e.File == "" || e.Line == 0 || e.Command == "" {
		return fmt.Sprintf("%s: reconciliation entry missing file, line, or command", e.ID)
	}
	if !reconciliationVerdicts[e.Verdict] {
		return fmt.Sprintf("%s: verdict %q is outside the closed four-member vocabulary (renamed, superseded, obsolete-by-design, under-scoped)", e.ID, e.Verdict)
	}

	switch e.Verdict {
	case reconciliationRenamed:
		if e.Replacement == "" {
			return fmt.Sprintf("%s: renamed entry names no replacement command", e.ID)
		}
		if class := classifyCommand(index, e.Replacement); class != classOK {
			return fmt.Sprintf("%s: renamed entry's replacement command does not resolve (classified %s): %s", e.ID, class, e.Replacement)
		}
	case reconciliationSuperseded:
		if e.SupersedingPhase == "" || e.SupersedingCommit == "" || e.CoveringCommand == "" {
			return fmt.Sprintf("%s: superseded entry missing superseding phase, commit, or covering command", e.ID)
		}
		if ok, msg := reconciliationCommitExists(e.SupersedingCommit); !ok {
			return fmt.Sprintf("%s: superseding commit %s: %s", e.ID, e.SupersedingCommit, msg)
		}
		if class := classifyCommand(index, e.CoveringCommand); class != classOK {
			return fmt.Sprintf("%s: superseded entry's covering command does not resolve (classified %s): %s", e.ID, class, e.CoveringCommand)
		}
	case reconciliationObsoleteByDesign:
		if e.DeletedPackage == "" || e.DeletedSymbol == "" || e.DeletingPhase == "" || e.DeletingCommit == "" {
			return fmt.Sprintf("%s: obsolete-by-design entry missing deleted package, deleted symbol, deleting phase, or commit", e.ID)
		}
		if ok, msg := reconciliationCommitExists(e.DeletingCommit); !ok {
			return fmt.Sprintf("%s: deleting commit %s: %s", e.ID, e.DeletingCommit, msg)
		}
		if ok, msg := reconciliationSymbolAbsentFromTree(e.DeletedPackage, e.DeletedSymbol); !ok {
			return fmt.Sprintf("%s: %s", e.ID, msg)
		}
	case reconciliationUnderScoped:
		if e.Classification != classR2b {
			return fmt.Sprintf("%s: under-scoped verdict is permitted only for per-branch (R2b) findings, not %s", e.ID, e.Classification)
		}
		if e.MissingClause == "" || e.LandingPhase == "" {
			return fmt.Sprintf("%s: under-scoped entry missing clause or landing phase", e.ID)
		}
		if !debtRegisterOwningPhaseForm(e.LandingPhase) {
			return fmt.Sprintf("%s: under-scoped entry's landing phase %q is not one of the closed forms (P<NN>, CLOSED(<sha>), UNOWNED(<witness-id>))", e.ID, e.LandingPhase)
		}
	}
	return ""
}

// TestReconciliationVerdictsCarryTheirObligations is Task 1's own
// non-inertness proof: the closed vocabulary, its four checked
// obligations (each with a passing positive and passing negative
// subtest), and the count/staleness cross-check against a live corpus
// scan -- never against the shrinking pinnedFrontier literal, so this
// stays correct after Task 3 empties it.
func TestReconciliationVerdictsCarryTheirObligations(t *testing.T) {
	index := buildTestIndex(t)

	t.Run("verdict outside the closed vocabulary fails, naming the offender", func(t *testing.T) {
		bad := reconciliationEntry{ID: "D-00-00", File: "x", Line: 1, Command: "y", Verdict: "deprecated"}
		problem := reconciliationEntryProblem(index, bad)
		if problem == "" {
			t.Fatal("a verdict outside the closed vocabulary should fail, but nothing was reported")
		}
		if !strings.Contains(problem, "deprecated") {
			t.Fatalf("problem should name the offending verdict, got: %s", problem)
		}
	})

	t.Run("renamed", func(t *testing.T) {
		t.Run("replacement resolving to a real test passes", func(t *testing.T) {
			e := reconciliationEntry{ID: "D-00-01", File: "x", Line: 1, Command: "y", Verdict: reconciliationRenamed,
				Replacement: "go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned"}
			if problem := reconciliationEntryProblem(index, e); problem != "" {
				t.Fatalf("expected pass, got: %s", problem)
			}
		})
		t.Run("replacement resolving to nothing fails -- cannot launder a dead pattern into another dead pattern", func(t *testing.T) {
			e := reconciliationEntry{ID: "D-00-02", File: "x", Line: 1, Command: "y", Verdict: reconciliationRenamed,
				Replacement: "go test ./internal/compiler/session/... -run TestThisNameDoesNotExistAnywhereInTheModule"}
			problem := reconciliationEntryProblem(index, e)
			if problem == "" {
				t.Fatal("a replacement resolving to nothing should fail, but nothing was reported")
			}
		})
	})

	t.Run("superseded", func(t *testing.T) {
		t.Run("real commit and resolving covering command passes", func(t *testing.T) {
			e := reconciliationEntry{ID: "D-00-03", File: "x", Line: 1, Command: "y", Verdict: reconciliationSuperseded,
				SupersedingPhase: "P09", SupersedingCommit: "b8fe3df",
				CoveringCommand: "go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned"}
			if problem := reconciliationEntryProblem(index, e); problem != "" {
				t.Fatalf("expected pass, got: %s", problem)
			}
		})
		t.Run("nonexistent commit fails", func(t *testing.T) {
			e := reconciliationEntry{ID: "D-00-04", File: "x", Line: 1, Command: "y", Verdict: reconciliationSuperseded,
				SupersedingPhase: "P09", SupersedingCommit: "0000000000000000000000000000000000dead",
				CoveringCommand: "go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned"}
			if problem := reconciliationEntryProblem(index, e); problem == "" {
				t.Fatal("a nonexistent superseding commit should fail, but nothing was reported")
			}
		})
		t.Run("covering command resolving to nothing fails", func(t *testing.T) {
			e := reconciliationEntry{ID: "D-00-05", File: "x", Line: 1, Command: "y", Verdict: reconciliationSuperseded,
				SupersedingPhase: "P09", SupersedingCommit: "b8fe3df",
				CoveringCommand: "go test ./internal/compiler/session/... -run TestThisNameDoesNotExistAnywhereInTheModule"}
			if problem := reconciliationEntryProblem(index, e); problem == "" {
				t.Fatal("a covering command resolving to nothing should fail, but nothing was reported")
			}
		})
	})

	t.Run("obsolete-by-design", func(t *testing.T) {
		t.Run("real deleting commit and genuinely absent symbol passes", func(t *testing.T) {
			e := reconciliationEntry{ID: "D-00-06", File: "x", Line: 1, Command: "y", Verdict: reconciliationObsoleteByDesign,
				DeletedPackage: "internal/compiler/check", DeletedSymbol: "computeLoanLastUses",
				DeletingPhase: "P09-09", DeletingCommit: "b8fe3df"}
			if problem := reconciliationEntryProblem(index, e); problem != "" {
				t.Fatalf("expected pass, got: %s", problem)
			}
		})
		t.Run("symbol that still exists fails -- absence is never a pass", func(t *testing.T) {
			e := reconciliationEntry{ID: "D-00-07", File: "x", Line: 1, Command: "y", Verdict: reconciliationObsoleteByDesign,
				DeletedPackage: "internal/compiler/check", DeletedSymbol: "resolveBlame",
				DeletingPhase: "P09-09", DeletingCommit: "b8fe3df"}
			problem := reconciliationEntryProblem(index, e)
			if problem == "" {
				t.Fatal("a symbol that still exists should fail, but nothing was reported")
			}
			if !strings.Contains(problem, "still exists") {
				t.Fatalf("problem should say the symbol still exists, got: %s", problem)
			}
		})
		t.Run("nonexistent deleting commit fails", func(t *testing.T) {
			e := reconciliationEntry{ID: "D-00-08", File: "x", Line: 1, Command: "y", Verdict: reconciliationObsoleteByDesign,
				DeletedPackage: "internal/compiler/check", DeletedSymbol: "computeLoanLastUses",
				DeletingPhase: "P09-09", DeletingCommit: "0000000000000000000000000000000000dead"}
			if problem := reconciliationEntryProblem(index, e); problem == "" {
				t.Fatal("a nonexistent deleting commit should fail, but nothing was reported")
			}
		})
	})

	t.Run("under-scoped", func(t *testing.T) {
		t.Run("R2b finding with a clause and a valid landing phase passes", func(t *testing.T) {
			e := reconciliationEntry{ID: "D-00-09", File: "x", Line: 1, Command: "y", Classification: classR2b, Verdict: reconciliationUnderScoped,
				MissingClause: "specific failing branch never repointed", LandingPhase: "P20"}
			if problem := reconciliationEntryProblem(index, e); problem != "" {
				t.Fatalf("expected pass, got: %s", problem)
			}
		})
		t.Run("under-scoped on a non-R2b finding fails -- permitted only for per-branch findings", func(t *testing.T) {
			e := reconciliationEntry{ID: "D-00-10", File: "x", Line: 1, Command: "y", Classification: classR2, Verdict: reconciliationUnderScoped,
				MissingClause: "x", LandingPhase: "P20"}
			problem := reconciliationEntryProblem(index, e)
			if problem == "" {
				t.Fatal("an under-scoped verdict on a non-R2b finding should fail, but nothing was reported")
			}
			if !strings.Contains(problem, "per-branch") {
				t.Fatalf("problem should name the per-branch restriction, got: %s", problem)
			}
		})
	})

	// D-14-12: no inline suppression syntax and no count-keyed ceiling
	// anywhere in this mechanism. Deliberately verified EXTERNALLY (a
	// one-off shell grep during plan execution, recorded in the SUMMARY),
	// never as a Go subtest embedded in this same file -- naming the
	// banned tokens literally in a subtest here would make this file's
	// own descriptive comment about the check match the check itself.

	t.Run("every real reconciliation entry's obligation holds", func(t *testing.T) {
		registers, err := phaseArtifactGlob("14-evidence-instrument-and-honest-scoping", "PHASE-14-DEBT.md")
		if err != nil {
			t.Fatal(err)
		}
		if len(registers) != 1 {
			t.Fatalf("expected exactly one PHASE-14-DEBT.md, found %d: %v", len(registers), registers)
		}
		entries, err := parseReconciliationEntries(registers[0])
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) == 0 {
			t.Fatal("PHASE-14-DEBT.md carries zero reconciliation entries")
		}
		for _, e := range entries {
			if problem := reconciliationEntryProblem(index, e); problem != "" {
				t.Errorf("reconciliation entry failed: %s", problem)
			}
		}
	})

	t.Run("the reconciliation ledger exactly covers the live corpus's R1/R2/R3 findings", func(t *testing.T) {
		registers, err := phaseArtifactGlob("14-evidence-instrument-and-honest-scoping", "PHASE-14-DEBT.md")
		if err != nil {
			t.Fatal(err)
		}
		var entries []reconciliationEntry
		for _, path := range registers {
			parsed, parseErr := parseReconciliationEntries(path)
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			entries = append(entries, parsed...)
		}
		entrySet := make(map[violationRecord]bool, len(entries))
		for _, e := range entries {
			entrySet[violationRecord{File: e.File, Line: e.Line, Command: e.Command, Classification: e.Classification}] = true
		}

		measured := measuredViolations(t)
		var liveFindings int
		for _, v := range measured {
			if v.Classification != classR1 && v.Classification != classR2 && v.Classification != classR3 {
				continue
			}
			liveFindings++
			if !entrySet[v] {
				t.Errorf("live R1/R2/R3 finding has no reconciliation entry: %s %s:%d: %s", v.Classification, v.File, v.Line, v.Command)
			}
		}
		for key := range entrySet {
			found := false
			for _, v := range measured {
				if v == key {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("stale reconciliation entry: its (file, line, command) no longer appears in the live corpus scan: %s:%d: %s (never auto-pruned -- fix by removing the entry in a reviewed commit)", key.File, key.Line, key.Command)
			}
		}
		t.Logf("%d live R1/R2/R3 findings, %d reconciliation entries", liveFindings, len(entrySet))
	})
}

// TestPhase16EmitterCutsAreAmendedAndOwned is the anti-decay control for
// NAT-09/D-10-60. It deliberately uses parseDebtRegisterTable, the shared
// debt-register parser, so this semantic law cannot grow a second Markdown
// grammar that disagrees with TestDebtRegistersAreWellFormed.
func TestPhase16EmitterCutsAreAmendedAndOwned(t *testing.T) {
	requirementsPath := testsupport.ProjectPath(".planning", "milestones", "M003-REQUIREMENTS.md")
	requirements, err := os.ReadFile(requirementsPath)
	if err != nil {
		t.Fatal(err)
	}
	roadmap, err := os.ReadFile(testsupport.ProjectPath(".planning", "ROADMAP.md"))
	if err != nil {
		t.Fatal(err)
	}
	registers, err := phaseArtifactGlob("16-branch-match-emitter-port", "PHASE-16-DEBT.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(registers) != 1 {
		t.Fatalf("expected exactly one Phase 16 emitter debt register, found %d: %v", len(registers), registers)
	}
	debt, err := os.ReadFile(registers[0])
	if err != nil {
		t.Fatal(err)
	}
	if problems := phase16EmitterCutProblems(string(requirements), string(roadmap), string(debt)); len(problems) != 0 {
		t.Fatalf("Phase 16 NAT-09 amendment/debt mismatch:\n%s", strings.Join(problems, "\n"))
	}

	// The control changes exactly one family spelling in an otherwise-real
	// register. A passing control would prove this test only found a heading,
	// not the required amendment-to-row bijection.
	mutated := strings.Replace(string(debt), "emitLinearBorrowedByPointerPlain", "missingPointerFamily", 1)
	if problems := phase16EmitterCutProblems(string(requirements), string(roadmap), mutated); len(problems) == 0 {
		t.Fatal("seeded missing pointer family passed the NAT-09 amendment/debt control")
	}
	mutated = strings.Replace(string(debt), "| warning | P21 |", "| warning | UNOWNED(m004-native-emission-design) |", 1)
	if problems := phase16EmitterCutProblems(string(requirements), string(roadmap), mutated); len(problems) == 0 {
		t.Fatal("seeded unowned emitter cut passed the NAT-09 amendment/debt control")
	}
	mutatedRoadmap := strings.Replace(string(roadmap), "### Phase 21: Native Emission Ownership and Resource Discharge", "### Phase 21: Unrelated Phase", 1)
	if mutatedRoadmap == string(roadmap) {
		t.Fatal("seeded roadmap mutation did not change the registered Phase 21 heading")
	}
	if problems := phase16EmitterCutProblems(string(requirements), mutatedRoadmap, string(debt)); len(problems) == 0 {
		t.Fatal("seeded mistitled roadmap owner passed the NAT-09 amendment/debt control")
	}
	mutated = strings.Replace(string(debt), "checked resource ledger", "resource ledger", 1)
	if problems := phase16EmitterCutProblems(string(requirements), string(roadmap), mutated); len(problems) == 0 {
		t.Fatal("seeded missing family prerequisite passed the NAT-09 amendment/debt control")
	}
}

func phase16EmitterCutProblems(requirements, roadmap, debt string) []string {
	const amendment = "### NAT-09 — Phase 16 D-10-60 amendment (2026-09-19)"
	const ownerTitle = "Native Emission Ownership and Resource Discharge"
	families := []struct{ id, name string }{
		{"D-16-11", "emitLinearForeign"},
		{"D-16-12", "emitLinearBorrowedByPointer"},
		{"D-16-13", "emitLinearBorrowedByPointerPlain"},
	}
	var problems []string
	if !strings.Contains(requirements, amendment) || !strings.Contains(requirements, "cut-m004") || !strings.Contains(requirements, "PHASE-16-DEBT.md") {
		problems = append(problems, "NAT-09 lacks the dated Phase 16 D-10-60 cut-m004 amendment linked to PHASE-16-DEBT.md")
	}
	if !strings.Contains(requirements, "M004 Phase 21: "+ownerTitle) {
		problems = append(problems, "NAT-09 does not name the Phase 21 owner")
	}
	const shippedM003 = "✅ **M003 — Computation and Honest Instruments**"
	const provisionalM004 = "◷ **M004 — Native Emission Ownership and Resource Discharge**"
	m003At, m004At := strings.Index(roadmap, shippedM003), strings.Index(roadmap, provisionalM004)
	if m003At < 0 || m004At < 0 || m004At < m003At || !strings.Contains(roadmap, provisionalM004+" — provisional;") || !strings.Contains(roadmap, "### Phase 21: "+ownerTitle) {
		problems = append(problems, "ROADMAP.md does not place the named Phase 21 owner after M003")
	}
	columns, rows, err := parseDebtRegisterTable("PHASE-16-DEBT.md", debt)
	if err != nil {
		return append(problems, err.Error())
	}
	if len(rows) != len(families) {
		problems = append(problems, fmt.Sprintf("PHASE-16-DEBT.md has %d cut rows, want %d", len(rows), len(families)))
	}
	for _, column := range []string{"ID", "Item", "Landing phase"} {
		if _, ok := columns[column]; !ok {
			return append(problems, fmt.Sprintf("PHASE-16-DEBT.md lacks %q column", column))
		}
	}
	for _, family := range families {
		if !strings.Contains(requirements, family.name) {
			problems = append(problems, fmt.Sprintf("NAT-09 amendment omits %s", family.name))
		}
		found := false
		for _, row := range rows {
			if row[columns["ID"]] == family.id && strings.Contains(row[columns["Item"]], family.name) {
				found = true
				if row[columns["Landing phase"]] != "P21" {
					problems = append(problems, fmt.Sprintf("%s is not assigned to P21", family.id))
				}
			}
		}
		if !found {
			problems = append(problems, fmt.Sprintf("debt register has no bijective %s row for %s", family.id, family.name))
		}
		section := phase16DebtDetail(debt, family.id)
		for _, required := range []string{"M004 owner: Phase 21 — " + ownerTitle + ".", "Prerequisite:", "Reopening condition:", "Witness:", "`-flto` consequence:"} {
			if !strings.Contains(section, required) {
				problems = append(problems, fmt.Sprintf("%s detail omits %s", family.id, required))
			}
		}
		familyPrerequisites := map[string][]string{
			"D-16-11": {"checked resource ledger", "foreign-call blocks", "native witness demonstrating cleanup across every admitted foreign exit path"},
			"D-16-12": {"full discharge-pair design", "macOS and Linux evidence", "both hosts"},
			"D-16-13": {"shared-pointer alias and discharge contract", "family-specific ownership/alias proof"},
		}
		normalizedSection := strings.Join(strings.Fields(section), " ")
		for _, required := range familyPrerequisites[family.id] {
			if !strings.Contains(normalizedSection, required) {
				problems = append(problems, fmt.Sprintf("%s detail omits family prerequisite %q", family.id, required))
			}
		}
	}
	return problems
}

func phase16DebtDetail(debt, id string) string {
	anchor := "### " + id
	start := strings.Index(debt, anchor)
	if start == -1 {
		return ""
	}
	rest := debt[start+len(anchor):]
	if end := strings.Index(rest, "\n### "); end != -1 {
		return rest[:end]
	}
	return rest
}
