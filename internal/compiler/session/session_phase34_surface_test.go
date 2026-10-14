package session_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/surfacemanifest"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase34SurfaceWitnessSchema = "phase34.surface-witnesses/1"
const phase34SurfaceCensusSchema = "phase34.surface-census/1"

type phase34SurfaceWitnessFile struct {
	Path   string `json:"path"`
	Source string `json:"source"`
}

type phase34SurfaceWitness struct {
	ID                  string                      `json:"id"`
	Facet               string                      `json:"facet"`
	Context             string                      `json:"context"`
	SourcePath          string                      `json:"source_path"`
	Source              string                      `json:"source"`
	Files               []phase34SurfaceWitnessFile `json:"files"`
	ExpectedStatus      string                      `json:"expected_status"`
	ExpectedCode        string                      `json:"expected_code"`
	ExpectedHint        string                      `json:"expected_hint"`
	Exercise            string                      `json:"exercise"`
	ExpectedForm        string                      `json:"expected_form"`
	ExpectedArity       int                         `json:"expected_arity"`
	ExpectedInputs      []string                    `json:"expected_inputs"`
	ExpectedResult      string                      `json:"expected_result"`
	ExpectedConstraints []string                    `json:"expected_constraints"`
	ExpectedCategory    string                      `json:"expected_category"`
	ExpectedScope       string                      `json:"expected_scope"`
	ExpectedUnit        string                      `json:"expected_unit"`
	ExpectedValue       int64                       `json:"expected_value"`
	ExpectedMeaning     string                      `json:"expected_meaning"`
	ExpectedRoutes      []string                    `json:"expected_routes"`
}

type phase34SurfaceWitnessDocument struct {
	Schema    string                  `json:"schema"`
	Boundary  string                  `json:"boundary"`
	Witnesses []phase34SurfaceWitness `json:"witnesses"`
	Limits    []phase34ExpectedLimit  `json:"limits"`
	Calls     phase34ExpectedCalls    `json:"calls"`
}

type phase34ExpectedLimit struct {
	ID       string   `json:"id"`
	Category string   `json:"category"`
	Scope    string   `json:"scope"`
	Unit     string   `json:"unit"`
	Value    int64    `json:"value"`
	Meaning  string   `json:"meaning"`
	Routes   []string `json:"routes"`
}

type phase34ExpectedCalls struct {
	DeclarationArity      int      `json:"declaration_arity"`
	OrdinaryArity         []int    `json:"ordinary_arity"`
	OrdinaryContexts      []string `json:"ordinary_contexts"`
	ScalarExpressionCalls bool     `json:"scalar_expression_calls"`
	Constraints           []string `json:"constraints"`
}

type phase34CensusBranch struct {
	File        string   `json:"file"`
	Function    string   `json:"function"`
	Selector    string   `json:"selector"`
	Values      []string `json:"values"`
	Facet       string   `json:"facet"`
	Disposition string   `json:"disposition"`
	WitnessIDs  []string `json:"witness_ids"`
	Reason      string   `json:"reason"`
}

type phase34CensusDocument struct {
	Schema   string                `json:"schema"`
	Boundary string                `json:"boundary"`
	Files    []string              `json:"files"`
	Branches []phase34CensusBranch `json:"branches"`
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
	ScalarCalls bool
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
	if witness.SourcePath != "" {
		path := testsupport.ProjectPath(filepath.FromSlash(witness.SourcePath))
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read source witness %s: %v", witness.ID, err)
		}
		if witness.Exercise != "" && !strings.Contains(string(source), witness.Exercise) {
			t.Fatalf("source witness %s does not contain its literal exercise %q", witness.ID, witness.Exercise)
		}
		result, err := session.CheckCommandFile(path)
		if err != nil {
			t.Fatalf("CheckCommandFile witness %s from %s: %v", witness.ID, witness.SourcePath, err)
		}
		code, hint := firstPhase34Diagnostic(result.Diagnostics)
		return string(result.Status), code, hint
	}
	root := t.TempDir()
	if witness.Exercise != "" && !strings.Contains(witness.Source, witness.Exercise) {
		t.Fatalf("source witness %s does not contain its literal exercise %q", witness.ID, witness.Exercise)
	}
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
	code, hint := firstPhase34Diagnostic(result.Diagnostics)
	return string(result.Status), code, hint
}

