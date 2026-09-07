package session

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
)

// TestStageBreakdownCoversEveryPipelineStage is Task 1's tracer: with
// LANG_OBSERVE_TIMING=1 and a cold cache/change-state pair (forcing a real
// compile+link, not a cache-artifact-reuse), running the native-differential
// lane end-to-end over a real corpus produces a stage_breakdown with one
// entry per StageNames() name, in that exact pipeline order.
func TestStageBreakdownCoversEveryPipelineStage(t *testing.T) {
	t.Setenv("LANG_OBSERVE_TIMING", "1")
	corpus := nat03CorpusPath("testdata/phase1")
	phase6TestRoots(t)

	result, err := VerifyPhase6ChangedRisk(context.Background(), corpus, native.DefaultRunner())
	if err != nil {
		t.Fatalf("VerifyPhase6ChangedRisk error: %v", err)
	}

	lane := phase6FindLane(result, phase6NativeDifferentialLane)
	if lane == nil {
		t.Fatalf("no lane %s in result; lanes=%+v", phase6NativeDifferentialLane, result.Lanes)
	}

	wantStages := StageNames()
	if len(lane.StageBreakdown) != len(wantStages) {
		t.Fatalf("StageBreakdown has %d entries, want %d (%v): %+v", len(lane.StageBreakdown), len(wantStages), wantStages, lane.StageBreakdown)
	}
	for index, want := range wantStages {
		got := lane.StageBreakdown[index]
		if got.Stage != want {
			t.Fatalf("StageBreakdown[%d].Stage = %q, want %q (breakdown=%+v)", index, got.Stage, want, lane.StageBreakdown)
		}
	}
}

// TestStageRecorderRefusesUnknownStage covers Task 1's remaining acceptance
// criteria: Stop on an unknown stage, Stop without a matching Start, and
// Start on an unknown stage all return a typed *StageRecordError rather
// than silently recording a zero elapsed value.
func TestStageRecorderRefusesUnknownStage(t *testing.T) {
	recorder := &StageRecorder{}

	if err := recorder.Stop("nonexistent"); err == nil {
		t.Fatal("Stop(\"nonexistent\") did not return an error")
	} else if _, ok := err.(*StageRecordError); !ok {
		t.Fatalf("Stop(\"nonexistent\") returned %T, want *StageRecordError", err)
	}

	if err := recorder.Stop("parse"); err == nil {
		t.Fatal("Stop(\"parse\") without a prior Start(\"parse\") did not return an error")
	} else if _, ok := err.(*StageRecordError); !ok {
		t.Fatalf("Stop(\"parse\") without Start returned %T, want *StageRecordError", err)
	}

	if err := recorder.Start("nonexistent"); err == nil {
		t.Fatal("Start(\"nonexistent\") did not return an error")
	} else if _, ok := err.(*StageRecordError); !ok {
		t.Fatalf("Start(\"nonexistent\") returned %T, want *StageRecordError", err)
	}

	t.Setenv("LANG_OBSERVE_TIMING", "1")
	if breakdown := recorder.Breakdown(); breakdown != nil {
		t.Fatalf("Breakdown() = %+v, want nil -- no stage was ever successfully started and stopped", breakdown)
	}
}

// TestStageRecorderFileDeclaresNoMutableStateOrGoroutines is Task 1's
// remaining acceptance criterion 5: a go/ast assertion proves
// session_phase6_stages.go declares no package-level mutable state and
// starts no goroutine -- the structural proof that StageRecorder is the
// fixed-size-array mechanism the plan requires, not a global registry or a
// tracing runtime in disguise (PROJECT.md's runtime-posture constraint).
func TestStageRecorderFileDeclaresNoMutableStateOrGoroutines(t *testing.T) {
	fileSet := token.NewFileSet()
	path := nat03CorpusPath("internal/compiler/session/session_phase6_stages.go")
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		t.Fatalf("parser.ParseFile: %v", err)
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		t.Fatalf("package-level var declaration found at %v -- session_phase6_stages.go must declare no package-level mutable state", fileSet.Position(genDecl.Pos()))
	}

	ast.Inspect(file, func(n ast.Node) bool {
		if goStmt, ok := n.(*ast.GoStmt); ok {
			t.Fatalf("go statement found at %v -- session_phase6_stages.go must start no goroutine", fileSet.Position(goStmt.Pos()))
		}
		return true
	})
}

