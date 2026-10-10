package cgen_test

import (
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
)

// TestPhase272ScalarClassifierRefusesOwnershipBodies is the D-27.2-04 negative
// guard. A CFG body that carries an ownership-kernel element must keep hitting
// the "multi-function foreign-call bodies" refusal even after the scalar
// classifier learns to accept copy/const/write-only bodies. It uses
// EmitProgramForTest so the hand-built body reaches the classifier without
// corevalidate refusing it first, and so no tracked public emitter is called.
func TestPhase272ScalarClassifierRefusesOwnershipBodies(t *testing.T) {
	build := func(placeType string, middle core.OperationKind) core.Program {
		const fn = "fn:main"
		return core.Program{Functions: []core.Function{{
			ID:            fn,
			Name:          "main",
			EntryPointID:  fn + ":point:entry",
			ReturnPointID: fn + ":point:return",
			Parameter:     core.Parameter{ID: fn + ":param", Name: "n", Type: "U64"},
			ReturnType:    "U64",
			Linear: &core.LinearBody{
				ID: fn + ":linear",
				Types: []core.TypeFact{
					{ID: "type:u64", Shape: core.TypeRef{Constructor: "U64"}},
					{ID: "type:" + placeType, Shape: core.TypeRef{Constructor: placeType}},
				},
				Places: []core.Place{{ID: fn + ":place:1", Name: "held", TypeID: "type:" + placeType}},
				Operations: []core.LinearOperation{
					{ID: fn + ":op:1", PointID: fn + ":point:block:0", Kind: middle, SourceID: fn + ":param", TargetID: fn + ":place:1", TypeID: "type:" + placeType},
					{ID: fn + ":op:2", PointID: fn + ":point:block:0", Kind: core.OpReturn, SourceID: fn + ":place:1", TypeID: "type:u64"},
				},
				Blocks: []core.Block{{ID: fn + ":block:entry", PointID: fn + ":point:block:0", OperationIDs: []string{fn + ":op:1", fn + ":op:2"}}},
			},
		}}}
	}
	const refusal = "multi-function foreign-call bodies are not supported"
	t.Run("owner-place", func(t *testing.T) {
		_, err := cgen.EmitProgramForTest(build("Owner", core.OpCopy))
		if err == nil || !strings.Contains(err.Error(), refusal) {
			t.Fatalf("owner-typed place with blocks: err=%v, want %q", err, refusal)
		}
	})
	t.Run("move-op", func(t *testing.T) {
		_, err := cgen.EmitProgramForTest(build("U64", core.OpMove))
		if err == nil || !strings.Contains(err.Error(), refusal) {
			t.Fatalf("OpMove body with blocks: err=%v, want %q", err, refusal)
		}
	})
}
