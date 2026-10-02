package cgen_test

import (
	"encoding/json"
	"os"
	"strconv"
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

func phase25ExclusiveProgram(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase25", "exclusive_copy_accept.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("exclusive-copy fixture failed to check: %+v", checked.Diagnostics)
	}
	return source
}

func phase25SharedProgramWithCaller(t *testing.T) core.Program {
	t.Helper()
	return phase25SharedProgramWithCallerSource(t, phase25SharedProgram(t))
}

func phase25SharedProgramWithCallerSource(t *testing.T, source []byte) core.Program {
	t.Helper()
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("shared-copy fixture failed to check: %+v", checked.Diagnostics)
	}
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

func TestPhase25SharedPointerCopyABI(t *testing.T) {
	source := strings.Replace(string(phase25SharedProgram(t)), "  borrowed\n", "  let copied = borrowed\n  copied\n", 1)
	if source == string(phase25SharedProgram(t)) {
		t.Fatal("shared fixture return was not replaced with an explicit U64 copy")
	}
	program := phase25SharedProgramWithCallerSource(t, []byte(source))
	generation, err := cgen.EmitProgramNativeForTest(program)
	if err != nil {
		t.Fatalf("shared borrow-copy-return helper was refused: %v", err)
	}
	if !strings.Contains(generation, "const uint64_t *") || !strings.Contains(generation, "SCHWAY_SHARED_COPY(&") || !strings.Contains(generation, " = *") {
		t.Fatalf("shared U64 copy was not emitted through its const pointer ABI:\n%s", generation)
	}
}

func TestPhase25TransferredOwnerCallerRefusesNonExactPointerPaths(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*core.Program)
	}{
		{"return shared result", func(program *core.Program) {
			main := phase25Function(program, "main")
			main.Linear.Operations[5].SourceID = main.Linear.Operations[2].TargetID
		}},
		{"exclusive skips shared result", func(program *core.Program) {
			main := phase25Function(program, "main")
			main.Linear.Operations[3].SourceID = main.Linear.Operations[2].SourceID
		}},
		{"shared call targets exclusive helper", func(program *core.Program) {
			main := phase25Function(program, "main")
			main.Linear.Operations[2].CalleeID = phase25Function(program, "exclusive_copy").ID
		}},
		{"release order changes", func(program *core.Program) {
			main := phase25Function(program, "main")
			main.Linear.Operations[3], main.Linear.Operations[4] = main.Linear.Operations[4], main.Linear.Operations[3]
		}},
		{"extra call", func(program *core.Program) {
			main := phase25Function(program, "main")
			main.Linear.Operations = append(main.Linear.Operations, main.Linear.Operations[2])
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			program := phase25UtilityProgram(t)
			test.mutate(&program)
			if _, err := cgen.EmitProgramNativeForTest(program); err == nil {
				t.Fatal("non-exact transferred-owner caller unexpectedly reached application emission")
			}
			if cgen.InvocationSerializationReachedForTest() {
				t.Fatal("non-exact transferred-owner caller reached C serialization")
			}
		})
	}
}

func TestPhase25PreservesPhase24DirectTransferredOwnerPath(t *testing.T) {
	program := phase25UtilityProgram(t)
	main := phase25Function(&program, "main")
	operations := main.Linear.Operations
	main.Linear.Operations = []core.LinearOperation{operations[0], operations[1], operations[4], operations[5]}
	main.Linear.Operations[3].SourceID = operations[1].TargetID
	for index := range main.Linear.Operations {
		main.Linear.Operations[index].ID = main.ID + ":op:" + strconv.Itoa(index)
		main.Linear.Operations[index].PointID = main.ID + ":point:linear:" + strconv.Itoa(index)
	}
	main.Linear.Places = append([]core.Place(nil), main.Linear.Places[:3]...)
	filtered := make([]core.Function, 0, 2)
	for _, function := range program.Functions {
		if function.Name == "main" || function.Name == "acquire" {
			filtered = append(filtered, function)
		}
	}
	program.Functions = filtered
	generated, err := cgen.EmitApplication(program)
	if err != nil {
		t.Fatalf("Phase 24 direct transfer caller was refused: %v", err)
	}
	if !strings.Contains(generated, "schway_file_byte_use(") || !strings.Contains(generated, "paired release:") ||
		strings.Contains(generated, "SCHWAY_SHARED_COPY(") || strings.Contains(generated, "SCHWAY_EXCLUSIVE_COPY(") {
		t.Fatalf("Phase 24 direct result lowering changed while adding Phase 25 helpers:\n%s", generated)
	}
}

func phase25UtilityProgram(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase24", "transfer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("integrated utility fixture failed to check: %+v", checked.Diagnostics)
	}
	program := checked.Program
	program.Functions = append([]core.Function(nil), program.Functions...)
	for index := range program.Functions {
		if program.Functions[index].Linear == nil {
			continue
		}
		linear := *program.Functions[index].Linear
		linear.Types = append([]core.TypeFact(nil), linear.Types...)
		linear.Places = append([]core.Place(nil), linear.Places...)
		linear.Operations = append([]core.LinearOperation(nil), linear.Operations...)
		linear.Blocks = append([]core.Block(nil), linear.Blocks...)
		linear.Edges = append([]core.Edge(nil), linear.Edges...)
		program.Functions[index].Linear = &linear
	}
	return program
}

