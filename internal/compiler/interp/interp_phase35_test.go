package interp_test

import (
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
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
}

func assertPhase35PureExecution(t *testing.T, got interp.Execution, want string, pureOperationIDs []string) {
	t.Helper()
	if got.Outcome.Kind != "returned" || got.Outcome.Value != want {
		t.Fatalf("outcome = %+v, want returned value %q", got.Outcome, want)
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
