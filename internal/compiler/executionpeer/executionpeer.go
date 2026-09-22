// Package executionpeer independently validates lang.execution/2 causal
// documents. It deliberately knows only core facts and the public invocation
// grammar; it does not consult an execution producer.
package executionpeer

import (
	"fmt"
	"sort"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/execution"
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

type index struct {
	functions    map[string]core.Function
	operations   map[string]operation
	matchReturns map[string]string
	entry        string
	occurrences  map[string]occurrence
}

// Validate checks only observed causal structure. It intentionally does not
// require every statically admissible invocation to appear.
func Validate(program core.Program, document execution.Execution) error {
	if document.Schema != execution.Schema2 {
		return refusal("schema", -1, "", "", "", "expected lang.execution/2")
	}
	indexed, err := buildIndex(program)
	if err != nil {
		return err
	}
	return indexed.validate(document)
}

// ValidateFullCoverage is the intentionally separate straight-line fixture
// control. Unlike Validate, it requires every unfolded occurrence to occur.
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
	i := &index{functions: make(map[string]core.Function), operations: make(map[string]operation), matchReturns: make(map[string]string), occurrences: make(map[string]occurrence)}
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
				if _, ok := i.functions[op.CalleeID]; !ok {
					return nil, refusal("program", -1, "", "", function.ID, "call names unknown callee "+op.CalleeID)
				}
			}
		}
	}
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
	roots := make([]string, 0)
	for id, degree := range inDegree {
		if degree == 0 {
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
	parsed, err := execution.ParseInvocation(parent)
	if err != nil {
		return "", err
	}
	parsed.Segments = append(parsed.Segments, execution.InvocationSegment{OpCallID: callID, Ordinal: 0})
	return execution.FormatInvocation(parsed.EntryID, parsed.Segments)
}

func (i *index) validate(document execution.Execution) error {
	if len(document.Events) == 0 {
		return refusal("missing_event", -1, "", "", "", "document has no events")
	}
	seenPairs := make(map[string]bool)
	stack := make([]string, 0)
	pending := ""
	for n, event := range document.Events {
		parsed, parseErr := execution.ParseInvocation(event.Invocation)
		if parseErr != nil {
			return refusal("malformed_invocation", n, event.Invocation, "", "", parseErr.Error())
		}
		occ, found := i.occurrences[event.Invocation]
		if !found {
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
		pair := event.Invocation + "\x00" + event.ID
		if seenPairs[pair] {
			return refusal("duplicate_pair", n, event.Invocation, "", occ.functionID, "duplicate invocation and event ID")
		}
		seenPairs[pair] = true
		if len(stack) == 0 {
			stack = append(stack, event.Invocation)
			if event.Invocation != rootInvocation(i.entry) {
				return refusal("orphan_child", n, event.Invocation, rootInvocation(i.entry), occ.functionID, "first activation is not entry")
			}
		} else if stack[len(stack)-1] != event.Invocation {
			if pending == event.Invocation {
				stack = append(stack, event.Invocation)
				pending = ""
			} else {
				return refusal("preorder", n, event.Invocation, stack[len(stack)-1], occ.functionID, "activation does not follow its caller-owned edge")
			}
		}
		op, err := i.classify(event)
		if err != nil {
			return refusal("unknown_kind", n, event.Invocation, "", occ.functionID, err.Error())
		}
		if op.functionID != event.FunctionID {
			return refusal("ownership", n, event.Invocation, "", event.FunctionID, "event ID belongs to another function")
		}
		if event.Kind == "function.called" {
			child, childErr := appendInvocation(event.Invocation, op.value.ID)
			if childErr != nil {
				return refusal("malformed_invocation", n, event.Invocation, "", event.FunctionID, childErr.Error())
			}
			childOcc, exists := i.occurrences[child]
			if !exists || childOcc.parent != event.Invocation || childOcc.callID != op.value.ID {
				return refusal("call_edge", n, event.Invocation, "", event.FunctionID, "call is not a declared child occurrence")
			}
			if event.CalleeFunctionID == "" || event.CalleeFunctionID != op.value.CalleeID {
				return refusal("callee", n, event.Invocation, "", op.value.CalleeID, "callee function disagrees with call")
			}
			pending = child
		} else if event.CalleeFunctionID != "" {
			return refusal("callee", n, event.Invocation, "", "", "callee function is reserved for function.called")
		}
		if terminal(event.Kind) {
			if pending != "" {
				return refusal("preorder", n, event.Invocation, "", event.FunctionID, "function returned before observed child activation")
			}
			stack = stack[:len(stack)-1]
		}
	}
	if pending != "" {
		return refusal("orphan_child", len(document.Events), pending, stack[len(stack)-1], i.occurrences[pending].functionID, "caller edge has no child activation")
	}
	if len(stack) != 0 {
		return refusal("preorder", len(document.Events), stack[len(stack)-1], "", i.occurrences[stack[len(stack)-1]].functionID, "activation has no terminal event")
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
		if event.Kind == "function.called" && event.ID == id+":event:called" && op.value.Kind == core.OpCall {
			return op, nil
		}
		if event.ID == id+":event" && normalKind(op.value.Kind) == event.Kind {
			return op, nil
		}
		if event.ID == id+":event:returned" && op.value.Kind == core.OpReturn && event.Kind == "function.returned" {
			return op, nil
		}
		if event.ID == id+":event:failed" && op.value.Kind == core.OpFail && event.Kind == "function.failed" {
			return op, nil
		}
		if event.ID == id+":event:defected" && op.value.Kind == core.OpDefect && event.Kind == "function.defected" {
			return op, nil
		}
	}
	return operation{}, fmt.Errorf("unclassified event kind or ID")
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
