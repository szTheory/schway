package session

import (
	"bytes"
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	langast "github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/evidence"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase6Fixture(t *testing.T, name string) []byte {
	t.Helper()
	path := testsupport.ProjectPath("testdata", "phase6", name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", name, err)
	}
	return source
}

// hasDriverEligibleRepair reports whether any diagnostic in diagnostics
// carries a diagnostic.DriverEligible repair.
func hasDriverEligibleRepair(diagnostics []diagnostic.Diagnostic) bool {
	for _, problem := range diagnostics {
		for _, repair := range problem.Repairs {
			if diagnostic.DriverEligible(repair) {
				return true
			}
		}
	}
	return false
}

// injectorError, when non-nil, unwraps err to *InjectorError or returns nil.
func injectorError(err error) *InjectorError {
	var typed *InjectorError
	if errors.As(err, &typed) {
		return typed
	}
	return nil
}

// TestMatchDefectInjectorProducesExactlyOneDefect is Task 1's tracer
// end-to-end assertion: a real held-out fixture, mutated by a real
// injector, checked by the real checker, produces a real repair-bearing
// diagnostic.
func TestMatchDefectInjectorProducesExactlyOneDefect(t *testing.T) {
	source := phase6Fixture(t, "heldout_match_defect.schway")

	clean := Check(source)
	if len(clean.Diagnostics) != 0 {
		t.Fatalf("heldout_match_defect.schway must check clean unmutated: %+v", clean.Diagnostics)
	}

	mutated, err := MatchInjector{}.Inject(source)
	if err != nil {
		t.Fatalf("MatchInjector.Inject failed on an eligible fixture: %v", err)
	}
	if bytes.Equal(mutated, source) {
		t.Fatal("MatchInjector.Inject returned the source unmutated")
	}

	checked := Check(mutated)
	if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "match.non_exhaustive" {
		t.Fatalf("mutated fixture did not reject with match.non_exhaustive: %+v", checked.Diagnostics)
	}
	if !hasDriverEligibleRepair(checked.Diagnostics) {
		t.Fatalf("match.non_exhaustive diagnostic carries no DriverEligible repair: %+v", checked.Diagnostics)
	}

	noTarget := bytes.ReplaceAll(source, []byte(matchTargetMarker), []byte(""))
	_, err = MatchInjector{}.Inject(noTarget)
	typed := injectorError(err)
	if typed == nil || typed.Code != InjectorTargetMissingCode {
		t.Fatalf("MatchInjector.Inject on a fixture with no eligible arm did not refuse with %s: %v", InjectorTargetMissingCode, err)
	}
}

// TestPhase6DefectCorpusIsHeldOut asserts D-06-29's structural split: the
// heldout_ and derivation_ prefix sets are both non-empty and genuinely
// distinct (different file content per class), not merely differently
// named.
func TestPhase6DefectCorpusIsHeldOut(t *testing.T) {
	dir := testsupport.ProjectPath("testdata", "phase6")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var heldout, derivation []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".schway") {
			continue
		}
		switch {
		case strings.HasPrefix(entry.Name(), "heldout_"):
			heldout = append(heldout, entry.Name())
		case strings.HasPrefix(entry.Name(), "derivation_"):
			derivation = append(derivation, entry.Name())
		}
	}
	if len(heldout) == 0 {
		t.Fatal("no heldout_*.schway fixtures found in testdata/phase6")
	}
	if len(derivation) == 0 {
		t.Fatal("no derivation_*.schway fixtures found in testdata/phase6")
	}
	seen := make(map[string]bool, len(heldout))
	for _, name := range heldout {
		seen[name] = true
	}
	for _, name := range derivation {
		if seen[name] {
			t.Fatalf("%s appears in both the heldout_ and derivation_ sets", name)
		}
	}
	readmePath := filepath.Join(dir, "README")
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("testdata/phase6/README must exist and name both prefixes: %v", err)
	}
	if !strings.Contains(string(readme), "heldout_") || !strings.Contains(string(readme), "derivation_") {
		t.Fatal("testdata/phase6/README does not name both the heldout_ and derivation_ prefixes")
	}

	// Genuine distinctness, not just naming: each class's heldout and
	// derivation fixture must differ in content.
	for _, class := range []string{"match", "move", "borrow"} {
		heldoutBytes := phase6Fixture(t, "heldout_"+class+"_defect.schway")
		derivationBytes := phase6Fixture(t, "derivation_"+class+"_defect.schway")
		if bytes.Equal(heldoutBytes, derivationBytes) {
			t.Fatalf("%s class: heldout and derivation fixtures are byte-identical", class)
		}
	}

	// D-13-33: each class must be structurally distinct under the
	// identifier-independent summary. D-13-34's replacement branch restores
	// this evidence for the move and borrow fixtures.
	for _, class := range []string{"match", "move", "borrow"} {
		class := class
		t.Run(class+"_structural_distinctness", func(t *testing.T) {
			heldoutSource := phase6Fixture(t, "heldout_"+class+"_defect.schway")
			derivationSource := phase6Fixture(t, "derivation_"+class+"_defect.schway")
			heldoutSummary := computePhase6StructuralSummary(t, heldoutSource)
			derivationSummary := computePhase6StructuralSummary(t, derivationSource)
			t.Logf("%s class structural summary: heldout=%+v derivation=%+v", class, heldoutSummary, derivationSummary)
			if heldoutSummary == derivationSummary {
				t.Fatalf("%s class: heldout and derivation fixtures are structurally IDENTICAL (%+v) -- byte-inequality alone would have passed this pair, which is the exact M001 weakness D-13-33 exists to close", class, heldoutSummary)
			}
		})
	}
}

