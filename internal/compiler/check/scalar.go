package check

import (
	"fmt"
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
		} else {
			b.addOperation(core.LinearOperation{Kind: core.OpReturn, SourceID: resultID, TypeID: b.types["U64"]})
		}
	}
	if len(b.diagnostics) != 0 {
		return core.Function{}, b.diagnostics, b.work
	}
	b.linear.Blocks, b.linear.Edges = b.blocks, b.edges
	return core.Function{ID: functionID, Name: function.Name, EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return", Parameter: core.Parameter{ID: parameter.ID, Name: function.Parameter.Name, Type: "U64"}, ReturnType: "U64", Linear: &b.linear, Span: function.Span}, nil, b.work
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
