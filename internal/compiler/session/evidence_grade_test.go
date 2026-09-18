// evidence_grade_test.go hosts EVD-02's grade cap (D-14-01..D-14-08a): a
// closed five-member grade vocabulary, a total order over it, and a
// derivation ladder that computes a CEILING from a Per-Task Verification Map
// row's own evidence cell. A row's DECLARED grade may never exceed its
// derived ceiling -- that is the mechanized half. Register HONESTY (whether
// a row's prose is a truthful account) stays a human reading exactly as
// TestDebtRegistersAreWellFormed's own doc comment already draws that line
// for *-DEBT.md; this file only ever asserts the SHAPE property that the two
// grades which SATISFY a requirement (EXERCISED, MUTATION-KILLED) are
// mechanically decidable from a resolving Go test name, while the three that
// refuse (DEFINED, WIRED, REACHABLE) are refusals nobody can game by
// under-claiming.
//
// This is a SEPARATE authored law from debtRegisterGrades/debtRegisterGradeOrdinal
// (session_test.go, EVD-03/04's *-DEBT.md Grade/Witness columns) even though
// both use the same five-member vocabulary text -- D-14-13's "two authored
// laws, two generated views" ruling: *-VALIDATION.md rows are graded here,
// *-DEBT.md rows are graded there, and neither reuses the other's package
// vars, so each law's own evolution stays independent.
//
// The static test index this ladder resolves identifiers against is
// verification_groundedness_test.go's buildTestIndex/testIndex (D-14-10,
// plan 14-01) -- this file never re-derives a second resolver or shells out
// to `go test -list` itself.
package session_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// ---------------------------------------------------------------------
// (a) The closed five-member grade vocabulary and its total order
// (D-14-01, D-14-02). Declared beside this file's own enforcing tests,
// copying debtRegisterSeverities' shape -- a package-level map, never a
// free-form string.
// ---------------------------------------------------------------------

// validationGradeVocabulary is the closed five-member grade vocabulary a
// Per-Task Verification Map row's Grade cell must belong to. A sixth value
// is never admitted, however plausible it reads.
var validationGradeVocabulary = map[string]bool{
	"DEFINED":         true,
	"WIRED":           true,
	"REACHABLE":       true,
	"EXERCISED":       true,
	"MUTATION-KILLED": true,
}

// validationGradeOrder lists the vocabulary from lowest to highest -- the
// total order D-14-01's cap comparison depends on.
var validationGradeOrder = []string{"DEFINED", "WIRED", "REACHABLE", "EXERCISED", "MUTATION-KILLED"}

// validationGradeOrdinal ranks the vocabulary so "declared <= derived" is a
// numeric comparison, never a string special-case.
var validationGradeOrdinal = map[string]int{
	"DEFINED":         0,
	"WIRED":           1,
	"REACHABLE":       2,
	"EXERCISED":       3,
	"MUTATION-KILLED": 4,
}

// validationSatisfyingGrade is EVD-02's own bar: the two grades that SATISFY
// a requirement are EXERCISED and MUTATION-KILLED; everything below is a
// refusal (D-14-02).
const validationSatisfyingGrade = "EXERCISED"

// validationGradeCellPattern accepts only a bare closed-vocabulary token --
// no "(withdrawn)" suffix here (that escape hatch belongs to the *-DEBT.md
// law only; a VALIDATION row's grade is not a deferred debt claim).
var validationGradeCellPattern = regexp.MustCompile(`^(DEFINED|WIRED|REACHABLE|EXERCISED|MUTATION-KILLED)$`)

// validationGradeCellProblem reports why a raw Grade cell fails to name a
// member of the closed vocabulary, or "" when it does. An empty or
// whitespace-only cell is rejected outright and is never defaulted to any
// vocabulary member (D-14-01's "absence is a failure, not a floor").
func validationGradeCellProblem(cell string) string {
	trimmed := strings.TrimSpace(cell)
	if trimmed == "" {
		return "Grade cell is empty -- absence is never defaulted to a vocabulary member"
	}
	if !validationGradeCellPattern.MatchString(trimmed) {
		return fmt.Sprintf("Grade %q is outside the closed vocabulary %v", trimmed, validationGradeOrder)
	}
	return ""
}

