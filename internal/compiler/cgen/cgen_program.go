package cgen

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/codename-lang/lang/internal/compiler/callgraph"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/execution"
)

// maxInvocationPathTableNodes is the deliberately fixed D-15-16 ceiling on
// static activation occurrences in a native program. It is not a byte-size
// budget and is intentionally independent of pathoracle.MaxPaths.
const maxInvocationPathTableNodes = 4096

var executionOutputLimit = execution.MaxDocumentBytes

// programLiveResourcesForTest is a narrow mutation seam for the surviving
// emitter's schema-2 resource tail. Production derivation has no admitted
// resource-owning shapes yet, so it returns an empty collection; the seam
// proves serialization consumes the derivation rather than a fixed literal.
var programLiveResourcesForTest []string

// deriveProgramLiveResources owns the schema-2 resource result for every
// shape emitProgram admits. Foreign-resource accounting remains outside this
// phase: ordinary linear programs currently derive the semantically correct
// empty collection without introducing a resource ledger.
func deriveProgramLiveResources(program core.Program) []string {
	if programLiveResourcesForTest != nil {
		return append([]string(nil), programLiveResourcesForTest...)
	}
	_ = program
	return []string{}
}

type executionOutputExceededError struct {
	limit, observed int
}

func (e *executionOutputExceededError) Error() string {
	return fmt.Sprintf("cgen.execution_output_exceeded: limit %d observed %d", e.limit, e.observed)
}

func (e *executionOutputExceededError) Code() string  { return "cgen.execution_output_exceeded" }
func (e *executionOutputExceededError) Limit() int    { return e.limit }
func (e *executionOutputExceededError) Observed() int { return e.observed }

// ExecutionOutputExceededError recognizes the stable schema-2 byte-bound
// refusal without exposing its representation for mutation.
func ExecutionOutputExceededError(err error) (*executionOutputExceededError, bool) {
	e, ok := err.(*executionOutputExceededError)
	return e, ok
}

// invocationPreflightBypassForTest is a narrowly-scoped mutation seam. It is
// false in every production build; the test-only accessor proves that omitting
// the guard would otherwise let serialization begin for an over-limit graph.
var invocationPreflightBypassForTest bool

// invocationSerializationReachedForTest records the forbidden path reached by
// the mutation control. It is only read through export_test.go.
var invocationSerializationReachedForTest bool

// invocationPathTableExceededError is the stable, witness-carrying refusal
// returned before C serialization when static invocation unfolding would add
// node 4097. The incoming OpCall and parent occurrence make the first rejected
// occurrence auditable without constructing an unbounded path table.
type invocationPathTableExceededError struct {
	entry, callOperationID, calleeFunctionID string
	parentIndex                              int
	limit, observedAtLeast                   int
}

func (e *invocationPathTableExceededError) Error() string {
	return fmt.Sprintf("cgen.invocation_path_table_exceeded: entry %q limit %d observed_at_least %d first_overflow_parent %d call %q callee %q", e.entry, e.limit, e.observedAtLeast, e.parentIndex, e.callOperationID, e.calleeFunctionID)
}

func (e *invocationPathTableExceededError) Code() string {
	return "cgen.invocation_path_table_exceeded"
}
func (e *invocationPathTableExceededError) Entry() string                 { return e.entry }
func (e *invocationPathTableExceededError) Limit() int                    { return e.limit }
func (e *invocationPathTableExceededError) ObservedAtLeast() int          { return e.observedAtLeast }
func (e *invocationPathTableExceededError) FirstOverflowParentIndex() int { return e.parentIndex }
func (e *invocationPathTableExceededError) FirstOverflowCallOperationID() string {
	return e.callOperationID
}
func (e *invocationPathTableExceededError) FirstOverflowCalleeFunctionID() string {
	return e.calleeFunctionID
}

// InvocationPathTableExceededError recognizes this package's bounded native
// unfolding refusal without exposing its representation for mutation.
func InvocationPathTableExceededError(err error) (*invocationPathTableExceededError, bool) {
	e, ok := err.(*invocationPathTableExceededError)
	return e, ok
}

type invocationPreflightNode struct {
	functionID, incomingCallID string
	parentIndex                int
}

// invocationEventCapacity derives the shared event-array size from activation
// occurrences, not unique function declarations. Every supported whole-program
// body is straight-line and records exactly one event per operation, so this is
// the exact runtime capacity required by the already-bounded unfolding.
func invocationEventCapacity(nodes []invocationPreflightNode, byID map[string]core.Function) (int, error) {
	capacity := 0
	for index, node := range nodes {
		function, ok := byID[node.functionID]
		if !ok {
			return 0, fmt.Errorf("event-capacity node %d names unknown function %q", index, node.functionID)
		}
		if function.Linear == nil {
			if function.Match != nil {
				continue
			}
			return 0, fmt.Errorf("function %q: has no linear body", function.ID)
		}
		capacity += len(function.Linear.Operations)
	}
	return capacity, nil
}

