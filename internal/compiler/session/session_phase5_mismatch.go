// session_phase5_mismatch.go wires plan 05-10's dependency-free
// internal/compiler/reduce package into a real, seeded interpreter
// mismatch, producing a reduce.MismatchDocument (schway.mismatch/0, D-05-26).
// SignatureFromDisagreement lives HERE, not in reduce, because reduce must
// stay a leaf package (05-10-SUMMARY.md): this plan is the session -> reduce
// wiring direction, and a reduce -> session dependency in the other
// direction would create an import cycle.
package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/reduce"
)

// LaneMismatchReduce and its three control identifiers are D-05-27's own
// vocabulary: a reducer has exactly three distinct ways to be vacuous
// (no-op, predicate-too-loose, non-deterministic), and each control names
// one. Plan 05-14 -- not this plan -- adds these to
// Phase5RequiredControls(): a declared-but-not-yet-gate-wired control is
// exactly the "not run rendered as pass" shape D-05-17 forbids, so this
// plan deliberately stops short of that wiring (matching 05-11's own QLT-01
// registry precedent).
const LaneMismatchReduce = "lane:mismatch-reduce"

const (
	ControlReduceNoProgress        = "control:reduce.no_progress"
	ControlReducePredicateTooLoose = "control:reduce.predicate_too_loose"
	ControlReduceNondeterministic  = "control:reduce.nondeterministic"
	// ControlReduceReverified is D-11-37's fourth control, alongside the
	// three above: QLT-05's strict, cold-start re-verification (see
	// QLT05Reverify) reports this control on LaneMismatchReduce whenever
	// re-verification actually ran, regardless of its verdict -- the
	// control names that the CHECK executed, matching this file's own
	// "declared, not yet gate-wired" precedent (see the comment above)
	// rather than double-encoding pass/fail into the control set itself.
	// Declaring this control never bumps MismatchDocument's own schema
	// (schway.mismatch/0 stays pinned, D-05-39): QLT-05 is a gate claim
	// about the reducer's own output, not a document field.
	ControlReduceReverified = "control:reduce.reverified"
)

// mismatchReduceFixture is the seeded, reproducible mismatch source this
// wiring reduces: plan 05-07's AliasFactMutationRunner over
// false_restrict_hoist.schway, engineered specifically to produce a real
// interpreter-optimizer-observable divergence (D-05-05) -- a ready source
// of a genuine mismatch, per this plan's own <read_first> note.
const mismatchReduceFixture = "testdata/phase5/false_restrict_hoist.schway"

// SignatureFromDisagreement maps a real *Phase5EngineDisagreement into a
// reduce.Signature (D-05-24) -- the wiring plan 05-10 deferred to this
// plan. foreignCallSequence is supplied by the caller (see
// foreignCallSequenceFor) rather than derived here, since a
// Phase5EngineDisagreement carries no foreign-call-sequence fact of its
// own.
func SignatureFromDisagreement(disagreement *Phase5EngineDisagreement, foreignCallSequence []string) reduce.Signature {
	return reduce.Signature{
		Axis:                disagreement.Axis,
		OperationID:         disagreement.OperationID,
		EnginePair:          disagreement.EnginePair,
		ForeignCallSequence: append([]string(nil), foreignCallSequence...),
	}
}

// staticForeignCallSequenceFor is Knower 1 (D-11-33): it derives the
// ordered foreign symbol invocation sequence across the WHOLE program by
// walking every function -- never just program.Functions[0] -- in
// callgraph.Order's own deterministic reverse-postorder over the
// proven-acyclic call graph (session may import callgraph; reduce may
// not). It reads only program's own declared structure, never an
// execution document. This is the fix for the silent-slippage hole this
// plan closes: the old single-function-only walk returned nil for ANY
// multi-function program, so a dropped call to a callee containing a
// foreign acquisition compared nil to nil and was invisible.
func staticForeignCallSequenceFor(program core.Program) ([]string, error) {
	order, err := callgraph.Order(program)
	if err != nil {
		return nil, fmt.Errorf("mismatch-reduce: deriving static foreign-call sequence: %w", err)
	}
	byID := make(map[string]core.Function, len(program.Functions))
	for _, function := range program.Functions {
		byID[function.ID] = function
	}
	var sequence []string
	for _, functionID := range order {
		function, ok := byID[functionID]
		if !ok || function.ForeignContract == nil || function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpForeignCall {
				sequence = append(sequence, function.ForeignContract.Symbol)
			}
		}
	}
	return sequence, nil
}

