package syntax_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase34RecoverySpan struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type phase34RecoveryCause struct {
	Kind string               `json:"kind"`
	Span *phase34RecoverySpan `json:"span"`
}

type phase34RecoveryDiagnostic struct {
	Code        string                 `json:"code"`
	PrimarySpan phase34RecoverySpan    `json:"primary_span"`
	Causes      []phase34RecoveryCause `json:"causes"`
}

type phase34RecoveryDraft struct {
	ID                    string                      `json:"id"`
	Provenance            string                      `json:"provenance"`
	Delta                 string                      `json:"delta"`
	DiagnosticCount       int                         `json:"diagnostic_count"`
	RecoveredSiblingNames []string                    `json:"recovered_sibling_names"`
	Diagnostics           []phase34RecoveryDiagnostic `json:"diagnostics"`
}

type phase34RecoveryObservation struct {
	Diagnostics           []phase34RecoveryDiagnostic
	RecoveredSiblingNames []string
}

type phase35RecoveryAfterFile struct {
	Schema     string                     `json:"schema"`
	Date       string                     `json:"date"`
	Amendments []phase35RecoveryAmendment `json:"amendments"`
}

type phase35RecoveryAmendment struct {
	Fixture           string                    `json:"fixture"`
	Reason            string                    `json:"reason"`
	Phase34Projection phase35RecoveryProjection `json:"phase34_projection"`
	Phase35Projection phase35RecoveryProjection `json:"phase35_projection"`
}

type phase35RecoveryProjection struct {
	DiagnosticCount       int                         `json:"diagnostic_count"`
	RecoveredSiblingNames []string                    `json:"recovered_sibling_names"`
	Diagnostics           []phase34RecoveryDiagnostic `json:"diagnostics"`
}

// These observation-level seams exercise the corpus comparator's specific
// failure classifications. They are local to this test, default off, and
// never alter parser behavior or production output.
var (
	restoreCascadeForTest       bool
	swallowSiblingForTest       bool
	dropIndependentCauseForTest bool
)

func TestPhase34RecoveryCorpus(t *testing.T) {
	corpus, after := loadPhase34RecoveryCorpus(t)
	phase35After := loadPhase35RecoveryAfter(t)
	if len(corpus.Drafts) != len(after.Drafts) {
		t.Fatalf("draft-after rows = %d, want one exact row for each of %d drafts", len(after.Drafts), len(corpus.Drafts))
	}
	wants := make(map[string]phase34RecoveryDraft, len(after.Drafts))
	for _, want := range after.Drafts {
		if want.Provenance != "reconstructed-2026-10-13" {
			t.Errorf("%s provenance = %q, want reconstructed-2026-10-13", want.ID, want.Provenance)
		}
		if strings.TrimSpace(want.Delta) == "" {
			t.Errorf("%s has no intentional-delta explanation", want.ID)
		}
		if _, duplicate := wants[want.ID]; duplicate {
			t.Errorf("draft-after.json repeats %q", want.ID)
		}
		wants[want.ID] = want
	}

	for _, draft := range corpus.Drafts {
		want, ok := wants[draft.ID]
		if !ok {
			t.Errorf("draft-after.json has no row for %q", draft.ID)
			continue
		}
		parsed := syntax.Parse([]byte(draft.Source))
		if !bytes.Equal(parsed.Tree.Bytes(), []byte(draft.Source)) {
			t.Errorf("%s recovery changed the lossless source bytes", draft.ID)
		}
		got := observePhase34Recovery(parsed)
		var reason string
		if draft.ID == "crc32-reconstructed" {
			reason = comparePhase35RecoveryAfter(phase35After, want, got)
		} else {
			reason = comparePhase34Recovery(want, got)
		}
		if reason != "" {
			t.Errorf("%s recovery mismatch (%s): diagnostics=%+v siblings=%v", draft.ID, reason, got.Diagnostics, got.RecoveredSiblingNames)
		}
	}

	malformed := fixtureByID(t, corpus.Drafts, "malformed-nested-paren")
	malformedParse := syntax.Parse([]byte(malformed.Source))
	checked := session.Check([]byte(malformed.Source))
	if len(checked.Diagnostics) == 0 {
		t.Fatal("session.Check accepted malformed nested syntax")
	}
	if !reflect.DeepEqual(checked.Program, core.Program{}) {
		t.Fatalf("session.Check admitted rejected syntax to core: %+v", checked.Program)
	}
	if !bytes.Equal(checked.Tree.Bytes(), malformedParse.Tree.Bytes()) {
		t.Fatal("session.Check changed the lossless rejected source bytes")
	}

	missingCloser := fixtureByID(t, corpus.Drafts, "independent-missing-closer")
	missingParse := syntax.Parse([]byte(missingCloser.Source))
	if len(missingParse.Diagnostics) != 2 ||
		missingParse.Diagnostics[0].Code != "syntax.expected_rparen" || missingParse.Diagnostics[0].Primary.Start != 63 || missingParse.Diagnostics[0].Primary.End != 65 ||
		missingParse.Diagnostics[1].Code != "syntax.expected_rbrace" || missingParse.Diagnostics[1].Primary.Start != 74 || missingParse.Diagnostics[1].Primary.End != 74 {
		t.Fatalf("independent missing closers were not preserved: %+v", missingParse.Diagnostics)
	}
}

