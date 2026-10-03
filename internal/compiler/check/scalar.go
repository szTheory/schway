package check

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/szTheory/schway/internal/compiler/ability"
	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
)

type scalarBuilder struct {
	function    ast.FuncDecl
	functionID  string
	linear      core.LinearBody
	types       map[string]string
	places      map[string]core.Place
	names       map[string]string
	initialized map[string]bool
	blocks      []core.Block
	edges       []core.Edge
	current     int
	work        int
	diagnostics []diagnostic.Diagnostic
}

// scalarAnalysisBudgetForTest overrides the deterministic production budget
// so tests can prove exhaustion without constructing adversarial source.
var scalarAnalysisBudgetForTest int

func checkScalarFunction(functionID string, function ast.FuncDecl) (core.Function, []diagnostic.Diagnostic, int) {
	if function.Parameter.Type.Constructor != "U64" || function.ReturnType.Constructor != "U64" {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.scalar_signature", function.Span, "scalar CFG functions require U64 input and output")}, 1
	}
	b := &scalarBuilder{function: function, functionID: functionID, types: map[string]string{}, places: map[string]core.Place{}, names: map[string]string{}, initialized: map[string]bool{}}
	b.linear = core.LinearBody{ID: functionID + ":linear"}
	for _, name := range []string{"U64", "Bool"} {
		shape := core.TypeRef{Constructor: name}
		derived, err := ability.Derive(shape)
		if err != nil {
			return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.unknown", function.Span, err.Error())}, 1
		}
		id := fmt.Sprintf("%s:type:%d", functionID, len(b.linear.Types))
		b.linear.Types = append(b.linear.Types, core.TypeFact{ID: id, Shape: shape, Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses})
		b.types[name] = id
	}
	parameter := b.addPlace(function.Parameter.Name, "U64", false)
	b.names[function.Parameter.Name] = parameter.ID
	b.initialized[parameter.ID] = true
	b.newBlock("entry")
	b.lower(function.Body.Linear.Statements)
	if len(b.diagnostics) == 0 {
		resultID, ok := b.names[function.Body.Linear.Result]
		if !ok {
			b.diagnostics = append(b.diagnostics, diagnostic.Error("name.unknown_place", function.Body.Linear.Span, "scalar result is not an initialized local"))
		} else if !b.initialized[resultID] {
			b.refuse("name.uninitialized_place", function.Body.Linear.Span, "scalar result is not definitely initialized")
		} else {
			if b.places[resultID].TypeID != b.types["U64"] {
				b.refuse("type.scalar_result", function.Body.Linear.Span, "scalar function result must be U64")
				return core.Function{}, b.diagnostics, b.work
			}
			b.addOperation(core.LinearOperation{Kind: core.OpReturn, SourceID: resultID, TypeID: b.types["U64"]})
		}
	}
	if len(b.diagnostics) != 0 {
		return core.Function{}, b.diagnostics, b.work
	}
	b.linear.Blocks, b.linear.Edges = b.blocks, b.edges
	checked := core.Function{ID: functionID, Name: function.Name, EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return", Parameter: core.Parameter{ID: parameter.ID, Name: function.Parameter.Name, Type: "U64"}, ReturnType: "U64", Linear: &b.linear, Span: function.Span}
	analysisWork, refusal, category := checkScalarFixedPointDetailed(checked)
	b.work += analysisWork
	if refusal != "" {
		if refusal == "check.cfg_back_edge" {
			return core.Function{}, []diagnostic.Diagnostic{scalarBackEdgeDiagnostic(functionID, category, function.Span)}, b.work
		}
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error(refusal, function.Span, "scalar CFG analysis did not converge or found an uninitialized read")}, b.work
	}
	return checked, nil, b.work
}

func scalarBackEdgeDiagnostic(functionID, category string, span diagnostic.Span) diagnostic.Diagnostic {
	return diagnostic.Error(
		"check.cfg_back_edge", span,
		fmt.Sprintf("function %q's cyclic control-flow carries unsupported authority or event semantics", functionID),
		diagnostic.Cause{Kind: category, Detail: "cyclic scalar analysis refused this carried authority or event"},
	)
}

type checkerScalarFact struct {
	initialized bool
	known       bool
	boolType    bool
	u64         uint64
	boolean     bool
}

