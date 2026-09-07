package session

import (
	"context"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/debugmap"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// This file is deliberately `package session` (internal), matching
// session_phase6_explain_test.go's own established precedent: Task 3's
// pagination/sort tests need to construct exact synthetic QueryFact shapes
// (many facts sharing sort keys, engineered tie cases) that no real fixture
// happens to produce, and the sort/pagination helpers are intentionally
// unexported. Schema minting (Task 1) and the every-real-vocabulary sweep
// (Task 2) still drive QueryCommandFile end-to-end against real fixtures.

// --- Task 1: schema minting + one vocabulary end-to-end -------------------

func TestQuerySummarySchemaIsMinted(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
	operationID := firstDebugMapOperationID(t, path)

	result, err := QueryCommandFile(path, operationID, QueryOptions{})
	if err != nil {
		t.Fatalf("QueryCommandFile: %v", err)
	}
	if result.Query == nil {
		t.Fatalf("result.Query is nil")
	}
	if result.Query.Schema != protocol.QuerySchema {
		t.Fatalf("schema = %q, want %q", result.Query.Schema, protocol.QuerySchema)
	}
	if len(result.Query.Facts) != 1 {
		t.Fatalf("Facts = %+v, want exactly one resolved fact", result.Query.Facts)
	}
	if result.Query.Facts[0].Availability != string(debugmap.Available) {
		t.Fatalf("Facts[0].Availability = %q, want available", result.Query.Facts[0].Availability)
	}
	if result.Query.Facts[0].Vocabulary != "operation_id" {
		t.Fatalf("Facts[0].Vocabulary = %q, want operation_id", result.Query.Facts[0].Vocabulary)
	}

	// Two Results differing only in Query must have different Result.ID
	// (Finalize folds QueryID into identity).
	finalized := result.Finalize()
	withoutQuery := result
	withoutQuery.Query = nil
	if finalized.ID == withoutQuery.Finalize().ID {
		t.Fatalf("Query did not perturb Result.ID: %s", finalized.ID)
	}
	if finalized.Command != "query" {
		t.Fatalf("Command = %q, want query", finalized.Command)
	}
}

// TestQueryUnknownIDReportsNotCaptured proves the honest-absence contract
// (FND-04): an operation ID the map never produced is one not_captured
// fact, never an error and never a fabricated span.
func TestQueryUnknownIDReportsNotCaptured(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
	result, err := QueryCommandFile(path, "s1:owned.transfer:fn:relay:op:999999", QueryOptions{})
	if err != nil {
		t.Fatalf("QueryCommandFile: %v", err)
	}
	if result.Status != protocol.StatusPass {
		t.Fatalf("Status = %q, want pass (absence is honest data, not an error)", result.Status)
	}
	if len(result.Query.Facts) != 1 {
		t.Fatalf("Facts = %+v, want exactly one fact", result.Query.Facts)
	}
	fact := result.Query.Facts[0]
	if fact.Availability != string(debugmap.NotCaptured) {
		t.Fatalf("Availability = %q, want not_captured", fact.Availability)
	}
	if fact.Vocabulary != "operation_id" {
		t.Fatalf("Vocabulary = %q, want operation_id (recognized shape, absent value)", fact.Vocabulary)
	}
	if fact.Span != nil {
		t.Fatalf("Span = %+v, want nil for an absent fact (never a fabricated span)", fact.Span)
	}

	// A genuinely unrecognized address (no vocabulary shape at all) reports
	// the "unrecognized" sentinel, never an error.
	unrecognized, err := QueryCommandFile(path, "totally-unrecognized-address", QueryOptions{})
	if err != nil {
		t.Fatalf("QueryCommandFile (unrecognized): %v", err)
	}
	if unrecognized.Status != protocol.StatusPass {
		t.Fatalf("unrecognized Status = %q, want pass", unrecognized.Status)
	}
	if len(unrecognized.Query.Facts) != 1 || unrecognized.Query.Facts[0].Vocabulary != QueryVocabularyUnrecognized {
		t.Fatalf("unrecognized Facts = %+v, want one fact with vocabulary=unrecognized", unrecognized.Query.Facts)
	}
	if unrecognized.Query.Facts[0].Availability != string(debugmap.NotCaptured) {
		t.Fatalf("unrecognized Availability = %q, want not_captured", unrecognized.Query.Facts[0].Availability)
	}
}

// --- shared test helpers ----------------------------------------------------

// firstDebugMapOperationID returns a real operation ID debug-map reports
// for path, following runDebugMap's own construction exactly.
func firstDebugMapOperationID(t *testing.T, path string) string {
	t.Helper()
	built := buildDebugMap(t, path)
	if len(built.Entries) == 0 {
		t.Fatalf("debugmap.Build produced no entries for %s", path)
	}
	return built.Entries[0].OperationID
}

func buildDebugMap(t *testing.T, path string) debugmap.Map {
	t.Helper()
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("%s failed to parse: %+v", path, parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) > 0 {
		t.Fatalf("%s failed to check: %+v", path, checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("%s failed core validation: %+v", path, validated.Problems)
	}
	built, _, err := debugmap.Build(context.Background(), parsed.Program, validated.Program())
	if err != nil {
		t.Fatalf("debugmap.Build(%s): %v", path, err)
	}
	return built
}
