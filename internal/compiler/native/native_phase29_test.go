package native_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

func phase29Interpret(t *testing.T, source []byte, caseID uint64) interp.Execution {
	t.Helper()
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}
	run, err := interp.RunWithFixtureBytes(checked.Program, "main", strconv.FormatUint(caseID, 10), testsupport.Phase29FixtureBytes)
	if err != nil {
		t.Fatalf("interpreter case %d: %v", caseID, err)
	}
	return run
}

// TestPhase29AdlerBothEngines pins SC1 / SC2 / APP-10: every case id returns
// the pinned literal from the interpreter and from native builds at -O0, -O3
// and -O3 -flto. Expected values come only from the pinned table.
func TestPhase29AdlerBothEngines(t *testing.T) {
	phase29RequireClang(t)
	source := testsupport.ProjectPath(phase29FixtureSource)
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, lane := range phase29Lanes() {
		t.Run(lane.name, func(t *testing.T) {
			artifact := phase29BuildLane(t, source, lane.flags)
			for _, caseID := range testsupport.Phase29FixtureCaseIDs() {
				want, ok := testsupport.Phase29FixtureExpected(caseID)
				if !ok {
					t.Fatalf("case %d has no pinned literal", caseID)
				}
				outcome, stdout, stderr := phase29Run(t, artifact, caseID)
				expected := strconv.FormatUint(want, 10) + "\n"
				if outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout != expected || stderr != "" {
					t.Errorf("native %s case %d outcome=%+v stdout=%q stderr=%q, want %q", lane.name, caseID, outcome, stdout, stderr, expected)
				}
			}
		})
	}
	t.Run("interpreter", func(t *testing.T) {
		for _, caseID := range testsupport.Phase29FixtureCaseIDs() {
			want, _ := testsupport.Phase29FixtureExpected(caseID)
			run := phase29Interpret(t, data, caseID)
			if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != strconv.FormatUint(want, 10) {
				t.Errorf("interpreter case %d = %+v, want returned %d", caseID, run.Outcome, want)
			}
		}
	})
}

// phase29ReplaceOnce replaces old with new and fails unless old matched exactly
// once, so a mutant can never silently equal the original program.
func phase29ReplaceOnce(t *testing.T, text, old, replacement string) string {
	t.Helper()
	if got := strings.Count(text, old); got != 1 {
		t.Fatalf("mutation anchor %q matched %d times, want exactly 1", old, got)
	}
	return strings.Replace(text, old, replacement, 1)
}

var phase29ViewDataAssign = regexp.MustCompile(`(\w+\.data) = (\w+\.data);`)

// phase29SignedCharMutation retypes the byte view to signed char consistently,
// so the mutant still builds under -Werror -pedantic and is killed by its
// checksum, not by a compile error. The load widens through int and uint32_t
// so a byte >= 0x80 stays on the value path instead of trapping the checked
// U64 add.
func phase29SignedCharMutation(t *testing.T) func(string) string {
	return func(c string) string {
		const member = "const unsigned char *data;"
		if strings.Count(c, member) != 1 {
			t.Fatalf("view member matched %d times, want 1", strings.Count(c, member))
		}
		c = strings.Replace(c, member, "const signed char *data;", 1)
		sites := len(phase29ViewDataAssign.FindAllString(c, -1))
		if sites < 1 {
			t.Fatal("no generated site copies an unsigned data pointer into the view member")
		}
		c = phase29ViewDataAssign.ReplaceAllString(c, "$1 = (const signed char *)$2;")
		const load = "(uint64_t)schway_scalar_0.data["
		if strings.Count(c, load) < 1 {
			t.Fatal("no byte load matched the widening anchor")
		}
		return strings.ReplaceAll(c, load, "(uint64_t)(uint32_t)(int)schway_scalar_0.data[")
	}
}

