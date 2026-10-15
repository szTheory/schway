package session_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase36EngineCallMutationControls(t *testing.T) {
	root := testsupport.ProjectPath()
	claimOwners := []struct {
		claim, knower, witness, limit string
	}{
		{"independent helper values and propagated defects", "literal fixture plus interpreter/native execution", "TestPhase36HelperAnswersAndDefects; TestPhase36HelperOptimizationMatrix", "numeric values and terminal reasons are observed at engine boundaries"},
		{"source intent for call argument/result shape", "checker peer", "check.TestPhase36CallSourceIntentControls", "does not execute the helper body"},
		{"raw call shape and removal controls", "corevalidate and originvalidate peers", "TestPhase36RawCallShapeAndFreshness; TestPhase36RawCallRefusalRemovalControls; TestPhase36OriginCallFreshnessAndTransfer; TestPhase36OriginCallRefusalRemovalControls", "checks checked-core/origin contracts, not runtime values"},
		{"dynamic call/terminal attribution", "execution causal peer", "TestPhase36CallCausalAttributionControls", "deliberately accepts a structurally plausible false-success numeric trace; the literal engine assertion owns that claim"},
	}
	for _, owner := range claimOwners {
		t.Logf("claim=%q knower=%q witness=%q limit=%q", owner.claim, owner.knower, owner.witness, owner.limit)
	}

	for _, mutation := range phase36EngineCallMutations() {
		mutation := mutation
		t.Run(mutation.name, func(t *testing.T) {
			implementationPath := filepath.Join(root, filepath.FromSlash(mutation.sourcePath))
			original, err := os.ReadFile(implementationPath)
			if err != nil {
				t.Fatalf("read %s implementation: %v", mutation.engine, err)
			}
			mutated, counts := phase36ApplyEngineMutation(string(original), mutation.edits)
			if len(counts) != len(mutation.edits) || mutated == string(original) {
				t.Fatalf("%s mutation did not change its intended engine source: matched=%v edits=%d", mutation.name, counts, len(mutation.edits))
			}

			before, beforeErr := phase36RunEngineMutationChild(t, root, implementationPath, original, mutation.selector, mutation.childEnv)
			if beforeErr != nil {
				t.Fatalf("fresh no-op overlay failed before %s: %v\n%s", mutation.name, beforeErr, before)
			}
			phase36RequireEngineMutationChild(t, before, mutation.targetTest, "PASS", "")

			mutantEnv := append(append([]string(nil), mutation.childEnv...), mutation.mutantEnv...)
			output, runErr := phase36RunEngineMutationChild(t, root, implementationPath, []byte(mutated), mutation.selector, mutantEnv)
			if runErr == nil {
				t.Fatalf("%s mutant survived its named engine assertion\n%s", mutation.name, output)
			}
			phase36RequireEngineMutationChild(t, output, mutation.targetTest, "FAIL", mutation.failureText)
			t.Logf("mutant killed engine=%s claim=%q target=%s matched=%v assertion=%q", mutation.engine, mutation.claim, mutation.targetTest, counts, mutation.failureText)

			restored, restoredErr := phase36RunEngineMutationChild(t, root, implementationPath, original, mutation.selector, mutation.childEnv)
			if restoredErr != nil {
				t.Fatalf("restored positive failed after %s: %v\n%s", mutation.name, restoredErr, restored)
			}
			phase36RequireEngineMutationChild(t, restored, mutation.targetTest, "PASS", "")
		})
	}
}

type phase36EngineMutationEdit struct {
	before string
	after  string
}

type phase36EngineCallMutation struct {
	name, engine, claim string
	sourcePath          string
	edits               []phase36EngineMutationEdit
	selector            string
	targetTest          string
	failureText         string
	childEnv            []string
	mutantEnv           []string
}

