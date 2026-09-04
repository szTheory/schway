package corevalidate_test

import (
	"encoding/json"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
)

const branchTestSource = `module owned.branch_view

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

func validBranchProgram(t *testing.T) core.Program {
	t.Helper()
	checked := session.Check([]byte(branchTestSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}
	if result := corevalidate.Validate(checked.Program); !result.Valid {
		t.Fatalf("baseline branch program rejected: %+v", result.Problems)
	}
	return cloneCoreProgram(t, checked.Program)
}

func cloneCoreProgram(t *testing.T, program core.Program) core.Program {
	t.Helper()
	encoded, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	var clone core.Program
	if err := json.Unmarshal(encoded, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

// TestBlockEdgeValidationRules is 03-01-03's independent-validator falsifier
// for T-03-01/T-03-04: the new Blocks/Edges slices get the same per-slice
// order/uniqueness/referential-closure treatment as every other ordinal-
// bearing fact (core.type_order/core.place_order/core.operation_order and
// their `v.unique*` family). Each mutation below is otherwise-valid and
// rejected by exactly one new code.
func TestBlockEdgeValidationRules(t *testing.T) {
	tests := []struct {
		name string
		code string
		edit func(*core.Program)
	}{
		{
			name: "duplicate block id",
			code: "core.duplicate_block_id",
			edit: func(program *core.Program) {
				linear := program.Functions[0].Linear
				linear.Blocks[1].ID = linear.Blocks[0].ID
			},
		},
		{
			name: "duplicate edge id",
			code: "core.duplicate_edge_id",
			edit: func(program *core.Program) {
				linear := program.Functions[0].Linear
				linear.Edges[1].ID = linear.Edges[0].ID
			},
		},
		{
			name: "block references unknown operation",
			code: "core.unknown_operation_reference",
			edit: func(program *core.Program) {
				linear := program.Functions[0].Linear
				linear.Blocks[1].OperationIDs[0] = "not-a-real-operation-id"
			},
		},
		{
			name: "operation claimed by no block",
			code: "core.unknown_block",
			edit: func(program *core.Program) {
				linear := program.Functions[0].Linear
				linear.Blocks[1].OperationIDs = linear.Blocks[1].OperationIDs[1:]
			},
		},
		{
			name: "operation claimed by two blocks",
			code: "core.duplicate_operation_id",
			edit: func(program *core.Program) {
				linear := program.Functions[0].Linear
				linear.Blocks[2].OperationIDs = append(linear.Blocks[2].OperationIDs, linear.Blocks[1].OperationIDs[0])
			},
		},
		{
			name: "edge references unknown from block",
			code: "core.unknown_block",
			edit: func(program *core.Program) {
				linear := program.Functions[0].Linear
				linear.Edges[0].FromBlockID = "not-a-real-block-id"
			},
		},
		{
			name: "edge references unknown to block",
			code: "core.unknown_block",
			edit: func(program *core.Program) {
				linear := program.Functions[0].Linear
				linear.Edges[0].ToBlockID = "not-a-real-block-id"
			},
		},
		{
			name: "block successor references unknown block",
			code: "core.unknown_block",
			edit: func(program *core.Program) {
				linear := program.Functions[0].Linear
				linear.Blocks[0].Successors[0] = "not-a-real-block-id"
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program := validBranchProgram(t)
			test.edit(&program)
			result := corevalidate.Validate(program)
			if result.Valid {
				t.Fatalf("mutation %q was accepted", test.name)
			}
			if len(result.Problems) == 0 || result.Problems[0].Code != test.code {
				t.Fatalf("mutation %q code=%v want=%s", test.name, result.Problems, test.code)
			}
		})
	}
}

// TestBranchReturnMustBeLastInBlock confirms a branch-shaped function still
// requires exactly one OpReturn per block that carries operations, keyed to
// that block rather than to the whole flat operations list — the exact
// change replayBlocks makes to the pre-existing straight-line
// return-uniqueness rule.
//
// Renamed from this test's original name (03-01's TestLoanEndpointMutationMatrix)
// in 03-04 (deviation, see 03-04-SUMMARY.md): the name promised a loan-endpoint
// mutation matrix, but the test never touched a LoanEndpoint at all — it is an
// OpReturn-ordering falsifier. 03-04-03 needs the exact name
// TestLoanEndpointMutationMatrix for the real endpoint mutation matrix
// (moved/dropped/invented, T-03-06/T-03-14's actual required control), and Go
// forbids two functions of the same name in one package.
func TestBranchReturnMustBeLastInBlock(t *testing.T) {
	program := validBranchProgram(t)
	linear := program.Functions[0].Linear
	// Move the return operation out of order within its own block: swap the
	// two operation IDs at the tail of the first arm block so its Return no
	// longer sits last.
	first := linear.Blocks[1].OperationIDs
	if len(first) < 2 {
		t.Fatalf("expected at least two operations in the first arm block: %+v", first)
	}
	first[len(first)-1], first[len(first)-2] = first[len(first)-2], first[len(first)-1]
	result := corevalidate.Validate(program)
	if result.Valid {
		t.Fatal("a return operation not last in its own block was accepted")
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != "core.final_claim_mismatch" {
		t.Fatalf("code=%v want=core.final_claim_mismatch", result.Problems)
	}
}
