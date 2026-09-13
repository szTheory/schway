package session

import (
	"fmt"
	"strings"

	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/native"
)

// This file implements DX-04's five defect injectors (D-06-25) and their
// shared fail-closed marker guard (D-06-27.3). The guard exists because of
// this project's own three standing process rules, adopted after three gate
// failures that all shared one shape: a green test whose reachable input
// space omitted the hard case. An injector whose target marker or span has
// silently disappeared -- because an unrelated lowering change moved a line,
// say -- is exactly that shape: it would report a clean run while never
// having exercised the defect it claims to. Every injector below refuses
// (InjectorTargetMissingCode) rather than silently returning an unmutated
// source that happens to check clean; TestInjectorMarkerCountGuardIsNotInert
// demonstrates concretely what that silent failure would look like without
// the guard.

// InjectorTargetMissingCode is the one typed refusal code every injector in
// this file returns when its target marker's count is zero -- the
// session.go matched == -1 refusal (ReleaseOmissionMutationRunner,
// AliasFactMutationRunner), generalized across all five defect classes.
const InjectorTargetMissingCode = "phase6.injector_target_missing"

// InjectorError is the {Code}-only typed-failure shape this phase's new
// packages reuse (matching debugmap.Error / evidence.ValidationError /
// originvalidate.Error), so a driver or test can distinguish "marker
// vanished" from any other failure by Code alone rather than string
// matching Err's text.
type InjectorError struct {
	Code string
	Err  error
}

func (e *InjectorError) Error() string { return e.Code + ": " + e.Err.Error() }
func (e *InjectorError) Unwrap() error { return e.Err }

// Injector is the shared shape all five defect classes implement. Inject
// takes whatever artifact bytes that class's marker convention lives in --
// `.lang` source for match/move/borrow/stale_evidence, generated C for
// cleanup (D-06-25: it targets the emitter's own `lang:release-site`
// marker, which never appears in `.lang` source at all) -- and returns
// either a mutated artifact carrying exactly one mechanical defect, or a
// typed InjectorTargetMissingCode refusal.
type Injector interface {
	Name() string
	Inject(source []byte) ([]byte, error)
}

// markerGuard is the matched == -1 refusal, generalized: it refuses only
// when count is exactly zero (a target-count of one or more proceeds -- see
// TestEveryInjectorRefusesWhenMarkerDisappears's zero-versus-one boundary).
// want names what was being searched for, so the returned error names both
// the found count and the required count.
func markerGuard(count int, want string) error {
	if count == 0 {
		return &InjectorError{Code: InjectorTargetMissingCode, Err: fmt.Errorf("%s marker count is %d, want at least 1", want, count)}
	}
	return nil
}

// markerCount and lastMarkerLine are the shared line-oriented marker
// primitives every source-granularity injector (match/move/borrow) below
// uses. Where more than one line carries a class's marker, the LAST one is
// the injector's target -- the same stable "take the last match" choice
// ReleaseOmissionMutationRunner already makes for lang:release-site, stated
// here once rather than re-derived per injector (TestInjectorTargetChoiceIsSpecified).
func markerCount(lines []string, marker string) int {
	count := 0
	for _, line := range lines {
		if strings.Contains(line, marker) {
			count++
		}
	}
	return count
}

func lastMarkerLine(source []byte, marker string) (lines []string, index int) {
	lines = strings.Split(string(source), "\n")
	index = -1
	for lineIndex, line := range lines {
		if strings.Contains(line, marker) {
			index = lineIndex
		}
	}
	return lines, index
}

// matchTargetMarker marks the single match arm MatchInjector deletes.
const matchTargetMarker = "// lang:match-target"

// MatchInjector removes the marked arm from a bare-arm exhaustive match,
// producing a match.non_exhaustive defect at source granularity --
// exhaustiveness lives above the C layer the existing NAT-03 runners
// mutate (D-06-25). Every heldout_match_defect.lang-shaped fixture maps
// every arm's pattern to itself, so the resulting diagnostic's
// add_missing_arm repair (check.go's allArmsSelfMap gate) reproduces the
// deleted arm exactly (D-06-26).
type MatchInjector struct{}

func (MatchInjector) Name() string { return "match" }

func (MatchInjector) Inject(source []byte) ([]byte, error) {
	lines, index := lastMarkerLine(source, matchTargetMarker)
	if err := markerGuard(markerCount(lines, matchTargetMarker), "match arm target"); err != nil {
		return nil, err
	}
	mutated := append(append([]string(nil), lines[:index]...), lines[index+1:]...)
	return []byte(strings.Join(mutated, "\n")), nil
}

// matchInjectSkippingGuard is the guard-DISABLED variant
// TestInjectorMarkerCountGuardIsNotInert uses to demonstrate concretely
// what MatchInjector.Inject would do without markerGuard: given a source
// with no matchTargetMarker at all, it returns the source completely
// unmutated (not an error, not a partial edit) -- exactly the silent-pass
// theatre the guard exists to prevent. It is deliberately never called by
// MatchInjector.Inject itself.
func matchInjectSkippingGuard(source []byte) []byte {
	lines, index := lastMarkerLine(source, matchTargetMarker)
	if index == -1 {
		return append([]byte(nil), source...)
	}
	mutated := append(append([]string(nil), lines[:index]...), lines[index+1:]...)
	return []byte(strings.Join(mutated, "\n"))
}

