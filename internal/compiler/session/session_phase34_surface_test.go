package session_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/surfacemanifest"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase34SurfaceWitnessSchema = "phase34.surface-witnesses/1"

type phase34SurfaceWitnessFile struct {
	Path   string `json:"path"`
	Source string `json:"source"`
}

type phase34SurfaceWitness struct {
	ID                  string                      `json:"id"`
	Facet               string                      `json:"facet"`
	Context             string                      `json:"context"`
	Source              string                      `json:"source"`
	Files               []phase34SurfaceWitnessFile `json:"files"`
	ExpectedStatus      string                      `json:"expected_status"`
	ExpectedCode        string                      `json:"expected_code"`
	ExpectedHint        string                      `json:"expected_hint"`
	ExpectedArity       int                         `json:"expected_arity"`
	ExpectedConstraints []string                    `json:"expected_constraints"`
}

type phase34SurfaceWitnessDocument struct {
	Schema    string                  `json:"schema"`
	Boundary  string                  `json:"boundary"`
	Witnesses []phase34SurfaceWitness `json:"witnesses"`
}

type phase34SurfaceClaim struct {
	Facet       string
	Form        string
	Contexts    []string
	Constraints []string
	Arity       int
	Arities     []int
	Inputs      []string
	Result      string
	Scope       string
	Unit        string
	Value       int64
	Meaning     string
	Routes      []string
	Code        string
	Hint        string
}

type phase34ObservedSurfaceWitness struct {
	Witness phase34SurfaceWitness
	Status  string
	Code    string
	Hint    string
}

type phase34SurfaceProblem struct {
	Reason string
	Facet  string
}

func phase34ReadSurfaceWitnesses(t *testing.T) phase34SurfaceWitnessDocument {
	t.Helper()
	path := testsupport.ProjectPath("testdata", "phase34", "surface-witnesses.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read independent surface witnesses: %v", err)
	}
	var document phase34SurfaceWitnessDocument
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode independent surface witnesses: %v", err)
	}
	if document.Schema != phase34SurfaceWitnessSchema {
		t.Fatalf("surface witness schema = %q, want %q", document.Schema, phase34SurfaceWitnessSchema)
	}
	if document.Boundary == "" {
		t.Fatal("surface witness document must state its finite source and input boundary")
	}
	return document
}

func phase34CheckSurfaceWitness(t *testing.T, witness phase34SurfaceWitness) (string, string, string) {
	t.Helper()
	root := t.TempDir()
	writeSource := func(relative, source string) string {
		t.Helper()
		clean := filepath.Clean(filepath.FromSlash(relative))
		if clean == "." || filepath.IsAbs(clean) || clean == ".." || len(clean) >= 3 && clean[:3] == ".."+string(filepath.Separator) {
			t.Fatalf("witness %s contains unsafe relative source path %q", witness.ID, relative)
		}
		path := filepath.Join(root, clean)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create witness directory for %s: %v", witness.ID, err)
		}
		if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
			t.Fatalf("write witness source for %s: %v", witness.ID, err)
		}
		return path
	}
	entry := writeSource("main.schway", witness.Source)
	for _, file := range witness.Files {
		writeSource(file.Path, file.Source)
	}
	result, err := session.CheckCommandFile(entry)
	if err != nil {
		t.Fatalf("CheckCommandFile witness %s: %v", witness.ID, err)
	}
	code, hint := "", ""
	if len(result.Diagnostics) > 0 {
		code = result.Diagnostics[0].Code
		hint = result.Diagnostics[0].Message
	}
	return string(result.Status), code, hint
}

func phase34ObserveSurfaceWitnesses(t *testing.T, witnesses []phase34SurfaceWitness) []phase34ObservedSurfaceWitness {
	t.Helper()
	observed := make([]phase34ObservedSurfaceWitness, 0, len(witnesses))
	for _, witness := range witnesses {
		status, code, hint := phase34CheckSurfaceWitness(t, witness)
		observed = append(observed, phase34ObservedSurfaceWitness{Witness: witness, Status: status, Code: code, Hint: hint})
	}
	return observed
}

func phase34SurfaceClaims(manifest surfacemanifest.Manifest) []phase34SurfaceClaim {
	claims := make([]phase34SurfaceClaim, 0, len(manifest.Types)+len(manifest.Operators)+len(manifest.Statements)+len(manifest.Builtins)+len(manifest.Restrictions)+len(manifest.Limits)+len(manifest.Refusals)+2)
	for _, entry := range manifest.Types {
		claims = append(claims, phase34SurfaceClaim{Facet: "type:" + entry.ID, Form: entry.Syntax, Contexts: entry.Contexts, Constraints: entry.Constraints, Arity: entry.Arity})
	}
	for _, entry := range manifest.Operators {
		claims = append(claims, phase34SurfaceClaim{Facet: "operator:" + entry.ID, Form: entry.Syntax, Contexts: entry.Contexts, Constraints: entry.Constraints})
	}
	for _, entry := range manifest.Statements {
		claims = append(claims, phase34SurfaceClaim{Facet: "statement:" + entry.ID, Form: entry.Syntax, Contexts: entry.Contexts, Constraints: entry.Constraints})
	}
	for _, entry := range manifest.Builtins {
		claims = append(claims, phase34SurfaceClaim{Facet: "builtin:" + entry.ID, Contexts: entry.Contexts, Constraints: entry.Constraints, Arity: entry.Arity, Inputs: entry.Inputs, Result: entry.Result})
	}
	claims = append(claims, phase34SurfaceClaim{Facet: "call:declaration", Arity: manifest.Calls.DeclarationArity, Constraints: manifest.Calls.Constraints})
	claims = append(claims, phase34SurfaceClaim{Facet: "call:ordinary", Arities: manifest.Calls.OrdinaryArity, Contexts: manifest.Calls.OrdinaryContexts, Constraints: manifest.Calls.Constraints})
	if manifest.Calls.ScalarExpressionCalls {
		claims = append(claims, phase34SurfaceClaim{Facet: "call:scalar_expression", Contexts: []string{"scalar CFG expressions"}})
	}
	for _, entry := range manifest.Restrictions {
		claims = append(claims, phase34SurfaceClaim{Facet: "restriction:" + entry.ID, Form: entry.Syntax, Contexts: entry.Contexts, Constraints: entry.Constraints})
	}
	for _, entry := range manifest.Limits {
		claims = append(claims, phase34SurfaceClaim{Facet: "limit:" + entry.ID, Form: entry.Category, Scope: entry.Scope, Unit: entry.Unit, Value: entry.Value, Meaning: entry.Meaning, Routes: entry.Routes})
	}
	for _, entry := range manifest.Refusals {
		claims = append(claims, phase34SurfaceClaim{Facet: "refusal:" + entry.Code, Form: entry.Form, Code: entry.Code, Hint: entry.Hint})
	}
	return claims
}

