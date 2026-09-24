package session_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/pathoracle"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// peerDivergenceEntry is plan 10-08 Task 3's own D-10-54 upgrade: an
// ACCOUNTABLE register value, replacing a bare code string. An allowlist
// entry that names no debt row is how a gate dies silently -- every
// declared divergence must be traceable to an open decision, not a bare
// exemption. code is the peer's reported Problems[0].Code (T-07-10-05's
// original field); debtID names the PHASE-NN-DEBT.md row that owns this
// divergence's disposition; landingPhase mirrors that row's own Landing
// phase cell for a human reading this file without also opening the
// register. TestPeerDivergenceRegisterDebtIDsAreAccountable is the
// companion test asserting every debtID here resolves to a STILL-OPEN row.
type peerDivergenceEntry struct {
	code         string
	debtID       string
	landingPhase string
}

// peerDivergenceExpected is 07-10's named, commented, hand-maintained
// register of every fixture under testdata/ where `check` ADMITS (zero
// checked.Diagnostics) but corevalidate.Validate REFUSES independently.
// TestNoUndeclaredCheckPeerDivergenceAcrossCorpus asserts this is the
// EXACT set, in both directions (T-07-10-05): an undeclared new divergence
// fails, and a stale entry that no longer diverges also fails. Keyed by
// project-relative path (forward slashes); value is D-10-54's accountable
// peerDivergenceEntry.
var peerDivergenceExpected = map[string]peerDivergenceEntry{
	// D-03-02's interprocedural half (PVG-04/CR-02) was carried here as
	// PHASE-07-DEBT.md D-07-49: check.computeLoanLastUses had no "call"
	// case, so the exclusive loan on `buffer` was treated as ending before
	// `relay`'s call was ever reached and `escort` checked clean, while
	// corevalidate's independent OpCall replay refused with
	// core.move_while_borrowed. CLOSED in Phase 08 (Task 2/3): check now
	// independently refuses this fixture too (check.interprocedural_loan_liveness),
	// so it is no longer a check-admits/peer-refuses divergence and this
	// entry is retired -- see
	// TestRelayEscortWitnessRefusesInterproceduralLiveness (check package)
	// and TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed
	// (corevalidate package), which now assert BOTH sides refuse
	// independently, via different codes, for the same program.
	//
	// WR-01's user-visible half (07-REVIEW.md; check-side half carried as
	// debt, PHASE-07-DEBT.md D-07-50): two `fn helper` declarations share
	// one semanticID; check's buildCalleeContracts silently resolves the
	// call to the LAST declaration, so check itself stays clean, but
	// corevalidate's function-ID uniqueness check refuses independently.
	"testdata/phase07/duplicate_function_name.lang": {
		code:         "core.duplicate_function_id",
		debtID:       "D-07-50",
		landingPhase: "Not scheduled -- needs its own ratified diagnostic code",
	},

	// 08-03's ACCEPTING twin members (D-03-02's interprocedural half,
	// producer side): check's own new interprocedural loan-liveness law
	// (Phase 08) correctly admits these -- the callee's declared return
	// contract is owned, so no loan is propagated across the call boundary.
	// corevalidate's own loan-liveness re-derivation is still
	// INTRAPROCEDURAL this phase (unmodified, Phase 09's own charter:
	// "peer re-derivation and D-03-02 closure"): loanChainIndex.carriedLoans
	// sets parent[operation.TargetID] = operation.SourceID for EVERY
	// operation kind, including OpCall, so it unconditionally treats the
	// call's own result as still carrying the argument's loan, regardless
	// of what the callee's declared contract says, and refuses. See
	// twin_a_accept.lang's own header for the full argument.
	//
	// RETIRED (Phase 09, D-09-03): corevalidate's own interprocedural
	// loan-carry peer (corevalidate_peer_liveness.go) now consults the
	// callee's declared return contract before propagating a loan across
	// an OpCall boundary, exactly like check's own interprocedural law
	// does -- so both "testdata/phase08/twin_a_accept.lang" and
	// "testdata/phase08/relay_depth2_accept.lang" no longer diverge here.
	// See PHASE-09-DEBT.md D-09-51 for a SEPARATE, pre-existing
	// originvalidate finding this retirement unmasked -- one this test is
	// structurally blind to (it never calls originvalidate.ValidatePublished).
}

// phase07SpotCheckRegression is 07-VERIFICATION.md's own pre-07-10
// Behavioral Spot-Checks table, restated here as a mechanical assertion
// (must_haves backstop truth): every fixture below is NOT a declared
// divergence entry, and its `lang check` status/code must be byte-for-byte
// unchanged by this plan.
var phase07SpotCheckRegression = []struct {
	fixture string
	status  string
	code    string // empty when status is pass (no diagnostics expected)
}{
	{"call_basic.lang", protocol.StatusPass, ""},
	{"call_from_both_match_arms.lang", protocol.StatusPass, ""},
	{"deep_diamond_acyclic.lang", protocol.StatusPass, ""},
	{"call_type_mismatch.lang", protocol.StatusInvalid, "check.call_argument_type_mismatch"},
	{"call_uncallable_callee.lang", protocol.StatusInvalid, "core.callee_not_callable"},
	{"clean_but_unpublishable.lang", protocol.StatusInvalid, "core.origin_omitted"},
	{"cycle_indirect.lang", protocol.StatusInvalid, "core.call_graph_cycle"},
	{"cycle_mutual.lang", protocol.StatusInvalid, "core.call_graph_cycle"},
	{"cycle_self.lang", protocol.StatusInvalid, "core.call_graph_cycle"},
	{"cycle_through_match_arm.lang", protocol.StatusInvalid, "core.call_graph_cycle"},
	{"cycle_unreachable.lang", protocol.StatusInvalid, "core.call_graph_cycle"},
	{"foreign_symbol_shadowing.lang", protocol.StatusInvalid, "core.call_graph_cycle"},
}

