package session

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/evidence"
	"github.com/codename-lang/lang/internal/compiler/protocol"
)

// phase6EvidenceFixture builds a real, clean evidence manifest+source pair
// from the pure_match corpus's own toggle.lang, and writes both to files
// under t.TempDir(), returning their paths.
func phase6EvidenceFixture(t *testing.T) (manifestPath, sourcePath string, manifest evidence.Manifest, source []byte) {
	t.Helper()
	source, err := os.ReadFile(nat03CorpusPath("testdata/phase1/toggle.lang"))
	if err != nil {
		t.Fatalf("reading fixture source: %v", err)
	}
	facts, err := evidence.DefaultFacts(context.Background(), "clang")
	if err != nil {
		t.Skipf("env:clang toolchain unavailable, skipping evidence exercise: %v", err)
	}
	product, diagnostics, err := evidence.Build(source, facts)
	if err != nil {
		t.Fatalf("evidence.Build error: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("fixture did not build cleanly: %+v", diagnostics)
	}
	if err := evidence.Validate(product.Manifest, source, facts); err != nil {
		t.Fatalf("captured manifest did not validate against its own source: %v", err)
	}

	dir := t.TempDir()
	manifestPath = filepath.Join(dir, "manifest.json")
	sourcePath = filepath.Join(dir, "toggle.lang")
	if err := os.WriteFile(manifestPath, product.ManifestBytes, 0o644); err != nil {
		t.Fatalf("WriteFile manifest: %v", err)
	}
	if err := os.WriteFile(sourcePath, source, 0o644); err != nil {
		t.Fatalf("WriteFile source: %v", err)
	}
	return manifestPath, sourcePath, product.Manifest, source
}

// phase6WriteStaleManifest writes a copy of manifest with SourceDigest
// deliberately corrupted, at a NEW path, and returns that path.
func phase6WriteStaleManifest(t *testing.T, dir string, manifest evidence.Manifest) string {
	t.Helper()
	stale := manifest
	stale.SourceDigest = "sha256:deliberately-stale-control"
	encoded, err := json.Marshal(stale)
	if err != nil {
		t.Fatalf("marshal stale manifest: %v", err)
	}
	path := filepath.Join(dir, "stale-manifest.json")
	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		t.Fatalf("WriteFile stale manifest: %v", err)
	}
	return path
}

// TestEvidenceValidateIsCompactByDefault (Behavior Test 1): a binding
// manifest, validated without --expand, returns the compact summary with no
// expanded trace.
func TestEvidenceValidateIsCompactByDefault(t *testing.T) {
	manifestPath, sourcePath, _, _ := phase6EvidenceFixture(t)

	result := ValidateEvidenceExpanded(context.Background(), manifestPath, sourcePath, false)
	if result.Status != protocol.StatusPass {
		t.Fatalf("Status = %q, want pass; diagnostics=%v", result.Status, result.Diagnostics)
	}
	if result.Evidence == nil {
		t.Fatal("Evidence summary is nil")
	}
	if result.Evidence.Trace != nil {
		t.Fatalf("Trace = %+v, want nil (compact by default on a passing, non-expand-requested validation)", result.Evidence.Trace)
	}
}

// TestEvidenceExpandsTraceOnFailure (Behavior Test 2): a manifest whose
// bound digest no longer matches returns the compact summary PLUS the
// expanded mismatch trace, without the caller asking.
func TestEvidenceExpandsTraceOnFailure(t *testing.T) {
	_, sourcePath, manifest, _ := phase6EvidenceFixture(t)
	stalePath := phase6WriteStaleManifest(t, filepath.Dir(sourcePath), manifest)

	result := ValidateEvidenceExpanded(context.Background(), stalePath, sourcePath, false)
	if result.Status != protocol.StatusInvalid {
		t.Fatalf("Status = %q, want invalid", result.Status)
	}
	if result.Evidence == nil || result.Evidence.Trace == nil {
		t.Fatal("Trace is nil on a failing validation, want it populated even without --expand")
	}
	found := false
	for _, entry := range result.Evidence.Trace.Entries {
		if entry.Input == "source_digest" {
			found = true
			if entry.Match {
				t.Errorf("source_digest entry reports Match=true, want false (the manifest was deliberately corrupted)")
			}
			if entry.Recorded == entry.Recomputed {
				t.Errorf("source_digest Recorded == Recomputed (%q), want them to differ", entry.Recorded)
			}
		}
	}
	if !found {
		t.Fatal("no source_digest entry in the expanded trace")
	}
}