// validationGradeCompare returns <0, 0, >0 as a is below, equal to, or
// above b in the total order. Both arguments must already be validated
// members of the vocabulary.
func validationGradeCompare(a, b string) int {
	return validationGradeOrdinal[a] - validationGradeOrdinal[b]
}

// TestValidationGradeVocabularyIsClosed asserts the vocabulary has exactly
// five members and that a sixth, plausible-looking value is rejected by
// name.
func TestValidationGradeVocabularyIsClosed(t *testing.T) {
	if len(validationGradeVocabulary) != 5 {
		t.Fatalf("validationGradeVocabulary has %d members, want exactly 5: %v", len(validationGradeVocabulary), validationGradeVocabulary)
	}
	for _, want := range validationGradeOrder {
		if !validationGradeVocabulary[want] {
			t.Fatalf("validationGradeVocabulary is missing %q", want)
		}
	}
	for _, bad := range []string{"GREEN", "TRUSTED", "PASSING", ""} {
		if validationGradeVocabulary[bad] {
			t.Fatalf("validationGradeVocabulary wrongly admits %q", bad)
		}
		if problem := validationGradeCellProblem(bad); problem == "" {
			t.Fatalf("validationGradeCellProblem(%q) should name the offending value, got no problem", bad)
		} else if !strings.Contains(problem, bad) && bad != "" {
			t.Fatalf("validationGradeCellProblem(%q) = %q, want it to name the offending value", bad, problem)
		}
	}
}

// TestValidationGradeOrderIsTotal asserts the five grades compare as a
// total order: strict ordering holds across every distinct pair, and two
// equal grades compare neither above nor below one another. The cap's
// comparison is only well defined at the boundary if this holds.
func TestValidationGradeOrderIsTotal(t *testing.T) {
	for i, a := range validationGradeOrder {
		for j, b := range validationGradeOrder {
			cmp := validationGradeCompare(a, b)
			switch {
			case i < j && cmp >= 0:
				t.Fatalf("%s should compare below %s, got comparator %d", a, b, cmp)
			case i > j && cmp <= 0:
				t.Fatalf("%s should compare above %s, got comparator %d", a, b, cmp)
			case i == j && cmp != 0:
				t.Fatalf("%s should compare equal to %s (neither above nor below), got comparator %d", a, b, cmp)
			}
		}
	}
}

// ---------------------------------------------------------------------
// (b) The run record (D-14-03's exercised rung, the fail-closed gate).
// scripts/evidence-run-record.sh is the PRODUCER; this file is the
// CONSUMER. A missing run record is never silently treated as a pass --
// any row whose ceiling would otherwise reach EXERCISED simply cannot,
// which is what makes the gate fail closed for a row declaring EXERCISED+
// with no run record present.
// ---------------------------------------------------------------------

// runRecordEntry is one go test -json TestEvent line, decoded loosely: only
// the three fields this ladder consults.
type runRecordEntry struct {
	Action string `json:"Action"`
	Test   string `json:"Test"`
}

// runRecord is a parsed run record, collapsed to top-level test names only
// (D-14-04: the toolchain's listing -- and this record -- carries top-level
// tests only; a subtest's "Test" field contains a "/" and is deliberately
// excluded here, matching the evidence cell's own parent-test-only
// discipline).
type runRecord struct {
	passed  map[string]bool
	failed  map[string]bool
	skipped map[string]bool
}

// passedNoSkip reports whether name is confirmed EXERCISED: a recorded pass
// for that exact top-level name, with no recorded skip or fail for the same
// name anywhere in the record. A nil record (no run record present at all)
// always reports false -- this is the fail-closed path.
func (r *runRecord) passedNoSkip(name string) bool {
	if r == nil {
		return false
	}
	return r.passed[name] && !r.skipped[name] && !r.failed[name]
}

// parseRunRecord decodes newline-delimited go test -json output. A line
// that is not valid JSON (e.g. a build-failure line on stderr that leaked
// into the file, or a blank line) is skipped, never fatal -- the record is
// best-effort evidence, and an unparseable line simply contributes nothing
// rather than crashing the consumer.
func parseRunRecord(data []byte) *runRecord {
	record := &runRecord{passed: map[string]bool{}, failed: map[string]bool{}, skipped: map[string]bool{}}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry runRecordEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry.Test == "" || strings.Contains(entry.Test, "/") {
			continue
		}
		switch entry.Action {
		case "pass":
			record.passed[entry.Test] = true
		case "fail":
			record.failed[entry.Test] = true
		case "skip":
			record.skipped[entry.Test] = true
		}
	}
	return record
}

