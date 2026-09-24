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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
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
	Action  string `json:"Action"`
	Test    string `json:"Test"`
	Package string `json:"Package"`
	Pattern string `json:"Pattern"`
	Pairs   int    `json:"Pairs"`
}

// runRecordPairCompleteAction and runRecordBatchCompleteAction are the two
// completion-witness sentinel actions scripts/evidence-run-record.sh emits
// (plan 14-11). Both are disjoint from go test's own -json Action
// vocabulary (run/pause/cont/bench/pass/fail/skip/output), so parseRunRecord
// can never confuse a sentinel with a test result.
const (
	runRecordPairCompleteAction  = "record_pair_complete"
	runRecordBatchCompleteAction = "record_batch_complete"
)

// runRecord is a parsed run record, collapsed to top-level test names only
// (D-14-04: the toolchain's listing -- and this record -- carries top-level
// tests only; a subtest's "Test" field contains a "/" and is deliberately
// excluded here, matching the evidence cell's own parent-test-only
// discipline). coveredPairs/batchComplete/requestedPairs are the completion
// witness (D-14-?? / plan 14-11): coveredPairs is keyed the same way
// corpusRunRecord's own "seen" dedup map is keyed (Package, NUL, Pattern),
// and batchComplete is set only when the script's final sentinel was
// observed -- a script that times out or is killed mid-batch never writes
// that line, so batchComplete simply stays false rather than reporting a
// result the batch never finished producing.
type runRecord struct {
	passed           map[string]bool
	failed           map[string]bool
	skipped          map[string]bool
	coveredPairs     map[string]bool
	batchComplete    bool
	requestedPairs   int
	batchCompletions int
	pairCompletions  map[string]int
	elapsed          time.Duration
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
	record := &runRecord{passed: map[string]bool{}, failed: map[string]bool{}, skipped: map[string]bool{}, coveredPairs: map[string]bool{}, pairCompletions: map[string]int{}}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry runRecordEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		switch entry.Action {
		case runRecordPairCompleteAction:
			key := entry.Package + "\x00" + entry.Pattern
			record.coveredPairs[key] = true
			record.pairCompletions[key]++
			continue
		case runRecordBatchCompleteAction:
			record.batchComplete = true
			record.batchCompletions++
			record.requestedPairs = entry.Pairs
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

// complete reports whether record positively demonstrates it covered every
// pair in requested: the script's final batch sentinel was observed AND
// every requested pair has its own per-pair completion sentinel. A
// zero-length requested is the existing legal "nothing to cover" case and
// is vacuously complete regardless of the record -- an empty batch was
// never asked to finish anything. Otherwise a nil record (no run record
// present at all) reports not complete with every pair uncovered -- the
// existing fail-closed direction, preserved. The returned slice names
// exactly the uncovered pairs, in requested's own order.
func (r *runRecord) complete(requested []pkgPattern) (bool, []pkgPattern) {
	if len(requested) == 0 {
		return true, nil
	}
	if r == nil {
		return false, requested
	}
	var missing []pkgPattern
	for _, p := range requested {
		key := p.Package + "\x00" + p.Pattern
		if !r.coveredPairs[key] {
			missing = append(missing, p)
		}
	}
	return r.batchComplete && len(missing) == 0, missing
}

// pkgPattern is one (package, -run pattern) pair the run record must cover.
type pkgPattern struct {
	Package string
	Pattern string
}

// checkedInCorpusRecord is the provenance envelope for the expensive corpus
// execution. The JSONL body remains the producer's unmodified output; this
// envelope binds it to the exact consumer-derived pair list and records the
// revision and bounded external execution that produced it.
type checkedInCorpusRecord struct {
	Schema       string `json:"schema"`
	Revision     string `json:"revision"`
	PairDigest   string `json:"pair_digest_sha256"`
	RecordDigest string `json:"record_digest_sha256"`
	Completed    bool   `json:"completed"`
	ProducedAt   string `json:"produced_at"`
	Elapsed      string `json:"elapsed"`
}

const (
	checkedInCorpusRecordSchema = "phase16-validation-corpus-run-record/v1"
	checkedInCorpusRecordLimit  = 15 * time.Minute
)

var checkedInCorpusRecordRelPath = []string{"testdata", "phase16", "validation-corpus-run-record.jsonl"}
var checkedInCorpusManifestRelPath = []string{"testdata", "phase16", "validation-corpus-run-record.manifest.json"}

func corpusPairBytes(pairs []pkgPattern) ([]byte, error) {
	type pair struct {
		Package string `json:"package"`
		Pattern string `json:"pattern"`
	}
	exported := make([]pair, len(pairs))
	for i, p := range pairs {
		exported[i] = pair{Package: p.Package, Pattern: p.Pattern}
	}
	return json.Marshal(exported)
}

func sha256Hex(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

// checkedInCorpusRunRecord validates an artifact without spawning a child
// process. The record must be complete exactly once, match the currently
// requested pairs byte-for-byte, carry a non-empty revision, and be bound to
// its raw JSONL digest. Any stale or partial artifact refuses to grade.
func checkedInCorpusRunRecord(requested []pkgPattern, recordData, manifestData []byte) (*runRecord, error) {
	if len(requested) == 0 {
		return nil, fmt.Errorf("validation corpus requested pair list is empty")
	}
	var manifest checkedInCorpusRecord
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("decode checked-in validation corpus manifest: %w", err)
	}
	if manifest.Schema != checkedInCorpusRecordSchema || manifest.Revision == "" || !manifest.Completed {
		return nil, fmt.Errorf("checked-in validation corpus manifest lacks schema, revision, or completion witness")
	}
	pairData, err := corpusPairBytes(requested)
	if err != nil {
		return nil, err
	}
	if manifest.PairDigest != sha256Hex(pairData) {
		return nil, fmt.Errorf("checked-in validation corpus pair digest does not match the current requested corpus")
	}
	if manifest.RecordDigest != sha256Hex(recordData) {
		return nil, fmt.Errorf("checked-in validation corpus record digest does not match its JSONL body")
	}
	if _, err := time.Parse(time.RFC3339, manifest.ProducedAt); err != nil {
		return nil, fmt.Errorf("checked-in validation corpus produced_at is invalid: %w", err)
	}
	elapsed, err := time.ParseDuration(manifest.Elapsed)
	if err != nil || elapsed <= 0 || elapsed > checkedInCorpusRecordLimit {
		return nil, fmt.Errorf("checked-in validation corpus elapsed %q is outside the external %v bound", manifest.Elapsed, checkedInCorpusRecordLimit)
	}
	record := parseRunRecord(recordData)
	record.elapsed = elapsed
	complete, missing := record.complete(requested)
	if !complete || record.batchCompletions != 1 || record.requestedPairs != len(requested) {
		return nil, fmt.Errorf("checked-in validation corpus completion witness is incomplete, duplicated, or mismatched (batch=%d pairs=%d want=%d missing=%v)", record.batchCompletions, record.requestedPairs, len(requested), missing)
	}
	for _, p := range requested {
		if record.pairCompletions[p.Package+"\x00"+p.Pattern] != 1 {
			return nil, fmt.Errorf("checked-in validation corpus pair completion witness is not exactly once for %s %s", p.Package, p.Pattern)
		}
	}
	return record, nil
}

func loadCheckedInCorpusRunRecord(requested []pkgPattern) (*runRecord, error) {
	recordData, err := os.ReadFile(testsupport.ProjectPath(checkedInCorpusRecordRelPath...))
	if err != nil {
		return nil, fmt.Errorf("read checked-in validation corpus record: %w", err)
	}
	manifestData, err := os.ReadFile(testsupport.ProjectPath(checkedInCorpusManifestRelPath...))
	if err != nil {
		return nil, fmt.Errorf("read checked-in validation corpus manifest: %w", err)
	}
	return checkedInCorpusRunRecord(requested, recordData, manifestData)
}

// evidenceRunRecordScriptRelPath is scripts/evidence-run-record.sh's
// project-relative path -- the sole producer this file ever invokes.
var evidenceRunRecordScriptRelPath = []string{"scripts", "evidence-run-record.sh"}

// evidenceRunRecordTimeout bounds the one process this file spawns
// (TestSourceNeverSpawnsUnboundedProcesses' discipline: a deadline-carrying
// context, independently size-capped stdout/stderr writers, never
// Output()/CombinedOutput()). Generation is scoped to exactly the
// (package, pattern) pairs the caller asks for -- not a whole-module
// `go test ./...` re-run. corpusRunRecord's consolidatePkgPatterns call
// (below) folds every pair sharing a package into one alternation before
// this budget is spent.
//
// WR-01 (14-REVIEW.md): the previous 300s value's doc comment claimed a
// measured ~136s corpus-wide run, which was stale and false -- commit
// 0dcb460 actually measured corpus-wide runs ranging 269s to 347s, and
// PHASE-14-DEBT.md row D-14-121 recorded a 300.11s run against the
// then-300s ceiling, a near-miss with no real margin. 480s keeps the
// deadline clear of that worst-observed case (347s) while staying under
// Go's 600s default per-package binary ceiling (that whole-binary ceiling
// is NOT overridable from in-process code, only via `go test -timeout`, so
// this budget must stay well under it, not just under some number of its
// own choosing) -- which is exactly why Task 3's cost reduction (anchored,
// resolved-name batching instead of raw markdown pattern text) is
// load-bearing, not optional: it is what keeps the measured elapsed inside
// evidenceRunRecordMarginFraction of this budget.
const evidenceRunRecordTimeout = 480 * time.Second

// evidenceRunRecordMarginFraction is the fraction of evidenceRunRecordTimeout
// a run may consume before corpusRunRecord refuses to grade from it. A run
// that finishes but consumed more than three quarters of its own deadline
// is exactly the false outcome EVD-02 exists to refuse: a result produced by
// a near-timeout run is not distinguishable, from the outside, from a run
// that got lucky this one time and will time out the next. Reported as a
// named budget failure rather than allowed to produce grades.
const evidenceRunRecordMarginFraction = 0.75

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
	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start)
	if runErr != nil {
		t.Logf("evidence-run-record.sh reported an error (a failing test inside the batch is expected and non-fatal here): %v; stderr: %s", runErr, stderr.bytes())
	}
	data, readErr := os.ReadFile(out)
	if readErr != nil {
		t.Logf("evidence run record file unreadable, treating as absent: %v", readErr)
		return nil
	}
	record := parseRunRecord(data)
	record.elapsed = elapsed
	return record
}

