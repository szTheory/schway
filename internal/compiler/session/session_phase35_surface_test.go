package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/surfacemanifest"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase35SurfaceWitnessSchema = "phase35.surface-witnesses/1"
const phase35SurfaceCensusSchema = "phase35.surface-census/1"

func phase35ReadSurfaceWitnessAdditions(t *testing.T) phase34SurfaceWitnessDocument {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase35", "surface-witnesses.json"))
	if err != nil {
		t.Fatalf("read additive Phase 35 surface witnesses: %v", err)
	}
	var document phase34SurfaceWitnessDocument
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode additive Phase 35 surface witnesses: %v", err)
	}
	if document.Schema != phase35SurfaceWitnessSchema || document.Boundary == "" {
		t.Fatalf("invalid Phase 35 witness header: schema=%q boundary=%q", document.Schema, document.Boundary)
	}
	return document
}

func phase35ReadSurfaceCensusAdditions(t *testing.T) phase34CensusDocument {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase35", "surface-census.json"))
	if err != nil {
		t.Fatalf("read additive Phase 35 source census: %v", err)
	}
	var document phase34CensusDocument
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode additive Phase 35 source census: %v", err)
	}
	if document.Schema != phase35SurfaceCensusSchema || document.Boundary == "" {
		t.Fatalf("invalid Phase 35 census header: schema=%q boundary=%q", document.Schema, document.Boundary)
	}
	if !samePhase34Strings(document.Files, phase34SurfaceCensusFiles) {
		t.Fatalf("Phase 35 census file boundary = %v, want exactly %v", document.Files, phase34SurfaceCensusFiles)
	}
	return document
}

func TestPhase35SurfaceManifestWitnesses(t *testing.T) {
	phase35AssertImmutablePhase34Fixtures(t)
	document := phase34ReadSurfaceWitnesses(t)
	required := map[string]bool{
		"phase35-bit-and": false, "phase35-bit-or": false, "phase35-bit-xor": false,
		"phase35-shift-left-zero": false, "phase35-shift-right-logical": false,
		"phase35-precedence": false, "phase35-hex-spelling": false,
		"phase35-byte-unused": false, "phase35-depth-64": false, "phase35-depth-65": false,
		"phase35-nested-byte-index": false, "phase35-type-closers": false,
		"phase35-bool-operand": false, "phase35-shift-nonliteral": false,
		"phase35-shift-negative": false, "phase35-shift-64": false, "phase35-shift-empty": false,
		"phase35-hex-maximum-leading-zeroes": false, "phase35-hex-overflow": false, "phase35-hex-malformed": false,
		"phase35-byte-used": false,
	}
	for _, witness := range document.Witnesses {
		if _, ok := required[witness.ID]; ok {
			required[witness.ID] = true
		}
	}
	for id, found := range required {
		if !found {
			t.Errorf("independent Phase 35 witness %q is missing", id)
		}
	}
	if !containsPhase35Witness(document.Witnesses, "phase36-grouping-with-helper-call") {
		t.Error("composed grouping claim lacks its exact Phase 36 helper-call witness")
	}
	observed := phase34ObserveSurfaceWitnesses(t, document.Witnesses)
	if problems := phase34FullSurfaceProblems(surfacemanifest.Current(), document, observed); len(problems) != 0 {
		t.Fatalf("independent Phase 35 witnesses disagree with the complete source-free manifest: %v", problems)
	}
	if !document.Calls.ScalarExpressionCalls {
		t.Fatal("Phase 36 composition must replace the inherited scalar expression helper-call absence claim")
	}
}

func containsPhase35Witness(witnesses []phase34SurfaceWitness, id string) bool {
	for _, witness := range witnesses {
		if witness.ID == id {
			return true
		}
	}
	return false
}

