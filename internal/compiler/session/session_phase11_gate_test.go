package session_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const (
	ControlPhase11ScanClean            = "control:phase11.scan_clean"
	ControlPhase11NonVacuousWouldCarry = "control:phase11.non_vacuous_would_carry"
	ControlPhase11StructuralFloors     = "control:phase11.structural_floors"
	ControlPhase11InterpreterAgreesO0  = "control:phase11.interpreter_agrees_o0"
	ControlPhase11ManifestConsistency  = "control:phase11.manifest_consistency"
)

// phase11GateReport is VerifyPhase11ZeroAttributeGate's fully observed,
// conjunct-by-conjunct result (D-11-19): every value 11-MIDPHASE-GATE.md's
// written adjudication must cite verbatim as a concrete number or string,
// never a characterization.
type phase11GateReport struct {
	FunctionCount           int
	CallEdgeCount           int
	WouldCarryCount         int
	WouldCarryFunctionIDs   []string
	ScannedArtifactCount    int
	BannedAttributesFound   []string
	ManifestEmptyAttributes bool
	InterpreterAgreesAtO0   bool
	Result                  protocol.Result
}

// Passed reports whether every one of the gate's conjuncts held.
func (r phase11GateReport) Passed() bool {
	return r.Result.Status == protocol.StatusPass
}

// phase11WouldCarryRestrict is Knower B's own independent re-derivation
// (D-11-14/D-11-15) of exactly the structural condition
// cgen.SelectsByPointerLowering decides, reading only
// core.Function/core.LinearBody -- session's own view of the PROGRAM,
// never cgen's derivation read a second time. This is session's own
// third, independent restatement of the identical structural fact (D-12:
// zero shared helpers between the derivations); any semantic drift
// between this and cgen's own selectsByPointerLowering is exactly what
// TestAliasFactAgreesWithByPointerSelection (check package) already
// polices for check's sibling derivation.
func phase11WouldCarryRestrict(function core.Function) bool {
	linear := function.Linear
	if function.Match != nil || function.PublicOrigin != nil || linear == nil || len(linear.Blocks) > 0 {
		return false
	}
	operations := linear.Operations
	if len(operations) == 0 {
		return false
	}
	first := operations[0]
	if first.Kind != core.OpBorrowExclusive || first.SourceID != function.Parameter.ID || first.TargetID == "" {
		return false
	}
	for _, operation := range operations[1:] {
		if operation.SourceID == function.Parameter.ID {
			return false
		}
	}
	current := first.TargetID
	terminatorIndex := -1
	for index := 1; index < len(operations); index++ {
		operation := operations[index]
		if operation.SourceID != current {
			return false
		}
		if operation.Kind == core.OpReturn {
			terminatorIndex = index
			break
		}
		if operation.TargetID == "" {
			return false
		}
		current = operation.TargetID
	}
	return terminatorIndex == len(operations)-1
}

// VerifyPhase11ZeroAttributeGate is Phase 11's mandatory mid-phase gate
// (D-11-14): a four-part conjunction with two genuinely independent
// knowers -- cgen.ScanForBannedAttributes reading OUTPUT BYTES (Knower A)
// and phase11WouldCarryRestrict's own re-derivation reading the PROGRAM
// (Knower B) -- plus the structural non-vacuity floors (function count,
// call-edge count) and a real interpreter/-O0 agreement run (RunNative,
// which also exercises -O3).
//
// The by-pointer route proves an exact public refusal before it can load
// provenance-checked frozen C. The N=0 control has no by-pointer body and
// continues through admitted live emission.
func phase11ExpectedRefusal(fixture string) string {
	switch fixture {
	case "testdata/phase11/multi_function_gate_corpus.schway":
		return `function "s1:phase11.multi_function_gate_corpus:fn:touch": by-pointer bodies are not supported by whole-program native emission this phase`
	case "testdata/phase16/historical/phase11_gate_n_two.fixture":
		return `function "s1:phase11.gate_n_two:fn:touchTwo": by-pointer bodies are not supported by whole-program native emission this phase`
	default:
		return ""
	}
}

