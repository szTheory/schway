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
	"errors"
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
// exactly one of {ok, R1, R2, R2b, R3, unparseable}. A command containing
// "go test" that does not parse into (packages) is a violation classified
// unparseable, never a pass -- D-14-15's "no silent skips" rule. R2 (union
// groundedness) is checked before R2b (per-branch groundedness): when every
// branch fails, the union also fails, so that case classifies R2 exactly as
// it always has (plan 14-01's pinned R2 rows are unaffected); R2b only
// fires when the union passes (>=1 branch resolves) but at least one
// OTHER branch does not -- D-14-16's "detected and pinned, not enforced to
// zero this phase" case. Grep-shaped commands are classified via R3 (plan
// 14-06 Task 2), using index.root as the working directory so callers never
// need to thread a separate project-root parameter through.
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
		if parsed.Pattern != "" {
			if !classifyGroundedness(index, parsed.Packages, parsed.Pattern) {
				return classR2
			}
			if failing := classifyPerBranchGroundedness(index, parsed.Packages, parsed.Pattern); len(failing) > 0 {
				return classR2b
			}
		}
		return classOK
	}
	if isGrepShaped(command) {
		if grepClass, _ := classifyGrepGroundedness(index.root, command); grepClass != classOK {
			return grepClass
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

// enforcedTierDocuments returns the project-relative-sortable paths of
// every Tier-A (enforced) document under .planning/**, per D-14-11's
// role-based scope rule -- the plan 14-06 Task 3 production discovery
// entry point, replacing the narrower VALIDATION/VERIFICATION-only
// tierADocuments (D-14-06 basename scope) for measuredViolations and the
// corpus floors. tierADocuments itself is left untouched: plan 14-06 Task 2's
// dedicated grep/per-branch tests keep using it deliberately, since their
// job is demonstrating detection exists, not exercising the full scope.
func enforcedTierDocuments(t testing.TB) []string {
	t.Helper()
	var docs []string
	for _, cd := range classifyAllPlanningDocuments(t) {
		if cd.Tier == tierEnforced {
			docs = append(docs, cd.Path)
		}
	}
	sort.Strings(docs)
	return docs
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
	docs := enforcedTierDocuments(t)
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
		// Line numbers shifted +2 at plan 14-09 (D-14-05's two new
		// frontmatter keys, evidence_vocabulary and graded_rows, inserted
		// ahead of the body of every migrated *-VALIDATION.md) -- a pure
		// coordinate shift, re-measured rather than assumed unchanged.
		{"08-VALIDATION.md", 53},
		{"08-VALIDATION.md", 65},
		{"09-VALIDATION.md", 87},
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
// (plan 14-06 Task 3, re-measured per D-14-19 rather than assumed from
// ROADMAP.md/CONTEXT.md prose). It is NOT a count and NOT a ceiling:
// TestVerificationGroundednessFrontierIsPinned asserts SET EQUALITY between
// this literal and a freshly computed run on every invocation, so an
// addition, a removal, and a count-neutral swap (one dead pattern traded
// for another, leaving the total unchanged) all fail. There is no inline
// suppression syntax anywhere in this file and no per-file ignore list --
// the only way to shrink this literal is to fix the underlying document and
// edit this slice in a reviewed commit, which is exactly what later plans
// in this phase (and QLT-10) do.
//
// This literal grew from plan 14-01's 26 entries to 125 for two combined
// reasons, both by design, neither a regression:
//
//	(1) D-14-11's role-based scope (plan 14-06 Task 1) replaced the
//	    VALIDATION/VERIFICATION-only discovery with the full
//	    .planning/**/*.md tree; no document in this corpus yet declares the
//	    `verification_role: proposal` frontmatter exemption, so every
//	    RESEARCH/PLAN/CONTEXT/DISCUSSION-LOG/research-tree document and
//	    every other non-exempted document defaults enforced (D-14-11's
//	    explicit "exemption is opt-in only" rule) and its command-shaped
//	    table cells are now in scope -- most new entries are the SAME
//	    already-known dead/elided patterns quoted a second (or third) time
//	    in a RESEARCH.md's own test-map table or a SUMMARY.md's own
//	    frontier transcription, not new distinct defects.
//	(2) R3 (grep execution) and R2b (per-branch groundedness) are wired in
//	    for the first time (plan 14-06 Task 2): R3 added 3 new findings
//	    (archival breakage -- a cited grep no longer matches its target
//	    file), and R2b added 11 new findings (an alternation `-run` pattern
//	    whose union resolves but whose SPECIFIC named branch does not),
//	    consistent with D-14-16's prediction that the per-branch defect
//	    mass would be the larger of the two.
//
// Shrinking this literal toward empty is Tier-B frontmatter declarations on
// genuine proposals (closing reason (1)) plus real reconciliation of the
// R1/R2/R2b/R3 findings (closing reason (2)) -- both explicitly deferred to
// later plans in this phase and to QLT-10 (D-14-16, D-14-18).
//
// Re-measured at plan 14-09 (125 -> 124 entries), attributed rather than
// assumed unchanged, per this literal's own discipline:
//
//	(1) EVERY (file, line) in the fourteen migrated *-VALIDATION.md
//	    documents shifted by exactly +2 (D-14-05's two new frontmatter
//	    keys, evidence_vocabulary and graded_rows, inserted ahead of the
//	    body): every pinned record naming one of those fourteen files
//	    moved down two lines, a pure coordinate shift, not a new or
//	    removed finding.
//	(2) One entry was REMOVED for real, not shifted: 14-RESEARCH.md's own
//	    citation of `go test ./internal/compiler/session/... -run
//	    TestValidationRowGradesAreEarned -v` is no longer a dead pattern
//	    -- plan 14-09 Task 1 implemented that exact test, so the pattern
//	    now resolves. This is the frontier moving for a genuine reason,
//	    the same shape every other frontier fixture in this milestone is
//	    required to demonstrate.
//
// enforced-tier document count is unchanged across the migration (verified
// in this plan's SUMMARY): retiring the Status column did not make any
// document undetectable, confirming D-14-11's Status/Grade dual-keying
// (plan 14-06) did exactly the job it was built for.
//
// Re-measured at plan 14-10 (124 -> 58 entries): every runnability (R1,
// 34 entries), groundedness (R2, 29 entries) and grep (R3, 3 entries)
// finding is now reconciled outside the archive under a debt row with a
// witness in PHASE-14-DEBT.md (D-14-55..D-14-120), so
// TestVerificationGroundednessFrontierIsPinned excludes every reconciled
// finding from its comparison (reconciledFindings, below) rather than
// this literal continuing to pin all three classes. Per-branch (R2b, 10
// entries) is untouched -- D-14-16's sizing decision keeps it pinned,
// not enforced to zero, this phase; every R2b entry now also carries a
// closed-vocabulary landing phase (r2bLandingPhases, below), all P20 per
// ROADMAP.md's QLT-10 assignment. This is the frontier moving for a
// genuine reason, the same shape every other frontier fixture in this
// milestone is required to demonstrate -- verified by
// TestVerificationGroundednessThreeClassesAreEmpty.
var pinnedFrontier = []violationRecord{
	{File: ".planning/STANDING-VERDICTS.md", Line: 21, Command: "go test -fuzz", Classification: classUnparseable},
	{File: ".planning/milestones/M001-phases/02-owned-values-and-abilities/02-RESEARCH.md", Line: 85, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md", Line: 552, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-VERIFICATION.md", Line: 86, Command: "go test ./internal/compiler/corevalidate -run '^(TestForeignPolicyValueNotIdentifierRefused|TestForeignContractCommentSafetyRefused)$' -v", Classification: classR2b},
	{File: ".planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-VERIFICATION.md", Line: 88, Command: "go test ./internal/compiler/corevalidate -run '^(TestInteriorMergeDivergentHistoriesRefused|TestInteriorMergeAgreeingHistoriesAccepted|TestCyclicOkEdgeChainRefusedNotHung|TestAcyclicChainsStillValidateUnderCycleGuard)$' -v", Classification: classR2b},
	{File: ".planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-VERIFICATION.md", Line: 89, Command: "go test ./internal/compiler/cgen -run '^(TestForeignSymbolInjectionNeverReachesGeneratedC|TestForeignEmittersRefuseNonIdentifierSymbolIndependently|TestExistingEmittersAreByteIdentical)$' -v", Classification: classR2b},
	{File: ".planning/milestones/M001-phases/05-native-equivalence-and-adversarial-evidence/05-RESEARCH.md", Line: 418, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M001-phases/05-native-equivalence-and-adversarial-evidence/05-VALIDATION.md", Line: 24, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M001-phases/05-native-equivalence-and-adversarial-evidence/05-VERIFICATION.md", Line: 56, Command: "go test ./internal/compiler/session/... -run \"TestNAT03|TestAssertMutationMovesAnAxis|TestEveryNAT03\" -v", Classification: classR2b},
	{File: ".planning/milestones/M001-phases/06-agent-feedback-and-performance-ratification/06-VALIDATION.md", Line: 25, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/07-calls-signatures-and-call-graph-refusal/07-02-PLAN.md", Line: 465, Command: "go test -race", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/07-calls-signatures-and-call-graph-refusal/07-04-PLAN.md", Line: 321, Command: "go test -race", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/07-calls-signatures-and-call-graph-refusal/07-05-PLAN.md", Line: 397, Command: "go test -race", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/07-calls-signatures-and-call-graph-refusal/07-06-PLAN.md", Line: 512, Command: "go test -race", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-RESEARCH.md", Line: 495, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md", Line: 26, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-RESEARCH.md", Line: 697, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 46, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 86, Command: "go test ./internal/compiler/corevalidate -run 'PeerLiveness|LoanChainIndex' -v", Classification: classR2b},
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 89, Command: "go test ./internal/compiler/corevalidate -run 'ImportsStayIndependent|ImportIndependence' -v", Classification: classR2b},
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 90, Command: "go test ./internal/compiler/corevalidate ./internal/compiler/check -run 'Seam.*StillRefuses|EndpointFault' -v", Classification: classR2b},
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 97, Command: "go test ./internal/compiler/corevalidate -run 'ClosureCostScaling|GrowthExponent' -v", Classification: classR2b},
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 99, Command: "go test ./internal/compiler/check -run 'OrderingStability|DiagnosticSelectionOrder' -v", Classification: classR2b},
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 102, Command: "go test ./internal/compiler/check -run 'UseAfterMove|BorrowRequiresShare|TransferRequiresTake' -v", Classification: classR2b},
	{File: ".planning/milestones/M002-phases/10-trusted-interprocedural-oracle/10-RESEARCH.md", Line: 180, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/10-trusted-interprocedural-oracle/10-RESEARCH.md", Line: 180, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/10-trusted-interprocedural-oracle/10-RESEARCH.md", Line: 216, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/10-trusted-interprocedural-oracle/10-RESEARCH.md", Line: 216, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/10-trusted-interprocedural-oracle/10-RESEARCH.md", Line: 674, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/10-trusted-interprocedural-oracle/10-VALIDATION.md", Line: 27, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md", Line: 47, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-RESEARCH.md", Line: 1039, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md", Line: 26, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/12-result-payloads/12-RESEARCH.md", Line: 867, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/12-result-payloads/12-VALIDATION.md", Line: 24, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-RESEARCH.md", Line: 572, Command: "go test", Classification: classUnparseable},
	{File: ".planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md", Line: 25, Command: "go test", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-01-PLAN.md", Line: 280, Command: "go test", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md", Line: 158, Command: "go test", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md", Line: 159, Command: "go test", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md", Line: 161, Command: "go test", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md", Line: 164, Command: "go test", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md", Line: 167, Command: "go test", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md", Line: 168, Command: "go test", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md", Line: 180, Command: "go test", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md", Line: 182, Command: "go test", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-DISCUSSION-LOG.md", Line: 108, Command: "go test -list", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-RESEARCH.md", Line: 59, Command: "go test -list", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-RESEARCH.md", Line: 59, Command: "go test -json", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-RESEARCH.md", Line: 205, Command: "go test -run", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-RESEARCH.md", Line: 446, Command: "go test", Classification: classUnparseable},
	{File: ".planning/phases/14-evidence-instrument-and-honest-scoping/14-RESEARCH.md", Line: 502, Command: "go test", Classification: classUnparseable},
	{File: ".planning/research/M003/EVIDENCE-AND-DEBT.md", Line: 828, Command: "go test -list", Classification: classUnparseable},
	{File: ".planning/research/STACK.md", Line: 29, Command: "go test", Classification: classUnparseable},
	{File: ".planning/research/STACK.md", Line: 30, Command: "go test -fuzz", Classification: classUnparseable},
	{File: ".planning/research/STACK.md", Line: 492, Command: "go test -fuzz", Classification: classUnparseable},
	{File: ".planning/research/STACK.md", Line: 509, Command: "go test -fuzz", Classification: classUnparseable},
	{File: ".planning/research/SUMMARY.md", Line: 468, Command: "go test -fuzz", Classification: classUnparseable},
}