// TestPhase6HeldoutPairsAreStructurallyDistinct is the live D-13-34 closure
// witness. It pins distinct structure for all three held-out/derivation
// classes independently of identifier spelling.
func TestPhase6HeldoutPairsAreStructurallyDistinct(t *testing.T) {
	for _, class := range []string{"match", "move", "borrow"} {
		class := class
		t.Run(class, func(t *testing.T) {
			heldoutSummary := computePhase6StructuralSummary(t, phase6Fixture(t, "heldout_"+class+"_defect.schway"))
			derivationSummary := computePhase6StructuralSummary(t, phase6Fixture(t, "derivation_"+class+"_defect.schway"))
			if heldoutSummary == derivationSummary {
				t.Fatalf("D-13-34: %s class heldout/derivation fixtures are structurally identical: %+v", class, heldoutSummary)
			}
		})
	}
}

// phase6StructuralSummary is D-13-33's identifier-independent structural
// predicate for testdata/phase6's distinctness control: alpha-renaming a
// fixture (module name, function name, parameter/binding names, alternative
// names) leaves every one of these components unchanged, so -- unlike
// byte-inequality -- structural equality here is undefeated by a pure
// rename. testdata/phase6 has zero interprocedural fixtures, so this
// intentionally omits the D-13-26 topology triple (function count / call-edge
// count / hop distance), which degenerates to a constant across the whole
// corpus and would tell us nothing.
type phase6StructuralSummary struct {
	bindingCount  int
	matchArmCount int
	borrowCount   int
	takeCount     int
	maxDepth      int
}

// summarizePhase6LinearBody counts bindings, takes, and borrows in one flat
// linear body (this language's LinearBody has no nested-block construct of
// its own; nesting arises only from a match arm carrying a body, handled by
// the caller).
func summarizePhase6LinearBody(body *langast.LinearBody) (bindingCount, takeCount, borrowCount int) {
	if body == nil {
		return 0, 0, 0
	}
	bindingCount = len(body.Bindings)
	for _, binding := range body.Bindings {
		switch binding.RHS.Kind {
		case "take":
			takeCount++
		case "borrow":
			borrowCount++
		}
	}
	return bindingCount, takeCount, borrowCount
}

// computePhase6StructuralSummary parses source independently (never
// consulting a cached/shared program) and aggregates phase6StructuralSummary
// across every declared function, so a multi-function fixture (none exist in
// testdata/phase6 today, but the predicate must not silently assume
// single-function input) is summarized correctly too.
func computePhase6StructuralSummary(t *testing.T, source []byte) phase6StructuralSummary {
	t.Helper()
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("computePhase6StructuralSummary: parse failed: %+v", parsed.Diagnostics)
	}
	var summary phase6StructuralSummary
	for _, function := range parsed.Program.Funcs {
		if function.Body.Linear != nil {
			bindingCount, takeCount, borrowCount := summarizePhase6LinearBody(function.Body.Linear)
			summary.bindingCount += bindingCount
			summary.takeCount += takeCount
			summary.borrowCount += borrowCount
			if summary.maxDepth < 1 {
				summary.maxDepth = 1
			}
			continue
		}
		summary.matchArmCount += len(function.Body.MatchExpr.Arms)
		depth := 1
		for _, arm := range function.Body.MatchExpr.Arms {
			if arm.Body != nil {
				bindingCount, takeCount, borrowCount := summarizePhase6LinearBody(arm.Body)
				summary.bindingCount += bindingCount
				summary.takeCount += takeCount
				summary.borrowCount += borrowCount
				if depth < 2 {
					depth = 2
				}
			}
		}
		if summary.maxDepth < depth {
			summary.maxDepth = depth
		}
	}
	return summary
}

