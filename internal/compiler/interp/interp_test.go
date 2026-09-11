package interp

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// interpProjectRoot mirrors nat03ProjectRoot's own technique
// (internal/compiler/session/session_phase5_alias.go): a
// runtime.Caller(0)-anchored resolution, avoiding a testsupport import
// (testsupport pulls in session, and session imports interp -- an interp
// test importing testsupport would be a cycle).
func interpProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

// checkedCallBasicProgram checks and corevalidates testdata/phase07/call_basic.lang
// directly (syntax.Parse + check.Program + corevalidate.Validate), never
// through package session -- session imports interp, so an interp test
// importing session would be a cycle.
func checkedCallBasicProgram(t *testing.T) core.Program {
	t.Helper()
	path := filepath.Join(interpProjectRoot(), "testdata", "phase07", "call_basic.lang")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("parse: unexpected diagnostics: %v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) > 0 {
		t.Fatalf("check: unexpected diagnostics: %v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("corevalidate rejected: %v", validated.Problems)
	}
	return validated.Program()
}

// moveAsCopyProbeProgram hand-builds a minimal two-function core.Program
// (never fed through the parser) whose caller re-reads its OWN call
// argument place immediately after the call. No LEGAL Lang source can
// express this: a non-copyable argument used a second time after a call is
// exactly what check's and corevalidate's OWN independent move-tracking
// both refuse (D-07-11's consume rule, re-derived independently on each
// side) -- so this probe is deliberately synthetic, mirroring the
// established in-repo precedent for exercising an engine directly against
// a forged core.Program (see callgraph's own T-07-35 doc comment). It is
// run via runFrameStack directly, bypassing Run's own corevalidate.Validate
// gate, which would otherwise refuse this shape identically regardless of
// moveAsCopyForTest -- exactly the point: the probe isolates interp's OWN
// mechanism from the two peers this test proves are blind to it.
func moveAsCopyProbeProgram() (program core.Program, caller core.Function) {
	nonCopyType := core.TypeFact{ID: "probe:type:buffer"} // no AbilityCopy: non-copyable (D-07-11)

	calleeParam := core.Parameter{ID: "probe:callee:place:param", Name: "value", Type: "Buffer"}
	callee := core.Function{
		ID: "probe:callee:fn", Name: "identity", Parameter: calleeParam, ReturnType: "Buffer",
		Linear: &core.LinearBody{
			ID:     "probe:callee:linear",
			Types:  []core.TypeFact{nonCopyType},
			Places: []core.Place{{ID: calleeParam.ID, Name: "value", TypeID: nonCopyType.ID}},
			Operations: []core.LinearOperation{
				{ID: "probe:callee:op:0", Kind: core.OpReturn, SourceID: calleeParam.ID, TypeID: nonCopyType.ID},
			},
		},
	}

	callerParam := core.Parameter{ID: "probe:caller:place:param", Name: "buffer", Type: "Buffer"}
	callTarget := "probe:caller:place:1"
	staleReadTarget := "probe:caller:place:2"
	caller = core.Function{
		ID: "probe:caller:fn", Name: "main", Parameter: callerParam, ReturnType: "Buffer",
		Linear: &core.LinearBody{
			ID:    "probe:caller:linear",
			Types: []core.TypeFact{nonCopyType},
			Places: []core.Place{
				{ID: callerParam.ID, Name: "buffer", TypeID: nonCopyType.ID},
				{ID: callTarget, Name: "result", TypeID: nonCopyType.ID},
				{ID: staleReadTarget, Name: "stale", TypeID: nonCopyType.ID},
			},
			Operations: []core.LinearOperation{
				// The call MOVES buffer (non-copyable): its own SourceID is
				// the caller's argument place.
				{ID: "probe:caller:op:0", Kind: core.OpCall, SourceID: callerParam.ID, TargetID: callTarget, TypeID: nonCopyType.ID, CalleeID: callee.ID},
				// The stale re-read: legal ONLY if the caller-side delete
				// was skipped (moveAsCopyForTest).
				{ID: "probe:caller:op:1", Kind: core.OpCopy, SourceID: callerParam.ID, TargetID: staleReadTarget, TypeID: nonCopyType.ID},
				{ID: "probe:caller:op:2", Kind: core.OpReturn, SourceID: staleReadTarget, TypeID: nonCopyType.ID},
			},
		},
	}

	return core.Program{Schema: core.Schema1, Module: "probe.move_as_copy", Functions: []core.Function{callee, caller}}, caller
}

// TestMoveAsCopyMutationKilled is Task 2's D-10-41 mutation test (QLT-08):
// partitionFrameForCall's caller-side delete of a non-copyable argument
// place is the entire mechanism realizing OWN-05b's ownership-transfer fact
// at a call boundary. With that delete suppressed via moveAsCopyForTest
// (move silently becomes copy), the caller's moved-from place stays
// readable after the call -- and ONLY interp's own observable execution
// behavior catches this: check and corevalidate execute nothing, so they
// are structurally blind to it (D-10-36). That pairing -- one peer catches
// what the other two cannot even in principle -- is what makes "interp is
// independent" falsifiable rather than rhetorical, and it is this phase's
// mutation obligation for the ownership fact (D-10-41).
//
// Four-beat body (pathoracle_test.go's precedent): assert clean,
// save/override/defer-restore, assert an observable effect, assert the
// specific divergence.
func TestMoveAsCopyMutationKilled(t *testing.T) {
	program, caller := moveAsCopyProbeProgram()

	// Beat 1: assert clean. The unmutated interp deletes the moved-from
	// argument place, so the stale re-read finds an uninitialized place
	// and the run correctly refuses.
	base := newFlatFrame(caller, map[string]string{caller.Parameter.ID: "AB"})
	cleanExecution, cleanErr := runFrameStack(program, base)
	if cleanErr == nil {
		t.Fatalf("expected the clean run to refuse the stale re-read of a moved-from place, got success: %+v", cleanExecution)
	}
	cleanBytes, err := CanonicalBytes(cleanExecution)
	if err != nil {
		t.Fatalf("CanonicalBytes (clean): %v", err)
	}

	// Beat 2: override, restored via defer.
	previous := moveAsCopyForTest
	moveAsCopyForTest = true
	defer func() { moveAsCopyForTest = previous }()

	// Beat 3: assert an observable effect. The mutated run no longer
	// refuses -- it wrongly succeeds, reading the stale value back out,
	// exactly the OWN-05b violation this mutant proves only interp itself
	// can catch.
	mutatedBase := newFlatFrame(caller, map[string]string{caller.Parameter.ID: "AB"})
	mutatedExecution, mutatedErr := runFrameStack(program, mutatedBase)
	if mutatedErr != nil {
		t.Fatalf("expected the mutated run to succeed (skipping the caller-side delete), got error: %v", mutatedErr)
	}
	if mutatedExecution.Outcome.Value != "AB" {
		t.Fatalf("expected the mutated run's stale re-read to return the original argument value %q, got %q", "AB", mutatedExecution.Outcome.Value)
	}
	mutatedBytes, err := CanonicalBytes(mutatedExecution)
	if err != nil {
		t.Fatalf("CanonicalBytes (mutated): %v", err)
	}

	// Beat 4: assert the specific divergence -- the clean run's refusal
	// and the mutated run's wrongful success produce different canonical
	// bytes, never merely "some error occurred".
	if bytes.Equal(cleanBytes, mutatedBytes) {
		t.Fatalf("expected the move-as-copy mutation to change interp's canonical bytes, but clean and mutated runs matched: %s", cleanBytes)
	}

	// check and corevalidate execute nothing (D-10-36): re-running BOTH on
	// a real, independently checked program (call_basic.lang) while
	// moveAsCopyForTest is still engaged from Beat 2 produces the
	// IDENTICAL verdict either way, because neither ever consults an
	// interp-internal, unexported runtime seam. This is the pairing the
	// mutant's whole point rests on -- check and corevalidate are
	// structurally blind to interp's own runtime behavior, never merely
	// coincidentally in agreement with it.
	realParsed := syntax.Parse(mustReadFixture(t, "call_basic.lang"))
	realChecked := check.Program(realParsed.Program)
	if len(realChecked.Diagnostics) != 0 {
		t.Fatalf("expected check.Program's diagnostics to be unaffected by moveAsCopyForTest, got: %v", realChecked.Diagnostics)
	}
	realValidated := corevalidate.Validate(realChecked.Program)
	if !realValidated.Valid {
		t.Fatalf("expected corevalidate.Validate's verdict to be unaffected by moveAsCopyForTest, got: %v", realValidated.Problems)
	}
}

// mustReadFixture reads a named file under testdata/phase07.
func mustReadFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join(interpProjectRoot(), "testdata", "phase07", name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return source
}

// TestCallExecutesAcrossOneFrame is Task 1's tracer test (SEM-08, OWN-05b,
// D-10-21/D-10-26): call_basic.lang's main calls identity across a real
// heap frame boundary and gets back identity's own returned value, with at
// least one emitted event attributed to the callee's own function ID
// (D-10-32) -- proof the callee frame genuinely ran rather than being
// faked by a pass-through. The same program run twice must produce
// byte-identical CanonicalBytes (D-10-26): every ordered output the frame
// stack produces comes from an ordered slice, never a Go map range.
func TestCallExecutesAcrossOneFrame(t *testing.T) {
	program := checkedCallBasicProgram(t)

	var calleeID string
	for _, function := range program.Functions {
		if function.Name == "identity" {
			calleeID = function.ID
		}
	}
	if calleeID == "" {
		t.Fatalf("call_basic.lang's checked program has no function named %q", "identity")
	}

	execution, err := Run(program, "main", "7")
	if err != nil {
		t.Fatalf("Run(main, %q) returned an unexpected error: %v", "7", err)
	}
	if execution.Outcome.Kind != "returned" {
		t.Fatalf("expected outcome kind %q, got %q", "returned", execution.Outcome.Kind)
	}
	if execution.Outcome.Value != "7" {
		t.Fatalf("expected the callee's own returned value %q, got %q", "7", execution.Outcome.Value)
	}

	foundCalleeEvent := false
	for _, event := range execution.Events {
		if event.FunctionID == calleeID {
			foundCalleeEvent = true
			break
		}
	}
	if !foundCalleeEvent {
		t.Fatalf("expected at least one event attributed to the callee's own function ID %q; events: %+v", calleeID, execution.Events)
	}

	firstBytes, err := CanonicalBytes(execution)
	if err != nil {
		t.Fatalf("CanonicalBytes: %v", err)
	}
	secondExecution, err := Run(program, "main", "7")
	if err != nil {
		t.Fatalf("second Run(main, %q) returned an unexpected error: %v", "7", err)
	}
	secondBytes, err := CanonicalBytes(secondExecution)
	if err != nil {
		t.Fatalf("CanonicalBytes (second run): %v", err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("expected two runs of the same program to produce byte-identical CanonicalBytes, got:\n%s\nvs\n%s", firstBytes, secondBytes)
	}
}
