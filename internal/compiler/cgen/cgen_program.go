package cgen

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
)

// maxInvocationPathTableNodes is the deliberately fixed D-15-16 ceiling on
// static activation occurrences in a native program. It is not a byte-size
// budget and is intentionally independent of pathoracle.MaxPaths.
const maxInvocationPathTableNodes = 4096

// EmitProgramNativeForTest exposes the direct whole-program native route to
// cross-package differential tests. It deliberately bypasses both public
// dispatchers and invokes exactly emitProgram(program, true), so it cannot
// silently certify the retained N=1 legacy route before Plan 16-09's cutover.
// Callers must supply an already checked and validated core.Program.
func EmitProgramNativeForTest(program core.Program) (string, error) {
	return emitProgram(program, true)
}

var executionOutputLimit = execution.MaxDocumentBytes

// These bounds are shared by the terminal writer and its output-size
// preflight. A Buffer payload is rejected above four bytes and Byte payloads
// are emitted as unsigned decimal values.
const (
	maxTerminalBufferPayloadBytes = 4
	maxTerminalBytePayloadValue   = 255
	maxTerminalBytePayloadDigits  = 3
)

type terminalPayloadEncoding uint8

const (
	terminalPayloadOmitted terminalPayloadEncoding = iota
	terminalPayloadBuffer
	terminalPayloadByte
	terminalPayloadNestedTag
)

func terminalPayloadEncodingFor(field programPayloadField, dataTypes []core.DataType) terminalPayloadEncoding {
	if field.cType == "SCHWAY_BUFFER" {
		return terminalPayloadBuffer
	}
	if field.cType == "unsigned char" {
		if _, nested := findProgramDataTypeByName(dataTypes, field.payloadType); nested {
			return terminalPayloadNestedTag
		}
		return terminalPayloadByte
	}
	return terminalPayloadOmitted
}

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
				capacity++ // A bare match records its one selected-arm return.
				continue
			}
			return 0, fmt.Errorf("function %q: has no linear body", function.ID)
		}
		capacity += len(function.Linear.Operations)
	}
	return capacity, nil
}

