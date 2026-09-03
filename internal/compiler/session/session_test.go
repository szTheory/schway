package session_test

import (
	"testing"

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