func firstPhase34Diagnostic(diagnostics []diagnostic.Diagnostic) (string, string) {
	if len(diagnostics) == 0 {
		return "", ""
	}
	return diagnostics[0].Code, diagnostics[0].Message
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
	claims = append(claims, phase34SurfaceClaim{Facet: "call:scalar_expression", Contexts: []string{"scalar CFG expressions"}, Constraints: manifest.Calls.Constraints, ScalarCalls: manifest.Calls.ScalarExpressionCalls})
	for _, entry := range manifest.Restrictions {
		claims = append(claims, phase34SurfaceClaim{Facet: "restriction:" + entry.ID, Form: entry.Syntax, Contexts: entry.Contexts, Constraints: entry.Constraints})
	}
	for _, entry := range manifest.Refusals {
		claims = append(claims, phase34SurfaceClaim{Facet: "refusal:" + entry.Code, Form: entry.Form, Code: entry.Code, Hint: entry.Hint})
	}
	return claims
}

// phase34SurfaceProblems compares only the explicit finite candidate set. A
// "*" candidate additionally includes every producer claim, so newly added
// manifest facets cannot hide outside the independently authored witness set.
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
	if containsPhase34SurfaceValue(candidates, "*") {
		for facet := range claims {
			candidates = append(candidates, facet)
		}
		for facet := range witnesses {
			candidates = append(candidates, facet)
		}
	}
	for _, facet := range candidates {
		if seen[facet] {
			continue
		}
		seen[facet] = true
		if strings.HasPrefix(facet, "probe:") {
			for _, row := range witnesses[facet] {
				if row.Status != row.Witness.ExpectedStatus || row.Code != row.Witness.ExpectedCode || row.Witness.ExpectedHint != "" && row.Hint != row.Witness.ExpectedHint {
					add("witness_outcome_changed", facet)
				}
			}
			continue
		}
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
			if strings.HasPrefix(facet, "refusal:") {
				add("missing_refusal", facet)
			} else {
				add("invented_admission", facet)
			}
			continue
		}
		var expectedContexts []string
		var expectedConstraint []string
		constraintSet := false
		for _, row := range rows {
			witness := row.Witness
			if (witness.ExpectedStatus == "pass" || witness.ExpectedCode == "check.unexecutable_shape") && witness.Context != "" {
				expectedContexts = append(expectedContexts, witness.Context)
			}
			if !constraintSet {
				expectedConstraint = witness.ExpectedConstraints
				constraintSet = true
			} else if !samePhase34Strings(expectedConstraint, witness.ExpectedConstraints) {
				add("constraint_changed", facet)
			}
			if row.Status != witness.ExpectedStatus || row.Code != witness.ExpectedCode {
				if strings.HasPrefix(facet, "refusal:") {
					add("missing_refusal", facet)
				} else {
					add("witness_outcome_changed", facet)
				}
			}
			if row.Hint != witness.ExpectedHint && witness.ExpectedHint != "" {
				add("hint_changed", facet)
			}
			if (witness.ExpectedStatus == "pass" || witness.ExpectedCode == "check.unexecutable_shape") && witness.Context != "" && !containsPhase34SurfaceValue(claim.Contexts, witness.Context) {
				add("constraint_changed", facet)
			}
			if witness.ExpectedConstraints != nil && !samePhase34Strings(claim.Constraints, witness.ExpectedConstraints) {
				add("constraint_changed", facet)
			}
			if witness.ExpectedForm != "" && claim.Form != witness.ExpectedForm {
				add("constraint_changed", facet)
			}
			arityIsPublished := strings.HasPrefix(facet, "type:") || strings.HasPrefix(facet, "builtin:") || facet == "call:declaration" || facet == "call:ordinary"
			if witness.ExpectedArity != 0 {
				switch {
				case facet == "call:ordinary":
					if !containsPhase34Int(claim.Arities, witness.ExpectedArity) {
						add("constraint_changed", facet)
					}
				case arityIsPublished && claim.Arity != witness.ExpectedArity:
					add("constraint_changed", facet)
				case strings.HasPrefix(facet, "operator:") && witness.ExpectedArity != 2:
					add("constraint_changed", facet)
				}
			}
			if witness.ExpectedInputs != nil && !samePhase34Strings(claim.Inputs, witness.ExpectedInputs) || witness.ExpectedResult != "" && claim.Result != witness.ExpectedResult {
				add("constraint_changed", facet)
			}
			if witness.ExpectedHint != "" && claim.Hint != witness.ExpectedHint {
				add("hint_changed", facet)
			}
			if witness.ExpectedCode != "" && claim.Code != witness.ExpectedCode && strings.HasPrefix(facet, "refusal:") {
				add("missing_refusal", facet)
			}
		}
		if len(expectedContexts) > 0 {
			expectedContexts = uniquePhase34Strings(expectedContexts)
			for _, expected := range expectedContexts {
				if !containsPhase34SurfaceValue(claim.Contexts, expected) {
					add("constraint_changed", facet)
				}
			}
			for _, actual := range claim.Contexts {
				if !containsPhase34SurfaceValue(expectedContexts, actual) {
					add("invented_admission", facet)
				}
			}
		}
	}
	return problems
}

