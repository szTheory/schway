package corevalidate_test

// This file lives in package corevalidate_test (external), following
// corevalidate_summary_peer_test.go:1-2's own documented distinction: it
// needs to import originvalidate and session as ORDINARY test dependencies
// to drive TestPeerFirstHopWinsMatchesRecomputeOriginPerReturn's
// equivalence proof against the producer's own RecomputeOrigin, and to
// parse/check real .schway fixtures directly. This is NOT the independence
// property under test -- that property is corevalidate.go's OWN production
// source never importing originvalidate, check, or callgraph
// (TestArbitraryMaskCannotEnterCoreValidation and its siblings in
// corevalidate_test.go / corevalidate_exclusive_test.go, which read
// corevalidate.go's file text directly and are unaffected by what this
// _test.go file imports).

import (
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// functionByName returns the first function in program named name, failing
// the test if none matches.
func functionByName(t *testing.T, program core.Program, name string) core.Function {
	t.Helper()
	for _, function := range program.Functions {
		if function.Name == name {
			return function
		}
	}
	t.Fatalf("expected program to declare a function named %q", name)
	return core.Function{}
}

// singleFunctionOriginProgram wraps one already-checked core.Function in a
// fresh, minimal core.Program -- used to isolate ONE function from a
// multi-function fixture (relay_escort_witness.schway) whose OWN program-wide
// check verdict is refused for reasons unrelated to the isolated function
// under test.
func singleFunctionOriginProgram(module string, function core.Function) core.Program {
	return core.Program{Schema: core.Schema1, Module: module, ModuleID: "s1:" + module + ":module", Functions: []core.Function{function}}
}

// peerCallableFor validates program and returns the peer's own Callable
// verdict for functionID, failing the test if the function is missing from
// PeerSignatures.
func peerCallableFor(t *testing.T, program core.Program, functionID string) bool {
	t.Helper()
	result := corevalidate.Validate(program)
	signature, ok := result.PeerSignatures()[functionID]
	if !ok {
		t.Fatalf("function %s missing from peer signatures", functionID)
	}
	return signature.Callable
}

// functionHasOpCall reports whether function's body contains any
// core.OpCall operation.
func functionHasOpCall(function *core.Function) bool {
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

// TestPeerCallableContainmentMatchesPublishProblemsFor is Task 1 Test 4
// (D-09-19): peerCallable's new containment check reaches the same verdict
// as originvalidate.PublishProblemsFor for honest single- and multi-hop
// declarations (Callable == true) and for mutated understated/mismatched
// declarations an honest producer can never construct itself (Callable ==
// false, matching PublishProblemsFor's own refusal).
func TestPeerCallableContainmentMatchesPublishProblemsFor(t *testing.T) {
	t.Run("single shared hop: honest declaration agrees Callable true", func(t *testing.T) {
		program := loadCheckedProgram(t, "phase3", "public_view.schway")
		function := functionByName(t, program, "view")
		if function.PublicOrigin == nil || function.PublicOrigin.Access != "shared" {
			t.Fatalf("expected an honest shared PublicOrigin, got %+v", function.PublicOrigin)
		}
		if problems := originvalidate.PublishProblemsFor(function, originvalidate.BuildCalleeOriginFacts(program)); len(problems) != 0 {
			t.Fatalf("expected the producer to admit this honest declaration, got %+v", problems)
		}
		if !peerCallableFor(t, program, function.ID) {
			t.Fatal("expected the peer to agree Callable == true on an honest shared declaration")
		}
	})

	t.Run("single exclusive hop: honest declaration agrees Callable true", func(t *testing.T) {
		// relay_escort_witness.schway's own `escort` function refuses via
		// check.interprocedural_loan_liveness (D-03-02's interprocedural
		// half, closed at the check layer in Phase 08); check never clears
		// result.Program on that refusal, so `relay` -- the single-hop,
		// non-refused function within the SAME checked program -- is still
		// a genuine, fully-formed core.Function this test can isolate to
		// exercise an honest single-hop EXCLUSIVE declaration (no such
		// fixture exists as a standalone, checked-clean program).
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "relay_escort_witness.schway"))
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(source)
		relay := functionByName(t, checked.Program, "relay")
		if relay.PublicOrigin == nil || relay.PublicOrigin.Access != "exclusive" {
			t.Fatalf("expected relay's honest declaration to be exclusive, got %+v", relay.PublicOrigin)
		}
		if problems := originvalidate.PublishProblemsFor(relay, originvalidate.BuildCalleeOriginFacts(checked.Program)); len(problems) != 0 {
			t.Fatalf("expected the producer to admit relay's honest declaration, got %+v", problems)
		}
		program := singleFunctionOriginProgram("test.exclusive_hop", relay)
		if !peerCallableFor(t, program, relay.ID) {
			t.Fatal("expected the peer to agree Callable == true on an honest exclusive declaration")
		}
	})

	// D-09-19/D-09-16: the plan's own <behavior> prose for this task
	// describes "a function whose body takes an exclusive borrow and then
	// a shared borrow of the already-derived place" as deriving
	// "exclusive," with "the first hop's mode" winning. That description
	// conflicts with originvalidate.RecomputeOriginPerReturn's own
	// documented and mechanical rule (originvalidate.go:130-142): its walk
	// is BACKWARD from the return, so the hop CLOSEST to the return (the
	// LAST hop in forward/chronological order -- here, the shared
	// reborrow) is "first seen" in that walk and wins; the farther
	// exclusive hop must NOT overwrite it.
	// testdata/phase3/public_view_mixed_access.schway is this EXACT real,
	// committed, checked-clean fixture (03-REVIEW.md CR-01's own
	// regression witness), and its own doc comment confirms the required
	// answer is "shared," not "exclusive." This test asserts the
	// mechanically correct, producer-equivalent answer (matching
	// RecomputeOrigin, proven generally by
	// TestPeerFirstHopWinsMatchesRecomputeOriginPerReturn below) rather
	// than the plan's inverted prose -- see the executor's SUMMARY.md
	// deviation entry for the full account. Reporting the disagreement
	// rather than silently forcing the peer to match invented prose is
	// exactly what Test 3's own escape hatch instructs.
	t.Run("two-hop reborrow chain: closest-to-return hop wins, matching RecomputeOrigin", func(t *testing.T) {
		program := loadCheckedProgram(t, "phase3", "public_view_mixed_access.schway")
		function := functionByName(t, program, "view")
		calleeContracts := originvalidate.BuildCalleeOriginFacts(program)
		recomputedPaths, recomputedAccess, ok := originvalidate.RecomputeOrigin(function, calleeContracts)
		if !ok || recomputedAccess != "shared" {
			t.Fatalf("expected RecomputeOrigin to derive a shared access for this fixture's own body, got paths=%v access=%q ok=%v", recomputedPaths, recomputedAccess, ok)
		}
		if function.PublicOrigin.Access != "exclusive" {
			t.Fatalf("expected the fixture's own declaration to be exclusive (the mismatch this fixture exists to exercise), got %q", function.PublicOrigin.Access)
		}
		if problems := originvalidate.PublishProblemsFor(function, calleeContracts); len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
			t.Fatalf("expected the producer to refuse this honest-but-mismatched declaration with core.origin_access_mismatch, got %+v", problems)
		}
		if peerCallableFor(t, program, function.ID) {
			t.Fatal("expected the peer to independently refuse (Callable == false), matching the producer's core.origin_access_mismatch verdict")
		}

		// Correcting the declaration to the body's ACTUAL derived access
		// (shared) makes the peer agree Callable == true -- proving the
		// refusal above tracked the real mismatch, not a bug in the
		// peer's own two-hop derivation.
		corrected := program
		corrected.Functions = append([]core.Function{}, program.Functions...)
		for i := range corrected.Functions {
			if corrected.Functions[i].ID == function.ID {
				corrected.Functions[i].PublicOrigin = &core.PublicOrigin{Paths: recomputedPaths, Access: recomputedAccess}
			}
		}
		if !peerCallableFor(t, corrected, function.ID) {
			t.Fatal("expected the peer to agree Callable == true once the declaration is corrected to the body's real derived access")
		}
	})

	t.Run("core.origin_understated: peer now independently refuses", func(t *testing.T) {
		program := loadCheckedProgram(t, "phase3", "public_view_understated.schway")
		program.Functions[0].PublicOrigin.Paths = []string{}
		function := program.Functions[0]
		if problems := originvalidate.PublishProblemsFor(function, originvalidate.BuildCalleeOriginFacts(program)); len(problems) != 1 || problems[0].Code != "core.origin_understated" {
			t.Fatalf("expected the mutated declaration to be refused as core.origin_understated, got %+v", problems)
		}
		if peerCallableFor(t, program, function.ID) {
			t.Fatal("expected the peer to independently refuse (Callable == false) a core.origin_understated function")
		}
	})

	t.Run("core.origin_access_mismatch: peer now independently refuses", func(t *testing.T) {
		program := loadCheckedProgram(t, "phase3", "public_view_impossible.schway")
		program.Functions[0].PublicOrigin.Access = "exclusive"
		function := program.Functions[0]
		if problems := originvalidate.PublishProblemsFor(function, originvalidate.BuildCalleeOriginFacts(program)); len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
			t.Fatalf("expected the mutated declaration to be refused as core.origin_access_mismatch, got %+v", problems)
		}
		if peerCallableFor(t, program, function.ID) {
			t.Fatal("expected the peer to independently refuse (Callable == false) a core.origin_access_mismatch function")
		}
	})
}

