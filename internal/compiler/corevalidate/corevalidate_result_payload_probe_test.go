package corevalidate_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
)

// Phase 12 criterion 3 -- the `Result` payload / interprocedural-origin
// pre-flight probe, run BEFORE planning per the ROADMAP's "Depends on" line.
//
// The question, from 12-CONTEXT.md D-12-02: does a payload flowing out of a
// callee as part of a returned value create an origin or ownership
// obligation that the four independent admission peers (check, corevalidate,
// originvalidate, pathoracle) cannot already express with the Phase 08-11
// machinery?
//
// BRANCH A and BRANCH B were pre-registered in
// .planning/phases/12-result-payloads/12-CONTEXT.md (D-12-04) and committed
// BEFORE this file was authored or run, so the verdict below cannot be
// rationalized after the fact:
//
//   BRANCH A (accepted, phase proceeds): every peer's verdict on the probe
//   shapes is fully attributable to an already-catalogued carry-forward gap
//   -- D-10-C01, D-11-51, or D-11-52 -- with no new divergence root cause.
//
//   BRANCH B (slip to M003): fires when EITHER (a) a peer's verdict traces
//   to a divergence root cause not already named in those three items, OR
//   (b) closing the divergence requires a new dispatch-site case arm of a
//   kind not already inventoried across the six dispatch sites. Filling an
//   existing empty switch arm is in-scope maintenance, NOT a new
//   interprocedural rule -- that exclusion is the whole gate.
//
// Payload-carrying alternatives do not exist yet (core.DataType.Alternatives
// is []string, bare nullary tags), so the probe cannot run a real payload
// fixture. It instead drives the two shapes the payload question REDUCES to,
// both already expressible:
//
//   1. testdata/phase07/call_from_both_match_arms.schway -- whose own header
//      names it "Phase 12's forward guard, since a future `Result` match
//      arm's own call must be picked up the same way". This is the
//      match-plus-call half: a payload match arm containing a call.
//   2. testdata/phase08/relay_depth2_accept.schway -- a callee's result
//      forwarded out through a second hop. This is the forwarding half: a
//      payload constructed in a callee and returned through a relay.
//
// Each assertion below is SINGLE-OUTCOME, never a disjunction -- an
// either/or assertion would be satisfied whichever way a peer answers,
// which is the vacuity this probe exists to avoid.

// probeVerdict records all four peers' answers for one fixture.
type probeVerdict struct {
	fixture             string
	checkClean          bool
	coreValid           bool
	coreProblemCodes    []string
	originProblemCodes  []string
	pathoracleErrCode   string
	pathoracleEndpoints int
}

// runFourPeerProbe drives all four admission peers over one fixture and
// returns their verdicts. check is consulted via loadCheckedProgram (which
// fails the test if check itself reports diagnostics, so a clean load IS
// check's verdict); the other three are called directly, never through
// session, so no peer is consulted through another peer.
func runFourPeerProbe(t *testing.T, phase, fixture string) (core.Program, probeVerdict) {
	t.Helper()

	// check: loadCheckedProgram fails on any diagnostic, so reaching the
	// next line means check admitted the program with zero diagnostics.
	program := loadCheckedProgram(t, phase, fixture)
	verdict := probeVerdict{fixture: phase + "/" + fixture, checkClean: true}

	// corevalidate: the independent reachability-closure peer.
	result := corevalidate.Validate(program)
	verdict.coreValid = result.Valid
	for _, problem := range result.Problems {
		verdict.coreProblemCodes = append(verdict.coreProblemCodes, problem.Code)
	}

	// originvalidate: the body-blind published-origin walker.
	for _, problem := range originvalidate.ValidatePublished(program) {
		verdict.originProblemCodes = append(verdict.originProblemCodes, problem.Code)
	}

	// pathoracle: the independent path enumerator. Exercised per function
	// over the same program, via the narrow callee lookup D-10-02's shape
	// established -- never a whole-core.Program handoff.
	lookup := pathoracle.BuildCalleeLookup(program)
	for _, function := range program.Functions {
		endpoints, paths, err := pathoracle.RecomputeEndpoints(function, lookup)
		if err != nil {
			verdict.pathoracleErrCode = err.Error()
			break
		}
		verdict.pathoracleEndpoints += len(endpoints)
		_ = paths
	}

	return program, verdict
}

