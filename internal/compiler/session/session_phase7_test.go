package session_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestVerifyPhase7ControlsAndWork runs Phase 07's own control-and-work
// gate end-to-end and asserts every required control fires with nonzero
// RecomputedWork, mirroring TestVerifyPhase5ControlsAndWork's /
// TestVerifyPhase6ControlsAndWork's own shape.
func TestVerifyPhase7ControlsAndWork(t *testing.T) {
	result, err := session.VerifyPhase7ControlsAndWork(context.Background())
	if err != nil {
		t.Fatalf("VerifyPhase7ControlsAndWork returned an error: %v", err)
	}
	if result.Status != protocol.StatusPass {
		t.Fatalf("VerifyPhase7ControlsAndWork status = %s, want pass; lanes=%+v diagnostics=%+v", result.Status, result.Lanes, result.Diagnostics)
	}
	for _, required := range session.Phase7RequiredControls() {
		found := false
		for _, lane := range result.Lanes {
			for _, control := range lane.Controls {
				if control == required {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("VerifyPhase7ControlsAndWork never fired required control %s", required)
		}
	}
	for _, lane := range result.Lanes {
		if lane.RecomputedWork == 0 {
			t.Fatalf("lane %s reported zero RecomputedWork", lane.ID)
		}
	}
}

// phase7VerifierScriptText reads scripts/verify-phase7.sh's own text,
// mirroring phase5VerifierScriptText's / phase6VerifierScriptText's
// precedent exactly.
func phase7VerifierScriptText(t testing.TB) string {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("scripts", "verify-phase7.sh"))
	if err != nil {
		t.Fatalf("read scripts/verify-phase7.sh: %v", err)
	}
	return string(data)
}

// phase7OwnControlIdentifiers extracts the exact set of `control:` tokens
// from the script's OWN required-control block (the `for control in ...`
// loop following the "Phase 07's own required-control set" comment),
// mirroring phase5OwnControlIdentifiers's / phase6OwnControlIdentifiers's
// anchor-plus-block-extraction technique exactly.
func phase7OwnControlIdentifiers(t testing.TB, text string) []string {
	t.Helper()
	anchor := "Phase 07's own required-control set"
	anchorIndex := strings.Index(text, anchor)
	if anchorIndex == -1 {
		t.Fatalf("scripts/verify-phase7.sh is missing its own required-control-set comment anchor")
	}
	rest := text[anchorIndex:]
	doneIndex := strings.Index(rest, "\ndone")
	if doneIndex == -1 {
		t.Fatalf("scripts/verify-phase7.sh's own required-control block has no closing done")
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

// TestPhase7RequiredControlsMatchScript asserts set equality, in both
// directions, between session.Phase7RequiredControls() and the
// identifiers named in scripts/verify-phase7.sh's own required-control
// block (T-07-24/D-07-41): a control added to one and forgotten in the
// other fails here rather than silently drifting apart.
func TestPhase7RequiredControlsMatchScript(t *testing.T) {
	text := phase7VerifierScriptText(t)
	fromScript := phase7OwnControlIdentifiers(t, text)
	scriptSet := make(map[string]bool, len(fromScript))
	for _, control := range fromScript {
		scriptSet[control] = true
	}
	sessionSet := make(map[string]bool, len(session.Phase7RequiredControls()))
	for _, control := range session.Phase7RequiredControls() {
		sessionSet[control] = true
	}
	for control := range sessionSet {
		if !scriptSet[control] {
			t.Fatalf("session.Phase7RequiredControls() names %s, which the script's own required-control block does not", control)
		}
	}
	for control := range scriptSet {
		if !sessionSet[control] {
			t.Fatalf("the script's required-control block names %s, which session.Phase7RequiredControls() does not", control)
		}
	}
	if len(scriptSet) != len(sessionSet) {
		t.Fatalf("script control set (%d) and session control set (%d) differ in size: script=%v session=%v", len(scriptSet), len(sessionSet), fromScript, session.Phase7RequiredControls())
	}
}
