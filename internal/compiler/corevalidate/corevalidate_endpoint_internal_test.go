package corevalidate

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestValidatorRecomputesLoanEndpoints is 03-04-01's algorithmic falsifier
// for recomputeLoanEndpoints, exercised directly against the unexported
// validator method (package corevalidate, not corevalidate_test) — exactly
// as check_test.go tests loanLivenessFixpoint/materializeLoanEndpoints
// directly rather than only through a full check.Program(...) round trip.
// recomputeLoanEndpoints is unreachable from the external test package, and
// the full Validate() pipeline's replayBlocks enforces "exactly one Return
// per non-empty block" — a constraint the general multi-successor
// divergence shape below deliberately does not satisfy (its "entry" block
// both carries an operation and has two successors, matching check.go's own
// TestEdgeSpecificLiveOut CFG exactly, not any shape checkBranch emits
// today). This proves the RECOMPUTATION mechanism itself, independent of
// whatever real topology the checker happens to produce.
func TestValidatorRecomputesLoanEndpoints(t *testing.T) {
	t.Run("straight-line body has no CFG endpoints to recompute", func(t *testing.T) {
		function := straightLineOnlyFunction()
		v := &validator{}
		if got := v.recomputeLoanEndpoints(&function); got != nil {
			t.Fatalf("straight-line body (no Blocks) recomputed non-nil endpoints: %+v", got)
		}
	})

	t.Run("single-successor block: point endpoint where the loan is actually consumed", func(t *testing.T) {
		function := edgeDivergenceFunction()
		v := &validator{}
		endpoints := v.recomputeLoanEndpoints(&function)
		want := wantEdgeDivergenceEndpoints(function)
		if !reflect.DeepEqual(endpoints, want) {
			t.Fatalf("recomputed endpoints mismatch:\ngot=%+v\nwant=%+v", endpoints, want)
		}
	})

	t.Run("counted work grows with newly visited places", func(t *testing.T) {
		function := edgeDivergenceFunction()
		v := &validator{}
		v.recomputeLoanEndpoints(&function)
		if v.checks == 0 {
			t.Fatal("recomputeLoanEndpoints performed no counted work")
		}
	})

	t.Run("a loan never referenced after its own creation ends at its own birth", func(t *testing.T) {
		function := neverReferencedLoanFunction()
		v := &validator{}
		endpoints := v.recomputeLoanEndpoints(&function)
		want := []core.LoanEndpoint{{
			ID:               function.ID + ":point:" + function.ID + ":block:only:0:" + function.ID + ":loan:0",
			LoanID:           function.ID + ":loan:0",
			Kind:             "point",
			BlockID:          function.ID + ":block:only",
			AfterOperationID: function.ID + ":op:0",
		}}
		if !reflect.DeepEqual(endpoints, want) {
			t.Fatalf("never-referenced loan endpoint mismatch:\ngot=%+v\nwant=%+v", endpoints, want)
		}
	})
}

// TestValidatorImportsStayIndependent asserts corevalidate's own Go import
// list (parsed from source, not assumed from a doc comment) contains
// neither "compiler/check" nor "compiler/ast" — T-03-13's import-independence
// falsifier, in the style of the existing TestArbitraryMaskCannotEnterCoreValidation
// architectural-invariant test.
func TestValidatorImportsStayIndependent(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "corevalidate")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			if strings.HasSuffix(path, "/compiler/check") || strings.HasSuffix(path, "/compiler/ast") {
				t.Fatalf("%s imports %s, which corevalidate must never depend on", entry.Name(), path)
			}
		}
	}
}

func straightLineOnlyFunction() core.Function {
	functionID := "s1:straight:fn:relay"
	typeID := functionID + ":type:0"
	parameterID := functionID + ":place:0"
	targetID := functionID + ":place:1"
	return core.Function{
		ID: functionID, Name: "relay",
		EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter: core.Parameter{ID: parameterID, Name: "source", Type: "Byte"}, ReturnType: "Byte",
		Linear: &core.LinearBody{
			ID: functionID + ":linear",
			Types: []core.TypeFact{{
				ID: typeID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
				Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
				NegativeWitnesses: []core.AbilityWitness{},
			}},
			Places: []core.Place{{ID: parameterID, Name: "source", TypeID: typeID}, {ID: targetID, Name: "copy", TypeID: typeID}},
			Operations: []core.LinearOperation{
				{ID: functionID + ":op:0", PointID: functionID + ":point:linear:0", Kind: core.OpCopy, SourceID: parameterID, TargetID: targetID, TypeID: typeID},
				{ID: functionID + ":op:1", PointID: functionID + ":point:linear:1", Kind: core.OpReturn, SourceID: parameterID, TypeID: typeID},
			},
		},
	}
}

