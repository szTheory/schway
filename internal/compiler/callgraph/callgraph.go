// Package callgraph is Phase 07 Stage 2a's whole-program call-graph
// acyclicity check (SEM-07, D-07-14/D-07-17/D-07-18). It is consumed by
// check, in production, immediately before check.Program returns a
// core.Program -- never by corevalidate, ast, interp, or cgen.
// corevalidate runs its own, independently implemented traversal (07-07);
// this package restates nothing from it. Unlike pathoracle (consumed only
// by tests and the verify harness), callgraph sits on check's own
// production admission path, so every fault-injection seam this package
// declares is unexported (D-07-42).
//
// It reads nothing but a core.Program's own declared Functions and their
// Linear.Operations. It imports neither corevalidate, ast, interp, cgen,
// nor check itself -- TestImportsStayIndependent enforces this statically.
//
// Node identity is the function ID (D-07-30): this package does NOT reuse
// corevalidate's unexported, Name-matching isDeclaredFunctionName, which
// answers a different question (foreign symbol vs. Lang function at
// resolution time) and stays exactly where it is, exercised by its own
// check-side fixture in 07-07. Edges come from
// core.LinearOperation.CalleeID on operations with Kind == core.OpCall,
// and nothing else, enumerated across a function's whole Linear.Operations
// list -- never reconstructed from Block/Successor position (D-07-28).
//
// The traversal is an iterative three-color (white/gray/black) depth-first
// search over an explicit stack -- never native Go recursion, so the
// compile-time traversal costs no Go stack. The distinguisher is on-stack
// (gray) re-entry; a black node is a legitimately revisited shared node in
// an acyclic subgraph, never a cycle. The black set is also what bounds
// the traversal at O(V+E) rather than exponentially on a diamond-laden
// graph (T-07-33) -- Task 3's gray-versus-visited mutation makes that
// distinctness falsifiable rather than merely asserted.
package callgraph

import (
	"fmt"
	"sort"

	"github.com/codename-lang/lang/internal/compiler/core"
)

// color is the three-state DFS marker. The zero value is white
// (unvisited), so a color map never needs pre-population: an absent key
// reads as white.
type color int

const (
	white color = iota
	gray
	black
)

// MaxCycleCauses bounds the number of cycle_member causes check's
// core.call_graph_cycle diagnostic will ever emit (D-07-15). Today's
// language has no construct that could produce a legal, meaningfully
// larger cycle witness than a handful of members, so this bound is a
// genuine, finite, fail-closed ceiling against a future call-graph shape
// (mirroring pathoracle.MaxPaths's doc-comment pattern) rather than a
// tuning knob: it bounds only the emitted DIAGNOSTIC, never the traversal
// itself (Order still visits every node and edge regardless of cycle
// length), and it is never raised to make a test pass.
const MaxCycleCauses = 32

// TruncatedCycleBound is the stable truncation code check's
// core.call_graph_cycle diagnostic reports when a cycle's true member
// count exceeds MaxCycleCauses: check appends exactly one cause carrying
// this code and drops the overflow, matching the
// "truncated:evidence.trace_bound" / "truncated:explain.depth" /
// "truncated:query.page_bound" convention. Declared here (not in
// protocol, which would close an import cycle through interp -- see
// protocol.go's own note) alongside its sibling bound, MaxCycleCauses.
const TruncatedCycleBound = "truncated:core.call_cycle_bound"

// rotationDisabledForTest and witnessSelectionDisabledForTest are Task 2's
// own fault-injection seams (D-07-42): unexported, false by production
// default, exercised ONLY via callgraph_export_test.go's setters (same
// shape as session's phase07-lane overrides), restored via defer in every
// test that engages them, never assigned on a production code path. They
// exist to prove D-07-16's canonical rotation and D-07-43's
// lexicographically-smallest cross-cycle witness selection are each
// independently load-bearing: disabling rotation alone breaks the
// same-cycle/different-discovery-order identity stability test; disabling
// witness selection alone breaks the two-distinct-cycles test, with
// rotation still applied to whichever candidate happened to be found
// first.
var (
	rotationDisabledForTest         = false
	witnessSelectionDisabledForTest = false
)

// cycleError is callgraph's private witness-carrying typed error for a
// discovered cycle: the canonically-rotated member function IDs
// (D-07-16 -- the lexicographically smallest function ID sits at index 0)
// and, parallel to members, the OpCall operation ID that realizes each
// member's own outgoing edge to the NEXT member (wrapping around), so
// check's diagnostic construction can project every cycle_member cause's
// span, and the diagnostic's own Primary span, from operation IDs without
// re-deriving anything (D-07-35). When several cycles exist in the graph,
// this is the one whose canonical rotation is lexicographically smallest
// across every cycle discovered (D-07-43) -- deterministic regardless of
// root or adjacency traversal order.
type cycleError struct {
	members          []string
	edgeOperationIDs []string
}

func (e *cycleError) Error() string {
	return fmt.Sprintf("callgraph.cycle: call graph contains a cycle of length %d: %v", len(e.members), e.members)
}

