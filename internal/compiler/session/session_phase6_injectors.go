package session

import (
	"fmt"
	"strings"
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
// (Task 3) demonstrates concretely what that silent failure would look like
// without the guard.

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
// TestEveryInjectorRefusesWhenMarkerDisappears's zero-versus-one boundary,
// Task 3). want names what was being searched for, so the returned error
// names both the found count and the required count.
func markerGuard(count int, want string) error {
	if count == 0 {
		return &InjectorError{Code: InjectorTargetMissingCode, Err: fmt.Errorf("%s marker count is %d, want at least 1", want, count)}
	}
	return nil
}

// markerCount and lastMarkerLine are the shared line-oriented marker
// primitives every source-granularity injector (match/move/borrow) uses.
// Where more than one line carries a class's marker, the LAST one is the
// injector's target -- the same stable "take the last match" choice
// ReleaseOmissionMutationRunner already makes for lang:release-site, stated
// here once rather than re-derived per injector.
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
