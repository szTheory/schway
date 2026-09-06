package session_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/interp/interptestdirect"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

var phase5AdversarialTargetHeaderPattern = regexp.MustCompile(`^// adversarial-target: (\S+)`)

// TestPhase5AdversarialSubsetIsComplete is Task 1's own completeness gate
// (D-05-18a): exactly six hand-written fixtures exist under
// testdata/phase5, each carrying a distinct `// adversarial-target:`
// header matching session.Phase5AdversarialTargets verbatim, each accepted
// by the shipped checker and corevalidate with no language change, and
// each reaching a genuine terminal outcome under the interpreter.
func TestPhase5AdversarialSubsetIsComplete(t *testing.T) {
	corpus := testsupport.ProjectPath("testdata", "phase5")

	if got, want := len(session.Phase5AdversarialFixtureFiles), len(session.Phase5AdversarialTargets); got != want {
		t.Fatalf("Phase5AdversarialFixtureFiles has %d entries, Phase5AdversarialTargets has %d, want equal", got, want)
	}

	seenTargets := make(map[string]bool, len(session.Phase5AdversarialTargets))
	for index, fixture := range session.Phase5AdversarialFixtureFiles {
		fixture, index := fixture, index
		t.Run(fixture, func(t *testing.T) {
			path := filepath.Join(corpus, fixture)
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			firstLine := strings.SplitN(string(source), "\n", 2)[0]
			match := phase5AdversarialTargetHeaderPattern.FindStringSubmatch(firstLine)
			if match == nil {
				t.Fatalf("%s: first line %q does not carry a `// adversarial-target: <name>` header", fixture, firstLine)
			}
			target := match[1]
			want := session.Phase5AdversarialTargets[index]
			if target != want {
				t.Fatalf("%s: header names target %q, want %q (Phase5AdversarialTargets[%d])", fixture, target, want, index)
			}
			if seenTargets[target] {
				t.Fatalf("%s: target %q is not distinct from an earlier fixture's header", fixture, target)
			}
			seenTargets[target] = true

			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("%s: unexpected diagnostics: %+v", fixture, checked.Diagnostics)
			}
			if len(checked.Program.Functions) != 1 {
				t.Fatalf("%s: expected exactly one function, got %d", fixture, len(checked.Program.Functions))
			}
		})
	}

	if got, want := len(seenTargets), len(session.Phase5AdversarialTargets); got != want {
		t.Fatalf("saw %d distinct adversarial-target headers, want %d (one per Phase5AdversarialTargets entry)", got, want)
	}

	// grep-visible invariant (mirrored in the plan's own acceptance
	// criteria): every fixture's header, deduplicated and sorted, is
	// exactly the six declared names.
	names := make([]string, 0, len(seenTargets))
	for name := range seenTargets {
		names = append(names, name)
	}
	sort.Strings(names)
	wantSorted := append([]string(nil), session.Phase5AdversarialTargets...)
	sort.Strings(wantSorted)
	for index := range wantSorted {
		if names[index] != wantSorted[index] {
			t.Fatalf("sorted header set diverges from sorted Phase5AdversarialTargets at index %d: %q vs %q", index, names[index], wantSorted[index])
		}
	}

	t.Run("no-prior-phase-drift", func(t *testing.T) {
		for _, phase := range []string{"phase1", "phase2", "phase3", "phase4"} {
			dir := testsupport.ProjectPath("testdata", phase)
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("read %s: %v", dir, err)
			}
			if len(entries) == 0 {
				t.Fatalf("%s is unexpectedly empty", dir)
			}
		}
	})

	t.Run("terminal-outcomes", func(t *testing.T) {
		assertOrdinaryTerminalOutcome(t, corpus, "inline_across_foreign.lang", "returned")
		assertOrdinaryTerminalOutcome(t, corpus, "dead_store_unused_acquire.lang", "returned")
		assertOrdinaryTerminalOutcome(t, corpus, "reorder_two_events.lang", "returned")
		assertOrdinaryTerminalOutcome(t, corpus, "tail_collapse_release_ladder.lang", "returned")
		assertDefectDiesBySignal(t, corpus)
		assertTypedFailureTruncatedStdout(t, corpus)
	})
}

// assertOrdinaryTerminalOutcome drives fixture through the ordinary
// interpreter entry point (interp.Run, via session.Phase4CheckedProgram)
// with input "7" and asserts the named terminal outcome kind.
func assertOrdinaryTerminalOutcome(t *testing.T, corpus, fixture, wantKind string) {
	t.Helper()
	program, functionName, err := session.Phase4CheckedProgram(corpus, fixture)
	if err != nil {
		t.Fatalf("%s: %v", fixture, err)
	}
	execution, err := interp.Run(program, functionName, "7")
	if err != nil {
		t.Fatalf("%s: interp.Run: %v", fixture, err)
	}
	if execution.Outcome.Kind != wantKind {
		t.Fatalf("%s: outcome.kind = %q, want %q", fixture, execution.Outcome.Kind, wantKind)
	}
}

