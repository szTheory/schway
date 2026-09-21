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

const (
	cgenImportPath           = "github.com/codename-lang/lang/internal/compiler/cgen"
	phase16AdmittedDynamic   = "admitted-dynamic-schema2"
	phase16RefusalWithFrozen = "refusal-frozen-witness"
)

type phase16ConsumerRegistry struct {
	Schema  string                       `json:"schema"`
	Entries []phase16ConsumerRegistryRow `json:"entries"`
}

type phase16ConsumerRegistryRow struct {
	Call           string `json:"call"`
	Classification string `json:"classification"`
	Witness        string `json:"witness,omitempty"`
}

// phase16PublicEmitterCalls resolves import aliases, same-package calls, and
// rejects dot imports. Non-cgen unqualified identifiers are local shadows.
func phase16PublicEmitterCalls(t *testing.T, root string) []string {
	t.Helper()
	var calls []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_generated.go") {
			return err
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
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
			switch fun := call.Fun.(type) {
			case *ast.SelectorExpr:
				if receiver, ok := fun.X.(*ast.Ident); ok && aliases[receiver.Name] && (fun.Sel.Name == "Emit" || fun.Sel.Name == "EmitNative") {
					name = fun.Sel.Name
				}
			case *ast.Ident:
				if file.Name.Name == "cgen" && (fun.Name == "Emit" || fun.Name == "EmitNative") {
					name = fun.Name
				}
			}
			if name != "" {
				relative, relErr := filepath.Rel(testsupport.ProjectPath(), path)
				if relErr != nil {
					t.Fatal(relErr)
				}
				calls = append(calls, filepath.ToSlash(relative)+":"+name+":"+strconv.Itoa(fset.Position(call.Pos()).Line))
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

func TestPhase16PublicEmitterConsumerInventory(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase16", "public-emitter-consumers.json"))
	if err != nil {
		t.Fatal(err)
	}
	var registry phase16ConsumerRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		t.Fatal(err)
	}
	if registry.Schema != "phase16.public-emitter-consumers/2" {
		t.Fatalf("unexpected registry schema %q", registry.Schema)
	}
	actual := phase16PublicEmitterCalls(t, testsupport.ProjectPath("internal", "compiler"))
	if problems := phase16ConsumerRegistryProblems(actual, registry); len(problems) != 0 {
		t.Fatal(strings.Join(problems, "\n"))
	}
}

// phase16ConsumerRegistryProblems is the shared, fail-closed law for the
// source-derived inventory and its hand-maintained disposition registry.
func phase16ConsumerRegistryProblems(actual []string, registry phase16ConsumerRegistry) []string {
	var problems []string
	seen := make(map[string]phase16ConsumerRegistryRow, len(registry.Entries))
	for _, row := range registry.Entries {
		if row.Call == "" || (row.Classification != phase16AdmittedDynamic && row.Classification != phase16RefusalWithFrozen) {
			problems = append(problems, "invalid classification for registry row "+row.Call)
			continue
		}
		if row.Classification == phase16RefusalWithFrozen && !strings.HasPrefix(row.Witness, "probe:") {
			problems = append(problems, "missing refusal witness for registry row "+row.Call)
		}
		if _, duplicate := seen[row.Call]; duplicate {
			problems = append(problems, "duplicate registry entry "+row.Call)
			continue
		}
		seen[row.Call] = row
	}
	for _, call := range actual {
		if _, ok := seen[call]; !ok {
			problems = append(problems, "missing registry entry for source call "+call)
		}
	}
	for call := range seen {
		index := sort.SearchStrings(actual, call)
		if index == len(actual) || actual[index] != call {
			problems = append(problems, "stale registry entry "+call)
		}
	}
	if len(actual) != len(seen) {
		problems = append(problems, "public emitter inventory cardinality drift: source="+strconv.Itoa(len(actual))+" registry="+strconv.Itoa(len(seen)))
	}
	sort.Strings(problems)
	return problems
}

func TestPhase16EmitterInventoryMutationControls(t *testing.T) {
	// The live scanner has explicit alias, same-package, dot-import, and local
	// shadow branches; these literals prevent their intent from becoming vague.
	for _, required := range []string{"aliases[receiver.Name]", "file.Name.Name == \"cgen\"", "cgen dot import"} {
		data, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "session", "session_phase16_emitter_inventory_test.go"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), required) {
			t.Fatalf("missing mutation control for %q", required)
		}
	}
}
