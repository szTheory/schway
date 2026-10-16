package executionpeer_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase371AdlerDefectFunction = `fn adler32(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  if n < 1 {
    defect "legacy child defect"
  } else {
    n = n + 0
  }
  n
}
`

const phase371DynamicDefectSource = `module phase371.dynamic_child_defect
export { fn main }
fn increment(value: U64) -> U64 {
  let result = value + 1
  result
}
fn main(input: U64) -> U64 {
  var value = input
  var index = 0
  while index < 3 {
    value = increment(value)
    index = index + 1
  }
  if input == 1 {
    defect "dynamic caller defect"
  } else {
    value = value
  }
  value
}
`

func TestPhase371LegacyDefectClosesAllActiveFrames(t *testing.T) {
	program, entry, leaf, call, sourceDigest := phase371LegacyDefectProgram(t)
	t.Logf("phase371 source witness sha256=%x", sourceDigest)
	if leaf.Parameter.Type != "ByteView" || leaf.ReturnType != "U64" {
		t.Fatalf("callee signature = %s -> %s, want the legacy non-scalar ByteView -> U64 route", leaf.Parameter.Type, leaf.ReturnType)
	}
	t.Run("legacy-static", func(t *testing.T) {
		defectRun, err := interp.RunWithFileBytes(program, "main", "phase371-empty", func(string) ([]byte, string) {
			return []byte{}, ""
		})
		if err != nil {
			t.Fatalf("interpreter legacy child defect: %v", err)
		}
		if defectRun.Outcome.Kind != execution.OutcomeDefect || defectRun.Outcome.Value != "" {
			t.Fatalf("defect outcome = %+v, want an empty-valued defect", defectRun.Outcome)
		}
		if len(defectRun.Events) == 0 {
			t.Fatal("interpreter legacy child defect produced no events")
		}
		last := defectRun.Events[len(defectRun.Events)-1]
		child, err := execution.FormatInvocation(entry.ID, []execution.InvocationSegment{{OpCallID: call.ID, Ordinal: 0}})
		if err != nil {
			t.Fatalf("format legacy child invocation: %v", err)
		}
		if last.Kind != "function.defected" || last.FunctionID != leaf.ID || last.Invocation != child || last.Output != "legacy child defect" {
			t.Fatalf("final interpreter event = %+v, want leaf %q at %q with the exact defect reason", last, leaf.ID, child)
		}
		peerDocument, err := session.ProjectExecutionSchema2(program, defectRun)
		if err != nil {
			t.Fatalf("project legacy defect trace: %v", err)
		}
		if err := executionpeer.Validate(program, peerDocument); err != nil {
			t.Fatalf("peer rejected actual legacy child defect trace: %v", err)
		}
	})
	t.Run("ordinary-nested-return", func(t *testing.T) {
		returnedRun, err := interp.RunWithFileBytes(program, "main", "phase371-nonempty", func(string) ([]byte, string) {
			return []byte("xx"), ""
		})
		if err != nil {
			t.Fatalf("interpreter nested return: %v", err)
		}
		if returnedRun.Outcome.Kind != execution.OutcomeReturned || returnedRun.Outcome.Value != "2" {
			t.Fatalf("nested return outcome = %+v, want returned 2", returnedRun.Outcome)
		}
		returnedPeerDocument, err := session.ProjectExecutionSchema2(program, returnedRun)
		if err != nil {
			t.Fatalf("project nested return trace: %v", err)
		}
		if err := executionpeer.Validate(program, returnedPeerDocument); err != nil {
			t.Fatalf("peer rejected ordinary nested return trace: %v", err)
		}
	})
	t.Run("dynamic-scalar", func(t *testing.T) {
		program, entry, callee, call := phase371DynamicDefectProgram(t)
		if callee.Parameter.Type != "U64" || callee.ReturnType != "U64" {
			t.Fatalf("dynamic callee signature = %s -> %s, want U64 -> U64", callee.Parameter.Type, callee.ReturnType)
		}
		run, err := interp.RunApplication(program, "main", "18446744073709551613")
		if err != nil {
			t.Fatalf("interpreter dynamic child overflow: %v", err)
		}
		if run.Execution.Outcome.Kind != execution.OutcomeDefect || run.Execution.Outcome.Value != "" {
			t.Fatalf("dynamic defect outcome = %+v, want an empty-valued defect", run.Execution.Outcome)
		}
		root, err := execution.FormatInvocation(entry.ID, nil)
		if err != nil {
			t.Fatalf("format dynamic root invocation: %v", err)
		}
		var calls, returned, childTerminals int
		for _, event := range run.Execution.Events {
			if event.Kind == "function.called" && event.CalleeFunctionID == callee.ID {
				calls++
				if event.Invocation != root {
					t.Fatalf("dynamic call event parent = %q, want root %q", event.Invocation, root)
				}
			}
			if event.Kind == "function.returned" && event.FunctionID == callee.ID {
				returned++
			}
			if event.FunctionID == callee.ID && (event.Kind == "function.returned" || event.Kind == "function.defected") {
				childTerminals++
				invocation, err := execution.ParseInvocation(event.Invocation)
				if err != nil || len(invocation.Segments) != 1 || invocation.Segments[0].OpCallID != call.ID || invocation.Segments[0].Ordinal != childTerminals-1 {
					t.Fatalf("dynamic child terminal %d has wrong occurrence identity: event=%+v parsed=%+v err=%v", childTerminals, event, invocation, err)
				}
			}
		}
		if calls != 3 || returned != 2 || childTerminals != 3 {
			t.Fatalf("successful earlier activations/call attempts/terminals = %d/%d/%d, want 2/3/3", returned, calls, childTerminals)
		}
		last := run.Execution.Events[len(run.Execution.Events)-1]
		if last.Kind != "function.defected" || last.FunctionID != callee.ID {
			t.Fatalf("final dynamic event = %+v, want the callee defect", last)
		}
		peerDocument, err := session.ProjectExecutionSchema2(program, run.Execution)
		if err != nil {
			t.Fatalf("project dynamic defect trace: %v", err)
		}
		if err := executionpeer.Validate(program, peerDocument); err != nil {
			t.Fatalf("peer rejected actual dynamic child defect trace: %v", err)
		}
	})
}

