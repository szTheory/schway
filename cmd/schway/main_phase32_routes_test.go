package main

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/exitcode"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// Phase 32 public-route audit (D-10b, D-10d, D-09a).
//
// Scope, stated as a non-claim: this audit enumerates the cmd/schway routes
// and the examples/** entry programs only. The `fn main` entries under
// testdata/ are fixtures and sit outside the receipt claim. Global flags
// (--json, --depth, --kind, --cursor, --expand) are parsed by helpers outside
// the audited functions and are not routes.
//
// The routes come from the go/parser AST of run, runApplication and
// runApplicationRun in main.go. A route key is the index-ordered literals the
// dispatch conditions compare against args[N], plus any switch case literals
// under them. Every route and every example entry program needs one row: a
// `covered` row naming receipt programs, or a `not covered because` row whose
// reason names an existing Test function. A route added later fails the audit
// instead of escaping the final receipts.

// phase32ReceiptPrograms is the D-09a receipt program set.
var phase32ReceiptPrograms = []string{
	"checksum", "adler_low", "adler32_cli", "adler32_fixture", "byte_index_oob", "sum_to_n", "fizzbuzz",
}

// phase32RouteRow is one audit row. A row with Programs is covered by the
// final receipts; any other row must carry a Reason naming a Test function.
//
// Modules classifies how the route treats an import-bearing entry (D-13, D-23):
// "module-aware" loads imports, "refuses-imports" reads a single file and
// refuses an import-bearing source naming RefusalCode (default
// module.imports_not_loaded), and "no-entry-source" takes no entry source.
// ReportsRefusalAsPass marks the routes (explain, query) that surface the
// refusal diagnostic as a passing lookup, so they exit 0 rather than 2.
type phase32RouteRow struct {
	Key                  string
	Programs             []string
	Reason               string
	Modules              string
	RefusalCode          string
	ReportsRefusalAsPass bool
}

const (
	phase32ModuleAware    = "module-aware"
	phase32RefusesImports = "refuses-imports"
	phase32NoEntrySource  = "no-entry-source"
)

var phase32AllSeven = phase32ReceiptPrograms