func phase34AssertFrozenRecoveryInputs(t *testing.T) {
	t.Helper()
	want := map[string]string{
		"testdata/phase34/drafts.json":         "8774033822c9f47db8236d00a347e7e176a9524437314c182be8555f16a1b9df",
		"testdata/phase34/draft-baseline.json": "6607c9b5e3657c0e65ba6e46d16a1dc281153994f27ca1ab8c645f9e935b8fae",
		"testdata/phase34/draft-after.json":    "1e11ff4a8af6cc03c5d6e3200d0cdf9ef7dad41ee28e6a6fcd3aa2362b272bf3",
		"testdata/phase34/compat-deltas.json":  "356d6f087023b46542636bfa244266afe1fbcd73cdb89120cc3d33802f558011",
	}
	for path, digestWant := range want {
		contents, err := os.ReadFile(testsupport.ProjectPath(path))
		if err != nil {
			t.Fatalf("read frozen Phase 34 recovery input %s: %v", path, err)
		}
		digest := sha256.Sum256(contents)
		if got := hex.EncodeToString(digest[:]); got != digestWant {
			t.Errorf("frozen Phase 34 recovery input %s changed: sha256 %s, want %s", path, got, digestWant)
		}
	}
}

func loadPhase35RecoveryAfter(t *testing.T) phase35RecoveryAfterFile {
	t.Helper()
	var after phase35RecoveryAfterFile
	readPhase34JSON(t, "testdata/phase35/recovery-after.json", &after)
	return after
}

func fixtureAfterByID(t *testing.T, drafts []phase34RecoveryDraft, id string) phase34RecoveryDraft {
	t.Helper()
	for _, draft := range drafts {
		if draft.ID == id {
			return draft
		}
	}
	t.Fatalf("draft-after fixture %q is missing", id)
	return phase34RecoveryDraft{}
}

func comparePhase35RecoveryAfter(file phase35RecoveryAfterFile, historical phase34RecoveryDraft, got phase34RecoveryObservation) string {
	if file.Schema != "phase35.recovery-after/1" || file.Date != "2026-10-14" {
		return "header"
	}
	if len(file.Amendments) != 1 {
		return "amendment_count"
	}
	amendment := file.Amendments[0]
	if amendment.Fixture != "crc32-reconstructed" {
		return "wrong_fixture"
	}
	if strings.TrimSpace(amendment.Reason) == "" {
		return "missing_reason"
	}
	wantHistorical := phase35RecoveryProjection{
		DiagnosticCount:       historical.DiagnosticCount,
		RecoveredSiblingNames: historical.RecoveredSiblingNames,
		Diagnostics:           historical.Diagnostics,
	}
	if !reflect.DeepEqual(amendment.Phase34Projection, wantHistorical) {
		return "wrong_phase34_projection"
	}
	actual := phase35RecoveryProjection{
		DiagnosticCount:       len(got.Diagnostics),
		RecoveredSiblingNames: got.RecoveredSiblingNames,
		Diagnostics:           got.Diagnostics,
	}
	if !reflect.DeepEqual(amendment.Phase35Projection, actual) {
		return "wrong_phase35_projection"
	}
	return ""
}

func clonePhase35Projection(projection phase35RecoveryProjection) phase35RecoveryProjection {
	clone := projection
	clone.RecoveredSiblingNames = append([]string(nil), projection.RecoveredSiblingNames...)
	clone.Diagnostics = append([]phase34RecoveryDiagnostic(nil), projection.Diagnostics...)
	for index := range clone.Diagnostics {
		clone.Diagnostics[index].Causes = append([]phase34RecoveryCause(nil), projection.Diagnostics[index].Causes...)
	}
	return clone
}

