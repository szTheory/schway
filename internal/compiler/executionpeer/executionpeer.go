// Package executionpeer independently validates lang.execution/2 causal
// documents. It deliberately knows only core facts and the public invocation
// grammar; it does not consult an execution producer.
package executionpeer

import (
	"fmt"
	"sort"
	"strconv"
	"unicode/utf8"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
)

// MaxInvocations is the fixed, fail-closed bound on statically unfolded
// activation occurrences. It counts the entry activation.
const MaxInvocations = 4096

// Error is an actionable peer refusal. Index is -1 for failures that happen
// before an event can be consumed (for example entry ambiguity or expansion).
type Error struct {
	Class            string
	Index            int
	Invocation       string
	ExpectedParent   string
	ExpectedFunction string
	Detail           string
}

// Code gives callers a stable, namespaced classification without forcing them
// to parse the diagnostic's contextual text.
func (e *Error) Code() string { return "executionpeer." + e.Class }

func (e *Error) Error() string {
	message := fmt.Sprintf("executionpeer.%s: event index %d", e.Class, e.Index)
	if e.Invocation != "" {
		message += fmt.Sprintf(" invocation %q", e.Invocation)
	}
	if e.ExpectedParent != "" {
		message += fmt.Sprintf(" expected parent %q", e.ExpectedParent)
	}
	if e.ExpectedFunction != "" {
		message += fmt.Sprintf(" expected function %q", e.ExpectedFunction)
	}
	if e.Detail != "" {
		message += ": " + e.Detail
	}
	return message
}

func refusal(class string, index int, invocation, parent, function, detail string) error {
	return &Error{Class: class, Index: index, Invocation: invocation, ExpectedParent: parent, ExpectedFunction: function, Detail: detail}
}

type occurrence struct{ functionID, parent, callID string }
type operation struct {
	functionID string
	value      core.LinearOperation
}

type failureProvenance struct {
	constructor             string
	errorTargetID           string
	errorEdgeID             string
	expectedFailOperationID string
	originOperationID       string
}

type index struct {
	program             core.Program
	functions           map[string]core.Function
	operations          map[string]operation
	matchReturns        map[string]string
	entry               string
	dynamicCallIdentity bool
	occurrences         map[string]occurrence
	loopOperations      map[string]bool
	scalarLoopCopies    map[string]bool
	eventFreeKinds      map[core.OperationKind]bool
}

// Validate checks only observed causal structure. It intentionally does not
// require every statically admissible invocation to appear.
func Validate(program core.Program, document execution.Execution) error {
	if document.Schema != execution.Schema2 {
		return refusal("schema", -1, "", "", "", "expected lang.execution/2")
	}
	if err := preflightDocumentCapacity(document); err != nil {
		return err
	}
	indexed, err := buildIndex(program)
	if err != nil {
		return err
	}
	return indexed.validate(document)
}

func preflightDocumentCapacity(document execution.Execution) error {
	if len(document.Events) > execution.MaxDocumentBytes/128 {
		return refusal("document_capacity", -1, "", "", "", "observed event count exceeds bounded peer capacity")
	}
	used := 0
	add := func(size int) bool {
		if size < 0 || size > execution.MaxDocumentBytes-used {
			return false
		}
		used += size
		return true
	}
	// Count the actual Execution envelope and event array without marshaling
	// or allocating a second copy of untrusted event strings.
	if !add(len(`{"schema":`)+jsonStringSize(document.Schema)) ||
		!add(len(`,"outcome":{"kind":`)+jsonStringSize(document.Outcome.Kind)) ||
		!add(len(`,"value":`)+jsonStringSize(document.Outcome.Value)) || !add(1) ||
		!add(len(`,"events":[`)) {
		return refusal("document_capacity", -1, "", "", "", "execution document exceeds bounded peer capacity")
	}
	for index, event := range document.Events {
		size, ok := eventJSONSize(event)
		if !ok || (index > 0 && !add(1)) || !add(size) {
			return refusal("document_capacity", -1, "", "", "", "execution document exceeds bounded peer capacity")
		}
	}
	if !add(len(`],"live_resources":[`)) {
		return refusal("document_capacity", -1, "", "", "", "execution document exceeds bounded peer capacity")
	}
	for index, resource := range document.LiveResources {
		if (index > 0 && !add(1)) || !add(jsonStringSize(resource)) {
			return refusal("document_capacity", -1, "", "", "", "execution document exceeds bounded peer capacity")
		}
	}
	if !add(2) { // close live_resources array and document object
		return refusal("document_capacity", -1, "", "", "", "execution document exceeds bounded peer capacity")
	}
	return nil
}

func eventJSONSize(event execution.Event) (int, bool) {
	size := 2 // braces
	fields := 0
	add := func(n int) bool {
		if n < 0 || size > execution.MaxDocumentBytes-n-1 {
			return false
		}
		if fields > 0 {
			size++
		}
		size += n
		fields++
		return true
	}
	stringField := func(key, value string, optional bool) bool {
		if optional && value == "" {
			return true
		}
		return add(len(key) + jsonStringSize(value))
	}
	for _, field := range []struct{ key, value string }{
		{`"schema":`, event.Schema}, {`"id":`, event.ID}, {`"kind":`, event.Kind}, {`"function_id":`, event.FunctionID},
	} {
		if !stringField(field.key, field.value, false) {
			return 0, false
		}
	}
	for _, field := range []struct{ key, value string }{
		{`"input":`, event.Input}, {`"output":`, event.Output}, {`"source_place":`, event.SourcePlace},
		{`"target_place":`, event.TargetPlace}, {`"type_id":`, event.TypeID}, {`"invocation":`, event.Invocation},
		{`"callee_function_id":`, event.CalleeFunctionID},
	} {
		if !stringField(field.key, field.value, true) {
			return 0, false
		}
	}
	if event.Occurrence != 0 && !add(len(`"occurrence":`)+len(strconv.FormatUint(event.Occurrence, 10))) {
		return 0, false
	}
	return size, true
}

