package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/evidence"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// This file is a _test.go file, so it is explicitly EXEMPT from the
// import-boundary lint (import_boundary_test.go scans only non-test files):
// it imports internal/ packages to DRIVE the fixtures and injectors the
// real driver (repair.go, main.go) never touches, exactly the way
// session_phase6_injectors_test.go already imports internal packages to
// drive AllInjectors() while the injectors themselves stay production code.

func phase6Fixture(t testing.TB, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase6", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mustWriteFile(t testing.TB, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func decodeCheckJSON(t testing.TB, raw []byte) checkResult {
	t.Helper()
	var decoded checkResult
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decoding lang --json check output: %v (raw: %s)", err, raw)
	}
	return decoded
}

// repairPackageDir is cmd/lang-repair's own directory, resolved the same
// way testsupport.ProjectPath resolves every other fixture and package
// path in the tree. This file is itself a _test.go file, so importing
// internal/compiler/testsupport here is allowed -- the import boundary
// (import_boundary_test.go) applies only to cmd/lang-repair's non-test
// files (main.go, repair.go), never to the tests that scan them.
func repairPackageDir(t testing.TB) string {
	t.Helper()
	return testsupport.ProjectPath("cmd", "lang-repair")
}

// forEachNonTestFile walks every non-test .go file directly inside dir,
// fully parsed (not ImportsOnly), calling visit on each parsed file. Shared
// by every go/ast structural test in this package
// (TestRepairDriverDecodesNoProseFields,
// TestRepairDriverSourceNeverReferencesHeldoutFixtures, and
// import_boundary_test.go's own boundary tests).
func forEachNonTestFile(t testing.TB, dir string, visit func(name string, fset *token.FileSet, file *ast.File)) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		visit(entry.Name(), fset, file)
	}
}

// TestRepairDriverFixesOneDefectEndToEnd is the tracer (Task 1): one
// heldout_move_*.lang fixture, mutated by the REAL move injector, repaired
// by the REAL driver spawning the REAL built `lang` binary as a subprocess,
// and independently re-verified clean by a separate `lang --json check`
// invocation. Real binary, real subprocess, real JSON, one class, one path.
func TestRepairDriverFixesOneDefectEndToEnd(t *testing.T) {
	langBinary := testsupport.BuildCLI(t)

	original := phase6Fixture(t, "heldout_move_defect.lang")
	mutated, err := session.MoveInjector{}.Inject(original)
	if err != nil {
		t.Fatalf("injecting move defect: %v", err)
	}

	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "heldout_move_defect.lang")
	mustWriteFile(t, sourcePath, mutated)

	outcome, err := Repair(context.Background(), langBinary, sourcePath)
	if err != nil {
		t.Fatalf("Repair: %v", err)
	}
	if outcome.Status != OutcomeRepaired {
		t.Fatalf("got status %q, want %q (diagnosis=%q repair=%q)", outcome.Status, OutcomeRepaired, outcome.DiagnosisCode, outcome.RepairKind)
	}

	repaired, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(repaired, original) {
		t.Fatal("repaired bytes are not byte-identical to the pre-defect original")
	}

	// Independent re-verification: a SECOND, wholly separate `lang --json
	// check` subprocess invocation, outside the driver's own internal
	// reverify, reports the post-repair source clean.
	verify := testsupport.RunCLI(t, langBinary, nil, "--json", "check", sourcePath)
	decoded := decodeCheckJSON(t, verify.Stdout)
	if decoded.Status != statusPass {
		t.Fatalf("independent re-verification reports status %q, want %q", decoded.Status, statusPass)
	}
}

