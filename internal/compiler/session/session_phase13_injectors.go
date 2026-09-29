package session

import (
	"fmt"
	"strings"
)

// This file implements Phase 13's three interprocedural defect injectors
// (D-13-24), cloning the Injector/markerGuard shape
// session_phase6_injectors.go established for Phase 6's five defect
// classes: a marker constant, a Name(), an Inject([]byte) ([]byte, error)
// built on lastMarkerLine/markerCount/markerGuard, and a
// <name>InjectSkippingGuard twin that is never called from the real
// Inject path. Each injector below produces exactly one mechanical
// change from a clean, marker-carrying base program, reproducing one of
// DX-07's three Set A target interprocedural diagnostic codes (D-13-09):
// check.interprocedural_loan_liveness, syntax.fallible_call_not_consumed,
// and check.call_argument_type_mismatch. See testdata/phase13/README for
// the heldout_/derivation_ corpus split these injectors operate against
// (D-13-24 through D-13-31).

// loanTargetMarker marks the take-statement InterproceduralLoanInjector
// reorders ahead of its immediately preceding statement -- by
// construction, the call whose own operation is the loan's last use in
// the clean base (D-13-09b's callIsLastUse backward gate). Reproduces
// the exact swap-fixable shape 13-01's tracer fixture
// (derivation_interprocedural_loan_defect.schway) established by hand,
// now produced mechanically and reused across a shared-callee,
// multi-caller topology (D-13-28).
const loanTargetMarker = "// schway:interprocedural-loan-target"

// InterproceduralLoanInjector swaps the marked statement with the
// statement immediately preceding it, reproducing
// check.interprocedural_loan_liveness by reordering a take BEFORE the
// call that is its loan's true last use in the clean base.
type InterproceduralLoanInjector struct{}

func (InterproceduralLoanInjector) Name() string { return "interprocedural_loan" }

func (InterproceduralLoanInjector) Inject(source []byte) ([]byte, error) {
	lines, index := lastMarkerLine(source, loanTargetMarker)
	if err := markerGuard(markerCount(lines, loanTargetMarker), "interprocedural loan target"); err != nil {
		return nil, err
	}
	if index == 0 {
		return nil, &InjectorError{Code: InjectorTargetMissingCode, Err: fmt.Errorf("interprocedural_loan: marked line has no preceding statement to swap with")}
	}
	mutated := append([]string(nil), lines...)
	mutated[index-1], mutated[index] = mutated[index], mutated[index-1]
	return []byte(strings.Join(mutated, "\n")), nil
}

// interproceduralLoanInjectSkippingGuard is the guard-DISABLED variant
// TestPhase13InjectorGuardsAreNotInert uses to demonstrate concretely
// what InterproceduralLoanInjector.Inject would do without markerGuard:
// given a source with no loanTargetMarker at all, it returns the source
// completely unmutated. Deliberately never called by
// InterproceduralLoanInjector.Inject itself.
func interproceduralLoanInjectSkippingGuard(source []byte) []byte {
	lines, index := lastMarkerLine(source, loanTargetMarker)
	if index == -1 {
		return append([]byte(nil), source...)
	}
	mutated := append([]string(nil), lines...)
	if index > 0 {
		mutated[index-1], mutated[index] = mutated[index], mutated[index-1]
	}
	return []byte(strings.Join(mutated, "\n"))
}

// fallibleConsumeTargetMarker marks the `try`-wrapped fallible call
// FallibleConsumeInjector strips `try` from.
const fallibleConsumeTargetMarker = "// schway:fallible-consume-target"

// FallibleConsumeInjector removes the leading "try " from the marked
// binding's right-hand side, turning a legally-consumed fallible call
// into a bare call -- reproducing syntax.fallible_call_not_consumed
// (D-13-09 Set A class 2).
type FallibleConsumeInjector struct{}

func (FallibleConsumeInjector) Name() string { return "fallible_consume" }

