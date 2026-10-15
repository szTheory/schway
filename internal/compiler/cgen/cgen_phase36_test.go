package cgen_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
)

const phase36NativeDynamicInvocationSource = `module phase36.native_dynamic_invocation

export {
  fn main
}

fn increment(value: U64) -> U64 {
  let one = 1
  let result = value + one
  result
}

fn nested(value: U64) -> U64 {
  var index = 0
  var result = value
  while index < 2 {
    result = increment(result)
    index = index + 1
  }
  result
}

fn below_ten(value: U64) -> Bool {
  let result = value < 10
  result
}

fn choose(value: Bool) -> U64 {
  var result = 0
  if value { result = 7 } else { result = 9 }
  result
}

fn identity_bool(value: Bool) -> Bool {
  let result = value
  result
}

fn main(input: U64) -> U64 {
  var index = 0
  var result = input
  var first = 0
  var second = 0
  var condition = input < 10
  var selected = 0
  while index < 2 {
    first = nested(result)
    second = nested(first)
    condition = below_ten(second)
    selected = choose(condition)
    condition = identity_bool(condition)
    if condition { result = selected } else { result = selected + 1 }
    index = index + 1
  }
  result
}
`

func TestPhase36NativeDynamicInvocationPaths(t *testing.T) {
	checked := session.Check([]byte(phase36NativeDynamicInvocationSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	for _, want := range []string{
		"schway_invocation_arena",
		"schway_reserve_invocation",
		"SCHWAY_INVOCATION_ORDINAL_MAX",
	} {
		if !containsPhase36(generated, want) {
			t.Fatalf("generated C is missing dynamic invocation support %q", want)
		}
	}

	got, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{"3"})
	if err != nil || len(got.Pairs) != 1 {
		t.Fatalf("native O0 run: pairs=%d err=%v", len(got.Pairs), err)
	}
	nativeExecution := got.Pairs[0].Execution
	if nativeExecution.Outcome.Kind != "returned" || nativeExecution.Outcome.Value != "10" {
		t.Fatalf("native literal answer = %+v, want returned 10", nativeExecution.Outcome)
	}
	for _, lane := range []struct {
		name         string
		optimization string
		lto          bool
	}{
		{name: "O3", optimization: "-O3"},
		{name: "O3-LTO", optimization: "-O3", lto: true},
	} {
		t.Run(lane.name, func(t *testing.T) {
			runner := native.DefaultRunner()
			runner.LTO = lane.lto
			result, runErr := runner.Run(context.Background(), generated, lane.optimization, []string{"3"})
			if runErr != nil || len(result.Pairs) != 1 {
				t.Fatalf("native %s run: pairs=%d err=%v", lane.name, len(result.Pairs), runErr)
			}
			actual := result.Pairs[0].Execution.Outcome
			if actual.Kind != "returned" || actual.Value != "10" {
				t.Fatalf("native %s literal answer = %+v, want returned 10", lane.name, actual)
			}
		})
	}

	entryID := phase36FunctionID(t, checked.Program.Functions, "main")
	nestedCalls := phase36CallIDs(t, checked.Program.Functions, "main", "nested")
	innerCalls := phase36CallIDs(t, checked.Program.Functions, "nested", "increment")
	if len(nestedCalls) != 2 || len(innerCalls) != 1 {
		t.Fatalf("static call sites: main->nested=%v nested->increment=%v, want two sites and one repeated site", nestedCalls, innerCalls)
	}
	root := phase36Invocation(t, entryID)
	wantPaths := []string{root}
	for parentOrdinal := 0; parentOrdinal < 2; parentOrdinal++ {
		for _, callID := range nestedCalls {
			nestedPath := phase36Invocation(t, entryID,
				execution.InvocationSegment{OpCallID: callID, Ordinal: parentOrdinal})
			wantPaths = append(wantPaths, nestedPath)
			for childOrdinal := 0; childOrdinal < 2; childOrdinal++ {
				wantPaths = append(wantPaths, phase36Invocation(t, entryID,
					execution.InvocationSegment{OpCallID: callID, Ordinal: parentOrdinal},
					execution.InvocationSegment{OpCallID: innerCalls[0], Ordinal: childOrdinal}))
			}
		}
	}
	seen := make(map[string]bool, len(wantPaths))
	for _, event := range nativeExecution.Events {
		if event.Invocation == "" {
			t.Fatalf("event %q has an empty invocation", event.ID)
		}
		seen[event.Invocation] = true
	}
	for _, want := range wantPaths {
		if !seen[want] {
			t.Errorf("final serialized execution is missing retained dynamic path %q", want)
		}
	}
	for path := range seen {
		if _, err := execution.ParseInvocation(path); err != nil {
			t.Errorf("generated invocation %q is not canonical: %v", path, err)
		}
	}
}

func containsPhase36(value, fragment string) bool {
	return strings.Contains(value, fragment)
}

func phase36FunctionID(t *testing.T, functions []core.Function, name string) string {
	t.Helper()
	for _, function := range functions {
		if function.Name == name {
			return function.ID
		}
	}
	t.Fatalf("function %q is absent", name)
	return ""
}

func phase36CallIDs(t *testing.T, functions []core.Function, caller, callee string) []string {
	t.Helper()
	var result []string
	for _, function := range functions {
		if function.Name != caller || function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCall && operation.CalleeID == phase36FunctionID(t, functions, callee) {
				result = append(result, operation.ID)
			}
		}
	}
	return result
}

