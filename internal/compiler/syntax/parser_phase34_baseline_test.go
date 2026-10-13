package syntax_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const (
	phase34DraftSchema         = "phase34.reconstructed-drafts/1"
	phase34DraftBaselineID     = "phase34.draft-baseline/1"
	phase34DraftProvenance     = "reconstructed-2026-10-13"
	phase34DraftBaselineSHA256 = "6607c9b5e3657c0e65ba6e46d16a1dc281153994f27ca1ab8c645f9e935b8fae"
)

var phase34DraftIDs = []string{
	"logical-and", "logical-or", "logical-not", "schway-arity-two", "foreign-arity-two",
	"if-no-else", "bool-tail-unbound", "subtraction", "malformed-nested-paren", "malformed-nested-block",
	"sibling-after-error", "two-independent-causes", "independent-missing-closer", "eof-empty", "eof-unclosed",
	"comments-strings-arrows", "admitted-ownership", "malformed-near-miss", "crc32-reconstructed", "word-counter-reconstructed",
}

type phase34DraftCorpus struct {
	Schema string         `json:"schema"`
	Drafts []phase34Draft `json:"drafts"`
}

type phase34Draft struct {
	ID                   string   `json:"id"`
	Provenance           string   `json:"provenance"`
	CauseRelation        string   `json:"cause_relation"`
	CauseGroup           string   `json:"cause_group"`
	ExpectedSiblingNames []string `json:"expected_sibling_names"`
	Source               string   `json:"source"`
}

type phase34DraftBaseline struct {
	Schema   string                      `json:"schema"`
	Revision string                      `json:"revision"`
	Drafts   []phase34DiagnosticBaseline `json:"drafts"`
}

type phase34DiagnosticBaseline struct {
	ID                    string                   `json:"id"`
	Provenance            string                   `json:"provenance"`
	Source                string                   `json:"source"`
	SourceBytes           int                      `json:"source_bytes"`
	TreeBytesEqual        bool                     `json:"tree_bytes_equal"`
	DiagnosticCount       int                      `json:"diagnostic_count"`
	Diagnostics           []phase34BaselineProblem `json:"diagnostics"`
	RecoveredSiblingNames []string                 `json:"recovered_sibling_names"`
}

