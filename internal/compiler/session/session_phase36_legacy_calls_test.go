package session_test

import (
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// Each peer receives the raw program directly. A PathToken probe returning
// U64 keeps its fallible ownership contract; its scalar result must not turn
// it into a U64/Bool helper or excuse malformed cleanup/error attribution.
func TestPhase36LegacyFallibleCallRoutes(t *testing.T) {
	baseline := checkedPhase24ErrorProgram(t)
	t.Run("positive", func(t *testing.T) { assertErrorPeers(t, baseline, true) })
	t.Run("no-op", func(t *testing.T) { assertErrorPeers(t, cloneTransferProgram(baseline), true) })
	for _, row := range []struct {
		name string
		edit func(*core.LinearOperation)
	}{
		{"missing-error-target", func(op *core.LinearOperation) { op.ErrTargetID = "" }},
		{"missing-ok-edge", func(op *core.LinearOperation) { op.OkEdgeID = "" }},
		{"missing-error-edge", func(op *core.LinearOperation) { op.ErrEdgeID = "" }},
		{"scalar-argument-for-path-probe", func(op *core.LinearOperation) { op.SourceID = op.TargetID }},
	} {
		t.Run(row.name, func(t *testing.T) {
			mutant := cloneTransferProgram(baseline)
			main := phase24Named(&mutant, "main")
			call := phase24Calls(main, phase24Named(&mutant, "probe").ID)[0]
			row.edit(&main.Linear.Operations[phase24OperationIndex(main, call.ID)])
			assertErrorPeers(t, mutant, false)
			assertErrorPeers(t, baseline, true)
		})
	}
	t.Run("wrong-owner-cleanup", func(t *testing.T) {
		mutant := cloneTransferProgram(baseline)
		main := phase24Named(&mutant, "main")
		probeCall := phase24Calls(main, phase24Named(&mutant, "probe").ID)[0]
		block := phase24EdgeBlock(main, probeCall.ErrEdgeID)
		main.Linear.Operations[phase24OperationIndex(main, block.OperationIDs[0])].ReleasesOperationID = "forged-owner"
		assertErrorPeers(t, mutant, false)
		assertErrorPeers(t, baseline, true)
	})
}

func TestPhase36LegacyFileOwnerPeerRoutes(t *testing.T) {
	for _, name := range []string{"checksum.schway", "adler_low.schway"} {
		t.Run(name, func(t *testing.T) {
			checked, err := session.CheckFile(testsupport.ProjectPath("examples", "checksum", name))
			if err != nil || len(checked.Diagnostics) != 0 {
				t.Fatalf("file-owner source check: %v %+v", err, checked.Diagnostics)
			}
			baseline := checked.Program
			assertErrorPeers(t, baseline, true)
			assertErrorPeers(t, cloneTransferProgram(baseline), true)
			mutant := cloneTransferProgram(baseline)
			main := phase24Named(&mutant, "main")
			// Preserve the old lifecycle refusal for a wrong trailing owner
			// argument, or a wrong release source in the five-operation row.
			if name == "adler_low.schway" {
				main.Linear.Operations[4].SourceID = main.Linear.Operations[0].TargetID
			} else {
				main.Linear.Operations[3].SourceID = main.Linear.Operations[1].TargetID
			}
			assertErrorPeers(t, mutant, false)
			problems := originvalidate.ValidatePublished(mutant)
			if len(problems) == 0 || !strings.HasPrefix(problems[0].Code, "core.file_bytes_owner_origin") {
				t.Fatalf("owner lifecycle refusal was masked: %+v", problems)
			}
			assertErrorPeers(t, baseline, true)
		})
	}
}
