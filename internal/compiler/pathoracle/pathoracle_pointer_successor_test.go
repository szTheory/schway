package pathoracle_test

import (
	"os"
	"reflect"
	"strconv"
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
