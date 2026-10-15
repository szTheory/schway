package session_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
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
	SourcePath          string                      `json:"source_path"`
	Signature           string                      `json:"signature"`
	Source              string                      `json:"source"`
	GeneratedCallDepth  int                         `json:"generated_call_depth"`
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
	Schema           string                      `json:"schema"`
	Boundary         string                      `json:"boundary"`
	Calls            phase34ExpectedCalls        `json:"calls"`
	Limits           []phase34ExpectedLimit      `json:"limits"`
	SupersededLimits []phase34LimitSupersession  `json:"superseded_limits"`
	Supersedes       []phase34ScopedSupersession `json:"supersedes"`
	Witnesses        []phase36SurfaceWitness     `json:"witnesses"`
}

type phase34LimitSupersession struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
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
	Schema        string                      `json:"schema"`
	Boundary      string                      `json:"boundary"`
	Files         []string                    `json:"files"`
	Branches      []phase34CensusBranch       `json:"branches"`
	LimitContexts []phase36LimitContext       `json:"limit_contexts"`
	Supersedes    []phase34CensusSupersession `json:"supersedes"`
}

type phase34CensusSupersession struct {
	Facet                 string   `json:"facet"`
	PriorBranchKeys       []string `json:"prior_branch_keys"`
	ReplacementBranchKeys []string `json:"replacement_branch_keys"`
	Reason                string   `json:"reason"`
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
	census := phase34ReadCensus(t)
	if problems := phase34CensusProblems(census, nil, witnessDocument.Witnesses); len(problems) != 0 {
		t.Fatalf("composed Phase 34/35/36 source branch census differs from selected producers; expected Phase 36 call census IDs to be accounted: %v", problems)
	}
	phase36Census := phase36ReadSurfaceCensus(t)
	wantSupersessions := map[string]bool{
		"type:bool|internal/compiler/ability/ability.go|deriveAt|shape.Constructor|Bool|internal/compiler/ability/ability.go|deriveAt|shape.Constructor|Bool": true,
		"restriction:scalar_expression_grouping|internal/compiler/syntax/parser.go|scalarExprAtom|leftToken.Kind|TokenLParen|internal/compiler/syntax/parser.go|scalarExprAtom|leftToken.Kind|TokenLParen": true,
		"call:ordinary|internal/compiler/check/check.go|resolveCallBinding|diagnostic-code|check.call_arity_unsupported,internal/compiler/check/check.go|resolveCallBinding|if-condition|len(binding.RHS.Arguments) != 1|internal/compiler/check/check.go|resolveCallBinding|diagnostic-code|check.call_arity_unsupported,internal/compiler/check/check.go|resolveCallBinding|if-condition|len(binding.RHS.Arguments) != 1": true,
		"refusal:surface.unsupported_expression_depth|internal/compiler/syntax/parser.go|nestedScalarExpr|if-condition|p.expressionDepth >= MaxExpressionDepth|internal/compiler/syntax/parser.go|nestedScalarExpr|if-condition|p.expressionDepth >= MaxExpressionDepth": true,
	}
	for _, row := range phase36Census.Supersedes {
		key := row.Facet + "|" + strings.Join(row.PriorBranchKeys, ",") + "|" + strings.Join(row.ReplacementBranchKeys, ",")
		if !wantSupersessions[key] || row.Reason == "" {
			t.Errorf("unexpected or unreasoned exact Phase 36 census supersession %q", key)
		}
		delete(wantSupersessions, key)
	}
	if len(wantSupersessions) != 0 {
		t.Errorf("Phase 36 census omitted exact changed-claim branch supersessions: %v", wantSupersessions)
	}
	for _, row := range phase36Census.Branches {
		if row.Disposition == "non-surface/internal" && row.Reason == "" {
			t.Fatalf("Phase 36 internal branch lacks a named exclusion: %+v", row)
		}
	}
	required := map[string]string{
		"internal/compiler/check/scalar.go|expression|expr.Kind|call":                                                                      "restriction:scalar_expression_calls",
		"internal/compiler/check/scalar.go|lowerScalarCall|if-condition|!ok || contract.ID == \"\"":                                        "probe:phase36-unresolved-callee",
		"internal/compiler/syntax/parser.go|scalarCall|if-condition|p.peek().Kind == TokenRParen":                                          "restriction:scalar_expression_calls",
		"internal/compiler/syntax/parser.go|scalarCall|if-condition|p.peek().Kind == TokenComma":                                           "probe:phase36-scalar-call-arity",
		"internal/compiler/syntax/parser.go|scalarExprAtom|if-condition|leftToken.Kind == TokenIdentifier && p.peek().Kind == TokenLParen": "restriction:scalar_expression_calls",
		"internal/compiler/syntax/parser.go|scalarLetStart|if-condition|value.Kind == TokenIdentifier && next.Kind == TokenLParen":         "restriction:scalar_expression_calls",
		"internal/compiler/syntax/parser.go|nestedScalarExpr|if-condition|p.expressionDepth >= MaxExpressionDepth":                         "refusal:surface.unsupported_expression_depth",
	}
	for _, row := range census.Branches {
		key := phase34RawCensusKey(phase34RawCensusBranch{File: row.File, Function: row.Function, Selector: row.Selector, Values: row.Values})
		if facet, ok := required[key]; ok {
			if !phase34CensusFacetMatches(row.Facet, facet) {
				t.Errorf("required call census branch %s has facet %q; want %q", key, row.Facet, facet)
			}
			delete(required, key)
		}
	}
	if len(required) != 0 {
		t.Errorf("Phase 36 call census omits required source branches: %v", required)
	}
}

