// This file proves OWN-05's shared-fact half (D-09-36): call-site
// ownership transfer has ONE meaning because `check` and `corevalidate`
// independently derive the SAME per-call-site move-vs-borrow
// classification and the SAME resulting loan/ownership state after the
// call, each from the callee's declared core.FunctionSignature -- never
// from a shared implementation. `check`'s own admission-time consult
// (buildCallSignatureTable, sourced from originvalidate.BuildInterface --
// the PRODUCER's own read) and corevalidate's Result.PeerSignatures()
// (corevalidate's own, separately-implemented structural re-derivation,
// D-07-20/D-07-22) are exactly the two genuinely independent readers
// buildCallSignatureTable's own doc comment names.
//
// D-09-38: no shared "convention classification" helper is extracted here
// for Phase 10's `interp` peer to reuse. `check` and `corevalidate` already
// each read signature.Parameters[0].Mode independently, sharing nothing
// beyond the struct shape; Phase 10's `interp` peer re-derives the
// move/borrow decision from its own view of the callee signature. The one
// comparison helper this file needs (callSiteAgreement) lives HERE, in the
// test file, and is used only by this file's own tests.
//
// D-09-37: this test proves two of OWN-05's three derivers (`check` and
// `corevalidate`). The third is Phase 10's `interp` peer, verified there by
// TRU-03 and in Phase 11 by NAT-06. OWN-05 is NOT closed by this test
// alone; the requirement-document split (OWN-05a/OWN-05b) is plan 09-10's
// own change, not this plan's.
package check

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/ability"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
)

// callTransferAgreementFixture names one admitted testdata fixture this
// sweep drives, and the phase directory it lives under (mirroring
// readPhase07Fixture/readPhase08Fixture's own two-directory convention).
type callTransferAgreementFixture struct {
	phase string
	name  string
	// wantStraightLine/wantBlocks record which of corevalidate's two
	// summary-peer replay sites THIS fixture's own body shape must drive
	// (D-07-21): a program cannot mix a match-bodied function with a
	// straight-line one (core.mixed_body_versions), so every legal program
	// exercises exactly one of the two, never both.
	wantStraightLine bool
	wantBlocks       bool
}

// callTransferAgreementFixtures is the swept corpus: every already-shipped
// admitted fixture under testdata/phase07 and testdata/phase08 containing
// at least one Lang-to-Lang call. twin_a_accept.lang and
// relay_depth2_accept.lang are the exact fixtures whose corevalidate
// divergence Phase 09's earlier plans closed (session_peer_gate_test.go's
// peerDivergenceExpected no longer names either) -- this test's own
// baseline check below fails loudly if that ever regresses.
var callTransferAgreementFixtures = []callTransferAgreementFixture{
	{phase: "phase07", name: "call_basic.lang", wantStraightLine: true},
	{phase: "phase07", name: "call_argument_used_once.lang", wantStraightLine: true},
	{phase: "phase07", name: "call_from_both_match_arms.lang", wantBlocks: true},
	{phase: "phase08", name: "twin_a_accept.lang", wantStraightLine: true},
	{phase: "phase08", name: "relay_depth2_accept.lang", wantStraightLine: true},
	{phase: "phase08", name: "match_arm_call.lang", wantBlocks: true},
}

func readCallTransferAgreementFixture(t *testing.T, fixture callTransferAgreementFixture) []byte {
	t.Helper()
	switch fixture.phase {
	case "phase07":
		return readPhase07Fixture(t, fixture.name)
	case "phase08":
		return readPhase08Fixture(t, fixture.name)
	default:
		t.Fatalf("unknown fixture phase %q", fixture.phase)
		return nil
	}
}

// callSiteClassification is the one per-call-site fact both peers must
// independently agree on (D-09-36): the callee's declared Parameters[0]
// convention (the move-vs-borrow classification an argument at this call
// site receives) and whether the callee's declared Return.Mode means the
// call's own target place inherits a borrow of the caller's argument (the
// resulting loan/ownership state after the call).
type callSiteClassification struct {
	parameterMode        string
	returnsBorrowOfParam bool
}

func classifyFromSignature(signature core.FunctionSignature) callSiteClassification {
	classification := callSiteClassification{returnsBorrowOfParam: signature.Return.Mode == "shared" || signature.Return.Mode == "exclusive"}
	if len(signature.Parameters) > 0 {
		classification.parameterMode = signature.Parameters[0].Mode
	}
	return classification
}