// TestStageTimingIsGatedOnObserveTiming is Task 2's Behavior Tests 1 and 2:
// with the timing switch unset, no lane carries a stage_breakdown key at
// all (not an empty array, not zeros); with the switch set, the
// native-differential lane carries a full, pipeline-ordered breakdown.
func TestStageTimingIsGatedOnObserveTiming(t *testing.T) {
	corpus := nat03CorpusPath("testdata/phase1")

	t.Run("unset", func(t *testing.T) {
		phase6TestRoots(t)
		result, err := VerifyPhase6ChangedRisk(context.Background(), corpus, native.DefaultRunner())
		if err != nil {
			t.Fatalf("VerifyPhase6ChangedRisk error: %v", err)
		}
		encoded, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if strings.Contains(string(encoded), "stage_breakdown") {
			t.Fatalf("stage_breakdown key present with LANG_OBSERVE_TIMING unset: %s", encoded)
		}
		lane := phase6FindLane(result, phase6NativeDifferentialLane)
		if lane == nil {
			t.Fatalf("no lane %s in result; lanes=%+v", phase6NativeDifferentialLane, result.Lanes)
		}
		if lane.StageBreakdown != nil {
			t.Fatalf("lane.StageBreakdown = %+v, want nil with timing unset", lane.StageBreakdown)
		}
	})

	t.Run("set", func(t *testing.T) {
		t.Setenv("LANG_OBSERVE_TIMING", "1")
		phase6TestRoots(t)
		result, err := VerifyPhase6ChangedRisk(context.Background(), corpus, native.DefaultRunner())
		if err != nil {
			t.Fatalf("VerifyPhase6ChangedRisk error: %v", err)
		}
		lane := phase6FindLane(result, phase6NativeDifferentialLane)
		if lane == nil {
			t.Fatalf("no lane %s in result; lanes=%+v", phase6NativeDifferentialLane, result.Lanes)
		}
		if !reflect.DeepEqual(stageTimingNames(lane.StageBreakdown), StageNames()) {
			t.Fatalf("StageBreakdown stages = %v, want %v (full pipeline order)", stageTimingNames(lane.StageBreakdown), StageNames())
		}
	})
}

// stageTimingNames projects a []protocol.StageTiming down to its ordered
// Stage names, for comparison against StageNames().
func stageTimingNames(entries []protocol.StageTiming) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Stage)
	}
	return names
}

// TestStageBreakdownAbsentWhenUnobserved is Task 2's Behavior Test 5's
// session-level half: with LANG_OBSERVE_TIMING unset, the marshalled verify
// Result carries no stage-breakdown key on any lane, and the absence
// itself -- not full-document byte identity, which lane.ElapsedNS's own
// unconditional per-invocation wall-clock recording already precludes,
// independent of this plan -- is stable across repeated invocations.
func TestStageBreakdownAbsentWhenUnobserved(t *testing.T) {
	corpus := nat03CorpusPath("testdata/phase1")

	for run := 0; run < 2; run++ {
		phase6TestRoots(t)
		result, err := VerifyPhase6ChangedRisk(context.Background(), corpus, native.DefaultRunner())
		if err != nil {
			t.Fatalf("run %d: VerifyPhase6ChangedRisk error: %v", run, err)
		}
		encoded, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("run %d: marshal: %v", run, err)
		}
		if strings.Contains(string(encoded), "stage_breakdown") {
			t.Fatalf("run %d: stage_breakdown present with LANG_OBSERVE_TIMING unset: %s", run, encoded)
		}
		for _, lane := range result.Lanes {
			if lane.StageBreakdown != nil {
				t.Fatalf("run %d: lane %s.StageBreakdown = %+v, want nil", run, lane.ID, lane.StageBreakdown)
			}
		}
	}
}

// TestTimingEnvironmentIsReadInExactlyOnePlace is Task 2's go/ast scan over
// every non-test .go file under internal/ and cmd/, asserting exactly one
// reference to the LANG_OBSERVE_TIMING string -- inside
// TimingObservationEnabled. Two readers is how the gate drifts, and a
// drifted gate is how a pinned JSON test becomes intermittently red
// (D-06-21).
func TestTimingEnvironmentIsReadInExactlyOnePlace(t *testing.T) {
	const envVarName = "LANG_OBSERVE_TIMING"
	roots := []string{
		nat03CorpusPath("internal"),
		nat03CorpusPath("cmd"),
	}
	fileSet := token.NewFileSet()

	type site struct {
		file string
		line int
		fn   string
	}
	var sites []site

	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, parseErr := parser.ParseFile(fileSet, path, nil, 0)
			if parseErr != nil {
				return parseErr
			}
			var currentFunc string
			ast.Inspect(file, func(n ast.Node) bool {
				if fn, ok := n.(*ast.FuncDecl); ok {
					currentFunc = fn.Name.Name
				}
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				value, unquoteErr := strconv.Unquote(lit.Value)
				if unquoteErr != nil {
					return true
				}
				if value == envVarName {
					sites = append(sites, site{file: path, line: fileSet.Position(lit.Pos()).Line, fn: currentFunc})
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}

	if len(sites) != 1 {
		t.Fatalf("found %d reference(s) to %q, want exactly 1: %+v", len(sites), envVarName, sites)
	}
	if sites[0].fn != "TimingObservationEnabled" {
		t.Fatalf("the sole %q reference is inside %s (%s:%d), want TimingObservationEnabled", envVarName, sites[0].fn, sites[0].file, sites[0].line)
	}
}
