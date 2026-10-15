package session_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/surfacemanifest"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase36SurfaceWitnessSchema = "phase36.surface-witnesses/1"

type phase36SurfaceWitness struct {
	ID                  string                      `json:"id"`
	Facet               string                      `json:"facet"`
	Context             string                      `json:"context"`
	Signature           string                      `json:"signature"`
	Source              string                      `json:"source"`
	Files               []phase34SurfaceWitnessFile `json:"files"`
	ExpectedStatus      string                      `json:"expected_status"`
	ExpectedCode        string                      `json:"expected_code"`
	ExpectedHint        string                      `json:"expected_hint"`
	Exercise            string                      `json:"exercise"`
	ExpectedForm        string                      `json:"expected_form"`
	ExpectedConstraints []string                    `json:"expected_constraints"`
}

type phase36SurfaceWitnessDocument struct {
	Schema    string                  `json:"schema"`
	Boundary  string                  `json:"boundary"`
	Calls     phase34ExpectedCalls    `json:"calls"`
	Witnesses []phase36SurfaceWitness `json:"witnesses"`
}

func TestPhase36SurfaceManifestCallWitnesses(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase36", "surface-witnesses.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read independent Phase 36 scalar-call witnesses: %v", err)
	}
	var document phase36SurfaceWitnessDocument
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode independent Phase 36 scalar-call witnesses: %v", err)
	}
	if document.Schema != phase36SurfaceWitnessSchema || document.Boundary == "" || len(document.Witnesses) == 0 {
		t.Fatalf("invalid Phase 36 witness header: schema=%q boundary=%q witnesses=%d", document.Schema, document.Boundary, len(document.Witnesses))
	}

	wantPairs := map[string]string{
		"phase36-pair-u64-u64":   "U64 -> U64",
		"phase36-pair-u64-bool":  "U64 -> Bool",
		"phase36-pair-bool-u64":  "Bool -> U64",
		"phase36-pair-bool-bool": "Bool -> Bool",
	}
	wantContexts := map[string]bool{
		"scalar binding and assignment RHS":         false,
		"scalar arithmetic and comparison operands": false,
		"scalar branch and loop conditions":         false,
		"scalar byte_at index expressions":          false,
		"nested scalar helper arguments":            false,
	}
	wantNegativeCodes := map[string]string{
		"phase36-refuse-arity-two":          "syntax.scalar_call_arity",
		"phase36-refuse-argument-type":      "check.call_argument_type_mismatch",
		"phase36-refuse-private-import":     "module.not_exported",
		"phase36-refuse-recursion":          "core.call_graph_cycle",
		"phase36-refuse-bool-literal-tail":  "surface.unsupported_bool_literal_tail",
		"phase36-refuse-byteview-signature": "type.scalar_signature",
	}
	seenIDs := map[string]bool{}
	var observedInput []phase34SurfaceWitness
	for _, witness := range document.Witnesses {
		if witness.ID == "" || seenIDs[witness.ID] {
			t.Fatalf("duplicate or empty Phase 36 witness ID %q", witness.ID)
		}
		seenIDs[witness.ID] = true
		if pair, ok := wantPairs[witness.ID]; ok {
			if witness.Facet != "restriction:scalar_expression_calls" || witness.Signature != pair || witness.ExpectedStatus != "pass" {
				t.Errorf("signature witness %s = facet %q pair %q status %q; want %q and a passing complete program", witness.ID, witness.Facet, witness.Signature, witness.ExpectedStatus, pair)
			}
			delete(wantPairs, witness.ID)
		}
		if _, ok := wantContexts[witness.Context]; ok && witness.ExpectedStatus == "pass" {
			wantContexts[witness.Context] = true
		}
		if expected, ok := wantNegativeCodes[witness.ID]; ok {
			if witness.ExpectedStatus != "invalid" || witness.ExpectedCode != expected {
				t.Errorf("negative witness %s expects %q/%q, want invalid/%q", witness.ID, witness.ExpectedStatus, witness.ExpectedCode, expected)
			}
			delete(wantNegativeCodes, witness.ID)
		}
		if witness.ID == "phase36-imported-exported-helper" {
			if len(witness.Files) != 1 || !strings.Contains(witness.Files[0].Source, "export { fn adjust }") {
				t.Errorf("imported helper witness does not independently declare its exported local target: %+v", witness.Files)
			}
		}
		observedInput = append(observedInput, phase34SurfaceWitness{
			ID: witness.ID, Facet: witness.Facet, Context: witness.Context, Files: witness.Files,
			Source: witness.Source, ExpectedStatus: witness.ExpectedStatus, ExpectedCode: witness.ExpectedCode,
			ExpectedHint: witness.ExpectedHint, Exercise: witness.Exercise, ExpectedForm: witness.ExpectedForm,
			ExpectedConstraints: witness.ExpectedConstraints,
		})
	}
	if len(wantPairs) != 0 {
		t.Errorf("missing exact scalar signature pair witnesses: %v", wantPairs)
	}
	for context, found := range wantContexts {
		if !found {
			t.Errorf("missing complete-program witness for admitted scalar call context %q", context)
		}
	}
	for id, code := range wantNegativeCodes {
		t.Errorf("missing complete-program negative control %s with code %s", id, code)
	}

	observed := phase34ObserveSurfaceWitnesses(t, observedInput)
	manifest := surfacemanifest.Current()
	if manifest.Schema != "schway.surface/1" || !manifest.Calls.ScalarExpressionCalls || !document.Calls.ScalarExpressionCalls {
		t.Fatalf("source-free scalar call discovery disagrees with the literal witness contract: schema=%q calls=%+v expected=%+v", manifest.Schema, manifest.Calls, document.Calls)
	}
	if manifest.Calls.DeclarationArity != document.Calls.DeclarationArity || !samePhase34Strings(manifest.Calls.OrdinaryContexts, document.Calls.OrdinaryContexts) ||
		!samePhase34Strings(manifest.Calls.Constraints, document.Calls.Constraints) {
		t.Fatalf("source-free call rules differ from independently authored literals: actual=%+v expected=%+v", manifest.Calls, document.Calls)
	}
	candidates := []string{"restriction:scalar_expression_calls"}
	for _, witness := range document.Witnesses {
		if strings.HasPrefix(witness.Facet, "probe:") {
			candidates = append(candidates, witness.Facet)
		}
	}
	if problems := phase34SurfaceProblems(manifest, observed, candidates); len(problems) != 0 {
		t.Fatalf("independent Phase 36 complete-program witnesses disagree with the source-free call contract: %v", problems)
	}
}