func jsonStringSize(value string) int {
	size := 2 // quotes
	for index := 0; index < len(value); {
		r, width := utf8.DecodeRuneInString(value[index:])
		if r == utf8.RuneError && width == 1 {
			size += 6 // encoding/json replaces an invalid byte with \\ufffd
			index++
			continue
		}
		switch {
		case r == '"' || r == '\\' || r == '\b' || r == '\f' || r == '\n' || r == '\r' || r == '\t':
			size += 2
		case r < 0x20 || r == '<' || r == '>' || r == '&' || r == '\u2028' || r == '\u2029':
			size += 6
		default:
			size += width
		}
		index += width
	}
	return size
}

// ValidateFullCoverage is the intentionally separate straight-line fixture
// control. Unlike Validate, it requires every unfolded occurrence to occur.
// It is a test-only fixture control with no production caller by design;
// wiring it into admission would be a new admission rule requiring a new
// decision.
func ValidateFullCoverage(program core.Program, document execution.Execution) error {
	if err := Validate(program, document); err != nil {
		return err
	}
	indexed, err := buildIndex(program)
	if err != nil {
		return err
	}
	seen := make(map[string]bool)
	for _, event := range document.Events {
		seen[event.Invocation] = true
	}
	for invocation := range indexed.occurrences {
		if !seen[invocation] {
			return refusal("full_coverage_missing", -1, invocation, "", indexed.occurrences[invocation].functionID, "statically admissible invocation was not observed")
		}
	}
	return nil
}

func buildIndex(program core.Program) (*index, error) {
	i := &index{
		program:   program,
		functions: make(map[string]core.Function), operations: make(map[string]operation),
		matchReturns: make(map[string]string), occurrences: make(map[string]occurrence),
		loopOperations: make(map[string]bool), scalarLoopCopies: make(map[string]bool),
		eventFreeKinds: eventFreeOperationKinds(),
	}
	for _, function := range program.Functions {
		if function.ID == "" {
			return nil, refusal("program", -1, "", "", "", "function ID is missing")
		}
		if _, exists := i.functions[function.ID]; exists {
			return nil, refusal("program", -1, "", "", function.ID, "duplicate function ID")
		}
		i.functions[function.ID] = function
		if function.Match != nil && function.Linear == nil {
			i.matchReturns[function.ID+":match:return"] = function.ID
			for _, arm := range function.Match.Arms {
				i.matchReturns[arm.ID+":event:returned"] = function.ID
			}
		}
	}
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, op := range function.Linear.Operations {
			if op.ID == "" {
				return nil, refusal("program", -1, "", "", function.ID, "operation ID is missing")
			}
			if _, exists := i.operations[op.ID]; exists {
				return nil, refusal("program", -1, "", "", function.ID, "duplicate operation ID")
			}
			i.operations[op.ID] = operation{functionID: function.ID, value: op}
			if op.Kind == core.OpCall {
				callee, ok := i.functions[op.CalleeID]
				if !ok {
					return nil, refusal("program", -1, "", "", function.ID, "call names unknown callee "+op.CalleeID)
				}
				if scalarValueFunction(callee) {
					i.dynamicCallIdentity = true
				}
			}
		}
	}
	i.indexLoopOperations(program)
	entry, err := resolveEntry(i.functions, i.operations)
	if err != nil {
		return nil, err
	}
	i.entry = entry
	root, err := execution.FormatInvocation(entry, nil)
	if err != nil {
		return nil, refusal("invocation", -1, "", "", entry, err.Error())
	}
	i.occurrences[root] = occurrence{functionID: entry}
	queue := []string{root}
	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]
		function := i.functions[i.occurrences[parent].functionID]
		if function.Linear == nil {
			continue
		}
		for _, op := range function.Linear.Operations {
			if op.Kind != core.OpCall {
				continue
			}
			child, err := appendInvocation(parent, op.ID)
			if err != nil {
				return nil, refusal("invocation", -1, parent, "", function.ID, err.Error())
			}
			if len(i.occurrences) >= MaxInvocations {
				return nil, refusal("traversal_exhausted", -1, child, parent, op.CalleeID, fmt.Sprintf("limit %d", MaxInvocations))
			}
			i.occurrences[child] = occurrence{functionID: op.CalleeID, parent: parent, callID: op.ID}
			queue = append(queue, child)
		}
	}
	return i, nil
}

func (i *index) typeConstructor(functionID, typeID string) (string, bool) {
	function, ok := i.functions[functionID]
	if !ok || function.Linear == nil || typeID == "" {
		return "", false
	}
	constructor := ""
	for _, fact := range function.Linear.Types {
		if fact.ID != typeID {
			continue
		}
		if fact.Shape.Constructor == "" || constructor != "" {
			return "", false
		}
		constructor = fact.Shape.Constructor
	}
	return constructor, constructor != ""
}

func (i *index) place(functionID, placeID string) (core.Place, bool) {
	function, ok := i.functions[functionID]
	if !ok || function.Linear == nil || placeID == "" {
		return core.Place{}, false
	}
	var found core.Place
	for _, candidate := range function.Linear.Places {
		if candidate.ID != placeID {
			continue
		}
		if found.ID != "" {
			return core.Place{}, false
		}
		found = candidate
	}
	return found, found.ID != ""
}

