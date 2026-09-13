package session

import (
	"fmt"
	"sort"
	"time"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/debugmap"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// ExplainError is this file's stable typed failure, matching the
// {Code}-only shape debugmap.Error and originvalidate.Error already use so
// callers can dispatch on Code identically.
type ExplainError struct{ Code string }

func (e *ExplainError) Error() string { return e.Code }

// ErrExplainDiagnosticNotFound is returned (wrapped into a protocol.Result,
// never as a Go error from ExplainCommandFile) when SRC's parse/check
// diagnostics contain no diagnostic with the requested ID.
var ErrExplainDiagnosticNotFound = &ExplainError{Code: "explain.diagnostic_not_found"}

// ErrExplainFunctionResolutionDisagreement is returned (wrapped into a
// protocol.Result, never as a Go error from ExplainCommandFile) when
// D-13-15's peer-re-derived function attribution -- once from the core
// function table, once from the AST -- disagrees for the same span. A
// disagreement is refused, never silently resolved by preferring either
// peer.
var ErrExplainFunctionResolutionDisagreement = &ExplainError{Code: "explain.function_resolution_disagreement"}

// resolveExplainDepth applies D-06-03's default: requested <= 0 means "no
// --depth supplied", which resolves to protocol.ExplainDefaultDepth.
func resolveExplainDepth(requested int) int {
	if requested <= 0 {
		return protocol.ExplainDefaultDepth
	}
	return requested
}

// ExplainCommandFile is the CLI seam for `lang explain SRC ID`: it parses
// and checks SRC exactly like `check` (D-06-02 forbids any persisted store
// or daemon, so the diagnostic must be re-derived from source on every cold
// invocation), locates the diagnostic whose ID equals diagnosticID among
// whichever diagnostics that pass actually produced, and synthesizes a
// bounded cause DAG under lang.explain/0 (D-06-04) from that diagnostic's
// existing flat Causes list. The graph is recomputed fresh every call and
// never persisted — two cold invocations on the same input produce
// byte-identical output (D-06-02's testable determinism obligation).
func ExplainCommandFile(path, diagnosticID string, depth int) (protocol.Result, error) {
	started := time.Now()
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil {
		return protocol.Result{}, err
	}
	parsed := syntax.Parse(source)
	result := protocol.New("explain", protocol.StatusPass)

	diagnostics := parsed.Diagnostics
	work := 1
	// checkedProgram stays the zero core.Program when parsed.Diagnostics
	// already carries a syntax error (check never runs) or when check
	// itself wipes result.Program after refusing a cyclic call graph
	// (check.go's checkCallGraphAcyclic branch) -- both are the honest "no
	// function data available" case buildExplainFunctionTable's own doc
	// comment describes, never a bug to work around.
	var checkedProgram core.Program
	if len(diagnostics) == 0 {
		checked := check.Program(parsed.Program)
		diagnostics = checked.Diagnostics
		work = checked.Work
		checkedProgram = checked.Program
	}

	target, found := findExplainDiagnostic(diagnostics, diagnosticID)
	if !found {
		result.Status = protocol.StatusOperational
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error(ErrExplainDiagnosticNotFound.Code, diagnostic.Span{}, "no diagnostic with that id was found")}
		return completeCommand(result, started, work), nil
	}

	table := buildExplainFunctionTable(checkedProgram, parsed.Program)
	nodes, edges, functions, truncated, graphWork, graphErr := buildExplainGraph(target, resolveExplainDepth(depth), table)
	if graphErr != nil {
		result.Status = protocol.StatusOperational
		code := ErrExplainFunctionResolutionDisagreement.Code
		if typed, ok := graphErr.(*ExplainError); ok {
			code = typed.Code
		}
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error(code, diagnostic.Span{}, graphErr.Error())}
		return completeCommand(result, started, work+graphWork), nil
	}
	result.Explain = &protocol.ExplainSummary{
		Schema: protocol.ExplainSchema, RootID: target.ID, Nodes: nodes, Edges: edges, Truncated: truncated, Functions: functions,
	}
	return completeCommand(result, started, work+graphWork), nil
}

