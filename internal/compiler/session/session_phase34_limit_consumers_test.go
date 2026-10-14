package session_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/surfacemanifest"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// This finite consumer census is separate from the manifest and its literal
// witness document. It follows named function references in seven current
// source files; it does not claim whole-program or branch-sensitive analysis.
func TestPhase34SurfaceLimitConsumers(t *testing.T) {
	graph := phase34LimitConsumerGraph(t)
	roots := map[string]string{
		"app verify":          "session.VerifyApplicationCasesFile",
		"build":               "session.BuildApplicationCommandFile",
		"check":               "session.CheckCommandFile",
		"debug-map":           "session.DebugMapCommandFile",
		"evidence":            "session.EvidenceCommandFile",
		"evidence --validate": "session.ValidateEvidenceCommandFile",
		"explain":             "session.ExplainCommandFile",
		"format":              "session.FormatCommandFile",
		"interface core":      "session.InterfaceCoreCommandFile",
		"interface export":    "session.InterfaceExportCommandFile",
		"query":               "session.QueryCommandFile",
		"run":                 "session.RunInterpreterCommandFile",
		"verify":              "session.VerifyCorpusFile",
	}
	expected := map[string][]string{}
	for route, root := range roots {
		if graph[root] == nil {
			t.Fatalf("consumer census lost command entry %s", root)
		}
		if phase34ConsumerReaches(graph, root, "syntax.Parse", map[string]bool{}) {
			expected["parser.functions"] = append(expected["parser.functions"], route)
		}
		if phase34ConsumerReaches(graph, root, "check.Program", map[string]bool{}) {
			expected["scalar.analysis_work"] = append(expected["scalar.analysis_work"], route)
		}
		if phase34ConsumerReaches(graph, root, "session.Check", map[string]bool{}) || phase34ConsumerReaches(graph, root, "session.checkLoaded", map[string]bool{}) {
			expected["check.work"] = append(expected["check.work"], route)
		}
	}
	for _, routes := range expected {
		sort.Strings(routes)
	}
	if len(expected["parser.functions"]) != 13 || len(expected["scalar.analysis_work"]) != 12 || len(expected["check.work"]) != 9 {
		t.Fatalf("consumer census went inert or changed boundary: %v", expected)
	}
	for _, id := range []string{"parser.alternatives", "parser.declarations", "parser.linear_bindings", "parser.primary_diagnostics", "parser.source_bytes", "parser.tokens", "parser.type_depth", "parser.type_nodes"} {
		expected[id] = expected["parser.functions"]
	}
	expected["parser.expression_depth"] = expected["parser.functions"]
	manifest := surfacemanifest.Current()
	if problems := phase34LimitConsumerProblems(manifest, expected); len(problems) != 0 {
		t.Fatalf("manifest differs from current consumers: %v", problems)
	}
	// The conditional debug-map guard is intentionally not inferred by the
	// reference census. Its two paths have executable Phase 30 controls.
	for _, limit := range manifest.Limits {
		if limit.ID == "check.work" && !strings.Contains(limit.Meaning, "debug-map applies this guard only to import-bearing programs") {
			t.Fatal("check.work hides debug-map's conditional guard")
		}
		if limit.ID == "check.work" && !strings.Contains(limit.Meaning, "query applies it only through lane/corpus queries") {
			t.Fatal("check.work hides query's delegated corpus guard")
		}
	}
	for _, control := range []struct{ id, route, reason string }{
		{"parser.functions", "interface core", "missing_consumer_route"},
		{"check.work", "interface export", "missing_consumer_route"},
		{"scalar.analysis_work", "query", "missing_consumer_route"},
		{"check.work", "format", "unsupported_consumer_route"},
	} {
		mutant := surfacemanifest.Current()
		for i := range mutant.Limits {
			if mutant.Limits[i].ID != control.id {
				continue
			}
			var routes []string
			for _, route := range mutant.Limits[i].Routes {
				if route != control.route {
					routes = append(routes, route)
				}
			}
			if control.reason == "unsupported_consumer_route" {
				routes = append(routes, control.route)
			}
			mutant.Limits[i].Routes = routes
		}
		want := control.reason + " " + control.id + " " + control.route
		if !phase34ContainsProblem(phase34LimitConsumerProblems(mutant, expected), want) {
			t.Fatalf("route mutation did not reach %q", want)
		}
	}
	if problems := phase34LimitConsumerProblems(surfacemanifest.Current(), expected); len(problems) != 0 {
		t.Fatalf("route controls did not restore: %v", problems)
	}
}