// TestCheckCommandSurfacesPeerRefusal is Task 1's core proof: CheckCommandFile
// now takes the refusing union of `check`'s own diagnostics and
// corevalidate.Validate's independent verdict, in fixed precedence order
// (07-10 checkpoint, SEM-04), closing 07-REVIEW.md CR-04 / PVG-03.
func TestCheckCommandSurfacesPeerRefusal(t *testing.T) {
	// Test 1: relay_escort_witness.lang reported the PEER's own
	// core.move_while_borrowed refusal at the CLI from 07-10 through Phase
	// 07 (before 07-10 it reported status: pass, exit-0-equivalent
	// StatusPass). Phase 08 closes the interprocedural half of D-03-02:
	// check ITSELF now refuses this fixture, via its own new
	// check.interprocedural_loan_liveness law
	// (TestRelayEscortWitnessRefusesInterproceduralLiveness, check
	// package), which runs and reports BEFORE CheckCommandFile ever
	// consults the peer -- the peer's own core.move_while_borrowed refusal
	// (TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed,
	// corevalidate package) still independently exists, it is simply no
	// longer the code surfaced at the CLI for this fixture.
	t.Run("relay_escort_witness flips to the peer's refusal", func(t *testing.T) {
		result, err := session.CheckCommandFile(phase07Fixture(t, "relay_escort_witness.lang"))
		if err != nil {
			t.Fatalf("CheckCommandFile returned an error: %v", err)
		}
		if result.Status != protocol.StatusInvalid {
			t.Fatalf("status = %s, want %s", result.Status, protocol.StatusInvalid)
		}
		if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "check.interprocedural_loan_liveness" {
			t.Fatalf("diagnostics = %+v, want exactly one check.interprocedural_loan_liveness", result.Diagnostics)
		}
	})

	// Test 2: duplicate_function_name.lang (WR-01's user-visible half)
	// refuses with the peer's own core.duplicate_function_id.
	t.Run("duplicate_function_name refuses with the peer's code", func(t *testing.T) {
		result, err := session.CheckCommandFile(phase07Fixture(t, "duplicate_function_name.lang"))
		if err != nil {
			t.Fatalf("CheckCommandFile returned an error: %v", err)
		}
		if result.Status != protocol.StatusInvalid {
			t.Fatalf("status = %s, want %s", result.Status, protocol.StatusInvalid)
		}
		if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "core.duplicate_function_id" {
			t.Fatalf("diagnostics = %+v, want exactly one core.duplicate_function_id", result.Diagnostics)
		}
	})

	// Test 3 (no regression): every fixture NOT in the declared divergence
	// register still reports exactly the status/code
	// 07-VERIFICATION.md's Behavioral Spot-Checks table recorded pre-07-10.
	t.Run("no regression across the rest of the phase07 corpus", func(t *testing.T) {
		for _, spot := range phase07SpotCheckRegression {
			spot := spot
			t.Run(spot.fixture, func(t *testing.T) {
				result, err := session.CheckCommandFile(phase07Fixture(t, spot.fixture))
				if err != nil {
					t.Fatalf("CheckCommandFile returned an error: %v", err)
				}
				if result.Status != spot.status {
					t.Fatalf("status = %s, want %s", result.Status, spot.status)
				}
				if spot.code == "" {
					if len(result.Diagnostics) != 0 {
						t.Fatalf("expected zero diagnostics, got %+v", result.Diagnostics)
					}
					return
				}
				// Edge 1 (adjacency): whether the shared code came from
				// check's own diagnostics or the peer, the union never
				// duplicates it -- exactly one diagnostic is reported.
				if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != spot.code {
					t.Fatalf("diagnostics = %+v, want exactly one %s", result.Diagnostics, spot.code)
				}
			})
		}
	})

	// Test 4 (edge 3, ordering): call_type_mismatch.lang has BOTH a check
	// diagnostic (check.call_argument_type_mismatch) and would separately
	// be refused by the peer's own OpCall replay -- the reported code is
	// check's, because check's diagnostics are reported before the peer
	// ever runs (precedence is asserted, not assumed).
	t.Run("edge 3: check diagnostics take precedence over the peer", func(t *testing.T) {
		result, err := session.CheckCommandFile(phase07Fixture(t, "call_type_mismatch.lang"))
		if err != nil {
			t.Fatalf("CheckCommandFile returned an error: %v", err)
		}
		if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "check.call_argument_type_mismatch" {
			t.Fatalf("diagnostics = %+v, want exactly one check.call_argument_type_mismatch (check's own code, not a core.* peer code)", result.Diagnostics)
		}
	})

	// Test 5 (edge 2, empty/fail-closed): !Valid with an EMPTY Problems
	// slice is structurally impossible from any real program, so it is
	// driven through a hand-built corevalidate.Result. The fallback must
	// never report an empty code and must never behave as StatusPass.
	t.Run("edge 2: empty Problems fails closed to the named fallback code", func(t *testing.T) {
		diag := session.PeerRefusalDiagnosticForTest(corevalidate.Result{Valid: false}, "phase07.example")
		if diag.Code != session.PeerRefusalUnnamedCodeForTest {
			t.Fatalf("code = %q, want the fail-closed fallback %q", diag.Code, session.PeerRefusalUnnamedCodeForTest)
		}
		if diag.Code == "" {
			t.Fatal("fail-closed fallback code must never be empty")
		}
	})
}

