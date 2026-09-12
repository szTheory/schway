package session

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "embed"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// ---------------------------------------------------------------------
// QLT-03: the call-graph-shape reachability register (11-06-PLAN.md).
//
// This is a NEW artifact, deliberately NOT an extension of
// qlt01_registry.json (D-11-43): qlt01's row key is spike_id over an
// unrelated universe (a spike's control mechanism), while this register's
// row key is an axis-product CELL over the call-graph-shape taxonomy.
// Merging the two would make both audits ambiguous about what a missing
// row means.
//
// Rows are an axis PRODUCT, not a list (D-11-44): a cell cannot be absent,
// and an unclassified cell defaults to `gap` and FAILS the audit. Cells
// are enumerated mechanically (qlt03EnumerateCells) and `reached` is
// recomputed mechanically from testsupport.GenerateCallGraphCorpus's own
// sweep; only `unreachable` rows are hand-written, and every hand-written
// row names a falsifier test that goes red if the claim stops being true.
//
// The taxonomy is five axes (D-11-45): edge topology; argument mode across
// the edge; callee body shape; returned-value origin class; composition-
// depth bucket. Rather than the full 5-axis product (2,560 cells,
// explicitly rejected), the register's row set is the union of two
// designed slices: a FULL enumeration of argument-mode x return-origin
// (20 cells, indexed by those two axes alone) plus PAIRWISE coverage over
// the remaining three axes (edge topology x callee body shape x
// composition-depth bucket, 15 cells via a Latin-square-style
// construction that covers every pairwise combination of those three axes
// without requiring the full 5x3x3 product). This is why a QLT03Row's
// CellKey is built from only the axes ITS OWN slice varies over, not a
// fixed 5-tuple: the two slices probe genuinely different corners of the
// taxonomy, and forcing every row into one 5-tuple shape would require
// picking meaningless placeholder values for the axes a given slice does
// not vary, which is exactly the kind of decorative completeness this
// register's own must_haves reject.
// ---------------------------------------------------------------------

//go:embed qlt03_shape_register.json
var qlt03RegisterBytes []byte

// ---- Axis 1: edge topology -------------------------------------------

// QLT03EdgeTopologies is declared identically to
// testsupport.CallGraphCorpusShapes() -- TestQLT03AxisCompletenessAgainstCorpusShapes
// cross-checks the two mechanically so they can never silently drift apart
// (a shape the generator produces that this axis does not name is a
// register gap).
func QLT03EdgeTopologies() []string {
	return testsupport.CallGraphCorpusShapes()
}

// ---- Axis 2: argument mode across the edge -----------------------------

const (
	ArgMove            = "move"
	ArgCopy            = "copy"
	ArgBorrowShared    = "borrow_shared"
	ArgBorrowExclusive = "borrow_exclusive"
)

// QLT03ArgumentModes is the closed, four-value argument-mode axis (D-11-45).
func QLT03ArgumentModes() []string {
	return []string{ArgMove, ArgCopy, ArgBorrowShared, ArgBorrowExclusive}
}

// ---- Axis 3: callee body shape -----------------------------------------

const (
	ShapeRelay    = "relay"
	ShapeLeafUse  = "leaf_use"
	ShapeLeafPass = "leaf_pass"
)

// QLT03CalleeBodyShapes is the closed, three-value callee-body-shape axis,
// matching testsupport/callgraphcorpus.go's own three body-building
// primitives (corpusRelay, corpusLeafUse, corpusLeafPass) exactly -- every
// function the generator can build is mechanically one of these three
// shapes, determined structurally (qlt03ClassifyBodyShape) rather than by
// trusting which builder function was called.
func QLT03CalleeBodyShapes() []string {
	return []string{ShapeRelay, ShapeLeafUse, ShapeLeafPass}
}

// ---- Axis 4: returned-value origin class --------------------------------

const (
	RetFreshOwned           = "fresh_owned"
	RetParamForwarded       = "param_forwarded"
	RetBorrowOfOwnParam     = "borrow_of_own_param"
	RetBorrowOfCalleeResult = "borrow_of_callee_result"
	RetForeignDerived       = "foreign_derived"
)

