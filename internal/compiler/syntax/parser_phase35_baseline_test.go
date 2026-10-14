package syntax_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase35NestedBaselineFile struct {
	Schema   string                       `json:"schema"`
	Revision string                       `json:"revision"`
	Indices  []phase35NestedBaselineEntry `json:"nested_index_cases"`
}

type phase35NestedBaselineEntry struct {
	Depth       int                      `json:"depth"`
	Source      string                   `json:"source"`
	SourceSHA   string                   `json:"source_sha256"`
	Parse       phase35NestedObservation `json:"parse"`
	Check       phase35NestedObservation `json:"check"`
	CoreSchema  string                   `json:"core_schema,omitempty"`
	CoreSHA     string                   `json:"core_sha256,omitempty"`
	SemanticIDs []string                 `json:"semantic_ids,omitempty"`
	Emit        phase35NestedEmission    `json:"emit"`
	EmitNative  phase35NestedEmission    `json:"emit_native"`
	Application phase35NestedEmission    `json:"emit_application"`
}

type phase35NestedObservation struct {
	Outcome     string                       `json:"outcome"`
	Diagnostics []phase35NestedDiagnosticRow `json:"diagnostics,omitempty"`
}

type phase35NestedDiagnosticRow struct {
	ID      string `json:"id"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Primary struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"primary_span"`
}

type phase35NestedEmission struct {
	Outcome string `json:"outcome"`
	Schema  string `json:"schema,omitempty"`
	Error   string `json:"error,omitempty"`
	CSource string `json:"c_source,omitempty"`
}

func TestPhase35NestedIndexBaseline(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase35", "numeric-policy-baseline.json"))
	if err != nil {
		t.Fatalf("read nested-index baseline: %v", err)
	}
	var fixture phase35NestedBaselineFile
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode nested-index baseline: %v", err)
	}
	if fixture.Schema != "phase35.numeric-policy-baseline/1" || fixture.Revision == "" {
		t.Fatalf("invalid nested-index baseline header: schema=%q revision=%q", fixture.Schema, fixture.Revision)
	}
	if err := exec.Command("git", "cat-file", "-e", fixture.Revision+"^{commit}").Run(); err != nil {
		t.Fatalf("frozen nested-index revision %s is unavailable: %v", fixture.Revision, err)
	}
	wantDepths := []int{63, 64, 65, 128}
	if len(fixture.Indices) != len(wantDepths) {
		t.Fatalf("nested-index depths=%d, want %v", len(fixture.Indices), wantDepths)
	}
	for index, expectedDepth := range wantDepths {
		row := fixture.Indices[index]
		if row.Depth != expectedDepth {
			t.Fatalf("nested-index row %d has depth %d, want %d", index, row.Depth, expectedDepth)
		}
		source := []byte(row.Source)
		if phase35NestedSHA(source) != row.SourceSHA {
			t.Fatalf("depth %d source bytes differ from their frozen digest", row.Depth)
		}
		expr := "0"
		for nested := 0; nested < row.Depth; nested++ {
			expr = "byte_at(input, " + expr + ")"
		}
		wantSource := strings.Join([]string{
			fmt.Sprintf("module phase35.index.depth_%d", row.Depth),
			"export {", "  fn main", "  fn nested", "}",
			"fn main(input: U64) -> U64 {", "  input", "}",
			"fn nested(input: ByteView) -> U64 {", "  let value = " + expr, "  value", "}", "",
		}, "\n")
		if row.Source != wantSource {
			t.Fatalf("depth %d source does not contain exactly %d nested byte_at index calls", row.Depth, row.Depth)
		}
		if row.Parse.Outcome != "admitted" || row.Check.Outcome != "admitted" {
			t.Fatalf("depth %d historical parse/check outcomes=%s/%s, want admitted/admitted", row.Depth, row.Parse.Outcome, row.Check.Outcome)
		}
		for _, observation := range []phase35NestedObservation{row.Parse, row.Check} {
			for _, problem := range observation.Diagnostics {
				if problem.Primary.Start < 0 || problem.Primary.End < problem.Primary.Start || problem.Primary.End > len(source) {
					t.Fatalf("depth %d diagnostic %s has out-of-source span %+v", row.Depth, problem.Code, problem.Primary)
				}
			}
		}
		if row.CoreSchema == "" || row.CoreSHA == "" || len(row.SemanticIDs) == 0 {
			t.Fatalf("depth %d admitted historical core lacks schema/hash/identity witnesses", row.Depth)
		}
		for route, emission := range map[string]phase35NestedEmission{"Emit": row.Emit, "EmitNative": row.EmitNative, "EmitApplication": row.Application} {
			if emission.Outcome != "emitted" || emission.CSource == "" || emission.Error != "" {
				t.Fatalf("depth %d %s historical route is not an exact emitted C witness: %+v", row.Depth, route, emission)
			}
		}
	}
}

func phase35NestedSHA(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
