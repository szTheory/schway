package session

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/native"
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