func TestPhase371ForeignFailureOpFailAttribution(t *testing.T) {
	program, acquire, use := phase371Phase24FailureProgram(t)
	for _, scenario := range []struct {
		name     string
		failure  core.LinearOperation
		outcomes map[string]interp.ForeignOutcome
	}{
		{
			name:    "AcquireError.AcquireFailed",
			failure: acquire,
			outcomes: map[string]interp.ForeignOutcome{
				acquire.ID: {Kind: "failure", Type: "AcquireError", Value: "AcquireFailed"},
			},
		},
		{
			name:    "UseError.UnsupportedByte",
			failure: use,
			outcomes: map[string]interp.ForeignOutcome{
				acquire.ID: {Kind: "success", Type: acquire.Foreign.ResultType, Value: "67"},
				use.ID:     {Kind: "failure", Type: "UseError", Value: "UnsupportedByte"},
			},
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			model, err := interp.RunWithForeignOutcomes(program, "main", "opaque-path-token", scenario.outcomes)
			if err != nil {
				t.Fatalf("checked Phase24 model execution: %v", err)
			}
			if model.EvidenceScope != interp.EvidenceScopeModelOnly || model.ActualHostIO || model.PhysicalCleanup {
				t.Fatalf("failure execution crossed the model-only evidence boundary: %+v", model)
			}
			document, err := session.ProjectExecutionSchema2(program, model.Execution)
			if err != nil {
				t.Fatalf("project checked Phase24 failure trace: %v", err)
			}
			if document.Outcome.Kind != execution.OutcomeTypedFailure {
				t.Fatalf("model outcome = %+v, want typed_failure", document.Outcome)
			}
			failureIndex := -1
			for index, event := range document.Events {
				if event.ID == scenario.failure.ID+":event" && event.Kind == "foreign.failed" {
					failureIndex = index
					if event.FunctionID != phase371FailureFunction(program, scenario.failure) || event.SourcePlace != scenario.failure.SourceID || event.TargetPlace != scenario.failure.TargetID || event.TypeID != scenario.failure.Foreign.Fails {
						t.Fatalf("foreign failure attribution = %+v, want indexed operation %+v", event, scenario.failure)
					}
					break
				}
			}
			if failureIndex < 0 {
				t.Fatalf("model trace omitted exact indexed foreign.failed for %q: %+v", scenario.failure.ID, document.Events)
			}
			failureEvent := document.Events[failureIndex]
			var calleeTerminal, rootTerminal bool
			var parentRouteWork bool
			for _, event := range document.Events[failureIndex+1:] {
				if event.Kind == "resource.released" && event.Invocation == rootInvocationForPhase371(t, program) {
					parentRouteWork = true
				}
				if event.Kind != "function.failed" {
					continue
				}
				if event.Invocation == failureEvent.Invocation && event.FunctionID == failureEvent.FunctionID {
					calleeTerminal = true
				}
				if event.Invocation == rootInvocationForPhase371(t, program) {
					rootTerminal = true
				}
			}
			if !calleeTerminal || !rootTerminal {
				t.Fatalf("foreign failure did not close one child frame and reach the root error terminal: events=%+v", document.Events[failureIndex:])
			}
			if scenario.name == "UseError.UnsupportedByte" && !parentRouteWork {
				t.Fatalf("parent did not perform its admitted error-route release work after child failure: events=%+v", document.Events[failureIndex:])
			}
			if !executionpeerValidatePhase371(program, document) {
				t.Fatalf("peer rejected checked %s foreign.failed → OpFail trace: %v; events=%+v", scenario.name, executionpeer.Validate(program, document), document.Events)
			}
		})
	}

	cleanOutcomes := map[string]interp.ForeignOutcome{
		acquire.ID: {Kind: "success", Type: acquire.Foreign.ResultType, Value: "65"},
		use.ID:     {Kind: "success", Type: "U64", Value: "65"},
	}
	clean, err := interp.RunWithForeignOutcomes(program, "main", "opaque-path-token", cleanOutcomes)
	if err != nil {
		t.Fatalf("checked Phase24 clean model execution: %v", err)
	}
	if clean.Execution.Outcome.Kind != execution.OutcomeReturned {
		t.Fatalf("clean model outcome = %+v, want returned", clean.Execution.Outcome)
	}
	cleanDocument, err := session.ProjectExecutionSchema2(program, clean.Execution)
	if err != nil {
		t.Fatalf("project checked Phase24 success trace: %v", err)
	}
	if err := executionpeer.Validate(program, cleanDocument); err != nil {
		t.Fatalf("peer rejected checked Phase24 foreign.called and ordinary-return trace: %v", err)
	}
}