// TestRepairDriverDecodesNoProseFields asserts, by go/ast, that no struct
// declared in cmd/lang-repair binds either prose field of the diagnostic
// document -- the human-readable "message", or any cause/repair "detail" --
// to a JSON tag. A field that is never decoded cannot be scraped
// (T-06-BOUNDARY-03).
func TestRepairDriverDecodesNoProseFields(t *testing.T) {
	forEachNonTestFile(t, repairPackageDir(t), func(name string, fset *token.FileSet, file *ast.File) {
		ast.Inspect(file, func(n ast.Node) bool {
			st, isStruct := n.(*ast.StructType)
			if !isStruct {
				return true
			}
			for _, field := range st.Fields.List {
				if field.Tag == nil {
					continue
				}
				tag := strings.Trim(field.Tag.Value, "`")
				if strings.Contains(tag, `json:"message`) || strings.Contains(tag, `json:"detail`) {
					t.Fatalf("%s:%s declares a struct field bound to a prose JSON field (%s) -- the driver may only decode structural fields",
						name, fset.Position(field.Pos()), tag)
				}
			}
			return true
		})
	})
}

// TestRepairDriverBoundedReaderRejectsOversizedStdout demonstrates the
// bounded-writer half of T-06-BOUNDARY-05 directly: a probe emitting more
// than the declared cap is observably overflowed rather than silently
// accumulated without limit.
func TestRepairDriverBoundedReaderRejectsOversizedStdout(t *testing.T) {
	writer := newBoundedWriter(8)
	oversized := bytes.Repeat([]byte("a"), 16)
	if _, err := writer.Write(oversized); err != nil {
		t.Fatalf("Write must never itself error (the subprocess must not see a write failure): %v", err)
	}
	if !writer.overflowed() {
		t.Fatal("expected the writer to report overflow for a stream larger than its declared cap")
	}

	// Boundary: exactly at the cap must NOT overflow.
	exact := newBoundedWriter(8)
	if _, err := exact.Write(bytes.Repeat([]byte("a"), 8)); err != nil {
		t.Fatal(err)
	}
	if exact.overflowed() {
		t.Fatal("stdout exactly at the cap must not be reported as overflowed")
	}
}

// TestRepairDriverSourceNeverReferencesHeldoutFixtures asserts, by go/ast,
// that no non-test file under cmd/lang-repair names any heldout_ fixture
// path -- the driver's kind-to-edit mapping (README's rule) may consult
// only derivation_ fixtures, if it consults any fixture at all. The
// shipped driver consults none (it is fully generic over Span/Replacement),
// which trivially satisfies this, but the test still stands as a live
// falsifier against a future regression that hardcodes a heldout_ path.
func TestRepairDriverSourceNeverReferencesHeldoutFixtures(t *testing.T) {
	forEachNonTestFile(t, repairPackageDir(t), func(name string, fset *token.FileSet, file *ast.File) {
		ast.Inspect(file, func(n ast.Node) bool {
			lit, isLit := n.(*ast.BasicLit)
			if !isLit || lit.Kind != token.STRING {
				return true
			}
			if strings.Contains(lit.Value, "heldout_") {
				t.Fatalf("%s:%s references a heldout_ fixture path %s -- the driver's mapping may consult only derivation_ fixtures",
					name, fset.Position(lit.Pos()), lit.Value)
			}
			return true
		})
	})
}

