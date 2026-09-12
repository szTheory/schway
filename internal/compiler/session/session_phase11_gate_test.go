package session_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// phase11GateCorpusSource reads the committed mid-phase gate corpus.
func phase11GateCorpusSource(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase11", "multi_function_gate_corpus.lang"))
	if err != nil {
		t.Fatal(err)
	}
	return source
}

// phase11ZeroWouldCarryCorpus is TestPhase11GateFailsAtNZero's own N=0
// witness: two functions, one call edge, and NEITHER function's body
// selects the by-pointer structural shape (both operate on a Byte
// parameter with no borrow chain at all) -- a deliberately vacuous
// corpus, engineered to make the gate FAIL at the N=0 boundary.
const phase11ZeroWouldCarryCorpus = `
module phase11.gate_n_zero

export {
  fn main
}

fn identity(value: Byte) -> Byte {
  value
}

fn main(value: Byte) -> Byte {
  let result = identity(value)
  result
}
`

// phase11TwoWouldCarryCorpus is TestPhase11GateCountsAdjacentWouldCarryFunctions's
// own N=2 witness: `touch` and `touchTwo` are BOTH would-have-carried-
// restrict functions (the identical exclusive-borrow-then-reborrow-to-
// terminator shape, on two distinct Buffer parameters), alongside the
// same main/identity call edge the N=1 corpus uses.
const phase11TwoWouldCarryCorpus = `
module phase11.gate_n_two

export {
  fn main
}

fn identity(value: Byte) -> Byte {
  value
}

fn main(value: Byte) -> Byte {
  let result = identity(value)
  result
}

fn touch(buffer: Buffer) -> Buffer {
  let first = borrow mut buffer
  let second = borrow first
  let third = borrow second
  third
}

fn touchTwo(buffer: Buffer) -> Buffer {
  let first = borrow mut buffer
  let second = borrow first
  let third = borrow second
  third
}
`

// TestPhase11ZeroAttributeGate is D-11-14's own full-conjunction pass: on
// the committed gate corpus, every one of the gate's four conjuncts holds,
// asserted individually (not merely the final boolean).
func TestPhase11ZeroAttributeGate(t *testing.T) {
	report, err := session.VerifyPhase11ZeroAttributeGate(context.Background(), phase11GateCorpusSource(t), native.DefaultRunner(), cgen.AttributesSuppressed)
	if err != nil {
		t.Fatalf("VerifyPhase11ZeroAttributeGate: %v", err)
	}
	if !report.Passed() {
		t.Fatalf("expected the gate to pass, got status %q with lanes: %+v", report.Result.Status, report.Result.Lanes)
	}
	if report.FunctionCount != 3 {
		t.Fatalf("expected function count 3, got %d", report.FunctionCount)
	}
	if report.CallEdgeCount != 1 {
		t.Fatalf("expected call-edge count 1, got %d", report.CallEdgeCount)
	}
	if report.WouldCarryCount != 1 {
		t.Fatalf("expected would-carry count 1, got %d (%v)", report.WouldCarryCount, report.WouldCarryFunctionIDs)
	}
	if len(report.BannedAttributesFound) != 0 {
		t.Fatalf("expected no banned attributes found, got %v", report.BannedAttributesFound)
	}
	if !report.ManifestEmptyAttributes {
		t.Fatal("expected the manifest consistency check to report empty")
	}
	if !report.InterpreterAgreesAtO0 {
		t.Fatal("expected the interpreter to agree with -O0")
	}
	if report.ScannedArtifactCount != 2 {
		t.Fatalf("expected 2 scanned artifacts (whole-program TU + 1 counterfactual), got %d", report.ScannedArtifactCount)
	}
}

