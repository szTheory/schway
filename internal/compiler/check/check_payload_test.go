package check

import (
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// readPhase12Fixture reads a testdata/phase12/*.lang fixture by name, the
// package check internal-test sibling of readTestdataFixture (which is
// hardcoded to testdata/phase3) -- Phase 12's fixtures live in their own
// directory.
func readPhase12Fixture(t *testing.T, name string) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase12", name))
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	return source
}

// TestPayloadPatternRefusals is Phase 12 Plan 03 Task 1's table over
// D-12-15's three named refusals -- a pattern that is wrong in a way
// set-membership exhaustiveness cannot express (wrong binder arity, a
// binder on a nullary alternative, or a missing binder on a
// payload-carrying alternative). The fourth row asserts Plan 02's own
// tracer fixture still checks clean, so this table cannot pass by
// refusing everything.
func TestPayloadPatternRefusals(t *testing.T) {
	cases := []struct {
		name         string
		fixture      string
		expectedCode string
	}{
		{name: "arity mismatch", fixture: "payload_arity_mismatch.lang", expectedCode: "check.payload_arity_mismatch"},
		{name: "binder on nullary alternative", fixture: "payload_binder_on_nullary.lang", expectedCode: "check.binder_on_nullary_alternative"},
		{name: "missing payload binder", fixture: "payload_missing_binder.lang", expectedCode: "check.missing_payload_binder"},
		{name: "tracer fixture stays clean", fixture: "payload_tracer.lang", expectedCode: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := readPhase12Fixture(t, tc.fixture)
			program := mustParseProgram(t, source)
			result := Program(program)

			if tc.expectedCode == "" {
				if len(result.Diagnostics) != 0 {
					t.Fatalf("%s: expected zero diagnostics, got %+v", tc.fixture, result.Diagnostics)
				}
				return
			}

			if len(result.Diagnostics) != 1 {
				t.Fatalf("%s: expected exactly one diagnostic, got %d: %+v", tc.fixture, len(result.Diagnostics), result.Diagnostics)
			}
			if result.Diagnostics[0].Code != tc.expectedCode {
				t.Fatalf("%s: expected code %q, got %q (%+v)", tc.fixture, tc.expectedCode, result.Diagnostics[0].Code, result.Diagnostics[0])
			}
		})
	}
}

// TestResourcePayloadRefused is Phase 12 Plan 03 Task 2's proof of D-12-27's
// fail-closed resource-payload refusal: a companion case (payloads that do
// NOT structurally contain a Phase-4 tracked-resource-derived value) is
// what stops the refusal from reading as a blanket payload ban.
func TestResourcePayloadRefused(t *testing.T) {
	t.Run("resource-carrying payload refused", func(t *testing.T) {
		source := readPhase12Fixture(t, "payload_resource_refused.lang")
		program := mustParseProgram(t, source)
		result := Program(program)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("expected exactly one diagnostic, got %d: %+v", len(result.Diagnostics), result.Diagnostics)
		}
		if result.Diagnostics[0].Code != "check.resource_payload_refused" {
			t.Fatalf("expected check.resource_payload_refused, got %q (%+v)", result.Diagnostics[0].Code, result.Diagnostics[0])
		}
	})

	// Companion: a program with NO foreign declarations at all has nothing
	// payloadTypeNamesForeignReturnType can match, so Byte, Buffer, and a
	// nullary-ADT payload -- the three payload shapes this maturity
	// supports (LANGUAGE-MATURITY.md) -- all stay accepted.
	t.Run("companion: Byte, Buffer, and nullary-ADT payloads with no matching foreign return type stay accepted", func(t *testing.T) {
		source := []byte(`module result.payload_resource_companion

export {
  type Fault
  type Wrapper
  fn identity
}

data Fault =
  | Broken

data Wrapper =
  | HoldsByte(Byte)
  | HoldsBuffer(Buffer)
  | HoldsFault(Fault)

fn identity(w: Wrapper) -> Wrapper {
  match w {
    HoldsByte(v) => HoldsByte(v)
    HoldsBuffer(v) => HoldsBuffer(v)
    HoldsFault(v) => HoldsFault(v)
  }
}
`)
		parsed := syntax.Parse(source)
		if len(parsed.Diagnostics) != 0 {
			t.Fatalf("companion fixture failed to parse: %+v", parsed.Diagnostics)
		}
		result := Program(parsed.Program)
		if len(result.Diagnostics) != 0 {
			t.Fatalf("expected zero diagnostics (no foreign block, nothing is resource-derived), got %+v", result.Diagnostics)
		}
	})
}