// testSourceClassRepair drives one source-granularity defect class (match,
// move, borrow) through the real driver against a real built `lang`
// binary, and applies the full D-06-26 success oracle: clean check AND
// byte-identity with the pre-defect original.
func testSourceClassRepair(t *testing.T, langBinary string, injector session.Injector, fixture string) {
	t.Helper()
	original := phase6Fixture(t, fixture)
	mutated, err := injector.Inject(original)
	if err != nil {
		t.Fatalf("%s: injecting defect: %v", injector.Name(), err)
	}

	dir := t.TempDir()
	sourcePath := filepath.Join(dir, fixture)
	mustWriteFile(t, sourcePath, mutated)

	outcome, err := Repair(context.Background(), langBinary, sourcePath)
	if err != nil {
		t.Fatalf("%s: Repair: %v", injector.Name(), err)
	}
	if outcome.Status != OutcomeRepaired {
		t.Fatalf("%s: got status %q, want %q", injector.Name(), outcome.Status, OutcomeRepaired)
	}
	// The gate is single-pass, bounded, and O(1) per class: exactly one
	// diagnose call and one reverify call, never a loop.
	if outcome.SubprocessCount != 2 {
		t.Fatalf("%s: subprocess invocation count is %d, want fixed at 2 (single pass, no repair loop)", injector.Name(), outcome.SubprocessCount)
	}

	repaired, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	// The injector's own out-of-band `// lang:*-target` marker comment is a
	// testdata-authoring device (README), never part of the program the
	// checker's repair reconstructs -- match/move delete or corrupt the
	// WHOLE marked line, so a correct repair cannot be expected to restore
	// a comment the checker never even knew existed. Strip it from BOTH
	// sides of the comparison basis so the oracle compares the actual
	// program text, not testdata bookkeeping.
	if !oracleAccepts(stripInjectorMarker(repaired), stripInjectorMarker(original), true) {
		t.Fatalf("%s: repaired bytes are not byte-identical to the pre-defect original", injector.Name())
	}

	verify := testsupport.RunCLI(t, langBinary, nil, "--json", "check", sourcePath)
	decoded := decodeCheckJSON(t, verify.Stdout)
	if decoded.Status != statusPass {
		t.Fatalf("%s: independent re-verification reports status %q, want %q", injector.Name(), decoded.Status, statusPass)
	}
}

// oracleAccepts is D-06-26's success oracle for the four source-granularity
// classes, made an explicit function so TestRepairOracleRejectsDeleteTheCode
// can demonstrate it rejecting the degenerate case: byte-identity ALONE is
// not enough and a clean check ALONE is not enough -- both are required, so
// a repair that merely satisfies the checker without reversing the
// injected defect is rejected.
func oracleAccepts(repaired, original []byte, cleanCheck bool) bool {
	return cleanCheck && bytes.Equal(repaired, original)
}

// injectorMarkerSuffixes are the trailing testdata-authoring markers
// (testdata/phase6/README) MatchInjector/MoveInjector/BorrowInjector locate
// their target line by. Each is stripped, on both sides of a byte-identity
// comparison, because these markers are out-of-band scoring bookkeeping,
// never program text a repair is expected to reproduce.
var injectorMarkerSuffixes = []string{
	" // lang:match-target",
	" // lang:move-target",
	" // lang:borrow-target",
}

func stripInjectorMarker(source []byte) []byte {
	text := string(source)
	for _, marker := range injectorMarkerSuffixes {
		text = strings.ReplaceAll(text, marker, "")
	}
	return []byte(text)
}