// schema2ExecutionDocumentSize calculates the exact canonical byte count of
// the straight-line document represented by the bounded occurrence table. The
// event order does not affect JSON length; field presence and spellings mirror
// emitProgramFunction and emitEventSupportSchema2.
func schema2ExecutionDocumentSize(entry core.Function, nodes []invocationPreflightNode, paths invocationPathTable, byID map[string]core.Function, liveResources []string) (int, error) {
	events := make([]execution.Event, 0)
	for index, node := range nodes {
		function, ok := byID[node.functionID]
		if !ok || (function.Linear == nil && function.Match == nil) {
			return 0, fmt.Errorf("execution-size node %d names invalid function %q", index, node.functionID)
		}
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			event := execution.Event{
				Schema: execution.Schema2, FunctionID: function.ID, Invocation: paths.invocations[index],
				SourcePlace: operation.SourceID, TargetPlace: operation.TargetID, TypeID: operation.TypeID,
			}
			switch operation.Kind {
			case core.OpCopy:
				event.ID, event.Kind = operation.ID+":event", "value.copied"
			case core.OpMove:
				event.ID, event.Kind = operation.ID+":event", "value.transferred"
			case core.OpBorrowShared:
				event.ID, event.Kind = operation.ID+":event", "value.borrowed"
			case core.OpBorrowExclusive:
				event.ID, event.Kind = operation.ID+":event", "value.borrowed_exclusive"
			case core.OpCall:
				event.ID, event.Kind, event.CalleeFunctionID = operation.ID+":event:called", "function.called", operation.CalleeID
			case core.OpReturn:
				event.ID, event.Kind, event.TargetPlace = operation.ID+":event:returned", "function.returned", ""
			default:
				return 0, fmt.Errorf("operation %q has unsupported schema-2 output kind %q", operation.ID, operation.Kind)
			}
			events = append(events, event)
		}
	}
	outcomeValue := ""
	if entry.Match != nil {
		for _, arm := range entry.Match.Arms {
			if len(arm.Value) > len(outcomeValue) {
				outcomeValue = arm.Value
			}
		}
	} else {
		var err error
		outcomeValue, _, _, err = linearInput(entry)
		if err != nil {
			return 0, err
		}
	}
	document := execution.Execution{
		Schema: execution.Schema2, Outcome: execution.Outcome{Kind: execution.OutcomeReturned, Value: outcomeValue},
		Events: events, LiveResources: liveResources,
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		return 0, fmt.Errorf("size schema-2 execution document: %w", err)
	}
	return len(encoded) + 1, nil // generated stdout terminates the document with '\n'
}

// preflightInvocationPathTable unfolds the validated call DAG in the same
// depth-first, declared-operation order native execution uses. It records only
// bounded occurrence ancestry for the later path-table emitter; crucially it
// refuses before a generated-C builder, emitted table, or C serialization can
// be allocated. Entry is occurrence zero (node one to users).
func preflightInvocationPathTable(program core.Program, entry core.Function) ([]invocationPreflightNode, error) {
	byID := make(map[string]core.Function, len(program.Functions))
	for _, function := range program.Functions {
		byID[function.ID] = function
	}
	nodes := []invocationPreflightNode{{functionID: entry.ID, parentIndex: -1}}
	for index := 0; index < len(nodes); index++ {
		function, ok := byID[nodes[index].functionID]
		if !ok {
			return nil, fmt.Errorf("emitProgram: preflight node names unknown function %q", nodes[index].functionID)
		}
		if function.Linear == nil {
			if function.Match != nil {
				continue
			}
			return nil, fmt.Errorf("function %q: has no linear body", function.ID)
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind != core.OpCall {
				continue
			}
			if len(nodes) == maxInvocationPathTableNodes && !invocationPreflightBypassForTest {
				return nil, &invocationPathTableExceededError{
					entry: entry.ID, limit: maxInvocationPathTableNodes, observedAtLeast: maxInvocationPathTableNodes + 1,
					parentIndex: index, callOperationID: operation.ID, calleeFunctionID: operation.CalleeID,
				}
			}
			nodes = append(nodes, invocationPreflightNode{functionID: operation.CalleeID, incomingCallID: operation.ID, parentIndex: index})
		}
	}
	return nodes, nil
}

// invocationPathTable turns the bounded preorder produced by preflight into
// C literals. A child table is deliberately indexed by the caller occurrence
// rather than by function: the same generated caller can run beneath two
// parents and must select two different child activations (D-15-15).
type invocationPathTable struct {
	invocations []string
	children    map[string][]int
}

func buildInvocationPathTable(entryID string, nodes []invocationPreflightNode) (invocationPathTable, error) {
	table := invocationPathTable{invocations: make([]string, len(nodes)), children: map[string][]int{}}
	segments := make([][]execution.InvocationSegment, len(nodes))
	for index, node := range nodes {
		if node.parentIndex < 0 {
			segments[index] = nil
		} else {
			if node.parentIndex >= index {
				return invocationPathTable{}, fmt.Errorf("invocation node %d has non-preorder parent %d", index, node.parentIndex)
			}
			segments[index] = append(append([]execution.InvocationSegment{}, segments[node.parentIndex]...), execution.InvocationSegment{OpCallID: node.incomingCallID, Ordinal: 0})
			if table.children[node.incomingCallID] == nil {
				table.children[node.incomingCallID] = make([]int, len(nodes))
				for childIndex := range table.children[node.incomingCallID] {
					table.children[node.incomingCallID][childIndex] = -1
				}
			}
			table.children[node.incomingCallID][node.parentIndex] = index
		}
		invocation, err := execution.FormatInvocation(entryID, segments[index])
		if err != nil {
			return invocationPathTable{}, fmt.Errorf("invocation node %d: %w", index, err)
		}
		table.invocations[index] = invocation
	}
	return table, nil
}

// ensureInvocationChildTables keeps every emitted definition compilable,
// including deliberately emitted but unreachable functions. Their rows have
// no reachable child occurrence, yet their static call expression still
// needs a named literal table; it is never selected by a valid invocation.
func ensureInvocationChildTables(table *invocationPathTable, functions []core.Function) {
	for _, function := range functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind != core.OpCall || table.children[operation.ID] != nil {
				continue
			}
			indices := make([]int, len(table.invocations))
			for index := range indices {
				indices[index] = -1
			}
			table.children[operation.ID] = indices
		}
	}
}

func invocationChildTableNames(table invocationPathTable, names *cNames) map[string]string {
	callIDs := make([]string, 0, len(table.children))
	for callID := range table.children {
		callIDs = append(callIDs, callID)
	}
	sort.Strings(callIDs)
	tableNames := make(map[string]string, len(callIDs))
	for ordinal, callID := range callIDs {
		tableNames[callID] = names.allocate("lang_child_index_"+cName(callID), "invocation_child", ordinal)
	}
	return tableNames
}

