package session_test

import (
	"context"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
)

// TestSanitizeLaneRetainedPointerAlwaysReports proves D-05-14's always-on
// positive control fires on a fresh invocation of VerifyPhase5SanitizeLane,
// asserted by substring AND exit code, and that the lane overall passes.
func TestSanitizeLaneRetainedPointerAlwaysReports(t *testing.T) {
	result, err := session.VerifyPhase5SanitizeLane(context.Background(), native.DefaultRunner())
	if err != nil {
		requirePhase16M004Refusal(t, err, "foreign")
		return
	}
	if result.Status != protocol.StatusPass {
		t.Fatalf("expected pass, got status=%s diagnostics=%+v", result.Status, result.Diagnostics)
	}
	found := false
	for _, lane := range result.Lanes {
		for _, control := range lane.Controls {
			if control == session.ControlSanitizeRetainedPointer {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("%s did not appear as a detected control: lanes=%+v", session.ControlSanitizeRetainedPointer, result.Lanes)
	}
}

// TestSanitizeLaneAllocatorMismatchIsDetected proves D-05-08's dynamic
// allocator-mismatch control fires.
func TestSanitizeLaneAllocatorMismatchIsDetected(t *testing.T) {
	result, err := session.VerifyPhase5SanitizeLane(context.Background(), native.DefaultRunner())
	if err != nil {
		requirePhase16M004Refusal(t, err, "foreign")
		return
	}
	if result.Status != protocol.StatusPass {
		t.Fatalf("expected pass, got status=%s diagnostics=%+v", result.Status, result.Diagnostics)
	}
	found := false
	for _, lane := range result.Lanes {
		for _, control := range lane.Controls {
			if control == session.ControlSanitizeAllocatorMismatch {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("%s did not appear as a detected control: lanes=%+v", session.ControlSanitizeAllocatorMismatch, result.Lanes)
	}
}

// TestSanitizeLaneUBSanNoRecoverIsProven proves -fno-sanitize-recover=all
// actually took effect (D-05-13/D-05-14): the UBSan fixture reports
// "runtime error:" with a nonzero exit code.
func TestSanitizeLaneUBSanNoRecoverIsProven(t *testing.T) {
	result, err := session.VerifyPhase5SanitizeLane(context.Background(), native.DefaultRunner())
	if err != nil {
		requirePhase16M004Refusal(t, err, "foreign")
		return
	}
	if result.Status != protocol.StatusPass {
		t.Fatalf("expected pass, got status=%s diagnostics=%+v", result.Status, result.Diagnostics)
	}
	found := false
	for _, lane := range result.Lanes {
		for _, control := range lane.Controls {
			if control == session.ControlSanitizeUBSanNoRecover {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("%s did not appear as a detected control: lanes=%+v", session.ControlSanitizeUBSanNoRecover, result.Lanes)
	}
}

// TestSanitizeLaneMissingRuntimeIsOperational asserts that with a broken
// ClangPath, the lane is operational, never a pass -- D-05-15's posture,
// mirroring control:foreign.unwind_forbidden's own nm-absence handling.
func TestSanitizeLaneMissingRuntimeIsOperational(t *testing.T) {
	runner := native.DefaultRunner()
	runner.ClangPath = "definitely-not-a-real-clang-binary"
	result, err := session.VerifyPhase5SanitizeLane(context.Background(), runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != protocol.StatusOperational {
		t.Fatalf("expected operational status, got %s", result.Status)
	}
	for _, lane := range result.Lanes {
		if lane.Status == protocol.StatusPass {
			t.Fatalf("no lane may report a pass with the sanitizer runtime unavailable: lanes=%+v", result.Lanes)
		}
	}
}

// benignPositiveControlSource is a syntactically-similar, but genuinely
// non-defective, replacement for the retained-pointer fixture's compiled C:
// a clean allocate/free cycle with no dangling access, used ONLY to prove
// the lane goes red when the positive control stops reporting.
const benignPositiveControlSource = `#include <stdlib.h>
int main(int argc, char **argv) {
  if (argc != 2) return 64;
  unsigned char *block = malloc(1u);
  if (block == NULL) return 1;
  block[0] = 7u;
  unsigned char value = block[0];
  free(block);
  return value == 7u ? 0 : 1;
}
`

// TestSanitizeLaneCleanPositiveControlFailsTheLane is the mutation-kill
// proving the lane is not inert (D-05-14): with the positive control's
// OWN compiled C substituted (via the exported test-only seam
// session.Phase5RetainedPointerFixtureLoaderForTest) for a benign,
// non-defective program, VerifyPhase5SanitizeLane itself must go RED
// rather than silently passing on a clean run -- a clean sanitizer run is
// never evidence.
func TestSanitizeLaneCleanPositiveControlFailsTheLane(t *testing.T) {
	original := session.Phase5RetainedPointerFixtureLoaderForTest
	defer func() { session.Phase5RetainedPointerFixtureLoaderForTest = original }()
	session.Phase5RetainedPointerFixtureLoaderForTest = func() (string, error) {
		return benignPositiveControlSource, nil
	}

	result, err := session.VerifyPhase5SanitizeLane(context.Background(), native.DefaultRunner())
	if err != nil {
		requirePhase16M004Refusal(t, err, "foreign")
		return
	}
	if result.Status == protocol.StatusPass {
		t.Fatalf("lane must go red when the positive control stops reporting, got status=%s lanes=%+v", result.Status, result.Lanes)
	}
	for _, lane := range result.Lanes {
		for _, control := range lane.Controls {
			if control == session.ControlSanitizeRetainedPointer {
				t.Fatalf("%s must not appear as a detected control once the positive control is inert: lanes=%+v", session.ControlSanitizeRetainedPointer, result.Lanes)
			}
		}
	}
}

// TestCallbackEscapeIsDeclaredNeverDetected peers
// TestExpectedEscapesAreVisibleNotSolved (session_test.go): the callback-
// invocation residual must be declared under ExpectedEscapes and must
// never appear as a detected control anywhere in the lane's own report.
func TestCallbackEscapeIsDeclaredNeverDetected(t *testing.T) {
	result, err := session.VerifyPhase5SanitizeLane(context.Background(), native.DefaultRunner())
	if err != nil {
		requirePhase16M004Refusal(t, err, "foreign")
		return
	}
	declared := false
	for _, escape := range result.ExpectedEscapes {
		if escape == session.EscapeCallbackInvocationUnsubjected {
			declared = true
		}
	}
	if !declared {
		t.Fatalf("expected escape %s is not declared: %+v", session.EscapeCallbackInvocationUnsubjected, result.ExpectedEscapes)
	}
	for _, lane := range result.Lanes {
		for _, control := range lane.Controls {
			if control == session.EscapeCallbackInvocationUnsubjected {
				t.Fatalf("expected escape %s must never be presented as a detected control: lanes=%+v", session.EscapeCallbackInvocationUnsubjected, result.Lanes)
			}
		}
	}
}
