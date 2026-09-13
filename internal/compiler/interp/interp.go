package interp

import (
	"fmt"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/execution"
)

const Schema = execution.Schema0

// MaxCallDepth bounds interp's own explicit []frame call stack (D-10-21) at
// a genuinely reachable ceiling -- the exact INVERSE of
// pathoracle.MaxPaths' direction (pathoracle.go:56-67). MaxPaths (4096)
// sits well ABOVE its real reachable maximum (64) so it never fires on any
// program the parser can produce today, a deliberately generous ceiling
// held in reserve for a future CFG shape. MaxCallDepth must sit BELOW the
// real structural ceiling instead: the parser's own maxFunctions (1024,
// syntax/parser.go:16) combined with callgraph's cycle refusal (every
// admitted call graph is acyclic, so a program has strictly no more
// distinct functions to call through than it declares) bounds the deepest
// possible call chain at 1024. A bound set anywhere near that ceiling would
// be decorative -- indistinguishable in practice from never firing at all.
// 128 is comfortably below 1024 (a 129-function chain fixture leaves ample
// headroom for scaffolding) and orders of magnitude above any depth
// today's fixtures exercise, so it genuinely fires on a program the
// compiler legitimately admits (D-10-23), never only on a program the
// compiler would already have refused for an unrelated reason. This
// direction is deliberately OPPOSITE MaxPaths': a reader who pattern-
// matches the two constants will otherwise draw the wrong conclusion.
const MaxCallDepth = 128

// callDepthExceededDefectReason is the stable named reason a depth-exceeded
// refusal carries: exceeding MaxCallDepth (or its test-only override)
// produces this Outcome/Event as comparable execution DATA -- modeled like
// the existing core.OpDefect arm (terminalOutcome, below), never a bare Go
// error (D-10-25) -- because Phase 11's five-axis comparator needs the
// refusal as data when diffing against native's structurally different
// limit-hit behavior.
const callDepthExceededDefectReason = "language call depth exceeded MaxCallDepth"

// maxCallDepthOverride is Task 3's Pitfall-4 subprocess-probe seam,
// following TerminatorKindsOverride's discipline (pathoracle.go:69-77):
// nil in production, consulted only through maxCallDepth(). Unlike
// TerminatorKindsOverride, this seam is deliberately unexported -- only a
// same-package test may temporarily raise or disable MaxCallDepth to drive
// the probe's cap-disabled arm; it is never a production-mutable exported
// global.
var maxCallDepthOverride func() int

// maxCallDepth reports the effective call-depth ceiling: maxCallDepthOverride()
// when set (test-only), else the production MaxCallDepth constant.
func maxCallDepth() int {
	if maxCallDepthOverride != nil {
		return maxCallDepthOverride()
	}
	return MaxCallDepth
}

// frameDrainOrderForTest is Task 1's D-10-35 observation seam, following
// maxCallDepthOverride's nil-default discipline: nil in production,
// consulted only through frameDrainOrder(). When set and returning true, it
// REVERSES the stack-traversal order drainStackForAbruptExit walks
// (outermost-first instead of the defined innermost-first, D-10-33) --
// proving via interp.CanonicalBytes divergence that this plan's ordering is
// genuinely OBSERVED by the emitted event stream, never merely computed and
// silently discarded. Unexported, exercised only by same-package tests.
var frameDrainOrderForTest func() bool

// frameDrainOrder reports the stack-INDEX traversal order for draining every
// live frame on an abrupt exit (D-10-33): a strict LIFO flattening,
// innermost frame (the top of stack, the highest index) first, then outward
// to the base frame (index 0) last. This is what makes the abrupt-path
// composition "free": call frames are already a LIFO stack, so walking it
// from the top down IS the innermost-first rule, for any number of live
// frames. frameDrainOrderForTest (D-10-35), when engaged, reverses this to
// outermost-first.
func frameDrainOrder(n int) []int {
	order := make([]int, n)
	for i := 0; i < n; i++ {
		order[i] = n - 1 - i
	}
	if frameDrainOrderForTest != nil && frameDrainOrderForTest() {
		for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
			order[i], order[j] = order[j], order[i]
		}
	}
	return order
}