// TestCallSiteTransferClassificationAgreesAcrossPeers sweeps every
// call-bearing fixture `check` admits and asserts that, for every call
// site, check's OWN admission-time signature consult
// (buildCallSignatureTable, sourced from originvalidate.BuildInterface)
// and corevalidate's OWN independently re-derived peer signature
// (Result.PeerSignatures()) agree on both halves of D-09-36's shared fact.
// The two are compared as DATA -- structural equality of each side's own
// derived classification -- proving the OUTPUTS agree without constraining
// either side's IMPLEMENTATION, in the spirit of this package's own
// structuralFieldsEqual/ClosureDigest byte-equality precedents.
func TestCallSiteTransferClassificationAgreesAcrossPeers(t *testing.T) {
	var sawStraightLine, sawBlocks bool
	var comparedCallSites int

	for _, fixture := range callTransferAgreementFixtures {
		t.Run(fixture.phase+"/"+fixture.name, func(t *testing.T) {
			source := readCallTransferAgreementFixture(t, fixture)
			program := mustParseProgram(t, source)
			result := Program(program)
			if len(result.Diagnostics) != 0 {
				t.Fatalf("expected check to admit %s with zero diagnostics, got %+v", fixture.name, result.Diagnostics)
			}

			table, err := buildCallSignatureTable(result.Program)
			if err != nil {
				t.Fatalf("buildCallSignatureTable: %v", err)
			}

			coreResult := corevalidate.Validate(result.Program)
			if !coreResult.Valid {
				t.Fatalf("expected corevalidate to independently admit %s, got problems %+v", fixture.name, coreResult.Problems)
			}
			peerSignatures := coreResult.PeerSignatures()

			straightLine, blocks := coreResult.PeerSiteCoverage()
			if straightLine != fixture.wantStraightLine || blocks != fixture.wantBlocks {
				t.Fatalf("%s: PeerSiteCoverage() = (straightLine=%v, blocks=%v), want (straightLine=%v, blocks=%v) -- coverage must be asserted per fixture shape, never inferred from which fixtures happened to be supplied", fixture.name, straightLine, blocks, fixture.wantStraightLine, fixture.wantBlocks)
			}
			if straightLine {
				sawStraightLine = true
			}
			if blocks {
				sawBlocks = true
			}

			for _, function := range result.Program.Functions {
				if function.Linear == nil {
					continue
				}
				for _, operation := range function.Linear.Operations {
					if operation.Kind != core.OpCall {
						continue
					}
					checkSignature, ok := table.lookup(operation.CalleeID)
					if !ok {
						t.Fatalf("%s: check's own callSignatureTable has no entry for callee %s", fixture.name, operation.CalleeID)
					}
					peerSignature, ok := peerSignatures[operation.CalleeID]
					if !ok {
						t.Fatalf("%s: corevalidate's PeerSignatures() has no entry for callee %s", fixture.name, operation.CalleeID)
					}

					checkClassification := classifyFromSignature(checkSignature)
					peerClassification := classifyFromSignature(peerSignature)
					if checkClassification != peerClassification {
						t.Fatalf("%s: call site to %s -- check derived %+v, corevalidate derived %+v; the two peers disagree on OWN-05's shared fact", fixture.name, operation.CalleeID, checkClassification, peerClassification)
					}
					comparedCallSites++
				}
			}
		})
	}

	if !sawStraightLine {
		t.Fatal("no swept fixture exercised corevalidate's straight-line replay site -- PeerSiteCoverage() coverage is not proven, only inferred")
	}
	if !sawBlocks {
		t.Fatal("no swept fixture exercised corevalidate's blocks replay site -- PeerSiteCoverage() coverage is not proven, only inferred")
	}
	if comparedCallSites == 0 {
		t.Fatal("no call site was compared -- the sweep's fixture list no longer contains a Lang-to-Lang call")
	}
}

// syntheticBorrowReturningLeaf and syntheticOwnedCaller build a minimal,
// fully structurally-valid two-function core.Program directly (mirroring
// corevalidate_cycle_peer_test.go's own syntheticProgram/syntheticFunction
// convention, adapted here since this file must never import syntax --
// only core and corevalidate, D-09-38's own "never through the parser"
// discipline for a peer's own test). `leaf` declares a genuinely
// SHARED-borrow PublicOrigin over its own parameter (Return.Mode ==
// "shared" -- differing from Parameters[0].Mode, which D-07-01 keeps
// hardcoded "owned" today), and `caller` calls it once and returns the
// result directly with no conflicting move. This is Task 3(c)'s negative
// direction: a fixture whose callee's Return.Mode and Parameters[0].Mode
// genuinely differ, so the classifications compared above are not an
// artifact of every value in the corpus happening to be "owned".
func syntheticBorrowReturningLeaf() core.Function {
	const fnID = "s1:calltransferagreement:fn:leaf"
	const typeID = fnID + ":type:0"
	const parameterID = fnID + ":place:0"
	const viewID = fnID + ":place:1"
	derived, _ := ability.Derive(core.TypeRef{Constructor: "Buffer"})
	return core.Function{
		ID: fnID, Name: "leaf",
		EntryPointID: fnID + ":point:entry", ReturnPointID: fnID + ":point:return",
		Parameter:    core.Parameter{ID: parameterID, Name: "buffer", Type: "Buffer"},
		ReturnType:   "Buffer",
		PublicOrigin: &core.PublicOrigin{Paths: []string{"buffer"}, Access: "shared"},
		Linear: &core.LinearBody{
			ID: fnID + ":linear",
			Types: []core.TypeFact{{
				ID: typeID, Shape: core.TypeRef{Constructor: "Buffer", Arguments: []core.TypeRef{}},
				Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses,
			}},
			Places: []core.Place{
				{ID: parameterID, Name: "buffer", TypeID: typeID},
				{ID: viewID, Name: "view", TypeID: typeID},
			},
			Operations: []core.LinearOperation{
				{ID: fnID + ":op:0", PointID: fnID + ":point:linear:0", Kind: core.OpBorrowShared, SourceID: parameterID, TargetID: viewID, TypeID: typeID, LoanID: fnID + ":loan:0"},
				{ID: fnID + ":op:1", PointID: fnID + ":point:linear:1", Kind: core.OpReturn, SourceID: viewID, TypeID: typeID},
			},
		},
	}
}

