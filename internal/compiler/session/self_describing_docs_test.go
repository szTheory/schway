package session

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// This file machine-checks two of the project's self-describing artefacts
// (EVD-06, EVD-08): .planning/LANGUAGE-MATURITY.md's stated counts, and the
// QLT-02 budget manifest's suite wall-clock baseline.
//
// Independence discipline (the reasoning this file exists to apply): a
// self-describing document must never be allowed to certify itself by
// supplying the command that checks it. LANGUAGE-MATURITY.md carries its own
// "Re-verify cheaply" awk/grep block, and the tempting shortcut is to shell
// out to that block and compare its output to the document's stated numbers.
// That shortcut is refused on purpose: a document that supplies the command
// which checks it can be made green by editing that command, which is
// exactly the artifact-read defect this phase (EVD-01 through EVD-08) exists
// to retire. Every count below is instead re-derived independently in Go,
// walking the module tree from testsupport.ProjectPath, and this file never
// spawns a child process and never runs text read out of either document.
// TestSelfDescribingDocsGuardIsNotInert (Task 3) proves that independence
// rather than merely asserting it: a copy of the maturity document whose
// embedded re-verify line is rewritten to a command that would report a
// different number does not change this file's verdict, because that line
// is never executed.

// isLenFunctionsCall matches a `len(X.Functions)` call expression, the
// argument shape every single-function guard in this tree uses.
func isLenFunctionsCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return false
	}
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "len" {
		return false
	}
	sel, ok := call.Args[0].(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "Functions"
}

// isFunctionsGuardExpr matches `len(X.Functions) != 1`, the exact structural
// shape LANGUAGE-MATURITY.md's guard table counts (the same shape its own
// embedded `awk` re-verify line greps for as plain text). Reproducing it as
// a syntax-tree match rather than a text match is what makes this
// comment-immune: parser.ParseFile discards comments from the expression
// tree ast.Inspect walks, so a comment that merely MENTIONS the pattern
// (this tree has several, e.g. session.go's own "the old
// len(program.Functions) != 1 guard" prose) can never be miscounted as a
// guard here, unlike the document's own awk line.
func isFunctionsGuardExpr(n ast.Node) bool {
	be, ok := n.(*ast.BinaryExpr)
	if !ok || be.Op != token.NEQ {
		return false
	}
	if !isLenFunctionsCall(be.X) {
		return false
	}
	lit, ok := be.Y.(*ast.BasicLit)
	return ok && lit.Kind == token.INT && lit.Value == "1"
}

// functionsGuardScan is one pass's independently re-derived guard counts.
type functionsGuardScan struct {
	NonTestTotal        int
	NonTestByPackage    map[string]int
	NonTestFileCount    int
	NonTestPackageCount int
	AllTotal            int
}

// scanFunctionsGuards walks every *.go file under internal/ and cmd/ inside
// root, parses each with go/parser (never shelling out, never regex-matching
// raw text), and counts isFunctionsGuardExpr matches. Build-constrained
// files are resolved through go/build.Context.MatchFile, matching this
// package's existing static-scan convention (see the groundedness lint's own
// static test index).
func scanFunctionsGuards(root string) (functionsGuardScan, error) {
	var out functionsGuardScan
	out.NonTestByPackage = map[string]int{}
	nonTestFiles := map[string]bool{}
	nonTestPackages := map[string]bool{}

	fset := token.NewFileSet()
	buildCtx := build.Default

	for _, base := range []string{"internal", "cmd"} {
		start := filepath.Join(root, base)
		walkErr := filepath.Walk(start, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			dir, base := filepath.Split(path)
			matched, matchErr := buildCtx.MatchFile(dir, base)
			if matchErr != nil {
				return matchErr
			}
			if !matched {
				return nil
			}
			file, parseErr := parser.ParseFile(fset, path, nil, 0)
			if parseErr != nil {
				return parseErr
			}
			count := 0
			ast.Inspect(file, func(n ast.Node) bool {
				if isFunctionsGuardExpr(n) {
					count++
				}
				return true
			})
			if count == 0 {
				return nil
			}
			out.AllTotal += count
			if !strings.HasSuffix(path, "_test.go") {
				out.NonTestTotal += count
				out.NonTestByPackage[file.Name.Name] += count
				nonTestFiles[path] = true
				nonTestPackages[file.Name.Name] = true
			}
			return nil
		})
		if walkErr != nil {
			return functionsGuardScan{}, walkErr
		}
	}
	out.NonTestFileCount = len(nonTestFiles)
	out.NonTestPackageCount = len(nonTestPackages)
	return out, nil
}