type phase34BaselineProblem struct {
	ID      string                 `json:"id"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Primary diagnostic.Span        `json:"primary_span"`
	Causes  []phase34BaselineCause `json:"causes"`
}

type phase34BaselineCause struct {
	Kind   string           `json:"kind"`
	Detail string           `json:"detail,omitempty"`
	Span   *diagnostic.Span `json:"span,omitempty"`
	Unit   string           `json:"unit,omitempty"`
}

func phase34DraftsPath() string {
	return testsupport.ProjectPath("testdata", "phase34", "drafts.json")
}

func phase34DraftBaselinePath() string {
	return testsupport.ProjectPath("testdata", "phase34", "draft-baseline.json")
}

func phase34LoadDrafts(t *testing.T) phase34DraftCorpus {
	t.Helper()
	data, err := os.ReadFile(phase34DraftsPath())
	if err != nil {
		t.Fatal(err)
	}
	var corpus phase34DraftCorpus
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatalf("decode reconstructed drafts: %v", err)
	}
	return corpus
}

func phase34ObserveDrafts(t *testing.T, corpus phase34DraftCorpus) []phase34DiagnosticBaseline {
	t.Helper()
	rows := make([]phase34DiagnosticBaseline, 0, len(corpus.Drafts))
	for _, draft := range corpus.Drafts {
		source := []byte(draft.Source)
		parsed := syntax.Parse(source)
		row := phase34DiagnosticBaseline{
			ID: draft.ID, Provenance: draft.Provenance, Source: draft.Source,
			SourceBytes: len(source), TreeBytesEqual: bytes.Equal(parsed.Tree.Bytes(), source),
			DiagnosticCount: len(parsed.Diagnostics), RecoveredSiblingNames: phase34RecoveredSiblingNames(parsed.Program.Funcs, parsed.Diagnostics),
			Diagnostics: make([]phase34BaselineProblem, 0, len(parsed.Diagnostics)),
		}
		for _, problem := range parsed.Diagnostics {
			causes := make([]phase34BaselineCause, 0, len(problem.Causes))
			for _, cause := range problem.Causes {
				causes = append(causes, phase34BaselineCause{Kind: cause.Kind, Detail: cause.Detail, Span: cause.Span, Unit: cause.Unit})
			}
			row.Diagnostics = append(row.Diagnostics, phase34BaselineProblem{
				ID: problem.ID, Code: problem.Code, Message: problem.Message, Primary: problem.Primary, Causes: causes,
			})
		}
		rows = append(rows, row)
	}
	return rows
}

func phase34RecoveredSiblingNames(functions []ast.FuncDecl, diagnostics []diagnostic.Diagnostic) []string {
	names := make([]string, 0, len(functions))
	if len(diagnostics) == 0 {
		return names
	}
	firstDiagnostic := diagnostics[0].Primary.Start
	for _, function := range functions {
		if function.Span.Start > firstDiagnostic {
			names = append(names, function.Name)
		}
	}
	return names
}

func phase34LoadDraftBaseline(t *testing.T) phase34DraftBaseline {
	t.Helper()
	data, err := os.ReadFile(phase34DraftBaselinePath())
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	if got := hex.EncodeToString(digest[:]); got != phase34DraftBaselineSHA256 {
		t.Fatalf("immutable draft baseline changed: sha256 %s, want %s", got, phase34DraftBaselineSHA256)
	}
	var baseline phase34DraftBaseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatalf("decode draft baseline: %v", err)
	}
	return baseline
}

func TestPhase34DraftBaseline(t *testing.T) {
	corpus := phase34LoadDrafts(t)
	if corpus.Schema != phase34DraftSchema {
		t.Fatalf("draft schema %q, want %q", corpus.Schema, phase34DraftSchema)
	}
	if len(corpus.Drafts) != len(phase34DraftIDs) {
		t.Fatalf("draft corpus has %d fixtures, want %d", len(corpus.Drafts), len(phase34DraftIDs))
	}
	for index, draft := range corpus.Drafts {
		if draft.ID != phase34DraftIDs[index] {
			t.Fatalf("draft row %d is %q, want %q", index, draft.ID, phase34DraftIDs[index])
		}
		if draft.Provenance != phase34DraftProvenance || draft.CauseRelation == "" || draft.CauseGroup == "" {
			t.Fatalf("draft %q lacks explicit reconstructed provenance or cause metadata: %+v", draft.ID, draft)
		}
		parsed := syntax.Parse([]byte(draft.Source))
		if !bytes.Equal(parsed.Tree.Bytes(), []byte(draft.Source)) {
			t.Fatalf("draft %q did not preserve exact source bytes", draft.ID)
		}
	}

	baseline := phase34LoadDraftBaseline(t)
	if baseline.Schema != phase34DraftBaselineID || baseline.Revision == "" {
		t.Fatalf("invalid draft baseline header: schema=%q revision=%q", baseline.Schema, baseline.Revision)
	}
	if len(baseline.Drafts) != len(corpus.Drafts) {
		t.Fatalf("baseline has %d rows, want %d", len(baseline.Drafts), len(corpus.Drafts))
	}
	for index, draft := range corpus.Drafts {
		row := baseline.Drafts[index]
		if row.ID != draft.ID || row.Provenance != phase34DraftProvenance || row.Source != draft.Source {
			t.Fatalf("historical baseline row %d does not match immutable fixture %q", index, draft.ID)
		}
		if row.SourceBytes != len([]byte(draft.Source)) || !row.TreeBytesEqual {
			t.Fatalf("baseline %q does not preserve exact source bytes", draft.ID)
		}
		if row.DiagnosticCount != len(row.Diagnostics) {
			t.Fatalf("baseline %q count=%d but records=%d", draft.ID, row.DiagnosticCount, len(row.Diagnostics))
		}
		if !reflect.DeepEqual(row.RecoveredSiblingNames, draft.ExpectedSiblingNames) {
			t.Fatalf("baseline %q recovered siblings %v, want authored expectation %v", draft.ID, row.RecoveredSiblingNames, draft.ExpectedSiblingNames)
		}
		for diagnosticIndex, problem := range row.Diagnostics {
			if problem.ID == "" || problem.Code == "" {
				t.Fatalf("baseline %q diagnostic %d has no literal ID/code", draft.ID, diagnosticIndex)
			}
			if problem.Primary.Start < 0 || problem.Primary.End < problem.Primary.Start || problem.Primary.End > row.SourceBytes {
				t.Fatalf("baseline %q diagnostic %d has invalid byte span %+v", draft.ID, diagnosticIndex, problem.Primary)
			}
			for causeIndex, cause := range problem.Causes {
				if cause.Kind == "" {
					t.Fatalf("baseline %q diagnostic %d cause %d has no kind", draft.ID, diagnosticIndex, causeIndex)
				}
				if cause.Span != nil && (cause.Span.Start < 0 || cause.Span.End < cause.Span.Start || cause.Span.End > row.SourceBytes) {
					t.Fatalf("baseline %q diagnostic %d cause %d has invalid span %+v", draft.ID, diagnosticIndex, causeIndex, cause.Span)
				}
			}
		}
	}
	phase34AssertCauseRelationships(t, corpus, baseline)
}

func phase34AssertCauseRelationships(t *testing.T, corpus phase34DraftCorpus, baseline phase34DraftBaseline) {
	t.Helper()
	byID := make(map[string]phase34DiagnosticBaseline, len(baseline.Drafts))
	for _, row := range baseline.Drafts {
		byID[row.ID] = row
	}
	shared := 0
	for _, problem := range byID["sibling-after-error"].Diagnostics {
		for _, cause := range problem.Causes {
			if cause.Kind == "skipped_region" && cause.Span != nil {
				shared++
			}
		}
	}
	if shared == 0 {
		t.Fatal("sibling-after-error baseline lacks its real skipped-region recovery cause")
	}
	var independent []diagnostic.Span
	for _, problem := range byID["two-independent-causes"].Diagnostics {
		for _, cause := range problem.Causes {
			if cause.Kind == "skipped_region" && cause.Span != nil {
				independent = append(independent, *cause.Span)
			}
		}
	}
	if len(independent) != 2 {
		t.Fatalf("two-independent-causes baseline has %d skipped recovery spans, want 2", len(independent))
	}
	if independent[0].End > independent[1].Start && independent[1].End > independent[0].Start {
		t.Fatalf("authored independent recovery spans overlap: %v", independent)
	}
	for _, draft := range corpus.Drafts {
		if draft.ID == "crc32-reconstructed" || draft.ID == "word-counter-reconstructed" {
			if !strings.Contains(draft.CauseRelation, "original scratch source not recovered") && !strings.Contains(draft.CauseRelation, "original scratch file not recovered") {
				t.Fatalf("%s claims unsupported scratch provenance: %q", draft.ID, draft.CauseRelation)
			}
		}
	}
}
