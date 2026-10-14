package interp_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase35BitwiseShiftVectorsEventFree(t *testing.T) {
	const source = `module phase35.interp_shift
export { fn main }
fn main(unused: U64) -> U64 {
  let one = 1
  let shifted = one << 3
  shifted
}
`
	checked := session.Check([]byte(source))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("shift tracer diagnostics: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) != 1 || checked.Program.Functions[0].Linear == nil {
		t.Fatalf("shift tracer did not produce one linear function: %+v", checked.Program.Functions)
	}
	shiftCount := 0
	var pureOperationIDs []string
	for _, operation := range checked.Program.Functions[0].Linear.Operations {
		if operation.Kind == core.OpShiftLeftU64 {
			shiftCount++
			pureOperationIDs = append(pureOperationIDs, operation.ID)
			if operation.ConstU64 != "3" || operation.RightID != "" {
				t.Fatalf("shift tracer did not retain its checked literal immediate: %+v", operation)
			}
		}
	}
	if shiftCount != 1 {
		t.Fatalf("shift tracer has %d left-shift operations, want one", shiftCount)
	}
	got, err := interp.Run(checked.Program, "main", "0")
	if err != nil {
		t.Fatal(err)
	}
	assertPhase35PureExecution(t, got, "8", pureOperationIDs)

	for _, test := range []struct {
		name   string
		mutate func(*core.LinearOperation, *core.Function)
	}{
		{name: "out of range immediate", mutate: func(op *core.LinearOperation, _ *core.Function) { op.ConstU64 = "64" }},
		{name: "wrong result type", mutate: func(op *core.LinearOperation, _ *core.Function) { op.TypeID = "missing-u64-type" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			mutated := checked.Program
			mutated.Functions = append([]core.Function(nil), checked.Program.Functions...)
			mutated.Functions[0].Linear = &core.LinearBody{
				ID:         checked.Program.Functions[0].Linear.ID,
				Types:      append([]core.TypeFact(nil), checked.Program.Functions[0].Linear.Types...),
				Places:     append([]core.Place(nil), checked.Program.Functions[0].Linear.Places...),
				Operations: append([]core.LinearOperation(nil), checked.Program.Functions[0].Linear.Operations...),
			}
			function := &mutated.Functions[0]
			for index := range function.Linear.Operations {
				if function.Linear.Operations[index].Kind == core.OpShiftLeftU64 {
					test.mutate(&function.Linear.Operations[index], function)
					break
				}
			}
			if _, err := interp.Run(mutated, "main", "0"); err == nil || !strings.Contains(err.Error(), "core validation failed") {
				t.Fatalf("malformed shift reached interpreter evaluation: err=%v", err)
			}
		})
	}

	fixture := phase35ReadValueVectors(t)
	for _, vector := range fixture.Cases {
		t.Run("vector/"+vector.Name, func(t *testing.T) {
			checked := phase35CheckSource(t, phase35VectorSource(vector))
			pureOperationIDs := phase35PureOperationIDs(checked.Program)
			if len(pureOperationIDs) == 0 {
				t.Fatal("vector source did not lower a pure Phase 35 operation")
			}
			got, err := interp.Run(checked.Program, "main", "0")
			if err != nil {
				t.Fatal(err)
			}
			assertPhase35PureExecution(t, got, vector.ExpectedDecimal, pureOperationIDs)
		})
	}
}

