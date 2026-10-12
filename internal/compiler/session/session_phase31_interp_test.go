package session_test

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase31InterpFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "in.bin")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func phase31InterpCheck(t *testing.T, path string) session.CheckResult {
	t.Helper()
	checked, err := session.CheckFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics for %s: %+v", path, checked.Diagnostics)
	}
	return checked
}

// phase31InterpNoRegularCheck is the armed control: it reads like
// testsupport.ChecksumReadFile but omits the IsRegular test.
func phase31InterpNoRegularCheck(path string) ([]byte, string) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, "FileUnavailable"
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, testsupport.ChecksumLimit+1))
	if err != nil {
		return nil, "FileUnavailable"
	}
	if len(data) > testsupport.ChecksumLimit {
		return nil, "FileTooLarge"
	}
	if data == nil {
		data = []byte{}
	}
	return data, ""
}

// TestPhase31ChecksumInterpreter runs the checksum example in the interpreter
// from an independent Go host reader (D-07, D-09, D-16).
func TestPhase31ChecksumInterpreter(t *testing.T) {
	checked := phase31InterpCheck(t, testsupport.ProjectPath("examples", "checksum", "checksum.schway"))

	t.Run("wikipedia", func(t *testing.T) {
		run, err := interp.RunWithFileBytes(checked.Program, "main", phase31InterpFile(t, "Wikipedia"), testsupport.ChecksumReadFile)
		if err != nil {
			t.Fatal(err)
		}
		if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != "300286872" || len(run.LiveResources) != 0 {
			t.Fatalf("outcome = %+v live=%v, want returned 300286872 with no live resources", run.Outcome, run.LiveResources)
		}
		released := 0
		for _, event := range run.Events {
			if event.Kind == "resource.released" {
				released++
			}
		}
		if released != 1 {
			t.Fatalf("release events = %d, want 1", released)
		}
	})

	t.Run("missing", func(t *testing.T) {
		run, err := interp.RunWithFileBytes(checked.Program, "main", filepath.Join(t.TempDir(), "absent.bin"), testsupport.ChecksumReadFile)
		if err != nil {
			t.Fatal(err)
		}
		if run.Outcome.Kind != execution.OutcomeTypedFailure || run.Outcome.Value != "FileUnavailable" || len(run.LiveResources) != 0 {
			t.Fatalf("outcome = %+v live=%v, want typed_failure FileUnavailable", run.Outcome, run.LiveResources)
		}
	})

	t.Run("unknown-alternative-refused", func(t *testing.T) {
		_, err := interp.RunWithFileBytes(checked.Program, "main", "x", func(string) ([]byte, string) { return nil, "NotAnAlternative" })
		if err == nil || !strings.Contains(err.Error(), "NotAnAlternative") {
			t.Fatalf("err = %v, want refusal naming the unknown alternative", err)
		}
	})

	t.Run("trailing-call-tempdir", func(t *testing.T) {
		dir := t.TempDir()
		for _, name := range []string{"checksum.schway", "adler32.schway"} {
			data, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", name))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		source, err := os.ReadFile(filepath.Join(dir, "checksum.schway"))
		if err != nil {
			t.Fatal(err)
		}
		text := string(source)
		const header = "module examples.checksum.checksum\n"
		const body = "  let sum = adler32(view)\n  sum\n}\n"
		if strings.Count(text, header) != 1 || strings.Count(text, body) != 1 {
			t.Fatal("checksum.schway no longer has the expected header and main tail")
		}
		text = strings.Replace(text, header, "module examples.checksum.adler_low_probe\n", 1)
		text = strings.Replace(text, body, "  let sum = adler32(view)\n  let low = low16(sum)\n  low\n}\n\nfn low16(sum: U64) -> U64 {\n  var r = sum\n  r = r % 65536\n  r\n}\n", 1)
		probe := filepath.Join(dir, "adler_low_probe.schway")
		if err := os.WriteFile(probe, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		probed := phase31InterpCheck(t, probe)
		run, err := interp.RunWithFileBytes(probed.Program, "main", phase31InterpFile(t, "Wikipedia"), testsupport.ChecksumReadFile)
		if err != nil {
			t.Fatal(err)
		}
		if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != "920" || len(run.LiveResources) != 0 {
			t.Fatalf("outcome = %+v live=%v, want returned 920", run.Outcome, run.LiveResources)
		}
	})

	t.Run("skip-isregular-control", func(t *testing.T) {
		if _, alternative := testsupport.ChecksumReadFile("/dev/null"); alternative != "FileUnavailable" {
			t.Fatalf("ChecksumReadFile(/dev/null) alternative = %q, want FileUnavailable", alternative)
		}
		run, err := interp.RunWithFileBytes(checked.Program, "main", "/dev/null", phase31InterpNoRegularCheck)
		if err != nil {
			t.Fatal(err)
		}
		if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != "1" {
			t.Fatalf("control outcome = %+v, want returned 1", run.Outcome)
		}
	})

	t.Run("oracle-rows", func(t *testing.T) {
		count := 0
		for _, row := range testsupport.ChecksumOracleRows() {
			if row.App != "checksum" {
				continue
			}
			count++
			path := filepath.Join(t.TempDir(), "in.bin")
			if err := os.WriteFile(path, row.Input, 0o600); err != nil {
				t.Fatal(err)
			}
			run, err := interp.RunWithFileBytes(checked.Program, "main", path, testsupport.ChecksumReadFile)
			if err != nil {
				t.Fatalf("%s: %v", row.Name, err)
			}
			if len(run.LiveResources) != 0 {
				t.Fatalf("%s: live resources %v", row.Name, run.LiveResources)
			}
			if row.Outcome == "FileTooLarge" {
				if run.Outcome.Kind != execution.OutcomeTypedFailure || run.Outcome.Value != "FileTooLarge" {
					t.Fatalf("%s: outcome %+v, want typed_failure FileTooLarge", row.Name, run.Outcome)
				}
				continue
			}
			if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != strconv.FormatUint(row.Want, 10) {
				t.Fatalf("%s: outcome %+v, want returned %d", row.Name, run.Outcome, row.Want)
			}
			released := 0
			for _, event := range run.Events {
				if event.Kind == "resource.released" {
					released++
				}
			}
			if released != 1 {
				t.Fatalf("%s: release events = %d, want 1", row.Name, released)
			}
		}
		if count == 0 {
			t.Fatal("no checksum oracle rows")
		}
	})
}