// dynamicForeignCallSequenceFor is Knower 2 (D-11-33): the SAME fact,
// independently derived from a real -O0 run's own event stream, which
// already carries FunctionID per event ("foreign.called", shared verbatim
// across the interpreter/native engines -- see execution.Event). It shares
// no helper with staticForeignCallSequenceFor beyond the execution.Event
// type itself: it reads the execution document's own chronological event
// order, never program's declared function order, and never calls the
// static walk.
func dynamicForeignCallSequenceFor(program core.Program, run execution.Execution) []string {
	symbolByFunctionID := make(map[string]string, len(program.Functions))
	for _, function := range program.Functions {
		if function.ForeignContract != nil {
			symbolByFunctionID[function.ID] = function.ForeignContract.Symbol
		}
	}
	var sequence []string
	for _, event := range run.Events {
		if event.Kind != foreignCalledEventKind {
			continue
		}
		if symbol, ok := symbolByFunctionID[event.FunctionID]; ok {
			sequence = append(sequence, symbol)
		}
	}
	return sequence
}

// foreignCalledEventKind is the event.Kind every engine (interp, native)
// emits for one completed foreign acquisition -- see
// internal/compiler/interp/interp.go's own "foreign.called" event and
// native.go's event-kind allowlist, which both share this literal.
const foreignCalledEventKind = "foreign.called"

// foreignCallSequenceDisagreementForTest is this plan's own fault-injection
// seam (mirrors this repository's testOnlyMoves/grayVsVisitedMutationForTest
// shape, D-07-42): when non-nil, it perturbs the DYNAMIC derivation's
// result immediately before foreignCallSequenceFor's own comparison runs,
// proving the agreement check is load-bearing rather than trivially true.
// Unexported, nil in production, set only by a same-package test
// (export_test.go) that defers the restore immediately.
var foreignCallSequenceDisagreementForTest func([]string) []string

// foreignCallSequenceFor is D-11-33's two-knower guard against reduction
// slippage: the ordered foreign symbol invocation sequence, derived
// TWICE independently (staticForeignCallSequenceFor, program structure;
// dynamicForeignCallSequenceFor, the -O0 run's own event stream) and
// compared. A candidate that silently dropped a call to a foreign-
// acquiring callee -- or a genuine reducer bug that perturbed one
// derivation but not the other -- fails this comparison rather than
// being compared nil-to-nil, which is exactly the slippage hole this
// plan closes ("never one derivation read twice" applied to the one fact
// that guards against it). nil, nil for a program with no ForeignContract
// anywhere (e.g. this plan's own seed fixture, a plain borrow chain),
// matching Signature's own documented zero-value semantics.
func foreignCallSequenceFor(program core.Program, o0Run execution.Execution) ([]string, error) {
	static, err := staticForeignCallSequenceFor(program)
	if err != nil {
		return nil, err
	}
	dynamic := dynamicForeignCallSequenceFor(program, o0Run)
	if foreignCallSequenceDisagreementForTest != nil {
		dynamic = foreignCallSequenceDisagreementForTest(dynamic)
	}
	if !equalStringSlices(static, dynamic) {
		return nil, fmt.Errorf("reduce.foreign_call_sequence_disagreement: static derivation %v disagrees with dynamic derivation %v -- a real reducer bug, never silently ignored", static, dynamic)
	}
	return static, nil
}

