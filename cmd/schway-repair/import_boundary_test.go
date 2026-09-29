package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// forbiddenImportSubstr is the BROAD form D-06-28 requires: "never import
// internal/", not one named sibling package -- unlike
// TestValidatorImportsStayIndependent's own two-named-package form.
const forbiddenImportSubstr = "/internal/"

// scanForbiddenImports parses every non-test .go file directly inside dir
// (go/parser.ImportsOnly, mirroring TestValidatorImportsStayIndependent's
// and TestOracleImportsStayIndependent's own established mechanism) and
// returns one violation string per forbidden import found, naming both the
// file and the import path.
func scanForbiddenImports(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fileSet := token.NewFileSet()
	var violations []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			return nil, err
		}
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			if strings.Contains(path, forbiddenImportSubstr) || strings.HasPrefix(path, "internal/") {
				violations = append(violations, fmt.Sprintf("%s imports %s, which cmd/schway-repair must never depend on (D-06-28)", entry.Name(), path))
			}
		}
	}
	return violations, nil
}

// TestRepairDriverImportsStayOutsideInternal is D-06-28's structural
// boundary: no non-test .go file under cmd/schway-repair may import anything
// under internal/. Demonstrated live (not merely asserted) during 06-13's
// execution by seeding a violation, confirming this test fails the build,
// and reverting -- recorded in 06-13-SUMMARY.md.
func TestRepairDriverImportsStayOutsideInternal(t *testing.T) {
	violations, err := scanForbiddenImports(repairPackageDir(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("cmd/schway-repair imports internal/:\n%s", strings.Join(violations, "\n"))
	}
}

// TestPhase17HeldoutPathsAbsentFromProduction prevents the repair driver from
// acquiring knowledge of the sealed TYP-05 corpus. Tests may name a fixture;
// non-test command sources must operate solely on the caller-supplied path.
func TestPhase17HeldoutPathsAbsentFromProduction(t *testing.T) {
	for _, forbidden := range []string{"testdata/phase17", "heldout_call_argument_mismatch.schway"} {
		forEachNonTestFile(t, repairPackageDir(t), func(name string, _ *token.FileSet, file *ast.File) {
			ast.Inspect(file, func(n ast.Node) bool {
				literal, ok := n.(*ast.BasicLit)
				if ok && literal.Kind == token.STRING && strings.Contains(literal.Value, forbidden) {
					t.Fatalf("%s names sealed Phase 17 fixture material %q", name, forbidden)
				}
				return true
			})
		})
	}
}

// TestImportBoundaryTestIsNotInert proves scanForbiddenImports actually
// detects what it claims to detect, rather than passing vacuously on an
// empty or well-behaved file list (this project's own three-gate-failure
// lesson: a green test whose reachable input space omitted the hard case).
func TestImportBoundaryTestIsNotInert(t *testing.T) {
	dir := t.TempDir()
	fixture := "package fixture\n\nimport \"github.com/szTheory/schway/internal/compiler/protocol\"\n\nvar _ = protocol.StatusPass\n"
	mustWriteFile(t, filepath.Join(dir, "fixture.go"), []byte(fixture))
	violations, err := scanForbiddenImports(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) == 0 {
		t.Fatal("scanForbiddenImports did not detect a forbidden import in a positive-control fixture -- the scan is inert")
	}
	if !strings.Contains(violations[0], "fixture.go") || !strings.Contains(violations[0], "internal/compiler/protocol") {
		t.Fatalf("violation message does not name both the file and the import path: %q", violations[0])
	}
}

// selectorCallIdent returns the receiver identifier name of a call of the
// shape pkg.Func(...), or "" if call is not that shape.
func selectorCallIdent(call *ast.CallExpr) (pkg, fn string, ok bool) {
	sel, isSel := call.Fun.(*ast.SelectorExpr)
	if !isSel {
		return "", "", false
	}
	ident, isIdent := sel.X.(*ast.Ident)
	if !isIdent {
		return "", "", false
	}
	return ident.Name, sel.Sel.Name, true
}

// TestRepairDriverNeverOpensSourceOutsideSpan covers both halves of
// T-06-BOUNDARY-02: statically, that every os.ReadFile/os.Open/os.Create/
// os.WriteFile call site in the driver takes a variable (never a hardcoded
// literal path other than the command-line source); and at runtime, that a
// repair whose span exceeds the source length is REFUSED with a named
// error rather than clamped or partially applied.
func TestRepairDriverNeverOpensSourceOutsideSpan(t *testing.T) {
	t.Run("static: file I/O call sites never use a hardcoded literal path", func(t *testing.T) {
		forEachNonTestFile(t, repairPackageDir(t), func(name string, fset *token.FileSet, file *ast.File) {
			ast.Inspect(file, func(n ast.Node) bool {
				call, isCall := n.(*ast.CallExpr)
				if !isCall {
					return true
				}
				pkg, fn, ok := selectorCallIdent(call)
				if !ok || pkg != "os" {
					return true
				}
				switch fn {
				case "ReadFile", "Open", "WriteFile", "Create":
					if len(call.Args) == 0 {
						return true
					}
					if _, isLit := call.Args[0].(*ast.BasicLit); isLit {
						t.Fatalf("%s:%s calls os.%s with a hardcoded literal path -- must take the single command-line source path",
							name, fset.Position(call.Pos()), fn)
					}
				}
				return true
			})
		})
	})

	t.Run("runtime: an out-of-range span is refused, not clamped", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "source.schway")
		original := []byte("fn identity(value: Byte) -> Byte {\n  value\n}\n")
		mustWriteFile(t, path, original)
		outOfRange := jsonRepair{Kind: "synthetic", Span: &jsonSpan{Start: 0, End: len(original) + 100}, Replacement: "x", Applicability: applicabilityMachineApplicable}
		err := applyRepair(path, outOfRange)
		if err == nil {
			t.Fatal("expected an out-of-range span to be refused, got nil error")
		}
		var driverErr *DriverError
		if !errors.As(err, &driverErr) || driverErr.Code != CodeSpanOutOfRange {
			t.Fatalf("expected %s, got %v", CodeSpanOutOfRange, err)
		}
		after, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if string(after) != string(original) {
			t.Fatal("file was modified despite an out-of-range span being refused")
		}
	})
}

// TestRepairDriverSpawnsOnlyTheLangBinary asserts every exec.Command/
// exec.CommandContext call site in the driver uses the --lang path
// variable, never a literal program name (T-06-BOUNDARY-06).
func TestRepairDriverSpawnsOnlyTheLangBinary(t *testing.T) {
	forEachNonTestFile(t, repairPackageDir(t), func(name string, fset *token.FileSet, file *ast.File) {
		ast.Inspect(file, func(n ast.Node) bool {
			call, isCall := n.(*ast.CallExpr)
			if !isCall {
				return true
			}
			pkg, fn, ok := selectorCallIdent(call)
			if !ok || pkg != "exec" {
				return true
			}
			programArgIndex := -1
			switch fn {
			case "Command":
				programArgIndex = 0
			case "CommandContext":
				programArgIndex = 1
			default:
				return true
			}
			if len(call.Args) <= programArgIndex {
				return true
			}
			if _, isLit := call.Args[programArgIndex].(*ast.BasicLit); isLit {
				t.Fatalf("%s:%s calls exec.%s with a literal program name -- must use the --lang path variable",
					name, fset.Position(call.Pos()), fn)
			}
			return true
		})
	})
}