func emitInvocationPathTable(out *strings.Builder, table invocationPathTable, tableNames map[string]string) {
	out.WriteString("static const char *lang_invocations[] = {\n")
	for _, invocation := range table.invocations {
		fmt.Fprintf(out, "  %s,\n", strconv.Quote(invocation))
	}
	out.WriteString("};\n\n")

	callIDs := make([]string, 0, len(table.children))
	for callID := range table.children {
		callIDs = append(callIDs, callID)
	}
	sort.Strings(callIDs)
	for _, callID := range callIDs {
		tableName := tableNames[callID]
		fmt.Fprintf(out, "static const unsigned int %s[] = {\n", tableName)
		for _, childIndex := range table.children[callID] {
			if childIndex < 0 {
				out.WriteString("  0u,\n")
			} else {
				fmt.Fprintf(out, "  %du,\n", childIndex)
			}
		}
		out.WriteString("};\n\n")
	}
}

// callBoundaryAttributeSetForTest is
// TestEmittedAttributeSetCommentIsDerivedNotLiteral's own fault-injection
// seam (D-04-12): nil (the default, and the ONLY value any production
// caller ever observes) means "use emitCallBoundaryAttributeSet's real,
// always-empty derivation"; a non-nil slice substitutes a counterfactual
// non-empty set so a test can prove the rendered comment tracks the
// derivation itself rather than restating a literal beside it. Exported
// (via export_test.go's SetCallBoundaryAttributeSetForTest) only for the
// external cgen_test package; never set on a production path.
var callBoundaryAttributeSetForTest []string

// emitCallBoundaryAttributeSet is D-11-09's own named derivation point:
// under this package's architecture, emitCall (this file) never mints an
// optimizer-visible attribute at any Lang-to-Lang call boundary (D-11-04),
// so absent the test seam above this always returns an empty, non-nil
// slice. Every caller -- emitCallBoundaryAttributeComment below, and any
// future gate that asks the same question -- reads THIS one derivation,
// so a later change to emitCall's own emission automatically changes what
// every caller reports (D-04-12: no hand-written comment beside a
// generated value that can drift from it).
func emitCallBoundaryAttributeSet(functions []core.Function) []string {
	if callBoundaryAttributeSetForTest != nil {
		return append([]string{}, callBoundaryAttributeSetForTest...)
	}
	_ = functions
	return []string{}
}

// emitCallBoundaryAttributeComment renders
// emitCallBoundaryAttributeSet's own value (sorted, so the statement is
// byte-stable regardless of the program's own function declaration
// order) as a generated C comment naming the set's size, whether it is
// empty, and D-11-09 by identifier -- NAT-05's explicit, visible, empty
// call-boundary attribute set.
func emitCallBoundaryAttributeComment(functions []core.Function) string {
	set := append([]string{}, emitCallBoundaryAttributeSet(functions)...)
	sort.Strings(set)
	if len(set) == 0 {
		return "/* call-boundary attribute set: EMPTY (0 entries). Per D-11-09, cgen.emitCall never mints an optimizer-visible alias or capture attribute at any Lang-to-Lang call boundary this phase -- NAT-05 is satisfied by this explicit, generated empty set, not by silence. */\n"
	}
	return fmt.Sprintf("/* call-boundary attribute set: %d entries: %s (D-11-09). */\n", len(set), strings.Join(set, ", "))
}

// emitCallLookup carries the whole-program name/type tables emitCall needs
// to resolve an OpCall's CalleeID into its already-allocated generated C
// function name and C return-type name (D-11-04/D-11-08). A nil
// *emitCallLookup always fails to resolve -- this is deliberate, not an
// oversight: emitLinear's and emitBranchOperations' own OpCall arms (both
// single-function paths, cgen.go) call resolve on a nil lookup because a
// one-function program can never legally contain an OpCall (self-recursion
// is refused as a call-graph cycle before either arm could run), so those
// two call sites always take the "unresolved" branch. emitProgram (this
// file) is the one caller that ever supplies a real, populated lookup.
type emitCallLookup struct {
	indexByFunctionID map[string]int
	functionNames     []string
	returnTypeNames   []string
}

func (l *emitCallLookup) resolve(calleeID string) (functionName, returnTypeName string, ok bool) {
	if l == nil {
		return "", "", false
	}
	index, known := l.indexByFunctionID[calleeID]
	if !known {
		return "", "", false
	}
	return l.functionNames[index], l.returnTypeNames[index], true
}

// emitCall is D-11-04's single writer of a Lang-to-Lang call anywhere in
// cgen: it emits ONE C call expression assigning the resolved callee's
// result to the operation's own TargetID local, using the same inline
// provenance comment convention (`/* label: operationID */`) every
// surrounding straight-line arm already uses. It never mints an
// optimizer-visible attribute on either side of the call -- `restrict` is a
// property of a definition and its own prototype (D-05-01..D-05-04), never
// of a call site, and D-11-09 declares Phase 11 emits zero call-boundary
// alias attributes at all.
func emitCall(out *strings.Builder, calleeTypeName, targetLocal, calleeName, argumentLocal string, operation core.LinearOperation) {
	fmt.Fprintf(out, "  %s %s = %s(%s); /* call: %s */\n", calleeTypeName, targetLocal, calleeName, argumentLocal, operation.ID)
	fmt.Fprintf(out, "  (void)%s;\n", targetLocal)
}

func emitProgramCall(out *strings.Builder, calleeTypeName, targetLocal, calleeName, argumentLocal, childIndex string, operation core.LinearOperation) {
	fmt.Fprintf(out, "  %s %s = %s(%s, %s); /* call: %s */\n", calleeTypeName, targetLocal, calleeName, argumentLocal, childIndex, operation.ID)
	fmt.Fprintf(out, "  (void)%s;\n", targetLocal)
}

