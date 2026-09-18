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
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

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
// (d) Illocutionary-role scoping (D-14-11, plan 14-06 Task 1) -- Tier A
// (asserting, enforced) vs Tier B (proposing, exempt), with a promotion rule
// so renaming a file out of the enforced tier is never an escape.
// ---------------------------------------------------------------------

// documentTier is the outcome of classifyDocumentTier: every discovered
// document is classified into exactly one of these two values, never a
// third "undetermined" bucket (D-14-11).
type documentTier string

const (
	tierEnforced documentTier = "enforced"
	tierExempt   documentTier = "exempt"
)

// enforcedBasenamePattern matches a basename that identifies a document as
// ASSERTING evidence -- validation, verification, summary, acceptance
// testing (UAT), or a mid-phase-gate record. These basenames are always
// enforced, independent of frontmatter or content.
var enforcedBasenamePattern = regexp.MustCompile(`(VALIDATION|VERIFICATION|SUMMARY|UAT|MIDPHASE-GATE)\.md$`)

// proposingBasenamePattern matches a basename that PROPOSES rather than
// asserts. Matching this pattern is necessary but never sufficient for
// exemption -- D-14-11 additionally requires a positive frontmatter
// declaration, so renaming a file into this shape is not itself an escape.
var proposingBasenamePattern = regexp.MustCompile(`-(RESEARCH|PLAN|CONTEXT|DISCUSSION-LOG)\.md$`)

// researchTreeSegment identifies the milestone/project research tree
// (.planning/research/**), exempt by the identical declaration-gated rule
// as the proposing basenames.
const researchTreeSegment = "/research/"

// verificationRoleProposalPattern is the positive frontmatter declaration a
// Tier-B-shaped document must carry to earn exemption
// (`verification_role: proposal`). Read only from inside the leading
// `---`-delimited frontmatter block, mirroring every other frontmatter
// field already read in this corpus (`phase:`, `status:`, ...).
var verificationRoleProposalPattern = regexp.MustCompile(`(?m)^verification_role:\s*proposal\s*$`)

// hasProposalFrontmatter reports whether content's leading frontmatter block
// declares the proposing role. A filename accident is never sufficient.
func hasProposalFrontmatter(content string) bool {
	if !strings.HasPrefix(content, "---\n") {
		return false
	}
	rest := content[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		return false
	}
	return verificationRoleProposalPattern.MatchString(rest[:end])
}

// verdictEmojiPattern matches the freeform Status column's verdict glyphs
// (this file's own legend, mirrored from every *-VALIDATION.md: "⬜
// pending · ✅ green · ❌ red · ⚠️ flaky").
var verdictEmojiPattern = regexp.MustCompile(`[✅❌⚠️⬜]`)

// gradeVocabularyPattern matches the closed grade vocabulary (D-14-01/D-14-05)
// that eventually replaces the Status column, as a whole word so prose that
// merely mentions one of these words is not mistaken for a graded cell.
var gradeVocabularyPattern = regexp.MustCompile(`\b(DEFINED|WIRED|REACHABLE|EXERCISED|MUTATION-KILLED)\b`)

// cellCarriesVerdictToken reports whether an already-unescaped, trimmed
// table cell carries a verdict signal under EITHER the current Status
// column convention or the Grade column convention D-14-05/14-09 migrate
// to -- the keying caveat this plan must not leave briefly undetectable
// from either side.
func cellCarriesVerdictToken(cell string) bool {
	return verdictEmojiPattern.MatchString(cell) || gradeVocabularyPattern.MatchString(cell)
}

// dividerCellPattern matches one GFM table divider cell (`---`, `:--`,
// `--:`, `:-:`).
var dividerCellPattern = regexp.MustCompile(`^:?-{1,}:?$`)

func isTableRowLine(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " \t"), "|")
}