// measuredViolations runs the classifier once over the whole Tier-A
// corpus, sorted deterministically by (File, Line).
func measuredViolations(t testing.TB) []violationRecord {
	t.Helper()
	index := buildTestIndex(t)
	docs := enforcedTierDocuments(t)
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
// between pinnedFrontier and a freshly measured run, with every
// reconciled finding (plan 14-10, D-14-12) EXCLUDED first -- never
// containment, never a count. Both directions are checked and each
// failure names the specific offending record (D-14-18). A reconciled
// finding is still, textually, present in the archive it was found in
// (never rewritten in place); excluding it here does not hide it from
// TestReconciliationVerdictsCarryTheirObligations, which re-checks its
// obligation on every run and re-fails the moment the underlying archive
// text changes without the reconciliation entry being updated.
func TestVerificationGroundednessFrontierIsPinned(t *testing.T) {
	allMeasured := measuredViolations(t)
	reconciled := reconciledFindings(t)

	var measured []violationRecord
	for _, v := range allMeasured {
		if reconciled[v] {
			continue
		}
		measured = append(measured, v)
	}

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
// Plan 14-06 Task 3: corpus floors, so the lint cannot go inert by finding
// nothing (D-14-20).
// ---------------------------------------------------------------------

// enforcedTierDocumentFloor and verificationCommandFloor are measured at
// authoring time by running the FINISHED discovery and extraction once
// (`enforcedTierDocuments` + `extractCommands` over every result) and
// reading off what came out: 427 enforced-tier documents, 533 extracted
// verification commands. Per D-14-20's own instruction, these are NOT
// copied from ROADMAP.md/CONTEXT.md/plan prose -- research already found
// the document-count figures stated there are reproducible under neither
// the narrow nor the broad reading, so copying them would pin the lint to
// a number nothing in this codebase actually produced. A small buffer below
// the exact measured values (427/533) keeps the floor meaningful (it still
// fails hard the moment the glob or the extractor silently narrows) without
// making every single-document addition/removal a floor-breaking event --
// exactly the ROADMAP-vs-code-shape distinction D-14-06 draws for the grade
// bar.
const (
	enforcedTierDocumentFloor = 420
	verificationCommandFloor  = 520
)

// TestVerificationGroundednessCorpusIsNotEmpty asserts D-14-20's two
// floors. A lint whose glob silently stops matching is the same failure
// class as a `-run` pattern that silently stops matching, and a floor is
// what makes that class visible -- exactly like
// TestVerificationGroundednessIsNotInert makes classification silence
// visible, this test makes DISCOVERY silence visible.
func TestVerificationGroundednessCorpusIsNotEmpty(t *testing.T) {
	docs := enforcedTierDocuments(t)
	if len(docs) < enforcedTierDocumentFloor {
		t.Fatalf("enforced-tier document count dropped to %d, below the floor of %d -- the discovery glob may have silently narrowed", len(docs), enforcedTierDocumentFloor)
	}

	totalCommands := 0
	for _, doc := range docs {
		occurrences, err := extractCommands(doc)
		if err != nil {
			t.Fatalf("extractCommands(%s): %v", doc, err)
		}
		totalCommands += len(occurrences)
	}
	if totalCommands < verificationCommandFloor {
		t.Fatalf("extracted verification command count dropped to %d, below the floor of %d -- the extractor may have silently narrowed", totalCommands, verificationCommandFloor)
	}
	t.Logf("corpus floors: %d enforced-tier documents (floor %d), %d verification commands (floor %d)", len(docs), enforcedTierDocumentFloor, totalCommands, verificationCommandFloor)
}

// staticIndexAccuracyControlTimeout bounds the one `go test ./... -list .`
// process this accuracy control spawns. D-14-10 measured this at 3.35s
// warm on this tree; the generous ceiling exists so a wedged toolchain
// cannot hang the suite, not as a performance budget.
const staticIndexAccuracyControlTimeout = 60 * time.Second

// TestStaticTestIndexMatchesGoTestList is D-14-10's accuracy control: the
// static, go/parser-built test index (buildTestIndex) is the resolution
// mechanism every classification in this file trusts, precisely because it
// is cheap (measured ~7ms vs 3.35s warm for the toolchain's own listing).
// A cheap resolver that silently drifts from the toolchain's own notion of
// "what tests exist" would make every R2/R2b/R3 classification built on top
// of it quietly wrong in either direction. This is the ONE place in this
// file that shells out to the real `go test ./... -list .` -- a single
// whole-module invocation, not one per package (the research doc's "146
// shell-outs" framing overstates even the -list path) -- and asserts SET
// EQUALITY against the static index, in both directions, so a name present
// in one and absent from the other is a named failure. If this control ever
// comes under wall-clock pressure, the recorded escape hatch is
// content-addressing it in a later phase (QLT-12) -- never deleting it.
func TestStaticTestIndexMatchesGoTestList(t *testing.T) {
	index := buildTestIndex(t)
	staticNames := make(map[string]bool)
	for _, names := range index.byImportPath {
		for name := range names {
			staticNames[name] = true
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), staticIndexAccuracyControlTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "./...", "-list", ".")
	cmd.Dir = testsupport.ProjectPath()
	var stdout, stderr groundednessBoundedWriter
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go test ./... -list .: %v (stderr: %s)", err, stderr.bytes())
	}
	if stdout.overflowed() || stderr.overflowed() {
		t.Fatalf("go test ./... -list . exceeded %d bytes of output", groundednessMaxStreamBytes)
	}

	// `go test ./... -list .` output is, per package: one listed name per
	// line, plus a trailing "ok <pkg> <duration>" or "?  <pkg> [no test
	// files]" summary line. Neither summary form matches
	// testFuncNamePattern's exact Test/Fuzz/Benchmark/Example identifier
	// shape, so filtering on that pattern -- the SAME pattern the static
	// resolver itself requires -- discards summary lines without a second,
	// looser parser.
	listedNames := make(map[string]bool)
	for _, line := range strings.Split(string(stdout.bytes()), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && testFuncNamePattern.MatchString(line) {
			listedNames[line] = true
		}
	}

	for name := range staticNames {
		if !listedNames[name] {
			t.Errorf("static index names %q, but `go test ./... -list .` does not -- resolution mechanism drift (spoofing risk, D-14-11 threat register)", name)
		}
	}
	for name := range listedNames {
		if !staticNames[name] {
			t.Errorf("`go test ./... -list .` names %q, but the static index does not -- resolution mechanism drift", name)
		}
	}
	t.Logf("accuracy control: static index resolved %d names, go test -list resolved %d names", len(staticNames), len(listedNames))
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

// ---------------------------------------------------------------------
// Plan 14-06 Task 2: execute grep-shaped commands (R3), and detect and pin
// per-branch groundedness (R2b) -- D-14-15's third rule, D-14-16, D-14-17.
// ---------------------------------------------------------------------

// classR3 and classR2b extend the classification vocabulary this plan adds.
// classR3 is the grep-groundedness finding (D-14-15's third rule, the
// archival-breakage class); classR2b is a per-branch alternation finding,
// DETECTED AND PINNED this phase but not enforced to zero (D-14-16).
const (
	classR3  classification = "R3"
	classR2b classification = "R2b"
)

// grepShapedPattern matches the `grep`/`rg` prefix verificationCommandPattern
// already anchors extraction to.
var grepShapedPattern = regexp.MustCompile(`^(grep|rg)\b`)

func isGrepShaped(command string) bool {
	return grepShapedPattern.MatchString(command)
}

// hasShellMetacharacterOutsideQuotes reports whether command contains a
// pipe, redirect, conjunction (`&&`), or semicolon OUTSIDE any
// single/double-quoted argument, or a command-substitution opener (`$(`)
// outside SINGLE quotes specifically. A grep pattern's own regex
// alternation (`grep -nE 'TBD|FIXME|XXX'`) uses `|` INSIDE quotes and must
// never be misclassified as a shell pipe -- this is what keeps R3 from
// silently swallowing every alternation-pattern grep command in the
// corpus. Command substitution is the one metacharacter double quotes do
// NOT neutralize in POSIX shell semantics (only single quotes suppress all
// expansion), so `$(` is checked independent of inDouble -- gated only by
// inSingle.
func hasShellMetacharacterOutsideQuotes(command string) bool {
	inSingle, inDouble := false, false
	runes := []rune(command)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case r == '\'' && !inDouble:
			inSingle = !inSingle
			continue
		case r == '"' && !inSingle:
			inDouble = !inDouble
			continue
		}
		if r == '$' && !inSingle && i+1 < len(runes) && runes[i+1] == '(' {
			return true
		}
		if inSingle || inDouble {
			continue // Inside a quoted argument: literal content, not shell syntax.
		}
		switch {
		case r == '|' || r == '>' || r == ';':
			return true
		case r == '&' && i+1 < len(runes) && runes[i+1] == '&':
			return true
		}
	}
	return false
}