// TestC03ResultPayloadOriginAcrossOpCall is the probe proper. It drives the
// four peers over both reduction shapes and asserts the verdicts are
// attributable to the catalogued gaps, never to a new root cause.
func TestC03ResultPayloadOriginAcrossOpCall(t *testing.T) {
	t.Run("match_arm_call_shape", func(t *testing.T) {
		// Phase 12's own forward guard, per this fixture's header. A
		// payload match arm's call must be enumerated the same way.
		_, verdict := runFourPeerProbe(t, "phase07", "call_from_both_match_arms.schway")
		t.Logf("probe verdict: %+v", verdict)

		if !verdict.coreValid {
			t.Fatalf("BRANCH B SIGNAL: corevalidate refuses the match-arm-call shape (%v). 12-CONTEXT.md D-12-04 pre-registered BRANCH A for this shape; a refusal here is a divergence root cause NOT named in D-10-C01/D-11-51/D-11-52, so Phase 12 must be replanned or slipped to M003 rather than absorbing a new rule", verdict.coreProblemCodes)
		}
		if len(verdict.originProblemCodes) != 0 {
			t.Fatalf("BRANCH B SIGNAL: originvalidate refuses the match-arm-call shape (%v) -- not a catalogued gap", verdict.originProblemCodes)
		}
		if verdict.pathoracleErrCode != "" {
			t.Fatalf("BRANCH B SIGNAL: pathoracle errored on the match-arm-call shape (%q) -- not a catalogued gap", verdict.pathoracleErrCode)
		}
		t.Log("BRANCH A for this shape: all four peers agree, no new divergence root cause")
	})

	t.Run("forwarded_callee_result_shape", func(t *testing.T) {
		// A callee's result forwarded out through a relay -- the shape a
		// payload constructed in a callee and returned would take.
		//
		// This fixture's OWN header records the expected, already-known
		// divergence: "corevalidate's own loan-liveness re-derivation ...
		// unconditionally treats an OpCall's target as inheriting every
		// loan its SourceID carries, so it also refuses this fixture
		// (core.move_while_borrowed) even though `check`'s own
		// interprocedural law -- the one under test -- correctly admits
		// it." That header predates Phase 09's peer work, so whether the
		// divergence SURVIVES is itself part of the probe.
		_, verdict := runFourPeerProbe(t, "phase08", "relay_depth2_accept.schway")
		t.Logf("probe verdict: %+v", verdict)

		// Single-outcome: whichever way corevalidate answers, the answer
		// must be attributable. An unexpected PROBLEM CODE is the BRANCH B
		// signal, not the presence or absence of a refusal per se.
		for _, code := range verdict.coreProblemCodes {
			switch code {
			case "core.move_while_borrowed", "core.callee_not_callable":
				// Both are catalogued: the former is this fixture's own
				// recorded residual, the latter is D-10-C01's signature
				// refusal (and D-10-C04's flip target).
			default:
				t.Fatalf("BRANCH B SIGNAL: corevalidate raised %q on the forwarding shape -- a divergence root cause NOT named in D-10-C01/D-11-51/D-11-52. Per 12-CONTEXT.md D-12-04 this means Phase 12 is replanned or slipped to M003; it does not silently absorb a new kernel rule. Full problem set: %v", code, verdict.coreProblemCodes)
			}
		}
		if verdict.pathoracleErrCode != "" {
			t.Fatalf("BRANCH B SIGNAL: pathoracle errored on the forwarding shape (%q) -- not a catalogued gap", verdict.pathoracleErrCode)
		}
		t.Log("BRANCH A for this shape: every peer verdict is attributable to a catalogued gap")
	})
}