// TestPhase11GateIsNonVacuous asserts the corpus's own non-vacuity floors
// explicitly, each with a failure message naming which floor was not met.
func TestPhase11GateIsNonVacuous(t *testing.T) {
	report, err := session.VerifyPhase11ZeroAttributeGate(context.Background(), phase11GateCorpusSource(t), native.DefaultRunner(), cgen.AttributesSuppressed)
	if err != nil {
		t.Fatalf("VerifyPhase11ZeroAttributeGate: %v", err)
	}
	if report.FunctionCount < 2 {
		t.Fatalf("non-vacuity floor not met: function count %d < 2", report.FunctionCount)
	}
	if report.CallEdgeCount < 1 {
		t.Fatalf("non-vacuity floor not met: call-edge count %d < 1", report.CallEdgeCount)
	}
	if report.WouldCarryCount < 1 {
		t.Fatalf("non-vacuity floor not met: would-carry count %d < 1", report.WouldCarryCount)
	}
}

// TestPhase11GateFailsAtNZero is D-11-14's own falsifying half: on a
// corpus where NO function would have carried restrict, the gate FAILS --
// asserted by a real fail, not by the lane being skipped. Both sides of
// the N threshold are asserted (this test is N=0's; TestPhase11ZeroAttributeGate
// is N=1's), since passing only at N=1 proves nothing without also
// failing at N=0.
func TestPhase11GateFailsAtNZero(t *testing.T) {
	report, err := session.VerifyPhase11ZeroAttributeGate(context.Background(), []byte(phase11ZeroWouldCarryCorpus), native.DefaultRunner(), cgen.AttributesSuppressed)
	if err != nil {
		t.Fatalf("VerifyPhase11ZeroAttributeGate: %v", err)
	}
	if report.WouldCarryCount != 0 {
		t.Fatalf("expected this corpus to have zero would-carry functions, got %d", report.WouldCarryCount)
	}
	if report.Passed() {
		t.Fatal("expected the gate to FAIL at N=0, got a pass")
	}
}

// TestPhase11GateCountsAdjacentWouldCarryFunctions: when TWO functions in
// one program would each have carried restrict, N == 2 and the emitted
// attribute set is still exactly empty -- adjacency does not merge or
// double-count the would-have-carried population.
func TestPhase11GateCountsAdjacentWouldCarryFunctions(t *testing.T) {
	report, err := session.VerifyPhase11ZeroAttributeGate(context.Background(), []byte(phase11TwoWouldCarryCorpus), native.DefaultRunner(), cgen.AttributesSuppressed)
	if err != nil {
		t.Fatalf("VerifyPhase11ZeroAttributeGate: %v", err)
	}
	if report.WouldCarryCount != 2 {
		t.Fatalf("expected would-carry count 2, got %d (%v)", report.WouldCarryCount, report.WouldCarryFunctionIDs)
	}
	if len(report.BannedAttributesFound) != 0 {
		t.Fatalf("expected the emitted attribute set to remain exactly empty, found: %v", report.BannedAttributesFound)
	}
	if !report.Passed() {
		t.Fatalf("expected the gate to pass, got status %q with lanes: %+v", report.Result.Status, report.Result.Lanes)
	}
}

// TestPhase11GateMutationKill is D-11-18's own anti-vacuity proof:
// re-running the SAME gate corpus through the SAME lane under the
// JUSTIFIED profile makes the lane go RED. A gate green under both
// profiles would be measuring nothing.
func TestPhase11GateMutationKill(t *testing.T) {
	report, err := session.VerifyPhase11ZeroAttributeGate(context.Background(), phase11GateCorpusSource(t), native.DefaultRunner(), cgen.AttributesJustified)
	if err != nil {
		t.Fatalf("VerifyPhase11ZeroAttributeGate: %v", err)
	}
	if report.Passed() {
		t.Fatal("expected the lane to go RED under the justified profile, got a pass")
	}
	if len(report.BannedAttributesFound) == 0 {
		t.Fatal("expected the justified profile to surface a banned attribute in the counterfactual artifact")
	}
}

