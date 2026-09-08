package session

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/pathoracle"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// Phase7RequiredControls is the complete Phase 07 required-control list as
// of plan 07-04 (D-07-41), mirroring Phase4RequiredControls' /
// Phase5RequiredControls' shape verbatim: a flat slice of exact identifier
// strings, one per line, no computation. scripts/verify-phase7.sh
// duplicates this list VERBATIM, and TestPhase7RequiredControlsMatchScript
// asserts the two are set-equal in both directions, so the gate cannot
// silently shrink by dropping a control from either copy (T-07-24).
//
// All three identifiers are fired together by this file's own
// lane:kind-exhaustive-dispatch-phase07 (VerifyPhase7ControlsAndWork): the
// CLI-observable lane drives the SAME phase-07 fixtures and the SAME
// recognized-not-executed assertion core_test.go's in-process control
// (control:kind.exhaustive_dispatch.phase07_in_process) and Task 3's
// mutation-kill test (control:dispatch.recognized_not_executed) exercise
// in-process -- so a CLI run of this lane genuinely proves all three
// claims, not merely the lane's own identity
// (control:kind.exhaustive_dispatch.phase07_lane). Recording all three
// here, rather than only the lane's own identifier, is what lets 07-07's
// phase-wide completeness matrix compare against a single authoritative
// list without a second, divergent registry.
func Phase7RequiredControls() []string {
	return []string{
		ControlKindExhaustiveDispatchPhase07InProcess,
		ControlKindExhaustiveDispatchPhase07Lane,
		ControlDispatchRecognizedNotExecuted,
	}
}

// Control identifiers introduced by plan 07-04 (D-07-41), following the
// session-package's control:<area>.<specific> naming convention.
const (
	// ControlKindExhaustiveDispatchPhase07InProcess names
	// internal/compiler/core/core_test.go's
	// TestAllOperationKindsHandledAtEverySite fixture list once it carries
	// core.OpCall: the in-process control (A-05) exercised via `go test`,
	// never repurposing Phase 4's own control:kind.exhaustive_dispatch
	// identifier or fixture list (T-07-23).
	ControlKindExhaustiveDispatchPhase07InProcess = "control:kind.exhaustive_dispatch.phase07_in_process"
	// ControlKindExhaustiveDispatchPhase07Lane names this file's own
	// CLI-observable lane (`lang verify testdata/phase07`), the phase-07
	// sibling of Phase 4's lane:kind-exhaustive-dispatch, with its own
	// distinct lane ID and its own phase07DispatchFixtures list.
	ControlKindExhaustiveDispatchPhase07Lane = "control:kind.exhaustive_dispatch.phase07_lane"
	// ControlDispatchRecognizedNotExecuted names the D-07-39 claim both
	// controls assert: a function containing core.OpCall is RECOGNIZED at
	// the interp site (interp.Run returns the named interp.ErrCallUnsupported,
	// never a crash and never a silently faked success) but Phase 07 defines
	// no call-stack execution semantics for it. Task 3's mutation-kill test
	// (TestPhase7DispatchControlsMutationKilled) proves this assertion is
	// load-bearing: folding core.OpCall into a grouped copy/move/borrow arm
	// makes it go red.
	ControlDispatchRecognizedNotExecuted = "control:dispatch.recognized_not_executed"
)

// phase07DispatchFixtures is the literal, hand-maintained, phase-scoped
// fixture list (A-05) this lane drives through corevalidate, pathoracle,
// originvalidate, interp, and cgen -- the CLI-observable sibling of
// core_test.go's own fixtures slice. Phase 07 does not extend or reuse
// Phase 4's dispatchFixtures (session.go); it carries its own.
var phase07DispatchFixtures = []string{"call_basic.lang", "call_from_both_match_arms.lang"}

// phase07FunctionHasOpCall mirrors core_test.go's functionHasOpCall (a
// deliberate, small duplication across the in-process and CLI-observable
// controls, matching A-05's requirement that each control carry its own
// literal fixture/logic rather than share a single implementation two
// independent proofs could both silently rot alongside).
func phase07FunctionHasOpCall(function core.Function) bool {
	if function.Linear == nil {
		return false
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpCall {
			return true
		}
	}
	return false
}

// phase07LinearProbeInput mirrors core_test.go's linearProbeInput: a
// straight-line function's sole declared parameter type selects a fixed,
// hand-picked probe input string for interp.Run. Both testdata/phase07
// fixtures' straight-line functions declare a Byte parameter.
func phase07LinearProbeInput(function core.Function) (string, bool) {
	switch function.Parameter.Type {
	case "Byte":
		return "7", true
	case "Buffer":
		return "01020304", true
	default:
		return "", false
	}
}

