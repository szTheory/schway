package session

import (
	"fmt"
	"sort"
	"time"

	"github.com/codename-lang/lang/internal/compiler/check"
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
	if len(diagnostics) == 0 {
		checked := check.Program(parsed.Program)
		diagnostics = checked.Diagnostics
		work = checked.Work
	}

	target, found := findExplainDiagnostic(diagnostics, diagnosticID)
	if !found {
		result.Status = protocol.StatusOperational
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error(ErrExplainDiagnosticNotFound.Code, diagnostic.Span{}, "no diagnostic with that id was found")}
		return completeCommand(result, started, work), nil
	}

	nodes, edges, truncated, graphWork := buildExplainGraph(target, resolveExplainDepth(depth))
	result.Explain = &protocol.ExplainSummary{
		Schema: protocol.ExplainSchema, RootID: target.ID, Nodes: nodes, Edges: edges, Truncated: truncated,
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
func buildExplainGraph(diag diagnostic.Diagnostic, maxDepth int) ([]protocol.ExplainNode, []protocol.ExplainEdge, string, int) {
	rootSpan := diag.Primary
	items := []explainBuilt{{
		node: protocol.ExplainNode{
			ID: diag.ID, Kind: diag.Code, Detail: diag.Message, Span: &rootSpan,
			Availability: string(debugmap.Available),
		},
		depth: 0,
	}}
	edges := make([]protocol.ExplainEdge, 0, len(diag.Causes))
	byKey := make(map[string]int, len(diag.Causes))
	truncated := ""
	work := 1

	for index, cause := range diag.Causes {
		work++
		if len(items) >= protocol.ExplainMaxNodes {
			truncated = "truncated:explain.node_budget"
			break
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
				parentSpan := items[candidate].node.Span
				if !explainSpanStrictlyContains(parentSpan, cause.Span) {
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
		items = append(items, explainBuilt{
			node:  protocol.ExplainNode{ID: causeID, Kind: cause.Kind, Detail: cause.Detail, Span: cause.Span, Availability: availability},
			depth: depth,
		})
		edges = append(edges, protocol.ExplainEdge{From: items[parent].node.ID, To: causeID, Kind: edgeKind})
		if key != "" {
			if _, ok := byKey[key]; !ok {
				byKey[key] = len(items) - 1
			}
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

	return nodes, edges, truncated, work
}