// equalStringSlices reports whether a and b hold the same strings in the
// same order -- treating a nil slice and an empty slice as equal, which is
// the comparison foreignCallSequenceFor's own agreement check needs (a
// program with no foreign acquisitions anywhere yields nil from both
// derivations).
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// causalChainFor derives the ordered acquisition/borrow/control-edge steps
// a repair agent needs (D-05-26): every borrow, foreign acquisition, and
// foreign release operation in the reduced function's own Linear body, in
// program order.
func causalChainFor(program core.Program) []reduce.CausalStep {
	if len(program.Functions) != 1 || program.Functions[0].Linear == nil {
		return nil
	}
	function := program.Functions[0]
	var steps []reduce.CausalStep
	for _, operation := range function.Linear.Operations {
		switch operation.Kind {
		case core.OpBorrowShared:
			steps = append(steps, reduce.CausalStep{Kind: "borrow_shared", OperationID: operation.ID, Place: operation.SourceID})
		case core.OpBorrowExclusive:
			steps = append(steps, reduce.CausalStep{Kind: "borrow_exclusive", OperationID: operation.ID, Place: operation.SourceID})
		case core.OpForeignCall:
			steps = append(steps, reduce.CausalStep{Kind: "foreign_acquire", OperationID: operation.ID, Place: operation.SourceID})
		case core.OpRelease:
			steps = append(steps, reduce.CausalStep{Kind: "foreign_release", OperationID: operation.ID, Place: operation.SourceID})
		}
	}
	return steps
}

// projectEvents narrows a real engine's []execution.Event into
// []reduce.EventRecord (T-05-47) -- reduce.EventWindow's own bounded,
// addressable projection, never the full execution.Event schema.
func projectEvents(events []execution.Event) []reduce.EventRecord {
	records := make([]reduce.EventRecord, 0, len(events))
	for _, event := range events {
		records = append(records, reduce.EventRecord{
			Kind: event.Kind, OperationID: event.ID, FunctionID: event.FunctionID,
			Detail: fmt.Sprintf("source=%s target=%s type=%s", event.SourcePlace, event.TargetPlace, event.TypeID),
		})
	}
	return records
}

// mismatchEvidenceID computes a content-derived identifier binding a
// MismatchDocument back to the exact reduced core and axis/engine-pair it
// was produced from (the additive evidence_id field -- Claude's Discretion,
// see 05-12-SUMMARY.md). It follows this project's own "content identity
// only" discipline (evidence.DigestClaim): a pure SHA-256 over the
// reduced program plus the axis/engine-pair facts, never a timestamp or
// address.
func mismatchEvidenceID(program core.Program, axis, enginePair string) string {
	identity := struct {
		Schema     string
		Program    core.Program
		Axis       string
		EnginePair string
	}{Schema: reduce.MismatchSchema, Program: program, Axis: axis, EnginePair: enginePair}
	encoded, err := json.Marshal(identity)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(encoded)
	return "evidence:sha256-v1:" + hex.EncodeToString(sum[:16])
}

