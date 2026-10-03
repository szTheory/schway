package interp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/syntax"
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

// checkedCallBasicProgram checks and corevalidates testdata/phase07/call_basic.schway
// directly (syntax.Parse + check.Program + corevalidate.Validate), never
// through package session -- session imports interp, so an interp test
// importing session would be a cycle.
func checkedCallBasicProgram(t *testing.T) core.Program {
	t.Helper()
	path := filepath.Join(interpProjectRoot(), "testdata", "phase07", "call_basic.schway")
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

type phase23ModelOutcome struct {
	kind  string
	type_ string
	value string
}

func checkedPhase23FileByteProgram(t *testing.T) core.Program {
	t.Helper()
	path := filepath.Join(interpProjectRoot(), "examples", "phase23", "file_byte.schway")
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

func phase23OutcomesForByte(t *testing.T, program core.Program, byteValue uint64) map[string]phase23ModelOutcome {
	t.Helper()
	if len(program.Functions) != 1 || program.Functions[0].Linear == nil {
		t.Fatalf("expected one linear file-byte function, got %+v", program.Functions)
	}
	outcomes := make(map[string]phase23ModelOutcome)
	for _, operation := range program.Functions[0].Linear.Operations {
		if operation.Foreign == nil {
			continue
		}
		switch operation.Foreign.Mode {
		case "acquire":
			outcomes[operation.ID] = phase23ModelOutcome{kind: "success", type_: "FileByteOwner", value: fmt.Sprint(byteValue)}
		case "borrow":
			outcomes[operation.ID] = phase23ModelOutcome{kind: "success", type_: "U64", value: fmt.Sprint(byteValue)}
		}
	}
	if len(outcomes) != 2 {
		t.Fatalf("expected distinct acquire and borrow outcomes keyed by operation, got %+v", outcomes)
	}
	return outcomes
}

func runPhase23Model(program core.Program, input string, outcomes map[string]phase23ModelOutcome) (ModelResult, error) {
	modeled := make(map[string]ForeignOutcome, len(outcomes))
	for operationID, outcome := range outcomes {
		modeled[operationID] = ForeignOutcome{Kind: outcome.kind, Type: outcome.type_, Value: outcome.value}
	}
	return RunWithForeignOutcomes(program, "main", input, modeled)
}

func TestPhase23ModelByteOutcomesAreDeterministicAndReleaseLocally(t *testing.T) {
	program := checkedPhase23FileByteProgram(t)
	for _, test := range []struct {
		name      string
		byteValue uint64
		want      string
	}{
		{name: "0x41", byteValue: 65, want: "65"},
		{name: "0x42", byteValue: 66, want: "66"},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := runPhase23Model(program, "model-only-path-token", phase23OutcomesForByte(t, program, test.byteValue))
			if err != nil {
				t.Fatalf("model execution failed: %v", err)
			}
			if result.Execution.Outcome.Kind != "returned" || result.Execution.Outcome.Value != test.want {
				t.Fatalf("modeled outcome = %+v, want returned:%s", result.Execution.Outcome, test.want)
			}
			if len(result.Execution.LiveResources) != 0 {
				t.Fatalf("model left local obligations live: %v", result.Execution.LiveResources)
			}
			releaseEvents := 0
			for _, event := range result.Execution.Events {
				if event.Kind == "resource.released" {
					releaseEvents++
				}
			}
			if releaseEvents != 1 {
				t.Fatalf("modeled local release events = %d, want exactly one", releaseEvents)
			}
			if result.EvidenceScope != EvidenceScopeModelOnly || result.ActualHostIO || result.PhysicalCleanup {
				t.Fatalf("model result overstates its evidence scope: %+v", result)
			}
		})
	}
}

func phase23AcquireFailureOutcome(t *testing.T, program core.Program) map[string]phase23ModelOutcome {
	t.Helper()
	for _, operation := range program.Functions[0].Linear.Operations {
		if operation.Foreign != nil && operation.Foreign.Mode == "acquire" {
			return map[string]phase23ModelOutcome{operation.ID: {kind: "failure", type_: "AcquireError", value: "AcquireFailed"}}
		}
	}
	t.Fatal("checked file-byte entry has no acquire operation")
	return nil
}

func phase23UseFailureOutcomes(t *testing.T, program core.Program) map[string]phase23ModelOutcome {
	t.Helper()
	outcomes := phase23OutcomesForByte(t, program, 67)
	for _, operation := range program.Functions[0].Linear.Operations {
		if operation.Foreign != nil && operation.Foreign.Mode == "borrow" {
			outcomes[operation.ID] = phase23ModelOutcome{kind: "failure", type_: "UseError", value: "UnsupportedByte"}
			return outcomes
		}
	}
	t.Fatal("checked file-byte entry has no borrowed use operation")
	return nil
}

func TestPhase23ModelAcquireFailureCreatesNoOwner(t *testing.T) {
	program := checkedPhase23FileByteProgram(t)
	result, err := runPhase23Model(program, "model-only-path-token", phase23AcquireFailureOutcome(t, program))
	if err != nil {
		t.Fatalf("modeled acquisition failure: %v", err)
	}
	if result.Execution.Outcome.Kind != execution.OutcomeTypedFailure || result.Execution.Outcome.Value != "AcquireFailed" {
		t.Fatalf("modeled acquisition failure = %+v", result.Execution.Outcome)
	}
	if len(result.Execution.LiveResources) != 0 {
		t.Fatalf("failed acquisition minted a live owner: %v", result.Execution.LiveResources)
	}
	for _, event := range result.Execution.Events {
		if event.Kind == "resource.released" {
			t.Fatalf("failed acquisition must not release a nonexistent owner: %+v", event)
		}
	}
}

func TestPhase24RepeatedHelperTypedErrorDrainsEachActivation(t *testing.T) {
	path := filepath.Join(interpProjectRoot(), "examples", "phase24", "error.schway")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse error fixture: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check error fixture: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("validate error fixture: %+v", validated.Problems)
	}
	program := validated.Program()
	var acquireID, useID string
	for _, function := range program.Functions {
		for _, operation := range function.Linear.Operations {
			if operation.Foreign == nil {
				continue
			}
			if operation.Foreign.Mode == "acquire" {
				acquireID = operation.ID
			}
			if function.Name == "probe" && operation.Foreign.Mode == "borrow" {
				useID = operation.ID
			}
		}
	}
	if acquireID == "" || useID == "" {
		t.Fatal("error fixture acquire/use operation missing")
	}
	result, err := RunWithForeignOutcomes(program, "main", "opaque-path-token", map[string]ForeignOutcome{
		acquireID: {Kind: "success", Type: "FileByteOwner", Value: "67"},
		useID:     {Kind: "failure", Type: "UseError", Value: "UnsupportedByte"},
	})
	if err != nil {
		t.Fatalf("model repeated typed error: %v", err)
	}
	if result.Execution.Outcome.Kind != execution.OutcomeTypedFailure || result.Execution.Outcome.Value != "UnsupportedByte" {
		t.Fatalf("model outcome=%+v; want exact UseError.UnsupportedByte failure", result.Execution.Outcome)
	}
	var acquisitions, releases []Event
	for _, event := range result.Execution.Events {
		if event.ID == acquireID+":event" {
			acquisitions = append(acquisitions, event)
		}
		if event.Kind == "resource.released" {
			releases = append(releases, event)
		}
	}
	if len(acquisitions) != 3 {
		t.Fatalf("dynamic acquisition events=%d; want three", len(acquisitions))
	}
	seenActivations := map[string]bool{}
	for _, event := range acquisitions {
		if event.Invocation == "" || seenActivations[event.Invocation] {
			t.Fatalf("repeated acquisition activation is empty or colliding: %+v", acquisitions)
		}
		seenActivations[event.Invocation] = true
	}
	if len(releases) != 3 || releases[0].FunctionID != functionIDByName(program, "probe") || releases[1].FunctionID != functionIDByName(program, "main") || releases[2].FunctionID != functionIDByName(program, "main") {
		t.Fatalf("release function order=%v; want callee C then caller B,A", eventFunctions(releases))
	}
	var typedFailure Event
	for _, event := range result.Execution.Events {
		if event.Kind == "function.failed" {
			typedFailure = event
		}
	}
	if typedFailure.TypeID != "UseError" || len(result.Execution.LiveResources) != 0 {
		t.Fatalf("failure event=%+v live=%v; want original UseError and zero live owners", typedFailure, result.Execution.LiveResources)
	}
	if result.ActualHostIO || result.PhysicalCleanup || result.EvidenceScope != EvidenceScopeModelOnly {
		t.Fatalf("model execution overstated evidence: %+v", result)
	}

	failedAcquire, err := RunWithForeignOutcomes(program, "main", "opaque-path-token", map[string]ForeignOutcome{
		acquireID: {Kind: "failure", Type: "AcquireError", Value: "AcquireFailed"},
	})
	if err != nil {
		t.Fatalf("model failed helper acquisition: %v", err)
	}
	if failedAcquire.Execution.Outcome.Kind != execution.OutcomeTypedFailure || failedAcquire.Execution.Outcome.Value != "AcquireFailed" || len(failedAcquire.Execution.LiveResources) != 0 {
		t.Fatalf("failed helper acquisition outcome=%+v live=%v; want typed failure and no owner", failedAcquire.Execution.Outcome, failedAcquire.Execution.LiveResources)
	}
	for _, event := range failedAcquire.Execution.Events {
		if event.Kind == "resource.released" {
			t.Fatalf("failed helper acquisition released a nonexistent owner: %+v", event)
		}
	}

	_, err = RunWithForeignOutcomes(program, "main", "opaque-path-token", map[string]ForeignOutcome{
		acquireID: {Kind: "success", Type: "FileByteOwner", Value: "65"},
		useID:     {Kind: "failure", Type: "UseError", Value: "UnsupportedByte"},
	})
	if err == nil || !strings.Contains(err.Error(), "cannot report UnsupportedByte") {
		t.Fatalf("model accepted UnsupportedByte for supported byte 0x41: %v", err)
	}
}

func functionIDByName(program core.Program, name string) string {
	for _, function := range program.Functions {
		if function.Name == name {
			return function.ID
		}
	}
	return ""
}

func eventFunctions(events []Event) []string {
	functions := make([]string, len(events))
	for index, event := range events {
		functions[index] = event.FunctionID
	}
	return functions
}

