package corevalidate

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
)

// phase30CrossModuleProgram hand-builds a two-function core program: an
// entry-module `caller` (Module "") that calls a "lib" function `callee`
// declared Exported with a matching import fact. It does not run check, so a
// test can edit the module facts directly and observe corevalidate alone.
// blocks selects the branch-shaped (replayBlocks) replay site instead of the
// straight-line (replayStraightLine) one.
func phase30CrossModuleProgram(blocks bool) core.Program {
	callerID := "s1:test:fn:caller"
	calleeID := "s1:lib:fn:callee"
	callerType := callerID + ":type:0"
	callerParam := callerID + ":place:0"
	callerTarget := callerID + ":place:1"
	calleeType := calleeID + ":type:0"
	calleeParam := calleeID + ":place:0"
	byteFact := core.TypeFact{
		ID: callerType, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
		Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
		NegativeWitnesses: []core.AbilityWitness{},
	}
	calleeFact := byteFact
	calleeFact.ID = calleeType
	callerLinear := &core.LinearBody{
		ID:     callerID + ":linear",
		Types:  []core.TypeFact{byteFact},
		Places: []core.Place{{ID: callerParam, Name: "value", TypeID: callerType}, {ID: callerTarget, Name: "result", TypeID: callerType}},
		Operations: []core.LinearOperation{
			{ID: callerID + ":op:0", PointID: callerID + ":point:linear:0", Kind: core.OpCall, SourceID: callerParam, TargetID: callerTarget, TypeID: callerType, CalleeID: calleeID},
			{ID: callerID + ":op:1", PointID: callerID + ":point:linear:1", Kind: core.OpReturn, SourceID: callerTarget, TypeID: callerType},
		},
	}
	if blocks {
		entry, done := callerID+":block:entry", callerID+":block:done"
		callerLinear.Blocks = []core.Block{
			{ID: entry, PointID: callerID + ":point:entry", OperationIDs: []string{callerID + ":op:0"}, Successors: []string{done}},
			{ID: done, PointID: callerID + ":point:done", OperationIDs: []string{callerID + ":op:1"}},
		}
		callerLinear.Edges = []core.Edge{{ID: callerID + ":edge:entry:done", FromBlockID: entry, ToBlockID: done, Pattern: "next"}}
	}
	return core.Program{
		Schema: core.Schema1, Module: "test", ModuleID: "s1:test:module:test",
		Imports: []core.ModuleImport{{From: "", Module: "lib", Name: "callee"}},
		Functions: []core.Function{
			{
				ID: callerID, Name: "caller", EntryPointID: callerID + ":point:entry", ReturnPointID: callerID + ":point:return",
				Parameter: core.Parameter{ID: callerParam, Name: "value", Type: "Byte"}, ReturnType: "Byte",
				Linear: callerLinear,
			},
			{
				ID: calleeID, Name: "callee", Module: "lib", Exported: true,
				EntryPointID: calleeID + ":point:entry", ReturnPointID: calleeID + ":point:return",
				Parameter: core.Parameter{ID: calleeParam, Name: "value", Type: "Byte"}, ReturnType: "Byte",
				Linear: &core.LinearBody{
					ID:         calleeID + ":linear",
					Types:      []core.TypeFact{calleeFact},
					Places:     []core.Place{{ID: calleeParam, Name: "value", TypeID: calleeType}},
					Operations: []core.LinearOperation{{ID: calleeID + ":op:0", PointID: calleeID + ":point:linear:0", Kind: core.OpReturn, SourceID: calleeParam, TypeID: calleeType}},
				},
			},
		},
	}
}

func firstProblemCode(result Result) string {
	if len(result.Problems) == 0 {
		return ""
	}
	return result.Problems[0].Code
}

