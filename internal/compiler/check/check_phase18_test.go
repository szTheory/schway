package check_test

import (
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase18ComputedScrutineeProductionPathAccepted(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase18", "computed_match.schway")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("computed terminal match was refused: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) != 1 {
		t.Fatalf("checked function count = %d, want 1", len(checked.Program.Functions))
	}
	function := checked.Program.Functions[0]
	if function.Match == nil || function.Match.Scrutinee != "computed" || function.Match.ScrutineeID == "" || function.Linear == nil {
		t.Fatalf("computed match did not reach checked branch core: %+v", function)
	}
	matchedPlace := false
	for _, place := range function.Linear.Places {
		if place.ID == function.Match.ScrutineeID && place.Name == "computed" {
			matchedPlace = true
		}
	}
	if !matchedPlace {
		t.Fatalf("match scrutinee identity %q does not identify the computed place", function.Match.ScrutineeID)
	}
	if len(function.Linear.Operations) < 2 || function.Linear.Operations[0].Kind == "return" {
		t.Fatalf("linear prefix/branch operations are missing or malformed: %+v", function.Linear.Operations)
	}
}

func TestPhase18ComputedScrutineeRefusals(t *testing.T) {
	base := `module phase18.computed_refusal
export {
  type Choice
  fn select
}
data Choice =
  | Left
  | Right
`
	cases := []struct {
		name, function, want string
	}{
		{
			name: "out of scope",
			function: `fn select(input: Choice) -> Choice {
  let computed = input
  match missing {
    Left => Left
    Right => Right
  }
}`,
			want: "name.unknown_scrutinee",
		},
		{
			name: "shadows parameter",
			function: `fn select(computed: Choice) -> Choice {
  let computed = computed
  match computed {
    Left => Left
    Right => Right
  }
}`,
			want: "name.unknown_scrutinee",
		},
		{
			name: "non data parameter",
			function: `fn select(input: Byte) -> Byte {
  let computed = input
  match computed {
    Left => Left
    Right => Right
  }
}`,
			want: "type.unknown",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checked := session.Check([]byte(base + tc.function))
			for _, diagnostic := range checked.Diagnostics {
				if diagnostic.Code == tc.want {
					return
				}
			}
			t.Fatalf("diagnostics = %+v, want stable code %q", checked.Diagnostics, tc.want)
		})
	}
}

func TestPhase18ResultComputedMatchChecker(t *testing.T) {
	fixturePath := testsupport.ProjectPath("testdata", "phase18", "result_computed_match.schway")
	source, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read Result computed-match fixture: %v", err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("Result computed-match fixture diagnostics = %+v, want none", checked.Diagnostics)
	}
	functions := map[string]core.Function{}
	for _, function := range checked.Program.Functions {
		functions[function.Name] = function
	}
	callee, calleeOK := functions["produce"]
	caller, callerOK := functions["main"]
	if !calleeOK || !callerOK || callee.Match == nil || caller.Match == nil || callee.Linear == nil || caller.Linear == nil {
		t.Fatalf("checked fixture lacks both branch functions: %+v", checked.Program.Functions)
	}
	if callee.Parameter.Type != "Resource" || callee.ReturnType != "Result" {
		t.Fatalf("callee contract = %s -> %s, want Resource -> Result", callee.Parameter.Type, callee.ReturnType)
	}
	calleeScrutinee, ok := phase18Place(callee, callee.Match.ScrutineeID)
	if !ok || calleeScrutinee.Name != "computed" || phase18TypeName(callee, calleeScrutinee.TypeID) != "Resource" {
		t.Fatalf("callee scrutinee place = %+v, want computed Resource", calleeScrutinee)
	}
	if len(callee.Match.Arms) != 1 || callee.Match.Arms[0].Pattern != "Raw" || callee.Match.Arms[0].Value != "Accepted" || callee.Match.Arms[0].ValuePlaceID == "" {
		t.Fatalf("callee match arm lacks the typed Accepted value place: %+v", callee.Match.Arms)
	}
	calleeValuePlace, ok := phase18Place(callee, callee.Match.Arms[0].ValuePlaceID)
	if !ok || phase18TypeName(callee, calleeValuePlace.TypeID) != "Result" {
		t.Fatalf("callee value place = %+v, want Result", calleeValuePlace)
	}
	calleeReturnFound := false
	for _, operation := range callee.Linear.Operations {
		if operation.Kind == core.OpReturn && operation.SourceID == callee.Match.Arms[0].ValuePlaceID && operation.TypeID == calleeValuePlace.TypeID {
			calleeReturnFound = true
		}
	}
	if !calleeReturnFound {
		t.Fatalf("callee OpReturn does not read its Result arm value place %q", callee.Match.Arms[0].ValuePlaceID)
	}
	callerScrutinee, ok := phase18Place(caller, caller.Match.ScrutineeID)
	if !ok || callerScrutinee.Name != "result" || phase18TypeName(caller, callerScrutinee.TypeID) != "Result" {
		t.Fatalf("caller scrutinee place = %+v, want call-produced Result", callerScrutinee)
	}
	callTargetsScrutinee := false
	for _, operation := range caller.Linear.Operations {
		if operation.Kind == core.OpCall && operation.TargetID == caller.Match.ScrutineeID {
			callTargetsScrutinee = true
		}
	}
	if !callTargetsScrutinee {
		t.Fatalf("caller scrutinee %q is not the call target", caller.Match.ScrutineeID)
	}

	wrongCalleeValue := strings.Replace(string(source), "Raw => Accepted", "Raw => Raw", 1)
	phase18RequireDiagnostic(t, []byte(wrongCalleeValue), "type.return_mismatch")

	wrongCallerPattern := strings.Replace(string(source), "    Accepted => Accepted", "    Raw => Accepted", 1)
	phase18RequireDiagnostic(t, []byte(wrongCallerPattern), "match.unreachable")

	omittedResultAlternative := strings.Replace(string(source), "    Rejected => Rejected\n", "", 1)
	phase18RequireDiagnostic(t, []byte(omittedResultAlternative), "match.non_exhaustive")

	ordinaryMismatch := `module phase18.ordinary_return_mismatch
export { fn wrong }
data Resource = | Raw
data Result = | Classified
fn wrong(resource: Resource) -> Result {
  match resource {
    Raw => { resource }
  }
}`
	phase18RequireDiagnostic(t, []byte(ordinaryMismatch), "type.return_mismatch")
}