// TestEvidenceExpandsTraceOnRequest (Behavior Test 3): a binding manifest
// validated WITH --expand returns the expanded trace on explicit request,
// even though it also passes.
func TestEvidenceExpandsTraceOnRequest(t *testing.T) {
	manifestPath, sourcePath, _, _ := phase6EvidenceFixture(t)

	result := ValidateEvidenceExpanded(context.Background(), manifestPath, sourcePath, true)
	if result.Status != protocol.StatusPass {
		t.Fatalf("Status = %q, want pass", result.Status)
	}
	if result.Evidence == nil || result.Evidence.Trace == nil {
		t.Fatal("Trace is nil despite expand=true")
	}
	if len(result.Evidence.Trace.Entries) == 0 {
		t.Fatal("Trace has zero entries")
	}
	for _, entry := range result.Evidence.Trace.Entries {
		if !entry.Match {
			t.Errorf("entry %s reports Match=false on a binding manifest: recorded=%q recomputed=%q", entry.Input, entry.Recorded, entry.Recomputed)
		}
	}
}

// TestExpansionNeverChangesTheVerdict (Behavior Test 5): the same manifest
// validates identically -- same Status, same diagnostic ID list -- whether
// expand is true or false. Expansion is a projection, never a
// re-derivation that could itself disagree.
func TestExpansionNeverChangesTheVerdict(t *testing.T) {
	_, sourcePath, manifest, _ := phase6EvidenceFixture(t)
	stalePath := phase6WriteStaleManifest(t, filepath.Dir(sourcePath), manifest)

	compact := ValidateEvidenceExpanded(context.Background(), stalePath, sourcePath, false)
	expanded := ValidateEvidenceExpanded(context.Background(), stalePath, sourcePath, true)

	if compact.Status != expanded.Status {
		t.Fatalf("Status differs: compact=%q expanded=%q", compact.Status, expanded.Status)
	}
	compactIDs := diagnosticCodes(compact.Diagnostics)
	expandedIDs := diagnosticCodes(expanded.Diagnostics)
	if !reflect.DeepEqual(compactIDs, expandedIDs) {
		t.Fatalf("diagnostic codes differ: compact=%v expanded=%v", compactIDs, expandedIDs)
	}

	// Same exercise on a PASSING manifest, for completeness.
	manifestPath, sourcePath2, _, _ := phase6EvidenceFixture(t)
	compactPass := ValidateEvidenceExpanded(context.Background(), manifestPath, sourcePath2, false)
	expandedPass := ValidateEvidenceExpanded(context.Background(), manifestPath, sourcePath2, true)
	if compactPass.Status != expandedPass.Status {
		t.Fatalf("passing-manifest Status differs: compact=%q expanded=%q", compactPass.Status, expandedPass.Status)
	}
}

func diagnosticCodes(diagnostics []diagnostic.Diagnostic) []string {
	codes := make([]string, 0, len(diagnostics))
	for _, problem := range diagnostics {
		codes = append(codes, problem.Code)
	}
	return codes
}

// TestEvidenceTraceRespectsOutputBound directly exercises
// phase6BoundTraceEntries with enough synthetic entries to exceed
// MaxEvidenceTraceBytes, and asserts it stops adding entries and reports
// the stable truncation code rather than growing unboundedly.
func TestEvidenceTraceRespectsOutputBound(t *testing.T) {
	entries := make([]protocol.TraceEntry, 0, 5000)
	for i := 0; i < 5000; i++ {
		entries = append(entries, protocol.TraceEntry{
			Input:      "synthetic-input-with-a-reasonably-long-name",
			Recorded:   "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Recomputed: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		})
	}

	bounded, truncated := phase6BoundTraceEntries(entries)
	if truncated != TruncatedEvidenceTraceBound {
		t.Fatalf("truncated = %q, want %q", truncated, TruncatedEvidenceTraceBound)
	}
	if len(bounded) >= len(entries) {
		t.Fatalf("bounded has %d entries, want fewer than the %d synthetic entries given", len(bounded), len(entries))
	}
	encoded, err := json.Marshal(bounded)
	if err != nil {
		t.Fatalf("marshal bounded entries: %v", err)
	}
	if len(encoded) > MaxEvidenceTraceBytes+1 {
		t.Fatalf("bounded entries encode to %d bytes, want <= %d", len(encoded), MaxEvidenceTraceBytes+1)
	}

	small := entries[:2]
	boundedSmall, truncatedSmall := phase6BoundTraceEntries(small)
	if truncatedSmall != "" {
		t.Fatalf("truncated = %q on a small entry set, want empty", truncatedSmall)
	}
	if len(boundedSmall) != len(small) {
		t.Fatalf("bounded has %d entries, want all %d", len(boundedSmall), len(small))
	}
}