func TestPhase36SurfaceManifestCallControls(t *testing.T) {
	document := phase34ReadSurfaceWitnesses(t)
	phase36Document := phase36ReadSurfaceWitnesses(t)
	observed := phase34ObserveSurfaceWitnesses(t, document.Witnesses)
	manifest := surfacemanifest.Current()
	if document.Calls.DeclarationArity != 1 || !samePhase34Ints(document.Calls.OrdinaryArity, []int{1}) || !document.Calls.ScalarExpressionCalls {
		t.Fatalf("Phase 36 literal call controls do not describe one-argument scalar calls: %+v", document.Calls)
	}
	wantSupersessions := map[string]bool{
		"type:bool|bool-local-state,bool-comparison-results|phase36-bool-local-state,phase36-bool-comparison-results,phase36-bool-function-parameter-and-result":                                              true,
		"restriction:scalar_expression_calls|restriction-scalar-expressions|phase36-pair-u64-u64":                                                                                                             true,
		"restriction:scalar_expression_grouping|phase35-grouping|phase36-grouping-with-helper-call":                                                                                                           true,
		"call:declaration|call-declaration-arity-one|phase36-call-declaration-arity-one":                                                                                                                      true,
		"call:ordinary|call-linear-binding-arity-one,call-match-arm-arity-one,call-ordinary-arity-two|phase36-call-linear-binding-arity-one,phase36-call-match-arm-arity-one,phase36-call-ordinary-arity-two": true,
		"call:scalar_expression|call-scalar-context|phase36-call-scalar-expression":                                                                                                                           true,
		"probe:scalar-helper-call|probe-scalar-helper-call|phase36-probe-scalar-helper-call":                                                                                                                  true,
	}
	for _, row := range phase36Document.Supersedes {
		key := row.Facet + "|" + strings.Join(row.PriorWitnessIDs, ",") + "|" + strings.Join(row.ReplacementWitnessIDs, ",")
		if !wantSupersessions[key] || row.Reason == "" {
			t.Errorf("unexpected or unreasoned exact witness supersession %q", key)
		}
		delete(wantSupersessions, key)
	}
	if len(wantSupersessions) != 0 {
		t.Errorf("Phase 36 omitted exact inherited witness supersessions: %v", wantSupersessions)
	}
	if problems := phase34FullSurfaceProblems(manifest, document, observed); len(problems) != 0 {
		for _, row := range observed {
			if row.Status != row.Witness.ExpectedStatus || row.Code != row.Witness.ExpectedCode {
				t.Logf("baseline witness mismatch %s/%s: got %s/%s want %s/%s", row.Witness.ID, row.Witness.Facet, row.Status, row.Code, row.Witness.ExpectedStatus, row.Witness.ExpectedCode)
			}
		}
		t.Fatalf("control baseline must be clean before mutations: %v", problems)
	}
	census := phase34ReadCensus(t)
	if problems := phase34CensusProblems(census, nil, document.Witnesses); len(problems) != 0 {
		t.Fatalf("census control baseline must be clean before mutations: %v", problems)
	}
	assert := func(name, reason, facet string, problems []phase34SurfaceProblem) {
		t.Helper()
		if len(problems) != 1 || problems[0] != (phase34SurfaceProblem{Reason: reason, Facet: facet}) {
			t.Errorf("%s = %v, want the named finding %s for %s", name, problems, reason, facet)
		}
	}

	withoutCalls := phase34CopyManifest(t, manifest)
	withoutCalls.Restrictions = phase34FilterEntries(withoutCalls.Restrictions, "scalar_expression_calls")
	assert("remove witnessed scalar-call admission", "missing_admission", "restriction:scalar_expression_calls", phase34SurfaceProblems(withoutCalls, observed, []string{"restriction:scalar_expression_calls"}))

	invented := phase34CopyManifest(t, manifest)
	invented.Restrictions = append(invented.Restrictions, surfacemanifest.Entry{ID: "scalar_indirect_calls", Syntax: "value(expression)", Contexts: []string{"arbitrary indirect calls"}, Constraints: []string{"invented admission"}})
	assert("invent an unwitnessed indirect-call admission", "invented_admission", "restriction:scalar_indirect_calls", phase34SurfaceProblems(invented, observed, []string{"restriction:scalar_indirect_calls"}))

	withoutDepthRefusal := phase34CopyManifest(t, manifest)
	withoutDepthRefusal.Refusals = phase34FilterRefusals(withoutDepthRefusal.Refusals, "surface.unsupported_expression_depth")
	assert("remove the required depth refusal", "missing_refusal", "refusal:surface.unsupported_expression_depth", phase34SurfaceProblems(withoutDepthRefusal, observed, []string{"refusal:surface.unsupported_expression_depth"}))

	wrongHint := phase34CopyManifest(t, manifest)
	phase34SetRefusalHint(wrongHint.Refusals, "surface.unsupported_expression_depth", "mutated depth hint")
	assert("change the required depth hint", "hint_changed", "refusal:surface.unsupported_expression_depth", phase34SurfaceProblems(wrongHint, observed, []string{"refusal:surface.unsupported_expression_depth"}))

	falseRefusalWitnesses := append([]phase34ObservedSurfaceWitness(nil), observed...)
	for index := range falseRefusalWitnesses {
		if falseRefusalWitnesses[index].Witness.ID == "phase36-pair-u64-u64" {
			falseRefusalWitnesses[index].Witness.ExpectedStatus = "invalid"
			falseRefusalWitnesses[index].Witness.ExpectedCode = "check.call_arity_unsupported"
		}
	}
	assert("label an accepted complete scalar call as refused", "witness_outcome_changed", "restriction:scalar_expression_calls", phase34SurfaceProblems(manifest, falseRefusalWitnesses, []string{"restriction:scalar_expression_calls"}))

	missingConstraint := phase34CopyManifest(t, manifest)
	phase34SetStatementConstraints(missingConstraint.Restrictions, "scalar_expression_calls", nil)
	assert("remove the exact scalar-call restriction", "constraint_changed", "restriction:scalar_expression_calls", phase34SurfaceProblems(missingConstraint, observed, []string{"restriction:scalar_expression_calls"}))

	widened := phase34CopyManifest(t, manifest)
	phase34SetStatementConstraints(widened.Restrictions, "scalar_expression_calls", []string{"Any callable value and arbitrary expression are admitted."})
	assert("widen scalar-call restrictions", "constraint_changed", "restriction:scalar_expression_calls", phase34SurfaceProblems(widened, observed, []string{"restriction:scalar_expression_calls"}))

	withoutCallBranch := census
	withoutCallBranch.Branches = filterSurfaceCensusBranches(census.Branches, "internal/compiler/check/scalar.go", "expression", "call")
	assert("omit the scalar checker call branch", "unaccounted_source_branch", "internal/compiler/check/scalar.go|expression|expr.Kind|call", phase34CensusProblems(withoutCallBranch, nil, document.Witnesses))

	parserPath := "internal/compiler/syntax/parser.go"
	parserSource, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(parserPath)))
	if err != nil {
		t.Fatal(err)
	}
	parserMutation := phase34ReplaceOnce(t, string(parserSource), "func (p *parser) scalarCall(callee Token) *ast.Expr {", "func (p *parser) scalarCall(callee Token) *ast.Expr {\n\tif p.peek().Kind == TokenRBrace { return nil }")
	assert("add an unaccounted parser branch", "unaccounted_source_branch", parserPath+"|scalarCall|if-condition|p.peek().Kind == TokenRBrace", phase34CensusProblems(census, map[string]string{parserPath: parserMutation}, document.Witnesses))

	graph := phase34LimitConsumerGraph(t)
	parserRoutes := make([]string, 0, 13)
	for route, root := range map[string]string{
		"app verify": "session.VerifyApplicationCasesFile", "build": "session.BuildApplicationCommandFile", "check": "session.CheckCommandFile", "debug-map": "session.DebugMapCommandFile", "evidence": "session.EvidenceCommandFile", "evidence --validate": "session.ValidateEvidenceCommandFile", "explain": "session.ExplainCommandFile", "format": "session.FormatCommandFile", "interface core": "session.InterfaceCoreCommandFile", "interface export": "session.InterfaceExportCommandFile", "query": "session.QueryCommandFile", "run": "session.RunInterpreterCommandFile", "verify": "session.VerifyCorpusFile",
	} {
		if graph[root] == nil {
			t.Fatalf("consumer census lost command entry %s", root)
		}
		if phase34ConsumerReaches(graph, root, "syntax.Parse", map[string]bool{}) {
			parserRoutes = append(parserRoutes, route)
		}
	}
	sort.Strings(parserRoutes)
	expectedRoutes := map[string][]string{"parser.expression_depth": parserRoutes}
	missingRoute := phase34CopyManifest(t, manifest)
	phase36EditLimitRoutes(missingRoute.Limits, "parser.expression_depth", "run", false)
	if !phase34ContainsProblem(phase34LimitConsumerProblems(missingRoute, expectedRoutes), "missing_consumer_route parser.expression_depth run") {
		t.Error("missing parser.expression_depth route control did not name run")
	}
	extraRoute := phase34CopyManifest(t, manifest)
	phase36EditLimitRoutes(extraRoute.Limits, "parser.expression_depth", "app run", true)
	if !phase34ContainsProblem(phase34LimitConsumerProblems(extraRoute, expectedRoutes), "unsupported_consumer_route parser.expression_depth app run") {
		t.Error("extra parser.expression_depth route control did not name app run")
	}
	if problems := phase34FullSurfaceProblems(manifest, document, observed); len(problems) != 0 {
		t.Fatalf("restored positive surface comparison is not clean: %v", problems)
	}
	if problems := phase34CensusProblems(census, nil, document.Witnesses); len(problems) != 0 {
		t.Fatalf("restored positive census comparison is not clean: %v", problems)
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
		parts := strings.SplitN(row.Source, ":", 2)
		if len(parts) != 2 {
			t.Errorf("limit context %s source binding %q must name path:literal", row.ID, row.Source)
			continue
		}
		data, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(parts[0])))
		if err != nil || !strings.Contains(string(data), parts[1]) {
			t.Errorf("limit context %s source witness %q is absent (read error %v)", row.ID, row.Source, err)
		}
		delete(want, row.ID)
	}
	if len(want) != 0 {
		t.Errorf("missing exact Phase 36 limit contexts: %v", want)
	}
	graph := phase34LimitConsumerGraph(t)
	roots := map[string]string{
		"app verify": "session.VerifyApplicationCasesFile", "build": "session.BuildApplicationCommandFile", "check": "session.CheckCommandFile", "debug-map": "session.DebugMapCommandFile", "evidence": "session.EvidenceCommandFile", "evidence --validate": "session.ValidateEvidenceCommandFile", "explain": "session.ExplainCommandFile", "format": "session.FormatCommandFile", "interface core": "session.InterfaceCoreCommandFile", "interface export": "session.InterfaceExportCommandFile", "query": "session.QueryCommandFile", "run": "session.RunInterpreterCommandFile", "verify": "session.VerifyCorpusFile",
	}
	var routes []string
	for route, root := range roots {
		if graph[root] == nil {
			t.Fatalf("Phase 36 limit census lost command entry %s", root)
		}
		if phase34ConsumerReaches(graph, root, "syntax.Parse", map[string]bool{}) {
			routes = append(routes, route)
		}
	}
	sort.Strings(routes)
	if len(routes) != 13 {
		t.Fatalf("64-level source expression parser routes = %v, want all 13 registered routes", routes)
	}
	expected := map[string][]string{"parser.expression_depth": routes}
	if problems := phase34LimitConsumerProblems(surfacemanifest.Current(), expected); len(problems) != 0 {
		t.Fatalf("64-level source expression routes differ from the independent consumer census: %v", problems)
	}
	for _, row := range document.LimitContexts {
		if row.ID == "application.evidence_bytes" && strings.Contains(strings.ToLower(row.Meaning), "iteration limit") {
			t.Fatal("evidence byte capture is mislabeled as a semantic iteration bound")
		}
	}
}

