package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase37FileRunRoute struct {
	key         string
	handler     string
	moduleClass string
	sourceClass string
	owner       string
	calls       []string
}

var phase37FileRunRoutes = []phase37FileRunRoute{
	{
		key:         "run --engine=interpreter SOURCE -- PATH",
		handler:     "runFileApplication",
		moduleClass: phase32ModuleAware,
		sourceClass: "module-aware-source",
		owner:       "TestPhase37PublicFileRunTracer",
		calls:       []string{"parseFileRun", "CheckCommandFile", "RunInterpreterFileApplication", "runNativeFileApplication"},
	},
	{
		key:         "run --engine=native SOURCE --manifest MANIFEST -- PATH",
		handler:     "runFileApplication",
		moduleClass: phase32ModuleAware,
		sourceClass: "module-aware-source",
		owner:       "TestPhase37PublicFileRunTracer",
		calls:       []string{"parseFileRun", "CheckCommandFile", "RunInterpreterFileApplication", "runNativeFileApplication"},
	},
}

// phase37EnumerateFileRunRoutes adds the two separator-marked source forms to
// the older finite route census only while the early dispatcher and parser
// remain present. The route grammar tests own the accepted/rejected operands.
func phase37EnumerateFileRunRoutes(mainGoSource string) []string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "main.go", mainGoSource, 0)
	if err != nil {
		return nil
	}
	functions := map[string]*ast.FuncDecl{}
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Body != nil {
			functions[function.Name.Name] = function
		}
	}
	run := functions["run"]
	predicate := functions["hasSeparatorMarkedSourceRun"]
	parserFunction := functions["parseFileRun"]
	handler := functions["runFileApplication"]
	if run == nil || predicate == nil || parserFunction == nil || handler == nil {
		return nil
	}
	runCalls := phase34CallNames(run)
	if !runCalls["hasSeparatorMarkedSourceRun"] || !runCalls["runFileApplication"] {
		return nil
	}
	var dispatchPosition, extractorPosition token.Pos
	ast.Inspect(run.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, selected := call.Fun.(*ast.Ident)
		if !selected {
			return true
		}
		switch selector.Name {
		case "hasSeparatorMarkedSourceRun":
			if dispatchPosition == token.NoPos || call.Pos() < dispatchPosition {
				dispatchPosition = call.Pos()
			}
		case "extractJSON":
			if extractorPosition == token.NoPos || call.Pos() < extractorPosition {
				extractorPosition = call.Pos()
			}
		}
		return true
	})
	if dispatchPosition == token.NoPos || extractorPosition == token.NoPos || dispatchPosition >= extractorPosition {
		return nil
	}
	return []string{phase37FileRunRoutes[0].key, phase37FileRunRoutes[1].key}
}

func phase37FileRunInventoryProblems(routes []string, routeRows []phase32RouteRow, presentationRows []phase34DiagnosticRoute, testNames map[string]bool, mainFunctions map[string]*ast.FuncDecl) []string {
	observed := map[string]bool{}
	for _, route := range routes {
		observed[route] = true
	}
	problems := []string{}
	for _, expected := range phase37FileRunRoutes {
		if !observed[expected.key] {
			problems = append(problems, "missing_route "+expected.key)
		}
		phase32Matches := []phase32RouteRow{}
		for _, row := range routeRows {
			if row.Key == expected.key {
				phase32Matches = append(phase32Matches, row)
			}
		}
		switch len(phase32Matches) {
		case 0:
			problems = append(problems, "missing_public_route "+expected.key)
		case 1:
			if phase32Matches[0].Modules != expected.moduleClass {
				problems = append(problems, "wrong_module_class "+expected.key)
			}
			if !strings.Contains(phase32Matches[0].Reason, expected.owner) || len(phase32Matches[0].Programs) != 0 {
				problems = append(problems, "unowned_public_route "+expected.key)
			}
		default:
			problems = append(problems, "duplicate_public_route "+expected.key)
		}

		presentationMatches := []phase34DiagnosticRoute{}
		for _, row := range presentationRows {
			if row.Key == expected.key {
				presentationMatches = append(presentationMatches, row)
			}
		}
		switch len(presentationMatches) {
		case 0:
			problems = append(problems, "missing_presentation_route "+expected.key)
		case 1:
			row := presentationMatches[0]
			if row.Handler != expected.handler {
				problems = append(problems, "handler_drift "+expected.key)
			}
			if row.SourceClass != expected.sourceClass {
				problems = append(problems, "source_class_drift "+expected.key)
			}
			if row.ModuleClass != expected.moduleClass {
				problems = append(problems, "wrong_presentation_module_class "+expected.key)
			}
			if row.Owner != expected.owner || !testNames[row.Owner] || row.SourceReason == "" {
				problems = append(problems, "unowned_presentation_route "+expected.key)
			}
			function := mainFunctions[row.Handler]
			if function == nil {
				problems = append(problems, "missing_handler "+expected.key)
			} else {
				calls := phase34CallNames(function)
				for _, call := range expected.calls {
					if !calls[call] {
						problems = append(problems, "missing_handler_call "+expected.key+" "+call)
					}
				}
			}
		default:
			problems = append(problems, "duplicate_presentation_route "+expected.key)
		}
	}
	sort.Strings(problems)
	return problems
}