func isDividerRowLine(line string) bool {
	trimmed := strings.Trim(strings.TrimSpace(line), "|")
	if trimmed == "" {
		return false
	}
	for _, cell := range strings.Split(trimmed, "|") {
		if !dividerCellPattern.MatchString(strings.TrimSpace(cell)) {
			return false
		}
	}
	return true
}

// rowCells splits and unescapes one table row line into trimmed cells,
// reusing the backtick-aware splitter and GFM unescape rule D-14-15/D-14-17
// already established -- no second cell-splitting algorithm.
func rowCells(line string) []string {
	raw := splitTableRow(line)
	cells := make([]string, len(raw))
	for i, cell := range raw {
		cells[i] = strings.TrimSpace(unescapeCell(cell))
	}
	return cells
}

// markdownTable is one header+divider+body table discovered in a document.
type markdownTable struct {
	header   []string
	rows     [][]string
	rowLines []int // 1-indexed source line of each row in rows
}

// parseMarkdownTables scans doc content for every header+divider+body table
// -- the structure the verdict-column keying rule (D-14-11's caveat) needs,
// which a flat per-line scan (extractCommands) cannot resolve because it
// has no notion of "this cell is under the Status/Grade column".
func parseMarkdownTables(content string) []markdownTable {
	lines := strings.Split(content, "\n")
	var tables []markdownTable
	i := 0
	for i < len(lines) {
		if isTableRowLine(lines[i]) && i+1 < len(lines) && isTableRowLine(lines[i+1]) && isDividerRowLine(lines[i+1]) {
			table := markdownTable{header: rowCells(lines[i])}
			j := i + 2
			for j < len(lines) && isTableRowLine(lines[j]) {
				table.rows = append(table.rows, rowCells(lines[j]))
				table.rowLines = append(table.rowLines, j+1)
				j++
			}
			tables = append(tables, table)
			i = j
			continue
		}
		i++
	}
	return tables
}

// verdictColumnIndex returns the index of a "Status" or "Grade" header
// column (case-insensitive, exact after trim), or -1 if neither is present
// -- this IS the D-14-11 keying caveat: detection keys on the column name,
// never on scanning the whole row for an emoji (which would falsely promote
// e.g. a RESEARCH.md "File Exists?" ❌ cell that has nothing to do with a
// verification verdict).
func verdictColumnIndex(header []string) int {
	for idx, cell := range header {
		switch strings.ToLower(strings.TrimSpace(cell)) {
		case "status", "grade":
			return idx
		}
	}
	return -1
}

// rowHasCommandSpan reports whether any cell in row contains a
// backtick-delimited span matching verificationCommandPattern.
func rowHasCommandSpan(row []string) bool {
	for _, cell := range row {
		for _, span := range extractCodeSpans(cell) {
			if verificationCommandPattern.MatchString(strings.TrimSpace(span)) {
				return true
			}
		}
	}
	return false
}

// documentHasCommandAndVerdictRow implements D-14-11's promotion clause: a
// table row holding both a command code-span (any cell) and a verdict token
// in the row's own Status/Grade column promotes the WHOLE document to Tier
// A, independent of basename or frontmatter -- this is what makes renaming
// a file useless as an evasion.
func documentHasCommandAndVerdictRow(content string) bool {
	for _, table := range parseMarkdownTables(content) {
		verdictIdx := verdictColumnIndex(table.header)
		if verdictIdx == -1 {
			continue
		}
		for _, row := range table.rows {
			if verdictIdx >= len(row) || !cellCarriesVerdictToken(row[verdictIdx]) {
				continue
			}
			if rowHasCommandSpan(row) {
				return true
			}
		}
	}
	return false
}