var phase32RouteRows = []phase32RouteRow{
	// Routes covered by the final receipts.
	{Key: "app", Modules: "no-entry-source", Programs: phase32AllSeven},
	{Key: "app run", Modules: "no-entry-source", Programs: phase32AllSeven},
	{Key: "build --manifest --output", Modules: "module-aware", Programs: []string{"checksum", "adler_low", "adler32_cli", "adler32_fixture", "byte_index_oob"}},
	{Key: "build --output --manifest", Modules: "module-aware", Programs: []string{"checksum", "adler_low", "adler32_cli", "adler32_fixture", "byte_index_oob"}},
	{Key: "build --output", Modules: "module-aware", Programs: []string{"sum_to_n", "fizzbuzz"}},

	// Routes the receipts do not exercise; each names the test that owns it.
	{Key: "--json app", Modules: "no-entry-source", Reason: "not covered because it is a routing shim into runApplication; the application path is owned by TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus"},
	{Key: "--json surface", Modules: "no-entry-source", Reason: "not covered because it is the global JSON placement shim owned by TestPhase34SurfaceRoute"},
	{Key: "app verify", Modules: "module-aware", Reason: "not covered because replay verification is owned by TestPhase22AppVerifyIndependentIdentityCases and TestPhase273AppVerifyCapacityCLI"},
	{Key: "explain", Modules: "refuses-imports", ReportsRefusalAsPass: true, Reason: "not covered because it is owned by TestExplainCLIReturnsBoundedCauseDAG"},
	{Key: "query", Modules: "refuses-imports", ReportsRefusalAsPass: true, Reason: "not covered because it is owned by TestQueryCLIResolvesEveryStableIDVocabulary"},
	{Key: "format", Modules: "no-entry-source", Reason: "not covered because it is owned by TestFormatIdempotent and TestConcurrentReadOnlyCommands"},
	{Key: "format --check", Modules: "no-entry-source", Reason: "not covered because it is owned by TestFormatCheck and TestConcurrentReadOnlyCommands"},
	{Key: "check", Modules: "module-aware", Reason: "not covered because it is owned by TestCLIExitTaxonomy and TestPhase30ModuleRefusalExitCodes"},
	{Key: "evidence", Modules: "refuses-imports", Reason: "not covered because it is owned by TestEvidenceCLI"},
	{Key: "evidence --validate", Modules: "refuses-imports", RefusalCode: "evidence.source_invalid", Reason: "not covered because it is owned by TestEvidenceCLI"},
	{Key: "verify", Modules: "no-entry-source", Reason: "not covered because corpus verification is owned by TestVerifyCLI and TestVerifyPhase2CLI"},
	{Key: "stats", Modules: "no-entry-source", Reason: "not covered because the CLI computes statistics from stdin and is owned by TestStatsCLIComputesGoStatistics"},
	{Key: "surface", Modules: "no-entry-source", Reason: "not covered because the source-free manifest route is owned by TestPhase34SurfaceRoute"},
	{Key: "interface export", Modules: "refuses-imports", Reason: "not covered because it is owned by TestInterfaceCheckIsBodyBlindCLI"},
	{Key: "interface core", Modules: "refuses-imports", Reason: "not covered because it is owned by TestInterfaceCheckIsBodyBlindCLI"},
	{Key: "interface check", Modules: "no-entry-source", Reason: "not covered because it is owned by TestInterfaceCheckIsBodyBlindCLI"},
	{Key: "debug-map", Modules: "module-aware", Reason: "not covered because it is owned by TestDebugMapCLIAvailability"},
	{Key: "run --engine= interpreter", Modules: "module-aware", Reason: "not covered because engine agreement is owned by TestPhase29AdlerBothEngines and TestPhase30ModuleRefusalExitCodes"},
	{Key: "run --engine= native", Modules: "module-aware", Reason: "not covered because engine agreement is owned by TestPhase29AdlerBothEngines and TestPhase30ModuleRefusalExitCodes"},
	{Key: "run --engine=interpreter SOURCE -- PATH", Modules: "module-aware", Reason: "not covered because the separator-marked source route is exercised by TestPhase37PublicFileRunTracer"},
	{Key: "run --engine=native SOURCE --manifest MANIFEST -- PATH", Modules: "module-aware", Reason: "not covered because the separator-marked source route is exercised by TestPhase37PublicFileRunTracer"},

	// Entry programs: the seven receipt programs.
	{Key: "examples/checksum/checksum.schway", Programs: []string{"checksum"}},
	{Key: "examples/checksum/adler_low.schway", Programs: []string{"adler_low"}},
	{Key: "examples/checksum/adler32_cli.schway", Programs: []string{"adler32_cli"}},
	{Key: "examples/phase29/adler32_fixture.schway", Programs: []string{"adler32_fixture"}},
	{Key: "examples/phase29/byte_index_oob.schway", Programs: []string{"byte_index_oob"}},
	{Key: "examples/sum_to_n.schway", Programs: []string{"sum_to_n"}},
	{Key: "examples/fizzbuzz.schway", Programs: []string{"fizzbuzz"}},

	// Entry programs the receipts do not run.
	{Key: "examples/bitwise.schway", Reason: "not covered because it was added in Phase 35 after the M006 receipt boundary; both public conformance engines are owned by TestPhase35PublicConformanceWitness"},
	{Key: "examples/phase22/identity.schway", Reason: "not covered because it is owned by TestPhase22IdentityApplicationBuildAndRunCLI"},
	{Key: "examples/phase23/file_byte.schway", Reason: "not covered because it is owned by TestPhase23PublicFileByte"},
	{Key: "examples/phase24/error.schway", Reason: "not covered because its ownership discharge is owned by TestPhase24ActivationPeerDerivesRepeatedAcquisitions"},
	{Key: "examples/phase24/transfer.schway", Reason: "not covered because its ownership discharge is owned by TestPhase24TransferPeerValidatesAcquisitionDerivedDischarge"},
	{Key: "examples/phase26/checked_add_overflow.schway", Reason: "not covered because it is owned by TestPhase26PublicAppCLI"},
	{Key: "examples/phase272/straight_line_copy.schway", Reason: "not covered because it is owned by TestPhase272StraightLineCopyCLI"},
	{Key: "examples/phase28/unused_function.schway", Reason: "not covered because it is owned by TestUnreachableFunctionProgramRuns"},
	{Key: "examples/helper_calls/crc_step.schway", Modules: phase32ModuleAware, Reason: "not covered because its module-aware U64 entry returns tool.run_input_required with exit 64 on both engines until explicit input is supplied; TestPhase36PublicHelperConformance owns that CLI result"},
	{Key: "examples/helper_calls/importer.schway", Modules: phase32ModuleAware, Reason: "not covered because the closed imported crc_step module loads and its U64 entry returns tool.run_input_required with exit 64 on both engines; TestPhase36PublicHelperConformance owns that CLI result"},
	{Key: "examples/helper_calls/public.schway", Modules: phase32ModuleAware, Reason: "not covered because the imported unused-Byte entry returns 1994146192 on both conformance engines; TestPhase36PublicHelperConformance owns that public result"},
	{Key: "examples/crc32/crc32.schway", Modules: phase32ModuleAware, Reason: "not covered because it is outside the M006 receipt set; the imported CRC-32 consumer is owned by TestPhase37CRC32Consumers"},
	{Key: "examples/crc32/crc_low.schway", Modules: phase32ModuleAware, Reason: "not covered because it is outside the M006 receipt set; the imported low-16 CRC consumer is owned by TestPhase37CRC32Consumers"},
}