func phase371FailureFunction(program core.Program, operation core.LinearOperation) string {
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, candidate := range function.Linear.Operations {
			if candidate.ID == operation.ID {
				return function.ID
			}
		}
	}
	return ""
}

func phase371Phase24FailureProgram(t *testing.T) (core.Program, core.LinearOperation, core.LinearOperation) {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase24", "error.schway"))
	if err != nil {
		t.Fatalf("read actual Phase24 error source: %v", err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("actual Phase24 source check diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("actual Phase24 source core validation problems: %+v", validated.Problems)
	}
	program := validated.Program()
	var acquire, use core.LinearOperation
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Foreign == nil {
				continue
			}
			switch {
			case function.Name == "acquire" && operation.Foreign.Mode == "acquire":
				acquire = operation
			case function.Name == "probe" && operation.Foreign.Mode == "borrow":
				use = operation
			}
		}
	}
	if acquire.ID == "" || use.ID == "" || acquire.Foreign.Fails != "AcquireError" || use.Foreign.Fails != "UseError" {
		t.Fatalf("checked Phase24 foreign failure operations = acquire:%+v use:%+v", acquire, use)
	}
	return program, acquire, use
}

func phase371Phase24FailureDocument(t *testing.T, program core.Program, acquire, use core.LinearOperation, scenario string) execution.Execution {
	t.Helper()
	outcomes := make(map[string]interp.ForeignOutcome)
	switch scenario {
	case "acquire":
		outcomes[acquire.ID] = interp.ForeignOutcome{Kind: "failure", Type: "AcquireError", Value: "AcquireFailed"}
	case "use":
		outcomes[acquire.ID] = interp.ForeignOutcome{Kind: "success", Type: acquire.Foreign.ResultType, Value: "67"}
		outcomes[use.ID] = interp.ForeignOutcome{Kind: "failure", Type: "UseError", Value: "UnsupportedByte"}
	case "success":
		outcomes[acquire.ID] = interp.ForeignOutcome{Kind: "success", Type: acquire.Foreign.ResultType, Value: "65"}
		outcomes[use.ID] = interp.ForeignOutcome{Kind: "success", Type: "U64", Value: "65"}
	default:
		t.Fatalf("unsupported Phase24 trace scenario %q", scenario)
	}
	model, err := interp.RunWithForeignOutcomes(program, "main", "opaque-path-token", outcomes)
	if err != nil {
		t.Fatalf("checked Phase24 %s model execution: %v", scenario, err)
	}
	if model.EvidenceScope != interp.EvidenceScopeModelOnly || model.ActualHostIO || model.PhysicalCleanup {
		t.Fatalf("Phase24 %s trace crossed model-only evidence boundary: %+v", scenario, model)
	}
	document, err := session.ProjectExecutionSchema2(program, model.Execution)
	if err != nil {
		t.Fatalf("project checked Phase24 %s trace: %v", scenario, err)
	}
	return document
}

func rootInvocationForPhase371(t *testing.T, program core.Program) string {
	t.Helper()
	for _, function := range program.Functions {
		if function.Name == "main" {
			invocation, err := execution.FormatInvocation(function.ID, nil)
			if err != nil {
				t.Fatalf("format Phase24 root invocation: %v", err)
			}
			return invocation
		}
	}
	t.Fatal("checked Phase24 source has no main function")
	return ""
}

func executionpeerValidatePhase371(program core.Program, document execution.Execution) bool {
	return executionpeer.Validate(program, document) == nil
}