// drainStackForAbruptExit flattens EVERY live resource across the whole
// frame stack for an abrupt exit -- the foreign process-root landing pad
// extended to N frames, and the SEM-08 depth refusal (PHASE-10-DEBT.md
// D-10-31: these are the ONLY two constructs that skip more than one frame,
// since OpCall is never a terminator and Lang has no exceptions or
// unwinding).
//
// D-10-31 FINDING (stated here, not only in PHASE-10-DEBT.md): core.go:559-564
// documents OpCall as never a terminator, and this language has no
// exceptions and no unwinding -- OpFail is an ordinary typed value
// propagated ONE boundary at a time, never unwinding through several. SEM-09's
// second clause therefore covers EXACTLY these two constructs -- the foreign
// landing pad above, extended here to N frames, and the depth refusal at
// runFrameStack's core.OpCall arm -- and nothing else. A test exercising
// only OpReturn/OpFail popping proves SEM-09's FIRST clause only; it says
// nothing about this function.
//
// Per frameDrainOrder's LIFO flattening, within each frame the
// order is the existing single-frame rule (liveResourcePlaces' own
// first-acquired-order iteration, UNCHANGED by this plan): composing frames
// changes nothing about within-frame ordering. Returns both the ordered
// "resource.leaked" events and the flattened live_resources place list,
// both an empty (never nil) slice when nothing is live anywhere on the
// stack. Every event's FunctionID is the OWNING frame's own function ID
// (D-10-32) -- never a depth value -- so this is sound only as long as
// recursion stays refused by callgraph (a function cannot appear twice on
// the stack, so no two live frames ever share a FunctionID).
func drainStackForAbruptExit(stack []frame) ([]Event, []string) {
	events := []Event{}
	liveResources := []string{}
	leakIndex := 0
	for _, frameIndex := range frameDrainOrder(len(stack)) {
		f := &stack[frameIndex]
		places := liveResourcePlaces(f.operations, f.live, f.liveOrder)
		for _, place := range places {
			events = append(events, Event{
				Schema: execution.Schema1, ID: fmt.Sprintf("%s:event:leaked:%d", f.function.ID, leakIndex), Kind: "resource.leaked", FunctionID: f.function.ID,
				SourcePlace: place,
			})
			leakIndex++
		}
		liveResources = append(liveResources, places...)
	}
	return events, liveResources
}

// moveAsCopyForTest is Task 2's D-10-41 fault-injection seam (QLT-08): when
// true, partitionFrameForCall skips the caller-side delete of the argument
// place, turning a move into a copy -- the caller's moved-from place stays
// readable after the call, exactly the OWN-05b violation this seam proves
// only interp itself (never check or corevalidate, which execute nothing)
// can catch. Unexported, false by default, exercised only by the
// same-package test TestMoveAsCopyMutationKilled (interp_test.go): never an
// exported package-level mutable var on a production path.
var moveAsCopyForTest = false

type Outcome = execution.Outcome
type Event = execution.Event
type Execution = execution.Execution