// moveTargetMarker marks the take-expression MoveInjector corrupts.
const moveTargetMarker = "// lang:move-target"

// MoveInjector corrupts the marked take-expression's source identifier back
// to the function's own parameter name -- the parameter is always the
// place a straight-line body's first take already moved out of, so
// referencing it again on the marked line reproduces exactly the affine
// use-after-move defect class D-06-25 names, at source granularity.
type MoveInjector struct{}

func (MoveInjector) Name() string { return "move" }

func (MoveInjector) Inject(source []byte) ([]byte, error) {
	lines, index := lastMarkerLine(source, moveTargetMarker)
	if err := markerGuard(markerCount(lines, moveTargetMarker), "move target"); err != nil {
		return nil, err
	}
	parameter, ok := extractFirstParameterName(source)
	if !ok {
		return nil, &InjectorError{Code: InjectorTargetMissingCode, Err: fmt.Errorf("move: no function parameter found to reintroduce as the stale reference")}
	}
	mutatedLine, ok := replaceTakeSource(lines[index], parameter)
	if !ok {
		return nil, &InjectorError{Code: InjectorTargetMissingCode, Err: fmt.Errorf("move: marked line has no take-expression to corrupt")}
	}
	lines[index] = mutatedLine
	return []byte(strings.Join(lines, "\n")), nil
}

// extractFirstParameterName parses the parameter name out of this fixture
// shape's single `fn NAME(PARAM: TYPE) -> ...` declaration -- deliberately
// minimal (this phase's language has exactly one parameter per function;
// PROJECT.md), not a general parser.
func extractFirstParameterName(source []byte) (string, bool) {
	text := string(source)
	fnIndex := strings.Index(text, "fn ")
	if fnIndex == -1 {
		return "", false
	}
	openIndex := strings.Index(text[fnIndex:], "(")
	if openIndex == -1 {
		return "", false
	}
	openIndex += fnIndex
	colonIndex := strings.Index(text[openIndex:], ":")
	if colonIndex == -1 {
		return "", false
	}
	colonIndex += openIndex
	name := strings.TrimSpace(text[openIndex+1 : colonIndex])
	if name == "" {
		return "", false
	}
	return name, true
}

// replaceTakeSource rewrites the identifier immediately following the
// first "take " on line with replacement, leaving everything else
// (including any trailing marker comment) untouched.
func replaceTakeSource(line, replacement string) (string, bool) {
	const takeKeyword = "take "
	takeIndex := strings.Index(line, takeKeyword)
	if takeIndex == -1 {
		return line, false
	}
	start := takeIndex + len(takeKeyword)
	end := start
	for end < len(line) && isIdentByte(line[end]) {
		end++
	}
	if end == start {
		return line, false
	}
	return line[:start] + replacement + line[end:], true
}

