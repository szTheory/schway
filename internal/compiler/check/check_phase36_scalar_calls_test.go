package check_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func TestPhase36ScalarCallLoweringAndTypePairs(t *testing.T) {
	source := []byte(`module phase36.scalar_calls
export { fn main }
fn identity_u64(input: U64) -> U64 { let result = input result }
fn less_than_one(input: U64) -> Bool { let result = input < 1 result }
fn bool_number(input: Bool) -> U64 { let result = 1 result }
fn identity_bool(input: Bool) -> Bool { let result = input result }
fn inner(input: U64) -> U64 { let result = input + 1 result }
fn outer(input: U64) -> U64 { let result = input + 2 result }
fn sibling(input: U64) -> U64 { let result = input + 3 result }
fn index_value(input: U64) -> U64 { let result = input result }
fn dynamic_zero(input: U64) -> U64 { let result = 0 result }
fn call_u64_u64(input: U64) -> U64 { let result = identity_u64(input) result }
fn call_u64_bool(input: U64) -> Bool { let result = less_than_one(input) result }
fn call_bool_u64(input: Bool) -> U64 { let result = bool_number(input) result }
fn call_bool_bool(input: Bool) -> Bool { let result = identity_bool(input) result }
fn compose(input: U64) -> U64 {
  let result = outer(inner(input + 1)) + sibling(input + 2)
  result
}

fn contexts(input: U64) -> U64 {
  var result = inner(input)
  result = sibling(result + 1)
  if less_than_one(result) {
    result = inner(result)
  } else {
    result = sibling(result)
  }
  result
}
fn byte_index(view: ByteView) -> U64 {
  var result = 0
  result = byte_at(view, index_value(result))
  result
}
fn divide_by_dynamic_zero(input: U64) -> U64 {
  let result = input % dynamic_zero(input)
  result
}
fn main(input: U64) -> U64 { let result = call_u64_u64(input) result }
`)
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("scalar call source did not parse: %+v", parsed.Diagnostics)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("well-typed scalar helper call source was refused: %+v", checked.Diagnostics)
	}

	functionIDs := make(map[string]string, len(checked.Program.Functions))
	for _, function := range checked.Program.Functions {
		functionIDs[function.Name] = function.ID
	}
	for _, pair := range []struct {
		caller string
		callee string
		result string
	}{
		{"call_u64_u64", "identity_u64", "U64"},
		{"call_u64_bool", "less_than_one", "Bool"},
		{"call_bool_u64", "bool_number", "U64"},
		{"call_bool_bool", "identity_bool", "Bool"},
	} {
		caller := phase36Function(t, checked.Program, pair.caller)
		callee := phase36Function(t, checked.Program, pair.callee)
		calls := phase36CallOperations(caller)
		if len(calls) != 1 {
			t.Fatalf("%s call operations = %+v, want one OpCall", pair.caller, calls)
		}
		call := calls[0]
		if call.SourceID != caller.Parameter.ID || call.CalleeID != callee.ID || call.TargetID == "" || call.TargetID == call.SourceID {
			t.Fatalf("%s call identity/source/target = %+v, parameter=%s callee=%s", pair.caller, call, caller.Parameter.ID, callee.ID)
		}
		places := make(map[string]core.Place, len(caller.Linear.Places))
		for _, place := range caller.Linear.Places {
			places[place.ID] = place
		}
		place, exists := places[call.TargetID]
		if !exists || place.TypeID != call.TypeID {
			t.Fatalf("%s call target place is absent or disagrees with its TypeID: %+v", pair.caller, call)
		}
		constructor := ""
		for _, fact := range caller.Linear.Types {
			if fact.ID == call.TypeID {
				constructor = fact.Shape.Constructor
			}
		}
		if constructor != pair.result || callee.ReturnType != pair.result {
			t.Fatalf("%s call result constructor=%q callee return=%q, want %q", pair.caller, constructor, callee.ReturnType, pair.result)
		}
	}
	compose := phase36Function(t, checked.Program, "compose")
	composeCalls := phase36CallOperations(compose)
	if len(composeCalls) != 3 {
		t.Fatalf("nested/sibling call count = %d, want 3: %+v", len(composeCalls), composeCalls)
	}
	wantCallers := []string{functionIDs["inner"], functionIDs["outer"], functionIDs["sibling"]}
	for index, call := range composeCalls {
		if call.CalleeID != wantCallers[index] {
			t.Fatalf("call order[%d] = %q, want %q", index, call.CalleeID, wantCallers[index])
		}
		for earlier := 0; earlier < index; earlier++ {
			if call.TargetID == composeCalls[earlier].TargetID {
				t.Fatalf("call results reused static target %q", call.TargetID)
			}
		}
	}
	if composeCalls[1].SourceID != composeCalls[0].TargetID {
		t.Fatalf("outer call did not consume the completed inner result: inner=%+v outer=%+v", composeCalls[0], composeCalls[1])
	}
	composePositions := make(map[string]int, len(compose.Linear.Operations))
	var addPositions []int
	for index, operation := range compose.Linear.Operations {
		composePositions[operation.ID] = index
		if operation.Kind == core.OpAddChecked {
			addPositions = append(addPositions, index)
		}
	}
	callPositions := make([]int, len(composeCalls))
	for index, call := range composeCalls {
		if sourceIndex, targetIndex := composePositions[call.SourceID], composePositions[call.ID]; sourceIndex >= targetIndex {
			t.Fatalf("call %s ran before its argument producer %s", call.ID, call.SourceID)
		}
		callPositions[index] = composePositions[call.ID]
	}
	if len(addPositions) != 3 || !(addPositions[0] < callPositions[0] && callPositions[0] < callPositions[1] &&
		callPositions[1] < addPositions[1] && addPositions[1] < callPositions[2] && callPositions[2] < addPositions[2]) {
		t.Fatalf("computed nested/sibling arguments were not materialized left-to-right before the parent add: adds=%v calls=%v", addPositions, callPositions)
	}

	contexts := phase36Function(t, checked.Program, "contexts")
	contextCalls := phase36CallOperations(contexts)
	if len(contextCalls) != 5 {
		t.Fatalf("initializer/assignment/condition/branch calls = %+v, want 5 calls", contextCalls)
	}
	wantContextCallees := []string{functionIDs["inner"], functionIDs["sibling"], functionIDs["less_than_one"], functionIDs["inner"], functionIDs["sibling"]}
	for index, call := range contextCalls {
		if call.CalleeID != wantContextCallees[index] {
			t.Fatalf("context call order[%d] = %q, want %q", index, call.CalleeID, wantContextCallees[index])
		}
	}
	contextPositions := make(map[string]int, len(contexts.Linear.Operations))
	var stores []core.LinearOperation
	for index, operation := range contexts.Linear.Operations {
		contextPositions[operation.ID] = index
		if operation.Kind == core.OpScalarStore {
			stores = append(stores, operation)
		}
	}
	if len(stores) < 4 || stores[0].SourceID != contextCalls[0].TargetID ||
		contextPositions[contextCalls[0].ID] >= contextPositions[stores[0].ID] ||
		contextPositions[contextCalls[1].ID] >= contextPositions[stores[1].ID] {
		t.Fatalf("call results were not initialized/stored after var and assignment evaluation: calls=%+v stores=%+v", contextCalls, stores)
	}
	branches := 0
	for _, operation := range contexts.Linear.Operations {
		if operation.Kind == core.OpBranch {
			branches++
			if operation.SourceID != contextCalls[2].TargetID {
				t.Fatalf("condition branch reads %q, want Bool call result %q", operation.SourceID, contextCalls[2].TargetID)
			}
		}
	}
	if branches != 1 {
		t.Fatalf("condition branch count = %d, want 1", branches)
	}

	byteIndex := phase36Function(t, checked.Program, "byte_index")
	indexCalls := phase36CallOperations(byteIndex)
	var byteRead core.LinearOperation
	for _, operation := range byteIndex.Linear.Operations {
		if operation.Kind == core.OpBytesAt {
			byteRead = operation
		}
	}
	if len(indexCalls) != 1 || byteRead.Kind != core.OpBytesAt || byteRead.RightID != indexCalls[0].TargetID {
		t.Fatalf("call in byte_at index was not materialized before the byte read: calls=%+v byte_at=%+v", indexCalls, byteRead)
	}

	divide := phase36Function(t, checked.Program, "divide_by_dynamic_zero")
	zeroCalls := phase36CallOperations(divide)
	var remainder core.LinearOperation
	for _, operation := range divide.Linear.Operations {
		if operation.Kind == core.OpRemainderU64 {
			remainder = operation
		}
	}
	if len(zeroCalls) != 1 || remainder.Kind != core.OpRemainderU64 || remainder.RightID != zeroCalls[0].TargetID {
		t.Fatalf("dynamic helper result did not remain an initialized unknown divisor: call=%+v remainder=%+v", zeroCalls, remainder)
	}

	byteSource := []byte(`module phase36.byte_call_use
export { fn main }
fn identity_u64(input: U64) -> U64 { let result = input result }
fn main(input: Byte) -> U64 {
  var result = 0
  result = identity_u64(input)
  result
}
`)
	byteUse := session.Check(byteSource)
	if len(byteUse.Diagnostics) != 1 || byteUse.Diagnostics[0].Code != "type.scalar_byte_parameter_used" {
		t.Fatalf("unused Byte bridge allowed its formal inside a call argument: %+v", byteUse.Diagnostics)
	}
}