// grepFileOperand extracts a grep/rg command's trailing file operand by
// POSIX-tokenizing (reusing posixTokenize, no second tokenizer) and taking
// the LAST non-flag positional token after the pattern. `-e`/`-f` consume a
// following value token; every other flag token (`-c`, `-n`, `-E`, `-nE`,
// `-r`, ...) does not. A command with only one positional token (the
// pattern, e.g. a recursive `grep -rn NAME` with no explicit path) has no
// file operand and is reported as such -- R3 only applies to a LITERAL file
// operand, never to an implied recursive scope.
func grepFileOperand(command string) (operand string, ok bool) {
	tokens, err := posixTokenize(command)
	if err != nil || len(tokens) == 0 {
		return "", false
	}
	var positionals []string
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if strings.HasPrefix(tok, "-") {
			if tok == "-e" || tok == "-f" {
				i++
			}
			continue
		}
		positionals = append(positionals, tok)
	}
	if len(positionals) < 2 {
		return "", false
	}
	return positionals[len(positionals)-1], true
}

// classifyGrepGroundedness applies R3: a grep/rg-shaped command with a
// literal existing file operand and no shell metacharacter is executed
// DIRECTLY -- argv form via os/exec, never a shell -- and must yield at
// least one match; zero matches (grep's exit status 1) is a finding, naming
// the archival-breakage class D-14-15 requires this lint to also catch. A
// command carrying a pipe, redirect, conjunction, semicolon, or command
// substitution is never executed; it is classified by shape only (the
// caller's R1 check already ran, and this function reports executed=false
// so a test can assert no subprocess ran). A missing file operand is a
// finding, never a skip -- classified without spawning a process, since
// there is nothing to run against.
func classifyGrepGroundedness(projectRoot, command string) (result classification, executed bool) {
	if hasShellMetacharacterOutsideQuotes(command) {
		return classOK, false
	}
	operand, hasOperand := grepFileOperand(command)
	if !hasOperand {
		return classOK, false
	}
	resolved := operand
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(projectRoot, operand)
	}
	if _, statErr := os.Stat(resolved); statErr != nil {
		return classR3, false
	}
	tokens, err := posixTokenize(command)
	if err != nil {
		return classUnparseable, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), groundednessGitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, tokens[0], tokens[1:]...)
	cmd.Dir = projectRoot
	var stdout, stderr groundednessBoundedWriter
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if runErr == nil {
		return classOK, true
	}
	// Any non-zero exit (grep's documented "no lines selected" status 1, or
	// any other failure) is a finding -- D-14-15's "no silent skips" rule
	// applies here exactly as it does to an unparseable go-test cell.
	return classR3, true
}

