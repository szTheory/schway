package pathoracle_test

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase25PointerPath(t *testing.T) {
	program := checkedProgram(t, "phase3", "branch_one_arm_shared_accept.schway")
	function := functionNamed(t, program, "choose")
	endpoints, _, err := pathoracle.RecomputeEndpoints(function, nil)
	if err != nil {
		t.Fatalf("independent path replay rejected the shared-loan witness: %v", err)
	}
	if len(endpoints) != 1 {
		t.Fatalf("want one independently-derived shared endpoint, got %+v", endpoints)
	}

	// Missing and shifted serialized endpoint claims cannot steer the oracle:
	// it recomputes directly from this function's operations and concrete CFG.
	for _, mutation := range []string{"missing", "shifted"} {
		t.Run(mutation, func(t *testing.T) {
			mutated := function
			linear := *function.Linear
			linear.LoanEndpoints = append([]core.LoanEndpoint(nil), function.Linear.LoanEndpoints...)
			mutated.Linear = &linear
			switch mutation {
			case "missing":
				linear.LoanEndpoints = nil
			case "shifted":
				linear.LoanEndpoints[0].AfterOperationID = function.Parameter.ID
			}
			got, _, err := pathoracle.RecomputeEndpoints(mutated, nil)
			if err != nil {
				t.Fatalf("path replay rejected endpoint mutation: %v", err)
			}
			if !reflect.DeepEqual(got, endpoints) {
				t.Fatalf("path oracle consumed declared endpoint mutation: got %+v, want independently-derived %+v", got, endpoints)
			}
		})
	}

	// A source reference with no preceding loan birth is a retained/forged
	// loan escape across the path boundary and must fail closed.
	forged := function
	forgedLinear := *function.Linear
	forgedLinear.Operations = append([]core.LinearOperation(nil), function.Linear.Operations...)
	forged.Linear = &forgedLinear
	for i := range forgedLinear.Operations {
		if forgedLinear.Operations[i].Kind == core.OpReturn {
			forgedLinear.Operations[i].SourceID = function.ID + ":unknown-loan-place"
		}
	}
	if _, _, err := pathoracle.RecomputeEndpoints(forged, nil); err == nil {
		t.Fatal("path oracle accepted a terminal reference with no independently-derived loan origin")
	}
}

func TestPhase25PointerPathLiveOverlap(t *testing.T) {
	program := checkedProgram(t, "phase3", "sequential_shared_then_exclusive_accept.schway")
	function := functionNamed(t, program, "relay")
	mutated := phase25CloneOwnerTransferProgram(core.Program{Functions: []core.Function{function}}).Functions[0]
	ops := mutated.Linear.Operations
	sharedUse, exclusiveBorrow := -1, -1
	for i, op := range ops {
		if op.Kind == core.OpBorrowShared && sharedUse < 0 {
			for j := i + 1; j < len(ops); j++ {
				if ops[j].SourceID == op.TargetID {
					sharedUse = j
					break
				}
			}
		}
		if op.Kind == core.OpBorrowExclusive {
			exclusiveBorrow = i
		}
	}
	if sharedUse < 0 || exclusiveBorrow < 0 || sharedUse >= exclusiveBorrow {
		t.Fatalf("accepted sequential witness did not have the expected first-loan last use before the exclusive borrow: %+v", ops)
	}
	firstUse := ops[sharedUse]
	moved := append([]core.LinearOperation(nil), ops[:sharedUse]...)
	moved = append(moved, ops[sharedUse+1:]...)
	insertion := exclusiveBorrow
	moved = append(moved[:insertion], append([]core.LinearOperation{firstUse}, moved[insertion:]...)...)
	mutated.Linear.Operations = moved
	for _, endpoints := range [][]core.LoanEndpoint{nil, {{ID: "forged", LoanID: "not-a-real-loan", Kind: "point", AfterOperationID: "fake"}}} {
		t.Run(map[bool]string{true: "forged", false: "empty"}[len(endpoints) != 0], func(t *testing.T) {
			candidate := mutated
			linear := *mutated.Linear
			linear.LoanEndpoints = endpoints
			candidate.Linear = &linear
			if err := pathoracle.ValidateLocalOwnerPaths(core.Program{Functions: []core.Function{candidate}}); err == nil {
				t.Fatal("path oracle accepted a real shared/exclusive overlap")
			}
			if _, _, err := pathoracle.RecomputeEndpoints(candidate, nil); err != nil {
				t.Fatalf("straight-line endpoint synthesis contract changed: %v", err)
			}
		})
	}
}

