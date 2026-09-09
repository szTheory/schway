package corevalidate_test

import (
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// checkFallibleForeignReachFixture parses and checks 07-12's own standing
// witness (testdata/phase07/call_fallible_foreign_reach.lang) through
// check.Program directly, mirroring mustCheckPhase07Fixture's own path
// resolution.
func checkFallibleForeignReachFixture(t testing.TB) check.Result {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "call_fallible_foreign_reach.lang"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("call_fallible_foreign_reach.lang failed to parse: %+v", parsed.Diagnostics)
	}
	result := check.Program(parsed.Program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("call_fallible_foreign_reach.lang: unexpected check diagnostics: %+v", result.Diagnostics)
	}
	return result
}

func functionIDByName(program core.Program, name string) string {
	for _, function := range program.Functions {
		if function.Name == name {
			return function.ID
		}
	}
	return ""
}

// TestPeerForeignReachIsClosureDerived is Task 2 Test 1's Foreign half
// (CR-03/PVG-02): corevalidate's own PeerSignatures()["...fn:main"]
// carries the SAME joined Foreign as the producer's own summary --
// derived independently, over corevalidate's own postorder, and asserted
// equal.
func TestPeerForeignReachIsClosureDerived(t *testing.T) {
	checked := checkFallibleForeignReachFixture(t)
	mainID := functionIDByName(checked.Program, "main")
	if mainID == "" {
		t.Fatal("expected call_fallible_foreign_reach.lang to declare main")
	}

	producerSummary, err := originvalidate.BuildInterface(checked.Program)
	if err != nil {
		t.Fatalf("originvalidate.BuildInterface: %v", err)
	}
	var producerMain core.FunctionSignature
	for _, function := range producerSummary.Functions {
		if function.ID == mainID {
			producerMain = function
		}
	}

	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("expected call_fallible_foreign_reach.lang to corevalidate-validate, got problems: %+v", validated.Problems)
	}
	peerMain, ok := validated.PeerSignatures()[mainID]
	if !ok {
		t.Fatal("main missing from peer signatures")
	}

	if peerMain.Foreign != producerMain.Foreign {
		t.Fatalf("expected the peer's independently-joined Foreign to match the producer's: peer=%+v producer=%+v", peerMain.Foreign, producerMain.Foreign)
	}
	wantForeign := core.ForeignReach{Allocator: "libc_malloc", Unwind: "forbidden", NonlocalExit: "forbidden"}
	if peerMain.Foreign != wantForeign {
		t.Fatalf("expected the peer to join main's Foreign to %+v, got %+v", wantForeign, peerMain.Foreign)
	}
}

// TestPeerFailsIsClosureDerived is Task 2 Test 1's Fails half
// (CR-03/PVG-02): the peer's own joined Fails matches the producer's.
func TestPeerFailsIsClosureDerived(t *testing.T) {
	checked := checkFallibleForeignReachFixture(t)
	mainID := functionIDByName(checked.Program, "main")
	if mainID == "" {
		t.Fatal("expected call_fallible_foreign_reach.lang to declare main")
	}

	producerSummary, err := originvalidate.BuildInterface(checked.Program)
	if err != nil {
		t.Fatalf("originvalidate.BuildInterface: %v", err)
	}
	var producerMain core.FunctionSignature
	for _, function := range producerSummary.Functions {
		if function.ID == mainID {
			producerMain = function
		}
	}

	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("expected call_fallible_foreign_reach.lang to corevalidate-validate, got problems: %+v", validated.Problems)
	}
	peerMain, ok := validated.PeerSignatures()[mainID]
	if !ok {
		t.Fatal("main missing from peer signatures")
	}

	if peerMain.Fails != producerMain.Fails {
		t.Fatalf("expected the peer's independently-joined Fails to match the producer's: peer=%q producer=%q", peerMain.Fails, producerMain.Fails)
	}
	if peerMain.Fails != "ProbeError" {
		t.Fatalf("expected the peer to join main's Fails to %q, got %q", "ProbeError", peerMain.Fails)
	}
}