func TestPhase371ForeignFailureAttributionControls(t *testing.T) {
	program, acquire, use := phase371Phase24FailureProgram(t)
	useFailure := phase371Phase24FailureDocument(t, program, acquire, use, "use")
	acquireFailure := phase371Phase24FailureDocument(t, program, acquire, use, "acquire")
	clean := phase371Phase24FailureDocument(t, program, acquire, use, "success")
	if err := executionpeer.Validate(program, useFailure); err != nil {
		t.Fatalf("fresh UseError child-to-root failure rejected: %v", err)
	}
	if err := executionpeer.Validate(program, acquireFailure); err != nil {
		t.Fatalf("fresh AcquireError child-to-root failure rejected: %v", err)
	}
	if err := executionpeer.Validate(program, clean); err != nil {
		t.Fatalf("fresh ordinary-return trace rejected: %v", err)
	}
	if err := executionpeer.Validate(program, clonePhase371Execution(clean)); err != nil {
		t.Fatalf("no-op ordinary-return trace rejected: %v", err)
	}

	root := rootInvocationForPhase371(t, program)
	foreignIndex := phase371FindEventIndex(useFailure, func(event execution.Event) bool { return event.Kind == "foreign.failed" })
	childTerminalIndex := phase371FindEventIndex(useFailure, func(event execution.Event) bool {
		return event.Kind == "function.failed" && event.Invocation != root
	})
	rootTerminalIndex := phase371FindEventIndex(useFailure, func(event execution.Event) bool {
		return event.Kind == "function.failed" && event.Invocation == root
	})
	callIndex := phase371FindEventIndex(useFailure, func(event execution.Event) bool {
		return event.Kind == "function.called" && event.CalleeFunctionID == phase371FunctionID(program, "probe")
	})
	if foreignIndex < 0 || childTerminalIndex < 0 || rootTerminalIndex < 0 || callIndex < 0 {
		t.Fatalf("checked UseError trace omitted required call/failure/terminal events: %+v", useFailure.Events)
	}
	if !phase371HasEventAfter(useFailure, childTerminalIndex, func(event execution.Event) bool { return event.Kind == "resource.released" && event.Invocation == root }) {
		t.Fatal("child failure did not leave the parent active for admitted error-route release work")
	}
	if err := executionpeer.Validate(program, clonePhase371Execution(useFailure)); err != nil {
		t.Fatalf("deep-copy failure trace rejected: %v", err)
	}

	wrongCallID := "missing-operation:event"
	for _, control := range []struct {
		name     string
		wantCode string
		mutate   func(*core.Program, *execution.Execution)
	}{
		{"nonexistent-foreign-operation", "executionpeer.unknown_kind", func(_ *core.Program, document *execution.Execution) { document.Events[foreignIndex].ID = wrongCallID }},
		{"different-foreign-operation", "executionpeer.ownership", func(_ *core.Program, document *execution.Execution) {
			document.Events[foreignIndex].ID = acquire.ID + ":event"
		}},
		{"wrong-function", "executionpeer.function", func(_ *core.Program, document *execution.Execution) {
			document.Events[foreignIndex].FunctionID = phase371FunctionID(program, "main")
		}},
		{"wrong-invocation", "executionpeer.function", func(_ *core.Program, document *execution.Execution) { document.Events[foreignIndex].Invocation = root }},
		{"wrong-source", "executionpeer.foreign_failure_attribution", func(_ *core.Program, document *execution.Execution) {
			document.Events[foreignIndex].SourcePlace = "forged-source"
		}},
		{"wrong-target", "executionpeer.foreign_failure_attribution", func(_ *core.Program, document *execution.Execution) {
			document.Events[foreignIndex].TargetPlace = "forged-target"
		}},
		{"wrong-type", "executionpeer.foreign_failure_attribution", func(_ *core.Program, document *execution.Execution) {
			document.Events[foreignIndex].TypeID = "AcquireError"
		}},
		{"failure-from-different-call", "executionpeer.unknown_kind", func(_ *core.Program, document *execution.Execution) {
			document.Events[foreignIndex].ID = phase371MainCallID(program, "acquire") + ":event"
		}},
		{"wrong-foreign-contract", "executionpeer.foreign_failure_attribution", func(candidate *core.Program, _ *execution.Execution) {
			phase371MutateOperation(candidate, use.ID, func(operation *core.LinearOperation) { operation.Foreign.Fails = "AcquireError" })
		}},
		{"wrong-foreign-error-edge", "executionpeer.foreign_failure_attribution", func(candidate *core.Program, _ *execution.Execution) {
			phase371MutateOperation(candidate, use.ID, func(operation *core.LinearOperation) { operation.ErrEdgeID = operation.OkEdgeID })
		}},
		{"wrong-caller-source", "executionpeer.call_attribution", func(_ *core.Program, document *execution.Execution) {
			document.Events[callIndex].SourcePlace = "forged-source"
		}},
		{"wrong-caller-target", "executionpeer.call_attribution", func(_ *core.Program, document *execution.Execution) {
			document.Events[callIndex].TargetPlace = "forged-target"
		}},
		{"wrong-caller-type", "executionpeer.call_attribution", func(_ *core.Program, document *execution.Execution) {
			document.Events[callIndex].TypeID = "forged-type"
		}},
		{"wrong-caller-error-target", "executionpeer.call_error_route", func(candidate *core.Program, _ *execution.Execution) {
			phase371MutateOperation(candidate, phase371MainCallID(program, "probe"), func(operation *core.LinearOperation) { operation.ErrTargetID = operation.TargetID })
		}},
		{"wrong-caller-error-edge", "executionpeer.call_error_route", func(candidate *core.Program, _ *execution.Execution) {
			phase371MutateOperation(candidate, phase371MainCallID(program, "probe"), func(operation *core.LinearOperation) { operation.ErrEdgeID = operation.OkEdgeID })
		}},
		{"wrong-terminal-source", "executionpeer.terminal_attribution", func(_ *core.Program, document *execution.Execution) {
			document.Events[rootTerminalIndex].SourcePlace = "forged-source"
		}},
		{"wrong-terminal-constructor", "executionpeer.failure_provenance", func(_ *core.Program, document *execution.Execution) {
			document.Events[rootTerminalIndex].TypeID = "ResourceError"
		}},
		{"wrong-terminal-operation", "executionpeer.terminal_attribution", func(_ *core.Program, document *execution.Execution) {
			document.Events[childTerminalIndex].ID = phase371OpFailID(program, "probe", 5) + ":event:failed"
		}},
		{"wrong-root-kind", "executionpeer.terminal_outcome", func(_ *core.Program, document *execution.Execution) {
			document.Outcome.Kind = execution.OutcomeReturned
		}},
		{"missing-terminal", "executionpeer.preorder", func(_ *core.Program, document *execution.Execution) {
			document.Events = document.Events[:rootTerminalIndex]
		}},
		{"duplicate-terminal", "executionpeer.closed_activation", func(_ *core.Program, document *execution.Execution) {
			duplicate := document.Events[rootTerminalIndex]
			duplicate.Occurrence++
			document.Events = append(document.Events, duplicate)
		}},
	} {
		t.Run(control.name, func(t *testing.T) {
			if err := executionpeer.Validate(program, clonePhase371Execution(useFailure)); err != nil {
				t.Fatalf("pre-mutation failure positive rejected: %v", err)
			}
			candidate := clonePhase371CoreProgram(program)
			mutant := clonePhase371Execution(useFailure)
			control.mutate(&candidate, &mutant)
			err := executionpeer.Validate(candidate, mutant)
			if peerErr, ok := err.(*executionpeer.Error); !ok || peerErr.Code() != control.wantCode {
				t.Fatalf("forged control returned %v, want %s", err, control.wantCode)
			}
			if err := executionpeer.Validate(program, clonePhase371Execution(useFailure)); err != nil {
				t.Fatalf("restored failure positive rejected after %s: %v", control.name, err)
			}
		})
	}

	t.Run("stale-cross-activation-failure", func(t *testing.T) {
		entryID := phase371FunctionID(program, "main")
		calls := phase371MainCalls(program, "acquire")
		if len(calls) < 2 {
			t.Fatalf("Phase24 source has %d main acquire call sites, want at least two", len(calls))
		}
		stale := clonePhase371Execution(acquireFailure)
		stale.Events[foreignIndexForPhase371(acquireFailure)].Invocation = mustPhase371Invocation(t, entryID, calls[1].ID)
		if err := executionpeer.Validate(program, stale); err == nil {
			t.Fatal("peer accepted a failure event reassigned to a sibling activation")
		}
		if err := executionpeer.Validate(program, clonePhase371Execution(acquireFailure)); err != nil {
			t.Fatalf("restored AcquireError positive rejected after stale-activation control: %v", err)
		}
	})

	t.Run("clean-return-relabeled-as-failure", func(t *testing.T) {
		mutant := clonePhase371Execution(clean)
		mutant.Outcome.Kind = execution.OutcomeTypedFailure
		if err := executionpeer.Validate(program, mutant); err == nil {
			t.Fatal("peer accepted a clean ordinary-return trace relabeled as typed failure")
		}
	})

	t.Run("failure-payload-is-outside-peer-knowledge", func(t *testing.T) {
		mutant := clonePhase371Execution(useFailure)
		mutant.Outcome.Value = "changed-without-payload-proof"
		if err := executionpeer.Validate(program, mutant); err != nil {
			t.Fatalf("peer inferred an absent failure payload: %v", err)
		}
	})
}

