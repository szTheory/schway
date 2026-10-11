package originvalidate_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase29FixtureOriginForge deep-copies the checked fixture program through
// its canonical JSON form and hands the copy's main to mutate. The checked
// program is never edited in place and no production code is touched.
func phase29FixtureOriginForge(t *testing.T, mutate func(f *core.Function)) core.Program {
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

// TestPhase29FixtureOriginMatchers forges each seeded fixture-shape fault on
// its own and requires originvalidate's own owner/origin matcher to refuse it
// with originvalidate's code (core.fixture_owner_origin). The unforged copy is
// accepted, so each refusal is attributable to the forged field alone.
func TestPhase29FixtureOriginMatchers(t *testing.T) {
	if problems := originvalidate.ValidatePublished(phase29FixtureOriginForge(t, func(*core.Function) {})); len(problems) != 0 {
		t.Fatalf("unforged fixture refused: %+v", problems)
	}
	renumber := func(f *core.Function) {
		for index := range f.Linear.Operations {
			f.Linear.Operations[index].ID = fmt.Sprintf("%s:op:%d", f.ID, index)
			f.Linear.Operations[index].PointID = fmt.Sprintf("%s:point:linear:%d", f.ID, index)
		}
	}
	cases := []struct {
		name  string
		forge func(f *core.Function)
	}{
		{"wrong acquire symbol", func(f *core.Function) {
			f.Linear.Operations[0].Foreign.Symbol = "schway_fixture_bytes_acquire_other"
		}},
		{"wrong acquire abi", func(f *core.Function) {
			f.Linear.Operations[0].Foreign.ABIType = "schway_fixture_bytes_acquire_other_fn"
		}},
		{"missing release", func(f *core.Function) {
			ops := f.Linear.Operations
			f.Linear.Operations = append(append([]core.LinearOperation{}, ops[:3]...), ops[4])
			renumber(f)
		}},
		{"release ordered before the call", func(f *core.Function) {
			ops := f.Linear.Operations
			ops[2], ops[3] = ops[3], ops[2]
			renumber(f)
		}},
		{"view returned from main", func(f *core.Function) {
			ops := f.Linear.Operations
			ops[4].SourceID = ops[1].TargetID
		}},
		{"second call reusing the moved view", func(f *core.Function) {
			ops := f.Linear.Operations
			second := ops[2]
			second.TargetID = f.ID + ":place:4"
			f.Linear.Places = append(f.Linear.Places, core.Place{ID: second.TargetID, Name: "again", TypeID: ops[2].TypeID})
			f.Linear.Operations = append(append(append([]core.LinearOperation{}, ops[:3]...), second), ops[3:]...)
			renumber(f)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			problems := originvalidate.ValidatePublished(phase29FixtureOriginForge(t, tc.forge))
			if len(problems) == 0 {
				t.Fatal("forged fixture was accepted")
			}
			if problems[0].Code != "core.fixture_owner_origin" {
				t.Fatalf("first problem = %+v, want code core.fixture_owner_origin", problems[0])
			}
		})
	}
}
