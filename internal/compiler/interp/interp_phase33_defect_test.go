package interp

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
)

const phase33OwnerOperationID = "main:op:0"

func phase33AbortLikeRow(kind core.OperationKind, reason string) (core.Program, Execution) {
	const operationID = "f:op:1"
	return core.Program{Functions: []core.Function{{ID: "f", Name: "f", Linear: &core.LinearBody{
			Operations: []core.LinearOperation{{ID: operationID, Kind: kind, Reason: reason}},
		}}}}, Execution{
			Outcome: execution.Outcome{Kind: execution.OutcomeDefect},
			Events: []Event{
				{Kind: "foreign.called", ID: "owner:event"},
				{Kind: "function.defected", ID: operationID + ":event:defected", FunctionID: "f", Output: reason},
			},
			LiveResources: []string{phase33OwnerOperationID},
		}
}

func TestPhase33AbortLikeNarrowed(t *testing.T) {
	accepted := []struct {
		name   string
		kind   core.OperationKind
		reason string
	}{
		{"byte index", core.OpBytesAt, execution.DefectReasonByteIndexOutOfBounds},
		{"checked add", core.OpAddChecked, execution.DefectReasonU64AddOverflow},
		{"remainder", core.OpRemainderU64, execution.DefectReasonU64RemainderByZero},
		{"user defect", core.OpDefect, "boom"},
	}
	for _, row := range accepted {
		t.Run("accept/"+row.name, func(t *testing.T) {
			program, result := phase33AbortLikeRow(row.kind, row.reason)
			if !isAbortLikeDefect(program, result, phase33OwnerOperationID) {
				t.Fatal("valid abort-like defect refused")
			}
		})
	}
	type refusedRow struct {
		name    string
		program core.Program
		result  Execution
	}
	refuse := func(name string, build func() (core.Program, Execution)) refusedRow {
		program, result := build()
		return refusedRow{name: name, program: program, result: result}
	}
	refused := []refusedRow{
		refuse("wrong reason bytes", func() (core.Program, Execution) {
			return phase33AbortLikeRow(core.OpAddChecked, execution.DefectReasonU64AddOverflow+"!")
		}),
		refuse("extra live resource", func() (core.Program, Execution) {
			p, e := phase33AbortLikeRow(core.OpAddChecked, execution.DefectReasonU64AddOverflow)
			e.LiveResources = []string{"main:op:0", "x:op:9"}
			return p, e
		}),
		refuse("missing owner", func() (core.Program, Execution) {
			p, e := phase33AbortLikeRow(core.OpAddChecked, execution.DefectReasonU64AddOverflow)
			e.LiveResources = nil
			return p, e
		}),
		refuse("duplicate owner", func() (core.Program, Execution) {
			p, e := phase33AbortLikeRow(core.OpAddChecked, execution.DefectReasonU64AddOverflow)
			e.LiveResources = []string{"main:op:0", "main:op:0"}
			return p, e
		}),
		refuse("wrong function adjacency", func() (core.Program, Execution) {
			p, e := phase33AbortLikeRow(core.OpAddChecked, execution.DefectReasonU64AddOverflow)
			p.Functions[0].ID = "a"
			p.Functions = append(p.Functions, core.Function{ID: "b", Name: "b", Linear: &core.LinearBody{}})
			e.Events[len(e.Events)-1].FunctionID = "b"
			return p, e
		}),
		refuse("call", func() (core.Program, Execution) { return phase33AbortLikeRow(core.OpCall, "x") }),
		refuse("write", func() (core.Program, Execution) { return phase33AbortLikeRow(core.OpWriteU64, "x") }),
		refuse("unknown op", func() (core.Program, Execution) {
			p, e := phase33AbortLikeRow(core.OpAddChecked, execution.DefectReasonU64AddOverflow)
			p.Functions[0].Linear.Operations[0].ID = "missing"
			return p, e
		}),
		refuse("wrong event kind", func() (core.Program, Execution) {
			p, e := phase33AbortLikeRow(core.OpAddChecked, execution.DefectReasonU64AddOverflow)
			e.Events[1].Kind = "function.returned"
			return p, e
		}),
		refuse("returned outcome", func() (core.Program, Execution) {
			p, e := phase33AbortLikeRow(core.OpAddChecked, execution.DefectReasonU64AddOverflow)
			e.Outcome.Kind = execution.OutcomeReturned
			return p, e
		}),
		refuse("typed failure", func() (core.Program, Execution) {
			p, e := phase33AbortLikeRow(core.OpAddChecked, execution.DefectReasonU64AddOverflow)
			e.Outcome.Kind = execution.OutcomeTypedFailure
			return p, e
		}),
		refuse("empty events", func() (core.Program, Execution) {
			p, e := phase33AbortLikeRow(core.OpAddChecked, execution.DefectReasonU64AddOverflow)
			e.Events = nil
			return p, e
		}),
	}
	for _, row := range refused {
		t.Run("refuse/"+row.name, func(t *testing.T) {
			if isAbortLikeDefect(row.program, row.result, phase33OwnerOperationID) {
				t.Fatal("invalid abort-like defect accepted")
			}
		})
	}
}

func TestPhase33AbortLikeWideningMutantKilled(t *testing.T) {
	t.Run("live resource acceptance", func(t *testing.T) {
		program, result := phase33AbortLikeRow(core.OpAddChecked, execution.DefectReasonU64AddOverflow)
		result.LiveResources = []string{phase33OwnerOperationID, "x:op:9"}
		if isAbortLikeDefect(program, result, phase33OwnerOperationID) {
			t.Fatal("production predicate accepted the extra live resource before mutation")
		}
		original := abortLikeAcceptLive
		defer func() { abortLikeAcceptLive = original }()
		abortLikeAcceptLive = func([]string, string) bool { return true }
		if !isAbortLikeDefect(program, result, phase33OwnerOperationID) {
			t.Fatal("widening abortLikeAcceptLive did not accept the extra-resource row")
		}
		abortLikeAcceptLive = original
		if isAbortLikeDefect(program, result, phase33OwnerOperationID) {
			t.Fatal("restored production live-resource rule accepted the extra-resource row")
		}
	})

	t.Run("reason matching", func(t *testing.T) {
		program, result := phase33AbortLikeRow(core.OpAddChecked, execution.DefectReasonU64AddOverflow+"!")
		if isAbortLikeDefect(program, result, phase33OwnerOperationID) {
			t.Fatal("production predicate accepted the wrong reason before mutation")
		}
		original := abortLikeReasonMatches
		defer func() { abortLikeReasonMatches = original }()
		abortLikeReasonMatches = func(string, string) bool { return true }
		if !isAbortLikeDefect(program, result, phase33OwnerOperationID) {
			t.Fatal("widening abortLikeReasonMatches did not accept the wrong-reason row")
		}
		abortLikeReasonMatches = original
		if isAbortLikeDefect(program, result, phase33OwnerOperationID) {
			t.Fatal("restored production reason matcher accepted the wrong-reason row")
		}
	})
}