// Code reports core.CallGraphCycle, the stable diagnostic code SEM-07
// names, shared verbatim (as an inert string constant) with corevalidate's
// own independent re-derivation in 07-07.
func (e *cycleError) Code() string { return core.CallGraphCycle }

// CycleError type-asserts err as callgraph's own cycle witness, mirroring
// the errors.As-style accessor convention used elsewhere in this
// repository (pathoracle.PathCapError, session.EngineMismatch).
func CycleError(err error) (*cycleError, bool) {
	e, ok := err.(*cycleError)
	return e, ok
}

// Members reports the cycle's witness path of function IDs, canonically
// rotated (D-07-16) so the lexicographically smallest function ID sits at
// index 0.
func (e *cycleError) Members() []string { return append([]string(nil), e.members...) }

// MemberEdgeOperationIDs reports, parallel to Members, the OpCall
// operation ID realizing each member's own outgoing edge to the next
// member in the (already-rotated) cycle order, wrapping from the last
// member back to the first (D-07-35): check projects each cycle_member
// cause's span from these operation IDs, never from a Span this package
// or core.LinearOperation carries.
func (e *cycleError) MemberEdgeOperationIDs() []string {
	return append([]string(nil), e.edgeOperationIDs...)
}

// ClosingOperationID reports the OpCall operation ID whose edge closes the
// canonical cycle -- the wrap-around edge from the last member back to the
// first -- the span check projects onto the diagnostic's own Primary.
func (e *cycleError) ClosingOperationID() string {
	if len(e.edgeOperationIDs) == 0 {
		return ""
	}
	return e.edgeOperationIDs[len(e.edgeOperationIDs)-1]
}

// unresolvedCalleeError is callgraph's own independent re-derivation of
// D-07-45's unresolved-callee refusal: an OpCall whose CalleeID names no
// declared function must never be silently dropped from the graph -- a
// dropped edge is precisely how a cycle escapes detection -- so this
// package refuses with the SAME stable code check and corevalidate each
// independently re-derive on their own side, rather than skipping the
// edge. This is exercised directly against a forged/synthetic core.Program
// (T-07-35): callgraph must behave correctly on programs no parser
// produced, never merely on programs check has already vetted.
type unresolvedCalleeError struct {
	functionID, operationID, calleeID string
}

func (e *unresolvedCalleeError) Error() string {
	return fmt.Sprintf(
		"callgraph.unresolved_callee: function %q operation %q names callee %q, which resolves to no declared function",
		e.functionID, e.operationID, e.calleeID,
	)
}

// Code reports core.CallCalleeUnresolved -- D-07-45's identity, minted in
// 07-03 -- never a fourth spelling of the cycle code.
func (e *unresolvedCalleeError) Code() string { return core.CallCalleeUnresolved }

// UnresolvedCalleeError type-asserts err as callgraph's own unresolved-
// callee refusal.
func UnresolvedCalleeError(err error) (*unresolvedCalleeError, bool) {
	e, ok := err.(*unresolvedCalleeError)
	return e, ok
}

// buildAdjacency enumerates every declared function as a node (D-07-30:
// node identity is the function ID) and every core.OpCall edge (D-07-29:
// edges come from CalleeID and nothing else), deduped so E is bounded by
// V-squared rather than by raw operation count, with adjacency lists
// sorted by callee ID at build time (load-bearing for Task 2's
// deterministic witness selection). It also returns, per edge, the
// lexicographically smallest OpCall operation ID that realizes it -- the
// representative edge operation check projects a span from (D-07-35).
//
// An OpCall whose CalleeID resolves to no declared function REFUSES
// (unresolvedCalleeError) rather than being silently dropped -- a dropped
// edge is how a cycle escapes detection (D-07-45).
func buildAdjacency(program core.Program) (map[string][]string, map[[2]string]string, error) {
	declared := make(map[string]bool, len(program.Functions))
	for _, function := range program.Functions {
		declared[function.ID] = true
	}

	type edgeKey = [2]string
	edgeSeen := make(map[edgeKey]bool)
	edgeOperation := make(map[edgeKey]string)

	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind != core.OpCall {
				continue
			}
			if !declared[operation.CalleeID] {
				return nil, nil, &unresolvedCalleeError{
					functionID: function.ID, operationID: operation.ID, calleeID: operation.CalleeID,
				}
			}
			key := edgeKey{function.ID, operation.CalleeID}
			edgeSeen[key] = true
			if existing, ok := edgeOperation[key]; !ok || operation.ID < existing {
				edgeOperation[key] = operation.ID
			}
		}
	}

	adjacency := make(map[string][]string, len(declared))
	for id := range declared {
		adjacency[id] = nil
	}
	for key := range edgeSeen {
		adjacency[key[0]] = append(adjacency[key[0]], key[1])
	}
	for id := range adjacency {
		sort.Strings(adjacency[id])
	}
	return adjacency, edgeOperation, nil
}

// stackFrame is one entry of the DFS's own explicit stack: the node's own
// ID and the index of the next adjacency-list child still to visit. Using
// an explicit stack of these, rather than a Go call stack, is what makes
// the traversal cost no Go stack (D-07-17/D-07-18).
type stackFrame struct {
	id             string
	nextChildIndex int
}

