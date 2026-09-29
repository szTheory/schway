package check

import (
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// readPhase12Fixture reads a testdata/phase12/*.schway fixture by name, the
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
		{name: "arity mismatch", fixture: "payload_arity_mismatch.schway", expectedCode: "check.payload_arity_mismatch"},
		{name: "binder on nullary alternative", fixture: "payload_binder_on_nullary.schway", expectedCode: "check.binder_on_nullary_alternative"},
		{name: "missing payload binder", fixture: "payload_missing_binder.schway", expectedCode: "check.missing_payload_binder"},
		{name: "tracer fixture stays clean", fixture: "payload_tracer.schway", expectedCode: ""},
		{name: "duplicate payload type", fixture: "payload_duplicate_payload_type.schway", expectedCode: "check.duplicate_payload_type"},
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
		source := readPhase12Fixture(t, "payload_resource_refused.schway")
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

// TestDuplicatePayloadTypeRefused is Phase 12 Plan 06 Task 2's control for
// CR-01/D-12-44's check.duplicate_payload_type refusal. It asserts the
// diagnostic CODE directly, in the same style as TestPayloadPatternRefusals,
// because interp.alternativeNameForPayloadType and
// cgen.alternativeNameForPayloadType independently reimplement the SAME
// flawed first-match-by-PayloadType derivation -- the phase's own
// three-engine convergence tests are structurally incapable of catching this
// defect class, since both engines would agree on the same wrong answer. A
// convergence test is not evidence here; only a direct code assertion is.
// Reverting Task 1's check.go walk must turn this test red (observed and
// recorded in 12-06-SUMMARY.md).
func TestDuplicatePayloadTypeRefused(t *testing.T) {
	cases := []struct {
		name         string
		fixture      string
		source       string
		expectedCode string
	}{
		{name: "colliding payload types refused", fixture: "payload_duplicate_payload_type.schway", expectedCode: "check.duplicate_payload_type"},
		{name: "tracer fixture stays clean (distinct payload types)", fixture: "payload_tracer.schway", expectedCode: ""},
		{
			name: "three nullary alternatives on one data type stay clean",
			source: `module result.duplicate_payload_type_nullary_companion

export {
  type Outcome
  fn identity
}

data Outcome =
  | First
  | Second
  | Third

fn identity(result: Outcome) -> Outcome {
  match result {
    First => First
    Second => Second
    Third => Third
  }
}
`,
			expectedCode: "",
		},
		{
			name: "one Buffer payload in each of two different data types stays clean",
			source: `module result.duplicate_payload_type_cross_type_companion

export {
  type Left
  type Right
  fn identityLeft
  fn identityRight
}

data Left =
  | HoldsLeft(Buffer)

data Right =
  | HoldsRight(Buffer)

fn identityLeft(l: Left) -> Left {
  match l {
    HoldsLeft(v) => HoldsLeft(v)
  }
}

fn identityRight(r: Right) -> Right {
  match r {
    HoldsRight(v) => HoldsRight(v)
  }
}
`,
			expectedCode: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var source []byte
			if tc.fixture != "" {
				source = readPhase12Fixture(t, tc.fixture)
			} else {
				source = []byte(tc.source)
			}
			program := mustParseProgram(t, source)
			result := Program(program)

			if tc.expectedCode == "" {
				duplicateCount := 0
				for _, diag := range result.Diagnostics {
					if diag.Code == "check.duplicate_payload_type" {
						duplicateCount++
					}
				}
				if duplicateCount != 0 {
					t.Fatalf("expected zero check.duplicate_payload_type diagnostics, got %d: %+v", duplicateCount, result.Diagnostics)
				}
				return
			}

			if len(result.Diagnostics) == 0 {
				t.Fatalf("expected at least one diagnostic, got none")
			}
			if result.Diagnostics[0].Code != tc.expectedCode {
				t.Fatalf("expected first diagnostic code %q, got %q (%+v)", tc.expectedCode, result.Diagnostics[0].Code, result.Diagnostics[0])
			}
		})
	}
}

// TestBinderOnNullaryAlternativeNamesConstructionSide is Phase 12 Plan 06
// Task 3's control for WR-01: driving the review's own `Empty => Ok(w)`
// worked example (a nullary pattern alternative matched with a construction
// binder naming a payload source for a different, payload-carrying value
// alternative) still yields check.binder_on_nullary_alternative -- the code
// is unchanged, this is a message-precision fix only -- but the message now
// names the CONSTRUCTION side rather than the pattern side, so a developer
// is sent to the right half of the arm. Asserted on a substring of the new
// wording, not the full message.
func TestBinderOnNullaryAlternativeNamesConstructionSide(t *testing.T) {
	source := []byte(`module result.binder_on_nullary_construction_side

export {
  type Outcome
  fn identity
}

data Outcome =
  | Empty
  | Ok(Buffer)

fn identity(result: Outcome) -> Outcome {
  match result {
    Empty => Ok(w)
    Ok(v) => Ok(v)
  }
}
`)
	program := mustParseProgram(t, source)
	result := Program(program)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected exactly one diagnostic, got %d: %+v", len(result.Diagnostics), result.Diagnostics)
	}
	diag := result.Diagnostics[0]
	if diag.Code != "check.binder_on_nullary_alternative" {
		t.Fatalf("expected check.binder_on_nullary_alternative, got %q (%+v)", diag.Code, diag)
	}
	if !strings.Contains(diag.Message, "construction binder") {
		t.Fatalf("expected message to name the construction side, got %q", diag.Message)
	}
}

