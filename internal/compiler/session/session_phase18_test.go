package session

import (
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/testsupport"
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
	checked := Check(source)
	if len(checked.Diagnostics) == 0 {
		t.Fatalf("fixture %s passed without reaching the pre-implementation frontier", fixture)
	}
	problem := checked.Diagnostics[0]
	if problem.Code != "syntax.expected_linear_result" {
		t.Fatalf("fixture %s first diagnostic = %q, want syntax.expected_linear_result: %+v", fixture, problem.Code, checked.Diagnostics)
	}
	if problem.Primary.End > len(source) || problem.Primary.End <= problem.Primary.Start || string(source[problem.Primary.Start:problem.Primary.End]) != "match" {
		t.Fatalf("fixture %s frontier span does not identify its terminal match: %+v", fixture, problem.Primary)
	}
}

func TestPhase18ResultFixtureFrontier(t *testing.T) {
	assertPhase18ComputedFrontier(t, "result_computed_match.lang", "let result = identity(input)", "match result")
}

func TestPhase18PayloadFixtureFrontier(t *testing.T) {
	assertPhase18ComputedFrontier(t, "payload_return.lang", "let result = identity(value)", "Ok(payload) => payload", "match result")
}
