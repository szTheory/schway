package session_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestTogglePipeline(t *testing.T) {
	result, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase1", "toggle.lang"))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
	}
	if result.Program.Schema != "lang.core/0" || len(result.Program.Functions) != 1 {
		t.Fatalf("unexpected core: %+v", result.Program)
	}
	function := result.Program.Functions[0]
	if function.EntryPointID == "" || function.ReturnPointID == "" || function.Match.PointID == "" || len(function.Match.Arms) == 0 || function.Match.Arms[0].EdgeID == "" {
		t.Fatalf("typed core omitted stable point/edge identities: %+v", function)
	}
}

func TestNativeToggleO0O3(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	result, diagnostics, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("native run failed: err=%v diagnostics=%+v", err, diagnostics)
	}
	if len(result.O0.Pairs) != 2 || len(result.O3.Pairs) != 2 {
		t.Fatalf("unexpected native results: O0=%+v O3=%+v", result.O0, result.O3)
	}
	if !strings.Contains(result.CSource, "typedef enum LANG_SWITCH") || !strings.Contains(result.CSource, "switch (LANG_STATE)") {
		t.Fatalf("generated C is not reviewable S1 lowering:\n%s", result.CSource)
	}
}

func TestNativeToolFailureIsOperational(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	_, diagnostics, err := session.RunNativeFile(context.Background(), path, native.Runner{ClangPath: testsupport.ProjectPath("missing-clang")})
	if len(diagnostics) != 0 {
		t.Fatalf("tool absence became source diagnostics: %+v", diagnostics)
	}
	var toolError *native.ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.tool_missing" {
		t.Fatalf("expected native.tool_missing, got %T %v", err, err)
	}
}

func TestNonExhaustiveMatch(t *testing.T) {
	result, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase1", "non_exhaustive.lang"))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "match.non_exhaustive" {
		t.Fatalf("expected match.non_exhaustive, got %+v", result.Diagnostics)
	}
	if got := result.Diagnostics[0].Causes[0].Detail; got != "On" {
		t.Fatalf("expected missing On, got %q", got)
	}
}

func TestCLIToggleTracer(t *testing.T) {
	result, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase1", "toggle.lang"))
	if err != nil || len(result.Diagnostics) != 0 {
		t.Fatalf("check tracer failed: err=%v diagnostics=%+v", err, result.Diagnostics)
	}
}

func TestInterpreterDeterministic(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	first, diagnostics, err := session.RunInterpreterFile(path)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("first run failed: err=%v diagnostics=%+v", err, diagnostics)
	}
	second, diagnostics, err := session.RunInterpreterFile(path)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("second run failed: err=%v diagnostics=%+v", err, diagnostics)
	}
	firstBytes, _ := json.Marshal(first)
	secondBytes, _ := json.Marshal(second)
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("nondeterministic executions:\n%s\n%s", firstBytes, secondBytes)
	}
	if len(first) != 2 || first[0].Outcome.Value != "On" || first[1].Outcome.Value != "Off" {
		t.Fatalf("unexpected toggle executions: %+v", first)
	}
}

func TestConcurrentReadOnlyCommands(t *testing.T) {
	projectRoot := testsupport.ProjectPath()
	fixture := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	before, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	beforeDigest := sha256.Sum256(before)

	binary := filepath.Join(t.TempDir(), "lang")
	build := exec.Command("go", "build", "-o", binary, "./cmd/lang")
	build.Dir = projectRoot
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}

	type commandResult struct {
		output []byte
		err    error
	}
	formatResults := make([]commandResult, 2)
	checkResults := make([]commandResult, 2)
	var group sync.WaitGroup
	for index := 0; index < 2; index++ {
		index := index
		group.Add(2)
		go func() {
			defer group.Done()
			formatResults[index].output, formatResults[index].err = exec.Command(binary, "format", fixture).CombinedOutput()
		}()
		go func() {
			defer group.Done()
			checkResults[index].output, checkResults[index].err = exec.Command(binary, "format", "--check", fixture).CombinedOutput()
		}()
	}
	group.Wait()

	for index, result := range append(formatResults, checkResults...) {
		if result.err != nil {
			t.Fatalf("concurrent command %d failed: %v\n%s", index, result.err, result.output)
		}
	}
	if !bytes.Equal(formatResults[0].output, formatResults[1].output) || !bytes.Equal(formatResults[0].output, before) {
		t.Fatalf("concurrent format output drifted:\nfirst=%q\nsecond=%q", formatResults[0].output, formatResults[1].output)
	}
	if !bytes.Equal(checkResults[0].output, checkResults[1].output) {
		t.Fatalf("concurrent check output drifted:\nfirst=%q\nsecond=%q", checkResults[0].output, checkResults[1].output)
	}
	after, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if afterDigest := sha256.Sum256(after); afterDigest != beforeDigest {
		t.Fatalf("read-only commands changed fixture digest: before=%x after=%x", beforeDigest, afterDigest)
	}
}

func TestVerifyCorpus(t *testing.T) {
	result := session.VerifyCorpusFile(context.Background(), testsupport.ProjectPath("testdata", "phase1"), native.DefaultRunner())
	if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
		t.Fatalf("verify failed: status=%s diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	if len(result.Lanes) != 5 || result.Metrics.RecomputedWork == 0 || result.Metrics.ElapsedNS <= 0 {
		t.Fatalf("verify omitted work or observations: lanes=%+v metrics=%+v", result.Lanes, result.Metrics)
	}
	for _, lane := range result.Lanes {
		if lane.Status != "pass" || lane.RecomputedWork == 0 || lane.PeakRSSStatus == "" {
			t.Fatalf("incomplete lane: %+v", lane)
		}
	}
}

func TestVerifyMutationControls(t *testing.T) {
	corpus := testsupport.ProjectPath("testdata", "phase1")
	engine := session.VerifyCorpus(context.Background(), corpus, native.DefaultRunner(), session.VerifyOptions{ForceEngineMismatch: true})
	if engine.Status != protocol.StatusMismatch || protocol.ExitCode(engine.Status) != 4 || len(engine.Diagnostics) == 0 || engine.Diagnostics[0].Code != "native.engine_mismatch" {
		t.Fatalf("forced engine mismatch escaped: status=%s diagnostics=%+v", engine.Status, engine.Diagnostics)
	}
	stale := session.VerifyCorpus(context.Background(), corpus, native.DefaultRunner(), session.VerifyOptions{ForceStaleManifest: true})
	if stale.Status == protocol.StatusPass || protocol.ExitCode(stale.Status) == 0 || len(stale.Diagnostics) == 0 || stale.Diagnostics[0].Code != "evidence.source_mismatch" {
		t.Fatalf("forced stale manifest escaped: status=%s diagnostics=%+v", stale.Status, stale.Diagnostics)
	}
}
