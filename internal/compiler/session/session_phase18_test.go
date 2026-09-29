package session_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func assertPhase18ComputedFrontier(t *testing.T, fixture string, sourceNeedles ...string) {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase18", fixture))
	if err != nil {
		t.Fatalf("read %s: %v", fixture, err)
	}
	for _, needle := range sourceNeedles {
		if !strings.Contains(string(source), needle) {
			t.Fatalf("fixture %s is missing source witness %q", fixture, needle)
		}
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture %s terminal match did not parse: %+v", fixture, parsed.Diagnostics)
	}
	if len(parsed.Program.Funcs) == 0 || parsed.Program.Funcs[len(parsed.Program.Funcs)-1].Body.Linear == nil || parsed.Program.Funcs[len(parsed.Program.Funcs)-1].Body.Linear.TerminalMatch == nil {
		t.Fatalf("fixture %s does not preserve its terminal computed match", fixture)
	}
}

func TestPhase18ResultFixtureFrontier(t *testing.T) {
	assertPhase18ComputedFrontier(t, "result_computed_match.schway", "let result = produce(input)", "match result")
}

// TestPhase18ResultComputedMatch follows the Result value from its Lang call
// target into the caller's terminal match, admits the checked source through
// both independent peers, and compares the four production execution routes.
func TestPhase18ResultComputedMatch(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase18", "result_computed_match.schway"))
	if err != nil {
		t.Fatalf("read Result caller fixture: %v", err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check Result caller: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) != 2 {
		t.Fatalf("Result caller has %d functions, want identity and main", len(checked.Program.Functions))
	}
	var main *core.Function
	for index := range checked.Program.Functions {
		if checked.Program.Functions[index].Name == "main" {
			main = &checked.Program.Functions[index]
		}
	}
	if main == nil || main.Match == nil || main.Linear == nil || main.Match.Scrutinee != "result" || main.Match.ScrutineeID == "" {
		t.Fatalf("main does not match its computed Result place: %+v", main)
	}
	scrutineeType := ""
	for _, place := range main.Linear.Places {
		if place.ID != main.Match.ScrutineeID {
			continue
		}
		for _, fact := range main.Linear.Types {
			if fact.ID == place.TypeID {
				scrutineeType = fact.Shape.Constructor
			}
		}
	}
	if scrutineeType != "Result" {
		t.Fatalf("main scrutinee type = %q, want Result", scrutineeType)
	}
	callReturnedScrutinee := false
	for _, operation := range main.Linear.Operations {
		if operation.Kind == core.OpCall && operation.CalleeID != "" && operation.TargetID == main.Match.ScrutineeID {
			callReturnedScrutinee = true
		}
	}
	if !callReturnedScrutinee {
		t.Fatalf("match scrutinee %q was not produced by an OpCall in main: %+v", main.Match.ScrutineeID, main.Linear.Operations)
	}
	if peer := corevalidate.Validate(checked.Program); !peer.Valid {
		t.Fatalf("independent core admission: %+v", peer.Problems)
	}
	if _, err := originvalidate.BuildInterface(checked.Program); err != nil {
		t.Fatalf("independent origin admission: %v", err)
	}

	for _, input := range []string{"Raw"} {
		engines := phase11RunFourTiersWithSupplier(t, context.Background(), checked.Program, "main", input, cgen.EmitNative, "public EmitNative")
		engines["interpreter"] = phase16ProjectInterpreterSchema2(t, checked.Program, engines["interpreter"])
		if len(engines) != 4 {
			t.Fatalf("input %q produced %d execution engines, want interpreter, O0, O3, and O3-LTO", input, len(engines))
		}
		for name, got := range engines {
			if got.Outcome.Kind != "returned" || got.Outcome.Value != "Accepted" {
				t.Fatalf("input %q engine %s outcome = %+v, want Accepted", input, name, got.Outcome)
			}
		}
		if err := session.Phase5CompareProgramEngines("testdata/phase18/result_computed_match.schway:"+input, checked.Program, engines); err != nil {
			t.Fatalf("input %q four-engine comparison: %v", input, err)
		}
	}
}

func TestPhase18ResultComputedMatchAdmission(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase18", "result_computed_match.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check: %+v", checked.Diagnostics)
	}
	if peer := corevalidate.Validate(checked.Program); !peer.Valid {
		t.Fatalf("core peer: %+v", peer.Problems)
	}
	if _, err := originvalidate.BuildInterface(checked.Program); err != nil {
		t.Fatalf("origin peer: %v", err)
	}
	var foundCallee, foundCaller bool
	for _, fn := range checked.Program.Functions {
		if fn.Name == "produce" {
			foundCallee = true
			if fn.ReturnType != "Result" || fn.Match == nil || len(fn.Match.Arms) != 1 || fn.Match.Arms[0].ValuePlaceID == "" {
				t.Fatalf("produce typed arm contract: %+v", fn)
			}
		}
		if fn.Name == "main" {
			foundCaller = true
			if fn.Match == nil || fn.Match.ScrutineeID == "" {
				t.Fatalf("main computed match absent: %+v", fn)
			}
		}
	}
	if !foundCallee || !foundCaller {
		t.Fatal("source fixture does not contain produce and main")
	}
}