// TestDuplicateFunctionDeclarationRefusedAtCLI is the standing negative
// control for WR-01's user-visible half (D-07-50): before 07-10, this
// program reported status: pass, exit 0. It is now refused, incidentally,
// once CheckCommandFile consults the peer -- named as its own test per
// 07-10's <artifacts_this_phase_produces>, distinct from the table-driven
// assertion above so a future regression on this specific fixture names
// itself unambiguously in test output.
func TestDuplicateFunctionDeclarationRefusedAtCLI(t *testing.T) {
	result, err := session.CheckCommandFile(phase07Fixture(t, "duplicate_function_name.lang"))
	if err != nil {
		t.Fatalf("CheckCommandFile returned an error: %v", err)
	}
	if result.Status != protocol.StatusInvalid {
		t.Fatalf("status = %s, want %s (WR-01's user-visible half must be refused, not silently admitted)", result.Status, protocol.StatusInvalid)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "core.duplicate_function_id" {
		t.Fatalf("diagnostics = %+v, want exactly one core.duplicate_function_id", result.Diagnostics)
	}
}

// peerGateLinearProbeInput mirrors core_test.go's linearProbeInput and
// session_phase7.go's phase07LinearProbeInput exactly in spirit -- a third,
// independent copy of the SAME probe-input convention already established
// twice in this tree. This is NOT a re-derivation of any fact under test:
// it only synthesizes a runnable literal argument for interp.Run from a
// straight-line function's own declared parameter type, so plan 10-08's
// accept-side interp dispatch can reuse the identical convention rather
// than inventing a fourth one. A function whose parameter type has no
// known literal probe form is simply not dispatched through interp here,
// mirroring the existing corpus-wide dispatch lanes' own skip-not-fail
// convention.
func peerGateLinearProbeInput(function core.Function) (string, bool) {
	switch function.Parameter.Type {
	case "Byte":
		return "7", true
	case "Buffer":
		return "01020304", true
	default:
		return "", false
	}
}

// assertThreeWayEndpointAgreement is plan 10-08 Task 2's core proof
// (D-10-52): for every function in a checked core.Program, `check`'s own
// materialized core.LoanEndpoint set (function.Linear.LoanEndpoints),
// corevalidate's independently-recomputed set (Result.LoanEndpoints,
// D-10-52), and pathoracle's independently-recomputed set
// (RecomputeEndpoints, composition-aware since plan 10-03) must agree
// EXACTLY. This runs for EVERY corpus fixture with a structurally real
// core.Program -- check's own admission diagnostics never gate it, because
// pathoracle's own composition-depth definition is check/corevalidate-
// verdict-agnostic by design (session_composition_depth_test.go's own
// precedent): a fixture check itself refuses, like
// testdata/phase10/relay_depth3_refuse.lang, still has a fully-formed
// core.Program whose declared core.OpCall chain genuinely reaches the
// declared composition depth, and all three static peers must still agree
// on its structure. This is the "three-way" half of "three-way on refuse,
// four-way on accept" -- it never varies by admit/refuse verdict, only the
// fourth (interp) participant does; see
// assertInterpFrameModelAgreesWithStaticVerdict.
//
// SCOPING FINDING (empirically discovered while building this gate, stated
// rather than papered over): corevalidate.Validate is a FAIL-FAST,
// first-problem-wins replayer (v.check short-circuits v.run() the instant
// any problem is recorded), unlike check's own materializeLoanEndpoints or
// pathoracle.RecomputeEndpoints, neither of which ever short-circuits on an
// unrelated problem elsewhere in the program. A branched (CFG-carrying)
// function that corevalidate refuses for a reason UNRELATED to its own
// loan-endpoint recomputation -- one caught during an earlier structural or
// per-operation replay step, before v.loanEndpointsMatch is ever reached
// for that function -- legitimately has an EMPTY entry in
// Result.LoanEndpoints, diverging from check's and pathoracle's
// structurally-complete sets, for a reason that has nothing to do with any
// of the three peers' endpoint-derivation logic disagreeing. This is a
// property of corevalidate's own fail-fast architecture, not an endpoint
// bug, so the comparison is skipped for a function with CFG Blocks whose
// program corevalidate itself did not fully validate (!validated.Valid);
// see testdata/phase3/branch_one_arm_shared_reject.lang for the fixture
// that surfaced this while building the gate. It is NEVER skipped for a
// function with no Blocks at all (every peer trivially reports empty
// there, corevalidate included, regardless of fail-fast completion) or for
// any function once corevalidate.Validate is fully Valid.
func assertThreeWayEndpointAgreement(t *testing.T, fixture string, program core.Program, validated corevalidate.Result) {
	t.Helper()
	if len(program.Functions) == 0 {
		return // e.g. a fixture whose own check.Work exceeded MaxCheckWork before a real core.Program was ever built.
	}
	calleeLookup := pathoracle.BuildCalleeLookup(program)
	corevalidateEndpoints := validated.LoanEndpoints()
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue // a match-only function (no Linear body) has no core.LoanEndpoint concept for any of the three peers.
		}
		if !validated.Valid && len(function.Linear.Blocks) > 0 {
			// See the scoping finding above: corevalidate's fail-fast
			// replay may never have reached this function's own
			// loanEndpointsMatch call, so its endpoint entry is not
			// meaningfully comparable here.
			continue
		}
		checkSet := function.Linear.LoanEndpoints
		coreSet := corevalidateEndpoints[function.ID]
		pathSet, _, oracleErr := pathoracle.RecomputeEndpoints(function, calleeLookup)
		if oracleErr != nil {
			t.Errorf("%s: pathoracle.RecomputeEndpoints(%s): %v", fixture, function.ID, oracleErr)
			continue
		}
		if reflect.DeepEqual(checkSet, coreSet) && reflect.DeepEqual(checkSet, pathSet) {
			continue
		}
		t.Errorf(
			"%s: function %s -- three-way-on-refuse/four-way-on-accept endpoint differential found an UNDECLARED divergence: check=%+v corevalidate=%+v pathoracle=%+v",
			fixture, function.ID, checkSet, coreSet, pathSet,
		)
	}
}

