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
	CheckRoute          string                      `json:"check_route"`
	Signature           string                      `json:"signature"`
	Source              string                      `json:"source"`
	Files               []phase34SurfaceWitnessFile `json:"files"`
	ExpectedStatus      string                      `json:"expected_status"`
	ExpectedCode        string                      `json:"expected_code"`
	ExpectedHint        string                      `json:"expected_hint"`
	Exercise            string                      `json:"exercise"`
	ExpectedForm        string                      `json:"expected_form"`
	ExpectedArity       int                         `json:"expected_arity"`
	ExpectedConstraints []string                    `json:"expected_constraints"`
}

type phase36SurfaceWitnessDocument struct {
	Schema     string                      `json:"schema"`
	Boundary   string                      `json:"boundary"`
	Calls      phase34ExpectedCalls        `json:"calls"`
	Supersedes []phase34ScopedSupersession `json:"supersedes"`
	Witnesses  []phase36SurfaceWitness     `json:"witnesses"`
}

const phase36SurfaceCensusSchema = "phase36.surface-census/1"

type phase36LimitContext struct {
	ID             string `json:"id"`
	Value          int    `json:"value"`
	Unit           string `json:"unit"`
	Meaning        string `json:"meaning"`
	Source         string `json:"source"`
	Classification string `json:"classification"`
}

type phase36SurfaceCensusDocument struct {
	Schema        string                `json:"schema"`
	Boundary      string                `json:"boundary"`
	Files         []string              `json:"files"`
	Branches      []phase34CensusBranch `json:"branches"`
	LimitContexts []phase36LimitContext `json:"limit_contexts"`
}

func phase36ReadSurfaceWitnesses(t *testing.T) phase36SurfaceWitnessDocument {
	t.Helper()
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
	return document
}

func phase36ReadSurfaceCensus(t *testing.T) phase36SurfaceCensusDocument {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase36", "surface-census.json"))
	if err != nil {
		t.Fatalf("read independent Phase 36 call census: %v", err)
	}
	var document phase36SurfaceCensusDocument
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode independent Phase 36 call census: %v", err)
	}
	if document.Schema != phase36SurfaceCensusSchema || document.Boundary == "" || len(document.Branches) == 0 {
		t.Fatalf("invalid Phase 36 census header: schema=%q boundary=%q branches=%d", document.Schema, document.Boundary, len(document.Branches))
	}
	if !samePhase34Strings(document.Files, phase34SurfaceCensusFiles) {
		t.Fatalf("Phase 36 census boundary files = %v, want exactly %v", document.Files, phase34SurfaceCensusFiles)
	}
	return document
}

func TestPhase36SurfaceManifestCallCensus(t *testing.T) {
	witnessDocument := phase34ReadSurfaceWitnesses(t)
	phase36Witnesses := phase36ReadSurfaceWitnesses(t)
	for _, witness := range phase36Witnesses.Witnesses {
		witnessDocument.Witnesses = append(witnessDocument.Witnesses, phase34SurfaceWitness{
			ID: witness.ID, Facet: witness.Facet, Context: witness.Context, CheckRoute: witness.CheckRoute,
			Files: witness.Files, Source: witness.Source, ExpectedStatus: witness.ExpectedStatus,
			ExpectedCode: witness.ExpectedCode, ExpectedHint: witness.ExpectedHint, Exercise: witness.Exercise,
			ExpectedForm: witness.ExpectedForm, ExpectedArity: witness.ExpectedArity, ExpectedConstraints: witness.ExpectedConstraints,
		})
	}
	census := phase34ReadCensus(t)
	phase36Census := phase36ReadSurfaceCensus(t)
	census.Branches = append(census.Branches, phase36Census.Branches...)
	if problems := phase34CensusProblems(census, nil, witnessDocument.Witnesses); len(problems) != 0 {
		t.Fatalf("composed Phase 34/35/36 source branch census differs from selected producers; expected Phase 36 call census IDs to be accounted: %v", problems)
	}
	for _, row := range phase36Census.Branches {
		if row.Disposition == "non-surface/internal" && row.Reason == "" {
			t.Fatalf("Phase 36 internal branch lacks a named exclusion: %+v", row)
		}
	}
}

func TestPhase36SurfaceManifestCallControls(t *testing.T) {
	witnessDocument := phase36ReadSurfaceWitnesses(t)
	if witnessDocument.Calls.DeclarationArity != 1 || !samePhase34Ints(witnessDocument.Calls.OrdinaryArity, []int{1}) || !witnessDocument.Calls.ScalarExpressionCalls {
		t.Fatalf("Phase 36 literal call controls do not describe one-argument scalar calls: %+v", witnessDocument.Calls)
	}
	if len(witnessDocument.Supersedes) == 0 {
		t.Fatal("Phase 36 call composition must name exact inherited witness supersessions")
	}
	manifest := surfacemanifest.Current()
	if !manifest.Calls.ScalarExpressionCalls {
		t.Fatal("scalar expression call admission is absent from the source-free manifest")
	}
}

func TestPhase36SurfaceLimitConsumers(t *testing.T) {
	document := phase36ReadSurfaceCensus(t)
	want := map[string]phase36LimitContext{
		"parser.expression_depth":    {ID: "parser.expression_depth", Value: 64, Unit: "levels", Classification: "source-expression-depth"},
		"interpreter.call_depth":     {ID: "interpreter.call_depth", Value: 128, Unit: "frames", Classification: "runtime-call-stack"},
		"cgen.invocation_path_nodes": {ID: "cgen.invocation_path_nodes", Value: 4096, Unit: "static nodes", Classification: "static-preflight"},
		"executionpeer.invocations":  {ID: "executionpeer.invocations", Value: 4096, Unit: "static occurrences", Classification: "static-preflight"},
		"execution.document_bytes":   {ID: "execution.document_bytes", Value: 16 * 1024 * 1024, Unit: "bytes", Classification: "runtime-capture-bytes"},
		"application.evidence_bytes": {ID: "application.evidence_bytes", Value: 64 * 1024, Unit: "bytes", Classification: "runtime-capture-bytes"},
		"application.output_bytes":   {ID: "application.output_bytes", Value: 65536, Unit: "bytes", Classification: "runtime-capture-bytes"},
	}
	for _, row := range document.LimitContexts {
		expected, ok := want[row.ID]
		if !ok || row.Value != expected.Value || row.Unit != expected.Unit || row.Classification != expected.Classification || row.Meaning == "" || row.Source == "" {
			t.Errorf("limit context %q = %+v; want independently classified value %d %s (%s) with a named source", row.ID, row, expected.Value, expected.Unit, expected.Classification)
			continue
		}
		delete(want, row.ID)
	}
	if len(want) != 0 {
		t.Errorf("missing exact Phase 36 limit contexts: %v", want)
	}
}

func samePhase34Ints(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
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
		"phase36-refuse-arity-two":          "check.call_arity_unsupported",
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
			ID: witness.ID, Facet: witness.Facet, Context: witness.Context, CheckRoute: witness.CheckRoute, Files: witness.Files,
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
	for _, row := range observed {
		witness := row.Witness
		if row.Status != witness.ExpectedStatus || row.Code != witness.ExpectedCode || witness.ExpectedHint != "" && row.Hint != witness.ExpectedHint {
			t.Errorf("witness %s (%s): got status=%q code=%q hint=%q; want status=%q code=%q hint=%q", witness.ID, witness.Facet, row.Status, row.Code, row.Hint, witness.ExpectedStatus, witness.ExpectedCode, witness.ExpectedHint)
		}
	}
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