func TestPhase23ModelUseFailureKeepsOwnerLiveUntilRelease(t *testing.T) {
	program := checkedPhase23FileByteProgram(t)
	result, err := runPhase23Model(program, "model-only-path-token", phase23UseFailureOutcomes(t, program))
	if err != nil {
		t.Fatalf("modeled borrowed-use failure: %v", err)
	}
	if result.Execution.Outcome.Kind != execution.OutcomeTypedFailure || result.Execution.Outcome.Value != "UnsupportedByte" {
		t.Fatalf("modeled use failure = %+v", result.Execution.Outcome)
	}
	if len(result.Execution.LiveResources) != 0 {
		t.Fatalf("modeled failure path left a local owner live: %v", result.Execution.LiveResources)
	}
	failedAt, releasedAt := -1, -1
	for index, event := range result.Execution.Events {
		if event.Kind == "foreign.failed" {
			failedAt = index
		}
		if event.Kind == "resource.released" {
			releasedAt = index
		}
	}
	if failedAt < 0 || releasedAt <= failedAt {
		t.Fatalf("modeled error must retain the acquired owner until local release: %+v", result.Execution.Events)
	}
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
	base := newFlatFrame(caller, map[string]value{caller.Parameter.ID: {tag: "", payload: "AB"}})
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
	mutatedBase := newFlatFrame(caller, map[string]value{caller.Parameter.ID: {tag: "", payload: "AB"}})
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
	// a real, independently checked program (call_basic.schway) while
	// moveAsCopyForTest is still engaged from Beat 2 produces the
	// IDENTICAL verdict either way, because neither ever consults an
	// interp-internal, unexported runtime seam. This is the pairing the
	// mutant's whole point rests on -- check and corevalidate are
	// structurally blind to interp's own runtime behavior, never merely
	// coincidentally in agreement with it.
	realParsed := syntax.Parse(mustReadFixture(t, "call_basic.schway"))
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

func TestPhase18CallComputedMatchPrefix(t *testing.T) {
	path := filepath.Join(interpProjectRoot(), "testdata", "phase18", "result_computed_match.schway")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check: %+v", checked.Diagnostics)
	}
	if peer := corevalidate.Validate(checked.Program); !peer.Valid {
		t.Fatalf("core peer: %+v", peer.Problems)
	}
	result, err := Run(checked.Program, "main", "Raw")
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome.Kind != "returned" || result.Outcome.Value != "Accepted" {
		t.Fatalf("outcome = %+v, want Accepted", result.Outcome)
	}
}

// TestCallExecutesAcrossOneFrame is Task 1's tracer test (SEM-08, OWN-05b,
// D-10-21/D-10-26): call_basic.schway's main calls identity across a real
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
		t.Fatalf("call_basic.schway's checked program has no function named %q", "identity")
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

// TestInvocationThreadsThroughAllEventPaths is the /2 producer contract:
// every event emitted by a multi-function activation carries the activation's
// canonical occurrence path, including abrupt terminal paths.  The diamond
// adds the important shared-leaf case where static event IDs repeat but their
// invocation-qualified identities do not.
func TestInvocationThreadsThroughAllEventPaths(t *testing.T) {
	for _, entry := range interpOracleCorpus(t) {
		entry := entry
		t.Run(entry.name, func(t *testing.T) {
			result, err := entry.run(t)
			if err != nil {
				t.Fatal(err)
			}
			if result.Schema != execution.Schema2 {
				return // single-function producers deliberately remain legacy.
			}
			for index, event := range result.Events {
				if event.Schema != execution.Schema2 {
					t.Fatalf("event[%d] schema = %q, want /2", index, event.Schema)
				}
				if _, err := execution.ParseInvocation(event.Invocation); err != nil {
					t.Fatalf("event[%d] invocation %q is not canonical: %v", index, event.Invocation, err)
				}
			}
		})
	}

	program := checkedProgramFromFixture(t, "phase11", "multi_function_diamond_call.schway")
	result, err := Run(program, "main", "7")
	if err != nil {
		t.Fatal(err)
	}
	if result.Schema != execution.Schema2 {
		t.Fatalf("diamond schema = %q, want %q", result.Schema, execution.Schema2)
	}
	seen := map[string]bool{}
	leafInvocations := map[string]bool{}
	for _, event := range result.Events {
		if event.Schema != execution.Schema2 || event.Invocation == "" {
			t.Fatalf("/2 event lacks /2 identity: %+v", event)
		}
		key := event.Invocation + "\x00" + event.ID
		if seen[key] {
			t.Fatalf("duplicate /2 event identity %q", key)
		}
		seen[key] = true
		if strings.Contains(event.FunctionID, ":fn:leaf") && event.Kind == "function.returned" {
			leafInvocations[event.Invocation] = true
		}
	}
	if len(leafInvocations) != 2 {
		t.Fatalf("shared leaf ran under %d invocation(s), want 2: %+v", len(leafInvocations), result.Events)
	}
}

// TestExecutionSchemaSelectionPreservesLegacy freezes both legacy producer
// boundaries: adding /2 must not alter a bare-match /0 document or a
// frame-based /1 document.
func TestExecutionSchemaSelectionPreservesLegacy(t *testing.T) {
	bareProgram := checkedProgramFromFixture(t, "phase1", "toggle.schway")
	bare, err := Run(bareProgram, bareProgram.Functions[0].Name, "Off")
	if err != nil {
		t.Fatal(err)
	}
	if bare.Schema != execution.Schema0 || len(bare.Events) != 1 || bare.Events[0].Schema != execution.Schema0 || bare.Events[0].Invocation != "" {
		t.Fatalf("single-function bare-match execution moved from legacy /0: %+v", bare)
	}

	program := checkedProgramFromFixture(t, "phase4", "defect_terminal.schway")
	result, err := Run(program, "triage", "Go")
	if err != nil {
		t.Fatal(err)
	}
	if result.Schema != execution.Schema1 {
		t.Fatalf("single-function schema = %q, want %q", result.Schema, execution.Schema1)
	}
	got, err := CanonicalBytes(result)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(oracleGoldenDir(), "single_frame_return.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("legacy /1 canonical bytes moved\n got: %s\nwant: %s", got, want)
	}
}

func functionCalledEvents(events []Event) []Event {
	called := []Event{}
	for _, event := range events {
		if event.Kind == "function.called" {
			called = append(called, event)
		}
	}
	return called
}

// TestFunctionCalledPreorderAndOwnership makes caller-to-callee causality a
// first-class /2 observation. The diamond also proves that a caller resumes
// only after the complete child subsequence.
func TestFunctionCalledPreorderAndOwnership(t *testing.T) {
	program := checkedProgramFromFixture(t, "phase11", "multi_function_diamond_call.schway")
	result, err := Run(program, "main", "7")
	if err != nil {
		t.Fatal(err)
	}
	called := functionCalledEvents(result.Events)
	if len(called) != 4 {
		t.Fatalf("function.called projection has %d events, want 4: %+v", len(called), called)
	}
	for _, event := range called {
		if event.Schema != execution.Schema2 || event.Invocation == "" || event.CalleeFunctionID == "" {
			t.Fatalf("called event lacks /2 caller/callee identity: %+v", event)
		}
		if event.ID != strings.TrimSuffix(event.ID, ":event:called")+":event:called" {
			t.Fatalf("called event ID is not canonical: %+v", event)
		}
		childStart := indexOfInvocationEvent(result.Events, event.ID, event.Invocation)
		if childStart < 0 {
			t.Fatalf("called event disappeared from full sequence: %+v", event)
		}
		for next := childStart + 1; next < len(result.Events); next++ {
			candidate := result.Events[next]
			if candidate.FunctionID == event.CalleeFunctionID {
				break
			}
			if candidate.Invocation == event.Invocation && candidate.FunctionID == event.FunctionID {
				t.Fatalf("caller event resumed before callee subsequence for %+v; next caller event %+v", event, candidate)
			}
			if next == len(result.Events)-1 {
				t.Fatalf("no callee event follows called edge %+v", event)
			}
		}
	}
}

func indexOfInvocationEvent(events []Event, id, invocation string) int {
	for index, event := range events {
		if event.ID == id && event.Invocation == invocation {
			return index
		}
	}
	return -1
}

// TestFunctionCalledProjectionRemoval compares only causal edges: removing
// left's reachable leaf call leaves left itself reachable, so the projection
// must lose exactly that one caller-owned function.called record.
func TestFunctionCalledProjectionRemoval(t *testing.T) {
	full := checkedProgramFromFixture(t, "phase11", "multi_function_diamond_call.schway")
	fullResult, err := Run(full, "main", "7")
	if err != nil {
		t.Fatal(err)
	}
	fullCalled := functionCalledEvents(fullResult.Events)

	sourcePath := filepath.Join(interpProjectRoot(), "testdata", "phase11", "multi_function_diamond_call.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	withoutLeafCall := strings.Replace(string(source), "  let result = leaf(value)\n  result", "  value", 1)
	parsed := syntax.Parse([]byte(withoutLeafCall))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse removal control: %v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check removal control: %v", checked.Diagnostics)
	}
	reduced, err := Run(checked.Program, "main", "7")
	if err != nil {
		t.Fatal(err)
	}
	reducedCalled := functionCalledEvents(reduced.Events)
	if got, want := len(fullCalled)-len(reducedCalled), 1; got != want {
		t.Fatalf("function.called projection loss = %d, want %d; full=%+v reduced=%+v", got, want, fullCalled, reducedCalled)
	}
}

// TestRejectedCallEmitsNoCalledEdge covers both admissions that do not become
// successful calls: unresolved callees fail before an execution exists, and
// a depth refusal returns only events from already-admitted ancestors.
func TestRejectedCallEmitsNoCalledEdge(t *testing.T) {
	t.Run("unresolved", func(t *testing.T) {
		program, caller := moveAsCopyProbeProgram()
		caller.Linear.Operations[0].CalleeID = "missing:callee"
		program.Functions[1] = caller
		base := newFlatFrame(caller, map[string]value{caller.Parameter.ID: {payload: "V"}})
		if _, err := runProgramFrameStack(program, base); err == nil {
			t.Fatal("unresolved call unexpectedly executed")
		}
	})
	t.Run("depth", func(t *testing.T) {
		program := generateAndCheckCallDepthChain(t, 3)
		previous := maxCallDepthOverride
		maxCallDepthOverride = func() int { return 2 }
		defer func() { maxCallDepthOverride = previous }()
		result, err := Run(program, "link0", "7")
		if err != nil {
			t.Fatal(err)
		}
		for _, event := range functionCalledEvents(result.Events) {
			if strings.Contains(event.ID, "link1") {
				t.Fatalf("depth-rejected call emitted a called edge: %+v", event)
			}
		}
		if got := len(functionCalledEvents(result.Events)); got != 1 {
			t.Fatalf("depth refusal admitted %d called edges, want only the first successful call", got)
		}
	})
}

// forbiddenOwnershipAccessors is corevalidate.Result's own ownership-bearing
// accessor set (D-10-36/D-10-37): PeerSignatures and PeerSiteCoverage both
// expose the summary peer's own call-site derivation contract. Validate,
// Valid, Problems, and Program stay permitted -- they are the fail-closed
// precondition Run's own corevalidate.Validate(program) call already relies
// on, never the ownership fact itself.
var forbiddenOwnershipAccessors = []string{"PeerSignatures", "PeerSiteCoverage"}

// scanForForbiddenOwnershipAccessors parses source with go/parser and
// reports every forbidden accessor name (forbiddenOwnershipAccessors)
// referenced anywhere as a selector expression (`x.Name`), regardless of
// x's own static type -- a purely syntactic scan, matching this repo's
// established import/reference-scanning guard-test technique (see
// originvalidate_test.go's TestOriginValidatorImportsStayIndependent).
func scanForForbiddenOwnershipAccessors(filename string, source []byte) ([]string, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filename, source, 0)
	if err != nil {
		return nil, err
	}
	var found []string
	ast.Inspect(file, func(n ast.Node) bool {
		selector, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		for _, forbidden := range forbiddenOwnershipAccessors {
			if selector.Sel.Name == forbidden {
				found = append(found, forbidden)
			}
		}
		return true
	})
	return found, nil
}

