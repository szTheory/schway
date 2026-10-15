package corevalidate

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
)

// Call-only scalar analysis belongs to scalar callers. A scalar result alone
// does not turn the existing PathToken ownership/error CFG into a scalar CFG.
// Explicit scalar operations still select it and enforce the authority law.
func TestPhase36CallCFGApplicability(t *testing.T) {
	for _, row := range []struct {
		name, parameter, argument, result string
		scalar                            bool
	}{
		{"u64-u64", "U64", "U64", "U64", true},
		{"u64-bool", "U64", "U64", "Bool", true},
		{"bool-u64", "Bool", "Bool", "U64", true},
		{"bool-bool", "Bool", "Bool", "Bool", true},
		{"byte-wrapper", "Byte", "U64", "U64", true},
		{"view-caller-scalar-helper", "ByteView", "U64", "U64", true},
		{"file-owner-view-call", "PathToken", "ByteView", "U64", false},
		{"file-owner-scalar-tail", "PathToken", "U64", "U64", false},
		{"fallible-path-probe", "PathToken", "PathToken", "U64", false},
		{"owner-transfer", "PathToken", "PathToken", "FileByteOwner", false},
		{"fixture-owner-view-call", "U64", "ByteView", "U64", false},
	} {
		t.Run(row.name, func(t *testing.T) {
			function := core.Function{Parameter: core.Parameter{Type: row.parameter}, Linear: &core.LinearBody{
				Types: []core.TypeFact{
					{ID: "argument", Shape: core.TypeRef{Constructor: row.argument}},
					{ID: "result", Shape: core.TypeRef{Constructor: row.result}},
				},
				Places:     []core.Place{{ID: "source", TypeID: "argument"}, {ID: "target", TypeID: "result"}},
				Operations: []core.LinearOperation{{Kind: core.OpCall, SourceID: "source", TargetID: "target", TypeID: "result"}},
			}}
			if got := hasScalarCFG(&function); got != row.scalar {
				t.Fatalf("call-only scalar CFG = %v, want %v", got, row.scalar)
			}
			function.Linear.Operations = append(function.Linear.Operations, core.LinearOperation{Kind: core.OpBranch})
			if !hasScalarCFG(&function) {
				t.Fatal("explicit scalar branch lost authority validation")
			}
			function.Linear.Operations = nil
			if hasScalarCFG(&function) {
				t.Fatal("empty body selected scalar CFG")
			}
		})
	}
}

// The old five-operation owner route and six-operation scalar-tail route must
// pass without weakening release ordering or allowing another view call.
func TestPhase36LegacyFileOwnerCalls(t *testing.T) {
	for _, row := range []struct{ name, tail string }{
		{"view-result", phase31PeerTail5},
		{"scalar-tail", phase31PeerTail6},
	} {
		t.Run(row.name, func(t *testing.T) {
			program, index := phase31PeerForge(t, row.tail, func(*core.Program, *core.Function) {})
			if result := Validate(program); !result.Valid {
				t.Fatalf("legacy owner call refused: %+v", result.Problems)
			}
			function := &program.Functions[index]
			types := map[string]core.TypeFact{}
			places := map[string]core.Place{}
			for _, fact := range function.Linear.Types {
				types[fact.ID] = fact
			}
			for _, place := range function.Linear.Places {
				places[place.ID] = place
			}
			call := function.Linear.Operations[2]
			var callee core.Function
			for _, candidate := range program.Functions {
				if candidate.ID == call.CalleeID {
					callee = candidate
				}
			}
			// Drive the contract gate alone: a release for the wrong owner must
			// not gain mixed-call admission through the compatibility seam.
			function.Linear.Operations[3].SourceID = call.SourceID
			v := &validator{program: program}
			if v.checkCallTypeContractForCaller(*function, callee, places[call.SourceID], call, places, types) {
				t.Fatal("mixed-call context accepted a forged owner lifecycle")
			}
			if len(v.problems) == 0 || v.problems[0].Code != "core.scalar_call_context_invalid" {
				t.Fatalf("wrong owner control did not reach mixed-call gate: %+v", v.problems)
			}
		})
	}
}