// alphaRenamePhase6MatchDefect returns a byte-for-byte structural copy of
// derivation_match_defect.schway with every identifier renamed (module suffix,
// data type name, alternative names, function name, parameter name) --
// exactly M001's documented weakness shape (item -> buffer,
// moved_once -> delivered): a rename that changes no structural component at
// all, so its structural summary must come out IDENTICAL to the original.
func alphaRenamePhase6MatchDefect(source []byte) []byte {
	renamed := string(source)
	replacements := []struct{ from, to string }{
		{"derivation_match_defect", "alpha_renamed_derivation_match_defect"},
		{"Mode", "Status"},
		{"Idle", "Dormant"},
		{"Active", "Running"},
		{"relay", "route"},
		{"state", "condition"},
	}
	for _, replacement := range replacements {
		pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(replacement.from) + `\b`)
		renamed = pattern.ReplaceAllString(renamed, replacement.to)
	}
	return []byte(renamed)
}

// TestPhase6DefectCorpusDistinctnessGuardIsNotInert is D-13-33's required
// mutation kill, mirroring 13-04's TestCorpusTopologyGuardIsNotInert
// (session_phase13_injectors_test.go) at intraprocedural scale: alpha-
// renaming a derivation fixture must not change its structural summary,
// proving that had the renamed copy been submitted as the held-out member in
// its place, TestPhase6DefectCorpusIsHeldOut's structural-equality check
// above would have caught it -- exactly the discrimination byte-inequality
// alone could never make. derivation_match_defect.schway is the vehicle
// because the match class's real heldout/derivation pair is the one class
// whose structural summaries genuinely differ today (2 arms vs 3), so this
// test also proves the predicate discriminates the real pair, not merely
// that it fails to reject a rename.
func TestPhase6DefectCorpusDistinctnessGuardIsNotInert(t *testing.T) {
	original := phase6Fixture(t, "derivation_match_defect.schway")
	renamed := alphaRenamePhase6MatchDefect(original)
	dir := t.TempDir()
	path := filepath.Join(dir, "alpha_renamed_derivation_match_defect.schway")
	if err := os.WriteFile(path, renamed, 0o644); err != nil {
		t.Fatal(err)
	}
	renamedSource, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(renamedSource, original) {
		t.Fatal("alpha-rename produced byte-identical output -- rename did not actually happen")
	}
	originalSummary := computePhase6StructuralSummary(t, original)
	renamedSummary := computePhase6StructuralSummary(t, renamedSource)
	if renamedSummary != originalSummary {
		t.Fatalf("control is INERT in the wrong direction: alpha-renaming changed the structural summary (original=%+v, renamed=%+v) -- rename must preserve structure for this test to demonstrate anything", originalSummary, renamedSummary)
	}
	// This IS the not-inert proof: had this renamed copy been submitted as
	// the held-out member of the match class in place of the real
	// heldout_match_defect.schway, TestPhase6DefectCorpusIsHeldOut's
	// `heldoutSummary == derivationSummary` check above would have fired on
	// exactly this pair -- demonstrated here by showing the renamed copy's
	// summary is identical to the original derivation fixture's summary.
	heldoutSummary := computePhase6StructuralSummary(t, phase6Fixture(t, "heldout_match_defect.schway"))
	if heldoutSummary == originalSummary {
		t.Fatalf("predicate is not discriminating: the real heldout_match_defect.schway pair is already structurally equal to derivation (%+v) -- this test cannot demonstrate a violation on a pair the predicate cannot tell apart in the first place", heldoutSummary)
	}
}

// TestMoveDefectInjectorProducesExactlyOneDefect mirrors
// TestMatchDefectInjectorProducesExactlyOneDefect for the move class.
func TestMoveDefectInjectorProducesExactlyOneDefect(t *testing.T) {
	source := phase6Fixture(t, "heldout_move_defect.schway")

	clean := Check(source)
	if len(clean.Diagnostics) != 0 {
		t.Fatalf("heldout_move_defect.schway must check clean unmutated: %+v", clean.Diagnostics)
	}

	mutated, err := MoveInjector{}.Inject(source)
	if err != nil {
		t.Fatalf("MoveInjector.Inject failed on an eligible fixture: %v", err)
	}
	if bytes.Equal(mutated, source) {
		t.Fatal("MoveInjector.Inject returned the source unmutated")
	}

	checked := Check(mutated)
	if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "ownership.use_after_move" {
		t.Fatalf("mutated fixture did not reject with ownership.use_after_move: %+v", checked.Diagnostics)
	}
	if !hasDriverEligibleRepair(checked.Diagnostics) {
		t.Fatalf("ownership.use_after_move diagnostic carries no DriverEligible repair: %+v", checked.Diagnostics)
	}

	noTarget := bytes.ReplaceAll(source, []byte(moveTargetMarker), []byte(""))
	_, err = MoveInjector{}.Inject(noTarget)
	typed := injectorError(err)
	if typed == nil || typed.Code != InjectorTargetMissingCode {
		t.Fatalf("MoveInjector.Inject on a fixture with no eligible take did not refuse with %s: %v", InjectorTargetMissingCode, err)
	}
}