// TestInterpDoesNotReadCorevalidateOwnershipFields is Task 3's OWN-05b
// guard (D-10-36/D-10-37, QLT-08): today interp references neither
// corevalidate.Result.PeerSignatures nor .PeerSiteCoverage, but that is
// true by OMISSION, not by enforcement (D-10-36) -- nothing today converts
// a future reference into a build failure or a red test. This test
// converts it into "cannot without a red test": it scans every non-test
// .go file in this package via go/parser, resolved through the existing
// interpProjectRoot() helper (never testsupport, which pulls in session,
// which imports interp -- a cycle), and fails if any references either
// forbidden accessor.
//
// NARROWED CLAIM (D-10-37, PHASE-10-DEBT.md): this proves independence of
// DERIVATION MECHANISM for the ownership-transfer fact specifically,
// nested inside a shared, unrelated validation dependency -- NOT the
// mutual non-import independence check and corevalidate have from each
// other. interp's Run still calls corevalidate.Validate(program) as its
// own fail-closed precondition (import kept deliberately: hoisting
// Validate to Run's own callers would weaken interp's fail-closed posture
// and touch every call site, for an import-graph purity the current
// one-value domain does not need), so a bug in Validate's own derivation
// would feed interp bad input too (Knight and Leveson's correlated-fault
// result). The claim this test proves is narrower: interp never
// additionally CONSULTS corevalidate's own ownership-bearing verdict to
// decide its OWN runtime behavior -- OWN-05b's fact falls out of interp's
// own execution alone.
//
// The negative control proves the guard is load-bearing rather than
// vacuously green: it runs the identical scan over a small synthetic
// source string containing a forbidden accessor reference and asserts the
// scan reports it -- a guard that has never been seen to fire is not
// evidence.
func TestInterpDoesNotReadCorevalidateOwnershipFields(t *testing.T) {
	t.Run("real_scan", func(t *testing.T) {
		dir := filepath.Join(interpProjectRoot(), "internal", "compiler", "interp")
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			found, err := scanForForbiddenOwnershipAccessors(path, source)
			if err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			if len(found) > 0 {
				t.Fatalf("%s references forbidden ownership-bearing accessor(s) %v -- interp must derive OWN-05b from its own observable execution, never by reading corevalidate's verdict", entry.Name(), found)
			}
		}
	})

	t.Run("negative_control", func(t *testing.T) {
		const synthetic = `package fake

type result struct{}

func (result) PeerSignatures() int { return 0 }

func probe(r result) int {
	return r.PeerSignatures()
}
`
		found, err := scanForForbiddenOwnershipAccessors("synthetic.go", []byte(synthetic))
		if err != nil {
			t.Fatalf("parse synthetic source: %v", err)
		}
		if len(found) == 0 {
			t.Fatal("expected the scan to report a violation on synthetic source containing a forbidden accessor reference -- the guard has never been observed to fire")
		}
	})
}

// generateCallDepthChainSource emits a genuine `.schway` module of n chained
// single-call functions, generalizing testdata/phase07/call_basic.schway's
// two-function template to n links (Task 2): link0 calls link1 calls
// link2 ... calls link(n-1), which is the base case with no call and simply
// returns its own parameter. link0 is the sole exported entry point. The
// full contract (module shape, per-link template, entry point name, and
// why the chain is generated rather than committed as a 129-function
// `.schway` file) is recorded in
// testdata/phase10/call_depth_chain_generator.md so a future reader can
// reconstruct the fixture without reading this function.
func generateCallDepthChainSource(n int) []byte {
	var b strings.Builder
	b.WriteString("module phase10.call_depth_chain\n\n")
	b.WriteString("export {\n  fn link0\n}\n\n")
	for i := n - 1; i >= 0; i-- {
		if i == n-1 {
			fmt.Fprintf(&b, "fn link%d(value: Byte) -> Byte {\n  value\n}\n\n", i)
			continue
		}
		fmt.Fprintf(&b, "fn link%d(value: Byte) -> Byte {\n  let result = link%d(value)\n  result\n}\n\n", i, i+1)
	}
	return []byte(b.String())
}

// generateAndCheckCallDepthChain generates an n-link call-depth chain
// (generateCallDepthChainSource) and drives it through the REAL pipeline --
// syntax.Parse -> check.Program -> corevalidate.Validate -- asserting a
// clean result at EACH stage before returning the checked core.Program, so
// a failure at an earlier stage is reported as itself rather than being
// mistaken for a depth refusal (Task 2, D-10-24: never a hand-built
// core.Program). Parameterized by n so this ONE helper drives Task 1's
// both-directions boundary test (n = MaxCallDepth, n = MaxCallDepth+1),
// Task 2's pipeline-conformance test, and Task 3's subprocess probe.
func generateAndCheckCallDepthChain(t *testing.T, n int) core.Program {
	t.Helper()
	source := generateCallDepthChainSource(n)
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("parse (n=%d): unexpected diagnostics: %v", n, parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) > 0 {
		t.Fatalf("check (n=%d): unexpected diagnostics: %v", n, checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("corevalidate (n=%d) rejected: %v", n, validated.Problems)
	}
	return validated.Program()
}

// TestCallDepthAtAndOverTheCap is Task 1's boundary test (SEM-08,
// D-10-23): a chain of exactly MaxCallDepth functions must execute to
// completion and return normally, while a chain one function deeper must
// terminate in the named depth-exceeded refusal -- both driven through the
// REAL pipeline via generateAndCheckCallDepthChain, never a hand-built
// core.Program.
func TestCallDepthAtAndOverTheCap(t *testing.T) {
	t.Run("at_the_cap", func(t *testing.T) {
		program := generateAndCheckCallDepthChain(t, MaxCallDepth)
		result, err := Run(program, "link0", "7")
		if err != nil {
			t.Fatalf("Run at MaxCallDepth (%d): unexpected error: %v", MaxCallDepth, err)
		}
		if result.Outcome.Kind != "returned" {
			t.Fatalf("expected outcome kind %q at the cap, got %q", "returned", result.Outcome.Kind)
		}
		if result.Outcome.Value != "7" {
			t.Fatalf("expected the chain's own base-case value %q, got %q", "7", result.Outcome.Value)
		}
	})

	t.Run("over_the_cap", func(t *testing.T) {
		program := generateAndCheckCallDepthChain(t, MaxCallDepth+1)
		result, err := Run(program, "link0", "7")
		if err != nil {
			t.Fatalf("Run over MaxCallDepth: unexpected error: %v", err)
		}

		validKind := false
		for _, kind := range execution.TerminalOutcomeKinds() {
			if result.Outcome.Kind == kind {
				validKind = true
				break
			}
		}
		if !validKind {
			t.Fatalf("expected the refusal's outcome kind %q to be a member of execution.TerminalOutcomeKinds()", result.Outcome.Kind)
		}

		foundReason := false
		for _, event := range result.Events {
			if event.Output == callDepthExceededDefectReason {
				foundReason = true
			}
		}
		if !foundReason {
			t.Fatalf("expected an event carrying the named depth-limit reason %q, got events: %+v", callDepthExceededDefectReason, result.Events)
		}

		firstBytes, err := CanonicalBytes(result)
		if err != nil {
			t.Fatalf("CanonicalBytes (first): %v", err)
		}
		if !bytes.Contains(firstBytes, []byte(callDepthExceededDefectReason)) {
			t.Fatalf("expected CanonicalBytes to contain the named depth-limit reason %q, got: %s", callDepthExceededDefectReason, firstBytes)
		}

		second, err := Run(program, "link0", "7")
		if err != nil {
			t.Fatalf("second Run over MaxCallDepth: unexpected error: %v", err)
		}
		secondBytes, err := CanonicalBytes(second)
		if err != nil {
			t.Fatalf("CanonicalBytes (second): %v", err)
		}
		if !bytes.Equal(firstBytes, secondBytes) {
			t.Fatalf("expected two runs of the same over-depth program to produce byte-identical CanonicalBytes, got:\n%s\nvs\n%s", firstBytes, secondBytes)
		}
	})
}

// TestCallDepthExceeded is Task 2's pipeline-conformance test (SEM-08): a
// genuine 129-function chain -- one function past MaxCallDepth -- is
// generated, parsed, checked, and corevalidated
// (generateAndCheckCallDepthChain, which itself asserts a clean result at
// each of those three stages), and interp.Run on the checked program must
// reach the named depth-exceeded refusal, never a parse/check/corevalidate
// diagnostic mistaken for it. See
// testdata/phase10/call_depth_chain_generator.md for the generator's full
// contract. D-10-24: this NEVER hand-builds a core.Program -- the whole
// point is that the compiler's real pipeline genuinely admits this program.
func TestCallDepthExceeded(t *testing.T) {
	const n = MaxCallDepth + 1
	program := generateAndCheckCallDepthChain(t, n)

	result, err := Run(program, "link0", "9")
	if err != nil {
		t.Fatalf("Run(%d-function chain): unexpected error: %v", n, err)
	}
	if result.Outcome.Kind != execution.OutcomeDefect {
		t.Fatalf("expected the depth-exceeded refusal's outcome kind %q, got %q", execution.OutcomeDefect, result.Outcome.Kind)
	}
	found := false
	for _, event := range result.Events {
		if event.Output == callDepthExceededDefectReason {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected an event carrying the named depth-limit reason %q, got events: %+v", callDepthExceededDefectReason, result.Events)
	}
}

// probeReducedCallDepthCap and probeCallDepthMultiplier are Task 3's
// Pitfall-4 probe constants (D-10-44): the production MaxCallDepth (128)
// times a multiplier large enough to meaningfully exercise a pinned small
// host stack would exceed the parser's own maxFunctions ceiling (1024,
// syntax/parser.go:16), so the probe expresses its multiplier against this
// smaller, in-test-only cap instead. Both constants -- and the derived
// probeChainDepth -- are fixed in source BEFORE this probe's first passing
// run, so the "threshold chosen after seeing the result" attack is
// foreclosed; the commit that first makes TestNativeStackHeadroomIndependentOfCallDepth
// pass states in its own message that these constants predate that first
// green run, so review can confirm the ordering from git history alone.
const (
	probeReducedCallDepthCap  = 8
	probeCallDepthMultiplier  = 100
	probeChainDepth           = probeReducedCallDepthCap * probeCallDepthMultiplier // 800, held well under maxFunctions=1024
	probeMaxStackBytes        = 1 << 20                                             // 1 MiB: a deliberately small, pinned host ceiling
	probeSubprocessTimeout    = 30 * time.Second
	probeChildEnv             = "SCHWAY_INTERP_CALL_DEPTH_PROBE_CHILD"
	probeArmEnv               = "SCHWAY_INTERP_CALL_DEPTH_PROBE_ARM"
	probeArmCapDisabled       = "cap_disabled"
	probeArmCapEnabled        = "cap_enabled"
	probeMaxCapturedOutputLen = 1 << 16
)

// probeBoundedWriter caps captured child-process output at
// probeMaxCapturedOutputLen, mirroring testsupport.boundedWriter's shape
// (internal/compiler/testsupport/testsupport.go) without depending on its
// unexported type -- an interp test cannot import testsupport (it pulls in
// session, which imports interp: a cycle). Excess bytes are discarded, never
// buffered, so a runaway child cannot exhaust test memory.
type probeBoundedWriter struct {
	buffer bytes.Buffer
	total  int
}

func (w *probeBoundedWriter) Write(data []byte) (int, error) {
	w.total += len(data)
	remaining := probeMaxCapturedOutputLen - w.buffer.Len()
	if remaining > 0 {
		if len(data) > remaining {
			w.buffer.Write(data[:remaining])
		} else {
			w.buffer.Write(data)
		}
	}
	return len(data), nil
}

func (w *probeBoundedWriter) bytes() []byte { return w.buffer.Bytes() }

// runCallDepthProbeSubprocess re-execs the current test binary
// (os.Args[0]) restricted to this one test via -test.run, guarded by
// probeChildEnv so the child branches into runCallDepthProbeChild instead
// of recursing, with probeArmEnv selecting which arm the child runs.
// exec.CommandContext with a real deadline (never context.Background()) and
// two independently bounded probeBoundedWriter streams (never .Output() or
// .CombinedOutput()'s unbounded merge, and never .StdoutPipe()'s unbounded
// pipe read) -- this repository's own TestSourceNeverSpawnsUnboundedProcesses
// guard (internal/compiler/native/native_test.go, D-02-01) forbids both.
func runCallDepthProbeSubprocess(t *testing.T, arm string) (stdout, stderr []byte, err error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), probeSubprocessTimeout)
	defer cancel()

	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeStackHeadroomIndependentOfCallDepth$", "-test.v")
	command.Env = append(os.Environ(), probeChildEnv+"=1", probeArmEnv+"="+arm)

	var stdoutWriter, stderrWriter probeBoundedWriter
	command.Stdout = &stdoutWriter
	command.Stderr = &stderrWriter

	runErr := command.Run()
	return stdoutWriter.bytes(), stderrWriter.bytes(), runErr
}