func phase36Invocation(t *testing.T, entryID string, segments ...execution.InvocationSegment) string {
	t.Helper()
	value, err := execution.FormatInvocation(entryID, segments)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestPhase36CallEventCapacityBoundary(t *testing.T) {
	checked := session.Check([]byte(phase36NativeDynamicInvocationSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{"3"})
	if err != nil || len(baseline.Pairs) != 1 {
		t.Fatalf("baseline run pairs=%d err=%v", len(baseline.Pairs), err)
	}
	eventCount := len(baseline.Pairs[0].Execution.Events)
	if eventCount < 2 {
		t.Fatalf("baseline event count=%d; expected a nontrivial trace", eventCount)
	}

	t.Run("exact", func(t *testing.T) {
		limited := phase36ReplaceDefine(t, generated, "SCHWAY_EVENT_CAPACITY", fmt.Sprintf("%du", eventCount))
		result, runErr := native.DefaultRunner().Run(context.Background(), limited, "-O0", []string{"3"})
		if runErr != nil || len(result.Pairs) != 1 || len(result.Pairs[0].Execution.Events) != eventCount {
			t.Fatalf("exact event capacity result=%+v err=%v; want %d events", result, runErr, eventCount)
		}
	})

	t.Run("plus-one", func(t *testing.T) {
		limited := phase36ReplaceDefine(t, generated, "SCHWAY_EVENT_CAPACITY", fmt.Sprintf("%du", eventCount-1))
		_, runErr := native.DefaultRunner().Run(context.Background(), limited, "-O0", []string{"3"})
		var toolError *native.ToolError
		if !strings.Contains(fmt.Sprint(runErr), native.CodeEventCapacityExhausted) || !errors.As(runErr, &toolError) || toolError.Code != native.CodeEventCapacityExhausted {
			t.Fatalf("one-below event capacity err=%v; want %s", runErr, native.CodeEventCapacityExhausted)
		}
	})

	t.Run("encoded-output-exact-and-plus-one", func(t *testing.T) {
		source := []byte(phase36NativeDynamicInvocationSource)
		build := func(t *testing.T, limit int) string {
			t.Helper()
			var application string
			var emitErr error
			if limit > 0 {
				restore := cgen.SetApplicationEvidenceOutputLimitForTest(limit)
				defer restore()
			}
			application, emitErr = cgen.EmitApplication(checked.Program)
			if emitErr != nil {
				t.Fatalf("EmitApplication evidence limit %d: %v", limit, emitErr)
			}
			artifact := filepath.Join(t.TempDir(), "phase36-evidence-capacity")
			if _, buildErr := native.BuildApplication(context.Background(), source, application, artifact); buildErr != nil {
				t.Fatalf("BuildApplication evidence limit %d: %v", limit, buildErr)
			}
			return artifact
		}
		capture := func(t *testing.T, artifact string) (native.RunOutcome, native.ApplicationEvidenceReport, error, string) {
			t.Helper()
			var stdout, stderr bytes.Buffer
			outcome, report, runErr := native.RunApplicationWithEvidence(context.Background(), artifact, "3", filepath.Join(t.TempDir(), "report.json"), native.EvidenceEvents, &stdout, &stderr)
			if stderr.Len() != 0 {
				t.Fatalf("evidence limit changed app stderr: %q", stderr.String())
			}
			return outcome, report, runErr, stdout.String()
		}

		baselineOutcome, baselineReport, baselineErr, baselineOutput := capture(t, build(t, 0))
		if baselineErr != nil || baselineOutcome.Kind != native.RunExited || baselineOutcome.ExitCode != 0 || baselineReport.CaptureStatus != native.EvidenceStatusComplete || baselineReport.Execution == nil || baselineOutput != "10\n" {
			t.Fatalf("baseline app capture outcome=%+v report=%+v output=%q err=%v", baselineOutcome, baselineReport, baselineOutput, baselineErr)
		}
		executionJSON, err := json.Marshal(baselineReport.Execution)
		if err != nil {
			t.Fatal(err)
		}
		captureBytes := len(`{"schema":"lang.app-capture/1","status":"complete","execution":`) + len(executionJSON) + len("}\n")
		if captureBytes <= 2 {
			t.Fatalf("calculated dynamic capture document size=%d", captureBytes)
		}

		exactOutcome, exactReport, exactErr, exactOutput := capture(t, build(t, captureBytes))
		if exactErr != nil || exactOutcome.Kind != native.RunExited || exactOutcome.ExitCode != 0 || exactReport.CaptureStatus != native.EvidenceStatusComplete || exactOutput != "10\n" {
			t.Fatalf("exact encoded-output capacity outcome=%+v report=%+v output=%q err=%v; want complete at %d bytes", exactOutcome, exactReport, exactOutput, exactErr, captureBytes)
		}

		belowOutcome, belowReport, belowErr, belowOutput := capture(t, build(t, captureBytes-1))
		var toolError *native.ToolError
		if !errors.As(belowErr, &toolError) || toolError.Code != "native.evidence_capacity_exhausted" || belowOutcome.Kind != native.RunExited || belowOutcome.ExitCode != 0 || belowReport.ProcessOutcome != belowOutcome || belowReport.CaptureStatus != native.EvidenceStatusCapacityExhausted || belowReport.Execution != nil || belowOutput != "10\n" {
			t.Fatalf("one-below encoded-output capacity outcome=%+v report=%+v output=%q err=%v; want independent capacity verdict", belowOutcome, belowReport, belowOutput, belowErr)
		}
	})
}

func TestPhase36NativeOrdinalAndPathBoundaries(t *testing.T) {
	checked := session.Check([]byte(phase36NativeDynamicInvocationSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{"3"})
	if err != nil || len(baseline.Pairs) != 1 {
		t.Fatalf("baseline run pairs=%d err=%v", len(baseline.Pairs), err)
	}
	uniquePaths := make(map[string]struct{})
	for _, event := range baseline.Pairs[0].Execution.Events {
		if event.Invocation != "" {
			uniquePaths[event.Invocation] = struct{}{}
		}
	}
	pathBytes := 0
	for path := range uniquePaths {
		if strings.HasPrefix(path, "inv:entry:") && !strings.Contains(path, "/") {
			continue
		}
		pathBytes += len(path) + 1
	}
	if pathBytes < 2 {
		t.Fatalf("baseline retained child path bytes=%d; want dynamic children", pathBytes)
	}
	t.Run("path-exact", func(t *testing.T) {
		limited := phase36ReplaceDefine(t, generated, "SCHWAY_INVOCATION_CAPACITY", fmt.Sprintf("%du", pathBytes))
		result, runErr := native.DefaultRunner().Run(context.Background(), limited, "-O0", []string{"3"})
		if runErr != nil || len(result.Pairs) != 1 {
			t.Fatalf("exact path capacity result=%+v err=%v; want a complete run at %d bytes", result, runErr, pathBytes)
		}
	})
	t.Run("path-plus-one", func(t *testing.T) {
		limited := phase36ReplaceDefine(t, generated, "SCHWAY_INVOCATION_CAPACITY", fmt.Sprintf("%du", pathBytes-1))
		_, runErr := native.DefaultRunner().Run(context.Background(), limited, "-O0", []string{"3"})
		var toolError *native.ToolError
		if !errors.As(runErr, &toolError) || toolError.Code != native.CodeEventCapacityExhausted {
			t.Fatalf("one-below path capacity err=%v; want %s", runErr, native.CodeEventCapacityExhausted)
		}
	})

	ordinalSource := `module phase36.ordinal_boundary

export {
  fn main
}

fn identity(value: U64) -> U64 {
  value
}

fn main(input: U64) -> U64 {
  var index = 0
  var result = input
  while index < input {
    result = identity(result)
    index = index + 1
  }
  result
}
`
	ordinalChecked := session.Check([]byte(ordinalSource))
	if len(ordinalChecked.Diagnostics) != 0 {
		t.Fatalf("ordinal source diagnostics: %+v", ordinalChecked.Diagnostics)
	}
	ordinalC, err := cgen.EmitNative(ordinalChecked.Program)
	if err != nil {
		t.Fatal(err)
	}
	counterPattern := regexp.MustCompile(`uint64_t ([A-Za-z0-9_]*call_ordinal[A-Za-z0-9_]*) = 0u;`)
	counter := counterPattern.FindStringSubmatch(ordinalC)
	if len(counter) != 2 {
		t.Fatal("generated ordinal counter declaration is missing")
	}
	marker := counter[0]
	counterName := counter[1]
	maxDeclaration := "uint64_t " + counterName + " = SCHWAY_INVOCATION_ORDINAL_MAX;"
	maxC := strings.Replace(ordinalC, marker, maxDeclaration, 1)
	maxResult, maxErr := native.DefaultRunner().Run(context.Background(), maxC, "-O0", []string{"1"})
	if maxErr != nil || len(maxResult.Pairs) != 1 || maxResult.Pairs[0].Execution.Outcome.Value != "1" {
		t.Fatalf("maximum ordinal run result=%+v err=%v; the canonical maximum must fit", maxResult, maxErr)
	}
	maximumPath := false
	for _, event := range maxResult.Pairs[0].Execution.Events {
		if strings.HasSuffix(event.Invocation, "#2147483647") {
			maximumPath = true
		}
	}
	if !maximumPath {
		t.Fatal("maximum ordinal path was not retained in the successful trace")
	}
	_, nextErr := native.DefaultRunner().Run(context.Background(), maxC, "-O0", []string{"2"})
	var nextToolError *native.ToolError
	if !errors.As(nextErr, &nextToolError) || nextToolError.Code != native.CodeEventCapacityExhausted {
		t.Fatalf("ordinal maximum plus one err=%v; want %s before wrap or narrowing", nextErr, native.CodeEventCapacityExhausted)
	}

	overDeclaration := "uint64_t " + counterName + " = (SCHWAY_INVOCATION_ORDINAL_MAX + UINT64_C(1));"
	overC := strings.Replace(ordinalC, marker, overDeclaration, 1)
	_, overErr := native.DefaultRunner().Run(context.Background(), overC, "-O0", []string{"1"})
	var overToolError *native.ToolError
	if !errors.As(overErr, &overToolError) || overToolError.Code != native.CodeEventCapacityExhausted {
		t.Fatalf("out-of-range initial ordinal err=%v; want checked capacity exit", overErr)
	}
}

func phase36ReplaceDefine(t *testing.T, generated, name, value string) string {
	t.Helper()
	prefix := "#define " + name + " "
	start := strings.Index(generated, prefix)
	if start < 0 {
		t.Fatalf("generated C lacks %s", name)
	}
	end := strings.IndexByte(generated[start:], '\n')
	if end < 0 {
		t.Fatalf("generated %s definition has no line ending", name)
	}
	end += start
	return generated[:start] + prefix + value + generated[end:]
}

const phase36EmitterCallControlSource = `module phase36.emitter_call_control

export {
  fn main
}

fn increment(value: U64) -> U64 {
  let one = 1
  let result = value + one
  result
}

fn main(input: U64) -> U64 {
  let result = increment(input)
  let answer = result + input
  answer
}
`

func TestPhase36EmitterCallInventoryControls(t *testing.T) {
	checked := session.Check([]byte(phase36EmitterCallControlSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	var call core.LinearOperation
	callCount := 0
	for _, function := range checked.Program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCall {
				call = operation
				callCount++
			}
		}
	}
	if callCount != 1 {
		t.Fatalf("checked core call inventory=%d, want exactly one call site", callCount)
	}
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	callMarker := "/* call: " + call.ID + " */"
	if count := strings.Count(generated, "/* call:"); count != callCount || strings.Count(generated, callMarker) != 1 {
		t.Fatalf("generated dynamic call inventory=%d marker_count=%d, want one checked site %q", count, strings.Count(generated, callMarker), call.ID)
	}

	markerOffset := strings.Index(generated, callMarker)
	lineStart := strings.LastIndex(generated[:markerOffset], "\n") + 1
	lineEnd := strings.IndexByte(generated[markerOffset:], '\n') + markerOffset
	callLine := generated[lineStart:lineEnd]
	callPattern := regexp.MustCompile(`^(\s*)(schway_scalar_[0-9]+) = ([A-Za-z0-9_]+)\((schway_scalar_[0-9]+), ([A-Za-z0-9_]+\[invocation_index\])\); /\* call: (.+) \*/$`)
	parts := callPattern.FindStringSubmatch(callLine)
	if len(parts) != 7 || parts[6] != call.ID {
		t.Fatalf("generated call line lacks a typed argument/result and child invocation index: %q", callLine)
	}
	functionStart := strings.LastIndex(generated[:lineStart], "\nstatic ") + 1
	if functionStart == 0 {
		t.Fatal("dynamic call site has no enclosing emitted user function")
	}
	functionEndOffset := strings.Index(generated[lineEnd:], "\n}\n")
	if functionEndOffset < 0 {
		t.Fatal("dynamic call site has no enclosing function end")
	}
	functionRegion := generated[functionStart : lineEnd+functionEndOffset+3]
	if strings.Count(functionRegion, "schway_record_event(\"function.called\"") != 1 || strings.Count(functionRegion, "schway_reserve_invocation(") != 1 {
		t.Fatalf("user function call region does not enumerate one call event and one path reservation:\n%s", functionRegion)
	}
	calledOffset := strings.Index(functionRegion, "schway_record_event(\"function.called\"")
	reserveOffset := strings.Index(functionRegion, "schway_reserve_invocation(")
	callOffset := strings.Index(functionRegion, callMarker)
	if calledOffset < 0 || reserveOffset <= calledOffset || callOffset <= reserveOffset {
		t.Fatalf("call event, path reservation and value call are out of order: called=%d reserve=%d call=%d", calledOffset, reserveOffset, callOffset)
	}
	for _, guard := range []string{
		"if (schway_event_count >= SCHWAY_EVENT_CAPACITY)",
		"if (*ordinal > SCHWAY_INVOCATION_ORDINAL_MAX)",
		"required > SCHWAY_INVOCATION_CAPACITY - schway_invocation_used",
	} {
		if !strings.Contains(generated, guard) {
			t.Fatalf("generated native call support is missing live capacity guard %q", guard)
		}
	}

	runValue := func(t *testing.T, source string) string {
		t.Helper()
		result, runErr := native.DefaultRunner().Run(context.Background(), source, "-O0", []string{"5"})
		if runErr != nil || len(result.Pairs) != 1 || result.Pairs[0].Execution.Outcome.Kind != "returned" {
			t.Fatalf("native call run pairs=%d err=%v", len(result.Pairs), runErr)
		}
		return result.Pairs[0].Execution.Outcome.Value
	}
	if got := runValue(t, generated); got != "11" {
		t.Fatalf("clean call result=%q, want 11", got)
	}

	droppedArgument := fmt.Sprintf("%s%s = %s(UINT64_C(0), %s); /* call: %s */", parts[1], parts[2], parts[3], parts[5], parts[6])
	argumentMutant := generated[:lineStart] + droppedArgument + generated[lineEnd:]
	if got := runValue(t, argumentMutant); got != "6" {
		t.Fatalf("dropped-argument mutant result=%q, want the detectable wrong value 6", got)
	}
	droppedResult := fmt.Sprintf("%s(void)%s(%s, %s); /* call: %s */", parts[1], parts[3], parts[4], parts[5], parts[6])
	resultMutant := generated[:lineStart] + droppedResult + generated[lineEnd:]
	if got := runValue(t, resultMutant); got != "5" {
		t.Fatalf("dropped-result mutant result=%q, want the detectable wrong value 5", got)
	}

	pathMutant := strings.Replace(generated, "*child = &schway_invocation_arena[offset];", "*child = parent;", 1)
	pathRun, pathErr := native.DefaultRunner().Run(context.Background(), pathMutant, "-O0", []string{"5"})
	if pathErr != nil || len(pathRun.Pairs) != 1 || pathRun.Pairs[0].Execution.Outcome.Value != "11" {
		t.Fatalf("path-lifetime mutant changed the scalar answer: result=%+v err=%v", pathRun, pathErr)
	}
	wantPath := phase36Invocation(t, phase36FunctionID(t, checked.Program.Functions, "main"), execution.InvocationSegment{OpCallID: call.ID, Ordinal: 0})
	for _, event := range pathRun.Pairs[0].Execution.Events {
		if event.Invocation == wantPath {
			t.Fatalf("reused-parent-path mutant unexpectedly retained child identity %q", wantPath)
		}
	}

	wrappedSource := `module phase36.wrapped_scalar_call

export {
  fn main
}

fn increment(value: U64) -> U64 {
  let one = 1
  let result = value + one
  result
}

fn relay(value: U64) -> U64 {
  let result = increment(value)
  result
}

fn main(input: U64) -> U64 {
  let result = relay(input)
  let answer = result + input
  answer
}
`
	wrapped := session.Check([]byte(wrappedSource))
	if len(wrapped.Diagnostics) != 0 {
		t.Fatalf("wrapped scalar-call diagnostics: %+v", wrapped.Diagnostics)
	}
	wrappedC, wrappedErr := cgen.EmitNative(wrapped.Program)
	if wrappedErr != nil {
		t.Fatalf("EmitNative refused a scalar helper whose body is one checked OpCall and return: %v", wrappedErr)
	}
	wrappedResult, wrappedRunErr := native.DefaultRunner().Run(context.Background(), wrappedC, "-O0", []string{"5"})
	if wrappedRunErr != nil || len(wrappedResult.Pairs) != 1 || wrappedResult.Pairs[0].Execution.Outcome.Value != "11" {
		t.Fatalf("wrapped scalar call native result=%+v err=%v; want returned 11", wrappedResult, wrappedRunErr)
	}
	wrappedMainID := phase36FunctionID(t, wrapped.Program.Functions, "main")
	wrappedRelayID := phase36FunctionID(t, wrapped.Program.Functions, "relay")
	mainToRelay := phase36CallIDs(t, wrapped.Program.Functions, "main", "relay")
	relayToIncrement := phase36CallIDs(t, wrapped.Program.Functions, "relay", "increment")
	if len(mainToRelay) != 1 || len(relayToIncrement) != 1 {
		t.Fatalf("wrapped call inventory main->relay=%v relay->increment=%v", mainToRelay, relayToIncrement)
	}
	outer := phase36Invocation(t, wrappedMainID, execution.InvocationSegment{OpCallID: mainToRelay[0], Ordinal: 0})
	inner := phase36Invocation(t, wrappedMainID, execution.InvocationSegment{OpCallID: mainToRelay[0], Ordinal: 0}, execution.InvocationSegment{OpCallID: relayToIncrement[0], Ordinal: 0})
	if wrappedRelayID == "" {
		t.Fatal("relay function identity was not assigned")
	}
	wrappedPaths := make(map[string]bool)
	for _, event := range wrappedResult.Pairs[0].Execution.Events {
		wrappedPaths[event.Invocation] = true
	}
	if !wrappedPaths[outer] || !wrappedPaths[inner] {
		t.Fatalf("wrapped dynamic paths missing after helper return: outer=%q present=%v inner=%q present=%v", outer, wrappedPaths[outer], inner, wrappedPaths[inner])
	}
}