func isIdentByte(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

// borrowTargetMarker marks the shared borrow BorrowInjector escalates.
const borrowTargetMarker = "// lang:borrow-target"

// BorrowInjector escalates the marked shared (`borrow`) binding to
// exclusive (`borrow mut`), producing exactly the overlapping
// exclusive+shared loan-conflict defect class D-06-25 names -- two
// overlapping SHARED loans are legal (testdata/phase3/shared_shared_accept.lang),
// so the escalation, and only the escalation, is what makes the mutated
// fixture reject.
type BorrowInjector struct{}

func (BorrowInjector) Name() string { return "borrow" }

func (BorrowInjector) Inject(source []byte) ([]byte, error) {
	lines, index := lastMarkerLine(source, borrowTargetMarker)
	if err := markerGuard(markerCount(lines, borrowTargetMarker), "borrow target"); err != nil {
		return nil, err
	}
	line := lines[index]
	if strings.Contains(line, "borrow mut ") {
		return nil, &InjectorError{Code: InjectorTargetMissingCode, Err: fmt.Errorf("borrow: marked line is already an exclusive borrow, nothing to escalate")}
	}
	const sharedBorrow = "= borrow "
	if !strings.Contains(line, sharedBorrow) {
		return nil, &InjectorError{Code: InjectorTargetMissingCode, Err: fmt.Errorf("borrow: marked line has no shared borrow to escalate")}
	}
	lines[index] = strings.Replace(line, sharedBorrow, "= borrow mut ", 1)
	return []byte(strings.Join(lines, "\n")), nil
}

// CleanupInjector is a thin adapter over the shipped
// ReleaseOmissionMutationRunner (D-06-25: "reuse the existing
// release-omission mutation runner directly", not a reimplementation).
// Unlike its siblings, source here is generated C, not `.lang`: the
// `lang:release-site` marker cgen emits never appears in `.lang` source at
// all. There is exactly one release-marker scan in this whole package --
// ReleaseOmissionMutationRunner.Mutate's own -- which
// TestCleanupInjectorReusesReleaseOmissionRunner asserts by go/ast.
type CleanupInjector struct{}

func (CleanupInjector) Name() string { return "cleanup" }

func (CleanupInjector) Inject(source []byte) ([]byte, error) {
	runner := NewReleaseOmissionMutationRunner(native.Runner{})
	mutated, err := runner.Mutate(string(source))
	if err != nil {
		// ReleaseOmissionMutationRunner.Mutate returns its own
		// native.backend_control_invalid code (session.go's pre-existing
		// convention for THAT runner's callers); every Phase 6 injector
		// must additionally answer to the one uniform
		// InjectorTargetMissingCode AllInjectors()-driven tests expect, so
		// the underlying error is wrapped, never replaced -- Unwrap still
		// reaches the original native.ToolError.
		return nil, &InjectorError{Code: InjectorTargetMissingCode, Err: fmt.Errorf("cleanup: %w", err)}
	}
	return []byte(mutated), nil
}

// evidenceSubjectMarker marks a fixture as eligible stale-evidence subject
// material. Unlike the other four classes, it does not name a mutation
// SITE (there is no source defect to inject, D-06-25) -- its disappearance
// means "this input was never meant to be run through this injector at
// all", which is still the same fail-closed shape: refuse rather than
// silently re-touch an arbitrary, unvetted file.
const evidenceSubjectMarker = "// lang:evidence-subject"

// StaleEvidenceInjector re-touches source captured evidence was bound to,
// so the manifest's recorded SHA-256 no longer matches -- D-06-25's
// stale-evidence class is NOT a source defect, so unlike its siblings the
// mutation here is a deliberately inert, legitimate-looking edit (an
// appended comment line), not a corruption. The locator is
// ValidateEvidenceCommandFile's own mismatch report (session.go); this
// injector computes no source diff and consults none
// (TestStaleEvidenceInjectorBreaksManifestBinding's go/ast assertion,
// D-06-27.2's structured-channel discipline extended to the locator side).
type StaleEvidenceInjector struct{}

func (StaleEvidenceInjector) Name() string { return "stale_evidence" }

func (StaleEvidenceInjector) Inject(source []byte) ([]byte, error) {
	lines, _ := lastMarkerLine(source, evidenceSubjectMarker)
	if err := markerGuard(markerCount(lines, evidenceSubjectMarker), "evidence subject"); err != nil {
		return nil, err
	}
	retouched := append(append([]byte(nil), source...), []byte("\n// lang:stale-evidence-retouch\n")...)
	return retouched, nil
}

// AllInjectors returns all five defect injectors, enumerated once here so
// TestEveryInjectorRefusesWhenMarkerDisappears (and any future caller) is
// driven from this single list rather than a hand-written one -- a sixth
// injector added later without updating this function is invisible to
// those tests, but a sixth injector added HERE without its own guard is
// exactly what those tests catch.
func AllInjectors() []Injector {
	return []Injector{
		MatchInjector{},
		MoveInjector{},
		BorrowInjector{},
		CleanupInjector{},
		StaleEvidenceInjector{},
		// Phase 13 (D-13-24, D-13-30a): the three interprocedural defect
		// injectors, defined in session_phase13_injectors.go, registered
		// here so TestEveryInjectorRefusesWhenMarkerDisappears and
		// TestInjectorTargetChoiceIsSpecified pick them up automatically.
		InterproceduralLoanInjector{},
		FallibleConsumeInjector{},
		CallArgumentTypeInjector{},
	}
}

// ExerciseResult is the minimal pass/fail envelope
// RunDefectInjectionExercise reports, so a refusal at the injector boundary
// (Inject returning an error) and a refusal further downstream (the
// mutated artifact somehow checking clean, meaning the defect was never
// observed) both surface as the SAME kind of failure to a caller -- never a
// pass (D-06-27.3, TestInjectorRefusalPropagatesToExerciseFailure).
type ExerciseResult struct {
	Injector    string
	Status      string // "pass" | "fail"
	Err         error
	Diagnostics []diagnostic.Diagnostic
}

// RunDefectInjectionExercise drives one source-granularity injector
// (match/move/borrow) through Inject and, only on success, session.Check,
// and asserts the mutated source is REJECTED -- a mutated source that
// checks clean is exactly as much a failure as an injector refusal, since
// it means the defect was never actually observed.
func RunDefectInjectionExercise(injector Injector, source []byte) ExerciseResult {
	mutated, err := injector.Inject(source)
	if err != nil {
		return ExerciseResult{Injector: injector.Name(), Status: "fail", Err: err}
	}
	checked := Check(mutated)
	if len(checked.Diagnostics) == 0 {
		return ExerciseResult{Injector: injector.Name(), Status: "fail", Err: fmt.Errorf("%s: mutated source checked clean, defect was not observed", injector.Name())}
	}
	return ExerciseResult{Injector: injector.Name(), Status: "pass", Diagnostics: checked.Diagnostics}
}
