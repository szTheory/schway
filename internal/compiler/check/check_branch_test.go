package check_test

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
)

const branchSource = `module owned.branch_view

export {
  type Switch
  fn direction
}

data Switch =
  | On
  | Off

fn direction(flag: Switch) -> Switch {
  match flag {
    On => {
      let held = take flag
      held
    }
    Off => {
      let held = take flag
      held
    }
  }
}
`

// TestArmBodyLowersToBlocksAndEdges is 03-01-02's structural falsifier: a
// match arm's value position holding a full linear body must yield a typed
// core with more than one block and a real join (OWN-03 success criterion
// 2), while the flat Operations list and every existing order/uniqueness
// invariant stay intact.
func TestArmBodyLowersToBlocksAndEdges(t *testing.T) {
	checked := session.Check([]byte(branchSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) != 1 {
		t.Fatalf("want one function, got %d", len(checked.Program.Functions))
	}
	function := checked.Program.Functions[0]
	if function.Match == nil || function.Linear == nil {
		t.Fatalf("branch function must carry both Match and Linear: %+v", function)
	}
	blocks := function.Linear.Blocks
	if len(blocks) < 4 {
		t.Fatalf("want at least entry + 2 arm blocks + join, got %d blocks: %+v", len(blocks), blocks)
	}
	var entry, join core.Block
	armBlocks := 0
	for _, block := range blocks {
		switch block.ID {
		case function.ID + ":block:entry":
			entry = block
		case function.ID + ":block:join":
			join = block
		default:
			armBlocks++
		}
	}
	if entry.ID == "" || join.ID == "" {
		t.Fatalf("missing entry or join block: %+v", blocks)
	}
	if armBlocks != 2 {
		t.Fatalf("want 2 arm blocks, got %d", armBlocks)
	}
	if len(entry.Successors) != 2 {
		t.Fatalf("entry block must have one successor per arm: %+v", entry)
	}
	if len(join.Successors) != 0 {
		t.Fatalf("join block must be terminal: %+v", join)
	}
	// A real join: both arm blocks' sole successor is the same join block.
	joinTargets := 0
	for _, block := range blocks {
		if block.ID == entry.ID || block.ID == join.ID {
			continue
		}
		if len(block.Successors) != 1 || block.Successors[0] != join.ID {
			t.Fatalf("arm block %q does not join at %q: %+v", block.ID, join.ID, block)
		}
		if len(block.OperationIDs) == 0 {
			t.Fatalf("arm block %q carries no operations", block.ID)
		}
		joinTargets++
	}
	if joinTargets != 2 {
		t.Fatalf("want 2 arm blocks joining, got %d", joinTargets)
	}
	if len(function.Linear.Edges) != 4 {
		t.Fatalf("want 4 edges (entry->arm, arm->join per arm), got %d: %+v", len(function.Linear.Edges), function.Linear.Edges)
	}
	// The existing flat Operations authority is unchanged in shape: still a
	// plain ordered slice, referenced by ID from the block records.
	if len(function.Linear.Operations) == 0 {
		t.Fatalf("expected flat operations list to be populated")
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("independent validator rejected a valid branch program: %+v", validated.Problems)
	}
}

// TestArmBodySchemaCrossLock proves the extended body/schema cross-lock
// (03-PATTERNS I-7, corevalidate.go:106/109 extended): a match carrying any
// arm body requires lang.core/1, an all-bare-arm match still requires
// lang.core/0, and the reverse pairing is rejected fail-closed by the
// independent validator, both directions.
func TestArmBodySchemaCrossLock(t *testing.T) {
	checked := session.Check([]byte(branchSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}
	if checked.Program.Schema != core.Schema1 {
		t.Fatalf("branch program schema = %q, want %q", checked.Program.Schema, core.Schema1)
	}

	bareSource := `module owned.bare_switch

export {
  type Switch
  fn toggle
}

data Switch =
  | On
  | Off

fn toggle(flag: Switch) -> Switch {
  match flag {
    On => Off
    Off => On
  }
}
`
	bareChecked := session.Check([]byte(bareSource))
	if len(bareChecked.Diagnostics) != 0 {
		t.Fatalf("bare diagnostics: %+v", bareChecked.Diagnostics)
	}
	if bareChecked.Program.Schema != core.Schema {
		t.Fatalf("bare-arm program schema = %q, want %q (Phase 1 byte-identical)", bareChecked.Program.Schema, core.Schema)
	}
	if bareChecked.Program.Functions[0].Linear != nil {
		t.Fatalf("bare-arm match unexpectedly carries a Linear body: %+v", bareChecked.Program.Functions[0])
	}

	// Reverse pairing 1: a bare-arm-only program forcibly relabeled /1 is
	// rejected — match body without any arm body requires exactly /0.
	mislabeledBare := bareChecked.Program
	mislabeledBare.Schema = core.Schema1
	if result := corevalidate.Validate(mislabeledBare); result.Valid {
		t.Fatalf("validator admitted a bare-arm match mislabeled lang.core/1")
	}

	// Reverse pairing 2: the branch program forcibly relabeled /0 is
	// rejected — a match with an arm body requires exactly /1.
	mislabeledBranch := checked.Program
	mislabeledBranch.Schema = core.Schema
	if result := corevalidate.Validate(mislabeledBranch); result.Valid {
		t.Fatalf("validator admitted an arm-body match mislabeled lang.core/0")
	}
}