func TestPhase25PointerPathBorrowedResultEscape(t *testing.T) {
	program := checkedProgram(t, "phase25", "exclusive_copy_accept.schway")
	function := functionNamed(t, program, "exclusive_copy")
	mutated := phase25CloneOwnerTransferProgram(core.Program{Functions: []core.Function{function}}).Functions[0]
	ops := mutated.Linear.Operations
	if len(ops) != 3 || ops[0].Kind != core.OpBorrowExclusive || ops[1].Kind != core.OpCopy || ops[2].Kind != core.OpReturn {
		t.Fatalf("expected exact exclusive U64-copy witness, got %+v", ops)
	}
	borrowedPlace := ops[0].TargetID
	mutated.Linear.Operations[2].SourceID = borrowedPlace
	for _, endpoints := range [][]core.LoanEndpoint{nil, {{ID: "forged", LoanID: "not-a-real-loan", Kind: "point", AfterOperationID: "fake"}}} {
		t.Run(map[bool]string{true: "forged", false: "empty"}[len(endpoints) != 0], func(t *testing.T) {
			candidate := mutated
			linear := *mutated.Linear
			linear.LoanEndpoints = endpoints
			candidate.Linear = &linear
			if err := pathoracle.ValidateLocalOwnerPaths(core.Program{Functions: []core.Function{candidate}}); err == nil {
				t.Fatal("path oracle accepted a genuine borrowed-result escape")
			}
		})
	}
	if err := pathoracle.ValidateLocalOwnerPaths(core.Program{Functions: []core.Function{function}}); err != nil {
		t.Fatalf("path oracle rejected the copied U64 result: %v", err)
	}
}

func TestPhase25PointerPathDefersCallResultProvenance(t *testing.T) {
	program := checkedProgram(t, "phase08", "twin_a_accept.schway")
	function := functionNamed(t, program, "escort")
	if !pathHasCallForTest(function) {
		t.Fatal("interprocedural fixture no longer contains a call")
	}
	if err := pathoracle.ValidateLocalOwnerPaths(core.Program{Functions: []core.Function{function}}); err != nil {
		t.Fatalf("bounded local replay must defer call-result provenance to contract-aware peers: %v", err)
	}
}

func TestPhase25PointerPathAllowsDeclaredBorrowReturn(t *testing.T) {
	program := checkedProgram(t, "phase6", "heldout_borrow_defect.schway")
	function := functionNamed(t, program, "relay")
	if function.PublicOrigin == nil || function.PublicOrigin.Access != "shared" {
		t.Fatalf("fixture return origin changed: got %+v, want declared shared", function.PublicOrigin)
	}
	if err := pathoracle.ValidateLocalOwnerPaths(core.Program{Functions: []core.Function{function}}); err != nil {
		t.Fatalf("bounded local replay rejected an explicitly declared borrowed return: %v", err)
	}
}

func TestPhase25PointerPathChecksOverlapAcrossCalls(t *testing.T) {
	program := checkedProgram(t, "phase11", "multi_function_relay_depth2.schway")
	function := functionNamed(t, program, "relay")
	mutated := phase25CloneOwnerTransferProgram(core.Program{Functions: []core.Function{function}}).Functions[0]
	var call core.LinearOperation
	var exclusiveIndex int = -1
	for index, operation := range mutated.Linear.Operations {
		if operation.Kind == core.OpCall {
			call = operation
		}
		if operation.Kind == core.OpMove {
			// This is the fixture's final `take buffer`; turn it into a
			// second live direct loan of the same owner.
			exclusiveIndex = index
			mutated.Linear.Operations[index].Kind = core.OpBorrowExclusive
			mutated.Linear.Operations[index].LoanID = "injected-exclusive-loan"
		}
	}
	if call.Kind != core.OpCall || exclusiveIndex < 0 {
		t.Fatalf("expected call and final move in relay operations: %+v", mutated.Linear.Operations)
	}
	call.ID = function.ID + ":op:injected-call"
	call.PointID = function.ID + ":point:injected-call"
	ops := mutated.Linear.Operations
	ops = append(ops[:exclusiveIndex+1], append([]core.LinearOperation{call}, ops[exclusiveIndex+1:]...)...)
	mutated.Linear.Operations = ops
	ops[len(ops)-1].SourceID = mutated.Parameter.ID
	for index := range ops {
		ops[index].ID = fmt.Sprintf("%s:op:%d", function.ID, index)
		ops[index].PointID = fmt.Sprintf("%s:point:linear:%d", function.ID, index)
	}
	if err := pathoracle.ValidateLocalOwnerPaths(core.Program{Functions: []core.Function{mutated}}); err == nil || !strings.HasPrefix(err.Error(), "pathoracle.pointer_borrow_conflict:") {
		t.Fatalf("local replay must retain overlap detection across an OpCall use, got %v", err)
	}
}