// mismatchPredicate builds the real reduce.Predicate ReduceSeededAliasMismatch
// drives Reduce with: for each candidate core.Program, independently
// re-derive whether it STILL reproduces a real, observable mismatch (via
// the SAME corevalidate -> cgen -> AliasFactMutationRunner -> compare
// pipeline the seed itself went through), and apply D-05-24's real
// Interesting conjunction against seed. A candidate that fails to
// re-validate, fails to emit, or loses the mutation runner's own marker
// (e.g. because a move changed cgen's lowering selection) is REJECTED as
// uninteresting -- never a fatal error -- since that is exactly what
// "candidate stopped reproducing the bug" means for this reducer.
func mismatchPredicate(baseRunner native.Runner, fixturePath, fallbackInput string, seed reduce.Signature) reduce.Predicate {
	return func(ctx context.Context, candidate core.Program) (reduce.Signature, bool, error) {
		if err := ctx.Err(); err != nil {
			return reduce.Signature{}, false, err
		}
		validated := corevalidate.Validate(candidate)
		if !validated.Valid {
			return reduce.Signature{}, false, nil
		}
		candidateProgram := validated.Program()
		if len(candidateProgram.Functions) != 1 {
			return reduce.Signature{}, false, nil
		}
		cSource, err := cgen.EmitNative(candidateProgram)
		if err != nil {
			// A reduced M004 candidate has altered canonical bytes and therefore
			// cannot borrow any digest-bound frozen artifact. Its only valid
			// native observation is the public refusal; it is not a reducer hit.
			if strings.Contains(err.Error(), "by-pointer bodies are not supported") || strings.Contains(err.Error(), "foreign-call bodies are not supported") {
				return reduce.Signature{}, false, nil
			}
			return reduce.Signature{}, false, nil
		}
		input := fallbackInput
		if candidateInputs, ok := interpreterInputs(candidateProgram); ok && len(candidateInputs) > 0 {
			input = candidateInputs[0]
		}
		candidateRunner := NewAliasFactMutationRunner(baseRunner, fixturePath)
		o0, err := candidateRunner.Run(ctx, cSource, "-O0", []string{input})
		if err != nil {
			return reduce.Signature{}, false, nil
		}
		o3, err := candidateRunner.Run(ctx, cSource, "-O3", []string{input})
		if err != nil {
			return reduce.Signature{}, false, nil
		}
		if len(o0.Pairs) != 1 || len(o3.Pairs) != 1 {
			return reduce.Signature{}, false, nil
		}
		engines := map[string]execution.Execution{"O0": o0.Pairs[0].Execution, "O3": o3.Pairs[0].Execution}
		compareErr := Phase5CompareEngines("mismatch-reduce-candidate", engines)
		if compareErr == nil {
			return reduce.Signature{}, false, nil
		}
		disagreement, ok := compareErr.(*Phase5EngineDisagreement)
		if !ok {
			return reduce.Signature{}, false, nil
		}
		candidateForeignSequence, err := foreignCallSequenceFor(candidateProgram, o0.Pairs[0].Execution)
		if err != nil {
			// A disagreement between the two independent foreign-call-
			// sequence derivations is a real reducer bug (D-11-33), never a
			// silently-rejected candidate -- surface it as a hard failure.
			return reduce.Signature{}, false, err
		}
		candidateSignature := SignatureFromDisagreement(disagreement, candidateForeignSequence)
		return candidateSignature, reduce.Interesting(seed, candidateSignature), nil
	}
}

// ReduceSeededAliasMismatch drives plan 05-07's AliasFactMutationRunner
// over false_restrict_hoist.schway to a REAL -O0-vs--O3 divergence, then
// reduces it via reduce.Reduce, emitting a schway.mismatch/0 document that is
// self-sufficient for an AI repair agent (D-05-26): its causal_chain and
// event_window are populated from the SAME seeded run that produced the
// divergence, never fabricated.
func ReduceSeededAliasMismatch(ctx context.Context) (reduce.MismatchDocument, error) {
	fixturePath := nat03CorpusPath(mismatchReduceFixture)
	source, err := os.ReadFile(fixturePath)
	if err != nil {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: reading fixture %s: %w", fixturePath, err)
	}
	checked := Check(source)
	if len(checked.Diagnostics) != 0 {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: fixture failed to check: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: fixture rejected by corevalidate: %+v", validated.Problems)
	}
	program := validated.Program()
	if len(program.Functions) != 1 {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: fixture must declare exactly one function")
	}

	baseRunner := phase5DefaultRunner()
	seedRunner := NewAliasFactMutationRunner(baseRunner, fixturePath)
	cSource, err := Phase16ControlNativeC(program, fixturePath)
	if err != nil {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: emitting C: %w", err)
	}
	nativeInputs, ok := interpreterInputs(program)
	if !ok || len(nativeInputs) == 0 {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: fixture has no derivable native input")
	}
	nativeInput := nativeInputs[0]

	o0, err := seedRunner.Run(ctx, cSource, "-O0", []string{nativeInput})
	if err != nil {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: -O0 run: %w", err)
	}
	o3, err := seedRunner.Run(ctx, cSource, "-O3", []string{nativeInput})
	if err != nil {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: -O3 run: %w", err)
	}
	if len(o0.Pairs) != 1 || len(o3.Pairs) != 1 {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: expected exactly one execution per optimization level, got -O0=%d -O3=%d", len(o0.Pairs), len(o3.Pairs))
	}

	engines := map[string]execution.Execution{"O0": o0.Pairs[0].Execution, "O3": o3.Pairs[0].Execution}
	compareErr := Phase5CompareEngines("false_restrict_hoist.schway", engines)
	disagreement, ok := compareErr.(*Phase5EngineDisagreement)
	if compareErr == nil || !ok {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: %s produced no real divergence to reduce -- the seed mutation was not exercised", ControlAliasFalseNoAlias)
	}

	seedForeignSequence, err := foreignCallSequenceFor(program, o0.Pairs[0].Execution)
	if err != nil {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: %w", err)
	}
	seedSignature := SignatureFromDisagreement(disagreement, seedForeignSequence)
	predicate := mismatchPredicate(baseRunner, fixturePath, nativeInput, seedSignature)
	entryFunction, err := callgraph.EntryFunction(program)
	if err != nil {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: resolving entry function: %w", err)
	}
	result, err := reduce.Reduce(ctx, reduce.Seed{Program: program, EntryFunctionID: entryFunction.ID}, predicate)
	if err != nil {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: %w", err)
	}

	eventWindow := map[string][]reduce.EventRecord{
		"O0": projectEvents(o0.Pairs[0].Execution.Events),
		"O3": projectEvents(o3.Pairs[0].Execution.Events),
	}
	causes := []diagnostic.Cause{{Kind: "reduce.diverging_alias_claim", Detail: disagreement.Detail}}
	evidenceID := mismatchEvidenceID(result.Program, disagreement.Axis, disagreement.EnginePair)

	return reduce.NewMismatchDocument(
		disagreement.Axis, disagreement.EnginePair, disagreement.OperationID,
		result.Program, result.Source, result.Minimality,
		result.TotalRecomputedWork, result.Attempts,
		eventWindow, causalChainFor(result.Program), causes, evidenceID,
	)
}

