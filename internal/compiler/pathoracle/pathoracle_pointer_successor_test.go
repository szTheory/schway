package pathoracle_test

import (
	"reflect"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
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