// TestPeerFirstHopWinsMatchesRecomputeOriginPerReturn is Task 1 Test 3
// (D-09-19): across every checked-clean corpus fixture (corpusFixtures,
// corevalidate_summary_peer_test.go) whose function declares a
// PublicOrigin and whose body never crosses an OpCall, the peer's own
// forward-derived access agrees with originvalidate.RecomputeOrigin's
// combined access -- proven by declaring EXACTLY the recomputed answer and
// observing the peer's containment check accept it, never assumed. A
// function whose body crosses an OpCall is excluded: this equivalence
// proof compares originvalidate.RecomputeOrigin's answer for ONE function
// against corevalidate's peer, which propagates across a call boundary
// using its OWN forward-derived peerLoanCarryFact map (D-09-03), never
// originvalidate's declared-contract map (D-10-01/D-10-03/D-10-06) -- the
// two packages' cross-function propagation mechanisms are independent by
// design (T-03-02/T-03-03), so a cross-call comparison here would not be
// testing the SAME derivation twice, only two different ones that happen
// to often agree. Matching peerDeriveOriginFacts' own deliberate choice
// never to cross an OpCall in its general propagation (D-09-20's
// function-local boundary is reserved for the foreign class alone). If any
// intraprocedural fixture disagreed, that would falsify the equivalence
// claim; this test reports such a disagreement via t.Fatalf rather than
// adjusting the peer to match by construction.
func TestPeerFirstHopWinsMatchesRecomputeOriginPerReturn(t *testing.T) {
	checkedAny := false
	for _, program := range corpusFixtures(t) {
		calleeContracts := originvalidate.BuildCalleeOriginFacts(program)
		for i := range program.Functions {
			function := &program.Functions[i]
			if function.PublicOrigin == nil {
				continue
			}
			if functionHasOpCall(function) {
				continue
			}
			checkedAny = true
			recomputedPaths, recomputedAccess, ok := originvalidate.RecomputeOrigin(*function, calleeContracts)
			if !ok {
				t.Fatalf("function %s declares a PublicOrigin but RecomputeOrigin reports not-derived -- an inconsistent corpus fixture", function.ID)
			}
			original := function.PublicOrigin
			function.PublicOrigin = &core.PublicOrigin{Paths: append([]string{}, recomputedPaths...), Access: recomputedAccess}
			if !peerCallableFor(t, program, function.ID) {
				t.Fatalf("EQUIVALENCE FALSIFIED for function %s: peer's forward-derived access disagrees with RecomputeOriginPerReturn's answer (paths=%v access=%q) -- declaring exactly the recomputed origin was refused", function.ID, recomputedPaths, recomputedAccess)
			}
			function.PublicOrigin = original
		}
	}
	if !checkedAny {
		t.Fatal("expected at least one intraprocedural PublicOrigin-declaring function to be compared")
	}
}