// emitProgram is Phase 11's whole-program C17 translation-unit assembler
// (D-11-01/D-11-03): the additive path Emit and EmitNative dispatch to
// whenever a validated program declares more than one function. Per
// D-11-03 it owns every whole-TU concern itself -- includes, typedefs,
// every support block, the resource ledger, every function's prototype,
// every function's definition, and exactly one `main` invoking the
// program's resolved entry function (callgraph.EntryFunction) -- while
// each function's own body writes statements and events only.
//
// Order of emission: the #include block; the LANG_BUFFER typedef when any
// emitted function's parameter type is Buffer; emitEventSupport (called,
// never forked) sized to the WHOLE program's total operation count, since
// every function shares one events buffer; emitDefectSupport when any
// function contains an OpDefect; every function prototype, in
// callgraph.Order's order; every function definition, in the same order;
// then `main`.
//
// This phase's corpus is entirely straight-line (D-11-24's tracer scope):
// no core.Match, no core.Block-based control flow, and no
// core.ForeignContract on any function this assembler emits. A function
// carrying any of those shapes is refused here with a named error rather
// than silently mis-emitted -- nothing in this phase's testdata/phase11
// corpus exercises a multi-function branch or foreign body; that remains a
// documented scope limit for a later phase, not a live gap this one
// papers over.
//
// executionJSON is accepted for signature symmetry with Emit/EmitNative's
// existing emitMatch(program, executionJSON) call convention and for a
// future multi-function branch body; it is unused today because this
// phase's assembler always writes the lang.execution/1 JSON document (the
// single behavior emitLinear's own single-function path already has).
func emitProgram(program core.Program, executionJSON bool) (string, error) {
	_ = executionJSON

	order, err := callgraph.Order(program)
	if err != nil {
		return "", err
	}
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		return "", err
	}

	byID := make(map[string]core.Function, len(program.Functions))
	for _, function := range program.Functions {
		byID[function.ID] = function
	}
	functions := make([]core.Function, 0, len(order))
	for _, id := range order {
		function, ok := byID[id]
		if !ok {
			return "", fmt.Errorf("emitProgram: call-graph order names unknown function %q", id)
		}
		if function.Match == nil && function.Linear == nil {
			return "", fmt.Errorf("function %q: has no linear body", function.ID)
		}
		if function.Match != nil && function.Linear == nil {
			return "", fmt.Errorf("function %q: match-only bodies are not supported by whole-program native emission this phase", function.ID)
		}
		if function.Match == nil && len(function.Linear.Blocks) > 0 {
			return "", fmt.Errorf("function %q: multi-function foreign-call bodies are not supported by native emission this phase", function.ID)
		}
		if function.ForeignContract != nil {
			return "", fmt.Errorf("function %q: multi-function foreign contracts are not supported by native emission this phase", function.ID)
		}
		functions = append(functions, function)
	}

	// D-15-18: graph and entry refusal always win. The supported-shape
	// contract above must also be established before schema-2 preflight,
	// whose size estimator assumes every emitted body is linear. Once those
	// checks have succeeded, bound every dynamic activation occurrence before
	// allocating emission tables or beginning generated-C serialization.
	preflightNodes, err := preflightInvocationPathTable(program, entry)
	if err != nil {
		return "", err
	}
	paths, err := buildInvocationPathTable(entry.ID, preflightNodes)
	if err != nil {
		return "", err
	}
	eventCapacity, err := invocationEventCapacity(preflightNodes, byID)
	if err != nil {
		return "", err
	}
	// C17 has no zero-length arrays. A zero-operation match still writes the
	// shared event document with count zero, so reserve one inert slot while
	// retaining the actual count as the sole serialization authority.
	if eventCapacity == 0 {
		eventCapacity = 1
	}
	liveResources := deriveProgramLiveResources(program)
	executionBytes, err := schema2ExecutionDocumentSize(entry, preflightNodes, paths, byID, liveResources)
	if err != nil {
		return "", err
	}
	if executionBytes > executionOutputLimit {
		return "", &executionOutputExceededError{limit: executionOutputLimit, observed: executionBytes}
	}
	ensureInvocationChildTables(&paths, functions)

	// Two-tier name allocation (D-11-08/T-11-06): ONE global cNames
	// allocates every function's own C name, iterating in
	// callgraph.Order's order so allocation is deterministic. Each
	// function's own return/parameter C type name is derived the same way
	// linearInput already derives it for the single-function path.
	globals := newCNames(linearFixedNames...)
	functionNames := make([]string, len(functions))
	typeNames := make([]string, len(functions))
	branchTypes := make(map[string]programBranchType)
	for index, function := range functions {
		functionNames[index] = globals.allocate(cName(function.Name), "function", index)
		if function.Match != nil {
			branchType, err := newProgramBranchType(function, program.DataTypes, globals)
			if err != nil {
				return "", fmt.Errorf("function %q: %w", function.ID, err)
			}
			branchTypes[function.ID] = branchType
			typeNames[index] = branchType.typeName
		} else {
			_, _, typeName, err := linearInput(function)
			if err != nil {
				return "", fmt.Errorf("function %q: %w", function.ID, err)
			}
			typeNames[index] = typeName
		}
	}
	// globalNames snapshots every name the global allocator has handed out
	// (function names plus the reserved list) so each function's own FRESH
	// per-function cNames (below) is seeded with the reserved list PLUS
	// every globally allocated name -- this is what keeps the second
	// function's places named lang_value_x rather than lang_value_x_2
	// (D-11-08): C block scope already makes locals independent; the seed
	// is what makes shadowing impossible.
	childTableNames := invocationChildTableNames(paths, globals)
	globalNames := make([]string, 0, len(globals.used))
	for name := range globals.used {
		globalNames = append(globalNames, name)
	}
	indexByFunctionID := make(map[string]int, len(functions))
	for index, function := range functions {
		indexByFunctionID[function.ID] = index
	}
	lookup := &emitCallLookup{indexByFunctionID: indexByFunctionID, functionNames: functionNames, returnTypeNames: typeNames}

	needsBuffer := false
	needsByte := false
	needsDefect := false
	for _, function := range functions {
		switch function.Parameter.Type {
		case "Buffer":
			needsBuffer = true
		case "Byte":
			needsByte = true
		}
		if functionHasDefect(function) {
			needsDefect = true
		}
	}

	invocationSerializationReachedForTest = true
	var out strings.Builder
	out.WriteString("/* generated by Codename Lang; schema lang.c17/0 */\n")
	out.WriteString("/* Moves below are authority transitions; C value assignment makes no ABI or zero-copy claim. */\n")
	out.WriteString(emitCallBoundaryAttributeComment(functions))
	out.WriteString("#include <stddef.h>\n#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n\n")
	emitInvocationPathTable(&out, paths, childTableNames)
	for _, function := range functions {
		if branchType, ok := branchTypes[function.ID]; ok {
			emitProgramBranchType(&out, branchType)
		}
	}
	if needsBuffer {
		out.WriteString("typedef struct LANG_BUFFER {\n  unsigned char bytes[4];\n  size_t length;\n} LANG_BUFFER;\n\n")
	}
	emitEventSupportSchema2(&out, eventCapacity, executionOutputLimit)
	if needsDefect {
		emitDefectSupport(&out)
	}
	if needsBuffer {
		emitProgramBufferWriter(&out, "LANG_BUFFER")
	}
	if needsByte {
		emitProgramByteWriter(&out)
	}
	if err := emitProgramLiveResourcesWriter(&out, liveResources); err != nil {
		return "", err
	}

	// Prototypes before definitions (D-11-03): what makes a forward-
	// referenced callee legal C17.
	for index := range functions {
		fmt.Fprintf(&out, "static %s %s(%s, unsigned int);\n", typeNames[index], functionNames[index], typeNames[index])
	}
	out.WriteString("\n")

	for index, function := range functions {
		var err error
		if branchType, ok := branchTypes[function.ID]; ok {
			err = emitProgramBranchFunction(&out, function, branchType, functionNames[index], globalNames, lookup, childTableNames)
		} else {
			err = emitProgramFunction(&out, function, typeNames[index], functionNames[index], globalNames, lookup, childTableNames)
		}
		if err != nil {
			return "", err
		}
	}

	entryIndex, ok := indexByFunctionID[entry.ID]
	if !ok {
		return "", fmt.Errorf("emitProgram: resolved entry %q is absent from the emitted function set", entry.ID)
	}
	input, initializer, entryTypeName, err := linearInput(entry)
	entryBranch, entryIsBranch := branchTypes[entry.ID]
	if entryIsBranch {
		entryTypeName = entryBranch.typeName
	} else if err != nil {
		return "", err
	}
	out.WriteString("int main(int argc, char **argv) {\n")
	// Task 2 consumes this literal table while recording /2 events. It is
	// already emitted in Task 1 so all internal calls share one stable index;
	// retain an explicit harmless reference until the event writer reads it.
	out.WriteString("  (void)lang_invocations;\n")
	// A nullary match can produce no operation events at all; retain a harmless
	// reference so the shared schema-2 support remains valid under -Werror.
	out.WriteString("  (void)lang_record_event;\n")
	// A declared-but-never-called function (D-11-05's unreachable-function
	// contract: it is still emitted as a real C definition, dead code that
	// does not change entry resolution) would otherwise be flagged
	// unused-function under this project's -Werror build. Taking each
	// function's own address here is a harmless, side-effect-free reference
	// -- never a call -- that keeps every emitted definition provably
	// reachable from a linker's perspective without changing which
	// function `main` actually invokes.
	for index := range functions {
		fmt.Fprintf(&out, "  (void)%s;\n", functionNames[index])
	}
	// lang_write_buffer_hex/lang_write_byte (emitProgramBufferWriter/
	// emitProgramByteWriter above) are emitted whenever ANY function in the
	// program declares that parameter type -- not only when the RESOLVED
	// ENTRY function does. A declared-but-never-called Buffer/Byte
	// function (D-11-05's own unreachable-function contract) whose type
	// differs from the entry's own type would otherwise leave the
	// corresponding writer genuinely unreferenced, failing this project's
	// -Werror -Wunused-function build (Rule 1: a real, reachable bug this
	// gate corpus's own touch function, declared Buffer-typed and never
	// called, first exposed). The same harmless, side-effect-free
	// address-taking reference used for user functions above applies here
	// too.
	if needsBuffer {
		out.WriteString("  (void)lang_write_buffer_hex;\n")
	}
	if needsByte {
		out.WriteString("  (void)lang_write_byte;\n")
	}
	out.WriteString("  if (argc != 2) return 64;\n")
	if entryIsBranch {
		fmt.Fprintf(&out, "  %s lang_entry_input;\n", entryTypeName)
		for index, alternative := range entryBranch.alternatives {
			prefix := "if"
			if index > 0 {
				prefix = "else if"
			}
			fmt.Fprintf(&out, "  %s (strcmp(argv[1], %s) == 0) lang_entry_input = %s;\n", prefix, strconv.Quote(alternative.source), alternative.cName)
		}
		out.WriteString("  else return 65;\n")
	} else {
		fmt.Fprintf(&out, "  if (strcmp(argv[1], %s) != 0) return 65;\n", strconv.Quote(input))
		fmt.Fprintf(&out, "  %s lang_entry_input = %s;\n", entryTypeName, initializer)
	}
	fmt.Fprintf(&out, "  %s lang_entry_output = %s(lang_entry_input, 0u);\n", entryTypeName, functionNames[entryIndex])
	if entryIsBranch {
		fmt.Fprintf(&out, "  const char *lang_entry_name = %s(lang_entry_output);\n", entryBranch.nameFunction)
		out.WriteString("  if (lang_entry_name == NULL) return 70;\n")
		out.WriteString("  if (!lang_write_literal(\"{\\\"schema\\\":\\\"lang.execution/2\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\")) return 74;\n")
		out.WriteString("  if (!lang_write_json_string(lang_entry_name)) return 74;\n")
	} else if entryTypeName == "LANG_BUFFER" {
		out.WriteString("  if (!lang_write_literal(\"{\\\"schema\\\":\\\"lang.execution/2\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\\\"\")) return 74;\n")
		out.WriteString("  if (!lang_write_buffer_hex(&lang_entry_output)) return 74;\n")
	} else {
		out.WriteString("  if (!lang_write_literal(\"{\\\"schema\\\":\\\"lang.execution/2\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\\\"\")) return 74;\n")
		out.WriteString("  if (!lang_write_byte(lang_entry_output)) return 74;\n")
	}
	if entryIsBranch {
		out.WriteString("  if (!lang_write_literal(\"},\\\"events\\\":[\")) return 74;\n")
	} else {
		out.WriteString("  if (!lang_write_literal(\"\\\"},\\\"events\\\":[\")) return 74;\n")
	}
	out.WriteString("  if (!lang_write_events()) return 74;\n")
	out.WriteString("  if (!lang_write_literal(\"],\\\"live_resources\\\":\")) return 74;\n")
	out.WriteString("  if (!lang_write_live_resources()) return 74;\n")
	out.WriteString("  if (!lang_write_literal(\"}\\n\")) return 74;\n")
	out.WriteString("  return 0;\n}\n")

	return out.String(), nil
}