// phase32UsageOmissions is the recorded set of routes present in dispatch but
// absent from the usage text (D-10d). D-14: the usage text lists every route, so
// the set is empty.
var phase32UsageOmissions = []string{}

// phase32ArgIndex reports N for an `args[N]` expression with a literal index.
func phase32ArgIndex(expr ast.Expr) (int, bool) {
	index, ok := expr.(*ast.IndexExpr)
	if !ok {
		return 0, false
	}
	base, ok := index.X.(*ast.Ident)
	if !ok || base.Name != "args" {
		return 0, false
	}
	lit, ok := index.Index.(*ast.BasicLit)
	if !ok || lit.Kind != token.INT {
		return 0, false
	}
	n, err := strconv.Atoi(lit.Value)
	if err != nil {
		return 0, false
	}
	return n, true
}

func phase32StringLiteral(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return value, true
}

// phase32ConditionLiterals collects the args[N] literals named by a condition.
func phase32ConditionLiterals(expr ast.Expr, into map[int]string) {
	ast.Inspect(expr, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.BinaryExpr:
			if typed.Op != token.EQL && typed.Op != token.NEQ {
				return true
			}
			if n, ok := phase32ArgIndex(typed.X); ok {
				if value, ok := phase32StringLiteral(typed.Y); ok {
					into[n] = value
				}
			}
			if n, ok := phase32ArgIndex(typed.Y); ok {
				if value, ok := phase32StringLiteral(typed.X); ok {
					into[n] = value
				}
			}
		case *ast.CallExpr:
			selector, ok := typed.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "HasPrefix" || len(typed.Args) != 2 {
				return true
			}
			if n, ok := phase32ArgIndex(typed.Args[0]); ok {
				if value, ok := phase32StringLiteral(typed.Args[1]); ok {
					into[n] = value
				}
			}
		}
		return true
	})
}

func phase32JoinLiterals(prefix string, literals map[int]string) string {
	indexes := make([]int, 0, len(literals))
	for n := range literals {
		indexes = append(indexes, n)
	}
	sort.Ints(indexes)
	words := []string{}
	if prefix != "" {
		words = append(words, prefix)
	}
	for _, n := range indexes {
		words = append(words, literals[n])
	}
	return strings.Join(words, " ")
}

// phase32WalkIf records a key for each leaf dispatch condition. A condition
// that only narrows a nested dispatch (a nested if or a switch with literal
// cases under it) is not itself a route.
func phase32WalkIf(statement *ast.IfStmt, prefix string, inherited map[int]string, keys map[string]bool) {
	literals := map[int]string{}
	for n, value := range inherited {
		literals[n] = value
	}
	own := map[int]string{}
	phase32ConditionLiterals(statement.Cond, own)
	if len(own) == 0 {
		return
	}
	for n, value := range own {
		literals[n] = value
	}
	narrowed := false
	for _, inner := range statement.Body.List {
		switch typed := inner.(type) {
		case *ast.IfStmt:
			before := len(keys)
			phase32WalkIf(typed, prefix, literals, keys)
			if len(keys) != before || phase32HasLiteralCond(typed) {
				narrowed = true
			}
		case *ast.SwitchStmt:
			for _, clause := range typed.Body.List {
				caseClause, ok := clause.(*ast.CaseClause)
				if !ok {
					continue
				}
				for _, expr := range caseClause.List {
					if value, ok := phase32StringLiteral(expr); ok {
						keys[phase32JoinLiterals(prefix, literals)+" "+value] = true
						narrowed = true
					}
				}
			}
		}
	}
	if !narrowed {
		keys[phase32JoinLiterals(prefix, literals)] = true
	}
}

func phase32HasLiteralCond(statement *ast.IfStmt) bool {
	own := map[int]string{}
	phase32ConditionLiterals(statement.Cond, own)
	return len(own) > 0
}

// phase32EnumerateRoutes returns the sorted public route keys found in run,
// runApplication and runApplicationRun.
func phase32EnumerateRoutes(mainGoSource string) []string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "main.go", mainGoSource, 0)
	if err != nil {
		return []string{"parse_error " + err.Error()}
	}
	prefixes := map[string]string{"run": "", "runApplication": "app", "runApplicationRun": "app"}
	keys := map[string]bool{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil || function.Body == nil {
			continue
		}
		prefix, audited := prefixes[function.Name.Name]
		if !audited {
			continue
		}
		for _, statement := range function.Body.List {
			if ifStatement, ok := statement.(*ast.IfStmt); ok {
				phase32WalkIf(ifStatement, prefix, map[int]string{}, keys)
			}
		}
	}
	for _, route := range phase37EnumerateFileRunRoutes(mainGoSource) {
		keys[route] = true
	}
	return phase32SortedKeys(keys)
}