func TestPhase35InterpreterWrongAnswerControls(t *testing.T) {
	fixture := phase35ReadValueVectors(t)
	wantOps := map[string]bool{"xor": false, "and": false, "or": false, "shift_left": false, "shift_right": false}
	for _, vector := range fixture.Cases {
		if _, needed := wantOps[vector.Operation]; !needed || wantOps[vector.Operation] {
			continue
		}
		wantOps[vector.Operation] = true
		t.Run(vector.Operation, func(t *testing.T) {
			checked := phase35CheckSource(t, phase35VectorSource(vector))
			got, err := interp.Run(checked.Program, "main", "0")
			if err != nil {
				t.Fatal(err)
			}
			if err := phase35OutcomeMismatch(got, vector.ExpectedDecimal); err != nil {
				t.Fatalf("independently pinned answer mismatch: %v", err)
			}
			wrong := got
			wrong.Outcome.Value = "1"
			if wrong.Outcome.Value == vector.ExpectedDecimal {
				wrong.Outcome.Value = "2"
			}
			if err := phase35OutcomeMismatch(wrong, vector.ExpectedDecimal); err == nil {
				t.Fatal("copied wrong-answer substitution passed the exact expected-value comparator")
			}
			restored, err := interp.Run(checked.Program, "main", "0")
			if err != nil {
				t.Fatal(err)
			}
			assertPhase35PureExecution(t, restored, vector.ExpectedDecimal, phase35PureOperationIDs(checked.Program))
		})
	}
	for operation, seen := range wantOps {
		if !seen {
			t.Errorf("no independent wrong-answer control for %s", operation)
		}
	}
}

func TestPhase35UnusedByteRuntime(t *testing.T) {
	path := testsupport.ProjectPath("examples", "bitwise.schway")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read tracked Byte-unused witness: %v", err)
	}
	checked := phase35CheckSource(t, string(source))
	if len(checked.Program.Functions) != 1 || checked.Program.Functions[0].Parameter.Type != "Byte" {
		t.Fatalf("witness entry did not retain its Byte formal: %+v", checked.Program.Functions)
	}
	function := checked.Program.Functions[0]
	seen := map[core.OperationKind]int{}
	returnIndex := -1
	for index, operation := range function.Linear.Operations {
		if operation.Kind == core.OpReturn {
			returnIndex = index
		}
		if operation.Kind == core.OpBranch || operation.Kind == core.OpFail || operation.Kind == core.OpDefect {
			t.Fatalf("Byte-unused witness is not a straight-line runtime path: %+v", operation)
		}
		if phase35PureOperation(operation.Kind) {
			seen[operation.Kind]++
			if returnIndex >= 0 {
				t.Fatalf("pure operation %s occurs after the return site", operation.Kind)
			}
		}
		if operation.SourceID == function.Parameter.ID || operation.RightID == function.Parameter.ID ||
			operation.TargetID == function.Parameter.ID || operation.StoreTargetID == function.Parameter.ID ||
			operation.PayloadTargetID == function.Parameter.ID {
			t.Fatalf("Byte formal was read, written, or converted by %s: %+v", operation.Kind, operation)
		}
	}
	if returnIndex != len(function.Linear.Operations)-1 {
		t.Fatalf("witness return is not the final operation: return index=%d operations=%+v", returnIndex, function.Linear.Operations)
	}
	for _, kind := range []core.OperationKind{core.OpBitXorU64, core.OpBitAndU64, core.OpBitOrU64, core.OpShiftLeftU64, core.OpShiftRightU64} {
		if seen[kind] != 1 {
			t.Fatalf("witness contains %d %s operations, want one reached by its straight-line body", seen[kind], kind)
		}
	}
	got, err := interp.Run(checked.Program, "main", "7")
	if err != nil {
		t.Fatal(err)
	}
	assertPhase35PureExecution(t, got, "264", phase35PureOperationIDs(checked.Program))

	const locked = "x ^ (y << 3) & 0xFF"
	const alternate = "(x ^ (y << 3)) & 0xFF"
	alternateSource := strings.Replace(string(source), locked, alternate, 1)
	if alternateSource == string(source) {
		t.Fatal("the locked witness expression was not found for its discriminating association control")
	}
	alternateProgram := phase35CheckSource(t, alternateSource)
	alternateRun, err := interp.Run(alternateProgram.Program, "main", "7")
	if err != nil {
		t.Fatal(err)
	}
	assertPhase35PureExecution(t, alternateRun, "8", phase35PureOperationIDs(alternateProgram.Program))
}