// TestBorrowDefectInjectorProducesExactlyOneDefect mirrors
// TestMatchDefectInjectorProducesExactlyOneDefect for the borrow class.
func TestBorrowDefectInjectorProducesExactlyOneDefect(t *testing.T) {
	source := phase6Fixture(t, "heldout_borrow_defect.schway")

	clean := Check(source)
	if len(clean.Diagnostics) != 0 {
		t.Fatalf("heldout_borrow_defect.schway must check clean unmutated: %+v", clean.Diagnostics)
	}

	mutated, err := BorrowInjector{}.Inject(source)
	if err != nil {
		t.Fatalf("BorrowInjector.Inject failed on an eligible fixture: %v", err)
	}
	if bytes.Equal(mutated, source) {
		t.Fatal("BorrowInjector.Inject returned the source unmutated")
	}

	checked := Check(mutated)
	if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "ownership.borrow_conflict" {
		t.Fatalf("mutated fixture did not reject with ownership.borrow_conflict: %+v", checked.Diagnostics)
	}
	if !hasDriverEligibleRepair(checked.Diagnostics) {
		t.Fatalf("ownership.borrow_conflict diagnostic carries no DriverEligible repair: %+v", checked.Diagnostics)
	}

	noTarget := bytes.ReplaceAll(source, []byte(borrowTargetMarker), []byte(""))
	_, err = BorrowInjector{}.Inject(noTarget)
	typed := injectorError(err)
	if typed == nil || typed.Code != InjectorTargetMissingCode {
		t.Fatalf("BorrowInjector.Inject on a fixture with no eligible loan did not refuse with %s: %v", InjectorTargetMissingCode, err)
	}
}

// injectorsSourceAST parses this package's own session_phase6_injectors.go,
// the shared fixture every go/ast assertion below inspects.
func injectorsSourceAST(t *testing.T) (*ast.File, []byte) {
	t.Helper()
	path := testsupport.ProjectPath("internal", "compiler", "session", "session_phase6_injectors.go")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, source, 0)
	if err != nil {
		t.Fatal(err)
	}
	return file, source
}

// containsCallTo reports whether file's AST contains any call whose callee
// (a selector or plain identifier) has the given name.
func containsCallTo(file *ast.File, name string) bool {
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fn := call.Fun.(type) {
		case *ast.Ident:
			if fn.Name == name {
				found = true
			}
		case *ast.SelectorExpr:
			if fn.Sel.Name == name {
				found = true
			}
		}
		return true
	})
	return found
}