// corpusStats walks the whole tree rooted at root (skipping .git) counting
// *.schway files and their total newline count — reimplementing, in Go, what
// the document's own "re-verify cheaply" shell block computes with `find`
// and `wc -l`, rather than running that block.
func corpusStats(root string) (programs int, lines int, err error) {
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".schway") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		programs++
		lines += strings.Count(string(content), "\n")
		return nil
	})
	if walkErr != nil {
		return 0, 0, walkErr
	}
	return programs, lines, nil
}

var (
	guardTotalSentenceRe = regexp.MustCompile("\\*\\*(\\d+) `len\\(Functions\\) != 1` guards across (\\d+) files in (\\d+) packages\\*\\* \\((\\d+) including tests\\)")
	guardTableRowRe      = regexp.MustCompile("^\\| `([A-Za-z0-9_]+)` \\| (\\d+) \\|")
	corpusSentenceRe     = regexp.MustCompile("\\*\\*(\\d+) `\\.schway` programs, ([\\d,]+) lines total\\*\\*")
)

// normalizeWhitespace collapses runs of whitespace (including newlines) to a
// single space, so a bolded sentence that happens to be hard-wrapped across
// lines in the markdown source is still matched by a single-line regexp.
func normalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// checkLanguageMaturityDoc re-reads docPath and root fresh on every call (no
// shared state, no caching), parses the document's stated numbers out of its
// text, independently re-derives the same facts from root, and returns one
// finding per mismatch, each naming the document's path, line number, and
// both the stated and re-derived values. Factored to take a document path
// and a source root so TestSelfDescribingDocsGuardIsNotInert (Task 3) can run
// it against temp copies without touching the real tree; the real-corpus
// test (TestLanguageMaturityCountsAreCurrent) calls the same function.
func checkLanguageMaturityDoc(docPath, root string) ([]string, error) {
	raw, err := os.ReadFile(docPath)
	if err != nil {
		return nil, err
	}
	content := string(raw)
	docLines := strings.Split(content, "\n")

	scan, err := scanFunctionsGuards(root)
	if err != nil {
		return nil, err
	}
	programs, lines, err := corpusStats(root)
	if err != nil {
		return nil, err
	}

	lineOf := func(anchor string) int {
		for i, l := range docLines {
			if strings.Contains(l, anchor) {
				return i + 1
			}
		}
		return -1
	}

	var findings []string

	// --- guard total, file count, package count, including-tests total ---
	m := guardTotalSentenceRe.FindStringSubmatch(normalizeWhitespace(content))
	if m == nil {
		findings = append(findings, fmt.Sprintf("%s: could not locate the guard-total sentence in the expected '**N `len(Functions) != 1` guards across N files in N packages** (N including tests)' shape", docPath))
	} else {
		statedTotal, _ := strconv.Atoi(m[1])
		statedFiles, _ := strconv.Atoi(m[2])
		statedPackages, _ := strconv.Atoi(m[3])
		statedAll, _ := strconv.Atoi(m[4])
		line := lineOf("guards across")
		if statedTotal != scan.NonTestTotal {
			findings = append(findings, fmt.Sprintf("%s:%d: stated non-test guard total %d does not match re-derived %d", docPath, line, statedTotal, scan.NonTestTotal))
		}
		if statedFiles != scan.NonTestFileCount {
			findings = append(findings, fmt.Sprintf("%s:%d: stated guard file count %d does not match re-derived %d", docPath, line, statedFiles, scan.NonTestFileCount))
		}
		if statedPackages != scan.NonTestPackageCount {
			findings = append(findings, fmt.Sprintf("%s:%d: stated guard package count %d does not match re-derived %d", docPath, line, statedPackages, scan.NonTestPackageCount))
		}
		if statedAll != scan.AllTotal {
			findings = append(findings, fmt.Sprintf("%s:%d: stated including-tests guard total %d does not match re-derived %d", docPath, line, statedAll, scan.AllTotal))
		}
	}

	// --- per-package breakdown table ---
	seenPackages := map[string]bool{}
	inGuardTable := false
	for i, l := range docLines {
		if strings.HasPrefix(strings.TrimSpace(l), "| Package | Guards |") {
			inGuardTable = true
			continue
		}
		if inGuardTable && strings.TrimSpace(l) == "" {
			inGuardTable = false
			continue
		}
		if !inGuardTable {
			continue
		}
		mm := guardTableRowRe.FindStringSubmatch(l)
		if mm == nil {
			continue
		}
		pkg := mm[1]
		stated, _ := strconv.Atoi(mm[2])
		seenPackages[pkg] = true
		derived := scan.NonTestByPackage[pkg]
		if stated != derived {
			findings = append(findings, fmt.Sprintf("%s:%d: stated guard count for package %q is %d, does not match re-derived %d", docPath, i+1, pkg, stated, derived))
		}
	}
	for pkg, derived := range scan.NonTestByPackage {
		if !seenPackages[pkg] {
			findings = append(findings, fmt.Sprintf("%s: package %q has %d re-derived guards but no row in the guard breakdown table", docPath, pkg, derived))
		}
	}

	// --- corpus program count + line count ---
	cm := corpusSentenceRe.FindStringSubmatch(normalizeWhitespace(content))
	if cm == nil {
		findings = append(findings, fmt.Sprintf("%s: could not locate the corpus-count sentence in the expected '**N `.schway` programs, N lines total**' shape", docPath))
	} else {
		statedPrograms, _ := strconv.Atoi(cm[1])
		statedLines, _ := strconv.Atoi(strings.ReplaceAll(cm[2], ",", ""))
		line := lineOf("lines total**")
		if statedPrograms != programs {
			findings = append(findings, fmt.Sprintf("%s:%d: stated corpus program count %d does not match re-derived %d", docPath, line, statedPrograms, programs))
		}
		if statedLines != lines {
			findings = append(findings, fmt.Sprintf("%s:%d: stated corpus line count %d does not match re-derived %d", docPath, line, statedLines, lines))
		}
	}

	return findings, nil
}