// runCallDepthProbeChild is the re-exec'd child's own logic (Task 3,
// D-10-42/D-10-43): it pins runtime/debug.SetMaxStack to a small, fixed
// ceiling (probeMaxStackBytes) for determinism across platforms and across
// the Linux and macOS CI runners, then drives a genuine probeChainDepth-
// function chain -- generated and checked through the real pipeline exactly
// like every other test in this file -- under one of two arms selected by
// probeArmEnv.
//
// This is genuinely NEW machinery for this repository: no subprocess,
// SetMaxStack, or TestMain re-exec pattern existed anywhere in this
// codebase before this test (10-RESEARCH.md's "No Analog Found" note). It
// converts Assumption A1 ([ASSUMED], 10-RESEARCH.md Assumptions Log) --
// that Go's stack-exhaustion runtime.throw is fatal and unrecoverable via
// recover() -- from an assumption into a directly observed result for this
// repository's toolchain: this test never exercises that path at all,
// because it exists precisely to demonstrate the STRONGER, honest finding
// (D-10-43) that language call depth never touches host stack in the first
// place. Because interp's call stack is an explicit heap []frame slice
// (D-10-21), Lang-level call depth costs O(1) host stack; the result below
// is therefore not a safety-margin ratio or a near-miss stress test -- it
// is a direct measurement that the host-stack limit and the language-level
// call-depth bound are two STRUCTURALLY UNRELATED limits, never one limit
// wearing two names (Roadmap criterion 2's Pitfall-4 Gate).
func runCallDepthProbeChild(t *testing.T) {
	t.Helper()
	debug.SetMaxStack(probeMaxStackBytes)

	switch arm := os.Getenv(probeArmEnv); arm {
	case probeArmCapDisabled:
		// Cap lifted through the nil-default maxCallDepthOverride seam
		// (never a production-mutable exported global): the chain runs
		// probeChainDepth deep -- many multiples past MaxCallDepth -- and
		// must exit cleanly despite the pinned small host ceiling, because
		// language call depth was never consuming host stack to begin
		// with.
		previous := maxCallDepthOverride
		maxCallDepthOverride = func() int { return probeChainDepth + 1 }
		defer func() { maxCallDepthOverride = previous }()

		program := generateAndCheckCallDepthChain(t, probeChainDepth)
		result, err := Run(program, "link0", "1")
		if err != nil {
			t.Fatalf("cap-disabled arm: unexpected error at chain depth %d under a %d-byte host stack ceiling: %v", probeChainDepth, probeMaxStackBytes, err)
		}
		if result.Outcome.Kind != "returned" {
			t.Fatalf("cap-disabled arm: expected outcome kind %q, got %q", "returned", result.Outcome.Kind)
		}
		fmt.Println("cap_disabled: ok")

	case probeArmCapEnabled:
		// The cap stays at its declared production value (MaxCallDepth,
		// maxCallDepthOverride left nil): the SAME probeChainDepth chain is
		// driven under the identical pinned small host stack ceiling, and
		// the named depth-exceeded refusal must fire well before the
		// chain's own end -- the language-level bound is what stops
		// execution, never a host stack limit that was never actually
		// threatened.
		program := generateAndCheckCallDepthChain(t, probeChainDepth)
		result, err := Run(program, "link0", "1")
		if err != nil {
			t.Fatalf("cap-enabled arm: unexpected error: %v", err)
		}
		if result.Outcome.Kind != execution.OutcomeDefect {
			t.Fatalf("cap-enabled arm: expected the depth-exceeded refusal's outcome kind %q, got %q", execution.OutcomeDefect, result.Outcome.Kind)
		}
		found := false
		for _, event := range result.Events {
			if event.Output == callDepthExceededDefectReason {
				found = true
			}
		}
		if !found {
			t.Fatalf("cap-enabled arm: expected an event carrying the named depth-limit reason %q, got events: %+v", callDepthExceededDefectReason, result.Events)
		}
		fmt.Println("cap_enabled: ok")

	default:
		t.Fatalf("unknown or missing probe arm %q (%s)", arm, probeArmEnv)
	}
}

// TestNativeStackHeadroomIndependentOfCallDepth is Task 3's Pitfall-4 gate
// (Roadmap criterion 2, D-10-42/D-10-43/D-10-44): a SUBPROCESS probe, never
// an in-process headroom-ratio measurement -- an in-process measurement is
// an assertion wearing observation's clothes and is REJECTED as this gate's
// evidence, usable only as a supplementary fast sanity check. The parent
// asserts on the child's own EXIT STATUS and STDOUT/STDERR, never on an
// in-process value.
//
// Two arms, both required: cap_disabled proves the language-level call
// stack survives probeChainDepth (many multiples past MaxCallDepth) under a
// pinned small host stack ceiling with no host collapse; cap_enabled proves
// the named depth-exceeded refusal fires FIRST, before that same host
// ceiling is ever approached, at the SAME chain depth. See
// runCallDepthProbeChild's own doc comment for the full framing and its
// recorded observation for Assumption A1.
// --- Plan 10-05: cross-frame drain ordering (SEM-09 clause 2, D-10-33) ---

// drainByteType is the shared, ability-free (non-copyable) TypeFact every
// Task 1 drain-order fixture below uses: none of these fixtures depend on
// OWN-05b's copy-vs-move distinction, so a single non-copyable type keeps
// every builder uniform.
var drainByteType = core.TypeFact{ID: "drain:type:byte"}

// newDrainAcquireReleaseBlocks builds the two-block shape every synthetic
// resource-lifecycle function below shares: an entry block with exactly one
// OpForeignCall acquisition (never itself a terminator: the ok edge always
// leads to a second block), and a second block holding whatever operations
// the caller supplies (a release+return, or a call to another function, or
// nothing at all before a defect fires). The acquisition is always made
// TRACKED (D-04-07's convention) by a same-named, always-present OpRelease
// operation appended to the function's own Operations list -- present for
// newBlockFrame's tracked-map scan (which walks ALL declared operations,
// never only the reachable ones) even when that release operation itself
// never executes, mirroring cgen's own newResourceLedger scan.
func newDrainAcquireReleaseBlocks(functionID, paramID, acquireTargetID string, afterAcquireOps []string, extraOperations []core.LinearOperation) *core.LinearBody {
	acquireOp := functionID + ":op:acquire"
	releaseOp := functionID + ":op:release_for_tracking"
	entryBlock := functionID + ":block:entry"
	afterBlock := functionID + ":block:after_acquire"
	okEdge := functionID + ":edge:ok"

	operations := []core.LinearOperation{
		{ID: acquireOp, Kind: core.OpForeignCall, SourceID: paramID, TargetID: acquireTargetID, ErrTargetID: functionID + ":place:err", TypeID: drainByteType.ID, OkEdgeID: okEdge},
	}
	operations = append(operations, extraOperations...)
	// The tracking release: appended to Operations (so newBlockFrame's scan
	// sees it and marks acquireOp live) but deliberately NOT referenced by
	// any block's OperationIDs -- it never actually executes on any path
	// these fixtures drive, exactly as PHASE-10-DEBT.md's D-10-31 finding
	// requires: only the two named abrupt constructs ever observe this
	// acquisition's fate here, never a normal release.
	operations = append(operations, core.LinearOperation{
		ID: releaseOp, Kind: core.OpRelease, SourceID: acquireTargetID, TypeID: drainByteType.ID, ReleasesOperationID: acquireOp,
	})

	return &core.LinearBody{
		ID:    functionID + ":linear",
		Types: []core.TypeFact{drainByteType},
		Places: []core.Place{
			{ID: paramID, Name: "value", TypeID: drainByteType.ID},
			{ID: acquireTargetID, Name: "acquired", TypeID: drainByteType.ID},
		},
		Operations: operations,
		Blocks: []core.Block{
			{ID: entryBlock, OperationIDs: []string{acquireOp}, Successors: []string{afterBlock}},
			{ID: afterBlock, OperationIDs: afterAcquireOps},
		},
		Edges: []core.Edge{{ID: okEdge, FromBlockID: entryBlock, ToBlockID: afterBlock}},
	}
}

// frameDrainNormalReturnProgram hand-builds a synthetic two-function
// core.Program (never fed through the parser or corevalidate, following
// moveAsCopyProbeProgram's own established precedent for isolating interp's
// mechanism directly): "caller" acquires its own resource, calls "callee"
// (which acquires and releases ITS OWN resource before returning normally),
// then releases its own resource and returns. On normal return, the
// callee's frame is drained (its own release event emitted) entirely before
// it pops -- and therefore entirely before the caller's own release event,
// which can only happen after control resumes in the caller's frame. This
// composes for free (no interp code change required for this path): it is
// simply the natural consequence of frames executing their own operations,
// in order, before the terminator that pops them.
func frameDrainNormalReturnProgram() (program core.Program, entry core.Function) {
	calleeID := "drain:normal:callee"
	calleeParam := calleeID + ":place:param"
	calleeAcquired := calleeID + ":place:acquired"
	calleeReturn := calleeID + ":op:return"
	callee := core.Function{
		ID: calleeID, Name: "callee", Parameter: core.Parameter{ID: calleeParam, Name: "value", Type: "Byte"}, ReturnType: "Byte",
		Linear: newDrainAcquireReleaseBlocks(calleeID, calleeParam, calleeAcquired,
			[]string{calleeID + ":op:real_release", calleeReturn}, nil),
	}
	// Append the REAL (reachable) release+return pair into the "after
	// acquire" block -- distinct from newDrainAcquireReleaseBlocks' own
	// always-unreached tracking release, this one genuinely executes.
	callee.Linear.Operations = append(callee.Linear.Operations,
		core.LinearOperation{ID: calleeID + ":op:real_release", Kind: core.OpRelease, SourceID: calleeAcquired, TypeID: drainByteType.ID, ReleasesOperationID: calleeID + ":op:acquire"},
		core.LinearOperation{ID: calleeReturn, Kind: core.OpReturn, SourceID: calleeParam, TypeID: drainByteType.ID},
	)

	callerID := "drain:normal:caller"
	callerParam := callerID + ":place:param"
	callerAcquired := callerID + ":place:acquired"
	callResult := callerID + ":place:call_result"
	callOp := callerID + ":op:call"
	callerRelease := callerID + ":op:real_release"
	callerReturn := callerID + ":op:return"
	caller := core.Function{
		ID: callerID, Name: "caller", Parameter: core.Parameter{ID: callerParam, Name: "value", Type: "Byte"}, ReturnType: "Byte",
		Linear: newDrainAcquireReleaseBlocks(callerID, callerParam, callerAcquired,
			[]string{callOp, callerRelease, callerReturn}, nil),
	}
	caller.Linear.Places = append(caller.Linear.Places, core.Place{ID: callResult, Name: "call_result", TypeID: drainByteType.ID})
	caller.Linear.Operations = append(caller.Linear.Operations,
		core.LinearOperation{ID: callOp, Kind: core.OpCall, SourceID: callerParam, TargetID: callResult, TypeID: drainByteType.ID, CalleeID: calleeID},
		core.LinearOperation{ID: callerRelease, Kind: core.OpRelease, SourceID: callerAcquired, TypeID: drainByteType.ID, ReleasesOperationID: callerID + ":op:acquire"},
		core.LinearOperation{ID: callerReturn, Kind: core.OpReturn, SourceID: callResult, TypeID: drainByteType.ID},
	)

	return core.Program{Schema: core.Schema1, Module: "drain.normal_return", Functions: []core.Function{callee, caller}}, caller
}