// TestRunRecordCarriesACompletionWitness is plan 14-11 Task 1's tracer: one
// thin path wired end to end through every layer this plan's remaining
// tasks expand on -- producer script -> record file -> parser -> in-memory
// record -> assertion. It invokes the REAL script (via generateRunRecord,
// never a hand-rolled second spawn) for two cheap pairs in
// internal/compiler/core, a package that runs in about a second, and
// asserts the returned record positively reports which pairs it covered,
// that the whole batch finished, how long it took, and that a real test
// name in that batch is confirmed passed-no-skip.
func TestRunRecordCarriesACompletionWitness(t *testing.T) {
	pairs := []pkgPattern{
		{Package: "./internal/compiler/core", Pattern: "^TestConventionOverrideNotExpressibleInCore$"},
		{Package: "./internal/compiler/core", Pattern: "^TestParameterModeDecodeIsClosedSet$"},
	}
	record := generateRunRecord(t, pairs)
	if record == nil {
		t.Fatal("generateRunRecord returned a nil record for a real two-pair batch")
	}
	complete, missing := record.complete(pairs)
	if !complete {
		t.Fatalf("record.complete = false, missing pairs: %v", missing)
	}
	if record.elapsed <= 0 {
		t.Fatalf("record.elapsed = %v, want > 0 for a real subprocess batch", record.elapsed)
	}
	t.Logf("measured elapsed for the two-pair tracer batch: %v", record.elapsed)
	if !record.passedNoSkip("TestConventionOverrideNotExpressibleInCore") {
		t.Fatal("expected TestConventionOverrideNotExpressibleInCore to be passedNoSkip in the tracer batch's record")
	}
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
	if assertGoTestsInvocationPattern.MatchString(evidence) {
		return deriveAssertGoTestsCeiling(index, record, evidence)
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
	if verificationCommandPattern.MatchString(evidence) || strings.Contains(evidence, "scripts/") {
		// grep/rg/awk/sed/git/./scripts/go run, or any other shell
		// invocation naming a scripts/ path (e.g. `sh scripts/verify-phaseN.sh`,
		// which does not start with one of verificationCommandPattern's
		// anchored prefixes but is still a command-line invocation, not a
		// test name): a fixture path or a command-line invocation rather
		// than a test name (D-14-03).
		return "REACHABLE", nil
	}
	return "DEFINED", nil
}

// assertGoTestsInvocationPattern recognizes an M001-era
// `[env VAR=val] sh scripts/assert-go-tests.sh [--self-test] <package> <TestName>...`
// evidence cell -- the corpus's OWN "one or more exact Go identifiers" shape
// D-14-03 describes, predating this phase's `go test -run` convention.
var assertGoTestsInvocationPattern = regexp.MustCompile(`assert-go-tests\.sh`)

// deriveAssertGoTestsCeiling resolves an assert-go-tests.sh invocation by
// the SAME exact-match discipline the script itself uses
// (scripts/assert-go-tests.sh:8-45): every named test argument must exist
// verbatim in the named package's test set. Any one unresolved name derives
// WIRED for the whole row -- there is no partial credit, matching D-14-03's
// "resolve each... unresolved => ceiling WIRED."
func deriveAssertGoTestsCeiling(index *testIndex, record *runRecord, evidence string) (ceiling string, matchedNames []string) {
	tokens, err := posixTokenize(evidence)
	if err != nil {
		return "DEFINED", nil
	}
	scriptIdx := -1
	for i, tok := range tokens {
		if strings.Contains(tok, "assert-go-tests.sh") {
			scriptIdx = i
			break
		}
	}
	if scriptIdx == -1 || scriptIdx+1 >= len(tokens) {
		return "DEFINED", nil
	}
	rest := tokens[scriptIdx+1:]
	if len(rest) > 0 && rest[0] == "--self-test" {
		rest = rest[1:]
	}
	if len(rest) < 2 {
		return "DEFINED", nil
	}
	packageNames, resolvedPkg := index.resolvePackageNames(rest[0])
	names := rest[1:]
	if !resolvedPkg {
		return "WIRED", nil
	}
	for _, name := range names {
		if !packageNames[name] {
			return "WIRED", nil
		}
	}
	matchedNames = append(matchedNames, names...)
	for _, name := range names {
		if record.passedNoSkip(name) {
			return "EXERCISED", matchedNames
		}
	}
	return "WIRED", matchedNames
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

// ---------------------------------------------------------------------
// (e) Plan 14-09 Task 2 -- derive over the whole archived corpus. This
// section parses the PRIMARY Per-Task Verification Map table (the first
// header+divider+rows block whose header names "Automated Command",
// immediately following the "## Per-Task Verification Map" heading) in
// each of the fourteen documents. Scope note (recorded, not silent): three
// SECONDARY supplement sub-tables under that same heading --
// 02-VALIDATION.md's two "Post-Gate Supplement" tables and
// 11-VALIDATION.md's "Requirement -> Test Map" table -- are out of this
// migration's scope and retain their original Status column; see this
// plan's own SUMMARY for the recorded reason.
// ---------------------------------------------------------------------

// validationSection extracts the text from "## Per-Task Verification Map"
// up to (not including) the next top-level "## " heading, or to EOF.
func validationSection(text string) (string, int, bool) {
	lines := strings.Split(text, "\n")
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "## Per-Task Verification Map") {
			start = i
			break
		}
	}
	if start == -1 {
		return "", 0, false
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n"), start, true
}

// validationRawRow is one data row of the primary table, with RAW
// (still-escaped) cell text keyed by header name, plus the 1-based line
// number within the whole file.
type validationRawRow struct {
	Line  int
	Cells map[string]string // header name -> raw cell text (framing pipes and outer whitespace trimmed by splitTableRow's caller)
}

// primaryValidationTable locates the FIRST header line inside the
// "## Per-Task Verification Map" section whose cells include "Automated
// Command", and returns its header order, the header's own line number, and
// every following data row up to the next divider-less non-table line or
// the next such header line (a repeated header inside the same block, as
// several M001 documents use for readability, starts a NEW logical
// sub-table and is therefore where this primary table stops, keeping the
// scope to exactly one sub-table as this migration's SUMMARY records).
func primaryValidationTable(text string) (header []string, rows []validationRawRow, ok bool) {
	section, sectionStartLine, found := validationSection(text)
	if !found {
		return nil, nil, false
	}
	lines := strings.Split(section, "\n")
	headerIdx := -1
	for i, line := range lines {
		if !strings.HasPrefix(strings.TrimLeft(line, " \t"), "|") {
			continue
		}
		cells := rawRowCells(line)
		for _, c := range cells {
			if strings.TrimSpace(c) == "Automated Command" {
				headerIdx = i
				header = cells
				break
			}
		}
		if headerIdx != -1 {
			break
		}
	}
	if headerIdx == -1 {
		return nil, nil, false
	}
	// headerIdx+1 is expected to be the divider row; data starts at +2.
	for i := headerIdx + 2; i < len(lines); i++ {
		line := lines[i]
		if !strings.HasPrefix(strings.TrimLeft(line, " \t"), "|") {
			break
		}
		cells := rawRowCells(line)
		isRepeatedHeader := false
		for _, c := range cells {
			if strings.TrimSpace(c) == "Automated Command" {
				isRepeatedHeader = true
			}
		}
		if isRepeatedHeader {
			break
		}
		if len(cells) != len(header) {
			continue // a divider-shaped or malformed line; skip rather than misalign columns
		}
		named := make(map[string]string, len(header))
		for col, name := range header {
			named[strings.TrimSpace(name)] = cells[col]
		}
		rows = append(rows, validationRawRow{Line: sectionStartLine + i + 1, Cells: named})
	}
	return header, rows, true
}

// rawRowCells splits a table row into its RAW (still-escaped) cell texts,
// with the outer framing empty cells (from a line that starts and ends with
// "|") trimmed away, and each cell's own leading/trailing space stripped.
// Built on splitTableRow (D-14-15) so backtick-aware splitting is shared,
// never re-derived.
func rawRowCells(line string) []string {
	cells := splitTableRow(line)
	if len(cells) >= 2 && strings.TrimSpace(cells[0]) == "" {
		cells = cells[1:]
	}
	if len(cells) >= 1 && strings.TrimSpace(cells[len(cells)-1]) == "" {
		cells = cells[:len(cells)-1]
	}
	for i, c := range cells {
		cells[i] = strings.TrimSpace(c)
	}
	return cells
}

// rowIdentifier returns the row's Task ID (10/6-column shape) for
// diagnostics; every primary table in this corpus carries one.
func rowIdentifier(row validationRawRow) string {
	if v, ok := row.Cells["Task ID"]; ok {
		return v
	}
	if v, ok := row.Cells["Req ID"]; ok {
		return v
	}
	return "?"
}

// rowEvidence returns the row's unescaped Automated Command cell -- the
// evidence text deriveCeiling consumes. A cell may legitimately carry two
// backtick spans (D-14-15); this ladder derives from the FIRST span only,
// since every real corpus row cited in this plan's read_first uses exactly
// one command per Automated Command cell for its primary claim.
func rowEvidence(row validationRawRow) string {
	raw, ok := row.Cells["Automated Command"]
	if !ok {
		return ""
	}
	unescaped := unescapeCell(raw)
	spans := extractCodeSpans(unescaped)
	if len(spans) == 0 {
		return strings.TrimSpace(unescaped)
	}
	return strings.TrimSpace(spans[0])
}

// pkgPatternsFor collects every (package, pattern) pair a row's evidence
// requires the run record to cover -- both the `go test -run` shape and the
// assert-go-tests.sh exact-name shape, expressed as a `go test` pattern the
// run-record producer can execute directly.
func pkgPatternsFor(evidence string) []pkgPattern {
	evidence = strings.TrimSpace(evidence)
	if strings.HasPrefix(evidence, "go test") {
		parsed, ok := parseGoTestCommand(evidence)
		if !ok || parsed.Pattern == "" || len(parsed.Packages) == 0 {
			return nil
		}
		var pairs []pkgPattern
		for _, pkg := range parsed.Packages {
			pairs = append(pairs, pkgPattern{Package: pkg, Pattern: parsed.Pattern})
		}
		return pairs
	}
	if assertGoTestsInvocationPattern.MatchString(evidence) {
		tokens, err := posixTokenize(evidence)
		if err != nil {
			return nil
		}
		scriptIdx := -1
		for i, tok := range tokens {
			if strings.Contains(tok, "assert-go-tests.sh") {
				scriptIdx = i
				break
			}
		}
		if scriptIdx == -1 || scriptIdx+1 >= len(tokens) {
			return nil
		}
		rest := tokens[scriptIdx+1:]
		if len(rest) > 0 && rest[0] == "--self-test" {
			rest = rest[1:]
		}
		if len(rest) < 2 {
			return nil
		}
		pattern := "^(" + strings.Join(rest[1:], "|") + ")$"
		return []pkgPattern{{Package: rest[0], Pattern: pattern}}
	}
	return nil
}

// evidenceRunRecordSelfCitedTests is the closed set of top-level test names
// that may never be cited as an evidence cell's own confirming run: today
// this holds only TestValidationRowGradesAreEarnedOverArchivedCorpus, the
// corpus-wide test itself. A row citing it as evidence would ask the run in
// which a claim is being graded to also grade that same claim -- a claim
// may never certify itself. Declared as a set (not a single constant) so a
// future second self-certifying test has a place to land without widening
// the check's shape.
var evidenceRunRecordSelfCitedTests = map[string]bool{
	"TestValidationRowGradesAreEarnedOverArchivedCorpus": true,
}

// patternNames splits an anchored "^(NameA|NameB)$" pattern (resolvedPkgPatterns'
// and pkgPatternsFor's own output shape) back into its individual top-level
// names. A pattern that is not in that anchored shape is returned unchanged
// as its own single-element list -- callers only ever feed this function
// patterns this file itself produced.
func patternNames(pattern string) []string {
	trimmed := strings.TrimPrefix(pattern, "^(")
	trimmed = strings.TrimSuffix(trimmed, ")$")
	if trimmed == pattern {
		return []string{pattern}
	}
	return strings.Split(trimmed, "|")
}

// selfCitationError scans pairs (already resolved to exact top-level names
// via resolvedPkgPatterns) for a member of evidenceRunRecordSelfCitedTests
// and, if found, returns an error naming the citing document, the row, and
// the reason -- self-citation is refused by name, never silently dropped
// from the batch.
func selfCitationError(doc, taskID string, pairs []pkgPattern) error {
	for _, p := range pairs {
		for _, name := range patternNames(p.Pattern) {
			if evidenceRunRecordSelfCitedTests[name] {
				return fmt.Errorf("%s: row %s cites %s as evidence -- a claim may not be graded by the run in which it is being graded", doc, taskID, name)
			}
		}
	}
	return nil
}

// resolvedPkgPatterns is WR-02's fix: it collects every (package, pattern)
// pair a row's evidence requires the run record to cover, built from
// RESOLVED top-level test NAMES rather than the raw pattern text a markdown
// cell wrote. For the `go test -run` shape, each package operand is
// resolved through index.resolvePackageNames and matched against the
// pattern's first path segment using the SAME regexp deriveCeiling itself
// performs (D-14-03) -- so the producer executes exactly the name set the
// consumer will consult, anchored, and a cell that resolves to zero names
// contributes no pair (the existing WIRED-ceiling path, unchanged). The
// assert-go-tests.sh shape already names exact top-level identifiers and is
// already anchored (pkgPatternsFor's own "^(...)$" construction, built from
// the SAME index.resolvePackageNames primitive), so it is routed through
// pkgPatternsFor unchanged rather than re-derived a second way.
func resolvedPkgPatterns(index *testIndex, evidence string) []pkgPattern {
	evidence = strings.TrimSpace(evidence)
	if assertGoTestsInvocationPattern.MatchString(evidence) {
		return pkgPatternsFor(evidence)
	}
	if !strings.HasPrefix(evidence, "go test") {
		return nil
	}
	parsed, ok := parseGoTestCommand(evidence)
	if !ok || parsed.Pattern == "" || len(parsed.Packages) == 0 {
		return nil
	}
	segment := strings.SplitN(parsed.Pattern, "/", 2)[0]
	re, err := regexp.Compile(segment)
	if err != nil {
		return nil
	}
	var pairs []pkgPattern
	for _, operand := range parsed.Packages {
		names, resolved := index.resolvePackageNames(operand)
		if !resolved {
			continue
		}
		var matched []string
		for name := range names {
			if re.MatchString(name) {
				matched = append(matched, name)
			}
		}
		if len(matched) == 0 {
			continue // zero resolved names: the WIRED-ceiling path, unchanged
		}
		sortStrings(matched)
		pairs = append(pairs, pkgPattern{Package: operand, Pattern: "^(" + strings.Join(matched, "|") + ")$"})
	}
	return pairs
}

// validationStatusEligibleForGrading keeps draft/planned validation contracts
// out of completed evidence grading until validate-phase records their actual
// status. Complete is retained for archived milestone artifacts.
func validationStatusEligibleForGrading(status string) bool {
	switch strings.TrimSpace(status) {
	case "validated", "partial", "complete":
		return true
	default:
		return false
	}
}

func validationDocumentStatus(markdown string) string {
	lines := strings.Split(markdown, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return ""
	}
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "---" {
			break
		}
		key, value, ok := strings.Cut(line, ":")
		if ok && strings.TrimSpace(key) == "status" {
			return strings.Trim(strings.TrimSpace(value), `"'`)
		}
	}
	return ""
}