func phase36Function(t *testing.T, program core.Program, name string) core.Function {
	t.Helper()
	for _, function := range program.Functions {
		if function.Name == name {
			return function
		}
	}
	t.Fatalf("checked function %q not found", name)
	return core.Function{}
}

func phase36CallOperations(function core.Function) []core.LinearOperation {
	var calls []core.LinearOperation
	if function.Linear == nil {
		return calls
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpCall {
			calls = append(calls, operation)
		}
	}
	return calls
}

func TestPhase36CallImportAndRefusalBoundaries(t *testing.T) {
	entry := func(importOrder string) string {
		return "module app.main\n\n" + importOrder + "\nexport { fn main }\n" +
			"fn main(input: U64) -> U64 { let result = convert(input + 1) result }\n"
	}
	libAFirst := "module alpha\nexport {\n  fn convert\n  fn anchor\n}\n" +
		"fn convert(input: U64) -> U64 { let result = input + 2 result }\n" +
		"fn anchor(input: U64) -> U64 { let result = input result }\n"
	libAReordered := "module alpha\nexport {\n  fn anchor\n  fn convert\n}\n" +
		"fn anchor(input: U64) -> U64 { let result = input result }\n" +
		"fn convert(input: U64) -> U64 { let result = input + 2 result }\n"
	libB := "module beta\nexport { fn other }\n" +
		"fn convert(input: U64) -> U64 { let result = input + 99 result }\n" +
		"fn other(input: U64) -> U64 { let result = input result }\n"

	var calleeIDs []string
	for _, fixture := range []struct {
		name        string
		entry       string
		alphaSource string
	}{
		{"alpha-first", entry("import alpha { fn convert }\nimport beta { fn other }"), libAFirst},
		{"declaration-and-import-order-permuted", entry("import beta { fn other }\nimport alpha { fn convert }"), libAReordered},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			path := phase36WriteModuleFixture(t, map[string]string{
				"main.schway":  fixture.entry,
				"alpha.schway": fixture.alphaSource,
				"beta.schway":  libB,
			})
			checked, err := session.CheckFile(path)
			if err != nil || len(checked.Diagnostics) != 0 {
				t.Fatalf("imported scalar call diagnostics=%+v err=%v", checked.Diagnostics, err)
			}
			main := phase36Function(t, checked.Program, "main")
			calls := phase36CallOperations(main)
			alphaConvert := phase36FunctionInModule(t, checked.Program, "alpha", "convert")
			if len(calls) != 1 || calls[0].CalleeID != alphaConvert.ID {
				t.Fatalf("imported same-named helper resolved to the wrong stable identity: calls=%+v alpha=%s", calls, alphaConvert.ID)
			}
			calleeIDs = append(calleeIDs, calls[0].CalleeID)
		})
	}
	if len(calleeIDs) != 2 || calleeIDs[0] != calleeIDs[1] {
		t.Fatalf("declaration/import order changed imported callee identity: %v", calleeIDs)
	}

	t.Run("private target is not importable", func(t *testing.T) {
		path := phase36WriteModuleFixture(t, map[string]string{
			"main.schway": "module app.main\nimport lib { fn hidden }\nexport { fn main }\nfn main(input: U64) -> U64 { let result = hidden(input) result }\n",
			"lib.schway":  "module lib\nexport {\n}\nfn hidden(input: U64) -> U64 { let result = input result }\n",
		})
		checked, err := session.CheckFile(path)
		if err != nil || len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "module.not_exported" {
			t.Fatalf("private imported call diagnostics=%+v err=%v, want module.not_exported", checked.Diagnostics, err)
		}
	})
	t.Run("same name imported from two modules collides", func(t *testing.T) {
		path := phase36WriteModuleFixture(t, map[string]string{
			"main.schway":  "module app.main\nimport alpha { fn convert }\nimport beta { fn convert }\nexport { fn main }\nfn main(input: U64) -> U64 { let result = convert(input) result }\n",
			"alpha.schway": "module alpha\nexport { fn convert }\nfn convert(input: U64) -> U64 { let result = input result }\n",
			"beta.schway":  "module beta\nexport { fn convert }\nfn convert(input: U64) -> U64 { let result = input result }\n",
		})
		checked, err := session.CheckFile(path)
		if err != nil || len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "module.import_collision" {
			t.Fatalf("colliding imported call diagnostics=%+v err=%v, want module.import_collision", checked.Diagnostics, err)
		}
	})

	for _, row := range []struct {
		name   string
		source string
		code   string
	}{
		{
			"wrong-argument-type",
			"module phase36.bad_argument\nexport { fn main }\nfn bool_value(input: Bool) -> U64 { let result = 1 result }\nfn main(input: U64) -> U64 { let result = bool_value(input) result }\n",
			"check.call_argument_type_mismatch",
		},
		{
			"wrong-result-tail-type",
			"module phase36.bad_result\nexport { fn main }\nfn bool_value(input: Bool) -> Bool { let result = input result }\nfn main(input: Bool) -> U64 { let result = bool_value(input) result }\n",
			"type.scalar_result",
		},
		{
			"unit-result-is-not-a-value",
			"module phase36.unit_call\nexport { fn main }\nfn sink(input: U64) -> Unit { write_newline() }\nfn main(input: U64) -> U64 { let result = sink(input) result }\n",
			"type.scalar_signature",
		},
		{
			"byteview-is-not-a-scalar-argument",
			"module phase36.view_call\nexport { fn main }\nfn identity(input: U64) -> U64 { let result = input result }\nfn view_caller(view: ByteView) -> U64 { let result = identity(view) result }\nfn main(input: U64) -> U64 { let result = input result }\n",
			"check.call_argument_type_mismatch",
		},
		{
			"unused-byte-bridge-cannot-pass-its-formal",
			"module phase36.byte_call\nexport { fn main }\nfn identity(input: U64) -> U64 { let result = input result }\nfn main(input: Byte) -> U64 { var result = 0 result = identity(input) result }\n",
			"type.scalar_byte_parameter_used",
		},
		{
			"resource-value-is-not-a-scalar-signature",
			"module phase36.resource_call\nexport { fn main }\nfn identity(input: U64) -> U64 { let result = input result }\nfn resource_caller(owner: Buffer) -> U64 { let result = identity(owner) result }\nfn main(input: U64) -> U64 { let result = input result }\n",
			"check.call_argument_type_mismatch",
		},
		{
			"uninitialized-name-is-not-a-call-argument",
			"module phase36.uninitialized_call\nexport { fn main }\nfn identity(input: U64) -> U64 { let result = input result }\nfn main(input: U64) -> U64 { let result = identity(missing) result }\n",
			"name.unknown_place",
		},
		{
			"immutable-result-target",
			"module phase36.immutable_call\nexport { fn main }\nfn identity(input: U64) -> U64 { let result = input result }\nfn main(input: U64) -> U64 { let result = input result = identity(input) result }\n",
			"ownership.scalar_assignment",
		},
	} {
		t.Run("refusal/"+row.name, func(t *testing.T) {
			checked := session.Check([]byte(row.source))
			if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != row.code {
				t.Fatalf("diagnostics=%+v, want exactly %s", checked.Diagnostics, row.code)
			}
		})
	}

	t.Run("direct recursion in an unused function is refused", func(t *testing.T) {
		source := []byte("module phase36.direct_cycle\nexport { fn main }\nfn recur(input: U64) -> U64 { let result = recur(input) result }\nfn main(input: U64) -> U64 { let result = input result }\n")
		checked := session.Check(source)
		if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != core.CallGraphCycle {
			t.Fatalf("unused direct recursion diagnostics=%+v, want %s", checked.Diagnostics, core.CallGraphCycle)
		}
	})
	t.Run("mutual recursion is refused", func(t *testing.T) {
		source := []byte("module phase36.mutual_cycle\nexport { fn main }\nfn alpha(input: U64) -> U64 { let result = beta(input) result }\nfn beta(input: U64) -> U64 { let result = alpha(input) result }\nfn main(input: U64) -> U64 { let result = input result }\n")
		checked := session.Check(source)
		if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != core.CallGraphCycle {
			t.Fatalf("mutual recursion diagnostics=%+v, want %s", checked.Diagnostics, core.CallGraphCycle)
		}
	})
	t.Run("recursion after a defect is refused even in unreachable source", func(t *testing.T) {
		source := []byte("module phase36.unreachable_cycle\nexport { fn main }\nfn recur(input: U64) -> U64 { var result = 0 if 1 < 0 { result = recur(input) } else { result = input } result }\nfn main(input: U64) -> U64 { let result = input result }\n")
		checked := session.Check(source)
		if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != core.CallGraphCycle {
			t.Fatalf("unreachable recursion diagnostics=%+v, want %s", checked.Diagnostics, core.CallGraphCycle)
		}
	})
}

