package check_test

import (
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestPhase18ComputedScrutineeFrontierPinned(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase18", "computed_match.lang")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) == 0 {
		t.Fatal("fixture did not reach the parser's terminal-result refusal")
	}
	diagnostic := checked.Diagnostics[0]
	if diagnostic.Code != "syntax.expected_linear_result" {
		t.Fatalf("diagnostic code = %q, want syntax.expected_linear_result at the unsupported terminal match: %+v", diagnostic.Code, diagnostic)
	}
	if diagnostic.Primary.Start <= 0 || diagnostic.Primary.End <= diagnostic.Primary.Start {
		t.Fatalf("computed-scrutinee diagnostic has no source location: %+v", diagnostic.Primary)
	}
	if diagnostic.Primary.End > len(source) || string(source[diagnostic.Primary.Start:diagnostic.Primary.End]) != "match" {
		t.Fatalf("diagnostic span = %q, want the terminal match token", source[diagnostic.Primary.Start:diagnostic.Primary.End])
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
	checked := session.Check(source)
	if len(checked.Diagnostics) == 0 || checked.Diagnostics[0].Code != "syntax.expected_linear_result" {
		t.Fatalf("want the pinned terminal-match parser frontier, got %+v", checked.Diagnostics)
	}
	problem := checked.Diagnostics[0]
	if problem.Primary.End > len(source) || string(source[problem.Primary.Start:problem.Primary.End]) != "match" {
		t.Fatalf("diagnostic does not point at terminal match: %+v", problem.Primary)
	}
}