// testCleanupClassRepair covers the cleanup class, whose defect (a release
// omission) lives entirely in generated C the compiler produces AFTER
// `lang check` runs -- there is no `.lang`-level diagnostic for it at all,
// so it cannot go through the driver's check-repair JSON protocol the way
// match/move/borrow do. Its class-specific mechanism (mirroring how
// stale-evidence gets its own distinct oracle in this same task): first
// prove the injected defect is REAL by observing a genuine resource leak
// during native execution, then repair it the only mechanically sound way
// available for an artifact the driver's protocol does not govern --
// re-deriving the generated C exactly once, fresh from the untouched
// `.lang` source, rather than patching the mutated artifact. The 06-14
// differential-behavior gate covers the same class from the mismatch side.
func testCleanupClassRepair(t *testing.T) {
	t.Helper()
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.lang"))
	if err != nil {
		t.Fatal(err)
	}
	original, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	mutated, err := session.CleanupInjector{}.Inject([]byte(original))
	if err != nil {
		t.Fatalf("injecting cleanup defect: %v", err)
	}

	runner := native.DefaultRunner()
	runner.ForeignSources = []string{native.ForeignResourceSourcePath()}

	// A release omission leaves a resource live at return. native.Runner's
	// own ExpectValue contract (validateExecution) requires LiveResources
	// to be EMPTY on a normal return, so the mutated C's execution document
	// violates that contract and Run reports a typed
	// native.invalid_execution error -- that error IS the leak's
	// observable signature at this API surface. Proving the injected
	// defect is real means proving Run rejects it this way, not that it
	// silently succeeds with a leak nobody checked for.
	_, mutatedErr := runner.Run(context.Background(), string(mutated), "-O0", []string{"7"})
	var toolErr *native.ToolError
	if !errors.As(mutatedErr, &toolErr) || toolErr.Code != "native.invalid_execution" {
		t.Fatalf("mutated C did not fail the leak-detecting execution contract as expected: %v", mutatedErr)
	}

	// Single-pass repair: re-derive the generated C exactly once, fresh
	// from the untouched `.lang` source.
	repaired, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if repaired != original {
		t.Fatal("recompiled C is not byte-identical to the pre-defect original")
	}
	repairedResult, err := runner.Run(context.Background(), repaired, "-O0", []string{"7"})
	if err != nil {
		t.Fatalf("repaired C failed to compile/run: %v", err)
	}
	if len(repairedResult.Pairs) != 1 || len(repairedResult.Pairs[0].Execution.LiveResources) != 0 {
		t.Fatalf("repaired C still reports live (leaked) resources: %+v", repairedResult)
	}
}

// staleEvidenceRepair is the shared result testStaleEvidenceClassRepair and
// TestStaleEvidenceRepairRebindsManifest both build on.
type staleEvidenceRepair struct {
	sourcePath      string
	manifestPath    string
	subprocessCount int
}

// repairStaleEvidenceClass drives the stale-evidence class end to end
// through the shipped `lang` binary's own evidence commands only (`lang
// evidence FILE` and `lang evidence --validate MANIFEST FILE`) -- never a
// source diff. D-06-26: stale evidence is a manifest whose SHA-256 no
// longer binds its source, repaired by recapturing a manifest that binds.
func repairStaleEvidenceClass(t *testing.T, langBinary string) staleEvidenceRepair {
	t.Helper()
	dir := t.TempDir()
	original := phase6Fixture(t, "stale_evidence_subject.lang")
	sourcePath := filepath.Join(dir, "stale_evidence_subject.lang")
	mustWriteFile(t, sourcePath, original)
	manifestPath := filepath.Join(dir, "manifest.json")

	invocations := 0
	capture := func() {
		invocations++
		result := testsupport.RunCLI(t, langBinary, nil, "evidence", sourcePath)
		if result.Exit != 0 {
			t.Fatalf("lang evidence failed: exit=%d stderr=%s", result.Exit, result.Stderr)
		}
		mustWriteFile(t, manifestPath, result.Stdout)
	}
	validateStatus := func() string {
		invocations++
		result := testsupport.RunCLI(t, langBinary, nil, "--json", "evidence", "--validate", manifestPath, sourcePath)
		var decoded struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(result.Stdout, &decoded); err != nil {
			t.Fatalf("decoding evidence validation result: %v (raw: %s)", err, result.Stdout)
		}
		return decoded.Status
	}

	capture()
	if status := validateStatus(); status != "pass" {
		t.Fatalf("pre-defect manifest failed to validate: %s", status)
	}

	mutated, err := session.StaleEvidenceInjector{}.Inject(original)
	if err != nil {
		t.Fatalf("injecting stale-evidence defect: %v", err)
	}
	mustWriteFile(t, sourcePath, mutated)

	if status := validateStatus(); status == "pass" {
		t.Fatal("mutated source unexpectedly validated against the stale manifest -- the injected defect was not observed")
	}

	// Single-pass repair: recapture the manifest against the retouched
	// source exactly once.
	capture()
	if status := validateStatus(); status != "pass" {
		t.Fatalf("recaptured manifest failed to validate: %s", status)
	}

	return staleEvidenceRepair{sourcePath: sourcePath, manifestPath: manifestPath, subprocessCount: invocations}
}