// edgeDivergenceFunction mirrors check.go's TestEdgeSpecificLiveOut exactly:
// a loan is born in "entry" (which fans out to two successors), used only
// along "used", and never referenced along "unused" — the general
// multi-successor divergence case, unreachable from any real
// checkBranch-produced fixture this phase (D-10) but proven directly here.
func edgeDivergenceFunction() core.Function {
	functionID := "s1:edge:fn:relay"
	typeID := functionID + ":type:0"
	ownerID := functionID + ":place:0"
	viewID := functionID + ":place:1"
	loanID := functionID + ":loan:0"
	entryID := functionID + ":block:entry"
	usedID := functionID + ":block:used"
	unusedID := functionID + ":block:unused"
	return core.Function{
		ID: functionID, Name: "relay",
		EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter: core.Parameter{ID: ownerID, Name: "owner", Type: "Byte"}, ReturnType: "Byte",
		Linear: &core.LinearBody{
			ID: functionID + ":linear",
			Types: []core.TypeFact{{
				ID: typeID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
				Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
				NegativeWitnesses: []core.AbilityWitness{},
			}},
			Places: []core.Place{
				{ID: ownerID, Name: "owner", TypeID: typeID},
				{ID: viewID, Name: "view", TypeID: typeID},
			},
			Operations: []core.LinearOperation{
				{ID: functionID + ":op:0", PointID: functionID + ":point:linear:0", Kind: core.OpBorrowShared, SourceID: ownerID, TargetID: viewID, LoanID: loanID, TypeID: typeID},
				{ID: functionID + ":op:1", PointID: functionID + ":point:linear:1", Kind: core.OpReturn, SourceID: viewID, TypeID: typeID},
				{ID: functionID + ":op:2", PointID: functionID + ":point:linear:2", Kind: core.OpReturn, SourceID: ownerID, TypeID: typeID},
			},
			Blocks: []core.Block{
				{ID: entryID, PointID: functionID + ":point:entry", OperationIDs: []string{functionID + ":op:0"}, Successors: []string{usedID, unusedID}},
				{ID: usedID, PointID: functionID + ":point:used", OperationIDs: []string{functionID + ":op:1"}, Successors: nil},
				{ID: unusedID, PointID: functionID + ":point:unused", OperationIDs: []string{functionID + ":op:2"}, Successors: nil},
			},
			Edges: []core.Edge{
				{ID: functionID + ":edge:entry:used", FromBlockID: entryID, ToBlockID: usedID, Pattern: "used"},
				{ID: functionID + ":edge:entry:unused", FromBlockID: entryID, ToBlockID: unusedID, Pattern: "unused"},
			},
		},
	}
}

func wantEdgeDivergenceEndpoints(function core.Function) []core.LoanEndpoint {
	loanID := function.ID + ":loan:0"
	usedID := function.ID + ":block:used"
	edgeUnusedID := function.ID + ":edge:entry:unused"
	endpoints := []core.LoanEndpoint{
		{ID: edgeUnusedID + ":" + loanID, LoanID: loanID, Kind: "edge", EdgeID: edgeUnusedID},
		{
			ID: function.ID + ":point:" + usedID + ":0:" + loanID, LoanID: loanID, Kind: "point",
			BlockID: usedID, AfterOperationID: function.ID + ":op:1",
		},
	}
	if endpoints[0].ID > endpoints[1].ID {
		endpoints[0], endpoints[1] = endpoints[1], endpoints[0]
	}
	return endpoints
}

// neverReferencedLoanFunction is a loan created and never referenced
// again — the degenerate fallback recomputeLoanEndpoints must still cover
// (mirrors materializeLoanEndpoints' own `!recorded[operation.LoanID]`
// branch on the checker side): its endpoint is its own birth operation.
func neverReferencedLoanFunction() core.Function {
	functionID := "s1:unused:fn:relay"
	typeID := functionID + ":type:0"
	ownerID := functionID + ":place:0"
	viewID := functionID + ":place:1"
	loanID := functionID + ":loan:0"
	onlyID := functionID + ":block:only"
	return core.Function{
		ID: functionID, Name: "relay",
		EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter: core.Parameter{ID: ownerID, Name: "owner", Type: "Byte"}, ReturnType: "Byte",
		Linear: &core.LinearBody{
			ID: functionID + ":linear",
			Types: []core.TypeFact{{
				ID: typeID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
				Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
				NegativeWitnesses: []core.AbilityWitness{},
			}},
			Places: []core.Place{
				{ID: ownerID, Name: "owner", TypeID: typeID},
				{ID: viewID, Name: "view", TypeID: typeID},
			},
			Operations: []core.LinearOperation{
				{ID: functionID + ":op:0", PointID: functionID + ":point:linear:0", Kind: core.OpBorrowShared, SourceID: ownerID, TargetID: viewID, LoanID: loanID, TypeID: typeID},
				{ID: functionID + ":op:1", PointID: functionID + ":point:linear:1", Kind: core.OpReturn, SourceID: ownerID, TypeID: typeID},
			},
			Blocks: []core.Block{
				{ID: onlyID, PointID: functionID + ":point:only", OperationIDs: []string{functionID + ":op:0", functionID + ":op:1"}, Successors: nil},
			},
		},
	}
}