// pkgPattern is one (package, -run pattern) pair the run record must cover.
type pkgPattern struct {
	Package string
	Pattern string
}

// evidenceRunRecordScriptRelPath is scripts/evidence-run-record.sh's
// project-relative path -- the sole producer this file ever invokes.
var evidenceRunRecordScriptRelPath = []string{"scripts", "evidence-run-record.sh"}

// evidenceRunRecordTimeout bounds the one process this file spawns
// (TestSourceNeverSpawnsUnboundedProcesses' discipline: a deadline-carrying
// context, independently size-capped stdout/stderr writers, never
// Output()/CombinedOutput()). Generation is scoped to exactly the
// (package, pattern) pairs the caller asks for -- not a whole-module
// `go test ./...` re-run -- so this budget is generous relative to the
// small, targeted batch it actually bounds.
const evidenceRunRecordTimeout = 180 * time.Second

// generateRunRecord invokes scripts/evidence-run-record.sh once, batching
// every requested (package, pattern) pair into a single process spawn, and
// parses whatever it wrote. A script failure (nonzero exit, unreadable
// output) is logged and reported as "no run record" -- fail-closed, never a
// silent pass dressed up as a record.
func generateRunRecord(t testing.TB, pairs []pkgPattern) *runRecord {
	t.Helper()
	if len(pairs) == 0 {
		return parseRunRecord(nil)
	}
	script := testsupport.ProjectPath(evidenceRunRecordScriptRelPath...)
	dir := t.TempDir()
	out := filepath.Join(dir, "run-record.json")
	args := []string{script, out}
	for _, p := range pairs {
		args = append(args, p.Package, p.Pattern)
	}
	ctx, cancel := context.WithTimeout(context.Background(), evidenceRunRecordTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", args...)
	cmd.Dir = testsupport.ProjectPath()
	var stdout, stderr groundednessBoundedWriter
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Logf("evidence-run-record.sh reported an error (a failing test inside the batch is expected and non-fatal here): %v; stderr: %s", err, stderr.bytes())
	}
	data, readErr := os.ReadFile(out)
	if readErr != nil {
		t.Logf("evidence run record file unreadable, treating as absent: %v", readErr)
		return nil
	}
	return parseRunRecord(data)
}

// ---------------------------------------------------------------------
// (c) The derivation ladder (D-14-03) -- deriveCeiling computes the
// ceiling a row's evidence cell earns. Reuses verification_groundedness_test.go's
// testIndex/parseGoTestCommand/resolvePackageNames as the resolution
// primitive, and session_test.go's debtRegisterCallsiteTokenPattern/
// debtRegisterCallSiteWitnessMatches for the zero-call-site rung -- no
// second resolver is invented anywhere in this file.
// ---------------------------------------------------------------------

// compileTimeEvidencePattern recognizes compile-time evidence (13-VALIDATION.md's
// `go build ./...` / `BUILD_OK` row): this rung derives the WIRED ceiling,
// never EXERCISED, because a successful build proves the code compiles, not
// that any assertion ran (D-14-03).
var compileTimeEvidencePattern = regexp.MustCompile(`^go (build|vet)\b|^BUILD_OK$`)