// TestC03PeerDeriveOriginFactsOpCallGapStillOpen pins D-10-C01 -- the gap
// the probe's attribution depends on. 12-CONTEXT.md D-12-28/D-12-30 defer
// resource-carrying payloads precisely BECAUSE this gap is open; if it
// closes, this test goes red and that deferral must be revisited rather than
// continuing to cite an obsolete premise.
//
// The gap cannot be pinned by loading an existing fixture, and that is
// itself part of the finding: D-10-C01's own text records that the gap
// "constrained which fixtures Phase 10 plans 10-07/10-08 could express", so
// no committed fixture declares a borrow-returning PublicOrigin sourced from
// forwarding a callee's result -- the shape is inexpressible in testdata
// BECAUSE the gap is open.
//
// So the shape is synthesized at core level instead, Q-01 style: clone
// testdata/phase08/relay_depth2_accept.schway's checked program (where `relay`
// forwards `leaf`'s result and declares no origin at all) and attach a
// borrow-returning PublicOrigin to `relay`. That is a SINGLE additive field
// edit -- every function ID, operation ID, and operation Kind is untouched,
// asserted below -- so the peers' verdict is attributable to the added
// declared origin alone.
func TestC03PeerDeriveOriginFactsOpCallGapStillOpen(t *testing.T) {
	program := loadCheckedProgram(t, "phase08", "relay_depth2_accept.schway")

	// Baseline: with no declared origin, both peers admit this program.
	// Phase 09's D-09-03 closed the older loan-carry divergence here, so a
	// clean baseline is expected and is asserted rather than assumed --
	// without it, a refusal below could not be attributed to the edit.
	if baseline := corevalidate.Validate(program); !baseline.Valid {
		t.Fatalf("baseline is not clean: corevalidate already refuses the unmodified fixture (%+v), so nothing below is attributable to the synthesized origin", baseline.Problems)
	}
	if baseline := originvalidate.ValidatePublished(program); len(baseline) != 0 {
		t.Fatalf("baseline is not clean: originvalidate already refuses the unmodified fixture (%+v)", baseline)
	}

	relayIndex := -1
	for i, function := range program.Functions {
		if function.Name == "relay" {
			relayIndex = i
			break
		}
	}
	if relayIndex < 0 {
		t.Fatal("expected testdata/phase08/relay_depth2_accept.schway to declare a function named relay")
	}
	if program.Functions[relayIndex].PublicOrigin != nil {
		t.Fatal("expected relay to declare no PublicOrigin in the committed fixture -- the synthesized edit assumes it is absent")
	}

	synthesized := cloneCheckedProgramForTest(t, program)
	synthesized.Functions[relayIndex].PublicOrigin = &core.PublicOrigin{
		Paths:  []string{program.Functions[relayIndex].Parameter.Name},
		Access: "shared",
	}

	// Single additive field edit -- prove it before drawing any verdict.
	assertOnlyPublicOriginChanged(t, program, synthesized, relayIndex)

	coreResult := corevalidate.Validate(synthesized)
	originProblems := originvalidate.ValidatePublished(synthesized)
	t.Logf("synthesized borrow-returning origin on relay: corevalidate.Valid=%v problems=%+v; originvalidate problems=%+v", coreResult.Valid, coreResult.Problems, originProblems)

	// D-10-C01's signature: a declared origin satisfiable only by walking
	// THROUGH an OpCall is refused, because peerDeriveOriginFacts has no
	// case core.OpCall and therefore cannot see the callee edge that would
	// justify it. Single-outcome assertion: if BOTH peers now admit the
	// synthesized shape, the gap has closed.
	if coreResult.Valid && len(originProblems) == 0 {
		t.Fatal("D-10-C01 appears CLOSED: a borrow-returning PublicOrigin sourced from forwarding a callee's result is now admitted by both corevalidate and originvalidate. 12-CONTEXT.md D-12-28/D-12-30 defer resource-carrying payloads because this gap is open -- re-evaluate that deferral, the probe's attribution set, and PHASE-11-DEBT.md's D-10-C01 row")
	}
	t.Log("D-10-C01 still open: the synthesized forwarded-origin shape is refused, so payload-forwarding fixtures hit the same wall -- 12-CONTEXT.md D-12-28's premise holds")
}