func checkScalarFixedPoint(function core.Function) (int, string) {
	work, code, _ := checkScalarFixedPointDetailed(function)
	return work, code
}

func checkScalarFixedPointDetailed(function core.Function) (int, string, string) {
	blocks := append([]core.Block(nil), function.Linear.Blocks...)
	sort.Slice(blocks, func(i, j int) bool { return blocks[i].ID < blocks[j].ID })
	ops := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		ops[operation.ID] = operation
	}
	types := make(map[string]string, len(function.Linear.Types))
	for _, fact := range function.Linear.Types {
		types[fact.ID] = fact.Shape.Constructor
	}
	places := make(map[string]core.Place, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	out := map[string]map[string]checkerScalarFact{}
	reachable := map[string]bool{}
	work := 0
	budget := scalarAnalysisBudgetForTest
	if budget <= 0 {
		budget = 65536
	}
	for sweep := 0; ; sweep++ {
		changed := false
		for _, block := range blocks {
			incoming := map[string]checkerScalarFact{}
			hasIncoming := false
			if block.PointID == function.EntryPointID {
				incoming[function.Parameter.ID] = checkerScalarFact{initialized: true}
				hasIncoming = true
			}
			for _, edge := range function.Linear.Edges {
				if edge.ToBlockID != block.ID || !reachable[edge.FromBlockID] {
					continue
				}
				if !hasIncoming {
					incoming = cloneCheckerScalarState(out[edge.FromBlockID])
					hasIncoming = true
					continue
				}
				for placeID := range places {
					a, b := incoming[placeID], out[edge.FromBlockID][placeID]
					merged := checkerScalarFact{initialized: a.initialized && b.initialized, boolType: a.boolType && b.boolType}
					if merged.initialized && a.known && b.known && a.boolType == b.boolType && a.u64 == b.u64 && a.boolean == b.boolean {
						merged.known, merged.u64, merged.boolean = true, a.u64, a.boolean
					}
					incoming[placeID] = merged
				}
			}
			if !hasIncoming {
				continue
			}
			state := cloneCheckerScalarState(incoming)
			for _, id := range block.OperationIDs {
				work++
				if work > budget {
					return work, "check.scalar_analysis_limit", ""
				}
				op := ops[id]
				source := state[op.SourceID]
				switch op.Kind {
				case core.OpConst:
					value, err := strconv.ParseUint(op.ConstU64, 10, 64)
					if err != nil {
						return work, "check.scalar_constant_invalid", ""
					}
					state[op.TargetID] = checkerScalarFact{initialized: true, known: true, u64: value}
				case core.OpAddChecked:
					right := state[op.RightID]
					if !source.initialized || !right.initialized {
						return work, "check.scalar_uninitialized", ""
					}
					fact := checkerScalarFact{initialized: true}
					if source.known && right.known && ^uint64(0)-source.u64 >= right.u64 {
						fact.known, fact.u64 = true, source.u64+right.u64
					}
					state[op.TargetID] = fact
				case core.OpLessU64:
					right := state[op.RightID]
					if !source.initialized || !right.initialized {
						return work, "check.scalar_uninitialized", ""
					}
					fact := checkerScalarFact{initialized: true, boolType: true}
					if source.known && right.known {
						fact.known, fact.boolean = true, source.u64 < right.u64
					}
					state[op.TargetID] = fact
				case core.OpScalarStore:
					if !source.initialized {
						return work, "check.scalar_uninitialized", ""
					}
					state[op.StoreTargetID] = checkerScalarFact{initialized: true, known: source.known, boolType: types[places[op.StoreTargetID].TypeID] == "Bool", u64: source.u64, boolean: source.boolean}
				case core.OpCopy:
					if !source.initialized || (types[places[op.SourceID].TypeID] != "U64" && types[places[op.SourceID].TypeID] != "Bool") || (types[places[op.TargetID].TypeID] != "U64" && types[places[op.TargetID].TypeID] != "Bool") || places[op.SourceID].TypeID != places[op.TargetID].TypeID {
						return work, "check.cfg_back_edge", scalarAuthorityCategory(op, types, places)
					}
					state[op.TargetID] = source
				case core.OpBranch, core.OpReturn, core.OpDefect:
					if !source.initialized {
						return work, "check.scalar_uninitialized", ""
					}
				default:
					return work, "check.cfg_back_edge", scalarAuthorityCategory(op, types, places)
				}
			}
			if !reachable[block.ID] || !equalCheckerScalarState(out[block.ID], state) {
				out[block.ID] = state
				reachable[block.ID] = true
				changed = true
			}
		}
		if !changed {
			return work, "", ""
		}
		if sweep >= budget {
			return work, "check.scalar_analysis_limit", ""
		}
	}
}

