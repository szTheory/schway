package corevalidate_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase29FixtureForge deep-copies the checked fixture program through its
// canonical JSON form and hands the copy's main to mutate. The checked
// program is never edited in place and no production code is touched.
func phase29FixtureForge(t *testing.T, mutate func(f *core.Function)) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase29", "adler32_fixture.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture diagnostics: %+v", checked.Diagnostics)
	}
	raw, err := json.Marshal(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	var program core.Program
	if err := json.Unmarshal(raw, &program); err != nil {
		t.Fatal(err)
	}
	for i := range program.Functions {
		if program.Functions[i].Name == "main" {
			mutate(&program.Functions[i])
		}
	}
	return program
}

// TestPhase29FixtureLifecycleReplay forges each seeded lifecycle fault on its
// own and requires corevalidate to refuse it with its own first code; the
// unforged copy is accepted, so each refusal is attributable to the forged
// field alone. Some faults are caught by the generic replay laws before the
// fixture lifecycle matcher runs; the matcher itself is pinned directly by
// TestPhase29FixtureLifecycleShapeDirect (internal test).
func TestPhase29FixtureLifecycleReplay(t *testing.T) {
	if result := corevalidate.Validate(phase29FixtureForge(t, func(*core.Function) {})); !result.Valid {
		t.Fatalf("unforged fixture refused: %+v", result.Problems)
	}
	// renumber keeps operation and point IDs in position order so the generic
	// operation-order law passes and the fixture matchers see the forged shape.
	renumber := func(f *core.Function) {
		for index := range f.Linear.Operations {
			f.Linear.Operations[index].ID = fmt.Sprintf("%s:op:%d", f.ID, index)
			f.Linear.Operations[index].PointID = fmt.Sprintf("%s:point:linear:%d", f.ID, index)
		}
	}
	cases := []struct {
		name  string
		forge func(f *core.Function)
		code  string
	}{
		{"wrong acquire symbol", func(f *core.Function) {
			f.Linear.Operations[0].Foreign.Symbol = "schway_fixture_bytes_acquire_other"
		}, "foreign.operation_contract_invalid"},
		{"wrong acquire abi", func(f *core.Function) {
			f.Linear.Operations[0].Foreign.ABIType = "schway_fixture_bytes_acquire_other_fn"
		}, "foreign.operation_contract_invalid"},
		{"missing release", func(f *core.Function) {
			ops := f.Linear.Operations
			f.Linear.Operations = append(append([]core.LinearOperation{}, ops[:3]...), ops[4])
			renumber(f)
		}, "core.callee_frame_not_drained"},
		{"release ordered before the call", func(f *core.Function) {
			ops := f.Linear.Operations
			ops[2], ops[3] = ops[3], ops[2]
			renumber(f)
		}, "core.release_while_view_live"}, // was core.invalid_target before plan 29-09 (D-22); that code is observed with the law toggled off, see corevalidate_phase29_release_test.go
		{"view returned from main", func(f *core.Function) {
			ops := f.Linear.Operations
			ops[4].SourceID = ops[1].TargetID
		}, "core.release_while_view_live"}, // the returned view is read after the release; was core.final_claim_mismatch before plan 29-09 (D-22)
		{"second call reusing the moved view", func(f *core.Function) {
			ops := f.Linear.Operations
			second := ops[2]
			second.TargetID = f.ID + ":place:4"
			f.Linear.Places = append(f.Linear.Places, core.Place{ID: second.TargetID, Name: "again", TypeID: ops[2].TypeID})
			f.Linear.Operations = append(append(append([]core.LinearOperation{}, ops[:3]...), second), ops[3:]...)
			renumber(f)
		}, "core.place_uninitialized"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := corevalidate.Validate(phase29FixtureForge(t, tc.forge))
			if result.Valid {
				t.Fatal("forged fixture was accepted")
			}
			if len(result.Problems) == 0 || result.Problems[0].Code != tc.code {
				t.Fatalf("problems = %+v, want first code %q", result.Problems, tc.code)
			}
		})
	}
}
