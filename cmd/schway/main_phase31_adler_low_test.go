package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase31LowBuild(t *testing.T, name string) string {
	t.Helper()
	source := testsupport.ProjectPath("examples", "checksum", name+".schway")
	manifest := testsupport.ProjectPath("examples", "checksum", "file_bytes.bindings.json")
	artifact := filepath.Join(t.TempDir(), name)
	code, stdout, stderr := captureLangRun(t, []string{"build", source, "--manifest", manifest, "--output", artifact})
	if code != 0 || stdout != "built "+artifact+"\n" || stderr != "" {
		t.Fatalf("%s build code=%d stdout=%q stderr=%q", name, code, stdout, stderr)
	}
	return artifact
}

// TestPhase31AdlerLowPublicRoute pins the MOD-06 second consumer (D-03, D-16):
// adler_low imports the one library and prints its low 16 bits.
func TestPhase31AdlerLowPublicRoute(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("env:clang toolchain unavailable on this host")
	}
	artifact := phase31LowBuild(t, "adler_low")
	for _, row := range testsupport.ChecksumOracleRows() {
		if row.App != "adler_low" {
			continue
		}
		t.Run(row.Name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "in.bin")
			if err := os.WriteFile(path, row.Input, 0o600); err != nil {
				t.Fatal(err)
			}
			code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", path})
			if row.Outcome == "FileTooLarge" {
				if code != 65 || stdout != "" || stderr != "FileError.FileTooLarge\n" {
					t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
				}
				return
			}
			want := strconv.FormatUint(row.Want, 10) + "\n"
			if code != 0 || stdout != want || stderr != "" {
				t.Fatalf("code=%d stdout=%q stderr=%q, want 0 %q empty", code, stdout, stderr, want)
			}
		})
	}
	t.Run("same-file-different-result", func(t *testing.T) {
		checksum := phase31LowBuild(t, "checksum")
		in := phase31RouteInput(t, "Wikipedia")
		code, stdout, stderr := captureLangRun(t, []string{"app", "run", checksum, "--", in})
		if code != 0 || stdout != "300286872\n" || stderr != "" {
			t.Fatalf("checksum code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
		code, stdout, stderr = captureLangRun(t, []string{"app", "run", artifact, "--", in})
		if code != 0 || stdout != "920\n" || stderr != "" {
			t.Fatalf("adler_low code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})
}