// frameDrainSingleFrameBaselineProgram is the pre-phase single-frame shape
// (no call at all): one function that acquires and releases its own
// resource before returning. Used to prove drop ordering at a single frame
// is byte-identical whether or not the multi-frame drain machinery exists
// at all -- composing frames must change nothing about within-frame
// ordering.
func frameDrainSingleFrameBaselineProgram() (program core.Program, entry core.Function) {
	id := "drain:baseline:fn"
	param := id + ":place:param"
	acquired := id + ":place:acquired"
	releaseOp := id + ":op:real_release"
	returnOp := id + ":op:return"
	fn := core.Function{
		ID: id, Name: "baseline", Parameter: core.Parameter{ID: param, Name: "value", Type: "Byte"}, ReturnType: "Byte",
		Linear: newDrainAcquireReleaseBlocks(id, param, acquired, []string{releaseOp, returnOp}, nil),
	}
	fn.Linear.Operations = append(fn.Linear.Operations,
		core.LinearOperation{ID: releaseOp, Kind: core.OpRelease, SourceID: acquired, TypeID: drainByteType.ID, ReleasesOperationID: id + ":op:acquire"},
		core.LinearOperation{ID: returnOp, Kind: core.OpReturn, SourceID: param, TypeID: drainByteType.ID},
	)
	return core.Program{Schema: core.Schema1, Module: "drain.single_frame_baseline", Functions: []core.Function{fn}}, fn
}

// frameDrainZeroAcquisitionCalleeProgram: "caller" acquires its own
// resource, calls a callee that acquires NOTHING and returns immediately
// (a flat body, no blocks at all), then releases its own resource and
// returns. Proves a zero-acquisition callee frame pops emitting zero drop
// events and leaves the caller's own drop sequence unchanged -- and, since
// the callee acquires nothing, that its own popped Execution-shape
// (exercised indirectly here through the caller's own event stream) never
// reports a null live_resources.
func frameDrainZeroAcquisitionCalleeProgram() (program core.Program, entry core.Function) {
	calleeID := "drain:zero:callee"
	calleeParam := calleeID + ":place:param"
	callee := core.Function{
		ID: calleeID, Name: "callee", Parameter: core.Parameter{ID: calleeParam, Name: "value", Type: "Byte"}, ReturnType: "Byte",
		Linear: &core.LinearBody{
			ID:     calleeID + ":linear",
			Types:  []core.TypeFact{drainByteType},
			Places: []core.Place{{ID: calleeParam, Name: "value", TypeID: drainByteType.ID}},
			Operations: []core.LinearOperation{
				{ID: calleeID + ":op:return", Kind: core.OpReturn, SourceID: calleeParam, TypeID: drainByteType.ID},
			},
		},
	}

	callerID := "drain:zero:caller"
	callerParam := callerID + ":place:param"
	callerAcquired := callerID + ":place:acquired"
	callResult := callerID + ":place:call_result"
	callOp := callerID + ":op:call"
	callerRelease := callerID + ":op:real_release"
	callerReturn := callerID + ":op:return"
	caller := core.Function{
		ID: callerID, Name: "caller", Parameter: core.Parameter{ID: callerParam, Name: "value", Type: "Byte"}, ReturnType: "Byte",
		Linear: newDrainAcquireReleaseBlocks(callerID, callerParam, callerAcquired,
			[]string{callOp, callerRelease, callerReturn}, nil),
	}
	caller.Linear.Places = append(caller.Linear.Places, core.Place{ID: callResult, Name: "call_result", TypeID: drainByteType.ID})
	caller.Linear.Operations = append(caller.Linear.Operations,
		core.LinearOperation{ID: callOp, Kind: core.OpCall, SourceID: callerParam, TargetID: callResult, TypeID: drainByteType.ID, CalleeID: calleeID},
		core.LinearOperation{ID: callerRelease, Kind: core.OpRelease, SourceID: callerAcquired, TypeID: drainByteType.ID, ReleasesOperationID: callerID + ":op:acquire"},
		core.LinearOperation{ID: callerReturn, Kind: core.OpReturn, SourceID: callResult, TypeID: drainByteType.ID},
	)

	return core.Program{Schema: core.Schema1, Module: "drain.zero_acquisition_callee", Functions: []core.Function{callee, caller}}, caller
}

// frameDrainNoAcquisitionProgram is a single function with NO acquisitions
// at all -- proves live_resources serializes as [] (never null) for a
// function whose live-tracking map and liveOrder slice both stay empty for
// its entire execution.
func frameDrainNoAcquisitionProgram() (program core.Program, entry core.Function) {
	id := "drain:noacquire:fn"
	param := id + ":place:param"
	fn := core.Function{
		ID: id, Name: "noacquire", Parameter: core.Parameter{ID: param, Name: "value", Type: "Byte"}, ReturnType: "Byte",
		Linear: &core.LinearBody{
			ID:         id + ":linear",
			Types:      []core.TypeFact{drainByteType},
			Places:     []core.Place{{ID: param, Name: "value", TypeID: drainByteType.ID}},
			Operations: []core.LinearOperation{{ID: id + ":op:return", Kind: core.OpReturn, SourceID: param, TypeID: drainByteType.ID}},
		},
	}
	return core.Program{Schema: core.Schema1, Module: "drain.no_acquisition", Functions: []core.Function{fn}}, fn
}

// frameDrainNonlocalMultiFrameProgram is the ABRUPT-path fixture (D-10-31's
// foreign process-root landing pad, extended to N > 1 frames): "caller"
// acquires its own resource A and calls "callee"; callee's OWN
// ForeignContract declares nonlocal_exit "possible" and its body performs
// TWO OpForeignCall operations -- the first acquires callee's own resource
// B (tracked, live), the second is callee's own SECOND foreign call within
// its own frame, which triggers the process-root landing pad per this
// engine's existing single-frame convention (nonlocalExitCalls == 2). At
// that moment the stack holds exactly two live frames: callee (innermost,
// resource B still live) and caller (outermost, resource A still live --
// its own release, guarded behind the call, never executes because control
// never returns to it). This is the genuine N-frame nonlocal-exit shape
// PHASE-10-DEBT.md's D-10-31 names -- never only a single-frame exercise of
// OpForeignCall's own existing landing pad.
func frameDrainNonlocalMultiFrameProgram() (program core.Program, entry core.Function) {
	calleeID := "drain:nonlocal:callee"
	calleeParam := calleeID + ":place:param"
	calleeAcquired := calleeID + ":place:acquired"
	calleeTrigger := calleeID + ":op:trigger"
	entryBlock := calleeID + ":block:entry"
	afterBlock := calleeID + ":block:after_acquire"
	okEdge := calleeID + ":edge:ok"
	callee := core.Function{
		ID: calleeID, Name: "callee", Parameter: core.Parameter{ID: calleeParam, Name: "value", Type: "Byte"}, ReturnType: "Byte",
		ForeignContract: &core.ForeignContract{Symbol: "schway_nonlocal_probe", Allocator: "libc_malloc", NonlocalExit: "possible", Fails: "ProbeError"},
		Linear: &core.LinearBody{
			ID:    calleeID + ":linear",
			Types: []core.TypeFact{drainByteType},
			Places: []core.Place{
				{ID: calleeParam, Name: "value", TypeID: drainByteType.ID},
				{ID: calleeAcquired, Name: "acquired", TypeID: drainByteType.ID},
			},
			Operations: []core.LinearOperation{
				{ID: calleeID + ":op:acquire", Kind: core.OpForeignCall, SourceID: calleeParam, TargetID: calleeAcquired, ErrTargetID: calleeID + ":place:err", TypeID: drainByteType.ID, OkEdgeID: okEdge},
				// The trigger: callee's OWN second OpForeignCall. Its
				// OkEdgeID/ErrTargetID are never consulted -- the
				// nonlocalExitCalls==2 branch returns before reaching that
				// logic (interp.go's runFrameStack), exactly like
				// nonlocal_exit_probe.schway's own second `try` call.
				{ID: calleeTrigger, Kind: core.OpForeignCall, SourceID: calleeParam, TargetID: calleeID + ":place:trigger_unused", TypeID: drainByteType.ID},
				// Tracking release for callee's own acquisition, unreached
				// on this path exactly like newDrainAcquireReleaseBlocks'
				// own convention.
				{ID: calleeID + ":op:release_for_tracking", Kind: core.OpRelease, SourceID: calleeAcquired, TypeID: drainByteType.ID, ReleasesOperationID: calleeID + ":op:acquire"},
			},
			Blocks: []core.Block{
				{ID: entryBlock, OperationIDs: []string{calleeID + ":op:acquire"}, Successors: []string{afterBlock}},
				{ID: afterBlock, OperationIDs: []string{calleeTrigger}},
			},
			Edges: []core.Edge{{ID: okEdge, FromBlockID: entryBlock, ToBlockID: afterBlock}},
		},
	}

	callerID := "drain:nonlocal:caller"
	callerParam := callerID + ":place:param"
	callerAcquired := callerID + ":place:acquired"
	callResult := callerID + ":place:call_result"
	callOp := callerID + ":op:call"
	caller := core.Function{
		ID: callerID, Name: "caller", Parameter: core.Parameter{ID: callerParam, Name: "value", Type: "Byte"}, ReturnType: "Byte",
		Linear: newDrainAcquireReleaseBlocks(callerID, callerParam, callerAcquired, []string{callOp}, nil),
	}
	caller.Linear.Places = append(caller.Linear.Places, core.Place{ID: callResult, Name: "call_result", TypeID: drainByteType.ID})
	caller.Linear.Operations = append(caller.Linear.Operations,
		core.LinearOperation{ID: callOp, Kind: core.OpCall, SourceID: callerParam, TargetID: callResult, TypeID: drainByteType.ID, CalleeID: calleeID},
	)

	return core.Program{Schema: core.Schema1, Module: "drain.nonlocal_multi_frame", Functions: []core.Function{callee, caller}}, caller
}

