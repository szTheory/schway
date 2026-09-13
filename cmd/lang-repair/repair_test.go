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

	// 13-06 Task 2 (D-13-29): the three interprocedural classes, each
	// driven from a SEALED held-out fixture (never the derivation fixture
	// the repair was hand-tuned against). interprocedural_loan reuses
	// D-13-28's twin pair alpha half -- already exercised end-to-end by
	// TestTwinPairBlame, but registered here too as this project's own
	// established per-class registration point (13-ANTITHEATER-CONTRACT.md).
	//
	// call_argument_type is INTENTIONALLY excluded from this repaired-only
	// list: 13-06 (D-13-10a) empirically found use_matching_argument's
	// Replacement is a byte-identical no-op on every real trigger, so
	// check.go no longer emits it, and the class's honest driver outcome is
	// `unrepairable` -- asserted separately by
	// TestHeldoutOutcomeSetContainsNoLaundering and
	// TestUnrepairableDefectFailsTheGate, never claimed as `repaired` here.
	t.Run("interprocedural_loan", func(t *testing.T) {
		testHeldoutInterproceduralClassRepair(t, langBinary, "heldout_shared_callee_twin_alpha.lang", session.InterproceduralLoanInjector{}, "check.interprocedural_loan_liveness", "move_after_interprocedural_loan")
	})
	t.Run("fallible_consume", func(t *testing.T) {
		testHeldoutInterproceduralClassRepair(t, langBinary, "heldout_fallible_call_unconsumed.lang", session.FallibleConsumeInjector{}, "syntax.fallible_call_not_consumed", "wrap_call_in_try")
	})
}

// testHeldoutInterproceduralClassRepair drives one testdata/phase13
// held-out fixture, mutated by the REAL injector, through the real driver
// against a real built `lang` binary, and asserts the EXACT outcome
// string `repaired` (D-13-29 -- never a substring or boolean), the exact
// diagnosis code, the exact repair kind, and SubprocessCount 2 (single
// pass). Unlike testSourceClassRepair (phase6), this does NOT require
// byte-identity with a pre-defect original: an interprocedural repair may
// legitimately produce a different-but-equally-clean-checking program
// (D-13-11's semantics-preserving argument for the swap class covers
// exactly this).
func testHeldoutInterproceduralClassRepair(t *testing.T, langBinary string, fixture string, injector session.Injector, wantCode, wantKind string) {
	t.Helper()
	fixturePath := testsupport.ProjectPath("testdata", "phase13", fixture)
	original, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	mutated, err := injector.Inject(original)
	if err != nil {
		t.Fatalf("%s: injecting: %v", injector.Name(), err)
	}
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, fixture)
	mustWriteFile(t, sourcePath, mutated)

	outcome, err := Repair(context.Background(), langBinary, sourcePath)
	if err != nil {
		t.Fatalf("%s: Repair: %v", injector.Name(), err)
	}
	if outcome.Status != OutcomeRepaired {
		t.Fatalf("%s: got status %q, want exactly %q (diagnosis=%q repair=%q)", injector.Name(), outcome.Status, OutcomeRepaired, outcome.DiagnosisCode, outcome.RepairKind)
	}
	if outcome.DiagnosisCode != wantCode {
		t.Fatalf("%s: got diagnosis code %q, want %q", injector.Name(), outcome.DiagnosisCode, wantCode)
	}
	if outcome.RepairKind != wantKind {
		t.Fatalf("%s: got repair kind %q, want %q", injector.Name(), outcome.RepairKind, wantKind)
	}
	if outcome.SubprocessCount != 2 {
		t.Fatalf("%s: got subprocess count %d, want 2 (single pass: one diagnose, one reverify)", injector.Name(), outcome.SubprocessCount)
	}

	verify := testsupport.RunCLI(t, langBinary, nil, "--json", "check", sourcePath)
	decoded := decodeCheckJSON(t, verify.Stdout)
	if decoded.Status != statusPass {
		t.Fatalf("%s: independent re-verification reports status %q, want %q", injector.Name(), decoded.Status, statusPass)
	}
	if len(decoded.Diagnostics) != 0 {
		t.Fatalf("%s: independent re-verification reports %d diagnostics, want 0: %+v", injector.Name(), len(decoded.Diagnostics), decoded.Diagnostics)
	}

	afterTest, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterTest, original) {
		t.Fatalf("%s: %s was modified by this test -- the driver must operate on the TempDir copy only", injector.Name(), fixture)
	}
}

