package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/protocol"
)

// Phase16ControlNativeC is the production session boundary for native C.
// The fixture parameter remains for operational-call compatibility, but cannot
// select evidence: cgen.EmitNative is the sole emission authority.
func Phase16ControlNativeC(program core.Program, fixture string) (string, error) {
	return cgen.EmitNative(program)
}

// Phase16M004RefusalFamily classifies only the two current M004 native
// emission refusal identities. Historical Phase 4/5 gates use this at the
// boundary where their formerly executable foreign fixtures are now
// refusal-first and backed by digest-bound frozen evidence.
func Phase16M004RefusalFamily(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	switch {
	case strings.Contains(message, "multi-function foreign-call bodies are not supported by native emission this phase"):
		return "foreign"
	case strings.Contains(message, "by-pointer bodies are not supported by whole-program native emission this phase"):
		return "by-pointer"
	default:
		return ""
	}
}

// Phase5RequiredControls is the complete Phase 5 required-control list as
// of plan 05-09 (D-05-17): every control this gate's own lanes fire,
// copying Phase4RequiredControls' shape verbatim -- a flat slice of exact
// identifier strings, one per line, no computation.
// scripts/verify-phase5.sh (this plan) duplicates this list VERBATIM, and
// TestPhase5RequiredControlsMatchScript asserts the two are set-equal in
// both directions, so the gate cannot silently shrink by dropping a
// control from either copy (T-05-32).
//
// control:foreign.no_unproven_attributes is plan 05-04's control, narrowed
// (not deleted) by that plan; the first ten identifiers are Phase 5
// controls introduced by plans 05-04 through 05-08. Plan 05-14 (this plan)
// closes the phase by adding the final five: the three reducer vacuity
// controls plan 05-12 built but deliberately left unwired
// (LaneMismatchReduce, session_phase5_mismatch.go), and the two QLT-01
// registry-audit controls plan 05-11 built but deliberately left unwired
// (LaneQLT01RegistryAudit, qlt01.go) -- both plans left the wiring to this
// plan precisely so the Go control set and scripts/verify-phase5.sh's own
// required-control block are extended TOGETHER, in one commit, and
// TestPhase5RequiredControlsMatchScript never observes a transiently
// divergent pair (D-05-17).
func Phase5RequiredControls() []string {
	return []string{
		"control:foreign.no_unproven_attributes",
		"control:alias.false_no_alias",
		"control:interpreter-o0-o3-lto",
		"control:diagnostic.reject_program_id_equivalence",
		"control:compare.field_routing_unrouted",
		"control:native.sanitize.retained_pointer",
		"control:native.sanitize.ubsan_no_recover",
		"control:native.sanitize.allocator_mismatch",
		"control:native.sanitize.use_after_free",
		"control:core.attribute_unjustified",
		"control:reduce.no_progress",
		"control:reduce.predicate_too_loose",
		"control:reduce.nondeterministic",
		"control:qlt01.registry_incomplete",
		"control:qlt01.stale_control_reference",
	}
}

// Phase5ExpectedEscapes is Phase 5's own declared, gate-visible residual
// set: the callback-invocation half of NAT-03's stale-callback-retention
// mutation (D-05-07), and plan 05-13's coordinated source-to-core false
// claim (D-05-30) -- neither claimed solved, neither ever permitted to
// appear as a detected lane control (TestBothPhase5EscapesAreVisible,
// TestCoordinatedLieEscapeIsNeverDetected).
func Phase5ExpectedEscapes() []string {
	return []string{EscapeCallbackInvocationUnsubjected, EscapeCoordinatedSourceToCoreFalseClaim}
}

// Phase5ClangPathOverrideForTest is a mutable seam ONLY for
// TestPhase5ToolMissingIsNamedObligationNotPass (D-05-17): the test
// substitutes a broken ClangPath to prove the sanitizer controls surface
// as named, gate-visible operational obligations -- never a silent pass --
// without widening VerifyPhase5ControlsAndWork's own plan-specified
// signature. Production callers must never set this. Empty means "use the
// default toolchain resolution", mirroring native.DefaultRunner's own
// default.
var Phase5ClangPathOverrideForTest = ""