func phase34LimitConsumerGraph(t *testing.T) map[string][]string {
	t.Helper()
	graph := map[string][]string{}
	for _, relative := range []string{
		"session/session.go", "session/session_modules.go", "session/session_moduledigests.go",
		"session/session_app_verify.go", "session/session_phase6_explain.go", "session/session_phase6_query.go",
		"evidence/evidence.go",
	} {
		file, err := parser.ParseFile(token.NewFileSet(), testsupport.ProjectPath("internal/compiler", relative), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		aliases := map[string]string{}
		for _, item := range file.Imports {
			path, err := strconv.Unquote(item.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Base(path)
			if item.Name != nil {
				name = item.Name.Name
			}
			aliases[name] = filepath.Base(path)
		}
		indexReferences := func(key string, body ast.Node) {
			graph[key] = []string{}
			ast.Inspect(body, func(node ast.Node) bool {
				switch node := node.(type) {
				case *ast.Ident:
					graph[key] = append(graph[key], file.Name.Name+"."+node.Name)
				case *ast.SelectorExpr:
					// Function values such as syntax.Parse passed to evidence's
					// local callback are consumer references as well as calls.
					if base, ok := node.X.(*ast.Ident); ok && aliases[base.Name] != "" {
						graph[key] = append(graph[key], aliases[base.Name]+"."+node.Sel.Name)
					}
				}
				return true
			})
		}
		for _, declaration := range file.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				if declaration.Recv == nil && declaration.Body != nil {
					indexReferences(file.Name.Name+"."+declaration.Name.Name, declaration.Body)
				}
			case *ast.GenDecl:
				// Query's closed dispatch table holds function values; follow
				// those references from the actual command entry as well.
				for _, specification := range declaration.Specs {
					if value, ok := specification.(*ast.ValueSpec); ok {
						for _, name := range value.Names {
							indexReferences(file.Name.Name+"."+name.Name, value)
						}
					}
				}
			}
		}
	}
	return graph
}

func phase34ConsumerReaches(graph map[string][]string, from, target string, seen map[string]bool) bool {
	if from == target {
		return true
	}
	if seen[from] {
		return false
	}
	seen[from] = true
	for _, next := range graph[from] {
		if phase34ConsumerReaches(graph, next, target, seen) {
			return true
		}
	}
	return false
}

func phase34LimitConsumerProblems(manifest surfacemanifest.Manifest, expected map[string][]string) []string {
	var problems []string
	for id, want := range expected {
		var got []string
		for _, limit := range manifest.Limits {
			if limit.ID == id {
				got = limit.Routes
			}
		}
		for _, route := range want {
			if !phase34ContainsProblem(got, route) {
				problems = append(problems, "missing_consumer_route "+id+" "+route)
			}
		}
		for _, route := range got {
			if !phase34ContainsProblem(want, route) {
				problems = append(problems, "unsupported_consumer_route "+id+" "+route)
			}
		}
	}
	sort.Strings(problems)
	return problems
}

func phase34ContainsProblem(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestPhase34InterfaceLimitWitnesses(t *testing.T) {
	// Counts are separately authored: no value or route is read from Current
	// to size these executable probes. Both interface producers must reach the
	// same parser and checked-work refusals before any artifact is published.
	for _, route := range []struct {
		name string
		run  func(string, string) (protocol.Result, error)
	}{
		{"interface export", session.InterfaceExportCommandFile},
		{"interface core", session.InterfaceCoreCommandFile},
	} {
		t.Run(route.name, func(t *testing.T) {
			var at strings.Builder
			at.WriteString("module limit.functions\nexport { fn main }\n")
			for i := 0; i < 1024; i++ {
				name := fmt.Sprintf("f%d", i)
				if i == 0 {
					name = "main"
				}
				fmt.Fprintf(&at, "fn %s(n: U64) -> U64 { n }\n", name)
			}
			atSource := []byte(at.String())
			overFunctions := append(append([]byte{}, atSource...), []byte("fn extra(n: U64) -> U64 { n }\n")...)
			overWork := phase30OverWorkNoImportSource()
			if parsed := syntax.Parse(overWork); len(parsed.Diagnostics) != 0 {
				t.Fatalf("work probe missed checker: %v", parsed.Diagnostics)
			}
			checked := session.Check(overWork)
			workCode, _ := firstPhase34Diagnostic(checked.Diagnostics)
			if checked.Work <= 100000 || workCode != "check.work_limit" {
				t.Fatalf("work probe did not exceed separately pinned 100000: %+v", checked)
			}
			for _, row := range []struct {
				name   string
				source []byte
				code   string
			}{
				{"at-function-cap", atSource, ""},
				{"above-function-cap", overFunctions, "syntax.input_limit"},
				{"above-work-cap", overWork, "check.work_limit"},
			} {
				root := t.TempDir()
				input, output := filepath.Join(root, "input.schway"), filepath.Join(root, "artifact.json")
				if err := os.WriteFile(input, row.source, 0o600); err != nil {
					t.Fatal(err)
				}
				result, err := route.run(input, output)
				if err != nil {
					t.Fatalf("%s: %v", row.name, err)
				}
				if row.code == "" {
					if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
						t.Fatalf("exact cap was refused: %+v", result)
					}
					if _, err := os.Stat(output); err != nil {
						t.Fatalf("exact cap published no artifact: %v", err)
					}
				} else {
					code, _ := firstPhase34Diagnostic(result.Diagnostics)
					if result.Status != protocol.StatusInvalid || code != row.code {
						t.Fatalf("%s: result=%+v, want %s", row.name, result, row.code)
					}
					if _, err := os.Stat(output); !os.IsNotExist(err) {
						t.Fatalf("refused source published artifact: %v", err)
					}
				}
			}
		})
	}
}