// testHeldoutInterproceduralClassUnrepairable drives one testdata/phase13
// held-out fixture through the real driver and asserts the EXACT outcome
// string `unrepairable` with SubprocessCount 1 -- a defect for which
// check.go emits zero driver-eligible repairs is reported unrepairable,
// never skipped and never counted as a pass (D-13-29, FND-04).
func testHeldoutInterproceduralClassUnrepairable(t *testing.T, langBinary string, fixture string, injector session.Injector, wantCode string) {
	t.Helper()
	fixturePath := testsupport.ProjectPath("testdata", "phase13", fixture)
	original, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	mutated, err := injector.Inject(original)
	if err != nil {
		t.Fatalf("%s: injecting: %v", injector.Name(), err)
	}
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, fixture)
	mustWriteFile(t, sourcePath, mutated)

	outcome, err := Repair(context.Background(), langBinary, sourcePath)
	if err != nil {
		t.Fatalf("%s: Repair: %v", injector.Name(), err)
	}
	if outcome.Status != OutcomeUnrepairable {
		t.Fatalf("%s: got status %q, want exactly %q", injector.Name(), outcome.Status, OutcomeUnrepairable)
	}
	if outcome.DiagnosisCode != wantCode {
		t.Fatalf("%s: got diagnosis code %q, want %q", injector.Name(), outcome.DiagnosisCode, wantCode)
	}
	if outcome.SubprocessCount != 1 {
		t.Fatalf("%s: got subprocess count %d, want 1", injector.Name(), outcome.SubprocessCount)
	}

	afterTest, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterTest, original) {
		t.Fatalf("%s: %s was modified by this test -- the driver must operate on the TempDir copy only", injector.Name(), fixture)
	}
}

// heldoutInterproceduralCase is one row of
// TestHeldoutOutcomeSetContainsNoLaundering's corpus-wide sweep.
type heldoutInterproceduralCase struct {
	name             string
	fixture          string
	injector         session.Injector
	claimedAsReached bool
}