// assertInterpFrameModelAgreesWithStaticVerdict is plan 10-08 Task 2's
// fourth, ACCEPT-ONLY participant (D-10-53). `interp` cannot produce a
// core.LoanEndpoint set -- its answer is an ordered Execution -- so this is
// deliberately NOT an equality comparison against the three static peers'
// sets, and it never runs on the refuse side (a refused program's declared
// contract is exactly what interp's own Run re-validates via
// corevalidate.Validate before executing anything, so running it there
// would prove nothing new). The comparison is METAMORPHIC: the static
// verdict already said ADMIT (both check and corevalidate agree), so
// interp actually executing the SAME checked core.Program cleanly, for
// every function and every reachable input, is this phase's riskiest
// assumption -- that interp's frame model agrees with the two admission
// layers about what is legal across a call boundary -- made falsifiable.
// interp's ONE genuine contribution here is named as its own separate
// claim, frame-model/static-verdict agreement, never folded into a peer
// count: this function proves that claim, and nothing here ever asks
// interp for a LoanEndpoint.
func assertInterpFrameModelAgreesWithStaticVerdict(t *testing.T, fixture string, program core.Program) {
	t.Helper()
	for _, function := range program.Functions {
		switch {
		case function.Match != nil:
			if function.Match.ScrutineeID != "" && function.Match.ScrutineeID != function.Parameter.ID {
				// A computed match is selected from a prefix value, so its arm
				// labels are not necessarily valid function inputs. Exercise
				// every declared alternative of the actual parameter type.
				for _, dataType := range program.DataTypes {
					if dataType.Name != function.Parameter.Type {
						continue
					}
					for _, alternative := range dataType.Alternatives {
						if _, err := interp.Run(program, function.Name, alternative); err != nil {
							t.Errorf("%s: interp.Run(%s, %q) disagreed with the static ACCEPT verdict (D-10-53's frame-model/static-verdict agreement claim): %v", fixture, function.Name, alternative, err)
						}
					}
				}
				continue
			}
			for _, arm := range function.Match.Arms {
				if _, err := interp.Run(program, function.Name, arm.Pattern); err != nil {
					t.Errorf("%s: interp.Run(%s, %q) disagreed with the static ACCEPT verdict (D-10-53's frame-model/static-verdict agreement claim): %v", fixture, function.Name, arm.Pattern, err)
				}
			}
		case function.Linear != nil:
			input, ok := peerGateLinearProbeInput(function)
			if !ok {
				continue
			}
			if _, err := interp.Run(program, function.Name, input); err != nil {
				t.Errorf("%s: interp.Run(%s, %q) disagreed with the static ACCEPT verdict (D-10-53's frame-model/static-verdict agreement claim): %v", fixture, function.Name, input, err)
			}
		}
	}
}