// suiteWallClockMetric is the QLT-02 budget-manifest metric name for the
// full-suite wall-clock baseline EVD-08 adds (Task 2): a cold, observed
// measurement of `go test ./...` taken during this phase, never a figure
// copied out of prose.
const suiteWallClockMetric = "suite_wall_clock_ns"

// checkSuiteWallClockObservedRow re-parses manifestPath directly (never
// through the package's go:embed, so it also works unchanged against a temp
// copy) and asserts exactly one gate_type "observed" row names
// suiteWallClockMetric. Factored to take a manifest path so
// TestSelfDescribingDocsGuardIsNotInert can run it against a temp copy with
// the row removed, without touching the checked-in manifest.
func checkSuiteWallClockObservedRow(manifestPath string) ([]string, error) {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	var rows []QLT02BudgetRow
	if jsonErr := json.Unmarshal(raw, &rows); jsonErr != nil {
		return nil, fmt.Errorf("%s: invalid JSON: %w", manifestPath, jsonErr)
	}
	count := 0
	var findings []string
	for _, row := range rows {
		if row.Metric != suiteWallClockMetric {
			continue
		}
		count++
		if row.GateType != QLT02GateTypeObserved {
			findings = append(findings, fmt.Sprintf("%s: %s row has gate_type %q, want %q", manifestPath, suiteWallClockMetric, row.GateType, QLT02GateTypeObserved))
		}
	}
	if count == 0 {
		findings = append(findings, fmt.Sprintf("%s: no %s row found", manifestPath, suiteWallClockMetric))
	} else if count > 1 {
		findings = append(findings, fmt.Sprintf("%s: %d %s rows found, want exactly 1", manifestPath, count, suiteWallClockMetric))
	}
	return findings, nil
}

// TestLanguageMaturityCountsAreCurrent is EVD-06's real-corpus check: every
// count .planning/LANGUAGE-MATURITY.md states about the tree must equal this
// file's independent re-derivation of that same count.
func TestLanguageMaturityCountsAreCurrent(t *testing.T) {
	docPath := testsupport.ProjectPath(".planning", "LANGUAGE-MATURITY.md")
	root := testsupport.ProjectPath()
	findings, err := checkLanguageMaturityDoc(docPath, root)
	if err != nil {
		t.Fatalf("checkLanguageMaturityDoc: %v", err)
	}
	for _, f := range findings {
		t.Error(f)
	}
}

