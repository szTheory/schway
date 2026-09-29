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

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const (
	cgenImportPath           = "github.com/szTheory/schway/internal/compiler/cgen"
	phase16AdmittedDynamic   = "admitted-dynamic-schema2"
	phase16RefusalWithFrozen = "refusal-frozen-witness"
	phase16TypedRefusal      = "typed-refusal"
	phase16EvidenceValidator = "frozen-evidence-validator"
)

type phase16ConsumerRegistry struct {
	Schema  string                       `json:"schema"`
	Entries []phase16ConsumerRegistryRow `json:"entries"`
}

type phase16ConsumerRegistryRow struct {
	Call             string   `json:"call"`
	Classification   string   `json:"classification"`
	Witness          string   `json:"witness,omitempty"`
	EvidenceFixtures []string `json:"evidence_fixtures,omitempty"`
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
		if row.Call == "" || (row.Classification != phase16AdmittedDynamic && row.Classification != phase16RefusalWithFrozen && row.Classification != phase16TypedRefusal && row.Classification != phase16EvidenceValidator) {
			problems = append(problems, "invalid classification for registry row "+row.Call)
			continue
		}
		if (row.Classification == phase16RefusalWithFrozen || row.Classification == phase16TypedRefusal || row.Classification == phase16EvidenceValidator) && !strings.HasPrefix(row.Witness, "probe:") {
			problems = append(problems, "missing refusal witness for registry row "+row.Call)
		}
		if row.Classification == phase16TypedRefusal && len(row.EvidenceFixtures) != 0 {
			problems = append(problems, "typed refusal unexpectedly maps frozen evidence for registry row "+row.Call)
		}
		if row.Classification == phase16EvidenceValidator && row.Witness == "probe:TestPhase16Phase11FrozenEvidenceBindsCanonicalProgram" && len(row.EvidenceFixtures) == 0 {
			problems = append(problems, "frozen evidence validator has no evidence fixtures "+row.Call)
		}
		if row.Classification == phase16RefusalWithFrozen && strings.HasPrefix(row.Call, "internal/compiler/session/session_phase11_gate_test.go:EmitNative:") && len(row.EvidenceFixtures) == 0 {
			problems = append(problems, "Phase 11 frozen refusal row has no evidence fixtures "+row.Call)
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
	actual := phase16PublicEmitterCalls(t, testsupport.ProjectPath("internal", "compiler"))
	for _, mutate := range []struct {
		name      string
		wantClass string
		apply     func(*phase16ConsumerRegistry)
	}{
		{
			name:      "duplicate current row",
			wantClass: "duplicate registry entry ",
			apply: func(registry *phase16ConsumerRegistry) {
				registry.Entries = append(registry.Entries, registry.Entries[0])
			},
		},
		{
			name:      "stale well-formed row",
			wantClass: "stale registry entry ",
			apply: func(registry *phase16ConsumerRegistry) {
				registry.Entries = append(registry.Entries, phase16ConsumerRegistryRow{Call: "internal/compiler/session/not_present.go:EmitNative:1", Classification: phase16AdmittedDynamic})
			},
		},
		{
			name:      "missing current row",
			wantClass: "missing registry entry for source call ",
			apply: func(registry *phase16ConsumerRegistry) {
				registry.Entries = registry.Entries[1:]
			},
		},
		{
			name:      "invalid classification",
			wantClass: "invalid classification for registry row ",
			apply: func(registry *phase16ConsumerRegistry) {
				registry.Entries[0].Classification = "invented-classification"
			},
		},
		{
			name:      "missing refusal witness",
			wantClass: "missing refusal witness for registry row ",
			apply: func(registry *phase16ConsumerRegistry) {
				for i := range registry.Entries {
					if registry.Entries[i].Classification == phase16RefusalWithFrozen {
						registry.Entries[i].Witness = ""
						return
					}
				}
				t.Fatal("fixture has no refusal row to mutate")
			},
		},
	} {
		t.Run(mutate.name, func(t *testing.T) {
			got := phase16InventoryFixture(t)
			mutate.apply(&got)
			problems := phase16ConsumerRegistryProblems(actual, got)
			for _, problem := range problems {
				if strings.HasPrefix(problem, mutate.wantClass) {
					return
				}
			}
			t.Fatalf("mutation did not produce %q: %v", mutate.wantClass, problems)
		})
	}

	// These sentinels protect the scanner's alias, same-package, dot-import,
	// and local-shadow branches while the table above exercises the registry law.
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

func phase16InventoryFixture(t *testing.T) phase16ConsumerRegistry {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase16", "public-emitter-consumers.json"))
	if err != nil {
		t.Fatal(err)
	}
	var registry phase16ConsumerRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		t.Fatal(err)
	}
	return registry
}