func assertPhase35PureExecution(t *testing.T, got interp.Execution, want string, pureOperationIDs []string) {
	t.Helper()
	if err := phase35OutcomeMismatch(got, want); err != nil {
		t.Fatal(err)
	}
	returns := 0
	for _, event := range got.Events {
		if event.Kind == "function.returned" {
			returns++
		}
		if event.Kind == "function.defected" {
			t.Fatalf("pure scalar execution produced defect/reason event: %+v", event)
		}
		for _, operationID := range pureOperationIDs {
			if strings.HasPrefix(event.ID, operationID+":event") {
				t.Fatalf("pure operation %q emitted an event: %+v", operationID, event)
			}
		}
	}
	if returns != 1 || len(got.Events) == 0 || got.Events[len(got.Events)-1].Kind != "function.returned" {
		t.Fatalf("pure scalar program events = %+v, want existing step evidence and one terminal return", got.Events)
	}
	if len(got.LiveResources) != 0 {
		t.Fatalf("pure scalar program left resources live: %v", got.LiveResources)
	}
}

func phase35OutcomeMismatch(got interp.Execution, want string) error {
	if got.Outcome.Kind != "returned" || got.Outcome.Value != want {
		return fmt.Errorf("outcome = %+v, want returned value %q", got.Outcome, want)
	}
	return nil
}

type phase35ValueVectorFixture struct {
	Schema string               `json:"schema"`
	Cases  []phase35ValueVector `json:"cases"`
}

type phase35ValueVector struct {
	Name            string `json:"name"`
	Operation       string `json:"op"`
	Left            string `json:"left"`
	Right           string `json:"right"`
	Count           string `json:"count"`
	Expression      string `json:"expression"`
	X               string `json:"x"`
	Y               string `json:"y"`
	ExpectedDecimal string `json:"expected_decimal"`
}

func phase35ReadValueVectors(t *testing.T) phase35ValueVectorFixture {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase35", "value-vectors.json"))
	if err != nil {
		t.Fatalf("read independent Phase 35 answer vectors: %v", err)
	}
	var fixture phase35ValueVectorFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode independent Phase 35 answer vectors: %v", err)
	}
	if fixture.Schema != "schway.phase35.u64-value-vectors/1" || len(fixture.Cases) == 0 {
		t.Fatalf("Phase 35 answer fixture has invalid schema or no cases: %+v", fixture)
	}
	return fixture
}

func phase35VectorSource(vector phase35ValueVector) string {
	body := ""
	if vector.Expression != "" {
		body = fmt.Sprintf("  let x = %s\n  let y = %s\n  let result = %s\n", vector.X, vector.Y, vector.Expression)
	} else {
		operator := map[string]string{"xor": "^", "and": "&", "or": "|"}[vector.Operation]
		if operator != "" {
			body = fmt.Sprintf("  let left = %s\n  let right = %s\n  let result = left %s right\n", vector.Left, vector.Right, operator)
		} else {
			operator = map[string]string{"shift_left": "<<", "shift_right": ">>"}[vector.Operation]
			body = fmt.Sprintf("  let left = %s\n  let result = left %s %s\n", vector.Left, operator, vector.Count)
		}
	}
	return "module phase35.interp_vector\nexport { fn main }\nfn main(unused: U64) -> U64 {\n" + body + "  result\n}\n"
}

func phase35CheckSource(t *testing.T, source string) session.CheckResult {
	t.Helper()
	checked := session.Check([]byte(source))
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) != 1 {
		t.Fatalf("Phase 35 interpreter source failed checking: diagnostics=%+v functions=%d\n%s", checked.Diagnostics, len(checked.Program.Functions), source)
	}
	return checked
}

func phase35PureOperation(kind core.OperationKind) bool {
	switch kind {
	case core.OpBitXorU64, core.OpBitAndU64, core.OpBitOrU64, core.OpShiftLeftU64, core.OpShiftRightU64:
		return true
	default:
		return false
	}
}

func phase35PureOperationIDs(program core.Program) []string {
	var operationIDs []string
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if phase35PureOperation(operation.Kind) {
				operationIDs = append(operationIDs, operation.ID)
			}
		}
	}
	return operationIDs
}