func phase36WriteModuleFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, "main.schway")
}

func phase36FunctionInModule(t *testing.T, program core.Program, module, name string) core.Function {
	t.Helper()
	for _, function := range program.Functions {
		if function.Module == module && function.Name == name {
			return function
		}
	}
	t.Fatalf("no %s.%s function in checked program", module, name)
	return core.Function{}
}

func TestPhase36CallSourceIntentControls(t *testing.T) {
	for _, mode := range []string{"argument", "result"} {
		t.Run(mode, func(t *testing.T) {
			original, err := os.ReadFile("scalar.go")
			if err != nil {
				t.Fatal(err)
			}
			mutation := phase36CallIntentMutation(mode)
			mutated := strings.Replace(string(original), mutation.before, mutation.after, 1)
			if mutated == string(original) {
				t.Fatalf("%s source-intent mutation anchor was not found", mode)
			}

			phase36RunSourceIntentChild(t, mode, nil, false, "")
			phase36RunSourceIntentChild(t, mode, original, false, "")
			phase36RunSourceIntentChild(t, mode, []byte(mutated), true, mutation.marker)
			phase36RunSourceIntentChild(t, mode, original, false, "")
		})
	}
}

func TestPhase36CallSourceIntentChild(t *testing.T) {
	mode := os.Getenv("PHASE36_CALL_INTENT_MODE")
	var source []byte
	switch mode {
	case "argument":
		source = []byte("module phase36.intent_argument\nexport { fn main }\nfn identity(input: U64) -> U64 { let result = input result }\nfn main(input: U64) -> U64 { let result = identity(input + 4) result }\n")
	case "result":
		source = []byte("module phase36.intent_result\nexport { fn main }\nfn identity(input: Bool) -> Bool { let result = input result }\nfn main(input: Bool) -> Bool { let result = identity(input) result }\n")
	default:
		return
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		if mode == "result" && len(checked.Diagnostics) == 1 && checked.Diagnostics[0].Code == "type.scalar_result" {
			t.Fatalf("SOURCE_INTENT_RESULT_MISMATCH: declared Bool helper result was rebound as U64: %+v", checked.Diagnostics)
		}
		t.Fatalf("source-intent positive was refused: %+v", checked.Diagnostics)
	}
	main := phase36Function(t, checked.Program, "main")
	calls := phase36CallOperations(main)
	if len(calls) != 1 {
		t.Fatalf("source-intent fixture has %d OpCall operations, want one", len(calls))
	}
	call := calls[0]
	places := make(map[string]core.Place, len(main.Linear.Places))
	for _, place := range main.Linear.Places {
		places[place.ID] = place
	}
	switch mode {
	case "argument":
		var computedArgument string
		for _, operation := range main.Linear.Operations {
			if operation.Kind == core.OpAddChecked {
				computedArgument = operation.TargetID
			}
		}
		if computedArgument == "" || call.SourceID != computedArgument {
			t.Fatalf("SOURCE_INTENT_ARGUMENT_MISMATCH: OpCall source=%q, computed source argument=%q", call.SourceID, computedArgument)
		}
	case "result":
		resultType := ""
		for _, fact := range main.Linear.Types {
			if fact.ID == call.TypeID {
				resultType = fact.Shape.Constructor
			}
		}
		if places[call.TargetID].TypeID != call.TypeID || resultType != "Bool" {
			t.Fatalf("SOURCE_INTENT_RESULT_MISMATCH: call target=%+v operation type=%s constructor=%s", places[call.TargetID], call.TypeID, resultType)
		}
	}
}