// VerifyPhase7ControlsAndWork is Phase 07's own control-and-work gate
// (D-07-41), reachable through `lang verify testdata/phase07` via
// cmd/lang/main.go's isPhase7Corpus dispatch (mirroring isPhase5Corpus/
// isPhase6Corpus). It drives phase07DispatchFixtures through check,
// corevalidate, pathoracle, originvalidate, interp, and cgen exactly like
// core_test.go's in-process control, and requires core.OpCall to be
// encountered by at least one operation across the corpus.
//
// D-07-39 -- recognition, not execution: where a function contains a
// core.OpCall, interp.Run returning the named interp.ErrCallUnsupported is
// treated as "handled" (recognized), never as a lane failure. Any OTHER
// interp error still fails this lane.
//
// A-02 -- cgen's runtime behaviour for core.OpCall is exercised by NEITHER
// this lane NOR core_test.go's in-process control, because cgen.Emit hard-
// fails on len(program.Functions) != 1 (cgen.go:22) before its OpCall arm
// could ever run, and no legal OpCall-bearing program in this phase's
// corpus has exactly one function. This lane's own `if len(program.Functions)
// == 1` gate below is therefore never taken for either phase07 fixture --
// declared here, in PHASE-07-DEBT.md, and in 07-04-SUMMARY.md, never implied.
func VerifyPhase7ControlsAndWork(ctx context.Context) (protocol.Result, error) {
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)

	addLane := func(id, status string, controls []string, work int, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: protocol.LaneSchema1, ID: id, Status: status,
			Controls: append([]string{}, controls...), RecomputedWork: work,
			ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
		})
		result.Metrics.RecomputedWork += work
	}
	fail := func(status, code, message string) (protocol.Result, error) {
		result.Status = status
		result.Diagnostics = append(result.Diagnostics, diagnostic.Error(code, diagnostic.Span{}, message))
		result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
		return result.Finalize(), nil
	}

	corpus := nat03CorpusPath("testdata/phase07")
	laneStarted := time.Now()
	encounteredKinds := make(map[core.OperationKind]bool)
	dispatchWork := 0
	for _, fixtureName := range phase07DispatchFixtures {
		fixtureSource, fixtureErr := readBoundedFile(filepath.Join(corpus, fixtureName), syntax.MaxSourceBytes)
		if fixtureErr != nil {
			addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
			return fail(protocol.StatusOperational, "verify.fixture_missing", fixtureName)
		}
		fixtureChecked := Check(fixtureSource)
		if len(fixtureChecked.Diagnostics) != 0 {
			addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
			return fail(protocol.StatusInvalid, "verify.fixture_rejected", fixtureName)
		}
		dispatchWork += fixtureChecked.Work
		fixtureValidated := corevalidate.Validate(fixtureChecked.Program)
		if !fixtureValidated.Valid {
			addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+fixtureValidated.Checks, laneStarted)
			return fail(protocol.StatusInvalid, "verify.fixture_rejected", fixtureName)
		}
		dispatchWork += fixtureValidated.Checks
		dispatchProgram := fixtureValidated.Program()
		for _, function := range dispatchProgram.Functions {
			if function.Linear != nil {
				for _, operation := range function.Linear.Operations {
					encounteredKinds[operation.Kind] = true
				}
				if function.Linear.ID != "" {
					if _, _, oracleErr := pathoracle.RecomputeEndpoints(function); oracleErr != nil {
						addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
						return fail(protocol.StatusOperational, "verify.control_incomplete", "pathoracle dispatch error for "+fixtureName)
					}
				}
			}
			_ = originvalidate.RecomputeOriginPerReturn(function)
			dispatchWork++

			hasCall := phase07FunctionHasOpCall(function)
			switch {
			case function.Match != nil:
				for _, arm := range function.Match.Arms {
					if _, interpErr := interp.Run(dispatchProgram, function.Name, arm.Pattern); interpErr != nil {
						// D-07-39: recognized, not executed -- see the
						// doc comment above.
						if !hasCall || !errors.Is(interpErr, interp.ErrCallUnsupported) {
							addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
							return fail(protocol.StatusOperational, "verify.control_incomplete", "interp dispatch error for "+fixtureName)
						}
					}
					dispatchWork++
				}
			case function.Linear != nil:
				input, ok := phase07LinearProbeInput(function)
				if ok {
					if _, interpErr := interp.Run(dispatchProgram, function.Name, input); interpErr != nil {
						// D-07-39: recognized, not executed -- see the
						// doc comment above.
						if !hasCall || !errors.Is(interpErr, interp.ErrCallUnsupported) {
							addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
							return fail(protocol.StatusOperational, "verify.control_incomplete", "interp dispatch error for "+fixtureName)
						}
					}
					dispatchWork++
				}
			}
		}
		// A-02: this gate is never taken for either phase07 fixture -- see
		// the doc comment above.
		if len(dispatchProgram.Functions) == 1 {
			if _, cgenErr := cgen.Emit(dispatchProgram); cgenErr != nil {
				addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
				return fail(protocol.StatusOperational, "verify.control_incomplete", "cgen dispatch error for "+fixtureName)
			}
			dispatchWork++
		}
	}
	for _, kind := range []core.OperationKind{core.OpCall} {
		if !encounteredKinds[kind] {
			addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
			return fail(protocol.StatusInvalid, "verify.control_missing", "control:kind.exhaustive_dispatch.phase07_lane (kind "+string(kind)+" never encountered)")
		}
	}
	addLane("lane:kind-exhaustive-dispatch-phase07", "pass", Phase7RequiredControls(), dispatchWork+1, laneStarted)

	for _, required := range Phase7RequiredControls() {
		if !hasControl(result.Lanes, required) {
			return fail(protocol.StatusInvalid, "verify.control_missing", required)
		}
	}
	for _, lane := range result.Lanes {
		if lane.RecomputedWork == 0 {
			return fail(protocol.StatusInvalid, "verify.zero_work", lane.ID)
		}
	}
	result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
	return result.Finalize(), nil
}
