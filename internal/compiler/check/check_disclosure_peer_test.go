package check

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// ---------------------------------------------------------------------
// Phase 09 Plan 06, Task 3(c) (D-09-30): the cross-peer identity
// assertion -- the independence instrument Phase 08's own gate anticipated
// when it sent both disclosure questions (D-08-26's refusal half and its
// accepted-program half) here together.
//
// What this constrains: both peers may read ONLY the declared,
// body-blind contract fields SEM-05 permits (ParameterContract.Mode /
// ReturnContract.Mode) -- a CONTRACT-CONFORMANCE check, proven on
// OUTPUTS, exactly as structuralFieldsEqual and ClosureDigest
// byte-equality already prove agreement without constraining
// implementation.
//
// What this deliberately does NOT constrain: worklist-vs-closure,
// traversal order, memoization, or any other internal mechanism. check's
// own set is read from originvalidate's independently-built interface (a
// stored lookup, buildCallSignatureTable); corevalidate's own set is
// independently RE-DERIVED from each function's own declared shape
// (derivePeerSignature). The two mechanisms differ completely; only the
// DISCLOSED FIELD NAME VOCABULARY must agree.
//
// Why this is strictly STRONGER than verdict agreement alone: two
// independent derivations can agree on accept/reject while one secretly
// reads a THIRD field -- a body fact that would break SEM-05's
// body-blindness, or Return.Paths instead of Return.Mode. Verdict
// comparison (session_peer_gate_test.go's peerDivergenceExpected,
// syntheticShapeDivergenceExpected) never catches that kind of latent
// divergence; a field-set mismatch catches it immediately, permanently,
// at zero runtime cost, because the set is a compile-time constant
// (D-09-29) checked once here rather than re-derived per program.
// ---------------------------------------------------------------------

// collectCheckConsultedFields installs interproceduralConsultObserved
// (this package's own D-08-26 seam) around the given closure, and returns
// the recorded field-name set (never per-callee attribution -- only the
// flat set this cross-peer comparison needs), with the seam restored via
// a deferred call before returning.
func collectCheckConsultedFields(run func()) []string {
	recorded := map[string]bool{}
	previous := interproceduralConsultObserved
	interproceduralConsultObserved = func(_, field string) { recorded[field] = true }
	defer func() { interproceduralConsultObserved = previous }()

	run()

	fields := make([]string, 0, len(recorded))
	for field := range recorded {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return fields
}

// assertIdenticalFieldSets asserts a and b are identical as SORTED sets --
// D-09-30's own comparison shape, applied here rather than a byte-for-byte
// slice equality so ordering differences between the two independently-
// written recorders can never cause a false failure.
func assertIdenticalFieldSets(t *testing.T, label string, checkFields, peerFields []string) {
	t.Helper()
	sortedCheck := append([]string(nil), checkFields...)
	sortedPeer := append([]string(nil), peerFields...)
	sort.Strings(sortedCheck)
	sort.Strings(sortedPeer)
	if len(sortedCheck) == 0 {
		t.Fatalf("%s: check's own consulted-field set is empty -- nothing to compare", label)
	}
	if len(sortedPeer) == 0 {
		t.Fatalf("%s: corevalidate's own consulted-field set is empty -- nothing to compare", label)
	}
	if len(sortedCheck) != len(sortedPeer) {
		t.Fatalf("%s: check consulted %v, corevalidate consulted %v -- sets differ in size", label, sortedCheck, sortedPeer)
	}
	for i := range sortedCheck {
		if sortedCheck[i] != sortedPeer[i] {
			t.Fatalf("%s: check consulted %v, corevalidate consulted %v -- sets are not identical", label, sortedCheck, sortedPeer)
		}
	}
}

// TestPeerAndCheckDisclosedFieldSetsAreIdentical drives more than one
// program shape (a real testdata/phase08 `.lang` fixture AND a synthetic
// call-graph shape, so identity is never an artifact of a single fixture)
// through BOTH peers and asserts their consulted-field sets are identical
// sorted sets.
func TestPeerAndCheckDisclosedFieldSetsAreIdentical(t *testing.T) {
	t.Run("real .lang fixture", func(t *testing.T) {
		source, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "phase08", "relay_depth2_accept.lang"))
		if err != nil {
			t.Fatalf("read fixture: %v", err)
		}
		program := mustParseProgram(t, source)

		var checkedProgram core.Program
		checkFields := collectCheckConsultedFields(func() {
			result := Program(program)
			if len(result.Diagnostics) != 0 {
				t.Fatalf("fixture failed check.Program: %+v", result.Diagnostics)
			}
			checkedProgram = result.Program
		})

		validated := corevalidate.Validate(checkedProgram)
		peerFields := validated.PeerConsultedFields()

		assertIdenticalFieldSets(t, "relay_depth2_accept.lang", checkFields, peerFields)
	})

	t.Run("synthetic diamond call-graph shape", func(t *testing.T) {
		program, err := testsupport.GenerateCallGraphCorpus("diamond", 16)
		if err != nil {
			t.Fatalf("testsupport.GenerateCallGraphCorpus(diamond, 16): %v", err)
		}
		// check's own consult runs directly over the bare core.Program
		// (D-09-50: check.Program only accepts ast.Program, so this
		// differential drives buildInterproceduralSummaries directly,
		// exactly like check_peer_shape_differential_test.go's own
		// checkSyntheticProgramVerdict).
		table, err := buildCallSignatureTable(program)
		if err != nil {
			t.Fatalf("buildCallSignatureTable: %v", err)
		}
		checkFields := collectCheckConsultedFields(func() {
			buildInterproceduralSummaries(program, table)
		})

		completed := completeSyntheticProgramForCorevalidate(program)
		completed.Schema = core.Schema1
		completed.Module = "diskcheck"
		completed.ModuleID = "s1:diskcheck:module:diskcheck"
		validated := corevalidate.Validate(completed)
		if !validated.Valid {
			t.Fatalf("corevalidate.Validate reported invalid on the completed synthetic program: %v", validated.Problems)
		}
		peerFields := validated.PeerConsultedFields()

		assertIdenticalFieldSets(t, "synthetic diamond shape", checkFields, peerFields)
	})
}
