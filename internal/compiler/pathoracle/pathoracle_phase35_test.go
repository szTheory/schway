package pathoracle_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
)

func TestPhase35PureOperationAncestry(t *testing.T) {
	for _, kind := range []core.OperationKind{
		core.OpBitXorU64,
		core.OpBitAndU64,
		core.OpBitOrU64,
		core.OpShiftLeftU64,
		core.OpShiftRightU64,
	} {
		t.Run(string(kind), func(t *testing.T) {
			function := phase35LoanedScalarFunction(kind, false)
			program := core.Program{Functions: []core.Function{function}}
			if err := pathoracle.ValidateLocalOwnerPaths(program); err != nil {
				t.Fatalf("pure %s result retained borrowed-view ancestry: %v", kind, err)
			}

			function = phase35LoanedScalarFunction(kind, true)
			endpoints, _, err := pathoracle.RecomputeEndpoints(function, nil)
			if err != nil {
				t.Fatalf("path replay for %s failed: %v", kind, err)
			}
			if got := phase35EndpointAfter(t, endpoints, "loan:view"); got != "op:pure" {
				t.Fatalf("loan ended after %q, want the %s source-use operation", got, kind)
			}
		})
	}
}

func phase35LoanedScalarFunction(kind core.OperationKind, withBlock bool) core.Function {
	const (
		viewType = "type:view"
		u64Type  = "type:u64"
	)
	operations := []core.LinearOperation{
		{ID: "op:borrow", Kind: core.OpBorrowShared, SourceID: "place:arg", TargetID: "place:view", TypeID: viewType, LoanID: "loan:view"},
		{ID: "op:index", Kind: core.OpConst, TargetID: "place:index", TypeID: u64Type, ConstU64: "0"},
		{ID: "op:read", Kind: core.OpBytesAt, SourceID: "place:view", RightID: "place:index", TargetID: "place:byte", TypeID: u64Type},
		{ID: "op:one", Kind: core.OpConst, TargetID: "place:one", TypeID: u64Type, ConstU64: "1"},
		{ID: "op:pure", Kind: kind, SourceID: "place:byte", TargetID: "place:result", TypeID: u64Type, RightID: "place:one", ConstU64: "1"},
		{ID: "op:borrow-exclusive", Kind: core.OpBorrowExclusive, SourceID: "place:arg", TargetID: "place:exclusive", TypeID: viewType, LoanID: "loan:exclusive"},
		{ID: "op:return", Kind: core.OpReturn, SourceID: "place:result", TypeID: u64Type},
	}
	if kind == core.OpShiftLeftU64 || kind == core.OpShiftRightU64 {
		operations[4].RightID = ""
	}
	function := core.Function{
		ID: "function:main", Name: "main", Parameter: core.Parameter{ID: "place:arg", Type: "ByteView"},
		ReturnType: "U64", EntryPointID: "point:entry",
		Linear: &core.LinearBody{
			ID: "linear:main",
			Types: []core.TypeFact{
				{ID: viewType, Shape: core.TypeRef{Constructor: "ByteView"}},
				{ID: u64Type, Shape: core.TypeRef{Constructor: "U64"}},
			},
			Places: []core.Place{
				{ID: "place:arg", TypeID: viewType},
				{ID: "place:view", TypeID: viewType},
				{ID: "place:index", TypeID: u64Type},
				{ID: "place:byte", TypeID: u64Type},
				{ID: "place:one", TypeID: u64Type},
				{ID: "place:result", TypeID: u64Type},
				{ID: "place:exclusive", TypeID: viewType},
			},
			Operations: operations,
		},
	}
	if withBlock {
		operationIDs := make([]string, len(operations))
		for index, operation := range operations {
			operationIDs[index] = operation.ID
		}
		function.Linear.Blocks = []core.Block{{ID: "block:entry", PointID: "point:entry", OperationIDs: operationIDs}}
	}
	return function
}

func phase35EndpointAfter(t *testing.T, endpoints []core.LoanEndpoint, loanID string) string {
	t.Helper()
	for _, endpoint := range endpoints {
		if endpoint.LoanID == loanID && endpoint.Kind == "point" {
			return endpoint.AfterOperationID
		}
	}
	t.Fatalf("no point endpoint found for %s in %+v", loanID, endpoints)
	return ""
}