// assertOnlyPublicOriginChanged proves the synthesized edit touched nothing
// but one function's PublicOrigin: function count, every function ID, every
// operation count, and every operation ID and Kind are identical. This is
// the companion assertion that makes the verdict above attributable to the
// edit rather than to collateral damage from the clone.
func assertOnlyPublicOriginChanged(t *testing.T, before, after core.Program, editedIndex int) {
	t.Helper()
	if len(before.Functions) != len(after.Functions) {
		t.Fatalf("function count changed: before=%d after=%d", len(before.Functions), len(after.Functions))
	}
	for i := range before.Functions {
		b, a := before.Functions[i], after.Functions[i]
		if b.ID != a.ID || b.Name != a.Name || b.ReturnType != a.ReturnType {
			t.Fatalf("function %d identity changed: %q/%q/%q vs %q/%q/%q", i, b.ID, b.Name, b.ReturnType, a.ID, a.Name, a.ReturnType)
		}
		if i == editedIndex {
			if a.PublicOrigin == nil {
				t.Fatal("the synthesized edit did not take effect")
			}
		} else if (b.PublicOrigin == nil) != (a.PublicOrigin == nil) {
			t.Fatalf("function %d PublicOrigin presence changed -- the edit was not confined to index %d", i, editedIndex)
		}
		if (b.Linear == nil) != (a.Linear == nil) {
			t.Fatalf("function %d Linear presence changed", i)
		}
		if b.Linear == nil {
			continue
		}
		if len(b.Linear.Operations) != len(a.Linear.Operations) {
			t.Fatalf("function %d operation count changed: before=%d after=%d", i, len(b.Linear.Operations), len(a.Linear.Operations))
		}
		for j := range b.Linear.Operations {
			if b.Linear.Operations[j].ID != a.Linear.Operations[j].ID {
				t.Fatalf("function %d operation %d ID changed", i, j)
			}
			if b.Linear.Operations[j].Kind != a.Linear.Operations[j].Kind {
				t.Fatalf("function %d operation %d Kind changed", i, j)
			}
		}
	}
}

// TestC03ProbeFixturesAreUnmodified is the companion clean-edit assertion,
// mirroring Phase 11's TestQ01RewrittenProgramStillChecks. The probe above
// draws its verdict from fixtures it must NOT have perturbed -- it reads
// them, it never rewrites them -- so this test proves the probe's inputs are
// the committed fixtures themselves. That is what makes the verdict
// attributable to the peers' behavior alone rather than to the probe
// harness.
func TestC03ProbeFixturesAreUnmodified(t *testing.T) {
	for _, fixture := range []struct{ phase, name string }{
		{"phase07", "call_from_both_match_arms.schway"},
		{"phase08", "relay_depth2_accept.schway"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			first, _ := runFourPeerProbe(t, fixture.phase, fixture.name)
			second := loadCheckedProgram(t, fixture.phase, fixture.name)

			if len(first.Functions) != len(second.Functions) {
				t.Fatalf("probe perturbed the fixture: function count %d vs %d", len(first.Functions), len(second.Functions))
			}
			for i := range first.Functions {
				before, after := first.Functions[i], second.Functions[i]
				if before.ID != after.ID {
					t.Fatalf("probe perturbed function %d ID: %q vs %q", i, before.ID, after.ID)
				}
				if (before.Linear == nil) != (after.Linear == nil) {
					t.Fatalf("probe perturbed function %d Linear presence", i)
				}
				if before.Linear == nil {
					continue
				}
				if len(before.Linear.Operations) != len(after.Linear.Operations) {
					t.Fatalf("probe perturbed function %d operation count: %d vs %d", i, len(before.Linear.Operations), len(after.Linear.Operations))
				}
				for j := range before.Linear.Operations {
					if before.Linear.Operations[j].ID != after.Linear.Operations[j].ID {
						t.Fatalf("probe perturbed function %d operation %d ID", i, j)
					}
					if before.Linear.Operations[j].Kind != after.Linear.Operations[j].Kind {
						t.Fatalf("probe perturbed function %d operation %d Kind", i, j)
					}
				}
			}
		})
	}
}
