package syntax_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase34RecoveryDraft struct {
	ID                    string   `json:"id"`
	Source                string   `json:"source"`
	DiagnosticCount       int      `json:"diagnostic_count"`
	RecoveredSiblingNames []string `json:"recovered_sibling_names"`
	Diagnostics           []struct {
		Code        string `json:"code"`
		PrimarySpan struct {
			Start int `json:"start"`
			End   int `json:"end"`
		} `json:"primary_span"`
		Causes []struct {
			Kind string `json:"kind"`
			Span *struct {
				Start int `json:"start"`
				End   int `json:"end"`
			} `json:"span"`
		} `json:"causes"`
	} `json:"diagnostics"`
}

func TestPhase34RecoveryCorpus(t *testing.T) {
	var corpus struct {
		Drafts []struct {
			ID     string `json:"id"`
			Source string `json:"source"`
		} `json:"drafts"`
	}
	readPhase34JSON(t, "testdata/phase34/drafts.json", &corpus)
	var after struct {
		Drafts []phase34RecoveryDraft `json:"drafts"`
	}
	readPhase34JSON(t, "testdata/phase34/draft-after.json", &after)

	var source string
	for _, draft := range corpus.Drafts {
		if draft.ID == "sibling-after-error" {
			source = draft.Source
			break
		}
	}
	if source == "" {
		t.Fatal("reconstructed sibling-after-error fixture is missing")
	}
	var want *phase34RecoveryDraft
	for i := range after.Drafts {
		if after.Drafts[i].ID == "sibling-after-error" {
			want = &after.Drafts[i]
			break
		}
	}
	if want == nil {
		t.Fatal("draft-after.json has no sibling-after-error row")
	}

	parsed := syntax.Parse([]byte(source))
	if !bytes.Equal(parsed.Tree.Bytes(), []byte(source)) {
		t.Fatal("rejected parse changed the lossless source bytes")
	}
	if got, expected := len(parsed.Diagnostics), want.DiagnosticCount; got != expected {
		t.Fatalf("sibling-after-error diagnostic count = %d, want %d: %+v", got, expected, parsed.Diagnostics)
	}
	if len(want.Diagnostics) != len(parsed.Diagnostics) {
		t.Fatalf("sibling-after-error has %d expected diagnostic rows, got %d", len(want.Diagnostics), len(parsed.Diagnostics))
	}
	for i, diagnostic := range parsed.Diagnostics {
		expected := want.Diagnostics[i]
		if diagnostic.Code != expected.Code || diagnostic.Primary.Start != expected.PrimarySpan.Start || diagnostic.Primary.End != expected.PrimarySpan.End {
			t.Fatalf("diagnostic %d = %s@%d:%d, want %s@%d:%d", i, diagnostic.Code, diagnostic.Primary.Start, diagnostic.Primary.End, expected.Code, expected.PrimarySpan.Start, expected.PrimarySpan.End)
		}
		if len(diagnostic.Causes) != len(expected.Causes) {
			t.Fatalf("diagnostic %d has %d causes, want %d", i, len(diagnostic.Causes), len(expected.Causes))
		}
		for j, cause := range diagnostic.Causes {
			wantCause := expected.Causes[j]
			if cause.Kind != wantCause.Kind || (cause.Span == nil) != (wantCause.Span == nil) {
				t.Fatalf("diagnostic %d cause %d = %+v, want %+v", i, j, cause, wantCause)
			}
			if cause.Span != nil && (cause.Span.Start != wantCause.Span.Start || cause.Span.End != wantCause.Span.End) {
				t.Fatalf("diagnostic %d cause %d span = %d:%d, want %d:%d", i, j, cause.Span.Start, cause.Span.End, wantCause.Span.Start, wantCause.Span.End)
			}
		}
	}
	var siblings []string
	for _, function := range parsed.Program.Funcs {
		siblings = append(siblings, function.Name)
	}
	if !reflect.DeepEqual(siblings, want.RecoveredSiblingNames) {
		t.Fatalf("recovered functions = %v, want %v", siblings, want.RecoveredSiblingNames)
	}

	checked := session.Check([]byte(source))
	if len(checked.Diagnostics) == 0 {
		t.Fatal("session.Check accepted a source tree with parser recovery diagnostics")
	}
	if !reflect.DeepEqual(checked.Program, core.Program{}) {
		t.Fatalf("session.Check admitted rejected syntax to core: %+v", checked.Program)
	}
	if !bytes.Equal(checked.Tree.Bytes(), []byte(source)) {
		t.Fatal("session.Check changed the lossless rejected source bytes")
	}
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
