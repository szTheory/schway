package native_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase33DefectVariant struct {
	name, callee, reason, stderr string
	caseID                       uint64
	input                        []byte
}

func phase33DefectVariants() []phase33DefectVariant {
	return []phase33DefectVariant{
		{"byte-index-oob", `fn adler32(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  var x = byte_at(bytes, n)
  x
}
`, "byte index out of bounds", "schway: byte index out of bounds\n", 2, []byte("x")},
		{"add-overflow", `fn adler32(bytes: ByteView) -> U64 {
  var m = 18446744073709551615
  var n = byte_len(bytes)
  m = m + n
  m
}
`, "U64 addition overflow", "schway: U64 addition overflow\n", 2, []byte("x")},
		{"remainder-by-zero", `fn adler32(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  var z = n % n
  z
}
`, "U64 remainder by zero", "schway: U64 remainder by zero\n", 0, []byte{}},
		{"user-defect", `fn adler32(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  if n < 1000 {
    defect "boom"
  } else {
    n = n + 0
  }
  n
}
`, "boom", "schway app: boom\n", 2, []byte("x")},
	}
}

func phase33MainAcquireID(t *testing.T, program core.Program) string {
	t.Helper()
	for _, function := range program.Functions {
		if function.Name != "main" || function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Foreign != nil && operation.Foreign.Mode == "acquire" {
				return operation.ID
			}
		}
	}
	t.Fatal("main has no acquire operation")
	return ""
}

func phase33AssertParity(t *testing.T, run interp.Execution, err error, ownerID, reason string, outcome native.RunOutcome, stdout, stderr, wantStderr string) {
	t.Helper()
	if err != nil {
		t.Fatalf("interpreter rejected abort-like defect: %v", err)
	}
	if run.Outcome.Kind != execution.OutcomeDefect || len(run.Events) == 0 || run.Events[len(run.Events)-1].Output != reason {
		t.Fatalf("interpreter = outcome %+v events=%+v, want defect reason %q", run.Outcome, run.Events, reason)
	}
	if len(run.LiveResources) != 1 || run.LiveResources[0] != ownerID {
		t.Fatalf("interpreter live resources = %v, want [%s]", run.LiveResources, ownerID)
	}
	if outcome.Kind != native.RunExited || outcome.ExitCode != 65 || stdout != "" || stderr != wantStderr {
		t.Fatalf("native outcome=%+v stdout=%q stderr=%q, want exit 65, empty stdout, %q", outcome, stdout, stderr, wantStderr)
	}
}

func TestPhase33DefectParityFixtureRoute(t *testing.T) {
	phase29RequireClang(t)
	lanes := phase29Lanes()
	for _, variant := range phase33DefectVariants() {
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
			run, interpErr := interp.RunWithFixtureBytes(checked.Program, "main", fmtUint(variant.caseID), testsupport.Phase29FixtureBytes)
			artifact := phase29BuildLane(t, path, lanes[0].flags)
			outcome, stdout, stderr := phase29Run(t, artifact, variant.caseID)
			phase33AssertParity(t, run, interpErr, phase33MainAcquireID(t, checked.Program), variant.reason, outcome, stdout, stderr, variant.stderr)
		})
	}
}

func TestPhase33DefectParityFileRoute(t *testing.T) {
	phase29RequireClang(t)
	for _, variant := range phase33DefectVariants() {
		t.Run(variant.name, func(t *testing.T) {
			dir := phase31ControlCopy(t)
			phase31ControlMutateLibrary(t, dir, func(source string) string {
				start := strings.Index(source, "fn adler32(")
				if start < 0 {
					t.Fatal("adler32 library function not found")
				}
				return source[:start] + variant.callee
			})
			artifact := phase31ControlBuild(t, dir, phase31ControlFlags())
			outcome, stdout, stderr := phase31ControlNative(t, artifact, variant.input)
			run := phase31ControlInterp(t, dir, variant.input)
			checked, err := session.CheckFile(filepath.Join(dir, "checksum.schway"))
			if err != nil || len(checked.Diagnostics) != 0 {
				t.Fatalf("check file route: diagnostics=%+v err=%v", checked.Diagnostics, err)
			}
			phase33AssertParity(t, run, nil, phase33MainAcquireID(t, checked.Program), variant.reason, outcome, stdout, stderr, variant.stderr)
		})
	}
}

func fmtUint(value uint64) string { return strconv.FormatUint(value, 10) }