// phase34SurfaceProblems compares only the explicit finite candidate set. It
// uses literal witness expectations and checked observations independently of
// the producer's claim data. Later tasks expand candidates from one operator
// to the authored witness matrix and source census.
func phase34SurfaceProblems(manifest surfacemanifest.Manifest, observed []phase34ObservedSurfaceWitness, candidates []string) []phase34SurfaceProblem {
	claims := map[string]phase34SurfaceClaim{}
	for _, claim := range phase34SurfaceClaims(manifest) {
		claims[claim.Facet] = claim
	}
	witnesses := map[string][]phase34ObservedSurfaceWitness{}
	for _, witness := range observed {
		witnesses[witness.Witness.Facet] = append(witnesses[witness.Witness.Facet], witness)
	}
	seen := map[string]bool{}
	var problems []phase34SurfaceProblem
	add := func(reason, facet string) {
		for _, existing := range problems {
			if existing.Reason == reason && existing.Facet == facet {
				return
			}
		}
		problems = append(problems, phase34SurfaceProblem{Reason: reason, Facet: facet})
	}
	for _, facet := range candidates {
		if seen[facet] {
			continue
		}
		seen[facet] = true
		claim, hasClaim := claims[facet]
		rows := witnesses[facet]
		if !hasClaim {
			for _, row := range rows {
				if row.Witness.ExpectedStatus == "pass" {
					add("missing_admission", facet)
				} else {
					add("missing_refusal", facet)
				}
			}
			continue
		}
		if len(rows) == 0 {
			add("invented_admission", facet)
			continue
		}
		for _, row := range rows {
			witness := row.Witness
			if row.Status != witness.ExpectedStatus || row.Code != witness.ExpectedCode {
				if strings.HasPrefix(facet, "refusal:") {
					add("missing_refusal", facet)
				} else {
					add("witness_outcome_changed", facet)
				}
			}
			if !containsPhase34SurfaceValue(claim.Contexts, witness.Context) || !reflect.DeepEqual(claim.Constraints, witness.ExpectedConstraints) {
				add("constraint_changed", facet)
			}
			arityIsPublished := strings.HasPrefix(facet, "type:") || strings.HasPrefix(facet, "builtin:") || strings.HasPrefix(facet, "call:")
			if witness.ExpectedArity != 0 && arityIsPublished && claim.Arity != witness.ExpectedArity {
				add("constraint_changed", facet)
			}
			if witness.ExpectedHint != "" && claim.Hint != witness.ExpectedHint {
				add("hint_changed", facet)
			}
		}
	}
	return problems
}

func TestPhase34SurfaceManifestWitnesses(t *testing.T) {
	document := phase34ReadSurfaceWitnesses(t)
	if len(document.Witnesses) != 1 {
		t.Fatalf("task 1 requires exactly its single independent witness, got %d", len(document.Witnesses))
	}
	witness := document.Witnesses[0]
	if witness.ID != "u64-add-scalar-operands" || witness.Facet != "operator:u64.add" {
		t.Fatalf("task 1 witness identity changed: id=%q facet=%q", witness.ID, witness.Facet)
	}
	manifest := surfacemanifest.Current()
	observed := phase34ObserveSurfaceWitnesses(t, document.Witnesses)
	clean := phase34SurfaceProblems(manifest, observed, []string{witness.Facet})
	if len(clean) != 0 {
		t.Fatalf("clean witness/manifest comparison: %v", clean)
	}
	if witness.ExpectedArity != 2 {
		t.Fatalf("binary source witness arity = %d, want independent literal 2", witness.ExpectedArity)
	}

	// A copied manifest-only logical operator has no witness. The same
	// comparator must name the invented facet, while the clean checked source
	// remains accepted and the real manifest remains untouched.
	mutated := manifest
	mutated.Operators = append(append([]surfacemanifest.Entry(nil), manifest.Operators...), surfacemanifest.Entry{
		ID:          "u64.logical_and",
		Syntax:      "&&",
		Contexts:    []string{"scalar Bool operands"},
		Constraints: []string{"No independent complete-source witness exists."},
	})
	problems := phase34SurfaceProblems(mutated, observed, []string{witness.Facet, "operator:u64.logical_and"})
	if len(problems) != 1 || problems[0].Reason != "invented_admission" || problems[0].Facet != "operator:u64.logical_and" {
		t.Fatalf("manifest-only operator control = %v, want invented_admission for operator:u64.logical_and", problems)
	}
}

func containsPhase34SurfaceValue(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