// classifyDocumentTier implements D-14-11's scope rule in full. Order
// matters: the promotion check runs FIRST so no basename and no exemption
// declaration can suppress a command+verdict row; the evidence-basename
// check runs next so an asserting document is enforced regardless of
// frontmatter; exemption is checked last and is opt-in ONLY (a positive
// `verification_role: proposal` declaration on a proposing-shaped basename
// or under the research tree) -- everything else defaults enforced, so
// omitting the declaration is never a silent escape either.
func classifyDocumentTier(path string, content string) documentTier {
	if documentHasCommandAndVerdictRow(content) {
		return tierEnforced
	}
	if enforcedBasenamePattern.MatchString(filepath.Base(path)) {
		return tierEnforced
	}
	proposingShaped := proposingBasenamePattern.MatchString(filepath.Base(path)) ||
		strings.Contains(filepath.ToSlash(path), researchTreeSegment)
	if proposingShaped && hasProposalFrontmatter(content) {
		return tierExempt
	}
	return tierEnforced
}

// classifiedDocument pairs a discovered document with its tier.
type classifiedDocument struct {
	Path string
	Tier documentTier
}

// allPlanningMarkdownDocuments walks every *.md file under .planning/**.
// D-14-11's promotion clause ("any OTHER .planning/**/*.md...") is not
// bounded to the VALIDATION/VERIFICATION glob phaseArtifactGlob resolves,
// so the discovery step this plan extends must walk the full tree, not just
// the archived-and-live phase directories.
func allPlanningMarkdownDocuments(t testing.TB) []string {
	t.Helper()
	root := testsupport.ProjectPath(".planning")
	var docs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			docs = append(docs, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("allPlanningMarkdownDocuments: %v", err)
	}
	sort.Strings(docs)
	return docs
}

