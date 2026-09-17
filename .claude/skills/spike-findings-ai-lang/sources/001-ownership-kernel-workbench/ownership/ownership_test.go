package ownership

import (
	"encoding/json"
	"testing"
)

func TestFixturesMatchExpectationsAndImplementationsAgree(t *testing.T) {
	file, err := LoadFixtures("../fixtures/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range file.Fixtures {
		fixture := fixture
		t.Run(fixture.Program.ID, func(t *testing.T) {
			comparison := Compare(fixture.Program, CheckerOptions{})
			if !comparison.Agreement {
				t.Fatalf("oracle/checker disagreement: oracle=%+v checker=%+v", comparison.Oracle.Diagnostic, comparison.Checker.Diagnostic)
			}
			if comparison.Oracle.Valid != fixture.ExpectValid {
				t.Fatalf("valid=%v, want %v; diagnostic=%+v", comparison.Oracle.Valid, fixture.ExpectValid, comparison.Oracle.Diagnostic)
			}
			if !fixture.ExpectValid && comparison.Oracle.Diagnostic.Code != fixture.ExpectCode {
				t.Fatalf("diagnostic=%q, want %q", comparison.Oracle.Diagnostic.Code, fixture.ExpectCode)
			}
		})
	}
}

func TestLastUseNormalizationInsertsLoanEnd(t *testing.T) {
	program := Program{ID: "normalize", Operations: []Operation{
		{Kind: Declare, Place: "p"},
		{Kind: BorrowShared, Place: "p", Loan: "l"},
		{Kind: ReadLoan, Loan: "l"},
		{Kind: MutatePlace, Place: "p"},
	}}
	normalized := Normalize(program)
	if len(normalized.Operations) != 5 {
		t.Fatalf("got %d operations, want 5", len(normalized.Operations))
	}
	end := normalized.Operations[3]
	if end.Kind != EndLoan || end.Loan != "l" || !end.Synthetic {
		t.Fatalf("unexpected inferred end: %+v", end)
	}
}

func TestIndependentNormalizersAgreeOnFixtureAndGeneratedCorpus(t *testing.T) {
	file, err := LoadFixtures("../fixtures/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	programs := make([]Program, 0, len(file.Fixtures)+7381)
	for _, fixture := range file.Fixtures {
		programs = append(programs, fixture.Program)
	}
	programs = append(programs, GeneratePrograms(4)...)
	for _, program := range programs {
		checker, err := json.Marshal(Normalize(program))
		if err != nil {
			t.Fatal(err)
		}
		oracle, err := json.Marshal(NormalizeOracle(program))
		if err != nil {
			t.Fatal(err)
		}
		if string(checker) != string(oracle) {
			t.Fatalf("normalizer disagreement for %s:\nchecker=%s\noracle=%s", program.ID, checker, oracle)
		}
	}
}

func TestResourcesReleaseInReverseOrderOnError(t *testing.T) {
	program := Program{ID: "cleanup", Operations: []Operation{
		{Kind: Declare, Place: "file", Resource: true},
		{Kind: Declare, Place: "lock", Resource: true},
		{Kind: ExitError},
	}}
	result := RunOracle(program)
	if !result.Valid {
		t.Fatal(result.Diagnostic)
	}
	var releases []string
	for _, event := range result.Events {
		if event.Kind == "release_auto" {
			releases = append(releases, event.Place)
		}
	}
	if len(releases) != 2 || releases[0] != "lock" || releases[1] != "file" {
		t.Fatalf("release order=%v, want [lock file]", releases)
	}
}

func TestNestedScopeReleasesInnerResourcesBeforeOuterResource(t *testing.T) {
	program := Program{ID: "nested-cleanup", Operations: []Operation{
		{Kind: Declare, Place: "outer", Resource: true},
		{Kind: BeginScope, Scope: "inner"},
		{Kind: Declare, Place: "first", Resource: true},
		{Kind: Declare, Place: "second", Resource: true},
		{Kind: EndScope, Scope: "inner"},
	}}
	comparison := Compare(program, CheckerOptions{})
	if !comparison.Agreement || !comparison.Oracle.Valid {
		t.Fatalf("unexpected comparison: %+v", comparison)
	}
	releases := []string{}
	for _, event := range comparison.Oracle.Events {
		if event.Kind == "release_auto" {
			releases = append(releases, event.Place+":"+event.Outcome)
		}
	}
	want := []string{"second:scope", "first:scope", "outer:success"}
	if len(releases) != len(want) {
		t.Fatalf("release order=%v, want %v", releases, want)
	}
	for index := range want {
		if releases[index] != want[index] {
			t.Fatalf("release order=%v, want %v", releases, want)
		}
	}
}

func TestExhaustiveShortSequencesAgree(t *testing.T) {
	result := FindMismatch(4, CheckerOptions{})
	if result.Mismatch != nil {
		bytes, _ := json.Marshal(result)
		t.Fatalf("unexpected mismatch after %d programs: %s", result.Programs, bytes)
	}
	if result.Programs != 7381 {
		t.Fatalf("enumerated %d programs, want 7381", result.Programs)
	}
}

func TestFaultInjectionProducesMinimalCounterexample(t *testing.T) {
	result := FindMismatch(3, CheckerOptions{AllowMoveWhileShared: true})
	if result.Mismatch == nil || result.MinimalInput == nil {
		t.Fatal("fault injection produced no mismatch")
	}
	// The later loan read is load-bearing: without it, last-use inference ends
	// the unused loan immediately and moving the owner is legal.
	if len(result.MinimalInput.Operations) != 4 {
		t.Fatalf("counterexample has %d operations, want minimal length 4", len(result.MinimalInput.Operations))
	}
	if result.Mismatch.Oracle.Diagnostic == nil || result.Mismatch.Oracle.Diagnostic.Code != "ownership.move_while_borrowed" {
		t.Fatalf("unexpected oracle result: %+v", result.Mismatch.Oracle)
	}
}

func TestMachineReadableResultIsStable(t *testing.T) {
	program := Program{ID: "stable", Operations: []Operation{
		{Kind: Declare, Place: "p"},
		{Kind: Move, Place: "p", Target: "q"},
		{Kind: ReadPlace, Place: "p"},
	}}
	first, err := json.Marshal(RunOracle(program))
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(RunOracle(program))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("JSON output changed:\n%s\n%s", first, second)
	}
	result := RunOracle(program)
	if result.Operations != 3 {
		t.Fatalf("operations=%d, want 3", result.Operations)
	}
	if len(result.Diagnostic.Repairs) == 0 {
		t.Fatal("ownership diagnostic has no semantic repair candidates")
	}
}