// frameDrainDepthRefusalMultiFrameProgram is the SECOND abrupt-path fixture
// (the SEM-08 depth refusal, D-10-31's second named construct): a synthetic
// three-function chain, each acquiring its OWN resource before calling the
// next. Driven with maxCallDepthOverride pinned to 2 (restored by the
// caller via defer), pushing the third frame exceeds the cap: at that
// moment the stack holds exactly two live frames (fn0 outermost, fn1
// innermost), each with its own still-live resource, and fn2 is never
// reached at all. This proves the SAME LIFO flattening applies on this
// construct as on the nonlocal landing pad, independently of it.
func frameDrainDepthRefusalMultiFrameProgram() (program core.Program, entry core.Function) {
	build := func(id, calleeID string) core.Function {
		param := id + ":place:param"
		acquired := id + ":place:acquired"
		callOp := id + ":op:call"
		fn := core.Function{
			ID: id, Name: id, Parameter: core.Parameter{ID: param, Name: "value", Type: "Byte"}, ReturnType: "Byte",
			Linear: newDrainAcquireReleaseBlocks(id, param, acquired, []string{callOp}, nil),
		}
		if calleeID != "" {
			callResult := id + ":place:call_result"
			fn.Linear.Places = append(fn.Linear.Places, core.Place{ID: callResult, Name: "call_result", TypeID: drainByteType.ID})
			fn.Linear.Operations = append(fn.Linear.Operations,
				core.LinearOperation{ID: callOp, Kind: core.OpCall, SourceID: param, TargetID: callResult, TypeID: drainByteType.ID, CalleeID: calleeID},
			)
		} else {
			// fn2 is never reached (the cap refuses before its own push),
			// but it must still be a structurally closed function.
			returnOp := id + ":op:return"
			fn.Linear.Blocks[1].OperationIDs = []string{returnOp}
			fn.Linear.Operations = append(fn.Linear.Operations,
				core.LinearOperation{ID: returnOp, Kind: core.OpReturn, SourceID: param, TypeID: drainByteType.ID},
			)
		}
		return fn
	}

	fn2 := build("drain:depth:fn2", "")
	fn1 := build("drain:depth:fn1", fn2.ID)
	fn0 := build("drain:depth:fn0", fn1.ID)

	return core.Program{Schema: core.Schema1, Module: "drain.depth_refusal_multi_frame", Functions: []core.Function{fn0, fn1, fn2}}, fn0
}

// runDrainProgram runs a builder's entry function directly through
// runFrameStack (bypassing Run's own corevalidate.Validate gate, following
// moveAsCopyProbeProgram's established precedent): every fixture above is a
// deliberately synthetic core.Program these Task 1 tests use to isolate
// interp's OWN cross-frame drain mechanism from check/corevalidate/parse.
func runDrainProgram(program core.Program, entry core.Function, input string) (Execution, error) {
	values := map[string]value{entry.Parameter.ID: {tag: "", payload: input}}
	var base frame
	if len(entry.Linear.Blocks) == 0 {
		base = newFlatFrame(entry, values)
	} else {
		base = newBlockFrame(entry, values, entry.Linear.Blocks[0].ID)
	}
	return runFrameStack(program, base)
}

// eventKindSequence extracts the ordered Kind sequence from an Execution's
// Events -- the shape every ordering assertion below compares, exactly as
// this plan's must_haves require: NEVER a frame field or a live map read
// directly, only interp.CanonicalBytes / the ordered Events slice itself.
func eventKindSequence(result Execution) []string {
	kinds := make([]string, len(result.Events))
	for i, event := range result.Events {
		kinds[i] = event.Kind + ":" + event.FunctionID
	}
	return kinds
}

// indexOfEvent finds functionID/kind's own event and returns its BYTE OFFSET
// within interp.CanonicalBytes' serialized Events array -- never an index
// into the Events slice directly. JSON array element order is byte order, so
// comparing two such offsets is exactly "comparing interp.CanonicalBytes
// output" (this plan's own mandatory observation channel), never a read of
// result.Events' own slice position (which, while not a `frame`-internal
// field, this helper avoids entirely to keep every ordering assertion on
// the single sanctioned channel).
func indexOfEvent(t *testing.T, result Execution, functionID, kind string) int {
	t.Helper()
	var id string
	for _, event := range result.Events {
		if event.FunctionID == functionID && event.Kind == kind {
			id = event.ID
			break
		}
	}
	if id == "" {
		t.Fatalf("expected an event with FunctionID %q and Kind %q, got: %+v", functionID, kind, eventKindSequence(result))
	}
	encoded, err := CanonicalBytes(result)
	if err != nil {
		t.Fatalf("CanonicalBytes: %v", err)
	}
	offset := bytes.Index(encoded, []byte(`"id":"`+id+`"`))
	if offset < 0 {
		t.Fatalf("expected canonical bytes to contain event id %q, got: %s", id, encoded)
	}
	return offset
}

// assertCanonicalOrder asserts that eventIDs appear, in the GIVEN order, as
// strictly increasing byte offsets inside interp.CanonicalBytes' own
// serialized output -- the sanctioned observation channel for every
// ordering assertion in this test (never a read of result.Events' own slice
// position, and never a frame-internal field or live map).
func assertCanonicalOrder(t *testing.T, result Execution, eventIDs ...string) {
	t.Helper()
	encoded, err := CanonicalBytes(result)
	if err != nil {
		t.Fatalf("CanonicalBytes: %v", err)
	}
	previousOffset := -1
	previousID := ""
	for _, id := range eventIDs {
		offset := bytes.Index(encoded, []byte(`"id":"`+id+`"`))
		if offset < 0 {
			t.Fatalf("expected canonical bytes to contain event id %q, got: %s", id, encoded)
		}
		if offset <= previousOffset {
			t.Fatalf("expected event id %q to follow %q in canonical bytes order, got offsets %d then %d: %s", id, previousID, previousOffset, offset, encoded)
		}
		previousOffset, previousID = offset, id
	}
}

// TestFrameDrainOrder is Plan 10-05's Task 1 test (SEM-09 clause 2,
// D-10-33/D-10-35): drop and cleanup obligations run in a DEFINED order
// across frames, on both the normal-return and the abrupt path, observed
// exclusively through interp.CanonicalBytes / the Events slice -- never
// against a frame's own live map or values map.
func TestFrameDrainOrder(t *testing.T) {
	t.Run("normal_return_across_one_boundary", func(t *testing.T) {
		program, entry := frameDrainNormalReturnProgram()
		result, err := runDrainProgram(program, entry, "7")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Outcome.Kind != "returned" {
			t.Fatalf("expected outcome kind %q, got %q: %+v", "returned", result.Outcome.Kind, result)
		}
		calleeReleaseIdx := indexOfEvent(t, result, "drain:normal:callee", "resource.released")
		callerReleaseIdx := indexOfEvent(t, result, "drain:normal:caller", "resource.released")
		if calleeReleaseIdx >= callerReleaseIdx {
			t.Fatalf("expected the callee's own drop event to precede the caller's first drop event, got sequence: %v", eventKindSequence(result))
		}
	})

	t.Run("single_frame_baseline_equivalence", func(t *testing.T) {
		program, entry := frameDrainSingleFrameBaselineProgram()
		result, err := runDrainProgram(program, entry, "7")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Events) != 3 {
			t.Fatalf("expected exactly 3 events (pre-phase single-frame baseline), got %d: %v", len(result.Events), eventKindSequence(result))
		}
		assertCanonicalOrder(t, result,
			"drain:baseline:fn:op:acquire:event",
			"drain:baseline:fn:op:real_release:event",
			"drain:baseline:fn:op:return:event:returned",
		)
	})

	t.Run("zero_acquisition_callee_frame", func(t *testing.T) {
		program, entry := frameDrainZeroAcquisitionCalleeProgram()
		result, err := runDrainProgram(program, entry, "7")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, event := range result.Events {
			if event.FunctionID == "drain:zero:callee" && (event.Kind == "resource.released" || event.Kind == "resource.leaked") {
				t.Fatalf("expected the zero-acquisition callee frame to emit zero DROP events, got: %+v", event)
			}
		}
		if len(result.Events) != 4 {
			t.Fatalf("expected exactly 4 events (caller's own sequence unreordered/unmerged), got %d: %v", len(result.Events), eventKindSequence(result))
		}
		assertCanonicalOrder(t, result,
			"drain:zero:caller:op:acquire:event",
			"drain:zero:callee:op:return:event:returned",
			"drain:zero:caller:op:real_release:event",
			"drain:zero:caller:op:return:event:returned",
		)
	})

	t.Run("live_resources_empty_array_never_null", func(t *testing.T) {
		program, entry := frameDrainNoAcquisitionProgram()
		result, err := runDrainProgram(program, entry, "7")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.LiveResources == nil {
			t.Fatalf("expected a non-nil, empty LiveResources slice, got nil")
		}
		if len(result.LiveResources) != 0 {
			t.Fatalf("expected zero live resources, got: %v", result.LiveResources)
		}
		encoded, err := CanonicalBytes(result)
		if err != nil {
			t.Fatalf("CanonicalBytes: %v", err)
		}
		if !bytes.Contains(encoded, []byte(`"live_resources":[]`)) {
			t.Fatalf("expected CanonicalBytes to serialize live_resources as [], never null, got: %s", encoded)
		}
		if bytes.Contains(encoded, []byte(`"live_resources":null`)) {
			t.Fatalf("live_resources serialized as null: %s", encoded)
		}
	})

	t.Run("foreign_nonlocal_landing_pad_across_multiple_frames", func(t *testing.T) {
		program, entry := frameDrainNonlocalMultiFrameProgram()
		result, err := runDrainProgram(program, entry, "7")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Outcome.Kind != execution.OutcomeDefect {
			t.Fatalf("expected outcome kind %q, got %q: %+v", execution.OutcomeDefect, result.Outcome.Kind, result)
		}
		// LIFO flattening: innermost (callee, resource B) leaked first,
		// then outward to the caller's own resource A -- proven against
		// the ordered Events slice, never against frame.live directly.
		calleeLeakIdx := indexOfEvent(t, result, "drain:nonlocal:callee", "resource.leaked")
		callerLeakIdx := indexOfEvent(t, result, "drain:nonlocal:caller", "resource.leaked")
		if calleeLeakIdx >= callerLeakIdx {
			t.Fatalf("expected the innermost frame's (callee) leak event to precede the outer frame's (caller), got sequence: %v", eventKindSequence(result))
		}
		wantLive := []string{"drain:nonlocal:callee:place:acquired", "drain:nonlocal:caller:place:acquired"}
		if len(result.LiveResources) != len(wantLive) {
			t.Fatalf("expected live_resources %v, got %v", wantLive, result.LiveResources)
		}
		for i := range wantLive {
			if result.LiveResources[i] != wantLive[i] {
				t.Fatalf("expected live_resources[%d] = %q (innermost first), got %q (full: %v)", i, wantLive[i], result.LiveResources[i], result.LiveResources)
			}
		}
	})

	t.Run("depth_refusal_with_live_resources_in_multiple_frames", func(t *testing.T) {
		previous := maxCallDepthOverride
		maxCallDepthOverride = func() int { return 2 }
		defer func() { maxCallDepthOverride = previous }()

		program, entry := frameDrainDepthRefusalMultiFrameProgram()
		result, err := runDrainProgram(program, entry, "7")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Outcome.Kind != execution.OutcomeDefect {
			t.Fatalf("expected outcome kind %q, got %q: %+v", execution.OutcomeDefect, result.Outcome.Kind, result)
		}
		found := false
		for _, event := range result.Events {
			if event.Output == callDepthExceededDefectReason {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected an event carrying the named depth-limit reason %q, got: %v", callDepthExceededDefectReason, eventKindSequence(result))
		}
		fn1LeakIdx := indexOfEvent(t, result, "drain:depth:fn1", "resource.leaked")
		fn0LeakIdx := indexOfEvent(t, result, "drain:depth:fn0", "resource.leaked")
		if fn1LeakIdx >= fn0LeakIdx {
			t.Fatalf("expected the innermost frame's (fn1) leak event to precede the outer frame's (fn0), got sequence: %v", eventKindSequence(result))
		}
		wantLive := []string{"drain:depth:fn1:place:acquired", "drain:depth:fn0:place:acquired"}
		if len(result.LiveResources) != len(wantLive) {
			t.Fatalf("expected live_resources %v, got %v", wantLive, result.LiveResources)
		}
		for i := range wantLive {
			if result.LiveResources[i] != wantLive[i] {
				t.Fatalf("expected live_resources[%d] = %q (innermost first), got %q (full: %v)", i, wantLive[i], result.LiveResources[i], result.LiveResources)
			}
		}
	})

	t.Run("frame_drain_order_seam_flips_canonical_bytes", func(t *testing.T) {
		program, entry := frameDrainNonlocalMultiFrameProgram()

		clean, err := runDrainProgram(program, entry, "7")
		if err != nil {
			t.Fatalf("clean run: unexpected error: %v", err)
		}
		cleanBytes, err := CanonicalBytes(clean)
		if err != nil {
			t.Fatalf("CanonicalBytes (clean): %v", err)
		}

		previous := frameDrainOrderForTest
		frameDrainOrderForTest = func() bool { return true }
		defer func() { frameDrainOrderForTest = previous }()

		flipped, err := runDrainProgram(program, entry, "7")
		if err != nil {
			t.Fatalf("flipped run: unexpected error: %v", err)
		}
		flippedBytes, err := CanonicalBytes(flipped)
		if err != nil {
			t.Fatalf("CanonicalBytes (flipped): %v", err)
		}

		if bytes.Equal(cleanBytes, flippedBytes) {
			t.Fatalf("expected frameDrainOrderForTest to change interp's canonical bytes, but clean and flipped runs matched: %s", cleanBytes)
		}
		if execution.Equal(clean, flipped) {
			t.Fatalf("expected execution.Equal to also observe the flipped drain order as a genuine divergence")
		}
	})
}