func phase11RefusalFirstFrozen(t *testing.T, program core.Program, fixture string) (string, error) {
	t.Helper()
	_, refusal := cgen.EmitNative(program)
	want := phase11ExpectedRefusal(fixture)
	if want == "" || refusal == nil || refusal.Error() != want {
		return "", fmt.Errorf("public refusal changed for %q: want %q, got %v", fixture, want, refusal)
	}
	return phase16FileFrozenEvidenceC(t, program, fixture)
}

func verifyPhase11ZeroAttributeGate(t *testing.T, ctx context.Context, source []byte, fixture string, runner native.Runner) (phase11GateReport, error) {
	report := phase11GateReport{}
	result := protocol.New("verify", protocol.StatusPass)

	addLane := func(id, status string, controls []string, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: protocol.LaneSchema1, ID: id, Status: status,
			Controls: append([]string{}, controls...), RecomputedWork: 1,
			ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
		})
	}
	markFail := func(status string) {
		if result.Status == protocol.StatusPass {
			result.Status = status
		}
	}

	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		return report, fmt.Errorf("phase11 gate: corpus failed to check: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return report, fmt.Errorf("phase11 gate: corevalidate rejected corpus: %+v", validated.Problems)
	}
	program := validated.Program()

	// Structural floors + Knower B, both derived from the PROGRAM alone,
	// never from cgen's own emission.
	report.FunctionCount = len(program.Functions)
	callEdges := 0
	var wouldCarry []core.Function
	for _, function := range program.Functions {
		if function.Linear != nil {
			for _, operation := range function.Linear.Operations {
				if operation.Kind == core.OpCall {
					callEdges++
				}
			}
		}
		if phase11WouldCarryRestrict(function) {
			wouldCarry = append(wouldCarry, function)
			report.WouldCarryFunctionIDs = append(report.WouldCarryFunctionIDs, function.ID)
		}
	}
	report.CallEdgeCount = callEdges
	report.WouldCarryCount = len(wouldCarry)

	floorsStarted := time.Now()
	if report.FunctionCount >= 2 && report.CallEdgeCount >= 1 {
		addLane("lane:phase11-structural-floors", protocol.StatusPass, []string{ControlPhase11StructuralFloors}, floorsStarted)
	} else {
		addLane("lane:phase11-structural-floors", protocol.StatusInvalid, nil, floorsStarted)
		markFail(protocol.StatusInvalid)
	}

	vacuityStarted := time.Now()
	if report.WouldCarryCount >= 1 {
		addLane("lane:phase11-non-vacuous-would-carry", protocol.StatusPass, []string{ControlPhase11NonVacuousWouldCarry}, vacuityStarted)
	} else {
		addLane("lane:phase11-non-vacuous-would-carry", protocol.StatusInvalid, nil, vacuityStarted)
		markFail(protocol.StatusInvalid)
	}

	// Knower A scans only the exact artifact that the current route is allowed
	// to supply. By-pointer programs first prove the current public refusal;
	// only after that exact witness passes may the test-only frozen loader read C.
	scanStarted := time.Now()
	var artifacts []string
	if report.WouldCarryCount > 0 {
		frozen, err := phase11RefusalFirstFrozen(t, program, fixture)
		if err != nil {
			return report, fmt.Errorf("phase11 gate: refusal-first frozen evidence: %w", err)
		}
		artifacts = append(artifacts, frozen)
		entry, entryErr := callgraph.EntryFunction(program)
		if entryErr != nil {
			return report, fmt.Errorf("phase11 gate: resolve entry: %w", entryErr)
		}
		input := phase11EntryInput(t, entry.Parameter.Type)
		engines := phase11RunFourTiersWithSupplier(t, ctx, program, entry.Name, input, func(core.Program) (string, error) {
			return frozen, nil
		}, "refusal-gated frozen Phase 11 evidence")
		if err := session.Phase5CompareProgramEngines(fixture, program, engines); err != nil {
			return report, fmt.Errorf("phase11 gate: four-tier frozen-evidence disagreement: %w", err)
		}
		report.InterpreterAgreesAtO0 = true
	} else {
		// N=0 has no by-pointer body, so it remains an admitted live-emission
		// falsification control.
		artifact, emitErr := cgen.EmitNative(program)
		if emitErr != nil {
			return report, fmt.Errorf("phase11 gate: emitting admitted N=0 artifact: %w", emitErr)
		}
		artifacts = append(artifacts, artifact)
		_, diags, runErr := session.RunNative(ctx, source, runner)
		if runErr != nil {
			addLane("lane:phase11-interpreter-agrees-o0", protocol.StatusOperational, nil, scanStarted)
			markFail(protocol.StatusOperational)
		} else if len(diags) != 0 {
			addLane("lane:phase11-interpreter-agrees-o0", protocol.StatusInvalid, nil, scanStarted)
			markFail(protocol.StatusInvalid)
		} else {
			report.InterpreterAgreesAtO0 = true
		}
	}
	report.ScannedArtifactCount = len(artifacts)
	report.BannedAttributesFound = cgen.ScanForBannedAttributes(artifacts...)
	if len(report.BannedAttributesFound) == 0 {
		addLane("lane:phase11-scan-clean", protocol.StatusPass, []string{ControlPhase11ScanClean}, scanStarted)
	} else {
		addLane("lane:phase11-scan-clean", protocol.StatusInvalid, nil, scanStarted)
		markFail(protocol.StatusInvalid)
	}

	// This historical manifest check records the empty attribute set; it is
	// not presented as a public emission result for refused by-pointer code.
	manifestStarted := time.Now()
	report.ManifestEmptyAttributes = len(report.BannedAttributesFound) == 0
	if report.ManifestEmptyAttributes {
		addLane("lane:phase11-manifest-consistency", protocol.StatusPass, []string{ControlPhase11ManifestConsistency}, manifestStarted)
	} else {
		addLane("lane:phase11-manifest-consistency", protocol.StatusInvalid, nil, manifestStarted)
		markFail(protocol.StatusInvalid)
	}
	if report.InterpreterAgreesAtO0 {
		addLane("lane:phase11-interpreter-agrees-o0", protocol.StatusPass, []string{ControlPhase11InterpreterAgreesO0}, scanStarted)
	}
	report.Result = result
	return report, nil
}