// TestCleanupInjectorReusesReleaseOmissionRunner asserts, by go/ast, that
// CleanupInjector delegates to ReleaseOmissionMutationRunner and that this
// file declares no second release-marker scan of its own (D-06-25), then
// proves the delegation functionally against a real generated C source.
func TestCleanupInjectorReusesReleaseOmissionRunner(t *testing.T) {
	file, source := injectorsSourceAST(t)
	if !containsCallTo(file, "NewReleaseOmissionMutationRunner") {
		t.Fatal("session_phase6_injectors.go does not call NewReleaseOmissionMutationRunner")
	}
	if !containsCallTo(file, "Mutate") {
		t.Fatal("session_phase6_injectors.go does not call ReleaseOmissionMutationRunner's Mutate method")
	}
	if strings.Contains(string(source), releaseMarker) {
		t.Fatalf("session_phase6_injectors.go contains its own %q literal -- a second release-marker scan", releaseMarker)
	}

	checked, err := CheckFile(testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cgen.EmitNative(checked.Program); err == nil || !strings.Contains(err.Error(), "multi-function foreign-call bodies are not supported") {
		t.Fatalf("EmitNative must retain the named M004 refusal, got %v", err)
	}
	cSource, err := phase16InternalFrozenEvidenceC(checked.Program, "testdata/phase4/acquire_three_success.schway")
	if err != nil {
		t.Fatalf("refusal-first frozen control C: %v", err)
	}
	originalCount := strings.Count(cSource, releaseMarker)
	if originalCount == 0 {
		t.Fatal("fixture's generated C carries no release-site marker; test setup is broken")
	}

	mutated, err := CleanupInjector{}.Inject([]byte(cSource))
	if err != nil {
		t.Fatalf("CleanupInjector.Inject failed on an eligible C source: %v", err)
	}
	if strings.Count(string(mutated), releaseMarker) != originalCount-1 {
		t.Fatalf("CleanupInjector.Inject did not remove exactly one release-site marker: got %d, want %d",
			strings.Count(string(mutated), releaseMarker), originalCount-1)
	}

	noTarget := strings.ReplaceAll(cSource, releaseMarker, "")
	_, err = CleanupInjector{}.Inject([]byte(noTarget))
	typed := injectorError(err)
	if typed == nil || typed.Code != InjectorTargetMissingCode {
		t.Fatalf("CleanupInjector.Inject on C source with no release-site marker did not refuse with %s: %v", InjectorTargetMissingCode, err)
	}
}

// TestStaleEvidenceInjectorBreaksManifestBinding proves the stale-evidence
// class end-to-end: a manifest captured over the clean subject validates,
// the SAME manifest fails to validate against the re-touched subject, and
// -- by go/ast -- the injector computes no source diff (it imports no
// differential/reduce machinery at all).
func TestStaleEvidenceInjectorBreaksManifestBinding(t *testing.T) {
	file, _ := injectorsSourceAST(t)
	for _, imported := range file.Imports {
		path := strings.Trim(imported.Path.Value, `"`)
		if strings.Contains(path, "/compiler/reduce") {
			t.Fatalf("session_phase6_injectors.go imports %s, a source-diff/differential package the stale-evidence locator must never consult (D-06-25)", path)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	source := phase6Fixture(t, "stale_evidence_subject.schway")

	facts, err := evidence.DefaultFacts(ctx, "clang")
	if err != nil {
		t.Skipf("env:clang toolchain unavailable, skipping stale-evidence exercise: %v", err)
	}
	product, diagnostics, err := evidence.Build(source, facts)
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("stale_evidence_subject.schway did not build cleanly: %+v", diagnostics)
	}
	if err := evidence.Validate(product.Manifest, source, facts); err != nil {
		t.Fatalf("captured manifest did not validate against its own source before injection: %v", err)
	}

	mutated, err := StaleEvidenceInjector{}.Inject(source)
	if err != nil {
		t.Fatalf("StaleEvidenceInjector.Inject failed on an eligible subject: %v", err)
	}
	if bytes.Equal(mutated, source) {
		t.Fatal("StaleEvidenceInjector.Inject returned the source unmutated")
	}
	if err := evidence.Validate(product.Manifest, mutated, facts); err == nil {
		t.Fatal("the captured manifest still validated against the re-touched source; the binding was not broken")
	}

	noTarget := bytes.ReplaceAll(source, []byte(evidenceSubjectMarker), []byte(""))
	_, err = StaleEvidenceInjector{}.Inject(noTarget)
	typed := injectorError(err)
	if typed == nil || typed.Code != InjectorTargetMissingCode {
		t.Fatalf("StaleEvidenceInjector.Inject on a fixture with no evidence-subject marker did not refuse with %s: %v", InjectorTargetMissingCode, err)
	}
}

// lineDiffCount reports how many lines differ between original and mutated
// under either a same-length or a one-line-shorter alignment -- the two
// shapes this plan's four source-granularity injectors ever produce.
func lineDiffCount(t *testing.T, name string, original, mutated []byte) int {
	t.Helper()
	origLines := strings.Split(string(original), "\n")
	mutLines := strings.Split(string(mutated), "\n")
	switch len(mutLines) - len(origLines) {
	case 0:
		diff := 0
		for index := range origLines {
			if origLines[index] != mutLines[index] {
				diff++
			}
		}
		return diff
	case -1:
		diff := 0
		mutIndex := 0
		for _, origLine := range origLines {
			if mutIndex < len(mutLines) && origLine == mutLines[mutIndex] {
				mutIndex++
				continue
			}
			diff++
		}
		return diff
	default:
		t.Fatalf("%s: unexpected line-count delta %d (orig=%d mutated=%d)", name, len(mutLines)-len(origLines), len(origLines), len(mutLines))
		return -1
	}
}

// TestEveryInjectorProducesExactlyOneMechanicalChange covers all four
// source-granularity classes (D-06-26's byte-identity oracle depends on
// this): match deletes exactly one arm line, move/borrow each rewrite
// exactly one line, cleanup deletes exactly one generated C line.
func TestEveryInjectorProducesExactlyOneMechanicalChange(t *testing.T) {
	matchSource := phase6Fixture(t, "heldout_match_defect.schway")
	matchMutated, err := MatchInjector{}.Inject(matchSource)
	if err != nil {
		t.Fatal(err)
	}
	if got := lineDiffCount(t, "match", matchSource, matchMutated); got != 1 {
		t.Fatalf("match: got %d differing lines, want 1", got)
	}

	moveSource := phase6Fixture(t, "heldout_move_defect.schway")
	moveMutated, err := MoveInjector{}.Inject(moveSource)
	if err != nil {
		t.Fatal(err)
	}
	if got := lineDiffCount(t, "move", moveSource, moveMutated); got != 1 {
		t.Fatalf("move: got %d differing lines, want 1", got)
	}

	borrowSource := phase6Fixture(t, "heldout_borrow_defect.schway")
	borrowMutated, err := BorrowInjector{}.Inject(borrowSource)
	if err != nil {
		t.Fatal(err)
	}
	if got := lineDiffCount(t, "borrow", borrowSource, borrowMutated); got != 1 {
		t.Fatalf("borrow: got %d differing lines, want 1", got)
	}

	checked, err := CheckFile(testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.schway"))
	if err != nil {
		t.Fatal(err)
	}
	cSource, err := phase16InternalFrozenEvidenceC(checked.Program, "testdata/phase4/acquire_three_success.schway")
	if err != nil {
		t.Fatal(err)
	}
	cleanupMutated, err := CleanupInjector{}.Inject([]byte(cSource))
	if err != nil {
		t.Fatal(err)
	}
	if got := lineDiffCount(t, "cleanup", []byte(cSource), cleanupMutated); got != 1 {
		t.Fatalf("cleanup: got %d differing lines, want 1", got)
	}
}

// TestInjectorTargetChoiceIsSpecified asserts, for each source-granularity
// injector, that an ambiguous (two-marker) input resolves to the LAST
// marked line, deterministically across repeated invocations -- the same
// choice ReleaseOmissionMutationRunner already makes for schway:release-site.
func TestInjectorTargetChoiceIsSpecified(t *testing.T) {
	// match: two marked arms, both otherwise-eligible; the LAST one is removed.
	matchSource := []byte("data Signal =\n  | Red\n  | Green\n\nfn relay(state: Signal) -> Signal {\n  match state {\n    Red => Red // schway:match-target\n    Green => Green // schway:match-target\n  }\n}")
	first, err := MatchInjector{}.Inject(matchSource)
	if err != nil {
		t.Fatal(err)
	}
	second, err := MatchInjector{}.Inject(matchSource)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("MatchInjector.Inject is not deterministic across repeated invocations on the same ambiguous input")
	}
	if strings.Contains(string(first), "Green => Green") || !strings.Contains(string(first), "Red => Red") {
		t.Fatalf("MatchInjector.Inject did not remove the LAST marked arm: %q", first)
	}

	// move: two marked take-expressions; the LAST is corrupted.
	moveSource := []byte("fn relay(buffer: Buffer) -> Buffer {\n  let a = take buffer // schway:move-target\n  let b = take a // schway:move-target\n  b\n}")
	moveFirst, err := MoveInjector{}.Inject(moveSource)
	if err != nil {
		t.Fatal(err)
	}
	moveSecond, err := MoveInjector{}.Inject(moveSource)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(moveFirst, moveSecond) {
		t.Fatal("MoveInjector.Inject is not deterministic across repeated invocations on the same ambiguous input")
	}
	if !strings.Contains(string(moveFirst), "let a = take buffer") || strings.Contains(string(moveFirst), "let b = take a") {
		t.Fatalf("MoveInjector.Inject did not corrupt the LAST marked take: %q", moveFirst)
	}

	// borrow: two marked shared borrows; the LAST is escalated.
	borrowSource := []byte("fn relay(buffer: Buffer) -> Buffer {\n  let a = borrow buffer // schway:borrow-target\n  let b = borrow buffer // schway:borrow-target\n  b\n}")
	borrowFirst, err := BorrowInjector{}.Inject(borrowSource)
	if err != nil {
		t.Fatal(err)
	}
	borrowSecond, err := BorrowInjector{}.Inject(borrowSource)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(borrowFirst, borrowSecond) {
		t.Fatal("BorrowInjector.Inject is not deterministic across repeated invocations on the same ambiguous input")
	}
	if !strings.Contains(string(borrowFirst), "let a = borrow buffer") || !strings.Contains(string(borrowFirst), "let b = borrow mut buffer") {
		t.Fatalf("BorrowInjector.Inject did not escalate the LAST marked borrow: %q", borrowFirst)
	}

	// interprocedural_loan (Phase 13, D-13-30a extension): two marked
	// take-statements, each immediately preceded by a call; the LAST
	// marked statement is swapped with its own predecessor, the FIRST
	// marked pair is left untouched.
	loanSource := []byte("fn relay(buffer: Buffer) -> Buffer {\n  let a = sink(buffer)\n  let b = take buffer // schway:interprocedural-loan-target\n  let c = sink(buffer)\n  let d = take buffer // schway:interprocedural-loan-target\n  d\n}")
	loanFirst, err := InterproceduralLoanInjector{}.Inject(loanSource)
	if err != nil {
		t.Fatal(err)
	}
	loanSecond, err := InterproceduralLoanInjector{}.Inject(loanSource)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(loanFirst, loanSecond) {
		t.Fatal("InterproceduralLoanInjector.Inject is not deterministic across repeated invocations on the same ambiguous input")
	}
	loanText := string(loanFirst)
	if strings.Index(loanText, "let d = take buffer") == -1 || strings.Index(loanText, "let d = take buffer") > strings.Index(loanText, "let c = sink(buffer)") {
		t.Fatalf("InterproceduralLoanInjector.Inject did not swap the LAST marked statement ahead of its predecessor: %q", loanFirst)
	}
	if strings.Index(loanText, "let a = sink(buffer)") == -1 || strings.Index(loanText, "let a = sink(buffer)") > strings.Index(loanText, "let b = take buffer") {
		t.Fatalf("InterproceduralLoanInjector.Inject disturbed the FIRST marked pair, which should be untouched: %q", loanFirst)
	}

	// fallible_consume (Phase 13, D-13-30a extension): two marked `= try`
	// bindings; the LAST is stripped of `try`, the FIRST is untouched.
	consumeSource := []byte("fn main(request: Byte) -> Byte {\n  let a = try open(request) // schway:fallible-consume-target\n  let b = try open(request) // schway:fallible-consume-target\n  b\n}")
	consumeFirst, err := FallibleConsumeInjector{}.Inject(consumeSource)
	if err != nil {
		t.Fatal(err)
	}
	consumeSecond, err := FallibleConsumeInjector{}.Inject(consumeSource)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(consumeFirst, consumeSecond) {
		t.Fatal("FallibleConsumeInjector.Inject is not deterministic across repeated invocations on the same ambiguous input")
	}
	if !strings.Contains(string(consumeFirst), "let a = try open(request)") || strings.Contains(string(consumeFirst), "let b = try open(request)") {
		t.Fatalf("FallibleConsumeInjector.Inject did not strip `try` from the LAST marked binding: %q", consumeFirst)
	}
	if !strings.Contains(string(consumeFirst), "let b = open(request)") {
		t.Fatalf("FallibleConsumeInjector.Inject did not leave the LAST marked binding as a bare call: %q", consumeFirst)
	}

	// call_argument_type (Phase 13, D-13-30a extension): two marked `fn`
	// declaration lines; the LAST has Byte/Buffer toggled, the FIRST is
	// untouched.
	typeSource := []byte("fn first(value: Byte) -> Byte { // schway:call-argument-target\n  value\n}\n\nfn second(value: Byte) -> Byte { // schway:call-argument-target\n  value\n}")
	typeFirst, err := CallArgumentTypeInjector{}.Inject(typeSource)
	if err != nil {
		t.Fatal(err)
	}
	typeSecond, err := CallArgumentTypeInjector{}.Inject(typeSource)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(typeFirst, typeSecond) {
		t.Fatal("CallArgumentTypeInjector.Inject is not deterministic across repeated invocations on the same ambiguous input")
	}
	if !strings.Contains(string(typeFirst), "fn first(value: Byte) -> Byte {") {
		t.Fatalf("CallArgumentTypeInjector.Inject disturbed the FIRST marked declaration, which should be untouched: %q", typeFirst)
	}
	if !strings.Contains(string(typeFirst), "fn second(value: Buffer) -> Buffer {") {
		t.Fatalf("CallArgumentTypeInjector.Inject did not toggle the LAST marked declaration: %q", typeFirst)
	}
}

// markerAbsentInput builds, for each of the five injectors, an input shaped
// like an eligible one but with its marker stripped out -- the "target
// vanished" scenario TestEveryInjectorRefusesWhenMarkerDisappears drives
// from AllInjectors() so a future sixth injector without this treatment
// fails loudly instead of being silently skipped.
func markerAbsentInput(t *testing.T, name string) []byte {
	t.Helper()
	switch name {
	case "match":
		return bytes.ReplaceAll(phase6Fixture(t, "heldout_match_defect.schway"), []byte(matchTargetMarker), []byte(""))
	case "move":
		return bytes.ReplaceAll(phase6Fixture(t, "heldout_move_defect.schway"), []byte(moveTargetMarker), []byte(""))
	case "borrow":
		return bytes.ReplaceAll(phase6Fixture(t, "heldout_borrow_defect.schway"), []byte(borrowTargetMarker), []byte(""))
	case "cleanup":
		checked, err := CheckFile(testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.schway"))
		if err != nil {
			t.Fatal(err)
		}
		cSource, err := phase16InternalFrozenEvidenceC(checked.Program, "testdata/phase4/acquire_three_success.schway")
		if err != nil {
			t.Fatal(err)
		}
		return []byte(strings.ReplaceAll(cSource, releaseMarker, ""))
	case "stale_evidence":
		return bytes.ReplaceAll(phase6Fixture(t, "stale_evidence_subject.schway"), []byte(evidenceSubjectMarker), []byte(""))
	case "interprocedural_loan":
		return bytes.ReplaceAll(phase13InjectorLoanBase, []byte(loanTargetMarker), []byte(""))
	case "fallible_consume":
		return bytes.ReplaceAll(phase13InjectorConsumeBase, []byte(fallibleConsumeTargetMarker), []byte(""))
	case "call_argument_type":
		return bytes.ReplaceAll(phase13InjectorArgumentBase, []byte(callArgumentTargetMarker), []byte(""))
	default:
		t.Fatalf("markerAbsentInput: unhandled injector %q -- add a case here (this IS the test AllInjectors() drives)", name)
		return nil
	}
}

// TestEveryInjectorRefusesWhenMarkerDisappears is D-06-27.3's central
// assertion, driven from AllInjectors() rather than a hand-written list:
// every injector, given an input shaped like an eligible one but with its
// target marker stripped, refuses with InjectorTargetMissingCode naming
// both the found (0) and required (>=1) counts, and returns no output.
func TestEveryInjectorRefusesWhenMarkerDisappears(t *testing.T) {
	for _, injector := range AllInjectors() {
		injector := injector
		t.Run(injector.Name(), func(t *testing.T) {
			input := markerAbsentInput(t, injector.Name())
			mutated, err := injector.Inject(input)
			typed := injectorError(err)
			if typed == nil || typed.Code != InjectorTargetMissingCode {
				t.Fatalf("%s: did not refuse with %s when its target vanished: mutated=%q err=%v", injector.Name(), InjectorTargetMissingCode, mutated, err)
			}
			if mutated != nil {
				t.Fatalf("%s: refusal still returned a non-nil mutated source", injector.Name())
			}
			if !strings.Contains(typed.Error(), "0") {
				t.Fatalf("%s: refusal message %q does not name the found count", injector.Name(), typed.Error())
			}
		})
	}
}

// TestInjectorMarkerCountGuardIsNotInert demonstrates concretely what
// MatchInjector.Inject would do WITHOUT markerGuard: matchInjectSkippingGuard,
// its guard-disabled twin, silently returns an unmutated source that checks
// clean when the marker is absent -- exactly the theatre the guard exists
// to prevent -- while the guarded path refuses.
func TestInjectorMarkerCountGuardIsNotInert(t *testing.T) {
	source := bytes.ReplaceAll(phase6Fixture(t, "heldout_match_defect.schway"), []byte(matchTargetMarker), []byte(""))

	unguarded := matchInjectSkippingGuard(source)
	if !bytes.Equal(unguarded, source) {
		t.Fatalf("matchInjectSkippingGuard mutated a marker-absent source; it should have returned it unchanged")
	}
	uncheckedResult := Check(unguarded)
	if len(uncheckedResult.Diagnostics) != 0 {
		t.Fatalf("guard-disabled path's unmutated source did not check clean: %+v", uncheckedResult.Diagnostics)
	}

	_, err := MatchInjector{}.Inject(source)
	typed := injectorError(err)
	if typed == nil || typed.Code != InjectorTargetMissingCode {
		t.Fatalf("guarded MatchInjector.Inject did not refuse on the same input the guard-disabled path silently accepted: %v", err)
	}
}

// TestInjectorRefusalPropagatesToExerciseFailure drives MatchInjector
// through RunDefectInjectionExercise on a marker-absent input and asserts
// the reported Status is "fail", not merely that Inject returned an error
// -- the refusal must propagate all the way to the exercise's own result,
// not stop at the injector boundary.
func TestInjectorRefusalPropagatesToExerciseFailure(t *testing.T) {
	source := bytes.ReplaceAll(phase6Fixture(t, "heldout_match_defect.schway"), []byte(matchTargetMarker), []byte(""))
	result := RunDefectInjectionExercise(MatchInjector{}, source)
	if result.Status != "fail" {
		t.Fatalf("RunDefectInjectionExercise on a vanished target reported status %q, want \"fail\"", result.Status)
	}
	if result.Err == nil {
		t.Fatal("RunDefectInjectionExercise reported failure with no error attached")
	}

	// The eligible, correctly-marked fixture is the control: it must pass.
	eligible := phase6Fixture(t, "heldout_match_defect.schway")
	passResult := RunDefectInjectionExercise(MatchInjector{}, eligible)
	if passResult.Status != "pass" {
		t.Fatalf("RunDefectInjectionExercise on an eligible fixture reported status %q, want \"pass\": err=%v", passResult.Status, passResult.Err)
	}
}