// TestPhase29AdlerMutantsKilled pins D-07: each Schway source mutant and the
// signed-char C mutant produces a result different from the pinned literal on
// case 5 or 6; the unmutated program is the control.
func TestPhase29AdlerMutantsKilled(t *testing.T) {
	phase29RequireClang(t)
	original, err := os.ReadFile(testsupport.ProjectPath(phase29FixtureSource))
	if err != nil {
		t.Fatal(err)
	}
	text := string(original)
	mutants := []struct {
		name   string
		mutate func(t *testing.T, text string) string
	}{
		{"wrong-modulus", func(t *testing.T, s string) string {
			s = phase29ReplaceOnce(t, s, "a = a % 65521", "a = a % 65520")
			return phase29ReplaceOnce(t, s, "b = b % 65521", "b = b % 65520")
		}},
		{"off-by-one-start", func(t *testing.T, s string) string {
			return phase29ReplaceOnce(t, s, "var i = 0", "var i = 1")
		}},
		{"swapped-halves", func(t *testing.T, s string) string {
			s = phase29ReplaceOnce(t, s, "var r = b", "var r = a")
			return phase29ReplaceOnce(t, s, "r = r + a", "r = r + b")
		}},
	}
	vectors := []uint64{5, 6}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			mutated := mutant.mutate(t, text)
			if mutated == text {
				t.Fatal("mutant equals the original program")
			}
			path := filepath.Join(t.TempDir(), "mutant.schway")
			if err := os.WriteFile(path, []byte(mutated), 0o600); err != nil {
				t.Fatal(err)
			}
			artifact := phase29BuildLane(t, path, phase29Lanes()[0].flags)
			killed := false
			for _, caseID := range vectors {
				want, _ := testsupport.Phase29FixtureExpected(caseID)
				outcome, stdout, _ := phase29Run(t, artifact, caseID)
				nativeDiffers := outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout != strconv.FormatUint(want, 10)+"\n"
				run := phase29Interpret(t, []byte(mutated), caseID)
				interpDiffers := run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != strconv.FormatUint(want, 10)
				t.Logf("case %d pinned %d: native exit=%d stdout=%q differs=%v; interpreter %s %q differs=%v", caseID, want, outcome.ExitCode, stdout, nativeDiffers, run.Outcome.Kind, run.Outcome.Value, interpDiffers)
				if nativeDiffers && interpDiffers {
					killed = true
				}
			}
			if !killed {
				t.Fatalf("%s survived on cases %v on both engines", mutant.name, vectors)
			}
		})
	}
	t.Run("signed-char", func(t *testing.T) {
		restore := native.SetApplicationCSourceMutationForTest(phase29SignedCharMutation(t))
		defer restore()
		artifact := phase29BuildLane(t, testsupport.ProjectPath(phase29FixtureSource), phase29Lanes()[0].flags)
		for _, caseID := range vectors {
			want, _ := testsupport.Phase29FixtureExpected(caseID)
			outcome, stdout, stderr := phase29Run(t, artifact, caseID)
			if outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stderr != "" {
				t.Fatalf("case %d outcome=%+v stderr=%q, want a clean exit 0 with a wrong checksum", caseID, outcome, stderr)
			}
			got, err := strconv.ParseUint(strings.TrimSpace(stdout), 10, 64)
			if err != nil {
				t.Fatalf("case %d stdout %q is not a U64: %v", caseID, stdout, err)
			}
			t.Logf("signed-char case %d: observed %d, pinned %d", caseID, got, want)
			if got == want {
				t.Fatalf("signed-char mutant survived case %d with the pinned literal %d", caseID, want)
			}
		}
	})
	t.Run("control", func(t *testing.T) {
		artifact := phase29BuildLane(t, testsupport.ProjectPath(phase29FixtureSource), phase29Lanes()[0].flags)
		for _, caseID := range vectors {
			want, _ := testsupport.Phase29FixtureExpected(caseID)
			_, stdout, _ := phase29Run(t, artifact, caseID)
			if stdout != strconv.FormatUint(want, 10)+"\n" {
				t.Fatalf("unmutated control case %d stdout=%q, want %d", caseID, stdout, want)
			}
		}
	})
}
