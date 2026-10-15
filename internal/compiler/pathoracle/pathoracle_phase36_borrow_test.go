package pathoracle

import (
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// A local borrow that ends at a copy does not grant the call result loan
// authority. These witnesses retain both access modes without an owner wrapper.
func TestPhase36ScalarCallLocalBorrowCopies(t *testing.T) {
	for _, mode := range []string{"borrow", "borrow mut"} {
		t.Run(mode, func(t *testing.T) {
			program := phase36LocalBorrowProgram(t, mode)
			callee := phase36PathCallFunction(t, &program, "copy_value")
			if err := validateStraightLinePointerLoans(*callee); err != nil {
				t.Fatalf("local copy failed independent pointer replay: %v", err)
			}
			if err := ValidateLocalOwnerPaths(program); err != nil {
				t.Fatalf("owned scalar result from local %s copy was refused: %v", mode, err)
			}
			if !pathScalarCalleeValueOnly(*callee) {
				t.Fatal("local borrow-copy callee was classified as transferring authority")
			}
		})
	}
}

func TestPhase36ScalarCallLocalBorrowControls(t *testing.T) {
	for _, mode := range []string{"borrow", "borrow mut"} {
		t.Run(mode, func(t *testing.T) {
			baseline := phase36LocalBorrowProgram(t, mode)
			if err := ValidateLocalOwnerPaths(phase36ClonePathProgram(t, baseline)); err != nil {
				t.Fatalf("no-op local borrow-copy baseline refused: %v", err)
			}
			for _, row := range []struct {
				name string
				edit func(*core.Function)
			}{
				{"borrowed return", func(f *core.Function) { f.Linear.Operations[2].SourceID = f.Linear.Operations[0].TargetID }},
				{"declared provenance", func(f *core.Function) { f.PublicOrigin = &core.PublicOrigin{Access: "shared"} }},
				{"missing loan", func(f *core.Function) { f.Linear.Operations[0].LoanID = "" }},
				{"loan on copy", func(f *core.Function) { f.Linear.Operations[1].LoanID = "forged" }},
				{"loan on return", func(f *core.Function) { f.Linear.Operations[2].LoanID = "forged" }},
				{"nested call", func(f *core.Function) {
					f.Linear.Operations[1].Kind = core.OpCall
					f.Linear.Operations[1].CalleeID = f.ID
				}},
				{"CFG body", func(f *core.Function) {
					f.Linear.Blocks = []core.Block{{ID: "block:entry", PointID: f.EntryPointID, OperationIDs: []string{f.Linear.Operations[0].ID, f.Linear.Operations[1].ID, f.Linear.Operations[2].ID}}}
				}},
				{"backedge", func(f *core.Function) {
					f.Linear.Edges = []core.Edge{{ID: "edge:back", FromBlockID: "block:entry", ToBlockID: "block:entry"}}
				}},
				{"resource place", func(f *core.Function) { f.Linear.Types[0].Shape.Constructor = "FileByteOwner" }},
				{"missing copy", func(f *core.Function) {
					f.Linear.Operations = append(f.Linear.Operations[:1], f.Linear.Operations[2:]...)
				}},
			} {
				t.Run(row.name, func(t *testing.T) {
					mutant := phase36ClonePathProgram(t, baseline)
					row.edit(phase36PathCallFunction(t, &mutant, "copy_value"))
					phase36AssertLocalBorrowContractRefused(t, mutant)
				})
			}
			// The matched three-operation shape is insufficient if any operation
			// smuggles resource/error authority into the otherwise owned result.
			for _, field := range []struct {
				name string
				edit func(*core.LinearOperation)
			}{
				{"callee", func(op *core.LinearOperation) { op.CalleeID = "forged-callee" }},
				{"foreign", func(op *core.LinearOperation) { op.Foreign = &core.ForeignOperationContract{Mode: "acquire"} }},
				{"allocator", func(op *core.LinearOperation) { op.Allocator = "libc_malloc" }},
				{"release", func(op *core.LinearOperation) { op.ReleasesOperationID = "forged-acquire" }},
				{"error target", func(op *core.LinearOperation) { op.ErrTargetID = "forged-error" }},
				{"ok edge", func(op *core.LinearOperation) { op.OkEdgeID = "forged-ok" }},
				{"error edge", func(op *core.LinearOperation) { op.ErrEdgeID = "forged-err" }},
				{"payload", func(op *core.LinearOperation) { op.PayloadTargetID = "forged-payload" }},
				{"store", func(op *core.LinearOperation) { op.StoreTargetID = "forged-store" }},
			} {
				for index, kind := range []string{"borrow", "copy", "return"} {
					t.Run(kind+"/"+field.name, func(t *testing.T) {
						mutant := phase36ClonePathProgram(t, baseline)
						callee := phase36PathCallFunction(t, &mutant, "copy_value")
						field.edit(&callee.Linear.Operations[index])
						phase36AssertLocalBorrowContractRefused(t, mutant)
					})
				}
			}
			if err := ValidateLocalOwnerPaths(baseline); err != nil {
				t.Fatalf("restored local borrow-copy baseline refused: %v", err)
			}
		})
	}
}

func phase36AssertLocalBorrowContractRefused(t *testing.T, program core.Program) {
	t.Helper()
	if err := validateScalarCallContracts(program); err == nil || !strings.HasPrefix(err.Error(), "pathoracle.scalar_call_context:") {
		t.Fatalf("raw local borrow authority mutation did not fail the repaired scalar contract gate: %v", err)
	}
}

func phase36LocalBorrowProgram(t *testing.T, mode string) core.Program {
	t.Helper()
	source := `module phase36.local_borrow
export { fn main }
fn copy_value(value: U64) -> U64 {
  let borrowed = ` + mode + ` value
  let copied = borrowed
  copied
}
fn main(input: U64) -> U64 {
  let result = copy_value(input)
  result
}
`
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("borrow-copy source parse failed: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("borrow-copy source check failed: %+v", checked.Diagnostics)
	}
	return checked.Program
}
