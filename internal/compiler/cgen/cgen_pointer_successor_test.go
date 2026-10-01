package cgen_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase25SharedProgram(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase25", "shared_copy_accept.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("shared-copy fixture failed to check: %+v", checked.Diagnostics)
	}
	return source
}

func phase25SharedProgramWithCaller(t *testing.T) core.Program {
	t.Helper()
	checked := session.Check(phase25SharedProgram(t))
	program := checked.Program
	program.Functions = append([]core.Function(nil), checked.Program.Functions...)
	helper := program.Functions[0]
	typeID := "phase25:main:type:u64"
	parameterID := "phase25:main:place:parameter"
	resultID := "phase25:main:place:result"
	callerID := "phase25:main"
	program.Functions = append(program.Functions, core.Function{
		ID: callerID, Name: "main", EntryPointID: callerID + ":point:entry", ReturnPointID: callerID + ":point:return",
		Parameter: core.Parameter{ID: parameterID, Name: "input", Type: "U64"}, ReturnType: "U64",
		Linear: &core.LinearBody{
			ID:     callerID + ":linear",
			Types:  []core.TypeFact{{ID: typeID, Shape: core.TypeRef{Constructor: "U64"}}},
			Places: []core.Place{{ID: parameterID, Name: "input", TypeID: typeID}, {ID: resultID, Name: "result", TypeID: typeID}},
			Operations: []core.LinearOperation{
				{ID: callerID + ":op:0", PointID: callerID + ":point:linear:0", Kind: core.OpCall, SourceID: parameterID, TargetID: resultID, TypeID: typeID, CalleeID: helper.ID},
				{ID: callerID + ":op:1", PointID: callerID + ":point:linear:1", Kind: core.OpReturn, SourceID: resultID, TypeID: typeID},
			},
		},
	})
	return program
}

func TestPhase25SharedPointerABI(t *testing.T) {
	generated, err := cgen.EmitProgramNativeForTest(phase25SharedProgramWithCaller(t))
	if err != nil {
		t.Fatalf("shared pointer helper was refused: %v", err)
	}
	if !strings.Contains(generated, "const uint64_t *") {
		t.Fatalf("generated C has no const pointer parameter:\n%s", generated)
	}
	if !strings.Contains(generated, "SCHWAY_SHARED_COPY(&") {
		t.Fatalf("ordinary Schway call does not pass the address of its U64 argument:\n%s", generated)
	}
	if !strings.Contains(generated, " = *") {
		t.Fatalf("shared helper does not copy through its pointer parameter:\n%s", generated)
	}
}

func TestPhase25SharedPointerManifest(t *testing.T) {
	checked := session.Check(phase25SharedProgram(t))
	manifest, err := cgen.EmitForeignManifest(checked.Program)
	if err != nil {
		t.Fatalf("shared pointer manifest was refused: %v", err)
	}
	var document struct {
		EmittedAttributes []json.RawMessage            `json:"emitted_attributes"`
		Pointers          []map[string]json.RawMessage `json:"emitted_pointer_parameters"`
	}
	if err := json.Unmarshal([]byte(manifest), &document); err != nil {
		t.Fatal(err)
	}
	if document.EmittedAttributes == nil || len(document.EmittedAttributes) != 0 {
		t.Fatalf("shared pointer ABI must carry no unsupported attributes: %s", manifest)
	}
	if len(document.Pointers) != 1 {
		t.Fatalf("manifest pointer entry count=%d, want one: %s", len(document.Pointers), manifest)
	}
	pointer := document.Pointers[0]
	if len(pointer) != 4 || string(pointer["c_type"]) != `"const uint64_t *"` || string(pointer["access"]) != `"shared"` ||
		pointer["core_node"] == nil || pointer["parameter"] == nil {
		t.Fatalf("manifest pointer entry does not match the checked shared ABI fact: %s", manifest)
	}
}

func TestPhase25SharedPointerABIFailClosed(t *testing.T) {
	program := phase25SharedProgramWithCaller(t)
	helper := &program.Functions[0]
	borrow := helper.Linear.Operations[0]
	reborrow := borrow
	reborrow.ID += ":wider"
	reborrow.PointID += ":wider"
	reborrow.SourceID = borrow.TargetID
	reborrow.TargetID += ":wider"
	reborrow.LoanID += ":wider"
	helper.Linear.Places = append(helper.Linear.Places, core.Place{ID: reborrow.TargetID, Name: "wider", TypeID: borrow.TypeID})
	helper.Linear.Operations = []core.LinearOperation{borrow, reborrow, {ID: "return:wider", Kind: core.OpReturn, SourceID: reborrow.TargetID, TypeID: reborrow.TypeID}}
	if _, err := cgen.EmitProgramNativeForTest(program); err == nil {
		t.Fatal("wider shared pointer chain unexpectedly reached production emission")
	}
	if cgen.InvocationSerializationReachedForTest() {
		t.Fatal("unsupported shared pointer chain reached C serialization")
	}
}