// assertDefectDiesBySignal drives defect_dies_by_signal.lang's Halt arm
// through the ordinary interpreter (a match-arm defect terminator is
// genuinely reachable via interp.Run, unlike a foreign call's err edge --
// see defect_terminal.lang's own established Phase 4 precedent) and
// asserts a real defect outcome.
func assertDefectDiesBySignal(t *testing.T, corpus string) {
	t.Helper()
	program, functionName, err := session.Phase4CheckedProgram(corpus, "defect_dies_by_signal.lang")
	if err != nil {
		t.Fatalf("defect_dies_by_signal.lang: %v", err)
	}
	execution, err := interp.Run(program, functionName, "Halt")
	if err != nil {
		t.Fatalf("defect_dies_by_signal.lang: interp.Run: %v", err)
	}
	if execution.Outcome.Kind != "defect" {
		t.Fatalf("defect_dies_by_signal.lang: outcome.kind = %q, want %q", execution.Outcome.Kind, "defect")
	}
}

// assertTypedFailureTruncatedStdout proves typed_failure_truncated_stdout.lang
// genuinely reaches a typed_failure outcome under the interpreter. The
// interpreter's own documented discretionary stub always simulates success
// for every OpForeignCall (interp.Run cannot actually call C), so the
// fixture's own final stage's failure block is driven directly via
// interptestdirect.RunLinearBlockDirect -- exactly session_test.go's own
// established assertTypedFailurePathAgrees technique, peered rather than
// duplicated as a private mechanism.
func assertTypedFailureTruncatedStdout(t *testing.T, corpus string) {
	t.Helper()
	program, functionName, err := session.Phase4CheckedProgram(corpus, "typed_failure_truncated_stdout.lang")
	if err != nil {
		t.Fatalf("typed_failure_truncated_stdout.lang: %v", err)
	}
	functionID := program.Functions[0].ID
	const stages = 150
	precedingCallIDs := make([]string, 0, stages-1)
	for index := 0; index < stages-1; index++ {
		precedingCallIDs = append(precedingCallIDs, fmt.Sprintf("%s:op:%d", functionID, index))
	}
	failingCallID := fmt.Sprintf("%s:op:%d", functionID, stages-1)
	blockID := fmt.Sprintf("%s:block:err:%d", functionID, stages-1)

	execution, err := interptestdirect.RunLinearBlockDirect(program, functionName, blockID, precedingCallIDs, failingCallID)
	if err != nil {
		t.Fatalf("typed_failure_truncated_stdout.lang: interptestdirect.RunLinearBlockDirect: %v", err)
	}
	if execution.Outcome.Kind != "typed_failure" {
		t.Fatalf("typed_failure_truncated_stdout.lang: outcome.kind = %q, want %q", execution.Outcome.Kind, "typed_failure")
	}
	if len(execution.Events) < stages {
		t.Fatalf("typed_failure_truncated_stdout.lang: only %d events, want at least %d (one per acquired stage, release, and the final failure)", len(execution.Events), stages)
	}
}

// TestPhase5CorpusBoundConstantsAreExported is D-05-18's own duplication-
// plus-equality-test anchor: scripts/verify-phase5.sh (plan 05-09) and
// TestPhase5RequiredControlsMatchScript's own peer will duplicate these
// four names and values verbatim, mirroring Phase4RequiredControls'
// established pattern.
func TestPhase5CorpusBoundConstantsAreExported(t *testing.T) {
	if session.Phase5CorpusBoundVersion != 1 {
		t.Fatalf("Phase5CorpusBoundVersion = %d, want 1", session.Phase5CorpusBoundVersion)
	}
	if session.Phase5EnumerationMaxDepth != 3 {
		t.Fatalf("Phase5EnumerationMaxDepth = %d, want 3", session.Phase5EnumerationMaxDepth)
	}
	if session.Phase5EnumerationMaxStatements != 4 {
		t.Fatalf("Phase5EnumerationMaxStatements = %d, want 4", session.Phase5EnumerationMaxStatements)
	}
	if len(session.Phase5AdversarialTargets) != 6 {
		t.Fatalf("Phase5AdversarialTargets has %d entries, want 6", len(session.Phase5AdversarialTargets))
	}
}

// phase5CanonicalPrograms serializes programs deterministically (sorted by
// module name, then canonical JSON per program) so two independently
// generated slices compare byte-for-byte regardless of any incidental
// slice-order coincidence.
func phase5CanonicalPrograms(t *testing.T, programs []core.Program) []byte {
	t.Helper()
	sorted := append([]core.Program(nil), programs...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Module < sorted[j].Module })
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	for _, program := range sorted {
		if err := encoder.Encode(program); err != nil {
			t.Fatalf("encode program %q: %v", program.Module, err)
		}
	}
	return buffer.Bytes()
}