func phase36EditLimitRoutes(limits []surfacemanifest.Limit, id, route string, add bool) {
	for index := range limits {
		if limits[index].ID != id {
			continue
		}
		if add {
			limits[index].Routes = append(limits[index].Routes, route)
			return
		}
		filtered := limits[index].Routes[:0]
		for _, existing := range limits[index].Routes {
			if existing != route {
				filtered = append(filtered, existing)
			}
		}
		limits[index].Routes = filtered
		return
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
		"phase36-refuse-arity-two":              "check.call_arity_unsupported",
		"phase36-refuse-scalar-call-arity":      "syntax.scalar_call_arity",
		"phase36-refuse-argument-type":          "check.call_argument_type_mismatch",
		"phase36-refuse-unresolved-scalar-call": "core.call_callee_unresolved",
		"phase36-refuse-private-import":         "module.not_exported",
		"phase36-refuse-recursion":              "core.call_graph_cycle",
		"phase36-refuse-bool-literal-tail":      "surface.unsupported_bool_literal_tail",
		"phase36-refuse-byteview-signature":     "type.scalar_signature",
	}
	wantContextCases := map[string]struct{ context, exercise string }{
		"phase36-assignment-rhs":                  {"scalar binding and assignment RHS", "result = adjust(n)"},
		"phase36-arithmetic-operand-and-grouping": {"scalar arithmetic and comparison operands", "n + (adjust(n))"},
		"phase36-pair-u64-bool":                   {"scalar branch and loop conditions", "if is_small(n)"},
		"phase36-loop-condition":                  {"scalar branch and loop conditions", "while below(i)"},
		"phase36-byte-index":                      {"scalar byte_at index expressions", "byte_at(bytes, offset(0))"},
		"phase36-nested-argument":                 {"nested scalar helper arguments", "outer(inner(n))"},
	}
	wantPairDeclarations := map[string]string{
		"phase36-pair-u64-u64":   "fn adjust(value: U64) -> U64",
		"phase36-pair-u64-bool":  "fn is_small(value: U64) -> Bool",
		"phase36-pair-bool-u64":  "fn choose(value: Bool) -> U64",
		"phase36-pair-bool-bool": "fn relay(value: Bool) -> Bool",
	}
	seenIDs := map[string]bool{}
	var observedInput []phase34SurfaceWitness
	for _, witness := range document.Witnesses {
		if witness.ID == "" || seenIDs[witness.ID] {
			t.Fatalf("duplicate or empty Phase 36 witness ID %q", witness.ID)
		}
		seenIDs[witness.ID] = true
		if pair, ok := wantPairs[witness.ID]; ok {
			if witness.Facet != "restriction:scalar_expression_calls" || witness.Signature != pair || witness.ExpectedStatus != "pass" || witness.ExpectedArity != 1 {
				t.Errorf("signature witness %s = facet %q pair %q status %q; want %q and a passing complete program", witness.ID, witness.Facet, witness.Signature, witness.ExpectedStatus, pair)
			}
			if declaration := wantPairDeclarations[witness.ID]; declaration == "" || !strings.Contains(witness.Source, declaration) {
				t.Errorf("signature witness %s source does not bind its literal pair declaration %q", witness.ID, declaration)
			}
			delete(wantPairDeclarations, witness.ID)
			delete(wantPairs, witness.ID)
		}
		if expected, ok := wantContextCases[witness.ID]; ok {
			if witness.Context != expected.context || witness.Exercise != expected.exercise || !strings.Contains(witness.Source, expected.exercise) {
				t.Errorf("context witness %s = context %q exercise %q; want literal context %q and source exercise %q", witness.ID, witness.Context, witness.Exercise, expected.context, expected.exercise)
			}
			delete(wantContextCases, witness.ID)
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
			ID: witness.ID, Facet: witness.Facet, Context: witness.Context, CheckRoute: witness.CheckRoute, SourcePath: witness.SourcePath, Files: witness.Files,
			Source: phase36WitnessSource(witness), ExpectedStatus: witness.ExpectedStatus, ExpectedCode: witness.ExpectedCode,
			ExpectedHint: witness.ExpectedHint, Exercise: witness.Exercise, ExpectedForm: witness.ExpectedForm,
			ExpectedConstraints: witness.ExpectedConstraints,
		})
	}
	if len(wantPairs) != 0 {
		t.Errorf("missing exact scalar signature pair witnesses: %v", wantPairs)
	}
	if len(wantPairDeclarations) != 0 {
		t.Errorf("pair declaration controls were not resolved: %v", wantPairDeclarations)
	}
	if len(wantContextCases) != 0 {
		t.Errorf("missing source-bound scalar call context witnesses: %v", wantContextCases)
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

func phase36WitnessSource(witness phase36SurfaceWitness) string {
	if witness.GeneratedCallDepth <= 0 {
		return witness.Source
	}
	call := strings.Repeat("identity(", witness.GeneratedCallDepth) + "n" + strings.Repeat(")", witness.GeneratedCallDepth)
	return "module phase36.call_depth\nexport { fn main }\nfn identity(value: U64) -> U64 {\n  let result = value\n  result\n}\nfn main(n: U64) -> U64 {\n  let result = " + call + "\n  result\n}\n"
}
