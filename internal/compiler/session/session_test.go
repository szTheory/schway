package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/native"
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
