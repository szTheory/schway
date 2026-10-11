package check

import (
	"encoding/json"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
)

// phase29ForgeMain deep-copies the checked fixture main through its canonical
// JSON form and hands the copy to mutate. The checked program is never edited in
// place and no production code is touched.
func phase29ForgeMain(t *testing.T, mutate func(f *core.Function)) core.Function {
	t.Helper()
	program, codes := phase29CheckSource(t, []byte(phase29FixtureSource(t)))
	if len(codes) != 0 {
		t.Fatalf("fixture diagnostics: %v", codes)
	}
	raw, err := json.Marshal(phase29FixtureMain(t, program))
	if err != nil {
		t.Fatal(err)
	}
	var function core.Function
	if err := json.Unmarshal(raw, &function); err != nil {
		t.Fatal(err)
	}
	mutate(&function)
	return function
}

// phase29EarlyRelease moves the fixture's OpRelease before its OpCall, keeping
// the operation IDs attached to their operations.
func phase29EarlyRelease(f *core.Function) {
	ops := f.Linear.Operations
	ops[2], ops[3] = ops[3], ops[2]
}

// TestPhase29ReleaseBeforeViewUseRefused is D-22's seeded fault for check: the
// fixture core with its OpRelease moved before its OpCall is refused with
// check.release_while_view_live; with only that law disabled the observed result
// is the closed-shape refusal instead, so the law is load-bearing.
func TestPhase29ReleaseBeforeViewUseRefused(t *testing.T) {
	if diagnostics := checkFixtureOwnerPostAssembly(phase29ForgeMain(t, func(*core.Function) {})); len(diagnostics) != 0 {
		t.Fatalf("unforged fixture refused: %+v", diagnostics)
	}
	forged := phase29ForgeMain(t, phase29EarlyRelease)
	diagnostics := checkFixtureOwnerPostAssembly(forged)
	if len(diagnostics) == 0 || diagnostics[0].Code != "check.release_while_view_live" {
		t.Fatalf("diagnostics = %+v, want first code check.release_while_view_live", diagnostics)
	}
	causes := map[string]string{}
	for _, cause := range diagnostics[0].Causes {
		causes[cause.Kind] = cause.Detail
	}
	if causes["release"] != forged.Linear.Operations[2].ID || causes["view_read_later"] != forged.Linear.Operations[3].ID {
		t.Fatalf("causes = %v, want the release and the later read operation IDs", causes)
	}
	disableReleaseWhileViewLiveLawForTest = true
	defer func() { disableReleaseWhileViewLiveLawForTest = false }()
	diagnostics = checkFixtureOwnerPostAssembly(forged)
	if len(diagnostics) == 0 || diagnostics[0].Code != "check.fixture_owner_shape_unsupported" {
		t.Fatalf("law disabled: diagnostics = %+v, want the closed-shape code", diagnostics)
	}
}