func phase371FindEventIndex(document execution.Execution, match func(execution.Event) bool) int {
	for index, event := range document.Events {
		if match(event) {
			return index
		}
	}
	return -1
}

func phase371HasEventAfter(document execution.Execution, start int, match func(execution.Event) bool) bool {
	for index := start + 1; index < len(document.Events); index++ {
		if match(document.Events[index]) {
			return true
		}
	}
	return false
}

func phase371FunctionID(program core.Program, name string) string {
	for _, function := range program.Functions {
		if function.Name == name {
			return function.ID
		}
	}
	return ""
}

func phase371MainCalls(program core.Program, callee string) []core.LinearOperation {
	mainID := phase371FunctionID(program, "main")
	var calls []core.LinearOperation
	for _, function := range program.Functions {
		if function.ID != mainID || function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCall && strings.HasSuffix(operation.CalleeID, ":fn:"+callee) {
				calls = append(calls, operation)
			}
		}
	}
	return calls
}

func phase371MainCallID(program core.Program, callee string) string {
	calls := phase371MainCalls(program, callee)
	if len(calls) == 0 {
		return ""
	}
	return calls[0].ID
}

func phase371OpFailID(program core.Program, functionName string, ordinal int) string {
	for _, function := range program.Functions {
		if function.Name != functionName || function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpFail && strings.HasSuffix(operation.ID, fmt.Sprintf(":op:%d", ordinal)) {
				return operation.ID
			}
		}
	}
	return ""
}