func phase32SortedKeys(keys map[string]bool) []string {
	out := make([]string, 0, len(keys))
	for key := range keys {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

// phase32EnumerateEntries returns the example-relative paths of .schway files
// that declare fn main.
func phase32EnumerateEntries(examplesRoot string) []string {
	root := filepath.Dir(examplesRoot)
	entries := []string{}
	_ = filepath.WalkDir(examplesRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".schway" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		for _, function := range syntax.Parse(data).Program.Funcs {
			if function.Name == "main" {
				relative, relErr := filepath.Rel(root, path)
				if relErr == nil {
					entries = append(entries, filepath.ToSlash(relative))
				}
				break
			}
		}
		return nil
	})
	sort.Strings(entries)
	return entries
}

var phase32TestNameToken = regexp.MustCompile(`Test[A-Za-z0-9_]+`)

// phase32ScanTestNames collects top-level Test function names under dirs.
func phase32ScanTestNames(dirs ...string) map[string]bool {
	names := map[string]bool{}
	fset := token.NewFileSet()
	for _, dir := range dirs {
		_ = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, parseErr := parser.ParseFile(fset, path, nil, 0)
			if parseErr != nil {
				return nil
			}
			for _, declaration := range file.Decls {
				if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv == nil && strings.HasPrefix(function.Name.Name, "Test") {
					names[function.Name.Name] = true
				}
			}
			return nil
		})
	}
	return names
}

// phase32RouteAuditProblems compares the enumerated routes and entries with
// the rows. Problems are sorted for stable output.
func phase32RouteAuditProblems(routes, entries []string, rows []phase32RouteRow, testNames map[string]bool) []string {
	observed := map[string]bool{}
	for _, key := range routes {
		observed[key] = true
	}
	for _, key := range entries {
		observed[key] = true
	}
	known := map[string]bool{}
	for _, program := range phase32ReceiptPrograms {
		known[program] = true
	}
	rowKeys := map[string]bool{}
	problems := []string{}
	for _, row := range rows {
		rowKeys[row.Key] = true
		if !observed[row.Key] {
			problems = append(problems, "stale_row "+row.Key)
		}
		if len(row.Programs) > 0 {
			for _, program := range row.Programs {
				if !known[program] {
					problems = append(problems, "unknown_program "+row.Key+" "+program)
				}
			}
			continue
		}
		named := false
		for _, name := range phase32TestNameToken.FindAllString(row.Reason, -1) {
			if testNames[name] {
				named = true
			}
		}
		if !strings.HasPrefix(row.Reason, "not covered because") || !named {
			problems = append(problems, "bad_reason "+row.Key)
		}
	}
	for key := range observed {
		if !rowKeys[key] {
			problems = append(problems, "uncovered_route "+key)
			problems = append(problems, "missing_route "+key)
		}
	}
	sort.Strings(problems)
	return problems
}

// phase32UsageTokens splits usage text into words, additionally exposing the
// parts of `--flag=value|value` tokens.
func phase32UsageTokens(usage string) map[string]bool {
	tokens := map[string]bool{}
	for _, word := range regexp.MustCompile(`[\s\[\]|]+`).Split(usage, -1) {
		if word == "" {
			continue
		}
		tokens[word] = true
		if at := strings.Index(word, "="); at >= 0 {
			tokens[word[:at+1]] = true
			tokens[word[at+1:]] = true
		}
	}
	return tokens
}

// phase32UsageProblems cross-checks the usage text against the routes: every
// subcommand the usage names must be a route, and the routes the usage omits
// must equal the recorded omission set.
func phase32UsageProblems(usage string, routes []string, omissions []string) []string {
	problems := []string{}
	first := map[string]bool{}
	for _, route := range routes {
		first[strings.Fields(route)[0]] = true
	}
	body := strings.TrimPrefix(usage, "usage: schway [--json] ")
	for _, alternative := range strings.Split(body, " | ") {
		fields := strings.Fields(alternative)
		if len(fields) == 0 {
			continue
		}
		if !first[fields[0]] {
			problems = append(problems, "usage_names_missing_route "+fields[0])
		}
	}
	tokens := phase32UsageTokens(usage)
	omitted := []string{}
	for _, route := range routes {
		for _, word := range strings.Fields(route) {
			if !tokens[word] {
				omitted = append(omitted, route)
				break
			}
		}
	}
	sort.Strings(omitted)
	want := append([]string(nil), omissions...)
	sort.Strings(want)
	if strings.Join(omitted, "\n") != strings.Join(want, "\n") {
		problems = append(problems, "usage_omissions_changed got="+strings.Join(omitted, ",")+" want="+strings.Join(want, ","))
	}
	return problems
}

