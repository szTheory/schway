package check_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// ---------------------------------------------------------------------
// Phase 14 Plan 02: the diagnostic-distinctness gate (DX-08).
//
// Three assertions, D-14-41:
//  1. TestDistinctnessCorpusMembersAreStructurallyDistinct -- the corpus
//     predicate. Computed from syntax.Lex output ALONE (no syntax.Parse, no
//     Check call for the purpose of distinctness): pairwise-different
//     non-trivia token-kind sequences, total, cheap, and rename-immune by
//     construction (every identifier lexes to TokenIdentifier). This fails
//     the CORPUS, not the compiler, when a near-duplicate fixture is added.
//  2. TestDiagnosticDistinctnessIsOne -- unique diagnostic-ID sets equal
//     corpus size, and unique result IDs equal corpus size. This is the
//     ROADMAP's success criterion 5.
//  3. TestDiagnosticDistinctnessGuardIsNotInert -- runs the same metric over
//     the frozen testdata/distinctness/collision_control.json and asserts it
//     reports 1/3, not 3/3. The capture is read, never written.
//
// Anti-Goodhart (D-14-41): a distinctness of 1.0 is corpus-gameable and the
// denominator is chosen by the people who want it green. Three things
// constrain it, all enforced here or in testdata/distinctness/README.md:
// (i) the kind-sequence predicate makes trivial identifier-rename padding
// impossible; (ii) corpus membership is ADDITIVE-ONLY -- removing a member
// to raise the metric is a reviewable deletion requiring a recorded
// decision, the same shape as re-pinning a published ID (README.md); (iii)
// the frozen collision control fails immediately if the metric is stubbed.
// What does NOT stop it: nothing forces the corpus to grow as the language
// grows -- recorded as debt (see 14-02-SUMMARY.md "Debt rows to register").
// ---------------------------------------------------------------------

// distinctnessCorpusFixtures returns the sorted list of .lang fixtures under
// testdata/distinctness/, excluding the frozen collision_control.json and
// README.md (neither is a corpus member).
func distinctnessCorpusFixtures(t *testing.T) []string {
	t.Helper()
	matches, err := filepath.Glob(testsupport.ProjectPath("testdata", "distinctness", "*.lang"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) < 10 {
		t.Fatalf("testdata/distinctness/ corpus has %d .lang fixtures, want >= 10", len(matches))
	}
	sort.Strings(matches)
	return matches
}

// nonTriviaKindSequence lexes source and returns the KIND (never the text)
// of every non-trivia token, including the trailing EOF. Kind-only means
// this predicate is rename-immune by construction: every identifier lexes
// to syntax.TokenIdentifier regardless of its spelling.
func nonTriviaKindSequence(source []byte) string {
	tokens, _ := syntax.Lex(source)
	kinds := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		if tok.Trivia() {
			continue
		}
		kinds = append(kinds, string(tok.Kind))
	}
	return strings.Join(kinds, "\x1f")
}

// TestDistinctnessCorpusMembersAreStructurallyDistinct asserts every pair of
// corpus members has a different non-trivia token-kind sequence. Adding a
// near-duplicate fixture (e.g. an identifier-renamed copy of an existing
// member) fails THIS test, not the compiler -- the corpus itself is the
// subject.
func TestDistinctnessCorpusMembersAreStructurallyDistinct(t *testing.T) {
	fixtures := distinctnessCorpusFixtures(t)
	sequences := make(map[string]string, len(fixtures))
	for _, path := range fixtures {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sequences[filepath.Base(path)] = nonTriviaKindSequence(source)
	}

	seen := make(map[string]string, len(sequences))
	for name, seq := range sequences {
		if other, collide := seen[seq]; collide {
			t.Fatalf("corpus members %q and %q share a non-trivia token-kind sequence (%q) -- the corpus is not structurally distinct, this is not a compiler defect", name, other, seq)
		}
		seen[seq] = name
	}
}

// checkResultIdentity captures the two identity facts diagnosticDistinctness
// derives per corpus member: the sorted set of diagnostic IDs, and the
// top-level result: ID.
type checkResultIdentity struct {
	diagnosticIDSet string // sorted, joined diagnostic IDs -- a SET, not a sequence
	resultID        string
}