// VerifyMismatchReduceLane wraps ReduceSeededAliasMismatch in this file's
// own protocol.Lane, reporting nonzero RecomputedWork equal to the
// reduction's own TotalRecomputedWork (this plan's own instruction).
// NOT wired into VerifyPhase5ControlsAndWork/Phase5RequiredControls by
// this plan -- plan 05-14 is the designated point that extends the gate
// with the reducer and QLT-01 registry controls together (matching 05-11's
// own QLT-01 registry precedent of leaving gate-wiring to its own plan).
func VerifyMismatchReduceLane(ctx context.Context) (protocol.Result, error) {
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)

	document, err := ReduceSeededAliasMismatch(ctx)
	if err != nil {
		result.Status = protocol.StatusMismatch
		result.Diagnostics = append(result.Diagnostics, diagnostic.Error("verify.mismatch_reduce_failed", diagnostic.Span{}, err.Error()))
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: protocol.LaneSchema1, ID: LaneMismatchReduce, Status: protocol.StatusMismatch,
			RecomputedWork: 1, ElapsedNS: time.Since(started).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
		})
		result.Metrics.RecomputedWork++
		result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
		return result.Finalize(), nil
	}

	work := document.TotalRecomputedWork
	if work == 0 {
		work = 1
	}
	result.Lanes = append(result.Lanes, protocol.Lane{
		Schema: protocol.LaneSchema1, ID: LaneMismatchReduce, Status: protocol.StatusPass,
		Controls:       []string{ControlReduceNoProgress, ControlReducePredicateTooLoose, ControlReduceNondeterministic},
		RecomputedWork: work, ElapsedNS: time.Since(started).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
	})
	result.Metrics.RecomputedWork += work
	result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
	return result.Finalize(), nil
}

// ---------------------------------------------------------------------
// QLT-05 strict, cold-start re-verification (D-11-32/D-11-37).
// ---------------------------------------------------------------------

// QLT05SignatureDeriver computes a FRESH reduce.Signature for a candidate
// program, entirely independently of anything Reduce's own search already
// computed: ok reports whether the candidate reproduces ANY observable
// divergence at all (a different failure mode from a strict-field
// mismatch, which QLT05Reverify itself decides). Production wiring
// (qlt05RealSignatureDeriver) mirrors mismatchPredicate's own real
// validate/emit/compile/run pipeline byte for byte, but is invoked
// completely independently -- it never calls mismatchPredicate, never
// consults reduce.Interesting, and shares no cached artifact with the
// search that produced the reduced program in the first place. This is
// the exported seam a caller supplies QLT05Reverify with; it is exported
// only so this file's own doc comment can name it as the trust boundary,
// not because a second production implementation is expected.
type QLT05SignatureDeriver func(ctx context.Context, candidate core.Program) (reduce.Signature, bool, error)

