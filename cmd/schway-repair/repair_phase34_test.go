package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
)

func TestPhase34RepairJSONCompatibility(t *testing.T) {
	original := phase6Fixture(t, "heldout_move_defect.schway")
	broken, err := session.MoveInjector{}.Inject(original)
	if err != nil {
		t.Fatal(err)
	}

	inputPath := filepath.Join(t.TempDir(), "move.schway")
	if err := os.WriteFile(inputPath, broken, 0o600); err != nil {
		t.Fatal(err)
	}
	command, err := session.CheckCommandFile(inputPath)
	if err != nil || command.Status != "invalid" || command.Presentation == nil {
		t.Fatalf("check command = %+v, err=%v; want enriched invalid source diagnostic", command, err)
	}
	raw, err := protocol.JSON(command)
	if err != nil {
		t.Fatal(err)
	}
	decoded := decodeCheckJSON(t, raw)
	selected, _, ok := selectRepair(decoded)
	if !ok || selected.Applicability != applicabilityMachineApplicable || selected.Span == nil || selected.Replacement == "" {
		t.Fatalf("actual enriched check JSON has no eligible byte repair: %+v", decoded.Diagnostics)
	}
	if selected.Span.Start < 0 || selected.Span.End < selected.Span.Start || selected.Span.End > len(broken) {
		t.Fatalf("actual byte span is outside source: %+v / %d", selected.Span, len(broken))
	}

	baselineOutcome, baselineBytes := phase34RepairRun(t, broken, raw)
	wantBytes := phase34Splice(broken, selected.Span, selected.Replacement)
	if baselineOutcome.Status != OutcomeRepaired || baselineOutcome.SubprocessCount != 2 || !bytes.Equal(baselineBytes, wantBytes) {
		t.Fatalf("baseline repair = %+v, bytes changed as %q; want one exact eligible byte edit %q", baselineOutcome, baselineBytes, wantBytes)
	}

	enrichedMutation := phase34MutatePresentationFields(t, raw)
	mutatedCheck := decodeCheckJSON(t, enrichedMutation)
	mutatedRepair, _, ok := selectRepair(mutatedCheck)
	if !ok || !reflect.DeepEqual(selected, mutatedRepair) {
		t.Fatalf("location/prose mutation changed selected byte authority: before=%+v after=%+v", selected, mutatedRepair)
	}
	mutatedOutcome, mutatedBytes := phase34RepairRun(t, broken, enrichedMutation)
	if mutatedOutcome != baselineOutcome || !bytes.Equal(mutatedBytes, baselineBytes) {
		t.Fatalf("location/prose changed repair result: before=(%+v,%q) after=(%+v,%q)", baselineOutcome, baselineBytes, mutatedOutcome, mutatedBytes)
	}

	crossUnit := phase34SetRepairUnit(t, raw, "lib.foreign")
	crossOutcome, crossBytes := phase34RepairRun(t, broken, crossUnit)
	if crossOutcome.Status != OutcomeUnrepairable || crossOutcome.DeclineReason != DeclineNoneEligible || crossOutcome.SubprocessCount != 1 || !bytes.Equal(crossBytes, broken) {
		t.Fatalf("cross-unit repair was not declined: outcome=%+v sourceChanged=%t", crossOutcome, !bytes.Equal(crossBytes, broken))
	}
}

func phase34RepairRun(t *testing.T, source, checkJSON []byte) (Outcome, []byte) {
	t.Helper()
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "input.schway")
	if err := os.WriteFile(sourcePath, source, 0o600); err != nil {
		t.Fatal(err)
	}
	langBinary := phase30FakeLang(t, directory, string(checkJSON))
	outcome, err := Repair(context.Background(), langBinary, sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	result, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	return outcome, result
}

func phase34Splice(source []byte, span *jsonSpan, replacement string) []byte {
	result := make([]byte, 0, len(source)-(span.End-span.Start)+len(replacement))
	result = append(result, source[:span.Start]...)
	result = append(result, replacement...)
	result = append(result, source[span.End:]...)
	return result
}

func phase34MutatePresentationFields(t *testing.T, raw []byte) []byte {
	t.Helper()
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	diagnostics, ok := document["diagnostics"].([]any)
	if !ok || len(diagnostics) == 0 {
		t.Fatalf("enriched JSON has no diagnostics: %s", raw)
	}
	var selected map[string]any
	var repair map[string]any
	for _, rawDiagnostic := range diagnostics {
		item, ok := rawDiagnostic.(map[string]any)
		if !ok {
			continue
		}
		repairs, _ := item["repairs"].([]any)
		for _, rawRepair := range repairs {
			candidate, ok := rawRepair.(map[string]any)
			if ok && candidate["applicability"] == applicabilityMachineApplicable {
				selected, repair = item, candidate
				break
			}
		}
		if repair != nil {
			break
		}
	}
	if selected == nil || repair == nil {
		t.Fatalf("actual enriched JSON has no machine-applicable repair: %s", raw)
	}
	mutateSpan := func(raw any) {
		span, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("presentation span is absent: %v", raw)
		}
		span["line"], span["col"] = float64(901), float64(77)
	}
	mutateSpan(selected["primary_span"])
	mutateSpan(repair["span"])
	selected["message"] = "mutated explanatory hint must never steer a byte edit"
	selected["hint"] = "invented prose must not gain edit authority"
	repair["detail"] = "mutated repair explanation"
	repair["hint"] = "mutated repair hint"
	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func phase34SetRepairUnit(t *testing.T, raw []byte, unit string) []byte {
	t.Helper()
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	diagnostics, ok := document["diagnostics"].([]any)
	if !ok || len(diagnostics) == 0 {
		t.Fatalf("check JSON has no diagnostics: %s", raw)
	}
	for _, rawDiagnostic := range diagnostics {
		item, ok := rawDiagnostic.(map[string]any)
		if !ok {
			continue
		}
		repairs, _ := item["repairs"].([]any)
		for _, rawRepair := range repairs {
			candidate, ok := rawRepair.(map[string]any)
			if ok && candidate["applicability"] == applicabilityMachineApplicable {
				candidate["unit"] = unit
				encoded, err := json.Marshal(document)
				if err != nil {
					t.Fatal(err)
				}
				return encoded
			}
		}
	}
	t.Fatalf("check JSON has no machine-applicable repair: %s", raw)
	return nil
}