// TestHeldoutOutcomeSetContainsNoLaundering is 13-06 Task 2's Test 2
// (D-13-29): collects the outcome string from EVERY held-out
// interprocedural case in one run and asserts the observed set contains
// no `already_clean` (an injector going inert) anywhere, and no
// `unrepairable` for either of the two classes claimed as genuinely
// reached (interprocedural_loan, fallible_consume).
//
// call_argument_type is EXCLUDED from the claimed-as-reached set BY NAME,
// not a catch-all -- 13-06 (D-13-10a) found use_matching_argument's
// Replacement is a byte-identical no-op on every real trigger (see this
// file's TestTwinPairBlame doc comment and 13-06-SUMMARY.md's "D-13-10a
// adjudication"). check.go no longer emits it on any partition, so BOTH
// its held-out cases -- the exact-one-match
// heldout_call_argument_mismatch.lang and the ambiguous-two-match
// heldout_call_argument_ambiguous.lang -- are EXPECTED to report
// unrepairable, and asserted as such here: the honest outcome for a class
// with no span-local fix at this language's current maturity, never
// laundered into a claimed pass.
func TestHeldoutOutcomeSetContainsNoLaundering(t *testing.T) {
	langBinary := testsupport.BuildCLI(t)

	cases := []heldoutInterproceduralCase{
		{"interprocedural_loan_alpha", "heldout_shared_callee_twin_alpha.lang", session.InterproceduralLoanInjector{}, true},
		{"interprocedural_loan_mirror", "heldout_shared_callee_twin_mirror.lang", session.InterproceduralLoanInjector{}, true},
		{"fallible_consume", "heldout_fallible_call_unconsumed.lang", session.FallibleConsumeInjector{}, true},
		{"call_argument_type_one_match", "heldout_call_argument_mismatch.lang", session.CallArgumentTypeInjector{}, false},
		{"call_argument_type_ambiguous", "heldout_call_argument_ambiguous.lang", session.CallArgumentTypeInjector{}, false},
	}

	observed := make(map[string]bool)
	for _, c := range cases {
		fixturePath := testsupport.ProjectPath("testdata", "phase13", c.fixture)
		original, err := os.ReadFile(fixturePath)
		if err != nil {
			t.Fatal(err)
		}
		mutated, err := c.injector.Inject(original)
		if err != nil {
			t.Fatalf("%s: injecting: %v", c.name, err)
		}
		dir := t.TempDir()
		sourcePath := filepath.Join(dir, c.fixture)
		mustWriteFile(t, sourcePath, mutated)

		outcome, err := Repair(context.Background(), langBinary, sourcePath)
		if err != nil {
			t.Fatalf("%s: Repair: %v", c.name, err)
		}
		observed[outcome.Status] = true

		if outcome.Status == OutcomeAlreadyClean {
			t.Fatalf("%s: outcome was already_clean -- the injector went inert", c.name)
		}
		if c.claimedAsReached && outcome.Status != OutcomeRepaired {
			t.Fatalf("%s: claimed-as-reached class reported %q, want %q", c.name, outcome.Status, OutcomeRepaired)
		}
		if !c.claimedAsReached && outcome.Status != OutcomeUnrepairable {
			t.Fatalf("%s: expected %q for the class excluded from claimed-as-reached (D-13-10a), got %q", c.name, OutcomeUnrepairable, outcome.Status)
		}

		afterTest, err := os.ReadFile(fixturePath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(afterTest, original) {
			t.Fatalf("%s: %s was modified by this test -- the driver must operate on the TempDir copy only", c.name, c.fixture)
		}
	}

	if observed[OutcomeAlreadyClean] {
		t.Fatal("observed outcome set contains already_clean")
	}
}

// TestWrapCallInTryReverifyFailed is 13-06 Task 2's Test 3: the honest
// negative for wrap_call_in_try. check.go's checkFallibleLinear
// (check.go:3391) only accepts two body shapes for a function containing a
// try_call: a single try binding immediately returned, or a sequence where
// EVERY binding is itself fallible and the result is the function's own
// parameter. Wrapping a call in `try` inside a function whose body has an
// ordinary, non-fallible binding ahead of the fallible call satisfies
// NEITHER shape once re-parsed, so the diagnosis TRADES one diagnostic
// (syntax.fallible_call_not_consumed) for a DIFFERENT one
// (check.foreign_call_shape_unsupported) rather than clearing to zero --
// the driver's single-pass structure means this is reported as failure,
// never looped. Constructed inline (not from the sealed held-out corpus:
// this is a structural negative, not a scored held-out case) and verified
// directly against the real driver before writing these assertions.
func TestWrapCallInTryReverifyFailed(t *testing.T) {
	langBinary := testsupport.BuildCLI(t)

	const source = `module phase13.wrap_call_in_try_reverify_failed

export {
  fn main
}

foreign C {

  fn lang_res_open(request: Byte) -> Byte {
    unwind: forbidden
    nonlocal_exit: forbidden
    allocator: "libc_malloc"
    fails: AcquireError
  }
}

data AcquireError =
  | OpenFailed

fn acquire(request: Byte) -> Byte {
  let extra = request
  let handle = lang_res_open(request)
  handle
}

fn main(request: Byte) -> Byte {
  let opened = acquire(request)
  opened
}
`
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "wrap_call_in_try_reverify_failed.lang")
	mustWriteFile(t, sourcePath, []byte(source))

	// Construction check: the un-repaired source must diagnose exactly
	// syntax.fallible_call_not_consumed with a wrap_call_in_try repair,
	// confirming the enclosing-shape trap is reached via the SAME path
	// every other wrap_call_in_try case takes, not a different one.
	diagnosis := testsupport.RunCLI(t, langBinary, nil, "--json", "check", sourcePath)
	decoded := decodeCheckJSON(t, diagnosis.Stdout)
	if len(decoded.Diagnostics) != 1 || decoded.Diagnostics[0].Code != "syntax.fallible_call_not_consumed" {
		t.Fatalf("test construction error: expected exactly one syntax.fallible_call_not_consumed diagnostic, got %+v", decoded.Diagnostics)
	}
	_, code, ok := selectRepair(decoded)
	if !ok || code != "syntax.fallible_call_not_consumed" {
		t.Fatalf("test construction error: expected a driver-eligible repair on syntax.fallible_call_not_consumed, got code=%q ok=%v", code, ok)
	}

	outcome, err := Repair(context.Background(), langBinary, sourcePath)
	if err != nil {
		t.Fatalf("Repair: %v", err)
	}
	if outcome.Status != OutcomeReverifyFailed {
		t.Fatalf("got status %q, want exactly %q (diagnosis=%q repair=%q)", outcome.Status, OutcomeReverifyFailed, outcome.DiagnosisCode, outcome.RepairKind)
	}
	if outcome.DiagnosisCode != "syntax.fallible_call_not_consumed" {
		t.Fatalf("got diagnosis code %q, want %q", outcome.DiagnosisCode, "syntax.fallible_call_not_consumed")
	}
	if outcome.RepairKind != "wrap_call_in_try" {
		t.Fatalf("got repair kind %q, want %q", outcome.RepairKind, "wrap_call_in_try")
	}
	if outcome.SubprocessCount != 2 {
		t.Fatalf("got subprocess count %d, want 2 (single pass: one diagnose, one reverify)", outcome.SubprocessCount)
	}

	// The traded-in diagnostic must genuinely be a DIFFERENT code, not the
	// same one persisting -- confirming this is the enclosing-shape trap,
	// not a no-op.
	verify := testsupport.RunCLI(t, langBinary, nil, "--json", "check", sourcePath)
	verifyDecoded := decodeCheckJSON(t, verify.Stdout)
	if verifyDecoded.Status == statusPass {
		t.Fatal("test construction error: post-repair source unexpectedly checks clean")
	}
	if len(verifyDecoded.Diagnostics) != 1 || verifyDecoded.Diagnostics[0].Code != "check.foreign_call_shape_unsupported" {
		t.Fatalf("expected exactly one check.foreign_call_shape_unsupported diagnostic after the repair, got %+v", verifyDecoded.Diagnostics)
	}
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

// TestRepairDriverFixesInterproceduralLoanLivenessSinglePass is 13-01
// Task 3's own regression, protecting the tracer path this plan's Task 1
// shipped: check.interprocedural_loan_liveness now carries a real,
// driver-eligible move_after_interprocedural_loan repair for the BACKWARD
// direction (D-08-08), and the untouched cmd/lang-repair driver applies it
// through the JSON protocol alone -- no kind-to-edit table, no source
// change to repair.go/main.go (D-13-32, mechanically asserted by this
// plan's <verify> blocks via `git diff --quiet`).
//
// D-13-13's own reasoning for reusing NO existing repair kind: this test
// asserts all four Outcome fields by EXACT string equality, never substring
// or prefix, so a future accidental reuse of move_after_last_borrow_use (or
// any other kind) fails loudly here rather than passing on a
// coincidental substring match.
//
// This plan's own 13-ANTITHEATER-CONTRACT.md (Task 2, "Obligations for
// plans 13-03, 13-05 and 13-06", item 6) records that registering a new
// defect class's coverage in antitheater_test.go's own hardcoded class
// lists is the one MANDATORY per-class obligation there; this repair_test.go
// function is the analogous registration for
// TestRepairDriverFixesEveryDefectClassSinglePass's OWN class list
// (repair_test.go:430-447) -- that list stays untouched by this plan
// because it drives session.Injector-mutated testdata/phase6/heldout_*
// fixtures, a shape this plan's real testdata/phase13/ source fixture does
// not share; this stands as its own, separate single-pass proof instead.
func TestRepairDriverFixesInterproceduralLoanLivenessSinglePass(t *testing.T) {
	langBinary := testsupport.BuildCLI(t)

	fixturePath := testsupport.ProjectPath("testdata", "phase13", "derivation_interprocedural_loan_defect.lang")
	original, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "derivation_interprocedural_loan_defect.lang")
	mustWriteFile(t, sourcePath, original)

	outcome, err := Repair(context.Background(), langBinary, sourcePath)
	if err != nil {
		t.Fatalf("Repair: %v", err)
	}
	if outcome.Status != OutcomeRepaired {
		t.Fatalf("got status %q, want %q", outcome.Status, OutcomeRepaired)
	}
	if outcome.DiagnosisCode != "check.interprocedural_loan_liveness" {
		t.Fatalf("got diagnosis code %q, want %q", outcome.DiagnosisCode, "check.interprocedural_loan_liveness")
	}
	if outcome.RepairKind != "move_after_interprocedural_loan" {
		t.Fatalf("got repair kind %q, want %q", outcome.RepairKind, "move_after_interprocedural_loan")
	}
	if outcome.SubprocessCount != 2 {
		t.Fatalf("got subprocess count %d, want 2 (single pass: one diagnose, one reverify)", outcome.SubprocessCount)
	}

	// Independent re-verification: a SECOND, wholly separate `lang --json
	// check` subprocess invocation, outside the driver's own internal
	// reverify, reports the post-repair source clean.
	verify := testsupport.RunCLI(t, langBinary, nil, "--json", "check", sourcePath)
	decoded := decodeCheckJSON(t, verify.Stdout)
	if decoded.Status != statusPass {
		t.Fatalf("independent re-verification reports status %q, want %q", decoded.Status, statusPass)
	}
	if len(decoded.Diagnostics) != 0 {
		t.Fatalf("independent re-verification reports %d diagnostics, want 0: %+v", len(decoded.Diagnostics), decoded.Diagnostics)
	}

	// The driver operates on the TempDir copy only -- the checked-in
	// corpus fixture must be byte-unchanged by this test.
	afterTest, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterTest, original) {
		t.Fatal("testdata/phase13/derivation_interprocedural_loan_defect.lang was modified by this test -- the driver must operate on the TempDir copy only")
	}
}

