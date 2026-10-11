package corevalidate_test

import (
	"fmt"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
)

// phase29Renumber keeps operation and point IDs in position order so the
// generic operation-order law passes and the forged shape reaches replay.
func phase29Renumber(f *core.Function) {
	for index := range f.Linear.Operations {
		f.Linear.Operations[index].ID = fmt.Sprintf("%s:op:%d", f.ID, index)
		f.Linear.Operations[index].PointID = fmt.Sprintf("%s:point:linear:%d", f.ID, index)
	}
}

// phase29EarlyRelease moves the fixture's OpRelease before its OpCall. The
// release keeps naming the acquire's operation ID (op:0), which renumbering
// leaves unchanged.
func phase29EarlyRelease(f *core.Function) {
	ops := f.Linear.Operations
	ops[2], ops[3] = ops[3], ops[2]
	phase29Renumber(f)
}

// TestPhase29ReleaseBeforeViewUseRefused is D-22's seeded fault for corevalidate:
// the fixture core with its OpRelease before its OpCall is refused with
// core.release_while_view_live; with only that law disabled the observed first
// code is the generic replay refusal core.invalid_target, so the law is
// load-bearing.
func TestPhase29ReleaseBeforeViewUseRefused(t *testing.T) {
	if result := corevalidate.Validate(phase29FixtureForge(t, func(*core.Function) {})); !result.Valid {
		t.Fatalf("unforged fixture refused: %+v", result.Problems)
	}
	result := corevalidate.Validate(phase29FixtureForge(t, phase29EarlyRelease))
	if result.Valid || len(result.Problems) == 0 || result.Problems[0].Code != "core.release_while_view_live" {
		t.Fatalf("problems = %+v, want first code core.release_while_view_live", result.Problems)
	}
	defer corevalidate.SetDisableReleaseWhileViewLiveLawForTest(true)()
	result = corevalidate.Validate(phase29FixtureForge(t, phase29EarlyRelease))
	if result.Valid || len(result.Problems) == 0 || result.Problems[0].Code != "core.invalid_target" {
		t.Fatalf("law disabled: problems = %+v, want first code core.invalid_target", result.Problems)
	}
}
