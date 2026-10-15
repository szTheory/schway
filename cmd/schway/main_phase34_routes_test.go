package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/exitcode"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase34SourceRoutePresentation(t *testing.T) {
	source := "module draft.route\nexport { fn main }\nfn main(n: U64) -> U64 { n + }\n"
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "route.schway")
	if err := os.WriteFile(sourcePath, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	code, jsonOutput, stderr := captureLangRun(t, []string{"--json", "check", sourcePath})
	if code != exitcode.Invalid || stderr != "" {
		t.Fatalf("JSON check code=%d stderr=%q stdout=%q", code, stderr, jsonOutput)
	}
	var checkResult struct {
		Diagnostics []struct {
			ID      string `json:"id"`
			Code    string `json:"code"`
			Primary struct {
				Line int `json:"line"`
				Col  int `json:"col"`
			} `json:"primary_span"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(jsonOutput), &checkResult); err != nil || len(checkResult.Diagnostics) == 0 {
		t.Fatalf("decode check diagnostic: err=%v output=%q", err, jsonOutput)
	}
	want := checkResult.Diagnostics[0]
	if want.ID == "" || want.Code == "" || want.Primary.Line != 3 || want.Primary.Col < 1 {
		t.Fatalf("check diagnostic=%+v, want a stable located diagnostic", want)
	}
	var stdout string

	routes := []struct {
		name string
		args []string
		json bool
	}{
		{name: "interpreter", args: []string{"run", "--engine=interpreter", sourcePath}, json: true},
		{name: "native", args: []string{"run", "--engine=native", sourcePath}, json: true},
		{name: "interpreter-file", args: []string{"run", "--engine=interpreter", sourcePath, "--", filepath.Join(directory, "input.bin")}},
		{name: "native-file", args: []string{"run", "--engine=native", sourcePath, "--manifest", testsupport.ProjectPath("examples", "checksum", "file_bytes.bindings.json"), "--", filepath.Join(directory, "input.bin")}},
		{name: "build", args: []string{"build", sourcePath, "--output", filepath.Join(directory, "app")}, json: true},
	}
	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			code, stdout, stderr := captureLangRun(t, route.args)
			assertPhase34RouteText(t, route.name, code, stdout, stderr, sourcePath, want.ID, want.Code, want.Primary.Col)

			if !route.json {
				return
			}
			jsonArgs := append([]string{"--json"}, route.args...)
			jsonCode, jsonStdout, jsonStderr := captureLangRun(t, jsonArgs)
			if jsonCode != exitcode.Invalid || jsonStderr != "" {
				t.Fatalf("JSON %s code=%d stderr=%q stdout=%q", route.name, jsonCode, jsonStderr, jsonStdout)
			}
			var result struct {
				Diagnostics []struct {
					ID      string `json:"id"`
					Code    string `json:"code"`
					Primary struct {
						Line int `json:"line"`
						Col  int `json:"col"`
					} `json:"primary_span"`
				} `json:"diagnostics"`
			}
			if err := json.Unmarshal([]byte(jsonStdout), &result); err != nil || len(result.Diagnostics) == 0 {
				t.Fatalf("decode JSON %s diagnostic: err=%v output=%q", route.name, err, jsonStdout)
			}
			got := result.Diagnostics[0]
			if got.ID != want.ID || got.Code != want.Code || got.Primary.Line != 3 || got.Primary.Col != want.Primary.Col {
				t.Fatalf("JSON %s diagnostic=%+v, want check identity/location %+v", route.name, got, want)
			}
		})
	}

	// The separator-marked source routes retain the imported module's relative
	// source location through the full admission path.
	importRoot := t.TempDir()
	importedPath := filepath.Join(importRoot, "lib", "bad.schway")
	if err := os.MkdirAll(filepath.Dir(importedPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(importedPath, []byte("module lib.bad\nexport { fn bad }\nfn bad(n: U64) -> U64 {\n  let result = n +\n  result\n}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	importEntryPath := filepath.Join(importRoot, "main.schway")
	importEntry := "module examples.imported_route\nimport lib.bad { fn bad }\nexport { fn main }\nfn main(path: PathToken) -> U64 {\n  let answer = 1\n  answer\n}\n"
	if err := os.WriteFile(importEntryPath, []byte(importEntry), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, route := range []struct {
		name string
		args []string
	}{
		{name: "interpreter", args: []string{"run", "--engine=interpreter", importEntryPath, "--", filepath.Join(importRoot, "input.bin")}},
		{name: "native", args: []string{"run", "--engine=native", importEntryPath, "--manifest", testsupport.ProjectPath("examples", "checksum", "file_bytes.bindings.json"), "--", filepath.Join(importRoot, "input.bin")}},
	} {
		t.Run("imported-file-diagnostic/"+route.name, func(t *testing.T) {
			code, stdout, stderr := captureLangRun(t, route.args)
			if code != exitcode.Invalid || stdout != "" || !strings.Contains(stderr, "lib/bad.schway") || strings.Contains(stderr, importEntryPath+":") {
				t.Fatalf("imported source diagnostic = (%d,%q,%q), want imported relative source ownership", code, stdout, stderr)
			}
		})
	}

	casesPath := filepath.Join(directory, "cases.json")
	if err := os.WriteFile(casesPath, []byte(`{"schema":"lang.replay-cases/1","cases":[{"id":"bad","kind":"source","input":"7","expected":{"kind":"returned","value":"7"},"foreign_outcomes":[]}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	reportPath := filepath.Join(directory, "report.json")
	if err := os.WriteFile(reportPath, []byte("stale report"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr = captureLangRun(t, []string{"app", "verify", sourcePath, "--cases", casesPath, "--report", reportPath})
	assertPhase34RouteText(t, "app verify", code, stdout, stderr, sourcePath, want.ID, want.Code, want.Primary.Col)
	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("source refusal did not replace the stale report: %v", err)
	}
	var report session.ReplayReport
	if err := json.Unmarshal(reportBytes, &report); err != nil || report.Verified || report.Status != session.ReplayStatusFail {
		t.Fatalf("app verify refusal report=%+v err=%v", report, err)
	}

	// A failed report publish must keep the source refusal visible, retain the
	// established invalid-source exit, and include the report failure.
	missingParentReport := filepath.Join(directory, "missing", "report.json")
	code, stdout, stderr = captureLangRun(t, []string{"app", "verify", sourcePath, "--cases", casesPath, "--report", missingParentReport})
	assertPhase34RouteText(t, "app verify report-write failure", code, stdout, stderr, sourcePath, want.ID, want.Code, want.Primary.Col)
	if !strings.Contains(stderr, "verification report") {
		t.Fatalf("report-write failure was lost: stderr=%q", stderr)
	}
}

func assertPhase34RouteText(t *testing.T, route string, code int, stdout, stderr, path, id, diagnosticCode string, column int) {
	t.Helper()
	if code != exitcode.Invalid || stdout != "" {
		t.Fatalf("%s code=%d stdout=%q stderr=%q, want invalid source on stderr", route, code, stdout, stderr)
	}
	for _, part := range []string{fmt.Sprintf("%s:3:%d", path, column), id, diagnosticCode, "n + }", "^"} {
		if !strings.Contains(stderr, part) {
			t.Errorf("%s stderr lacks %q: %q", route, part, stderr)
		}
	}
}

type phase34DiagnosticRoute struct {
	Key          string
	Handler      string
	Calls        []string
	ModuleClass  string
	SourceClass  string
	Owner        string
	SourceReason string
}

var phase34DiagnosticRoutes = []phase34DiagnosticRoute{
	{Key: "app", Handler: "runApplication", Calls: []string{"runApplicationRun", "runApplicationVerify"}, ModuleClass: phase32NoEntrySource, SourceClass: "artifact-runtime", Owner: "TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus", SourceReason: "retained artifact input; no Schway source operand"},
	{Key: "app run", Handler: "runApplicationRun", Calls: []string{"RunApplication", "RunApplicationWithEvidence"}, ModuleClass: phase32NoEntrySource, SourceClass: "artifact-runtime", Owner: "TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus", SourceReason: "retained artifact input; no Schway source operand"},
	{Key: "--json app", Handler: "runApplication", Calls: []string{"runApplicationRun", "runApplicationVerify"}, ModuleClass: phase32NoEntrySource, SourceClass: "artifact-runtime", Owner: "TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus", SourceReason: "global JSON placement shim"},
	{Key: "app verify", Handler: "runApplicationVerify", Calls: []string{"VerifyApplicationCasesFile"}, ModuleClass: phase32ModuleAware, SourceClass: "module-aware-source", Owner: "TestPhase34SourceRoutePresentation", SourceReason: "source refusal uses the checked entry/import snapshots"},
	{Key: "build --manifest --output", Handler: "runBuild", Calls: []string{"BuildApplicationCommandFile"}, ModuleClass: phase32ModuleAware, SourceClass: "module-aware-source", Owner: "TestPhase34SourceRoutePresentation", SourceReason: "same command result handles either manifest placement"},
	{Key: "build --output --manifest", Handler: "runBuild", Calls: []string{"BuildApplicationCommandFile"}, ModuleClass: phase32ModuleAware, SourceClass: "module-aware-source", Owner: "TestPhase34SourceRoutePresentation", SourceReason: "same command result handles either manifest placement"},
	{Key: "build --output", Handler: "runBuild", Calls: []string{"BuildApplicationCommandFile"}, ModuleClass: phase32ModuleAware, SourceClass: "module-aware-source", Owner: "TestPhase34SourceRoutePresentation", SourceReason: "preserves the legacy receipt/diagnostics API"},
	{Key: "explain", Handler: "runExplain", Calls: []string{"ExplainCommandFile"}, ModuleClass: phase32RefusesImports, SourceClass: "single-file-byte-identity", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "summary spans are byte identity and synthetic lookup failures stay unlocated"},
	{Key: "query", Handler: "runQuery", Calls: []string{"QueryCommandFile"}, ModuleClass: phase32RefusesImports, SourceClass: "single-file-byte-identity", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "query fact spans remain byte identity; import refusal is unchanged"},
	{Key: "format", Handler: "runFormat", Calls: []string{"FormatCommandFile"}, ModuleClass: phase32NoEntrySource, SourceClass: "single-file-source", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "format retains the bytes read before canonical projection"},
	{Key: "format --check", Handler: "runFormat", Calls: []string{"FormatCommandFile"}, ModuleClass: phase32NoEntrySource, SourceClass: "single-file-source", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "format.non_canonical is synthetic and remains unlocated"},
	{Key: "check", Handler: "runCheck", Calls: []string{"CheckCommandFile"}, ModuleClass: phase32ModuleAware, SourceClass: "module-aware-source", Owner: "TestPhase34SourceRoutePresentation", SourceReason: "entry and imported diagnostics share the one loaded snapshot"},
	{Key: "evidence", Handler: "runEvidence", Calls: []string{"EvidenceCommandFile"}, ModuleClass: phase32RefusesImports, SourceClass: "single-file-source", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "tool and artifact errors remain unlocated"},
	{Key: "evidence --validate", Handler: "runEvidenceValidation", Calls: []string{"ValidateEvidenceExpanded"}, ModuleClass: phase32RefusesImports, SourceClass: "artifact-diagnostic", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "manifest/source digest mismatch is a synthetic evidence diagnostic"},
	{Key: "verify", Handler: "runVerify", Calls: []string{"VerifyCorpusFile", "VerifyPhase5ControlsAndWork", "VerifyPhase6ControlsAndWork", "VerifyPhase7ControlsAndWork"}, ModuleClass: phase32NoEntrySource, SourceClass: "corpus-diagnostic", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "corpus failures name the failed lane/fixture, not a retained source snapshot"},
	{Key: "stats", Handler: "runStats", ModuleClass: phase32NoEntrySource, SourceClass: "stdin-operational", Owner: "TestStatsCLIComputesGoStatistics", SourceReason: "stdin samples and usage errors have no source operand"},
	{Key: "surface", Handler: "runSurface", Calls: []string{"usageResult", "problemResult"}, ModuleClass: phase32NoEntrySource, SourceClass: "source-free-manifest", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "invalid raw arguments are usage/global diagnostics"},
	{Key: "--json surface", Handler: "runSurface", Calls: []string{"usageResult", "problemResult"}, ModuleClass: phase32NoEntrySource, SourceClass: "source-free-manifest", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "global JSON placement shim; usage errors have no source"},
	{Key: "interface export", Handler: "runInterfaceExport", Calls: []string{"InterfaceExportCommandFile"}, ModuleClass: phase32RefusesImports, SourceClass: "single-file-source", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "peer/global interface failures stay unlocated unless a source span exists"},
	{Key: "interface core", Handler: "runInterfaceCore", Calls: []string{"InterfaceCoreCommandFile"}, ModuleClass: phase32RefusesImports, SourceClass: "single-file-source", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "peer/global interface failures stay unlocated unless a source span exists"},
	{Key: "interface check", Handler: "runInterfaceCheck", Calls: []string{"InterfaceCheckCommandFile"}, ModuleClass: phase32NoEntrySource, SourceClass: "artifact-diagnostic", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "summary/core decode and digest failures carry no source position"},
	{Key: "debug-map", Handler: "runDebugMap", Calls: []string{"DebugMapCommandFile"}, ModuleClass: phase32ModuleAware, SourceClass: "module-aware-source", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "uses loader-confirmed snapshots and relative imported labels"},
	{Key: "run --engine= interpreter", Handler: "runInterpreter", Calls: []string{"RunInterpreterCommandFile"}, ModuleClass: phase32ModuleAware, SourceClass: "module-aware-source", Owner: "TestPhase34SourceRoutePresentation", SourceReason: "entry/imported refusal occurs before interpreting"},
	{Key: "run --engine= native", Handler: "runNative", Calls: []string{"RunNativeCommandFile"}, ModuleClass: phase32ModuleAware, SourceClass: "module-aware-source", Owner: "TestPhase34SourceRoutePresentation", SourceReason: "entry/imported refusal occurs before native execution"},
	{Key: "run --engine=interpreter SOURCE -- PATH", Handler: "runFileApplication", Calls: []string{"parseFileRun", "CheckCommandFile", "RunInterpreterFileApplication", "runNativeFileApplication"}, ModuleClass: phase32ModuleAware, SourceClass: "module-aware-source", Owner: "TestPhase37PublicFileRunTracer", SourceReason: "full module-aware admission keeps entry and imported diagnostic snapshots before file execution"},
	{Key: "run --engine=native SOURCE --manifest MANIFEST -- PATH", Handler: "runFileApplication", Calls: []string{"parseFileRun", "CheckCommandFile", "RunInterpreterFileApplication", "runNativeFileApplication"}, ModuleClass: phase32ModuleAware, SourceClass: "module-aware-source", Owner: "TestPhase37PublicFileRunTracer", SourceReason: "native file application uses the admitted source closure before a private build and one run"},
}

func TestPhase34DiagnosticRouteInventory(t *testing.T) {
	mainPath := testsupport.ProjectPath("cmd", "schway", "main.go")
	mainSource, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	routes := phase32EnumerateRoutes(string(mainSource))
	testNames := phase32ScanTestNames(testsupport.ProjectPath("cmd"), testsupport.ProjectPath("internal"))
	mainFunctions := phase34ParseFunctions(t, mainPath)
	if problems := phase34DiagnosticRouteProblems(routes, phase34DiagnosticRoutes, phase32RouteRows, testNames, mainFunctions); len(problems) != 0 {
		t.Fatalf("diagnostic route inventory problems:\n%s", strings.Join(problems, "\n"))
	}
	entries := phase32EnumerateEntries(testsupport.ProjectPath("examples"))
	if problems := phase32RouteAuditProblems(routes, entries, phase32RouteRows, testNames); len(problems) != 0 {
		t.Fatalf("public route audit problems:\n%s", strings.Join(problems, "\n"))
	}
	if problems := phase32ModuleProblems(phase32RouteRows, check.ModuleAwareRoutes(), nil); len(problems) != 0 {
		t.Fatalf("module boundary changed:\n%s", strings.Join(problems, "\n"))
	}
	if problems := phase34SessionConstructorProblems(t); len(problems) != 0 {
		t.Fatalf("session constructor inventory problems:\n%s", strings.Join(problems, "\n"))
	}

	t.Run("deleted-route-row", func(t *testing.T) {
		rows := phase34WithoutDiagnosticRoute(phase34DiagnosticRoutes, "query")
		if problems := phase34DiagnosticRouteProblems(routes, rows, phase32RouteRows, testNames, mainFunctions); !phase34HasProblem(problems, "missing_route query") {
			t.Fatalf("problems=%v, want missing_route query", problems)
		}
	})
	t.Run("stale-route-row", func(t *testing.T) {
		rows := append(append([]phase34DiagnosticRoute(nil), phase34DiagnosticRoutes...), phase34DiagnosticRoute{Key: "ghost", Handler: "runCheck", ModuleClass: phase32ModuleAware, SourceClass: "module-aware-source", Owner: "TestPhase34DiagnosticRouteInventory", SourceReason: "negative control"})
		if problems := phase34DiagnosticRouteProblems(routes, rows, phase32RouteRows, testNames, mainFunctions); !phase34HasProblem(problems, "stale_route ghost") {
			t.Fatalf("problems=%v, want stale_route ghost", problems)
		}
	})
	t.Run("wrong-module-class", func(t *testing.T) {
		rows := append([]phase34DiagnosticRoute(nil), phase34DiagnosticRoutes...)
		for index := range rows {
			if rows[index].Key == "check" {
				rows[index].ModuleClass = phase32NoEntrySource
			}
		}
		if problems := phase34DiagnosticRouteProblems(routes, rows, phase32RouteRows, testNames, mainFunctions); !phase34HasProblem(problems, "wrong_module_class check") {
			t.Fatalf("problems=%v, want wrong_module_class check", problems)
		}
	})

	missingPath := filepath.Join(t.TempDir(), "missing.schway")
	code, stdout, stderr := captureLangRun(t, []string{"check", missingPath})
	if code != exitcode.Operational || stdout != "" || strings.Contains(stderr, missingPath+":") || !strings.Contains(stderr, "tool.read_failed") {
		t.Fatalf("missing-source operational result code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	code, stdout, stderr = captureLangRun(t, []string{"surface", "--json", "--depth=3"})
	if code != exitcode.Usage || stderr != "" {
		t.Fatalf("surface usage result code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	var usage struct {
		Diagnostics []struct {
			Primary struct {
				Line int `json:"line"`
				Col  int `json:"col"`
			} `json:"primary_span"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(stdout), &usage); err != nil || len(usage.Diagnostics) == 0 || usage.Diagnostics[0].Primary.Line != 0 || usage.Diagnostics[0].Primary.Col != 0 {
		t.Fatalf("surface usage fabricated coordinates: %s err=%v", stdout, err)
	}
}

func phase34DiagnosticRouteProblems(routes []string, rows []phase34DiagnosticRoute, routeRows []phase32RouteRow, testNames map[string]bool, mainFunctions map[string]*ast.FuncDecl) []string {
	observed := map[string]bool{}
	for _, route := range routes {
		observed[route] = true
	}
	phase32 := map[string]phase32RouteRow{}
	for _, row := range routeRows {
		phase32[row.Key] = row
	}
	rowKeys := map[string]bool{}
	problems := []string{}
	for _, row := range rows {
		if rowKeys[row.Key] {
			problems = append(problems, "duplicate_route "+row.Key)
		}
		rowKeys[row.Key] = true
		if !observed[row.Key] {
			problems = append(problems, "stale_route "+row.Key)
		}
		if _, ok := mainFunctions[row.Handler]; !ok {
			problems = append(problems, "missing_handler "+row.Key+" "+row.Handler)
		} else {
			calls := phase34CallNames(mainFunctions[row.Handler])
			for _, name := range row.Calls {
				if !calls[name] {
					problems = append(problems, "missing_dispatch_target "+row.Key+" "+name)
				}
			}
		}
		if !testNames[row.Owner] || row.SourceReason == "" {
			problems = append(problems, "unowned_route "+row.Key)
		}
		if prior, ok := phase32[row.Key]; !ok || prior.Modules != row.ModuleClass {
			problems = append(problems, "wrong_module_class "+row.Key)
		}
	}
	for route := range observed {
		if !rowKeys[route] {
			problems = append(problems, "missing_route "+route)
		}
	}
	sort.Strings(problems)
	return problems
}

func phase34WithoutDiagnosticRoute(rows []phase34DiagnosticRoute, key string) []phase34DiagnosticRoute {
	result := make([]phase34DiagnosticRoute, 0, len(rows)-1)
	for _, row := range rows {
		if row.Key != key {
			result = append(result, row)
		}
	}
	return result
}

func phase34HasProblem(problems []string, want string) bool {
	for _, problem := range problems {
		if problem == want {
			return true
		}
	}
	return false
}

func phase34ParseFunctions(t *testing.T, path string) map[string]*ast.FuncDecl {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	functions := map[string]*ast.FuncDecl{}
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Body != nil {
			functions[function.Name.Name] = function
		}
	}
	return functions
}

func phase34CallNames(function *ast.FuncDecl) map[string]bool {
	calls := map[string]bool{}
	ast.Inspect(function.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch target := call.Fun.(type) {
		case *ast.Ident:
			calls[target.Name] = true
		case *ast.SelectorExpr:
			calls[target.Sel.Name] = true
		}
		return true
	})
	return calls
}

func phase34SessionConstructorProblems(t *testing.T) []string {
	t.Helper()
	paths := []string{
		testsupport.ProjectPath("internal", "compiler", "session", "session.go"),
		testsupport.ProjectPath("internal", "compiler", "session", "session_app_verify.go"),
		testsupport.ProjectPath("internal", "compiler", "session", "session_phase6_evidence.go"),
		testsupport.ProjectPath("internal", "compiler", "session", "session_phase6_explain.go"),
		testsupport.ProjectPath("internal", "compiler", "session", "session_phase6_query.go"),
	}
	actual := map[string]bool{}
	for _, path := range paths {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv != nil {
				continue
			}
			name := function.Name.Name
			if strings.HasSuffix(name, "CommandFile") || name == "BuildApplicationFile" || name == "ValidateEvidenceExpanded" || name == "VerifyApplicationCasesFile" || name == "VerifyCorpusFile" {
				actual[name] = true
			}
		}
	}
	expected := []string{
		"BuildApplicationCommandFile", "BuildApplicationFile", "CheckCommandFile", "DebugMapCommandFile",
		"EvidenceCommandFile", "ExplainCommandFile", "FormatCommandFile", "InterfaceCheckCommandFile",
		"InterfaceCoreCommandFile", "InterfaceExportCommandFile", "QueryCommandFile", "RunInterpreterCommandFile",
		"RunNativeCommandFile", "ValidateEvidenceCommandFile", "ValidateEvidenceExpanded", "VerifyApplicationCasesFile",
		"VerifyCorpusFile",
	}
	wanted := map[string]bool{}
	for _, name := range expected {
		wanted[name] = true
	}
	problems := []string{}
	for name := range wanted {
		if !actual[name] {
			problems = append(problems, "missing_constructor "+name)
		}
	}
	for name := range actual {
		if !wanted[name] {
			problems = append(problems, "unclassified_constructor "+name)
		}
	}
	called := map[string]bool{}
	for _, row := range phase34DiagnosticRoutes {
		for _, name := range row.Calls {
			called[name] = true
		}
	}
	apiOnly := map[string]bool{"BuildApplicationFile": true, "ValidateEvidenceCommandFile": true}
	for name := range wanted {
		if !called[name] && !apiOnly[name] {
			problems = append(problems, "constructor_without_dispatch_or_reason "+name)
		}
	}
	sort.Strings(problems)
	return problems
}
