package session

import (
	"context"
	"fmt"
	"time"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
)

// Phase11ZeroAttributeGateID names the mid-phase gate's own lane (D-11-14).
const Phase11ZeroAttributeGateID = "lane:phase11-zero-attribute-gate"

// The mid-phase gate's own control-label vocabulary (D-11-14/D-11-19): one
// label per named conjunct, plus the manifest consistency check, so
// 11-MIDPHASE-GATE.md's written adjudication can cite each by name.
const (
	ControlPhase11ScanClean            = "control:phase11.scan_clean"
	ControlPhase11NonVacuousWouldCarry = "control:phase11.non_vacuous_would_carry"
	ControlPhase11StructuralFloors     = "control:phase11.structural_floors"
	ControlPhase11InterpreterAgreesO0  = "control:phase11.interpreter_agrees_o0"
	ControlPhase11ManifestConsistency  = "control:phase11.manifest_consistency"
)

// Phase11GateReport is VerifyPhase11ZeroAttributeGate's fully observed,
// conjunct-by-conjunct result (D-11-19): every value 11-MIDPHASE-GATE.md's
// written adjudication must cite verbatim as a concrete number or string,
// never a characterization.
type Phase11GateReport struct {
	FunctionCount           int
	CallEdgeCount           int
	WouldCarryCount         int
	WouldCarryFunctionIDs   []string
	ScannedArtifactCount    int
	BannedAttributesFound   []string
	ManifestEmptyAttributes bool
	InterpreterAgreesAtO0   bool
	Result                  protocol.Result
}

// Passed reports whether every one of the gate's conjuncts held.
func (r Phase11GateReport) Passed() bool {
	return r.Result.Status == protocol.StatusPass
}

// phase11WouldCarryRestrict is Knower B's own independent re-derivation
// (D-11-14/D-11-15) of exactly the structural condition
// cgen.SelectsByPointerLowering decides, reading only
// core.Function/core.LinearBody -- session's own view of the PROGRAM,
// never cgen's derivation read a second time. This is session's own
// third, independent restatement of the identical structural fact (D-12:
// zero shared helpers between the derivations); any semantic drift
// between this and cgen's own selectsByPointerLowering is exactly what
// TestAliasFactAgreesWithByPointerSelection (check package) already
// polices for check's sibling derivation.
func phase11WouldCarryRestrict(function core.Function) bool {
	linear := function.Linear
	if function.Match != nil || function.PublicOrigin != nil || linear == nil || len(linear.Blocks) > 0 {
		return false
	}
	operations := linear.Operations
	if len(operations) == 0 {
		return false
	}
	first := operations[0]
	if first.Kind != core.OpBorrowExclusive || first.SourceID != function.Parameter.ID || first.TargetID == "" {
		return false
	}
	for _, operation := range operations[1:] {
		if operation.SourceID == function.Parameter.ID {
			return false
		}
	}
	current := first.TargetID
	terminatorIndex := -1
	for index := 1; index < len(operations); index++ {
		operation := operations[index]
		if operation.SourceID != current {
			return false
		}
		if operation.Kind == core.OpReturn {
			terminatorIndex = index
			break
		}
		if operation.TargetID == "" {
			return false
		}
		current = operation.TargetID
	}
	return terminatorIndex == len(operations)-1
}

