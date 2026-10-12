package session_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// Phase 30 runnable gain (SC1, D-17): the Phase 29 Adler-32 program with the
// library in its own module file, imported rather than copied. The pinned
// answers come from testsupport.Phase29FixtureExpected (independent literals),
// never from Schway itself.

const phase30AdlerLibraryName = "adler32.schway"
const phase30AdlerEntryName = "adler32_cli.schway"

func phase30AdlerDir() string {
	return testsupport.ProjectPath("examples", "checksum")
}

// phase30CopyDir copies every file of src into a fresh TempDir and returns it.
func phase30CopyDir(t *testing.T, src string) string {
	t.Helper()
	dst := t.TempDir()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dst, entry.Name()), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dst
}

// phase30MutateLibrary replaces the single `var a = 1` line of the copied
// library with `var a = 2`, which changes every case including the empty one.
func phase30MutateLibrary(t *testing.T, dir string) {
	t.Helper()
	path := filepath.Join(dir, phase30AdlerLibraryName)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Count(text, "var a = 1\n") != 1 {
		t.Fatalf("library no longer has exactly one `var a = 1` line")
	}
	if err := os.WriteFile(path, []byte(strings.Replace(text, "var a = 1\n", "var a = 2\n", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
}

// adlerInterpMismatches runs every pinned case through the interpreter and
// returns the case ids whose answer differs from the pinned literal.
func adlerInterpMismatches(t *testing.T, entryPath string) []int {
	t.Helper()
	checked, err := session.CheckFile(entryPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics for %s: %+v", entryPath, checked.Diagnostics)
	}
	var mismatches []int
	for _, caseID := range testsupport.Phase29FixtureCaseIDs() {
		want, ok := testsupport.Phase29FixtureExpected(caseID)
		if !ok {
			t.Fatalf("case %d has no pinned checksum", caseID)
		}
		run, err := interp.RunWithFixtureBytes(checked.Program, "main", strconv.FormatUint(caseID, 10), testsupport.Phase29FixtureBytes)
		if err != nil {
			t.Fatalf("case %d: %v", caseID, err)
		}
		if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != strconv.FormatUint(want, 10) || len(run.LiveResources) != 0 {
			mismatches = append(mismatches, int(caseID))
		}
	}
	return mismatches
}

// TestPhase30AdlerModuleInterpreter is the tracer: the imported library gives
// the ten pinned Phase 29 answers in the interpreter, and the merged
// program's adler32 belongs to module adler32 (D-17).
func TestPhase30AdlerModuleInterpreter(t *testing.T) {
	entry := filepath.Join(phase30AdlerDir(), phase30AdlerEntryName)
	checked, err := session.CheckFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}
	found := 0
	for _, function := range checked.Program.Functions {
		if function.Name == "adler32" {
			found++
			if function.Module != "adler32" {
				t.Fatalf("adler32 function Module = %q, want adler32", function.Module)
			}
		}
	}
	if found != 1 {
		t.Fatalf("merged program has %d adler32 functions, want 1", found)
	}
	// The tracer: case 3 ("Wikipedia") first.
	tracer, err := interp.RunWithFixtureBytes(checked.Program, "main", "3", testsupport.Phase29FixtureBytes)
	if err != nil {
		t.Fatal(err)
	}
	if tracer.Outcome.Kind != execution.OutcomeReturned || tracer.Outcome.Value != "300286872" || len(tracer.LiveResources) != 0 {
		t.Fatalf("case 3 = %+v live=%v, want returned 300286872 with no live resources", tracer.Outcome, tracer.LiveResources)
	}
	if mismatches := adlerInterpMismatches(t, entry); len(mismatches) != 0 {
		t.Fatalf("interpreter answers differ from the pinned values for cases %v", mismatches)
	}

	t.Run("mutated-library", func(t *testing.T) {
		dir := phase30CopyDir(t, phase30AdlerDir())
		phase30MutateLibrary(t, dir)
		mismatches := adlerInterpMismatches(t, filepath.Join(dir, phase30AdlerEntryName))
		if len(mismatches) != len(testsupport.Phase29FixtureCaseIDs()) {
			t.Fatalf("mutated library differs for cases %v, want all %d", mismatches, len(testsupport.Phase29FixtureCaseIDs()))
		}
	})
}

// phase30AdlerDefinitions lists the .schway files under dir that define
// `fn adler32(`.
func phase30AdlerDefinitions(t *testing.T, dir string) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".schway" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(data), "fn adler32(") {
			found = append(found, filepath.Base(path))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return found
}

// TestPhase30LibraryIsSingleCopy pins that the library is imported, never
// copied (T-30-33): exactly one file defines fn adler32 and the merged
// program has exactly one such function, in module adler32.
func TestPhase30LibraryIsSingleCopy(t *testing.T) {
	dir := phase30AdlerDir()
	defs := phase30AdlerDefinitions(t, dir)
	if len(defs) != 1 || defs[0] != phase30AdlerLibraryName {
		t.Fatalf("files defining fn adler32( = %v, want exactly [%s]", defs, phase30AdlerLibraryName)
	}
	entryBytes, err := os.ReadFile(filepath.Join(dir, phase30AdlerEntryName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(entryBytes), "import adler32 {") {
		t.Fatal("entry does not import adler32")
	}
	checked, err := session.CheckFile(filepath.Join(dir, phase30AdlerEntryName))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, function := range checked.Program.Functions {
		if function.Name == "adler32" {
			count++
			if function.Module != "adler32" {
				t.Fatalf("adler32 Module = %q, want adler32", function.Module)
			}
		}
	}
	if count != 1 {
		t.Fatalf("merged program has %d adler32 functions, want 1", count)
	}

	t.Run("copied-library-is-detected", func(t *testing.T) {
		clone := phase30CopyDir(t, dir)
		library, err := os.ReadFile(filepath.Join(clone, phase30AdlerLibraryName))
		if err != nil {
			t.Fatal(err)
		}
		text := string(library)
		body := text[strings.Index(text, "fn adler32("):]
		entryPath := filepath.Join(clone, phase30AdlerEntryName)
		entry, err := os.ReadFile(entryPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(entryPath, append(entry, []byte("\n"+body)...), 0o600); err != nil {
			t.Fatal(err)
		}
		if defs := phase30AdlerDefinitions(t, clone); len(defs) != 2 {
			t.Fatalf("clone with a copied library has definitions in %v, want two files", defs)
		}
	})
}
