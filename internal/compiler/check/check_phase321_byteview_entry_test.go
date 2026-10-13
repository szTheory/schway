package check_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase321ByteViewMain = `module test.byteview_main

export {
  fn main
}

fn main(b: ByteView) -> U64 {
  var i = 0
  var n = byte_len(b)
  while i < n {
    i = i + 1
  }
  i
}
`

func phase321Write(t *testing.T, dir, name, source string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestPhase321ByteViewEntryRefused pins D-08/D-22: the entry-module main with a
// ByteView parameter is refused with check.unexecutable_shape, while library
// and non-entry ByteView functions still check (armed refuse/accept pair).
func TestPhase321ByteViewEntryRefused(t *testing.T) {
	dir := t.TempDir()

	refused := phase321Write(t, dir, "byteview_main.schway", phase321ByteViewMain)
	checked, err := session.CheckFile(refused)
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 1 {
		t.Fatalf("want exactly one diagnostic, got %+v", checked.Diagnostics)
	}
	problem := checked.Diagnostics[0]
	if problem.Code != "check.unexecutable_shape" || problem.Schema != "lang.diagnostic/1" {
		t.Fatalf("unexpected diagnostic %+v", problem)
	}
	if problem.Primary.Start == 0 && problem.Primary.End == 0 {
		t.Fatalf("diagnostic carries no span: %+v", problem)
	}
	source := phase321ByteViewMain
	if got := source[problem.Primary.Start:problem.Primary.End]; got != "b" {
		t.Fatalf("primary span %q does not cover the parameter", got)
	}
	if len(problem.Repairs) != 1 || problem.Repairs[0].Kind != "use_executable_shape" || !strings.Contains(problem.Repairs[0].Detail, "U64") {
		t.Fatalf("missing repair naming supported inputs: %+v", problem.Repairs)
	}
	if len(checked.Program.Functions) != 0 {
		t.Fatalf("refused function admitted: %+v", checked.Program.Functions)
	}

	// Non-entry ByteView function (renamed) still checks.
	renamed := phase321Write(t, dir, "byteview_lib.schway",
		strings.NewReplacer("test.byteview_main", "test.byteview_lib", "fn main", "fn scan").Replace(phase321ByteViewMain))
	assertCheckClean(t, renamed)

	// Imported module with main(b: ByteView) is not the entry module.
	phase321Write(t, dir, "byteview_dep.schway", `module byteview_dep

export {
  fn helper
}

fn helper(n: U64) -> U64 {
  n
}

fn main(b: ByteView) -> U64 {
  let n = byte_len(b)
  n
}
`)
	entry := phase321Write(t, dir, "entry_u64.schway", `module test.entry_u64

import byteview_dep {
  fn helper
}

export {
  fn main
}

fn main(n: U64) -> U64 {
  let r = helper(n)
  r
}
`)
	assertCheckClean(t, entry)

	// Entry module with imports and a ByteView main is refused.
	phase321Write(t, dir, "lib2.schway", `module lib2

export {
  fn helper
}

fn helper(n: U64) -> U64 {
  n
}
`)
	importing := phase321Write(t, dir, "entry_bv.schway", `module test.entry_bv

import lib2 {
  fn helper
}

export {
  fn main
}

fn main(b: ByteView) -> U64 {
  let n = byte_len(b)
  n
}
`)
	checked, err = session.CheckFile(importing)
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "check.unexecutable_shape" {
		t.Fatalf("importing ByteView main must be refused: %+v", checked.Diagnostics)
	}

	// D-08 negative controls.
	for _, rel := range []string{
		"examples/checksum/adler32.schway",
		"examples/checksum/checksum.schway",
		"examples/checksum/adler32_cli.schway",
		"examples/checksum/adler_low.schway",
		"examples/phase29/adler32_fixture.schway",
	} {
		assertCheckClean(t, testsupport.ProjectPath(strings.Split(rel, "/")...))
	}
}

func assertCheckClean(t *testing.T, path string) {
	t.Helper()
	checked, err := session.CheckFile(path)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s: want clean check, got %+v", path, checked.Diagnostics)
	}
}