func phase36EngineCallMutations() []phase36EngineCallMutation {
	interpArgument := phase36EngineCallMutation{
		name: "interpreter-same-typed-argument-substitution", engine: "interpreter", claim: "numeric answer",
		sourcePath: "internal/compiler/interp/interp.go",
		edits:      []phase36EngineMutationEdit{{before: "argument := caller.values[operation.SourceID]", after: "argument := caller.values[caller.function.Parameter.ID]"}},
		selector:   "^TestPhase36HelperAnswersAndDefects$/^local-four-step-loop$", targetTest: "TestPhase36HelperAnswersAndDefects/local-four-step-loop", failureText: "interpreter outcome=",
	}
	interpResult := phase36EngineCallMutation{
		name: "interpreter-dropped-helper-result", engine: "interpreter", claim: "numeric answer",
		sourcePath: "internal/compiler/interp/interp.go",
		edits:      []phase36EngineMutationEdit{{before: "caller.values[returnTarget] = returnValue", after: "caller.values[returnTarget] = value{u64: 0, isU64: true}"}},
		selector:   "^TestPhase36HelperAnswersAndDefects$/^local-four-step-loop$", targetTest: "TestPhase36HelperAnswersAndDefects/local-four-step-loop", failureText: "interpreter outcome=",
	}
	interpOverflow := phase36EngineCallMutation{
		name: "interpreter-swallowed-helper-overflow", engine: "interpreter", claim: "callee overflow terminal",
		sourcePath: "internal/compiler/interp/interp.go",
		edits:      []phase36EngineMutationEdit{{before: "if ^uint64(0)-sourceValue.u64 < right.u64 {", after: "if false && ^uint64(0)-sourceValue.u64 < right.u64 {"}},
		selector:   "^TestPhase36HelperAnswersAndDefects$/^u64-helper-overflow$", targetTest: "TestPhase36HelperAnswersAndDefects/u64-helper-overflow", failureText: "interpreter outcome=",
	}
	interpRemainder := phase36EngineCallMutation{
		name: "interpreter-swallowed-helper-remainder", engine: "interpreter", claim: "callee remainder terminal",
		sourcePath: "internal/compiler/interp/interp.go",
		edits: []phase36EngineMutationEdit{
			{before: "if right.u64 == 0 {", after: "if false && right.u64 == 0 {"},
			{before: "top.values[operation.TargetID] = value{u64: sourceValue.u64 % right.u64, isU64: true}", after: "top.values[operation.TargetID] = value{u64: sourceValue.u64 % (right.u64 | 1), isU64: true}"},
		},
		selector: "^TestPhase36HelperAnswersAndDefects$/^bool-helper-zero-remainder$", targetTest: "TestPhase36HelperAnswersAndDefects/bool-helper-zero-remainder", failureText: "interpreter outcome=",
	}
	mutations := []phase36EngineCallMutation{interpArgument, interpResult, interpOverflow, interpRemainder}
	for _, lane := range []string{"O0", "O3", "O3-LTO"} {
		lanePattern := "^TestPhase36EngineMutationLaneChild$/^local-four-step-loop$/^" + lane + "$"
		laneTest := "TestPhase36EngineMutationLaneChild/local-four-step-loop/" + lane
		laneEnv := []string{"PHASE36_ENGINE_MUTATION_CASE=local-four-step-loop", "PHASE36_ENGINE_MUTATION_LANE=" + lane}
		mutations = append(mutations,
			phase36EngineCallMutation{
				name: "native-" + lane + "-same-typed-argument-substitution", engine: "native-cgen", claim: "numeric answer",
				sourcePath: "internal/compiler/cgen/cgen_program.go",
				edits: []phase36EngineMutationEdit{{
					before: `fmt.Fprintf(out, "%s%s = %s(%s, %s); /* call: %s */\n", indent, targetLocal, calleeName, argumentLocal, childTable, operation.ID)`,
					after:  `fmt.Fprintf(out, "%s%s = %s((uint64_t)(%s ^ UINT64_C(1)), %s); /* call: %s */\n", indent, targetLocal, calleeName, argumentLocal, childTable, operation.ID)`,
				}},
				selector: lanePattern, targetTest: laneTest, failureText: "native " + lane + " outcome=", childEnv: laneEnv,
			},
			phase36EngineCallMutation{
				name: "native-" + lane + "-dropped-helper-result", engine: "native-cgen", claim: "numeric answer",
				sourcePath: "internal/compiler/cgen/cgen_program.go",
				edits: []phase36EngineMutationEdit{{
					before: `fmt.Fprintf(out, "%s%s = %s(%s, %s); /* call: %s */\n", indent, targetLocal, calleeName, argumentLocal, childTable, operation.ID)`,
					after:  `fmt.Fprintf(out, "%s%s = %s(%s, %s); /* call: %s */\n%s%s = 0;\n", indent, targetLocal, calleeName, argumentLocal, childTable, operation.ID, indent, targetLocal)`,
				}},
				selector: lanePattern, targetTest: laneTest, failureText: "native " + lane + " outcome=", childEnv: laneEnv,
			},
		)
		for _, defect := range []struct{ name, lanePattern, laneTest string }{
			{"u64-helper-overflow", "^TestPhase36HelperOptimizationMatrix$/^u64-helper-overflow$/^" + lane + "$", "TestPhase36HelperOptimizationMatrix/u64-helper-overflow/" + lane},
			{"bool-helper-zero-remainder", "^TestPhase36HelperOptimizationMatrix$/^bool-helper-zero-remainder$/^" + lane + "$", "TestPhase36HelperOptimizationMatrix/bool-helper-zero-remainder/" + lane},
		} {
			mutations = append(mutations, phase36EngineCallMutation{
				name: "native-" + lane + "-swallowed-" + defect.name, engine: "native-cgen", claim: "dynamic callee defect terminal",
				sourcePath:  "internal/compiler/cgen/cgen_program.go",
				edits:       []phase36EngineMutationEdit{{before: `\\\"kind\\\":\\\"defect\\\",\\\"value\\\":\\\"\\\"`, after: `\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\\\"0\\\"`}},
				selector:    "^TestPhase36EngineMutationLaneChild$/^" + defect.name + "$/^" + lane + "$",
				targetTest:  "TestPhase36EngineMutationLaneChild/" + defect.name + "/" + lane,
				failureText: "native " + lane + " outcome=", childEnv: []string{"PHASE36_ENGINE_MUTATION_CASE=" + defect.name, "PHASE36_ENGINE_MUTATION_LANE=" + lane},
				mutantEnv: []string{"PHASE36_ENGINE_MUTATION_EXPECT_RETURNED=1"},
			})
		}
	}
	return mutations
}