func TestPhase35SurfaceManifestCensus(t *testing.T) {
	phase35AssertImmutablePhase34Fixtures(t)
	document := phase34ReadSurfaceWitnesses(t)
	census := phase34ReadCensus(t)
	if problems := phase34CensusProblems(census, nil, document.Witnesses); len(problems) != 0 {
		t.Fatalf("composed Phase 34/35 source branch census differs from selected producers: %v", problems)
	}
	for _, row := range census.Branches {
		if row.Disposition == "non-surface/internal" && row.Reason == "" {
			t.Fatalf("internal source branch lacks a named exclusion: %+v", row)
		}
	}
	phase35Required := map[string]bool{
		"operator:u64.bit_and": false, "operator:u64.bit_or": false, "operator:u64.bit_xor": false,
		"operator:u64.shift_left": false, "operator:u64.shift_right": false,
		"restriction:u64_hex_spelling": false, "refusal:surface.unsupported_shift_amount": false,
		"refusal:surface.unsupported_expression_depth": false,
	}
	for _, row := range census.Branches {
		for _, facet := range stringsSplitPhase35Facets(row.Facet) {
			if _, ok := phase35Required[facet]; ok {
				phase35Required[facet] = true
			}
		}
	}
	for facet, found := range phase35Required {
		if !found {
			t.Errorf("Phase 35 source census does not reach required producer/refusal %q", facet)
		}
	}
}