func phase36CallIntentMutation(mode string) struct{ before, after, marker string } {
	switch mode {
	case "argument":
		return struct{ before, after, marker string }{
			before: "argument, argumentType := b.expression(expr.Left)",
			after:  "_, argumentType := b.expression(expr.Left)\n\targument := b.names[b.function.Parameter.Name]",
			marker: "SOURCE_INTENT_ARGUMENT_MISMATCH",
		}
	case "result":
		return struct{ before, after, marker string }{
			before: "place := b.addPlace(\"call\", contract.ReturnType, false)",
			after:  "contract.ReturnType = \"U64\"\n\tplace := b.addPlace(\"call\", contract.ReturnType, false)",
			marker: "SOURCE_INTENT_RESULT_MISMATCH",
		}
	default:
		panic("unknown source-intent mutation mode")
	}
}

func phase36RunSourceIntentChild(t *testing.T, mode string, replacement []byte, wantFailure bool, marker string) {
	t.Helper()
	args := []string{"test"}
	if replacement != nil {
		sourcePath, err := filepath.Abs("scalar.go")
		if err != nil {
			t.Fatal(err)
		}
		dir := t.TempDir()
		replacementPath := filepath.Join(dir, "scalar.go")
		if err := os.WriteFile(replacementPath, replacement, 0o600); err != nil {
			t.Fatal(err)
		}
		overlayBytes, err := json.Marshal(struct {
			Replace map[string]string `json:"Replace"`
		}{Replace: map[string]string{sourcePath: replacementPath}})
		if err != nil {
			t.Fatal(err)
		}
		overlayPath := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(overlayPath, overlayBytes, 0o600); err != nil {
			t.Fatal(err)
		}
		args = append(args, "-overlay", overlayPath)
	}
	args = append(args, ".", "-run", "^TestPhase36CallSourceIntentChild$", "-count=1", "-v", "-timeout=600s")
	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "GOCACHE=/tmp/ai-lang-verification-gocache", "PHASE36_CALL_INTENT_MODE="+mode)
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Run()
	text := output.String()
	if wantFailure {
		if err == nil || !strings.Contains(text, "--- FAIL: TestPhase36CallSourceIntentChild") {
			t.Fatalf("%s mutant did not fail its child test: err=%v\n%s", mode, err, text)
		}
		if marker != "" && (!strings.Contains(text, marker) || strings.Contains(text, "[build failed]")) {
			t.Fatalf("%s mutant did not reach named assertion %s: err=%v\n%s", mode, marker, err, text)
		}
		return
	}
	if err != nil || !strings.Contains(text, "=== RUN   TestPhase36CallSourceIntentChild") || !strings.Contains(text, "--- PASS: TestPhase36CallSourceIntentChild") {
		t.Fatalf("%s clean/no-op/restored source-intent child failed: err=%v\n%s", mode, err, text)
	}
}