func Run(program core.Program, functionName, input string) (Execution, error) {
	validated := corevalidate.Validate(program)
	if !validated.Valid {
		return Execution{}, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	program = validated.Program()
	function, ok := findFunction(program, functionName)
	if !ok {
		return Execution{}, fmt.Errorf("function %q is absent from checked core", functionName)
	}
	if !function.HasClosedBody() {
		return Execution{}, fmt.Errorf("function %q has invalid body union", functionName)
	}
	if function.Linear != nil && function.Match == nil {
		return runLinear(program, function, input)
	}
	for _, arm := range function.Match.Arms {
		if arm.Pattern != input {
			continue
		}
		if arm.BlockID == "" {
			event := Event{
				Schema: Schema, ID: arm.ID + ":event:returned", Kind: "function.returned",
				FunctionID: function.ID, Input: input, Output: arm.Value,
			}
			return Execution{
				Schema: Schema, Outcome: Outcome{Kind: "returned", Value: arm.Value},
				Events: []Event{event}, LiveResources: []string{},
			}, nil
		}
		return runBranchArm(program, function, arm, input)
	}
	return Execution{}, fmt.Errorf("checked match %q has no arm for %q", function.Match.ID, input)
}

// runBranchArm executes the selected match arm's own block on interp's
// explicit frame stack (D-10-21/D-10-39): the unselected arms' operations
// are never visited, so they produce no events (03-01-02's behavior
// clause). Its own core.OpCall arm now routes through partitionFrameForCall
// and runFrameStack -- the same helper and driver runLinear uses -- rather
// than a hand-copied loop (D-10-39 permits this ONE in-package helper
// across all three execution paths).
func runBranchArm(program core.Program, function core.Function, arm core.MatchArm, input string) (Execution, error) {
	found := false
	for _, candidate := range function.Linear.Blocks {
		if candidate.ID == arm.BlockID {
			found = true
			break
		}
	}
	if !found {
		return Execution{}, fmt.Errorf("branch arm %q references unknown block %q", arm.ID, arm.BlockID)
	}
	base := newArmFrame(function, map[string]value{function.Parameter.ID: {tag: "", payload: input}}, arm.BlockID)
	return runFrameStack(program, base)
}

// runLinearBlocks executes a fallible-call function's Blocks/Edges
// (D-04-04) on interp's explicit frame stack (D-10-21/D-10-39), starting at
// the entry block. An OpForeignCall never terminates its block by itself --
// per Claude's Discretion (04-PATTERNS Pattern 3 note), the interpreter
// cannot actually call C, so it models the call as a fixed literal-outcome
// stub that always succeeds this phase, unconditionally following the ok
// edge. Its own core.OpCall arm routes through the SAME
// partitionFrameForCall helper and runFrameStack driver runLinear and
// runBranchArm use, rather than a third hand-copied loop.
func runLinearBlocks(program core.Program, function core.Function, input string) (Execution, error) {
	base := newBlockFrame(function, map[string]value{function.Parameter.ID: {tag: "", payload: input}}, function.ID+":block:entry")
	return runFrameStack(program, base)
}

// nonlocalExitDefectReason is duplicated VERBATIM from cgen.go's own
// constant of the same name -- the two literal strings are kept in sync by
// comment and convention on both sides, not by import, since interp and
// cgen model the SAME shared probe convention through entirely different
// mechanisms (a Go call counter here, a real longjmp there).
const nonlocalExitDefectReason = "foreign nonlocal exit detected at process-root landing pad"

// liveResourcePlaces projects the live-tracking map into the acquisition's
// own TARGET PLACE id, not its operation id -- the same identifier
// convention cgen.go's resourceLedger uses for its own lang_resource_ids
// array, so the two engines report the identical strings for a nonlocal-
// exit-triggered defect's live_resources field and its per-event
// SourcePlace. This is a narrower, DIFFERENT convention than
// liveResourceList's operation-id shape (used by every OTHER terminator
// this phase), scoped only to the nonlocal-exit path this plan adds.
func liveResourcePlaces(operations map[string]core.LinearOperation, live map[string]bool, order []string) []string {
	result := []string{}
	for _, id := range order {
		if live[id] {
			result = append(result, operations[id].TargetID)
		}
	}
	return result
}

// liveResourceList projects the live-tracking map into the deterministic
// slice Execution.LiveResources carries: acquisition op IDs still marked
// live, in first-acquired order. Returns an empty (never nil) slice when
// nothing is live, matching every pre-plan-02 terminal record's shape.
func liveResourceList(live map[string]bool, order []string) []string {
	result := []string{}
	for _, id := range order {
		if live[id] {
			result = append(result, id)
		}
	}
	return result
}

// RunLinearBlockDirect (task 04-07-02's narrow, test-only entry point) and
// its foreignFailureLiteralDirect helper moved to the interptestdirect
// package (WR-02 of the Phase 4 code review): the function's own doc
// comment stated plainly that it "is not used by any production CLI path",
// so it should not ship as an exported, discoverable member of interp's
// real production API surface. See interp/interptestdirect's package doc
// for the full rationale and the moved implementation.

func CanonicalBytes(execution Execution) ([]byte, error) {
	return execution2bytes(execution)
}

func execution2bytes(value Execution) ([]byte, error) { return execution.CanonicalBytes(value) }

func runLinear(program core.Program, function core.Function, input string) (Execution, error) {
	// A straight-line body that carries Blocks/Edges is Phase 4's fallible-
	// call shape (D-04-04): the flat Operations list alone is not enough to
	// know which edge to follow, so it is walked block-by-block instead.
	// Every pre-Phase-4 linear (non-Match) function leaves Blocks empty --
	// only checkFallibleLinear ever populates it for a Match-less function --
	// so this branch changes nothing for any existing program.
	if len(function.Linear.Blocks) > 0 {
		return runLinearBlocks(program, function, input)
	}
	base := newFlatFrame(function, map[string]value{function.Parameter.ID: {tag: "", payload: input}})
	return runFrameStack(program, base)
}

// frame is one activation record on interp's own explicit call stack
// (D-10-21): a heap-allocated slice element standing in for what would
// otherwise be a native Go call, so Lang-level call depth costs O(1) Go
// stack regardless of how deeply a Lang program nests calls -- the same
// iterative, explicit-stack discipline this compiler already applies to
// its own whole-program traversals elsewhere (the SHAPE is reused, never
// the code: this package imports nothing from that traversal, D-10-39).
//
// Run's own top-level single-frame execution becomes the base element of
// this stack (D-10-26), never a special case: the base frame is simply
// the one frame with hasCaller == false.
type frame struct {
	function core.Function

	// values is this frame's OWN place map (D-10-26): a moved-from place
	// in a caller's frame is genuinely absent from this map -- not merely
	// hidden -- and a callee's frame starts as a disjoint namespace seeded
	// only by partitionFrameForCall.
	values    map[string]value
	live      map[string]bool
	liveOrder []string

	// returnTarget is the place ID in this frame's OWN CALLER that the
	// callee's returned value binds into when this frame pops on a
	// core.OpReturn. Only meaningful when hasCaller is true.
	returnTarget string
	hasCaller    bool

	// operations, blocks, and edges index this frame's OWN function body
	// by ID, built once when the frame is pushed. blocks/edges are nil
	// for a flat (non-block) body.
	operations map[string]core.LinearOperation
	blocks     map[string]core.Block
	edges      map[string]core.Edge

	// Execution cursor. For a flat body (blocks == nil), ids holds the
	// function's own ordered operation ID list and idx indexes into it
	// directly. For a block-based body, currentBlockID names the active
	// block and idx indexes into that block's own OperationIDs.
	ids            []string
	idx            int
	currentBlockID string
	// singleBlockOnly is true only for a match-arm frame: such a frame
	// never follows a block successor -- exhausting its one block without
	// a terminator is an error, exactly as before this plan.
	singleBlockOnly bool

	// tracked/nonlocalExitPolicy/nonlocalExitCalls are runLinearBlocks' own
	// fallible-call resource bookkeeping (D-04-07), carried per-frame so a
	// callee's own resource accounting never leaks into its caller's.
	// Zero-valued and unused by a flat or single-block-arm frame.
	tracked            map[string]bool
	nonlocalExitPolicy bool
	nonlocalExitCalls  int

	// placeTypes and types index this frame's OWN function's declared
	// places and type facts by ID -- present on every frame shape (not
	// just a block-based one), since partitionFrameForCall consults them
	// on every core.OpCall to decide whether the argument place is
	// consumed (D-07-11: a copyable argument is COPIED, the caller's
	// place stays live; a non-copyable one is MOVED).
	placeTypes map[string]string
	types      map[string]core.TypeFact
}

// value is Phase 12's D-12-17 widened place value: tag names the
// alternative a payload-carrying value was constructed as ("" for every
// ordinary, non-payload value), and payload carries either the whole
// scalar value (when tag == "") or the extracted/carried payload content.
// D-12-18: every non-payload write sets tag: "", and String() below
// special-cases tag == "" to return payload bare -- a scalar value's own
// serialized-facing representation is therefore byte-identical to today's
// plain string BY CONSTRUCTION. Never encode tag and payload into one
// string (D-12-20's named in-band-signalling anti-pattern: the literal
// "err" sentinel at this file's OpForeignCall case already shares this
// same string namespace) and never add parallel tags/payloads maps (the
// named parallel-map anti-pattern) -- this single struct is the one
// widened representation every frame.values entry uses.
type value struct {
	tag     string
	payload string
}

// String returns value's own plain-string projection: tag when the value
// is payload-tagged, otherwise payload verbatim (D-12-18).
func (v value) String() string {
	if v.tag != "" {
		return v.tag
	}
	return v.payload
}

// newFlatFrame builds a frame for a flat (non-block) linear body: the
// function's own Operations list, walked once, in order.
func newFlatFrame(function core.Function, values map[string]value) frame {
	ops := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	ids := make([]string, len(function.Linear.Operations))
	for i, operation := range function.Linear.Operations {
		ops[operation.ID] = operation
		ids[i] = operation.ID
	}
	return frame{
		function: function, values: values, live: map[string]bool{}, operations: ops, ids: ids,
		placeTypes: placeTypeIndex(function), types: typeFactIndex(function),
	}
}

// newBlockFrame builds a frame for a block-based linear body (Blocks/Edges
// populated), starting at startBlockID, with runLinearBlocks' own
// fallible-call resource bookkeeping seeded from the function's own
// declared operations.
func newBlockFrame(function core.Function, values map[string]value, startBlockID string) frame {
	ops := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		ops[operation.ID] = operation
	}
	blocks := make(map[string]core.Block, len(function.Linear.Blocks))
	for _, block := range function.Linear.Blocks {
		blocks[block.ID] = block
	}
	edges := make(map[string]core.Edge, len(function.Linear.Edges))
	for _, edge := range function.Linear.Edges {
		edges[edge.ID] = edge
	}
	tracked := make(map[string]bool)
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpRelease && operation.ReleasesOperationID != "" {
			tracked[operation.ReleasesOperationID] = true
		}
	}
	nonlocalExitPolicy := function.ForeignContract != nil && function.ForeignContract.NonlocalExit != "" && function.ForeignContract.NonlocalExit != "forbidden"
	return frame{
		function: function, values: values, live: map[string]bool{},
		operations: ops, blocks: blocks, edges: edges, currentBlockID: startBlockID,
		tracked: tracked, nonlocalExitPolicy: nonlocalExitPolicy,
		placeTypes: placeTypeIndex(function), types: typeFactIndex(function),
	}
}