func (FallibleConsumeInjector) Inject(source []byte) ([]byte, error) {
	lines, index := lastMarkerLine(source, fallibleConsumeTargetMarker)
	if err := markerGuard(markerCount(lines, fallibleConsumeTargetMarker), "fallible consume target"); err != nil {
		return nil, err
	}
	const tryAssign = "= try "
	line := lines[index]
	if !strings.Contains(line, tryAssign) {
		return nil, &InjectorError{Code: InjectorTargetMissingCode, Err: fmt.Errorf("fallible_consume: marked line has no `= try ` call to strip")}
	}
	mutated := append([]string(nil), lines...)
	mutated[index] = strings.Replace(line, tryAssign, "= ", 1)
	return []byte(strings.Join(mutated, "\n")), nil
}

// fallibleConsumeInjectSkippingGuard is the guard-DISABLED twin: given
// marker-free source, it returns the input completely unchanged. Never
// called by FallibleConsumeInjector.Inject itself.
func fallibleConsumeInjectSkippingGuard(source []byte) []byte {
	lines, index := lastMarkerLine(source, fallibleConsumeTargetMarker)
	if index == -1 {
		return append([]byte(nil), source...)
	}
	const tryAssign = "= try "
	line := lines[index]
	if !strings.Contains(line, tryAssign) {
		return append([]byte(nil), source...)
	}
	mutated := append([]string(nil), lines...)
	mutated[index] = strings.Replace(line, tryAssign, "= ", 1)
	return []byte(strings.Join(mutated, "\n"))
}

// callArgumentTargetMarker marks the callee's own `fn` declaration line
// CallArgumentTypeInjector toggles between Byte and Buffer.
const callArgumentTargetMarker = "// schway:call-argument-target"

// CallArgumentTypeInjector toggles every occurrence of "Byte" to
// "Buffer" (or the reverse, if the line names no "Byte") on the marked
// callee declaration line -- both the declared parameter type and the
// declared return type live on that one line for every fixture this
// injector targets, so the toggle keeps the callee internally
// sameType-consistent while changing which type a caller must supply,
// reproducing check.call_argument_type_mismatch at every call site that
// still passes the ORIGINAL type (D-13-09 Set A class 3).
type CallArgumentTypeInjector struct{}

func (CallArgumentTypeInjector) Name() string { return "call_argument_type" }

func (CallArgumentTypeInjector) Inject(source []byte) ([]byte, error) {
	lines, index := lastMarkerLine(source, callArgumentTargetMarker)
	if err := markerGuard(markerCount(lines, callArgumentTargetMarker), "call argument type target"); err != nil {
		return nil, err
	}
	mutatedLine, ok := toggleByteBuffer(lines[index])
	if !ok {
		return nil, &InjectorError{Code: InjectorTargetMissingCode, Err: fmt.Errorf("call_argument_type: marked line names neither Byte nor Buffer")}
	}
	mutated := append([]string(nil), lines...)
	mutated[index] = mutatedLine
	return []byte(strings.Join(mutated, "\n")), nil
}

// callArgumentTypeInjectSkippingGuard is the guard-DISABLED twin: given
// marker-free source, it returns the input completely unchanged. Never
// called by CallArgumentTypeInjector.Inject itself.
func callArgumentTypeInjectSkippingGuard(source []byte) []byte {
	lines, index := lastMarkerLine(source, callArgumentTargetMarker)
	if index == -1 {
		return append([]byte(nil), source...)
	}
	mutatedLine, ok := toggleByteBuffer(lines[index])
	if !ok {
		return append([]byte(nil), source...)
	}
	mutated := append([]string(nil), lines...)
	mutated[index] = mutatedLine
	return []byte(strings.Join(mutated, "\n"))
}

// toggleByteBuffer swaps every "Byte" on line for "Buffer", or every
// "Buffer" for "Byte" if the line contains no "Byte" -- the single
// mechanical edit that keeps a `fn NAME(param: T) -> T {` declaration
// line's own sameType invariant intact (both the parameter and return
// type occurrences flip together) while changing which type a caller
// must supply to remain admissible.
func toggleByteBuffer(line string) (string, bool) {
	switch {
	case strings.Contains(line, "Byte"):
		return strings.ReplaceAll(line, "Byte", "Buffer"), true
	case strings.Contains(line, "Buffer"):
		return strings.ReplaceAll(line, "Buffer", "Byte"), true
	default:
		return line, false
	}
}