// splitTopLevelAlternation splits a go-test -run pattern's first segment on
// `|`. The pattern has already been through D-14-17's markdown-unescape
// ruling by the time it reaches here (extractCommands/unescapeCell already
// ran), and every alternation pattern actually present in this corpus is a
// flat `A|B|...` with no grouping metacharacter, so a literal top-level
// split is exact.
func splitTopLevelAlternation(pattern string) []string {
	return strings.Split(pattern, "|")
}

// classifyPerBranchGroundedness applies R2b (D-14-16): split the run
// pattern's first segment on top-level alternation and require EVERY
// branch to resolve to at least one name in the union index -- this is
// what turns `-run 'PeerLiveness|LoanChainIndex'` from a pass into a
// finding when only one branch resolves. Returns every branch that fails to
// resolve; the caller decides how to record that (Task 2 logs it, Task 3
// wires it into the pinned frontier as classR2b, DETECTED AND PINNED but
// not enforced to zero this phase per D-14-16's explicit sizing decision).
func classifyPerBranchGroundedness(index *testIndex, packages []string, pattern string) (failingBranches []string) {
	names := make(map[string]bool)
	for _, operand := range packages {
		set, _ := index.resolvePackageNames(operand)
		for name := range set {
			names[name] = true
		}
	}
	segment := strings.SplitN(pattern, "/", 2)[0]
	for _, branch := range splitTopLevelAlternation(segment) {
		re, err := regexp.Compile(branch)
		if err != nil {
			failingBranches = append(failingBranches, branch)
			continue
		}
		matched := false
		for name := range names {
			if re.MatchString(name) {
				matched = true
				break
			}
		}
		if !matched {
			failingBranches = append(failingBranches, branch)
		}
	}
	return failingBranches
}

