package corevalidate_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// corpusFixtures walks testdata/phase1..phase4 plus testdata/phase07 and
// returns every fixture that checks clean (zero session.Check diagnostics)
// -- the same "never reaches a real producer otherwise" gate
// originvalidate_test.go's TestInterfaceV1FieldInvariantsAcrossCorpus and
// TestPublishProblemsForMatchesValidatePublishedAcrossCorpus already use.
// This is D-07-25's Stage 0 gate corpus.
func corpusFixtures(t *testing.T) []core.Program {
	t.Helper()
	var programs []core.Program
	for _, phase := range []string{"phase1", "phase2", "phase3", "phase4", "phase07"} {
		dir := testsupport.ProjectPath("testdata", phase)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lang") {
				continue
			}
			source, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatalf("read %s/%s: %v", phase, entry.Name(), err)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) > 0 {
				continue
			}
			programs = append(programs, checked.Program)
		}
	}
	if len(programs) == 0 {
		t.Fatal("expected at least one checked-clean fixture across the corpus")
	}
	return programs
}

// signatureByID indexes an Interface's Functions by ID for lookup against
// corevalidate's PeerSignatures map, which is already ID-keyed.
func signatureByID(summary core.Interface) map[string]core.FunctionSignature {
	out := make(map[string]core.FunctionSignature, len(summary.Functions))
	for _, function := range summary.Functions {
		out[function.ID] = function
	}
	return out
}

// structuralFieldsEqual compares every field of a lang.interface/1
// FunctionSignature EXCEPT Callable and ClosureDigest: ClosureDigest is a
// digest over the signature itself (comparing it is circular, not
// independent evidence), and Callable's narrowed peer agreement is this
// file's own separate claim (TestSummaryPeerCallableAgreesOnOriginOmittedClass),
// not this blanket structural claim.
func structuralFieldsEqual(a, b core.FunctionSignature) bool {
	return a.ID == b.ID &&
		a.Name == b.Name &&
		reflect.DeepEqual(a.Parameters, b.Parameters) &&
		reflect.DeepEqual(a.Return, b.Return) &&
		reflect.DeepEqual(a.Abilities, b.Abilities) &&
		a.Fails == b.Fails &&
		reflect.DeepEqual(a.Foreign, b.Foreign)
}

// TestSummaryPeerStructuralFieldsMatchProducerAcrossCorpus is 07-02 Task 2's
// Test 1 (D-07-25): over every function in the whole Stage 0 gate corpus,
// corevalidate's independently-derived PeerSignatures structural fields
// (Parameters, Return, Abilities, Fails, Foreign) equal
// originvalidate.BuildInterface's producer output, field for field, with
// zero divergence. Callable is asserted separately
// (TestSummaryPeerCallableAgreesOnOriginOmittedClass) because D-07-33
// deliberately narrows it.
// relayEscortWitnessModule is 07-05 Task 2's D-03-02/D-07-44 named
// exception to this test's own "checked-clean implies corevalidate-valid"
// invariant: testdata/phase07/relay_escort_witness.lang is DELIBERATELY a
// fixture that checks clean under `check`'s current (intraprocedural) loan
// liveness but is independently refused by `corevalidate`'s own replay with
// core.move_while_borrowed -- the INTERPROCEDURAL half of D-03-02, left
// open until Phase 08/09 closes it (see check_test.go's
// TestRelayEscortWitnessChecksCleanPendingInterproceduralLiveness for the
// full account of why `check` misses it and corevalidate does not). This is
// the ONE named exception in the whole corpus; every other fixture keeps
// the strict invariant below. See
// TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed
// for the decisive, tested assertion of the divergence itself -- this is
// not a silent skip.
const relayEscortWitnessModule = "phase07.relay_escort_witness"