func findExplainDiagnostic(diagnostics []diagnostic.Diagnostic, id string) (diagnostic.Diagnostic, bool) {
	for _, candidate := range diagnostics {
		if candidate.ID == id {
			return candidate, true
		}
	}
	return diagnostic.Diagnostic{}, false
}

// explainBuilt pairs one synthesized ExplainNode with the depth it was
// attached at — depth is not part of ExplainNode's own JSON shape (it is a
// synthesis-time bookkeeping fact, not a published field), so it travels
// alongside the node only until buildExplainGraph's final sort.
type explainBuilt struct {
	node  protocol.ExplainNode
	depth int
}

// explainCorrelationKinds are the diagnostic.Cause Kind values that name a
// place/loan/binding identity in their Detail field — verified directly
// against every `diagnostic.Cause{Kind: ...}` construction site in
// internal/compiler/check/check.go before this was written. Two causes
// sharing one of these Kind+Detail pairs correlate to the same binding.
var explainCorrelationKinds = map[string]bool{
	"place": true, "owner": true, "loan": true, "transfer_target": true,
}

// explainCorrelationKey returns the empty string when kind/detail carry no
// binding identity to correlate on.
func explainCorrelationKey(kind, detail string) string {
	if detail == "" || !explainCorrelationKinds[kind] {
		return ""
	}
	return kind + ":" + detail
}

// explainSpanStrictlyContains reports whether parent strictly contains
// child: parent's bounds are at least as wide as child's, and the two spans
// are not identical.
func explainSpanStrictlyContains(parent, child *diagnostic.Span) bool {
	if parent == nil || child == nil {
		return false
	}
	if parent.Start > child.Start || parent.End < child.End {
		return false
	}
	return parent.Start != child.Start || parent.End != child.End
}

func explainSpanWidth(span *diagnostic.Span) int {
	if span == nil {
		return 0
	}
	return span.End - span.Start
}

func explainSpanBounds(span *diagnostic.Span) (int, int) {
	if span == nil {
		return 0, 0
	}
	return span.Start, span.End
}

// explainFunctionRef names one function attribution candidate: ID (always
// core-derived -- ast.FuncDecl carries no ID of its own, D-13-15), Name, and
// the function's WHOLE declaration span.
type explainFunctionRef struct {
	ID   string
	Name string
	Span diagnostic.Span
}

// explainFunctionTable holds D-13-15's two independently-derived function
// tables: coreFuncs from core.Program.Functions (ID+Name+Span) and
// astFuncs from ast.Program.Funcs (Name+Span only). Both empty is a
// legitimate, honest "no function data available" state -- e.g. a syntax
// error (check never ran) or check.go's checkCallGraphAcyclic branch, which
// wipes result.Program to its zero value after a cycle refusal -- and
// resolveExplainFunction treats an empty coreFuncs as "attribution not
// attempted", never as a peer disagreement.
type explainFunctionTable struct {
	coreFuncs []explainFunctionRef
	astFuncs  []explainFunctionRef // ID always "": peer-comparison only.
}

// buildExplainFunctionTable derives both tables directly from
// ExplainCommandFile's own existing locals: program is checked.Program (the
// core function table), parsedProgram is parsed.Program (the AST). No new
// state is plumbed from elsewhere (D-13-15's key link).
func buildExplainFunctionTable(program core.Program, parsedProgram ast.Program) explainFunctionTable {
	table := explainFunctionTable{}
	for _, function := range program.Functions {
		table.coreFuncs = append(table.coreFuncs, explainFunctionRef{ID: function.ID, Name: function.Name, Span: function.Span})
	}
	for _, function := range parsedProgram.Funcs {
		table.astFuncs = append(table.astFuncs, explainFunctionRef{Name: function.Name, Span: function.Span})
	}
	return table
}

