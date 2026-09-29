package syntax_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestSkippedRegionCauseSeparatesSpiralTrio is DX-08's RED/GREEN behavioral
// contract (14-02-PLAN.md Task 2 <behavior>). Before the skipped_region
// cause fix, all three spiral trio members collide on one diagnostic ID and
// one result: ID (frozen at testdata/distinctness/collision_control.json).
// After the fix:
//   - the declaration-recovery diagnostic's causes[0].span extent differs
//     across all three programs (wide / narrower / one-token region)
//   - the three programs mint three distinct diagnostic IDs and three
//     distinct result: IDs
//   - primary_span stays identical and one token wide across all three
func TestSkippedRegionCauseSeparatesSpiralTrio(t *testing.T) {
	fixtures := []string{"spiral_full.schway", "spiral_narrow.schway", "spiral_bare.schway"}

	type observed struct {
		fixture      string
		primarySpan  string
		causeSpan    string
		causeWidth   int
		diagnosticID string
		resultID     string
	}

	var results []observed
	for _, name := range fixtures {
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "distinctness", name))
		if err != nil {
			t.Fatal(err)
		}
		parsed := syntax.Parse(source)

		var declRecovery *syntaxDiagnosticView
		for _, d := range parsed.Diagnostics {
			if d.Code == "syntax.expected_declaration" {
				view := syntaxDiagnosticView{
					primaryStart: d.Primary.Start, primaryEnd: d.Primary.End,
					id: d.ID,
				}
				if len(d.Causes) > 0 && d.Causes[0].Kind == "skipped_region" && d.Causes[0].Span != nil {
					view.hasCause = true
					view.causeStart = d.Causes[0].Span.Start
					view.causeEnd = d.Causes[0].Span.End
				}
				declRecovery = &view
				break
			}
		}
		if declRecovery == nil {
			t.Fatalf("%s: no syntax.expected_declaration diagnostic found", name)
		}
		if !declRecovery.hasCause {
			t.Fatalf("%s: syntax.expected_declaration diagnostic carries no skipped_region cause", name)
		}

		fileResult, err := session.CheckCommandFile(testsupport.ProjectPath("testdata", "distinctness", name))
		if err != nil {
			t.Fatalf("%s: CheckCommandFile: %v", name, err)
		}

		results = append(results, observed{
			fixture:      name,
			primarySpan:  spanKey(declRecovery.primaryStart, declRecovery.primaryEnd),
			causeSpan:    spanKey(declRecovery.causeStart, declRecovery.causeEnd),
			causeWidth:   declRecovery.causeEnd - declRecovery.causeStart,
			diagnosticID: declRecovery.id,
			resultID:     fileResult.ID,
		})
	}

	// primary_span must be identical and one token wide (the "if" token)
	// across all three programs.
	wantPrimary := results[0].primarySpan
	for _, r := range results {
		if r.primarySpan != wantPrimary {
			t.Fatalf("primary_span moved across spiral trio: %s has %s, want %s (from %s)", r.fixture, r.primarySpan, wantPrimary, results[0].fixture)
		}
	}

	// causes[0].span must differ across all three (wide / narrower / one-token).
	if results[0].causeSpan == results[1].causeSpan || results[1].causeSpan == results[2].causeSpan || results[0].causeSpan == results[2].causeSpan {
		t.Fatalf("skipped_region cause spans are not pairwise distinct: %+v", results)
	}
	if !(results[0].causeWidth > results[1].causeWidth && results[1].causeWidth > results[2].causeWidth) {
		t.Fatalf("skipped_region cause widths are not strictly decreasing full>narrow>bare: %+v", results)
	}

	// Diagnostic IDs must be pairwise distinct.
	if results[0].diagnosticID == results[1].diagnosticID || results[1].diagnosticID == results[2].diagnosticID || results[0].diagnosticID == results[2].diagnosticID {
		t.Fatalf("diagnostic IDs are not pairwise distinct: %+v", results)
	}

	// result: IDs must be pairwise distinct.
	if results[0].resultID == results[1].resultID || results[1].resultID == results[2].resultID || results[0].resultID == results[2].resultID {
		t.Fatalf("result: IDs are not pairwise distinct: %+v", results)
	}
}

// TestValidProgramUnaffectedBySkippedRegionFix pins that a program which
// parses successfully produces zero diagnostics before and after the fix --
// the fix only ever adds a cause to an already-produced
// syntax.expected_declaration diagnostic, so a clean parse is untouched by
// construction.
func TestValidProgramUnaffectedBySkippedRegionFix(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", "toggle.schway"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("toggle.schway should parse clean, got diagnostics: %+v", parsed.Diagnostics)
	}
}

type syntaxDiagnosticView struct {
	primaryStart, primaryEnd int
	hasCause                 bool
	causeStart, causeEnd     int
	id                       string
}

func spanKey(start, end int) string {
	return fmt.Sprintf("%d:%d", start, end)
}