// classifyAllPlanningDocuments classifies every *.md document under
// .planning/** into exactly one of {enforced, exempt} (D-14-11).
func classifyAllPlanningDocuments(t testing.TB) []classifiedDocument {
	t.Helper()
	var classified []classifiedDocument
	for _, path := range allPlanningMarkdownDocuments(t) {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		classified = append(classified, classifiedDocument{
			Path: path,
			Tier: classifyDocumentTier(path, string(data)),
		})
	}
	return classified
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
// Plan 14-06 Task 1: scope by illocutionary role, with promotion and no
// filename escape (D-14-11).
// ---------------------------------------------------------------------

// TestVerificationGroundednessScopeByIllocutionaryRole exercises every
// bullet of Task 1's <behavior> list against synthetic documents in
// t.TempDir() -- the real .planning/** tree is never written to.
func TestVerificationGroundednessScopeByIllocutionaryRole(t *testing.T) {
	t.Run("evidence basename is enforced regardless of content", func(t *testing.T) {
		content := "# prose only\n\nno tables, no frontmatter, nothing to classify.\n"
		if got := classifyDocumentTier("09-VALIDATION.md", content); got != tierEnforced {
			t.Fatalf("classifyDocumentTier(evidence basename) = %s, want enforced", got)
		}
	})

	t.Run("evidence basename with unresolvable commands and no proposing frontmatter is still enforced -- renaming is not an escape", func(t *testing.T) {
		content := "# doc\n\n| Task | Automated Command | Status |\n" +
			"|---|---|---|\n" +
			"| T1 | `go test ./internal/compiler/<package>/... -run <TestName>` | ⬜ pending |\n"
		if got := classifyDocumentTier("weird-name-VALIDATION.md", content); got != tierEnforced {
			t.Fatalf("classifyDocumentTier(evidence basename, unresolvable content) = %s, want enforced", got)
		}
	})

	t.Run("proposing basename without a positive frontmatter declaration defaults enforced -- exemption is opt-in only", func(t *testing.T) {
		content := "# no frontmatter at all\n\nProposes ten tests never written.\n"
		if got := classifyDocumentTier("04-RESEARCH.md", content); got != tierEnforced {
			t.Fatalf("classifyDocumentTier(proposing basename, no declaration) = %s, want enforced (exemption is opt-in)", got)
		}
	})

	t.Run("proposing basename WITH a positive frontmatter declaration is exempt", func(t *testing.T) {
		content := "---\nphase: \"14\"\nverification_role: proposal\n---\n\n" +
			"Proposes a test that does not exist: `go test ./internal/compiler/example/... -run TestNeverWritten`.\n"
		if got := classifyDocumentTier("14-RESEARCH.md", content); got != tierExempt {
			t.Fatalf("classifyDocumentTier(proposing basename, declared) = %s, want exempt", got)
		}
	})

	t.Run("research tree path WITH declaration is exempt", func(t *testing.T) {
		content := "---\nverification_role: proposal\n---\n\nSpike notes, no tables.\n"
		path := testsupport.ProjectPath(".planning", "research", "M003", "notes.md")
		if got := classifyDocumentTier(path, content); got != tierExempt {
			t.Fatalf("classifyDocumentTier(research tree, declared) = %s, want exempt", got)
		}
	})

	t.Run("a declared-exempt document carrying a verdict token beside a command span is PROMOTED to enforced", func(t *testing.T) {
		content := "---\nverification_role: proposal\n---\n\n" +
			"| Task | Automated Command | Status |\n" +
			"|---|---|---|\n" +
			"| T1 | `go test ./internal/compiler/example/... -run TestExampleThing` | ✅ green |\n"
		if got := classifyDocumentTier("14-RESEARCH.md", content); got != tierEnforced {
			t.Fatalf("classifyDocumentTier(declared exempt, promoted) = %s, want enforced (promotion overrides exemption)", got)
		}
	})

	t.Run("detection succeeds when the Per-Task Verification Map still carries a Status column", func(t *testing.T) {
		content := "## Per-Task Verification Map\n\n" +
			"| Task ID | Automated Command | Status |\n" +
			"|---|---|---|\n" +
			"| T1 | `go test ./internal/compiler/example/... -run TestExampleThing` | ✅ green |\n"
		if !documentHasCommandAndVerdictRow(content) {
			t.Fatal("documentHasCommandAndVerdictRow(Status column) = false, want true")
		}
	})

	t.Run("detection succeeds when the Per-Task Verification Map's Status column has been replaced by a Grade column", func(t *testing.T) {
		content := "## Per-Task Verification Map\n\n" +
			"| Task ID | Automated Command | Grade |\n" +
			"|---|---|---|\n" +
			"| T1 | `go test ./internal/compiler/example/... -run TestExampleThing` | WIRED |\n"
		if !documentHasCommandAndVerdictRow(content) {
			t.Fatal("documentHasCommandAndVerdictRow(Grade column) = false, want true")
		}
	})

	t.Run("a non-verdict column carrying an emoji (e.g. a RESEARCH.md File Exists? cell) is never mistaken for a verdict", func(t *testing.T) {
		content := "| Req ID | Automated Command | File Exists? |\n" +
			"|---|---|---|\n" +
			"| RES-01 | `go test ./internal/compiler/session/... -run TestResourceReleaseOrder` | ❌ Wave 0 — new fixture |\n"
		if documentHasCommandAndVerdictRow(content) {
			t.Fatal("documentHasCommandAndVerdictRow(File Exists? column) = true, want false -- only Status/Grade columns carry a verdict")
		}
	})

	t.Run("every real document under .planning/** classifies into exactly one of {enforced, exempt}", func(t *testing.T) {
		classified := classifyAllPlanningDocuments(t)
		if len(classified) == 0 {
			t.Fatal("no .planning/**/*.md documents discovered")
		}
		enforcedCount, exemptCount := 0, 0
		for _, cd := range classified {
			switch cd.Tier {
			case tierEnforced:
				enforcedCount++
			case tierExempt:
				exemptCount++
			default:
				t.Errorf("%s classified into unknown tier %q", cd.Path, cd.Tier)
			}
			t.Logf("%s: %s", cd.Tier, projectRelativePath(testsupport.ProjectPath(), cd.Path))
		}
		t.Logf("classified %d documents: %d enforced, %d exempt", len(classified), enforcedCount, exemptCount)
	})
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

// ---------------------------------------------------------------------
// Task 3: prove the lint is not inert (mirrors
// TestInjectorMarkerCountGuardIsNotInert's temp-copy-and-seed-one-fault
// shape, session_phase6_injectors_test.go:823-863).
// ---------------------------------------------------------------------

// nonInertBaseDocument is a real Tier-A document that produces ZERO
// violations against the current tree, chosen so a seeded fault's single
// new violation is unambiguously attributable to the seed alone.
const nonInertBaseDocument = ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VERIFICATION.md"

// nonInertSentinelName is the R2 seed's identifier. Its non-existence in
// the module is confirmed against the static index itself before use,
// mirroring scripts/assert-go-tests.sh --self-test's sentinel discipline
// (a nonexistent name, verified nonexistent, not merely asserted).
const nonInertSentinelName = "TestZZZNonexistentSentinelForGroundednessNotInertProof"

// seedCopy copies nonInertBaseDocument into t.TempDir() with one extra
// table row appended and returns the copy's path -- the real .planning/**
// tree is never touched.
func seedCopy(t *testing.T, extraRow string) string {
	t.Helper()
	base := testsupport.ProjectPath(strings.Split(nonInertBaseDocument, "/")...)
	data, err := os.ReadFile(base)
	if err != nil {
		t.Fatalf("read base document: %v", err)
	}
	mutated := string(data)
	if !strings.HasSuffix(mutated, "\n") {
		mutated += "\n"
	}
	mutated += extraRow
	dest := filepath.Join(t.TempDir(), "seeded.md")
	if err := os.WriteFile(dest, []byte(mutated), 0o644); err != nil {
		t.Fatalf("write seeded copy: %v", err)
	}
	return dest
}

// groundednessGitTimeout bounds the one process this file spawns. A
// `git status` over .planning/ is sub-second on any tree this project has
// ever had; the generous ceiling exists so a wedged git cannot hang the
// suite, not as a performance budget.
const groundednessGitTimeout = 30 * time.Second

// groundednessMaxStreamBytes caps each captured child stream independently.
const groundednessMaxStreamBytes = 1 << 20

// groundednessBoundedWriter is this file's local copy of the module's
// bounded-capture shape (testsupport.boundedWriter, native.boundedWriter):
// it retains at most groundednessMaxStreamBytes+1 bytes so overflow is
// observable rather than silently truncated into a plausible-looking
// result. Duplicated rather than exported because both existing copies are
// deliberately package-private and this file is package session_test.
type groundednessBoundedWriter struct {
	buffer bytes.Buffer
	total  int
}

func (w *groundednessBoundedWriter) Write(data []byte) (int, error) {
	w.total += len(data)
	remaining := groundednessMaxStreamBytes + 1 - w.buffer.Len()
	if remaining > len(data) {
		remaining = len(data)
	}
	if remaining > 0 {
		_, _ = w.buffer.Write(data[:remaining])
	}
	return len(data), nil
}

func (w *groundednessBoundedWriter) bytes() []byte { return w.buffer.Bytes() }

func (w *groundednessBoundedWriter) overflowed() bool { return w.total > groundednessMaxStreamBytes }

// TestVerificationGroundednessIsNotInert seeds exactly one fault per
// mechanizable classification (R1 elision, R2 dead pattern, an unparseable
// go test cell) into a temp copy of a clean Tier-A document and asserts
// the classifier reports exactly that violation and nothing else. A guard
// proven red on only one classification is inert for the others, so all
// three are required here.
func TestVerificationGroundednessIsNotInert(t *testing.T) {
	gitStatusPlanning := func(t *testing.T) string {
		t.Helper()
		// Spawned the way every other spawn in this module is
		// (TestSourceNeverSpawnsUnboundedProcesses): a deadline-carrying
		// context, and independently size-capped stdout/stderr writers --
		// never CombinedOutput/Output, which read the child into one
		// unbounded buffer.
		ctx, cancel := context.WithTimeout(context.Background(), groundednessGitTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, "git", "status", "--porcelain", ".planning")
		cmd.Dir = testsupport.ProjectPath()
		var stdout, stderr groundednessBoundedWriter
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("git status --porcelain .planning: %v (stderr: %s)", err, stderr.bytes())
		}
		if stdout.overflowed() || stderr.overflowed() {
			t.Fatalf("git status --porcelain .planning exceeded %d bytes of output", groundednessMaxStreamBytes)
		}
		return string(stdout.bytes())
	}
	// Captured before any seeding so the assertion below is a diff against
	// this run's own activity, not an assumption that the working tree
	// starts clean (STATE.md/state.json are legitimately touched by
	// unrelated GSD tracking machinery during phase execution).
	statusBefore := gitStatusPlanning(t)

	index := buildTestIndex(t)

	for importPath, names := range index.byImportPath {
		if names[nonInertSentinelName] {
			t.Fatalf("sentinel %q unexpectedly exists in %s -- pick a different sentinel", nonInertSentinelName, importPath)
		}
	}

	baseFull := testsupport.ProjectPath(strings.Split(nonInertBaseDocument, "/")...)
	baseline, err := classifyDocument(index, baseFull)
	if err != nil {
		t.Fatalf("classifyDocument(base): %v", err)
	}
	if len(baseline) != 0 {
		t.Fatalf("base document %s already produces %d violation(s); it must be clean for seeded-fault attribution to be unambiguous: %+v", nonInertBaseDocument, len(baseline), baseline)
	}

	assertExactlyOne := func(t *testing.T, row string, want classification) {
		t.Helper()
		copyPath := seedCopy(t, row)
		unmodifiedBaseline, err := classifyDocument(index, baseFull)
		if err != nil {
			t.Fatalf("classifyDocument(base, re-checked): %v", err)
		}
		if len(unmodifiedBaseline) != 0 {
			t.Fatalf("unmodified base document produced %d violation(s), want 0: %+v", len(unmodifiedBaseline), unmodifiedBaseline)
		}
		violations, err := classifyDocument(index, copyPath)
		if err != nil {
			t.Fatalf("classifyDocument(seeded): %v", err)
		}
		if len(violations) != 1 {
			t.Fatalf("expected exactly one seeded violation, got %d: %+v", len(violations), violations)
		}
		if violations[0].Classification != want {
			t.Fatalf("expected classification %s, got %s: %+v", want, violations[0].Classification, violations[0])
		}
	}

	t.Run("R2 dead pattern", func(t *testing.T) {
		row := "| seed | seed | seed | seed | seed | seed | `go test ./internal/compiler/session -run " + nonInertSentinelName + "` | seed | seed |\n"
		assertExactlyOne(t, row, classR2)
	})

	t.Run("R1 elision", func(t *testing.T) {
		row := "| seed | seed | seed | seed | seed | seed | `go test ./internal/compiler/session -run TestSeedElided…` | seed | seed |\n"
		assertExactlyOne(t, row, classR1)
	})

	t.Run("unparseable go test cell", func(t *testing.T) {
		row := "| seed | seed | seed | seed | seed | seed | `go test -flag-with-no-package-or-pattern` | seed | seed |\n"
		assertExactlyOne(t, row, classUnparseable)
	})

	statusAfter := gitStatusPlanning(t)
	if statusAfter != statusBefore {
		t.Fatalf("git status --porcelain .planning changed during this test run -- a seeded copy leaked into the real tree (seeded copies must live only under t.TempDir()):\nbefore:\n%s\nafter:\n%s", statusBefore, statusAfter)
	}
}