func foreignIndexForPhase371(document execution.Execution) int {
	return phase371FindEventIndex(document, func(event execution.Event) bool { return event.Kind == "foreign.failed" })
}

func mustPhase371Invocation(t *testing.T, entryID, callID string) string {
	t.Helper()
	invocation, err := execution.FormatInvocation(entryID, []execution.InvocationSegment{{OpCallID: callID, Ordinal: 0}})
	if err != nil {
		t.Fatalf("format Phase24 sibling activation: %v", err)
	}
	return invocation
}

func clonePhase371CoreProgram(program core.Program) core.Program {
	clone := program
	clone.DataTypes = append([]core.DataType(nil), program.DataTypes...)
	clone.Functions = append([]core.Function(nil), program.Functions...)
	for functionIndex := range clone.Functions {
		function := &clone.Functions[functionIndex]
		if function.Linear == nil {
			continue
		}
		linear := *function.Linear
		linear.Types = append([]core.TypeFact(nil), function.Linear.Types...)
		linear.Places = append([]core.Place(nil), function.Linear.Places...)
		linear.Operations = append([]core.LinearOperation(nil), function.Linear.Operations...)
		linear.Blocks = append([]core.Block(nil), function.Linear.Blocks...)
		linear.Edges = append([]core.Edge(nil), function.Linear.Edges...)
		for operationIndex := range linear.Operations {
			if linear.Operations[operationIndex].Foreign != nil {
				contract := *linear.Operations[operationIndex].Foreign
				linear.Operations[operationIndex].Foreign = &contract
			}
		}
		for blockIndex := range linear.Blocks {
			linear.Blocks[blockIndex].OperationIDs = append([]string(nil), linear.Blocks[blockIndex].OperationIDs...)
			linear.Blocks[blockIndex].Successors = append([]string(nil), linear.Blocks[blockIndex].Successors...)
		}
		function.Linear = &linear
	}
	return clone
}

func phase371MutateOperation(program *core.Program, operationID string, mutate func(*core.LinearOperation)) bool {
	for functionIndex := range program.Functions {
		function := &program.Functions[functionIndex]
		if function.Linear == nil {
			continue
		}
		for operationIndex := range function.Linear.Operations {
			operation := &function.Linear.Operations[operationIndex]
			if operation.ID == operationID {
				mutate(operation)
				return true
			}
		}
	}
	return false
}