// phase32ModuleCommandWord reduces a module-aware route key to the command
// word ModuleAwareRoutes names: the run and build variants collapse.
func phase32ModuleCommandWord(key string) string {
	switch {
	case strings.HasPrefix(key, "run "):
		return "run"
	case strings.HasPrefix(key, "build "):
		return "build"
	}
	return key
}

func phase32IsEntryKey(key string) bool { return strings.HasPrefix(key, "examples/") }

func phase32RefusalCode(row phase32RouteRow) string {
	if row.RefusalCode != "" {
		return row.RefusalCode
	}
	return "module.imports_not_loaded"
}

// phase32ModuleProblems classifies the routes (D-13, D-23). Rows are walked in
// declared order and the problems are sorted. observe, when non-nil, runs a
// classified route on the import-bearing fixture and returns its exit code and
// combined output.
func phase32ModuleProblems(rows []phase32RouteRow, moduleAware []string, observe func(row phase32RouteRow) (int, string)) []string {
	problems := []string{}
	words := map[string]bool{}
	moduleSet := map[string]bool{}
	for _, word := range moduleAware {
		moduleSet[word] = true
	}
	for _, row := range rows {
		if phase32IsEntryKey(row.Key) {
			continue
		}
		word := phase32ModuleCommandWord(row.Key)
		if (row.Modules == phase32ModuleAware) != moduleSet[word] {
			problems = append(problems, "wrong_module_class "+row.Key)
		}
		switch row.Modules {
		case phase32ModuleAware:
			words[word] = true
		case phase32RefusesImports, phase32NoEntrySource:
		default:
			problems = append(problems, "unclassified_route "+row.Key)
		}
	}
	got := []string{}
	for word := range words {
		got = append(got, word)
	}
	sort.Strings(got)
	want := append([]string(nil), moduleAware...)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		problems = append(problems, "module_aware_set_drift got="+strings.Join(got, ",")+" want="+strings.Join(want, ","))
	}
	if observe != nil {
		for _, row := range rows {
			if phase32IsEntryKey(row.Key) {
				continue
			}
			switch row.Modules {
			case phase32RefusesImports:
				code, output := observe(row)
				wantExit := exitcode.Invalid
				if row.ReportsRefusalAsPass {
					wantExit = exitcode.Success
				}
				if code != wantExit || !strings.Contains(output, phase32RefusalCode(row)) {
					problems = append(problems, "misclassified_route "+row.Key)
				}
			case phase32ModuleAware:
				code, output := observe(row)
				if strings.Contains(output, "module.imports_not_loaded") {
					problems = append(problems, "misclassified_route "+row.Key)
				} else if (row.Key == "check" || row.Key == "debug-map") && code != exitcode.Success {
					problems = append(problems, "misclassified_route "+row.Key)
				}
			}
		}
	}
	sort.Strings(problems)
	return problems
}

// phase32ModuleObserver runs one classified route on the import-bearing fixture
// examples/checksum/adler32_cli.schway. It never runs in parallel: the import
// seam is process-global.
type phase32ModuleObserver struct {
	t            *testing.T
	fixture      string
	bindings     string
	fileBindings string
	fileInput    string
	manifest     string
	refusal      string
	dir          string
}

func newPhase32ModuleObserver(t *testing.T) *phase32ModuleObserver {
	t.Helper()
	fixture := testsupport.ProjectPath("examples", "crc32", "crc32.schway")
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	refusal := ""
	for _, diagnostic := range session.Check(source).Diagnostics {
		if diagnostic.Code == "module.imports_not_loaded" {
			refusal = diagnostic.ID
		}
	}
	if refusal == "" {
		t.Fatal("single-file Check did not fail closed on the fixture's import")
	}
	dir := t.TempDir()
	product, _, err := session.EvidenceCommandFile(context.Background(), testsupport.ProjectPath("examples", "sum_to_n.schway"))
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "sum_to_n.evidence.json")
	if err := os.WriteFile(manifest, product.ManifestBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	fileInput := filepath.Join(dir, "file-input.bin")
	if err := os.WriteFile(fileInput, []byte("123456789"), 0o600); err != nil {
		t.Fatal(err)
	}
	return &phase32ModuleObserver{
		t: t, fixture: fixture, manifest: manifest, refusal: refusal, dir: dir,
		bindings:     testsupport.ProjectPath("examples", "phase29", "fixture_bytes.bindings.json"),
		fileBindings: testsupport.ProjectPath("examples", "checksum", "file_bytes.bindings.json"),
		fileInput:    fileInput,
	}
}

