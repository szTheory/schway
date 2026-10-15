package cgen_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
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

func phase36EscapedCallID(value string) string {
	const hex = "0123456789ABCDEF"
	var escaped strings.Builder
	for _, b := range []byte(value) {
		if b == ':' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || b == '-' || b == '.' || b == '_' || b == '~' {
			escaped.WriteByte(b)
			continue
		}
		escaped.WriteByte('%')
		escaped.WriteByte(hex[b>>4])
		escaped.WriteByte(hex[b&0x0f])
	}
	return escaped.String()
}

func phase36CheckEmitterCallSiteCensus(functionRegion string, callIDs []string) error {
	expectedCalls := make(map[string]int, len(callIDs))
	expectedEvents := make(map[string]int, len(callIDs))
	expectedReservations := make(map[string]int, len(callIDs))
	for _, callID := range callIDs {
		expectedCalls[callID]++
		expectedEvents[callID+":event:called"]++
		expectedReservations[phase36EscapedCallID(callID)]++
	}
	callPattern := regexp.MustCompile(`/\* call: ([^*]+) \*/`)
	eventPattern := regexp.MustCompile(`schway_record_event\("function\.called",\s*"([^"]+)"`)
	reservationPattern := regexp.MustCompile(`schway_reserve_invocation\([^,\n]+,\s*"([^"]+)"`)
	calls := callPattern.FindAllStringSubmatch(functionRegion, -1)
	events := eventPattern.FindAllStringSubmatch(functionRegion, -1)
	reservations := reservationPattern.FindAllStringSubmatch(functionRegion, -1)
	checkCategory := func(category string, expected map[string]int, matches [][]string) error {
		actual := make(map[string]int, len(matches))
		for _, match := range matches {
			actual[match[1]]++
		}
		ids := make(map[string]struct{}, len(expected)+len(actual))
		for id := range expected {
			ids[id] = struct{}{}
		}
		for id := range actual {
			ids[id] = struct{}{}
		}
		ordered := make([]string, 0, len(ids))
		for id := range ids {
			ordered = append(ordered, id)
		}
		sort.Strings(ordered)
		for _, id := range ordered {
			want, got := expected[id], actual[id]
			if got < want {
				return fmt.Errorf("missing %s site %q: got %d, want %d", category, id, got, want)
			}
			if got > want {
				return fmt.Errorf("extra %s site %q: got %d, want %d", category, id, got, want)
			}
		}
		return nil
	}
	if err := checkCategory("call", expectedCalls, calls); err != nil {
		return err
	}
	if err := checkCategory("event", expectedEvents, events); err != nil {
		return err
	}
	return checkCategory("reservation", expectedReservations, reservations)
}

func phase36FunctionRegionByHeader(generated, header string) (string, bool) {
	start := strings.Index(generated, header)
	if start < 0 {
		return "", false
	}
	bodyStart := start + len(header)
	functionEndOffset := strings.Index(generated[bodyStart:], "\n}\n")
	if functionEndOffset < 0 {
		return "", false
	}
	return generated[start : bodyStart+functionEndOffset+3], true
}

func phase36ReplaceLineContaining(generated, fragment, replacement string) (string, string, bool) {
	index := strings.Index(generated, fragment)
	if index < 0 {
		return generated, "", false
	}
	start := strings.LastIndex(generated[:index], "\n") + 1
	end := strings.IndexByte(generated[index:], '\n')
	if end < 0 {
		end = len(generated)
	} else {
		end += index
	}
	line := generated[start:end]
	return generated[:start] + replacement + generated[end:], line, true
}