// TestPhase18FiveAxis includes the diagnostic-ID axis on a refused Phase 18
// computed match. The four execution axes are exercised above on the accepted
// call-result source; reject programs have no execution document.
func TestPhase18FiveAxis(t *testing.T) {
	refused := []byte(`module phase18.result_refusal
export {
  type Result
  fn main
}
data Result =
  | Accepted
  | Rejected
fn main(input: Result) -> Result {
  match missing {
    Accepted => Accepted
    Rejected => Rejected
  }
}`)
	first := session.Check(refused)
	second := session.Check(refused)
	if len(first.Diagnostics) == 0 || len(second.Diagnostics) == 0 {
		t.Fatalf("negative control was not refused: first=%+v second=%+v", first.Diagnostics, second.Diagnostics)
	}
	if err := session.Phase5CompareDiagnosticIDs("phase18-result-computed-refusal", map[string]diagnostic.Diagnostic{
		"check-run-1": first.Diagnostics[0],
		"check-run-2": second.Diagnostics[0],
	}); err != nil {
		t.Fatalf("diagnostic-ID refusal comparison: %v", err)
	}
}

func TestPhase18PayloadFixtureFrontier(t *testing.T) {
	assertPhase18ComputedFrontier(t, "payload_return.schway", "let result = take value", "Ok(payload) => Ok(payload)", "match result")
}

func TestPhase18ComputedSourceFourTierDifferential(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase18", "computed_match.schway"))
	if err != nil {
		t.Fatalf("read computed-match fixture: %v", err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check computed-match fixture: %+v", checked.Diagnostics)
	}
	var function *core.Function
	for index := range checked.Program.Functions {
		if checked.Program.Functions[index].Name == "select" {
			function = &checked.Program.Functions[index]
			break
		}
	}
	if function == nil || function.Match == nil || function.Linear == nil || function.Match.Scrutinee != "computed" {
		t.Fatalf("source did not produce a checked computed branch: %+v", checked.Program.Functions)
	}
	if peer := corevalidate.Validate(checked.Program); !peer.Valid {
		t.Fatalf("independent core admission: %+v", peer.Problems)
	}
	for _, input := range []string{"Left", "Right"} {
		engines := phase11RunFourTiersWithSupplier(t, context.Background(), checked.Program, "select", input, cgen.EmitProgramNativeForTest, "emitProgram")
		engines["interpreter"] = phase16ProjectInterpreterSchema2(t, checked.Program, engines["interpreter"])
		if len(engines) != 4 {
			t.Fatalf("input %q produced %d execution engines, want interpreter, O0, O3, and O3-LTO", input, len(engines))
		}
		if err := session.Phase5CompareProgramEngines("testdata/phase18/computed_match.schway:"+input, checked.Program, engines); err != nil {
			t.Fatalf("input %q five-axis comparison: %v", input, err)
		}
	}
}