// TestForeignClosureJoinPeersIndependent is Task 2 Test 2/Test 5: proves
// independence directly rather than assuming it. With the PRODUCER's own
// foreignClosureJoinSeam engaged (originvalidate's join disabled), the
// PEER still publishes the joined reach -- and producer/peer parity
// BREAKS, since the producer is now wrongly reporting the zero value while
// the peer reports the correct joined one. With the PEER's own seam
// engaged instead (producer's join left on), parity breaks in the other
// direction. Each direction proves the corresponding side is genuinely
// DERIVING the join, not copying the other's answer.
//
// Producer reads: chainOrder (from callgraph.Order, the check-side
// traversal) plus calleeIDsForClosureDigest(function). Peer reads: its own
// checkCallGraphAcyclic-derived v.peerPostorder/v.peerAdjacency, built from
// the core artifact's own operation stream -- never callgraph's or
// originvalidate's.
func TestForeignClosureJoinPeersIndependent(t *testing.T) {
	checked := checkFallibleForeignReachFixture(t)
	mainID := functionIDByName(checked.Program, "main")

	t.Run("producer seamed off, peer still joins", func(t *testing.T) {
		restore := originvalidate.SetForeignClosureJoinSeam(true)
		defer restore()

		producerSummary, err := originvalidate.BuildInterface(checked.Program)
		if err != nil {
			t.Fatalf("originvalidate.BuildInterface: %v", err)
		}
		var producerMain core.FunctionSignature
		for _, function := range producerSummary.Functions {
			if function.ID == mainID {
				producerMain = function
			}
		}
		if producerMain.Foreign != (core.ForeignReach{}) {
			t.Fatalf("expected the seamed-off producer to wrongly publish the zero ForeignReach, got %+v", producerMain.Foreign)
		}

		validated := corevalidate.Validate(checked.Program)
		if !validated.Valid {
			t.Fatalf("expected the program to still corevalidate-validate, got problems: %+v", validated.Problems)
		}
		peerMain := validated.PeerSignatures()[mainID]
		if peerMain.Foreign == (core.ForeignReach{}) {
			t.Fatal("expected the peer to still independently join Foreign with the producer's own join seamed off")
		}
		if peerMain.Foreign == producerMain.Foreign {
			t.Fatal("expected producer/peer parity to BREAK when only the producer's join is disabled")
		}
	})

	t.Run("peer seamed off, producer still joins", func(t *testing.T) {
		restore := corevalidate.SetDisableForeignClosureJoinPeerForTest(true)
		defer restore()

		producerSummary, err := originvalidate.BuildInterface(checked.Program)
		if err != nil {
			t.Fatalf("originvalidate.BuildInterface: %v", err)
		}
		var producerMain core.FunctionSignature
		for _, function := range producerSummary.Functions {
			if function.ID == mainID {
				producerMain = function
			}
		}
		if producerMain.Foreign == (core.ForeignReach{}) {
			t.Fatal("expected the (unseamed) producer to still join Foreign")
		}

		validated := corevalidate.Validate(checked.Program)
		if !validated.Valid {
			t.Fatalf("expected the program to still corevalidate-validate, got problems: %+v", validated.Problems)
		}
		peerMain := validated.PeerSignatures()[mainID]
		if peerMain.Foreign != (core.ForeignReach{}) {
			t.Fatalf("expected the seamed-off peer to wrongly publish the zero ForeignReach, got %+v", peerMain.Foreign)
		}
		if peerMain.Foreign == producerMain.Foreign {
			t.Fatal("expected producer/peer parity to BREAK when only the peer's join is disabled")
		}
	})
}

// grep-based cross-package import boundary assertions (Task 2 Test 3):
// the peer's join must never import the producer's package, and the
// producer must never import the peer's -- both already asserted
// elsewhere in this package's own test suite
// (TestValidatorImportsStayIndependent); this file adds no new import
// boundary, it only exercises the join behavior above through the real
// public API on both sides.

// TestForeignClosureJoinPeerMutationMatrix is Task 3's peer-side kill for
// control:summary.foreign_reach_closure_derived and
// control:summary.fails_closure_derived: with the peer's own seams
// engaged, main WRONGLY publishes the zero ForeignReach / fails="" on the
// PEER side; restoring each seam restores the correct joined value.
func TestForeignClosureJoinPeerMutationMatrix(t *testing.T) {
	checked := checkFallibleForeignReachFixture(t)
	mainID := functionIDByName(checked.Program, "main")

	t.Run("foreign", func(t *testing.T) {
		restore := corevalidate.SetDisableForeignClosureJoinPeerForTest(true)
		mutated := corevalidate.Validate(checked.Program)
		restore()
		if !mutated.Valid {
			t.Fatalf("expected the program to still corevalidate-validate, got problems: %+v", mutated.Problems)
		}
		mutatedMain := mutated.PeerSignatures()[mainID]
		if mutatedMain.Foreign != (core.ForeignReach{}) {
			t.Fatalf("mutation (disabling the peer's Foreign join) had no observable effect: main still published %+v", mutatedMain.Foreign)
		}

		restored := corevalidate.Validate(checked.Program)
		if !restored.Valid {
			t.Fatalf("expected the program to still corevalidate-validate, got problems: %+v", restored.Problems)
		}
		restoredMain := restored.PeerSignatures()[mainID]
		if restoredMain.Foreign == (core.ForeignReach{}) {
			t.Fatal("expected the peer's Foreign join to be restored once the seam is disengaged")
		}
	})

	t.Run("fails", func(t *testing.T) {
		restore := corevalidate.SetDisableFailsClosureJoinPeerForTest(true)
		mutated := corevalidate.Validate(checked.Program)
		restore()
		if !mutated.Valid {
			t.Fatalf("expected the program to still corevalidate-validate, got problems: %+v", mutated.Problems)
		}
		mutatedMain := mutated.PeerSignatures()[mainID]
		if mutatedMain.Fails != "" {
			t.Fatalf("mutation (disabling the peer's Fails join) had no observable effect: main still published fails=%q", mutatedMain.Fails)
		}

		restored := corevalidate.Validate(checked.Program)
		if !restored.Valid {
			t.Fatalf("expected the program to still corevalidate-validate, got problems: %+v", restored.Problems)
		}
		restoredMain := restored.PeerSignatures()[mainID]
		if restoredMain.Fails == "" {
			t.Fatal("expected the peer's Fails join to be restored once the seam is disengaged")
		}
	})
}
