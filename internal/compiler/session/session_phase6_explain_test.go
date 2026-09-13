package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/debugmap"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// This file is deliberately `package session` (internal), unlike every
// other *_test.go sibling in this directory (`package session_test`).
// buildExplainGraph/resolveExplainDepth are intentionally unexported (the
// plan's own artifact list names only session.ExplainCommandFile as new
// exported surface) -- Tasks 2/3's <behavior> bullets need to construct
// exact depth/span/binding shapes no real in-tree fixture happens to
// produce (real Cause construction sites in check.go never repeat a
// Kind+Detail pair within one diagnostic), so this file exercises the
// synthesizer directly rather than only through the CLI seam. Schema
// minting (Task 1) and cross-corpus determinism (Task 3) still drive
// ExplainCommandFile end-to-end against real fixtures.

func spanPtr(start, end int) *diagnostic.Span {
	span := diagnostic.Span{Start: start, End: end}
	return &span
}

// --- Task 1: schema minting over a real fixture -----------------------

func TestExplainSummarySchemaIsMinted(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "use_after_move.lang")
	diagID := firstDiagnosticID(t, path)

	result, err := ExplainCommandFile(path, diagID, 0)
	if err != nil {
		t.Fatalf("ExplainCommandFile: %v", err)
	}
	if result.Explain == nil {
		t.Fatalf("result.Explain is nil")
	}
	if result.Explain.Schema != protocol.ExplainSchema {
		t.Fatalf("schema = %q, want %q", result.Explain.Schema, protocol.ExplainSchema)
	}
	if result.Explain.RootID == "" {
		t.Fatalf("RootID is empty")
	}
	if len(result.Explain.Nodes) == 0 {
		t.Fatalf("Nodes is empty")
	}
	validAvailability := map[string]bool{
		string(debugmap.Available): true, string(debugmap.OptimizedOut): true, string(debugmap.NotCaptured): true,
	}
	for _, node := range result.Explain.Nodes {
		if !validAvailability[node.Availability] {
			t.Fatalf("node %s has unknown availability %q", node.ID, node.Availability)
		}
	}

	// Two Results differing only in Explain must have different Result.ID
	// (Finalize folds ExplainID into identity).
	finalized := result.Finalize()
	withoutExplain := result
	withoutExplain.Explain = nil
	if finalized.ID == withoutExplain.Finalize().ID {
		t.Fatalf("Explain did not perturb Result.ID: %s", finalized.ID)
	}

	// usageResult's grammar is asserted at the CLI layer (main package);
	// here we just prove the command name round-trips through Finalize.
	if finalized.Command != "explain" {
		t.Fatalf("Command = %q, want explain", finalized.Command)
	}
}

// TestExplainDiagnosticNotFoundIsOperational proves the honest failure path:
// a request for an ID that does not exist in SRC's diagnostics is an
// operational failure carrying explain.diagnostic_not_found, never a panic
// or fabricated graph.
func TestExplainDiagnosticNotFoundIsOperational(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "use_after_move.lang")
	result, err := ExplainCommandFile(path, "diagnostic:does-not-exist", 0)
	if err != nil {
		t.Fatalf("ExplainCommandFile returned Go error: %v", err)
	}
	if result.Status != protocol.StatusOperational {
		t.Fatalf("Status = %q, want %q", result.Status, protocol.StatusOperational)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != ErrExplainDiagnosticNotFound.Code {
		t.Fatalf("Diagnostics = %+v, want one entry with code %s", result.Diagnostics, ErrExplainDiagnosticNotFound.Code)
	}
}

// --- Task 1: peer-re-derived function attribution (D-13-14/15/22/23) ----

