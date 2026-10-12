package main

import (
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
type phase32RouteRow struct {
	Key      string
	Programs []string
	Reason   string
}

var phase32AllSeven = phase32ReceiptPrograms

var phase32RouteRows = []phase32RouteRow{
	// Routes covered by the final receipts.
	{Key: "app", Programs: phase32AllSeven},
	{Key: "app run", Programs: phase32AllSeven},
	{Key: "build --manifest --output", Programs: []string{"checksum", "adler_low", "adler32_cli", "adler32_fixture", "byte_index_oob"}},
	{Key: "build --output --manifest", Programs: []string{"checksum", "adler_low", "adler32_cli", "adler32_fixture", "byte_index_oob"}},
	{Key: "build --output", Programs: []string{"sum_to_n", "fizzbuzz"}},

	// Routes the receipts do not exercise; each names the test that owns it.
	{Key: "--json app", Reason: "not covered because it is a routing shim into runApplication; the application path is owned by TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus"},
	{Key: "app verify", Reason: "not covered because replay verification is owned by TestPhase22AppVerifyIndependentIdentityCases and TestPhase273AppVerifyCapacityCLI"},
	{Key: "explain", Reason: "not covered because it is owned by TestExplainCLIReturnsBoundedCauseDAG"},
	{Key: "query", Reason: "not covered because it is owned by TestQueryCLIResolvesEveryStableIDVocabulary"},
	{Key: "format", Reason: "not covered because it is owned by TestFormatIdempotent and TestConcurrentReadOnlyCommands"},
	{Key: "format --check", Reason: "not covered because it is owned by TestFormatCheck and TestConcurrentReadOnlyCommands"},
	{Key: "check", Reason: "not covered because it is owned by TestCLIExitTaxonomy and TestPhase30ModuleRefusalExitCodes"},
	{Key: "evidence", Reason: "not covered because it is owned by TestEvidenceCLI"},
	{Key: "evidence --validate", Reason: "not covered because it is owned by TestEvidenceCLI"},
	{Key: "verify", Reason: "not covered because corpus verification is owned by TestVerifyCLI and TestVerifyPhase2CLI"},
	{Key: "stats", Reason: "not covered because the CLI computes statistics from stdin and is owned by TestStatsCLIComputesGoStatistics; usage text omits this route; tracked by the 2026-10-12 usage-text todo"},
	{Key: "interface export", Reason: "not covered because it is owned by TestInterfaceCheckIsBodyBlindCLI"},
	{Key: "interface core", Reason: "not covered because it is owned by TestInterfaceCheckIsBodyBlindCLI"},
	{Key: "interface check", Reason: "not covered because it is owned by TestInterfaceCheckIsBodyBlindCLI"},
	{Key: "debug-map", Reason: "not covered because it is owned by TestDebugMapCLIAvailability"},
	{Key: "run --engine= interpreter", Reason: "not covered because engine agreement is owned by TestPhase29AdlerBothEngines and TestPhase30ModuleRefusalExitCodes"},
	{Key: "run --engine= native", Reason: "not covered because engine agreement is owned by TestPhase29AdlerBothEngines and TestPhase30ModuleRefusalExitCodes"},

	// Entry programs: the seven receipt programs.
	{Key: "examples/checksum/checksum.schway", Programs: []string{"checksum"}},
	{Key: "examples/checksum/adler_low.schway", Programs: []string{"adler_low"}},
	{Key: "examples/checksum/adler32_cli.schway", Programs: []string{"adler32_cli"}},
	{Key: "examples/phase29/adler32_fixture.schway", Programs: []string{"adler32_fixture"}},
	{Key: "examples/phase29/byte_index_oob.schway", Programs: []string{"byte_index_oob"}},
	{Key: "examples/sum_to_n.schway", Programs: []string{"sum_to_n"}},
	{Key: "examples/fizzbuzz.schway", Programs: []string{"fizzbuzz"}},

	// Entry programs the receipts do not run.
	{Key: "examples/phase22/identity.schway", Reason: "not covered because it is owned by TestPhase22IdentityApplicationBuildAndRunCLI"},
	{Key: "examples/phase23/file_byte.schway", Reason: "not covered because it is owned by TestPhase23PublicFileByte"},
	{Key: "examples/phase24/error.schway", Reason: "not covered because its ownership discharge is owned by TestPhase24ActivationPeerDerivesRepeatedAcquisitions"},
	{Key: "examples/phase24/transfer.schway", Reason: "not covered because its ownership discharge is owned by TestPhase24TransferPeerValidatesAcquisitionDerivedDischarge"},
	{Key: "examples/phase26/checked_add_overflow.schway", Reason: "not covered because it is owned by TestPhase26PublicAppCLI"},
	{Key: "examples/phase272/straight_line_copy.schway", Reason: "not covered because it is owned by TestPhase272StraightLineCopyCLI"},
	{Key: "examples/phase28/unused_function.schway", Reason: "not covered because it is owned by TestUnreachableFunctionProgramRuns"},
}

// phase32UsageOmissions is the recorded set of routes present in dispatch but
// absent from the usage text (D-10d). The usage text is not edited here.
var phase32UsageOmissions = []string{"build --manifest --output", "build --output --manifest", "stats"}

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
	body := strings.TrimPrefix(usage, "usage: lang [--json] ")
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

func phase32WithoutRow(rows []phase32RouteRow, key string) []phase32RouteRow {
	out := []phase32RouteRow{}
	for _, row := range rows {
		if row.Key != key {
			out = append(out, row)
		}
	}
	return out
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

	if len(routes) < 20 || len(entries) != 14 {
		t.Fatalf("enumeration looks wrong: %d routes %v, %d entries %v", len(routes), routes, len(entries), entries)
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
	t.Run("usage-cross-check", func(t *testing.T) {
		usage := usageResult().Diagnostics[0].Message
		if problems := phase32UsageProblems(usage, routes, phase32UsageOmissions); len(problems) > 0 {
			t.Fatalf("usage cross-check problems: %v", problems)
		}
		mutated := usage + " | frobnicate FILE"
		if problems := phase32UsageProblems(mutated, routes, phase32UsageOmissions); !phase32Contains(problems, "usage_names_missing_route frobnicate") {
			t.Fatalf("problems = %v, want usage_names_missing_route frobnicate", problems)
		}
		if problems := phase32UsageProblems(usage, routes, nil); len(problems) == 0 {
			t.Fatal("dropping the recorded omissions did not fail the cross-check")
		}
	})
}