// errorRoute independently binds an error-producing operation to the exact
// error edge, target place, and terminal in its destination block. It reads
// only the indexed core; it does not trust checker or interpreter verdicts.
func (i *index) errorRoute(functionID string, operation core.LinearOperation) (core.Place, core.LinearOperation, bool) {
	function, ok := i.functions[functionID]
	if !ok || function.Linear == nil || operation.ErrTargetID == "" || operation.ErrEdgeID == "" {
		return core.Place{}, core.LinearOperation{}, false
	}
	target, ok := i.place(functionID, operation.ErrTargetID)
	if !ok {
		return core.Place{}, core.LinearOperation{}, false
	}
	if _, ok := i.typeConstructor(functionID, target.TypeID); !ok {
		return core.Place{}, core.LinearOperation{}, false
	}
	operationBlockID := ""
	for _, block := range function.Linear.Blocks {
		for _, operationID := range block.OperationIDs {
			if operationID != operation.ID {
				continue
			}
			if operationBlockID != "" {
				return core.Place{}, core.LinearOperation{}, false
			}
			operationBlockID = block.ID
		}
	}
	if operationBlockID == "" {
		return core.Place{}, core.LinearOperation{}, false
	}
	var sourceBlock, targetBlock core.Block
	var edge core.Edge
	var sourceCount, targetCount, edgeCount int
	for _, block := range function.Linear.Blocks {
		if block.ID == operationBlockID {
			sourceBlock = block
			sourceCount++
		}
	}
	for _, candidate := range function.Linear.Edges {
		if candidate.ID == operation.ErrEdgeID {
			edge = candidate
			edgeCount++
		}
	}
	if edgeCount != 1 || sourceCount != 1 || edge.FromBlockID != operationBlockID || edge.Pattern != "err" || !containsString(sourceBlock.Successors, edge.ToBlockID) {
		return core.Place{}, core.LinearOperation{}, false
	}
	for _, block := range function.Linear.Blocks {
		if block.ID == edge.ToBlockID {
			targetBlock = block
			targetCount++
		}
	}
	if targetCount != 1 || len(targetBlock.OperationIDs) == 0 {
		return core.Place{}, core.LinearOperation{}, false
	}
	var fail core.LinearOperation
	failCount := 0
	for _, operationID := range targetBlock.OperationIDs {
		candidate, exists := i.operations[operationID]
		if !exists || candidate.functionID != functionID {
			return core.Place{}, core.LinearOperation{}, false
		}
		if candidate.value.Kind == core.OpFail {
			fail = candidate.value
			failCount++
		}
	}
	if failCount != 1 || targetBlock.OperationIDs[len(targetBlock.OperationIDs)-1] != fail.ID || fail.SourceID != target.ID || fail.TypeID != target.TypeID {
		return core.Place{}, core.LinearOperation{}, false
	}
	return target, fail, true
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func (i *index) declaredFailureTypes(functionID string) map[string]bool {
	failures := make(map[string]bool)
	seenFunctions := make(map[string]bool)
	queue := []string{functionID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if seenFunctions[current] {
			continue
		}
		seenFunctions[current] = true
		function, ok := i.functions[current]
		if !ok || function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			switch operation.Kind {
			case core.OpFail:
				if constructor, known := i.typeConstructor(current, operation.TypeID); known {
					failures[constructor] = true
				}
			case core.OpForeignCall:
				if operation.Foreign != nil && operation.Foreign.Fails != "" {
					failures[operation.Foreign.Fails] = true
				}
			case core.OpCall:
				if operation.ErrTargetID != "" {
					if target, known := i.place(current, operation.ErrTargetID); known {
						if constructor, mapped := i.typeConstructor(current, target.TypeID); mapped {
							failures[constructor] = true
						}
					}
					queue = append(queue, operation.CalleeID)
				}
			}
		}
	}
	return failures
}

func (i *index) errorEnvelopeIncludes(calleeID, envelope, observedFailure string) bool {
	if envelope == "" {
		return false
	}
	var envelopeType *core.DataType
	for index := range i.program.DataTypes {
		candidate := &i.program.DataTypes[index]
		if candidate.Name != envelope {
			continue
		}
		if envelopeType != nil || len(candidate.AlternativeDetails) != 0 || len(candidate.Alternatives) == 0 {
			return false
		}
		envelopeType = candidate
	}
	if envelopeType == nil {
		return false
	}
	envelopeAlternatives := make(map[string]bool, len(envelopeType.Alternatives))
	for _, alternative := range envelopeType.Alternatives {
		if alternative == "" || envelopeAlternatives[alternative] {
			return false
		}
		envelopeAlternatives[alternative] = true
	}
	failureTypes := i.declaredFailureTypes(calleeID)
	if observedFailure != "" && !failureTypes[observedFailure] {
		return false
	}
	for failureType := range failureTypes {
		var failureDataType *core.DataType
		for index := range i.program.DataTypes {
			candidate := &i.program.DataTypes[index]
			if candidate.Name != failureType {
				continue
			}
			if failureDataType != nil || len(candidate.AlternativeDetails) != 0 || len(candidate.Alternatives) == 0 {
				return false
			}
			failureDataType = candidate
		}
		if failureDataType == nil {
			return false
		}
		for _, alternative := range failureDataType.Alternatives {
			if !envelopeAlternatives[alternative] {
				return false
			}
		}
	}
	return true
}