func scalarAuthorityCategory(op core.LinearOperation, types map[string]string, places map[string]core.Place) string {
	switch op.Kind {
	case core.OpMove:
		return "owner"
	case core.OpBorrowShared, core.OpBorrowExclusive:
		return "loan"
	case core.OpForeignCall, core.OpRelease:
		return "resource"
	case core.OpCopy:
		constructor := types[places[op.SourceID].TypeID]
		switch constructor {
		case "Resource", "File", "Handle":
			return "resource"
		case "Owner", "Box", "Buffer":
			return "owner"
		case "Loan", "Borrowed", "View":
			return "loan_derived_provenance"
		}
	default:
		return "unsupported_event_kind"
	}
	return "unsupported_event_kind"
}

func cloneCheckerScalarState(state map[string]checkerScalarFact) map[string]checkerScalarFact {
	clone := make(map[string]checkerScalarFact, len(state))
	for id, fact := range state {
		clone[id] = fact
	}
	return clone
}

func equalCheckerScalarState(left, right map[string]checkerScalarFact) bool {
	if len(left) != len(right) {
		return false
	}
	for id, fact := range left {
		if right[id] != fact {
			return false
		}
	}
	return true
}

func (b *scalarBuilder) addPlace(name, typeName string, mutable bool) core.Place {
	id := fmt.Sprintf("%s:place:%d", b.functionID, len(b.linear.Places))
	place := core.Place{ID: id, Name: name, TypeID: b.types[typeName], Mutable: mutable}
	b.linear.Places = append(b.linear.Places, place)
	b.places[id] = place
	return place
}

func (b *scalarBuilder) newBlock(name string) string {
	id := fmt.Sprintf("%s:block:%s:%d", b.functionID, name, len(b.blocks))
	if name == "entry" {
		id = b.functionID + ":block:entry"
	}
	point := b.functionID + ":point:entry"
	if len(b.blocks) != 0 {
		point = b.functionID + ":point:block:" + fmt.Sprint(len(b.blocks))
	}
	b.blocks = append(b.blocks, core.Block{ID: id, PointID: point, OperationIDs: []string{}, Successors: []string{}})
	b.current = len(b.blocks) - 1
	return id
}

func (b *scalarBuilder) edge(from, to, pattern string) string {
	id := fmt.Sprintf("%s:edge:%d", b.functionID, len(b.edges))
	b.edges = append(b.edges, core.Edge{ID: id, FromBlockID: from, ToBlockID: to, Pattern: pattern})
	for i := range b.blocks {
		if b.blocks[i].ID == from {
			b.blocks[i].Successors = append(b.blocks[i].Successors, to)
			break
		}
	}
	return id
}

func (b *scalarBuilder) addOperation(op core.LinearOperation) {
	op.ID = fmt.Sprintf("%s:op:%d", b.functionID, len(b.linear.Operations))
	op.PointID = fmt.Sprintf("%s:point:linear:%d", b.functionID, len(b.linear.Operations))
	if op.TypeID == "" {
		op.TypeID = b.types["U64"]
	}
	b.linear.Operations = append(b.linear.Operations, op)
	b.blocks[b.current].OperationIDs = append(b.blocks[b.current].OperationIDs, op.ID)
	b.work++
}