func TestPhase25PointerPathChecksEscapeInCallFunction(t *testing.T) {
	program := checkedProgram(t, "phase11", "multi_function_relay_depth2.schway")
	function := functionNamed(t, program, "relay")
	mutated := phase25CloneOwnerTransferProgram(core.Program{Functions: []core.Function{function}}).Functions[0]
	if !pathHasCallForTest(mutated) {
		t.Fatal("interprocedural fixture no longer contains a call")
	}
	for index := range mutated.Linear.Operations {
		if mutated.Linear.Operations[index].Kind == core.OpBorrowShared {
			mutated.Linear.Operations[len(mutated.Linear.Operations)-1].SourceID = mutated.Linear.Operations[index].TargetID
			break
		}
	}
	if err := pathoracle.ValidateLocalOwnerPaths(core.Program{Functions: []core.Function{mutated}}); err == nil || !strings.HasPrefix(err.Error(), "pathoracle.pointer_escape:") {
		t.Fatalf("local replay must retain direct borrowed-result escape detection in a function with calls, got %v", err)
	}
}

func pathHasCallForTest(function core.Function) bool {
	if function.Linear == nil {
		return false
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpCall {
			return true
		}
	}
	return false
}

func TestPhase25UtilityOwnerTransfer(t *testing.T) {
	program := phase25OwnerTransferProgram(t)
	if err := pathoracle.ValidateLocalOwnerPaths(program); err != nil {
		t.Fatalf("independent path peer rejected exact Phase 25 owner/result flow: %v", err)
	}
	for _, test := range []struct {
		name   string
		mutate func(*core.Program)
	}{
		{"swapped helpers", func(program *core.Program) {
			main := phase25OwnerFunction(program, "main")
			sharedID, exclusiveID := main.Linear.Operations[2].CalleeID, main.Linear.Operations[3].CalleeID
			main.Linear.Operations[2].CalleeID, main.Linear.Operations[3].CalleeID = exclusiveID, sharedID
		}},
		{"arbitrary helper", func(program *core.Program) {
			phase25OwnerFunction(program, "main").Linear.Operations[2].CalleeID = "untrusted:helper"
		}},
		{"missing exclusive call", func(program *core.Program) {
			main := phase25OwnerFunction(program, "main")
			main.Linear.Operations = append(main.Linear.Operations[:3], main.Linear.Operations[4:]...)
		}},
		{"tampered return", func(program *core.Program) {
			main := phase25OwnerFunction(program, "main")
			main.Linear.Operations[5].SourceID = main.Linear.Operations[2].TargetID
		}},
		{"exclusive argument bypasses shared", func(program *core.Program) {
			main := phase25OwnerFunction(program, "main")
			main.Linear.Operations[3].SourceID = main.Linear.Operations[1].TargetID
		}},
		{"reordered helpers", func(program *core.Program) {
			main := phase25OwnerFunction(program, "main")
			main.Linear.Operations[2], main.Linear.Operations[3] = main.Linear.Operations[3], main.Linear.Operations[2]
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			mutated := phase25CloneOwnerTransferProgram(phase25OwnerTransferProgram(t))
			test.mutate(&mutated)
			if err := pathoracle.ValidateLocalOwnerPaths(mutated); err == nil {
				t.Fatal("path peer accepted a non-exact Phase 25 owner/result chain")
			}
		})
	}
}

func TestPhase25PreservesExactPhase24DirectOwnerRoute(t *testing.T) {
	direct := phase25DirectOwnerTransferProgram(t)
	main := phase25OwnerFunction(&direct, "main")
	if len(direct.Functions) != 2 || main.Linear == nil || len(main.Linear.Operations) != 4 {
		t.Fatalf("direct-route fixture is not the exact Phase 24 shape: functions=%d operations=%d", len(direct.Functions), len(main.Linear.Operations))
	}
	if err := pathoracle.ValidateLocalOwnerPaths(direct); err != nil {
		t.Fatalf("independent path peer rejected the exact direct Phase 24 route: %v", err)
	}

	mutated := phase25CloneOwnerTransferProgram(phase25OwnerTransferProgram(t))
	main = phase25OwnerFunction(&mutated, "main")
	main.Linear.Operations[5].SourceID = main.Linear.Operations[1].TargetID
	if err := pathoracle.ValidateLocalOwnerPaths(mutated); err == nil {
		t.Fatal("path peer accepted a Phase 25 chain whose return was tampered back to the borrowed-use result")
	}
}

func phase25OwnerTransferProgram(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase24", "transfer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("integrated owner-transfer source diagnostics: %+v", checked.Diagnostics)
	}
	return phase25CloneOwnerTransferProgram(checked.Program)
}

func phase25CloneOwnerTransferProgram(program core.Program) core.Program {
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

func phase25OwnerFunction(program *core.Program, name string) *core.Function {
	for index := range program.Functions {
		if program.Functions[index].Name == name {
			return &program.Functions[index]
		}
	}
	panic("Phase 25 owner-transfer fixture function not found: " + name)
}

func phase25DirectOwnerTransferProgram(t *testing.T) core.Program {
	t.Helper()
	program := phase25OwnerTransferProgram(t)
	main := phase25OwnerFunction(&program, "main")
	ops := main.Linear.Operations
	main.Linear.Operations = []core.LinearOperation{ops[0], ops[1], ops[4], ops[5]}
	main.Linear.Operations[3].SourceID = ops[1].TargetID
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
	return program
}