// placeTypeIndex and typeFactIndex build the ID-keyed lookups every frame
// shape carries: a place's own declared TypeID, and a type ID's own
// TypeFact (abilities included) -- both read off the function's own
// Linear body, never re-derived.
func placeTypeIndex(function core.Function) map[string]string {
	placeTypes := make(map[string]string, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		placeTypes[place.ID] = place.TypeID
	}
	return placeTypes
}

func typeFactIndex(function core.Function) map[string]core.TypeFact {
	types := make(map[string]core.TypeFact, len(function.Linear.Types))
	for _, fact := range function.Linear.Types {
		types[fact.ID] = fact
	}
	return types
}

// newArmFrame builds a frame for a single selected match arm's own block
// (runBranchArm's shape): the arm's block lives in the same
// function.Linear.Blocks list a full block-based body uses, so this reuses
// newBlockFrame and marks the frame singleBlockOnly -- no successor is
// ever followed past the arm's own block.
func newArmFrame(function core.Function, values map[string]value, blockID string) frame {
	f := newBlockFrame(function, values, blockID)
	f.singleBlockOnly = true
	return f
}

// pushResult is partitionFrameForCall's outcome: either a genuine callee
// frame to push onto the stack (Frame non-nil), or an immediate
// resolution. A bare match arm (core.MatchArm.BlockID == "") needs no
// execution at all -- only its own declared literal value and event,
// exactly mirroring Run's own top-level immediate-arm shortcut -- so a
// callee whose selected arm is bare never grows the stack at all (D-10-26:
// this is not a special case of the frame model; it is simply the zero-
// operations-to-run case, same as the top-level entry point already
// handled before this plan).
type pushResult struct {
	frame          *frame
	immediateValue string
	immediateEvent Event
}