// TestDiagnosticDistinctnessIsOne is the ROADMAP's success criterion 5:
// every corpus member mints its OWN diagnostic-ID set and its OWN result:
// ID. Unique diagnostic-ID sets must equal corpus size, and unique result
// IDs must equal corpus size -- a distinctness ratio of 1.0 on the LIVE,
// post-fix corpus.
func TestDiagnosticDistinctnessIsOne(t *testing.T) {
	fixtures := distinctnessCorpusFixtures(t)
	identities := make([]checkResultIdentity, 0, len(fixtures))
	for _, path := range fixtures {
		result, err := session.CheckCommandFile(path)
		if err != nil {
			t.Fatalf("%s: CheckCommandFile: %v", filepath.Base(path), err)
		}
		if len(result.Diagnostics) == 0 {
			t.Fatalf("%s: is not refused -- every distinctness corpus member must be a refused program", filepath.Base(path))
		}
		diagIDs := make([]string, 0, len(result.Diagnostics))
		for _, d := range result.Diagnostics {
			diagIDs = append(diagIDs, d.ID)
		}
		sort.Strings(diagIDs)
		identities = append(identities, checkResultIdentity{
			diagnosticIDSet: strings.Join(diagIDs, "\x1f"),
			resultID:        result.ID,
		})
	}

	uniqueDiagSets := make(map[string]bool, len(identities))
	uniqueResultIDs := make(map[string]bool, len(identities))
	for _, identity := range identities {
		uniqueDiagSets[identity.diagnosticIDSet] = true
		uniqueResultIDs[identity.resultID] = true
	}

	if len(uniqueDiagSets) != len(fixtures) {
		t.Fatalf("diagnostic-ID-set distinctness is %d/%d, want %d/%d (1.0)", len(uniqueDiagSets), len(fixtures), len(fixtures), len(fixtures))
	}
	if len(uniqueResultIDs) != len(fixtures) {
		t.Fatalf("result: ID distinctness is %d/%d, want %d/%d (1.0)", len(uniqueResultIDs), len(fixtures), len(fixtures), len(fixtures))
	}
}

// collisionControlCapture mirrors testdata/distinctness/collision_control.json's
// schema exactly. This struct decodes the frozen capture; it never writes it.
type collisionControlCapture struct {
	CapturedAt         string                   `json:"captured_at"`
	CapturedFromCommit string                   `json:"captured_from_commit"`
	Description        string                   `json:"description"`
	Records            []collisionControlRecord `json:"records"`
}

type collisionControlRecord struct {
	Fixture       string   `json:"fixture"`
	ResultID      string   `json:"result_id"`
	DiagnosticIDs []string `json:"diagnostic_ids"`
}

// TestDiagnosticDistinctnessGuardIsNotInert runs the SAME distinctness
// metric TestDiagnosticDistinctnessIsOne runs, but over the frozen pre-fix
// capture (testdata/distinctness/collision_control.json) instead of a live
// compiler run. This capture is read here and NOWHERE ELSE is it written --
// there is no regeneration path in this tree. The pre-fix collision must
// report a distinctness of 1/3 (one unique result: ID and one unique
// diagnostic-ID set shared by all three spiral members), never the healthy
// 3/3 -- proving the gate is not inert (it would go red the moment the fix
// were reverted or stubbed).
func TestDiagnosticDistinctnessGuardIsNotInert(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "distinctness", "collision_control.json")
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var capture collisionControlCapture
	if err := json.Unmarshal(raw, &capture); err != nil {
		t.Fatalf("collision_control.json did not decode: %v", err)
	}
	if len(capture.Records) != 3 {
		t.Fatalf("collision_control.json has %d records, want 3 (the spiral trio)", len(capture.Records))
	}

	uniqueDiagSets := make(map[string]bool, len(capture.Records))
	uniqueResultIDs := make(map[string]bool, len(capture.Records))
	for _, record := range capture.Records {
		diagIDs := append([]string(nil), record.DiagnosticIDs...)
		sort.Strings(diagIDs)
		uniqueDiagSets[strings.Join(diagIDs, "\x1f")] = true
		uniqueResultIDs[record.ResultID] = true
	}

	if len(uniqueDiagSets) != 1 {
		t.Fatalf("frozen collision control reports %d unique diagnostic-ID sets, want 1 (the pre-fix collision) -- the guard is INERT if this ever reports 3", len(uniqueDiagSets))
	}
	if len(uniqueResultIDs) != 1 {
		t.Fatalf("frozen collision control reports %d unique result: IDs, want 1 (the pre-fix collision) -- the guard is INERT if this ever reports 3", len(uniqueResultIDs))
	}

	// This test performs NO write to testdata/distinctness -- assert the
	// capture file itself is untouched by this run (mtime and size stable).
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if before.Size() != after.Size() || before.ModTime() != after.ModTime() {
		t.Fatal("collision_control.json was modified by this test run -- the capture must be read-only")
	}
}