// phase11GateCorpusSource reads the committed mid-phase gate corpus.
func phase11GateCorpusSource(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase11", "multi_function_gate_corpus.schway"))
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

func phase11TwoWouldCarryCorpusSource(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase16", "historical", "phase11_gate_n_two.fixture"))
	if err != nil {
		t.Fatal(err)
	}
	return source
}

// TestPhase11ZeroAttributeGate is D-11-14's own full-conjunction pass: on
// the committed gate corpus, every one of the gate's four conjuncts holds,
// asserted individually (not merely the final boolean).
func TestPhase11ZeroAttributeGate(t *testing.T) {
	report, err := verifyPhase11ZeroAttributeGate(t, context.Background(), phase11GateCorpusSource(t), "testdata/phase11/multi_function_gate_corpus.schway", native.DefaultRunner())
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
	if report.ScannedArtifactCount != 1 {
		t.Fatalf("expected exactly one refusal-gated frozen artifact, got %d", report.ScannedArtifactCount)
	}
}

// TestPhase11GateIsNonVacuous asserts the corpus's own non-vacuity floors
// explicitly, each with a failure message naming which floor was not met.
func TestPhase11GateIsNonVacuous(t *testing.T) {
	report, err := verifyPhase11ZeroAttributeGate(t, context.Background(), phase11GateCorpusSource(t), "testdata/phase11/multi_function_gate_corpus.schway", native.DefaultRunner())
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
	report, err := verifyPhase11ZeroAttributeGate(t, context.Background(), []byte(phase11ZeroWouldCarryCorpus), "", native.DefaultRunner())
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
	report, err := verifyPhase11ZeroAttributeGate(t, context.Background(), phase11TwoWouldCarryCorpusSource(t), "testdata/phase16/historical/phase11_gate_n_two.fixture", native.DefaultRunner())
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

// TestPhase11GateMutationKill is D-11-18's anti-vacuity proof: a seeded
// restrict mutation in frozen evidence must turn the exact same scanner red.
func TestPhase11GateMutationKill(t *testing.T) {
	program := session.Check(phase11GateCorpusSource(t)).Program
	frozen, err := phase11RefusalFirstFrozen(t, program, "testdata/phase11/multi_function_gate_corpus.schway")
	if err != nil {
		t.Fatal(err)
	}
	mutated := phase11SeedRestrict(t, frozen)
	if found := cgen.ScanForBannedAttributes(mutated); len(found) == 0 {
		t.Fatal("expected seeded restrict mutation to be detected by banned-attribute scan")
	}
	if diff := phase11DiffOutsideBannedTokens(frozen, mutated); len(diff) != 0 {
		t.Fatalf("seeded mutation changed tokens outside banned attributes: %v", diff)
	}
}

func phase11SeedRestrict(t *testing.T, source string) string {
	t.Helper()
	lines := strings.Split(source, "\n")
	matched := 0
	for index, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "static ") || !strings.Contains(line, "invocation_index") || !strings.HasSuffix(strings.TrimSpace(line), "{") {
			continue
		}
		open := strings.IndexByte(line, '(')
		close := strings.IndexByte(line, ')')
		if open < 0 || close <= open {
			continue
		}
		firstArgument := strings.TrimSpace(strings.SplitN(line[open+1:close], ",", 2)[0])
		parts := strings.Fields(firstArgument)
		if len(parts) != 2 {
			continue
		}
		old := parts[0] + " " + parts[1]
		replacement := parts[0] + " restrict " + parts[1]
		lines[index] = strings.Replace(line, old, replacement, 1)
		matched++
	}
	if matched != 1 {
		t.Fatalf("expected one frozen C function parameter seam, found %d", matched)
	}
	return strings.Join(lines, "\n")
}

