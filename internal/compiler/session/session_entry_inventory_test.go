package session_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const callgraphImportPath = "github.com/szTheory/schway/internal/compiler/callgraph"

// entryFunctionCallsIn returns "label:line" for every callgraph.EntryFunction
// selector call in src. It resolves the import alias and fails closed on a
// dot import, like the Phase 16 emitter inventory.
func entryFunctionCallsIn(t *testing.T, label, filename string, src any) []string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		t.Fatal(err)
	}
	aliases := map[string]bool{}
	for _, spec := range file.Imports {
		if strings.Trim(spec.Path.Value, "\"") != callgraphImportPath {
			continue
		}
		if spec.Name != nil && spec.Name.Name == "." {
			t.Fatalf("fail-closed: callgraph dot import in %s", label)
		}
		name := "callgraph"
		if spec.Name != nil {
			name = spec.Name.Name
		}
		aliases[name] = true
	}
	var calls []string
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "EntryFunction" {
			if receiver, ok := selector.X.(*ast.Ident); ok && aliases[receiver.Name] {
				calls = append(calls, label+":"+strconv.Itoa(fset.Position(call.Pos()).Line))
			}
		}
		return true
	})
	return calls
}

// TestEntryFunctionCalledOnlyViaResolveEntry enforces D-04: outside the
// callgraph package and test files, callgraph.EntryFunction is called exactly
// once in session_entry.go (entryOf) and exactly once in cgen_program.go
// (cgen cannot import session).
func TestEntryFunctionCalledOnlyViaResolveEntry(t *testing.T) {
	root := testsupport.ProjectPath()
	var calls []string
	for _, dir := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if filepath.ToSlash(path) == filepath.ToSlash(filepath.Join(root, "internal", "compiler", "callgraph")) {
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
			calls = append(calls, entryFunctionCallsIn(t, filepath.ToSlash(relative), path, nil)...)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Strings(calls)
	var files []string
	for _, call := range calls {
		files = append(files, call[:strings.LastIndex(call, ":")])
	}
	want := []string{"internal/compiler/cgen/cgen_program.go", "internal/compiler/session/session_entry.go"}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("callgraph.EntryFunction call sites = %v, want exactly one in each of %v", calls, want)
	}

	t.Run("negative-control", func(t *testing.T) {
		synthetic := "package x\nimport cg \"" + callgraphImportPath + "\"\nfunc f() { _, _ = cg.EntryFunction(nil) }\n"
		if got := entryFunctionCallsIn(t, "synthetic.go", "synthetic.go", synthetic); len(got) != 1 {
			t.Fatalf("a disallowed aliased call must be reported, got %v", got)
		}
	})
}