func (b *scalarBuilder) lower(statements []ast.Statement) {
	for _, statement := range statements {
		b.work++
		switch statement.Kind {
		case "var":
			if _, exists := b.names[statement.Name]; exists {
				b.refuse("name.duplicate_binding", statement.Span, "scalar binding is already declared")
				continue
			}
			if statement.Expr != nil && statement.Expr.Kind == "number" {
				place := b.addPlace(statement.Name, "U64", true)
				b.names[statement.Name] = place.ID
				value, err := strconv.ParseUint(statement.Expr.Value, 0, 64)
				if err != nil {
					value, err = strconv.ParseUint(statement.Expr.Value, 10, 64)
				}
				if err != nil {
					b.refuse("type.invalid_u64", statement.Expr.Span, "invalid U64 literal")
					continue
				}
				b.addOperation(core.LinearOperation{Kind: core.OpConst, TargetID: place.ID, ConstU64: strconv.FormatUint(value, 10), TypeID: b.types["U64"]})
				b.initialized[place.ID] = true
			} else {
				value, typeName := b.expression(statement.Expr)
				if typeName != "U64" && typeName != "Bool" {
					b.refuse("type.scalar_binding", statement.Span, "var accepts only U64 or Bool scalars")
					continue
				}
				place := b.addPlace(statement.Name, typeName, true)
				b.names[statement.Name] = place.ID
				b.addOperation(core.LinearOperation{Kind: core.OpScalarStore, SourceID: value, StoreTargetID: place.ID, TypeID: b.types[typeName]})
				b.initialized[place.ID] = true
			}
		case "assign":
			target, exists := b.names[statement.Name]
			if !exists || !b.places[target].Mutable {
				b.refuse("ownership.scalar_assignment", statement.Span, "assignment requires a mutable local scalar")
				continue
			}
			value, typeName := b.expression(statement.Expr)
			if typeName != "U64" && typeName != "Bool" || b.places[target].TypeID != b.types[typeName] {
				b.refuse("type.scalar_assignment", statement.Span, "scalar assignment types must match")
				continue
			}
			b.addOperation(core.LinearOperation{Kind: core.OpScalarStore, SourceID: value, StoreTargetID: target, TypeID: b.places[target].TypeID})
			b.initialized[target] = true
		case "if":
			condition, conditionType := b.expression(statement.Expr)
			if conditionType != "Bool" {
				b.refuse("type.condition_not_bool", statement.Span, "if condition must be Bool")
				continue
			}
			parentIndex := b.current
			from := b.blocks[parentIndex].ID
			before := cloneInitialized(b.initialized)
			thenID := b.newBlock("if-then")
			elseIndex := len(b.blocks)
			elseID := fmt.Sprintf("%s:block:if-else:%d", b.functionID, elseIndex)
			joinID := fmt.Sprintf("%s:block:if-join:%d", b.functionID, elseIndex+1)
			trueEdge := b.edge(from, thenID, "true")
			falseEdge := b.edge(from, elseID, "false")
			b.current = parentIndex
			b.addOperation(core.LinearOperation{Kind: core.OpBranch, SourceID: condition, TrueEdgeID: trueEdge, FalseEdgeID: falseEdge, TypeID: b.types["Bool"]})
			b.current = parentIndex + 1
			b.initialized = cloneInitialized(before)
			b.lower(statement.Then)
			thenInitialized := cloneInitialized(b.initialized)
			thenEnd := len(b.blocks) - 1
			b.newNamedBlock(elseID)
			b.initialized = cloneInitialized(before)
			b.lower(statement.Else)
			elseInitialized := cloneInitialized(b.initialized)
			elseEnds := len(b.blocks) - 1
			b.blocks = append(b.blocks, core.Block{ID: joinID, PointID: b.functionID + ":point:block:" + fmt.Sprint(len(b.blocks)), OperationIDs: []string{}, Successors: []string{}})
			if !b.terminated(thenEnd) {
				b.edge(b.blocks[thenEnd].ID, joinID, "next")
			}
			if !b.terminated(elseEnds) {
				b.edge(b.blocks[elseEnds].ID, joinID, "next")
			}
			b.current = len(b.blocks) - 1
			b.initialized = make(map[string]bool)
			for id := range thenInitialized {
				if elseInitialized[id] {
					b.initialized[id] = true
				}
			}
		case "while":
			before := cloneInitialized(b.initialized)
			preheader := b.blocks[b.current].ID
			condID := b.newBlock("while-condition")
			b.edge(preheader, condID, "next")
			b.current = len(b.blocks) - 1
			condition, conditionType := b.expression(statement.Expr)
			if conditionType != "Bool" {
				b.refuse("type.condition_not_bool", statement.Span, "while condition must be Bool")
				continue
			}
			condID = b.blocks[b.current].ID
			bodyID := fmt.Sprintf("%s:block:while-body:%d", b.functionID, len(b.blocks))
			exitID := fmt.Sprintf("%s:block:while-exit:%d", b.functionID, len(b.blocks)+1)
			trueEdge := b.edge(condID, bodyID, "true")
			falseEdge := b.edge(condID, exitID, "false")
			b.addOperation(core.LinearOperation{Kind: core.OpBranch, SourceID: condition, TrueEdgeID: trueEdge, FalseEdgeID: falseEdge, TypeID: b.types["Bool"]})
			b.newBlock("while-body")
			b.initialized = cloneInitialized(before)
			b.lower(statement.Then)
			bodyEnd := len(b.blocks) - 1
			b.blocks = append(b.blocks, core.Block{ID: exitID, PointID: b.functionID + ":point:block:" + fmt.Sprint(len(b.blocks)), OperationIDs: []string{}, Successors: []string{}})
			if !b.terminated(bodyEnd) {
				b.edge(b.blocks[bodyEnd].ID, condID, "back")
			}
			b.current = len(b.blocks) - 1
			b.initialized = before
		case "defect":
			dummy := b.names[b.function.Parameter.Name]
			b.addOperation(core.LinearOperation{Kind: core.OpDefect, SourceID: dummy, Reason: statement.Value, TypeID: b.types["U64"]})
		default:
			b.refuse("check.scalar_statement", statement.Span, "unsupported scalar statement")
		}
	}
}