func TestPhase371DefectTerminalControls(t *testing.T) {
	t.Run("legacy-static", func(t *testing.T) {
		program, entry, leaf, _, _ := phase371LegacyDefectProgram(t)
		cleanRun, err := interp.RunWithFileBytes(program, "main", "phase371-clean", func(string) ([]byte, string) {
			return []byte("xx"), ""
		})
		if err != nil {
			t.Fatalf("interpreter clean static call: %v", err)
		}
		clean, err := session.ProjectExecutionSchema2(program, cleanRun)
		if err != nil {
			t.Fatalf("project clean static call: %v", err)
		}
		if clean.Outcome.Kind != execution.OutcomeReturned {
			t.Fatalf("clean static outcome = %+v, want returned", clean.Outcome)
		}
		if err := executionpeer.Validate(program, clean); err != nil {
			t.Fatalf("peer rejected clean static trace: %v", err)
		}
		if err := executionpeer.Validate(program, clonePhase371Execution(clean)); err != nil {
			t.Fatalf("peer rejected no-op clean static trace: %v", err)
		}

		defectRun, err := interp.RunWithFileBytes(program, "main", "phase371-child-defect", func(string) ([]byte, string) {
			return []byte{}, ""
		})
		if err != nil {
			t.Fatalf("interpreter static child defect: %v", err)
		}
		defect, err := session.ProjectExecutionSchema2(program, defectRun)
		if err != nil {
			t.Fatalf("project static child defect: %v", err)
		}
		if defect.Outcome.Kind != execution.OutcomeDefect || defect.Outcome.Value != "" || len(defect.Events) == 0 {
			t.Fatalf("static child defect document = %+v, want an empty-valued defect with events", defect)
		}
		if err := executionpeer.Validate(program, defect); err != nil {
			t.Fatalf("peer rejected actual legacy child defect trace: %v", err)
		}
		terminalIndex := len(defect.Events) - 1
		terminal := defect.Events[terminalIndex]
		if terminal.Kind != "function.defected" || terminal.FunctionID != leaf.ID || terminal.Output != "legacy child defect" {
			t.Fatalf("static child terminal = %+v, want the checked child defect", terminal)
		}
		rootInvocation, err := execution.FormatInvocation(entry.ID, nil)
		if err != nil {
			t.Fatalf("format static root invocation: %v", err)
		}
		controls := phase371TerminalMutations(terminalIndex, rootInvocation)
		for _, control := range controls {
			t.Run(control.name, func(t *testing.T) {
				if err := executionpeer.Validate(program, clonePhase371Execution(clean)); err != nil {
					t.Fatalf("pre-mutation clean positive rejected: %v", err)
				}
				mutant := clonePhase371Execution(defect)
				control.mutate(&mutant)
				if err := executionpeer.Validate(program, mutant); err == nil {
					t.Fatalf("peer accepted forged static terminal control %q", control.name)
				}
				if err := executionpeer.Validate(program, clonePhase371Execution(clean)); err != nil {
					t.Fatalf("restored clean positive rejected after %s: %v", control.name, err)
				}
			})
		}
		var callerActivity execution.Event
		for _, event := range clean.Events {
			if event.Kind == "resource.released" && event.FunctionID == entry.ID {
				callerActivity = event
			}
		}
		if callerActivity.ID == "" {
			t.Fatal("genuine clean trace has no caller cleanup activity after the child call")
		}
		postChild := clonePhase371Execution(defect)
		postChild.Events = append(postChild.Events, callerActivity)
		postChildError := executionpeer.Validate(program, postChild)
		peerError, ok := postChildError.(*executionpeer.Error)
		if !ok || peerError.Code() != "executionpeer.closed_activation" {
			t.Fatalf("legacy-static caller cleanup after child defect = %v, want executionpeer.closed_activation", postChildError)
		}
		if err := executionpeer.Validate(program, clonePhase371Execution(clean)); err != nil {
			t.Fatalf("restored clean static positive rejected after caller-after-defect control: %v", err)
		}
	})

	t.Run("dynamic-scalar", func(t *testing.T) {
		program, entry, callee, _ := phase371DynamicDefectProgram(t)
		defect := phase371DynamicDefectExecution(t, program, "18446744073709551613")
		if defect.Outcome.Kind != execution.OutcomeDefect || defect.Outcome.Value != "" {
			t.Fatalf("dynamic child defect = %+v, want empty-valued defect", defect.Outcome)
		}
		if defect.Events[len(defect.Events)-1].FunctionID != callee.ID {
			t.Fatalf("dynamic child terminal = %+v, want callee %q", defect.Events[len(defect.Events)-1], callee.ID)
		}
		if err := executionpeer.Validate(program, defect); err != nil {
			t.Fatalf("peer rejected clean dynamic child defect trace: %v", err)
		}
		clean := phase371DynamicDefectExecution(t, program, "2")
		if clean.Outcome.Kind != execution.OutcomeReturned {
			t.Fatalf("dynamic clean outcome = %+v, want returned", clean.Outcome)
		}
		if err := executionpeer.Validate(program, clean); err != nil {
			t.Fatalf("peer rejected clean dynamic trace: %v", err)
		}
		if err := executionpeer.Validate(program, clonePhase371Execution(clean)); err != nil {
			t.Fatalf("peer rejected no-op clean dynamic trace: %v", err)
		}
		terminalIndex := len(defect.Events) - 1
		rootInvocation, err := execution.FormatInvocation(entry.ID, nil)
		if err != nil {
			t.Fatalf("format dynamic root invocation: %v", err)
		}
		for _, control := range phase371TerminalMutations(terminalIndex, rootInvocation) {
			t.Run(control.name, func(t *testing.T) {
				if err := executionpeer.Validate(program, clonePhase371Execution(clean)); err != nil {
					t.Fatalf("pre-mutation dynamic positive rejected: %v", err)
				}
				mutant := clonePhase371Execution(defect)
				control.mutate(&mutant)
				if err := executionpeer.Validate(program, mutant); err == nil {
					t.Fatalf("peer accepted forged dynamic terminal control %q", control.name)
				}
				if err := executionpeer.Validate(program, clonePhase371Execution(clean)); err != nil {
					t.Fatalf("restored dynamic positive rejected after %s: %v", control.name, err)
				}
			})
		}
		callerDefect := phase371DynamicDefectExecution(t, program, "1")
		if callerDefect.Outcome.Kind != execution.OutcomeDefect {
			t.Fatalf("dynamic caller control = %+v, want root defect from the same program", callerDefect.Outcome)
		}
		var callerTerminal execution.Event
		for _, event := range callerDefect.Events {
			if event.Kind == "function.defected" && event.FunctionID == entry.ID {
				callerTerminal = event
			}
		}
		if callerTerminal.ID == "" {
			t.Fatal("dynamic caller defect trace has no root terminal event")
		}
		postChild := clonePhase371Execution(defect)
		postChild.Events = append(postChild.Events, callerTerminal)
		if err := executionpeer.Validate(program, postChild); err == nil {
			t.Fatal("peer accepted dynamic caller activity after child defect")
		}
		if err := executionpeer.Validate(program, clonePhase371Execution(clean)); err != nil {
			t.Fatalf("restored dynamic clean positive rejected after caller-after-defect control: %v", err)
		}
	})
}

type phase371TerminalControl struct {
	name   string
	mutate func(*execution.Execution)
}

