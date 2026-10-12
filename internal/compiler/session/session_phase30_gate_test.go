package session_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/syntax"
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
		"internal/compiler/modload/load.go",
		"internal/compiler/modload/merge.go",
		"internal/compiler/modload/resolve.go",
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

// gateFunctionSource builds a no-import source of exactly count trivial
// functions, exported entry first. withMain names the first function main.
func gateFunctionSource(module, prefix string, count int, withMain bool) []byte {
	var b strings.Builder
	entry := prefix + "0"
	if withMain {
		entry = "main"
	}
	fmt.Fprintf(&b, "module %s\n\nexport {\n  fn %s\n}\n\n", module, entry)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		if i == 0 && withMain {
			name = "main"
		}
		fmt.Fprintf(&b, "fn %s(n: U64) -> U64 {\n  n\n}\n\n", name)
	}
	return []byte(b.String())
}

// TestPhase30GateMergedProgramLimit is flag (f) of the D-14 gate (D-13, D-21).
// The per-file function cap is 1,024; a merge of two parsed programs that
// exceeds it reaches check.Program with no limit diagnostic today. Plan 30-04
// closes that gap with module.program_limit (D-21).
func TestPhase30GateMergedProgramLimit(t *testing.T) {
	const perFileCap = 1024
	if parsed := syntax.Parse(gateFunctionSource("gate.cap", "f", perFileCap, true)); len(parsed.Diagnostics) != 0 {
		t.Fatalf("%d functions must parse: %v", perFileCap, parsed.Diagnostics)
	} else if len(parsed.Program.Funcs) != perFileCap {
		t.Fatalf("parsed %d functions, want %d", len(parsed.Program.Funcs), perFileCap)
	}

	over := syntax.Parse(gateFunctionSource("gate.cap", "f", perFileCap+1, true))
	if len(over.Diagnostics) != 1 || over.Diagnostics[0].Code != "syntax.input_limit" {
		t.Fatalf("%d functions must be refused syntax.input_limit, got %v", perFileCap+1, over.Diagnostics)
	}

	first := syntax.Parse(gateFunctionSource("gate.first", "f", 600, true))
	second := syntax.Parse(gateFunctionSource("gate.second", "g", 425, false))
	if len(first.Diagnostics) != 0 || len(second.Diagnostics) != 0 {
		t.Fatalf("merge inputs must parse: %v %v", first.Diagnostics, second.Diagnostics)
	}
	merged := first.Program
	merged.Funcs = append(append(merged.Funcs[:0:0], merged.Funcs...), second.Program.Funcs...)
	if len(merged.Funcs) != perFileCap+1 {
		t.Fatalf("merged program has %d functions, want %d", len(merged.Funcs), perFileCap+1)
	}
	checked := check.Program(merged)
	for _, diag := range checked.Diagnostics {
		if diag.Code == "syntax.input_limit" || strings.HasSuffix(diag.Code, ".program_limit") {
			t.Fatalf("a merged-size limit already exists: %s", diag.Code)
		}
	}
	t.Logf("merged program of %d functions reached check.Program with %d diagnostics and no limit code", len(merged.Funcs), len(checked.Diagnostics))

	// One host, not pinned (D-28): logged only, never compared.
	started := time.Now()
	validated := corevalidate.Validate(checked.Program)
	t.Logf("corevalidate.Validate on the merged program: valid=%v elapsed=%s (one host, not pinned)", validated.Valid, time.Since(started))
}