func TestSummaryPeerStructuralFieldsMatchProducerAcrossCorpus(t *testing.T) {
	checkedAny := false
	for _, program := range corpusFixtures(t) {
		producerSummary, err := originvalidate.BuildInterface(program)
		if err != nil {
			t.Fatalf("BuildInterface: %v", err)
		}
		producerByID := signatureByID(producerSummary)

		result := corevalidate.Validate(program)
		if !result.Valid {
			if program.Module == relayEscortWitnessModule {
				continue
			}
			t.Fatalf("expected a checked-clean corpus program to also corevalidate-validate, got problems: %+v", result.Problems)
		}
		peerByID := result.PeerSignatures()

		for _, function := range program.Functions {
			checkedAny = true
			producerSignature, ok := producerByID[function.ID]
			if !ok {
				t.Fatalf("function %s missing from producer summary", function.ID)
			}
			peerSignature, ok := peerByID[function.ID]
			if !ok {
				t.Fatalf("function %s missing from peer signatures -- summary peer did not run", function.ID)
			}
			if !structuralFieldsEqual(producerSignature, peerSignature) {
				t.Fatalf("function %s: peer signature diverges from producer on structural fields:\nproducer=%+v\npeer=%+v", function.ID, producerSignature, peerSignature)
			}
		}
	}
	if !checkedAny {
		t.Fatal("expected at least one function to be compared across the corpus")
	}
}

// TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed is
// 07-05 Task 2's decisive, named assertion of the
// relayEscortWitnessModule exception above: `check` admits
// testdata/phase07/relay_escort_witness.lang (asserted by
// check_test.go's TestRelayEscortWitnessChecksCleanPendingInterproceduralLiveness),
// but corevalidate's own, independently-implemented replay refuses it with
// core.move_while_borrowed on escort's `take buffer` operation --
// `escort`'s exclusive loan on `buffer` (created for the call argument
// `borrowed`) is still active, from corevalidate's own bookkeeping's point
// of view, when `buffer` is moved. This is D-03-02's INTERPROCEDURAL half,
// caught by one peer and missed by the other -- exactly the divergence the
// two-validator architecture exists to surface, not hide. It remains open
// until Phase 08/09's interprocedural loan-liveness work makes both sides
// agree (by refusing, never by both silently accepting).
func TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "relay_escort_witness.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("expected check to admit the fixture (the D-03-02 finding), got %+v", checked.Diagnostics)
	}
	if checked.Program.Module != relayEscortWitnessModule {
		t.Fatalf("expected module %s, got %s", relayEscortWitnessModule, checked.Program.Module)
	}
	result := corevalidate.Validate(checked.Program)
	if result.Valid {
		t.Fatal("expected corevalidate to independently refuse this program, got Valid == true")
	}
	found := false
	for _, problem := range result.Problems {
		if problem.Code == "core.move_while_borrowed" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a core.move_while_borrowed problem, got %+v", result.Problems)
	}
}

