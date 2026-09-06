package session_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

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