// TestPhase11SuppressionIsDiffLocal is D-11-16's own diff-locality proof:
// emitting the corpus's own would-carry function under both profiles
// produces output that differs ONLY in bytes belonging to
// BannedOptimizerAttributes; NoreturnExemption (`_Noreturn`) -- not an
// alias promise -- stays absent from both, since this fixture's function
// never carries a defect terminator, so its absence is symmetric rather
// than a hidden divergence.
func TestPhase11SuppressionIsDiffLocal(t *testing.T) {
	checked := session.Check(phase11GateCorpusSource(t))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	var wouldCarryFunctionID string
	for _, function := range checked.Program.Functions {
		if function.Name == "touch" {
			wouldCarryFunctionID = function.ID
		}
	}
	if wouldCarryFunctionID == "" {
		t.Fatal("expected the gate corpus to declare a function named touch")
	}

	suppressed := phase11EmitSingleFunction(t, checked.Program, wouldCarryFunctionID, cgen.AttributesSuppressed)
	justified := phase11EmitSingleFunction(t, checked.Program, wouldCarryFunctionID, cgen.AttributesJustified)

	if suppressed == justified {
		t.Fatal("expected the two profiles to produce different output for a would-carry function")
	}
	if strings.Contains(suppressed, "restrict") {
		t.Fatal("expected the suppressed profile's output to contain no restrict token")
	}
	if !strings.Contains(justified, "restrict") {
		t.Fatal("expected the justified profile's output to contain the restrict token")
	}

	suppressedNoreturn := strings.Contains(suppressed, cgen.NoreturnExemption)
	justifiedNoreturn := strings.Contains(justified, cgen.NoreturnExemption)
	if suppressedNoreturn != justifiedNoreturn {
		t.Fatalf("expected NoreturnExemption presence to be symmetric across profiles, suppressed=%v justified=%v", suppressedNoreturn, justifiedNoreturn)
	}

	diffTokens := phase11DiffOutsideBannedTokens(suppressed, justified)
	if len(diffTokens) != 0 {
		t.Fatalf("expected the two profiles' output to differ ONLY in bytes belonging to BannedOptimizerAttributes, found non-banned divergent tokens: %v", diffTokens)
	}
}

// phase11EmitSingleFunction extracts functionID from program and emits it
// as a standalone single-function core.Program under profile, returning
// the generated C.
func phase11EmitSingleFunction(t *testing.T, program core.Program, functionID string, profile cgen.AttributeSuppressionProfile) string {
	t.Helper()
	var target core.Function
	found := false
	for _, function := range program.Functions {
		if function.ID == functionID {
			target = function
			found = true
		}
	}
	if !found {
		t.Fatalf("function %q not found in program", functionID)
	}
	single := core.Program{Schema: program.Schema, Module: program.Module, ModuleID: program.ModuleID, Functions: []core.Function{target}}
	restore := cgen.SetAttributeSuppressionProfileForTest(profile)
	defer restore()
	generated, err := cgen.EmitNative(single)
	if err != nil {
		t.Fatalf("EmitNative (single-function %q, profile %v): %v", functionID, profile, err)
	}
	return generated
}

// phase11DiffOutsideBannedTokens line-diffs a and b, stripping every
// cgen.BannedOptimizerAttributes token (and the whitespace it leaves
// behind) from each divergent line before re-comparing; a line still
// unequal after stripping is a structural divergence, not a banned-token-
// only one. Returns each such offending "lineA || lineB" pair.
func phase11DiffOutsideBannedTokens(a, b string) []string {
	linesA := strings.Split(a, "\n")
	linesB := strings.Split(b, "\n")
	max := len(linesA)
	if len(linesB) > max {
		max = len(linesB)
	}
	var offending []string
	for index := 0; index < max; index++ {
		var lineA, lineB string
		if index < len(linesA) {
			lineA = linesA[index]
		}
		if index < len(linesB) {
			lineB = linesB[index]
		}
		if lineA == lineB {
			continue
		}
		strippedA, strippedB := lineA, lineB
		for _, token := range cgen.BannedOptimizerAttributes {
			strippedA = strings.ReplaceAll(strippedA, token, "")
			strippedB = strings.ReplaceAll(strippedB, token, "")
		}
		strippedA = strings.Join(strings.Fields(strippedA), "")
		strippedB = strings.Join(strings.Fields(strippedB), "")
		if strippedA != strippedB {
			offending = append(offending, lineA+" || "+lineB)
		}
	}
	return offending
}
