package native_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase321InlineFixture splices a replacement adler32 callee into the
// committed fixture between `fn adler32(` and `fn main(` (D-26: built in
// memory, written to t.TempDir(), never a committed .schway file).
func phase321InlineFixture(t *testing.T, callee string) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath(phase29FixtureSource))
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	start := strings.Index(text, "fn adler32(")
	end := strings.Index(text, "fn main(")
	if start < 0 || end < start {
		t.Fatal("fixture source no longer has the expected adler32/main layout")
	}
	return []byte(text[:start] + callee + "\n" + text[end:])
}

// TestPhase321DefectParityWithLiveOwner pins D-09: each abort-like defect with
// the fixture owner live exits 65 natively with stderr equal to the
// interpreter Execution's final event output (D-25 widened predicate).
func TestPhase321DefectParityWithLiveOwner(t *testing.T) {
	phase29RequireClang(t)
	variants := []struct {
		name, callee string
		caseID       uint64
		prefix       string
	}{
		{"byte index out of bounds", `fn adler32(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  var x = byte_at(bytes, n)
  x
}
`, 2, "schway: "},
		{"checked add overflow", `fn adler32(bytes: ByteView) -> U64 {
  var m = 18446744073709551615
  var n = byte_len(bytes)
  m = m + n
  m
}
`, 2, "schway: "},
		{"remainder by zero", `fn adler32(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  var z = n % n
  z
}
`, 0, "schway: "},
		{"user defect", `fn adler32(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  if n < 1000 {
    defect "boom"
  } else {
    n = n + 0
  }
  n
}
`, 2, "schway app: "},
	}
	lanes := phase29Lanes()
	for index, variant := range variants {
		t.Run(variant.name, func(t *testing.T) {
			source := phase321InlineFixture(t, variant.callee)
			path := filepath.Join(t.TempDir(), "variant.schway")
			if err := os.WriteFile(path, source, 0o600); err != nil {
				t.Fatal(err)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("diagnostics: %+v", checked.Diagnostics)
			}
			input := strconv.FormatUint(variant.caseID, 10)
			run, err := interp.RunWithFixtureBytes(checked.Program, "main", input, testsupport.Phase29FixtureBytes)
			if err != nil {
				t.Fatalf("interpreter rejected the abort-like defect: %v", err)
			}
			if run.Outcome.Kind != execution.OutcomeDefect || len(run.LiveResources) != 1 || len(run.Events) == 0 {
				t.Fatalf("interpreter = %+v live=%v, want a defect with one live resource", run.Outcome, run.LiveResources)
			}
			reason := run.Events[len(run.Events)-1].Output
			want := variant.prefix + reason + "\n"
			chosen := []phase29Lane{lanes[0]}
			if index == 0 {
				chosen = append(chosen, lanes[len(lanes)-1])
			}
			for _, lane := range chosen {
				artifact := phase29BuildLane(t, path, lane.flags)
				outcome, stdout, stderr := phase29Run(t, artifact, variant.caseID)
				if outcome.Kind != native.RunExited || outcome.ExitCode != 65 || stdout != "" || stderr != want {
					t.Fatalf("%s outcome=%+v stdout=%q stderr=%q, want exit 65, empty stdout, %q", lane.name, outcome, stdout, stderr, want)
				}
			}
		})
	}
}