// TestVerificationGroundednessGrepExecution exercises Task 2's grep/R3
// <behavior> bullets against synthetic fixture files under t.TempDir().
func TestVerificationGroundednessGrepExecution(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "match.txt"), []byte("needle found here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "nomatch.txt"), []byte("nothing interesting\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("an existing file with a guaranteed match classifies clean", func(t *testing.T) {
		got, executed := classifyGrepGroundedness(root, "grep -c 'needle' match.txt")
		if !executed {
			t.Fatal("expected the grep subprocess to run")
		}
		if got != classOK {
			t.Fatalf("classifyGrepGroundedness(match) = %s, want ok", got)
		}
	})

	t.Run("an existing file with no match is a finding", func(t *testing.T) {
		got, executed := classifyGrepGroundedness(root, "grep -c 'needle' nomatch.txt")
		if !executed {
			t.Fatal("expected the grep subprocess to run")
		}
		if got != classR3 {
			t.Fatalf("classifyGrepGroundedness(no match) = %s, want R3", got)
		}
	})

	t.Run("a missing file operand is a finding, not a skip", func(t *testing.T) {
		got, executed := classifyGrepGroundedness(root, "grep -c 'needle' does-not-exist.txt")
		if executed {
			t.Fatal("expected no subprocess for a missing file operand")
		}
		if got != classR3 {
			t.Fatalf("classifyGrepGroundedness(missing file) = %s, want R3", got)
		}
	})

	t.Run("a command carrying a pipe is proven never executed and is classified by shape only", func(t *testing.T) {
		got, executed := classifyGrepGroundedness(root, "grep -c 'needle' match.txt | wc -l")
		if executed {
			t.Fatal("expected the piped command to never be executed")
		}
		if got != classOK {
			t.Fatalf("classifyGrepGroundedness(piped) = %s, want ok (shape-classified, never executed)", got)
		}
	})

	t.Run("a command carrying a redirect is never executed", func(t *testing.T) {
		_, executed := classifyGrepGroundedness(root, "grep -c 'needle' match.txt > out.txt")
		if executed {
			t.Fatal("expected the redirected command to never be executed")
		}
	})

	t.Run("a command carrying a conjunction is never executed", func(t *testing.T) {
		_, executed := classifyGrepGroundedness(root, "grep -c 'needle' match.txt && echo done")
		if executed {
			t.Fatal("expected the conjunction command to never be executed")
		}
	})

	t.Run("a command carrying a semicolon is never executed", func(t *testing.T) {
		_, executed := classifyGrepGroundedness(root, "grep -c 'needle' match.txt; echo done")
		if executed {
			t.Fatal("expected the semicolon-joined command to never be executed")
		}
	})

	t.Run("a command carrying a command substitution is never executed", func(t *testing.T) {
		_, executed := classifyGrepGroundedness(root, "grep -c \"$(cat match.txt)\" match.txt")
		if executed {
			t.Fatal("expected the command-substitution command to never be executed")
		}
	})

	t.Run("a pattern-internal alternation pipe inside quotes is not mistaken for a shell pipe", func(t *testing.T) {
		if hasShellMetacharacterOutsideQuotes(`grep -nE 'TBD|FIXME|XXX' match.txt`) {
			t.Fatal("a quoted alternation pipe was misclassified as a shell metacharacter")
		}
		got, executed := classifyGrepGroundedness(root, "grep -nE 'needle|absent' match.txt")
		if !executed {
			t.Fatal("expected the subprocess to run: the pipe is inside quotes, not a shell pipe")
		}
		if got != classOK {
			t.Fatalf("classifyGrepGroundedness(quoted alternation) = %s, want ok", got)
		}
	})

	t.Run("a recursive grep with no file operand is not R3-eligible", func(t *testing.T) {
		_, executed := classifyGrepGroundedness(root, "grep -rn needle")
		if executed {
			t.Fatal("expected no subprocess for a command with no file operand")
		}
	})

	// Argv-form-only invariant (Task 2 acceptance criteria): this file must
	// never pass a command string to a shell. Verified directly, executed
	// exactly the way every other spawn in this module is (bounded,
	// deadline-carrying, argv form). The search pattern is assembled at
	// runtime from parts, never written as one contiguous literal in this
	// file's own source -- otherwise this very assertion would be a
	// self-referential false positive, matching its own source line.
	t.Run("this file never passes a command string to a shell", func(t *testing.T) {
		shellInvocationPattern := strings.Join([]string{`sh`, `"`, `, `, `"`, `-c`, `"`}, "")
		out := runGroundednessSelfCheckGrep(t, shellInvocationPattern, "verification_groundedness_test.go")
		if strings.TrimSpace(out) != "0" {
			t.Fatalf("grep -c %q verification_groundedness_test.go = %q, want \"0\" (no shell invocation present)", shellInvocationPattern, strings.TrimSpace(out))
		}
	})
}

