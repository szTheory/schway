package cfg

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"ai-lang/ownership-kernel-workbench/ownership"
)

func TestFixturesAgreeWithBoundedPathOracle(t *testing.T) {
	file, err := LoadFixtures("../fixtures/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range file.Fixtures {
		fixture := fixture
		t.Run(fixture.Program.ID, func(t *testing.T) {
			comparison := Compare(fixture.Program, Options{}, 3)
			if !comparison.Agreement {
				bytes, _ := json.Marshal(comparison)
				t.Fatalf("analyzer/path disagreement: %s", bytes)
			}
			if comparison.AnalyzerValid != fixture.ExpectValid {
				t.Fatalf("valid=%v, want %v", comparison.AnalyzerValid, fixture.ExpectValid)
			}
			if !fixture.ExpectValid && !comparisonHasCode(comparison, fixture.ExpectCode) {
				t.Fatalf("expected diagnostic %q: %+v", fixture.ExpectCode, comparison.Paths)
			}
		})
	}
}

func TestBranchSpecificEndpointIsPlacedOnUnusedEdge(t *testing.T) {
	program := branchSpecificProgram()
	analysis := Analyze(program, Options{})
	want := "edge:entry:mutates:view"
	if !contains(EndpointShape(analysis), want) {
		t.Fatalf("endpoints=%v, want %s", EndpointShape(analysis), want)
	}
	if !contains(EndpointShape(analysis), "point:uses:0:view") {
		t.Fatalf("endpoints=%v, want point endpoint after final use", EndpointShape(analysis))
	}
}

func TestLoopExitGetsEdgeEndpoint(t *testing.T) {
	file, err := LoadFixtures("../fixtures/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var program Program
	for _, fixture := range file.Fixtures {
		if fixture.Program.ID == "CFG-LOOP-001" {
			program = fixture.Program
		}
	}
	analysis := Analyze(program, Options{})
	if !contains(EndpointShape(analysis), "edge:header:exit:view") {
		t.Fatalf("endpoints=%v, want loop-exit endpoint", EndpointShape(analysis))
	}
	if containsPrefix(EndpointShape(analysis), "point:body:") {
		t.Fatalf("loop-body loan must remain live across the back edge: %v", EndpointShape(analysis))
	}
}

func TestGeneratedDiamondProductsAgree(t *testing.T) {
	programs := GenerateDiamondPrograms()
	if len(programs) != 100 {
		t.Fatalf("generated=%d, want 100", len(programs))
	}
	for index, program := range programs {
		comparison := Compare(program, Options{}, 3)
		if !comparison.Agreement {
			bytes, _ := json.Marshal(comparison)
			t.Fatalf("mismatch at product %d: %s", index, bytes)
		}
	}
}

func TestQuickRandomizedRepresentationChangesPreserveAgreement(t *testing.T) {
	programs := GenerateDiamondPrograms()
	property := func(raw uint16, reverseSuccessors bool, reverseBlocks bool) bool {
		program := programs[int(raw)%len(programs)]
		if reverseSuccessors {
			for index := range program.Blocks {
				reverseStrings(program.Blocks[index].Successors)
			}
		}
		if reverseBlocks {
			for left, right := 0, len(program.Blocks)-1; left < right; left, right = left+1, right-1 {
				program.Blocks[left], program.Blocks[right] = program.Blocks[right], program.Blocks[left]
			}
		}
		comparison := Compare(program, Options{}, 3)
		return comparison.Agreement
	}
	config := &quick.Config{MaxCount: 500, Rand: rand.New(rand.NewSource(0xA11A57))}
	if err := quick.Check(property, config); err != nil {
		t.Fatal(err)
	}
}

func TestFaultInjectionFindsEdgeSpecificCounterexample(t *testing.T) {
	result := FindMismatch(Options{OmitEdgeEnds: true})
	if result.Mismatch == nil || result.MinimalInput == nil {
		t.Fatal("edge fault injection produced no mismatch")
	}
	if result.Programs > 100 {
		t.Fatalf("counterexample search examined %d programs", result.Programs)
	}
	if !hasDisagreeingPath(*result.Mismatch) {
		t.Fatalf("mismatch has no disagreeing path: %+v", result.Mismatch)
	}
}

func TestSuccessorOrderDoesNotChangeEndpointShapeOrValidity(t *testing.T) {
	program := branchSpecificProgram()
	left := Compare(program, Options{}, 3)
	program.Blocks[0].Successors[0], program.Blocks[0].Successors[1] = program.Blocks[0].Successors[1], program.Blocks[0].Successors[0]
	right := Compare(program, Options{}, 3)
	if !Equivalent(left, right) {
		t.Fatalf("successor order changed semantics:\nleft=%+v\nright=%+v", left, right)
	}
}

func TestLoanAlphaRenamePreservesEndpointStructure(t *testing.T) {
	program := branchSpecificProgram()
	before := Analyze(program, Options{})
	for blockIndex := range program.Blocks {
		for operationIndex := range program.Blocks[blockIndex].Operations {
			operation := &program.Blocks[blockIndex].Operations[operationIndex]
			if operation.Loan == "view" {
				operation.Loan = "renamed"
			}
		}
	}
	after := Analyze(program, Options{})
	stripLoan := func(values []string) []string {
		result := make([]string, len(values))
		for index, value := range values {
			result[index] = strings.ReplaceAll(value, "view", "loan")
			result[index] = strings.ReplaceAll(result[index], "renamed", "loan")
		}
		return result
	}
	if !reflect.DeepEqual(stripLoan(EndpointShape(before)), stripLoan(EndpointShape(after))) {
		t.Fatalf("alpha rename changed endpoint structure: %v vs %v", EndpointShape(before), EndpointShape(after))
	}
}

func TestLargeCFGConvergesWithoutPathExpansion(t *testing.T) {
	program := LargeDiamondChain(500)
	analysis := Analyze(program, Options{})
	if !analysis.Valid {
		t.Fatal(analysis.Diagnostic)
	}
	if analysis.Blocks != 1501 {
		t.Fatalf("blocks=%d, want 1501", analysis.Blocks)
	}
	if analysis.TransferEvaluations > analysis.Blocks*4 {
		t.Fatalf("transfer evaluations=%d exceed bounded worklist expectation for %d blocks", analysis.TransferEvaluations, analysis.Blocks)
	}
	if !contains(EndpointShape(analysis), "point:b0500-join:0:view") {
		t.Fatalf("final loan use endpoint missing; final endpoints=%v", EndpointShape(analysis))
	}
}

func TestStableAnalysisJSON(t *testing.T) {
	program := branchSpecificProgram()
	first, err := json.Marshal(Analyze(program, Options{}))
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(Analyze(program, Options{}))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("analysis JSON changed:\n%s\n%s", first, second)
	}
}

