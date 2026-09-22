package session_test

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestPhase17SourceFrontierPreWidening preserves the source-first evidence
// commit.  Task 2 replaces the historical type.return_mismatch expectations
// with the two call-contract diagnostics while retaining this provenance.
func TestPhase17SourceFrontierPreWidening(t *testing.T) {
	for _, tc := range []struct {
		fixture string
		codes   []string
	}{
		{"return_type_tracer.lang", []string{"type.return_mismatch", "type.return_mismatch"}},
		{"call_argument_type_mismatch.lang", []string{"type.return_mismatch", "check.call_argument_type_mismatch"}},
		{"call_return_type_unrepresentable.lang", []string{"type.return_mismatch", "check.call_return_type_unrepresentable"}},
	} {
		t.Run(tc.fixture, func(t *testing.T) {
			checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase17", tc.fixture))
			if err != nil {
				t.Fatalf("CheckFile: %v", err)
			}
			if len(checked.Diagnostics) != len(tc.codes) {
				t.Fatalf("diagnostic count = %d, want %d: %+v", len(checked.Diagnostics), len(tc.codes), checked.Diagnostics)
			}
			for i, code := range tc.codes {
				if checked.Diagnostics[i].Code != code {
					t.Fatalf("diagnostic[%d] = %q, want %q: %+v", i, checked.Diagnostics[i].Code, code, checked.Diagnostics)
				}
			}
		})
	}
}

func TestPhase17SourceFrontierFixturesAreParserValid(t *testing.T) {
	for _, fixture := range []string{
		"return_type_tracer.lang",
		"call_argument_type_mismatch.lang",
		"call_return_type_unrepresentable.lang",
	} {
		t.Run(fixture, func(t *testing.T) {
			result, err := session.FormatFile(testsupport.ProjectPath("testdata", "phase17", fixture))
			if err != nil || len(result.Diagnostics) != 0 {
				t.Fatalf("fixture must parse and format: result=%+v err=%v", result, err)
			}
		})
	}
}