// runGroundednessSelfCheckGrep runs `grep -c PATTERN FILE` over this
// package's own source directory, bounded exactly like every other spawn in
// this file, and returns stdout. Used only by the argv-form-only self-check
// above -- not part of the lint's production classification path.
func runGroundednessSelfCheckGrep(t *testing.T, pattern, relFile string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), groundednessGitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "grep", "-c", pattern, relFile)
	cmd.Dir = testsupport.ProjectPath("internal", "compiler", "session")
	var stdout, stderr groundednessBoundedWriter
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !(errors.As(err, &exitErr) && exitErr.ExitCode() == 1) {
			t.Fatalf("grep -c %q %s: %v (stderr: %s)", pattern, relFile, err, stderr.bytes())
		}
	}
	return string(stdout.bytes())
}

// TestVerificationGroundednessGrepOverRealCorpus proves the grep classifier
// finds the real, currently-live archival-breakage instance in the Tier-A
// corpus: `.planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md:115`
// (line 113 at plan 14-06's measurement; shifted +2 by plan 14-09's D-14-05
// frontmatter migration -- a pure coordinate shift, re-measured here)
// (`grep -c "S-008" .planning/ROADMAP.md`) returns zero matches today --
// "S-008" no longer appears in the live ROADMAP.md, exactly the archival
// breakage D-14-15's R3 exists to catch. (Task 2's read_first names this as
// "the grep row over the roadmap that returns zero matches since
// archival" under the 11-VALIDATION.md heading; the verified live instance
// is this 09-VALIDATION.md row -- recorded in the plan 14-06 SUMMARY.)
func TestVerificationGroundednessGrepOverRealCorpus(t *testing.T) {
	root := testsupport.ProjectPath()
	docs := tierADocuments(t)
	var grepFindings []violationRecord
	for _, doc := range docs {
		occurrences, err := extractCommands(doc)
		if err != nil {
			t.Fatalf("extractCommands(%s): %v", doc, err)
		}
		for _, occ := range occurrences {
			if !isGrepShaped(occ.Command) {
				continue
			}
			if !classifyRunnability(occ.Command) {
				continue // already R1 -- do not double-report
			}
			class, executed := classifyGrepGroundedness(root, occ.Command)
			t.Logf("grep-shaped (executed=%v) %s %s:%d: %s", executed, class, projectRelativePath(root, doc), occ.Line, occ.Command)
			if class == classR3 {
				grepFindings = append(grepFindings, violationRecord{
					File:           projectRelativePath(root, doc),
					Line:           occ.Line,
					Command:        occ.Command,
					Classification: classR3,
				})
			}
		}
	}
	found := false
	for _, v := range grepFindings {
		if strings.HasSuffix(v.File, "09-VALIDATION.md") && v.Line == 115 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected an R3 finding at .../09-VALIDATION.md:115 (grep -c \"S-008\" .planning/ROADMAP.md), found none among %d grep findings: %+v", len(grepFindings), grepFindings)
	}
}

