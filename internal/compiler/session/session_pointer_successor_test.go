package session_test

import (
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase25FamilyConflict(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"shared", "module phase25.shared_conflict\nexport { fn relay }\nfn relay(value: Buffer) -> Buffer {\n let first = borrow value\n let second = borrow mut value\n let observed = take first\n second\n}\n"},
		{"exclusive", string(mustPhase25Fixture(t, "exclusive_conflict_reject.schway"))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checked := session.Check([]byte(test.source))
			if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "ownership.borrow_conflict" {
				t.Fatalf("expected a semantic borrow conflict, got %+v", checked.Diagnostics)
			}
			problem := checked.Diagnostics[0]
			if problem.Primary.Start == problem.Primary.End || len(problem.Causes) < 3 {
				t.Fatalf("conflict lost primary source attribution or causes: %+v", problem)
			}
		})
	}
}

func TestPhase25FamilyEscape(t *testing.T) {
	result, err := session.CheckCommandFile(testsupport.ProjectPath("testdata", "phase25", "exclusive_escape_reject.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "invalid" || len(result.Diagnostics) != 1 {
		t.Fatalf("expected exclusive escape refusal, got %+v", result)
	}
	problem := result.Diagnostics[0]
	if problem.Code != "core.origin_omitted" {
		t.Fatalf("escape must identify its source and semantic boundary: %+v", problem)
	}
}

func TestPhase25IndependentPeer(t *testing.T) {
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase3", "sequential_shared_then_exclusive_accept.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("ended shared loan followed by exclusive access was rejected: %+v", checked.Diagnostics)
	}
	if result := corevalidate.Validate(checked.Program); !result.Valid {
		t.Fatalf("corevalidate rejected sequential families: %+v", result.Problems)
	}
	if problems := originvalidate.ValidatePublished(checked.Program); len(problems) != 0 {
		t.Fatalf("originvalidate rejected sequential families: %+v", problems)
	}
	if err := pathoracle.ValidateLocalOwnerPaths(checked.Program); err != nil {
		t.Fatalf("pathoracle rejected sequential families: %v", err)
	}
}

func TestPhase25UnsupportedPointerShape(t *testing.T) {
	for _, test := range []struct{ name, fixture, code string }{
		{"exclusive mutation conflicts with active loan", "exclusive_conflict_reject.schway", "ownership.borrow_conflict"},
		{"exclusive result escape stays unsupported", "exclusive_escape_reject.schway", "core.origin_omitted"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.fixture == "exclusive_escape_reject.schway" {
				result, err := session.CheckCommandFile(testsupport.ProjectPath("testdata", "phase25", test.fixture))
				if err != nil {
					t.Fatal(err)
				}
				if result.Status != "invalid" || len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != test.code {
					t.Fatalf("source control %q produced %+v, want %s", test.fixture, result.Diagnostics, test.code)
				}
				return
			}
			checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase25", test.fixture))
			if err != nil {
				t.Fatal(err)
			}
			if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != test.code {
				t.Fatalf("source control %q produced %+v, want %s", test.fixture, checked.Diagnostics, test.code)
			}
			if checked.Diagnostics[0].Primary.Start == checked.Diagnostics[0].Primary.End {
				t.Fatalf("source control %q lost its offending source span: %+v", test.fixture, checked.Diagnostics[0])
			}
		})
	}
}

func TestPhase25OwnershipDiagnosticBoundary(t *testing.T) {
	for _, test := range []struct{ name, phase, fixture, code string }{
		{"moved-from use", "phase2", "use_after_move.schway", "ownership.use_after_move"},
		{"discarded acquired owner", "phase23", "discard_owner.schway", "check.local_owner_discarded"},
	} {
		t.Run(test.name, func(t *testing.T) {
			checked, err := session.CheckFile(testsupport.ProjectPath("testdata", test.phase, test.fixture))
			if err != nil {
				t.Fatal(err)
			}
			if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != test.code {
				t.Fatalf("source control %q produced %+v, want %s", test.fixture, checked.Diagnostics, test.code)
			}
			if checked.Diagnostics[0].Primary.Start == checked.Diagnostics[0].Primary.End {
				t.Fatalf("source control %q lost source attribution: %+v", test.fixture, checked.Diagnostics[0])
			}
		})
	}
}

func mustPhase25Fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase25", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