// QLT03ReturnOrigins is the closed, five-value return-origin axis (D-11-45).
func QLT03ReturnOrigins() []string {
	return []string{RetFreshOwned, RetParamForwarded, RetBorrowOfOwnParam, RetBorrowOfCalleeResult, RetForeignDerived}
}

// ---- Axis 5: composition-depth bucket -----------------------------------

const (
	DepthOne      = "depth_1"
	DepthTwoThree = "depth_2_3"
	DepthFourPlus = "depth_4_plus"
)

// QLT03CompositionDepthBuckets is the closed, three-value composition-depth
// axis.
func QLT03CompositionDepthBuckets() []string {
	return []string{DepthOne, DepthTwoThree, DepthFourPlus}
}

func qlt03DepthBucket(depth int) string {
	switch {
	case depth <= 1:
		return DepthOne
	case depth <= 3:
		return DepthTwoThree
	default:
		return DepthFourPlus
	}
}

// ---- Q-03: the committed generator op-kind literal ----------------------

// QLT03CommittedGeneratorOpKinds is the committed literal
// TestQLT03GeneratorOpKindClosure compares the generator's ACTUALLY
// observed op-kind set against -- never against a second recomputation.
// The predicted value, per 11-CONTEXT.md's Q-03 experiment and Phase 09's
// own finding that the generator contains no borrow op at all, is exactly
// {call, copy, return}. Do NOT edit this literal to match a surprising
// observation: if the observed set diverges, the axes are wrong and the
// taxonomy must be revised in PHASE-11-DEBT.md BEFORE any row is written
// (this task's own escape hatch).
func QLT03CommittedGeneratorOpKinds() []core.OperationKind {
	return []core.OperationKind{core.OpCall, core.OpCopy, core.OpReturn}
}

// qlt03SweepSizes are the three corpus sizes every mechanical recomputation
// in this file sweeps -- matching 11-06-PLAN.md's own n in {8, 24, 48}.
var qlt03SweepSizes = []int{8, 24, 48}

