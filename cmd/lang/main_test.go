package main

import (
	"os"
	"testing"
)

// TestPhase6CorpusDispatchRequiresMarker asserts isPhase6Corpus recognizes
// testdata/phase6 by its own characteristic marker fixture
// (heldout_match_defect.lang, per testdata/phase6/README) and returns
// false for testdata/phase5 and for an empty temp directory -- a directory
// with no Phase 6 marker must never be silently treated as a Phase 6
// corpus (FND-04's empty-input edge, plan 06-15's acceptance criteria).
func TestPhase6CorpusDispatchRequiresMarker(t *testing.T) {
	if !isPhase6Corpus("../../testdata/phase6") {
		t.Fatal("isPhase6Corpus(testdata/phase6) = false, want true")
	}
	if isPhase6Corpus("../../testdata/phase5") {
		t.Fatal("isPhase6Corpus(testdata/phase5) = true, want false")
	}
	empty := t.TempDir()
	if isPhase6Corpus(empty) {
		t.Fatal("isPhase6Corpus(empty directory) = true, want false")
	}
	if isPhase5Corpus(empty) {
		t.Fatal("isPhase5Corpus(empty directory) = true, want false")
	}
	if _, err := os.Stat("../../testdata/phase6/heldout_match_defect.lang"); err != nil {
		t.Fatalf("testdata/phase6's own marker fixture is missing: %v", err)
	}
}

// TestPhase7CorpusDispatchRequiresMarker asserts isPhase7Corpus recognizes
// testdata/phase07 by its own characteristic marker fixture
// (call_basic.lang, the Phase 07 tracer fixture) and returns false for
// testdata/phase6 and for an empty temp directory -- a directory with no
// Phase 07 marker must never be silently treated as a Phase 07 corpus.
func TestPhase7CorpusDispatchRequiresMarker(t *testing.T) {
	if !isPhase7Corpus("../../testdata/phase07") {
		t.Fatal("isPhase7Corpus(testdata/phase07) = false, want true")
	}
	if isPhase7Corpus("../../testdata/phase6") {
		t.Fatal("isPhase7Corpus(testdata/phase6) = true, want false")
	}
	empty := t.TempDir()
	if isPhase7Corpus(empty) {
		t.Fatal("isPhase7Corpus(empty directory) = true, want false")
	}
	if _, err := os.Stat("../../testdata/phase07/call_basic.lang"); err != nil {
		t.Fatalf("testdata/phase07's own marker fixture is missing: %v", err)
	}
}