// wholeFunctionSpanSet returns the D-13-19 guard's lookup set: every
// core-derived function's own whole-declaration span, keyed by
// (start, end). A candidate parent node whose span exactly equals one of
// these is a function-scope node and may only be a caused_by parent, never
// a narrows parent.
func (t explainFunctionTable) wholeFunctionSpanSet() map[[2]int]bool {
	set := make(map[[2]int]bool, len(t.coreFuncs))
	for _, function := range t.coreFuncs {
		set[[2]int{function.Span.Start, function.Span.End}] = true
	}
	return set
}

// sortedFunctions returns table.coreFuncs as protocol.ExplainFunction
// entries, ordered by span start then ID (deterministic, independent of
// core.Program.Functions' own emission order) for ExplainSummary.Functions.
func (t explainFunctionTable) sortedFunctions() []protocol.ExplainFunction {
	if len(t.coreFuncs) == 0 {
		return nil
	}
	functions := make([]protocol.ExplainFunction, len(t.coreFuncs))
	for index, function := range t.coreFuncs {
		functions[index] = protocol.ExplainFunction{ID: function.ID, Name: function.Name, Span: function.Span}
	}
	sort.SliceStable(functions, func(left, right int) bool {
		if functions[left].Span.Start != functions[right].Span.Start {
			return functions[left].Span.Start < functions[right].Span.Start
		}
		return functions[left].ID < functions[right].ID
	})
	return functions
}

// narrowestContainingFunction returns the narrowest candidate whose span
// contains span (inclusive bounds -- a cause's own span may legitimately
// equal a function's whole span, e.g. a whole-function Primary), per
// D-13-15's "innermost containing function" resolution rule.
func narrowestContainingFunction(candidates []explainFunctionRef, span *diagnostic.Span) (explainFunctionRef, bool) {
	best := -1
	for index := range candidates {
		candidateSpan := candidates[index].Span
		if candidateSpan.Start > span.Start || candidateSpan.End < span.End {
			continue
		}
		if best == -1 {
			best = index
			continue
		}
		bestSpan := candidates[best].Span
		if (candidateSpan.End - candidateSpan.Start) < (bestSpan.End - bestSpan.Start) {
			best = index
		}
	}
	if best == -1 {
		return explainFunctionRef{}, false
	}
	return candidates[best], true
}

// resolveExplainFunction resolves span to its innermost containing function
// declaration, independently from BOTH table.coreFuncs and table.astFuncs
// (D-13-15's peer re-derivation), refusing -- never silently preferring
// either peer -- when the two derivations disagree. A nil span, or an
// empty coreFuncs table (explainFunctionTable's own "no function data
// available" case), always resolves to "no function" without error.
func resolveExplainFunction(table explainFunctionTable, span *diagnostic.Span) (explainFunctionRef, bool, error) {
	if span == nil || len(table.coreFuncs) == 0 {
		return explainFunctionRef{}, false, nil
	}
	coreMatch, coreOK := narrowestContainingFunction(table.coreFuncs, span)
	if !coreOK {
		// The core table is authoritative for FunctionID (ast.FuncDecl
		// carries no ID of its own), and it is legitimately INCOMPLETE:
		// check.go only appends a function to core.Program.Functions once
		// it checks cleanly (a function that itself carries the diagnostic
		// being explained, or any other diagnostic, is skipped). The AST
		// table still has every declared function regardless. That is not
		// a peer disagreement to refuse -- core simply has no ID to offer
		// for this span, so this is an honest "no function" report, the
		// same shape as a nil span or foreign-boundary span.
		return explainFunctionRef{}, false, nil
	}
	// Core DID claim a function for this span: independently verify the
	// SAME identity (Name + whole Span) is what the AST table resolves to
	// for the same span. Disagreement here IS refused (D-13-15) -- core.
	// Function.Span is supposed to be copied directly from the matching
	// ast.FuncDecl.Span (check.go), so any mismatch is a genuine
	// derivation bug, never a structural incompleteness to route around
	// silently.
	astMatch, astOK := narrowestContainingFunction(table.astFuncs, span)
	if !astOK || coreMatch.Name != astMatch.Name || coreMatch.Span != astMatch.Span {
		return explainFunctionRef{}, false, ErrExplainFunctionResolutionDisagreement
	}
	return coreMatch, true, nil
}

