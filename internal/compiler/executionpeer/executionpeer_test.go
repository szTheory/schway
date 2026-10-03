package executionpeer_test

import (
	"bytes"
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
)

const maxGoListBytes = 1 << 20

type boundedGoListWriter struct {
	buffer bytes.Buffer
	total  int
}

func (w *boundedGoListWriter) Write(data []byte) (int, error) {
	w.total += len(data)
	remaining := maxGoListBytes + 1 - w.buffer.Len()
	if remaining > len(data) {
		remaining = len(data)
	}
	if remaining > 0 {
		_, _ = w.buffer.Write(data[:remaining])
	}
	return len(data), nil
}

func (w *boundedGoListWriter) overflowed() bool { return w.total > maxGoListBytes }

func function(id string, operations ...core.LinearOperation) core.Function {
	return core.Function{ID: id, Name: id, Parameter: core.Parameter{ID: id + ":arg", Type: "Byte"}, Linear: &core.LinearBody{ID: id + ":body", Operations: operations}}
}
func call(id, callee string) core.LinearOperation {
	return core.LinearOperation{ID: id, Kind: core.OpCall, CalleeID: callee}
}
func returned(id string) core.LinearOperation {
	return core.LinearOperation{ID: id, Kind: core.OpReturn}
}
func inv(t *testing.T, entry string, calls ...string) string {
	t.Helper()
	segments := make([]execution.InvocationSegment, len(calls))
	for n, callID := range calls {
		segments[n] = execution.InvocationSegment{OpCallID: callID}
	}
	value, err := execution.FormatInvocation(entry, segments)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func diamondProgram() core.Program {
	return core.Program{Functions: []core.Function{
		function("entry", call("entry:a", "a"), call("entry:b", "b"), returned("entry:return")),
		function("a", call("a:leaf", "leaf"), returned("a:return")),
		function("b", call("b:leaf", "leaf"), returned("b:return")),
		function("leaf", returned("leaf:return")),
	}}
}

func event(invocation, id, kind, functionID string) execution.Event {
	return execution.Event{Schema: execution.Schema2, Invocation: invocation, ID: id, Kind: kind, FunctionID: functionID}
}
func called(invocation, id, functionID, callee string) execution.Event {
	value := event(invocation, id+":event:called", "function.called", functionID)
	value.CalleeFunctionID = callee
	return value
}
func ret(invocation, id, functionID string) execution.Event {
	return event(invocation, id+":event:returned", "function.returned", functionID)
}

func checkedAddOverflowFixture(t *testing.T) (core.Program, execution.Execution) {
	t.Helper()
	const functionID, operationID = "entry", "entry:add"
	invocation := inv(t, functionID)
	program := core.Program{Functions: []core.Function{{
		ID: functionID, Name: functionID, Parameter: core.Parameter{ID: "entry:arg", Type: "U64"}, ReturnType: "U64",
		Linear: &core.LinearBody{
			Types:      []core.TypeFact{{ID: "u64", Shape: core.TypeRef{Constructor: "U64"}}},
			Places:     []core.Place{{ID: "entry:left", TypeID: "u64"}, {ID: "entry:right", TypeID: "u64"}, {ID: "entry:sum", TypeID: "u64"}},
			Operations: []core.LinearOperation{{ID: operationID, Kind: core.OpAddChecked, SourceID: "entry:left", RightID: "entry:right", TargetID: "entry:sum", TypeID: "u64"}, returned("entry:return")},
		},
	}}}
	defect := event(invocation, operationID+":event:defected", "function.defected", functionID)
	defect.SourcePlace, defect.TypeID, defect.Output = "entry:left", "u64", "U64 addition overflow"
	document := execution.Execution{Schema: execution.Schema2, Outcome: execution.Outcome{Kind: execution.OutcomeDefect}, Events: []execution.Event{defect}}
	return program, document
}

func TestCheckedAddOverflowEventIsStructurallyAccepted(t *testing.T) {
	program, document := checkedAddOverflowFixture(t)
	if err := executionpeer.Validate(program, document); err != nil {
		t.Fatalf("source-attributed checked-add defect rejected: %v", err)
	}

	// The peer validates that this terminal event is structurally attributable
	// to a checked-add operation. It has no operand values, so this fixture does
	// not claim the independent peer proved that arithmetic actually overflowed.
	tests := []struct {
		name   string
		mutate func(*execution.Execution)
	}{
		{"wrong kind", func(doc *execution.Execution) { doc.Events[0].Kind = "function.failed" }},
		{"wrong reason", func(doc *execution.Execution) { doc.Events[0].Output = "forged reason" }},
		{"wrong source", func(doc *execution.Execution) { doc.Events[0].SourcePlace = "entry:right" }},
		{"wrong type", func(doc *execution.Execution) { doc.Events[0].TypeID = "bool" }},
		{"wrong outcome", func(doc *execution.Execution) { doc.Outcome.Kind = execution.OutcomeReturned }},
		{"nonempty defect value", func(doc *execution.Execution) { doc.Outcome.Value = "forged" }},
		{"event after terminal", func(doc *execution.Execution) {
			doc.Events = append(doc.Events, ret(inv(t, "entry"), "entry:return", "entry"))
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := document
			mutated.Events = append([]execution.Event(nil), document.Events...)
			test.mutate(&mutated)
			if err := executionpeer.Validate(program, mutated); err == nil {
				t.Fatal("forged checked-add overflow trace was accepted")
			}
		})
	}
}

func phase26OccurrenceFixture(t *testing.T) (core.Program, execution.Execution) {
	t.Helper()
	const functionID = "entry"
	const copyID = "entry:copy"
	const returnID = "entry:return"
	const loopID = "entry:block:loop"
	invocation := inv(t, functionID)
	program := core.Program{Functions: []core.Function{{
		ID: functionID, Name: functionID,
		Parameter: core.Parameter{ID: "entry:arg", Name: "n", Type: "U64"}, ReturnType: "U64",
		Linear: &core.LinearBody{
			Types: []core.TypeFact{{ID: "u64", Shape: core.TypeRef{Constructor: "U64"}}},
			Places: []core.Place{
				{ID: "entry:arg", Name: "n", TypeID: "u64"},
				{ID: "entry:copy-place", Name: "snapshot", TypeID: "u64"},
			},
			Operations: []core.LinearOperation{
				{ID: copyID, PointID: "entry:point:copy", Kind: core.OpCopy, SourceID: "entry:arg", TargetID: "entry:copy-place", TypeID: "u64"},
				{ID: returnID, PointID: "entry:point:return", Kind: core.OpReturn, SourceID: "entry:copy-place", TypeID: "u64"},
			},
			Blocks: []core.Block{
				{ID: "entry:block:entry", OperationIDs: []string{}, Successors: []string{"entry:block:loop"}},
				{ID: loopID, OperationIDs: []string{copyID}, Successors: []string{loopID, "entry:block:exit"}},
				{ID: "entry:block:exit", OperationIDs: []string{returnID}, Successors: []string{}},
			},
			Edges: []core.Edge{
				{ID: "entry:edge:entry-loop", FromBlockID: "entry:block:entry", ToBlockID: loopID},
				{ID: "entry:edge:back", FromBlockID: loopID, ToBlockID: loopID},
				{ID: "entry:edge:exit", FromBlockID: loopID, ToBlockID: "entry:block:exit"},
			},
		},
	}}}
	document := execution.Execution{Schema: execution.Schema2, Outcome: execution.Outcome{Kind: execution.OutcomeReturned, Value: "3"}, Events: []execution.Event{
		{Schema: execution.Schema2, ID: copyID + ":event", Kind: "value.copied", FunctionID: functionID, Invocation: invocation, SourcePlace: "entry:arg", TargetPlace: "entry:copy-place", TypeID: "u64", Occurrence: 0},
		{Schema: execution.Schema2, ID: copyID + ":event", Kind: "value.copied", FunctionID: functionID, Invocation: invocation, SourcePlace: "entry:arg", TargetPlace: "entry:copy-place", TypeID: "u64", Occurrence: 1},
		{Schema: execution.Schema2, ID: copyID + ":event", Kind: "value.copied", FunctionID: functionID, Invocation: invocation, SourcePlace: "entry:arg", TargetPlace: "entry:copy-place", TypeID: "u64", Occurrence: 2},
		{Schema: execution.Schema2, ID: returnID + ":event:returned", Kind: "function.returned", FunctionID: functionID, Invocation: invocation, SourcePlace: "entry:copy-place", TypeID: "u64"},
	}, LiveResources: []string{}}
	return program, document
}

func TestPhase26PeerOccurrenceOrder(t *testing.T) {
	program, document := phase26OccurrenceFixture(t)
	if err := executionpeer.Validate(program, document); err != nil {
		t.Fatalf("complete 0/1/2 copy trace rejected: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*core.Program, *execution.Execution)
	}{
		{name: "duplicate zero", mutate: func(_ *core.Program, doc *execution.Execution) { doc.Events[1].Occurrence = 0 }},
		{name: "gap", mutate: func(_ *core.Program, doc *execution.Execution) { doc.Events[1].Occurrence = 2 }},
		{name: "swapped", mutate: func(_ *core.Program, doc *execution.Execution) {
			doc.Events[1].Occurrence, doc.Events[2].Occurrence = 2, 1
		}},
		{name: "wrong invocation", mutate: func(_ *core.Program, doc *execution.Execution) {
			doc.Events[1].Invocation = inv(t, "entry", "entry:missing")
		}},
		{name: "wrong static site", mutate: func(_ *core.Program, doc *execution.Execution) { doc.Events[0].ID = "entry:forged:event" }},
		{name: "non-scalar cycle event", mutate: func(program *core.Program, doc *execution.Execution) {
			program.Functions[0].Linear.Operations[0].Kind = core.OpMove
			for index := range doc.Events[:3] {
				doc.Events[index].Kind = "value.transferred"
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, document := phase26OccurrenceFixture(t)
			test.mutate(&program, &document)
			if err := executionpeer.Validate(program, document); err == nil {
				t.Fatal("forged occurrence trace was accepted")
			}
		})
	}
}

func validPartial(t *testing.T) execution.Execution {
	root := inv(t, "entry")
	child := inv(t, "entry", "entry:a")
	leaf := inv(t, "entry", "entry:a", "a:leaf")
	return execution.Execution{Schema: execution.Schema2, Events: []execution.Event{called(root, "entry:a", "entry", "a"), called(child, "a:leaf", "a", "leaf"), ret(leaf, "leaf:return", "leaf"), ret(child, "a:return", "a"), ret(root, "entry:return", "entry")}}
}

func TestIndependentEntryResolution(t *testing.T) {
	if err := executionpeer.Validate(diamondProgram(), validPartial(t)); err != nil {
		t.Fatal(err)
	}
	ambiguous := core.Program{Functions: []core.Function{function("a", returned("a:return")), function("b", returned("b:return"))}}
	err := executionpeer.Validate(ambiguous, execution.Execution{Schema: execution.Schema2, Events: []execution.Event{}})
	if err == nil || !strings.Contains(err.Error(), "entry_ambiguous") {
		t.Fatalf("want named ambiguity, got %v", err)
	}
}

func TestInvocationMembershipTraversal(t *testing.T) {
	if err := executionpeer.Validate(diamondProgram(), validPartial(t)); err != nil {
		t.Fatal(err)
	}
	// The shared leaf has two distinct occurrence paths, and an unobserved
	// branch is valid for the general observed-structure validator.
	if err := executionpeer.ValidateFullCoverage(diamondProgram(), validPartial(t)); err == nil || !strings.Contains(err.Error(), "full_coverage_missing") {
		t.Fatalf("full control = %v, want missing branch", err)
	}
}

func TestInvocationMembershipTraversalBound(t *testing.T) {
	functions := make([]core.Function, executionpeer.MaxInvocations+1)
	for n := range functions {
		id := fmt.Sprintf("f:%d", n)
		operations := []core.LinearOperation{returned(id + ":return")}
		if n+1 < len(functions) {
			operations = append([]core.LinearOperation{call(id+":call", fmt.Sprintf("f:%d", n+1))}, operations...)
		}
		functions[n] = function(id, operations...)
	}
	err := executionpeer.Validate(core.Program{Functions: functions}, execution.Execution{Schema: execution.Schema2})
	if err == nil || !strings.Contains(err.Error(), "traversal_exhausted") || !strings.Contains(err.Error(), "limit 4096") {
		t.Fatalf("bound error = %v", err)
	}
}

func TestExecutionPeerImportBoundary(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "list", "-deps", "-f", "{{.ImportPath}}", "github.com/szTheory/schway/internal/compiler/executionpeer")
	var stdout, stderr boundedGoListWriter
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("go list dependency probe timed out: %v", ctx.Err())
	}
	if stdout.overflowed() || stderr.overflowed() {
		t.Fatalf("go list dependency probe exceeded %d bytes (stdout=%d stderr=%d)", maxGoListBytes, stdout.total, stderr.total)
	}
	if err != nil {
		t.Fatalf("go list dependency probe: %v: %s", err, stderr.buffer.Bytes())
	}
	for _, path := range strings.Fields(stdout.buffer.String()) {
		if strings.HasSuffix(path, "/interp") || strings.HasSuffix(path, "/cgen") {
			t.Fatalf("forbidden dependency %q", path)
		}
	}
}

func TestFullCoverageControlIsSeparate(t *testing.T) {
	partial := validPartial(t)
	if err := executionpeer.Validate(diamondProgram(), partial); err != nil {
		t.Fatalf("general validation overreached: %v", err)
	}
	if err := executionpeer.ValidateFullCoverage(diamondProgram(), partial); err == nil {
		t.Fatal("full coverage accepted partial observation")
	}
}

func TestValidateObservedCausalStructure(t *testing.T) {
	if err := executionpeer.Validate(diamondProgram(), validPartial(t)); err != nil {
		t.Fatal(err)
	}
}

func TestExecutionPeerFailuresAreActionable(t *testing.T) {
	base := validPartial(t)
	mutations := []struct {
		name, want string
		change     func(*execution.Execution)
	}{
		{"forged parent", "preorder", func(d *execution.Execution) { d.Events[1] = called(inv(t, "entry", "entry:b"), "b:leaf", "b", "leaf") }},
		{"callee", "callee", func(d *execution.Execution) { d.Events[0].CalleeFunctionID = "b" }},
		{"function", "function", func(d *execution.Execution) { d.Events[0].FunctionID = "a" }},
		{"duplicate pair", "duplicate_pair", func(d *execution.Execution) { d.Events[1] = d.Events[0] }},
		{"orphan", "orphan_child", func(d *execution.Execution) { d.Events = d.Events[1:] }},
		{"malformed", "malformed_invocation", func(d *execution.Execution) { d.Events[0].Invocation = "not-an-invocation" }},
		{"missing", "missing_field", func(d *execution.Execution) { d.Events[0].ID = "" }},
		{"kind", "unknown_kind", func(d *execution.Execution) { d.Events[0].Kind = "unknown" }},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			document := base
			document.Events = append([]execution.Event(nil), base.Events...)
			mutation.change(&document)
			err := executionpeer.Validate(diamondProgram(), document)
			if err == nil || !strings.Contains(err.Error(), mutation.want) || !strings.Contains(err.Error(), "event index") {
				t.Fatalf("error = %v, want actionable %q", err, mutation.want)
			}
		})
	}
}

func TestSourceParses(t *testing.T) {
	if _, err := parser.ParseFile(token.NewFileSet(), "executionpeer.go", nil, 0); err != nil {
		t.Fatal(err)
	}
}
