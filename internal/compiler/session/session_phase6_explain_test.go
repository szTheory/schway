package session

import (
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/debugmap"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// This file is deliberately `package session` (internal), unlike every
// other *_test.go sibling in this directory (`package session_test`).
// buildExplainGraph/resolveExplainDepth are intentionally unexported (the
// plan's own artifact list names only session.ExplainCommandFile as new
// exported surface) -- Tasks 2/3's <behavior> bullets need to construct
// exact depth/span/binding shapes no real in-tree fixture happens to
// produce (real Cause construction sites in check.go never repeat a
// Kind+Detail pair within one diagnostic), so this file exercises the
// synthesizer directly rather than only through the CLI seam. Schema
// minting (Task 1) and cross-corpus determinism (Task 3) still drive
// ExplainCommandFile end-to-end against real fixtures.

func spanPtr(start, end int) *diagnostic.Span {
	span := diagnostic.Span{Start: start, End: end}
	return &span
}

// --- Task 1: schema minting over a real fixture -----------------------

func TestExplainSummarySchemaIsMinted(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "use_after_move.lang")
	diagID := firstDiagnosticID(t, path)

	result, err := ExplainCommandFile(path, diagID, 0)
	if err != nil {
		t.Fatalf("ExplainCommandFile: %v", err)
	}
	if result.Explain == nil {
		t.Fatalf("result.Explain is nil")
	}
	if result.Explain.Schema != protocol.ExplainSchema {
		t.Fatalf("schema = %q, want %q", result.Explain.Schema, protocol.ExplainSchema)
	}
	if result.Explain.RootID == "" {
		t.Fatalf("RootID is empty")
	}
	if len(result.Explain.Nodes) == 0 {
		t.Fatalf("Nodes is empty")
	}
	validAvailability := map[string]bool{
		string(debugmap.Available): true, string(debugmap.OptimizedOut): true, string(debugmap.NotCaptured): true,
	}
	for _, node := range result.Explain.Nodes {
		if !validAvailability[node.Availability] {
			t.Fatalf("node %s has unknown availability %q", node.ID, node.Availability)
		}
	}

	// Two Results differing only in Explain must have different Result.ID
	// (Finalize folds ExplainID into identity).
	finalized := result.Finalize()
	withoutExplain := result
	withoutExplain.Explain = nil
	if finalized.ID == withoutExplain.Finalize().ID {
		t.Fatalf("Explain did not perturb Result.ID: %s", finalized.ID)
	}

	// usageResult's grammar is asserted at the CLI layer (main package);
	// here we just prove the command name round-trips through Finalize.
	if finalized.Command != "explain" {
		t.Fatalf("Command = %q, want explain", finalized.Command)
	}
}

// TestExplainDiagnosticNotFoundIsOperational proves the honest failure path:
// a request for an ID that does not exist in SRC's diagnostics is an
// operational failure carrying explain.diagnostic_not_found, never a panic
// or fabricated graph.
func TestExplainDiagnosticNotFoundIsOperational(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "use_after_move.lang")
	result, err := ExplainCommandFile(path, "diagnostic:does-not-exist", 0)
	if err != nil {
		t.Fatalf("ExplainCommandFile returned Go error: %v", err)
	}
	if result.Status != protocol.StatusOperational {
		t.Fatalf("Status = %q, want %q", result.Status, protocol.StatusOperational)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != ErrExplainDiagnosticNotFound.Code {
		t.Fatalf("Diagnostics = %+v, want one entry with code %s", result.Diagnostics, ErrExplainDiagnosticNotFound.Code)
	}
}

// --- shared fixture-corpus helpers --------------------------------------

// fixtureDiagnosticIDs parses (and, if the parse is clean, checks) path and
// returns every diagnostic ID it produced. It is the single source of truth
// the tests in this file use to discover real diagnostic IDs to drive
// ExplainCommandFile with.
func fixtureDiagnosticIDs(t *testing.T, path string) []string {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	parsed := syntax.Parse(source)
	diagnostics := parsed.Diagnostics
	if len(diagnostics) == 0 {
		checked := check.Program(parsed.Program)
		diagnostics = checked.Diagnostics
	}
	ids := make([]string, 0, len(diagnostics))
	for _, diag := range diagnostics {
		ids = append(ids, diag.ID)
	}
	return ids
}

// firstDiagnosticID is a small convenience wrapper for tests that only need
// one diagnostic ID from a known-rejecting fixture.
func firstDiagnosticID(t *testing.T, path string) string {
	t.Helper()
	ids := fixtureDiagnosticIDs(t, path)
	if len(ids) == 0 {
		t.Fatalf("%s produced no diagnostics", path)
	}
	return ids[0]
}