// TestVerificationGroundednessPerBranch exercises Task 2's per-branch/R2b
// <behavior> bullets.
func TestVerificationGroundednessPerBranch(t *testing.T) {
	index := &testIndex{byImportPath: map[string]map[string]bool{
		groundednessModulePath + "/internal/compiler/example": {
			"TestPeerLivenessFileImportsStayIndependent": true,
		},
	}}

	t.Run("every branch resolving is clean", func(t *testing.T) {
		failing := classifyPerBranchGroundedness(index, []string{"./internal/compiler/example/..."}, "PeerLiveness|TestPeerLivenessFileImportsStayIndependent")
		if len(failing) != 0 {
			t.Fatalf("want no failing branches, got %v", failing)
		}
	})

	t.Run("one branch resolving to zero names is a per-branch finding", func(t *testing.T) {
		failing := classifyPerBranchGroundedness(index, []string{"./internal/compiler/example/..."}, "PeerLiveness|LoanChainIndex")
		if len(failing) != 1 || failing[0] != "LoanChainIndex" {
			t.Fatalf("want exactly one failing branch (LoanChainIndex), got %v", failing)
		}
	})

	t.Run("per-branch findings are recorded, not enforced to zero (D-14-16)", func(t *testing.T) {
		root := testsupport.ProjectPath()
		realIndex := buildTestIndex(t)
		docs := tierADocuments(t)
		var perBranchCount int
		for _, doc := range docs {
			occurrences, err := extractCommands(doc)
			if err != nil {
				t.Fatalf("extractCommands(%s): %v", doc, err)
			}
			for _, occ := range occurrences {
				if !strings.HasPrefix(occ.Command, "go test") || !classifyRunnability(occ.Command) {
					continue
				}
				parsed, ok := parseGoTestCommand(occ.Command)
				if !ok || parsed.Pattern == "" || !strings.Contains(parsed.Pattern, "|") {
					continue
				}
				failing := classifyPerBranchGroundedness(realIndex, parsed.Packages, parsed.Pattern)
				if len(failing) > 0 {
					perBranchCount++
					t.Logf("R2b %s:%d: %s -- failing branches: %v", projectRelativePath(root, doc), occ.Line, occ.Command, failing)
				}
			}
		}
		t.Logf("total per-branch (R2b) findings over the real Tier-A corpus: %d", perBranchCount)
		if perBranchCount == 0 {
			t.Fatal("expected at least one real per-branch finding (D-14-16 predicts the defect mass is here); found none -- classifier may be broken")
		}
	})
}

// ---------------------------------------------------------------------
// Plan 14-10 Task 3: the reconciliation exclusion (D-14-18's frontier
// must move for a genuine reason) and the closed R2b landing-phase
// register (D-14-16).
// ---------------------------------------------------------------------

// reconciledFindings returns the set of (File,Line,Command,Classification)
// covered by a reconciliation entry across every *-DEBT.md register
// (session_test.go's parseReconciliationEntries via deriveEvidenceReconciliation,
// witness_registry_test.go). A finding in this set is EXCLUDED from
// pinnedFrontier's comparison -- it is still, textually, a dead pattern in
// the archive (never rewritten in place), but it is now adjudicated
// outside the archive under a verdict TestReconciliationVerdictsCarryTheirObligations
// re-checks on every run, so the lint no longer needs to also pin it.
func reconciledFindings(t testing.TB) map[violationRecord]bool {
	t.Helper()
	entries, err := deriveEvidenceReconciliation()
	if err != nil {
		t.Fatal(err)
	}
	set := make(map[violationRecord]bool, len(entries))
	for _, e := range entries {
		set[violationRecord{File: e.File, Line: e.Line, Command: e.Command, Classification: e.Classification}] = true
	}
	return set
}