func phase36InsertLineAfterContaining(generated, fragment, newLine string) (string, bool) {
	index := strings.Index(generated, fragment)
	if index < 0 {
		return generated, false
	}
	end := strings.IndexByte(generated[index:], '\n')
	if end < 0 {
		return generated, false
	}
	end += index + 1
	return generated[:end] + newLine + "\n" + generated[end:], true
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
	functionHeaderEnd := strings.IndexByte(functionRegion, '\n')
	if functionHeaderEnd < 0 {
		t.Fatal("dynamic call function has no complete C signature")
	}
	functionHeader := functionRegion[:functionHeaderEnd+1]
	if strings.Count(functionRegion, "schway_record_event(\"function.called\"") != 1 || strings.Count(functionRegion, "schway_reserve_invocation(") != 1 {
		t.Fatalf("user function call region does not enumerate one call event and one path reservation:\n%s", functionRegion)
	}
	if err := phase36CheckEmitterCallSiteCensus(functionRegion, []string{call.ID}); err != nil {
		t.Fatalf("clean user-function call-site census: %v", err)
	}
	if err := phase36CheckEmitterCallSiteCensus(functionRegion, []string{call.ID}); err != nil {
		t.Fatalf("no-op user-function call-site census: %v", err)
	}
	calledOffset := strings.Index(functionRegion, "schway_record_event(\"function.called\"")
	reserveOffset := strings.Index(functionRegion, "schway_reserve_invocation(")
	callOffset := strings.Index(functionRegion, callMarker)
	if calledOffset < 0 || reserveOffset <= calledOffset || callOffset <= reserveOffset {
		t.Fatalf("call event, path reservation and value call are out of order: called=%d reserve=%d call=%d", calledOffset, reserveOffset, callOffset)
	}
	eventFragment := fmt.Sprintf("schway_record_event(\"function.called\", %q", call.ID+":event:called")
	reserveFragment := fmt.Sprintf("schway_reserve_invocation(schway_invocations[invocation_index], %q,", phase36EscapedCallID(call.ID))
	mutations := []struct {
		name      string
		fragment  string
		inserted  string
		wantError string
	}{
		{name: "removed-call", fragment: callMarker, wantError: fmt.Sprintf("missing call site %q", call.ID)},
		{name: "removed-event", fragment: eventFragment, wantError: fmt.Sprintf("missing event site %q", call.ID+":event:called")},
		{name: "removed-reservation", fragment: reserveFragment, wantError: fmt.Sprintf("missing reservation site %q", phase36EscapedCallID(call.ID))},
		{name: "extra-call", fragment: callMarker, wantError: fmt.Sprintf("extra call site %q", call.ID+"-extra")},
		{name: "extra-event", fragment: eventFragment, wantError: fmt.Sprintf("extra event site %q", call.ID+":event:called:extra")},
		{name: "extra-reservation", fragment: reserveFragment, wantError: fmt.Sprintf("extra reservation site %q", phase36EscapedCallID(call.ID)+"-extra")},
	}
	for _, mutation := range mutations {
		t.Run("census-"+mutation.name, func(t *testing.T) {
			var mutated string
			var ok bool
			if strings.HasPrefix(mutation.name, "removed-") {
				mutated, _, ok = phase36ReplaceLineContaining(generated, mutation.fragment, "")
			} else {
				var extraLine string
				switch mutation.name {
				case "extra-call":
					extraLine = strings.Replace(callLine, callMarker, "/* call: "+call.ID+"-extra */", 1)
				case "extra-event":
					_, eventLine, found := phase36ReplaceLineContaining(generated, eventFragment, "")
					if !found {
						t.Fatal("could not locate emitted function.called event line")
					}
					extraLine = strings.Replace(eventLine, call.ID+":event:called", call.ID+":event:called:extra", 1)
				case "extra-reservation":
					_, reserveLine, found := phase36ReplaceLineContaining(generated, reserveFragment, "")
					if !found {
						t.Fatal("could not locate emitted invocation reservation line")
					}
					extraLine = strings.Replace(reserveLine, `"`+phase36EscapedCallID(call.ID)+`"`, `"`+phase36EscapedCallID(call.ID)+`-extra"`, 1)
				}
				mutated, ok = phase36InsertLineAfterContaining(generated, mutation.fragment, extraLine)
			}
			if !ok {
				t.Fatalf("could not apply %s mutation to the emitted user-function region", mutation.name)
			}
			mutatedRegion, found := phase36FunctionRegionByHeader(mutated, functionHeader)
			if !found {
				t.Fatalf("%s mutation lost the emitted user-function boundary", mutation.name)
			}
			censusErr := phase36CheckEmitterCallSiteCensus(mutatedRegion, []string{call.ID})
			if censusErr == nil || !strings.Contains(censusErr.Error(), mutation.wantError) {
				t.Fatalf("%s mutation census error=%v, want named finding containing %q", mutation.name, censusErr, mutation.wantError)
			}
			if restoreErr := phase36CheckEmitterCallSiteCensus(functionRegion, []string{call.ID}); restoreErr != nil {
				t.Fatalf("restored user-function census after %s mutation: %v", mutation.name, restoreErr)
			}
		})
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

	// Redirect an actual child invocation pointer to a local array in the
	// repeated nested helper. Each loop activation rewrites the same stack slot;
	// the event table retains that address after the helper returns.
	mutantChecked := session.Check([]byte(phase36NativeDynamicInvocationSource))
	if len(mutantChecked.Diagnostics) != 0 {
		t.Fatalf("stack-path mutation source diagnostics: %+v", mutantChecked.Diagnostics)
	}
	mutantGenerated, mutantEmitErr := cgen.EmitNative(mutantChecked.Program)
	if mutantEmitErr != nil {
		t.Fatalf("EmitNative for stack-path lifetime control: %v", mutantEmitErr)
	}
	mainID := phase36FunctionID(t, mutantChecked.Program.Functions, "main")
	outerCalls := phase36CallIDs(t, mutantChecked.Program.Functions, "main", "nested")
	innerCalls := phase36CallIDs(t, mutantChecked.Program.Functions, "nested", "increment")
	if len(outerCalls) != 2 || len(innerCalls) != 1 {
		t.Fatalf("stack-path repeated-site inventory outer=%v inner=%v", outerCalls, innerCalls)
	}
	firstRetainedPath := phase36Invocation(t, mainID,
		execution.InvocationSegment{OpCallID: outerCalls[0], Ordinal: 0},
		execution.InvocationSegment{OpCallID: innerCalls[0], Ordinal: 0})
	lastRetainedPath := phase36Invocation(t, mainID,
		execution.InvocationSegment{OpCallID: outerCalls[1], Ordinal: 1},
		execution.InvocationSegment{OpCallID: innerCalls[0], Ordinal: 1})
	innerMarker := "/* call: " + innerCalls[0] + " */"
	innerMarkerOffset := strings.Index(mutantGenerated, innerMarker)
	if innerMarkerOffset < 0 {
		t.Fatal("repeated nested helper has no emitted inner call marker")
	}
	innerLineStart := strings.LastIndex(mutantGenerated[:innerMarkerOffset], "\n") + 1
	innerFunctionStart := strings.LastIndex(mutantGenerated[:innerLineStart], "\nstatic ") + 1
	innerFunctionEndOffset := strings.Index(mutantGenerated[innerLineStart:], "\n}\n")
	if innerFunctionStart == 0 || innerFunctionEndOffset < 0 {
		t.Fatal("repeated nested helper call has no enclosing emitted function")
	}
	innerRegion := mutantGenerated[innerFunctionStart : innerLineStart+innerFunctionEndOffset+3]
	innerHeaderEnd := strings.IndexByte(innerRegion, '\n')
	if innerHeaderEnd < 0 {
		t.Fatal("repeated nested helper has no complete C signature")
	}
	innerHeader := innerRegion[:innerHeaderEnd+1]
	innerReserveFragment := fmt.Sprintf("schway_reserve_invocation(schway_invocations[invocation_index], %q,", phase36EscapedCallID(innerCalls[0]))
	_, innerReserveLine, reserveFound := phase36ReplaceLineContaining(mutantGenerated, innerReserveFragment, "")
	if !reserveFound {
		t.Fatal("could not locate repeated nested helper reservation")
	}
	reserveNames := regexp.MustCompile(`&([A-Za-z0-9_]+), &([A-Za-z0-9_]+)\)\) abort\(\);`).FindStringSubmatch(innerReserveLine)
	if len(reserveNames) != 3 {
		t.Fatalf("repeated nested helper reservation locals are not enumerable: %q", innerReserveLine)
	}
	childName := reserveNames[2]
	assignmentFragment := " = " + childName + ";"
	_, assignmentLine, assignmentFound := phase36ReplaceLineContaining(mutantGenerated, assignmentFragment, "")
	if !assignmentFound {
		t.Fatal("could not locate emitted invocation-table pointer assignment")
	}
	trimmedAssignment := strings.TrimSpace(assignmentLine)
	assignmentParts := strings.SplitN(trimmedAssignment, " = ", 2)
	if len(assignmentParts) != 2 || !strings.HasSuffix(assignmentParts[1], ";") {
		t.Fatalf("invocation-table assignment has unexpected form %q", assignmentLine)
	}
	indent := assignmentLine[:len(assignmentLine)-len(strings.TrimLeft(assignmentLine, " \t"))]
	const stackBuffer = "phase36_stack_path_control"
	const stackLength = "phase36_stack_path_control_length"
	stackDeclarations := "  char " + stackBuffer + "[512];\n  size_t " + stackLength + ";\n"
	buildStackMutation := func(redirectPointer bool) string {
		mutated := strings.Replace(mutantGenerated, innerHeader, innerHeader+stackDeclarations, 1)
		pointer := childName
		if redirectPointer {
			pointer = stackBuffer
		}
		replacement := strings.Join([]string{
			indent + stackLength + " = strlen(" + childName + ") + 1u;",
			indent + "if (" + stackLength + " > sizeof " + stackBuffer + ") abort();",
			indent + "memcpy(" + stackBuffer + ", " + childName + ", " + stackLength + ");",
			indent + assignmentParts[0] + " = " + pointer + ";",
		}, "\n")
		result, _, found := phase36ReplaceLineContaining(mutated, assignmentFragment, replacement)
		if !found {
			t.Fatal("could not apply stack-path pointer mutation")
		}
		return result
	}

	type phase36NativeControlRun struct {
		value  string
		kind   string
		events []execution.Event
	}
	runDynamicControl := func(t *testing.T, source string) phase36NativeControlRun {
		t.Helper()
		result, runErr := native.DefaultRunner().Run(context.Background(), source, "-O0", []string{"3"})
		if runErr != nil || len(result.Pairs) != 1 {
			t.Fatalf("native stack-path control pairs=%d err=%v", len(result.Pairs), runErr)
		}
		execution := result.Pairs[0].Execution
		return phase36NativeControlRun{value: execution.Outcome.Value, kind: execution.Outcome.Kind, events: execution.Events}
	}
	runRawStackMutant := func(t *testing.T, source string) phase36NativeControlRun {
		t.Helper()
		binaryPath, cleanup, compileErr := native.DefaultRunner().CompileOnly(context.Background(), source, "-O0")
		if compileErr != nil {
			t.Fatalf("installed Clang compile for unsafe stack-path mutant: %v", compileErr)
		}
		defer cleanup()
		command := exec.Command(binaryPath, "3")
		var stdout, stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr
		if runErr := command.Run(); runErr != nil {
			t.Fatalf("compiled stack-path mutant native run failed: %v stderr=%q", runErr, stderr.String())
		}
		var raw execution.Execution
		if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
			t.Fatalf("compiled stack-path mutant did not emit a parseable execution document: %v stdout=%q", err, stdout.String())
		}
		return phase36NativeControlRun{value: raw.Outcome.Value, kind: raw.Outcome.Kind, events: raw.Events}
	}
	retainedPaths := func(run phase36NativeControlRun, paths ...string) error {
		seen := make(map[string]bool, len(run.events))
		for _, event := range run.events {
			seen[event.Invocation] = true
		}
		for _, path := range paths {
			if !seen[path] {
				return fmt.Errorf("retained post-return invocation path %q missing", path)
			}
		}
		return nil
	}
	cleanRun := runDynamicControl(t, mutantGenerated)
	if cleanRun.kind != "returned" || cleanRun.value != "10" {
		t.Fatalf("clean path lifetime control outcome=%s:%s, want returned 10", cleanRun.kind, cleanRun.value)
	}
	if err := retainedPaths(cleanRun, firstRetainedPath, lastRetainedPath); err != nil {
		t.Fatalf("clean positive: %v", err)
	}
	noOpRun := runDynamicControl(t, buildStackMutation(false))
	if noOpRun.kind != "returned" || noOpRun.value != "10" {
		t.Fatalf("no-op path lifetime control outcome=%s:%s, want returned 10", noOpRun.kind, noOpRun.value)
	}
	if err := retainedPaths(noOpRun, firstRetainedPath, lastRetainedPath); err != nil {
		t.Fatalf("no-op positive: %v", err)
	}
	stackRun := runRawStackMutant(t, buildStackMutation(true))
	if stackRun.kind != "returned" || stackRun.value != "10" || len(stackRun.events) == 0 {
		t.Fatalf("stack-path mutant must compile and return the real answer before path assertion: outcome=%s:%s events=%d", stackRun.kind, stackRun.value, len(stackRun.events))
	}
	if err := retainedPaths(stackRun, firstRetainedPath); err == nil || !strings.Contains(err.Error(), "retained post-return invocation path") {
		t.Fatalf("unsafe returned-stack-path mutant escaped the named retained-path assertion: %v", err)
	}
	restoredRun := runDynamicControl(t, mutantGenerated)
	if restoredRun.kind != "returned" || restoredRun.value != "10" {
		t.Fatalf("restored path lifetime control outcome=%s:%s, want returned 10", restoredRun.kind, restoredRun.value)
	}
	if err := retainedPaths(restoredRun, firstRetainedPath, lastRetainedPath); err != nil {
		t.Fatalf("restored positive: %v", err)
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