// straightLineProgram and branchProgram are the fixed fixtures Test 2 uses
// to exercise the summary peer through each replay shape, per the
// 07-02-PLAN.md read_first pointers (pathoracle_test.go:46-76's
// TestOracleImportsStayIndependent shape, and testdata/phase3's own
// straight-line/branch pair).
func straightLineProgram(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", "public_view.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

func branchProgram(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", "branch_one_arm_shared_accept.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

// mergeProgramsForReplayCoverage builds ONE core.Program carrying both a
// straight-line function and a branch (match-with-linear-arms) function, so
// a single corevalidate.Validate call's Result.PeerSiteCoverage can assert
// BOTH replayStraightLine and replayBlocks actually ran the peer within the
// same run, rather than inferring coverage from two separate calls.
func mergeProgramsForReplayCoverage(t *testing.T) core.Program {
	t.Helper()
	straight := straightLineProgram(t)
	branch := branchProgram(t)
	merged := core.Program{
		Schema:    core.Schema1,
		Module:    "phase07.summary_peer_replay_coverage",
		ModuleID:  "s1:phase07.summary_peer_replay_coverage",
		DataTypes: append(append([]core.DataType{}, straight.DataTypes...), branch.DataTypes...),
		Functions: append(append([]core.Function{}, straight.Functions...), branch.Functions...),
	}
	return merged
}

// TestSummaryPeerRunsAtBothReplaySites is 07-02 Task 2's Test 2 (D-07-21): a
// match-bodied fixture exercises the peer through replayBlocks and a
// straight-line fixture exercises it through replayStraightLine, in the
// SAME Validate() call, and the test fails if either site did not run --
// wiring one and not the other reproduces the literal D-02-03/D-03-01
// repeat.
func TestSummaryPeerRunsAtBothReplaySites(t *testing.T) {
	merged := mergeProgramsForReplayCoverage(t)
	result := corevalidate.Validate(merged)
	if !result.Valid {
		t.Fatalf("expected the merged straight-line + branch program to validate, got problems: %+v", result.Problems)
	}
	straightLineRan, blocksRan := result.PeerSiteCoverage()
	if !straightLineRan {
		t.Fatal("expected the summary peer to have run at replayStraightLine")
	}
	if !blocksRan {
		t.Fatal("expected the summary peer to have run at replayBlocks")
	}
	if len(merged.Functions) != len(result.PeerSignatures()) {
		t.Fatalf("expected a peer signature for every function in the merged program, got %d for %d functions", len(result.PeerSignatures()), len(merged.Functions))
	}
}

// classifyPublicationProblem reports the single Code
// originvalidate.PublishProblemsFor returns for function within program, or
// "" if publication succeeds (no problem). Used only to SELECT which
// functions Test 3 may compare Callable over -- it does not feed back into
// either derivation under test.
func classifyPublicationProblem(t *testing.T, program core.Program, functionID string) string {
	t.Helper()
	for _, function := range program.Functions {
		if function.ID != functionID {
			continue
		}
		if problems := originvalidate.PublishProblemsFor(function); len(problems) > 0 {
			return problems[0].Code
		}
		return ""
	}
	t.Fatalf("function %s not found in program", functionID)
	return ""
}

// TestSummaryPeerCallableAgreesOnOriginOmittedClass is 07-02 Task 2's Test 3
// (D-07-33): the peer's narrowed Callable re-derivation agrees with the
// producer's for every function whose only publication problem class is
// core.origin_omitted, or which has no publication problem at all --
// including testdata/phase07/clean_but_unpublishable.lang.
func TestSummaryPeerCallableAgreesOnOriginOmittedClass(t *testing.T) {
	comparedOmitted := false
	for _, program := range corpusFixtures(t) {
		producerSummary, err := originvalidate.BuildInterface(program)
		if err != nil {
			t.Fatalf("BuildInterface: %v", err)
		}
		producerByID := signatureByID(producerSummary)
		result := corevalidate.Validate(program)
		peerByID := result.PeerSignatures()

		for _, function := range program.Functions {
			problemCode := classifyPublicationProblem(t, program, function.ID)
			if problemCode != "" && problemCode != "core.origin_omitted" {
				// D-07-33's narrowed-away classes: not this test's claim.
				continue
			}
			if problemCode == "core.origin_omitted" {
				comparedOmitted = true
			}
			producerCallable := producerByID[function.ID].Callable
			peerCallable := peerByID[function.ID].Callable
			if producerCallable != peerCallable {
				t.Fatalf("function %s (problem=%q): Callable diverges -- producer=%v peer=%v", function.ID, problemCode, producerCallable, peerCallable)
			}
		}
	}
	if !comparedOmitted {
		t.Fatal("expected at least one core.origin_omitted function to be compared (the corpus should include testdata/phase07/clean_but_unpublishable.lang and testdata/phase3's omitted fixtures)")
	}
}

// TestPeerDoesNotRederiveNarrowedClasses is 07-02 Task 2's Test 5 (D-07-33):
// names, by an explicit assertion rather than by silent agreement, that the
// peer does NOT independently re-derive core.origin_understated,
// core.origin_access_mismatch, or foreign-origin-omitted. It does this by
// mutating an honestly-checked program's PublicOrigin the same way
// originvalidate_test.go's own falsifiers do for these exact classes
// (check.go's honest producer can never construct these declarations
// itself), then showing the peer's Callable stays true -- it falsely agrees
// with a producer that WOULD refuse -- for the two origin-declared classes,
// and separately that foreign_origin_omitted.lang's peer Callable diverges
// from the producer specifically because the peer does not walk
// OpForeignCall.
func TestPeerDoesNotRederiveNarrowedClasses(t *testing.T) {
	t.Run("core.origin_understated: peer falsely agrees (Callable stays true)", func(t *testing.T) {
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", "public_view_understated.lang"))
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
		}
		program := checked.Program
		function := &program.Functions[0]
		if function.PublicOrigin == nil {
			t.Fatal("expected an honestly-declared PublicOrigin to mutate")
		}
		// Understate the declaration: an honest producer never constructs
		// this, so it is injected the same way TestOriginUnderstatedRejected
		// does, into a copy of the paths slice this function alone owns.
		function.PublicOrigin.Paths = []string{}

		if problems := originvalidate.PublishProblemsFor(*function); len(problems) != 1 || problems[0].Code != "core.origin_understated" {
			t.Fatalf("expected the mutated declaration to be refused as core.origin_understated, got %+v", problems)
		}
		result := corevalidate.Validate(program)
		peerSignature, ok := result.PeerSignatures()[function.ID]
		if !ok {
			t.Fatalf("function %s missing from peer signatures", function.ID)
		}
		if !peerSignature.Callable {
			t.Fatal("expected the peer to FALSELY agree (Callable == true) on a core.origin_understated function -- it does not independently re-derive this class (D-07-33)")
		}
	})

	t.Run("core.origin_access_mismatch: peer falsely agrees (Callable stays true)", func(t *testing.T) {
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", "public_view_impossible.lang"))
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
		}
		program := checked.Program
		function := &program.Functions[0]
		if function.PublicOrigin == nil {
			t.Fatal("expected an honestly-declared PublicOrigin to mutate")
		}
		function.PublicOrigin.Access = "exclusive"

		if problems := originvalidate.PublishProblemsFor(*function); len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
			t.Fatalf("expected the mutated declaration to be refused as core.origin_access_mismatch, got %+v", problems)
		}
		result := corevalidate.Validate(program)
		peerSignature, ok := result.PeerSignatures()[function.ID]
		if !ok {
			t.Fatalf("function %s missing from peer signatures", function.ID)
		}
		if !peerSignature.Callable {
			t.Fatal("expected the peer to FALSELY agree (Callable == true) on a core.origin_access_mismatch function -- it does not independently re-derive this class (D-07-33)")
		}
	})

	t.Run("foreign-origin-omitted: peer diverges from the refusing producer", func(t *testing.T) {
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_origin_omitted.lang"))
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
		}
		program := checked.Program
		function := program.Functions[0]
		problems := originvalidate.PublishProblemsFor(function)
		if len(problems) != 1 || problems[0].Code != "core.foreign_origin_omitted" {
			t.Fatalf("expected core.foreign_origin_omitted, got %+v", problems)
		}
		producerSummary, err := originvalidate.BuildInterface(program)
		if err != nil {
			t.Fatalf("BuildInterface: %v", err)
		}
		if producerSummary.Functions[0].Callable {
			t.Fatal("expected the producer to refuse Callable for a foreign-origin-omitted function")
		}
		result := corevalidate.Validate(program)
		peerSignature, ok := result.PeerSignatures()[function.ID]
		if !ok {
			t.Fatalf("function %s missing from peer signatures", function.ID)
		}
		if !peerSignature.Callable {
			t.Fatal("expected the peer's narrowed re-derivation to NOT catch foreign-origin-omitted -- Callable should stay true, diverging from the producer's false, demonstrating D-07-33's declared single-producer leak")
		}
	})
}
