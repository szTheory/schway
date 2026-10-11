package corevalidate

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func phase29FixtureInternalProgram(t *testing.T) (core.Program, int) {
	t.Helper()
	source, err := os.ReadFile("../../../examples/phase29/adler32_fixture.schway")
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check: %+v", checked.Diagnostics)
	}
	raw, err := json.Marshal(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	var program core.Program
	if err := json.Unmarshal(raw, &program); err != nil {
		t.Fatal(err)
	}
	for index := range program.Functions {
		if program.Functions[index].Name == "main" {
			return program, index
		}
	}
	t.Fatal("no main")
	return core.Program{}, 0
}

// TestPhase29FixtureLifecycleShapeDirect pins fixtureOwnerLifecycleValid on
// its own, so the matcher is exercised even for forged shapes that the generic
// replay laws would refuse first: release before the call, missing release,
// the view returned, a second call on the moved view, a call result that is
// not returned, and a release naming the wrong acquisition.
func TestPhase29FixtureLifecycleShapeDirect(t *testing.T) {
	forge := func(mutate func(f *core.Function)) bool {
		program, index := phase29FixtureInternalProgram(t)
		mutate(&program.Functions[index])
		v := &validator{program: program}
		return v.fixtureOwnerLifecycleValid(&program.Functions[index])
	}
	if !forge(func(*core.Function) {}) {
		t.Fatal("unforged fixture lifecycle refused")
	}
	cases := []struct {
		name   string
		mutate func(f *core.Function)
	}{
		{"release before the call", func(f *core.Function) {
			ops := f.Linear.Operations
			ops[2], ops[3] = ops[3], ops[2]
		}},
		{"missing release", func(f *core.Function) {
			ops := f.Linear.Operations
			f.Linear.Operations = append(append([]core.LinearOperation{}, ops[:3]...), ops[4])
		}},
		{"view returned", func(f *core.Function) {
			ops := f.Linear.Operations
			ops[4].SourceID = ops[1].TargetID
		}},
		{"second call reusing the view", func(f *core.Function) {
			ops := f.Linear.Operations
			second := ops[2]
			f.Linear.Operations = append(append(append([]core.LinearOperation{}, ops[:3]...), second), ops[3:]...)
		}},
		{"release names the wrong acquisition", func(f *core.Function) {
			f.Linear.Operations[3].ReleasesOperationID = f.Linear.Operations[1].ID
		}},
		{"acquire not from the parameter", func(f *core.Function) {
			f.Linear.Operations[0].SourceID = f.Linear.Operations[0].TargetID
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if forge(tc.mutate) {
				t.Fatal("forged lifecycle accepted")
			}
		})
	}
}
