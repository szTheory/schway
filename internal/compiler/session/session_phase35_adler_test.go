package session_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase35ExistingAdlerAnswers records the current checksum and adler_low
// answers through both engines before the Adler source is compacted.
func TestPhase35ExistingAdlerAnswers(t *testing.T) {
	dir := phase31CopyDir(t, testsupport.ProjectPath("examples", "checksum"))
	library, err := os.ReadFile(filepath.Join(dir, "adler32.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(library), "var r = b << 16\n") != 1 || strings.Contains(string(library), "r = r + r") {
		t.Fatal("Adler library must use exactly one b << 16 and no repeated doublings")
	}
	for _, app := range []string{"checksum", "adler_low"} {
		t.Run(app, func(t *testing.T) {
			entry := filepath.Join(dir, app+".schway")
			checked, err := session.CheckFile(entry)
			if err != nil || len(checked.Diagnostics) != 0 {
				t.Fatalf("check %s: err=%v diagnostics=%+v", app, err, checked.Diagnostics)
			}
			foundLibrary := 0
			for _, function := range checked.Program.Functions {
				if function.Name == "adler32" {
					foundLibrary++
					if function.Module != "adler32" {
						t.Fatalf("adler32 module = %q, want adler32", function.Module)
					}
				}
			}
			if foundLibrary != 1 {
				t.Fatalf("merged program has %d adler32 definitions, want one imported definition", foundLibrary)
			}

			artifact, _ := phase31CopyBuild(t, dir, app)
			rows := 0
			for _, row := range testsupport.ChecksumOracleRows() {
				if row.App != app {
					continue
				}
				rows++
				input := phase31CopyInput(t, string(row.Input))
				interpRun, err := interp.RunWithFileBytes(checked.Program, "main", input, testsupport.ChecksumReadFile)
				if err != nil {
					t.Fatalf("%s interpreter: %v", row.Name, err)
				}
				if len(interpRun.LiveResources) != 0 {
					t.Fatalf("%s interpreter live resources = %v", row.Name, interpRun.LiveResources)
				}
				if row.Outcome == "FileTooLarge" {
					if interpRun.Outcome.Kind != execution.OutcomeTypedFailure || interpRun.Outcome.Value != "FileTooLarge" {
						t.Fatalf("%s interpreter outcome = %+v, want FileTooLarge", row.Name, interpRun.Outcome)
					}
				} else {
					want := fmt.Sprint(row.Want)
					if interpRun.Outcome.Kind != execution.OutcomeReturned || interpRun.Outcome.Value != want {
						t.Fatalf("%s interpreter outcome = %+v, want returned %s", row.Name, interpRun.Outcome, want)
					}
					released := 0
					for _, event := range interpRun.Events {
						if event.Kind == "resource.released" {
							released++
						}
					}
					if released != 1 {
						t.Fatalf("%s interpreter release events = %d, want one", row.Name, released)
					}
				}

				var stdout, stderr bytes.Buffer
				run, err := native.RunApplication(context.Background(), artifact, input, &stdout, &stderr)
				if err != nil || run.Kind != native.RunExited {
					t.Fatalf("%s native outcome=%+v err=%v stdout=%q stderr=%q", row.Name, run, err, stdout.String(), stderr.String())
				}
				if row.Outcome == "FileTooLarge" {
					if run.ExitCode != 65 || stdout.Len() != 0 || stderr.String() != "FileError.FileTooLarge\n" {
						t.Fatalf("%s native=(%d,%q,%q), want (65,empty,FileError.FileTooLarge)", row.Name, run.ExitCode, stdout.String(), stderr.String())
					}
				} else {
					want := fmt.Sprintf("%d\n", row.Want)
					if run.ExitCode != 0 || stdout.String() != want || stderr.Len() != 0 {
						t.Fatalf("%s native=(%d,%q,%q), want (0,%q,empty)", row.Name, run.ExitCode, stdout.String(), stderr.String(), want)
					}
				}
			}
			if rows != 7 {
				t.Fatalf("oracle rows for %s = %d, want all seven", app, rows)
			}
		})
	}
}

// TestPhase35AdlerWrongShiftControl mutates only a temporary module copy and
// pins the resulting answer on both engines before restoring and rerunning it.
func TestPhase35AdlerWrongShiftControl(t *testing.T) {
	dir := phase31CopyDir(t, testsupport.ProjectPath("examples", "checksum"))
	library := filepath.Join(dir, "adler32.schway")
	phase31CopyReplaceOnce(t, library, "var r = b << 16\n", "var r = b << 15\n")
	check := func(want string) {
		t.Helper()
		checked, err := session.CheckFile(filepath.Join(dir, "checksum.schway"))
		if err != nil || len(checked.Diagnostics) != 0 {
			t.Fatalf("checksum check: err=%v diagnostics=%+v", err, checked.Diagnostics)
		}
		input := phase31CopyInput(t, "Wikipedia")
		interpreted, err := interp.RunWithFileBytes(checked.Program, "main", input, testsupport.ChecksumReadFile)
		if err != nil || interpreted.Outcome.Kind != execution.OutcomeReturned || interpreted.Outcome.Value != strings.TrimSuffix(want, "\n") {
			t.Fatalf("interpreter outcome=%+v err=%v, want %s", interpreted.Outcome, err, strings.TrimSpace(want))
		}
		artifact, _ := phase31CopyBuild(t, dir, "checksum")
		got := phase31CopyNative(t, artifact, "Wikipedia")
		if got != want {
			t.Fatalf("native output=%q, want %q", got, want)
		}
	}
	check("150143896\n")
	phase31CopyReplaceOnce(t, library, "var r = b << 15\n", "var r = b << 16\n")
	check("300286872\n")
}