func (i *index) validateForeignEvent(event execution.Event, operation operation) (failureProvenance, error) {
	function := i.functions[operation.functionID]
	contract := operation.value.Foreign
	if function.Linear == nil || contract == nil || event.ID != operation.value.ID+":event" {
		return failureProvenance{}, fmt.Errorf("foreign event is not bound to an indexed foreign operation")
	}
	source, sourceKnown := i.place(operation.functionID, operation.value.SourceID)
	target, targetKnown := i.place(operation.functionID, operation.value.TargetID)
	sourceType, sourceTypeKnown := i.typeConstructor(operation.functionID, source.TypeID)
	targetType, targetTypeKnown := i.typeConstructor(operation.functionID, target.TypeID)
	if !sourceKnown || !targetKnown || !sourceTypeKnown || !targetTypeKnown || sourceType != contract.ParameterType || targetType != contract.ResultType || operation.value.TypeID != target.TypeID {
		return failureProvenance{}, fmt.Errorf("foreign source, target, or contract type disagrees with indexed type facts")
	}
	if event.SourcePlace != operation.value.SourceID || event.TargetPlace != operation.value.TargetID || event.Input != "" || event.Output != "" || event.CalleeFunctionID != "" {
		return failureProvenance{}, fmt.Errorf("foreign event source or target disagrees with its indexed operation")
	}
	if event.Kind == "foreign.called" {
		if event.TypeID != operation.value.TypeID {
			return failureProvenance{}, fmt.Errorf("successful foreign event type disagrees with its checked result place")
		}
		return failureProvenance{}, nil
	}
	if event.Kind != "foreign.failed" || contract.Fails == "" || event.TypeID != contract.Fails {
		return failureProvenance{}, fmt.Errorf("foreign failure is absent from the indexed operation contract")
	}
	errTarget, failOperation, routeKnown := i.errorRoute(operation.functionID, operation.value)
	errConstructor, errTypeKnown := i.typeConstructor(operation.functionID, errTarget.TypeID)
	if !routeKnown || !errTypeKnown || errConstructor != contract.Fails {
		return failureProvenance{}, fmt.Errorf("foreign failure contract is not bound to its checked error target and OpFail route")
	}
	return failureProvenance{
		constructor: contract.Fails, errorTargetID: errTarget.ID, errorEdgeID: operation.value.ErrEdgeID,
		expectedFailOperationID: failOperation.ID, originOperationID: operation.value.ID,
	}, nil
}

func (i *index) validateCallFailureRoute(operation operation) (failureProvenance, error) {
	if operation.value.Kind != core.OpCall || operation.value.ErrTargetID == "" {
		return failureProvenance{}, fmt.Errorf("call has no checked error target")
	}
	target, failOperation, known := i.errorRoute(operation.functionID, operation.value)
	envelope, envelopeKnown := i.typeConstructor(operation.functionID, target.TypeID)
	if !known || !envelopeKnown || !i.errorEnvelopeIncludes(operation.value.CalleeID, envelope, "") {
		return failureProvenance{}, fmt.Errorf("call error edge, target, or failure envelope is not represented by checked core facts")
	}
	return failureProvenance{
		errorTargetID: target.ID, errorEdgeID: operation.value.ErrEdgeID,
		expectedFailOperationID: failOperation.ID,
	}, nil
}

func (i *index) indexLoopOperations(program core.Program) {
	for _, function := range program.Functions {
		if function.Linear == nil || len(function.Linear.Blocks) == 0 || len(function.Linear.Edges) == 0 {
			continue
		}
		adjacency := make(map[string][]string, len(function.Linear.Blocks))
		for _, edge := range function.Linear.Edges {
			adjacency[edge.FromBlockID] = append(adjacency[edge.FromBlockID], edge.ToBlockID)
		}
		cyclic := make(map[string]bool)
		for _, block := range function.Linear.Blocks {
			for _, successor := range adjacency[block.ID] {
				if reachableBlock(adjacency, successor, block.ID) {
					cyclic[block.ID] = true
					break
				}
			}
		}
		places := make(map[string]core.Place, len(function.Linear.Places))
		for _, place := range function.Linear.Places {
			places[place.ID] = place
		}
		types := make(map[string]string, len(function.Linear.Types))
		for _, fact := range function.Linear.Types {
			types[fact.ID] = fact.Shape.Constructor
		}
		operations := make(map[string]core.LinearOperation, len(function.Linear.Operations))
		for _, op := range function.Linear.Operations {
			operations[op.ID] = op
		}
		for _, block := range function.Linear.Blocks {
			if !cyclic[block.ID] {
				continue
			}
			for _, operationID := range block.OperationIDs {
				op, known := operations[operationID]
				if !known {
					continue
				}
				i.loopOperations[operationID] = true
				if scalarCycleCopy(function.Linear, op, places, types) {
					i.scalarLoopCopies[operationID] = true
				}
			}
		}
	}
}

func reachableBlock(adjacency map[string][]string, from, target string) bool {
	queue := []string{from}
	seen := make(map[string]bool)
	for len(queue) > 0 {
		block := queue[0]
		queue = queue[1:]
		if block == target {
			return true
		}
		if seen[block] {
			continue
		}
		seen[block] = true
		queue = append(queue, adjacency[block]...)
	}
	return false
}

func scalarCycleCopy(linear *core.LinearBody, op core.LinearOperation, places map[string]core.Place, types map[string]string) bool {
	if op.Kind != core.OpCopy || op.LoanID != "" {
		return false
	}
	source, sourceExists := places[op.SourceID]
	target, targetExists := places[op.TargetID]
	if !sourceExists || !targetExists || source.TypeID == "" || source.TypeID != target.TypeID || op.TypeID != source.TypeID {
		return false
	}
	shape := types[source.TypeID]
	if shape != "U64" && shape != "Bool" {
		return false
	}
	for _, endpoint := range linear.LoanEndpoints {
		if endpoint.AfterOperationID == op.ID {
			return false
		}
	}
	return true
}