// partitionFrameForCall is the ONE frame-partition helper D-10-39 permits
// across all three of interp's execution paths (runBranchArm,
// runLinearBlocks, runLinear): given the CALLER frame and the OpCall
// operation, it resolves the callee, consumes the argument place from the
// caller's own values map -- deleting it only when the argument is NOT
// copyable (D-07-11: a copyable argument is COPIED, so the caller's place
// stays live, mirroring resolveCallBinding's own consume predicate at the
// check layer exactly) -- and dispatches the callee's own body shape
// (flat, block-based, or match-arm-selected) exactly as Run's own
// top-level entry point does. This realizes the OWN-05b ownership-transfer
// fact as observable execution behavior (D-10-26): a genuinely moved-from
// place is gone from the caller for the call's duration, never merely a
// Mode string consulted after the fact.
func partitionFrameForCall(program core.Program, caller *frame, operation core.LinearOperation) (pushResult, error) {
	callee, ok := findFunctionByID(program, operation.CalleeID)
	if !ok {
		return pushResult{}, fmt.Errorf("call to unresolved callee %q", operation.CalleeID)
	}
	if !callee.HasClosedBody() {
		return pushResult{}, fmt.Errorf("callee %q has invalid body union", callee.ID)
	}

	argument := caller.values[operation.SourceID]
	if !moveAsCopyForTest && !argumentIsCopyable(caller, operation) {
		delete(caller.values, operation.SourceID)
	}
	seeded := map[string]value{callee.Parameter.ID: argument}

	if callee.Match != nil {
		argumentText := argument.String()
		for _, arm := range callee.Match.Arms {
			if arm.Pattern != argumentText {
				continue
			}
			if arm.BlockID == "" {
				return pushResult{
					immediateValue: arm.Value,
					immediateEvent: Event{
						Schema: Schema, ID: arm.ID + ":event:returned", Kind: "function.returned",
						FunctionID: callee.ID, Input: argumentText, Output: arm.Value,
					},
				}, nil
			}
			f := newArmFrame(callee, seeded, arm.BlockID)
			f.returnTarget = operation.TargetID
			f.hasCaller = true
			return pushResult{frame: &f}, nil
		}
		return pushResult{}, fmt.Errorf("checked match %q has no arm for %q", callee.Match.ID, argumentText)
	}

	var f frame
	if len(callee.Linear.Blocks) > 0 {
		f = newBlockFrame(callee, seeded, callee.ID+":block:entry")
	} else {
		f = newFlatFrame(callee, seeded)
	}
	f.returnTarget = operation.TargetID
	f.hasCaller = true
	return pushResult{frame: &f}, nil
}

