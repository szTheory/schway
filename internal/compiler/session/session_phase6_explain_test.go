package session

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

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
		nodes, edges, truncated, _ := buildExplainGraph(root, 1)
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
		nodes, _, truncated, _ := buildExplainGraph(root, protocol.ExplainDefaultDepth)
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
	if _, _, truncated, _ := buildExplainGraph(depthCase, 1); truncated != "truncated:explain.depth" {
		t.Fatalf("depth truncation code = %q, want truncated:explain.depth", truncated)
	}

	budgetCauses := make([]diagnostic.Cause, protocol.ExplainMaxNodes+10)
	for index := range budgetCauses {
		budgetCauses[index] = diagnostic.Cause{Kind: "detail", Detail: fmt.Sprintf("d%d", index)}
	}
	budgetCase := diagnostic.Diagnostic{ID: "diagnostic:budget", Code: "test.code", Primary: diagnostic.Span{Start: 0, End: 1}, Causes: budgetCauses}
	if _, _, truncated, _ := buildExplainGraph(budgetCase, protocol.ExplainDefaultDepth); truncated != "truncated:explain.node_budget" {
		t.Fatalf("node budget truncation code = %q, want truncated:explain.node_budget", truncated)
	}
}

// TestExplainZeroCauseDiagnosticReturnsSingleNode covers Test 5: an honest
// empty result, never an error and never a fabricated cause (FND-04).
func TestExplainZeroCauseDiagnosticReturnsSingleNode(t *testing.T) {
	root := diagnostic.Diagnostic{ID: "diagnostic:lonely", Code: "test.code", Primary: diagnostic.Span{Start: 5, End: 6}}
	nodes, edges, truncated, work := buildExplainGraph(root, protocol.ExplainDefaultDepth)
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
		_, edges, truncated, _ := buildExplainGraph(root, protocol.ExplainDefaultDepth)
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