// Order runs an iterative three-color (white/gray/black) depth-first
// search over program's whole call graph, roots = ALL declared functions
// (sorted by function ID for determinism), never merely the entry point.
// It returns the graph's reverse postorder on success.
//
// On a cycle, it returns a *cycleError (see CycleError) built from the
// on-stack (gray) re-entry that discovered it -- never from a visit count
// or a depth ceiling (D-07-17/D-07-18): a node already black is a
// legitimately revisited shared node in an acyclic subgraph, not a cycle.
// The traversal itself is never bounded; only the diagnostic callgraph's
// caller ultimately emits from a *cycleError is (Task 2).
func Order(program core.Program) ([]string, error) {
	adjacency, edgeOperation, err := buildAdjacency(program)
	if err != nil {
		return nil, err
	}

	roots := make([]string, 0, len(adjacency))
	for id := range adjacency {
		roots = append(roots, id)
	}
	sort.Strings(roots)

	colorOf := make(map[string]color, len(adjacency))
	onStackIndex := make(map[string]int, len(adjacency))
	var postorder []string
	var candidates []cycleWitness

	for _, root := range roots {
		if colorOf[root] != white {
			continue
		}
		stack := []stackFrame{{id: root}}
		colorOf[root] = gray
		onStackIndex[root] = 0

		for len(stack) > 0 {
			top := &stack[len(stack)-1]
			children := adjacency[top.id]
			if top.nextChildIndex < len(children) {
				child := children[top.nextChildIndex]
				top.nextChildIndex++
				switch colorOf[child] {
				case white:
					colorOf[child] = gray
					stack = append(stack, stackFrame{id: child})
					onStackIndex[child] = len(stack) - 1
				case gray:
					// A back edge to an on-stack (gray) node is a real
					// cycle. Record it as a candidate and KEEP GOING --
					// D-07-43 requires collecting every cycle the whole
					// deterministic traversal discovers, then choosing the
					// lexicographically-smallest-rotation witness among
					// them at the very end, rather than short-circuiting
					// on whichever cycle a single root ordering happens to
					// reach first.
					startIndex := onStackIndex[child]
					members := make([]string, 0, len(stack)-startIndex)
					for i := startIndex; i < len(stack); i++ {
						members = append(members, stack[i].id)
					}
					edgeIDs := make([]string, len(members))
					for i := range members {
						next := members[(i+1)%len(members)]
						edgeIDs[i] = edgeOperation[[2]string{members[i], next}]
					}
					candidates = append(candidates, cycleWitness{members: members, edgeOperationIDs: edgeIDs})
				case black:
					// A legitimately revisited shared node in an acyclic
					// subgraph (a diamond's shared leaf) -- not a cycle.
					// This is the exact distinction the gray-versus-
					// visited mutation (Task 3) collapses.
				}
			} else {
				colorOf[top.id] = black
				delete(onStackIndex, top.id)
				postorder = append(postorder, top.id)
				stack = stack[:len(stack)-1]
			}
		}
	}

	if len(candidates) > 0 {
		best := canonicalRotate(candidates[0])
		if !witnessSelectionDisabledForTest {
			for _, candidate := range candidates[1:] {
				rotated := canonicalRotate(candidate)
				if lexLess(rotated.members, best.members) {
					best = rotated
				}
			}
		}
		return nil, &cycleError{members: best.members, edgeOperationIDs: best.edgeOperationIDs}
	}

	reversePostorder := make([]string, len(postorder))
	for i, id := range postorder {
		reversePostorder[len(postorder)-1-i] = id
	}
	return reversePostorder, nil
}

// cycleWitness is one candidate cycle discovered mid-traversal: its
// (not-yet-rotated) member function IDs in discovery order, and, parallel
// to members, the OpCall operation ID realizing each member's own
// outgoing edge to the next member, wrapping around.
type cycleWitness struct {
	members          []string
	edgeOperationIDs []string
}

// canonicalRotate rotates a cycle witness so the lexicographically
// smallest function ID sits at index 0 (D-07-16), rotating members and
// their parallel edgeOperationIDs in lockstep so each member's own
// outgoing-edge operation ID stays correctly paired after rotation.
func canonicalRotate(witness cycleWitness) cycleWitness {
	n := len(witness.members)
	if n == 0 || rotationDisabledForTest {
		return witness
	}
	minIndex := 0
	for i := 1; i < n; i++ {
		if witness.members[i] < witness.members[minIndex] {
			minIndex = i
		}
	}
	members := make([]string, n)
	edgeIDs := make([]string, n)
	for i := 0; i < n; i++ {
		members[i] = witness.members[(minIndex+i)%n]
		edgeIDs[i] = witness.edgeOperationIDs[(minIndex+i)%n]
	}
	return cycleWitness{members: members, edgeOperationIDs: edgeIDs}
}

// lexLess reports whether a sorts before b under ordinary element-wise
// lexicographic slice comparison -- the D-07-43 tie-break rule selecting
// one witness among several distinct cycles.
func lexLess(a, b []string) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return len(a) < len(b)
}
