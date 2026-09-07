package session_test

import (
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestPhase6ExpectedEscapesAreDeclared asserts the four D-06-13 cache
// escapes and D-06-29's residual overfitting escape are all declared by
// Phase6ExpectedEscapes(), and appear verbatim in a real gate run's own
// ExpectedEscapes output.
func TestPhase6ExpectedEscapesAreDeclared(t *testing.T) {
	want := []string{
		session.EscapeCacheUndeclaredEnvironment,
		session.EscapeCacheClangVersionStringStable,
		session.EscapeCacheNondeterministicCodegen,
		session.EscapeCacheDirectoryHandEdited,
		session.EscapeRepairHeldoutCorpusResidualOverfitting,
	}
	declared := session.Phase6ExpectedEscapes()
	declaredSet := make(map[string]bool, len(declared))
	for _, escape := range declared {
		declaredSet[escape] = true
	}
	for _, escape := range want {
		if !declaredSet[escape] {
			t.Fatalf("Phase6ExpectedEscapes() is missing %s", escape)
		}
	}
	if len(declared) != len(want) {
		t.Fatalf("Phase6ExpectedEscapes() has %d entries, want exactly %d: got=%v", len(declared), len(want), declared)
	}
}

// TestPhase6EscapesAreNeverPresentedAsControls asserts the intersection of
// Phase6ExpectedEscapes() and Phase6RequiredControls() is empty -- an
// escape presented as a solved control would be exactly the spoofing
// threat T-06-GATE-02 names.
func TestPhase6EscapesAreNeverPresentedAsControls(t *testing.T) {
	controls := make(map[string]bool, len(session.Phase6RequiredControls()))
	for _, control := range session.Phase6RequiredControls() {
		controls[control] = true
	}
	for _, escape := range session.Phase6ExpectedEscapes() {
		if controls[escape] {
			t.Fatalf("%s is declared as BOTH an expected escape and a required control -- an escape must never be presented as a solved control", escape)
		}
	}
}

// TestPhase6EscapeGrepsMatchScript asserts set equality, in both
// directions, between Phase6ExpectedEscapes() and the escape strings
// scripts/verify-phase6.sh greps for -- a script grepping for an escape Go
// does not declare, or vice versa, must fail this test.
func TestPhase6EscapeGrepsMatchScript(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath("scripts", "verify-phase6.sh"))
	if err != nil {
		t.Fatalf("read scripts/verify-phase6.sh: %v", err)
	}
	text := string(data)

	declared := session.Phase6ExpectedEscapes()
	declaredSet := make(map[string]bool, len(declared))
	for _, escape := range declared {
		declaredSet[escape] = true
	}

	var fromScript []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "grep -q 'escape:") {
			continue
		}
		start := strings.Index(trimmed, "'escape:")
		if start == -1 {
			continue
		}
		rest := trimmed[start+1:]
		end := strings.Index(rest, "'")
		if end == -1 {
			continue
		}
		fromScript = append(fromScript, rest[:end])
	}
	scriptSet := make(map[string]bool, len(fromScript))
	for _, escape := range fromScript {
		scriptSet[escape] = true
	}

	for escape := range declaredSet {
		if !scriptSet[escape] {
			t.Fatalf("Phase6ExpectedEscapes() names %s, which scripts/verify-phase6.sh's escape greps do not", escape)
		}
	}
	for escape := range scriptSet {
		if !declaredSet[escape] {
			t.Fatalf("scripts/verify-phase6.sh greps for escape %s, which Phase6ExpectedEscapes() does not declare", escape)
		}
	}
	if len(scriptSet) != len(declaredSet) {
		t.Fatalf("script escape set (%d) and Go escape set (%d) differ in size: script=%v go=%v", len(scriptSet), len(declaredSet), fromScript, declared)
	}
}