func TestPhase11ByPointerRefusalFirstEvidence(t *testing.T) {
	for _, fixture := range []string{
		"testdata/phase11/multi_function_gate_corpus.schway",
		"testdata/phase16/historical/phase11_gate_n_two.fixture",
	} {
		source, err := os.ReadFile(testsupport.ProjectPath(fixture))
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("%s: unexpected diagnostics: %+v", fixture, checked.Diagnostics)
		}
		frozen, err := phase11RefusalFirstFrozen(t, checked.Program, fixture)
		if err != nil {
			t.Fatalf("%s: %v", fixture, err)
		}
		if frozen == "" {
			t.Fatalf("%s: refusal-first loader returned empty frozen C", fixture)
		}
	}
}

// TestPhase11SuppressionIsDiffLocal is D-11-16's own diff-locality proof:
// comparing frozen attribute-free evidence with a single seeded `restrict`
// token at the exact function-declaration seam proves suppression locality.
func TestPhase11SuppressionIsDiffLocal(t *testing.T) {
	checked := session.Check(phase11GateCorpusSource(t))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	suppressed, err := phase11RefusalFirstFrozen(t, checked.Program, "testdata/phase11/multi_function_gate_corpus.schway")
	if err != nil {
		t.Fatal(err)
	}
	justified := phase11SeedRestrict(t, suppressed)

	if suppressed == justified {
		t.Fatal("expected the seeded restrict mutation to change the frozen bytes")
	}
	if strings.Contains(suppressed, "restrict") {
		t.Fatal("expected frozen attribute-free C to contain no restrict token")
	}
	if !strings.Contains(justified, "restrict") {
		t.Fatal("expected the seeded frozen C to contain the restrict token")
	}
	if found := cgen.ScanForBannedAttributes(justified); len(found) == 0 {
		t.Fatal("expected the banned-attribute scan to detect seeded restrict")
	}

	suppressedNoreturn := strings.Contains(suppressed, cgen.NoreturnExemption)
	justifiedNoreturn := strings.Contains(justified, cgen.NoreturnExemption)
	if suppressedNoreturn != justifiedNoreturn {
		t.Fatalf("expected NoreturnExemption presence to be symmetric across frozen and seeded bytes, frozen=%v mutated=%v", suppressedNoreturn, justifiedNoreturn)
	}

	diffTokens := phase11DiffOutsideBannedTokens(suppressed, justified)
	if len(diffTokens) != 0 {
		t.Fatalf("expected frozen and seeded C to differ ONLY in banned-attribute bytes, found non-banned divergent tokens: %v", diffTokens)
	}
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