func TestPhase34RecoveryControls(t *testing.T) {
	corpus, after := loadPhase34RecoveryCorpus(t)
	t.Cleanup(func() {
		restoreCascadeForTest = false
		swallowSiblingForTest = false
		dropIndependentCauseForTest = false
	})
	if restoreCascadeForTest || swallowSiblingForTest || dropIndependentCauseForTest {
		t.Fatal("recovery test controls must be disabled by default")
	}
	wants := make(map[string]phase34RecoveryDraft, len(after.Drafts))
	for _, want := range after.Drafts {
		wants[want.ID] = want
	}
	fixtures := make(map[string]phase34Source, len(corpus.Drafts))
	for _, draft := range corpus.Drafts {
		fixtures[draft.ID] = draft
	}

	checkControl := func(id string) string {
		draft := fixtures[id]
		parsed := syntax.Parse([]byte(draft.Source))
		got := applyPhase34RecoveryControlForTest(id, observePhase34Recovery(parsed))
		return comparePhase34Recovery(wants[id], got)
	}
	if reason := checkControl("malformed-nested-paren"); reason != "" {
		t.Fatalf("normal nested-paren observation failed before controls: %s", reason)
	}

	restoreCascadeForTest = true
	if reason := checkControl("malformed-nested-paren"); reason != "cascade_count" {
		t.Fatalf("restoreCascadeForTest reached failure reason %q, want cascade_count", reason)
	}
	t.Log("restoreCascadeForTest reached cascade_count")
	restoreCascadeForTest = false
	if reason := checkControl("malformed-nested-paren"); reason != "" {
		t.Fatalf("restoreCascadeForTest did not restore normal output: %s", reason)
	}
	t.Log("restoreCascadeForTest restored normal output")

	swallowSiblingForTest = true
	if reason := checkControl("sibling-after-error"); reason != "missing_sibling" {
		t.Fatalf("swallowSiblingForTest reached failure reason %q, want missing_sibling", reason)
	}
	t.Log("swallowSiblingForTest reached missing_sibling")
	swallowSiblingForTest = false
	if reason := checkControl("sibling-after-error"); reason != "" {
		t.Fatalf("swallowSiblingForTest did not restore normal output: %s", reason)
	}
	t.Log("swallowSiblingForTest restored normal output")

	dropIndependentCauseForTest = true
	if reason := checkControl("two-independent-causes"); reason != "missing_independent_cause" {
		t.Fatalf("dropIndependentCauseForTest reached failure reason %q, want missing_independent_cause", reason)
	}
	t.Log("dropIndependentCauseForTest reached missing_independent_cause")
	dropIndependentCauseForTest = false
	if reason := checkControl("two-independent-causes"); reason != "" {
		t.Fatalf("dropIndependentCauseForTest did not restore normal output: %s", reason)
	}
	t.Log("dropIndependentCauseForTest restored normal output")

	t.Run("bounded_nested_delimiters", func(t *testing.T) {
		const depth = 4096
		nested := strings.Repeat("(", depth) + "broken" + strings.Repeat(")", depth)
		source := []byte("module phase34.deep\nexport { fn broken fn kept }\n" +
			"fn broken(n: U64) -> U64 { let value = call(n, " + nested + ") n }\n" +
			"fn kept(n: U64) -> U64 { n }\n")
		if len(source) > syntax.MaxSourceBytes {
			t.Fatalf("generated source length %d exceeds MaxSourceBytes %d", len(source), syntax.MaxSourceBytes)
		}
		parsed := syntax.Parse(source)
		if !bytes.Equal(parsed.Tree.Bytes(), source) {
			t.Fatal("deep recovery changed lossless source bytes")
		}
		if len(parsed.Tree.Tokens) >= syntax.MaxTokens {
			t.Fatalf("generated token stream reached MaxTokens: %d", len(parsed.Tree.Tokens))
		}
		if len(parsed.Diagnostics) != 1 || parsed.Diagnostics[0].Code != "syntax.expected_call_argument" {
			t.Fatalf("deep nested recovery diagnostics = %+v, want one owned call-argument cause", parsed.Diagnostics)
		}
		var names []string
		for _, function := range parsed.Program.Funcs {
			names = append(names, function.Name)
		}
		if !reflect.DeepEqual(names, []string{"broken", "kept"}) {
			t.Fatalf("deep recovery siblings = %v, want [broken kept]", names)
		}
	})

	t.Run("diagnostic_cap_retains_truncation_marker", func(t *testing.T) {
		var source strings.Builder
		source.WriteString("module phase34.cap\nexport { fn f00 }\n")
		for index := 0; index < 24; index++ {
			fmt.Fprintf(&source, "@\nfn f%02d(n: U64) -> U64 { n }\n", index)
		}
		parsed := syntax.Parse([]byte(source.String()))
		if len(parsed.Diagnostics) != 21 || parsed.Diagnostics[len(parsed.Diagnostics)-1].Code != "syntax.too_many_errors" {
			t.Fatalf("bounded error stream diagnostics = %d, final=%+v; want 20 primaries plus syntax.too_many_errors", len(parsed.Diagnostics), parsed.Diagnostics)
		}
	})
}