func phase36ApplyEngineMutation(source string, edits []phase36EngineMutationEdit) (string, []int) {
	mutated := source
	counts := make([]int, 0, len(edits))
	for _, edit := range edits {
		count := strings.Count(mutated, edit.before)
		counts = append(counts, count)
		if count != 1 {
			return source, counts
		}
		mutated = strings.Replace(mutated, edit.before, edit.after, 1)
	}
	return mutated, counts
}

func phase36RunEngineMutationChild(t *testing.T, root, implementationPath string, source []byte, selector string, testEnv []string) (string, error) {
	t.Helper()
	overlayDir := t.TempDir()
	overlaySource := filepath.Join(overlayDir, filepath.Base(implementationPath))
	if err := os.WriteFile(overlaySource, source, 0o600); err != nil {
		return "", fmt.Errorf("write engine source overlay: %w", err)
	}
	overlayPath := filepath.Join(overlayDir, "overlay.json")
	encoded, err := json.Marshal(struct {
		Replace map[string]string `json:"Replace"`
	}{Replace: map[string]string{implementationPath: overlaySource}})
	if err != nil {
		return "", fmt.Errorf("encode engine source overlay: %w", err)
	}
	if err := os.WriteFile(overlayPath, encoded, 0o600); err != nil {
		return "", fmt.Errorf("write engine source overlay: %w", err)
	}
	command := exec.Command("go", "test", "-overlay="+overlayPath, "./internal/compiler/session", "-run="+selector, "-count=1", "-v", "-timeout=600s")
	command.Dir = root
	command.Env = phase36EngineMutationGoEnv(os.Environ(), testEnv)
	output, runErr := command.CombinedOutput()
	return string(output), runErr
}

func phase36EngineMutationGoEnv(environment, testEnv []string) []string {
	filtered := make([]string, 0, len(environment)+len(testEnv)+2)
	for _, entry := range environment {
		if !strings.HasPrefix(entry, "GOCACHE=") && !strings.HasPrefix(entry, "GOTOOLCHAIN=") && !strings.HasPrefix(entry, "PHASE36_ENGINE_MUTATION_") {
			filtered = append(filtered, entry)
		}
	}
	filtered = append(filtered, "GOCACHE=/tmp/ai-lang-verification-gocache", "GOTOOLCHAIN=local")
	return append(filtered, testEnv...)
}