func phase18Place(function core.Function, id string) (core.Place, bool) {
	if function.Linear == nil {
		return core.Place{}, false
	}
	for _, place := range function.Linear.Places {
		if place.ID == id {
			return place, true
		}
	}
	return core.Place{}, false
}

func phase18TypeName(function core.Function, typeID string) string {
	if function.Linear == nil {
		return ""
	}
	for _, fact := range function.Linear.Types {
		if fact.ID == typeID {
			return fact.Shape.Constructor
		}
	}
	return ""
}

func phase18RequireDiagnostic(t *testing.T, source []byte, code string) {
	t.Helper()
	checked := session.Check(source)
	for _, diag := range checked.Diagnostics {
		if diag.Code == code {
			return
		}
	}
	t.Fatalf("diagnostics = %+v, want %s", checked.Diagnostics, code)
}

func TestPhase18LoanAcrossBranchFixture(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase18", "loan_across_branch.schway")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if !strings.Contains(string(source), "let view = borrow input") || !strings.Contains(string(source), "let observed = take view") {
		t.Fatal("fixture does not witness a pre-match borrow consumed in one arm")
	}
	parsed := syntax.Parse(source)
	foundTerminalMatch := false
	for _, function := range parsed.Program.Funcs {
		if function.Name == "select" && function.Body.Linear != nil && function.Body.Linear.TerminalMatch != nil {
			foundTerminalMatch = true
		}
	}
	if len(parsed.Diagnostics) != 0 || !foundTerminalMatch {
		t.Fatalf("terminal computed match did not parse into the linear-body form: diagnostics=%+v", parsed.Diagnostics)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("production checker rejected S-010 fixture: %+v", checked.Diagnostics)
	}
	var selectFn *core.Function
	for index := range checked.Program.Functions {
		if checked.Program.Functions[index].Name == "select" {
			selectFn = &checked.Program.Functions[index]
		}
	}
	if selectFn == nil || selectFn.Linear == nil || selectFn.Match == nil {
		t.Fatalf("checked source omitted select CFG: %+v", checked.Program.Functions)
	}
	var borrowID string
	for _, op := range selectFn.Linear.Operations {
		if op.Kind == core.OpBorrowShared {
			borrowID = op.LoanID
		}
	}
	if borrowID == "" {
		t.Fatal("production linear prefix omitted source borrow")
	}
	var pointUses, edgeEnds int
	var offEdge string
	for _, endpoint := range selectFn.Linear.LoanEndpoints {
		if endpoint.LoanID != borrowID {
			continue
		}
		switch endpoint.Kind {
		case "point":
			pointUses++
		case "edge":
			edgeEnds++
			offEdge = endpoint.EdgeID
		default:
			t.Fatalf("production endpoint introduced unsupported kind %q", endpoint.Kind)
		}
	}
	if pointUses != 1 || edgeEnds != 1 {
		t.Fatalf("borrow endpoints = %+v, want one point use and one diverging edge", selectFn.Linear.LoanEndpoints)
	}
	var offEdgeIsSibling bool
	for _, edge := range selectFn.Linear.Edges {
		if edge.ID == offEdge && edge.Pattern == "Off" {
			offEdgeIsSibling = true
		}
	}
	if !offEdgeIsSibling {
		t.Fatalf("unused sibling did not receive the edge endpoint: edge=%q", offEdge)
	}

	// The neither-arm source control preserves the admitted ownership shape
	// while proving that the one-arm edge endpoint disappears when the use is
	// removed. The both-arm control is a valid synthetic CFG in check_test.go;
	// spelling a second `take view` in source would move the same value twice.
	neither := strings.Replace(string(source), "      let observed = take view\n", "", 1)
	result := session.Check([]byte(neither))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("neither-arm control source rejected: %+v", result.Diagnostics)
	}
	var neitherFn *core.Function
	for index := range result.Program.Functions {
		if result.Program.Functions[index].Name == "select" {
			neitherFn = &result.Program.Functions[index]
		}
	}
	if neitherFn == nil || neitherFn.Linear == nil {
		t.Fatalf("neither-arm control omitted select CFG: %+v", result.Program.Functions)
	}
	for _, endpoint := range neitherFn.Linear.LoanEndpoints {
		if endpoint.LoanID == borrowID && endpoint.Kind == "edge" {
			t.Fatalf("neither-arm control retained one-arm edge classification: %+v", endpoint)
		}
	}
}