// argumentIsCopyable reports whether operation's own SourceID place carries
// core.AbilityCopy in the CALLER's declared type facts -- the exact
// predicate resolveCallBinding (check.go) applies when deciding whether a
// call argument is copied (place stays live) or moved (place consumed). An
// argument whose type fact cannot be resolved at all is treated as
// non-copyable (fail-closed), matching check's own "empty ability list is
// never the permitting default" rule.
func argumentIsCopyable(caller *frame, operation core.LinearOperation) bool {
	typeID, known := caller.placeTypes[operation.SourceID]
	if !known {
		return false
	}
	fact, known := caller.types[typeID]
	if !known {
		return false
	}
	for _, ability := range fact.Abilities {
		if ability == core.AbilityCopy {
			return true
		}
	}
	return false
}

// findFunctionByID resolves an OpCall's CalleeID against the whole
// program's declared functions -- unlike findFunction, which resolves by
// source Name for Run's own top-level entry point.
func findFunctionByID(program core.Program, id string) (core.Function, bool) {
	for _, function := range program.Functions {
		if function.ID == id {
			return function, true
		}
	}
	return core.Function{}, false
}

// currentOperationIDs reports the operation ID sequence the frame is
// currently walking: the flat body's own ordered list, or the active
// block's own OperationIDs for a block-based body.
func currentOperationIDs(f *frame) []string {
	if f.blocks == nil {
		return f.ids
	}
	return f.blocks[f.currentBlockID].OperationIDs
}

// terminalOutcome builds the Outcome and Event a core.OpReturn/OpFail/
// OpDefect terminator produces, attributed to the frame's OWN function ID
// (D-10-32) -- never a depth value, so a callee's events are distinguished
// from its caller's by identity alone, invariant under inlining.
func terminalOutcome(function core.Function, operation core.LinearOperation, value string) (Outcome, Event) {
	switch operation.Kind {
	case core.OpFail:
		return Outcome{Kind: "typed_failure", Value: value}, Event{
			Schema: execution.Schema1, ID: operation.ID + ":event:failed", Kind: "function.failed",
			FunctionID: function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID,
		}
	case core.OpDefect:
		return Outcome{Kind: execution.OutcomeDefect, Value: ""}, Event{
			Schema: execution.Schema1, ID: operation.ID + ":event:defected", Kind: "function.defected",
			FunctionID: function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID, Output: operation.Reason,
		}
	default: // core.OpReturn
		return Outcome{Kind: "returned", Value: value}, Event{
			Schema: execution.Schema1, ID: operation.ID + ":event:returned", Kind: "function.returned",
			FunctionID: function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID,
		}
	}
}