func resolveEntry(functions map[string]core.Function, operations map[string]operation) (string, error) {
	inDegree := make(map[string]int, len(functions))
	adjacency := make(map[string][]string, len(functions))
	for id := range functions {
		inDegree[id] = 0
	}
	for _, op := range operations {
		if op.value.Kind == core.OpCall {
			adjacency[op.functionID] = append(adjacency[op.functionID], op.value.CalleeID)
			inDegree[op.value.CalleeID]++
		}
	}
	// Phase 32.1 (D-15, D-26): independent re-derivation of the Phase 28 D-01
	// entry rule. Only entry-module (Module == "") functions are candidates;
	// exactly one `main` wins, several refuse, otherwise the unique root and
	// the closure tiebreak decide.
	mains := make([]string, 0)
	for id, function := range functions {
		if function.Name == "main" && function.Module == "" {
			mains = append(mains, id)
		}
	}
	sort.Strings(mains)
	if len(mains) == 1 {
		return mains[0], nil
	}
	if len(mains) > 1 {
		return "", refusal("entry_ambiguous", -1, "", "", "", fmt.Sprintf("duplicate main %v", mains))
	}
	roots := make([]string, 0)
	for id, degree := range inDegree {
		if degree == 0 && functions[id].Module == "" {
			roots = append(roots, id)
		}
	}
	sort.Strings(roots)
	if len(roots) == 1 {
		return roots[0], nil
	}
	best, count := -1, 0
	bestID := ""
	for _, root := range roots {
		visited := map[string]bool{root: true}
		queue := []string{root}
		size := 0
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			for _, callee := range adjacency[current] {
				if !visited[callee] {
					visited[callee] = true
					size++
					queue = append(queue, callee)
				}
			}
		}
		if size > best {
			best, count, bestID = size, 1, root
		} else if size == best {
			count++
		}
	}
	if count != 1 {
		return "", refusal("entry_ambiguous", -1, "", "", "", fmt.Sprintf("roots %v", roots))
	}
	return bestID, nil
}

func appendInvocation(parent, callID string) (string, error) {
	return appendInvocationOrdinal(parent, callID, 0)
}

func appendInvocationOrdinal(parent, callID string, ordinal int) (string, error) {
	parsed, err := execution.ParseInvocation(parent)
	if err != nil {
		return "", err
	}
	parsed.Segments = append(parsed.Segments, execution.InvocationSegment{OpCallID: callID, Ordinal: ordinal})
	return execution.FormatInvocation(parsed.EntryID, parsed.Segments)
}