func containsPhase34Int(values []int, want int) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func phase34ManifestContractProblems(manifest surfacemanifest.Manifest, expected phase34SurfaceWitnessDocument) []phase34SurfaceProblem {
	var problems []phase34SurfaceProblem
	add := func(facet string) {
		problems = append(problems, phase34SurfaceProblem{Reason: "constraint_changed", Facet: facet})
	}
	if manifest.Calls.DeclarationArity != expected.Calls.DeclarationArity || !reflect.DeepEqual(manifest.Calls.OrdinaryArity, expected.Calls.OrdinaryArity) ||
		!samePhase34Strings(manifest.Calls.OrdinaryContexts, expected.Calls.OrdinaryContexts) ||
		!samePhase34Strings(manifest.Calls.Constraints, expected.Calls.Constraints) {
		add("call:ordinary")
	}
	if manifest.Calls.ScalarExpressionCalls != expected.Calls.ScalarExpressionCalls {
		add("call:scalar_expression")
	}
	expectedLimits := make(map[string]phase34ExpectedLimit, len(expected.Limits))
	for _, limit := range expected.Limits {
		expectedLimits[limit.ID] = limit
	}
	seen := map[string]bool{}
	for _, actual := range manifest.Limits {
		facet := "limit:" + actual.ID
		want, ok := expectedLimits[actual.ID]
		if !ok {
			add(facet)
			continue
		}
		seen[actual.ID] = true
		if actual.Category != want.Category || actual.Scope != want.Scope || actual.Unit != want.Unit || actual.Value != want.Value ||
			actual.Meaning != want.Meaning || !samePhase34Strings(actual.Routes, want.Routes) {
			add(facet)
		}
	}
	for id := range expectedLimits {
		if !seen[id] {
			add("limit:" + id)
		}
	}
	return problems
}

func phase34FullSurfaceProblems(manifest surfacemanifest.Manifest, document phase34SurfaceWitnessDocument, observed []phase34ObservedSurfaceWitness) []phase34SurfaceProblem {
	problems := phase34SurfaceProblems(manifest, observed, []string{"*"})
	problems = append(problems, phase34ManifestContractProblems(manifest, document)...)
	return problems
}

func samePhase34Strings(left, right []string) bool {
	left = uniquePhase34Strings(left)
	right = uniquePhase34Strings(right)
	return reflect.DeepEqual(left, right)
}

func uniquePhase34Strings(values []string) []string {
	values = append([]string(nil), values...)
	sort.Strings(values)
	unique := values[:0]
	for _, value := range values {
		if len(unique) == 0 || unique[len(unique)-1] != value {
			unique = append(unique, value)
		}
	}
	return unique
}