// runFrameStack drives interp's explicit heap frame stack (D-10-21) with a
// for {} loop, pushing a new frame on core.OpCall and popping on a
// core.OpReturn terminator, binding the popped frame's returned value into
// its caller's own TargetID place. base is the outermost (program-entry)
// frame; every callee frame pushed above it shares this ONE driver, never
// a special case for the entry point (D-10-26). Every ordered output this
// loop produces -- the shared events slice, each frame's own liveOrder --
// comes from an ordered slice, never a Go map range, so repeated runs of
// the same program are byte-identical.
//
// core.OpFail and core.OpDefect always end the WHOLE Execution, even when
// raised from inside a callee frame: Lang has no exception handling this
// phase, so a callee's typed failure or defect is not caught by its
// caller. This is a documented scope limit (a functionality gap, not an
// architectural one) -- no fixture yet raises either from inside a call.
func runFrameStack(program core.Program, base frame) (Execution, error) {
	stack := []frame{base}
	events := make([]Event, 0, len(base.operations))

	for {
		top := &stack[len(stack)-1]
		ids := currentOperationIDs(top)
		if top.idx >= len(ids) {
			if top.blocks == nil {
				return Execution{}, fmt.Errorf("linear body %q has no return operation", top.function.Linear.ID)
			}
			if top.singleBlockOnly {
				return Execution{}, fmt.Errorf("branch arm referencing block %q has no return operation", top.currentBlockID)
			}
			block := top.blocks[top.currentBlockID]
			if len(block.Successors) != 1 {
				return Execution{}, fmt.Errorf("block %q has no terminator and an ambiguous successor set", block.ID)
			}
			top.currentBlockID = block.Successors[0]
			top.idx = 0
			continue
		}

		operationID := ids[top.idx]
		operation, known := top.operations[operationID]
		if !known {
			return Execution{}, fmt.Errorf("block or body references unknown operation %q", operationID)
		}
		sourceValue, initialized := top.values[operation.SourceID]
		if !initialized {
			return Execution{}, fmt.Errorf("operation %q reads uninitialized place %q", operation.ID, operation.SourceID)
		}

		switch operation.Kind {
		case core.OpCopy:
			top.values[operation.TargetID] = sourceValue
			events = append(events, ownedEvent(top.function, operation, "value.copied"))
			top.idx++
		case core.OpMove:
			delete(top.values, operation.SourceID)
			top.values[operation.TargetID] = sourceValue
			events = append(events, ownedEvent(top.function, operation, "value.transferred"))
			top.idx++
		case core.OpBorrowShared:
			top.values[operation.TargetID] = sourceValue
			events = append(events, ownedEvent(top.function, operation, "value.borrowed"))
			top.idx++
		case core.OpBorrowExclusive:
			top.values[operation.TargetID] = sourceValue
			events = append(events, ownedEvent(top.function, operation, "value.borrowed_exclusive"))
			top.idx++
		case core.OpConstructPayload:
			// D-12-05/D-12-14: builds a NEW tagged value from the source
			// payload's own content, wrapped with the alternative name this
			// operation's own PayloadType uniquely resolves to (D-12-25:
			// alternativeNameForPayloadType reads the SAME checked
			// core.DataType every other consumer reads, never a second
			// derivation). Construction consumes its source (the payload
			// the arm's own binder names), mirroring OpMove's delete.
			delete(top.values, operation.SourceID)
			altName := alternativeNameForPayloadType(program, top.function.Parameter.Type, operation.PayloadType)
			top.values[operation.TargetID] = value{tag: altName, payload: sourceValue.payload}
			events = append(events, ownedEvent(top.function, operation, "value.payload_constructed"))
			top.idx++
		case core.OpDestructurePayload:
			// D-12-05/D-12-14: extracts the source's own payload content
			// into PayloadTargetID (never the ordinary TargetID, which this
			// kind leaves unused), mirroring OpMove's delete of the
			// scrutinee alias it moves the payload out of.
			delete(top.values, operation.SourceID)
			top.values[operation.PayloadTargetID] = value{tag: "", payload: sourceValue.payload}
			events = append(events, Event{
				Schema: execution.Schema1, ID: operation.ID + ":event", Kind: "value.payload_destructured", FunctionID: top.function.ID,
				SourcePlace: operation.SourceID, TargetPlace: operation.PayloadTargetID, TypeID: operation.TypeID,
			})
			top.idx++
		case core.OpRelease:
			top.live[operation.ReleasesOperationID] = false
			events = append(events, ownedEvent(top.function, operation, "resource.released"))
			top.idx++
		case core.OpForeignCall:
			top.nonlocalExitCalls++
			if top.nonlocalExitPolicy && top.nonlocalExitCalls == 2 {
				events = append(events, Event{
					Schema: execution.Schema1, ID: top.function.ID + ":event:nonlocal_exit", Kind: "foreign.nonlocal_exit", FunctionID: top.function.ID,
				})
				// D-10-33: drain EVERY live frame on the stack, not only
				// top -- this landing pad is one of the two constructs
				// (D-10-31) that genuinely skip more than one frame.
				leakEvents, liveResources := drainStackForAbruptExit(stack)
				events = append(events, leakEvents...)
				events = append(events, Event{
					Schema: execution.Schema1, ID: top.function.ID + ":event:nonlocal_defect", Kind: "function.defected", FunctionID: top.function.ID,
					SourcePlace: top.function.Parameter.ID, TypeID: top.placeTypes[top.function.Parameter.ID], Output: nonlocalExitDefectReason,
				})
				return Execution{Schema: execution.Schema1, Outcome: Outcome{Kind: execution.OutcomeDefect, Value: ""}, Events: events, LiveResources: liveResources}, nil
			}
			top.values[operation.TargetID] = sourceValue
			top.values[operation.ErrTargetID] = value{tag: "", payload: "err"}
			events = append(events, Event{
				Schema: execution.Schema1, ID: operation.ID + ":event", Kind: "foreign.called", FunctionID: top.function.ID,
				SourcePlace: operation.SourceID, TargetPlace: operation.TargetID, TypeID: operation.TypeID,
			})
			if top.tracked[operation.ID] && !top.live[operation.ID] {
				top.live[operation.ID] = true
				top.liveOrder = append(top.liveOrder, operation.ID)
			}
			edge, known := top.edges[operation.OkEdgeID]
			if !known {
				return Execution{}, fmt.Errorf("foreign call %q references unknown ok edge %q", operation.ID, operation.OkEdgeID)
			}
			top.currentBlockID = edge.ToBlockID
			top.idx = 0
		case core.OpCall:
			top.idx++
			result, err := partitionFrameForCall(program, top, operation)
			if err != nil {
				return Execution{}, fmt.Errorf("operation %q: %w", operation.ID, err)
			}
			if result.frame == nil {
				// Immediate arm resolution (a bare match arm): nothing to
				// push, no callee frame ever ran -- bind directly into the
				// caller's own TargetID place.
				events = append(events, result.immediateEvent)
				top.values[operation.TargetID] = value{tag: "", payload: result.immediateValue}
				continue
			}
			if len(stack) >= maxCallDepth() {
				// Pushing this callee frame would exceed the depth ceiling
				// (D-10-21/D-10-23): refuse as a named, comparable Outcome
				// rather than growing the stack further. This is checked
				// AFTER partitionFrameForCall resolves the callee (never
				// before), because a bare match-arm callee needs no frame
				// at all and must never be refused for depth it would not
				// actually consume.
				//
				// D-10-33: the depth refusal is the SECOND construct
				// (D-10-31) that skips more than one frame at once -- drain
				// every live frame currently on the stack, same LIFO
				// flattening the nonlocal landing pad uses.
				leakEvents, liveResources := drainStackForAbruptExit(stack)
				events = append(events, leakEvents...)
				event := Event{
					Schema: execution.Schema1, ID: operation.ID + ":event:call_depth_exceeded", Kind: "function.defected",
					FunctionID: top.function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID, Output: callDepthExceededDefectReason,
				}
				return Execution{
					Schema: execution.Schema1, Outcome: Outcome{Kind: execution.OutcomeDefect, Value: ""},
					Events: append(events, event), LiveResources: liveResources,
				}, nil
			}
			stack = append(stack, *result.frame)
		case core.OpReturn, core.OpFail, core.OpDefect:
			outcome, event := terminalOutcome(top.function, operation, sourceValue.String())
			events = append(events, event)
			if operation.Kind == core.OpReturn && top.hasCaller {
				returnTarget, returnValue := top.returnTarget, outcome.Value
				stack = stack[:len(stack)-1]
				stack[len(stack)-1].values[returnTarget] = value{tag: "", payload: returnValue}
				continue
			}
			liveResources := []string{}
			if top.blocks != nil && !top.singleBlockOnly {
				liveResources = liveResourceList(top.live, top.liveOrder)
			}
			return Execution{Schema: execution.Schema1, Outcome: outcome, Events: events, LiveResources: liveResources}, nil
		default:
			return Execution{}, fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
		}
	}
}