func phase371TerminalMutations(terminalIndex int, rootInvocation string) []phase371TerminalControl {
	return []phase371TerminalControl{
		{"source", func(document *execution.Execution) { document.Events[terminalIndex].SourcePlace = "forged-source" }},
		{"type", func(document *execution.Execution) { document.Events[terminalIndex].TypeID = "forged-type" }},
		{"function", func(document *execution.Execution) { document.Events[terminalIndex].FunctionID = "forged-function" }},
		{"invocation", func(document *execution.Execution) { document.Events[terminalIndex].Invocation = rootInvocation }},
		{"reason", func(document *execution.Execution) { document.Events[terminalIndex].Output = "forged reason" }},
		{"outcome-kind", func(document *execution.Execution) { document.Outcome.Kind = execution.OutcomeReturned }},
		{"outcome-value", func(document *execution.Execution) { document.Outcome.Value = "forged value" }},
		{"missing-terminal", func(document *execution.Execution) { document.Events = document.Events[:terminalIndex] }},
		{"duplicate-terminal", func(document *execution.Execution) {
			repeated := document.Events[terminalIndex]
			repeated.Occurrence++
			document.Events = append(document.Events, repeated)
		}},
	}
}

func phase371DynamicDefectProgram(t *testing.T) (core.Program, core.Function, core.Function, core.LinearOperation) {
	t.Helper()
	checked := session.Check([]byte(phase371DynamicDefectSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("dynamic defect source diagnostics: %+v", checked.Diagnostics)
	}
	var entry, callee core.Function
	for _, function := range checked.Program.Functions {
		switch function.Name {
		case "main":
			entry = function
		case "increment":
			callee = function
		}
	}
	var call core.LinearOperation
	for _, operation := range entry.Linear.Operations {
		if operation.Kind == core.OpCall && operation.CalleeID == callee.ID {
			call = operation
			break
		}
	}
	if entry.ID == "" || callee.ID == "" || call.ID == "" {
		t.Fatalf("dynamic checked call missing: entry=%q callee=%q call=%q", entry.ID, callee.ID, call.ID)
	}
	return checked.Program, entry, callee, call
}

func phase371DynamicDefectExecution(t *testing.T, program core.Program, input string) execution.Execution {
	t.Helper()
	run, err := interp.RunApplication(program, "main", input)
	if err != nil {
		t.Fatalf("dynamic RunApplication(%q): %v", input, err)
	}
	return run.Execution
}

func clonePhase371Execution(document execution.Execution) execution.Execution {
	clone := document
	clone.Events = append([]execution.Event(nil), document.Events...)
	clone.LiveResources = append([]string(nil), document.LiveResources...)
	return clone
}

func phase371LegacyDefectProgram(t *testing.T) (core.Program, core.Function, core.Function, core.LinearOperation, [32]byte) {
	t.Helper()
	root := t.TempDir()
	entryPath := testsupport.ProjectPath("examples", "checksum", "checksum.schway")
	entrySource, err := os.ReadFile(entryPath)
	if err != nil {
		t.Fatalf("read actual checksum entry source: %v", err)
	}
	const entryTail = "  let sum = adler32(view)\n  sum\n}\n"
	if strings.Count(string(entrySource), entryTail) != 1 {
		t.Fatal("actual checksum entry no longer has the expected single call-and-return tail")
	}
	moduleSource, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", "adler32.schway"))
	if err != nil {
		t.Fatalf("read actual Adler module source: %v", err)
	}
	const functionStart = "fn adler32(bytes: ByteView) -> U64 {"
	if strings.Count(string(moduleSource), functionStart) != 1 || !strings.HasSuffix(string(moduleSource), "}\n") {
		t.Fatal("actual adler32.schway no longer has the expected single final function")
	}
	start := strings.Index(string(moduleSource), functionStart)
	moduleSource = append(append([]byte(nil), moduleSource[:start]...), []byte(phase371AdlerDefectFunction)...)
	for _, row := range []struct {
		name   string
		source []byte
	}{{"checksum.schway", entrySource}, {"adler32.schway", moduleSource}} {
		if err := os.WriteFile(filepath.Join(root, row.name), row.source, 0o600); err != nil {
			t.Fatalf("write temporary checked-source witness %s: %v", row.name, err)
		}
	}
	checked, err := session.CheckFile(filepath.Join(root, "checksum.schway"))
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("check temporary checksum source: err=%v diagnostics=%+v", err, checked.Diagnostics)
	}
	var entry, leaf core.Function
	for _, function := range checked.Program.Functions {
		switch function.Name {
		case "main":
			entry = function
		case "adler32":
			leaf = function
		}
	}
	if entry.ID == "" || entry.Linear == nil || leaf.ID == "" || leaf.Linear == nil {
		t.Fatalf("checked entry/helper missing: main=%+v adler32=%+v", entry, leaf)
	}
	var call core.LinearOperation
	for _, operation := range entry.Linear.Operations {
		if operation.Kind == core.OpCall && operation.CalleeID == leaf.ID {
			if call.ID != "" {
				t.Fatalf("entry has duplicate calls to the child: %q and %q", call.ID, operation.ID)
			}
			call = operation
		}
	}
	if call.ID == "" {
		t.Fatal("checked entry has no direct call to the non-scalar child")
	}
	boundSource := append(append([]byte(nil), entrySource...), moduleSource...)
	return checked.Program, entry, leaf, call, sha256.Sum256(boundSource)
}