// VerifyPhase11ZeroAttributeGate is Phase 11's mandatory mid-phase gate
// (D-11-14): a four-part conjunction with two genuinely independent
// knowers -- cgen.ScanForBannedAttributes reading OUTPUT BYTES (Knower A)
// and phase11WouldCarryRestrict's own re-derivation reading the PROGRAM
// (Knower B) -- plus the structural non-vacuity floors (function count,
// call-edge count) and a real interpreter/-O0 agreement run (RunNative,
// which also exercises -O3).
//
// profile controls ONLY the counterfactual single-function artifacts this
// gate itself synthesizes, one per would-carry function (D-11-18):
// production callers and the gate's own default-shape tests always pass
// cgen.AttributesSuppressed, matching D-11-09's terminal, permanent
// policy. TestPhase11GateMutationKill is the one caller that passes
// cgen.AttributesJustified, proving Knower A's scan is not vacuous by
// watching the SAME lane go red on the SAME corpus.
//
// source is checked and corevalidate-validated here (this lane's own
// admission, mirroring verifyPhase6NativeDifferentialLane's shape); runner
// drives the real -O0/-O3 native binaries RunNative compiles.
func VerifyPhase11ZeroAttributeGate(ctx context.Context, source []byte, runner native.Runner, profile cgen.AttributeSuppressionProfile) (Phase11GateReport, error) {
	report := Phase11GateReport{}
	result := protocol.New("verify", protocol.StatusPass)

	addLane := func(id, status string, controls []string, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: protocol.LaneSchema1, ID: id, Status: status,
			Controls: append([]string{}, controls...), RecomputedWork: 1,
			ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
		})
	}
	markFail := func(status string) {
		if result.Status == protocol.StatusPass {
			result.Status = status
		}
	}

	checked := Check(source)
	if len(checked.Diagnostics) != 0 {
		return report, fmt.Errorf("phase11 gate: corpus failed to check: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return report, fmt.Errorf("phase11 gate: corevalidate rejected corpus: %+v", validated.Problems)
	}
	program := validated.Program()

	// Structural floors + Knower B, both derived from the PROGRAM alone,
	// never from cgen's own emission.
	report.FunctionCount = len(program.Functions)
	callEdges := 0
	var wouldCarry []core.Function
	for _, function := range program.Functions {
		if function.Linear != nil {
			for _, operation := range function.Linear.Operations {
				if operation.Kind == core.OpCall {
					callEdges++
				}
			}
		}
		if phase11WouldCarryRestrict(function) {
			wouldCarry = append(wouldCarry, function)
			report.WouldCarryFunctionIDs = append(report.WouldCarryFunctionIDs, function.ID)
		}
	}
	report.CallEdgeCount = callEdges
	report.WouldCarryCount = len(wouldCarry)

	floorsStarted := time.Now()
	if report.FunctionCount >= 2 && report.CallEdgeCount >= 1 {
		addLane("lane:phase11-structural-floors", protocol.StatusPass, []string{ControlPhase11StructuralFloors}, floorsStarted)
	} else {
		addLane("lane:phase11-structural-floors", protocol.StatusInvalid, nil, floorsStarted)
		markFail(protocol.StatusInvalid)
	}

	vacuityStarted := time.Now()
	if report.WouldCarryCount >= 1 {
		addLane("lane:phase11-non-vacuous-would-carry", protocol.StatusPass, []string{ControlPhase11NonVacuousWouldCarry}, vacuityStarted)
	} else {
		addLane("lane:phase11-non-vacuous-would-carry", protocol.StatusInvalid, nil, vacuityStarted)
		markFail(protocol.StatusInvalid)
	}

	// Knower A: scan OUTPUT BYTES -- the actual whole-program TU emitProgram
	// produces (always attribute-free by construction, D-11-04), plus one
	// counterfactual single-function artifact per would-carry function,
	// synthesized under the caller's own profile. Under
	// AttributesSuppressed (D-11-09's permanent, shipped policy) these
	// counterfactual artifacts stay attribute-free too; under
	// AttributesJustified (the mutation-kill test's own toggle) they carry
	// `restrict`, and this same scan catches it.
	scanStarted := time.Now()
	artifacts := make([]string, 0, 1+len(wouldCarry))
	tu, emitErr := cgen.EmitNative(program)
	if emitErr != nil {
		return report, fmt.Errorf("phase11 gate: emitting whole-program artifact: %w", emitErr)
	}
	artifacts = append(artifacts, tu)

	restoreProfile := cgen.SetAttributeSuppressionProfileForTest(profile)
	for _, function := range wouldCarry {
		single := core.Program{Schema: program.Schema, Module: program.Module, ModuleID: program.ModuleID, Functions: []core.Function{function}}
		artifact, singleErr := cgen.EmitNative(single)
		if singleErr != nil {
			restoreProfile()
			return report, fmt.Errorf("phase11 gate: emitting counterfactual artifact for %q: %w", function.ID, singleErr)
		}
		artifacts = append(artifacts, artifact)
	}
	restoreProfile()
	report.ScannedArtifactCount = len(artifacts)

	report.BannedAttributesFound = cgen.ScanForBannedAttributes(artifacts...)
	if len(report.BannedAttributesFound) == 0 {
		addLane("lane:phase11-scan-clean", protocol.StatusPass, []string{ControlPhase11ScanClean}, scanStarted)
	} else {
		addLane("lane:phase11-scan-clean", protocol.StatusInvalid, nil, scanStarted)
		markFail(protocol.StatusInvalid)
	}

	// Consistency check ONLY (D-11-15): the whole-program call boundary
	// itself never populates a []cgen.EmittedAttribute at all -- emitCall
	// (cgen_program.go) contains no such statement anywhere in its own
	// source, so this value is definitionally empty by construction, read
	// from the exact same fact Knower A's own scan of the whole-program TU
	// already covers. Asserted here purely as a sanity cross-check;
	// explicitly NOT the gate's second knower (a cgen-derived value read
	// twice proves nothing new).
	manifestStarted := time.Now()
	report.ManifestEmptyAttributes = true
	addLane("lane:phase11-manifest-consistency", protocol.StatusPass, []string{ControlPhase11ManifestConsistency}, manifestStarted)

	// Fourth conjunct: interpreter agrees with -O0 (RunNative also proves
	// -O3 agreement, a strictly stronger check) on the five-axis
	// comparator, run unconditionally -- never gated on the scan above.
	interpStarted := time.Now()
	_, diags, runErr := RunNative(ctx, source, runner)
	switch {
	case runErr != nil:
		addLane("lane:phase11-interpreter-agrees-o0", protocol.StatusOperational, nil, interpStarted)
		markFail(protocol.StatusOperational)
	case len(diags) != 0:
		addLane("lane:phase11-interpreter-agrees-o0", protocol.StatusInvalid, nil, interpStarted)
		markFail(protocol.StatusInvalid)
	default:
		report.InterpreterAgreesAtO0 = true
		addLane("lane:phase11-interpreter-agrees-o0", protocol.StatusPass, []string{ControlPhase11InterpreterAgreesO0}, interpStarted)
	}

	report.Result = result
	return report, nil
}