// alternativeNameForPayloadType resolves an OpConstructPayload/
// OpDestructurePayload operation's own PayloadType fact back to the
// declaring data type's alternative name, reading the SAME checked
// core.DataType every other consumer of this fact reads (D-12-25) --
// interp derives no second, independent notion of "which alternative".
// This resolution is unambiguous only when a data type's alternatives
// declare distinct payload types (this tracer's own fixture, D-12-41); a
// data type with two alternatives sharing one payload type is future work,
// not this plan's scope.
func alternativeNameForPayloadType(program core.Program, parameterTypeName, payloadType string) string {
	for _, dataType := range program.DataTypes {
		if dataType.Name != parameterTypeName {
			continue
		}
		for _, detail := range dataType.AlternativeDetails {
			if detail.PayloadType == payloadType {
				return detail.Name
			}
		}
	}
	return ""
}

func ownedEvent(function core.Function, operation core.LinearOperation, kind string) Event {
	return Event{
		Schema: execution.Schema1, ID: operation.ID + ":event", Kind: kind, FunctionID: function.ID,
		SourcePlace: operation.SourceID, TargetPlace: operation.TargetID, TypeID: operation.TypeID,
	}
}

func findFunction(program core.Program, name string) (core.Function, bool) {
	for _, function := range program.Functions {
		if function.Name == name {
			return function, true
		}
	}
	return core.Function{}, false
}