// TestPeerForeignOriginOmittedIsRederived is Task 2 Test 1 (D-09-20):
// testdata/phase4/foreign_origin_omitted.schway's own honest ForeignContract
// (Alias == "borrow", no declared PublicOrigin) is independently refused
// by the peer, matching the producer's core.foreign_origin_omitted
// verdict.
func TestPeerForeignOriginOmittedIsRederived(t *testing.T) {
	program := loadCheckedProgram(t, "phase4", "foreign_origin_omitted.schway")
	function := program.Functions[0]
	if function.ForeignContract == nil || function.ForeignContract.Alias != "borrow" {
		t.Fatalf("expected the fixture's own foreign contract to declare alias \"borrow\", got %+v", function.ForeignContract)
	}
	if problems := originvalidate.PublishProblemsFor(function, originvalidate.BuildCalleeOriginFacts(program)); len(problems) != 1 || problems[0].Code != "core.foreign_origin_omitted" {
		t.Fatalf("expected the producer to refuse with core.foreign_origin_omitted, got %+v", problems)
	}
	if peerCallableFor(t, program, function.ID) {
		t.Fatal("expected the peer to independently refuse (Callable == false), matching the producer's core.foreign_origin_omitted verdict")
	}
}

// TestPeerForeignOriginOmittedDoesNotOverRefuse is Task 2 Test 2's negative
// control (D-09-20): the same function with Alias mutated outside
// {"borrow", "retain"} makes the peer agree Callable == true -- the class
// must not over-refuse a foreign call with no borrow/retain obligation.
func TestPeerForeignOriginOmittedDoesNotOverRefuse(t *testing.T) {
	program := loadCheckedProgram(t, "phase4", "foreign_origin_omitted.schway")
	program.Functions[0].ForeignContract.Alias = ""
	function := program.Functions[0]
	if problems := originvalidate.PublishProblemsFor(function, originvalidate.BuildCalleeOriginFacts(program)); len(problems) != 0 {
		t.Fatalf("expected the producer to admit an alias-less foreign contract, got %+v", problems)
	}
	if !peerCallableFor(t, program, function.ID) {
		t.Fatal("expected the peer's negative control (alias outside {borrow, retain}) to still agree Callable == true -- the class must not over-refuse")
	}
}

