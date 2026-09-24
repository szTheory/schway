package check_test

import (
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestPhase18ComputedScrutineeProductionPathAccepted(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase18", "computed_match.lang")
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

func TestPhase18LoanAcrossBranchFixture(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase18", "loan_across_branch.lang")
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
}