// TestInterpDeterministicAcrossRuns is Task 1's determinism assertion
// (D-10-26/D-10-57 clause 3): repeated runs of the SAME multi-frame program
// must produce byte-identical CanonicalBytes, proving drop ordering derives
// from ordered slices (the frame stack itself, each frame's own liveOrder)
// rather than Go map iteration order. Looped internally (25x) so a single
// process invocation already exercises Go's per-range map-iteration
// randomization repeatedly; run at `-count=10` externally (per this task's
// own <verify>) multiplies that further across separate process runs.
func TestInterpDeterministicAcrossRuns(t *testing.T) {
	program, entry := frameDrainNonlocalMultiFrameProgram()

	var first []byte
	for i := 0; i < 25; i++ {
		result, err := runDrainProgram(program, entry, "7")
		if err != nil {
			t.Fatalf("run %d: unexpected error: %v", i, err)
		}
		encoded, err := CanonicalBytes(result)
		if err != nil {
			t.Fatalf("run %d: CanonicalBytes: %v", i, err)
		}
		if i == 0 {
			first = encoded
			continue
		}
		if !bytes.Equal(first, encoded) {
			t.Fatalf("run %d diverged from run 0's canonical bytes:\nrun 0: %s\nrun %d: %s", i, first, i, encoded)
		}
	}
}