func (i *index) validate(document execution.Execution) error {
	if len(document.Events) == 0 {
		return refusal("missing_event", -1, "", "", "", "document has no events")
	}
	nextOccurrence := make(map[string]uint64)
	exhaustedOccurrence := make(map[string]bool)
	root := rootInvocation(i.entry)
	dynamicOccurrences := make(map[string]occurrence)
	expectedOccurrences := i.occurrences
	if i.dynamicCallIdentity {
		dynamicOccurrences[root] = occurrence{functionID: i.entry}
		expectedOccurrences = dynamicOccurrences
	}
	closedActivations := make(map[string]bool)
	nextOrdinals := make(map[string]uint64)
	observedErrorCalls := make(map[string]bool)
	failureByInvocation := make(map[string]failureProvenance)
	stack := make([]string, 0)
	pending := ""
	checkedAddDefect := -1
	checkedRemainderDefect := -1
	checkedBytesDefect := -1
	for n, event := range document.Events {
		parsed, parseErr := execution.ParseInvocation(event.Invocation)
		if parseErr != nil {
			return refusal("malformed_invocation", n, event.Invocation, "", "", parseErr.Error())
		}
		var occ occurrence
		var found bool
		if i.dynamicCallIdentity {
			occ, found = dynamicOccurrences[event.Invocation]
		} else {
			occ, found = i.occurrences[event.Invocation]
		}
		if !found {
			if i.dynamicCallIdentity {
				return refusal("unknown_invocation", n, event.Invocation, "", "", "invocation is not derived from an observed call edge")
			}
			return refusal("unknown_invocation", n, event.Invocation, "", "", "invocation is outside bounded membership")
		}
		if parsed.EntryID != i.entry {
			return refusal("entry", n, event.Invocation, "", i.entry, "wrong entry")
		}
		if event.Schema != execution.Schema2 {
			return refusal("schema", n, event.Invocation, "", occ.functionID, "event schema is not lang.execution/2")
		}
		if event.ID == "" || event.FunctionID == "" || event.Kind == "" {
			return refusal("missing_field", n, event.Invocation, "", occ.functionID, "event identity field is missing")
		}
		if event.FunctionID != occ.functionID {
			return refusal("function", n, event.Invocation, "", occ.functionID, "event function disagrees with invocation")
		}
		if closedActivations[event.Invocation] {
			pair := event.Invocation + "\x00" + event.ID
			if event.Occurrence < nextOccurrence[pair] {
				return refusal("duplicate_pair", n, event.Invocation, "", occ.functionID, "event pair was already observed in a closed activation")
			}
			return refusal("closed_activation", n, event.Invocation, "", occ.functionID, "event follows this activation's terminal")
		}
		op, err := i.classify(event)
		if err != nil {
			return refusal("unknown_kind", n, event.Invocation, "", occ.functionID, err.Error())
		}
		if i.eventFreeKinds[op.value.Kind] {
			return refusal("pure_operation_event", n, event.Invocation, "", occ.functionID, "pure value operation is event-free")
		}
		checkedOverflow := op.value.Kind == core.OpAddChecked && event.Kind == "function.defected"
		checkedRemainderZero := op.value.Kind == core.OpRemainderU64 && event.Kind == "function.defected"
		checkedBytesAt := op.value.Kind == core.OpBytesAt && event.Kind == "function.defected"
		checkedCall := i.dynamicCallIdentity && op.value.Kind == core.OpCall && event.Kind == "function.called"
		if (i.loopOperations[op.value.ID] || event.Occurrence > 0) && !i.scalarLoopCopies[op.value.ID] && !checkedOverflow && !checkedRemainderZero && !checkedBytesAt && !checkedCall {
			return refusal("loop_event", n, event.Invocation, "", occ.functionID, "cyclic or repeated event is not an admitted scalar copy")
		}
		if checkedOverflow {
			if event.Output != execution.DefectReasonU64AddOverflow {
				return refusal("checked_add_attribution", n, event.Invocation, "", occ.functionID, "checked-add defect reason disagrees with fixed overflow reason")
			}
			if event.SourcePlace != op.value.SourceID {
				return refusal("checked_add_attribution", n, event.Invocation, "", occ.functionID, "checked-add source place disagrees with operation source")
			}
			if event.TypeID != op.value.TypeID {
				return refusal("checked_add_attribution", n, event.Invocation, "", occ.functionID, "checked-add type disagrees with operation type")
			}
			if event.TargetPlace != "" {
				return refusal("checked_add_attribution", n, event.Invocation, "", occ.functionID, "checked-add defect must not name a target place")
			}
			if n != len(document.Events)-1 || document.Outcome.Kind != execution.OutcomeDefect || document.Outcome.Value != "" {
				return refusal("checked_add_terminal", n, event.Invocation, "", occ.functionID, "checked-add overflow must be the final event and carry an empty defect outcome")
			}
			checkedAddDefect = n
		} else if checkedRemainderZero {
			if event.Output != execution.DefectReasonU64RemainderByZero {
				return refusal("remainder_zero_attribution", n, event.Invocation, "", occ.functionID, "remainder defect reason disagrees with fixed zero-divisor reason")
			}
			if event.SourcePlace != op.value.SourceID {
				return refusal("remainder_zero_attribution", n, event.Invocation, "", occ.functionID, "remainder source place disagrees with operation source")
			}
			if event.TypeID != op.value.TypeID {
				return refusal("remainder_zero_attribution", n, event.Invocation, "", occ.functionID, "remainder type disagrees with operation type")
			}
			if event.TargetPlace != "" {
				return refusal("remainder_zero_attribution", n, event.Invocation, "", occ.functionID, "remainder defect must not name a target place")
			}
			if n != len(document.Events)-1 || document.Outcome.Kind != execution.OutcomeDefect || document.Outcome.Value != "" {
				return refusal("remainder_zero_terminal", n, event.Invocation, "", occ.functionID, "remainder by zero must be the final event and carry an empty defect outcome")
			}
			checkedRemainderDefect = n
		} else if checkedBytesAt {
			if event.Output != execution.DefectReasonByteIndexOutOfBounds {
				return refusal("byte_index_attribution", n, event.Invocation, "", occ.functionID, "byte-at defect reason disagrees with fixed bounds reason")
			}
			if event.SourcePlace != op.value.SourceID {
				return refusal("byte_index_attribution", n, event.Invocation, "", occ.functionID, "byte-at source place disagrees with the view place")
			}
			if event.TargetPlace != "" {
				return refusal("byte_index_attribution", n, event.Invocation, "", occ.functionID, "byte-at defect must not name a target place")
			}
			if n != len(document.Events)-1 || document.Outcome.Kind != execution.OutcomeDefect || document.Outcome.Value != "" {
				return refusal("byte_index_terminal", n, event.Invocation, "", occ.functionID, "byte index defect must be the final event and carry an empty defect outcome")
			}
			checkedBytesDefect = n
		} else if checkedAddDefect >= 0 {
			return refusal("checked_add_terminal", n, event.Invocation, "", occ.functionID, "event follows checked-add overflow terminal")
		} else if checkedRemainderDefect >= 0 {
			return refusal("remainder_zero_terminal", n, event.Invocation, "", occ.functionID, "event follows remainder-by-zero terminal")
		} else if checkedBytesDefect >= 0 {
			return refusal("byte_index_terminal", n, event.Invocation, "", occ.functionID, "event follows byte index defect terminal")
		}
		pair := event.Invocation + "\x00" + event.ID
		if event.Occurrence < nextOccurrence[pair] || (exhaustedOccurrence[pair] && event.Occurrence == ^uint64(0)) {
			return refusal("duplicate_pair", n, event.Invocation, "", occ.functionID, "event pair was already observed")
		}
		if exhaustedOccurrence[pair] || event.Occurrence != nextOccurrence[pair] {
			return refusal("occurrence", n, event.Invocation, "", occ.functionID, fmt.Sprintf("expected occurrence %d", nextOccurrence[pair]))
		}
		if event.Occurrence == ^uint64(0) {
			exhaustedOccurrence[pair] = true
		} else {
			nextOccurrence[pair] = event.Occurrence + 1
		}
		if pending != "" && event.Invocation != pending {
			return refusal("preorder", n, event.Invocation, pending, occ.functionID, "caller continued before the called child activation")
		}
		if len(stack) == 0 {
			stack = append(stack, event.Invocation)
			if event.Invocation != root {
				return refusal("orphan_child", n, event.Invocation, root, occ.functionID, "first activation is not entry")
			}
		} else if stack[len(stack)-1] != event.Invocation {
			if pending == event.Invocation {
				stack = append(stack, event.Invocation)
				pending = ""
			} else {
				return refusal("preorder", n, event.Invocation, stack[len(stack)-1], occ.functionID, "activation does not follow its caller-owned edge")
			}
		}
		if op.functionID != event.FunctionID {
			return refusal("ownership", n, event.Invocation, "", event.FunctionID, "event ID belongs to another function")
		}
		if op.value.Kind == core.OpForeignCall && (event.Kind == "foreign.called" || event.Kind == "foreign.failed") {
			provenance, attributionErr := i.validateForeignEvent(event, op)
			if attributionErr != nil {
				class := "foreign_attribution"
				if event.Kind == "foreign.failed" {
					class = "foreign_failure_attribution"
				}
				return refusal(class, n, event.Invocation, "", event.FunctionID, attributionErr.Error())
			}
			if event.Kind == "foreign.failed" {
				if _, alreadyFailed := failureByInvocation[event.Invocation]; alreadyFailed {
					return refusal("foreign_failure_provenance", n, event.Invocation, "", event.FunctionID, "activation already has an unconsumed failure")
				}
				failureByInvocation[event.Invocation] = provenance
			}
		}
		if event.Kind == "function.called" {
			if event.CalleeFunctionID == "" || event.CalleeFunctionID != op.value.CalleeID {
				return refusal("callee", n, event.Invocation, "", op.value.CalleeID, "callee function disagrees with call")
			}
			if op.value.ErrTargetID != "" {
				if event.SourcePlace != op.value.SourceID || event.TargetPlace != op.value.TargetID || event.TypeID != op.value.TypeID || event.Input != "" || event.Output != "" {
					return refusal("call_attribution", n, event.Invocation, "", event.FunctionID, "fallible call places or result type disagree with core, or scalar payload is present")
				}
				if _, routeErr := i.validateCallFailureRoute(op); routeErr != nil {
					return refusal("call_error_route", n, event.Invocation, "", event.FunctionID, routeErr.Error())
				}
				observedErrorCalls[event.Invocation+"\x00"+op.value.ID] = true
			}
			var child string
			if i.dynamicCallIdentity {
				if event.SourcePlace != op.value.SourceID || event.TargetPlace != op.value.TargetID || event.TypeID != op.value.TypeID || event.Input != "" || event.Output != "" {
					return refusal("call_attribution", n, event.Invocation, "", event.FunctionID, "call places or result type disagree with core, or scalar payload is present")
				}
				ordinalKey := event.Invocation + "\x00" + op.value.ID
				ordinal := nextOrdinals[ordinalKey]
				if ordinal > 2147483647 {
					return refusal("ordinal_exhausted", n, event.Invocation, "", event.FunctionID, "dynamic invocation ordinal exceeds the portable capture ceiling")
				}
				var childErr error
				child, childErr = appendInvocationOrdinal(event.Invocation, op.value.ID, int(ordinal))
				if childErr != nil {
					return refusal("malformed_invocation", n, event.Invocation, "", event.FunctionID, childErr.Error())
				}
				if _, exists := dynamicOccurrences[child]; exists {
					return refusal("call_edge", n, event.Invocation, "", event.FunctionID, "call reuses a previously derived child activation")
				}
				dynamicOccurrences[child] = occurrence{functionID: op.value.CalleeID, parent: event.Invocation, callID: op.value.ID}
				nextOrdinals[ordinalKey] = ordinal + 1
			} else {
				var childErr error
				child, childErr = appendInvocation(event.Invocation, op.value.ID)
				if childErr != nil {
					return refusal("malformed_invocation", n, event.Invocation, "", event.FunctionID, childErr.Error())
				}
				childOcc, exists := i.occurrences[child]
				if !exists || childOcc.parent != event.Invocation || childOcc.callID != op.value.ID {
					return refusal("call_edge", n, event.Invocation, "", event.FunctionID, "call is not a declared child occurrence")
				}
			}
			pending = child
		} else if event.CalleeFunctionID != "" {
			return refusal("callee", n, event.Invocation, "", "", "callee function is reserved for function.called")
		}
		if terminal(event.Kind) {
			// These are core attribution facts, not a replay of the returned value.
			switch op.value.Kind {
			case core.OpReturn, core.OpDefect:
				if event.SourcePlace != op.value.SourceID || event.TypeID != op.value.TypeID || event.TargetPlace != "" {
					return refusal("terminal_attribution", n, event.Invocation, "", event.FunctionID, "terminal places or type disagree with core")
				}
				if op.value.Kind == core.OpDefect && event.Output != op.value.Reason {
					return refusal("terminal_attribution", n, event.Invocation, "", event.FunctionID, "user-defect reason disagrees with core")
				}
			case core.OpFail:
				if event.SourcePlace != op.value.SourceID || event.TargetPlace != "" {
					return refusal("terminal_attribution", n, event.Invocation, "", event.FunctionID, "failure terminal source or target disagrees with core")
				}
				source, sourceKnown := i.place(op.functionID, op.value.SourceID)
				constructor, constructorKnown := i.typeConstructor(op.functionID, op.value.TypeID)
				if !sourceKnown || !constructorKnown || source.TypeID != op.value.TypeID {
					return refusal("terminal_attribution", n, event.Invocation, "", event.FunctionID, "failure terminal type fact does not match its source place")
				}
				if provenance, hasProvenance := failureByInvocation[event.Invocation]; hasProvenance {
					if op.value.ID != provenance.expectedFailOperationID || op.value.SourceID != provenance.errorTargetID || event.TypeID != provenance.constructor {
						return refusal("failure_provenance", n, event.Invocation, "", event.FunctionID, "failure terminal is not the OpFail reached through its activation's checked error target")
					}
				} else if event.TypeID != constructor {
					return refusal("terminal_attribution", n, event.Invocation, "", event.FunctionID, "failure event constructor disagrees with the source type fact")
				}
			}
			provenance, hasProvenance := failureByInvocation[event.Invocation]
			if hasProvenance && event.Kind != "function.failed" {
				return refusal("failure_provenance", n, event.Invocation, "", event.FunctionID, "activation with an admitted foreign failure did not reach its indexed OpFail")
			}
			// A child return does not determine its caller's eventual outcome.
			// A defect, however, terminates the whole observed execution.
			if event.Invocation == root || event.Kind == "function.defected" {
				want := execution.OutcomeReturned
				if event.Kind == "function.failed" {
					want = execution.OutcomeTypedFailure
				} else if event.Kind == "function.defected" {
					want = execution.OutcomeDefect
				}
				if document.Outcome.Kind != want || (want == execution.OutcomeDefect && document.Outcome.Value != "") {
					return refusal("terminal_outcome", n, event.Invocation, "", event.FunctionID, "final terminal and execution outcome disagree")
				}
			}
			if pending != "" {
				return refusal("preorder", n, event.Invocation, "", event.FunctionID, "function returned before observed child activation")
			}
			if hasProvenance {
				delete(failureByInvocation, event.Invocation)
				if occ.parent != "" {
					callOperation, exists := i.operations[occ.callID]
					callKey := occ.parent + "\x00" + occ.callID
					parentOccurrence, parentExists := expectedOccurrences[occ.parent]
					if !exists || !parentExists || callOperation.functionID != parentOccurrence.functionID || callOperation.value.Kind != core.OpCall || callOperation.value.CalleeID != occ.functionID || !observedErrorCalls[callKey] {
						return refusal("failure_provenance", n, event.Invocation, occ.parent, event.FunctionID, "failed activation has no observed caller-owned error edge")
					}
					parentProvenance, routeErr := i.validateCallFailureRoute(callOperation)
					if routeErr != nil {
						return refusal("call_error_route", n, event.Invocation, occ.parent, callOperation.functionID, routeErr.Error())
					}
					envelopeTarget, targetKnown := i.place(callOperation.functionID, callOperation.value.ErrTargetID)
					envelope, envelopeKnown := i.typeConstructor(callOperation.functionID, envelopeTarget.TypeID)
					if !targetKnown || !envelopeKnown || !i.errorEnvelopeIncludes(callOperation.value.CalleeID, envelope, provenance.constructor) {
						return refusal("call_error_route", n, event.Invocation, occ.parent, callOperation.functionID, "child failure constructor is outside the caller's checked error envelope")
					}
					if _, alreadyFailed := failureByInvocation[occ.parent]; alreadyFailed {
						return refusal("failure_provenance", n, event.Invocation, occ.parent, callOperation.functionID, "caller already has unconsumed failure provenance")
					}
					parentProvenance.constructor = provenance.constructor
					parentProvenance.originOperationID = provenance.originOperationID
					failureByInvocation[occ.parent] = parentProvenance
				}
			}
			closedActivations[event.Invocation] = true
			if event.Kind == "function.defected" {
				for _, activation := range stack {
					closedActivations[activation] = true
				}
				stack = stack[:0]
			} else {
				stack = stack[:len(stack)-1]
			}
		}
	}
	if pending != "" {
		return refusal("orphan_child", len(document.Events), pending, stack[len(stack)-1], expectedOccurrences[pending].functionID, "caller edge has no child activation")
	}
	if len(stack) != 0 {
		return refusal("preorder", len(document.Events), stack[len(stack)-1], "", expectedOccurrences[stack[len(stack)-1]].functionID, "activation has no terminal event")
	}
	return nil
}