// ---------------------------------------------------------------------
// 13-06 Task 1: D-13-28's twin pair -- criterion 3's re-check-clean gate
// ---------------------------------------------------------------------

// functionDeclSpan returns the byte range [start, end) of the function
// declaration named name within source, computed independently from the
// source bytes by literal-text scanning plus brace matching -- never by
// reading it back from check's own diagnostic payload (the plan's own
// acceptance criterion for TestTwinPairBlame's span assertion). start is
// the byte offset of the "fn " keyword and end is one past the function's
// own matching closing brace.
func functionDeclSpan(t testing.TB, source []byte, name string) (int, int) {
	t.Helper()
	marker := []byte("fn " + name + "(")
	start := bytes.Index(source, marker)
	if start == -1 {
		t.Fatalf("function %q not found in source", name)
	}
	relOpen := bytes.IndexByte(source[start:], '{')
	if relOpen == -1 {
		t.Fatalf("function %q has no opening brace", name)
	}
	openBrace := start + relOpen
	depth := 0
	for i := openBrace; i < len(source); i++ {
		switch source[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return start, i + 1
			}
		}
	}
	t.Fatalf("function %q has no matching closing brace", name)
	return 0, 0
}

// TestTwinPairBlame is 13-06 Task 1's Test 1 (D-13-28): both halves of the
// criterion-3 twin pair, driven through the real driver against a real
// built `lang` binary. Each half proves the repair-then-re-check-clean
// oracle: `Repair` reports the exact outcome string `repaired`, a wholly
// separate `lang --json check` invocation confirms zero diagnostics
// afterward, and the applied repair's own span -- captured independently
// via a direct diagnose+select BEFORE `Repair` ever touches the file --
// falls inside the expected function's own declaration span, computed
// here from the mutated source bytes via functionDeclSpan, never read
// back from the diagnostic.
//
// 13-06 (D-13-28 empirical adjudication -- see 13-06-SUMMARY.md's
// "Criterion 3 verdict" section for the full argument): the plan's own
// literal text expected the alpha half's true-fix location to be the
// SHARED CALLEE (`sink`)'s own declaration span. Verified directly against
// this exact fixture pair, through the real driver, before writing this
// test's assertions: the shipped move_after_interprocedural_loan repair
// (check.go's interproceduralLoanLivenessDiagnostic) operates ENTIRELY on
// the two statements inside the MUTATED CALLER's own body -- never inside
// `sink` -- for BOTH halves. This is consistent with, and a second
// instance of, D-13-02b's terminal finding that no B1-shaped
// (callee-contract) blame is constructible at this language's current
// maturity: the fixture shape D-13-28 originally envisioned (a defect
// whose only whole-program-clean fix lives in the shared callee) requires
// changing `sink`'s own declared contract independent of any one caller's
// misuse, which is exactly the shape `sameType`'s admission-time
// precondition makes unreachable. The twin pair's real, honestly
// constructible discrimination is DIFFERENT from what the plan's literal
// text assumed, and is asserted here: the repair is fully DATA-DRIVEN from
// the diagnostic's own recorded move/call operation IDs, so it correctly
// names whichever caller (`alpha` or `beta`) the injector actually broke
// -- never a fixed, hardcoded position. TestTwinPairBlameGuard is this
// test's own mutation kill, proving that a rule which ignores the
// diagnostic and always "fixes" a fixed caller instead fails the mirror
// half.
func TestTwinPairBlame(t *testing.T) {
	langBinary := testsupport.BuildCLI(t)

	cases := []struct {
		subtestName string
		fixture     string
		// mutatedFunc is the caller InterproceduralLoanInjector's marker
		// lives in for this fixture, and therefore the ONLY function this
		// run mutates -- and, per the empirical finding above, the SAME
		// function the shipped repair's own span must fall inside.
		mutatedFunc string
	}{
		{"alpha", "heldout_shared_callee_twin_alpha.lang", "alpha"},
		{"mirror", "heldout_shared_callee_twin_mirror.lang", "beta"},
	}

	for _, tc := range cases {
		t.Run(tc.subtestName, func(t *testing.T) {
			fixturePath := testsupport.ProjectPath("testdata", "phase13", tc.fixture)
			original, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatal(err)
			}
			mutated, err := session.InterproceduralLoanInjector{}.Inject(original)
			if err != nil {
				t.Fatalf("%s: injecting: %v", tc.subtestName, err)
			}

			dir := t.TempDir()
			sourcePath := filepath.Join(dir, tc.fixture)
			mustWriteFile(t, sourcePath, mutated)

			// Capture the repair's own span BEFORE Repair() ever writes
			// to the file -- Repair()'s own Outcome envelope carries no
			// span field, so this is an independent diagnose+select over
			// the identical, deterministic protocol Repair() itself uses.
			diagnosis := testsupport.RunCLI(t, langBinary, nil, "--json", "check", sourcePath)
			decoded := decodeCheckJSON(t, diagnosis.Stdout)
			repair, code, ok := selectRepair(decoded)
			if !ok {
				t.Fatalf("%s: expected a driver-eligible repair, got none (diagnostics: %+v)", tc.subtestName, decoded.Diagnostics)
			}
			if code != "check.interprocedural_loan_liveness" {
				t.Fatalf("%s: expected check.interprocedural_loan_liveness, got %q", tc.subtestName, code)
			}
			if repair.Kind != "move_after_interprocedural_loan" {
				t.Fatalf("%s: expected move_after_interprocedural_loan, got %q", tc.subtestName, repair.Kind)
			}

			outcome, err := Repair(context.Background(), langBinary, sourcePath)
			if err != nil {
				t.Fatalf("%s: Repair: %v", tc.subtestName, err)
			}
			if outcome.Status != OutcomeRepaired {
				t.Fatalf("%s: got status %q, want %q (diagnosis=%q repair=%q)", tc.subtestName, outcome.Status, OutcomeRepaired, outcome.DiagnosisCode, outcome.RepairKind)
			}

			verify := testsupport.RunCLI(t, langBinary, nil, "--json", "check", sourcePath)
			verifyDecoded := decodeCheckJSON(t, verify.Stdout)
			if verifyDecoded.Status != statusPass {
				t.Fatalf("%s: independent re-verification reports status %q, want %q", tc.subtestName, verifyDecoded.Status, statusPass)
			}
			if len(verifyDecoded.Diagnostics) != 0 {
				t.Fatalf("%s: independent re-verification reports %d diagnostics, want 0: %+v", tc.subtestName, len(verifyDecoded.Diagnostics), verifyDecoded.Diagnostics)
			}

			funcStart, funcEnd := functionDeclSpan(t, mutated, tc.mutatedFunc)
			if repair.Span == nil || repair.Span.Start < funcStart || repair.Span.End > funcEnd {
				t.Fatalf("%s: repair span %+v does not fall inside %s's own declaration span [%d,%d)", tc.subtestName, repair.Span, tc.mutatedFunc, funcStart, funcEnd)
			}

			afterTest, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(afterTest, original) {
				t.Fatalf("%s: %s was modified by this test -- the driver must operate on the TempDir copy only", tc.subtestName, tc.fixture)
			}
		})
	}
}