// emitProgramBufferWriter/emitProgramByteWriter are the multi-function
// assembler's own copies of emitLinearOutputSupport's scalar-value writer
// selection (cgen.go): a Buffer-parameterized program needs
// lang_write_buffer_hex, a Byte-parameterized one needs lang_write_byte.
// Kept here (not shared with emitLinearOutputSupport) because that
// function also unconditionally calls emitEventSupport, which emitProgram
// has already called once for the whole program (D-11-01: calling the
// single-function helper a second time would duplicate every LANG_EVENT
// declaration it contains).
func emitProgramBufferWriter(out *strings.Builder, typeName string) {
	fmt.Fprintf(out, "static int lang_write_buffer_hex(const %s *value) {\n", typeName)
	out.WriteString("  static const char hex[] = \"0123456789abcdef\";\n  size_t index;\n")
	out.WriteString("  if (value->length > sizeof value->bytes) return 0;\n")
	out.WriteString("  for (index = 0u; index < value->length; index++) {\n")
	out.WriteString("    char encoded[2] = {hex[value->bytes[index] >> 4u], hex[value->bytes[index] & 0x0fu]};\n")
	out.WriteString("    if (!lang_write_bytes(encoded, sizeof encoded)) return 0;\n  }\n  return 1;\n}\n\n")
}