func phase37WithoutPublicRoute(rows []phase32RouteRow, key string) []phase32RouteRow {
	result := []phase32RouteRow{}
	for _, row := range rows {
		if row.Key != key {
			result = append(result, row)
		}
	}
	return result
}

func phase37WithoutPresentationRoute(rows []phase34DiagnosticRoute, key string) []phase34DiagnosticRoute {
	result := []phase34DiagnosticRoute{}
	for _, row := range rows {
		if row.Key != key {
			result = append(result, row)
		}
	}
	return result
}

func phase37ProblemsContain(problems []string, want string) bool {
	for _, problem := range problems {
		if problem == want {
			return true
		}
	}
	return false
}

func TestPhase37FileRunRouteInventoryControls(t *testing.T) {
	mainPath := testsupport.ProjectPath("cmd", "schway", "main.go")
	mainSource, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	routes := append(phase32EnumerateRoutes(string(mainSource)), phase37EnumerateFileRunRoutes(string(mainSource))...)
	// phase32EnumerateRoutes adds the file forms itself once its source census
	// hook is installed; avoid double-counting in the assertion below.
	routes = phase32SortedKeys(func() map[string]bool {
		set := map[string]bool{}
		for _, route := range routes {
			set[route] = true
		}
		return set
	}())
	testNames := phase32ScanTestNames(testsupport.ProjectPath("cmd"), testsupport.ProjectPath("internal"))
	mainFunctions := phase34ParseFunctions(t, mainPath)
	if problems := phase37FileRunInventoryProblems(routes, phase32RouteRows, phase34DiagnosticRoutes, testNames, mainFunctions); len(problems) != 0 {
		t.Fatalf("file-run inventory problems:\n%s", strings.Join(problems, "\n"))
	}

	for _, expected := range phase37FileRunRoutes {
		t.Run(strings.ReplaceAll(expected.key, " ", "-"), func(t *testing.T) {
			if problems := phase37FileRunInventoryProblems(phase37WithoutRoute(routes, expected.key), phase32RouteRows, phase34DiagnosticRoutes, testNames, mainFunctions); !phase37ProblemsContain(problems, "missing_route "+expected.key) {
				t.Fatalf("problems=%v, want missing_route %q", problems, expected.key)
			}
			if problems := phase37FileRunInventoryProblems(routes, phase37WithoutPublicRoute(phase32RouteRows, expected.key), phase34DiagnosticRoutes, testNames, mainFunctions); !phase37ProblemsContain(problems, "missing_public_route "+expected.key) {
				t.Fatalf("problems=%v, want missing_public_route %q", problems, expected.key)
			}
			if problems := phase37FileRunInventoryProblems(routes, phase32RouteRows, phase37WithoutPresentationRoute(phase34DiagnosticRoutes, expected.key), testNames, mainFunctions); !phase37ProblemsContain(problems, "missing_presentation_route "+expected.key) {
				t.Fatalf("problems=%v, want missing_presentation_route %q", problems, expected.key)
			}
			publicRows := append([]phase32RouteRow(nil), phase32RouteRows...)
			for _, row := range phase32RouteRows {
				if row.Key == expected.key {
					publicRows = append(publicRows, row)
					break
				}
			}
			if problems := phase37FileRunInventoryProblems(routes, publicRows, phase34DiagnosticRoutes, testNames, mainFunctions); !phase37ProblemsContain(problems, "duplicate_public_route "+expected.key) {
				t.Fatalf("problems=%v, want duplicate_public_route %q", problems, expected.key)
			}
			presentationRows := append([]phase34DiagnosticRoute(nil), phase34DiagnosticRoutes...)
			for _, row := range phase34DiagnosticRoutes {
				if row.Key == expected.key {
					presentationRows = append(presentationRows, row)
					break
				}
			}
			if problems := phase37FileRunInventoryProblems(routes, phase32RouteRows, presentationRows, testNames, mainFunctions); !phase37ProblemsContain(problems, "duplicate_presentation_route "+expected.key) {
				t.Fatalf("problems=%v, want duplicate_presentation_route %q", problems, expected.key)
			}
			mutated := append([]phase34DiagnosticRoute(nil), phase34DiagnosticRoutes...)
			for index := range mutated {
				if mutated[index].Key == expected.key {
					mutated[index].Handler = "runNative"
				}
			}
			if problems := phase37FileRunInventoryProblems(routes, phase32RouteRows, mutated, testNames, mainFunctions); !phase37ProblemsContain(problems, "handler_drift "+expected.key) {
				t.Fatalf("problems=%v, want handler_drift %q", problems, expected.key)
			}
			mutated = append([]phase34DiagnosticRoute(nil), phase34DiagnosticRoutes...)
			for index := range mutated {
				if mutated[index].Key == expected.key {
					mutated[index].SourceClass = "single-file-source"
				}
			}
			if problems := phase37FileRunInventoryProblems(routes, phase32RouteRows, mutated, testNames, mainFunctions); !phase37ProblemsContain(problems, "source_class_drift "+expected.key) {
				t.Fatalf("problems=%v, want source_class_drift %q", problems, expected.key)
			}
		})
	}
}

func phase37WithoutRoute(routes []string, key string) []string {
	result := []string{}
	for _, route := range routes {
		if route != key {
			result = append(result, route)
		}
	}
	return result
}