func TestPhase34SurfaceManifestWitnesses(t *testing.T) {
	document := phase34ReadSurfaceWitnesses(t)
	var first *phase34SurfaceWitness
	for index := range document.Witnesses {
		if document.Witnesses[index].ID == "u64-add-scalar-operands" {
			first = &document.Witnesses[index]
			break
		}
	}
	if first == nil {
		t.Fatal("task 1's independent U64 addition witness is missing")
	}
	witness := *first
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

func TestPhase34SurfaceManifestCensus(t *testing.T) {
	witnessDocument := phase34ReadSurfaceWitnesses(t)
	census := phase34ReadCensus(t)
	if problems := phase34CensusProblems(census, nil, witnessDocument.Witnesses); len(problems) != 0 {
		t.Fatalf("source branch census has unaccounted or stale rows: %v", problems)
	}
	observed := phase34ObserveSurfaceWitnesses(t, witnessDocument.Witnesses)
	for _, row := range observed {
		if row.Status != row.Witness.ExpectedStatus || row.Code != row.Witness.ExpectedCode || row.Witness.ExpectedHint != "" && row.Hint != row.Witness.ExpectedHint {
			t.Errorf("witness %s (%s): got status=%q code=%q hint=%q; want status=%q code=%q hint=%q", row.Witness.ID, row.Witness.Facet, row.Status, row.Code, row.Hint, row.Witness.ExpectedStatus, row.Witness.ExpectedCode, row.Witness.ExpectedHint)
		}
	}
	if problems := phase34FullSurfaceProblems(surfacemanifest.Current(), witnessDocument, observed); len(problems) != 0 {
		t.Fatalf("independent witnesses disagree with the complete manifest: %v", problems)
	}
	for _, branch := range census.Branches {
		if branch.Disposition == "non-surface/internal" && branch.Reason == "" {
			t.Fatalf("internal branch lacks an explicit exclusion reason: %+v", branch)
		}
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

var phase34SurfaceCensusFiles = []string{
	"internal/compiler/syntax/parser.go",
	"internal/compiler/syntax/lexer.go",
	"internal/compiler/check/check.go",
	"internal/compiler/check/scalar.go",
	"internal/compiler/ability/ability.go",
	"internal/compiler/check/check_modules_routes.go",
	"internal/compiler/surface/refusals.go",
}

type phase34RawCensusBranch struct {
	File     string
	Function string
	Selector string
	Values   []string
}

func phase34ReadCensus(t *testing.T) phase34CensusDocument {
	t.Helper()
	path := testsupport.ProjectPath("testdata", "phase34", "surface-census.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read independent source census: %v", err)
	}
	var document phase34CensusDocument
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode independent source census: %v", err)
	}
	if document.Schema != phase34SurfaceCensusSchema || document.Boundary == "" {
		t.Fatalf("invalid finite source census header: schema=%q boundary=%q", document.Schema, document.Boundary)
	}
	if !samePhase34Strings(document.Files, phase34SurfaceCensusFiles) {
		t.Fatalf("census file boundary = %v, want exactly %v", document.Files, phase34SurfaceCensusFiles)
	}
	return document
}

func phase34GoExprLabel(expression ast.Node) string {
	var buffer bytes.Buffer
	if err := format.Node(&buffer, token.NewFileSet(), expression); err != nil {
		return fmt.Sprintf("<unformatted:%T>", expression)
	}
	text := buffer.String()
	if literal, ok := expression.(*ast.BasicLit); ok && literal.Kind == token.STRING {
		if decoded, err := strconv.Unquote(literal.Value); err == nil {
			return decoded
		}
	}
	return text
}

func phase34StringLiterals(node ast.Node) []string {
	var values []string
	ast.Inspect(node, func(inner ast.Node) bool {
		literal, ok := inner.(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		if value, err := strconv.Unquote(literal.Value); err == nil {
			values = append(values, value)
		}
		return true
	})
	return values
}

func phase34RawCensusKey(branch phase34RawCensusBranch) string {
	return branch.File + "|" + branch.Function + "|" + branch.Selector + "|" + strings.Join(branch.Values, "|")
}

func phase34UniqueRawBranches(branches []phase34RawCensusBranch) []phase34RawCensusBranch {
	unique := map[string]phase34RawCensusBranch{}
	for _, branch := range branches {
		unique[phase34RawCensusKey(branch)] = branch
	}
	result := make([]phase34RawCensusBranch, 0, len(unique))
	for _, branch := range unique {
		result = append(result, branch)
	}
	sort.Slice(result, func(i, j int) bool { return phase34RawCensusKey(result[i]) < phase34RawCensusKey(result[j]) })
	return result
}

func phase34CensusProblems(document phase34CensusDocument, overrides map[string]string, witnesses []phase34SurfaceWitness) []phase34SurfaceProblem {
	actual := phase34CollectCensusBranchesNoFatal(overrides)
	expected := map[string]phase34CensusBranch{}
	witnessIDs := map[string]bool{}
	witnessByID := map[string]phase34SurfaceWitness{}
	for _, witness := range witnesses {
		witnessIDs[witness.ID] = true
		witnessByID[witness.ID] = witness
	}
	var problems []phase34SurfaceProblem
	add := func(facet string) {
		for _, problem := range problems {
			if problem.Reason == "unaccounted_source_branch" && problem.Facet == facet {
				return
			}
		}
		problems = append(problems, phase34SurfaceProblem{Reason: "unaccounted_source_branch", Facet: facet})
	}
	for _, row := range document.Branches {
		raw := phase34RawCensusBranch{File: row.File, Function: row.Function, Selector: row.Selector, Values: row.Values}
		key := phase34RawCensusKey(raw)
		if _, duplicate := expected[key]; duplicate || row.Reason == "" || row.Facet == "" || !validPhase34Disposition(row.Disposition) {
			add("census row " + key)
		}
		expected[key] = row
		if row.Disposition != "non-surface/internal" && len(row.WitnessIDs) == 0 {
			add(row.Facet)
		}
		for _, witnessID := range row.WitnessIDs {
			if !witnessIDs[witnessID] {
				add(row.Facet + " witness " + witnessID)
			} else if row.Disposition != "non-surface/internal" && !phase34CensusFacetMatches(row.Facet, witnessByID[witnessID].Facet) {
				add(row.Facet + " links mismatched witness " + witnessID)
			}
		}
	}
	for _, branch := range actual {
		key := phase34RawCensusKey(branch)
		if _, ok := expected[key]; !ok {
			add(key)
		}
	}
	actualKeys := map[string]bool{}
	for _, branch := range actual {
		actualKeys[phase34RawCensusKey(branch)] = true
	}
	for key := range expected {
		if !actualKeys[key] {
			add(key)
		}
	}
	return problems
}

func phase34CollectCensusBranchesNoFatal(overrides map[string]string) []phase34RawCensusBranch {
	// Parse errors in a mutated control are reported as an unaccounted branch
	// by the caller, not converted into a test panic.
	var branches []phase34RawCensusBranch
	selectedFunctions := map[string]map[string]bool{
		"internal/compiler/syntax/parser.go": {
			"scalarBinaryOperator": true, "scalarBinaryExpr": true, "consumeUnsupportedBinaryOperator": true,
			"consumeUnsupportedLogicalNot": true, "scalarStatements": true, "scalarExprAtom": true, "linearBody": true,
			"refuseExtraDeclarationParameter": true,
		},
		"internal/compiler/syntax/lexer.go":               {"punctuation": true},
		"internal/compiler/check/check.go":                {"payloadFieldShape": true, "resolveCallBinding": true},
		"internal/compiler/check/scalar.go":               {"lower": true, "expression": true, "unsupportedBoolTail": true},
		"internal/compiler/ability/ability.go":            {"deriveAt": true},
		"internal/compiler/check/check_modules_routes.go": {"ModuleAwareRoutes": true},
	}
	for _, relative := range phase34SurfaceCensusFiles {
		var source []byte
		if override, ok := overrides[relative]; ok {
			source = []byte(override)
		} else {
			var err error
			source, err = os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(relative)))
			if err != nil {
				return []phase34RawCensusBranch{{File: relative, Function: "<read-error>", Selector: "source", Values: []string{err.Error()}}}
			}
		}
		file, err := parser.ParseFile(token.NewFileSet(), relative, source, parser.AllErrors)
		if err != nil {
			return []phase34RawCensusBranch{{File: relative, Function: "<parse-error>", Selector: "source", Values: []string{err.Error()}}}
		}
		functions := selectedFunctions[relative]
		ast.Inspect(file, func(node ast.Node) bool {
			declaration, ok := node.(*ast.FuncDecl)
			if !ok || declaration.Body == nil || !functions[declaration.Name.Name] {
				return true
			}
			name := declaration.Name.Name
			ast.Inspect(declaration.Body, func(inner ast.Node) bool {
				switch statement := inner.(type) {
				case *ast.SwitchStmt:
					if statement.Tag == nil {
						return true
					}
					selector := phase34GoExprLabel(statement.Tag)
					for _, rawCase := range statement.Body.List {
						clause := rawCase.(*ast.CaseClause)
						for _, value := range clause.List {
							branches = append(branches, phase34RawCensusBranch{File: relative, Function: name, Selector: selector, Values: []string{phase34GoExprLabel(value)}})
						}
					}
				case *ast.IfStmt:
					condition := phase34GoExprLabel(statement.Cond)
					if phase34RelevantCensusCondition(name, condition) {
						branches = append(branches, phase34RawCensusBranch{File: relative, Function: name, Selector: "if-condition", Values: []string{condition}})
					}
				case *ast.ReturnStmt:
					if name == "ModuleAwareRoutes" {
						for _, result := range statement.Results {
							literalValues := phase34StringLiterals(result)
							if len(literalValues) > 0 {
								branches = append(branches, phase34RawCensusBranch{File: relative, Function: name, Selector: "closed-route-list", Values: literalValues})
							}
							break
						}
					}
				}
				return true
			})
			if name == "resolveCallBinding" {
				ast.Inspect(declaration.Body, func(inner ast.Node) bool {
					literal, ok := inner.(*ast.BasicLit)
					if ok && literal.Kind == token.STRING && phase34GoExprLabel(literal) == "check.call_arity_unsupported" {
						branches = append(branches, phase34RawCensusBranch{File: relative, Function: name, Selector: "diagnostic-code", Values: []string{"check.call_arity_unsupported"}})
					}
					return true
				})
			}
			if relative == "internal/compiler/syntax/parser.go" || relative == "internal/compiler/check/scalar.go" {
				codes := map[string]bool{}
				ast.Inspect(declaration.Body, func(inner ast.Node) bool {
					literal, ok := inner.(*ast.BasicLit)
					if ok && literal.Kind == token.STRING {
						value := phase34GoExprLabel(literal)
						if strings.HasPrefix(value, "surface.unsupported_") {
							codes[value] = true
						}
					}
					return true
				})
				for code := range codes {
					branches = append(branches, phase34RawCensusBranch{File: relative, Function: "surface-refusal-emitter", Selector: "code-literal", Values: []string{code}})
				}
			}
			return false
		})
		if relative == "internal/compiler/surface/refusals.go" {
			codes := map[string]bool{}
			ast.Inspect(file, func(node ast.Node) bool {
				literal, ok := node.(*ast.BasicLit)
				if ok && literal.Kind == token.STRING {
					value := phase34GoExprLabel(literal)
					if strings.HasPrefix(value, "surface.unsupported_") {
						codes[value] = true
					}
				}
				return true
			})
			for code := range codes {
				branches = append(branches, phase34RawCensusBranch{File: relative, Function: "refusal-catalog", Selector: "code-literal", Values: []string{code}})
			}
		}
	}
	return phase34UniqueRawBranches(branches)
}

func validPhase34Disposition(value string) bool {
	return value == "admitted-with-context" || value == "refusal" || value == "non-surface/internal"
}

func phase34RelevantCensusCondition(function, condition string) bool {
	markers := map[string][]string{
		"linearBody":                       {"TokenDiscard", "TokenTry", "TokenTake", "TokenBorrow", "TokenMut", "TokenMatch"},
		"refuseExtraDeclarationParameter":  {"TokenComma"},
		"consumeUnsupportedBinaryOperator": {"TokenUnknown", "TokenPipe"},
		"resolveCallBinding":               {"len(binding.RHS.Arguments)"},
		"deriveAt":                         {"sealedLeaves"},
		"unsupportedBoolTail":              {"resultName"},
	}
	for _, marker := range markers[function] {
		if strings.Contains(condition, marker) {
			return true
		}
	}
	return false
}

func phase34CensusFacetMatches(facets, candidate string) bool {
	for _, facet := range strings.Split(facets, "|") {
		if facet == candidate {
			return true
		}
	}
	return false
}