func TestPhase35SurfaceManifestControls(t *testing.T) {
	document := phase34ReadSurfaceWitnesses(t)
	census := phase34ReadCensus(t)
	manifest := surfacemanifest.Current()
	observed := phase34ObserveSurfaceWitnesses(t, document.Witnesses)
	if problems := phase34FullSurfaceProblems(manifest, document, observed); len(problems) != 0 {
		t.Fatalf("control baseline must be clean before Phase 35 mutations: %v", problems)
	}
	if problems := phase34CensusProblems(census, nil, document.Witnesses); len(problems) != 0 {
		t.Fatalf("census control baseline must be clean before Phase 35 mutations: %v", problems)
	}
	assert := func(name, reason, facet string, problems []phase34SurfaceProblem) {
		t.Helper()
		unique := map[phase34SurfaceProblem]bool{}
		for _, problem := range problems {
			unique[problem] = true
		}
		if len(unique) != 1 || !unique[phase34SurfaceProblem{Reason: reason, Facet: facet}] {
			t.Errorf("%s = %v, want only distinct finding %s for %s", name, problems, reason, facet)
		}
	}

	withoutXor := phase34CopyManifest(t, manifest)
	withoutXor.Operators = phase34FilterEntries(withoutXor.Operators, "u64.bit_xor")
	assert("remove witnessed XOR admission", "missing_admission", "operator:u64.bit_xor", phase34FullSurfaceProblems(withoutXor, document, observed))

	falseAdmission := phase34CopyManifest(t, manifest)
	falseAdmission.Operators = append(falseAdmission.Operators, surfacemanifest.Entry{ID: "u64.logical_and", Syntax: "&&", Contexts: []string{"scalar Bool operands"}, Constraints: []string{"Invented Phase 35 claim."}})
	assert("invent unwitnessed logical operator", "invented_admission", "operator:u64.logical_and", phase34FullSurfaceProblems(falseAdmission, document, observed))

	withoutShiftRefusal := phase34CopyManifest(t, manifest)
	withoutShiftRefusal.Refusals = phase34FilterRefusals(withoutShiftRefusal.Refusals, "surface.unsupported_shift_amount")
	assert("remove shift amount refusal", "missing_refusal", "refusal:surface.unsupported_shift_amount", phase34FullSurfaceProblems(withoutShiftRefusal, document, observed))
	wrongShiftHint := phase34CopyManifest(t, manifest)
	phase34SetRefusalHint(wrongShiftHint.Refusals, "surface.unsupported_shift_amount", "mutated shift hint")
	assert("change shift amount refusal hint", "hint_changed", "refusal:surface.unsupported_shift_amount", phase34FullSurfaceProblems(wrongShiftHint, document, observed))

	for _, mutation := range []struct {
		name, facet, id string
		edit            func([]string) []string
	}{
		{"drop literal shift rule", "operator:u64.shift_left", "u64.shift_left", func([]string) []string { return nil }},
		{"drop low64 shift rule", "operator:u64.shift_left", "u64.shift_left", func([]string) []string { return []string{"literal counts are bounded"} }},
		{"drop hex value rule", "restriction:u64_hex_spelling", "u64_hex_spelling", func([]string) []string { return nil }},
	} {
		mutated := phase34CopyManifest(t, manifest)
		if strings.HasPrefix(mutation.facet, "operator:") {
			for index := range mutated.Operators {
				if mutated.Operators[index].ID == mutation.id {
					mutated.Operators[index].Constraints = mutation.edit(mutated.Operators[index].Constraints)
				}
			}
		} else {
			phase34SetStatementConstraints(mutated.Restrictions, mutation.id, mutation.edit(nil))
		}
		assert(mutation.name, "constraint_changed", mutation.facet, phase34FullSurfaceProblems(mutated, document, observed))
	}

	withoutDepthLimit := phase34CopyManifest(t, manifest)
	withoutDepthLimit.Limits = phase34FilterLimits(withoutDepthLimit.Limits, "parser.expression_depth")
	assert("remove nested-expression depth limit", "constraint_changed", "limit:parser.expression_depth", phase34ManifestContractProblems(withoutDepthLimit, document))

	withoutDepthWitness := document
	withoutDepthWitness.Witnesses = nil
	removedDepthWitnesses := 0
	for _, witness := range document.Witnesses {
		// Composed helper-call depth witnesses cover this same refusal.
		// Remove its complete copied-data coverage to exercise missing_refusal.
		if witness.Facet == "refusal:surface.unsupported_expression_depth" {
			removedDepthWitnesses++
			continue
		}
		withoutDepthWitness.Witnesses = append(withoutDepthWitness.Witnesses, witness)
	}
	if removedDepthWitnesses == 0 {
		t.Fatal("depth-refusal control has no active full-source witness to remove")
	}
	withoutDepthObserved := phase34ObserveSurfaceWitnesses(t, withoutDepthWitness.Witnesses)
	assert("remove over-depth full-source witness", "missing_refusal", "refusal:surface.unsupported_expression_depth", phase34SurfaceProblems(manifest, withoutDepthObserved, []string{"refusal:surface.unsupported_expression_depth"}))
	if problems := phase34SurfaceProblems(manifest, observed, []string{"refusal:surface.unsupported_expression_depth"}); len(problems) != 0 {
		t.Fatalf("restored over-depth full-source witnesses must cover the refusal: %v", problems)
	}

	parserPath := "internal/compiler/syntax/parser.go"
	parserSource, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(parserPath)))
	if err != nil {
		t.Fatal(err)
	}
	operatorMutation := phase34ReplaceOnce(t, string(parserSource), "case TokenPercent:\n\t\treturn 7", "case TokenStar:\n\t\treturn 8\n\tcase TokenPercent:\n\t\treturn 7")
	assert("add an unaccounted Phase 35 precedence branch", "unaccounted_source_branch", parserPath+"|scalarBinaryOperator|kind|TokenStar", phase34CensusProblems(census, map[string]string{parserPath: operatorMutation}, document.Witnesses))

	withoutGroupingBranch := census
	withoutGroupingBranch.Branches = filterSurfaceCensusBranches(census.Branches, parserPath, "scalarExprAtom", "TokenLParen")
	assert("remove an accounted Phase 35 grouping branch", "unaccounted_source_branch", parserPath+"|scalarExprAtom|leftToken.Kind|TokenLParen", phase34CensusProblems(withoutGroupingBranch, nil, document.Witnesses))

	if problems := phase34FullSurfaceProblems(manifest, document, observed); len(problems) != 0 {
		t.Fatalf("copied manifest controls did not restore: %v", problems)
	}
	if problems := phase34CensusProblems(census, nil, document.Witnesses); len(problems) != 0 {
		t.Fatalf("copied census controls did not restore: %v", problems)
	}
}

