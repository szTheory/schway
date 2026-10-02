package interp

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/execution"
)

const Schema = execution.Schema0

// schemaForProgram makes the producer boundary explicit: existing
// single-function executions retain the legacy schema selected by their
// producer (/0 for a bare match, /1 for frame-based execution), while a
// program with more than one declared function exposes dynamic activations
// and call causality through /2.
func schemaForProgram(program core.Program, legacySchema string) string {
	if len(program.Functions) > 1 {
		return execution.Schema2
	}
	return legacySchema
}

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
		places := frameLiveResourcePlaces(f)
		for _, place := range places {
			events = append(events, Event{
				Schema: f.eventSchema(), ID: fmt.Sprintf("%s:event:leaked:%d", f.function.ID, leakIndex), Kind: "resource.leaked", FunctionID: f.function.ID,
				Invocation:  f.invocation,
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

const EvidenceScopeModelOnly = "interpreter_model_only"

// ForeignOutcome is an independently supplied semantic result for one
// checked foreign operation. It is consumed by operation ID and never
// performs host IO or represents a physical allocation/free.
type ForeignOutcome struct {
	Kind  string `json:"kind"` // "success" or "failure"
	Type  string `json:"type"`
	Value string `json:"value"`
}

// ModelResult keeps semantic interpretation separate from host observations.
// ActualHostIO and PhysicalCleanup are always false for this runner.
type ModelResult struct {
	Execution       Execution `json:"execution"`
	EvidenceScope   string    `json:"evidence_scope"`
	ActualHostIO    bool      `json:"actual_host_io"`
	PhysicalCleanup bool      `json:"physical_cleanup"`
}

// RunWithForeignOutcomes executes the checked local-owner model using
// deterministic outcomes keyed by each core operation ID. The opaque
// input is treated only as a modeled PathToken; this function never opens it.
func RunWithForeignOutcomes(program core.Program, functionName, input string, outcomes map[string]ForeignOutcome) (ModelResult, error) {
	validated := corevalidate.Validate(program)
	if !validated.Valid {
		return ModelResult{}, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	program = validated.Program()
	if len(program.Functions) == 0 {
		return ModelResult{}, fmt.Errorf("model-only foreign outcomes require a checked entry function")
	}
	function, ok := findFunction(program, functionName)
	if !ok {
		return ModelResult{}, fmt.Errorf("function %q is absent from checked core", functionName)
	}
	phase24ErrorEntry := function.Name == "main" && function.Parameter.Type == "PathToken" && function.ReturnType == "U64" && phase24ErrorProgram(program)
	if !function.HasClosedBody() || function.Linear == nil || function.Match != nil || (len(function.Linear.Blocks) != 0 && !phase24ErrorEntry) || function.Parameter.Type != "PathToken" || function.ReturnType != "U64" {
		return ModelResult{}, fmt.Errorf("model-only foreign outcomes require the checked PathToken-to-U64 local-owner entry")
	}
	initial, err := inputValue(program, function, input)
	if err != nil {
		return ModelResult{}, err
	}
	base := newFlatFrame(function, map[string]value{function.Parameter.ID: initial})
	if phase24ErrorEntry {
		base = newBlockFrame(function, map[string]value{function.Parameter.ID: initial}, function.ID+":block:entry")
	}
	base.modeledOutcomes = make(map[string]ForeignOutcome, len(outcomes))
	base.modeledReusableOutcomes = map[string]bool{}
	base.modeledReusableUsed = map[string]bool{}
	for operationID, outcome := range outcomes {
		base.modeledOutcomes[operationID] = outcome
	}
	for _, candidate := range program.Functions {
		if candidate.Linear == nil {
			continue
		}
		for _, operation := range candidate.Linear.Operations {
			if operation.Foreign != nil && operation.Foreign.Mode == "acquire" {
				base.modeledReusableOutcomes[operation.ID] = true
			}
		}
	}
	result, err := runProgramFrameStack(program, base)
	if err != nil {
		return ModelResult{}, err
	}
	for operationID := range base.modeledReusableOutcomes {
		if base.modeledReusableUsed[operationID] {
			delete(base.modeledOutcomes, operationID)
		}
	}
	if len(base.modeledOutcomes) != 0 {
		return ModelResult{}, fmt.Errorf("modeled foreign outcomes were not consumed for operations %v", sortedOutcomeIDs(base.modeledOutcomes))
	}
	if len(result.LiveResources) != 0 {
		return ModelResult{}, fmt.Errorf("model execution reached a terminal outcome with live local resources %v", result.LiveResources)
	}
	return ModelResult{
		Execution: result, EvidenceScope: EvidenceScopeModelOnly,
		ActualHostIO: false, PhysicalCleanup: false,
	}, nil
}

func phase24ErrorProgram(program core.Program) bool {
	if program.Module != "phase24.error" {
		return false
	}
	names := map[string]bool{}
	for _, function := range program.Functions {
		names[function.Name] = true
	}
	return names["acquire"] && names["probe"] && names["main"]
}

func sortedOutcomeIDs(outcomes map[string]ForeignOutcome) []string {
	ids := make([]string, 0, len(outcomes))
	for id := range outcomes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func modeledOutcomeFor(f *frame, operation core.LinearOperation) (ForeignOutcome, bool) {
	activationKey := ownerActivationID(f, operation.ID)
	if outcome, ok := f.modeledOutcomes[activationKey]; ok {
		delete(f.modeledOutcomes, activationKey)
		return outcome, true
	}
	outcome, ok := f.modeledOutcomes[operation.ID]
	if !ok {
		return ForeignOutcome{}, false
	}
	if operation.Foreign != nil && operation.Foreign.Mode == "acquire" && f.modeledReusableOutcomes[operation.ID] {
		f.modeledReusableUsed[operation.ID] = true
	} else {
		delete(f.modeledOutcomes, operation.ID)
	}
	return outcome, true
}

func modeledForeignResult(program core.Program, frame *frame, operation core.LinearOperation, source value, outcome ForeignOutcome) (value, bool, error) {
	contract := operation.Foreign
	if contract == nil {
		return value{}, false, fmt.Errorf("foreign operation %q has no checked per-operation contract", operation.ID)
	}
	switch outcome.Kind {
	case "success":
		if outcome.Type != contract.ResultType {
			return value{}, false, fmt.Errorf("foreign operation %q modeled result type %q does not match its checked result type %q", operation.ID, outcome.Type, contract.ResultType)
		}
		switch contract.Mode {
		case "acquire":
			parsed, err := strconv.ParseUint(outcome.Value, 10, 8)
			if err != nil || strconv.FormatUint(parsed, 10) != outcome.Value {
				return value{}, false, fmt.Errorf("foreign acquire %q model value must be one canonical byte", operation.ID)
			}
		case "borrow":
			owner, ok := modeledOwnerForBorrow(frame, operation.SourceID)
			if !ok {
				return value{}, false, fmt.Errorf("foreign borrow %q has no live modeled acquire for source place %q", operation.ID, operation.SourceID)
			}
			if contract.ParameterType != owner.Foreign.ResultType {
				return value{}, false, fmt.Errorf("foreign borrow %q parameter type %q does not match acquired owner type %q", operation.ID, contract.ParameterType, owner.Foreign.ResultType)
			}
			if outcome.Value != source.payload {
				return value{}, false, fmt.Errorf("foreign borrow %q modeled byte %q does not match acquired byte %q", operation.ID, outcome.Value, source.payload)
			}
			if !canonicalModeledU64(outcome.Value) {
				return value{}, false, fmt.Errorf("foreign borrow %q model result must be one canonical U64", operation.ID)
			}
		default:
			return value{}, false, fmt.Errorf("foreign operation %q has unsupported modeled mode %q", operation.ID, contract.Mode)
		}
		return value{payload: outcome.Value}, true, nil
	case "failure":
		if contract.Fails == "" || outcome.Type != contract.Fails || !modeledFailureAlternative(program, outcome.Type, outcome.Value) {
			return value{}, false, fmt.Errorf("foreign operation %q modeled failure does not match its checked failure contract", operation.ID)
		}
		if contract.Mode != "acquire" && contract.Mode != "borrow" {
			return value{}, false, fmt.Errorf("foreign operation %q has unsupported modeled failure mode %q", operation.ID, contract.Mode)
		}
		if contract.Mode == "borrow" {
			if _, ok := modeledOwnerForBorrow(frame, operation.SourceID); !ok {
				return value{}, false, fmt.Errorf("foreign borrow %q has no live modeled owner for its failure path", operation.ID)
			}
			if source.payload == "65" || source.payload == "66" {
				return value{}, false, fmt.Errorf("foreign borrow %q cannot report UnsupportedByte for modeled byte %q", operation.ID, source.payload)
			}
		}
		return value{}, false, nil
	default:
		return value{}, false, fmt.Errorf("foreign operation %q has unsupported modeled outcome kind %q", operation.ID, outcome.Kind)
	}
}

func modeledOwnerForBorrow(frame *frame, sourcePlace string) (core.LinearOperation, bool) {
	ownerValue := frame.values[sourcePlace]
	if ownerValue.ownerActivationID != "" {
		if transferred, ok := frame.transferredOwners[ownerValue.ownerActivationID]; ok && transferred.active && transferred.acquisition.ID == ownerValue.ownerOperationID && transferred.placeID == sourcePlace {
			return transferred.acquisition, true
		}
	}
	for _, operation := range frame.operations {
		if operation.Kind == core.OpForeignCall && operation.TargetID == sourcePlace && operation.Foreign != nil && operation.Foreign.Mode == "acquire" && frame.live[operation.ID] {
			return operation, true
		}
	}
	return core.LinearOperation{}, false
}

func modeledFailureAlternative(program core.Program, typeName, alternative string) bool {
	for _, dataType := range program.DataTypes {
		if dataType.Name != typeName {
			continue
		}
		for _, candidate := range dataType.Alternatives {
			if candidate == alternative {
				return true
			}
		}
		return false
	}
	return false
}

func canonicalModeledU64(text string) bool {
	parsed, err := strconv.ParseUint(text, 10, 64)
	return err == nil && strconv.FormatUint(parsed, 10) == text
}

func validateModeledRelease(frame *frame, operation core.LinearOperation, source value) error {
	acquire, ok := frame.operations[operation.ReleasesOperationID]
	if !ok && source.ownerActivationID != "" {
		if transferred, exists := frame.transferredOwners[source.ownerActivationID]; exists && transferred.active && transferred.acquisition.ID == operation.ReleasesOperationID && transferred.placeID == operation.SourceID {
			acquire, ok = transferred.acquisition, true
		}
	}
	if !ok || acquire.Kind != core.OpForeignCall || acquire.Foreign == nil || acquire.Foreign.Mode != "acquire" {
		return fmt.Errorf("release %q does not name a checked modeled acquisition", operation.ID)
	}
	if source.ownerOperationID != "" && source.ownerOperationID != acquire.ID {
		return fmt.Errorf("release %q names a different acquisition than its owner value", operation.ID)
	}
	if operation.Foreign == nil || operation.Foreign.Mode != "consume" || operation.Foreign.Symbol != acquire.Foreign.Release || operation.Foreign.ParameterType != acquire.Foreign.ResultType || operation.Foreign.Allocator != acquire.Foreign.Allocator {
		return fmt.Errorf("release %q does not match its acquisition's checked consume contract", operation.ID)
	}
	return nil
}

func transferReturnedOwner(from, to *frame, returned value, targetPlace string) error {
	if returned.ownerActivationID == "" || returned.ownerOperationID == "" {
		return nil
	}
	var record transferredOwner
	if acquisition, ok := from.operations[returned.ownerOperationID]; ok && from.live[returned.ownerOperationID] && acquisition.Kind == core.OpForeignCall && acquisition.Foreign != nil && acquisition.Foreign.Mode == "acquire" {
		from.live[returned.ownerOperationID] = false
		record = transferredOwner{acquisition: acquisition, placeID: targetPlace, active: true}
	} else if existing, ok := from.transferredOwners[returned.ownerActivationID]; ok && existing.active && existing.acquisition.ID == returned.ownerOperationID {
		existing.active = false
		from.transferredOwners[returned.ownerActivationID] = existing
		record = transferredOwner{acquisition: existing.acquisition, placeID: targetPlace, active: true}
	} else {
		return fmt.Errorf("owning return from %q has no live acquisition for %q", from.function.Name, returned.ownerOperationID)
	}
	if to.transferredOwners == nil {
		to.transferredOwners = map[string]transferredOwner{}
	}
	if _, duplicate := to.transferredOwners[returned.ownerActivationID]; duplicate {
		return fmt.Errorf("owning return duplicated live acquisition %q", returned.ownerActivationID)
	}
	to.transferredOwners[returned.ownerActivationID] = record
	to.transferredOwnerOrder = append(to.transferredOwnerOrder, returned.ownerActivationID)
	return nil
}

func frameLiveResourceList(f *frame) []string {
	result := liveResourceList(f.live, f.liveOrder)
	for _, id := range f.transferredOwnerOrder {
		if owner, ok := f.transferredOwners[id]; ok && owner.active {
			result = append(result, id)
		}
	}
	return result
}

func frameLiveResourcePlaces(f *frame) []string {
	result := liveResourcePlaces(f.operations, f.live, f.liveOrder)
	for _, id := range f.transferredOwnerOrder {
		if owner, ok := f.transferredOwners[id]; ok && owner.active {
			result = append(result, owner.placeID)
		}
	}
	return result
}

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
	if function.Match != nil && function.Linear != nil && len(function.Linear.Blocks) > 0 && len(function.Linear.Blocks[0].OperationIDs) > 0 {
		return runLinearBlocks(program, function, input)
	}
	for _, arm := range function.Match.Arms {
		if arm.Pattern != input {
			continue
		}
		if arm.BlockID == "" {
			schema, invocation, err := entryIdentity(program, function.ID, Schema)
			if err != nil {
				return Execution{}, err
			}
			event := Event{
				Schema: schema, ID: arm.ID + ":event:returned", Kind: "function.returned", Invocation: invocation,
				FunctionID: function.ID, Input: input, Output: arm.Value,
			}
			if function.Parameter.Type != function.ReturnType {
				event.ID = function.ID + ":match:return"
				event.Input, event.Output = "", ""
				event.SourcePlace, event.TypeID = function.Parameter.ID, function.Parameter.Type
			}
			return Execution{
				Schema: schema, Outcome: Outcome{Kind: "returned", Value: arm.Value},
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
	initial, err := inputValue(program, function, input)
	if err != nil {
		return Execution{}, err
	}
	base := newArmFrame(function, map[string]value{function.Parameter.ID: initial}, arm.BlockID)
	return runProgramFrameStack(program, base)
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
	initial, err := inputValue(program, function, input)
	if err != nil {
		return Execution{}, err
	}
	base := newBlockFrame(function, map[string]value{function.Parameter.ID: initial}, function.ID+":block:entry")
	return runProgramFrameStack(program, base)
}

// nonlocalExitDefectReason is the stable Phase 4 interpreter probe result.
// Its wording preserves the historical native-emitter fixture convention;
// current whole-program native emission refuses foreign-call bodies.
const nonlocalExitDefectReason = "foreign nonlocal exit detected at process-root landing pad"

// liveResourcePlaces projects the live-tracking map into the acquisition's
// own TARGET PLACE id, not its operation id -- the identifier convention
// preserved by the historical native-emitter witness for its
// schway_resource_ids array. This is a narrower, DIFFERENT convention than
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
	initial, err := inputValue(program, function, input)
	if err != nil {
		return Execution{}, err
	}
	base := newFlatFrame(function, map[string]value{function.Parameter.ID: initial})
	return runProgramFrameStack(program, base)
}

// inputValue mirrors the production emitter's closed test-input convention:
// data parameters receive the requested alternative and the same canonical
// payload bytes that native main initializes for that alternative.
func inputValue(program core.Program, function core.Function, input string) (value, error) {
	if function.Parameter.Type == "U64" {
		parsed, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			return value{}, fmt.Errorf("function %q: invalid U64 input %q", function.ID, input)
		}
		return value{u64: parsed, isU64: true}, nil
	}
	for _, dataType := range program.DataTypes {
		if dataType.Name != function.Parameter.Type {
			continue
		}
		for _, alternative := range dataType.AlternativeDetails {
			if alternative.Name != input || alternative.PayloadType == "" {
				continue
			}
			switch alternative.PayloadType {
			case "Buffer":
				return value{tag: input, payload: "01020304"}, nil
			case "Byte":
				return value{tag: input, payload: "7"}, nil
			default:
				for _, nested := range program.DataTypes {
					if nested.Name != alternative.PayloadType || len(nested.Alternatives) == 0 {
						continue
					}
					return value{tag: input, payload: nested.Alternatives[0]}, nil
				}
				return value{tag: input, payload: input}, nil
			}
		}
	}
	return value{payload: input}, nil
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

	// schema and invocation are activation-local evidence context. They are
	// intentionally carried by every frame rather than reconstructed from a
	// function ID or stack depth, either of which collides at a shared leaf.
	schema     string
	invocation string
	entryID    string
	segments   []execution.InvocationSegment

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
	returnOkEdge string
	errorTarget  string
	errorEdge    string
	hasCaller    bool

	// transferredOwners carries an acquisition-derived owner obligation after
	// an owning return crosses a call frame. Its key includes the acquisition
	// operation and that frame's call activation, so repeated calls to one
	// helper cannot collapse distinct resources into one map entry.
	transferredOwners     map[string]transferredOwner
	transferredOwnerOrder []string

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
	tracked                 map[string]bool
	nonlocalExitPolicy      bool
	nonlocalExitCalls       int
	modeledOutcomes         map[string]ForeignOutcome
	modeledFailure          *ForeignOutcome
	modeledReusableOutcomes map[string]bool
	modeledReusableUsed     map[string]bool

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
	tag               string
	payload           string
	u64               uint64
	isU64             bool
	boolean           bool
	isBool            bool
	ownerOperationID  string
	ownerActivationID string
}

// transferredOwner keeps the acquisition contract and the owner's current
// receiving place together after OpReturn transfers it across a frame.
type transferredOwner struct {
	acquisition core.LinearOperation
	placeID     string
	active      bool
}

func ownerActivationID(f *frame, operationID string) string {
	activation := f.invocation
	if activation == "" {
		activation = f.entryID
		for _, segment := range f.segments {
			activation += "/" + segment.OpCallID + ":" + strconv.Itoa(segment.Ordinal)
		}
	}
	return operationID + "@" + activation
}

// disableEmptyTagSerializationSeamForTest is Phase 12 plan 04 Task 2's own
// D-12-38-shaped mutation-kill seam (mirroring cgen's
// SetOpCallGroupedArmForTest / check's callReturnTypeDerivationSeam shape):
// when true, value.String() below skips D-12-18's tag=="" special case
// entirely, so a SCALAR value's own serialized-facing projection stops being
// byte-identical to today's plain string. This exists to prove
// TestPayloadCorpusCharacterizationReplay is genuinely load-bearing rather
// than a self-comparison that would pass regardless of what this file does:
// disabling the seam must turn the replay red. Exposed to the external
// session_test package via interp/export_test.go's
// SetDisableEmptyTagSerializationForTest; never called from any production
// code path. false (the production default) means "apply D-12-18's
// special case for real".
var disableEmptyTagSerializationSeamForTest = false

// SetDisableEmptyTagSerializationForTest installs/restores
// disableEmptyTagSerializationSeamForTest. Callers MUST defer the returned
// restore func immediately. Following corevalidate.SetDisableCyclePeerForTest's
// own D-07-42 precedent exactly: this is production-visible (so the
// cross-package session_test package can reach it -- a _test.go-only export
// is invisible outside interp's own test binary), but a documented,
// clearly-named, test-only no-op unless a test explicitly calls it. Never
// called from any production code path in this repository.
func SetDisableEmptyTagSerializationForTest(disabled bool) (restore func()) {
	previous := disableEmptyTagSerializationSeamForTest
	disableEmptyTagSerializationSeamForTest = disabled
	return func() { disableEmptyTagSerializationSeamForTest = previous }
}

// String returns value's own plain-string projection: tag when the value
// is payload-tagged, otherwise payload verbatim (D-12-18).
func (v value) String() string {
	if disableEmptyTagSerializationSeamForTest {
		// Seam engaged: every value's projection carries both fields, so a
		// scalar's serialized bytes move even though nothing else about the
		// program changed -- exactly the drift
		// TestPayloadCorpusCharacterizationReplay must catch.
		return "tag=" + v.tag + "|payload=" + v.payload
	}
	if v.tag != "" {
		return v.tag
	}
	if v.isU64 {
		return strconv.FormatUint(v.u64, 10)
	}
	return v.payload
}

// terminalString preserves a returned payload alongside its selected
// alternative. Scrutinee dispatch continues to use String(), which remains
// the tag-only representation for payload-bearing ADTs.
func (v value) terminalString() string {
	if v.tag != "" && v.payload != "" {
		return v.tag + ":" + v.payload
	}
	return v.String()
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
		transferredOwners: map[string]transferredOwner{},
		placeTypes:        placeTypeIndex(function), types: typeFactIndex(function),
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
		transferredOwners: map[string]transferredOwner{},
		placeTypes:        placeTypeIndex(function), types: typeFactIndex(function),
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

func entryIdentity(program core.Program, entryID, legacySchema string) (string, string, error) {
	schema := schemaForProgram(program, legacySchema)
	if schema != execution.Schema2 {
		return schema, "", nil
	}
	invocation, err := execution.FormatInvocation(entryID, nil)
	if err != nil {
		return "", "", fmt.Errorf("entry invocation: %w", err)
	}
	return schema, invocation, nil
}

func runProgramFrameStack(program core.Program, base frame) (Execution, error) {
	schema, invocation, err := entryIdentity(program, base.function.ID, execution.Schema1)
	if err != nil {
		return Execution{}, err
	}
	base.schema, base.invocation, base.entryID = schema, invocation, base.function.ID
	return runFrameStack(program, base)
}

func (f *frame) eventSchema() string {
	if f.schema == "" {
		return execution.Schema1
	}
	return f.schema
}

func (f *frame) childInvocation(operation core.LinearOperation) string {
	if f.eventSchema() != execution.Schema2 {
		return ""
	}
	segments := append(append([]execution.InvocationSegment{}, f.segments...), execution.InvocationSegment{OpCallID: operation.ID, Ordinal: 0})
	invocation, err := execution.FormatInvocation(f.entryID, segments)
	if err != nil {
		panic(fmt.Sprintf("checked core has invalid invocation component: %v", err))
	}
	return invocation
}

func (f *frame) configureChild(child *frame, operation core.LinearOperation) {
	child.schema = f.eventSchema()
	child.entryID = f.entryID
	child.segments = append(append([]execution.InvocationSegment{}, f.segments...), execution.InvocationSegment{OpCallID: operation.ID, Ordinal: 0})
	child.invocation = f.childInvocation(operation)
	child.returnOkEdge = operation.OkEdgeID
	child.errorTarget = operation.ErrTargetID
	child.errorEdge = operation.ErrEdgeID
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
	calleeID       string
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
		scrutineeID := callee.Match.ScrutineeID
		if scrutineeID == "" {
			scrutineeID = callee.Parameter.ID
		}
		if scrutineeID != callee.Parameter.ID {
			f := newBlockFrame(callee, seeded, callee.ID+":block:entry")
			f.returnTarget = operation.TargetID
			f.hasCaller = true
			caller.configureChild(&f, operation)
			return pushResult{frame: &f, calleeID: callee.ID}, nil
		}
		argumentText := argument.String()
		for _, arm := range callee.Match.Arms {
			if arm.Pattern != argumentText {
				continue
			}
			if arm.BlockID == "" {
				event := Event{
					Schema: caller.eventSchema(), ID: arm.ID + ":event:returned", Kind: "function.returned",
					Invocation: caller.childInvocation(operation),
					FunctionID: callee.ID, Input: argumentText, Output: arm.Value,
				}
				if callee.Parameter.Type != callee.ReturnType {
					event.ID = callee.ID + ":match:return"
					event.Input, event.Output = "", ""
					event.SourcePlace, event.TypeID = callee.Parameter.ID, callee.Parameter.Type
				}
				return pushResult{
					immediateValue: arm.Value,
					immediateEvent: event,
					calleeID:       callee.ID,
				}, nil
			}
			f := newArmFrame(callee, seeded, arm.BlockID)
			f.returnTarget = operation.TargetID
			f.hasCaller = true
			caller.configureChild(&f, operation)
			return pushResult{frame: &f, calleeID: callee.ID}, nil
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
	caller.configureChild(&f, operation)
	return pushResult{frame: &f, calleeID: callee.ID}, nil
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
func terminalOutcome(f *frame, operation core.LinearOperation, value string) (Outcome, Event) {
	switch operation.Kind {
	case core.OpFail:
		failureType := operation.TypeID
		if f.modeledFailure != nil {
			value = f.modeledFailure.Value
			failureType = f.modeledFailure.Type
		}
		return Outcome{Kind: "typed_failure", Value: value}, Event{
			Schema: f.eventSchema(), ID: operation.ID + ":event:failed", Kind: "function.failed",
			FunctionID: f.function.ID, Invocation: f.invocation, SourcePlace: operation.SourceID, TypeID: failureType,
		}
	case core.OpDefect:
		return Outcome{Kind: execution.OutcomeDefect, Value: ""}, Event{
			Schema: f.eventSchema(), ID: operation.ID + ":event:defected", Kind: "function.defected",
			FunctionID: f.function.ID, Invocation: f.invocation, SourcePlace: operation.SourceID, TypeID: operation.TypeID, Output: operation.Reason,
		}
	default: // core.OpReturn
		return Outcome{Kind: "returned", Value: value}, Event{
			Schema: f.eventSchema(), ID: operation.ID + ":event:returned", Kind: "function.returned",
			FunctionID: f.function.ID, Invocation: f.invocation, SourcePlace: operation.SourceID, TypeID: operation.TypeID,
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
			if top.function.Match != nil && (len(block.Successors) > 1 || block.ID == top.function.ID+":block:entry") {
				scrutineeID := top.function.Match.ScrutineeID
				if scrutineeID == "" {
					for _, place := range top.function.Linear.Places {
						if place.Name == top.function.Match.Scrutinee {
							scrutineeID = place.ID
							break
						}
					}
				}
				scrutinee, initialized := top.values[scrutineeID]
				if !initialized || scrutineeID == "" {
					return Execution{}, fmt.Errorf("match %q scrutinee %q is not initialized at dispatch", top.function.Match.ID, top.function.Match.Scrutinee)
				}
				selected := ""
				for _, edge := range top.edges {
					if edge.FromBlockID == block.ID && edge.Pattern == scrutinee.String() {
						selected = edge.ToBlockID
						break
					}
				}
				if selected == "" {
					return Execution{}, fmt.Errorf("match %q has no edge for %q", top.function.Match.ID, scrutinee.String())
				}
				for _, arm := range top.function.Match.Arms {
					if arm.BlockID == selected && arm.ValuePlaceID != "" {
						top.values[arm.ValuePlaceID] = value{payload: arm.Value}
					}
				}
				top.currentBlockID = selected
				top.idx = 0
				continue
			}
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
		var sourceValue value
		modeledFailureContinuation := top.modeledFailure != nil && (operation.Kind == core.OpForeignCall || operation.Kind == core.OpRelease || operation.Kind == core.OpReturn)
		if top.modeledFailure != nil && operation.Kind != core.OpRelease && operation.Kind != core.OpReturn && operation.Kind != core.OpForeignCall && operation.Kind != core.OpFail {
			// A modeled foreign failure has already selected the checked error
			// edge. Do not evaluate successful-path calls or value operations
			// while walking toward cleanup and the enclosing typed failure.
			top.idx++
			continue
		}
		if operation.Kind == core.OpRelease && top.modeledFailure != nil {
			// A failed borrowed use still leaves its transferred owner available
			// for the unconditional cleanup operation. A failed acquisition has
			// no initialized value at this place and remains a no-op release.
			sourceValue = top.values[operation.SourceID]
		} else if operation.Kind != core.OpConst && !modeledFailureContinuation {
			var initialized bool
			sourceValue, initialized = top.values[operation.SourceID]
			if !initialized {
				return Execution{}, fmt.Errorf("operation %q reads uninitialized place %q", operation.ID, operation.SourceID)
			}
		}

		switch operation.Kind {
		case core.OpConst:
			parsed, err := strconv.ParseUint(operation.ConstU64, 10, 64)
			if err != nil {
				return Execution{}, fmt.Errorf("operation %q has invalid canonical U64 constant", operation.ID)
			}
			top.values[operation.TargetID] = value{u64: parsed, isU64: true}
			top.idx++
		case core.OpAddChecked:
			right, initialized := top.values[operation.RightID]
			if !initialized || !sourceValue.isU64 || !right.isU64 {
				return Execution{}, fmt.Errorf("operation %q has invalid checked-add operands", operation.ID)
			}
			if ^uint64(0)-sourceValue.u64 < right.u64 {
				return Execution{Outcome: Outcome{Kind: execution.OutcomeDefect}, Events: events, LiveResources: liveResourceList(top.live, top.liveOrder)}, nil
			}
			top.values[operation.TargetID] = value{u64: sourceValue.u64 + right.u64, isU64: true}
			top.idx++
		case core.OpLessU64:
			right, initialized := top.values[operation.RightID]
			if !initialized || !sourceValue.isU64 || !right.isU64 {
				return Execution{}, fmt.Errorf("operation %q has invalid less-than operands", operation.ID)
			}
			top.values[operation.TargetID] = value{boolean: sourceValue.u64 < right.u64, isBool: true}
			top.idx++
		case core.OpScalarStore:
			top.values[operation.StoreTargetID] = sourceValue
			top.idx++
		case core.OpBranch:
			if !sourceValue.isBool {
				return Execution{}, fmt.Errorf("operation %q branch condition is not Bool", operation.ID)
			}
			edgeID := operation.FalseEdgeID
			if sourceValue.boolean {
				edgeID = operation.TrueEdgeID
			}
			edge, known := top.edges[edgeID]
			if !known {
				return Execution{}, fmt.Errorf("operation %q references unknown branch edge", operation.ID)
			}
			top.currentBlockID, top.idx = edge.ToBlockID, 0
			continue
		case core.OpCopy:
			top.values[operation.TargetID] = sourceValue
			events = append(events, ownedEvent(top, operation, "value.copied"))
			top.idx++
		case core.OpMove:
			delete(top.values, operation.SourceID)
			top.values[operation.TargetID] = sourceValue
			events = append(events, ownedEvent(top, operation, "value.transferred"))
			top.idx++
		case core.OpBorrowShared:
			top.values[operation.TargetID] = sourceValue
			events = append(events, ownedEvent(top, operation, "value.borrowed"))
			top.idx++
		case core.OpBorrowExclusive:
			top.values[operation.TargetID] = sourceValue
			events = append(events, ownedEvent(top, operation, "value.borrowed_exclusive"))
			top.idx++
		case core.OpConstructPayload:
			// D-12-05/D-12-14: builds a NEW tagged value from the source
			// payload's own content, wrapped with the alternative name this
			// operation's own PayloadType uniquely resolves to via the ONE
			// shared, ambiguity-detecting derivation core.
			// AlternativeNameForPayloadType (IN-01/CR-01, D-12-25) --
			// interp derives no second, independent notion of "which
			// alternative". Construction consumes its source (the payload
			// the arm's own binder names), mirroring OpMove's delete.
			delete(top.values, operation.SourceID)
			var dataType core.DataType
			var dataTypeKnown bool
			for _, candidate := range program.DataTypes {
				if candidate.Name == top.function.Parameter.Type {
					dataType = candidate
					dataTypeKnown = true
					break
				}
			}
			if !dataTypeKnown {
				return Execution{}, fmt.Errorf("operation %q: no data type named %q", operation.ID, top.function.Parameter.Type)
			}
			altName, altErr := core.AlternativeNameForPayloadType(dataType, operation.PayloadType)
			if altErr != nil {
				return Execution{}, fmt.Errorf("operation %q: %w", operation.ID, altErr)
			}
			top.values[operation.TargetID] = value{tag: altName, payload: sourceValue.payload}
			events = append(events, ownedEvent(top, operation, "value.payload_constructed"))
			top.idx++
		case core.OpDestructurePayload:
			// D-12-05/D-12-14: extracts the source's own payload content
			// into PayloadTargetID (never the ordinary TargetID, which this
			// kind leaves unused), mirroring OpMove's delete of the
			// scrutinee alias it moves the payload out of.
			delete(top.values, operation.SourceID)
			top.values[operation.PayloadTargetID] = value{tag: "", payload: sourceValue.payload}
			events = append(events, Event{
				Schema: top.eventSchema(), ID: operation.ID + ":event", Kind: "value.payload_destructured", FunctionID: top.function.ID, Invocation: top.invocation,
				SourcePlace: operation.SourceID, TargetPlace: operation.PayloadTargetID, TypeID: operation.TypeID,
			})
			top.idx++
		case core.OpRelease:
			transferred, hasTransferred := top.transferredOwners[sourceValue.ownerActivationID]
			transferredLive := hasTransferred && transferred.active && transferred.acquisition.ID == operation.ReleasesOperationID && transferred.placeID == operation.SourceID
			if top.modeledOutcomes != nil {
				if !top.live[operation.ReleasesOperationID] && !transferredLive {
					top.idx++
					continue
				}
				if err := validateModeledRelease(top, operation, sourceValue); err != nil {
					return Execution{}, err
				}
			}
			if transferredLive {
				transferred.active = false
				top.transferredOwners[sourceValue.ownerActivationID] = transferred
			}
			top.live[operation.ReleasesOperationID] = false
			events = append(events, ownedEvent(top, operation, "resource.released"))
			top.idx++
		case core.OpForeignCall:
			if top.modeledOutcomes != nil {
				if top.modeledFailure != nil {
					top.idx++
					continue
				}
				outcome, supplied := modeledOutcomeFor(top, operation)
				if !supplied {
					return Execution{}, fmt.Errorf("modeled outcome for foreign operation %q is missing", operation.ID)
				}
				modeledValue, succeeded, err := modeledForeignResult(program, top, operation, sourceValue, outcome)
				if err != nil {
					return Execution{}, err
				}
				event := Event{
					Schema: top.eventSchema(), ID: operation.ID + ":event", Kind: "foreign.called", FunctionID: top.function.ID, Invocation: top.invocation,
					SourcePlace: operation.SourceID, TargetPlace: operation.TargetID, TypeID: operation.TypeID,
				}
				if !succeeded {
					top.modeledFailure = &outcome
					if operation.ErrTargetID != "" {
						top.values[operation.ErrTargetID] = value{tag: outcome.Type, payload: outcome.Value}
					}
					event.Kind = "foreign.failed"
					event.TypeID = outcome.Type
				} else {
					if operation.Foreign != nil && operation.Foreign.Mode == "acquire" {
						modeledValue.ownerOperationID = operation.ID
						modeledValue.ownerActivationID = ownerActivationID(top, operation.ID)
					}
					top.values[operation.TargetID] = modeledValue
					if operation.Foreign != nil && operation.Foreign.Mode == "acquire" {
						top.live[operation.ID] = true
						top.liveOrder = append(top.liveOrder, operation.ID)
					}
				}
				events = append(events, event)
				if top.blocks != nil && operation.OkEdgeID != "" {
					edgeID := operation.OkEdgeID
					if !succeeded {
						edgeID = operation.ErrEdgeID
					}
					edge, ok := top.edges[edgeID]
					if !ok {
						return Execution{}, fmt.Errorf("foreign call %q references unknown modeled result edge %q", operation.ID, edgeID)
					}
					top.currentBlockID = edge.ToBlockID
					top.idx = 0
				} else {
					top.idx++
				}
				continue
			}
			top.nonlocalExitCalls++
			if top.nonlocalExitPolicy && top.nonlocalExitCalls == 2 {
				events = append(events, Event{
					Schema: top.eventSchema(), ID: top.function.ID + ":event:nonlocal_exit", Kind: "foreign.nonlocal_exit", FunctionID: top.function.ID, Invocation: top.invocation,
				})
				// D-10-33: drain EVERY live frame on the stack, not only
				// top -- this landing pad is one of the two constructs
				// (D-10-31) that genuinely skip more than one frame.
				leakEvents, liveResources := drainStackForAbruptExit(stack)
				events = append(events, leakEvents...)
				events = append(events, Event{
					Schema: top.eventSchema(), ID: top.function.ID + ":event:nonlocal_defect", Kind: "function.defected", FunctionID: top.function.ID, Invocation: top.invocation,
					SourcePlace: top.function.Parameter.ID, TypeID: top.placeTypes[top.function.Parameter.ID], Output: nonlocalExitDefectReason,
				})
				return Execution{Schema: top.eventSchema(), Outcome: Outcome{Kind: execution.OutcomeDefect, Value: ""}, Events: events, LiveResources: liveResources}, nil
			}
			if operation.Foreign != nil && operation.Foreign.Mode == "acquire" {
				sourceValue.ownerOperationID = operation.ID
				sourceValue.ownerActivationID = ownerActivationID(top, operation.ID)
			}
			top.values[operation.TargetID] = sourceValue
			top.values[operation.ErrTargetID] = value{tag: "", payload: "err"}
			events = append(events, Event{
				Schema: top.eventSchema(), ID: operation.ID + ":event", Kind: "foreign.called", FunctionID: top.function.ID, Invocation: top.invocation,
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
				if top.eventSchema() == execution.Schema2 {
					events = append(events, calledEvent(top, operation, result.calleeID))
				}
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
					Schema: top.eventSchema(), ID: operation.ID + ":event:call_depth_exceeded", Kind: "function.defected",
					FunctionID: top.function.ID, Invocation: top.invocation, SourcePlace: operation.SourceID, TypeID: operation.TypeID, Output: callDepthExceededDefectReason,
				}
				return Execution{
					Schema: top.eventSchema(), Outcome: Outcome{Kind: execution.OutcomeDefect, Value: ""},
					Events: append(events, event), LiveResources: liveResources,
				}, nil
			}
			if top.modeledOutcomes != nil {
				result.frame.modeledOutcomes = top.modeledOutcomes
				result.frame.modeledReusableOutcomes = top.modeledReusableOutcomes
				result.frame.modeledReusableUsed = top.modeledReusableUsed
			}
			if top.eventSchema() == execution.Schema2 {
				events = append(events, calledEvent(top, operation, result.calleeID))
			}
			stack = append(stack, *result.frame)
		case core.OpReturn, core.OpFail, core.OpDefect:
			outcome, event := terminalOutcome(top, operation, sourceValue.terminalString())
			if operation.Kind == core.OpReturn && top.modeledFailure != nil {
				outcome = Outcome{Kind: execution.OutcomeTypedFailure, Value: top.modeledFailure.Value}
				event = Event{
					Schema: top.eventSchema(), ID: operation.ID + ":event:failed", Kind: "function.failed",
					FunctionID: top.function.ID, Invocation: top.invocation, SourcePlace: operation.SourceID, TypeID: top.modeledFailure.Type,
				}
			}
			events = append(events, event)
			if operation.Kind == core.OpFail && top.hasCaller && top.errorTarget != "" && top.errorEdge != "" {
				failure := ForeignOutcome{Kind: "failure", Type: operation.TypeID, Value: sourceValue.terminalString()}
				if top.modeledFailure != nil {
					failure = *top.modeledFailure
				}
				callerIndex := len(stack) - 2
				caller := &stack[callerIndex]
				edge, exists := caller.edges[top.errorEdge]
				if !exists {
					return Execution{}, fmt.Errorf("call failure %q references unknown caller error edge %q", operation.ID, top.errorEdge)
				}
				caller.values[top.errorTarget] = value{tag: failure.Type, payload: failure.Value}
				caller.modeledFailure = &failure
				caller.currentBlockID = edge.ToBlockID
				caller.idx = 0
				stack = stack[:len(stack)-1]
				continue
			}
			if operation.Kind == core.OpReturn && top.hasCaller {
				returnTarget := top.returnTarget
				returnOkEdge := top.returnOkEdge
				returnValue := top.values[operation.SourceID]
				if top.modeledFailure != nil {
					returnValue = value{payload: outcome.Value}
				}
				callerIndex := len(stack) - 2
				if err := transferReturnedOwner(top, &stack[callerIndex], returnValue, returnTarget); err != nil {
					return Execution{}, err
				}
				stack = stack[:len(stack)-1]
				caller := &stack[len(stack)-1]
				caller.values[returnTarget] = returnValue
				if returnOkEdge != "" {
					edge, exists := caller.edges[returnOkEdge]
					if !exists {
						return Execution{}, fmt.Errorf("call return %q references unknown caller ok edge %q", operation.ID, returnOkEdge)
					}
					caller.currentBlockID = edge.ToBlockID
					caller.idx = 0
				}
				continue
			}
			liveResources := frameLiveResourceList(top)
			return Execution{Schema: top.eventSchema(), Outcome: outcome, Events: events, LiveResources: liveResources}, nil
		default:
			return Execution{}, fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
		}
	}
}

func ownedEvent(f *frame, operation core.LinearOperation, kind string) Event {
	return Event{
		Schema: f.eventSchema(), ID: operation.ID + ":event", Kind: kind, FunctionID: f.function.ID, Invocation: f.invocation,
		SourcePlace: operation.SourceID, TargetPlace: operation.TargetID, TypeID: operation.TypeID,
	}
}

func calledEvent(f *frame, operation core.LinearOperation, calleeID string) Event {
	return Event{
		Schema: f.eventSchema(), ID: operation.ID + ":event:called", Kind: "function.called",
		FunctionID: f.function.ID, Invocation: f.invocation, CalleeFunctionID: calleeID,
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