func phase25Function(program *core.Program, name string) *core.Function {
	for index := range program.Functions {
		if program.Functions[index].Name == name {
			return &program.Functions[index]
		}
	}
	panic("Phase 25 fixture function not found: " + name)
}

func phase25ExclusiveProgramWithCaller(t *testing.T) core.Program {
	t.Helper()
	checked := session.Check(phase25ExclusiveProgram(t))
	program := checked.Program
	program.Functions = append([]core.Function(nil), checked.Program.Functions...)
	helper := program.Functions[0]
	typeID := "phase25:exclusive-main:type:u64"
	parameterID := "phase25:exclusive-main:place:parameter"
	resultID := "phase25:exclusive-main:place:result"
	callerID := "phase25:exclusive-main"
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
	program := phase25SharedProgramWithCaller(t)
	generated, err := cgen.EmitProgramNativeForTest(program)
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

func TestPhase25ExclusivePointerABI(t *testing.T) {
	generated, err := cgen.EmitProgramNativeForTest(phase25ExclusiveProgramWithCaller(t))
	if err != nil {
		t.Fatalf("exclusive pointer helper was refused: %v", err)
	}
	if !strings.Contains(generated, "uint64_t *") {
		t.Fatalf("generated C has no exclusive pointer parameter:\n%s", generated)
	}
	if !strings.Contains(generated, "SCHWAY_EXCLUSIVE_COPY(&") {
		t.Fatalf("ordinary Schway call does not pass the address of its U64 argument:\n%s", generated)
	}
	if !strings.Contains(generated, " = *") {
		t.Fatalf("exclusive helper does not copy through its pointer parameter:\n%s", generated)
	}
	var declaration string
	for _, line := range strings.Split(generated, "\n") {
		if strings.Contains(line, "SCHWAY_EXCLUSIVE_COPY(") && strings.Contains(line, "uint64_t *") {
			declaration = line
			break
		}
	}
	for _, forbidden := range []string{"restrict", "noalias", "capture", "align("} {
		if strings.Contains(declaration, forbidden) {
			t.Fatalf("exclusive pointer declaration emitted unsupported promise %q: %s", forbidden, declaration)
		}
	}
}

func TestPhase25ExclusivePointerManifest(t *testing.T) {
	checked := session.Check(phase25ExclusiveProgram(t))
	manifest, err := cgen.EmitForeignManifest(checked.Program)
	if err != nil {
		t.Fatalf("exclusive pointer manifest was refused: %v", err)
	}
	var document struct {
		EmittedAttributes []json.RawMessage            `json:"emitted_attributes"`
		Pointers          []map[string]json.RawMessage `json:"emitted_pointer_parameters"`
	}
	if err := json.Unmarshal([]byte(manifest), &document); err != nil {
		t.Fatal(err)
	}
	if document.EmittedAttributes == nil || len(document.EmittedAttributes) != 0 || len(document.Pointers) != 1 {
		t.Fatalf("exclusive pointer manifest must contain only one pointer entry and no attributes: %s", manifest)
	}
	pointer := document.Pointers[0]
	if len(pointer) != 4 || string(pointer["c_type"]) != `"uint64_t *"` || string(pointer["access"]) != `"exclusive"` ||
		pointer["core_node"] == nil || pointer["parameter"] == nil {
		t.Fatalf("manifest pointer entry does not match the checked exclusive ABI fact: %s", manifest)
	}
}

func TestPhase25ExclusivePointerRefusal(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*core.Function)
	}{
		{"mutation", func(function *core.Function) { function.Linear.Operations[1].Kind = core.OpMove }},
		{"forwarding", func(function *core.Function) {
			function.Linear.Operations[2].SourceID = function.Linear.Operations[0].TargetID
		}},
		{"retention", func(function *core.Function) {
			function.Linear.Operations[2].SourceID = function.Linear.Operations[0].TargetID
		}},
		{"callback", func(function *core.Function) { function.Match = &core.Match{} }},
		{"nonlocal_exit", func(function *core.Function) { function.Linear.Blocks = []core.Block{{ID: "unexpected"}} }},
		{"wider_pointer", func(function *core.Function) {
			function.Linear.Operations = append(function.Linear.Operations, function.Linear.Operations[1])
			function.Linear.Operations[2].Kind = core.OpBorrowExclusive
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			program := phase25ExclusiveProgramWithCaller(t)
			test.mutate(&program.Functions[0])
			if _, err := cgen.EmitProgramNativeForTest(program); err == nil {
				t.Fatal("unsupported exclusive shape unexpectedly reached production emission")
			}
			if cgen.InvocationSerializationReachedForTest() {
				t.Fatal("unsupported exclusive shape reached C serialization")
			}
		})
	}
}