// TestNoUndeclaredCheckPeerDivergenceAcrossCorpus is Task 1 Test 7
// (T-07-10-05)'s original verdict-divergence walk, extended by plan 10-08
// Task 2 (D-10-51: EXTENDED, never a second harness -- D-09-23 already
// ratified this tradeoff) into the phase's own criterion-4 gate. It walks
// every .lang file under testdata/ and, for every fixture, runs THREE
// independent comparisons:
//
//  1. The ORIGINAL verdict-divergence walk (T-07-10-05, unchanged): when
//     check admits but corevalidate.Validate independently refuses, the
//     set of such fixtures must equal peerDivergenceExpected EXACTLY, in
//     both directions -- an undeclared new divergence fails, and a stale
//     entry that no longer diverges also fails.
//  2. THREE-WAY, on every fixture regardless of admit/refuse: `check`,
//     `corevalidate`, and `pathoracle`'s independently-derived
//     core.LoanEndpoint sets must agree exactly (D-10-52,
//     assertThreeWayEndpointAgreement) -- rebuilding M001 Phase 3's
//     exhaustive endpoint enumeration (03-VALIDATION.md's own five-row
//     conflict matrix) at the declared, bounded composition depth of 3
//     (QLT-04; testdata/phase10/relay_depth3_accept.lang and
//     relay_depth3_refuse.lang, plan 10-07, are both walked here and
//     agree).
//  3. FOUR-WAY, on the ACCEPT side only (both check and corevalidate
//     admit): `interp`'s ordered Execution additionally participates,
//     compared METAMORPHICALLY rather than by equality
//     (assertInterpFrameModelAgreesWithStaticVerdict, D-10-53).
//
// This gate is billed EVERYWHERE as exactly what it proves: THREE-WAY ON
// REFUSE, FOUR-WAY ON ACCEPT -- never unqualified "four-way" (D-09-37's
// recorded overclaim precedent). `interp`'s one genuine contribution,
// frame-model/static-verdict agreement, is its own separate claim, never a
// fourth equality peer, because it cannot produce a core.LoanEndpoint at
// all.
//
// Criterion 4's ownership-fact (call-site parameter mode) half is
// documented here, honestly, as NEAR-VACUOUS by construction (D-10-40):
// core.ParameterContract.Mode is hardcoded "owned" everywhere
// (corevalidate.go:2163, whose own comment cites D-07-01's cardinality-1
// grammar), and core.LinearOperation carries no override field, so a
// natural-input differential on this fact CANNOT DISAGREE BY CONSTRUCTION.
// What makes it non-vacuous is Task 3's seeded-fault harness (a synthetic
// core.Program with Mode flipped), NOT this natural-corpus walk agreeing --
// this walk proves the endpoint/execution facts above, never "three
// engines agree" on the ownership fact.
//
// D-10-54's accountable divergence register (peerDivergenceEntry, above)
// and its companion test (TestPeerDivergenceRegisterDebtIDsAreAccountable,
// below) are this file's own contribution to Task 3. The two D-10-55/
// D-10-41 seeded-fault-and-companion-assertion pairs Task 3 also lands
// CANNOT live in this file or this package: both fault seams
// (pathoracle.SetForceContractHopForTest, corevalidate's own
// parameterContractModeOverrideForTest) are _test.go-only symbols, and Go
// never compiles another package's _test.go files into ITS OWN test
// binary -- reachable only from within their OWN package's test binary,
// which is the exact "package boundary is the actual independence proof"
// property forceContractHopForTest's own doc comment already states. They
// are landed instead where they are reachable:
//
//	| Mutant                                                                        | Catches                                                                              | Blind, and why                                                                                             | Lands in                                                                          |
//	|--------------------------------------------------------------------------------|----------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------|
//	| pathoracle: a stubbed per-function contract hop replacing real per-path composition (plan 10-03's forceContractHopForTest, D-10-55) | pathoracle only                                                                        | check/corevalidate never call into pathoracle at all -- the seam is unreachable from outside package pathoracle | pathoracle_test.go#TestCompositionDiscriminatesPerPathBorrow (plan 10-08 Task 3 extends its companion assertion to also cover Result.LoanEndpoints, Task 1's accessor) |
//	| interp: a Move treated as a Copy (plan 10-01, interp.go's moveAsCopyForTest)     | interp only                                                                            | check/corevalidate execute nothing; a Move's own STATIC type is unaffected by whether the RUNTIME treats it as a copy | interp_test.go#TestMoveAsCopyMutationKilled                                          |
//	| corevalidate: parameterContract.Mode hardcoded to the wrong value (this plan's parameterContractModeOverrideForTest, D-10-41's complementary half) | the criterion-4 differential (corevalidate's own peer signature vs. originvalidate's declared contract) | interp, which never reads a mode string at all -- it executes core.Operations, never contract metadata          | corevalidate_endpoint_test.go#TestDerivePeerSignatureModeMutantPairing               |
//
// Naming which peer alone catches which mutant, in both directions, is
// what makes "independent" falsifiable rather than rhetorical.
//
// `interp`'s OWN independence claim is also narrower than `check`'s and
// `corevalidate`'s mutual non-import independence (D-10-37): `interp`
// imports `corevalidate` and `Run`'s first act is
// `corevalidate.Validate(program)`, so a bug in `corevalidate.Validate`
// would feed `interp` bad input too. The defensible claim this gate relies
// on is independence of DERIVATION MECHANISM for the frame-model fact
// specifically, nested inside a shared, unrelated validation dependency --
// never the stronger mutual-non-import claim.
func TestNoUndeclaredCheckPeerDivergenceAcrossCorpus(t *testing.T) {
	root := testsupport.ProjectPath("testdata")
	found := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".lang" {
			return nil
		}
		checked, checkErr := session.CheckFile(path)
		if checkErr != nil {
			t.Fatalf("CheckFile(%s): %v", path, checkErr)
		}
		relative, relErr := filepath.Rel(testsupport.ProjectPath("."), path)
		if relErr != nil {
			t.Fatalf("filepath.Rel: %v", relErr)
		}
		relative = filepath.ToSlash(relative)

		checkAdmits := len(checked.Diagnostics) == 0
		validated := corevalidate.Validate(checked.Program)

		if checkAdmits && !validated.Valid {
			// check admits but the peer independently refuses -- the
			// ORIGINAL verdict-divergence fact (T-07-10-05), unchanged.
			code := session.PeerRefusalUnnamedCodeForTest
			if len(validated.Problems) > 0 {
				code = validated.Problems[0].Code
			}
			found[relative] = code
		}

		// D-10-52: three-way endpoint agreement, on every fixture,
		// regardless of admit/refuse.
		assertThreeWayEndpointAgreement(t, relative, checked.Program, validated)

		// D-10-53: interp participates on the ACCEPT side only -- both
		// admission layers must agree the program is admitted.
		if checkAdmits && validated.Valid {
			assertInterpFrameModelAgreesWithStaticVerdict(t, relative, checked.Program)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}

	for path, wantEntry := range peerDivergenceExpected {
		gotCode, ok := found[path]
		if !ok {
			t.Errorf("declared divergence %s no longer diverges (check admits, peer no longer refuses, or the fixture is missing) -- stale entry in peerDivergenceExpected", path)
			continue
		}
		if gotCode != wantEntry.code {
			t.Errorf("%s: peer code = %q, want declared %q", path, gotCode, wantEntry.code)
		}
	}
	for path, gotCode := range found {
		if _, declared := peerDivergenceExpected[path]; !declared {
			t.Errorf("UNDECLARED divergence: %s (check admits, corevalidate refuses with %s) is not in peerDivergenceExpected", path, gotCode)
		}
	}
}