// explainRealFunctionTable parses and checks a real .lang fixture, returning
// the real explainFunctionTable buildExplainFunctionTable derives from it
// (four real functions with real, non-overlapping spans for
// explain_three_function_chain.lang) alongside the raw parsed ast.Program,
// which callers use to write their OWN independent test-side resolution
// (D-13-23) rather than reusing anything production-side.
func explainRealFunctionTable(t *testing.T, path string) (explainFunctionTable, ast.Program) {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("%s: unexpected parse diagnostics: %+v", path, parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 1 {
		t.Fatalf("%s: want exactly one check diagnostic (check.interprocedural_loan_liveness), got %+v", path, checked.Diagnostics)
	}
	return buildExplainFunctionTable(checked.Program, parsed.Program), parsed.Program
}

// innerSpan returns a span strictly inside function's own whole-declaration
// span (never equal to it), roughly a third of the way through the body --
// deliberately not the Start/End boundary, so containment tests exercise a
// genuine "inside the body" cause span rather than an edge coordinate.
func innerSpan(function explainFunctionRef) *diagnostic.Span {
	width := function.Span.End - function.Span.Start
	start := function.Span.Start + width/3
	return &diagnostic.Span{Start: start, End: start + 2}
}

// testResolveFunctionOverAST is D-13-23's REQUIRED independent test-side
// resolver: innermost containing function, computed directly over
// ast.Program.Funcs by this test file's own loop -- it must never call
// (and does not call) narrowestContainingFunction/resolveExplainFunction,
// the production resolvers under test.
func testResolveFunctionOverAST(program ast.Program, span *diagnostic.Span) (string, bool) {
	best := -1
	for index, function := range program.Funcs {
		if function.Span.Start > span.Start || function.Span.End < span.End {
			continue
		}
		if best == -1 {
			best = index
			continue
		}
		bestWidth := program.Funcs[best].Span.End - program.Funcs[best].Span.Start
		candidateWidth := function.Span.End - function.Span.Start
		if candidateWidth < bestWidth {
			best = index
		}
	}
	if best == -1 {
		return "", false
	}
	return program.Funcs[best].Name, true
}

// TestExplainFunctionAttribution is D-13-23's falsifiable test: over the
// real four-function fixture (leaf, decoy, relay, caller -- decoy's byte
// range lies between leaf's and relay's), a synthetic diagnostic carries
// one cause per real function, each cause's span computed from that
// function's own real span. Each resulting node's function_id/function_name
// must equal the innermost AST function containing its span, verified
// against testResolveFunctionOverAST's independently-written resolver, not
// a read-back of the production node itself.
func TestExplainFunctionAttribution(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase13", "explain_three_function_chain.lang")
	table, astProgram := explainRealFunctionTable(t, path)

	names := []string{"leaf", "decoy", "relay", "caller"}
	causes := make([]diagnostic.Cause, len(names))
	wantByIndex := make([]string, len(names))
	for index, name := range names {
		var function explainFunctionRef
		found := false
		for _, candidate := range table.coreFuncs {
			if candidate.Name == name {
				function, found = candidate, true
				break
			}
		}
		if !found {
			t.Fatalf("function %q not present in the real core table -- fixture regressed", name)
		}
		span := innerSpan(function)
		causes[index] = diagnostic.Cause{Kind: "declared_here", Span: span}

		wantName, ok := testResolveFunctionOverAST(astProgram, span)
		if !ok {
			t.Fatalf("independent test resolver found no function for %q's own inner span", name)
		}
		if wantName != name {
			t.Fatalf("independent test resolver itself disagrees: span for %q resolved to %q -- fixture spans overlap", name, wantName)
		}
		wantByIndex[index] = name
	}

	root := diagnostic.Diagnostic{
		ID: "diagnostic:attribution", Code: "test.code", Primary: diagnostic.Span{Start: 0, End: 1}, Causes: causes,
	}
	nodes, _, functions, truncated, _, err := buildExplainGraph(root, protocol.ExplainDefaultDepth, table)
	if err != nil {
		t.Fatalf("buildExplainGraph: %v", err)
	}
	if truncated != "" {
		t.Fatalf("unexpected truncation: %q", truncated)
	}

	byID := make(map[string]protocol.ExplainNode, len(nodes))
	for _, node := range nodes {
		byID[node.ID] = node
	}
	for index, wantName := range wantByIndex {
		causeID := fmt.Sprintf("%s:cause:%d", root.ID, index)
		node, ok := byID[causeID]
		if !ok {
			t.Fatalf("cause %d (%s): node missing from graph", index, wantName)
		}
		if node.FunctionName != wantName {
			t.Fatalf("cause %d: function_name = %q, want %q", index, node.FunctionName, wantName)
		}
		if node.FunctionID == "" {
			t.Fatalf("cause %d: function_id is empty for a real, resolvable function", index)
		}
	}

	if len(functions) != len(table.coreFuncs) {
		t.Fatalf("ExplainSummary.Functions has %d entries, want %d (every consulted function)", len(functions), len(table.coreFuncs))
	}
}

// TestExplainNilSpanOmitsFunction covers D-13-23's nil-span rule: a cause
// with Span == nil emits NO function field and availability: not_captured,
// never a guessed function from a sibling or the root.
func TestExplainNilSpanOmitsFunction(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase13", "explain_three_function_chain.lang")
	table, _ := explainRealFunctionTable(t, path)

	root := diagnostic.Diagnostic{
		ID: "diagnostic:nilspan", Code: "test.code", Primary: diagnostic.Span{Start: 0, End: 1},
		Causes: []diagnostic.Cause{{Kind: "callee_return_contract", Detail: "no-span-cause"}},
	}
	nodes, _, _, _, _, err := buildExplainGraph(root, protocol.ExplainDefaultDepth, table)
	if err != nil {
		t.Fatalf("buildExplainGraph: %v", err)
	}
	found := false
	for _, node := range nodes {
		if node.ID == root.ID {
			continue
		}
		found = true
		if node.Span != nil {
			t.Fatalf("test constructed a nil-span cause but node has a span: %+v", node)
		}
		if node.FunctionID != "" || node.FunctionName != "" {
			t.Fatalf("nil-span node carries a function field: %+v", node)
		}
		if node.Availability != string(debugmap.NotCaptured) {
			t.Fatalf("nil-span node availability = %q, want %q", node.Availability, debugmap.NotCaptured)
		}
	}
	if !found {
		t.Fatalf("no cause node produced")
	}
}

// TestExplainFunctionIdentitySurvivesReorder is D-13-23's discriminating
// mutation: explain_three_function_chain_reordered.lang declares the SAME
// four functions in a different source order, so every byte offset shifts.
// A cause built from each fixture's own real inner span for the SAME
// function name must resolve to that SAME function_name in both fixtures,
// even though the spans themselves differ -- defeating any shortcut that
// hardcodes a position (e.g. "the first function", "the root's function")
// instead of genuinely resolving from the span.
func TestExplainFunctionIdentitySurvivesReorder(t *testing.T) {
	originalPath := testsupport.ProjectPath("testdata", "phase13", "explain_three_function_chain.lang")
	reorderedPath := testsupport.ProjectPath("testdata", "phase13", "explain_three_function_chain_reordered.lang")
	originalTable, _ := explainRealFunctionTable(t, originalPath)
	reorderedTable, _ := explainRealFunctionTable(t, reorderedPath)

	originalByName := map[string]explainFunctionRef{}
	for _, function := range originalTable.coreFuncs {
		originalByName[function.Name] = function
	}
	reorderedByName := map[string]explainFunctionRef{}
	for _, function := range reorderedTable.coreFuncs {
		reorderedByName[function.Name] = function
	}

	spanChanged := false
	for _, name := range []string{"leaf", "decoy", "relay", "caller"} {
		original, ok := originalByName[name]
		if !ok {
			t.Fatalf("function %q missing from original fixture", name)
		}
		reordered, ok := reorderedByName[name]
		if !ok {
			t.Fatalf("function %q missing from reordered fixture", name)
		}
		if original.Span != reordered.Span {
			spanChanged = true
		}

		for _, trial := range []struct {
			table explainFunctionTable
			span  *diagnostic.Span
		}{
			{originalTable, innerSpan(original)},
			{reorderedTable, innerSpan(reordered)},
		} {
			root := diagnostic.Diagnostic{
				ID: "diagnostic:reorder", Code: "test.code", Primary: diagnostic.Span{Start: 0, End: 1},
				Causes: []diagnostic.Cause{{Kind: "declared_here", Span: trial.span}},
			}
			nodes, _, _, _, _, err := buildExplainGraph(root, protocol.ExplainDefaultDepth, trial.table)
			if err != nil {
				t.Fatalf("buildExplainGraph: %v", err)
			}
			found := false
			for _, node := range nodes {
				if node.ID == root.ID {
					continue
				}
				found = true
				if node.FunctionName != name {
					t.Fatalf("function %q: resolved function_name = %q", name, node.FunctionName)
				}
			}
			if !found {
				t.Fatalf("function %q: no cause node produced", name)
			}
		}
	}
	if !spanChanged {
		t.Fatalf("reordered fixture did not actually shift any function's span -- mutation is not discriminating")
	}
}

// --- Task 2: typed edges, depth/node budget, stable truncation ---------

// TestExplainRespectsDepthAndNodeBudget covers Tests 1-3 of the plan's
// <behavior> list.
func TestExplainRespectsDepthAndNodeBudget(t *testing.T) {
	t.Run("depth=1 truncates a two-level chain to root+immediate causes", func(t *testing.T) {
		root := diagnostic.Diagnostic{
			ID: "diagnostic:root", Code: "test.code", Primary: diagnostic.Span{Start: 0, End: 100},
			Causes: []diagnostic.Cause{
				{Kind: "declared_here", Span: spanPtr(10, 90)}, // depth 1: narrows root
				{Kind: "declared_here", Span: spanPtr(20, 80)}, // depth 2: narrows cause 0
			},
		}
		nodes, edges, _, truncated, _, _ := buildExplainGraph(root, 1, explainFunctionTable{})
		if truncated != "truncated:explain.depth" {
			t.Fatalf("truncated = %q, want truncated:explain.depth", truncated)
		}
		if len(nodes) != 2 {
			t.Fatalf("nodes = %+v, want root + 1 immediate cause", nodes)
		}
		if len(edges) != 1 {
			t.Fatalf("edges = %+v, want exactly 1 edge", edges)
		}
	})

	t.Run("default depth is 3 when no depth is supplied", func(t *testing.T) {
		if got := resolveExplainDepth(0); got != protocol.ExplainDefaultDepth {
			t.Fatalf("resolveExplainDepth(0) = %d, want %d", got, protocol.ExplainDefaultDepth)
		}
		if protocol.ExplainDefaultDepth != 3 {
			t.Fatalf("protocol.ExplainDefaultDepth = %d, want 3 (D-06-03)", protocol.ExplainDefaultDepth)
		}
		if got := resolveExplainDepth(5); got != 5 {
			t.Fatalf("resolveExplainDepth(5) = %d, want 5 (explicit depth passes through)", got)
		}
	})

	t.Run("exceeding the node budget stops at the budget with the same truncation code family", func(t *testing.T) {
		causes := make([]diagnostic.Cause, protocol.ExplainMaxNodes+50)
		for index := range causes {
			causes[index] = diagnostic.Cause{Kind: "detail", Detail: fmt.Sprintf("d%d", index)}
		}
		root := diagnostic.Diagnostic{ID: "diagnostic:budget", Code: "test.code", Primary: diagnostic.Span{Start: 0, End: 1}, Causes: causes}
		nodes, _, _, truncated, _, _ := buildExplainGraph(root, protocol.ExplainDefaultDepth, explainFunctionTable{})
		if truncated != "truncated:explain.node_budget" {
			t.Fatalf("truncated = %q, want truncated:explain.node_budget", truncated)
		}
		if len(nodes) != protocol.ExplainMaxNodes {
			t.Fatalf("len(nodes) = %d, want exactly ExplainMaxNodes (%d)", len(nodes), protocol.ExplainMaxNodes)
		}
	})
}

// TestExplainTruncationCodeIsStable pins the exact two truncation code
// strings D-06-03 mints.
func TestExplainTruncationCodeIsStable(t *testing.T) {
	depthCase := diagnostic.Diagnostic{
		ID: "diagnostic:depth", Code: "test.code", Primary: diagnostic.Span{Start: 0, End: 100},
		Causes: []diagnostic.Cause{
			{Kind: "declared_here", Span: spanPtr(10, 90)},
			{Kind: "declared_here", Span: spanPtr(20, 80)},
		},
	}
	if _, _, _, truncated, _, _ := buildExplainGraph(depthCase, 1, explainFunctionTable{}); truncated != "truncated:explain.depth" {
		t.Fatalf("depth truncation code = %q, want truncated:explain.depth", truncated)
	}

	budgetCauses := make([]diagnostic.Cause, protocol.ExplainMaxNodes+10)
	for index := range budgetCauses {
		budgetCauses[index] = diagnostic.Cause{Kind: "detail", Detail: fmt.Sprintf("d%d", index)}
	}
	budgetCase := diagnostic.Diagnostic{ID: "diagnostic:budget", Code: "test.code", Primary: diagnostic.Span{Start: 0, End: 1}, Causes: budgetCauses}
	if _, _, _, truncated, _, _ := buildExplainGraph(budgetCase, protocol.ExplainDefaultDepth, explainFunctionTable{}); truncated != "truncated:explain.node_budget" {
		t.Fatalf("node budget truncation code = %q, want truncated:explain.node_budget", truncated)
	}
}

// TestExplainZeroCauseDiagnosticReturnsSingleNode covers Test 5: an honest
// empty result, never an error and never a fabricated cause (FND-04).
func TestExplainZeroCauseDiagnosticReturnsSingleNode(t *testing.T) {
	root := diagnostic.Diagnostic{ID: "diagnostic:lonely", Code: "test.code", Primary: diagnostic.Span{Start: 5, End: 6}}
	nodes, edges, _, truncated, work, _ := buildExplainGraph(root, protocol.ExplainDefaultDepth, explainFunctionTable{})
	if len(nodes) != 1 {
		t.Fatalf("nodes = %+v, want exactly 1 (the root)", nodes)
	}
	if len(edges) != 0 {
		t.Fatalf("edges = %+v, want none", edges)
	}
	if truncated != "" {
		t.Fatalf("truncated = %q, want empty", truncated)
	}
	if work != 1 {
		t.Fatalf("work = %d, want 1", work)
	}
	if nodes[0].ID != root.ID || nodes[0].Availability != string(debugmap.Available) {
		t.Fatalf("root node = %+v", nodes[0])
	}
}

// TestExplainEdgeKindVocabularyIsClosed covers Test 4 (same_binding,
// narrows, caused_by all reachable) directly, plus the corpus-wide closed
// enumeration the plan's acceptance criteria requires.
func TestExplainEdgeKindVocabularyIsClosed(t *testing.T) {
	t.Run("same_binding for shared identity, narrows for span containment, caused_by otherwise", func(t *testing.T) {
		root := diagnostic.Diagnostic{
			ID: "diagnostic:mixed", Code: "test.code", Primary: diagnostic.Span{Start: 0, End: 100},
			Causes: []diagnostic.Cause{
				{Kind: "place", Detail: "p:place:0"},           // 0: caused_by (first occurrence, no span)
				{Kind: "place", Detail: "p:place:0"},           // 1: same_binding -> 0
				{Kind: "declared_here", Span: spanPtr(10, 90)}, // 2: narrows -> root
				{Kind: "type", Detail: "p:type:0"},             // 3: caused_by (no correlation, no span)
			},
		}
		_, edges, _, truncated, _, _ := buildExplainGraph(root, protocol.ExplainDefaultDepth, explainFunctionTable{})
		if truncated != "" {
			t.Fatalf("unexpected truncation: %q", truncated)
		}
		kinds := map[string]int{}
		for _, edge := range edges {
			kinds[edge.Kind]++
		}
		for _, want := range []string{protocol.EdgeCausedBy, protocol.EdgeNarrows, protocol.EdgeSameBinding} {
			if kinds[want] == 0 {
				t.Fatalf("edge kind %q never emitted: edges=%+v", want, edges)
			}
		}
	})

	t.Run("closed over the whole in-tree rejecting-fixture set", func(t *testing.T) {
		validKinds := map[string]bool{protocol.EdgeCausedBy: true, protocol.EdgeNarrows: true, protocol.EdgeSameBinding: true}
		checked := 0
		for _, dir := range []string{"phase2", "phase3", "phase4", "phase5"} {
			for _, fixture := range rejectingFixtures(t, dir) {
				for _, diagID := range fixtureDiagnosticIDs(t, fixture) {
					result, err := ExplainCommandFile(fixture, diagID, 0)
					if err != nil {
						t.Fatalf("%s %s: %v", fixture, diagID, err)
					}
					if result.Explain == nil {
						t.Fatalf("%s %s: Explain is nil", fixture, diagID)
					}
					for _, edge := range result.Explain.Edges {
						checked++
						if !validKinds[edge.Kind] {
							t.Fatalf("%s %s: edge kind %q outside the closed vocabulary", fixture, diagID, edge.Kind)
						}
					}
				}
			}
		}
		if checked == 0 {
			t.Fatalf("no edges were checked -- the in-tree rejecting-fixture set produced no cause edges")
		}
	})
}

// --- Task 3: the determinism obligation ---------------------------------

// TestExplainCauseGraphIsDeterministic discharges D-06-02's stated
// obligation as an executable claim: ExplainCommandFile called twice on the
// same source and diagnostic ID produces byte-identical marshalled
// ExplainSummary bytes and the same Result.ID, over every rejecting fixture
// in the Phase 2-5 corpora (not one hand-picked case).
func TestExplainCauseGraphIsDeterministic(t *testing.T) {
	total := 0
	for _, dir := range []string{"phase2", "phase3", "phase4", "phase5"} {
		for _, fixture := range rejectingFixtures(t, dir) {
			for _, diagID := range fixtureDiagnosticIDs(t, fixture) {
				fixture, diagID := fixture, diagID
				name := fmt.Sprintf("%s/%s", filepath.Base(fixture), diagID)
				t.Run(name, func(t *testing.T) {
					first, err := ExplainCommandFile(fixture, diagID, 0)
					if err != nil {
						t.Fatalf("first call: %v", err)
					}
					second, err := ExplainCommandFile(fixture, diagID, 0)
					if err != nil {
						t.Fatalf("second call: %v", err)
					}
					firstBytes, err := json.Marshal(first.Explain)
					if err != nil {
						t.Fatalf("marshal first: %v", err)
					}
					secondBytes, err := json.Marshal(second.Explain)
					if err != nil {
						t.Fatalf("marshal second: %v", err)
					}
					if string(firstBytes) != string(secondBytes) {
						t.Fatalf("non-deterministic ExplainSummary bytes:\nfirst:  %s\nsecond: %s", firstBytes, secondBytes)
					}
					if first.Finalize().ID != second.Finalize().ID {
						t.Fatalf("non-deterministic Result.ID: first=%s second=%s", first.Finalize().ID, second.Finalize().ID)
					}
				})
				total++
			}
		}
	}
	if total == 0 {
		t.Fatalf("no rejecting fixtures with diagnostics found across testdata/phase2-5")
	}
}

// TestExplainNodeOrderIsStableOnTies constructs a diagnostic whose causes
// attach at out-of-insertion-order depths: cause 0 and cause 2 both end up
// at depth 1 (cause 2 attaches via `narrows` directly to root, skipping
// past cause 1 which is inserted between them), while cause 1 correlates
// via `same_binding` to cause 0 and lands one level deeper at depth 2. Raw
// insertion order is therefore [root, cause0(depth1), cause1(depth2),
// cause2(depth1)] -- NOT depth-ascending -- so the final
// sort.SliceStable(items, ...) in buildExplainGraph is load-bearing, not
// decorative. This was manually verified during 06-02's execution: deleting
// that sort call reproduces the raw insertion order below and this test
// goes red (recorded in 06-02-SUMMARY.md), proving the test is not inert.
func TestExplainNodeOrderIsStableOnTies(t *testing.T) {
	build := func() ([]string, []protocol.ExplainEdge) {
		root := diagnostic.Diagnostic{
			ID: "diagnostic:tie", Code: "test.code", Primary: diagnostic.Span{Start: 0, End: 100},
			Causes: []diagnostic.Cause{
				{Kind: "place", Detail: "p1"},                  // 0: caused_by root, depth 1
				{Kind: "place", Detail: "p1"},                  // 1: same_binding -> 0, depth 2
				{Kind: "declared_here", Span: spanPtr(10, 90)}, // 2: narrows root, depth 1
			},
		}
		nodes, edges, _, _, _, _ := buildExplainGraph(root, protocol.ExplainDefaultDepth, explainFunctionTable{})
		ids := make([]string, len(nodes))
		for index, node := range nodes {
			ids[index] = node.ID
		}
		return ids, edges
	}

	first, firstEdges := build()
	second, _ := build()
	if fmt.Sprint(first) != fmt.Sprint(second) {
		t.Fatalf("two identical calls produced different node order: %v vs %v", first, second)
	}

	// depth-ascending order interleaves cause2 (depth 1) BEFORE cause1
	// (depth 2), even though cause1 was inserted first -- this is exactly
	// what raw insertion order would get wrong.
	want := []string{"diagnostic:tie", "diagnostic:tie:cause:0", "diagnostic:tie:cause:2", "diagnostic:tie:cause:1"}
	if fmt.Sprint(first) != fmt.Sprint(want) {
		t.Fatalf("node order = %v, want %v (depth-ascending, not insertion order)", first, want)
	}

	wantEdgeKinds := map[string]string{
		"diagnostic:tie:cause:0": protocol.EdgeCausedBy,
		"diagnostic:tie:cause:1": protocol.EdgeSameBinding,
		"diagnostic:tie:cause:2": protocol.EdgeNarrows,
	}
	for _, edge := range firstEdges {
		if got, want := edge.Kind, wantEdgeKinds[edge.To]; got != want {
			t.Fatalf("edge to %s has kind %q, want %q", edge.To, got, want)
		}
	}
}

// TestExplainSynthesisOpensNoWritePath structurally asserts D-06-02's
// non-persistence claim: the explain synthesizer's own source file imports
// no file-write or cache package, so "never persisted" is a checked
// property rather than a claim.
func TestExplainSynthesisOpensNoWritePath(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "session", "session_phase6_explain.go"))
	if err != nil {
		t.Fatalf("read session_phase6_explain.go: %v", err)
	}
	forbidden := []string{"os.Create", "os.WriteFile", "os.OpenFile", `"cache"`, "ioutil.WriteFile"}
	for _, marker := range forbidden {
		if strings.Contains(string(source), marker) {
			t.Fatalf("session_phase6_explain.go references forbidden write/cache marker %q", marker)
		}
	}
}