func emitProgramByteWriter(out *strings.Builder) {
	out.WriteString("static int lang_write_byte(unsigned char value) {\n")
	out.WriteString("  char encoded[3];\n  int length = snprintf(encoded, sizeof encoded, \"%u\", (unsigned int)value);\n")
	out.WriteString("  return length > 0 && (size_t)length < sizeof encoded && lang_write_bytes(encoded, (size_t)length);\n}\n\n")
}

// emitProgramLiveResourcesWriter serializes the collection derived by
// emitProgram. The C writer receives canonical JSON computed from that one
// collection, so the output-size preflight and document tail share the same
// result without adding a resource ledger for unsupported foreign shapes.
func emitProgramLiveResourcesWriter(out *strings.Builder, liveResources []string) error {
	encoded, err := json.Marshal(liveResources)
	if err != nil {
		return fmt.Errorf("encode derived live resources: %w", err)
	}
	fmt.Fprintf(out, "static int lang_write_live_resources(void) {\n  return lang_write_literal(%s);\n}\n\n", strconv.Quote(string(encoded)))
	return nil
}

type programBranchAlternative struct {
	source, cName string
}

// programBranchType carries the checked nominal branch layout into the one
// whole-program writer. Phase 16 admits only nullary alternatives here; the
// payload record lowering remains behind its existing unsupported boundary.
type programBranchType struct {
	typeName, nameFunction string
	alternatives           []programBranchAlternative
	bySource               map[string]string
}

func newProgramBranchType(function core.Function, dataTypes []core.DataType, names *cNames) (programBranchType, error) {
	var dataType core.DataType
	for _, candidate := range dataTypes {
		if candidate.Name == function.Parameter.Type {
			dataType = candidate
			break
		}
	}
	if dataType.Name == "" || len(dataType.Alternatives) == 0 {
		return programBranchType{}, fmt.Errorf("branch C emitter cannot find alternatives for data type %q", function.Parameter.Type)
	}
	for _, detail := range dataType.AlternativeDetails {
		if detail.PayloadType != "" {
			return programBranchType{}, fmt.Errorf("branch payload shapes are not supported by whole-program native emission this phase")
		}
	}
	result := programBranchType{
		typeName:     names.allocate(cName(dataType.Name), "type", 0),
		bySource:     make(map[string]string, len(dataType.Alternatives)),
		alternatives: make([]programBranchAlternative, 0, len(dataType.Alternatives)),
	}
	for index, alternative := range dataType.Alternatives {
		cAlternative := names.allocate(result.typeName+"_"+cName(alternative), "alternative", index)
		result.alternatives = append(result.alternatives, programBranchAlternative{source: alternative, cName: cAlternative})
		result.bySource[alternative] = cAlternative
	}
	result.nameFunction = names.allocate(result.typeName+"_name", "type_name", 0)
	return result, nil
}

func emitProgramBranchType(out *strings.Builder, branchType programBranchType) {
	fmt.Fprintf(out, "typedef enum %s {\n", branchType.typeName)
	for index, alternative := range branchType.alternatives {
		fmt.Fprintf(out, "  %s = %d,\n", alternative.cName, index)
	}
	fmt.Fprintf(out, "} %s;\n\n", branchType.typeName)
	fmt.Fprintf(out, "static const char *%s(%s value) {\n  switch (value) {\n", branchType.nameFunction, branchType.typeName)
	for _, alternative := range branchType.alternatives {
		fmt.Fprintf(out, "    case %s: return %s;\n", alternative.cName, strconv.Quote(alternative.source))
	}
	out.WriteString("  }\n  return NULL;\n}\n\n")
}