// QLT05Reverification is QLT-05's strict, cold-start re-verification
// verdict (D-11-32): Verified is true only when EVERY strictly-compared
// field agrees; Reason is empty on success and otherwise names which
// field drifted, always prefixed with the reduce.reverification_signature_drift
// refusal ID reduce.RefusedShapes() already reserves for this exact
// purpose (11-08-SUMMARY.md). Control is always ControlReduceReverified
// once re-verification actually ran, independently of Verified -- the
// control names that the CHECK executed, matching this file's own
// "control declared, not double-encoding pass/fail" convention.
type QLT05Reverification struct {
	Verified bool
	Reason   string
	Control  string
}

// qlt05ReverificationSignatureDrift is the stable refusal ID
// reduce.RefusalReverificationSignatureDrift names verbatim
// (reduce.reverification_signature_drift, reduce.go's own reserved
// constant) -- session cannot reference reduce's constant directly without
// creating a second spelling risk, so it is duplicated here as a plain
// string literal, matching this file's own established
// byPointerParamMarker-style verbatim-duplication precedent
// (session_phase5_alias.go).
const qlt05ReverificationSignatureDrift = "reduce.reverification_signature_drift"

// QLT05Reverify implements QLT-05's strict, cold-start re-verification
// (D-11-32): it re-derives a candidate's Signature via deriver -- a FRESH
// check+validate+emit+compile+run cycle, never any artifact or verdict
// computed during Reduce's own search -- and compares it against seed by
// STRICT FIELD EQUALITY: Axis, EnginePair, OperationID, and the foreign-
// call sequence (Task 1's two-knower guard). There is deliberately no
// fallback to a shifted-position acceptance path here.
//
// predicate.go's Interesting (predicate.go lines 63-66) permits exactly
// one such fallback DURING the search, to tolerate an operation whose
// POSITION shifted under reduction but whose role in the program did not
// -- that relaxation is what lets the search keep narrowing past a move
// that renumbers or repositions the diverging operation. Inheriting it
// HERE, at re-verification, would satisfy QLT-05 by construction: any
// candidate the search already accepted trivially carries a matching
// role fact, so re-checking it under the SAME relaxation proves nothing
// new. Forbidding the fallback at re-verification is SAFE, not merely
// stricter-for-its-own-sake: every operation ID in this project is
// function-ID-prefixed (the "…:fn:main:op:0" form), and neither
// whole-program move (drop-call-site, drop-orphan-function) renumbers a
// surviving operation's own ID (11-08-SUMMARY.md's own
// TestOperationIDsUnchangedByWholeProgramMoves). A genuine ID shift is
// therefore structurally impossible under this reducer's own moves, so
// observing one here can only mean a real reducer bug -- reported as
// qlt05ReverificationSignatureDrift, never silently tolerated.
//
// cacheRoot is accepted ONLY to make QLT-05's cold-start property
// demonstrable in a test (TestQLT05ReverificationIsColdStart): this
// function and deriver's own real production wiring never read or write
// anything under it. Every fact QLT05Reverify's own comparison consults
// is freshly recomputed inside deriver's own call, on this call alone --
// there is no persistent cache anywhere on this path to consult in the
// first place (native.Runner's own compile step always uses a fresh
// os.MkdirTemp per run, never a shared directory).
func QLT05Reverify(ctx context.Context, seed reduce.Signature, reducedProgram core.Program, deriver QLT05SignatureDeriver, cacheRoot string) (QLT05Reverification, error) {
	_ = cacheRoot // never read; see doc comment above.
	candidate, ok, err := deriver(ctx, reducedProgram)
	if err != nil {
		return QLT05Reverification{}, err
	}
	if !ok {
		return QLT05Reverification{
			Verified: false,
			Reason:   qlt05ReverificationSignatureDrift + ": reduced program no longer reproduces any observable divergence",
			Control:  ControlReduceReverified,
		}, nil
	}
	if seed.Axis == "" || seed.Axis != candidate.Axis {
		return QLT05Reverification{Verified: false, Reason: qlt05ReverificationSignatureDrift + ": Axis differs", Control: ControlReduceReverified}, nil
	}
	if seed.EnginePair == "" || seed.EnginePair != candidate.EnginePair {
		return QLT05Reverification{Verified: false, Reason: qlt05ReverificationSignatureDrift + ": EnginePair differs", Control: ControlReduceReverified}, nil
	}
	// STRICT field equality on OperationID alone -- see the doc comment
	// above for why the search-time shifted-position fallback is both
	// forbidden and safe to forbid here.
	if seed.OperationID == "" || seed.OperationID != candidate.OperationID {
		return QLT05Reverification{Verified: false, Reason: qlt05ReverificationSignatureDrift + ": OperationID differs (no positional fallback permitted at re-verification)", Control: ControlReduceReverified}, nil
	}
	if !equalStringSlices(seed.ForeignCallSequence, candidate.ForeignCallSequence) {
		return QLT05Reverification{Verified: false, Reason: qlt05ReverificationSignatureDrift + ": ForeignCallSequence differs", Control: ControlReduceReverified}, nil
	}
	return QLT05Reverification{Verified: true, Control: ControlReduceReverified}, nil
}