// --- shared fixture-corpus helpers --------------------------------------

// rejectingFixtures returns the absolute paths of every .lang fixture under
// testdata/<dir> that produces at least one diagnostic at parse or check
// time, discovered dynamically rather than hand-listed -- the standing
// project rule (adopted after three gate failures shared one shape) that a
// property test's reachable input space must be the real corpus, not a
// curated subset.
func rejectingFixtures(t *testing.T, dir string) []string {
	t.Helper()
	root := testsupport.ProjectPath("testdata", dir)
	matches, err := filepath.Glob(filepath.Join(root, "*.lang"))
	if err != nil {
		t.Fatalf("glob %s: %v", root, err)
	}
	var rejecting []string
	for _, path := range matches {
		if len(fixtureDiagnosticIDs(t, path)) > 0 {
			rejecting = append(rejecting, path)
		}
	}
	sort.Strings(rejecting)
	return rejecting
}

// fixtureDiagnosticIDs parses (and, if the parse is clean, checks) path and
// returns every diagnostic ID it produced. It is the single source of truth
// the tests in this file use to discover real diagnostic IDs to drive
// ExplainCommandFile with.
func fixtureDiagnosticIDs(t *testing.T, path string) []string {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	parsed := syntax.Parse(source)
	diagnostics := parsed.Diagnostics
	if len(diagnostics) == 0 {
		checked := check.Program(parsed.Program)
		diagnostics = checked.Diagnostics
	}
	ids := make([]string, 0, len(diagnostics))
	for _, diag := range diagnostics {
		ids = append(ids, diag.ID)
	}
	return ids
}

// firstDiagnosticID is a small convenience wrapper for tests that only need
// one diagnostic ID from a known-rejecting fixture.
func firstDiagnosticID(t *testing.T, path string) string {
	t.Helper()
	ids := fixtureDiagnosticIDs(t, path)
	if len(ids) == 0 {
		t.Fatalf("%s produced no diagnostics", path)
	}
	return ids[0]
}