// emitProgramBranchFunction lowers a match/branch body into the same shared
// schema-2 event buffer and bare-return ABI as ordinary emitProgram functions.
// It intentionally never writes an execution document: main owns that once.
func emitProgramBranchFunction(out *strings.Builder, function core.Function, branchType programBranchType, functionName string, globalNames []string, lookup *emitCallLookup, childTableNames map[string]string) error {
	if function.Match == nil {
		return fmt.Errorf("function %q: branch writer requires a match body", function.ID)
	}
	if function.Linear == nil {
		fmt.Fprintf(out, "static %s %s(%s value, unsigned int invocation_index) {\n  (void)invocation_index;\n  switch (value) {\n", branchType.typeName, functionName, branchType.typeName)
		for _, arm := range function.Match.Arms {
			pattern, known := branchType.bySource[arm.Pattern]
			valueName, valueKnown := branchType.bySource[arm.Value]
			if !known || !valueKnown {
				return fmt.Errorf("function %q: match arm %q names unknown alternative", function.ID, arm.ID)
			}
			fmt.Fprintf(out, "    case %s: return %s;\n", pattern, valueName)
		}
		out.WriteString("  }\n  abort();\n}\n\n")
		return nil
	}

	places := make(map[string]core.Place, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	parameter, ok := places[function.Parameter.ID]
	if !ok {
		return fmt.Errorf("function %q: branch parameter place is absent", function.ID)
	}
	reserved := append(append([]string{}, linearFixedNames...), globalNames...)
	names := newCNames(reserved...)
	locals := make(map[string]string, len(function.Linear.Places))
	for index, place := range function.Linear.Places {
		locals[place.ID] = names.allocate(cLocal(place.Name), "place", index)
	}
	operations := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		operations[operation.ID] = operation
	}
	blocks := make(map[string]core.Block, len(function.Linear.Blocks))
	for _, block := range function.Linear.Blocks {
		blocks[block.ID] = block
	}

	fmt.Fprintf(out, "static %s %s(%s %s, unsigned int invocation_index) {\n  switch (%s) {\n", branchType.typeName, functionName, branchType.typeName, locals[parameter.ID], locals[parameter.ID])
	for _, arm := range function.Match.Arms {
		block, known := blocks[arm.BlockID]
		if !known {
			return fmt.Errorf("arm %q references unknown block %q", arm.ID, arm.BlockID)
		}
		pattern, known := branchType.bySource[arm.Pattern]
		if !known {
			return fmt.Errorf("arm %q names unknown alternative %q", arm.ID, arm.Pattern)
		}
		fmt.Fprintf(out, "    case %s: {\n", pattern)
		declared := map[string]bool{parameter.ID: true}
		for _, operationID := range block.OperationIDs {
			operation, known := operations[operationID]
			if !known {
				return fmt.Errorf("block references unknown operation %q", operationID)
			}
			if err := emitProgramBranchOperation(out, function, operation, branchType.typeName, locals, places, declared, lookup, childTableNames); err != nil {
				return err
			}
		}
		out.WriteString("    }\n")
	}
	out.WriteString("  }\n  abort();\n}\n\n")
	return nil
}

func emitProgramBranchOperation(out *strings.Builder, function core.Function, operation core.LinearOperation, typeName string, locals map[string]string, places map[string]core.Place, declared map[string]bool, lookup *emitCallLookup, childTableNames map[string]string) error {
	if _, known := places[operation.SourceID]; !known {
		return fmt.Errorf("operation %q has invalid source", operation.ID)
	}
	switch operation.Kind {
	case core.OpCopy, core.OpMove, core.OpBorrowShared, core.OpBorrowExclusive:
		if _, exists := places[operation.TargetID]; !exists || declared[operation.TargetID] {
			return fmt.Errorf("operation %q has invalid target", operation.ID)
		}
		label, eventKind := "copy", "value.copied"
		if operation.Kind == core.OpMove {
			label, eventKind = "authority transfer", "value.transferred"
		} else if operation.Kind == core.OpBorrowShared {
			label, eventKind = "shared borrow representation", "value.borrowed"
		} else if operation.Kind == core.OpBorrowExclusive {
			label, eventKind = "exclusive borrow representation", "value.borrowed_exclusive"
		}
		fmt.Fprintf(out, "      %s %s = %s; /* %s: %s */\n      (void)%s;\n", typeName, locals[operation.TargetID], locals[operation.SourceID], label, operation.ID, locals[operation.TargetID])
		fmt.Fprintf(out, "      if (!lang_record_event(%s, %s, %s, %s, %s, %s, lang_invocations[invocation_index], NULL)) abort();\n", strconv.Quote(eventKind), strconv.Quote(operation.ID+":event"), strconv.Quote(function.ID), strconv.Quote(operation.SourceID), strconv.Quote(operation.TargetID), strconv.Quote(operation.TypeID))
		declared[operation.TargetID] = true
	case core.OpReturn:
		fmt.Fprintf(out, "      if (!lang_record_event(%s, %s, %s, %s, NULL, %s, lang_invocations[invocation_index], NULL)) abort();\n      return %s;\n", strconv.Quote("function.returned"), strconv.Quote(operation.ID+":event:returned"), strconv.Quote(function.ID), strconv.Quote(operation.SourceID), strconv.Quote(operation.TypeID), locals[operation.SourceID])
	case core.OpCall:
		if _, exists := places[operation.TargetID]; !exists || declared[operation.TargetID] {
			return fmt.Errorf("operation %q has invalid target", operation.ID)
		}
		calleeName, calleeTypeName, ok := lookup.resolve(operation.CalleeID)
		if !ok {
			return fmt.Errorf("operation %q: call to unresolved callee %q", operation.ID, operation.CalleeID)
		}
		childTable, ok := childTableNames[operation.ID]
		if !ok {
			return fmt.Errorf("operation %q has no invocation child table", operation.ID)
		}
		fmt.Fprintf(out, "      if (!lang_record_event(%s, %s, %s, %s, %s, %s, lang_invocations[invocation_index], %s)) abort();\n", strconv.Quote("function.called"), strconv.Quote(operation.ID+":event:called"), strconv.Quote(function.ID), strconv.Quote(operation.SourceID), strconv.Quote(operation.TargetID), strconv.Quote(operation.TypeID), strconv.Quote(operation.CalleeID))
		emitProgramCall(out, calleeTypeName, locals[operation.TargetID], calleeName, locals[operation.SourceID], childTable+"[invocation_index]", operation)
		declared[operation.TargetID] = true
	case core.OpForeignCall, core.OpFail, core.OpDefect, core.OpConstructPayload, core.OpDestructurePayload:
		return fmt.Errorf("operation %q: unsupported branch operation kind %q in whole-program native emission", operation.ID, operation.Kind)
	default:
		return fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
	}
	return nil
}