func TestPhase35SurfaceLimitConsumers(t *testing.T) {
	graph := phase34LimitConsumerGraph(t)
	roots := map[string]string{
		"app verify": "session.VerifyApplicationCasesFile", "build": "session.BuildApplicationCommandFile",
		"check": "session.CheckCommandFile", "debug-map": "session.DebugMapCommandFile",
		"evidence": "session.EvidenceCommandFile", "evidence --validate": "session.ValidateEvidenceCommandFile",
		"explain": "session.ExplainCommandFile", "format": "session.FormatCommandFile",
		"interface core": "session.InterfaceCoreCommandFile", "interface export": "session.InterfaceExportCommandFile",
		"query": "session.QueryCommandFile", "run": "session.RunInterpreterCommandFile", "verify": "session.VerifyCorpusFile",
	}
	var parserRoutes []string
	for route, root := range roots {
		if graph[root] == nil {
			t.Fatalf("Phase 35 consumer census lost command entry %s", root)
		}
		if phase34ConsumerReaches(graph, root, "syntax.Parse", map[string]bool{}) {
			parserRoutes = append(parserRoutes, route)
		}
	}
	sort.Strings(parserRoutes)
	if len(parserRoutes) != 13 {
		t.Fatalf("expression-depth parser routes = %v, want all 13 named parser routes", parserRoutes)
	}
	expected := map[string][]string{"parser.expression_depth": parserRoutes}
	manifest := surfacemanifest.Current()
	if problems := phase34LimitConsumerProblems(manifest, expected); len(problems) != 0 {
		t.Fatalf("expression-depth limit routes differ from actual parser consumers: %v", problems)
	}
	for _, control := range []struct {
		route, reason string
	}{{"run", "missing_consumer_route"}, {"app run", "unsupported_consumer_route"}} {
		mutated := phase34CopyManifest(t, manifest)
		for index := range mutated.Limits {
			if mutated.Limits[index].ID != "parser.expression_depth" {
				continue
			}
			routes := append([]string(nil), mutated.Limits[index].Routes...)
			if control.reason == "missing_consumer_route" {
				filtered := routes[:0]
				for _, route := range routes {
					if route != control.route {
						filtered = append(filtered, route)
					}
				}
				routes = filtered
			} else {
				routes = append(routes, control.route)
			}
			mutated.Limits[index].Routes = routes
		}
		want := control.reason + " parser.expression_depth " + control.route
		if !phase34ContainsProblem(phase34LimitConsumerProblems(mutated, expected), want) {
			t.Fatalf("depth route mutation did not reach %q", want)
		}
	}
	if problems := phase34LimitConsumerProblems(surfacemanifest.Current(), expected); len(problems) != 0 {
		t.Fatalf("depth route controls did not restore: %v", problems)
	}
}

func phase35AssertImmutablePhase34Fixtures(t *testing.T) {
	t.Helper()
	for _, fixture := range []struct{ path, want string }{
		{"testdata/phase34/surface-witnesses.json", "823ac61fac34274ba6ee4cb8719ab0618443600fc976ca5587e1ffee76869742"},
		{"testdata/phase34/surface-census.json", "3934469136ed0f741dbbbca79428a98bc8134c7bfc0e5862ca440e85a94c256a"},
	} {
		data, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(fixture.path)))
		if err != nil {
			t.Fatal(err)
		}
		actual := sha256.Sum256(data)
		if got := hex.EncodeToString(actual[:]); got != fixture.want {
			t.Fatalf("immutable Phase 34 fixture %s digest = %s, want %s", fixture.path, got, fixture.want)
		}
	}
}

func stringsSplitPhase35Facets(value string) []string {
	return strings.Split(value, "|")
}

func filterSurfaceWitnesses(witnesses []phase34SurfaceWitness, omit string) []phase34SurfaceWitness {
	filtered := make([]phase34SurfaceWitness, 0, len(witnesses))
	for _, witness := range witnesses {
		if witness.ID != omit {
			filtered = append(filtered, witness)
		}
	}
	return filtered
}

func filterSurfaceCensusBranches(branches []phase34CensusBranch, file, function, value string) []phase34CensusBranch {
	filtered := make([]phase34CensusBranch, 0, len(branches))
	for _, branch := range branches {
		if branch.File == file && branch.Function == function && len(branch.Values) == 1 && branch.Values[0] == value {
			continue
		}
		filtered = append(filtered, branch)
	}
	return filtered
}

func phase34FilterLimits(limits []surfacemanifest.Limit, id string) []surfacemanifest.Limit {
	filtered := make([]surfacemanifest.Limit, 0, len(limits))
	for _, limit := range limits {
		if limit.ID != id {
			filtered = append(filtered, limit)
		}
	}
	return filtered
}