func TestExplicitEndIsRejectedToKeepExperimentScoped(t *testing.T) {
	program := Program{ID: "CFG-MANUAL-END", Entry: "entry", Blocks: []Block{{
		ID: "entry", Operations: []ownership.Operation{{Kind: ownership.EndLoan, Loan: "view"}},
	}}}
	analysis := Analyze(program, Options{})
	if analysis.Valid || analysis.Diagnostic.Code != "cfg.explicit_end_out_of_scope" {
		t.Fatalf("unexpected analysis: %+v", analysis)
	}
}

func TestMalformedGraphsProduceStableDiagnostics(t *testing.T) {
	cases := []struct {
		name    string
		program Program
		code    string
	}{
		{
			name:    "unknown entry",
			program: Program{ID: "BAD-ENTRY", Entry: "missing", Blocks: []Block{{ID: "entry"}}},
			code:    "cfg.unknown_entry",
		},
		{
			name:    "unknown successor",
			program: Program{ID: "BAD-EDGE", Entry: "entry", Blocks: []Block{{ID: "entry", Successors: []string{"missing"}}}},
			code:    "cfg.unknown_successor",
		},
		{
			name: "duplicate loan",
			program: Program{ID: "BAD-LOAN", Entry: "entry", Blocks: []Block{{ID: "entry", Operations: []ownership.Operation{
				{Kind: ownership.BorrowShared, Place: "p", Loan: "view"},
				{Kind: ownership.BorrowShared, Place: "q", Loan: "view"},
			}}}},
			code: "cfg.duplicate_loan",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			analysis := Analyze(test.program, Options{})
			if analysis.Valid || analysis.Diagnostic == nil || analysis.Diagnostic.Code != test.code {
				t.Fatalf("analysis=%+v, want %s", analysis, test.code)
			}
		})
	}
}

func branchSpecificProgram() Program {
	return Program{ID: "CFG-EDGE-MINIMAL", Entry: "entry", Blocks: []Block{
		{ID: "entry", Operations: []ownership.Operation{
			{Kind: ownership.Declare, Place: "p"},
			{Kind: ownership.BorrowShared, Place: "p", Loan: "view"},
		}, Successors: []string{"uses", "mutates"}},
		{ID: "uses", Operations: []ownership.Operation{{Kind: ownership.ReadLoan, Loan: "view"}}, Successors: []string{"done"}},
		{ID: "mutates", Operations: []ownership.Operation{{Kind: ownership.MutatePlace, Place: "p"}}, Successors: []string{"done"}},
		{ID: "done"},
	}}
}

func comparisonHasCode(comparison Comparison, code string) bool {
	for _, path := range comparison.Paths {
		if path.Analyzer.Diagnostic != nil && path.Analyzer.Diagnostic.Code == code {
			return true
		}
	}
	return false
}

func hasDisagreeingPath(comparison Comparison) bool {
	for _, path := range comparison.Paths {
		if !path.Agreement {
			return true
		}
	}
	return false
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsPrefix(values []string, prefix string) bool {
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func reverseStrings(values []string) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