// TestTwinPairBlameGuard is 13-06 Task 1's Test 2 (D-13-30d), the
// criterion-3 mutation kill -- ADAPTED from the plan's literal text per
// the empirical finding TestTwinPairBlame documents (see
// 13-06-SUMMARY.md's "Criterion 3 verdict"): since the shipped repair's
// own span ALWAYS coincides with the diagnostic's detection-site function
// (never a separate "compiler-named" function such as `sink`), the
// literal "apply the repair at the detection site instead of the
// compiler-named function" construction has no discrepancy left to
// exploit for this diagnostic class. The discrimination the twin pair
// DOES support, and this test kills, is the one both sealed fixtures'
// own file-header comments name explicitly: a rule that always "fixes" a
// FIXED, hardcoded caller position (here, `alpha` -- the first caller
// declared) regardless of which caller the diagnostic actually names.
//
// On the mirror half, `beta` is the actually-broken caller. This test
// constructs the naive fixed-position edit: it swaps `alpha`'s own two
// statements (which are clean and already correctly ordered in this
// fixture) into the SAME wrong order the injector itself produces,
// leaving `beta`'s real defect completely untouched. Re-checking the
// result must NOT check clean -- a non-empty diagnostics array is
// asserted, and the specific surviving code
// (check.interprocedural_loan_liveness) is named, not just "some
// difference".
func TestTwinPairBlameGuard(t *testing.T) {
	langBinary := testsupport.BuildCLI(t)

	fixturePath := testsupport.ProjectPath("testdata", "phase13", "heldout_shared_callee_twin_mirror.lang")
	original, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	mutated, err := session.InterproceduralLoanInjector{}.Inject(original)
	if err != nil {
		t.Fatalf("injecting: %v", err)
	}

	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "heldout_shared_callee_twin_mirror.lang")
	mustWriteFile(t, sourcePath, mutated)

	alphaStart, alphaEnd := functionDeclSpan(t, mutated, "alpha")
	alphaBody := mutated[alphaStart:alphaEnd]
	const callLine = "let routed = sink(borrowed)"
	const takeLine = "let delivered = take buffer"
	callIdx := bytes.Index(alphaBody, []byte(callLine))
	takeIdx := bytes.Index(alphaBody, []byte(takeLine))
	if callIdx == -1 || takeIdx == -1 || callIdx >= takeIdx {
		t.Fatalf("test construction error: expected alpha's own clean call-before-take ordering in the mirror fixture, got callIdx=%d takeIdx=%d", callIdx, takeIdx)
	}
	naiveSpan := jsonSpan{Start: alphaStart + callIdx, End: alphaStart + takeIdx + len(takeLine)}
	// The exact wrong-order shape InterproceduralLoanInjector itself
	// produces (take before call) -- applied to the WRONG (clean) caller.
	naiveReplacement := takeLine + "\n  " + callLine
	naive := jsonRepair{
		Kind:          "move_after_interprocedural_loan",
		Span:          &naiveSpan,
		Replacement:   naiveReplacement,
		Applicability: applicabilityMachineApplicable,
	}
	if err := applyRepair(sourcePath, naive); err != nil {
		t.Fatalf("applying naive fixed-position edit: %v", err)
	}

	verify := testsupport.RunCLI(t, langBinary, nil, "--json", "check", sourcePath)
	verifyDecoded := decodeCheckJSON(t, verify.Stdout)
	if verifyDecoded.Status == statusPass {
		t.Fatal("naive fixed-position edit (targeting alpha, which the diagnostic never named) unexpectedly re-checked clean -- the mutation kill is not discriminating")
	}
	if len(verifyDecoded.Diagnostics) == 0 {
		t.Fatal("expected a non-empty diagnostics array after the naive fixed-position edit -- beta's own real defect was never addressed")
	}
	survives := false
	for _, d := range verifyDecoded.Diagnostics {
		if d.Code == "check.interprocedural_loan_liveness" {
			survives = true
		}
	}
	if !survives {
		t.Fatalf("expected check.interprocedural_loan_liveness to survive (beta's own defect untouched), got %+v", verifyDecoded.Diagnostics)
	}

	afterTest, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterTest, original) {
		t.Fatal("heldout_shared_callee_twin_mirror.lang was modified by this test -- the driver must operate on the TempDir copy only")
	}
}
