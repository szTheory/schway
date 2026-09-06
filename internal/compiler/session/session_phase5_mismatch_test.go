package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/reduce"
	"github.com/codename-lang/lang/internal/compiler/session"
)

// TestMismatchDocumentEmittedOnSeededDivergence drives plan 05-07's
// AliasFactMutationRunner to a REAL divergence over
// testdata/phase5/false_restrict_hoist.lang and asserts a lang.mismatch/0
// document is emitted with a non-empty reduced_source, a diverging_axis
// matching the comparator's own verdict, and a minimality of fixpoint or
// budget_exhausted (D-05-26).
func TestMismatchDocumentEmittedOnSeededDivergence(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	document, err := session.ReduceSeededAliasMismatch(ctx)
	if err != nil {
		t.Fatalf("ReduceSeededAliasMismatch: %v", err)
	}
	if document.Schema != reduce.MismatchSchema {
		t.Fatalf("schema = %q, want %q", document.Schema, reduce.MismatchSchema)
	}
	if document.DivergingAxis != session.AxisTerminalOutcome {
		t.Fatalf("diverging_axis = %q, want the comparator's own verdict %q", document.DivergingAxis, session.AxisTerminalOutcome)
	}
	if document.ReducedSource == "" {
		t.Fatal("expected a non-empty reduced_source")
	}
	if document.Minimality != reduce.MinimalityFixpoint && document.Minimality != reduce.MinimalityBudgetExhausted {
		t.Fatalf("minimality = %q, want %q or %q", document.Minimality, reduce.MinimalityFixpoint, reduce.MinimalityBudgetExhausted)
	}
	if document.EnginePair == "" {
		t.Fatal("expected a non-empty engine_pair")
	}
	if document.EvidenceID == "" {
		t.Fatal("expected a non-empty evidence_id (the additive discretion field)")
	}
}

// TestMismatchDocumentIsSelfSufficient asserts the emitted document's
// causal_chain and event_window are both non-empty for the seeded
// divergence -- the repair-agent requirement that a fix be proposable from
// this document alone, without opening the full execution log (D-05-26).
func TestMismatchDocumentIsSelfSufficient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	document, err := session.ReduceSeededAliasMismatch(ctx)
	if err != nil {
		t.Fatalf("ReduceSeededAliasMismatch: %v", err)
	}
	if len(document.CausalChain) == 0 {
		t.Fatal("expected a non-empty causal_chain")
	}
	if len(document.EventWindow) == 0 {
		t.Fatal("expected a non-empty event_window")
	}
	for engine, events := range document.EventWindow {
		if len(events) == 0 {
			t.Fatalf("engine %q has an empty event window", engine)
		}
		if len(events) > reduce.EventWindowSize {
			t.Fatalf("engine %q has %d events, want at most %d (bounded)", engine, len(events), reduce.EventWindowSize)
		}
	}
	if len(document.Causes) == 0 {
		t.Fatal("expected at least one cause naming what diverged")
	}
}