// TestSelfDescribingDocsGuardIsNotInert proves both self-checks above are
// not inert (D-14-28's shape): one seeded fault per mechanizable kind, plus
// an unmodified-copy control for each artefact, all under t.TempDir() so the
// real tree is never touched.
func TestSelfDescribingDocsGuardIsNotInert(t *testing.T) {
	docPath := testsupport.ProjectPath(".planning", "LANGUAGE-MATURITY.md")
	root := testsupport.ProjectPath()
	manifestPath := testsupport.ProjectPath("internal", "compiler", "session", "qlt02_budget_manifest.json")

	originalDoc, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("read maturity doc: %v", err)
	}
	originalManifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	t.Run("unmodified maturity doc copy passes", func(t *testing.T) {
		copyPath := filepath.Join(t.TempDir(), "LANGUAGE-MATURITY.md")
		if writeErr := os.WriteFile(copyPath, originalDoc, 0o644); writeErr != nil {
			t.Fatal(writeErr)
		}
		findings, checkErr := checkLanguageMaturityDoc(copyPath, root)
		if checkErr != nil {
			t.Fatal(checkErr)
		}
		if len(findings) != 0 {
			t.Fatalf("unmodified copy reported findings: %v", findings)
		}
	})

	t.Run("changed stated count fails naming the line and both numbers", func(t *testing.T) {
		loc := guardTotalSentenceRe.FindStringIndex(string(originalDoc))
		if loc == nil {
			t.Fatal("guard-total sentence not found in the real document -- cannot seed this fault")
		}
		sentence := string(originalDoc[loc[0]:loc[1]])
		mutatedSentence := regexp.MustCompile(`\*\*(\d+) `).ReplaceAllStringFunc(sentence, func(s string) string {
			digits := regexp.MustCompile(`\d+`).FindString(s)
			n, _ := strconv.Atoi(digits)
			return strings.Replace(s, digits, strconv.Itoa(n+1), 1)
		})
		if mutatedSentence == sentence {
			t.Fatal("mutation did not change the guard-total sentence")
		}
		mutatedDoc := string(originalDoc[:loc[0]]) + mutatedSentence + string(originalDoc[loc[1]:])

		copyPath := filepath.Join(t.TempDir(), "LANGUAGE-MATURITY.md")
		if writeErr := os.WriteFile(copyPath, []byte(mutatedDoc), 0o644); writeErr != nil {
			t.Fatal(writeErr)
		}
		findings, checkErr := checkLanguageMaturityDoc(copyPath, root)
		if checkErr != nil {
			t.Fatal(checkErr)
		}
		if len(findings) == 0 {
			t.Fatal("changing the stated guard total by one did not fail the count self-check")
		}
		named := false
		for _, f := range findings {
			if strings.Contains(f, copyPath) {
				named = true
			}
		}
		if !named {
			t.Fatalf("findings do not name the document path: %v", findings)
		}
	})

	t.Run("rewritten re-verify command does not change the verdict", func(t *testing.T) {
		anchor := "Re-verify (approximate only"
		if !strings.Contains(string(originalDoc), anchor) {
			t.Fatalf("anchor %q not found in the real document -- cannot seed this fault", anchor)
		}
		mutatedDoc := strings.Replace(string(originalDoc), anchor,
			"Re-verify (REWRITTEN by a seeded fault to a command that would report a wrong number, e.g. `echo 999999`, and must never be executed", 1)
		if mutatedDoc == string(originalDoc) {
			t.Fatal("mutation did not change the document")
		}

		copyPath := filepath.Join(t.TempDir(), "LANGUAGE-MATURITY.md")
		if writeErr := os.WriteFile(copyPath, []byte(mutatedDoc), 0o644); writeErr != nil {
			t.Fatal(writeErr)
		}
		findings, checkErr := checkLanguageMaturityDoc(copyPath, root)
		if checkErr != nil {
			t.Fatal(checkErr)
		}
		if len(findings) != 0 {
			t.Fatalf("rewriting the embedded re-verify command changed the verdict (it must never be executed): %v", findings)
		}
	})

	t.Run("unmodified manifest copy passes", func(t *testing.T) {
		copyPath := filepath.Join(t.TempDir(), "qlt02_budget_manifest.json")
		if writeErr := os.WriteFile(copyPath, originalManifest, 0o644); writeErr != nil {
			t.Fatal(writeErr)
		}
		findings, checkErr := checkSuiteWallClockObservedRow(copyPath)
		if checkErr != nil {
			t.Fatal(checkErr)
		}
		if len(findings) != 0 {
			t.Fatalf("unmodified manifest copy reported findings: %v", findings)
		}
	})

	t.Run("manifest missing the suite wall-clock row fails", func(t *testing.T) {
		var rows []QLT02BudgetRow
		if jsonErr := json.Unmarshal(originalManifest, &rows); jsonErr != nil {
			t.Fatal(jsonErr)
		}
		var withoutRow []QLT02BudgetRow
		for _, row := range rows {
			if row.Metric == suiteWallClockMetric {
				continue
			}
			withoutRow = append(withoutRow, row)
		}
		if len(withoutRow) != len(rows)-1 {
			t.Fatalf("expected to remove exactly one row, removed %d of %d", len(rows)-len(withoutRow), len(rows))
		}
		encoded, marshalErr := json.MarshalIndent(withoutRow, "", "  ")
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}

		copyPath := filepath.Join(t.TempDir(), "qlt02_budget_manifest.json")
		if writeErr := os.WriteFile(copyPath, encoded, 0o644); writeErr != nil {
			t.Fatal(writeErr)
		}
		findings, checkErr := checkSuiteWallClockObservedRow(copyPath)
		if checkErr != nil {
			t.Fatal(checkErr)
		}
		if len(findings) == 0 {
			t.Fatal("removing the suite wall-clock row did not fail the manifest self-check")
		}
	})
}