func rootInvocation(entry string) string {
	value, _ := execution.FormatInvocation(entry, nil)
	return value
}
func terminal(kind string) bool {
	return kind == "function.returned" || kind == "function.failed" || kind == "function.defected"
}

func (i *index) classify(event execution.Event) (operation, error) {
	if functionID, ok := i.matchReturns[event.ID]; ok && event.Kind == "function.returned" {
		return operation{functionID: functionID}, nil
	}
	for _, op := range i.operations {
		id := op.value.ID
		if i.eventFreeKinds[op.value.Kind] && (event.ID == id+":event" || event.ID == id+":event:defected") {
			return op, nil
		}
		if event.Kind == "function.called" && event.ID == id+":event:called" && op.value.Kind == core.OpCall {
			return op, nil
		}
		if event.ID == id+":event" && normalKind(op.value.Kind) == event.Kind {
			return op, nil
		}
		if event.ID == id+":event" && event.Kind == "foreign.failed" && op.value.Kind == core.OpForeignCall && op.value.Foreign != nil && op.value.Foreign.Fails != "" {
			return op, nil
		}
		if event.ID == id+":event:returned" && op.value.Kind == core.OpReturn && event.Kind == "function.returned" {
			return op, nil
		}
		if event.ID == id+":event:failed" && op.value.Kind == core.OpFail && event.Kind == "function.failed" {
			return op, nil
		}
		if event.ID == id+":event:defected" && event.Kind == "function.defected" && (op.value.Kind == core.OpDefect || op.value.Kind == core.OpAddChecked || op.value.Kind == core.OpRemainderU64 || op.value.Kind == core.OpBytesAt) {
			return op, nil
		}
	}
	return operation{}, fmt.Errorf("unclassified event kind or ID")
}

func eventFreeOperationKinds() map[core.OperationKind]bool {
	return map[core.OperationKind]bool{
		core.OpBitXorU64:     true,
		core.OpBitAndU64:     true,
		core.OpBitOrU64:      true,
		core.OpShiftLeftU64:  true,
		core.OpShiftRightU64: true,
	}
}

func scalarValueFunction(function core.Function) bool {
	parameter := function.Parameter.Type
	result := function.ReturnType
	return (parameter == "U64" || parameter == "Bool") && (result == "U64" || result == "Bool")
}

func normalKind(kind core.OperationKind) string {
	switch kind {
	case core.OpCopy:
		return "value.copied"
	case core.OpMove:
		return "value.transferred"
	case core.OpBorrowShared:
		return "value.borrowed"
	case core.OpBorrowExclusive:
		return "value.borrowed_exclusive"
	case core.OpForeignCall:
		return "foreign.called"
	case core.OpRelease:
		return "resource.released"
	case core.OpConstructPayload:
		return "value.payload_constructed"
	case core.OpDestructurePayload:
		return "value.payload_destructured"
	}
	return ""
}