// qlt05RealSignatureDeriver builds a production QLT05SignatureDeriver that
// mirrors mismatchPredicate's own real pipeline (corevalidate.Validate,
// cgen.EmitNative, the same runner's -O0/-O3 pair, Phase5CompareEngines)
// but is invoked completely independently: it never calls
// mismatchPredicate itself and applies NO reduce.Interesting conjunction
// -- ok is false only when the candidate fails to validate, fails to
// emit or compile, or produces no real engine disagreement at all;
// QLT05Reverify's own strict comparison is solely responsible for
// deciding whether an observed disagreement still matches seed.
func qlt05RealSignatureDeriver(runner native.Runner, fixturePath, fallbackInput string) QLT05SignatureDeriver {
	return func(ctx context.Context, candidate core.Program) (reduce.Signature, bool, error) {
		if err := ctx.Err(); err != nil {
			return reduce.Signature{}, false, err
		}
		validated := corevalidate.Validate(candidate)
		if !validated.Valid {
			return reduce.Signature{}, false, nil
		}
		candidateProgram := validated.Program()
		cSource, err := cgen.EmitNative(candidateProgram)
		if err != nil {
			return reduce.Signature{}, false, nil
		}
		input := fallbackInput
		if candidateInputs, ok := interpreterInputs(candidateProgram); ok && len(candidateInputs) > 0 {
			input = candidateInputs[0]
		}
		candidateRunner := NewAliasFactMutationRunner(runner, fixturePath)
		o0, err := candidateRunner.Run(ctx, cSource, "-O0", []string{input})
		if err != nil {
			return reduce.Signature{}, false, nil
		}
		o3, err := candidateRunner.Run(ctx, cSource, "-O3", []string{input})
		if err != nil {
			return reduce.Signature{}, false, nil
		}
		if len(o0.Pairs) != 1 || len(o3.Pairs) != 1 {
			return reduce.Signature{}, false, nil
		}
		engines := map[string]execution.Execution{"O0": o0.Pairs[0].Execution, "O3": o3.Pairs[0].Execution}
		compareErr := Phase5CompareEngines("qlt05-reverify-candidate", engines)
		if compareErr == nil {
			return reduce.Signature{}, false, nil
		}
		disagreement, ok := compareErr.(*Phase5EngineDisagreement)
		if !ok {
			return reduce.Signature{}, false, nil
		}
		foreignSequence, err := foreignCallSequenceFor(candidateProgram, o0.Pairs[0].Execution)
		if err != nil {
			return reduce.Signature{}, false, err
		}
		return SignatureFromDisagreement(disagreement, foreignSequence), true, nil
	}
}