// TestInterfaceCommandsReportPeerRefusalAsInvalid is Task 2's core proof: a
// program corevalidate.Validate refuses is reported by both
// InterfaceExportCommandFile and InterfaceCoreCommandFile as a
// protocol.StatusInvalid result carrying the peer's own code, with a NIL
// error -- never the pre-07-10 `fmt.Errorf("core validation failed: …")`
// that reached the CLI as tool.operation_failed / exit 3 with the code
// discarded (07-REVIEW.md CR-04's `interface` half).
func TestInterfaceCommandsReportPeerRefusalAsInvalid(t *testing.T) {
	fixture := phase07Fixture(t, "duplicate_function_name.lang")

	// Test 1: InterfaceCoreCommandFile.
	t.Run("InterfaceCoreCommandFile reports StatusInvalid with a nil error", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "duplicate_function_name.core.json")
		result, err := session.InterfaceCoreCommandFile(fixture, outPath)
		if err != nil {
			t.Fatalf("InterfaceCoreCommandFile returned a non-nil error: %v", err)
		}
		if result.Status != protocol.StatusInvalid {
			t.Fatalf("status = %s, want %s", result.Status, protocol.StatusInvalid)
		}
		if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "core.duplicate_function_id" {
			t.Fatalf("diagnostics = %+v, want exactly one core.duplicate_function_id", result.Diagnostics)
		}
		// Test 3: no output artifact is written on a refusal.
		if _, statErr := os.Stat(outPath); !os.IsNotExist(statErr) {
			t.Fatalf("expected %s to not exist after a refusal, stat err = %v", outPath, statErr)
		}
	})

	// Test 2: InterfaceExportCommandFile, the same shape.
	t.Run("InterfaceExportCommandFile reports StatusInvalid with a nil error", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "duplicate_function_name.summary.json")
		result, err := session.InterfaceExportCommandFile(fixture, outPath)
		if err != nil {
			t.Fatalf("InterfaceExportCommandFile returned a non-nil error: %v", err)
		}
		if result.Status != protocol.StatusInvalid {
			t.Fatalf("status = %s, want %s", result.Status, protocol.StatusInvalid)
		}
		if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "core.duplicate_function_id" {
			t.Fatalf("diagnostics = %+v, want exactly one core.duplicate_function_id", result.Diagnostics)
		}
		if _, statErr := os.Stat(outPath); !os.IsNotExist(statErr) {
			t.Fatalf("expected %s to not exist after a refusal, stat err = %v", outPath, statErr)
		}
	})

	// Test 4: an accepting program behaves exactly as before -- same
	// StatusPass, artifact written, and a working round-trip through
	// InterfaceCheckCommandFile (the existing CoreDigest-bound contract).
	t.Run("accepting program: unchanged StatusPass, written artifacts, and a working round-trip", func(t *testing.T) {
		accepting := phase07Fixture(t, "call_basic.lang")
		dir := t.TempDir()
		summaryPath := filepath.Join(dir, "call_basic.summary.json")
		corePath := filepath.Join(dir, "call_basic.core.json")

		exportResult, err := session.InterfaceExportCommandFile(accepting, summaryPath)
		if err != nil {
			t.Fatalf("InterfaceExportCommandFile: %v", err)
		}
		if exportResult.Status != protocol.StatusPass {
			t.Fatalf("export status = %s, want %s", exportResult.Status, protocol.StatusPass)
		}
		if exportResult.Interface == nil {
			t.Fatal("expected a non-nil Interface projection on an accepting program")
		}

		coreResult, err := session.InterfaceCoreCommandFile(accepting, corePath)
		if err != nil {
			t.Fatalf("InterfaceCoreCommandFile: %v", err)
		}
		if coreResult.Status != protocol.StatusPass {
			t.Fatalf("core status = %s, want %s", coreResult.Status, protocol.StatusPass)
		}

		checkResult, err := session.InterfaceCheckCommandFile(summaryPath, corePath)
		if err != nil {
			t.Fatalf("InterfaceCheckCommandFile: %v", err)
		}
		if checkResult.Status != protocol.StatusPass {
			t.Fatalf("round-trip check status = %s, want %s (the digest bound at export time must still match the bytes InterfaceCoreCommandFile wrote)", checkResult.Status, protocol.StatusPass)
		}
	})

	// Test 5 (edge 2, fail-closed): !Valid with an empty Problems slice
	// yields StatusInvalid with the fallback code, on both paths -- driven
	// through the shared peerRefusalDiagnostic helper, since no real
	// program produces an empty Problems slice.
	t.Run("edge 2: empty Problems fails closed on both interface paths too", func(t *testing.T) {
		diag := session.PeerRefusalDiagnosticForTest(corevalidate.Result{Valid: false}, "phase07.example")
		if diag.Code != session.PeerRefusalUnnamedCodeForTest {
			t.Fatalf("code = %q, want the fail-closed fallback %q", diag.Code, session.PeerRefusalUnnamedCodeForTest)
		}
	})
}