func phase5DefaultRunner() native.Runner {
	runner := native.DefaultRunner()
	if Phase5ClangPathOverrideForTest != "" {
		runner.ClangPath = Phase5ClangPathOverrideForTest
	}
	return runner
}

// VerifyPhase5ControlsAndWork is Phase 5's own control-and-work gate
// (D-05-17), peering verifyForeignCorpus's own addLane/fail shape: it runs
// the alias-fact mutation lane, the two attribute-justification lanes
// (positive and negative, narrowing control:foreign.no_unproven_attributes
// per plan 05-04), the reject-program diagnostic-ID equivalence lane, the
// comparator's own fail-closed field-routing check, a real
// interpreter/-O0/-O3/-O3-LTO differential over one of the Phase 5
// adversarial fixtures, and the sanitizer lane -- collecting every
// control's fired status and RecomputedWork.
//
// A `tool_missing`/`sanitizer_inert` sanitizer lane is NEVER folded into a
// silent pass: when VerifyPhase5SanitizeLane itself does not reach every
// one of its own controls (an unavailable or non-instrumented sanitizer
// runtime), this function adds an explicit obligation lane naming all four
// sanitizer controls under the sanitizer lane's own reported status, so
// every declared control is always visible in the result even when it
// could not be exercised (T-05-33).
func VerifyPhase5ControlsAndWork(ctx context.Context) (protocol.Result, error) {
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)
	result.ExpectedEscapes = append([]string{}, Phase5ExpectedEscapes()...)

	addLane := func(id, status string, controls []string, work int, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: protocol.LaneSchema1, ID: id, Status: status,
			Controls: append([]string{}, controls...), RecomputedWork: work,
			ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
		})
		result.Metrics.RecomputedWork += work
	}
	markFail := func(status string) {
		if result.Status == protocol.StatusPass {
			result.Status = status
		}
	}

	runner := phase5DefaultRunner()

	// Lanes: control:foreign.no_unproven_attributes (positive, narrowed by
	// plan 05-04) and control:core.attribute_unjustified (negative,
	// D-05-03b's own falsifier), both driven off the same checked
	// restrict_borrow.schway fixture so the negative lane's corrupted claim
	// is compared against the identical program the positive lane proved
	// justified.
	attributesStarted := time.Now()
	restrictSource, err := os.ReadFile(nat03CorpusPath("testdata/phase5/restrict_borrow.schway"))
	switch {
	case err != nil:
		addLane("lane:foreign-no-unproven-attributes", protocol.StatusOperational, nil, 1, attributesStarted)
		addLane("lane:attribute-unjustified", protocol.StatusOperational, nil, 1, attributesStarted)
		markFail(protocol.StatusOperational)
	default:
		restrictChecked := Check(restrictSource)
		// Phase 11 (11-GUARD-LEDGER.md): KEPT. Bound to the fixed Phase 5
		// fixture restrict_borrow.schway, genuinely single-function by
		// construction; unrelated to the multi-function corpus this
		// phase widens.
		if len(restrictChecked.Diagnostics) != 0 || len(restrictChecked.Program.Functions) != 1 {
			addLane("lane:foreign-no-unproven-attributes", protocol.StatusInvalid, nil, 1, attributesStarted)
			addLane("lane:attribute-unjustified", protocol.StatusInvalid, nil, 1, attributesStarted)
			markFail(protocol.StatusInvalid)
		} else {
			function := restrictChecked.Program.Functions[0]
			loanID := ""
			if function.Linear != nil && len(function.Linear.Operations) > 0 {
				loanID = function.Linear.Operations[0].LoanID
			}

			validClaim := corevalidate.AttributeClaim{Attr: "restrict", CoreNode: function.ID, Parameter: function.Parameter.ID, JustifiedBy: loanID}
			if validateErr := corevalidate.ValidateEmittedAttributes(restrictChecked.Program, []corevalidate.AttributeClaim{validClaim}); validateErr != nil {
				addLane("lane:foreign-no-unproven-attributes", protocol.StatusInvalid, nil, 1, attributesStarted)
				markFail(protocol.StatusInvalid)
			} else {
				addLane("lane:foreign-no-unproven-attributes", protocol.StatusPass, []string{"control:foreign.no_unproven_attributes"}, 1, attributesStarted)
			}

			corruptClaim := corevalidate.AttributeClaim{Attr: "restrict", CoreNode: function.ID, Parameter: function.Parameter.ID, JustifiedBy: "gate-bogus-loan-id"}
			corruptErr := corevalidate.ValidateEmittedAttributes(restrictChecked.Program, []corevalidate.AttributeClaim{corruptClaim})
			var attributeErr *corevalidate.AttributeUnjustifiedError
			if corruptErr == nil || !errors.As(corruptErr, &attributeErr) || attributeErr.Code != "core.attribute_unjustified" {
				addLane("lane:attribute-unjustified", protocol.StatusInvalid, nil, 1, attributesStarted)
				markFail(protocol.StatusInvalid)
			} else {
				addLane("lane:attribute-unjustified", protocol.StatusPass, []string{"control:core.attribute_unjustified"}, 1, attributesStarted)
			}
		}
	}

	// Lane: control:alias.false_no_alias (D-05-05), reusing the exact
	// mutation runner and assertion NAT03Mutations' own row 5 cites.
	aliasStarted := time.Now()
	aliasRunner := NewAliasFactMutationRunner(runner, nat03CorpusPath("testdata/phase5/false_restrict_hoist.schway"))
	if aliasErr := VerifyAliasFalseNoAlias(ctx, aliasRunner); aliasErr != nil {
		addLane("lane:alias-false-no-alias", protocol.StatusMismatch, nil, aliasRunner.RecomputedWork()+1, aliasStarted)
		markFail(protocol.StatusMismatch)
	} else {
		addLane("lane:alias-false-no-alias", protocol.StatusPass, []string{ControlAliasFalseNoAlias}, aliasRunner.RecomputedWork()+1, aliasStarted)
	}

	// Lane: control:diagnostic.reject_program_id_equivalence (D-05-20):
	// two independent Check() runs over the same reject-program must agree
	// on the exact same diagnostic ID.
	diagnosticStarted := time.Now()
	rejectSource, rejectErr := os.ReadFile(nat03CorpusPath("testdata/phase4/foreign_call_target_not_foreign.schway"))
	if rejectErr != nil {
		addLane("lane:diagnostic-reject-program-id-equivalence", protocol.StatusOperational, nil, 1, diagnosticStarted)
		markFail(protocol.StatusOperational)
	} else {
		firstChecked := Check(rejectSource)
		secondChecked := Check(rejectSource)
		if len(firstChecked.Diagnostics) == 0 || len(secondChecked.Diagnostics) == 0 {
			addLane("lane:diagnostic-reject-program-id-equivalence", protocol.StatusInvalid, nil, 1, diagnosticStarted)
			markFail(protocol.StatusInvalid)
		} else {
			diagnostics := map[string]diagnostic.Diagnostic{"check-run-1": firstChecked.Diagnostics[0], "check-run-2": secondChecked.Diagnostics[0]}
			if compareErr := Phase5CompareDiagnosticIDs("foreign_call_target_not_foreign.schway", diagnostics); compareErr != nil {
				addLane("lane:diagnostic-reject-program-id-equivalence", protocol.StatusMismatch, nil, 2, diagnosticStarted)
				markFail(protocol.StatusMismatch)
			} else {
				addLane("lane:diagnostic-reject-program-id-equivalence", protocol.StatusPass, []string{ControlDiagnosticRejectProgramIDEquivalence}, 2, diagnosticStarted)
			}
		}
	}

	// Lane: control:compare.field_routing_unrouted (D-05-21): every field
	// reachable from execution.Execution is routed to either
	// Phase5ComparedComparisonFields or Phase5ExcludedComparisonFields --
	// never left unrouted, and never a stale entry naming a field that no
	// longer exists.
	routingStarted := time.Now()
	actualFields := reachableFieldPaths(reflect.TypeOf(execution.Execution{}), "Execution")
	unrouted := unroutedFields(actualFields, Phase5ComparedComparisonFields, Phase5ExcludedComparisonFields)
	stale := staleFields(actualFields, Phase5ComparedComparisonFields, Phase5ExcludedComparisonFields)
	if len(unrouted) != 0 || len(stale) != 0 {
		addLane("lane:compare-field-routing", protocol.StatusInvalid, nil, len(actualFields)+1, routingStarted)
		markFail(protocol.StatusInvalid)
	} else {
		addLane("lane:compare-field-routing", protocol.StatusPass, []string{"control:compare.field_routing_unrouted"}, len(actualFields)+1, routingStarted)
	}

	// Lane: control:interpreter-o0-o3-lto (D-05-19): a real
	// interpreter/-O0/-O3/-O3-LTO differential over one of the Phase 5
	// adversarial fixtures -- inline_across_foreign.schway, the fixture
	// D-05-18a specifically engineered to give `-flto` a real cross-TU
	// inlining opportunity the non-LTO tiers cannot take.
	ltoStarted := time.Now()
	ltoStatus, ltoControls, ltoWork := phase5RunInterpreterO0O3LTOLane(ctx, runner)
	addLane("lane:interpreter-o0-o3-lto", ltoStatus, ltoControls, ltoWork, ltoStarted)
	if ltoStatus != protocol.StatusPass {
		markFail(ltoStatus)
	}

	// Lane: the sanitizer lane's own four controls (D-05-08/D-05-09/
	// D-05-13/D-05-14). When VerifyPhase5SanitizeLane itself does not
	// pass -- an unavailable sanitizer runtime, or one of its own controls
	// failing to fire -- an explicit obligation lane names all four
	// sanitizer controls under the sanitizer lane's own reported status,
	// so a `tool_missing`/`sanitizer_inert` runtime is a NAMED OBLIGATION,
	// never an implicit pass (T-05-33).
	sanitizeStarted := time.Now()
	sanitizeResult, sanitizeErr := VerifyPhase5SanitizeLane(ctx, runner)
	if sanitizeErr != nil {
		return protocol.Result{}, sanitizeErr
	}
	result.Lanes = append(result.Lanes, sanitizeResult.Lanes...)
	result.Metrics.RecomputedWork += sanitizeResult.Metrics.RecomputedWork
	if sanitizeResult.Status != protocol.StatusPass {
		addLane("lane:native-sanitize-obligations", sanitizeResult.Status,
			[]string{ControlSanitizeRetainedPointer, ControlSanitizeUseAfterFree, ControlSanitizeAllocatorMismatch, ControlSanitizeUBSanNoRecover},
			1, sanitizeStarted)
		markFail(sanitizeResult.Status)
	}

	// Lane: the QLT-01 registry completeness audit (D-05-29), built by
	// plan 05-11 and deliberately left unwired until this plan so the
	// registry's two controls (control:qlt01.registry_incomplete,
	// control:qlt01.stale_control_reference) join Phase5RequiredControls()
	// in the same commit as the shell gate's own extension.
	qlt01Started := time.Now()
	qlt01Lane, qlt01Err := VerifyQLT01Registry(ctx)
	if qlt01Err != nil {
		addLane(LaneQLT01RegistryAudit, protocol.StatusOperational, nil, 1, qlt01Started)
		markFail(protocol.StatusOperational)
	} else {
		addLane(qlt01Lane.ID, qlt01Lane.Status, qlt01Lane.Controls, qlt01Lane.RecomputedWork, qlt01Started)
		if qlt01Lane.Status != protocol.StatusPass {
			markFail(qlt01Lane.Status)
		}
	}

	// Lane: the reducer's three vacuity controls (D-05-27), built by
	// plan 05-12 and deliberately left unwired until this plan
	// (LaneMismatchReduce, session_phase5_mismatch.go). Its own
	// VerifyMismatchReduceLane already returns a full protocol.Result with
	// its own lane and RecomputedWork, so it is merged directly rather
	// than re-derived through addLane.
	mismatchStarted := time.Now()
	mismatchResult, mismatchErr := VerifyMismatchReduceLane(ctx)
	if mismatchErr != nil {
		addLane(LaneMismatchReduce, protocol.StatusOperational, nil, 1, mismatchStarted)
		markFail(protocol.StatusOperational)
	} else {
		result.Lanes = append(result.Lanes, mismatchResult.Lanes...)
		result.Metrics.RecomputedWork += mismatchResult.Metrics.RecomputedWork
		if mismatchResult.Status != protocol.StatusPass {
			markFail(mismatchResult.Status)
		}
	}

	// Lane: the coordinated source-to-core false-claim escape (D-05-30),
	// declared by plan 05-13 and wired here alongside this plan's own
	// extension of Phase5ExpectedEscapes(). A pass here is attributed to
	// the named escape (never a silent absence of checking); an error
	// means the adversarial pair failed to construct as claimed, which is
	// a genuine gate failure, not an escape.
	coordinatedStarted := time.Now()
	coordinatedLane, coordinatedErr := VerifyCoordinatedLieEscape(ctx)
	if coordinatedErr != nil {
		addLane(LaneCoordinatedLieEscape, protocol.StatusOperational, nil, 1, coordinatedStarted)
		markFail(protocol.StatusOperational)
	} else {
		addLane(coordinatedLane.ID, coordinatedLane.Status, coordinatedLane.Controls, coordinatedLane.RecomputedWork, coordinatedStarted)
		if coordinatedLane.Status != protocol.StatusPass {
			markFail(coordinatedLane.Status)
		}
	}

	for _, required := range Phase5RequiredControls() {
		if !hasControl(result.Lanes, required) {
			markFail(protocol.StatusInvalid)
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("verify.control_missing", diagnostic.Span{}, required))
		}
	}
	for _, lane := range result.Lanes {
		if lane.RecomputedWork == 0 {
			markFail(protocol.StatusInvalid)
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("verify.zero_work", diagnostic.Span{}, lane.ID))
		}
	}

	result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
	return result.Finalize(), nil
}