func (o *phase32ModuleObserver) observe(row phase32RouteRow) (int, string) {
	o.t.Helper()
	out := func(name string) string {
		return filepath.Join(o.dir, strings.NewReplacer(" ", "_", "=", "_", "-", "_").Replace(row.Key)+"."+name)
	}
	var args []string
	switch row.Key {
	case "check":
		args = []string{"check", o.fixture}
	case "debug-map":
		args = []string{"debug-map", o.fixture}
	case "run --engine= interpreter":
		args = []string{"run", "--engine=interpreter", o.fixture}
	case "run --engine= native":
		args = []string{"run", "--engine=native", o.fixture}
	case "run --engine=interpreter SOURCE -- PATH":
		args = []string{"run", "--engine=interpreter", o.fixture, "--", o.fileInput}
	case "run --engine=native SOURCE --manifest MANIFEST -- PATH":
		args = []string{"run", "--engine=native", o.fixture, "--manifest", o.fileBindings, "--", o.fileInput}
	case "build --output":
		args = []string{"build", o.fixture, "--output", out("bin")}
	case "build --output --manifest":
		args = []string{"build", o.fixture, "--output", out("bin"), "--manifest", o.bindings}
	case "build --manifest --output":
		args = []string{"build", o.fixture, "--manifest", o.bindings, "--output", out("bin")}
	case "app verify":
		args = []string{"app", "verify", o.fixture, "--cases", testsupport.ProjectPath("examples", "phase22", "identity.cases.json"), "--report", out("report.json")}
	case "evidence":
		args = []string{"evidence", o.fixture}
	case "evidence --validate":
		args = []string{"evidence", "--validate", o.manifest, o.fixture}
	case "interface export":
		args = []string{"interface", "export", o.fixture, out("summary.json")}
	case "interface core":
		args = []string{"interface", "core", o.fixture, out("core.json")}
	case "explain":
		args = []string{"explain", o.fixture, o.refusal}
	case "query":
		// The text form of a query fact omits the diagnostic kind; JSON names it.
		args = []string{"--json", "query", o.fixture, o.refusal}
	default:
		o.t.Fatalf("no observation recipe for route %q", row.Key)
	}
	code, stdout, stderr := captureLangRun(o.t, args)
	return code, stdout + stderr
}

var phase32ProducerProgramsLine = regexp.MustCompile(`(?m)^programs=(.*)$`)

// phase32ProducerPrograms reads the `programs=` line from a milestone manifest.
// It returns nil when the line is absent.
func phase32ProducerPrograms(scriptText string) []string {
	match := phase32ProducerProgramsLine.FindStringSubmatch(scriptText)
	if match == nil {
		return nil
	}
	return strings.Split(match[1], ",")
}

// phase32ProducerProgramProblems compares the covered program set with the
// manifest's line. The producer line is the hub of the coupling.
func phase32ProducerProgramProblems(scriptText string, covered []string) []string {
	want := phase32ProducerPrograms(scriptText)
	if want == nil {
		return []string{"programs_line_missing"}
	}
	got := append([]string(nil), covered...)
	sort.Strings(got)
	sorted := append([]string(nil), want...)
	sort.Strings(sorted)
	if strings.Join(got, ",") != strings.Join(sorted, ",") {
		return []string{"programs_line_drift covered=" + strings.Join(got, ",") + " producer=" + strings.Join(sorted, ",")}
	}
	return nil
}

func phase32WithoutRow(rows []phase32RouteRow, key string) []phase32RouteRow {
	out := []phase32RouteRow{}
	for _, row := range rows {
		if row.Key != key {
			out = append(out, row)
		}
	}
	return out
}

func phase32HasPrefix(problems []string, prefix string) bool {
	for _, problem := range problems {
		if strings.HasPrefix(problem, prefix) {
			return true
		}
	}
	return false
}

func phase32Contains(problems []string, want string) bool {
	for _, problem := range problems {
		if problem == want {
			return true
		}
	}
	return false
}