// TestPhase30CorevalidateVisibilityPeer pins D-28's two codes, peers of
// check's module.not_exported, and proves corevalidate alone refuses each
// forged cross-module call at BOTH OpCall replay sites. The control engages
// disableVisibilityPeerForTest and observes the forged programs admitted, so
// it is the law -- not some other check -- that refuses them.
//
// Trust boundary: a coordinated lie (a forged Exported flag together with a
// matching import fact) is admitted by this law by design. That is the
// frontend trust boundary; the Phase 31 module digests are the cross-check.
func TestPhase30CorevalidateVisibilityPeer(t *testing.T) {
	if calleeNotExportedCode != "core.callee_not_exported" || calleeNotImportedCode != "core.callee_not_imported" {
		t.Fatalf("visibility codes drifted: %q %q", calleeNotExportedCode, calleeNotImportedCode)
	}
	type row struct {
		name   string
		mutate func(p *core.Program)
		want   string // "" means admitted
	}
	rows := []row{
		{"valid cross-module import", func(p *core.Program) {}, ""},
		{"callee not exported", func(p *core.Program) { p.Functions[1].Exported = false }, calleeNotExportedCode},
		{"import fact stripped", func(p *core.Program) { p.Imports = nil }, calleeNotImportedCode},
		{"import fact names the wrong function", func(p *core.Program) { p.Imports[0].Name = "other" }, calleeNotImportedCode},
		{"import fact names the wrong importer", func(p *core.Program) { p.Imports[0].From = "elsewhere" }, calleeNotImportedCode},
		{"library caller into entry-module callee", func(p *core.Program) {
			p.Functions[0].Module = "lib"
			p.Functions[1].Module = ""
			p.Functions[1].Exported = false
		}, calleeNotExportedCode},
		{"library caller into entry-module callee even flagged exported", func(p *core.Program) {
			p.Functions[0].Module = "lib"
			p.Functions[1].Module = ""
			p.Functions[1].Exported = true
		}, calleeNotExportedCode},
		{"same-module private call inside lib", func(p *core.Program) {
			p.Functions[0].Module = "lib"
			p.Functions[1].Exported = false
			p.Imports = nil
		}, ""},
	}
	for _, blocks := range []bool{false, true} {
		site := "straight-line"
		if blocks {
			site = "blocks"
		}
		for _, r := range rows {
			t.Run(site+"/"+r.name, func(t *testing.T) {
				program := phase30CrossModuleProgram(blocks)
				r.mutate(&program)
				result := Validate(program)
				if got := firstProblemCode(result); got != r.want || result.Valid != (r.want == "") {
					t.Fatalf("want first problem %q, got valid=%v problems=%+v", r.want, result.Valid, result.Problems)
				}
				if r.want == "" {
					return
				}
				// Control: with the law disabled the forged program is admitted.
				disableVisibilityPeerForTest = true
				defer func() { disableVisibilityPeerForTest = false }()
				if control := Validate(program); !control.Valid {
					t.Fatalf("control: disabling the visibility law must admit the forged program, got %+v", control.Problems)
				}
			})
		}
	}
}

// moduleFactMismatches compares the module facts of a validated program with
// the input's, through JSON, and reports each mismatch instead of failing.
func moduleFactMismatches(input, validated core.Program) []string {
	var out []string
	for _, field := range []string{"Module", "Exported", "Imports"} {
		a, b := pick(input, field), pick(validated, field)
		if a != b {
			out = append(out, fmt.Sprintf("%s: %s != %s", field, a, b))
		}
	}
	return out
}

func pick(p core.Program, field string) string {
	var v any
	switch field {
	case "Module":
		var xs []string
		for _, fn := range p.Functions {
			xs = append(xs, fn.Module)
		}
		v = xs
	case "Exported":
		var xs []bool
		for _, fn := range p.Functions {
			xs = append(xs, fn.Exported)
		}
		v = xs
	default:
		v = p.Imports
	}
	raw, _ := json.Marshal(v)
	return string(raw)
}

// TestPhase30ValidatedProgramKeepsModuleFacts proves admission preserves
// Function.Module, Function.Exported and Program.Imports through
// cloneProgram, and that the M-CLONE mutant (a lossy clone, engaged through
// the production seam) is observed.
func TestPhase30ValidatedProgramKeepsModuleFacts(t *testing.T) {
	program := phase30CrossModuleProgram(false)
	result := Validate(program)
	if !result.Valid {
		t.Fatalf("fixture must validate: %+v", result.Problems)
	}
	if mismatches := moduleFactMismatches(program, result.Program()); len(mismatches) != 0 {
		t.Fatalf("clone lost module facts: %v", mismatches)
	}
	if pick(program, "Module") != `["","lib"]` || pick(program, "Exported") != `[false,true]` {
		t.Fatalf("fixture facts are not the non-trivial values this test needs: %s %s", pick(program, "Module"), pick(program, "Exported"))
	}

	dropModuleFactsInCloneForTest = true
	defer func() { dropModuleFactsInCloneForTest = false }()
	lossy := Validate(program)
	mismatches := moduleFactMismatches(program, lossy.Program())
	if len(mismatches) != 3 {
		t.Fatalf("control: the lossy clone must be observed on Module, Exported and Imports, got %v", mismatches)
	}
}
