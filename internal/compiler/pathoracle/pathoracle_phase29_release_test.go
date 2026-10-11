package pathoracle_test

import (
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
)

// phase29EarlyRelease moves the fixture's OpRelease before its OpCall.
func phase29EarlyRelease(f *core.Function) {
	ops := f.Linear.Operations
	ops[2], ops[3] = ops[3], ops[2]
}

// TestPhase29ReleaseBeforeViewUseRefused is D-22's seeded fault for pathoracle:
// the fixture core with its OpRelease before its OpCall is refused with
// pathoracle.release_while_view_live; with only that law disabled the observed
// code is the closed replay's own pathoracle.fixture_owner_order, so the law is
// load-bearing.
func TestPhase29ReleaseBeforeViewUseRefused(t *testing.T) {
	if err := pathoracle.ValidateFixtureOwnerPaths(phase29FixtureForge(t, func(*core.Function) {})); err != nil {
		t.Fatalf("unforged fixture refused: %v", err)
	}
	err := pathoracle.ValidateFixtureOwnerPaths(phase29FixtureForge(t, phase29EarlyRelease))
	if err == nil || !strings.HasPrefix(err.Error(), "pathoracle.release_while_view_live:") {
		t.Fatalf("error = %v, want code pathoracle.release_while_view_live", err)
	}
	defer pathoracle.SetDisableReleaseWhileViewLiveLawForTest(true)()
	err = pathoracle.ValidateFixtureOwnerPaths(phase29FixtureForge(t, phase29EarlyRelease))
	if err == nil || !strings.HasPrefix(err.Error(), "pathoracle.fixture_owner_order:") {
		t.Fatalf("law disabled: error = %v, want code pathoracle.fixture_owner_order", err)
	}
}