func TestAllPrimaryValidationRowsExcludeDrafts(t *testing.T) {
	for _, status := range []string{"validated", "partial", "complete"} {
		if !validationStatusEligibleForGrading(status) {
			t.Errorf("eligible validation status %q was excluded", status)
		}
	}
	for _, status := range []string{"draft", "planned", "", "unknown"} {
		if validationStatusEligibleForGrading(status) {
			t.Errorf("incomplete validation status %q was included", status)
		}
	}
	if got := validationDocumentStatus("---\nstatus: draft\n---\n# Draft\n"); got != "draft" {
		t.Fatalf("validationDocumentStatus(draft) = %q", got)
	}
	if got := validationDocumentStatus("---\nstatus: \"validated\"\n---\n"); got != "validated" {
		t.Fatalf("validationDocumentStatus(quoted validated) = %q", got)
	}
}

// allPrimaryValidationRows returns rows from validated, partial, and complete
// *-VALIDATION.md primary tables only, sorted by (file, line) for deterministic
// iteration. Draft and planned documents have not collected final evidence.
func allPrimaryValidationRows(t testing.TB) map[string][]validationRawRow {
	t.Helper()
	docs, err := phaseArtifactGlob("*", "*-VALIDATION.md")
	if err != nil {
		t.Fatalf("phaseArtifactGlob *-VALIDATION.md: %v", err)
	}
	result := make(map[string][]validationRawRow)
	for _, doc := range docs {
		data, readErr := os.ReadFile(doc)
		if readErr != nil {
			t.Fatalf("read %s: %v", doc, readErr)
		}
		markdown := string(data)
		if !validationStatusEligibleForGrading(validationDocumentStatus(markdown)) {
			continue
		}
		_, rows, ok := primaryValidationTable(markdown)
		if !ok {
			continue
		}
		result[doc] = rows
	}
	return result
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

// ---------------------------------------------------------------------
// (f) Plan 14-09 Task 2 -- the satisfying bar (D-14-06). Split from
// derivation: derivation is total over ALL fourteen documents (every row
// gets a ceiling, always); the BAR (declared grade must be >= EXERCISED)
// is enforced only for phase fourteen onward, gated by this file-scoped,
// dated exemption map -- copying debtRegisterLandingPhaseExemptions'
// shape exactly. Exemptions are file-scoped and dated, never row-scoped.
// ---------------------------------------------------------------------

// validationGradeBarExemptions names every *-VALIDATION.md file exempted
// from the satisfying bar. It now names exactly thirteen frozen M001/M002
// documents, written before this cap existed -- 14-VALIDATION.md itself
// carried a fourteenth, differently-reasoned entry here from plan 14-09
// through plan 14-11 (its Per-Task Verification Map was still being
// filled in, and D-14-121 recorded a real, measured reason the removal
// attempt could not yet be trusted: the corpus-wide run finished at
// 300.11s, suspiciously close to the then-300s timeout, producing
// spurious ceiling failures across nine unrelated frozen files). Plan
// 14-11 fixed the root cause (raised the timeout to a measured-honest
// 480s, added a 0.75 margin fraction that fails closed on a near-timeout
// run, and anchored the producer/consumer name contract so the batch
// executes exactly what deriveCeiling consults) and plan 14-12 then
// removed the 14-VALIDATION.md entry for good, confirming the bar holds
// for this phase's own artifact with a complete run record measuring a
// 88.63s elapsed against the 480s budget (18.5%, comfortably inside the
// 75% margin) -- see D-14-121's closure and 14-12-SUMMARY.md. A
// file-scoped exemption is reserved for these thirteen frozen documents
// only; any document authored from Phase 14 on that needs a bar
// narrowing uses the row-scoped validationGradeBarRowExemptions instead,
// which requires a named debt-row owner and can never widen to cover a
// whole file the way this map does.
var validationGradeBarExemptions = map[string]string{
	"01-VALIDATION.md": "written before M001, before the grade cap existed (plan 14-09); frozen prior art",
	"02-VALIDATION.md": "written before M001, before the grade cap existed (plan 14-09); frozen prior art",
	"03-VALIDATION.md": "written before M001, before the grade cap existed (plan 14-09); frozen prior art",
	"04-VALIDATION.md": "written before M001, before the grade cap existed (plan 14-09); frozen prior art",
	"05-VALIDATION.md": "written before M001, before the grade cap existed (plan 14-09); frozen prior art",
	"06-VALIDATION.md": "written before M001, before the grade cap existed (plan 14-09); frozen prior art",
	"07-VALIDATION.md": "written before M002, before the grade cap existed (plan 14-09); frozen prior art",
	"08-VALIDATION.md": "written before M002, before the grade cap existed (plan 14-09); frozen prior art",
	"09-VALIDATION.md": "written before M002, before the grade cap existed (plan 14-09); frozen prior art",
	"10-VALIDATION.md": "written before M002, before the grade cap existed (plan 14-09); frozen prior art",
	"11-VALIDATION.md": "written before M002, before the grade cap existed (plan 14-09); frozen prior art",
	"12-VALIDATION.md": "written before M002, before the grade cap existed (plan 14-09); frozen prior art",
	"13-VALIDATION.md": "written before M002, before the grade cap existed (plan 14-09); frozen prior art",
}

// validationGradeBarRowExemptions is the honest, row-scoped escape hatch
// that replaces a whole-file exemption for any document authored from
// Phase 14 on: a closed map from "<file>:<Task ID>" to the D-14-NNN debt
// row that owns the narrowing. Every entry's cited row MUST carry a
// P<NN> or CLOSED(<sha>) Landing phase cell -- never UNOWNED(...) --
// mechanically checked by TestValidationGradeBarRowExemptionsAreOwned. A
// file-scoped exemption (validationGradeBarExemptions) is reserved for
// the thirteen frozen M001/M002 documents above; this map is where every
// later, honest narrowing lives instead.
//
// This plan's own read_first expected this map to ship EMPTY: the
// planner's assumption was that every pre-existing row in
// 14-VALIDATION.md's own Per-Task Verification Map would already clear
// the satisfying bar once the file-scoped exemption was removed. Running
// the real corpus test against the real archived rows falsified that
// assumption for five of them (see PHASE-14-DEBT.md D-14-123..D-14-127):
// two (14-01-T2, 14-10-T3) declare WIRED even though a complete run
// record now derives EXERCISED for their evidence, and three (14-02-T1,
// 14-02-T2, 14-03-T1) are structurally capped below EXERCISED by their
// evidence cell's own shape (a `go run` invocation, or a `go test`
// invocation with no `-run` pattern) and can never reach EXERCISED
// without rewriting the cell -- which this plan's own prohibition on
// rewriting an archived 14-01..14-10 row to force the bar to pass
// forbids. Each of the five is narrowed here instead, exactly per this
// plan's own Task 1 contingency text ("for an archived 14-01..14-10 row
// you may not rewrite, route it through Task 2's row-scoped narrowing
// with a debt witness").
var validationGradeBarRowExemptions = map[string]string{
	"14-VALIDATION.md:14-01-T2": "D-14-123",
	"14-VALIDATION.md:14-02-T1": "D-14-124",
	"14-VALIDATION.md:14-02-T2": "D-14-125",
	"14-VALIDATION.md:14-03-T1": "D-14-126",
	"14-VALIDATION.md:14-10-T3": "D-14-127",
}

// phase14DebtLandingPhases reads PHASE-14-DEBT.md's own Items table and
// returns a map from debt-row ID to its (already-trimmed) Landing phase
// cell -- the one lookup both TestValidationGradeBarRowExemptionsAreOwned
// and rowExemptionProblems use to resolve a validationGradeBarRowExemptions
// entry's cited debt row, built on parseDebtRegisterTable (session_test.go)
// rather than a second, ad hoc parser.
func phase14DebtLandingPhases(t testing.TB) map[string]string {
	t.Helper()
	registers, err := phaseArtifactGlob("14-evidence-instrument-and-honest-scoping", "PHASE-14-DEBT.md")
	if err != nil || len(registers) != 1 {
		t.Fatalf("expected exactly one PHASE-14-DEBT.md, found %d (err=%v)", len(registers), err)
	}
	data, err := os.ReadFile(registers[0])
	if err != nil {
		t.Fatalf("read %s: %v", registers[0], err)
	}
	columns, rows, parseErr := parseDebtRegisterTable(filepath.Base(registers[0]), string(data))
	if parseErr != nil {
		t.Fatalf("parse %s: %v", registers[0], parseErr)
	}
	result := make(map[string]string, len(rows))
	for _, row := range rows {
		result[row[columns["ID"]]] = strings.TrimSpace(row[columns["Landing phase"]])
	}
	return result
}

// rowExemptionProblems is the ONE predicate shared by
// validationGradeBarRowExemptions' production use and
// TestValidationGradeBarRowExemptionsAreOwned's seeded-fault proof (never
// a second copy), following suppressionProblems' precedent
// (witness_registry_test.go). register is phase14DebtLandingPhases' own
// output shape (debt-row ID -> trimmed Landing phase cell). Returns one
// problem string per faulty entry -- a cited row that does not exist, or
// one whose Landing phase cell is UNOWNED(...) -- naming the entry's key
// and the debt row every time, never just "failed".
func rowExemptionProblems(entries map[string]string, register map[string]string) []string {
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sortStrings(keys)
	var problems []string
	for _, key := range keys {
		debtID := entries[key]
		cell, exists := register[debtID]
		if !exists {
			problems = append(problems, fmt.Sprintf("validationGradeBarRowExemptions[%q] cites %s, which does not exist in PHASE-14-DEBT.md", key, debtID))
			continue
		}
		if debtRegisterUnownedPattern.MatchString(cell) {
			problems = append(problems, fmt.Sprintf("validationGradeBarRowExemptions[%q] cites %s, whose Landing phase cell %q is UNOWNED -- a row-scoped narrowing must name a real owner", key, debtID, cell))
		}
	}
	return problems
}

// TestValidationGradeBarRowExemptionsAreOwned proves rowExemptionProblems
// is not inert (three seeded faults over synthetic entries against the
// real, live PHASE-14-DEBT.md register) and then asserts the shipped
// validationGradeBarRowExemptions carries zero problems.
func TestValidationGradeBarRowExemptionsAreOwned(t *testing.T) {
	register := phase14DebtLandingPhases(t)

	t.Run("an entry citing a nonexistent row is refused, naming the entry", func(t *testing.T) {
		entries := map[string]string{"synthetic-doc.md:T1": "D-14-999999-does-not-exist"}
		problems := rowExemptionProblems(entries, register)
		if len(problems) != 1 {
			t.Fatalf("expected exactly one problem, got %d: %v", len(problems), problems)
		}
		for _, want := range []string{"synthetic-doc.md:T1", "D-14-999999-does-not-exist"} {
			if !strings.Contains(problems[0], want) {
				t.Fatalf("problem %q does not name %q", problems[0], want)
			}
		}
	})

	t.Run("an entry citing a real row whose cell is UNOWNED is refused, naming the row", func(t *testing.T) {
		// D-14-45 is a real, currently UNOWNED(...) row in the live register.
		if cell := register["D-14-45"]; !debtRegisterUnownedPattern.MatchString(cell) {
			t.Fatalf("fixture assumption broken: D-14-45's Landing phase is %q, no longer UNOWNED(...)", cell)
		}
		entries := map[string]string{"synthetic-doc.md:T2": "D-14-45"}
		problems := rowExemptionProblems(entries, register)
		if len(problems) != 1 {
			t.Fatalf("expected exactly one problem, got %d: %v", len(problems), problems)
		}
		if !strings.Contains(problems[0], "D-14-45") {
			t.Fatalf("problem %q does not name D-14-45", problems[0])
		}
	})

	t.Run("an entry citing a real row with a P<NN> cell passes", func(t *testing.T) {
		// D-14-48's Landing phase is P14.
		if cell := register["D-14-48"]; !debtRegisterPhaseIDPattern.MatchString(cell) {
			t.Fatalf("fixture assumption broken: D-14-48's Landing phase is %q, not a P<NN> cell", cell)
		}
		entries := map[string]string{"synthetic-doc.md:T3": "D-14-48"}
		if problems := rowExemptionProblems(entries, register); len(problems) != 0 {
			t.Fatalf("expected zero problems, got %v", problems)
		}
	})

	if problems := rowExemptionProblems(validationGradeBarRowExemptions, register); len(problems) > 0 {
		t.Fatalf("shipped validationGradeBarRowExemptions has %d unowned/nonexistent entries:\n%s", len(problems), strings.Join(problems, "\n"))
	}
}

// phase14ValidationRowFloor is the non-vacuity floor plan 14-10 established
// (31 real rows, one per task across plans 14-01..14-10) -- the count
// 14-VALIDATION.md's own Per-Task Verification Map must never fall below,
// so TestValidationGradeBarAppliesToPhase14 cannot be satisfied by emptying
// the table (D-14-121's permanent guard, second half).
const phase14ValidationRowFloor = 31

// TestValidationGradeBarAppliesToPhase14 is D-14-121's permanent guard: it
// fails, naming the file, if 14-VALIDATION.md is ever re-added to
// validationGradeBarExemptions, AND it fails, naming the count, if
// 14-VALIDATION.md's own Per-Task Verification Map ever shrinks below
// phase14ValidationRowFloor -- both halves are needed, because either one
// alone reopens the same finding D-14-121 records: re-adding the
// exemption directly un-checks the bar, and emptying the table reaches
// the same outcome indirectly (there would be nothing left for the bar to
// judge).
func TestValidationGradeBarAppliesToPhase14(t *testing.T) {
	if _, exempt := validationGradeBarExemptions["14-VALIDATION.md"]; exempt {
		t.Fatal("14-VALIDATION.md must never re-appear in validationGradeBarExemptions -- D-14-121's finding is permanently closed, not merely fixed once")
	}
	byDoc := allPrimaryValidationRows(t)
	var found bool
	for doc, rows := range byDoc {
		if filepath.Base(doc) != "14-VALIDATION.md" {
			continue
		}
		found = true
		if len(rows) < phase14ValidationRowFloor {
			t.Fatalf("14-VALIDATION.md has %d rows, below the %d floor plan 14-10 established -- the satisfying bar must not be satisfiable by emptying the table", len(rows), phase14ValidationRowFloor)
		}
	}
	if !found {
		t.Fatal("14-VALIDATION.md's primary table was not discovered at all")
	}
}

// ---------------------------------------------------------------------
// (g) The real archived-corpus scan (D-14-06's derivation-is-total
// requirement, wired against the now-migrated documents). Extends
// TestValidationRowGradesAreEarned's own name via substring match
// (-run 'TestValidationRowGradesAreEarned' matches both), following the
// same "keep new classifiers standalone, wire in the final task" shape
// plan 14-06 used.
// ---------------------------------------------------------------------

// evidenceRunRecordOnce memoizes ONE corpus-wide run-record generation per
// test binary process: `go test ./...` already builds and runs every
// package exactly once, so regenerating per subtest/per-run-of-this-test
// would multiply an already-measured-expensive cost for no additional
// information. Measured cost (this plan's own SUMMARY): several minutes
// for the full ~200-pair corpus, not the "milliseconds" this phase's
// threat model first assumed for a run-record read -- corrected here.
var (
	evidenceRunRecordOnce   sync.Once
	evidenceRunRecordCached *runRecord
	evidenceRunRecordFatal  string
)

// runRecordAdmissible is the ONE shared law between corpusRunRecord's fatal
// gate and its own non-inertness proof
// (TestRunRecordCompletenessGuardIsNotInert) -- never a second copy of the
// predicate. A record is admissible (nil error) only when it positively
// demonstrates it covered every requested pair AND its measured elapsed
// does not exceed evidenceRunRecordMarginFraction of
// evidenceRunRecordTimeout. Both failure messages name the measured
// numbers, never just "failed", so a CI log always carries the evidence a
// human needs without re-running anything.
func runRecordAdmissible(record *runRecord, requested []pkgPattern) error {
	complete, missing := record.complete(requested)
	if !complete {
		var elapsed time.Duration
		if record != nil {
			elapsed = record.elapsed
		}
		names := make([]string, 0, len(missing))
		for _, p := range missing {
			names = append(names, p.Package+" "+p.Pattern)
		}
		return fmt.Errorf("evidence run record is incomplete -- uncovered pairs %v (measured elapsed %v)", names, elapsed)
	}
	budget := time.Duration(float64(evidenceRunRecordTimeout) * evidenceRunRecordMarginFraction)
	if record.elapsed > budget {
		fraction := float64(record.elapsed) / float64(evidenceRunRecordTimeout)
		return fmt.Errorf("evidence run record consumed %.1f%% of its %v budget (measured elapsed %v, margin %.0f%%) -- a near-timeout run may not produce grades", fraction*100, evidenceRunRecordTimeout, record.elapsed, evidenceRunRecordMarginFraction*100)
	}
	return nil
}

func corpusRunRecord(t testing.TB, index *testIndex, byDoc map[string][]validationRawRow) *runRecord {
	t.Helper()
	evidenceRunRecordOnce.Do(func() {
		requested, err := requestedCorpusPairs(index, byDoc)
		if err != nil {
			evidenceRunRecordFatal = err.Error()
			return
		}
		record, err := loadCheckedInCorpusRunRecord(requested)
		if err != nil {
			evidenceRunRecordFatal = err.Error()
			return
		}
		t.Logf("checked-in evidence run record elapsed: %v (external bound %v)", record.elapsed, checkedInCorpusRecordLimit)
		evidenceRunRecordCached = record
	})
	if evidenceRunRecordFatal != "" {
		t.Fatalf("%s", evidenceRunRecordFatal)
	}
	return evidenceRunRecordCached
}

// requestedCorpusPairs is the single deterministic source for both the
// validation consumer and the external record producer.  Keeping the pair
// derivation here prevents a checked-in evidence record from claiming a
// guessed or stale subset of the archived validation corpus.
func requestedCorpusPairs(index *testIndex, byDoc map[string][]validationRawRow) ([]pkgPattern, error) {
	var pairs []pkgPattern
	seen := map[string]bool{}
	docs := make([]string, 0, len(byDoc))
	for doc := range byDoc {
		docs = append(docs, doc)
	}
	sort.Strings(docs)
	for _, doc := range docs {
		rows := byDoc[doc]
		for _, row := range rows {
			rowPairs := resolvedPkgPatterns(index, rowEvidence(row))
			if err := selfCitationError(doc, rowIdentifier(row), rowPairs); err != nil {
				return nil, err
			}
			for _, p := range rowPairs {
				key := p.Package + "\x00" + p.Pattern
				if !seen[key] {
					seen[key] = true
					pairs = append(pairs, p)
				}
			}
		}
	}
	return consolidatePkgPatterns(pairs), nil
}

// TestExportValidationCorpusPairs is an intentionally narrow producer seam.
// When AI_LANG_EVIDENCE_PAIR_OUTPUT is set it writes the exact pair list the
// grade consumer requests, using requestedCorpusPairs rather than a second
// parser.  Normal test runs leave no file behind.
func TestExportValidationCorpusPairs(t *testing.T) {
	path := os.Getenv("AI_LANG_EVIDENCE_PAIR_OUTPUT")
	if path == "" {
		t.Skip("producer seam is invoked only by the external record command; see probe:TestValidationCorpusPairExportMatchesConsumer")
	}
	pairs, err := requestedCorpusPairs(buildTestIndex(t), allPrimaryValidationRows(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := writeCorpusPairExport(path, pairs); err != nil {
		t.Fatal(err)
	}
}

func writeCorpusPairExport(path string, pairs []pkgPattern) error {
	data, err := corpusPairBytes(pairs)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func TestValidationCorpusPairExportMatchesConsumer(t *testing.T) {
	pairs, err := requestedCorpusPairs(buildTestIndex(t), allPrimaryValidationRows(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) == 0 {
		t.Fatal("requested corpus pair export is empty")
	}
	for _, pair := range pairs {
		if pair.Package == "" || pair.Pattern == "" {
			t.Fatalf("invalid requested corpus pair: %+v", pair)
		}
	}
	path := filepath.Join(t.TempDir(), "pairs.json")
	if err := writeCorpusPairExport(path, pairs); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want, err := corpusPairBytes(pairs)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatal("producer pair export differs from the validation consumer's exact requested list")
	}
	again, err := requestedCorpusPairs(buildTestIndex(t), allPrimaryValidationRows(t))
	if err != nil {
		t.Fatal(err)
	}
	againBytes, err := corpusPairBytes(again)
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != string(againBytes) {
		t.Fatal("consumer pair export is nondeterministic across fresh corpus maps")
	}
}

// consolidatePkgPatterns merges every pair sharing the same Package operand
// into ONE pair whose Pattern is the top-level alternation of every
// distinct pattern cited for that package. `go test`'s default per-package
// timeout (10m, unaffected by any flag this file controls) bounds the
// WHOLE test binary run for internal/compiler/session, not each subtest
// individually -- so the corpus-wide run record must spend its cost on
// actual test EXECUTION time, not on ~200 redundant process-start+compile
// overheads for a package this consolidation could have covered in one
// spawn. Measured effect: cuts the run-record producer's subprocess count
// from one per (package, pattern) pair (up to ~208) to one per distinct
// package (~20).
//
// Anchoring constraint (WR-02, plan 14-10's finding, restated so it is not
// re-litigated): each already-anchored per-cell pattern (resolvedPkgPatterns'
// "^(...)$" output) is wrapped in its own extra parens and joined by "|" --
// the anchors stay INSIDE each branch. Wrapping the WHOLE joined alternation
// in one outer "^(...)$" was tried and breaks classifyPerBranchGroundedness's
// (R2b) per-branch splitter, which expects to split on top-level "|" and
// re-derive each branch's own groundedness independently. Per-branch
// anchoring is therefore load-bearing, not cosmetic.
func consolidatePkgPatterns(pairs []pkgPattern) []pkgPattern {
	order := make([]string, 0, len(pairs))
	byPackage := make(map[string][]string, len(pairs))
	for _, p := range pairs {
		if _, ok := byPackage[p.Package]; !ok {
			order = append(order, p.Package)
		}
		byPackage[p.Package] = append(byPackage[p.Package], p.Pattern)
	}
	consolidated := make([]pkgPattern, 0, len(order))
	for _, pkg := range order {
		consolidated = append(consolidated, pkgPattern{
			Package: pkg,
			Pattern: "(" + strings.Join(byPackage[pkg], ")|(") + ")",
		})
	}
	return consolidated
}

// syntheticCompleteRunRecord builds a runRecord that reports complete for
// exactly requested, with the given elapsed -- the unmodified control this
// file's seeded-fault subtests each mutate one way.
func syntheticCompleteRunRecord(requested []pkgPattern, elapsed time.Duration) *runRecord {
	r := &runRecord{
		passed:         map[string]bool{},
		failed:         map[string]bool{},
		skipped:        map[string]bool{},
		coveredPairs:   map[string]bool{},
		batchComplete:  true,
		requestedPairs: len(requested),
		elapsed:        elapsed,
	}
	for _, p := range requested {
		r.coveredPairs[p.Package+"\x00"+p.Pattern] = true
	}
	return r
}

// TestRunRecordCompletenessGuardIsNotInert seeds one fault per
// mechanizable kind into a synthetic record and asserts runRecordAdmissible
// refuses each, plus the unmodified control that must accept (D-14-11's own
// non-inertness discipline, mirrored from TestValidationGradeCapIsNotInert).
func TestRunRecordCompletenessGuardIsNotInert(t *testing.T) {
	requested := []pkgPattern{{Package: "./internal/compiler/core", Pattern: "^TestSeeded$"}}

	t.Run("batch sentinel absent fails naming incompleteness", func(t *testing.T) {
		r := syntheticCompleteRunRecord(requested, time.Second)
		r.batchComplete = false
		err := runRecordAdmissible(r, requested)
		if err == nil {
			t.Fatal("expected an error when the batch-completion sentinel is absent")
		}
		if !strings.Contains(err.Error(), "incomplete") {
			t.Fatalf("error %q does not name incompleteness", err.Error())
		}
	})

	t.Run("one missing pair sentinel fails naming exactly that pair", func(t *testing.T) {
		r := syntheticCompleteRunRecord(requested, time.Second)
		delete(r.coveredPairs, requested[0].Package+"\x00"+requested[0].Pattern)
		err := runRecordAdmissible(r, requested)
		if err == nil {
			t.Fatal("expected an error when a requested pair's completion sentinel is missing")
		}
		if !strings.Contains(err.Error(), requested[0].Package) || !strings.Contains(err.Error(), requested[0].Pattern) {
			t.Fatalf("error %q does not name the missing pair %v", err.Error(), requested[0])
		}
	})

	t.Run("elapsed just over the margin fails naming the fraction", func(t *testing.T) {
		overMargin := time.Duration(float64(evidenceRunRecordTimeout)*evidenceRunRecordMarginFraction) + time.Second
		r := syntheticCompleteRunRecord(requested, overMargin)
		err := runRecordAdmissible(r, requested)
		if err == nil {
			t.Fatal("expected an error when elapsed exceeds the margin")
		}
		if !strings.Contains(err.Error(), "%") {
			t.Fatalf("error %q does not name the consumed fraction", err.Error())
		}
	})

	t.Run("the unmodified control stays green", func(t *testing.T) {
		r := syntheticCompleteRunRecord(requested, 10*time.Second)
		if err := runRecordAdmissible(r, requested); err != nil {
			t.Fatalf("unmodified complete record inside margin should be admissible, got: %v", err)
		}
	})
}

func TestCheckedInCorpusRecordRejectsTamperingAndVacuity(t *testing.T) {
	requested := []pkgPattern{{Package: "./internal/compiler/core", Pattern: "^TestSeeded$"}}
	recordData := []byte("{\"Action\":\"pass\",\"Test\":\"TestSeeded\"}\n{\"Action\":\"record_pair_complete\",\"Package\":\"./internal/compiler/core\",\"Pattern\":\"^TestSeeded$\"}\n{\"Action\":\"record_batch_complete\",\"Pairs\":1}\n")
	pairData, err := corpusPairBytes(requested)
	if err != nil {
		t.Fatal(err)
	}
	manifest := checkedInCorpusRecord{Schema: checkedInCorpusRecordSchema, Revision: "9a1de4c", PairDigest: sha256Hex(pairData), RecordDigest: sha256Hex(recordData), Completed: true, ProducedAt: "2026-09-20T23:00:00Z", Elapsed: "1s"}
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := checkedInCorpusRunRecord(requested, recordData, manifestData); err != nil {
		t.Fatalf("valid checked-in record rejected: %v", err)
	}
	t.Run("changed requested pair bytes refuse the old pair digest", func(t *testing.T) {
		changed := []pkgPattern{{Package: "./internal/compiler/core", Pattern: "^TestChanged$"}}
		if _, err := checkedInCorpusRunRecord(changed, recordData, manifestData); err == nil || !strings.Contains(err.Error(), "pair digest") {
			t.Fatalf("changed pair bytes should be refused by the bound pair digest, got: %v", err)
		}
	})
	t.Run("changed run-record bytes refuse the old record digest", func(t *testing.T) {
		changed := append(append([]byte(nil), recordData...), []byte("{\"Action\":\"output\",\"Output\":\"tampered\"}\n")...)
		if _, err := checkedInCorpusRunRecord(requested, changed, manifestData); err == nil || !strings.Contains(err.Error(), "record digest") {
			t.Fatalf("changed run-record bytes should be refused by the bound record digest, got: %v", err)
		}
	})
	t.Run("missing persisted pair completion refuses a re-digested body", func(t *testing.T) {
		pairWitness := []byte("{\"Action\":\"record_pair_complete\",\"Package\":\"./internal/compiler/core\",\"Pattern\":\"^TestSeeded$\"}\n")
		missingPair := []byte(strings.Replace(string(recordData), string(pairWitness), "", 1))
		mutated := manifest
		mutated.RecordDigest = sha256Hex(missingPair)
		data, err := json.Marshal(mutated)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := checkedInCorpusRunRecord(requested, missingPair, data); err == nil || !strings.Contains(err.Error(), "missing=") || !strings.Contains(err.Error(), requested[0].Pattern) {
			t.Fatalf("missing persisted pair completion should be refused after its new body digest is bound, got: %v", err)
		}
	})
	t.Run("missing persisted batch completion refuses a re-digested body", func(t *testing.T) {
		batchWitness := []byte("{\"Action\":\"record_batch_complete\",\"Pairs\":1}\n")
		missingBatch := []byte(strings.Replace(string(recordData), string(batchWitness), "", 1))
		mutated := manifest
		mutated.RecordDigest = sha256Hex(missingBatch)
		data, err := json.Marshal(mutated)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := checkedInCorpusRunRecord(requested, missingBatch, data); err == nil || !strings.Contains(err.Error(), "completion witness") {
			t.Fatalf("missing persisted batch completion should be refused after its new body digest is bound, got: %v", err)
		}
	})
	for name, mutate := range map[string]func(*checkedInCorpusRecord){
		"missing revision":    func(m *checkedInCorpusRecord) { m.Revision = "" },
		"pair digest drift":   func(m *checkedInCorpusRecord) { m.PairDigest = "00" },
		"record digest drift": func(m *checkedInCorpusRecord) { m.RecordDigest = "00" },
		"missing completion":  func(m *checkedInCorpusRecord) { m.Completed = false },
	} {
		t.Run(name, func(t *testing.T) {
			mutated := manifest
			mutate(&mutated)
			data, err := json.Marshal(mutated)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := checkedInCorpusRunRecord(requested, recordData, data); err == nil {
				t.Fatal("tampered checked-in record was accepted")
			}
		})
	}
}

// evidenceRunRecordAnchoredPatternShape recognizes resolvedPkgPatterns' and
// pkgPatternsFor's own "^(...)$" output shape.
var evidenceRunRecordAnchoredPatternShape = regexp.MustCompile(`^\^\(.*\)\$$`)

// TestEvidencePatternsResolveToAnchoredNames is plan 14-11 Task 3's proof
// that the producer executes exactly what the consumer consults (WR-02):
// (a) over the whole live corpus, every operand resolvedPkgPatterns would
// batch is anchored and every branch is an exact top-level test name
// present in the static index; (b) a synthetic two-cell fixture whose
// patterns are prefix-related consolidates to exactly the two intended
// names, sweeping in no unrelated third name; (c) a synthetic cell citing
// the corpus-wide test itself produces the self-citation refusal -- this
// subtest is the seeded fault that proves the refusal non-inert.
func TestEvidencePatternsResolveToAnchoredNames(t *testing.T) {
	t.Run("every live corpus operand is anchored, and every go-test-shaped branch is an exact top-level name", func(t *testing.T) {
		index := buildTestIndex(t)
		byDoc := allPrimaryValidationRows(t)
		var totalPairs, checkedGoTestNames int
		for _, rows := range byDoc {
			for _, row := range rows {
				evidence := rowEvidence(row)
				pairs := resolvedPkgPatterns(index, evidence)
				for _, p := range pairs {
					totalPairs++
					if !evidenceRunRecordAnchoredPatternShape.MatchString(p.Pattern) {
						t.Fatalf("pair %+v (evidence %q) is not anchored in the ^(...)$ shape", p, evidence)
					}
				}
				// Existence-against-the-index is checked only for the `go
				// test -run` shape: resolvedPkgPatterns derives THOSE
				// branches FROM the static index itself, so every branch it
				// produces is an exact top-level name by construction. The
				// assert-go-tests.sh shape is delegated unchanged to
				// pkgPatternsFor (D-14-11's own read_first constraint), and
				// -- like the corpus's own already-documented dead citation
				// (D-14-51, 04-VALIDATION.md:74, unrelated to this plan) --
				// may legitimately name a since-retired identifier; that is
				// capped at WIRED by deriveAssertGoTestsCeiling's own
				// existence check, not re-verified here.
				if strings.HasPrefix(evidence, "go test") {
					for _, p := range pairs {
						for _, name := range patternNames(p.Pattern) {
							checkedGoTestNames++
							if !isTopLevelTestName(index, name) {
								t.Fatalf("pair %+v (evidence %q) names %q, which is not an exact top-level test name in the static index", p, evidence, name)
							}
						}
					}
				}
			}
		}
		if totalPairs == 0 {
			t.Fatal("resolvedPkgPatterns produced zero pairs over the live corpus -- resolution has gone inert")
		}
		if checkedGoTestNames == 0 {
			t.Fatal("no `go test -run` shaped evidence cell was found to check name resolution against -- the live corpus fixture has gone stale")
		}
	})

	t.Run("prefix-related patterns consolidate to exactly the two intended names", func(t *testing.T) {
		index := &testIndex{
			root: testsupport.ProjectPath(),
			byImportPath: map[string]map[string]bool{
				"github.com/codename-lang/lang/internal/compiler/check": {
					"TestAlpha":         true,
					"TestAlphaExtended": true,
					"TestOmega":         true, // unrelated: must never be swept in
				},
			},
		}
		pairsA := resolvedPkgPatterns(index, "go test ./internal/compiler/check -run TestAlpha")
		pairsB := resolvedPkgPatterns(index, "go test ./internal/compiler/check -run TestAlphaExtended")
		all := append(append([]pkgPattern{}, pairsA...), pairsB...)
		consolidated := consolidatePkgPatterns(all)
		if len(consolidated) != 1 {
			t.Fatalf("expected exactly one consolidated pair (single package), got %d: %v", len(consolidated), consolidated)
		}
		// consolidatePkgPatterns' own output ("(branch1)|(branch2)") is not
		// itself in the single-level "^(...)$" shape patternNames parses --
		// it is the OUTER join of already-anchored branches (the
		// per-branch-anchoring constraint this task's own doc comment
		// records). Compile it and check what it actually matches, which is
		// what the run-record script itself will do with this exact string.
		re, err := regexp.Compile(consolidated[0].Pattern)
		if err != nil {
			t.Fatalf("consolidated pattern %q does not compile: %v", consolidated[0].Pattern, err)
		}
		for _, want := range []string{"TestAlpha", "TestAlphaExtended"} {
			if !re.MatchString(want) {
				t.Fatalf("consolidated pattern %q does not match intended name %q", consolidated[0].Pattern, want)
			}
		}
		if re.MatchString("TestOmega") {
			t.Fatalf("consolidated pattern %q wrongly matches the unrelated TestOmega", consolidated[0].Pattern)
		}
	})

	t.Run("a cell citing the corpus-wide test itself is refused as a self-citation", func(t *testing.T) {
		index := buildTestIndex(t)
		evidence := "go test ./internal/compiler/session -run TestValidationRowGradesAreEarnedOverArchivedCorpus$"
		pairs := resolvedPkgPatterns(index, evidence)
		if len(pairs) == 0 {
			t.Fatal("resolvedPkgPatterns produced zero pairs for a real production test name -- fixture has gone stale")
		}
		err := selfCitationError("synthetic-doc.md", "seed-self-citation", pairs)
		if err == nil {
			t.Fatal("expected selfCitationError to refuse a cell citing TestValidationRowGradesAreEarnedOverArchivedCorpus")
		}
		for _, want := range []string{"synthetic-doc.md", "seed-self-citation", "TestValidationRowGradesAreEarnedOverArchivedCorpus"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error %q does not name %q", err.Error(), want)
			}
		}
	})
}

// TestValidationRowGradesAreEarnedOverArchivedCorpus is D-14-06's real
// scan: every primary Per-Task Verification Map row across the fourteen
// documents must derive a ceiling (derivation is total -- a row this
// cannot classify is a hard failure, the signal that the vocabulary
// itself is wrong), every declared Grade must not exceed its derived
// ceiling, and -- for files NOT in validationGradeBarExemptions -- every
// declared grade must additionally satisfy the bar (>= EXERCISED).
func TestValidationRowGradesAreEarnedOverArchivedCorpus(t *testing.T) {
	index := buildTestIndex(t)
	byDoc := allPrimaryValidationRows(t)
	if len(byDoc) == 0 {
		t.Fatal("no *-VALIDATION.md primary table discovered -- the glob or table detector has gone inert")
	}
	record := corpusRunRecord(t, index, byDoc)

	docs := make([]string, 0, len(byDoc))
	for doc := range byDoc {
		docs = append(docs, doc)
	}
	sortStrings(docs)

	for _, doc := range docs {
		name := filepath.Base(doc)
		t.Run(name, func(t *testing.T) {
			rows := byDoc[doc]
			if len(rows) == 0 {
				t.Fatalf("%s: primary table discovered but has zero data rows", name)
			}
			_, exempt := validationGradeBarExemptions[name]
			for _, row := range rows {
				taskID := rowIdentifier(row)
				gradeCell, hasGrade := row.Cells["Grade"]
				nonInertCell, hasNonInert := row.Cells["Non-inertness"]
				if !hasGrade {
					t.Fatalf("%s: row %s has no Grade column -- derivation cannot classify a column that does not exist", name, taskID)
				}
				if !hasNonInert {
					t.Fatalf("%s: row %s has no Non-inertness column", name, taskID)
				}
				evidence := rowEvidence(row)
				if problem := validationRowProblem(index, record, taskID, gradeCell, nonInertCell, evidence); problem != "" {
					t.Fatalf("%s: %s", name, problem)
				}
				if !exempt {
					declared := strings.TrimSpace(gradeCell)
					if validationGradeOrdinal[declared] < validationGradeOrdinal[validationSatisfyingGrade] {
						if _, rowExempt := validationGradeBarRowExemptions[name+":"+taskID]; !rowExempt {
							t.Fatalf("%s: row %s declares %s, below the satisfying bar (%s) enforced for this file (not in validationGradeBarExemptions or validationGradeBarRowExemptions)", name, taskID, declared, validationSatisfyingGrade)
						}
					}
				}
			}
		})
	}
}

// TestValidationGradeBarExemptionsAreFileScoped asserts every exemption
// key is a bare filename (never a row identifier) and carries a non-empty
// reason -- the mechanical form of "exemptions are file-scoped and dated,
// never row-scoped" (D-14-06's own prohibition).
func TestValidationGradeBarExemptionsAreFileScoped(t *testing.T) {
	for key, reason := range validationGradeBarExemptions {
		if !enforcedBasenamePattern.MatchString(key) || strings.Contains(key, "/") || strings.Contains(key, "|") {
			t.Fatalf("validationGradeBarExemptions key %q does not look like a bare *-VALIDATION.md filename", key)
		}
		if strings.TrimSpace(reason) == "" {
			t.Fatalf("validationGradeBarExemptions[%q] has an empty reason", key)
		}
	}
}

// ---------------------------------------------------------------------
// (h) Witnesses for PHASE-14-DEBT.md's below-shipped-verdict findings
// (D-14-07). Two of the five findings this plan surfaced (09-VALIDATION.md:85,
// :93) are already witnessed by 14-01/14-06's own pinned groundedness
// frontier (TestVerificationGroundednessFrontierIsPinned): both rows are
// literal entries in that pinned literal, so if either pattern ever became
// findable again, the frontier pin itself would need re-measuring --
// exactly D-14-29's "ungrading is forced by execution." The remaining
// three findings are NOT covered by that lint (assert-go-tests.sh
// invocations and a bare package-only `go test` cell are both outside its
// `go test -run` scope), so this probe supplies their witness directly.
// ---------------------------------------------------------------------

// validationGradeArchivedDeadCitations names the four exact top-level
// identifiers this plan's derivation found cited in an archived
// *-VALIDATION.md row (04-VALIDATION.md:74, 06-VALIDATION.md:69,
// 06-VALIDATION.md:71) that do not exist anywhere in the current module:
// two were retired alongside the `computeLoanLastUses` deletion
// (04-DEBT.md's own recorded retirement), and two were renamed in plan
// 08-05 (TestOnlyRecomputedWorkIsGateEligible -> TestOnlyGateEligibleMetricsPassThrough)
// and its own reviewed follow-up (TestRecomputedWorkIsTheOnlyHardGate ->
// TestRecomputedWorkHardGateBoundComparison, per 08-REVIEW.md's WR-01,
// which named this exact staleness and recommended the rename that
// landed). Their absence is the claim; if any of them is ever declared
// again, this probe goes red (XPASS in the LLVM lit sense) and the
// corresponding debt row must be re-graded, never silently re-passed.
var validationGradeArchivedDeadCitations = []string{
	"TestLastUseDiscoveryWorkIsCounted",
	"TestLastUseDiscoveryWorkSeries",
	"TestOnlyRecomputedWorkIsGateEligible",
	"TestRecomputedWorkIsTheOnlyHardGate",
}

// TestValidationGradeCapArchivedDeadCitationsRemainAbsent is D-14-48/49's
// witness: the three names above must not exist as a top-level
// Test/Fuzz/Benchmark/Example identifier anywhere the static index covers.
func TestValidationGradeCapArchivedDeadCitationsRemainAbsent(t *testing.T) {
	index := buildTestIndex(t)
	for _, name := range validationGradeArchivedDeadCitations {
		if isTopLevelTestName(index, name) {
			t.Fatalf("%s now exists in the module -- the debt row citing it as a dead reference must be re-graded, not left as WIRED", name)
		}
	}
}

// TestValidationGradeCapBarePackageRowHasNoNamedTest is D-14-50's witness
// (12-VALIDATION.md's row 12-04-01): a `go test` cell naming packages but
// no -run/-list/-fuzz/-bench flag derives WIRED because it names no exact
// test identifier, not because anything is broken. Re-derives the row's
// own verbatim evidence text and asserts the ceiling is exactly WIRED, so
// a future edit that adds a pattern to this cell (which would change the
// ceiling) is caught here rather than leaving the debt row stale.
func TestValidationGradeCapBarePackageRowHasNoNamedTest(t *testing.T) {
	index := buildTestIndex(t)
	evidence := "go test ./internal/compiler/originvalidate/... ./internal/compiler/corevalidate/... -count=1"
	ceiling, matched := deriveCeiling(index, nil, evidence)
	if ceiling != "WIRED" {
		t.Fatalf("12-VALIDATION.md row 12-04-01's evidence now derives %s (matched %v), not WIRED -- re-grade the debt row", ceiling, matched)
	}
}

// ---------------------------------------------------------------------
// (i) Plan 14-09 Task 3 -- the cap's own non-inertness proof
// (scripts/assert-go-tests.sh:44-62's sentinel is the direct analogue:
// proving a nonexistent name is correctly rejected). One seeded fault per
// mechanizable kind, plus the unfaulted control, over synthetic rows only
// -- never the real .planning/** tree.
// ---------------------------------------------------------------------

// TestValidationGradeCapIsNotInert seeds one fault per mechanizable kind
// into a synthetic row and asserts the cap refuses each; the unfaulted
// control (declaring the derived ceiling exactly) must accept. A cap
// proven red on only one seeded fault is inert for the others.
func TestValidationGradeCapIsNotInert(t *testing.T) {
	index := syntheticIndex(t)
	record := syntheticRunRecord()

	t.Run("mutation-killed grade with a nonexistent non-inertness twin fails", func(t *testing.T) {
		problem := validationRowProblem(index, record, "seed-mk-nonexistent-twin",
			"MUTATION-KILLED", "TestSyntheticTwinDoesNotExist",
			"go test ./internal/compiler/check -run TestSyntheticPrimaryClaim")
		if problem == "" {
			t.Fatal("declaring MUTATION-KILLED with a nonexistent non-inertness twin must fail the cap")
		}
	})

	t.Run("exercised grade whose evidence names a nonexistent test fails", func(t *testing.T) {
		problem := validationRowProblem(index, record, "seed-exercised-nonexistent-evidence",
			"EXERCISED", "—",
			"go test ./internal/compiler/check -run TestDoesNotExistAnywhere")
		if problem == "" {
			t.Fatal("declaring EXERCISED with evidence naming a nonexistent test must fail the cap")
		}
	})

	t.Run("exercised grade with a resolving test but no run record fails", func(t *testing.T) {
		problem := validationRowProblem(index, nil, "seed-exercised-no-run-record",
			"EXERCISED", "—",
			"go test ./internal/compiler/check -run TestSyntheticPrimaryClaim")
		if problem == "" {
			t.Fatal("declaring EXERCISED with a resolving test but record == nil (no run record present) must fail the cap")
		}
	})

	t.Run("the unfaulted control accepts", func(t *testing.T) {
		if problem := validationRowProblem(index, record, "seed-control",
			"EXERCISED", "—",
			"go test ./internal/compiler/check -run TestSyntheticPrimaryClaim"); problem != "" {
			t.Fatalf("unfaulted control (declared grade == derived ceiling) should pass, got: %s", problem)
		}
	})
}