func syntheticOwnedCaller(calleeID string) core.Function {
	const fnID = "s1:calltransferagreement:fn:caller"
	const typeID = fnID + ":type:0"
	const parameterID = fnID + ":place:0"
	const resultID = fnID + ":place:1"
	derived, _ := ability.Derive(core.TypeRef{Constructor: "Buffer"})
	return core.Function{
		ID: fnID, Name: "caller",
		EntryPointID: fnID + ":point:entry", ReturnPointID: fnID + ":point:return",
		Parameter:  core.Parameter{ID: parameterID, Name: "buffer", Type: "Buffer"},
		ReturnType: "Buffer",
		Linear: &core.LinearBody{
			ID: fnID + ":linear",
			Types: []core.TypeFact{{
				ID: typeID, Shape: core.TypeRef{Constructor: "Buffer", Arguments: []core.TypeRef{}},
				Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses,
			}},
			Places: []core.Place{
				{ID: parameterID, Name: "buffer", TypeID: typeID},
				{ID: resultID, Name: "result", TypeID: typeID},
			},
			Operations: []core.LinearOperation{
				{ID: fnID + ":op:0", PointID: fnID + ":point:linear:0", Kind: core.OpCall, SourceID: parameterID, TargetID: resultID, TypeID: typeID, CalleeID: calleeID},
				{ID: fnID + ":op:1", PointID: fnID + ":point:linear:1", Kind: core.OpReturn, SourceID: resultID, TypeID: typeID},
			},
		},
	}
}

// TestCallSiteTransferClassificationAgreesWhenModesDiffer is Task 3(c)'s
// negative-direction case, built directly as a synthetic core.Program
// (never through the parser) rather than as a .lang fixture: `leaf`'s
// declared Return.Mode ("shared") differs from its own Parameters[0].Mode
// ("owned", D-07-01's still-hardcoded fact), so a comparison that
// accidentally read the WRONG field on either side would diverge here even
// though every fixture in the main sweep above happens to have
// Return.Mode == "owned" throughout.
func TestCallSiteTransferClassificationAgreesWhenModesDiffer(t *testing.T) {
	leaf := syntheticBorrowReturningLeaf()
	caller := syntheticOwnedCaller(leaf.ID)
	program := core.Program{Schema: core.Schema1, Module: "calltransferagreement", ModuleID: "s1:calltransferagreement:module:calltransferagreement", Functions: []core.Function{leaf, caller}}

	table, err := buildCallSignatureTable(program)
	if err != nil {
		t.Fatalf("buildCallSignatureTable: %v", err)
	}
	checkSignature, ok := table.lookup(leaf.ID)
	if !ok {
		t.Fatal("check's own callSignatureTable has no entry for leaf")
	}
	if checkSignature.Parameters[0].Mode == checkSignature.Return.Mode {
		t.Fatalf("fixture is not actually a negative-direction case: Parameters[0].Mode == Return.Mode == %q", checkSignature.Parameters[0].Mode)
	}

	coreResult := corevalidate.Validate(program)
	if !coreResult.Valid {
		t.Fatalf("expected corevalidate to admit the synthetic program, got problems %+v", coreResult.Problems)
	}
	peerSignature, ok := coreResult.PeerSignatures()[leaf.ID]
	if !ok {
		t.Fatal("corevalidate's PeerSignatures() has no entry for leaf")
	}

	checkClassification := classifyFromSignature(checkSignature)
	peerClassification := classifyFromSignature(peerSignature)
	if checkClassification != peerClassification {
		t.Fatalf("check derived %+v, corevalidate derived %+v for a callee whose Return.Mode and Parameters[0].Mode differ -- the two peers disagree", checkClassification, peerClassification)
	}
	if !checkClassification.returnsBorrowOfParam {
		t.Fatal("expected both peers to derive returnsBorrowOfParam == true for a callee declaring a shared-borrow return")
	}
}
