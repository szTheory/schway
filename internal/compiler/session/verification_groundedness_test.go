// verification_groundedness_test.go hosts the groundedness lint (EVD-01,
// D-14-09): a Go test, not a growth of scripts/assert-go-tests.sh. The two
// artifacts have non-overlapping jobs and neither subsumes the other.
// scripts/assert-go-tests.sh is the *selection-time* guard for callers
// actually running `go test` against a live target -- its --self-test
// sentinel proves selection, not document scanning, and that proof does not
// extend to reading markdown. This file is the *document-time* guard over
// .planning/**: it never runs the commands it finds, it only resolves
// whether each one *could* run. Neither artifact is deleted or folded into
// the other; scripts/assert-go-tests.sh's --self-test stays exactly as it
// is today.
//
// Markdown-cell unescape ruling (D-14-17, verified on this tree): a table
// cell's *rendered* form -- after undoing GFM's `\|`/`\\` escaping -- is
// the command a human actually copies and runs from the rendered table.
// `go test ... -run 'PeerLiveness\|LoanChainIndex'` (the raw, still-escaped
// text) prints "[no tests to run]" and exits 0; the unescaped
// `'PeerLiveness|LoanChainIndex'` resolves to
// TestPeerLivenessFileImportsStayIndependent. This lint therefore
// unescapes every cell (`\|`->`|` then `\\`->`\`) before extracting code
// spans, matching what a reader would actually run.
package session_test

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// groundednessModulePath is this module's own import path, read from
// go.mod's module directive rather than hardcoded twice.
const groundednessModulePath = "github.com/codename-lang/lang"

// ---------------------------------------------------------------------
// (a) Static test index (D-14-10) -- go/parser over every *_test.go file,
// never `go test -list`. Resolution elsewhere is by exact identifier only,
// mirroring scripts/assert-go-tests.sh:8-45's exact-match discipline.
// ---------------------------------------------------------------------

// testIndex maps an import path to the set of top-level Test/Fuzz/
// Benchmark/Example identifiers declared in that package's *_test.go
// files.
type testIndex struct {
	byImportPath map[string]map[string]bool
	root         string
}

var testFuncNamePattern = regexp.MustCompile(`^(Test|Fuzz|Benchmark|Example)([A-Z0-9_].*)?$`)

