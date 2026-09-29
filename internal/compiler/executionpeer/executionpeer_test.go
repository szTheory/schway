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