// deriveCeiling computes the ceiling one evidence cell earns, and (for a
// `go test` cell) the set of matched top-level test names the run record is
// consulted against. index and record may be reused across many calls;
// record may be nil, meaning "no run record present" (fail-closed: nothing
// can reach EXERCISED).
func deriveCeiling(index *testIndex, record *runRecord, evidence string) (ceiling string, matchedNames []string) {
	return "DEFINED", nil // RED: intentional stub, restored for GREEN
	evidence = strings.TrimSpace(evidence)
	if evidence == "" {
		return "DEFINED", nil
	}
	if m := debtRegisterCallsiteTokenPattern.FindStringSubmatch(evidence); m != nil {
		pkgPath, symbol, wantStr := m[1], m[2], m[3]
		var want int
		if _, err := fmt.Sscanf(wantStr, "%d", &want); err != nil {
			return "DEFINED", nil
		}
		if ok, _ := debtRegisterCallSiteWitnessMatches(pkgPath, symbol, want); ok {
			// A named production symbol with zero call sites derives WIRED
			// (structurally unreachable, mechanically confirmed); a nonzero
			// count is REACHABLE (invoked in production, not itself
			// exercised by this evidence cell).
			if want == 0 {
				return "WIRED", nil
			}
			return "REACHABLE", nil
		}
		return "DEFINED", nil
	}
	if compileTimeEvidencePattern.MatchString(evidence) {
		return "WIRED", nil
	}
	if strings.HasPrefix(evidence, "go test") {
		parsed, ok := parseGoTestCommand(evidence)
		if !ok {
			return "DEFINED", nil
		}
		names := make(map[string]bool)
		for _, operand := range parsed.Packages {
			set, _ := index.resolvePackageNames(operand)
			for name := range set {
				names[name] = true
			}
		}
		if parsed.Pattern == "" {
			return "WIRED", nil
		}
		segment := strings.SplitN(parsed.Pattern, "/", 2)[0]
		re, err := regexp.Compile(segment)
		if err != nil {
			return "DEFINED", nil
		}
		for name := range names {
			if re.MatchString(name) {
				matchedNames = append(matchedNames, name)
			}
		}
		if len(matchedNames) == 0 {
			return "WIRED", nil
		}
		for _, name := range matchedNames {
			if record.passedNoSkip(name) {
				return "EXERCISED", matchedNames
			}
		}
		return "WIRED", matchedNames
	}
	if verificationCommandPattern.MatchString(evidence) {
		// grep/rg/awk/sed/git/./scripts/go run: a fixture path or a
		// command-line invocation rather than a test name (D-14-03).
		return "REACHABLE", nil
	}
	return "DEFINED", nil
}

// nonInertnessNoneMarkers is the closed set of explicit "no twin declared"
// spellings a Non-inertness cell may carry. Anything else is either a real
// twin identifier or a formatting error -- there is no third, ambiguous
// reading.
var nonInertnessNoneMarkers = map[string]bool{"—": true, "-": true, "n/a": true, "none": true}

// upgradeForNonInertness applies D-14-03's mutation-killed rung: a
// non-empty, non-marker Non-inertness cell must name a DISTINCT top-level
// test that exists in index and is confirmed passed-no-skip by record. If
// it does, and the row's evidence-derived ceiling already reached
// EXERCISED, the ceiling upgrades to MUTATION-KILLED. Naming the same test
// as the primary evidence (not distinct) never upgrades.
func upgradeForNonInertness(index *testIndex, record *runRecord, ceiling string, matchedNames []string, nonInertness string) string {
	if ceiling != "EXERCISED" {
		return ceiling
	}
	cell := strings.TrimSpace(nonInertness)
	if cell == "" || nonInertnessNoneMarkers[strings.ToLower(cell)] {
		return ceiling
	}
	if !isTopLevelTestName(index, cell) {
		return ceiling
	}
	for _, primary := range matchedNames {
		if primary == cell {
			return ceiling // naming the same test as the primary does not upgrade
		}
	}
	if record.passedNoSkip(cell) {
		return "MUTATION-KILLED"
	}
	return ceiling
}

// isTopLevelTestName reports whether name is declared as a top-level
// Test/Fuzz/Benchmark/Example identifier in ANY package the static index
// covers -- the global existence check a Non-inertness twin must pass.
func isTopLevelTestName(index *testIndex, name string) bool {
	for _, names := range index.byImportPath {
		if names[name] {
			return true
		}
	}
	return false
}

// validationRowProblem checks one row's Grade and Non-inertness cells
// against the derived ceiling and returns a non-empty problem string when
// the declared grade exceeds it, is malformed, or (for EXERCISED+) has no
// confirming run record. Returns "" when the row is well formed.
func validationRowProblem(index *testIndex, record *runRecord, taskID, gradeCell, nonInertnessCell, evidenceCell string) string {
	if problem := validationGradeCellProblem(gradeCell); problem != "" {
		return fmt.Sprintf("row %s: %s", taskID, problem)
	}
	declared := strings.TrimSpace(gradeCell)
	ceiling, matchedNames := deriveCeiling(index, record, evidenceCell)
	ceiling = upgradeForNonInertness(index, record, ceiling, matchedNames, nonInertnessCell)
	if validationGradeOrdinal[declared] >= validationGradeOrdinal[validationSatisfyingGrade] && record == nil {
		return fmt.Sprintf("row %s: declares %s (>= %s) with no run record present -- the gate fails closed", taskID, declared, validationSatisfyingGrade)
	}
	if validationGradeCompare(declared, ceiling) > 0 {
		return fmt.Sprintf("row %s: declared grade %s exceeds the ceiling %s derived from its own evidence %q", taskID, declared, ceiling, evidenceCell)
	}
	return ""
}

