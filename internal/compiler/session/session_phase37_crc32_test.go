package session_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase37CRC32NativeTracer(t *testing.T) {
	exampleRoot := testsupport.ProjectPath("examples", "crc32")
	entryPath := filepath.Join(exampleRoot, "crc32.schway")
	libraryPath := filepath.Join(exampleRoot, "lib", "crc32.schway")
	entrySource, err := os.ReadFile(entryPath)
	if err != nil {
		t.Fatal(err)
	}
	librarySource, err := os.ReadFile(libraryPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(entrySource), "import lib.crc32 {") != 1 {
		t.Fatal("CRC entry must import lib.crc32 exactly once")
	}
	if !strings.Contains(string(librarySource), "export {\n  fn crc32\n}") || !strings.Contains(string(librarySource), "fn crc_byte(") {
		t.Fatal("CRC library must export crc32 and keep its crc_byte helper private")
	}

	checked, err := session.CheckFile(entryPath)
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("check CRC entry: err=%v diagnostics=%+v", err, checked.Diagnostics)
	}
	crc32Functions := 0
	for _, function := range checked.Program.Functions {
		if function.Name == "crc32" {
			crc32Functions++
			if function.Module != "lib.crc32" {
				t.Fatalf("crc32 module = %q, want lib.crc32", function.Module)
			}
		}
	}
	if crc32Functions != 1 {
		t.Fatalf("merged program has %d crc32 definitions, want one imported implementation", crc32Functions)
	}

	artifact := filepath.Join(t.TempDir(), "crc32")
	manifest := testsupport.ProjectPath("examples", "checksum", "file_bytes.bindings.json")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), entryPath, artifact, native.DefaultRunner(), manifest)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("build CRC app: err=%v diagnostics=%+v", err, diagnostics)
	}
	input := filepath.Join(t.TempDir(), "input.bin")
	if err := os.WriteFile(input, []byte("123456789"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, input, &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 {
		t.Fatalf("native app outcome=%+v err=%v stdout=%q stderr=%q", outcome, err, stdout.String(), stderr.String())
	}
	if stdout.String() != "3421780262\n" || stderr.String() != "" {
		t.Fatalf("native CRC result stdout=%q stderr=%q; want %q and empty stderr", stdout.String(), stderr.String(), "3421780262\n")
	}
}