// schema2ExecutionDocumentSize calculates a conservative canonical byte bound
// for the straight-line document represented by the bounded occurrence table.
// The event order does not affect JSON length; field presence and spellings
// mirror emitProgramFunction and emitEventSupportSchema2.
func schema2ExecutionDocumentSize(entry core.Function, nodes []invocationPreflightNode, paths invocationPathTable, byID map[string]core.Function, liveResources []string, dataTypes []core.DataType) (int, error) {
	events := make([]execution.Event, 0)
	for index, node := range nodes {
		function, ok := byID[node.functionID]
		if !ok || (function.Linear == nil && function.Match == nil) {
			return 0, fmt.Errorf("execution-size node %d names invalid function %q", index, node.functionID)
		}
		if function.Linear == nil {
			if function.Match != nil {
				events = append(events, execution.Event{
					Schema: execution.Schema2, ID: function.ID + ":match:return", Kind: "function.returned",
					FunctionID: function.ID, Invocation: paths.invocations[index], SourcePlace: function.Parameter.ID, TypeID: function.Parameter.Type,
				})
			}
			continue
		}
		for _, operation := range function.Linear.Operations {
			event := execution.Event{
				Schema: execution.Schema2, FunctionID: function.ID, Invocation: paths.invocations[index],
				SourcePlace: operation.SourceID, TargetPlace: operation.TargetID, TypeID: operation.TypeID,
			}
			switch operation.Kind {
			case core.OpConst:
				// Constants change local state but intentionally add no public event.
				continue
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
			case core.OpDestructurePayload:
				event.ID, event.Kind, event.TargetPlace = operation.ID+":event", "value.payload_destructured", operation.PayloadTargetID
			case core.OpConstructPayload:
				event.ID, event.Kind = operation.ID+":event", "value.payload_constructed"
			case core.OpDefect:
				event.ID, event.Kind, event.Output = operation.ID+":event:defected", "function.defected", operation.Reason
			case core.OpForeignCall, core.OpRelease:
				// The Phase 23 retained app records model output only. Its
				// operation ABI and cleanup are checked independently; compiler
				// event rows are not physical allocation evidence.
				continue
			default:
				return 0, fmt.Errorf("operation %q has unsupported schema-2 output kind %q", operation.ID, operation.Kind)
			}
			events = append(events, event)
		}
	}
	outcomeValue := ""
	if entry.Match != nil {
		var returnedType core.DataType
		for _, candidate := range dataTypes {
			if candidate.Name == entry.ReturnType {
				returnedType = candidate
				break
			}
		}
		if returnedType.Name == "" {
			return 0, fmt.Errorf("execution-size entry %q has unknown returned data type %q", entry.ID, entry.ReturnType)
		}
		layout := check.PayloadRecordLayout(returnedType)
		for _, alternative := range returnedType.Alternatives {
			value := alternative
			var layoutField *core.LayoutField
			for index := range layout.Fields {
				if layout.Fields[index].Name == "field_"+alternative {
					layoutField = &layout.Fields[index]
					break
				}
			}
			if detail := core.LookupAlternativeDetail(returnedType, alternative); detail.PayloadType != "" && layoutField != nil {
				switch terminalPayloadEncodingFor(programPayloadField{name: layoutField.Name, cType: layoutField.CType, payloadType: detail.PayloadType}, dataTypes) {
				case terminalPayloadBuffer:
					value += ":" + strings.Repeat("f", maxTerminalBufferPayloadBytes*2)
				case terminalPayloadByte:
					value += fmt.Sprintf(":%d", maxTerminalBytePayloadValue)
				case terminalPayloadNestedTag:
					nested, _ := findProgramDataTypeByName(dataTypes, detail.PayloadType)
					for _, nestedAlternative := range nested.Alternatives {
						candidate := value + ":" + nestedAlternative
						if encodedStringContentLength(candidate) > encodedStringContentLength(outcomeValue) {
							outcomeValue = candidate
						}
					}
					continue
				}
			}
			if encodedStringContentLength(value) > encodedStringContentLength(outcomeValue) {
				outcomeValue = value
			}
		}
	} else {
		if entry.ReturnType == "U64" {
			// The canonical decimal representation is at most 20 digits. Use
			// that bound here because the returned place may be an input, a
			// constant, or a copy of either.
			outcomeValue = "18446744073709551615"
		} else {
			var err error
			outcomeValue, _, _, err = linearInput(entry)
			if err != nil {
				return 0, err
			}
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

// encodedStringContentLength bounds the byte expansion in
// schway_write_json_string_content.
// Marshal's additional HTML/U+2028 escaping only makes the estimate more
// conservative for source names containing those characters.
func encodedStringContentLength(value string) int {
	encoded, _ := json.Marshal(value)
	return len(encoded) - 2 // exclude the enclosing quotes
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
		tableNames[callID] = names.allocate("schway_child_index_"+cName(callID), "invocation_child", ordinal)
	}
	return tableNames
}

func emitInvocationPathTable(out *strings.Builder, table invocationPathTable, tableNames map[string]string) {
	out.WriteString("static const char *schway_invocations[] = {\n")
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
// Order of emission: the #include block; the SCHWAY_BUFFER typedef when any
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
// phase's assembler always writes the schway.execution/1 JSON document (the
// single behavior emitLinear's own single-function path already has).
type programEntryShell uint8

const (
	programExecutionShell programEntryShell = iota
	programApplicationShell
)

func emitProgram(program core.Program, executionJSON bool) (string, error) {
	return emitProgramWithShell(program, programExecutionShell, executionJSON)
}

func emitProgramWithShell(program core.Program, shell programEntryShell, executionJSON bool) (string, error) {
	_ = executionJSON
	// Every caller observes whether THIS attempt crossed into C
	// serialization; ordered-refusal tests must not inherit a prior success.
	invocationSerializationReachedForTest = false

	order, err := callgraph.Order(program)
	if err != nil {
		return "", err
	}
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		return "", err
	}
	localFileByteEntry := isLocalFileByteFunction(entry)
	if hasLocalOwnerFacts(entry) {
		if !localFileByteEntry {
			return "", fmt.Errorf("entry function %q: unsupported local-owner C shape", entry.ID)
		}
		if err := validateLocalFileByteFunction(entry); err != nil {
			return "", err
		}
	}
	if shell == programApplicationShell && (entry.Match != nil || entry.Linear == nil || entry.ReturnType != "U64" || (entry.Parameter.Type != "U64" && !localFileByteEntry)) {
		return "", fmt.Errorf("application entry %q: only a linear U64-to-U64 entry is supported", entry.ID)
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
		if function.Match == nil && len(function.Linear.Blocks) > 0 {
			return "", fmt.Errorf("function %q: multi-function foreign-call bodies are not supported by native emission this phase", function.ID)
		}
		// Plan 16-05's human-selected cut-m004 applies to every program
		// cardinality. A legacy pointer-specialized body must not become
		// admissible merely because an otherwise ordinary caller is present.
		if function.Match == nil && (selectsByPointerLowering(function, function.Linear) || selectsByPointerLoweringSharedOnly(function, function.Linear)) {
			return "", fmt.Errorf("function %q: by-pointer bodies are not supported by whole-program native emission this phase", function.ID)
		}
		if function.ForeignContract != nil {
			return "", fmt.Errorf("function %q: multi-function foreign contracts are not supported by native emission this phase", function.ID)
		}
		functions = append(functions, function)
	}
	localFileByteProgram := false
	for _, function := range functions {
		if !hasLocalOwnerFacts(function) {
			continue
		}
		if !isLocalFileByteFunction(function) {
			return "", fmt.Errorf("function %q: unsupported local-owner C shape", function.ID)
		}
		if err := validateLocalFileByteFunction(function); err != nil {
			return "", err
		}
		localFileByteProgram = true
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
	executionBytes, err := schema2ExecutionDocumentSize(entry, preflightNodes, paths, byID, liveResources, program.DataTypes)
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
	// function's parameter and return C types are allocated independently.
	globals := newCNames(linearFixedNames...)
	functionNames := make([]string, len(functions))
	parameterTypeNames := make([]string, len(functions))
	returnTypeNames := make([]string, len(functions))
	branchTypes := make(map[string]programBranchType)
	branchTypeOrder := make([]string, 0)
	branchTypeFor := func(typeName, representationType string) (programBranchType, error) {
		if branchType, ok := branchTypes[typeName]; ok {
			return branchType, nil
		}
		branchType, err := newProgramBranchType(typeName, representationType, program.DataTypes, globals)
		if err != nil {
			return programBranchType{}, err
		}
		branchTypes[typeName] = branchType
		branchTypeOrder = append(branchTypeOrder, typeName)
		return branchType, nil
	}
	for index, function := range functions {
		functionNames[index] = globals.allocate(cName(function.Name), "function", index)
		if function.Match != nil {
			parameterBranchType, err := branchTypeFor(function.Parameter.Type, function.Parameter.Type)
			if err != nil {
				return "", fmt.Errorf("function %q: %w", function.ID, err)
			}
			returnBranchType, err := branchTypeFor(function.ReturnType, function.ReturnType)
			if err != nil {
				return "", fmt.Errorf("function %q: %w", function.ID, err)
			}
			parameterTypeNames[index] = parameterBranchType.typeName
			returnTypeNames[index] = returnBranchType.typeName
		} else {
			_, _, typeName, err := linearInput(function)
			if err != nil {
				return "", fmt.Errorf("function %q: %w", function.ID, err)
			}
			parameterTypeNames[index] = typeName
			switch function.ReturnType {
			case "Buffer":
				returnTypeNames[index] = "SCHWAY_BUFFER"
			case "Byte":
				returnTypeNames[index] = "unsigned char"
			case "U64":
				returnTypeNames[index] = "uint64_t"
			default:
				return "", fmt.Errorf("function %q: unsupported linear C return type %q", function.ID, function.ReturnType)
			}
		}
	}
	// globalNames snapshots every name the global allocator has handed out
	// (function names plus the reserved list) so each function's own FRESH
	// per-function cNames (below) is seeded with the reserved list PLUS
	// every globally allocated name -- this is what keeps the second
	// function's places named schway_value_x rather than schway_value_x_2
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
	lookup := &emitCallLookup{indexByFunctionID: indexByFunctionID, functionNames: functionNames, returnTypeNames: returnTypeNames}

	needsBuffer := false
	needsByte := false
	needsU64 := false
	needsDefect := false
	for _, function := range functions {
		switch function.Parameter.Type {
		case "Buffer":
			needsBuffer = true
		case "Byte":
			needsByte = true
		case "U64":
			needsU64 = true
		}
		switch function.ReturnType {
		case "Buffer":
			needsBuffer = true
		case "Byte":
			needsByte = true
		case "U64":
			needsU64 = true
		}
		if function.Linear != nil {
			for _, operation := range function.Linear.Operations {
				if operation.Kind == core.OpConst {
					needsU64 = true
				}
			}
		}
		if functionHasDefect(function) {
			needsDefect = true
		}
		if branchType, ok := branchTypes[function.Parameter.Type]; ok && branchType.hasPayload && branchType.hasCType("SCHWAY_BUFFER") {
			needsBuffer = true
		}
	}

	invocationSerializationReachedForTest = true
	var out strings.Builder
	out.WriteString("/* generated by Schway; schema schway.c17/0 */\n")
	out.WriteString("/* Moves below are authority transitions; C value assignment makes no ABI or zero-copy claim. */\n")
	out.WriteString(emitCallBoundaryAttributeComment(functions))
	out.WriteString("#include <stddef.h>\n#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n\n")
	if localFileByteProgram {
		out.WriteString("#include \"local/adapter.h\"\n\n")
	}
	if needsU64 {
		out.WriteString("#include <stdint.h>\n\n")
		out.WriteString("#if !defined(UINT64_MAX)\n#error \"Schway U64 requires exact-width uint64_t support\"\n#endif\n")
		out.WriteString("#if UINT64_MAX != 18446744073709551615ULL\n#error \"Schway U64 requires an exact 64-bit unsigned type\"\n#endif\n\n")
	}
	emitInvocationPathTable(&out, paths, childTableNames)
	if needsBuffer {
		out.WriteString("typedef struct SCHWAY_BUFFER {\n  unsigned char bytes[4];\n  size_t length;\n} SCHWAY_BUFFER;\n\n")
	}
	for _, typeName := range branchTypeOrder {
		emitProgramBranchType(&out, branchTypes[typeName])
	}
	entryReturnBranch, entryReturnIsBranch := branchTypes[entry.ReturnType]
	needsTerminalJSONContent := entryReturnIsBranch && entryReturnBranch.hasPayload
	emitEventSupportSchema2(&out, eventCapacity, executionOutputLimit, needsTerminalJSONContent, shell == programApplicationShell, execution.MaxApplicationEvidenceBytes)
	if needsDefect {
		emitDefectSupport(&out)
	}
	if needsBuffer {
		emitProgramBufferWriter(&out, "SCHWAY_BUFFER")
	}
	if needsByte {
		emitProgramByteWriter(&out)
	}
	if needsU64 {
		emitProgramU64Support(&out)
		if shell == programApplicationShell {
			emitProgramU64ApplicationWriter(&out)
		}
	}
	if err := emitProgramLiveResourcesWriter(&out, liveResources); err != nil {
		return "", err
	}

	// Prototypes before definitions (D-11-03): what makes a forward-
	// referenced callee legal C17.
	for index := range functions {
		fmt.Fprintf(&out, "static %s %s(%s, unsigned int);\n", returnTypeNames[index], functionNames[index], parameterTypeNames[index])
	}
	out.WriteString("\n")

	for index, function := range functions {
		var err error
		if function.Match != nil {
			err = emitProgramBranchFunction(&out, function, branchTypes[function.Parameter.Type], branchTypes[function.ReturnType], functionNames[index], globalNames, lookup, childTableNames)
		} else if isLocalFileByteFunction(function) {
			err = emitProgramLocalFileByteFunction(&out, function, parameterTypeNames[index], returnTypeNames[index], functionNames[index], globalNames)
		} else {
			err = emitProgramFunction(&out, function, parameterTypeNames[index], returnTypeNames[index], functionNames[index], globalNames, lookup, childTableNames)
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
	entryBranch, entryIsBranch := branchTypes[entry.Parameter.Type]
	entryIndexType := parameterTypeNames[entryIndex]
	entryOutputType := returnTypeNames[entryIndex]
	if entryReturnIsBranch && entryReturnBranch.hasPayload {
		emitProgramTerminalValueWriter(&out, entryReturnBranch, program.DataTypes)
	}
	if entryIsBranch {
		entryTypeName = entryIndexType
	} else if err != nil {
		return "", err
	}
	out.WriteString("int main(int argc, char **argv) {\n")
	// Task 2 consumes this literal table while recording /2 events. It is
	// already emitted in Task 1 so all internal calls share one stable index;
	// retain an explicit harmless reference until the event writer reads it.
	out.WriteString("  (void)schway_invocations;\n")
	// A nullary match can produce no operation events at all; retain a harmless
	// reference so the shared schema-2 support remains valid under -Werror.
	out.WriteString("  (void)schway_record_event;\n")
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
	// A directional ABI can declare a nominal input representation that is
	// intentionally distinct from the entry result. The result name writer is
	// used below; retain harmless references to every other generated name
	// writer so strict C builds do not reject the checked input-only types.
	for _, typeName := range branchTypeOrder {
		branchType := branchTypes[typeName]
		if !entryReturnIsBranch || branchType.nameFunction != entryReturnBranch.nameFunction {
			fmt.Fprintf(&out, "  (void)%s;\n", branchType.nameFunction)
		}
	}
	// schway_write_buffer_hex/schway_write_byte (emitProgramBufferWriter/
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
		out.WriteString("  (void)schway_write_buffer_hex;\n")
	}
	if needsByte {
		out.WriteString("  (void)schway_write_byte;\n")
	}
	if needsU64 {
		out.WriteString("  (void)schway_parse_u64_decimal;\n  (void)schway_write_u64;\n")
	}
	if shell == programApplicationShell {
		// The application keeps semantic event recording in the shared body
		// lowering, but does not serialize those events onto application stdout.
		// Taking the evidence writers' addresses keeps this shared support
		// translation unit valid under the strict -Wunused-function build.
		out.WriteString("  (void)schway_write_events;\n  (void)schway_write_live_resources;\n")
	}
	out.WriteString("  if (argc != 2) return 64;\n")
	if shell == programApplicationShell {
		if entry.Parameter.Type == "PathToken" {
			out.WriteString("  if (argv[1][0] == '\\0' || strlen(argv[1]) > 4096u) return 65;\n")
			out.WriteString("  const char *schway_entry_input = argv[1];\n")
		} else {
			out.WriteString("  if (strlen(argv[1]) > 4096u) return 65;\n")
			out.WriteString("  uint64_t schway_entry_input;\n")
			out.WriteString("  if (!schway_parse_u64_decimal(argv[1], &schway_entry_input)) return 65;\n")
		}
		fmt.Fprintf(&out, "  %s schway_entry_output = %s(schway_entry_input, 0u);\n", entryOutputType, functionNames[entryIndex])
		out.WriteString("  if (!schway_write_u64_plain(schway_entry_output) || !schway_write_literal(\"\\n\")) return 74;\n")
		out.WriteString("  const char *schway_evidence_path = getenv(\"SCHWAY_APP_EVIDENCE_PATH\");\n")
		out.WriteString("  if (schway_evidence_path != NULL && schway_evidence_path[0] != '\\0') {\n")
		out.WriteString("    FILE *schway_evidence_file = fopen(schway_evidence_path, \"wb\");\n")
		out.WriteString("    if (schway_evidence_file != NULL) {\n      schway_output_stream = schway_evidence_file;\n      schway_output_limit = SCHWAY_EVIDENCE_OUTPUT_LIMIT;\n      schway_output_count = 0u;\n")
		out.WriteString("      if (schway_event_overflow) {\n        (void)schway_write_literal(\"{\\\"schema\\\":\\\"schway.app-capture/1\\\",\\\"status\\\":\\\"capacity_exhausted\\\"}\\n\");\n      } else {\n")
		out.WriteString("        (void)schway_write_literal(\"{\\\"schema\\\":\\\"schway.app-capture/1\\\",\\\"status\\\":\\\"complete\\\",\\\"execution\\\":{\\\"schema\\\":\\\"schway.execution/2\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\");\n")
		out.WriteString("        (void)schway_write_u64(schway_entry_output);\n")
		out.WriteString("        (void)schway_write_literal(\"},\\\"events\\\":[\");\n        (void)schway_write_events();\n")
		out.WriteString("        (void)schway_write_literal(\"],\\\"live_resources\\\":\");\n        (void)schway_write_live_resources();\n")
		out.WriteString("        (void)schway_write_literal(\"}}\\n\");\n      }\n      (void)fclose(schway_evidence_file);\n    }\n  }\n")
		out.WriteString("  return 0;\n}\n")
		return out.String(), nil
	}
	if entryIsBranch {
		fmt.Fprintf(&out, "  %s schway_entry_input;\n", entryTypeName)
		for index, alternative := range entryBranch.alternatives {
			prefix := "if"
			if index > 0 {
				prefix = "else if"
			}
			if entryBranch.hasPayload {
				if field, ok := entryBranch.fields[alternative.source]; ok && field.payloadType != "" {
					fmt.Fprintf(&out, "  %s (strcmp(argv[1], %s) == 0) { schway_entry_input.tag = %s; schway_entry_input.%s = %s; }\n", prefix, strconv.Quote(alternative.source), alternative.cName, field.name, payloadCannedInitializer(field.payloadType))
				} else {
					fmt.Fprintf(&out, "  %s (strcmp(argv[1], %s) == 0) { schway_entry_input.tag = %s; }\n", prefix, strconv.Quote(alternative.source), alternative.cName)
				}
			} else {
				fmt.Fprintf(&out, "  %s (strcmp(argv[1], %s) == 0) schway_entry_input = %s;\n", prefix, strconv.Quote(alternative.source), alternative.cName)
			}
		}
		out.WriteString("  else return 65;\n")
	} else if entryTypeName == "uint64_t" {
		out.WriteString("  uint64_t schway_entry_input;\n")
		out.WriteString("  if (!schway_parse_u64_decimal(argv[1], &schway_entry_input)) return 65;\n")
	} else {
		fmt.Fprintf(&out, "  if (strcmp(argv[1], %s) != 0) return 65;\n", strconv.Quote(input))
		fmt.Fprintf(&out, "  %s schway_entry_input = %s;\n", entryTypeName, initializer)
	}
	fmt.Fprintf(&out, "  %s schway_entry_output = %s(schway_entry_input, 0u);\n", entryOutputType, functionNames[entryIndex])
	if entryReturnIsBranch {
		fmt.Fprintf(&out, "  const char *schway_entry_name = %s(schway_entry_output);\n", entryReturnBranch.nameFunction)
		out.WriteString("  if (schway_entry_name == NULL) return 70;\n")
		out.WriteString("  if (!schway_write_literal(\"{\\\"schema\\\":\\\"schway.execution/2\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\")) return 74;\n")
		if entryReturnBranch.hasPayload {
			out.WriteString("  if (!schway_write_terminal_value(schway_entry_output)) return 74;\n")
		} else {
			out.WriteString("  if (!schway_write_json_string(schway_entry_name)) return 74;\n")
		}
	} else if entryOutputType == "SCHWAY_BUFFER" {
		out.WriteString("  if (!schway_write_literal(\"{\\\"schema\\\":\\\"schway.execution/2\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\\\"\")) return 74;\n")
		out.WriteString("  if (!schway_write_buffer_hex(&schway_entry_output)) return 74;\n")
	} else if entryOutputType == "uint64_t" {
		out.WriteString("  if (!schway_write_literal(\"{\\\"schema\\\":\\\"schway.execution/2\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\")) return 74;\n")
		out.WriteString("  if (!schway_write_u64(schway_entry_output)) return 74;\n")
	} else {
		out.WriteString("  if (!schway_write_literal(\"{\\\"schema\\\":\\\"schway.execution/2\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",\\\"value\\\":\\\"\")) return 74;\n")
		out.WriteString("  if (!schway_write_byte(schway_entry_output)) return 74;\n")
	}
	if entryIsBranch || entryOutputType == "uint64_t" {
		out.WriteString("  if (!schway_write_literal(\"},\\\"events\\\":[\")) return 74;\n")
	} else {
		out.WriteString("  if (!schway_write_literal(\"\\\"},\\\"events\\\":[\")) return 74;\n")
	}
	out.WriteString("  if (!schway_write_events()) return 74;\n")
	out.WriteString("  if (!schway_write_literal(\"],\\\"live_resources\\\":\")) return 74;\n")
	out.WriteString("  if (!schway_write_live_resources()) return 74;\n")
	out.WriteString("  if (!schway_write_literal(\"}\\n\")) return 74;\n")
	out.WriteString("  return 0;\n}\n")

	return out.String(), nil
}

// emitProgramBufferWriter/emitProgramByteWriter are the multi-function
// assembler's own copies of emitLinearOutputSupport's scalar-value writer
// selection (cgen.go): a Buffer-parameterized program needs
// schway_write_buffer_hex, a Byte-parameterized one needs schway_write_byte.
// Kept here (not shared with emitLinearOutputSupport) because that
// function also unconditionally calls emitEventSupport, which emitProgram
// has already called once for the whole program (D-11-01: calling the
// single-function helper a second time would duplicate every SCHWAY_EVENT
// declaration it contains).
func emitProgramBufferWriter(out *strings.Builder, typeName string) {
	fmt.Fprintf(out, "static int schway_write_buffer_hex(const %s *value) {\n", typeName)
	out.WriteString("  static const char hex[] = \"0123456789abcdef\";\n  size_t index;\n")
	out.WriteString("  if (value->length > sizeof value->bytes) return 0;\n")
	out.WriteString("  for (index = 0u; index < value->length; index++) {\n")
	out.WriteString("    char encoded[2] = {hex[value->bytes[index] >> 4u], hex[value->bytes[index] & 0x0fu]};\n")
	out.WriteString("    if (!schway_write_bytes(encoded, sizeof encoded)) return 0;\n  }\n  return 1;\n}\n\n")
}

func isLocalFileByteFunction(function core.Function) bool {
	if function.Name != "main" || function.Parameter.Type != "PathToken" || function.ReturnType != "U64" || function.Linear == nil || len(function.Linear.Blocks) != 0 || len(function.Linear.Operations) != 4 {
		return false
	}
	operations := function.Linear.Operations
	acquire, use, release, returned := operations[0], operations[1], operations[2], operations[3]
	return acquire.Kind == core.OpForeignCall && acquire.Foreign != nil && acquire.Foreign.Mode == "acquire" && acquire.Foreign.Symbol == "schway_file_byte_acquire" &&
		use.Kind == core.OpForeignCall && use.Foreign != nil && use.Foreign.Mode == "borrow" && use.Foreign.Symbol == "schway_file_byte_use" && use.SourceID == acquire.TargetID &&
		release.Kind == core.OpRelease && release.Foreign != nil && release.Foreign.Mode == "consume" && release.Foreign.Symbol == acquire.Foreign.Release && release.SourceID == acquire.TargetID && release.ReleasesOperationID == acquire.ID &&
		returned.Kind == core.OpReturn && returned.SourceID == use.TargetID && returned.TypeID != "" && acquire.Foreign.Allocator == "libc_malloc" && release.Foreign.Allocator == acquire.Foreign.Allocator
}

// hasLocalOwnerFacts keeps malformed local-owner candidates on the admission
// path even when a mutation removes an operation or contract from the complete
// success shape. PathToken and FileByteOwner facts identify this deliberately
// narrow feature; an unrelated foreign operation or release is not evidence
// that a function uses the Phase 23 local-owner ABI.
func hasLocalOwnerFacts(function core.Function) bool {
	if function.Parameter.Type == "PathToken" || function.ReturnType == "FileByteOwner" {
		return true
	}
	if function.Linear == nil {
		return false
	}
	for _, typeFact := range function.Linear.Types {
		if typeFact.Shape.Constructor == "PathToken" || typeFact.Shape.Constructor == "FileByteOwner" {
			return true
		}
	}
	return false
}

// validateLocalFileByteFunction rechecks the narrow source/emission contract
// from candidate core immediately before any C serialization. The function
// shape alone is not enough: each emitted symbol, ABI typedef, operand and
// result type, status type, allocator, destructor pairing, and exit policy is
// pinned independently at the sole serializer boundary.
func validateLocalFileByteFunction(function core.Function) error {
	if !isLocalFileByteFunction(function) || function.Linear == nil || len(function.Linear.Edges) != 0 {
		return fmt.Errorf("function %q: unsupported local-owner C shape", function.ID)
	}
	operations := function.Linear.Operations
	acquire, use, release, returned := operations[0], operations[1], operations[2], operations[3]
	wantAcquire := core.ForeignOperationContract{
		Symbol: "schway_file_byte_acquire", ABIType: "schway_file_byte_acquire_fn", Mode: "acquire",
		ParameterType: "PathToken", ResultType: "FileByteOwner", Fails: "AcquireError",
		Allocator: "libc_malloc", Release: "schway_file_byte_release", Unwind: "forbidden", NonlocalExit: "forbidden",
	}
	wantUse := core.ForeignOperationContract{
		Symbol: "schway_file_byte_use", ABIType: "schway_file_byte_use_fn", Mode: "borrow",
		ParameterType: "FileByteOwner", ResultType: "U64", Fails: "UseError",
		Unwind: "forbidden", NonlocalExit: "forbidden",
	}
	wantRelease := core.ForeignOperationContract{
		Symbol: "schway_file_byte_release", ABIType: "schway_file_byte_release_fn", Mode: "consume",
		ParameterType: "FileByteOwner", ResultType: "Unit", Allocator: "libc_malloc",
		Unwind: "forbidden", NonlocalExit: "forbidden",
	}
	if acquire.Foreign == nil || *acquire.Foreign != wantAcquire {
		return fmt.Errorf("function %q: acquire operation contract does not match the selected FileByteOwner ABI", function.ID)
	}
	if use.Foreign == nil || *use.Foreign != wantUse {
		return fmt.Errorf("function %q: borrowed-use operation contract does not match the selected FileByteOwner ABI", function.ID)
	}
	if release.Foreign == nil || *release.Foreign != wantRelease {
		return fmt.Errorf("function %q: consuming-release operation contract does not match the selected FileByteOwner ABI", function.ID)
	}
	if acquire.ID == "" || use.ID == "" || release.ID == "" || returned.ID == "" ||
		acquire.SourceID != function.Parameter.ID || acquire.TargetID == "" ||
		use.SourceID != acquire.TargetID || use.TargetID == "" ||
		release.SourceID != acquire.TargetID || release.ReleasesOperationID != acquire.ID ||
		returned.SourceID != use.TargetID || acquire.TypeID != release.TypeID ||
		use.TypeID != returned.TypeID {
		return fmt.Errorf("function %q: local-owner operation places and discharge facts do not agree", function.ID)
	}
	if len(function.Linear.Types) != 3 || len(function.Linear.Places) != 3 {
		return fmt.Errorf("function %q: local-owner facts must contain exactly the path, owner, and byte-result places", function.ID)
	}
	types := make(map[string]core.TypeFact, len(function.Linear.Types))
	for _, typeFact := range function.Linear.Types {
		if typeFact.ID == "" {
			return fmt.Errorf("function %q: local-owner type fact has no identity", function.ID)
		}
		if _, duplicate := types[typeFact.ID]; duplicate {
			return fmt.Errorf("function %q: local-owner type fact identity is duplicated", function.ID)
		}
		types[typeFact.ID] = typeFact
	}
	places := make(map[string]core.Place, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		if place.ID == "" || place.TypeID == "" {
			return fmt.Errorf("function %q: local-owner place is missing its identity or type", function.ID)
		}
		if _, duplicate := places[place.ID]; duplicate {
			return fmt.Errorf("function %q: local-owner place identity is duplicated", function.ID)
		}
		places[place.ID] = place
	}
	parameter, parameterOK := places[function.Parameter.ID]
	owner, ownerOK := places[acquire.TargetID]
	value, valueOK := places[use.TargetID]
	if !parameterOK || !ownerOK || !valueOK || parameter.TypeID == owner.TypeID || owner.TypeID == value.TypeID || parameter.TypeID == value.TypeID ||
		types[parameter.TypeID].Shape.Constructor != "PathToken" ||
		types[owner.TypeID].Shape.Constructor != "FileByteOwner" ||
		types[value.TypeID].Shape.Constructor != "U64" ||
		acquire.TypeID != owner.TypeID || use.TypeID != value.TypeID || release.TypeID != owner.TypeID || returned.TypeID != value.TypeID {
		return fmt.Errorf("function %q: local-owner path, owner, and result type facts do not agree", function.ID)
	}
	return nil
}

func emitProgramLocalFileByteFunction(out *strings.Builder, function core.Function, parameterTypeName, returnTypeName, functionName string, globalNames []string) error {
	if !isLocalFileByteFunction(function) {
		return fmt.Errorf("function %q: unsupported local-owner C shape", function.ID)
	}
	operations := function.Linear.Operations
	acquire, use, release := operations[0], operations[1], operations[2]
	for _, contract := range []*core.ForeignOperationContract{acquire.Foreign, use.Foreign, release.Foreign} {
		if !validForeignSymbol(contract.Symbol) || !validForeignSymbol(contract.ABIType) || contract.Unwind != "forbidden" || contract.NonlocalExit != "forbidden" {
			return fmt.Errorf("function %q: local-owner ABI contract is not a closed C identifier shape", function.ID)
		}
	}
	places := make(map[string]core.Place, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	parameter, parameterOK := places[function.Parameter.ID]
	owner, ownerOK := places[acquire.TargetID]
	value, valueOK := places[use.TargetID]
	releaseOwner, releaseOwnerOK := places[release.SourceID]
	returned := operations[3]
	if !parameterOK || !ownerOK || !valueOK || !releaseOwnerOK || owner.TypeID != release.TypeID || releaseOwner.ID != owner.ID || value.TypeID != returned.TypeID {
		return fmt.Errorf("function %q: local-owner operation places do not agree", function.ID)
	}
	reserved := append(append([]string(nil), linearFixedNames...), globalNames...)
	names := newCNames(reserved...)
	parameterName := names.allocate(cLocal(parameter.Name), "place", 0)
	ownerName := names.allocate(cLocal(owner.Name), "place", 1)
	valueName := names.allocate(cLocal(value.Name), "place", 2)
	acquireResultName := names.allocate("schway_local_acquire_result", "ffi", 0)
	useResultName := names.allocate("schway_local_use_result", "ffi", 1)
	fmt.Fprintf(out, "static %s %s(%s %s, unsigned int invocation_index) {\n", returnTypeName, functionName, parameterTypeName, parameterName)
	fmt.Fprintf(out, "  schway_file_byte_acquire_result %s = %s(%s);\n", acquireResultName, acquire.Foreign.Symbol, parameterName)
	fmt.Fprintf(out, "  if (%s.status != 0) exit(65);\n", acquireResultName)
	fmt.Fprintf(out, "  schway_file_byte_owner %s = %s.owner;\n", ownerName, acquireResultName)
	fmt.Fprintf(out, "  if (%s.data == NULL || %s.length != UINT64_C(1)) { %s(%s); exit(65); }\n", ownerName, ownerName, release.Foreign.Symbol, ownerName)
	fmt.Fprintf(out, "  schway_file_byte_use_result %s = %s(%s);\n", useResultName, use.Foreign.Symbol, ownerName)
	fmt.Fprintf(out, "  %s(%s);\n", release.Foreign.Symbol, ownerName)
	fmt.Fprintf(out, "  if (%s.status != 0) { fputs(\"schway_file_byte_use: UnsupportedByte\\n\", stderr); exit(65); }\n", useResultName)
	fmt.Fprintf(out, "  uint64_t %s = %s.value;\n", valueName, useResultName)
	fmt.Fprintf(out, "  if (!schway_record_event(%s, %s, %s, %s, NULL, %s, schway_invocations[invocation_index], NULL)) abort();\n",
		strconv.Quote("function.returned"), strconv.Quote(returned.ID+":event:returned"), strconv.Quote(function.ID),
		strconv.Quote(returned.SourceID), strconv.Quote(returned.TypeID))
	fmt.Fprintf(out, "  return %s;\n", valueName)
	out.WriteString("}\n\n")
	return nil
}

func emitProgramByteWriter(out *strings.Builder) {
	out.WriteString("static int schway_write_byte(unsigned char value) {\n")
	out.WriteString("  char encoded[3];\n  int length = snprintf(encoded, sizeof encoded, \"%u\", (unsigned int)value);\n")
	out.WriteString("  return length > 0 && (size_t)length < sizeof encoded && schway_write_bytes(encoded, (size_t)length);\n}\n\n")
}

// emitProgramU64Support parses decimal command-line inputs without relying on
// host-sized integer conversions and writes the canonical U64 decimal value
// as a bounded JSON string through the shared output-limit writer.
func emitProgramU64Support(out *strings.Builder) {
	out.WriteString("static int schway_parse_u64_decimal(const char *text, uint64_t *out) {\n")
	out.WriteString("  uint64_t value = UINT64_C(0);\n  const unsigned char *cursor = (const unsigned char *)text;\n")
	out.WriteString("  if (cursor == NULL || *cursor == 0u) return 0;\n")
	out.WriteString("  for (; *cursor != 0u; ++cursor) {\n    unsigned int digit;\n    if (*cursor < (unsigned char)'0' || *cursor > (unsigned char)'9') return 0;\n    digit = (unsigned int)(*cursor - (unsigned char)'0');\n    if (value > (UINT64_MAX - digit) / UINT64_C(10)) return 0;\n    value = value * UINT64_C(10) + digit;\n  }\n  *out = value;\n  return 1;\n}\n\n")
	out.WriteString("static int schway_write_u64(uint64_t value) {\n  char digits[20];\n  size_t length = 0u;\n  size_t index;\n  do { digits[length++] = (char)('0' + (value % UINT64_C(10))); value /= UINT64_C(10); } while (value != 0u && length < sizeof digits);\n  if (value != 0u || !schway_write_bytes(\"\\\"\", 1u)) return 0;\n  for (index = length; index > 0u; --index) if (!schway_write_bytes(&digits[index - 1u], 1u)) return 0;\n  return schway_write_bytes(\"\\\"\", 1u);\n}\n\n")
}

// emitProgramU64ApplicationWriter emits the plain decimal result path used by
// retained applications. The conformance shell keeps schway_write_u64's JSON
// string encoding unchanged.
func emitProgramU64ApplicationWriter(out *strings.Builder) {
	out.WriteString("static int schway_write_u64_plain(uint64_t value) {\n  char digits[20];\n  size_t length = 0u;\n  do { digits[length++] = (char)('0' + (value % UINT64_C(10))); value /= UINT64_C(10); } while (value != 0u && length < sizeof digits);\n  if (value != 0u) return 0;\n  while (length > 0u) { --length; if (!schway_write_bytes(&digits[length], 1u)) return 0; }\n  return 1;\n}\n\n")
}

// emitProgramTerminalValueWriter makes a returned payload-bearing ADT value
// observable at the schema-2 terminal boundary. The tag and payload are read
// from the returned C value, so a wrong-slot construction cannot hide behind
// a compile-time alternative name.
func emitProgramTerminalValueWriter(out *strings.Builder, branchType programBranchType, dataTypes []core.DataType) {
	fmt.Fprintf(out, "static int schway_write_terminal_value(%s value) {\n", branchType.typeName)
	fmt.Fprintf(out, "  const char *tag_name = %s(value);\n  if (tag_name == NULL) return 0;\n", branchType.nameFunction)
	needsBufferScratch, needsByteScratch := false, false
	for _, alternative := range branchType.alternatives {
		field, ok := branchType.fields[alternative.source]
		if !ok {
			continue
		}
		encoding := terminalPayloadEncodingFor(field, dataTypes)
		needsBufferScratch = needsBufferScratch || encoding == terminalPayloadBuffer
		needsByteScratch = needsByteScratch || encoding == terminalPayloadByte
	}
	if needsBufferScratch {
		fmt.Fprintf(out, "  char payload_hex[%d];\n", maxTerminalBufferPayloadBytes*2)
	}
	if needsByteScratch {
		fmt.Fprintf(out, "  char payload_decimal[%d];\n  int payload_length;\n", maxTerminalBytePayloadDigits+1)
	}
	out.WriteString("  if (!schway_write_bytes(\"\\\"\", 1u) || !schway_write_json_string_content(tag_name)) return 0;\n  switch (value.tag) {\n")
	for _, alternative := range branchType.alternatives {
		field, hasPayload := branchType.fields[alternative.source]
		fmt.Fprintf(out, "    case %s:\n", alternative.cName)
		if !hasPayload {
			out.WriteString("      break;\n")
			continue
		}
		encoding := terminalPayloadEncodingFor(field, dataTypes)
		if nestedType, nested := findProgramDataTypeByName(dataTypes, field.payloadType); nested && encoding == terminalPayloadNestedTag {
			fmt.Fprintf(out, "      switch (value.%s) {\n", field.name)
			for index, nestedAlternative := range nestedType.Alternatives {
				fmt.Fprintf(out, "        case %du: if (!schway_write_bytes(\":\", 1u) || !schway_write_json_string_content(%s)) return 0; break;\n", index, strconv.Quote(nestedAlternative))
			}
			out.WriteString("        default: return 0;\n      }\n      break;\n")
			continue
		}
		switch encoding {
		case terminalPayloadBuffer:
			fmt.Fprintf(out, "      if (value.%s.length > %du) return 0;\n      static const char hex[] = \"0123456789abcdef\";\n      for (size_t i = 0u; i < value.%s.length; ++i) {\n        payload_hex[i * 2u] = hex[value.%s.bytes[i] >> 4u];\n        payload_hex[i * 2u + 1u] = hex[value.%s.bytes[i] & 15u];\n      }\n      if (!schway_write_bytes(\":\", 1u) || !schway_write_bytes(payload_hex, value.%s.length * 2u)) return 0;\n      break;\n", field.name, maxTerminalBufferPayloadBytes, field.name, field.name, field.name, field.name)
		case terminalPayloadByte:
			fmt.Fprintf(out, "      payload_length = snprintf(payload_decimal, sizeof payload_decimal, \"%%u\", (unsigned int)value.%s);\n      if (payload_length < 1 || (size_t)payload_length >= sizeof payload_decimal || !schway_write_bytes(\":\", 1u) || !schway_write_bytes(payload_decimal, (size_t)payload_length)) return 0;\n      break;\n", field.name)
		default:
			// Nested data payload serialization is outside this phase's
			// admitted witness; preserve existing tag semantics for it.
			out.WriteString("      break;\n")
		}
	}
	out.WriteString("    default: return 0;\n  }\n  return schway_write_bytes(\"\\\"\", 1u);\n}\n\n")
}

func isProgramNestedBytePayload(field programPayloadField, dataTypes []core.DataType) bool {
	if field.cType != "unsigned char" {
		return false
	}
	_, ok := findProgramDataTypeByName(dataTypes, field.payloadType)
	return ok
}

func findProgramDataTypeByName(dataTypes []core.DataType, name string) (core.DataType, bool) {
	for _, candidate := range dataTypes {
		if candidate.Name == name {
			return candidate, true
		}
	}
	return core.DataType{}, false
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
	fmt.Fprintf(out, "static int schway_write_live_resources(void) {\n  return schway_write_literal(%s);\n}\n\n", strconv.Quote(string(encoded)))
	return nil
}

type programBranchAlternative struct {
	source, cName string
}

type programPayloadField struct {
	name, cType, payloadType string
}

// programBranchType carries the checked nominal branch layout into the one
// whole-program writer. Phase 16 admits only nullary alternatives here; the
// payload record lowering remains behind its existing unsupported boundary.
type programBranchType struct {
	typeName, nameFunction string
	alternatives           []programBranchAlternative
	bySource               map[string]string
	fields                 map[string]programPayloadField
	hasPayload             bool
	dataType               core.DataType
}

func newProgramBranchType(typeName, representationType string, dataTypes []core.DataType, names *cNames) (programBranchType, error) {
	var dataType core.DataType
	for _, candidate := range dataTypes {
		if candidate.Name == representationType {
			dataType = candidate
			break
		}
	}
	if dataType.Name == "" || len(dataType.Alternatives) == 0 {
		return programBranchType{}, fmt.Errorf("branch C emitter cannot find alternatives for representation %q", representationType)
	}
	result := programBranchType{
		typeName:     names.allocate(cName(typeName), "type", 0),
		bySource:     make(map[string]string, len(dataType.Alternatives)),
		alternatives: make([]programBranchAlternative, 0, len(dataType.Alternatives)),
		fields:       make(map[string]programPayloadField, len(dataType.Alternatives)),
		dataType:     dataType,
	}
	for _, detail := range dataType.AlternativeDetails {
		if detail.PayloadType != "" {
			result.hasPayload = true
		}
	}
	if result.hasPayload {
		layout := check.PayloadRecordLayout(dataType)
		if len(layout.Fields) != len(dataType.Alternatives)+1 || layout.Fields[0].Name != "tag" || layout.Fields[0].CType != "unsigned char" {
			return programBranchType{}, fmt.Errorf("branch payload layout for data type %q is invalid", dataType.Name)
		}
		for index, alternative := range dataType.Alternatives {
			detail := core.LookupAlternativeDetail(dataType, alternative)
			field := layout.Fields[index+1]
			result.fields[alternative] = programPayloadField{name: field.Name, cType: field.CType, payloadType: detail.PayloadType}
		}
	}
	for index, alternative := range dataType.Alternatives {
		cAlternative := names.allocate(result.typeName+"_"+cName(alternative), "alternative", index)
		result.alternatives = append(result.alternatives, programBranchAlternative{source: alternative, cName: cAlternative})
		result.bySource[alternative] = cAlternative
	}
	result.nameFunction = names.allocate(result.typeName+"_name", "type_name", 0)
	return result, nil
}

func (t programBranchType) hasCType(cType string) bool {
	for _, field := range t.fields {
		if field.cType == cType {
			return true
		}
	}
	return false
}

func emitProgramBranchType(out *strings.Builder, branchType programBranchType) {
	if branchType.hasPayload {
		fmt.Fprintf(out, "typedef struct %s {\n  unsigned char tag;\n", branchType.typeName)
		for _, alternative := range branchType.alternatives {
			field := branchType.fields[alternative.source]
			fmt.Fprintf(out, "  %s %s;\n", field.cType, field.name)
		}
		fmt.Fprintf(out, "} %s;\n\n", branchType.typeName)
		for index, alternative := range branchType.alternatives {
			fmt.Fprintf(out, "#define %s %d\n", alternative.cName, index)
		}
		out.WriteString("\n")
	} else {
		fmt.Fprintf(out, "typedef enum %s {\n", branchType.typeName)
		for index, alternative := range branchType.alternatives {
			fmt.Fprintf(out, "  %s = %d,\n", alternative.cName, index)
		}
		fmt.Fprintf(out, "} %s;\n\n", branchType.typeName)
	}
	switchValue := "value"
	if branchType.hasPayload {
		switchValue = "value.tag"
	}
	fmt.Fprintf(out, "static const char *%s(%s value) {\n  switch (%s) {\n", branchType.nameFunction, branchType.typeName, switchValue)
	for _, alternative := range branchType.alternatives {
		fmt.Fprintf(out, "    case %s: return %s;\n", alternative.cName, strconv.Quote(alternative.source))
	}
	out.WriteString("  }\n  return NULL;\n}\n\n")
}

// emitProgramBranchFunction lowers a match/branch body into the same shared
// schema-2 event buffer and bare-return ABI as ordinary emitProgram functions.
// It intentionally never writes an execution document: main owns that once.
func emitProgramBranchFunction(out *strings.Builder, function core.Function, parameterBranchType, returnBranchType programBranchType, functionName string, globalNames []string, lookup *emitCallLookup, childTableNames map[string]string) error {
	if function.Match == nil {
		return fmt.Errorf("function %q: branch writer requires a match body", function.ID)
	}
	if function.Linear == nil {
		fmt.Fprintf(out, "static %s %s(%s value, unsigned int invocation_index) {\n  switch (value) {\n", returnBranchType.typeName, functionName, parameterBranchType.typeName)
		for _, arm := range function.Match.Arms {
			pattern, known := parameterBranchType.bySource[arm.Pattern]
			valueName, valueKnown := returnBranchType.bySource[arm.Value]
			if !known || !valueKnown {
				return fmt.Errorf("function %q: match arm %q names unknown alternative", function.ID, arm.ID)
			}
			fmt.Fprintf(out, "    case %s:\n      if (!schway_record_event(\"function.returned\", %s, %s, %s, NULL, %s, schway_invocations[invocation_index], NULL)) abort();\n      return %s;\n", pattern, strconv.Quote(function.ID+":match:return"), strconv.Quote(function.ID), strconv.Quote(function.Parameter.ID), strconv.Quote(function.Parameter.Type), valueName)
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

	fmt.Fprintf(out, "static %s %s(%s %s, unsigned int invocation_index) {\n", returnBranchType.typeName, functionName, parameterBranchType.typeName, locals[parameter.ID])
	declaredPrefix := map[string]bool{parameter.ID: true}
	entryBlock, hasEntry := blocks[function.ID+":block:entry"]
	if !hasEntry {
		return fmt.Errorf("function %q: branch entry block is absent", function.ID)
	}
	for _, operationID := range entryBlock.OperationIDs {
		operation, known := operations[operationID]
		if !known {
			return fmt.Errorf("entry block references unknown operation %q", operationID)
		}
		if err := emitProgramBranchOperation(out, function, operation, parameterBranchType, returnBranchType, locals, places, declaredPrefix, lookup, childTableNames); err != nil {
			return err
		}
	}
	scrutineeID := function.Match.ScrutineeID
	if scrutineeID == "" {
		scrutineeID = parameter.ID
	}
	if function.Match.ScrutineeID == "" && function.Match.Scrutinee != parameter.Name {
		for _, place := range function.Linear.Places {
			if place.Name == function.Match.Scrutinee {
				scrutineeID = place.ID
				break
			}
		}
	}
	switchPlace, known := places[scrutineeID]
	if !known || switchPlace.Name != function.Match.Scrutinee {
		return fmt.Errorf("function %q: match scrutinee place %q is absent", function.ID, function.Match.Scrutinee)
	}
	scrutineeBranchType := parameterBranchType
	for _, fact := range function.Linear.Types {
		if fact.ID == switchPlace.TypeID && fact.Shape.Constructor == function.ReturnType {
			scrutineeBranchType = returnBranchType
			break
		}
	}
	switchValue := locals[scrutineeID]
	if scrutineeBranchType.hasPayload {
		switchValue += ".tag"
	}
	fmt.Fprintf(out, "  switch (%s) {\n", switchValue)
	for _, arm := range function.Match.Arms {
		block, known := blocks[arm.BlockID]
		if !known {
			return fmt.Errorf("arm %q references unknown block %q", arm.ID, arm.BlockID)
		}
		pattern, known := scrutineeBranchType.bySource[arm.Pattern]
		if !known {
			return fmt.Errorf("arm %q names unknown alternative %q", arm.ID, arm.Pattern)
		}
		fmt.Fprintf(out, "    case %s: {\n", pattern)
		declared := make(map[string]bool, len(declaredPrefix))
		for placeID, isDeclared := range declaredPrefix {
			declared[placeID] = isDeclared
		}
		if arm.ValuePlaceID != "" {
			if _, exists := places[arm.ValuePlaceID]; !exists {
				return fmt.Errorf("arm %q references unknown value place %q", arm.ID, arm.ValuePlaceID)
			}
			if declared[arm.ValuePlaceID] {
				return fmt.Errorf("arm %q value place %q was already declared", arm.ID, arm.ValuePlaceID)
			}
			valueName, exists := returnBranchType.bySource[arm.Value]
			if !exists {
				return fmt.Errorf("arm %q names unknown return alternative %q", arm.ID, arm.Value)
			}
			fmt.Fprintf(out, "      %s %s = %s; /* selected arm value */\n", returnBranchType.typeName, locals[arm.ValuePlaceID], valueName)
			declared[arm.ValuePlaceID] = true
		}
		for _, operationID := range block.OperationIDs {
			operation, known := operations[operationID]
			if !known {
				return fmt.Errorf("block references unknown operation %q", operationID)
			}
			if err := emitProgramBranchOperation(out, function, operation, parameterBranchType, returnBranchType, locals, places, declared, lookup, childTableNames); err != nil {
				return err
			}
		}
		out.WriteString("    }\n")
	}
	out.WriteString("  }\n  abort();\n}\n\n")
	return nil
}

func emitProgramBranchOperation(out *strings.Builder, function core.Function, operation core.LinearOperation, parameterBranchType, returnBranchType programBranchType, locals map[string]string, places map[string]core.Place, declared map[string]bool, lookup *emitCallLookup, childTableNames map[string]string) error {
	if operation.Kind != core.OpConst {
		if _, known := places[operation.SourceID]; !known {
			return fmt.Errorf("operation %q has invalid source", operation.ID)
		}
	}
	switch operation.Kind {
	case core.OpConst:
		target, exists := places[operation.TargetID]
		if !exists || declared[operation.TargetID] {
			return fmt.Errorf("operation %q has invalid target", operation.ID)
		}
		fmt.Fprintf(out, "      uint64_t %s = UINT64_C(%s); /* constant: %s */\n      (void)%s;\n", locals[target.ID], operation.ConstU64, operation.ID, locals[target.ID])
		declared[operation.TargetID] = true
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
		valueType := programBranchTypeForTypeID(function, operation.TypeID, parameterBranchType, returnBranchType)
		fmt.Fprintf(out, "      %s %s = %s; /* %s: %s */\n      (void)%s;\n", valueType.typeName, locals[operation.TargetID], locals[operation.SourceID], label, operation.ID, locals[operation.TargetID])
		fmt.Fprintf(out, "      if (!schway_record_event(%s, %s, %s, %s, %s, %s, schway_invocations[invocation_index], NULL)) abort();\n", strconv.Quote(eventKind), strconv.Quote(operation.ID+":event"), strconv.Quote(function.ID), strconv.Quote(operation.SourceID), strconv.Quote(operation.TargetID), strconv.Quote(operation.TypeID))
		declared[operation.TargetID] = true
	case core.OpReturn:
		returnValue := locals[operation.SourceID]
		fmt.Fprintf(out, "      if (!schway_record_event(%s, %s, %s, %s, NULL, %s, schway_invocations[invocation_index], NULL)) abort();\n      return %s;\n", strconv.Quote("function.returned"), strconv.Quote(operation.ID+":event:returned"), strconv.Quote(function.ID), strconv.Quote(operation.SourceID), strconv.Quote(operation.TypeID), returnValue)
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
		fmt.Fprintf(out, "      if (!schway_record_event(%s, %s, %s, %s, %s, %s, schway_invocations[invocation_index], %s)) abort();\n", strconv.Quote("function.called"), strconv.Quote(operation.ID+":event:called"), strconv.Quote(function.ID), strconv.Quote(operation.SourceID), strconv.Quote(operation.TargetID), strconv.Quote(operation.TypeID), strconv.Quote(operation.CalleeID))
		emitProgramCall(out, calleeTypeName, locals[operation.TargetID], calleeName, locals[operation.SourceID], childTable+"[invocation_index]", operation)
		declared[operation.TargetID] = true
	case core.OpDestructurePayload:
		target, exists := places[operation.PayloadTargetID]
		if !exists || declared[operation.PayloadTargetID] {
			return fmt.Errorf("operation %q has invalid payload target", operation.ID)
		}
		alternative, err := branchPayloadAlternative(parameterBranchType, operation)
		if err != nil {
			return err
		}
		field, ok := parameterBranchType.fields[alternative]
		if !ok {
			return fmt.Errorf("operation %q: no checker-derived field for alternative %q", operation.ID, alternative)
		}
		fmt.Fprintf(out, "      %s %s = %s.%s; /* payload destructure: %s */\n      (void)%s;\n", field.cType, locals[target.ID], locals[operation.SourceID], field.name, operation.ID, locals[target.ID])
		fmt.Fprintf(out, "      if (!schway_record_event(%s, %s, %s, %s, %s, %s, schway_invocations[invocation_index], NULL)) abort();\n", strconv.Quote("value.payload_destructured"), strconv.Quote(operation.ID+":event"), strconv.Quote(function.ID), strconv.Quote(operation.SourceID), strconv.Quote(operation.PayloadTargetID), strconv.Quote(operation.TypeID))
		declared[operation.PayloadTargetID] = true
	case core.OpConstructPayload:
		target, exists := places[operation.TargetID]
		if !exists || declared[operation.TargetID] {
			return fmt.Errorf("operation %q has invalid target", operation.ID)
		}
		alternative, err := branchPayloadAlternative(parameterBranchType, operation)
		if err != nil {
			return err
		}
		field, ok := parameterBranchType.fields[alternative]
		if !ok {
			return fmt.Errorf("operation %q: no checker-derived field for alternative %q", operation.ID, alternative)
		}
		tag, ok := parameterBranchType.bySource[alternative]
		if !ok {
			return fmt.Errorf("operation %q: unknown alternative %q", operation.ID, alternative)
		}
		fmt.Fprintf(out, "      %s %s = {0}; /* payload construct: %s */\n      %s.tag = %s;\n", parameterBranchType.typeName, locals[target.ID], operation.ID, locals[target.ID], tag)
		if payloadSlotSwapForTest {
			// Keep D-12-38's test-only mutation seam on the whole-program
			// emitter path.  Leaving it only in emitBranchOperations made the
			// quality control silently inject nothing after Phase 16 routed
			// admitted branch programs through emitProgram.
			wrongField, wrongType, swapped := programWrongPayloadSlot(parameterBranchType, alternative)
			if swapped {
				payloadSlotSwapInjectedWriteCount++
				sourceType := payloadCTypeName(operation.PayloadType)
				switch {
				case wrongType == sourceType:
					fmt.Fprintf(out, "      %s.%s = %s; /* D-12-38 mutation: wrong-slot write */\n", locals[target.ID], wrongField, locals[operation.SourceID])
				case wrongType == "SCHWAY_BUFFER":
					fmt.Fprintf(out, "      %s.%s = (SCHWAY_BUFFER){{%s}, 1u}; /* D-12-38 mutation: wrong-slot write, widened */\n", locals[target.ID], wrongField, locals[operation.SourceID])
				default:
					fmt.Fprintf(out, "      %s.%s = %s.bytes[0]; /* D-12-38 mutation: wrong-slot write, truncated */\n", locals[target.ID], wrongField, locals[operation.SourceID])
				}
			} else {
				fmt.Fprintf(out, "      %s.%s = %s;\n", locals[target.ID], field.name, locals[operation.SourceID])
			}
		} else {
			fmt.Fprintf(out, "      %s.%s = %s;\n", locals[target.ID], field.name, locals[operation.SourceID])
		}
		fmt.Fprintf(out, "      (void)%s;\n", locals[target.ID])
		fmt.Fprintf(out, "      if (!schway_record_event(%s, %s, %s, %s, %s, %s, schway_invocations[invocation_index], NULL)) abort();\n", strconv.Quote("value.payload_constructed"), strconv.Quote(operation.ID+":event"), strconv.Quote(function.ID), strconv.Quote(operation.SourceID), strconv.Quote(operation.TargetID), strconv.Quote(operation.TypeID))
		declared[operation.TargetID] = true
	case core.OpDefect:
		emitProgramDefectTerminal(out, function, operation)
		fmt.Fprintf(out, "      schway_defect(%s);\n", strconv.Quote(operation.Reason))
	case core.OpForeignCall, core.OpFail:
		return fmt.Errorf("operation %q: unsupported branch operation kind %q in whole-program native emission", operation.ID, operation.Kind)
	default:
		return fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
	}
	return nil
}

func programBranchTypeForTypeID(function core.Function, typeID string, parameterType, returnType programBranchType) programBranchType {
	for _, fact := range function.Linear.Types {
		if fact.ID == typeID && fact.Shape.Constructor == function.ReturnType {
			return returnType
		}
	}
	return parameterType
}

func branchPayloadAlternative(branchType programBranchType, operation core.LinearOperation) (string, error) {
	alternative, err := core.AlternativeNameForPayloadType(branchType.dataType, operation.PayloadType)
	if err != nil {
		return "", fmt.Errorf("operation %q: %w", operation.ID, err)
	}
	return alternative, nil
}

// programWrongPayloadSlot finds a different payload-bearing alternative for
// D-12-38's test-only wrong-slot mutation.  It deliberately derives the
// target from the same checker-derived branch type that emission uses.
func programWrongPayloadSlot(branchType programBranchType, alternative string) (field, cType string, ok bool) {
	for _, detail := range branchType.dataType.AlternativeDetails {
		if detail.Name == alternative || detail.PayloadType == "" {
			continue
		}
		if candidate, exists := branchType.fields[detail.Name]; exists {
			return candidate.name, candidate.cType, true
		}
	}
	return "", "", false
}

// emitProgramDefectTerminal writes the schema-2 document before the only
// generated _Noreturn helper aborts. The common buffered event representation
// intentionally has no Output member, so the defect's reason-bearing event is
// appended directly instead of broadening every ordinary event's ABI.
func emitProgramDefectTerminal(out *strings.Builder, function core.Function, operation core.LinearOperation) {
	out.WriteString("      if (!schway_write_literal(\"{\\\"schema\\\":\\\"schway.execution/2\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"defect\\\",\\\"value\\\":\\\"\\\"},\\\"events\\\":[\")) abort();\n")
	out.WriteString("      if (!schway_write_events()) abort();\n")
	out.WriteString("      if (schway_event_count != 0u && !schway_write_bytes(\",\", 1u)) abort();\n")
	out.WriteString("      if (!schway_write_literal(\"{\\\"schema\\\":\\\"schway.execution/2\\\",\\\"id\\\":\")) abort();\n")
	fmt.Fprintf(out, "      if (!schway_write_json_string(%s)) abort();\n", strconv.Quote(operation.ID+":event:defected"))
	out.WriteString("      if (!schway_write_literal(\",\\\"kind\\\":\\\"function.defected\\\",\\\"function_id\\\":\")) abort();\n")
	fmt.Fprintf(out, "      if (!schway_write_json_string(%s)) abort();\n", strconv.Quote(function.ID))
	out.WriteString("      if (!schway_write_literal(\",\\\"output\\\":\")) abort();\n")
	fmt.Fprintf(out, "      if (!schway_write_json_string(%s)) abort();\n", strconv.Quote(operation.Reason))
	out.WriteString("      if (!schway_write_literal(\",\\\"source_place\\\":\")) abort();\n")
	fmt.Fprintf(out, "      if (!schway_write_json_string(%s)) abort();\n", strconv.Quote(operation.SourceID))
	out.WriteString("      if (!schway_write_literal(\",\\\"type_id\\\":\")) abort();\n")
	fmt.Fprintf(out, "      if (!schway_write_json_string(%s)) abort();\n", strconv.Quote(operation.TypeID))
	out.WriteString("      if (!schway_write_literal(\",\\\"invocation\\\":\")) abort();\n")
	out.WriteString("      if (!schway_write_json_string(schway_invocations[invocation_index])) abort();\n")
	out.WriteString("      if (!schway_write_literal(\"}],\\\"live_resources\\\":\")) abort();\n")
	out.WriteString("      if (!schway_write_live_resources()) abort();\n")
	out.WriteString("      if (!schway_write_literal(\"}\\n\")) abort();\n")
}

// emitProgramFunction writes ONE function's own C definition: its own
// fresh, per-function cNames (D-11-08: seeded with the reserved list plus
// every globally allocated name) locals allocator, its straight-line
// operations, and a bare C `return` -- never the whole-TU
// schway.execution/1 JSON tail emitLinear's own single-function path writes,
// since that tail belongs to `main` alone in the multi-function assembler
// (a non-entry callee's own return value is consumed by its caller's own
// OpCall, never written to stdout directly; the entry function's own
// return value is written to stdout exactly once, by `main`, after it
// returns).
func emitProgramFunction(out *strings.Builder, function core.Function, parameterTypeName, returnTypeName, functionName string, globalNames []string, lookup *emitCallLookup, childTableNames map[string]string) error {
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

	fmt.Fprintf(out, "static %s %s(%s %s, unsigned int invocation_index) {\n", returnTypeName, functionName, parameterTypeName, locals[parameter.ID])
	// Leaf functions do not consult a child table, but every internal
	// function still receives the occurrence index. Keep generated C clean
	// under -Werror while preserving the uniform ABI.
	out.WriteString("  (void)invocation_index;\n")
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpConst {
			fmt.Fprintf(out, "  (void)%s;\n", locals[parameter.ID])
			break
		}
	}
	declared := map[string]bool{parameter.ID: true}
	returned := false
	for _, operation := range function.Linear.Operations {
		var source core.Place
		if operation.Kind != core.OpConst {
			var sourceKnown bool
			source, sourceKnown = places[operation.SourceID]
			if !sourceKnown {
				return fmt.Errorf("operation %q has invalid source", operation.ID)
			}
		}
		switch operation.Kind {
		case core.OpConst:
			target, exists := places[operation.TargetID]
			if !exists || declared[operation.TargetID] {
				return fmt.Errorf("operation %q has invalid target", operation.ID)
			}
			fmt.Fprintf(out, "  uint64_t %s = UINT64_C(%s); /* constant: %s */\n  (void)%s;\n", locals[target.ID], operation.ConstU64, operation.ID, locals[target.ID])
			declared[target.ID] = true
		case core.OpCopy, core.OpMove, core.OpBorrowShared, core.OpBorrowExclusive:
			target, exists := places[operation.TargetID]
			if !exists || declared[operation.TargetID] {
				return fmt.Errorf("operation %q has invalid target", operation.ID)
			}
			label := "copy"
			marker := ""
			if operation.Kind == core.OpMove {
				label = "authority transfer"
				marker = " /* schway:mutation-site */"
			} else if operation.Kind == core.OpBorrowShared {
				label = "shared borrow representation"
			} else if operation.Kind == core.OpBorrowExclusive {
				label = "exclusive borrow representation"
			}
			fmt.Fprintf(out, "  %s %s = %s; /* %s: %s */%s\n", parameterTypeName, locals[target.ID], locals[source.ID], label, operation.ID, marker)
			fmt.Fprintf(out, "  (void)%s;\n", locals[target.ID])
			eventKind := "value.copied"
			if operation.Kind == core.OpMove {
				eventKind = "value.transferred"
			} else if operation.Kind == core.OpBorrowShared {
				eventKind = "value.borrowed"
			} else if operation.Kind == core.OpBorrowExclusive {
				eventKind = "value.borrowed_exclusive"
			}
			fmt.Fprintf(out, "  if (!schway_record_event(%s, %s, %s, %s, %s, %s, schway_invocations[invocation_index], NULL)) abort();\n",
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
			fmt.Fprintf(out, "  if (!schway_record_event(%s, %s, %s, %s, %s, %s, schway_invocations[invocation_index], %s)) abort();\n",
				strconv.Quote("function.called"), strconv.Quote(operation.ID+":event:called"), strconv.Quote(function.ID),
				strconv.Quote(operation.SourceID), strconv.Quote(operation.TargetID), strconv.Quote(operation.TypeID), strconv.Quote(operation.CalleeID))
			emitProgramCall(out, calleeTypeName, locals[target.ID], calleeName, locals[source.ID], childTable+"[invocation_index]", operation)
			declared[operation.TargetID] = true
		case core.OpReturn:
			// Unlike emitLinear's single-function OpReturn arm, this never
			// writes the schway.execution/1 JSON tail: that belongs to
			// `main` alone (emitProgram), written exactly once after the
			// resolved entry function returns. A callee's own event is
			// still recorded here, attributed to ITS OWN function ID
			// (interp.terminalOutcome's identical rule, D-10-32), so the
			// caller's own events and the callee's own function.returned
			// event both land in the shared events buffer in execution
			// order regardless of call depth.
			fmt.Fprintf(out, "  if (!schway_record_event(%s, %s, %s, %s, NULL, %s, schway_invocations[invocation_index], NULL)) abort(); /* returned place: %s */\n",
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