func phase36RequireEngineMutationChild(t *testing.T, output, targetTest, status, assertion string) {
	t.Helper()
	if !strings.Contains(output, "=== RUN   "+targetTest) || !strings.Contains(output, "--- "+status+": "+targetTest) {
		t.Fatalf("child output did not prove target %q ran and ended %s:\n%s", targetTest, status, output)
	}
	if assertion != "" && !strings.Contains(output, assertion) {
		t.Fatalf("child %s for %q missed intended assertion %q:\n%s", status, targetTest, assertion, output)
	}
}

// TestPhase36EngineMutationLaneChild is a one-case native consumer for the
// source-overlay mutation harness. Keeping each child to one real runner
// invocation prevents the matrix-wide run-count invariant from turning a
// correctly selected case into a false negative.
func TestPhase36EngineMutationLaneChild(t *testing.T) {
	caseName := os.Getenv("PHASE36_ENGINE_MUTATION_CASE")
	laneName := os.Getenv("PHASE36_ENGINE_MUTATION_LANE")
	if caseName == "" || laneName == "" {
		t.Skip("only invoked by the Phase36 engine mutation harness")
	}
	lanes := map[string]struct {
		optimization string
		lto          bool
	}{
		"O0":     {optimization: "-O0"},
		"O3":     {optimization: "-O3"},
		"O3-LTO": {optimization: "-O3", lto: true},
	}
	lane, ok := lanes[laneName]
	if !ok {
		t.Fatalf("unknown native mutation lane %q", laneName)
	}
	t.Run(caseName+"/"+laneName, func(t *testing.T) {
		var checked session.CheckResult
		var expected phase36HelperCaseRow
		if caseName == "local-four-step-loop" {
			fixture := phase36ReadHelperValues(t)
			checked = phase36CheckFile(t, testsupport.ProjectPath("examples", "helper_calls", "crc_step.schway"))
			expected = phase36FindHelperCase(t, fixture, caseName)
		} else {
			fixture := phase36ReadHelperValues(t)
			var source string
			found := false
			for _, defect := range phase36DefectPrograms(t, fixture) {
				if defect.caseName == caseName {
					source, expected, found = defect.source, defect.expected, true
					break
				}
			}
			if !found {
				t.Fatalf("unknown native mutation case %q", caseName)
			}
			checked = session.Check([]byte(source))
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("%s source diagnostics: %+v", caseName, checked.Diagnostics)
			}
		}
		entry, err := callgraph.EntryFunction(checked.Program)
		if err != nil {
			t.Fatalf("%s entry function: %v", caseName, err)
		}
		pruned, err := callgraph.PruneToReachable(checked.Program, entry.ID)
		if err != nil {
			t.Fatalf("%s prune reachable program: %v", caseName, err)
		}
		generated, err := cgen.EmitNative(pruned)
		if err != nil {
			t.Fatalf("%s EmitNative: %v", caseName, err)
		}
		runner := native.DefaultRunner().EnableCommandRecording()
		runner.LTO = lane.lto
		if expected.ExpectedExit != 0 {
			runner.Expect = native.ExpectDefect
			if os.Getenv("PHASE36_ENGINE_MUTATION_EXPECT_RETURNED") == "1" {
				runner.Expect = native.ExpectValue
			}
		}
		got, err := runner.Run(context.Background(), generated, lane.optimization, []string{expected.Input})
		if err != nil || len(got.Pairs) != 1 {
			t.Fatalf("native %s %s pairs=%d err=%v", laneName, caseName, len(got.Pairs), err)
		}
		if expected.ExpectedExit != 0 {
			helperName := map[string]string{
				"u64-helper-overflow":        "add_one",
				"bool-helper-zero-remainder": "remainder_to_bool",
			}[caseName]
			helperID := phase36FunctionID(checked.Program.Functions, helperName)
			if helperID == "" || phase36CallsTo(got.Pairs[0].Execution.Events, helperID) == 0 {
				t.Fatalf("native %s defect trace does not prove entry into helper %q", laneName, helperName)
			}
		}
		if err := phase36CompareExecution("native "+laneName, got.Pairs[0].Execution, expected); err != nil {
			t.Fatal(err)
		}
		if lane.lto {
			commands := runner.LastCommandLines()
			if len(commands) == 0 {
				t.Fatal("O3-LTO mutation lane recorded no compiler command")
			}
			for _, command := range commands {
				if !phase36HasArgument(command, "-flto") || !phase36HasArgument(command, "-O3") {
					t.Fatalf("O3-LTO mutation command lacks -O3 -flto: %v", command)
				}
			}
		}
	})
}
