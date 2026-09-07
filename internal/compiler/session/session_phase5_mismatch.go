// session_phase5_mismatch.go wires plan 05-10's dependency-free
// internal/compiler/reduce package into a real, seeded interpreter
// mismatch, producing a reduce.MismatchDocument (lang.mismatch/0, D-05-26).
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
	"time"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/reduce"
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
)

// mismatchReduceFixture is the seeded, reproducible mismatch source this
// wiring reduces: plan 05-07's AliasFactMutationRunner over
// false_restrict_hoist.lang, engineered specifically to produce a real
// interpreter-optimizer-observable divergence (D-05-05) -- a ready source
// of a genuine mismatch, per this plan's own <read_first> note.
const mismatchReduceFixture = "testdata/phase5/false_restrict_hoist.lang"

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

// foreignCallSequenceFor extracts the ordered foreign symbol invocations a
// program's own function makes, for Signature.ForeignCallSequence --
// nil for a program with no ForeignContract (e.g. this plan's own seed
// fixture, a plain borrow chain), matching Signature's own documented
// zero-value semantics.
func foreignCallSequenceFor(program core.Program) []string {
	if len(program.Functions) != 1 {
		return nil
	}
	function := program.Functions[0]
	if function.ForeignContract == nil || function.Linear == nil {
		return nil
	}
	var sequence []string
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpForeignCall {
			sequence = append(sequence, function.ForeignContract.Symbol)
		}
	}
	return sequence
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
		candidateSignature := SignatureFromDisagreement(disagreement, foreignCallSequenceFor(candidateProgram))
		return candidateSignature, reduce.Interesting(seed, candidateSignature), nil
	}
}

// ReduceSeededAliasMismatch drives plan 05-07's AliasFactMutationRunner
// over false_restrict_hoist.lang to a REAL -O0-vs--O3 divergence, then
// reduces it via reduce.Reduce, emitting a lang.mismatch/0 document that is
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
	cSource, err := cgen.EmitNative(program)
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
	compareErr := Phase5CompareEngines("false_restrict_hoist.lang", engines)
	disagreement, ok := compareErr.(*Phase5EngineDisagreement)
	if compareErr == nil || !ok {
		return reduce.MismatchDocument{}, fmt.Errorf("mismatch-reduce: %s produced no real divergence to reduce -- the seed mutation was not exercised", ControlAliasFalseNoAlias)
	}

	seedSignature := SignatureFromDisagreement(disagreement, foreignCallSequenceFor(program))
	predicate := mismatchPredicate(baseRunner, fixturePath, nativeInput, seedSignature)
	result, err := reduce.Reduce(ctx, program, predicate)
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
			RecomputedWork: 1, ElapsedNS: time.Since(started).Nanoseconds(), PeakRSSStatus: "unavailable",
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
		RecomputedWork: work, ElapsedNS: time.Since(started).Nanoseconds(), PeakRSSStatus: "unavailable",
	})
	result.Metrics.RecomputedWork += work
	result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
	return result.Finalize(), nil
}
