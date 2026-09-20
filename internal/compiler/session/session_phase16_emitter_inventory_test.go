package session_test

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

const cgenImportPath = "github.com/codename-lang/lang/internal/compiler/cgen"

// phase16ConsumerRegistry is intentionally small: the source-derived list is
// the authority and this file pins its cardinality and classification policy.
// A new callsite cannot silently become trusted just because it compiles.
type phase16ConsumerRegistry struct {
	Schema        string `json:"schema"`
	ExpectedCount int    `json:"expected_count"`
}

func phase16PublicEmitterCalls(t *testing.T, root string) []string {
	t.Helper()
	var calls []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_generated.go") {
			return err
		}
		fset := token.NewFileSet()
		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		aliases := map[string]bool{}
		for _, spec := range file.Imports {
			if strings.Trim(spec.Path.Value, "\"") != cgenImportPath {
				continue
			}
			if spec.Name != nil && spec.Name.Name == "." {
				t.Fatalf("fail-closed: cgen dot import in %s", path)
			}
			name := "cgen"
			if spec.Name != nil {
				name = spec.Name.Name
			}
			aliases[name] = true
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			name := ""
			switch fn := call.Fun.(type) {
			case *ast.SelectorExpr:
				if receiver, ok := fn.X.(*ast.Ident); ok && aliases[receiver.Name] && (fn.Sel.Name == "Emit" || fn.Sel.Name == "EmitNative") {
					name = fn.Sel.Name
				}
			case *ast.Ident:
				// Only package cgen can call its own exported functions without
				// a selector; local shadows are deliberately not counted.
				if file.Name.Name == "cgen" && (fn.Name == "Emit" || fn.Name == "EmitNative") {
					name = fn.Name
				}
			}
			if name != "" {
				position := fset.Position(call.Pos())
				calls = append(calls, filepath.ToSlash(strings.TrimPrefix(path, testsupport.ProjectPath()+string(filepath.Separator)))+":"+name+":"+itoa(position.Line))
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(calls)
	return calls
}

func itoa(v int) string { return strconv.Itoa(v) }

func TestPhase16PublicEmitterConsumerInventory(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase16", "public-emitter-consumers.json"))
	if err != nil {
		t.Fatal(err)
	}
	var registry phase16ConsumerRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		t.Fatal(err)
	}
	if registry.Schema != "phase16.public-emitter-consumers/1" {
		t.Fatalf("unexpected registry schema %q", registry.Schema)
	}
	calls := phase16PublicEmitterCalls(t, testsupport.ProjectPath("internal", "compiler"))
	if len(calls) != registry.ExpectedCount {
		t.Fatalf("public emitter inventory drift: got %d calls, registry permits %d; first=%v", len(calls), registry.ExpectedCount, calls[:min(5, len(calls))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