func testStaleEvidenceClassRepair(t *testing.T, langBinary string) {
	t.Helper()
	repaired := repairStaleEvidenceClass(t, langBinary)
	if repaired.subprocessCount != 5 {
		t.Fatalf("subprocess invocation count is %d, want fixed at 5 (single pass, no repair loop)", repaired.subprocessCount)
	}
}

// TestRepairDriverFixesEveryDefectClassSinglePass has one subtest per
// class, proving the supported command protocol is mechanically sufficient
// to fix representative match, move, borrow, cleanup, and stale-evidence
// defects (DX-04's SC3), each bounded and O(1) -- no repair loop anywhere.
func TestRepairDriverFixesEveryDefectClassSinglePass(t *testing.T) {
	langBinary := testsupport.BuildCLI(t)

	t.Run("match", func(t *testing.T) {
		testSourceClassRepair(t, langBinary, session.MatchInjector{}, "heldout_match_defect.lang")
	})
	t.Run("move", func(t *testing.T) {
		testSourceClassRepair(t, langBinary, session.MoveInjector{}, "heldout_move_defect.lang")
	})
	t.Run("borrow", func(t *testing.T) {
		testSourceClassRepair(t, langBinary, session.BorrowInjector{}, "heldout_borrow_defect.lang")
	})
	t.Run("cleanup", func(t *testing.T) {
		testCleanupClassRepair(t)
	})
	t.Run("stale_evidence", func(t *testing.T) {
		testStaleEvidenceClassRepair(t, langBinary)
	})
}

// TestRepairOracleRejectsDeleteTheCode is D-06-26's positive control: a
// repair that makes `lang check` clean WITHOUT reversing the injected
// defect (here: filling the missing match arm with a DIFFERENT, still-
// valid mapping instead of the original self-mapping one) must still be
// REJECTED by the oracle. Without this test the oracle could be satisfied
// by exactly the overfitting failure mode the automated-program-repair
// literature names, and by this project's own three-gate-failure lesson.
func TestRepairOracleRejectsDeleteTheCode(t *testing.T) {
	langBinary := testsupport.BuildCLI(t)

	original := phase6Fixture(t, "heldout_match_defect.lang")
	mutated, err := session.MatchInjector{}.Inject(original)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "heldout_match_defect.lang")
	mustWriteFile(t, sourcePath, mutated)

	diagnosis := testsupport.RunCLI(t, langBinary, nil, "--json", "check", sourcePath)
	decoded := decodeCheckJSON(t, diagnosis.Stdout)
	repair, _, ok := selectRepair(decoded)
	if !ok {
		t.Fatal("expected a driver-eligible repair for the injected match defect")
	}

	// The DEGENERATE variant: splice in a DIFFERENT, still-exhaustive arm
	// mapping ("Blue => Red") instead of reversing the injected defect
	// ("Blue => Blue") -- a plausible, clean-passing, WRONG program.
	degenerate := jsonRepair{
		Kind: repair.Kind, Span: repair.Span,
		Replacement:   "\n    Blue => Red",
		Applicability: repair.Applicability,
	}
	if err := applyRepair(sourcePath, degenerate); err != nil {
		t.Fatal(err)
	}

	verify := testsupport.RunCLI(t, langBinary, nil, "--json", "check", sourcePath)
	verifyDecoded := decodeCheckJSON(t, verify.Stdout)
	if verifyDecoded.Status != statusPass {
		t.Fatalf("test construction error: degenerate repair did not even pass check: %s", verifyDecoded.Status)
	}

	repaired, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(repaired, original) {
		t.Fatal("test construction error: degenerate repair should not be byte-identical to the original")
	}
	if oracleAccepts(repaired, original, verifyDecoded.Status == statusPass) {
		t.Fatal("oracle accepted the degenerate delete-the-offending-code repair (clean check but NOT a reversal of the injected defect)")
	}
}

