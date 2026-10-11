package pathoracle_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase29FixtureForge deep-copies the checked fixture program through its
// canonical JSON form and hands the copy's main to mutate. The checked program
// is never edited in place and no production code is touched.
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

// TestPhase29PathoracleFixtureReplay forges each fixture-owner fault on its own
// and requires the independent replay to refuse it with its own code; the
// unforged copy is accepted, so every refusal is attributable to the forged
// field alone.
func TestPhase29PathoracleFixtureReplay(t *testing.T) {
	if err := pathoracle.ValidateFixtureOwnerPaths(phase29FixtureForge(t, func(*core.Function) {})); err != nil {
		t.Fatalf("unforged fixture refused: %v", err)
	}
	cases := []struct {
		name  string
		forge func(f *core.Function)
		code  string
	}{
		{name: "release ordered before the call", forge: func(f *core.Function) {
			ops := f.Linear.Operations
			ops[2], ops[3] = ops[3], ops[2]
		}, code: "pathoracle.release_while_view_live"}, // was pathoracle.fixture_owner_order before plan 29-09 (D-22); that code is observed with the law toggled off
		{name: "missing release", forge: func(f *core.Function) {
			ops := f.Linear.Operations
			f.Linear.Operations = append(append([]core.LinearOperation{}, ops[:3]...), ops[4])
		}, code: "pathoracle.fixture_owner_terminal"},
		{name: "second release", forge: func(f *core.Function) {
			ops := f.Linear.Operations
			f.Linear.Operations = append(append([]core.LinearOperation{}, ops[:4]...), ops[3], ops[4])
		}, code: "pathoracle.fixture_owner_release"},
		{name: "acquire not sourced from the entry parameter", forge: func(f *core.Function) {
			f.Linear.Operations[0].SourceID = f.Linear.Operations[1].TargetID
		}, code: "pathoracle.fixture_owner_acquire"},
		{name: "release names another acquisition", forge: func(f *core.Function) {
			f.Linear.Operations[3].ReleasesOperationID = f.Linear.Operations[1].ID
		}, code: "pathoracle.fixture_owner_release"},
		{name: "call consumes a place other than the borrowed view", forge: func(f *core.Function) {
			f.Linear.Operations[2].SourceID = f.Linear.Operations[0].TargetID
		}, code: "pathoracle.fixture_owner_call"},
		{name: "borrow from a place that is not the owner", forge: func(f *core.Function) {
			f.Linear.Operations[1].SourceID = f.Parameter.ID
		}, code: "pathoracle.fixture_owner_borrow"},
		{name: "view returned instead of the call result", forge: func(f *core.Function) {
			f.Linear.Operations[4].SourceID = f.Linear.Operations[1].TargetID
		}, code: "pathoracle.release_while_view_live"}, // the returned view is read after the release; was pathoracle.fixture_owner_terminal before plan 29-09 (D-22)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := pathoracle.ValidateFixtureOwnerPaths(phase29FixtureForge(t, tc.forge))
			if err == nil {
				t.Fatal("forged fixture was accepted")
			}
			if !strings.HasPrefix(err.Error(), tc.code+":") {
				t.Fatalf("error = %q, want code %q", err, tc.code)
			}
		})
	}
}