// TestPeerOriginContainmentDisabledFalselyAgreesAgain is Task 3's QLT-08
// mutation-kill (D-09-19/D-09-20): with each new fault-injection seam
// engaged, the corresponding class's mutation case goes back to falsely
// agreeing (Callable == true); with the seam restored, it refuses again. A
// control that has never been seen to fail is a claim, not evidence -- this
// is what makes Task 1 and Task 2's containment/foreign checks evidence
// rather than an unfalsified assertion.
func TestPeerOriginContainmentDisabledFalselyAgreesAgain(t *testing.T) {
	t.Run("origin containment: disabling the seam reintroduces the understated false agreement", func(t *testing.T) {
		program := loadCheckedProgram(t, "phase3", "public_view_understated.schway")
		program.Functions[0].PublicOrigin.Paths = []string{}
		function := program.Functions[0]

		restore := corevalidate.SetDisablePeerOriginContainmentForTest(true)
		falselyAgrees := peerCallableFor(t, program, function.ID)
		restore()
		if !falselyAgrees {
			t.Fatal("expected the peer to falsely agree (Callable == true) with the containment seam disabled")
		}
		if peerCallableFor(t, program, function.ID) {
			t.Fatal("expected the peer to refuse once the containment seam is restored")
		}
	})

	t.Run("origin containment: disabling the seam reintroduces the access-mismatch false agreement", func(t *testing.T) {
		program := loadCheckedProgram(t, "phase3", "public_view_impossible.schway")
		program.Functions[0].PublicOrigin.Access = "exclusive"
		function := program.Functions[0]

		restore := corevalidate.SetDisablePeerOriginContainmentForTest(true)
		falselyAgrees := peerCallableFor(t, program, function.ID)
		restore()
		if !falselyAgrees {
			t.Fatal("expected the peer to falsely agree (Callable == true) with the containment seam disabled")
		}
		if peerCallableFor(t, program, function.ID) {
			t.Fatal("expected the peer to refuse once the containment seam is restored")
		}
	})

	t.Run("foreign-origin-omitted: disabling the seam reintroduces the false agreement", func(t *testing.T) {
		program := loadCheckedProgram(t, "phase4", "foreign_origin_omitted.schway")
		function := program.Functions[0]

		restore := corevalidate.SetDisableForeignOriginPeerForTest(true)
		falselyAgrees := peerCallableFor(t, program, function.ID)
		restore()
		if !falselyAgrees {
			t.Fatal("expected the peer to falsely agree (Callable == true) with the foreign-origin seam disabled")
		}
		if peerCallableFor(t, program, function.ID) {
			t.Fatal("expected the peer to refuse once the foreign-origin seam is restored")
		}
	})
}