// ---------------------------------------------------------------------
// (d) Task 1's own subtests over SYNTHETIC rows -- one per ladder rung,
// plus the empty-cell and no-run-record-present cases. The real archived
// corpus is scanned by TestValidationRowGradesAreEarnedOverArchivedCorpus
// (plan 14-09 Task 2), which extends this same file.
// ---------------------------------------------------------------------

func syntheticIndex(t testing.TB) *testIndex {
	t.Helper()
	return &testIndex{
		root: testsupport.ProjectPath(),
		byImportPath: map[string]map[string]bool{
			"github.com/codename-lang/lang/internal/compiler/check": {
				"TestSyntheticPrimaryClaim": true,
				"TestSyntheticTwinClaim":    true,
			},
		},
	}
}

func syntheticRunRecord() *runRecord {
	return &runRecord{
		passed:  map[string]bool{"TestSyntheticPrimaryClaim": true, "TestSyntheticTwinClaim": true},
		failed:  map[string]bool{},
		skipped: map[string]bool{},
	}
}

// TestValidationRowGradesAreEarned exercises every rung of deriveCeiling
// over synthetic rows (D-14-03's <behavior> cases), before Task 2 wires the
// real archived corpus into this same test name via
// TestValidationRowGradesAreEarnedOverArchivedCorpus.
func TestValidationRowGradesAreEarned(t *testing.T) {
	index := syntheticIndex(t)
	record := syntheticRunRecord()

	t.Run("exact resolving test with a passing no-skip run record derives exercised", func(t *testing.T) {
		ceiling, matched := deriveCeiling(index, record, "go test ./internal/compiler/check -run TestSyntheticPrimaryClaim")
		if ceiling != "EXERCISED" {
			t.Fatalf("ceiling = %s, want EXERCISED", ceiling)
		}
		if len(matched) != 1 || matched[0] != "TestSyntheticPrimaryClaim" {
			t.Fatalf("matchedNames = %v, want [TestSyntheticPrimaryClaim]", matched)
		}
	})

	t.Run("unresolved identifier derives wired", func(t *testing.T) {
		ceiling, _ := deriveCeiling(index, record, "go test ./internal/compiler/check -run TestDoesNotExistAnywhere")
		if ceiling != "WIRED" {
			t.Fatalf("ceiling = %s, want WIRED", ceiling)
		}
	})

	t.Run("distinct resolving passing non-inertness twin derives mutation-killed", func(t *testing.T) {
		ceiling, matched := deriveCeiling(index, record, "go test ./internal/compiler/check -run TestSyntheticPrimaryClaim")
		upgraded := upgradeForNonInertness(index, record, ceiling, matched, "TestSyntheticTwinClaim")
		if upgraded != "MUTATION-KILLED" {
			t.Fatalf("upgraded = %s, want MUTATION-KILLED", upgraded)
		}
	})

	t.Run("naming the same test as the primary does not derive mutation-killed", func(t *testing.T) {
		ceiling, matched := deriveCeiling(index, record, "go test ./internal/compiler/check -run TestSyntheticPrimaryClaim")
		upgraded := upgradeForNonInertness(index, record, ceiling, matched, "TestSyntheticPrimaryClaim")
		if upgraded != "EXERCISED" {
			t.Fatalf("upgraded = %s, want EXERCISED (no self-upgrade)", upgraded)
		}
	})

	t.Run("fixture path or command-line invocation derives reachable", func(t *testing.T) {
		ceiling, _ := deriveCeiling(index, record, "grep -c 'S-008' .planning/ROADMAP.md")
		if ceiling != "REACHABLE" {
			t.Fatalf("ceiling = %s, want REACHABLE", ceiling)
		}
	})

	t.Run("compile-time evidence derives wired, not exercised", func(t *testing.T) {
		ceiling, _ := deriveCeiling(index, record, "go build ./...")
		if ceiling != "WIRED" {
			t.Fatalf("ceiling = %s, want WIRED (compile-time evidence never reaches EXERCISED)", ceiling)
		}
		ceiling, _ = deriveCeiling(index, record, "BUILD_OK")
		if ceiling != "WIRED" {
			t.Fatalf("BUILD_OK ceiling = %s, want WIRED", ceiling)
		}
	})

	t.Run("zero call site derives wired", func(t *testing.T) {
		// resolveBlame is real production code (internal/compiler/check/check.go)
		// with zero call sites today (D-13-02b's own finding) -- a genuine,
		// currently-true zero, not a synthetic stand-in.
		ceiling, _ := deriveCeiling(index, record, "callsite:internal/compiler/check.resolveBlame=0")
		if ceiling != "WIRED" {
			t.Fatalf("ceiling = %s, want WIRED", ceiling)
		}
		ceiling, _ = deriveCeiling(index, record, "callsite:internal/compiler/check.resolveBlame=99")
		if ceiling != "DEFINED" {
			t.Fatalf("a wrong claimed call count should fail resolution and fall back to DEFINED, got %s", ceiling)
		}
	})

	t.Run("the defined grade is always permitted as the floor", func(t *testing.T) {
		if problem := validationRowProblem(index, record, "synthetic-floor", "DEFINED", "—", "go test ./internal/compiler/check -run TestDoesNotExistAnywhere"); problem != "" {
			t.Fatalf("DEFINED should always be permitted regardless of ceiling, got problem: %s", problem)
		}
	})

	t.Run("empty or absent grade cell fails, never defaulted", func(t *testing.T) {
		if problem := validationRowProblem(index, record, "synthetic-empty", "", "—", "go test ./internal/compiler/check -run TestSyntheticPrimaryClaim"); problem == "" {
			t.Fatalf("an empty Grade cell must fail, not default to a vocabulary member")
		}
		if problem := validationRowProblem(index, record, "synthetic-blank", "   ", "—", "go test ./internal/compiler/check -run TestSyntheticPrimaryClaim"); problem == "" {
			t.Fatalf("a whitespace-only Grade cell must fail, not default to a vocabulary member")
		}
	})

	t.Run("row declaring exercised or above with no run record present fails", func(t *testing.T) {
		if problem := validationRowProblem(index, nil, "synthetic-no-record", "EXERCISED", "—", "go test ./internal/compiler/check -run TestSyntheticPrimaryClaim"); problem == "" {
			t.Fatalf("declaring EXERCISED with record == nil must fail closed")
		}
		if problem := validationRowProblem(index, record, "synthetic-with-record", "EXERCISED", "—", "go test ./internal/compiler/check -run TestSyntheticPrimaryClaim"); problem != "" {
			t.Fatalf("the same row with a run record showing a pass should succeed, got: %s", problem)
		}
	})

	t.Run("a declared grade above the derived ceiling fails, naming row, declared grade and ceiling", func(t *testing.T) {
		problem := validationRowProblem(index, record, "synthetic-overclaim", "EXERCISED", "—", "grep -c 'S-008' .planning/ROADMAP.md")
		if problem == "" {
			t.Fatalf("declaring EXERCISED over a REACHABLE-ceiling grep cell must fail")
		}
		for _, want := range []string{"synthetic-overclaim", "EXERCISED", "REACHABLE"} {
			if !strings.Contains(problem, want) {
				t.Fatalf("problem %q does not name %q", problem, want)
			}
		}
	})

	t.Run("a declared grade at or below the ceiling passes", func(t *testing.T) {
		if problem := validationRowProblem(index, record, "synthetic-at-ceiling", "REACHABLE", "—", "grep -c 'S-008' .planning/ROADMAP.md"); problem != "" {
			t.Fatalf("declaring exactly the derived ceiling should pass, got: %s", problem)
		}
		if problem := validationRowProblem(index, record, "synthetic-below-ceiling", "WIRED", "—", "grep -c 'S-008' .planning/ROADMAP.md"); problem != "" {
			t.Fatalf("declaring below the derived ceiling should pass, got: %s", problem)
		}
	})
}
