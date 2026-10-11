package session_test

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase29FixtureManifestPath = "../../../examples/phase29/fixture_bytes.bindings.json"

// TestPhase29NativeFixtureVectors builds the real fixture program once with its
// binding manifest and runs the single artifact for every pinned case id (D-15,
// D-16): one artifact serves every vector, and an unknown case exits 65 with
// nothing printed.
func TestPhase29NativeFixtureVectors(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang unavailable on this host")
	}
	artifact := filepath.Join(t.TempDir(), "adler32_fixture")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), phase29FixturePath, artifact, native.DefaultRunner(), phase29FixtureManifestPath)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplicationFile diagnostics=%+v err=%v", diagnostics, err)
	}
	for caseID := uint64(0); caseID <= 9; caseID++ {
		want, ok := testsupport.Phase29FixtureExpected(caseID)
		if !ok {
			t.Fatalf("case %d has no pinned checksum", caseID)
		}
		var stdout, stderr bytes.Buffer
		outcome, err := native.RunApplication(context.Background(), artifact, strconv.FormatUint(caseID, 10), &stdout, &stderr)
		expected := strconv.FormatUint(want, 10) + "\n"
		if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != expected || stderr.Len() != 0 {
			t.Fatalf("case %d outcome=%+v stdout=%q stderr=%q err=%v, want %q exit 0", caseID, outcome, stdout.String(), stderr.String(), err, expected)
		}
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, "10", &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 65 || stdout.Len() != 0 {
		t.Fatalf("unknown case outcome=%+v stdout=%q stderr=%q err=%v, want exit 65 and empty stdout", outcome, stdout.String(), stderr.String(), err)
	}
}