// TestStaleEvidenceRepairRebindsManifest asserts the recaptured manifest's
// bound digest equals the digest of the `lang format`-canonical source, not
// merely that validation returned pass (D-06-26).
func TestStaleEvidenceRepairRebindsManifest(t *testing.T) {
	langBinary := testsupport.BuildCLI(t)
	repaired := repairStaleEvidenceClass(t, langBinary)

	manifestBytes, err := os.ReadFile(repaired.manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := evidence.DecodeStrict(manifestBytes)
	if err != nil {
		t.Fatal(err)
	}

	canonical := testsupport.RunCLI(t, langBinary, nil, "format", repaired.sourcePath)
	if canonical.Exit != 0 {
		t.Fatalf("lang format failed: exit=%d stderr=%s", canonical.Exit, canonical.Stderr)
	}
	want := evidence.ContentDigest(canonical.Stdout)
	if manifest.SourceDigest != want {
		t.Fatalf("recaptured manifest's bound digest %q does not match the current lang format-canonical source digest %q", manifest.SourceDigest, want)
	}
}

// TestUnrepairableDefectFailsTheGate: a defect for which the binary emits
// zero driver-eligible repairs is reported unrepairable, never skipped and
// never counted as a pass (FND-04 empty-input edge). testdata/phase1's own
// non_exhaustive.lang is exactly this case -- its arms don't self-map
// (Off => On), so check.go attaches no repair at all, per the comment in
// check.go's analyzeMatchArms.
func TestUnrepairableDefectFailsTheGate(t *testing.T) {
	langBinary := testsupport.BuildCLI(t)
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", "non_exhaustive.lang"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "non_exhaustive.lang")
	mustWriteFile(t, sourcePath, source)

	outcome, err := Repair(context.Background(), langBinary, sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Status != OutcomeUnrepairable {
		t.Fatalf("got status %q, want %q -- a defect with zero driver-eligible repairs must never be silently skipped or counted as a pass", outcome.Status, OutcomeUnrepairable)
	}
	if outcome.SubprocessCount != 1 {
		t.Fatalf("expected exactly 1 subprocess invocation for an unrepairable defect, got %d", outcome.SubprocessCount)
	}
}

// TestRepairSelectionIsSpecifiedOnTies: when a diagnostic carries two
// equally applicable (MachineApplicable) repairs, the driver's choice is
// specified (first eligible repair in document order) and stable across
// runs (FND-04 ordering edge). The shipped compiler never actually emits
// two MachineApplicable repairs on one diagnostic today, so this
// constructs the tie synthetically to exercise selectRepair's own rule
// directly.
func TestRepairSelectionIsSpecifiedOnTies(t *testing.T) {
	result := checkResult{
		Status: "invalid",
		Diagnostics: []jsonDiagnostic{
			{
				Code: "synthetic.tie",
				Repairs: []jsonRepair{
					{Kind: "second_choice", Span: &jsonSpan{Start: 5, End: 5}, Replacement: "b", Applicability: applicabilityMachineApplicable},
					{Kind: "first_choice", Span: &jsonSpan{Start: 0, End: 0}, Replacement: "a", Applicability: applicabilityMachineApplicable},
				},
			},
		},
	}
	first, _, ok := selectRepair(result)
	if !ok {
		t.Fatal("expected a driver-eligible repair")
	}
	for i := 0; i < 10; i++ {
		got, _, ok := selectRepair(result)
		if !ok || got != first {
			t.Fatalf("selection is not stable across repeated invocations: got %+v, want %+v", got, first)
		}
	}
	if first.Kind != "second_choice" {
		t.Fatalf("selection rule is document order (first eligible repair encountered); got %q, want %q", first.Kind, "second_choice")
	}
}