// applyExplainFunction sets node's FunctionID/FunctionName when has is
// true, and leaves both at their zero value (omitted from JSON) otherwise
// -- never a guessed or fabricated value (D-13-23).
func applyExplainFunction(node *protocol.ExplainNode, function explainFunctionRef, has bool) {
	if !has {
		return
	}
	node.FunctionID = function.ID
	node.FunctionName = function.Name
}

// explainBlameSpanSet returns the D-13-22 blame-site lookup set for one
// diagnostic: its own Primary span (the blamed site under D-13-01's
// contract-boundary rule, B2-default or otherwise) plus, when the
// diagnostic publishes D-13-07's blame_undetermined dual-site repairs, each
// published alternative site. Derived from the diagnostic payload alone --
// this file never imports the blame resolver itself (D-13-06 keeps blame
// derivation inside check).
func explainBlameSpanSet(diag diagnostic.Diagnostic) map[[2]int]bool {
	set := map[[2]int]bool{{diag.Primary.Start, diag.Primary.End}: true}
	for _, repair := range diag.Repairs {
		if repair.Kind == "blame_undetermined" && repair.Span != nil {
			set[[2]int{repair.Span.Start, repair.Span.End}] = true
		}
	}
	return set
}

// buildExplainGraph synthesizes the bounded cause DAG for one diagnostic:
// one root node (the diagnostic itself) plus one node per entry in its flat
// Causes list, with each cause attached to the most specific correlated
// ancestor already in the graph (D-06-03's edge derivation: same_binding
// when two nodes correlate to the same binding/place identity, narrows
// when a child span is strictly contained within its parent's span,
// caused_by otherwise). Exceeding maxDepth or protocol.ExplainMaxNodes
// stops expansion and returns a stable truncation code rather than
// growing further; a diagnostic with zero causes returns exactly one node,
// zero edges, and no truncation code (FND-04 empty-input edge). Nodes are
// sorted by (depth, span.start, span.end, id) and edges by (from, to,
// kind) before returning, so ties compare equal and stable (FND-04
// ordering edge) — this is part of the contract, not a convenience, and is
// pinned by TestExplainNodeOrderIsStableOnTies.
func buildExplainGraph(diag diagnostic.Diagnostic, maxDepth int, table explainFunctionTable) ([]protocol.ExplainNode, []protocol.ExplainEdge, []protocol.ExplainFunction, string, int, error) {
	rootSpan := diag.Primary
	rootFunction, rootHasFunction, err := resolveExplainFunction(table, &rootSpan)
	if err != nil {
		return nil, nil, nil, "", 0, err
	}
	rootNode := protocol.ExplainNode{
		ID: diag.ID, Kind: diag.Code, Detail: diag.Message, Span: &rootSpan,
		Availability: string(debugmap.Available),
	}
	applyExplainFunction(&rootNode, rootFunction, rootHasFunction)
	items := []explainBuilt{{node: rootNode, depth: 0}}
	edges := make([]protocol.ExplainEdge, 0, len(diag.Causes))
	byKey := make(map[string]int, len(diag.Causes))
	truncated := ""
	work := 1
	wholeFunctionSpans := table.wholeFunctionSpanSet()

	for index, cause := range diag.Causes {
		work++
		if len(items) >= protocol.ExplainMaxNodes {
			truncated = "truncated:explain.node_budget"
			break
		}

		causeFunction, causeHasFunction, causeErr := resolveExplainFunction(table, cause.Span)
		if causeErr != nil {
			return nil, nil, nil, "", 0, causeErr
		}
		causeFunctionID := ""
		if causeHasFunction {
			causeFunctionID = causeFunction.ID
		}

		parent := 0
		edgeKind := protocol.EdgeCausedBy
		key := explainCorrelationKey(cause.Kind, cause.Detail)
		if key != "" {
			if match, ok := byKey[key]; ok {
				parent = match
				edgeKind = protocol.EdgeSameBinding
			}
		}
		if edgeKind == protocol.EdgeCausedBy && cause.Span != nil {
			best := -1
			for candidate := range items {
				candidateNode := items[candidate].node
				parentSpan := candidateNode.Span
				if !explainSpanStrictlyContains(parentSpan, cause.Span) {
					continue
				}
				// D-13-19: a node whose span exactly equals a
				// core.Function.Span is a function-scope node and may only
				// be a caused_by parent -- pure byte containment alone
				// would otherwise make it the narrowest containing
				// ancestor of every later cause anywhere in its body,
				// asserting a containment relation the rule was never
				// written to mean once causes cross functions.
				if parentSpan != nil && wholeFunctionSpans[[2]int{parentSpan.Start, parentSpan.End}] {
					continue
				}
				// D-13-19: narrows additionally requires parent and child
				// to resolve to the SAME function_id -- cross-function
				// parenting stays caused_by, keeping the closed
				// three-value edge vocabulary intact. Both sides empty
				// (no function data available, e.g. a synthetic
				// diagnostic with an empty table) compares equal, so
				// single-function-context behavior is unperturbed.
				if candidateNode.FunctionID != causeFunctionID {
					continue
				}
				if best == -1 || explainSpanWidth(parentSpan) < explainSpanWidth(items[best].node.Span) {
					best = candidate
				}
			}
			if best != -1 {
				parent = best
				edgeKind = protocol.EdgeNarrows
			}
		}

		depth := items[parent].depth + 1
		if depth > maxDepth {
			truncated = "truncated:explain.depth"
			continue
		}

		causeID := fmt.Sprintf("%s:cause:%d", diag.ID, index)
		availability := string(debugmap.Available)
		if cause.Span == nil {
			availability = string(debugmap.NotCaptured)
		}
		causeNode := protocol.ExplainNode{ID: causeID, Kind: cause.Kind, Detail: cause.Detail, Span: cause.Span, Availability: availability}
		applyExplainFunction(&causeNode, causeFunction, causeHasFunction)
		items = append(items, explainBuilt{node: causeNode, depth: depth})
		edges = append(edges, protocol.ExplainEdge{From: items[parent].node.ID, To: causeID, Kind: edgeKind})
		if key != "" {
			if _, ok := byKey[key]; !ok {
				byKey[key] = len(items) - 1
			}
		}
	}

	// D-13-22: mark the blamed site(s) -- the diagnostic's own Primary span
	// (always the root node's own span, by construction above) plus any
	// blame_undetermined dual-site repair spans this diagnostic publishes.
	blameSpans := explainBlameSpanSet(diag)
	for index := range items {
		span := items[index].node.Span
		if span != nil && blameSpans[[2]int{span.Start, span.End}] {
			items[index].node.Blame = true
		}
	}

	sort.SliceStable(items, func(left, right int) bool {
		if items[left].depth != items[right].depth {
			return items[left].depth < items[right].depth
		}
		leftStart, leftEnd := explainSpanBounds(items[left].node.Span)
		rightStart, rightEnd := explainSpanBounds(items[right].node.Span)
		if leftStart != rightStart {
			return leftStart < rightStart
		}
		if leftEnd != rightEnd {
			return leftEnd < rightEnd
		}
		return items[left].node.ID < items[right].node.ID
	})
	nodes := make([]protocol.ExplainNode, len(items))
	for index, item := range items {
		nodes[index] = item.node
	}

	sort.SliceStable(edges, func(left, right int) bool {
		if edges[left].From != edges[right].From {
			return edges[left].From < edges[right].From
		}
		if edges[left].To != edges[right].To {
			return edges[left].To < edges[right].To
		}
		return edges[left].Kind < edges[right].Kind
	})

	return nodes, edges, table.sortedFunctions(), truncated, work, nil
}
