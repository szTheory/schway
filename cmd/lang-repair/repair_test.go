package main

import (
	"bytes"
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// This file is a _test.go file, so it is explicitly EXEMPT from the
// import-boundary lint (import_boundary_test.go scans only non-test files):
// it imports internal/ packages to DRIVE the fixtures and injectors the
// real driver (repair.go, main.go) never touches, exactly the way
// session_phase6_injectors_test.go already imports internal packages to
// drive AllInjectors() while the injectors themselves stay production code.

func phase6Fixture(t testing.TB, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase6", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mustWriteFile(t testing.TB, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func decodeCheckJSON(t testing.TB, raw []byte) checkResult {
	t.Helper()
	var decoded checkResult
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decoding lang --json check output: %v (raw: %s)", err, raw)
	}
	return decoded
}

// repairPackageDir is cmd/lang-repair's own directory, resolved the same
// way testsupport.ProjectPath resolves every other fixture and package
// path in the tree. This file is itself a _test.go file, so importing
// internal/compiler/testsupport here is allowed -- the import boundary
// (import_boundary_test.go) applies only to cmd/lang-repair's non-test
// files (main.go, repair.go), never to the tests that scan them.
func repairPackageDir(t testing.TB) string {
	t.Helper()
	return testsupport.ProjectPath("cmd", "lang-repair")
}

// forEachNonTestFile walks every non-test .go file directly inside dir,
// fully parsed (not ImportsOnly), calling visit on each parsed file. Shared
// by every go/ast structural test in this package
// (TestRepairDriverDecodesNoProseFields,
// TestRepairDriverSourceNeverReferencesHeldoutFixtures, and
// import_boundary_test.go's own boundary tests).
func forEachNonTestFile(t testing.TB, dir string, visit func(name string, fset *token.FileSet, file *ast.File)) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		visit(entry.Name(), fset, file)
	}
}

// TestRepairDriverFixesOneDefectEndToEnd is the tracer (Task 1): one
// heldout_move_*.lang fixture, mutated by the REAL move injector, repaired
// by the REAL driver spawning the REAL built `lang` binary as a subprocess,
// and independently re-verified clean by a separate `lang --json check`
// invocation. Real binary, real subprocess, real JSON, one class, one path.
func TestRepairDriverFixesOneDefectEndToEnd(t *testing.T) {
	langBinary := testsupport.BuildCLI(t)

	original := phase6Fixture(t, "heldout_move_defect.lang")
	mutated, err := session.MoveInjector{}.Inject(original)
	if err != nil {
		t.Fatalf("injecting move defect: %v", err)
	}

	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "heldout_move_defect.lang")
	mustWriteFile(t, sourcePath, mutated)

	outcome, err := Repair(context.Background(), langBinary, sourcePath)
	if err != nil {
		t.Fatalf("Repair: %v", err)
	}
	if outcome.Status != OutcomeRepaired {
		t.Fatalf("got status %q, want %q (diagnosis=%q repair=%q)", outcome.Status, OutcomeRepaired, outcome.DiagnosisCode, outcome.RepairKind)
	}

	repaired, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(repaired, original) {
		t.Fatal("repaired bytes are not byte-identical to the pre-defect original")
	}

	// Independent re-verification: a SECOND, wholly separate `lang --json
	// check` subprocess invocation, outside the driver's own internal
	// reverify, reports the post-repair source clean.
	verify := testsupport.RunCLI(t, langBinary, nil, "--json", "check", sourcePath)
	decoded := decodeCheckJSON(t, verify.Stdout)
	if decoded.Status != statusPass {
		t.Fatalf("independent re-verification reports status %q, want %q", decoded.Status, statusPass)
	}
}

// TestRepairDriverDecodesNoProseFields asserts, by go/ast, that no struct
// declared in cmd/lang-repair binds either prose field of the diagnostic
// document -- the human-readable "message", or any cause/repair "detail" --
// to a JSON tag. A field that is never decoded cannot be scraped
// (T-06-BOUNDARY-03).
func TestRepairDriverDecodesNoProseFields(t *testing.T) {
	forEachNonTestFile(t, repairPackageDir(t), func(name string, fset *token.FileSet, file *ast.File) {
		ast.Inspect(file, func(n ast.Node) bool {
			st, isStruct := n.(*ast.StructType)
			if !isStruct {
				return true
			}
			for _, field := range st.Fields.List {
				if field.Tag == nil {
					continue
				}
				tag := strings.Trim(field.Tag.Value, "`")
				if strings.Contains(tag, `json:"message`) || strings.Contains(tag, `json:"detail`) {
					t.Fatalf("%s:%s declares a struct field bound to a prose JSON field (%s) -- the driver may only decode structural fields",
						name, fset.Position(field.Pos()), tag)
				}
			}
			return true
		})
	})
}

// TestRepairDriverBoundedReaderRejectsOversizedStdout demonstrates the
// bounded-writer half of T-06-BOUNDARY-05 directly: a probe emitting more
// than the declared cap is observably overflowed rather than silently
// accumulated without limit.
func TestRepairDriverBoundedReaderRejectsOversizedStdout(t *testing.T) {
	writer := newBoundedWriter(8)
	oversized := bytes.Repeat([]byte("a"), 16)
	if _, err := writer.Write(oversized); err != nil {
		t.Fatalf("Write must never itself error (the subprocess must not see a write failure): %v", err)
	}
	if !writer.overflowed() {
		t.Fatal("expected the writer to report overflow for a stream larger than its declared cap")
	}

	// Boundary: exactly at the cap must NOT overflow.
	exact := newBoundedWriter(8)
	if _, err := exact.Write(bytes.Repeat([]byte("a"), 8)); err != nil {
		t.Fatal(err)
	}
	if exact.overflowed() {
		t.Fatal("stdout exactly at the cap must not be reported as overflowed")
	}
}

// TestRepairDriverSourceNeverReferencesHeldoutFixtures asserts, by go/ast,
// that no non-test file under cmd/lang-repair names any heldout_ fixture
// path -- the driver's kind-to-edit mapping (README's rule) may consult
// only derivation_ fixtures, if it consults any fixture at all. The
// shipped driver consults none (it is fully generic over Span/Replacement),
// which trivially satisfies this, but the test still stands as a live
// falsifier against a future regression that hardcodes a heldout_ path.
func TestRepairDriverSourceNeverReferencesHeldoutFixtures(t *testing.T) {
	forEachNonTestFile(t, repairPackageDir(t), func(name string, fset *token.FileSet, file *ast.File) {
		ast.Inspect(file, func(n ast.Node) bool {
			lit, isLit := n.(*ast.BasicLit)
			if !isLit || lit.Kind != token.STRING {
				return true
			}
			if strings.Contains(lit.Value, "heldout_") {
				t.Fatalf("%s:%s references a heldout_ fixture path %s -- the driver's mapping may consult only derivation_ fixtures",
					name, fset.Position(lit.Pos()), lit.Value)
			}
			return true
		})
	})
}
