package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPhase321ByteViewEntryCLI pins the D-08/D-01 seam: a ByteView main is a
// source refusal (exit 2) on check and on both run engines, never 64.
func TestPhase321ByteViewEntryCLI(t *testing.T) {
	dir := t.TempDir()
	bv := filepath.Join(dir, "byteview_main.schway")
	source := "module test.byteview_cli\n\nexport {\n  fn main\n}\n\nfn main(b: ByteView) -> U64 {\n  let n = byte_len(b)\n  n\n}\n"
	if err := os.WriteFile(bv, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"--json", "check", bv},
		{"--json", "run", "--engine=interpreter", bv},
		{"--json", "run", "--engine=native", bv},
	} {
		code, stdout, stderr := captureLangRun(t, args)
		if code != 2 {
			t.Fatalf("%v: exit %d, want 2 (stdout %s stderr %s)", args, code, stdout, stderr)
		}
		if !strings.Contains(stdout, "check.unexecutable_shape") {
			t.Fatalf("%v: missing code in %s", args, stdout)
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "..", "examples", "sum_to_n.schway"))
	if err != nil {
		t.Fatal(err)
	}
	ok := filepath.Join(dir, "sum_to_n.schway")
	if err := os.WriteFile(ok, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := captureLangRun(t, []string{"--json", "check", ok}); code != 0 {
		t.Fatalf("U64 main check: exit %d (stdout %s stderr %s)", code, stdout, stderr)
	}
}