func (b *scalarBuilder) newNamedBlock(id string) {
	b.blocks = append(b.blocks, core.Block{ID: id, PointID: b.functionID + ":point:block:" + fmt.Sprint(len(b.blocks)), OperationIDs: []string{}, Successors: []string{}})
	b.current = len(b.blocks) - 1
}

func (b *scalarBuilder) terminated(index int) bool {
	if len(b.blocks[index].OperationIDs) == 0 {
		return false
	}
	lastID := b.blocks[index].OperationIDs[len(b.blocks[index].OperationIDs)-1]
	for _, op := range b.linear.Operations {
		if op.ID == lastID {
			return op.Kind == core.OpDefect || op.Kind == core.OpReturn
		}
	}
	return false
}

func (b *scalarBuilder) expression(expr *ast.Expr) (string, string) {
	if expr == nil {
		return "", ""
	}
	b.work++
	switch expr.Kind {
	case "name":
		id, ok := b.names[expr.Name]
		if !ok {
			b.refuse("name.unknown_place", expr.Span, "unknown scalar local")
			return "", ""
		}
		if !b.initialized[id] {
			b.refuse("name.uninitialized_place", expr.Span, "scalar local is not definitely initialized")
			return id, ""
		}
		place := b.places[id]
		for name, typeID := range b.types {
			if typeID == place.TypeID {
				return id, name
			}
		}
		return id, ""
	case "number":
		value, err := strconv.ParseUint(expr.Value, 0, 64)
		if err != nil {
			value, err = strconv.ParseUint(expr.Value, 10, 64)
		}
		if err != nil {
			b.refuse("type.invalid_u64", expr.Span, "invalid U64 literal")
			return "", ""
		}
		place := b.addPlace("literal", "U64", false)
		b.addOperation(core.LinearOperation{Kind: core.OpConst, TargetID: place.ID, ConstU64: strconv.FormatUint(value, 10), TypeID: b.types["U64"]})
		b.initialized[place.ID] = true
		return place.ID, "U64"
	case "add", "less":
		left, leftType := b.expression(expr.Left)
		right, rightType := b.expression(expr.Right)
		if leftType != "U64" || rightType != "U64" {
			b.refuse("type.scalar_operand", expr.Span, "U64 operator requires U64 operands")
			return "", ""
		}
		resultType, kind := "U64", core.OpAddChecked
		if expr.Kind == "less" {
			resultType, kind = "Bool", core.OpLessU64
		}
		place := b.addPlace("scalar", ""+resultType, false)
		b.addOperation(core.LinearOperation{Kind: kind, SourceID: left, RightID: right, TargetID: place.ID, TypeID: b.types[resultType]})
		b.initialized[place.ID] = true
		return place.ID, resultType
	default:
		b.refuse("syntax.scalar_expression", expr.Span, "unsupported scalar expression")
		return "", ""
	}
}

func cloneInitialized(input map[string]bool) map[string]bool {
	clone := make(map[string]bool, len(input))
	for id, initialized := range input {
		clone[id] = initialized
	}
	return clone
}

func (b *scalarBuilder) refuse(code string, span diagnostic.Span, message string) {
	b.diagnostics = append(b.diagnostics, diagnostic.Error(code, span, message))
}