// TestPayloadDropObligation is Phase 12 Plan 03 Task 3's own witness for
// criterion 1's affine drop obligation (D-12-29): the checked program
// contains EXACTLY the expected number of destructure_payload operations
// -- one per payload-carrying alternative (Ok, Err), zero for the nullary
// Nope -- proving a lowering that silently skipped either payload arm's
// own obligation would be caught, which an interp-only run (exactly one
// arm executes per invocation) cannot catch on its own. corevalidate's
// own independent re-derivation (replayBlocks' OpConstructPayload/
// OpDestructurePayload cases) is asserted separately from check's own
// verdict, per D-12-29's "derived independently by check and
// corevalidate."
func TestPayloadDropObligation(t *testing.T) {
	source := readPhase12Fixture(t, "payload_drop_obligation.schway")
	program := mustParseProgram(t, source)
	result := Program(program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected zero diagnostics, got %+v", result.Diagnostics)
	}

	validated := corevalidate.Validate(result.Program)
	if !validated.Valid {
		t.Fatalf("corevalidate independently rejected the checked program: %+v", validated.Problems)
	}

	destructureCount := 0
	for _, function := range result.Program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpDestructurePayload {
				destructureCount++
			}
		}
	}
	const expectedDestructureCount = 2
	if destructureCount != expectedDestructureCount {
		t.Fatalf("expected exactly %d destructure_payload operations (one per payload-carrying alternative), got %d", expectedDestructureCount, destructureCount)
	}
}

// TestPayloadBindConsumes is Phase 12 Plan 03 Task 3's mechanical proof of
// D-12-14's "bind MOVES": an in-test core-level mutation (per the plan's
// own "a negative fixture OR an in-test core-level edit" allowance --
// there is no bare-value-arm SOURCE SYNTAX that could author a second read
// of an already-destructured place, so a .schway fixture cannot express
// this). Starting from the tracer's own CHECKED, corevalidate-accepted
// program, the "Ok" arm's OpConstructPayload is mutated to read the
// SCRUTINEE ALIAS PLACE directly (the place the arm's own
// OpDestructurePayload already cleared) instead of the properly
// destructured payload place. If binding merely NAMED the payload rather
// than consuming it, this would still validate; corevalidate must refuse
// it, proving the destructure genuinely moved the alias away.
func TestPayloadBindConsumes(t *testing.T) {
	source := readPhase12Fixture(t, "payload_tracer.schway")
	program := mustParseProgram(t, source)
	result := Program(program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", result.Diagnostics)
	}
	if valid := corevalidate.Validate(result.Program); !valid.Valid {
		t.Fatalf("fixture rejected by corevalidate before mutation: %+v", valid.Problems)
	}

	if len(result.Program.Functions) != 1 {
		t.Fatalf("expected exactly one function, got %d", len(result.Program.Functions))
	}
	function := &result.Program.Functions[0]
	if function.Linear == nil {
		t.Fatal("expected a Linear body on the checked function")
	}

	var aliasOp, constructOp *core.LinearOperation
	for i := range function.Linear.Operations {
		operation := &function.Linear.Operations[i]
		switch operation.Kind {
		case core.OpCopy:
			if aliasOp == nil {
				aliasOp = operation
			}
		case core.OpConstructPayload:
			if constructOp == nil {
				constructOp = operation
			}
		}
	}
	if aliasOp == nil || constructOp == nil {
		t.Fatal("expected to find both an alias OpCopy and an OpConstructPayload in the checked program")
	}

	// The mutation: point the construction at the ALREADY-CLEARED
	// scrutinee alias (aliasOp's own TargetID) instead of the properly
	// destructured payload place, carrying the alias's own TypeID so the
	// generic source.TypeID == operation.TypeID gate passes and the
	// mutation is caught by the initialized[] liveness gate specifically,
	// not by an unrelated type mismatch.
	constructOp.SourceID = aliasOp.TargetID
	constructOp.TypeID = aliasOp.TypeID

	mutated := corevalidate.Validate(result.Program)
	if mutated.Valid {
		t.Fatal("expected corevalidate to refuse constructing from the already-moved scrutinee alias, but it validated")
	}
	if len(mutated.Problems) == 0 {
		t.Fatal("expected at least one problem from the mutated program")
	}
	if mutated.Problems[0].Code != "core.place_uninitialized" {
		t.Fatalf("expected core.place_uninitialized, got %q (%+v)", mutated.Problems[0].Code, mutated.Problems[0])
	}
}