// buildTestIndex walks the module from testsupport.ProjectPath(), parsing
// every *_test.go file with go/parser.ParseFile(..., parser.SkipObjectResolution)
// -- never compiling, never shelling out to `go test -list`. Build
// constraints are honoured via go/build.Context.MatchFile so an
// OS/arch-excluded test file is not counted as live.
func buildTestIndex(t testing.TB) *testIndex {
	t.Helper()
	root := testsupport.ProjectPath()
	index := &testIndex{byImportPath: make(map[string]map[string]bool), root: root}
	fset := token.NewFileSet()
	buildCtx := build.Default

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			base := d.Name()
			if base == ".git" || base == "testdata" || (path != root && strings.HasPrefix(base, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		dir := filepath.Dir(path)
		name := filepath.Base(path)
		match, matchErr := buildCtx.MatchFile(dir, name)
		if matchErr != nil {
			return fmt.Errorf("MatchFile %s: %w", path, matchErr)
		}
		if !match {
			return nil
		}
		file, parseErr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if parseErr != nil {
			return fmt.Errorf("parse %s: %w", path, parseErr)
		}
		importPath := index.importPathFor(dir)
		names := index.byImportPath[importPath]
		if names == nil {
			names = make(map[string]bool)
			index.byImportPath[importPath] = names
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			if isTestLikeFunc(fn) {
				names[fn.Name.Name] = true
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("buildTestIndex: %v", err)
	}
	return index
}

func (index *testIndex) importPathFor(dir string) string {
	rel, err := filepath.Rel(index.root, dir)
	if err != nil {
		return dir
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		return groundednessModulePath
	}
	return groundednessModulePath + "/" + rel
}

// isTestLikeFunc reports whether fn is a top-level declaration matching the
// Test/Fuzz/Benchmark/Example naming convention with a valid signature --
// never a substring or loose match.
func isTestLikeFunc(fn *ast.FuncDecl) bool {
	name := fn.Name.Name
	if !testFuncNamePattern.MatchString(name) {
		return false
	}
	switch {
	case strings.HasPrefix(name, "Test"):
		return paramIsPointerTo(fn.Type.Params, "testing", "T")
	case strings.HasPrefix(name, "Benchmark"):
		return paramIsPointerTo(fn.Type.Params, "testing", "B")
	case strings.HasPrefix(name, "Fuzz"):
		return paramIsPointerTo(fn.Type.Params, "testing", "F")
	case strings.HasPrefix(name, "Example"):
		return fn.Type.Params == nil || len(fn.Type.Params.List) == 0
	}
	return false
}

func paramIsPointerTo(params *ast.FieldList, pkg, typeName string) bool {
	if params == nil || len(params.List) != 1 {
		return false
	}
	star, ok := params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == pkg && sel.Sel.Name == typeName
}

// resolvePackageNames returns the union of top-level test names for one
// package operand. An operand ending in "/..." expands recursively over
// every indexed import path sharing that prefix.
func (index *testIndex) resolvePackageNames(operand string) (map[string]bool, bool) {
	names := make(map[string]bool)
	if strings.HasSuffix(operand, "/...") {
		prefix := index.canonicalImportPath(strings.TrimSuffix(operand, "/..."))
		found := false
		for importPath, set := range index.byImportPath {
			if importPath == prefix || strings.HasPrefix(importPath, prefix+"/") {
				found = true
				for name := range set {
					names[name] = true
				}
			}
		}
		return names, found
	}
	importPath := index.canonicalImportPath(operand)
	set, ok := index.byImportPath[importPath]
	if !ok {
		return names, false
	}
	for name := range set {
		names[name] = true
	}
	return names, true
}

func (index *testIndex) canonicalImportPath(operand string) string {
	operand = strings.TrimSuffix(operand, "/...")
	switch {
	case operand == "." || operand == "":
		return groundednessModulePath
	case strings.HasPrefix(operand, "./"):
		return groundednessModulePath + "/" + strings.TrimPrefix(operand, "./")
	case strings.HasPrefix(operand, groundednessModulePath):
		return operand
	default:
		return groundednessModulePath + "/" + strings.TrimPrefix(operand, "/")
	}
}

// ---------------------------------------------------------------------
// (b) Tier-A document scanner (D-14-15) -- discovery goes through
// phaseArtifactGlob only; this file never globs the filesystem itself.
// ---------------------------------------------------------------------

// verificationCommandPattern matches a code span whose trimmed content is a
// verification command.
var verificationCommandPattern = regexp.MustCompile(`^(go test|go run|go vet|go build|grep|rg|awk|sed|git |\./?scripts/)`)

type commandOccurrence struct {
	File    string
	Line    int
	Command string
}

// tierADocuments discovers every Tier-A evidence document via
// phaseArtifactGlob (D-14-09, D-14-11), the sole path resolver used here.
func tierADocuments(t testing.TB) []string {
	t.Helper()
	validation, err := phaseArtifactGlob("*", "*-VALIDATION.md")
	if err != nil {
		t.Fatalf("phaseArtifactGlob VALIDATION: %v", err)
	}
	verification, err := phaseArtifactGlob("*", "*-VERIFICATION.md")
	if err != nil {
		t.Fatalf("phaseArtifactGlob VERIFICATION: %v", err)
	}
	docs := append([]string{}, validation...)
	docs = append(docs, verification...)
	sort.Strings(docs)
	return docs
}

// extractCommands scans a Tier-A document line by line. A line whose first
// non-space character is `|` is a table row. Cells are split on `|` not
// preceded by `\`, then each cell is unescaped (GFM cell semantics) before
// every backtick-delimited code span inside it is inspected. Extraction is
// per-code-span, not per-cell, because a single row legitimately carries
// two commands.
func extractCommands(path string) ([]commandOccurrence, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var occurrences []commandOccurrence
	for lineIndex, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(strings.TrimLeft(line, " \t"), "|") {
			continue
		}
		for _, cell := range splitTableRow(line) {
			unescaped := unescapeCell(cell)
			for _, span := range extractCodeSpans(unescaped) {
				trimmedSpan := strings.TrimSpace(span)
				if verificationCommandPattern.MatchString(trimmedSpan) {
					occurrences = append(occurrences, commandOccurrence{
						File:    path,
						Line:    lineIndex + 1,
						Command: trimmedSpan,
					})
				}
			}
		}
	}
	return occurrences, nil
}

// splitTableRow splits a markdown table row on `|` characters not preceded
// by a backslash AND not inside a backtick-delimited code span. GFM table
// rendering is code-span-aware: a literal, unescaped `|` inside backticks
// (e.g. `-run 'TestA|TestB'`) is real, measured corpus content (M001's
// 01-VALIDATION.md and others) that does not need `\|` escaping to render
// or run correctly -- only some rows in this corpus escape it (09-VALIDATION.md's
// `Mode.*Invalid\|DecodeMode`), so the splitter must tolerate both
// conventions without truncating the command mid-span.
func splitTableRow(line string) []string {
	var cells []string
	var current strings.Builder
	runes := []rune(line)
	inCode := false
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '`' {
			inCode = !inCode
			current.WriteRune(r)
			continue
		}
		if r == '|' && !inCode && (i == 0 || runes[i-1] != '\\') {
			cells = append(cells, current.String())
			current.Reset()
			continue
		}
		current.WriteRune(r)
	}
	cells = append(cells, current.String())
	return cells
}

// unescapeCell applies GFM table-cell unescaping: `\|` -> `|`, then
// `\\` -> `\`, in that order (D-14-15).
func unescapeCell(cell string) string {
	cell = strings.ReplaceAll(cell, `\|`, `|`)
	cell = strings.ReplaceAll(cell, `\\`, `\`)
	return cell
}

// extractCodeSpans returns every backtick-delimited span's inner text.
func extractCodeSpans(cell string) []string {
	parts := strings.Split(cell, "`")
	var spans []string
	for i := 1; i < len(parts); i += 2 {
		spans = append(spans, parts[i])
	}
	return spans
}

// ---------------------------------------------------------------------
// (c) Classification -- R1 (runnability) and R2 (groundedness) only in
// this task; R3 and R2b arrive in plan 14-06.
// ---------------------------------------------------------------------

type classification string

const (
	classOK          classification = "ok"
	classR1          classification = "R1"
	classR2          classification = "R2"
	classUnparseable classification = "unparseable"
)

type violationRecord struct {
	File           string
	Line           int
	Command        string
	Classification classification
}

var (
	angleBracketPlaceholder = regexp.MustCompile(`<[A-Za-z][^>]*>`)
	unfinishedTokens        = []string{"TODO", "TBD", "XXX"}
	ellipsisRunes           = []string{"…", "⋯", "⋮", "‥", "᠁"}
)

// classifyRunnability applies R1: after deleting every occurrence of the Go
// package wildcard suffix ("/..."), fail on a Unicode ellipsis, an ASCII
// three-dot sequence, an angle-bracket placeholder, or an unfinished-work
// token. Reports true when the command is runnable (not R1).
func classifyRunnability(command string) bool {
	scan := strings.ReplaceAll(command, "/...", "")
	for _, r := range ellipsisRunes {
		if strings.Contains(scan, r) {
			return false
		}
	}
	if strings.Contains(scan, "...") {
		return false
	}
	if angleBracketPlaceholder.MatchString(scan) {
		return false
	}
	for _, token := range unfinishedTokens {
		if strings.Contains(scan, token) {
			return false
		}
	}
	return true
}

// parsedGoTestCommand is the (packages, pattern) shape D-14-15's R2
// extracts from a `go test` command cell. Pattern is empty when the
// command names no -run/-list/-fuzz/-bench flag (a legitimate full-package
// invocation, not a groundedness question).
type parsedGoTestCommand struct {
	Packages []string
	Pattern  string
}

var (
	runFlagPattern    = regexp.MustCompile(`^-(run|list|fuzz|bench)$`)
	envAssignmentExpr = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)
)

// posixTokenize splits a command string into shell-style tokens honouring
// single and double quotes (no expansion) -- a naive field split
// mis-handles quoted -run arguments such as -run 'A|B'.
func posixTokenize(command string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	hasToken := false
	runes := []rune(command)
	i := 0
	for i < len(runes) {
		r := runes[i]
		switch {
		case r == '\'':
			hasToken = true
			i++
			for i < len(runes) && runes[i] != '\'' {
				current.WriteRune(runes[i])
				i++
			}
			if i >= len(runes) {
				return nil, fmt.Errorf("unterminated single quote")
			}
			i++
		case r == '"':
			hasToken = true
			i++
			for i < len(runes) && runes[i] != '"' {
				if runes[i] == '\\' && i+1 < len(runes) && (runes[i+1] == '"' || runes[i+1] == '\\') {
					current.WriteRune(runes[i+1])
					i += 2
					continue
				}
				current.WriteRune(runes[i])
				i++
			}
			if i >= len(runes) {
				return nil, fmt.Errorf("unterminated double quote")
			}
			i++
		case r == ' ' || r == '\t':
			if hasToken {
				tokens = append(tokens, current.String())
				current.Reset()
				hasToken = false
			}
			i++
		default:
			hasToken = true
			current.WriteRune(r)
			i++
		}
	}
	if hasToken {
		tokens = append(tokens, current.String())
	}
	return tokens, nil
}

// parseGoTestCommand tokenizes a `go test` command cell and extracts its
// package operands and -run/-run=/-list/-fuzz/-bench pattern. It returns
// ok=false when the command does not parse into at least (packages) --
// D-14-15's "no silent skips" rule requires the caller to treat that as a
// violation (unparseable), never a pass.
func parseGoTestCommand(command string) (parsedGoTestCommand, bool) {
	tokens, err := posixTokenize(command)
	if err != nil {
		return parsedGoTestCommand{}, false
	}
	i := 0
	for i < len(tokens) && envAssignmentExpr.MatchString(tokens[i]) {
		i++
	}
	if i < len(tokens) && tokens[i] == "env" {
		i++
		for i < len(tokens) && envAssignmentExpr.MatchString(tokens[i]) {
			i++
		}
	}
	if i >= len(tokens) || tokens[i] != "go" {
		return parsedGoTestCommand{}, false
	}
	i++
	if i >= len(tokens) || tokens[i] != "test" {
		return parsedGoTestCommand{}, false
	}
	i++
	var packages []string
	var pattern string
	for i < len(tokens) {
		tok := tokens[i]
		switch {
		case tok == "./..." || strings.HasPrefix(tok, "./") || strings.HasPrefix(tok, groundednessModulePath):
			packages = append(packages, tok)
			i++
		case runFlagPattern.MatchString(tok):
			i++
			if i >= len(tokens) {
				return parsedGoTestCommand{}, false
			}
			pattern = tokens[i]
			i++
		case strings.HasPrefix(tok, "-run=") || strings.HasPrefix(tok, "-list=") || strings.HasPrefix(tok, "-fuzz=") || strings.HasPrefix(tok, "-bench="):
			parts := strings.SplitN(tok, "=", 2)
			if len(parts) != 2 || parts[1] == "" {
				return parsedGoTestCommand{}, false
			}
			pattern = parts[1]
			i++
		default:
			i++
		}
	}
	if len(packages) == 0 {
		return parsedGoTestCommand{}, false
	}
	return parsedGoTestCommand{Packages: packages, Pattern: pattern}, true
}

// classifyGroundedness applies R2: build the union of top-level names
// across the named packages, take the pattern's first slash-separated
// segment, compile it as a Go regexp, and fail if it matches no name in
// the union.
func classifyGroundedness(index *testIndex, packages []string, pattern string) bool {
	names := make(map[string]bool)
	for _, operand := range packages {
		set, _ := index.resolvePackageNames(operand)
		for name := range set {
			names[name] = true
		}
	}
	segment := strings.SplitN(pattern, "/", 2)[0]
	re, err := regexp.Compile(segment)
	if err != nil {
		return false
	}
	for name := range names {
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

// classifyCommand classifies a single verification command occurrence into
// exactly one of {ok, R1, R2, unparseable}. A command containing "go test"
// that does not parse into (packages) is a violation classified
// unparseable, never a pass -- D-14-15's "no silent skips" rule.
func classifyCommand(index *testIndex, command string) classification {
	if !classifyRunnability(command) {
		return classR1
	}
	// HasPrefix, not Contains: the command has already matched
	// verificationCommandPattern's anchored prefix, so this only needs to
	// discriminate which prefix matched. A substring Contains check here
	// is a real bug -- "session.go testdata/..." contains the literal
	// substring "go test" (from ".go test[data]"), which would wrongly
	// route a `git diff` command into the go-test parser.
	if strings.HasPrefix(command, "go test") {
		parsed, ok := parseGoTestCommand(command)
		if !ok {
			return classUnparseable
		}
		if parsed.Pattern != "" && !classifyGroundedness(index, parsed.Packages, parsed.Pattern) {
			return classR2
		}
	}
	return classOK
}

// classifyDocument runs the classifier over one document path, returning
// every violation found. Factored so Task 3's non-inertness proof can run
// it over a temp copy without touching the real .planning/** tree.
func classifyDocument(index *testIndex, path string) ([]violationRecord, error) {
	occurrences, err := extractCommands(path)
	if err != nil {
		return nil, err
	}
	relFile := projectRelativePath(index.root, path)
	var violations []violationRecord
	for _, occ := range occurrences {
		class := classifyCommand(index, occ.Command)
		if class == classOK {
			continue
		}
		violations = append(violations, violationRecord{
			File:           relFile,
			Line:           occ.Line,
			Command:        occ.Command,
			Classification: class,
		})
	}
	return violations, nil
}

// projectRelativePath renders path relative to root (the project root, per
// testsupport.ProjectPath()) so the frontier pin (Task 2) is a portable
// literal rather than an absolute, host-specific path.
func projectRelativePath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// ---------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------

// TestVerificationGroundednessClassifier is the behavior-driven suite
// written before the classifier's real logic is wired up (RED phase): each
// subtest below exercises one bullet of the plan's <behavior> list.
func TestVerificationGroundednessClassifier(t *testing.T) {
	index := &testIndex{byImportPath: map[string]map[string]bool{
		groundednessModulePath + "/internal/compiler/example": {
			"TestExampleThing": true,
		},
	}}

	t.Run("resolves a live pattern as ok", func(t *testing.T) {
		cmd := "go test ./internal/compiler/example/... -run TestExampleThing -v"
		if got := classifyCommand(index, cmd); got != classOK {
			t.Fatalf("classifyCommand(%q) = %s, want ok", cmd, got)
		}
	})

	t.Run("classifies a dead pattern as R2", func(t *testing.T) {
		cmd := "go test ./internal/compiler/example/... -run TestDoesNotExistAnywhere -v"
		if got := classifyCommand(index, cmd); got != classR2 {
			t.Fatalf("classifyCommand(%q) = %s, want R2", cmd, got)
		}
	})

	t.Run("classifies a unicode ellipsis elision as R1", func(t *testing.T) {
		cmd := "go test ./internal/compiler/example/... -run 'TestExample…' -v"
		if got := classifyCommand(index, cmd); got != classR1 {
			t.Fatalf("classifyCommand(%q) = %s, want R1", cmd, got)
		}
	})

	t.Run("classifies an ascii three-dot elision as R1", func(t *testing.T) {
		cmd := "go test ./internal/compiler/example/... -run TestExample..."
		if got := classifyCommand(index, cmd); got != classR1 {
			t.Fatalf("classifyCommand(%q) = %s, want R1", cmd, got)
		}
	})

	t.Run("classifies an angle-bracket placeholder as R1", func(t *testing.T) {
		cmd := "go test ./internal/compiler/<package>/... -run <TestName>"
		if got := classifyCommand(index, cmd); got != classR1 {
			t.Fatalf("classifyCommand(%q) = %s, want R1", cmd, got)
		}
	})

	t.Run("classifies an unfinished-work token as R1", func(t *testing.T) {
		cmd := "go test ./internal/compiler/example/... -run TODO"
		if got := classifyCommand(index, cmd); got != classR1 {
			t.Fatalf("classifyCommand(%q) = %s, want R1", cmd, got)
		}
	})

	t.Run("does not treat the package wildcard suffix as an elision", func(t *testing.T) {
		cmd := "go test ./internal/compiler/example/... -run TestExampleThing"
		if got := classifyCommand(index, cmd); got != classOK {
			t.Fatalf("classifyCommand(%q) = %s, want ok (package wildcard must not register as elision)", cmd, got)
		}
	})

	t.Run("classifies an unparseable go test cell as unparseable, never a pass", func(t *testing.T) {
		cmd := "go test -weird-flag-with-no-package-operand"
		if got := classifyCommand(index, cmd); got != classUnparseable {
			t.Fatalf("classifyCommand(%q) = %s, want unparseable", cmd, got)
		}
	})
}

// TestVerificationGroundedness runs the lint end-to-end over the real
// .planning/** Tier-A corpus and asserts that the live dead patterns named
// in this task's own read_first are found by file and line.
func TestVerificationGroundedness(t *testing.T) {
	docs := tierADocuments(t)
	if len(docs) == 0 {
		t.Fatal("no Tier-A evidence documents found under .planning/**")
	}

	all := measuredViolations(t)
	for _, v := range all {
		t.Logf("%s %s:%d: %s", v.Classification, v.File, v.Line, v.Command)
	}

	mustContain := []struct {
		fileSuffix string
		line       int
	}{
		{"08-VALIDATION.md", 51},
		{"08-VALIDATION.md", 63},
		{"09-VALIDATION.md", 85},
	}
	for _, want := range mustContain {
		found := false
		for _, v := range all {
			if strings.HasSuffix(v.File, want.fileSuffix) && v.Line == want.line {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected a violation at .../%s:%d, found none among %d violations", want.fileSuffix, want.line, len(all))
		}
	}
}

// ---------------------------------------------------------------------
// Task 2: the pinned violation frontier (D-14-18, D-14-19).
// ---------------------------------------------------------------------

// pinnedFrontier is the EXACT violation set measured against this tree
// when this test was authored (D-14-18, re-measured per D-14-19 rather
// than assumed from ROADMAP.md/CONTEXT.md prose). It is NOT a count and
// NOT a ceiling: TestVerificationGroundednessFrontierIsPinned asserts SET
// EQUALITY between this literal and a freshly computed run on every
// invocation, so an addition, a removal, and a count-neutral swap (one
// dead pattern traded for another, leaving the total unchanged) all fail.
// There is no inline suppression syntax anywhere in this file and no
// per-file ignore list -- the only way to shrink this literal is to fix
// the underlying document and edit this slice in a reviewed commit, which
// is exactly what later plans in this phase (and QLT-10) do.
var pinnedFrontier = []violationRecord{
	{File: ".planning/milestones/M001-phases/05-native-equivalence-and-adversarial-evidence/05-VALIDATION.md", Line: 22, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M001-phases/06-agent-feedback-and-performance-ratification/06-VALIDATION.md", Line: 23, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/07-calls-signatures-and-call-graph-refusal/07-VERIFICATION.md", Line: 87, Command: "grep -nE 'TBD|FIXME|XXX'", Classification: classR1},
	{File: ".planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md", Line: 24, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md", Line: 51, Command: "go test ./internal/compiler/check/... -run TestComputeLoanLastUsesAndDerivePlaceLoansAgree", Classification: classR2},
	{File: ".planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md", Line: 63, Command: "go test ./internal/compiler/session/... -run TestAuditQLT02BudgetManifest", Classification: classR2},
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 44, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 85, Command: "go test ./internal/compiler/corevalidate -run 'LoanChainIndex' -v", Classification: classR2},
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 93, Command: "go test ./internal/compiler/corevalidate -run 'Mode.*Invalid|DecodeMode' -v", Classification: classR2},
	{File: ".planning/milestones/M002-phases/10-trusted-interprocedural-oracle/10-VALIDATION.md", Line: 25, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md", Line: 24, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md", Line: 26, Command: "go test ./internal/compiler/<touched-package>/...", Classification: classR1},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md", Line: 51, Command: "grep -c -E 'D-11-(02|07|11|12|13|27|36|40|42)' …/PHASE-11-DEBT.md", Classification: classR1},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md", Line: 56, Command: "go test ./internal/compiler/callgraph/... -run 'TestEntryFunction…' -v -count=1", Classification: classR1},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md", Line: 58, Command: "go test ./internal/compiler/cgen/... -run 'TestEmittedAttributeSet…' -v -count=1", Classification: classR1},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md", Line: 60, Command: "grep -c -E 'function count|call-edge count|N =…' …/11-MIDPHASE-GATE.md", Classification: classR1},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md", Line: 61, Command: "awk … | wc -l", Classification: classR1},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md", Line: 63, Command: "grep -c 'D-11-25' …/session_phase11_differential_test.go", Classification: classR1},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md", Line: 64, Command: "go test ./internal/compiler/session/... -run 'TestQLT03GeneratorOpKindClosure…' -v -count=1", Classification: classR1},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md", Line: 68, Command: "go test ./internal/compiler/cache/... -run 'TestDeclaredInputNames|TestCache…|TestNoClosureDigestInCache' -v -count=1", Classification: classR1},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md", Line: 69, Command: "grep -c -E 'QLT-06a|QLT-06b|strictly dominates…' …/11-QLT06-ABSTENTION.md", Classification: classR1},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md", Line: 71, Command: "go test ./internal/compiler/reduce/... -run 'TestDropCallSite|TestDropOrphanFunction|…' -v -count=1", Classification: classR1},
	{File: ".planning/milestones/M002-phases/12-result-payloads/12-VALIDATION.md", Line: 22, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/12-result-payloads/12-VALIDATION.md", Line: 24, Command: "go test ./internal/compiler/<package>/... -run <TestName> -count=1", Classification: classR1},
	{File: ".planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md", Line: 23, Command: "go test", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md", Line: 24, Command: "go test ./<changed-package>/...", Classification: classR1},
}

// measuredViolations runs the classifier once over the whole Tier-A
// corpus, sorted deterministically by (File, Line).
func measuredViolations(t testing.TB) []violationRecord {
	t.Helper()
	index := buildTestIndex(t)
	docs := tierADocuments(t)
	var measured []violationRecord
	for _, doc := range docs {
		violations, err := classifyDocument(index, doc)
		if err != nil {
			t.Fatalf("classifyDocument(%s): %v", doc, err)
		}
		measured = append(measured, violations...)
	}
	sort.Slice(measured, func(i, j int) bool {
		if measured[i].File != measured[j].File {
			return measured[i].File < measured[j].File
		}
		return measured[i].Line < measured[j].Line
	})
	return measured
}

// TestVerificationGroundednessFrontierIsPinned asserts SET EQUALITY
// between pinnedFrontier and a freshly measured run -- never containment,
// never a count. Both directions are checked and each failure names the
// specific offending record (D-14-18).
func TestVerificationGroundednessFrontierIsPinned(t *testing.T) {
	measured := measuredViolations(t)

	measuredSet := make(map[violationRecord]bool, len(measured))
	for _, v := range measured {
		measuredSet[v] = true
	}
	pinnedSet := make(map[violationRecord]bool, len(pinnedFrontier))
	for _, v := range pinnedFrontier {
		pinnedSet[v] = true
	}

	for _, v := range measured {
		if !pinnedSet[v] {
			t.Errorf("unexpected extra violation not in the pinned frontier: %s %s:%d: %s", v.Classification, v.File, v.Line, v.Command)
		}
	}
	for _, v := range pinnedFrontier {
		if !measuredSet[v] {
			t.Errorf("pinned frontier names a violation the tree no longer produces (remove it from pinnedFrontier once verified): %s %s:%d: %s", v.Classification, v.File, v.Line, v.Command)
		}
	}
}
