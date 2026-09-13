package pathoracle_test

import (
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/pathoracle"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestPayloadPathEnumerationTerminates is Task 3's own probe methodology
// (D-12-02) applied to the one admission peer that cannot be checked by a
// case arm: pathoracle never switches exhaustively on core.OperationKind
// (D-04-29) -- it walks only the core.OpBorrowShared/core.OpBorrowExclusive
// membership test at pathoracle.go:320-326 -- so "handled" for this file
// means the walk completes without error on a program containing the new
// payload kinds, exactly as core_test.go's own exhaustive-dispatch control
// doc comment already defines it. No case arm is added to pathoracle.go for
// this reason.
//
// Both plan 02's tracer fixture and plan 03's borrow-interaction fixture are
// driven, not just one: a payload fixture carrying no borrow (the tracer
// alone) produces a DEGENERATE enumeration -- zero loans, so the
// termination-within-MaxPaths claim would report pass without ever
// exercising the hazard it names (the Pitfall-1 vacuous control this plan's
// own prohibitions forbid). payload_borrow_interaction.lang's `choose`
// function is what actually meets real loan machinery; its `identity`
// function is the same bare-value payload shape as the tracer, included for
// completeness.
func TestPayloadPathEnumerationTerminates(t *testing.T) {
	tests := []struct {
		fixture           string
		function          string
		wantLoanEndpoints int
	}{
		{"testdata/phase12/payload_tracer.lang", "identity", 0},
		{"testdata/phase12/payload_borrow_interaction.lang", "choose", 1},
		{"testdata/phase12/payload_borrow_interaction.lang", "identity", 0},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.fixture+"/"+tc.function, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath(splitFixturePath(tc.fixture)...))
			if err != nil {
				t.Fatalf("read %s: %v", tc.fixture, err)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) > 0 {
				t.Fatalf("%s: unexpected diagnostics: %v", tc.fixture, checked.Diagnostics)
			}
			lookup := pathoracle.BuildCalleeLookup(checked.Program)

			found := false
			for _, function := range checked.Program.Functions {
				if function.Name != tc.function {
					continue
				}
				found = true
				endpoints, work, err := pathoracle.RecomputeEndpoints(function, lookup)
				// A nil error here IS the termination-within-MaxPaths proof:
				// RecomputeEndpoints calls EnumeratePaths with cap ==
				// pathoracle.MaxPaths internally and returns a named error
				// the instant that cap is exceeded, so err == nil means the
				// enumeration terminated within the declared cap.
				if err != nil {
					t.Fatalf("%s/%s: RecomputeEndpoints error: %v", tc.fixture, tc.function, err)
				}
				if work <= 0 {
					t.Fatalf("%s/%s: expected pathoracle to inspect at least one operation, got work=%d", tc.fixture, tc.function, work)
				}
				if len(endpoints) != tc.wantLoanEndpoints {
					t.Fatalf("%s/%s: expected %d loan endpoint(s), got %d: %+v", tc.fixture, tc.function, tc.wantLoanEndpoints, len(endpoints), endpoints)
				}
			}
			if !found {
				t.Fatalf("%s: function %q not found in checked program", tc.fixture, tc.function)
			}
		})
	}
}

func splitFixturePath(path string) []string {
	var parts []string
	start := 0
	for index := 0; index < len(path); index++ {
		if path[index] == '/' {
			parts = append(parts, path[start:index])
			start = index + 1
		}
	}
	parts = append(parts, path[start:])
	return parts
}