type phase34Source struct {
	ID     string `json:"id"`
	Source string `json:"source"`
}

func loadPhase34RecoveryCorpus(t *testing.T) (struct {
	Drafts []phase34Source `json:"drafts"`
}, struct {
	Drafts []phase34RecoveryDraft `json:"drafts"`
}) {
	t.Helper()
	var corpus struct {
		Drafts []phase34Source `json:"drafts"`
	}
	readPhase34JSON(t, "testdata/phase34/drafts.json", &corpus)
	var after struct {
		Drafts []phase34RecoveryDraft `json:"drafts"`
	}
	readPhase34JSON(t, "testdata/phase34/draft-after.json", &after)
	return corpus, after
}

func readPhase34JSON(t *testing.T, path string, target any) {
	t.Helper()
	contents, err := os.ReadFile(testsupport.ProjectPath(path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(contents, target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}

func fixtureByID(t *testing.T, drafts []phase34Source, id string) phase34Source {
	t.Helper()
	for _, draft := range drafts {
		if draft.ID == id {
			return draft
		}
	}
	t.Fatalf("draft fixture %q is missing", id)
	return phase34Source{}
}

func observePhase34Recovery(parsed syntax.ParseResult) phase34RecoveryObservation {
	got := phase34RecoveryObservation{
		Diagnostics:           make([]phase34RecoveryDiagnostic, 0, len(parsed.Diagnostics)),
		RecoveredSiblingNames: make([]string, 0, len(parsed.Program.Funcs)),
	}
	for _, problem := range parsed.Diagnostics {
		row := phase34RecoveryDiagnostic{
			Code:        problem.Code,
			PrimarySpan: phase34RecoverySpan{Start: problem.Primary.Start, End: problem.Primary.End},
			Causes:      make([]phase34RecoveryCause, 0, len(problem.Causes)),
		}
		for _, cause := range problem.Causes {
			entry := phase34RecoveryCause{Kind: cause.Kind}
			if cause.Span != nil {
				entry.Span = &phase34RecoverySpan{Start: cause.Span.Start, End: cause.Span.End}
			}
			row.Causes = append(row.Causes, entry)
		}
		got.Diagnostics = append(got.Diagnostics, row)
	}
	for _, function := range parsed.Program.Funcs {
		got.RecoveredSiblingNames = append(got.RecoveredSiblingNames, function.Name)
	}
	return got
}

func comparePhase34Recovery(want phase34RecoveryDraft, got phase34RecoveryObservation) string {
	if len(got.Diagnostics) > want.DiagnosticCount {
		return "cascade_count"
	}
	if len(got.Diagnostics) < want.DiagnosticCount {
		if want.ID == "two-independent-causes" {
			return "missing_independent_cause"
		}
		return "missing_cause"
	}
	for _, sibling := range want.RecoveredSiblingNames {
		if !containsString(got.RecoveredSiblingNames, sibling) {
			return "missing_sibling"
		}
	}
	if !reflect.DeepEqual(want.RecoveredSiblingNames, got.RecoveredSiblingNames) {
		return "unexpected_sibling"
	}
	if len(want.Diagnostics) != want.DiagnosticCount || !reflect.DeepEqual(want.Diagnostics, got.Diagnostics) {
		return "diagnostic_mismatch"
	}
	return ""
}

func applyPhase34RecoveryControlForTest(id string, got phase34RecoveryObservation) phase34RecoveryObservation {
	switch {
	case restoreCascadeForTest && id == "malformed-nested-paren" && len(got.Diagnostics) > 0:
		got.Diagnostics = append(got.Diagnostics, got.Diagnostics[0])
	case swallowSiblingForTest && id == "sibling-after-error" && len(got.RecoveredSiblingNames) > 0:
		got.RecoveredSiblingNames = got.RecoveredSiblingNames[:len(got.RecoveredSiblingNames)-1]
	case dropIndependentCauseForTest && id == "two-independent-causes" && len(got.Diagnostics) > 0:
		got.Diagnostics = got.Diagnostics[:len(got.Diagnostics)-1]
	}
	return got
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