// phase5RunInterpreterO0O3LTOLane drives inline_across_foreign.schway
// through the interpreter and three native builds -- -O0, -O3, and
// -O3 with LTO -- and compares all four via Phase5CompareEngines. Split
// out of VerifyPhase5ControlsAndWork as its own function so a failure at
// any step degrades to a single reported lane status rather than aborting
// the whole gate.
func phase5RunInterpreterO0O3LTOLane(ctx context.Context, runner native.Runner) (status string, controls []string, work int) {
	const fixture = "inline_across_foreign.schway"
	source, err := os.ReadFile(nat03CorpusPath("testdata/phase5/" + fixture))
	if err != nil {
		return protocol.StatusOperational, nil, 1
	}
	checked := Check(source)
	if len(checked.Diagnostics) != 0 {
		return protocol.StatusInvalid, nil, 1
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return protocol.StatusInvalid, nil, 1
	}
	program := validated.Program()
	// Phase 11 (11-GUARD-LEDGER.md): KEPT. Bound to the fixed Phase 5
	// fixture inline_across_foreign.schway, genuinely single-function by
	// construction -- its own cross-TU LTO opportunity is the foreign
	// boundary, not the multi-function corpus this phase widens. The
	// comparator this lane calls (Phase5CompareEngines) is unaffected;
	// only this caller's own admission stays scoped to one function.
	if len(program.Functions) != 1 {
		return protocol.StatusInvalid, nil, 1
	}
	functionName := program.Functions[0].Name

	interpreted, err := interp.Run(program, functionName, "7")
	if err != nil {
		return protocol.StatusOperational, nil, 1
	}
	cSource, err := Phase16ControlNativeC(program, "testdata/phase5/inline_across_foreign.schway")
	if err != nil {
		return protocol.StatusOperational, nil, 1
	}

	nativeRunner := runner
	nativeRunner.Expect = expectForOutcomeKind(interpreted.Outcome.Kind)
	for _, function := range program.Functions {
		if function.Name != functionName || function.ForeignContract == nil {
			continue
		}
		if sourcePaths := native.ForeignSourcePathsForSymbol(function.ForeignContract.Symbol); len(sourcePaths) != 0 {
			nativeRunner.ForeignSources = append(append([]string(nil), nativeRunner.ForeignSources...), sourcePaths...)
		}
	}

	o0, err := nativeRunner.Run(ctx, cSource, "-O0", []string{"7"})
	if err != nil || len(o0.Pairs) != 1 {
		return protocol.StatusOperational, nil, 1
	}
	o3, err := nativeRunner.Run(ctx, cSource, "-O3", []string{"7"})
	if err != nil || len(o3.Pairs) != 1 {
		return protocol.StatusOperational, nil, 2
	}
	ltoRunner := nativeRunner
	ltoRunner.LTO = true
	o3lto, err := ltoRunner.Run(ctx, cSource, "-O3", []string{"7"})
	if err != nil || len(o3lto.Pairs) != 1 {
		return protocol.StatusOperational, nil, 3
	}

	engines := map[string]execution.Execution{
		"interpreter": interpreted,
		"O0":          o0.Pairs[0].Execution,
		"O3":          o3.Pairs[0].Execution,
		"O3-LTO":      o3lto.Pairs[0].Execution,
	}
	if compareErr := Phase5CompareEngines(fixture, engines); compareErr != nil {
		return protocol.StatusMismatch, nil, 4
	}
	return protocol.StatusPass, []string{"control:interpreter-o0-o3-lto"}, 4
}