// QLT03ObservedGeneratorOpKinds unions the op-kind set actually emitted by
// every GenerateCallGraphCorpus(shape, n) across every CallGraphCorpusShapes()
// shape and every qlt03SweepSizes size, sorted for determinism.
func QLT03ObservedGeneratorOpKinds() ([]core.OperationKind, error) {
	seen := map[core.OperationKind]bool{}
	for _, n := range qlt03SweepSizes {
		for _, shape := range QLT03EdgeTopologies() {
			program, err := testsupport.GenerateCallGraphCorpus(shape, n)
			if err != nil {
				return nil, fmt.Errorf("qlt03: generating shape %q at n=%d: %w", shape, n, err)
			}
			for _, fn := range program.Functions {
				if fn.Linear == nil {
					continue
				}
				for _, op := range fn.Linear.Operations {
					seen[op.Kind] = true
				}
			}
		}
	}
	out := make([]core.OperationKind, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

// qlt03OpKindSetsEqual compares two OperationKind sets as sets (order
// independent), used by TestQLT03GeneratorOpKindClosure to compare the
// observed set against QLT03CommittedGeneratorOpKinds().
func qlt03OpKindSetsEqual(a, b []core.OperationKind) bool {
	if len(a) != len(b) {
		return false
	}
	setA := map[core.OperationKind]bool{}
	for _, k := range a {
		setA[k] = true
	}
	for _, k := range b {
		if !setA[k] {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------
// Structural classifiers -- purely mechanical functions of a
// core.Function/core.LinearOperation, never of which corpus builder
// produced them. These are what make "reached" mechanically recomputed
// rather than trusted from the generator's own intent.
// ---------------------------------------------------------------------

// qlt03FindProducingOp returns the operation within fn whose TargetID is
// place, or nil if no operation in this function produces that place
// (e.g. place is the function's own parameter, produced by nothing).
func qlt03FindProducingOp(fn core.Function, place string) *core.LinearOperation {
	if fn.Linear == nil {
		return nil
	}
	for i := range fn.Linear.Operations {
		if fn.Linear.Operations[i].TargetID == place {
			return &fn.Linear.Operations[i]
		}
	}
	return nil
}

// qlt03ClassifyArgumentMode classifies how the value crossing a given
// OpCall's edge arrived at the call site: the PROXIMATE producer of the
// call's SourceID place. A raw, unproduced place (the function's own
// parameter, or any other place with no producing operation) defaults to
// "move" -- the ability-decided default for an owned, single-ownership
// value absent an explicit copy/borrow annotation (D-07-11's
// consume-on-call precedent).
func qlt03ClassifyArgumentMode(fn core.Function, call core.LinearOperation) string {
	if call.SourceID == fn.Parameter.ID {
		return ArgMove
	}
	producer := qlt03FindProducingOp(fn, call.SourceID)
	if producer == nil {
		return ArgMove
	}
	switch producer.Kind {
	case core.OpBorrowShared:
		return ArgBorrowShared
	case core.OpBorrowExclusive:
		return ArgBorrowExclusive
	case core.OpCopy:
		return ArgCopy
	default:
		return ArgMove
	}
}

// qlt03ClassifyBodyShape classifies a function's own body shape: "relay" if
// it contains at least one OpCall; otherwise a leaf, distinguished by
// whether its terminal OpReturn sources the bare parameter ("leaf_pass") or
// a place derived from the parameter through an intervening operation such
// as OpCopy ("leaf_use"). Mirrors testsupport/callgraphcorpus.go's own
// three body-building primitives structurally, never by trusting which
// builder function was invoked.
func qlt03ClassifyBodyShape(fn core.Function) string {
	if fn.Linear == nil {
		return ShapeLeafPass
	}
	for _, op := range fn.Linear.Operations {
		if op.Kind == core.OpCall {
			return ShapeRelay
		}
	}
	var returnOp *core.LinearOperation
	for i := range fn.Linear.Operations {
		if fn.Linear.Operations[i].Kind == core.OpReturn {
			returnOp = &fn.Linear.Operations[i]
		}
	}
	if returnOp == nil || returnOp.SourceID == fn.Parameter.ID {
		return ShapeLeafPass
	}
	return ShapeLeafUse
}

// qlt03ClassifyReturnOrigin classifies a function's own terminal OpReturn by
// tracing its SourceID backward through Move/Copy pass-through operations
// to the operation that actually PRODUCED the returned value:
//   - the function's own parameter, reached with no intervening borrow ->
//     param_forwarded
//   - an OpBorrowShared/OpBorrowExclusive whose own underlying source
//     traces to the parameter -> borrow_of_own_param; traces to a call
//     result -> borrow_of_callee_result; traces to a foreign call ->
//     foreign_derived (the borrow is over a foreign-derived value)
//   - an OpCall's own target, forwarded with no borrow -> fresh_owned (an
//     owned value received across a call boundary is this function's own
//     fresh owned return, not a forwarding of ITS OWN parameter)
//   - an OpForeignCall's own target -> foreign_derived
//
// This is a purely structural trace over core.LinearOperation, never a
// property of which testsupport builder constructed the function.
func qlt03ClassifyReturnOrigin(fn core.Function) string {
	if fn.Linear == nil {
		return RetFreshOwned
	}
	var returnOp *core.LinearOperation
	for i := range fn.Linear.Operations {
		if fn.Linear.Operations[i].Kind == core.OpReturn {
			returnOp = &fn.Linear.Operations[i]
		}
	}
	if returnOp == nil {
		return RetFreshOwned
	}
	return qlt03TraceReturnOrigin(fn, returnOp.SourceID, map[string]bool{})
}

func qlt03TraceReturnOrigin(fn core.Function, place string, visited map[string]bool) string {
	if place == fn.Parameter.ID {
		return RetParamForwarded
	}
	if visited[place] {
		return RetFreshOwned
	}
	visited[place] = true
	producer := qlt03FindProducingOp(fn, place)
	if producer == nil {
		return RetFreshOwned
	}
	switch producer.Kind {
	case core.OpMove, core.OpCopy:
		return qlt03TraceReturnOrigin(fn, producer.SourceID, visited)
	case core.OpBorrowShared, core.OpBorrowExclusive:
		switch qlt03TraceReturnOrigin(fn, producer.SourceID, visited) {
		case RetParamForwarded:
			return RetBorrowOfOwnParam
		case RetForeignDerived:
			return RetForeignDerived
		default:
			return RetBorrowOfCalleeResult
		}
	case core.OpCall:
		return RetFreshOwned
	case core.OpForeignCall:
		return RetForeignDerived
	default:
		return RetFreshOwned
	}
}

// qlt03EdgeCompositionDepth measures the composition depth of a specific
// OpCall edge within fn: 1 for a call sourced directly from the function's
// own parameter (or any place not itself produced by a prior OpCall in
// this function), and 1 + the depth of the call that produced its source
// place for a chained forward (testdata/phase08's corpusForward shape,
// D-08-35's "forward" corpus).
func qlt03EdgeCompositionDepth(fn core.Function, call core.LinearOperation) int {
	depth := 1
	source := call.SourceID
	visited := map[string]bool{}
	for {
		producer := qlt03FindProducingOp(fn, source)
		if producer == nil || producer.Kind != core.OpCall {
			return depth
		}
		if visited[producer.ID] {
			return depth
		}
		visited[producer.ID] = true
		depth++
		source = producer.SourceID
	}
}

// ---------------------------------------------------------------------
// The mechanical sweep: one observation per OpCall edge across every
// declared shape and n in qlt03SweepSizes.
// ---------------------------------------------------------------------

type qlt03EdgeObservation struct {
	Topology     string
	ArgumentMode string
	CalleeShape  string
	ReturnOrigin string
	DepthBucket  string
	Depth        int
}

// qlt03SweepConfig controls which shapes/op-kinds a sweep includes,
// letting TestQLT03SwarmConfigurations run shape-omitting and
// op-kind-omitting configurations without duplicating the sweep body.
type qlt03SweepConfig struct {
	ExcludeShapes  map[string]bool
	ExcludeOpKinds map[core.OperationKind]bool
}

// qlt03SweepEdges walks every declared shape's generated corpus at size n
// and returns one qlt03EdgeObservation per OpCall edge encountered. cfg may
// be nil for the unconfigured (full) sweep.
func qlt03SweepEdges(n int, cfg *qlt03SweepConfig) ([]qlt03EdgeObservation, error) {
	var observations []qlt03EdgeObservation
	for _, shape := range QLT03EdgeTopologies() {
		if cfg != nil && cfg.ExcludeShapes[shape] {
			continue
		}
		program, err := testsupport.GenerateCallGraphCorpus(shape, n)
		if err != nil {
			return nil, fmt.Errorf("qlt03: generating shape %q at n=%d: %w", shape, n, err)
		}
		byID := make(map[string]core.Function, len(program.Functions))
		for _, fn := range program.Functions {
			byID[fn.ID] = fn
		}
		for _, fn := range program.Functions {
			if fn.Linear == nil {
				continue
			}
			for _, op := range fn.Linear.Operations {
				if op.Kind != core.OpCall {
					continue
				}
				callee, ok := byID[op.CalleeID]
				if !ok {
					continue
				}
				calleeShape := qlt03ClassifyBodyShape(callee)
				// The op-kind-omitting swarm configuration omits OpCopy:
				// the only op-kind whose omission does not collapse the
				// entire edge-observation mechanism (every function needs
				// an OpReturn to terminate, and OpCall IS the edge being
				// observed). Omitting OpCopy means dropping any callee
				// whose own body shape is produced by an OpCopy
				// (leaf_use).
				if cfg != nil && cfg.ExcludeOpKinds[core.OpCopy] && calleeShape == ShapeLeafUse {
					continue
				}
				depth := qlt03EdgeCompositionDepth(fn, op)
				observations = append(observations, qlt03EdgeObservation{
					Topology:     shape,
					ArgumentMode: qlt03ClassifyArgumentMode(fn, op),
					CalleeShape:  calleeShape,
					ReturnOrigin: qlt03ClassifyReturnOrigin(callee),
					DepthBucket:  qlt03DepthBucket(depth),
					Depth:        depth,
				})
			}
		}
	}
	return observations, nil
}

// qlt03Aggregate accumulates ReachedEvidence for one cell across a sweep.
type qlt03Aggregate struct {
	Occurrences int
	Shapes      map[string]bool
	MaxDepth    int
}

func (a *qlt03Aggregate) toEvidence() ReachedEvidence {
	return ReachedEvidence{
		Occurrences:    a.Occurrences,
		DistinctShapes: len(a.Shapes),
		MaxDepth:       a.MaxDepth,
	}
}

// qlt03AggregateByArgReturn aggregates observations across ALL sweep sizes
// into a map keyed by (argument_mode, return_origin), ignoring topology,
// callee shape and depth bucket other than for evidence accounting -- this
// is the FULL argument-mode x return-origin slice (20 cells).
func qlt03AggregateByArgReturn(observations []qlt03EdgeObservation) map[[2]string]*qlt03Aggregate {
	out := map[[2]string]*qlt03Aggregate{}
	for _, obs := range observations {
		key := [2]string{obs.ArgumentMode, obs.ReturnOrigin}
		agg := out[key]
		if agg == nil {
			agg = &qlt03Aggregate{Shapes: map[string]bool{}}
			out[key] = agg
		}
		agg.Occurrences++
		agg.Shapes[obs.Topology] = true
		if obs.Depth > agg.MaxDepth {
			agg.MaxDepth = obs.Depth
		}
	}
	return out
}

// qlt03AggregateByTopologyShapeDepth aggregates observations into a map
// keyed by (edge_topology, callee_body_shape, composition_depth_bucket) --
// the pairwise-coverage slice (15 cells).
func qlt03AggregateByTopologyShapeDepth(observations []qlt03EdgeObservation) map[[3]string]*qlt03Aggregate {
	out := map[[3]string]*qlt03Aggregate{}
	for _, obs := range observations {
		key := [3]string{obs.Topology, obs.CalleeShape, obs.DepthBucket}
		agg := out[key]
		if agg == nil {
			agg = &qlt03Aggregate{Shapes: map[string]bool{}}
			out[key] = agg
		}
		agg.Occurrences++
		agg.Shapes[obs.Topology] = true
		if obs.Depth > agg.MaxDepth {
			agg.MaxDepth = obs.Depth
		}
	}
	return out
}

// qlt03SweepAllSizes runs qlt03SweepEdges at every qlt03SweepSizes entry and
// concatenates the results -- the aggregate sweep the committed register's
// reached evidence is derived from.
func qlt03SweepAllSizes(cfg *qlt03SweepConfig) ([]qlt03EdgeObservation, error) {
	var all []qlt03EdgeObservation
	for _, n := range qlt03SweepSizes {
		obs, err := qlt03SweepEdges(n, cfg)
		if err != nil {
			return nil, err
		}
		all = append(all, obs...)
	}
	return all, nil
}

// ---------------------------------------------------------------------
// Mechanical cell enumeration (D-11-45): the register's declared row set.
// ---------------------------------------------------------------------

const (
	qlt03AxisSetArgReturn          = "argument_mode_x_return_origin"
	qlt03AxisSetTopologyShapeDepth = "edge_topology_x_callee_body_shape_x_composition_depth"

	// qlt03CellSeparator is the declared separator cell identity is joined
	// with (D-11-44): the row key is the axis tuple compared as an exact,
	// case-sensitive string.
	qlt03CellSeparator = "|"
)

// qlt03AxisOrder fixes a canonical axis ordering per axis set so the same
// logical cell always produces the same key regardless of map iteration
// order.
var qlt03AxisOrder = map[string][]string{
	qlt03AxisSetArgReturn:          {"argument_mode", "return_origin"},
	qlt03AxisSetTopologyShapeDepth: {"edge_topology", "callee_body_shape", "composition_depth_bucket"},
}

// QLT03CellKey builds the declared cell-identity string for a given axis
// set and axis-value map: the axis set name, then each axis's own
// "name=value" pair in canonical order, joined by qlt03CellSeparator.
// Compared as an exact, case-sensitive string (D-11-44) -- no case-folding,
// no whitespace trimming performed here; qlt03NearDuplicateKey (below)
// performs the SEPARATE near-duplicate check TestQLT03CellIdentityIsExactStringEquality
// exercises.
func QLT03CellKey(axisSet string, axes map[string]string) string {
	order := qlt03AxisOrder[axisSet]
	parts := make([]string, 0, len(order)+1)
	parts = append(parts, axisSet)
	for _, axis := range order {
		parts = append(parts, axis+"="+axes[axis])
	}
	return strings.Join(parts, qlt03CellSeparator)
}

// qlt03NearDuplicateFold normalizes a cell key for the NEAR-duplicate check
// only (trim surrounding whitespace, fold ASCII case) -- never used for the
// register's own real cell identity, which stays exact and case-sensitive.
// Two rows whose RAW keys fold to the same normalized string but are not
// byte-identical are reported as a duplicate-row error rather than
// silently accepted as two distinct cells (D-11-44's own worked edge case).
func qlt03NearDuplicateFold(key string) string {
	folded := strings.ToLower(key)
	folded = strings.Join(strings.Fields(folded), "")
	return folded
}

// qlt03NearDuplicateRowFailures returns one failure string per pair of rows
// whose RAW cell keys are NOT byte-identical but fold (case+whitespace) to
// the same normalized string -- D-11-44's own worked edge case.
func qlt03NearDuplicateRowFailures(rows []QLT03Row) []string {
	var failures []string
	folded := map[string][]string{}
	for _, row := range rows {
		key := row.CellKey()
		fold := qlt03NearDuplicateFold(key)
		folded[fold] = append(folded[fold], key)
	}
	for fold, keys := range folded {
		if len(keys) < 2 {
			continue
		}
		distinct := map[string]bool{}
		for _, k := range keys {
			distinct[k] = true
		}
		if len(distinct) > 1 {
			failures = append(failures, fmt.Sprintf("control:qlt03.duplicate_row_case_or_whitespace_variant fold=%s keys=%v", fold, keys))
		}
	}
	return failures
}

// QLT03CellSpec is one designed cell in the register's mechanical
// enumeration: an axis set plus the concrete axis values that identify it.
type QLT03CellSpec struct {
	AxisSet string
	Axes    map[string]string
}

// Key returns this cell's declared identity string.
func (c QLT03CellSpec) Key() string {
	return QLT03CellKey(c.AxisSet, c.Axes)
}

// qlt03EnumerateCells mechanically enumerates the register's declared row
// set (D-11-45): a FULL argument-mode x return-origin product (4 x 5 = 20
// cells) plus PAIRWISE coverage over the remaining three axes (edge
// topology x callee body shape x composition-depth bucket), built via a
// Latin-square-style formula (depth bucket = (topology index + shape
// index) mod 3) that covers every pairwise combination of
// (topology, depth) and (shape, depth) while the (topology, shape) pair is
// the full 5x3=15 product already. This totals 35 rows -- inside the
// plan's own "roughly 40" target, and nowhere near the full 2,560-cell
// product the plan explicitly rejects (D-11-45).
func qlt03EnumerateCells() []QLT03CellSpec {
	var cells []QLT03CellSpec
	for _, am := range QLT03ArgumentModes() {
		for _, ro := range QLT03ReturnOrigins() {
			cells = append(cells, QLT03CellSpec{
				AxisSet: qlt03AxisSetArgReturn,
				Axes:    map[string]string{"argument_mode": am, "return_origin": ro},
			})
		}
	}
	topologies := QLT03EdgeTopologies()
	shapes := QLT03CalleeBodyShapes()
	buckets := QLT03CompositionDepthBuckets()
	for ti, topology := range topologies {
		for si, shape := range shapes {
			depth := buckets[(ti+si)%len(buckets)]
			cells = append(cells, QLT03CellSpec{
				AxisSet: qlt03AxisSetTopologyShapeDepth,
				Axes: map[string]string{
					"edge_topology":            topology,
					"callee_body_shape":        shape,
					"composition_depth_bucket": depth,
				},
			})
		}
	}
	return cells
}

// qlt03InjectExtraRequiredCellForTest is Task 3's D-11-50 fault-injection
// seam: when true, the cell enumeration AuditQLT03Register consults
// includes one additional cell no committed row can ever satisfy, proving
// the audit is load-bearing (able to FAIL) without editing the committed
// qlt03_shape_register.json. Unexported, false in production, set only via
// SetQLT03InjectExtraRequiredCellForTest (export_test.go) by a same-package
// test that defers the restore immediately.
var qlt03InjectExtraRequiredCellForTest bool

func qlt03EnumerateCellsForAudit() []QLT03CellSpec {
	cells := qlt03EnumerateCells()
	if qlt03InjectExtraRequiredCellForTest {
		cells = append(cells, QLT03CellSpec{
			AxisSet: qlt03AxisSetArgReturn,
			Axes:    map[string]string{"argument_mode": "__fault_injected_for_test__", "return_origin": "__fault_injected_for_test__"},
		})
	}
	return cells
}

// ---------------------------------------------------------------------
// The register's discriminated-union row type and its closed
// proof-mechanism table.
// ---------------------------------------------------------------------

// ReachedEvidence is a mechanically recomputed cell's reach evidence: the
// occurrence count, the number of distinct shapes it was reached in, and
// the maximum composition depth observed. witness is not a boolean
// (D-11-48): Thin() reports whether this evidence counts as
// "reached_thin" -- fewer than 3 occurrences, in exactly 1 shape, at
// max_depth == 1 -- which COUNTS AS A GAP for gating.
type ReachedEvidence struct {
	Occurrences    int `json:"occurrences"`
	DistinctShapes int `json:"distinct_shapes"`
	MaxDepth       int `json:"max_depth"`
}

// Thin implements D-11-48's thin-reach gate.
func (e ReachedEvidence) Thin() bool {
	return e.Occurrences < 3 && e.DistinctShapes == 1 && e.MaxDepth == 1
}

// The closed set of admissible negative-proof mechanisms (D-11-46). The
// loader REJECTS any value outside this set, and rejects free text in
// place of a mechanism.
const (
	MechanismGrammar           = "grammar"
	MechanismRefused           = "refused"
	MechanismGenerator         = "generator"
	MechanismBoundedExhaustive = "bounded_exhaustive"
	MechanismGap               = "gap"
)

// QLT03Mechanisms is the closed, five-value proof-mechanism vocabulary.
func QLT03Mechanisms() []string {
	return []string{MechanismGrammar, MechanismRefused, MechanismGenerator, MechanismBoundedExhaustive, MechanismGap}
}

// UnreachableProof is a hand-written row's negative claim: exactly one
// mechanism from the closed set, plus its supporting evidence. Every
// non-gap row names a real, existing falsifier test by name -- the loader
// rejects a row whose named test does not exist as a symbol in the
// package's own test files.
type UnreachableProof struct {
	Mechanism     string `json:"mechanism"`
	Diagnostic    string `json:"diagnostic,omitempty"`
	Rationale     string `json:"rationale"`
	FalsifierTest string `json:"falsifier_test,omitempty"`
}

// QLT03Row is one register entry: a cell (identified by its own axis set
// and axis values) plus exactly one of Reached/Unreachable -- a
// discriminated union, never both, never neither.
type QLT03Row struct {
	AxisSet     string            `json:"axis_set"`
	Axes        map[string]string `json:"axes"`
	Reached     *ReachedEvidence  `json:"reached,omitempty"`
	Unreachable *UnreachableProof `json:"unreachable,omitempty"`
}

// CellKey returns this row's declared cell-identity string.
func (r QLT03Row) CellKey() string {
	return QLT03CellKey(r.AxisSet, r.Axes)
}

// ---------------------------------------------------------------------
// The loader.
// ---------------------------------------------------------------------

// qlt03PackageTestFuncs scans every *_test.go file in this package's own
// directory (via go/parser, not by importing the test binary) and returns
// the set of declared top-level func names -- used to reject an
// UnreachableProof.FalsifierTest that names a symbol that does not exist.
func qlt03PackageTestFuncs() (map[string]bool, error) {
	dir := filepath.Join(nat03ProjectRoot(), "internal", "compiler", "session")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("qlt03: reading package directory: %w", err)
	}
	funcs := map[string]bool{}
	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			return nil, fmt.Errorf("qlt03: parsing %s: %w", name, err)
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil {
				funcs[fn.Name.Name] = true
			}
		}
	}
	return funcs, nil
}

// LoadQLT03Register parses the embedded register, validates the
// discriminated-union invariant on every row, validates that every
// UnreachableProof.Mechanism is in the closed set (rejecting free text),
// and validates that every non-gap UnreachableProof.FalsifierTest names an
// existing test symbol in this package.
func LoadQLT03Register() ([]QLT03Row, error) {
	var rows []QLT03Row
	if err := json.Unmarshal(qlt03RegisterBytes, &rows); err != nil {
		return nil, fmt.Errorf("qlt03: failed to parse embedded register: %w", err)
	}

	testFuncs, err := qlt03PackageTestFuncs()
	if err != nil {
		return nil, err
	}
	mechanisms := map[string]bool{}
	for _, m := range QLT03Mechanisms() {
		mechanisms[m] = true
	}

	for i, row := range rows {
		hasReached := row.Reached != nil
		hasUnreachable := row.Unreachable != nil
		if hasReached == hasUnreachable {
			return nil, fmt.Errorf("qlt03: row %d (cell=%s) must have exactly one of reached/unreachable populated", i, row.CellKey())
		}
		if !hasUnreachable {
			continue
		}
		if !mechanisms[row.Unreachable.Mechanism] {
			return nil, fmt.Errorf("qlt03: row %d (cell=%s) has out-of-closed-set mechanism %q", i, row.CellKey(), row.Unreachable.Mechanism)
		}
		if row.Unreachable.Mechanism == MechanismGap {
			continue
		}
		if row.Unreachable.FalsifierTest == "" {
			return nil, fmt.Errorf("qlt03: row %d (cell=%s) mechanism %q requires a named falsifier test", i, row.CellKey(), row.Unreachable.Mechanism)
		}
		if !testFuncs[row.Unreachable.FalsifierTest] {
			return nil, fmt.Errorf("qlt03: row %d (cell=%s) names nonexistent falsifier test %q", i, row.CellKey(), row.Unreachable.FalsifierTest)
		}
	}
	return rows, nil
}

// ---------------------------------------------------------------------
// The audit.
// ---------------------------------------------------------------------

// AuditQLT03Register returns a list of audit failures, each naming its own
// control ID in the qlt01.go failure-with-control-ID style. It FAILS when:
// any enumerated cell is absent from the register; any row is
// unclassified; any row has both Reached and Unreachable populated; any
// Reached row is reached_thin (D-11-48, counts as a gap); any Unreachable
// row's mechanism is `gap` or outside the closed set; any two rows share a
// cell key.
func AuditQLT03Register(rows []QLT03Row) []string {
	var failures []string

	byKey := map[string]QLT03Row{}
	seenOrder := make([]string, 0, len(rows))
	duplicate := map[string]bool{}
	for _, row := range rows {
		key := row.CellKey()
		if _, exists := byKey[key]; exists {
			if !duplicate[key] {
				duplicate[key] = true
				failures = append(failures, fmt.Sprintf("control:qlt03.duplicate_cell_key cell=%s", key))
			}
			continue
		}
		byKey[key] = row
		seenOrder = append(seenOrder, key)
	}
	_ = seenOrder

	mechanisms := map[string]bool{}
	for _, m := range QLT03Mechanisms() {
		mechanisms[m] = true
	}

	for _, cell := range qlt03EnumerateCellsForAudit() {
		key := cell.Key()
		row, ok := byKey[key]
		if !ok {
			failures = append(failures, fmt.Sprintf("control:qlt03.cell_absent cell=%s", key))
			continue
		}
		switch {
		case row.Reached == nil && row.Unreachable == nil:
			failures = append(failures, fmt.Sprintf("control:qlt03.row_unclassified cell=%s", key))
		case row.Reached != nil && row.Unreachable != nil:
			failures = append(failures, fmt.Sprintf("control:qlt03.row_both_populated cell=%s", key))
		case row.Reached != nil:
			if row.Reached.Thin() {
				failures = append(failures, fmt.Sprintf("control:qlt03.reached_thin_counts_as_gap cell=%s", key))
			}
		case row.Unreachable != nil:
			if row.Unreachable.Mechanism == MechanismGap {
				failures = append(failures, fmt.Sprintf("control:qlt03.gap cell=%s", key))
			} else if !mechanisms[row.Unreachable.Mechanism] {
				failures = append(failures, fmt.Sprintf("control:qlt03.mechanism_out_of_set cell=%s mechanism=%s", key, row.Unreachable.Mechanism))
			}
		}
	}
	return failures
}