func TestPhase32PublicRouteAudit(t *testing.T) {
	mainSource, err := os.ReadFile(testsupport.ProjectPath("cmd", "schway", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	routes := phase32EnumerateRoutes(string(mainSource))
	entries := phase32EnumerateEntries(testsupport.ProjectPath("examples"))
	testNames := phase32ScanTestNames(testsupport.ProjectPath("cmd"), testsupport.ProjectPath("internal"))

	if len(routes) < 20 || len(entries) != 20 {
		t.Fatalf("enumeration looks wrong: %d routes %v, %d entries %v (want 20 current entry programs)", len(routes), routes, len(entries), entries)
	}
	if problems := phase32RouteAuditProblems(routes, entries, phase32RouteRows, testNames); len(problems) > 0 {
		t.Fatalf("route audit problems:\n%s", strings.Join(problems, "\n"))
	}

	covered := map[string]bool{}
	for _, row := range phase32RouteRows {
		for _, program := range row.Programs {
			covered[program] = true
		}
	}
	got := []string{}
	for program := range covered {
		got = append(got, program)
	}
	want := append([]string(nil), phase32ReceiptPrograms...)
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("covered program set = %v, want %v (D-09a)", got, want)
	}

	manifest, err := os.ReadFile(testsupport.ProjectPath("testdata", "receipts", "M006.manifest"))
	if err != nil {
		t.Fatal(err)
	}
	manifestText := string(manifest)
	if problems := phase32ProducerProgramProblems(manifestText, got); len(problems) > 0 {
		t.Fatalf("covered set differs from testdata/receipts/M006.manifest programs= (D-10b): %v", problems)
	}
	t.Run("programs-line-drift", func(t *testing.T) {
		original := "programs=" + strings.Join(phase32ReceiptPrograms, ",")
		if !strings.Contains(manifestText, "\n"+original+"\n") {
			t.Fatalf("manifest lacks the line %q", original)
		}
		drifted := strings.Replace(manifestText, original, "programs="+strings.Join(phase32ReceiptPrograms[:len(phase32ReceiptPrograms)-1], ","), 1)
		if problems := phase32ProducerProgramProblems(drifted, got); len(problems) == 0 {
			t.Fatal("a producer line without fizzbuzz did not report a problem")
		}
		if problems := phase32ProducerProgramProblems("no line here\n", got); !phase32Contains(problems, "programs_line_missing") {
			t.Fatalf("problems = %v, want programs_line_missing", problems)
		}
	})
	t.Run("uncovered-route", func(t *testing.T) {
		injected := append(append([]string(nil), routes...), "frobnicate")
		problems := phase32RouteAuditProblems(injected, entries, phase32RouteRows, testNames)
		if !phase32Contains(problems, "uncovered_route frobnicate") {
			t.Fatalf("problems = %v, want uncovered_route frobnicate", problems)
		}
	})
	t.Run("deleted-row", func(t *testing.T) {
		problems := phase32RouteAuditProblems(routes, entries, phase32WithoutRow(phase32RouteRows, "stats"), testNames)
		if !phase32Contains(problems, "uncovered_route stats") {
			t.Fatalf("problems = %v, want uncovered_route stats", problems)
		}
	})
	t.Run("stale-row", func(t *testing.T) {
		rows := append(append([]phase32RouteRow(nil), phase32RouteRows...), phase32RouteRow{Key: "ghost", Reason: "not covered because TestPhase32PublicRouteAudit"})
		problems := phase32RouteAuditProblems(routes, entries, rows, testNames)
		if !phase32Contains(problems, "stale_row ghost") {
			t.Fatalf("problems = %v, want stale_row ghost", problems)
		}
	})
	t.Run("bad-reason", func(t *testing.T) {
		rows := append([]phase32RouteRow(nil), phase32RouteRows...)
		for index := range rows {
			if rows[index].Key == "stats" {
				rows[index].Reason = "not covered because TestDoesNotExist owns it"
			}
		}
		problems := phase32RouteAuditProblems(routes, entries, rows, testNames)
		if !phase32Contains(problems, "bad_reason stats") {
			t.Fatalf("problems = %v, want bad_reason stats", problems)
		}
	})
	t.Run("unknown-program", func(t *testing.T) {
		rows := append(append([]phase32RouteRow(nil), phase32WithoutRow(phase32RouteRows, "stats")...), phase32RouteRow{Key: "stats", Programs: []string{"mystery"}})
		problems := phase32RouteAuditProblems(routes, entries, rows, testNames)
		if !phase32Contains(problems, "unknown_program stats mystery") {
			t.Fatalf("problems = %v, want unknown_program stats mystery", problems)
		}
	})
	t.Run("usage-names-module-aware-set", func(t *testing.T) {
		usage := usageResult().Diagnostics[0].Message
		for _, route := range check.ModuleAwareRoutes() {
			if !strings.Contains(usage, route) {
				t.Errorf("usage text does not name module-aware route %q: %s", route, usage)
			}
		}
	})
	t.Run("module-classification", func(t *testing.T) {
		observer := newPhase32ModuleObserver(t)
		aware := check.ModuleAwareRoutes()
		if problems := phase32ModuleProblems(phase32RouteRows, aware, observer.observe); len(problems) > 0 {
			t.Fatalf("module classification problems:\n%s", strings.Join(problems, "\n"))
		}
		t.Run("unclassified", func(t *testing.T) {
			rows := append([]phase32RouteRow(nil), phase32RouteRows...)
			for index := range rows {
				if rows[index].Key == "stats" {
					rows[index].Modules = ""
				}
			}
			if problems := phase32ModuleProblems(rows, aware, nil); !phase32Contains(problems, "unclassified_route stats") {
				t.Fatalf("problems = %v, want unclassified_route stats", problems)
			}
		})
		t.Run("set-drift", func(t *testing.T) {
			if problems := phase32ModuleProblems(phase32RouteRows, aware[1:], nil); !phase32HasPrefix(problems, "module_aware_set_drift") {
				t.Fatalf("problems = %v, want module_aware_set_drift", problems)
			}
			rows := append([]phase32RouteRow(nil), phase32RouteRows...)
			for index := range rows {
				if rows[index].Key == "debug-map" {
					rows[index].Modules = phase32RefusesImports
				}
			}
			if problems := phase32ModuleProblems(rows, aware, nil); !phase32HasPrefix(problems, "module_aware_set_drift") {
				t.Fatalf("problems = %v, want module_aware_set_drift after reclassifying debug-map", problems)
			}
		})
		t.Run("misclassified", func(t *testing.T) {
			swap := func(key, class string) []phase32RouteRow {
				rows := append([]phase32RouteRow(nil), phase32RouteRows...)
				for index := range rows {
					if rows[index].Key == key {
						rows[index].Modules = class
					}
				}
				return rows
			}
			if problems := phase32ModuleProblems(swap("evidence", phase32ModuleAware), aware, observer.observe); !phase32Contains(problems, "misclassified_route evidence") {
				t.Fatalf("problems = %v, want misclassified_route evidence (a refusing route labelled module-aware)", problems)
			}
			if problems := phase32ModuleProblems(swap("check", phase32RefusesImports), aware, observer.observe); !phase32Contains(problems, "misclassified_route check") {
				t.Fatalf("problems = %v, want misclassified_route check (a module-aware route labelled refusing)", problems)
			}
		})
		t.Run("import-seam", func(t *testing.T) {
			restore := check.SetIgnoreUnloadedImportsForTest(true)
			t.Cleanup(restore)
			problems := phase32ModuleProblems(phase32RouteRows, aware, observer.observe)
			for _, row := range phase32RouteRows {
				if row.Modules == phase32RefusesImports && phase32RefusalCode(row) == "module.imports_not_loaded" {
					if !phase32Contains(problems, "misclassified_route "+row.Key) {
						t.Errorf("with the import guard ignored, %q still refused; problems = %v", row.Key, problems)
					}
				}
			}
			restore()
			if problems := phase32ModuleProblems(phase32RouteRows, aware, observer.observe); len(problems) > 0 {
				t.Fatalf("restoring the import guard did not restore the classification: %v", problems)
			}
		})
		t.Run("adjacency-no-imports", func(t *testing.T) {
			noImports := testsupport.ProjectPath("examples", "sum_to_n.schway")
			for _, args := range [][]string{{"evidence", noImports}, {"check", noImports}} {
				code, stdout, stderr := captureLangRun(t, args)
				if code != exitcode.Success || strings.Contains(stdout+stderr, "module.imports_not_loaded") {
					t.Fatalf("%v on a source without imports: code=%d stdout=%q stderr=%q", args, code, stdout, stderr)
				}
			}
		})
	})
	t.Run("usage-cross-check", func(t *testing.T) {
		usage := usageResult().Diagnostics[0].Message
		if problems := phase32UsageProblems(usage, routes, phase32UsageOmissions); len(problems) > 0 {
			t.Fatalf("usage cross-check problems: %v", problems)
		}
		mutated := usage + " | frobnicate FILE"
		if problems := phase32UsageProblems(mutated, routes, phase32UsageOmissions); !phase32Contains(problems, "usage_names_missing_route frobnicate") {
			t.Fatalf("problems = %v, want usage_names_missing_route frobnicate", problems)
		}
		// Armed controls: removing a route's words from a copy of the usage must
		// be reported against the (empty) recorded omission set.
		withoutStats := strings.Replace(usage, " | stats", "", 1)
		if withoutStats == usage {
			t.Fatal("usage text has no \" | stats\" alternative to remove")
		}
		if problems := phase32UsageProblems(withoutStats, routes, phase32UsageOmissions); !phase32HasPrefix(problems, "usage_omissions_changed got=stats") {
			t.Fatalf("problems = %v, want usage_omissions_changed got=stats", problems)
		}
		withoutManifest := strings.ReplaceAll(usage, "--manifest", "")
		problems := phase32UsageProblems(withoutManifest, routes, phase32UsageOmissions)
		if !phase32HasPrefix(problems, "usage_omissions_changed") ||
			!strings.Contains(strings.Join(problems, "\n"), "build --manifest --output") ||
			!strings.Contains(strings.Join(problems, "\n"), "build --output --manifest") {
			t.Fatalf("problems = %v, want usage_omissions_changed naming both build-manifest routes", problems)
		}
	})
}