// assertAllocatorMismatchMovesAxis proves control:native.sanitize.allocator_mismatch
// against its cited fixture (testdata/phase5/allocator_mismatch.schway):
// ASan's own alloc-dealloc-mismatch report is a genuine divergence from a
// clean terminal outcome so extreme no comparable execution.Execution
// document is ever produced at all -- mapped onto axis:terminal-outcome,
// the same mapping assertLayoutMismatchMovesAxis uses for an equally
// total refusal (a normal return vs. an ASan-terminated process).
func assertAllocatorMismatchMovesAxis(ctx context.Context, mutation NAT03Mutation) error {
	cSource, err := compilePhase5SanitizeFixture("testdata/phase5/allocator_mismatch.schway")
	if err != nil {
		return err
	}
	runner := phase5DefaultRunner()
	runner.ForeignSources = native.ForeignLegacyArenaSourcePaths()
	report, runErr := runner.RunSanitized(ctx, cSource, []string{"7"})
	wantSignature := sanitizerSignatureFor("alloc-dealloc-mismatch")
	if runErr != nil || report.ExitCode == 0 || report.ReportSignature != wantSignature {
		return fmt.Errorf("%s produced no divergence: allocator-mismatch fixture did not report %q (exit=%d, check_kind=%q, err=%v)",
			mutation.ControlID, wantSignature, report.ExitCode, report.CheckKind, runErr)
	}
	return requireAxis(mutation, AxisTerminalOutcome)
}