// TestCheckCommandPeerConsultMutationKilled is Task 3's own observation
// that control:check.peer_consulted fails under its seeded fault (QLT-08):
// a control never seen to fail is not evidence. With checkCommandPeerSeam
// enabled, CheckCommandFile skips the peer entirely and WRONGLY reports
// StatusPass on relay_escort_witness.lang; with the seam restored, it
// correctly refuses with core.move_while_borrowed. Both directions are
// asserted in one test, so a control that is green for the wrong reason
// (e.g. a no-op seam) fails here.
func TestCheckCommandPeerConsultMutationKilled(t *testing.T) {
	// relay_escort_witness.lang was this test's fixture through Phase 07:
	// check itself admitted it, so a peer-consult seam that wrongly skips
	// the peer entirely made CheckCommandFile wrongly report StatusPass,
	// proving the peer consult load-bearing. Phase 08 closes the
	// interprocedural half of D-03-02 (Task 2/3): check now independently
	// refuses this exact fixture via its own new
	// check.interprocedural_loan_liveness law, BEFORE CheckCommandFile ever
	// reaches the peer-consult seam -- so this fixture can no longer
	// demonstrate the seam is load-bearing (it would report StatusInvalid
	// regardless of the seam). duplicate_function_name.lang is the
	// remaining declared divergence (peerDivergenceExpected): check admits
	// it cleanly and only the peer's independent function-ID uniqueness
	// check refuses it, which is exactly the shape this mutation-kill needs.
	fixture := phase07Fixture(t, "duplicate_function_name.lang")

	restore := session.SetCheckCommandPeerSeamForTest(true)
	wronglyPermissive, err := session.CheckCommandFile(fixture)
	restore()
	if err != nil {
		t.Fatalf("CheckCommandFile (seam on) returned an error: %v", err)
	}
	if wronglyPermissive.Status != protocol.StatusPass {
		t.Fatalf("with the seam ON, expected the mutation to wrongly report %s, got %s -- the seam is not load-bearing", protocol.StatusPass, wronglyPermissive.Status)
	}

	correctlyRefusing, err := session.CheckCommandFile(fixture)
	if err != nil {
		t.Fatalf("CheckCommandFile (seam restored) returned an error: %v", err)
	}
	if correctlyRefusing.Status != protocol.StatusInvalid || len(correctlyRefusing.Diagnostics) != 1 || correctlyRefusing.Diagnostics[0].Code != "core.duplicate_function_id" {
		t.Fatalf("with the seam restored, expected exactly one core.duplicate_function_id diagnostic, got status=%s diagnostics=%+v", correctlyRefusing.Status, correctlyRefusing.Diagnostics)
	}
}

// TestInterfacePeerRefusalMutationKilled is Task 3's own observation that
// control:interface.peer_refusal_is_invalid fails under its seeded fault:
// with interfacePeerRefusalSeam enabled, both interface command paths
// restore the pre-07-10 behaviour of a non-nil error with no StatusInvalid
// result (the code is discarded); with it restored, both report
// StatusInvalid carrying the peer's own code. One seam kills the control
// on both paths, asserted here on both.
func TestInterfacePeerRefusalMutationKilled(t *testing.T) {
	fixture := phase07Fixture(t, "duplicate_function_name.lang")

	t.Run("InterfaceCoreCommandFile", func(t *testing.T) {
		restore := session.SetInterfacePeerRefusalSeamForTest(true)
		wrongResult, wrongErr := session.InterfaceCoreCommandFile(fixture, filepath.Join(t.TempDir(), "out.core.json"))
		restore()
		if wrongErr == nil {
			t.Fatalf("with the seam ON, expected a non-nil error (the mutation restores discarding the peer's code), got result=%+v", wrongResult)
		}
		if wrongResult.Status == protocol.StatusInvalid {
			t.Fatal("with the seam ON, a non-nil error must not also carry a StatusInvalid result -- the seam is not load-bearing")
		}

		correctResult, correctErr := session.InterfaceCoreCommandFile(fixture, filepath.Join(t.TempDir(), "out.core.json"))
		if correctErr != nil {
			t.Fatalf("with the seam restored, expected a nil error, got %v", correctErr)
		}
		if correctResult.Status != protocol.StatusInvalid || len(correctResult.Diagnostics) != 1 || correctResult.Diagnostics[0].Code != "core.duplicate_function_id" {
			t.Fatalf("with the seam restored, expected exactly one core.duplicate_function_id diagnostic, got status=%s diagnostics=%+v", correctResult.Status, correctResult.Diagnostics)
		}
	})

	t.Run("InterfaceExportCommandFile", func(t *testing.T) {
		restore := session.SetInterfacePeerRefusalSeamForTest(true)
		wrongResult, wrongErr := session.InterfaceExportCommandFile(fixture, filepath.Join(t.TempDir(), "out.summary.json"))
		restore()
		if wrongErr == nil {
			t.Fatalf("with the seam ON, expected a non-nil error (the mutation restores discarding the peer's code), got result=%+v", wrongResult)
		}
		if wrongResult.Status == protocol.StatusInvalid {
			t.Fatal("with the seam ON, a non-nil error must not also carry a StatusInvalid result -- the seam is not load-bearing")
		}

		correctResult, correctErr := session.InterfaceExportCommandFile(fixture, filepath.Join(t.TempDir(), "out.summary.json"))
		if correctErr != nil {
			t.Fatalf("with the seam restored, expected a nil error, got %v", correctErr)
		}
		if correctResult.Status != protocol.StatusInvalid || len(correctResult.Diagnostics) != 1 || correctResult.Diagnostics[0].Code != "core.duplicate_function_id" {
			t.Fatalf("with the seam restored, expected exactly one core.duplicate_function_id diagnostic, got status=%s diagnostics=%+v", correctResult.Status, correctResult.Diagnostics)
		}
	})
}