func TestNativeStackHeadroomIndependentOfCallDepth(t *testing.T) {
	if os.Getenv(probeChildEnv) != "" {
		runCallDepthProbeChild(t)
		return
	}

	t.Run(probeArmCapDisabled, func(t *testing.T) {
		stdout, stderr, err := runCallDepthProbeSubprocess(t, probeArmCapDisabled)
		if err != nil {
			t.Fatalf("cap-disabled child exited with error: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}
		if !bytes.Contains(stdout, []byte("cap_disabled: ok")) {
			t.Fatalf("expected the cap-disabled child to report success on its own stdout; stdout: %s\nstderr: %s", stdout, stderr)
		}
	})

	t.Run(probeArmCapEnabled, func(t *testing.T) {
		stdout, stderr, err := runCallDepthProbeSubprocess(t, probeArmCapEnabled)
		if err != nil {
			t.Fatalf("cap-enabled child exited with error: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}
		if !bytes.Contains(stdout, []byte("cap_enabled: ok")) {
			t.Fatalf("expected the cap-enabled child to report the depth refusal fired first on its own stdout; stdout: %s\nstderr: %s", stdout, stderr)
		}
	})
}

// --- Plan 10-09 Task 3: closing the structural coverage floor's real gaps ---

// oracleBorrowSequenceType is the shared TypeFact oracleBorrowSequenceProgram
// uses: interp performs no ability check of its own (that is
// corevalidate's/check's job), so a single bare TypeFact is enough to
// exercise core.OpBorrowShared/core.OpBorrowExclusive/core.OpCopy/
// core.OpMove uniformly.
var oracleBorrowSequenceType = core.TypeFact{ID: "oracle:coverage:type:value"}

// oracleBorrowSequenceProgram hand-builds a synthetic single-function
// core.Program (never fed through the parser, following
// moveAsCopyProbeProgram's own established precedent) exercising, in order,
// core.OpBorrowShared, core.OpBorrowExclusive, core.OpCopy, and core.OpMove
// before a terminal core.OpReturn. shape selects which of interp's three
// body shapes wraps this SAME operation sequence: "flat" (runLinear, no
// Blocks at all), "block" (runLinearBlocks' own two-block-free single-block
// shape, entered via newBlockFrame), or "arm" (runBranchArm's shape, the
// identical block marked singleBlockOnly via newArmFrame). This closes the
// Operation-kind coverage matrix's real gaps (D-10-58): before this test,
// core.OpBorrowShared and core.OpBorrowExclusive were never exercised by
// ANY interp test at ANY path, and core.OpCopy/core.OpMove were each
// exercised at only one of the three paths (defect_terminal.schway's "Go" arm,
// runBranchArm only; TestMoveAsCopyMutationKilled's mutated run, runLinear
// only) rather than at every path a structural floor requires.
func oracleBorrowSequenceProgram(shape string) (program core.Program, entry core.Function, blockID string) {
	id := "oracle:coverage:" + shape
	param := id + ":place:param"
	sharedTarget := id + ":place:shared"
	exclusiveTarget := id + ":place:exclusive"
	copyTarget := id + ":place:copy"
	moveTarget := id + ":place:move"

	operations := []core.LinearOperation{
		{ID: id + ":op:0", Kind: core.OpBorrowShared, SourceID: param, TargetID: sharedTarget, TypeID: oracleBorrowSequenceType.ID},
		{ID: id + ":op:1", Kind: core.OpBorrowExclusive, SourceID: param, TargetID: exclusiveTarget, TypeID: oracleBorrowSequenceType.ID},
		{ID: id + ":op:2", Kind: core.OpCopy, SourceID: sharedTarget, TargetID: copyTarget, TypeID: oracleBorrowSequenceType.ID},
		{ID: id + ":op:3", Kind: core.OpMove, SourceID: exclusiveTarget, TargetID: moveTarget, TypeID: oracleBorrowSequenceType.ID},
		{ID: id + ":op:4", Kind: core.OpReturn, SourceID: moveTarget, TypeID: oracleBorrowSequenceType.ID},
	}
	places := []core.Place{
		{ID: param, Name: "value", TypeID: oracleBorrowSequenceType.ID},
		{ID: sharedTarget, Name: "shared", TypeID: oracleBorrowSequenceType.ID},
		{ID: exclusiveTarget, Name: "exclusive", TypeID: oracleBorrowSequenceType.ID},
		{ID: copyTarget, Name: "copy", TypeID: oracleBorrowSequenceType.ID},
		{ID: moveTarget, Name: "move", TypeID: oracleBorrowSequenceType.ID},
	}
	linear := &core.LinearBody{ID: id + ":linear", Types: []core.TypeFact{oracleBorrowSequenceType}, Places: places}

	switch shape {
	case "flat":
		linear.Operations = operations
	case "block", "arm":
		blockID = id + ":block:entry"
		linear.Operations = operations
		linear.Blocks = []core.Block{{ID: blockID, OperationIDs: []string{
			id + ":op:0", id + ":op:1", id + ":op:2", id + ":op:3", id + ":op:4",
		}}}
	default:
		panic("oracleBorrowSequenceProgram: unknown shape " + shape)
	}

	entry = core.Function{ID: id + ":fn", Name: shape, Parameter: core.Parameter{ID: param, Name: "value", Type: "Byte"}, ReturnType: "Byte", Linear: linear}
	return core.Program{Schema: core.Schema1, Module: "oracle.coverage." + shape, Functions: []core.Function{entry}}, entry, blockID
}

// TestOperationKindCoverageAcrossAllThreePaths is Task 3's structural
// coverage-floor closer (D-10-58): drives the IDENTICAL
// OpBorrowShared/OpBorrowExclusive/OpCopy/OpMove/OpReturn sequence through
// all three of interp's own body shapes -- flat (runLinear), block-based
// (runLinearBlocks), and match-arm (runBranchArm) -- proving every one of
// these five operation kinds is genuinely reachable and correctly handled
// at every path, not merely at whichever single path happened to exercise
// it first.
func TestOperationKindCoverageAcrossAllThreePaths(t *testing.T) {
	for _, shape := range []string{"flat", "block", "arm"} {
		shape := shape
		t.Run(shape, func(t *testing.T) {
			program, entry, blockID := oracleBorrowSequenceProgram(shape)
			values := map[string]value{entry.Parameter.ID: {tag: "", payload: "V"}}
			var base frame
			switch shape {
			case "flat":
				base = newFlatFrame(entry, values)
			case "block":
				base = newBlockFrame(entry, values, blockID)
			case "arm":
				base = newArmFrame(entry, values, blockID)
			}
			result, err := runFrameStack(program, base)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", shape, err)
			}
			if result.Outcome.Kind != "returned" || result.Outcome.Value != "V" {
				t.Fatalf("%s: expected outcome {returned, V}, got %+v", shape, result.Outcome)
			}
			wantKinds := []string{
				"value.borrowed", "value.borrowed_exclusive", "value.copied", "value.transferred", "function.returned",
			}
			if len(result.Events) != len(wantKinds) {
				t.Fatalf("%s: expected exactly %d events (one per non-terminal operation), got %d: %v", shape, len(wantKinds), len(result.Events), eventKindSequence(result))
			}
			for i, kind := range wantKinds {
				if result.Events[i].Kind != kind {
					t.Fatalf("%s: expected event[%d].Kind = %q, got %q (full sequence: %v)", shape, i, kind, result.Events[i].Kind, eventKindSequence(result))
				}
			}
		})
	}
}

// --- Plan 10-09 Task 3: closing the refusal-path table's untested rows ---

// TestRunRefusesCorevalidateInvalidProgram is the refusal-path table's
// "corevalidate precondition failure" row: Run's own first act
// (corevalidate.Validate) refuses an empty core.Program (no declared
// Schema) before ever attempting to find or execute a function, reported
// as a plain Go error naming corevalidate's own problem code -- this guard
// was previously untested at the interp layer, relying entirely on
// corevalidate's own test suite to prove the underlying check works.
func TestRunRefusesCorevalidateInvalidProgram(t *testing.T) {
	_, err := Run(core.Program{}, "main", "x")
	if err == nil {
		t.Fatal("expected Run to refuse an empty, corevalidate-invalid core.Program, got success")
	}
	if !strings.Contains(err.Error(), "core validation failed") {
		t.Fatalf("expected an error naming corevalidate's own refusal, got: %v", err)
	}
}

// TestRunRefusesAbsentFunction is the refusal-path table's "absent
// function" row: Run refuses a functionName absent from the checked
// program's own declared Functions, a lookup-failure guard distinct from
// every operation-level refusal runFrameStack itself can produce.
func TestRunRefusesAbsentFunction(t *testing.T) {
	program := checkedCallBasicProgram(t)
	_, err := Run(program, "does_not_exist", "x")
	if err == nil {
		t.Fatal("expected Run to refuse an absent function name, got success")
	}
	if !strings.Contains(err.Error(), "is absent from checked core") {
		t.Fatalf("expected an error naming the absent-function refusal, got: %v", err)
	}
}

// TestRunRefusesInvalidBodyUnion documents the refusal-path table's
// "invalid body union" row and its own finding: Run's own
// `!function.HasClosedBody()` guard is PROVABLY UNREACHABLE via Run() for
// any input, because Run's own first act (corevalidate.Validate,
// unconditional, called before the guard is ever reached) already refuses
// a function with neither Linear nor Match set as its own independent
// structural code core.invalid_body -- confirmed here directly rather than
// assumed. Run's own guard is defense-in-depth: fail-closed redundancy
// against a hypothetically-differently-checked corevalidate.Result, never
// a path this test (or any other) can drive through Run itself. A
// function with neither Linear nor Match still resolves by name and
// reports HasClosedBody() == false, proving the shape is real; only the
// ORDER Run calls its two checks in (corevalidate first, always) makes the
// second one unreachable.
func TestRunRefusesInvalidBodyUnion(t *testing.T) {
	program := checkedCallBasicProgram(t)
	for i, function := range program.Functions {
		if function.Name == "main" {
			program.Functions[i].Linear = nil
			program.Functions[i].Match = nil
		}
	}
	function, ok := findFunction(program, "main")
	if !ok {
		t.Fatal("expected main to still resolve by name after clearing its body union")
	}
	if function.HasClosedBody() {
		t.Fatal("expected a function with neither Linear nor Match set to report HasClosedBody() == false")
	}
	_, err := Run(program, "main", "7")
	if err == nil {
		t.Fatal("expected Run to refuse a program carrying an invalid body union, got success")
	}
	if !strings.Contains(err.Error(), "core validation failed") {
		t.Fatalf("expected corevalidate.Validate (Run's own first act) to catch this shape FIRST, reported as \"core validation failed\" -- proving Run's own HasClosedBody guard is unreachable in practice; got: %v", err)
	}
}

// TestRunFrameStackRefusesUnknownOperationKind is the refusal-path table's
// "unknown operation kind" row: runFrameStack's own switch default case
// refuses an operation whose Kind is not a member of
// core.AllOperationKinds() -- a synthetic, hand-built core.LinearOperation
// with a bogus Kind string, since no real .schway source or check.go emission
// path can produce one (check.go only ever emits from the closed
// core.OperationKind set).
func TestRunFrameStackRefusesUnknownOperationKind(t *testing.T) {
	id := "oracle:refusal:unknown_kind"
	param := id + ":place:param"
	fn := core.Function{
		ID: id + ":fn", Name: "bogus", Parameter: core.Parameter{ID: param, Name: "value", Type: "Byte"}, ReturnType: "Byte",
		Linear: &core.LinearBody{
			ID:     id + ":linear",
			Types:  []core.TypeFact{oracleBorrowSequenceType},
			Places: []core.Place{{ID: param, Name: "value", TypeID: oracleBorrowSequenceType.ID}},
			Operations: []core.LinearOperation{
				{ID: id + ":op:0", Kind: core.OperationKind("bogus_kind"), SourceID: param, TypeID: oracleBorrowSequenceType.ID},
			},
		},
	}
	base := newFlatFrame(fn, map[string]value{param: {tag: "", payload: "V"}})
	_, err := runFrameStack(core.Program{Schema: core.Schema1, Module: "oracle.refusal.unknown_kind", Functions: []core.Function{fn}}, base)
	if err == nil {
		t.Fatal("expected runFrameStack to refuse an operation with an unknown Kind, got success")
	}
	if !strings.Contains(err.Error(), "unknown kind") {
		t.Fatalf("expected an error naming the unknown-kind refusal, got: %v", err)
	}
}

// TestCallFromBothMatchArmsAcrossFrames closes the Operation-kind coverage
// matrix's remaining core.OpCall gap (D-10-58): testdata/phase07/
// call_from_both_match_arms.schway (D-07-28) was checked/corevalidated by
// earlier phases but never previously driven through interp.Run at all --
// core.OpCall was exercised at runLinear (call_basic.schway) and at a
// synthetic runLinear-shaped probe (moveAsCopyProbeProgram), but never at
// runBranchArm, where a match arm's own block itself contains the
// core.OpCall to another match-arm-bodied function. Both real arms are
// driven, proving the callee dispatch this arm's block reaches (itself
// match-bodied, D-10-26/partitionFrameForCall) executes identically to
// Run's own top-level match dispatch.
func TestCallFromBothMatchArmsAcrossFrames(t *testing.T) {
	program := checkedProgramFromFixture(t, "phase07", "call_from_both_match_arms.schway")
	for _, input := range []string{"A", "B"} {
		input := input
		t.Run(input, func(t *testing.T) {
			result, err := Run(program, "main", input)
			if err != nil {
				t.Fatalf("Run(main, %q): unexpected error: %v", input, err)
			}
			if result.Outcome.Kind != "returned" || result.Outcome.Value != input {
				t.Fatalf("expected outcome {returned, %s}, got %+v", input, result.Outcome)
			}
			foundHelperReturn := false
			for _, event := range result.Events {
				if event.Kind == "function.returned" && strings.HasSuffix(event.FunctionID, ":fn:helper") {
					foundHelperReturn = true
				}
			}
			if !foundHelperReturn {
				t.Fatalf("expected an event proving helper's own callee frame genuinely ran (a function.returned event attributed to helper's own FunctionID), got: %v", eventKindSequence(result))
			}
		})
	}
}

func TestPhase26CheckedAddInterpreter(t *testing.T) {
	const maximum = "18446744073709551615"
	source, err := os.ReadFile(filepath.Join(interpProjectRoot(), "examples", "phase26", "checked_add_overflow.schway"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("overflow witness parse diagnostics: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("overflow witness check diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("overflow witness core validation problems: %+v", validated.Problems)
	}
	for _, test := range []struct {
		input string
		kind  string
		value string
	}{{input: "0", kind: execution.OutcomeReturned, value: maximum}, {input: "1", kind: execution.OutcomeDefect, value: ""}} {
		result, err := Run(validated.Program(), "main", test.input)
		if err != nil {
			t.Fatalf("Run(main, %q): %v", test.input, err)
		}
		if result.Outcome.Kind != test.kind || result.Outcome.Value != test.value {
			t.Errorf("input %s outcome=%+v, want {%s, %q}", test.input, result.Outcome, test.kind, test.value)
		}
		if test.input == "1" {
			if len(result.Events) == 0 {
				t.Fatal("overflow returned no attributed defect event")
			}
			event := result.Events[len(result.Events)-1]
			if event.Kind != "function.defected" || event.Output != "U64 addition overflow" || event.ID == "" || event.SourcePlace == "" {
				t.Fatalf("overflow event=%+v, want source-attributed U64 addition overflow defect", event)
			}
		}
	}
}

func TestPhase26InterpreterRepeatedCopy(t *testing.T) {
	source, err := os.ReadFile(filepath.Join(interpProjectRoot(), "examples", "sum_to_n.schway"))
	if err != nil {
		t.Fatal(err)
	}
	source = append(source, []byte("\nfn entry(n: U64) -> U64 {\n  let result = main(n)\n  result\n}\n")...)
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("repeated-copy source parse diagnostics: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("repeated-copy source check diagnostics: %+v", checked.Diagnostics)
	}
	program := checked.Program
	var main *core.Function
	for index := range program.Functions {
		if program.Functions[index].Name == "main" {
			main = &program.Functions[index]
		}
	}
	if main == nil || main.Linear == nil {
		t.Fatal("checked sum_to_n has no linear main")
	}
	var sourcePlace core.Place
	for _, place := range main.Linear.Places {
		if place.Name == "i" {
			sourcePlace = place
		}
	}
	if sourcePlace.ID == "" {
		t.Fatal("checked sum_to_n has no loop counter place")
	}
	targetID := fmt.Sprintf("%s:place:%d", main.ID, len(main.Linear.Places))
	copyID := fmt.Sprintf("%s:op:%d", main.ID, len(main.Linear.Operations))
	main.Linear.Places = append(main.Linear.Places, core.Place{ID: targetID, Name: "snapshot", TypeID: sourcePlace.TypeID})
	main.Linear.Operations = append(main.Linear.Operations, core.LinearOperation{ID: copyID, PointID: fmt.Sprintf("%s:point:linear:%d", main.ID, len(main.Linear.Operations)), Kind: core.OpCopy, SourceID: sourcePlace.ID, TargetID: targetID, TypeID: sourcePlace.TypeID})
	inserted := false
	for index := range main.Linear.Blocks {
		block := &main.Linear.Blocks[index]
		if !strings.Contains(block.ID, "while-body") {
			continue
		}
		block.OperationIDs = append(block.OperationIDs, "")
		copy(block.OperationIDs[2:], block.OperationIDs[1:])
		block.OperationIDs[1] = copyID
		inserted = true
		break
	}
	if !inserted {
		t.Fatal("checked sum_to_n has no while body block")
	}
	for _, input := range []struct {
		value string
		count int
	}{{"3", 3}, {"0", 0}} {
		result, err := Run(program, "entry", input.value)
		if err != nil {
			t.Fatalf("Run(entry, %q): %v", input.value, err)
		}
		if result.Outcome.Kind != execution.OutcomeReturned || result.Outcome.Value != map[string]string{"3": "6", "0": "0"}[input.value] {
			t.Fatalf("input %s outcome=%+v", input.value, result.Outcome)
		}
		var document struct {
			Events []map[string]any `json:"events"`
		}
		encoded, err := CanonicalBytes(result)
		if err != nil || json.Unmarshal(encoded, &document) != nil {
			t.Fatalf("decode execution events: err=%v bytes=%s", err, encoded)
		}
		var visits []map[string]any
		for _, event := range document.Events {
			if event["id"] == copyID+":event" {
				visits = append(visits, event)
			}
		}
		if len(visits) != input.count {
			t.Fatalf("input %s scalar copy events=%d, want %d: %+v", input.value, len(visits), input.count, visits)
		}
		if input.count == 3 {
			firstID, firstInvocation := visits[0]["id"], visits[0]["invocation"]
			for index, event := range visits {
				ordinal, present := event["occurrence"].(float64)
				if event["id"] != firstID || event["invocation"] != firstInvocation || (index == 0 && present) || (index > 0 && (!present || ordinal != float64(index))) {
					t.Fatalf("copy visit %d identity=%+v, want stable site/invocation and ordinal %d (zero omitted)", index, event, index)
				}
			}
		}
	}
}