// r2bLandingPhases assigns a landing phase to every pinned R2b (per-branch)
// finding, per D-14-16's sizing decision: this phase drives runnability,
// groundedness and grep (R1/R2/R3) to empty via reconciliation, but
// per-branch (R2b) closure is judgment work explicitly deferred to
// QLT-10 (ROADMAP.md Phase 20). Keyed by the exact pinnedFrontier record
// so a mismatch (an R2b entry with no landing phase, or a landing phase
// for an entry no longer pinned) is a hard failure, never silently
// tolerated.
var r2bLandingPhases = map[violationRecord]string{
	{File: ".planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-VERIFICATION.md", Line: 86, Command: "go test ./internal/compiler/corevalidate -run '^(TestForeignPolicyValueNotIdentifierRefused|TestForeignContractCommentSafetyRefused)$' -v", Classification: classR2b}:                                                                                      "P20",
	{File: ".planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-VERIFICATION.md", Line: 88, Command: "go test ./internal/compiler/corevalidate -run '^(TestInteriorMergeDivergentHistoriesRefused|TestInteriorMergeAgreeingHistoriesAccepted|TestCyclicOkEdgeChainRefusedNotHung|TestAcyclicChainsStillValidateUnderCycleGuard)$' -v", Classification: classR2b}: "P20",
	{File: ".planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-VERIFICATION.md", Line: 89, Command: "go test ./internal/compiler/cgen -run '^(TestForeignSymbolInjectionNeverReachesGeneratedC|TestForeignEmittersRefuseNonIdentifierSymbolIndependently|TestExistingEmittersAreByteIdentical)$' -v", Classification: classR2b}:                                 "P20",
	{File: ".planning/milestones/M001-phases/05-native-equivalence-and-adversarial-evidence/05-VERIFICATION.md", Line: 56, Command: "go test ./internal/compiler/session/... -run \"TestNAT03|TestAssertMutationMovesAnAxis|TestEveryNAT03\" -v", Classification: classR2b}:                                                                                                           "P20",
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 86, Command: "go test ./internal/compiler/corevalidate -run 'PeerLiveness|LoanChainIndex' -v", Classification: classR2b}:                                                                                                                                              "P20",
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 89, Command: "go test ./internal/compiler/corevalidate -run 'ImportsStayIndependent|ImportIndependence' -v", Classification: classR2b}:                                                                                                                                "P20",
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 90, Command: "go test ./internal/compiler/corevalidate ./internal/compiler/check -run 'Seam.*StillRefuses|EndpointFault' -v", Classification: classR2b}:                                                                                                               "P20",
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 97, Command: "go test ./internal/compiler/corevalidate -run 'ClosureCostScaling|GrowthExponent' -v", Classification: classR2b}:                                                                                                                                        "P20",
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 99, Command: "go test ./internal/compiler/check -run 'OrderingStability|DiagnosticSelectionOrder' -v", Classification: classR2b}:                                                                                                                                      "P20",
	{File: ".planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md", Line: 102, Command: "go test ./internal/compiler/check -run 'UseAfterMove|BorrowRequiresShare|TransferRequiresTake' -v", Classification: classR2b}:                                                                                                                          "P20",
}

// TestVerificationGroundednessThreeClassesAreEmpty is plan 14-10 Task 3's
// own closing assertion (D-14-18, D-14-19): over a freshly measured run
// with reconciled findings excluded, the runnability (R1), groundedness
// (R2) and grep (R3) classes are empty, the per-branch (R2b) class is
// non-empty with every member carrying a closed-vocabulary landing
// phase, and the corpus floors still hold in the SAME run -- so "zero
// findings" is only reportable over a corpus that is still being read
// (D-14-64).
func TestVerificationGroundednessThreeClassesAreEmpty(t *testing.T) {
	measured := measuredViolations(t)
	reconciled := reconciledFindings(t)

	var r1, r2, r3, r2b int
	for _, v := range measured {
		if reconciled[v] {
			continue
		}
		switch v.Classification {
		case classR1:
			r1++
			t.Errorf("unreconciled R1 finding remains: %s:%d: %s", v.File, v.Line, v.Command)
		case classR2:
			r2++
			t.Errorf("unreconciled R2 finding remains: %s:%d: %s", v.File, v.Line, v.Command)
		case classR3:
			r3++
			t.Errorf("unreconciled R3 finding remains: %s:%d: %s", v.File, v.Line, v.Command)
		case classR2b:
			r2b++
			if _, owned := r2bLandingPhases[v]; !owned {
				t.Errorf("R2b finding has no landing phase: %s:%d: %s", v.File, v.Line, v.Command)
			} else if phase := r2bLandingPhases[v]; !debtRegisterOwningPhaseForm(phase) {
				t.Errorf("R2b finding %s:%d has landing phase %q outside the closed P<NN>|CLOSED(<sha>)|UNOWNED(<witness>) vocabulary", v.File, v.Line, phase)
			}
		}
	}
	if r2b == 0 {
		t.Error("expected the per-branch (R2b) class to be non-empty -- D-14-16's sizing decision keeps it pinned, not enforced to zero, this phase")
	}
	for key := range r2bLandingPhases {
		found := false
		for _, v := range measured {
			if v == key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("r2bLandingPhases names a finding the tree no longer produces: %s:%d: %s", key.File, key.Line, key.Command)
		}
	}
	t.Logf("post-reconciliation: R1=%d R2=%d R3=%d (all must be 0), R2b=%d (non-zero, every member owned)", r1, r2, r3, r2b)

	// The corpus floors, asserted in the SAME run (D-14-64): "zero
	// findings" over an empty corpus would be a vacuous, not honest,
	// claim.
	docs := enforcedTierDocuments(t)
	if len(docs) < enforcedTierDocumentFloor {
		t.Fatalf("enforced-tier document count dropped to %d, below the floor of %d", len(docs), enforcedTierDocumentFloor)
	}
	totalCommands := 0
	for _, doc := range docs {
		occurrences, err := extractCommands(doc)
		if err != nil {
			t.Fatalf("extractCommands(%s): %v", doc, err)
		}
		totalCommands += len(occurrences)
	}
	if totalCommands < verificationCommandFloor {
		t.Fatalf("extracted verification command count dropped to %d, below the floor of %d", totalCommands, verificationCommandFloor)
	}
	t.Logf("corpus floors held in the same run: %d enforced-tier documents, %d verification commands", len(docs), totalCommands)
}