// phase07Fixture resolves a testdata/phase07 fixture's absolute path,
// mirroring this package's other phase07 test helpers' precedent.
func phase07Fixture(t testing.TB, name string) string {
	t.Helper()
	if strings.Contains(name, string(filepath.Separator)) || strings.Contains(name, "/") {
		t.Fatalf("phase07Fixture expects a bare filename, got %q", name)
	}
	return testsupport.ProjectPath("testdata", "phase07", name)
}

// peerDivergenceDebtIDResolvesToOpenRow resolves debtID against every
// discovered *-DEBT.md register's own `## Items` table (D-10-54), reusing
// session_test.go's existing phaseArtifactGlob/debtRegisterTable parsing --
// never a second parser, following D-10-51's own "extend, don't duplicate"
// discipline applied to register-reading too. Returns found=false if no
// register names debtID at all; open=false if the row's own Landing phase
// cell reads as already resolved/closed (a leading "Resolved" or a
// "Superseded, closed" marker, the two closure conventions already used
// across this tree's own registers, e.g. PHASE-09-DEBT.md D-09-21/D-09-31).
func peerDivergenceDebtIDResolvesToOpenRow(t *testing.T, debtID string) (open bool, found bool) {
	t.Helper()
	registers, err := phaseArtifactGlob("*", "*-DEBT.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range registers {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		columns, rows := debtRegisterTable(t, filepath.Base(path), text)
		idColumn, hasID := columns["ID"]
		if !hasID {
			continue
		}
		landingColumn, hasLanding := columns["Landing phase"]
		for _, row := range rows {
			if row[idColumn] != debtID {
				continue
			}
			if !hasLanding {
				// A legacy register exempted from the Landing phase column
				// (debtRegisterLandingPhaseExemptions in session_test.go)
				// carries no closure signal at all -- treat as open rather
				// than guessing.
				return true, true
			}
			landing := strings.ToLower(strings.TrimSpace(row[landingColumn]))
			closed := strings.HasPrefix(landing, "resolved") || strings.Contains(landing, "superseded, closed")
			return !closed, true
		}
	}
	return false, false
}

// TestPeerDivergenceRegisterDebtIDsAreAccountable is plan 10-08 Task 3's
// own D-10-54 companion: every debtID peerDivergenceExpected declares must
// resolve to a row in some PHASE-NN-DEBT.md register, and that row must
// still be OPEN. An allowlist entry naming a nonexistent or already-closed
// debt row is exactly how a gate dies silently -- this test is the
// mechanical backstop. Includes a negative control over a synthetic
// unresolvable ID, proving the resolver can genuinely go red rather than
// vacuously reporting "open" for anything it cannot find.
func TestPeerDivergenceRegisterDebtIDsAreAccountable(t *testing.T) {
	for path, entry := range peerDivergenceExpected {
		path, entry := path, entry
		t.Run(path, func(t *testing.T) {
			open, found := peerDivergenceDebtIDResolvesToOpenRow(t, entry.debtID)
			if !found {
				t.Fatalf("declared divergence %s names debt ID %s, which resolves to no row in any *-DEBT.md register", path, entry.debtID)
			}
			if !open {
				t.Fatalf("declared divergence %s names debt ID %s, which resolves to an already-closed row -- promote or retire this entry", path, entry.debtID)
			}
		})
	}

	t.Run("negative control: a synthetic unresolvable ID fails closed", func(t *testing.T) {
		open, found := peerDivergenceDebtIDResolvesToOpenRow(t, "D-99-99")
		if found {
			t.Fatal("synthetic debt ID D-99-99 unexpectedly resolved to a real row -- this negative control's own ID collided with a real one, pick a different synthetic ID")
		}
		if open {
			t.Fatal("an unresolved debt ID must never report open=true")
		}
	})
}