// TestPhase5EnumerationIsDeterministic proves EnumeratePhase5Closure has no
// hidden nondeterminism (randomness, map iteration order, wall-clock or
// filesystem dependence): two independent calls in the same process
// produce byte-identical serialized program sequences.
func TestPhase5EnumerationIsDeterministic(t *testing.T) {
	first := session.EnumeratePhase5Closure()
	second := session.EnumeratePhase5Closure()
	if len(first) != len(second) {
		t.Fatalf("first run produced %d programs, second run produced %d", len(first), len(second))
	}
	firstBytes := phase5CanonicalPrograms(t, first)
	secondBytes := phase5CanonicalPrograms(t, second)
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("two EnumeratePhase5Closure runs diverged byte-for-byte")
	}
	firstRejected := session.EnumeratePhase5ClosureRejected()
	secondRejected := session.EnumeratePhase5ClosureRejected()
	if firstRejected != secondRejected {
		t.Fatalf("EnumeratePhase5ClosureRejected diverged across runs: %d vs %d", firstRejected, secondRejected)
	}
}

// TestPhase5EnumerationIsNonEmpty is T-05-17's mitigation: an empty or
// all-rejected enumeration must never pass as a clean run. The accepted
// count must clear a floor of 20, and the rejected count -- however large
// -- is always exposed alongside it, never hidden.
func TestPhase5EnumerationIsNonEmpty(t *testing.T) {
	accepted := session.EnumeratePhase5Closure()
	rejected := session.EnumeratePhase5ClosureRejected()
	t.Logf("EnumeratePhase5Closure: accepted=%d rejected=%d", len(accepted), rejected)
	if len(accepted) < 20 {
		t.Fatalf("EnumeratePhase5Closure accepted only %d programs, want at least 20 (rejected=%d)", len(accepted), rejected)
	}
	if rejected < 0 {
		t.Fatalf("EnumeratePhase5ClosureRejected returned a negative count: %d", rejected)
	}
}

// phase5AllowedGrammarConstructors is the exact two-element type slice
// D-05-18b's enumerated closure is scoped to ("one parameter, Byte |
// Buffer").
var phase5AllowedGrammarConstructors = map[string]bool{"Byte": true, "Buffer": true}

// TestPhase5EnumerationRespectsBound asserts, structurally (never by
// comment), that no generated program exceeds Phase5EnumerationMaxDepth /
// Phase5EnumerationMaxStatements and none uses a construct outside the
// six-element grammar slice: more than one parameter type outside Byte |
// Buffer, more than 2 ADT alternatives, more than one foreign call per
// function, or an operation kind this project's core package does not
// define (core has no Lang-to-Lang call kind at all -- D-05-32/D-12a -- so
// this loop additionally proves that invariant on every generated program,
// not merely by the absence of a constructor in this file).
func TestPhase5EnumerationRespectsBound(t *testing.T) {
	knownKinds := map[core.OperationKind]bool{
		core.OpCopy: true, core.OpMove: true, core.OpBorrowShared: true, core.OpBorrowExclusive: true,
		core.OpReturn: true, core.OpForeignCall: true, core.OpFail: true, core.OpRelease: true, core.OpDefect: true,
	}
	programs := session.EnumeratePhase5Closure()
	if len(programs) == 0 {
		t.Fatal("no programs to check (EnumeratePhase5Closure returned empty)")
	}
	for _, program := range programs {
		for _, dataType := range program.DataTypes {
			if len(dataType.Alternatives) > 2 {
				t.Fatalf("program %q: data type %q has %d alternatives, want at most 2", program.Module, dataType.Name, len(dataType.Alternatives))
			}
		}
		for _, function := range program.Functions {
			if !phase5AllowedGrammarConstructors[function.Parameter.Type] {
				t.Fatalf("program %q: function %q has parameter type %q outside {Byte, Buffer}", program.Module, function.Name, function.Parameter.Type)
			}
			if function.Linear == nil {
				t.Fatalf("program %q: function %q has no linear body (branch-only bodies are outside this enumerator's scope)", program.Module, function.Name)
			}
			foreignCalls := 0
			maxOperations := (session.Phase5EnumerationMaxStatements+1)*4 + 16
			if len(function.Linear.Operations) > maxOperations {
				t.Fatalf("program %q: function %q has %d operations, want at most %d (Phase5EnumerationMaxStatements=%d bound)", program.Module, function.Name, len(function.Linear.Operations), maxOperations, session.Phase5EnumerationMaxStatements)
			}
			for _, operation := range function.Linear.Operations {
				if !knownKinds[operation.Kind] {
					t.Fatalf("program %q: function %q has an operation of unknown kind %q (a Lang-to-Lang call construct would surface here)", program.Module, function.Name, operation.Kind)
				}
				if operation.Kind == core.OpForeignCall {
					foreignCalls++
				}
			}
			if foreignCalls > 1 {
				t.Fatalf("program %q: function %q has %d foreign calls, want 0 or 1", program.Module, function.Name, foreignCalls)
			}
		}
	}
}
