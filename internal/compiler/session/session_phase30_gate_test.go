package session_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const (
	gateDiagnosticImport = "github.com/szTheory/schway/internal/compiler/diagnostic"
	gateDebugmapImport   = "github.com/szTheory/schway/internal/compiler/debugmap"
)

// gateUnitConsumerFile reports whether src names diagnostic.Diagnostic,
// diagnostic.Cause, diagnostic.Repair or debugmap.Entry through an import,
// resolving aliases and failing closed on dot imports.
func gateUnitConsumerFile(t *testing.T, label string, src any) bool {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, label, src, 0)
	if err != nil {
		t.Fatal(err)
	}
	diagnosticNames := map[string]bool{}
	debugmapNames := map[string]bool{}
	for _, spec := range file.Imports {
		path := strings.Trim(spec.Path.Value, "\"")
		var set map[string]bool
		var base string
		switch path {
		case gateDiagnosticImport:
			set, base = diagnosticNames, "diagnostic"
		case gateDebugmapImport:
			set, base = debugmapNames, "debugmap"
		default:
			continue
		}
		name := base
		if spec.Name != nil {
			if spec.Name.Name == "." {
				t.Fatalf("fail-closed: dot import of %s in %s", path, label)
			}
			name = spec.Name.Name
		}
		set[name] = true
	}
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		receiver, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}
		switch {
		case diagnosticNames[receiver.Name]:
			switch selector.Sel.Name {
			case "Diagnostic", "Cause", "Repair":
				found = true
			}
		case debugmapNames[receiver.Name]:
			if selector.Sel.Name == "Entry" {
				found = true
			}
		}
		return true
	})
	return found
}

// TestPhase30GateUnitConsumerInventory is the consumer half of flag (c): the
// non-test files outside diagnostic and debugmap that name the identity-bearing
// diagnostic types or debugmap.Entry are pinned, so plan 30-02's Unit plumbing
// knows every consumer it can touch.
func TestPhase30GateUnitConsumerInventory(t *testing.T) {
	root := testsupport.ProjectPath()
	var files []string
	for _, dir := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			slash := filepath.ToSlash(path)
			if entry.IsDir() {
				if strings.HasSuffix(slash, "/internal/compiler/diagnostic") {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			relative, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			if gateUnitConsumerFile(t, path, nil) {
				files = append(files, filepath.ToSlash(relative))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Strings(files)
	// Cross-checked against RESEARCH "Unit plumbing" (2026-10-12). Listed there but
	// not observed here: debugmap/debugmap.go (defines Entry, so it names it
	// unqualified), pathoracle.go (no import-qualified use), and
	// cmd/schway-repair/repair.go (decodes JSON, imports no compiler package by
	// design). Observed but not named in RESEARCH: check/scalar.go,
	// session_app_verify.go, session_phase5*.go, session_phase6_evidence.go and
	// syntax/lexer.go.
	want := []string{
		"cmd/schway/main.go",
		"internal/compiler/check/check.go",
		"internal/compiler/check/scalar.go",
		"internal/compiler/evidence/evidence.go",
		"internal/compiler/protocol/protocol.go",
		"internal/compiler/reduce/mismatch.go",
		"internal/compiler/session/session.go",
		"internal/compiler/session/session_admit.go",
		"internal/compiler/session/session_app_verify.go",
		"internal/compiler/session/session_entry.go",
		"internal/compiler/session/session_phase5.go",
		"internal/compiler/session/session_phase5_compare.go",
		"internal/compiler/session/session_phase5_mismatch.go",
		"internal/compiler/session/session_phase6_evidence.go",
		"internal/compiler/session/session_phase6_explain.go",
		"internal/compiler/session/session_phase6_injectors.go",
		"internal/compiler/session/session_phase6_query.go",
		"internal/compiler/syntax/lexer.go",
		"internal/compiler/syntax/parser.go",
	}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("Unit consumer inventory changed:\n got %q\nwant %q", files, want)
	}

	t.Run("negative-control", func(t *testing.T) {
		synthetic := "package x\nimport d \"" + gateDiagnosticImport + "\"\nvar _ d.Cause\n"
		if !gateUnitConsumerFile(t, "synthetic.go", synthetic) {
			t.Fatal("a synthetic file naming diagnostic.Cause must be reported")
		}
		unrelated := "package x\nimport d \"" + gateDiagnosticImport + "\"\nvar _ = d.Error\n"
		if gateUnitConsumerFile(t, "unrelated.go", unrelated) {
			t.Fatal("a file naming only diagnostic.Error must not be reported")
		}
	})
}