// emitProgramFunction writes ONE function's own C definition: its own
// fresh, per-function cNames (D-11-08: seeded with the reserved list plus
// every globally allocated name) locals allocator, its straight-line
// operations, and a bare C `return` -- never the whole-TU
// lang.execution/1 JSON tail emitLinear's own single-function path writes,
// since that tail belongs to `main` alone in the multi-function assembler
// (a non-entry callee's own return value is consumed by its caller's own
// OpCall, never written to stdout directly; the entry function's own
// return value is written to stdout exactly once, by `main`, after it
// returns).
func emitProgramFunction(out *strings.Builder, function core.Function, typeName, functionName string, globalNames []string, lookup *emitCallLookup, childTableNames map[string]string) error {
	places := make(map[string]core.Place, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	parameter, ok := places[function.Parameter.ID]
	if !ok {
		return fmt.Errorf("function %q: linear parameter place is absent", function.ID)
	}

	reserved := make([]string, 0, len(linearFixedNames)+len(globalNames))
	reserved = append(reserved, linearFixedNames...)
	reserved = append(reserved, globalNames...)
	names := newCNames(reserved...)

	locals := make(map[string]string, len(function.Linear.Places))
	locals[parameter.ID] = names.allocate(cLocal(parameter.Name), "place", 0)
	for index, place := range function.Linear.Places {
		if place.ID == parameter.ID {
			continue
		}
		locals[place.ID] = names.allocate(cLocal(place.Name), "place", index)
	}

	fmt.Fprintf(out, "static %s %s(%s %s, unsigned int invocation_index) {\n", typeName, functionName, typeName, locals[parameter.ID])
	// Leaf functions do not consult a child table, but every internal
	// function still receives the occurrence index. Keep generated C clean
	// under -Werror while preserving the uniform ABI.
	out.WriteString("  (void)invocation_index;\n")
	declared := map[string]bool{parameter.ID: true}
	returned := false
	for _, operation := range function.Linear.Operations {
		source, sourceKnown := places[operation.SourceID]
		if !sourceKnown {
			return fmt.Errorf("operation %q has invalid source", operation.ID)
		}
		switch operation.Kind {
		case core.OpCopy, core.OpMove, core.OpBorrowShared, core.OpBorrowExclusive:
			target, exists := places[operation.TargetID]
			if !exists || declared[operation.TargetID] {
				return fmt.Errorf("operation %q has invalid target", operation.ID)
			}
			label := "copy"
			marker := ""
			if operation.Kind == core.OpMove {
				label = "authority transfer"
				marker = " /* lang:mutation-site */"
			} else if operation.Kind == core.OpBorrowShared {
				label = "shared borrow representation"
			} else if operation.Kind == core.OpBorrowExclusive {
				label = "exclusive borrow representation"
			}
			fmt.Fprintf(out, "  %s %s = %s; /* %s: %s */%s\n", typeName, locals[target.ID], locals[source.ID], label, operation.ID, marker)
			fmt.Fprintf(out, "  (void)%s;\n", locals[target.ID])
			eventKind := "value.copied"
			if operation.Kind == core.OpMove {
				eventKind = "value.transferred"
			} else if operation.Kind == core.OpBorrowShared {
				eventKind = "value.borrowed"
			} else if operation.Kind == core.OpBorrowExclusive {
				eventKind = "value.borrowed_exclusive"
			}
			fmt.Fprintf(out, "  if (!lang_record_event(%s, %s, %s, %s, %s, %s, lang_invocations[invocation_index], NULL)) abort();\n",
				strconv.Quote(eventKind), strconv.Quote(operation.ID+":event"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TargetID), strconv.Quote(operation.TypeID))
			declared[operation.TargetID] = true
		case core.OpCall:
			target, exists := places[operation.TargetID]
			if !exists || declared[operation.TargetID] {
				return fmt.Errorf("operation %q has invalid target", operation.ID)
			}
			calleeName, calleeTypeName, ok := lookup.resolve(operation.CalleeID)
			if !ok {
				return fmt.Errorf("operation %q: call to unresolved callee %q", operation.ID, operation.CalleeID)
			}
			childTable, ok := childTableNames[operation.ID]
			if !ok {
				return fmt.Errorf("operation %q has no invocation child table", operation.ID)
			}
			fmt.Fprintf(out, "  if (!lang_record_event(%s, %s, %s, %s, %s, %s, lang_invocations[invocation_index], %s)) abort();\n",
				strconv.Quote("function.called"), strconv.Quote(operation.ID+":event:called"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TargetID), strconv.Quote(operation.TypeID), strconv.Quote(operation.CalleeID))
			emitProgramCall(out, calleeTypeName, locals[target.ID], calleeName, locals[source.ID], childTable+"[invocation_index]", operation)
			declared[operation.TargetID] = true
		case core.OpReturn:
			// Unlike emitLinear's single-function OpReturn arm, this never
			// writes the lang.execution/1 JSON tail: that belongs to
			// `main` alone (emitProgram), written exactly once after the
			// resolved entry function returns. A callee's own event is
			// still recorded here, attributed to ITS OWN function ID
			// (interp.terminalOutcome's identical rule, D-10-32), so the
			// caller's own events and the callee's own function.returned
			// event both land in the shared events buffer in execution
			// order regardless of call depth.
			fmt.Fprintf(out, "  if (!lang_record_event(%s, %s, %s, %s, NULL, %s, lang_invocations[invocation_index], NULL)) abort(); /* returned place: %s */\n",
				strconv.Quote("function.returned"), strconv.Quote(operation.ID+":event:returned"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TypeID), operation.ID)
			fmt.Fprintf(out, "  return %s;\n", locals[source.ID])
			returned = true
		default:
			return fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
		}
	}
	if !returned {
		return fmt.Errorf("function %q: linear body has no return operation", function.ID)
	}
	out.WriteString("}\n\n")
	return nil
}
