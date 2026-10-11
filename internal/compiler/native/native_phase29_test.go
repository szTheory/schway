package native_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const (
	phase29FixtureSource   = "examples/phase29/adler32_fixture.schway"
	phase29OOBSource       = "examples/phase29/byte_index_oob.schway"
	phase29ManifestRelPath = "examples/phase29/fixture_bytes.bindings.json"
	phase29OOBStderr       = "schway: byte index out of bounds\n"
)

// phase29Lane is one optimization lane. The -O3 and -O3 -flto lanes are a
// test-only seam (D-21): production applicationFlags stay -O0.
type phase29Lane struct {
	name  string
	flags []string
}

var phase29BaseFlags = []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic"}

func phase29Lanes() []phase29Lane {
	return []phase29Lane{
		{"O0", append(append([]string(nil), phase29BaseFlags...), "-O0")},
		{"O3", append(append([]string(nil), phase29BaseFlags...), "-O3")},
		{"O3-LTO", append(append([]string(nil), phase29BaseFlags...), "-O3", "-flto")},
	}
}

func phase29RequireClang(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang unavailable on this host")
	}
}

// phase29BuildLane builds the Schway source at sourcePath with the fixture
// manifest under the lane's flags. Seam users must never be parallel.
func phase29BuildLane(t *testing.T, sourcePath string, flags []string) string {
	t.Helper()
	restore := native.SetApplicationFlagsForTest(flags)
	defer restore()
	artifact := filepath.Join(t.TempDir(), "program")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), sourcePath, artifact, native.DefaultRunner(), testsupport.ProjectPath(phase29ManifestRelPath))
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("build %s with %v: diagnostics=%+v err=%v", sourcePath, flags, diagnostics, err)
	}
	return artifact
}

func phase29Run(t *testing.T, artifact string, caseID uint64) (native.RunOutcome, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, strconv.FormatUint(caseID, 10), &stdout, &stderr)
	if err != nil {
		t.Fatalf("run case %d: %v", caseID, err)
	}
	return outcome, stdout.String(), stderr.String()
}

func phase29AssertOOB(t *testing.T, artifact string, caseID uint64) {
	t.Helper()
	outcome, stdout, stderr := phase29Run(t, artifact, caseID)
	if outcome.Kind != native.RunExited || outcome.ExitCode != 65 || stdout != "" || stderr != phase29OOBStderr {
		t.Fatalf("case %d outcome=%+v stdout=%q stderr=%q, want exit 65, empty stdout, %q", caseID, outcome, stdout, stderr, phase29OOBStderr)
	}
}

// TestPhase29OOBApplicationLanes pins SC3 / BYTE-02: reading one byte past the
// end exits 65 with the exact stderr at -O0, -O3 and -O3 -flto, so the
// optimizer cannot remove the observable check (T-29-28); the interpreter
// projects the same failure as the D-19 defect.
func TestPhase29OOBApplicationLanes(t *testing.T) {
	phase29RequireClang(t)
	source := testsupport.ProjectPath(phase29OOBSource)
	lanes := phase29Lanes()
	// Tracer order: the O3+LTO lane on case 2 first, then the full matrix.
	last := lanes[len(lanes)-1]
	artifact := phase29BuildLane(t, source, last.flags)
	phase29AssertOOB(t, artifact, 2)
	for _, lane := range lanes {
		t.Run(lane.name, func(t *testing.T) {
			artifact := phase29BuildLane(t, source, lane.flags)
			for _, caseID := range []uint64{0, 2} {
				phase29AssertOOB(t, artifact, caseID)
			}
		})
	}
	t.Run("interpreter", func(t *testing.T) {
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(data)
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("diagnostics: %+v", checked.Diagnostics)
		}
		for _, caseID := range []uint64{0, 2} {
			run, err := interp.RunWithFixtureBytes(checked.Program, "main", strconv.FormatUint(caseID, 10), testsupport.Phase29FixtureBytes)
			if err != nil {
				t.Fatalf("case %d: %v", caseID, err)
			}
			if run.Outcome.Kind != execution.OutcomeDefect || run.Outcome.Value != "" || len(run.Events) == 0 {
				t.Fatalf("case %d outcome=%+v, want empty defect", caseID, run.Outcome)
			}
			final := run.Events[len(run.Events)-1]
			if final.Kind != "function.defected" || final.Output != "byte index out of bounds" {
				t.Fatalf("case %d final event = %+v, want function.defected byte index out of bounds", caseID, final)
			}
		}
	})
}
