package cgen_test

import (
	"context"
	"errors"
	"fmt"
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
