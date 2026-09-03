package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/codename-lang/lang/internal/compiler/ability"
	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
)

type Result struct {
	Program     core.Program
	Diagnostics []diagnostic.Diagnostic
}

func Program(program ast.Program) Result {
	result := Result{Program: core.Program{Schema: core.Schema, Module: program.Module, ModuleID: semanticID(program.Module, "module", program.Module)}}
	types := make(map[string]core.DataType)
	for _, declaration := range program.Data {
		alternatives := make([]string, 0, len(declaration.Alternatives))
		for _, alternative := range declaration.Alternatives {
			alternatives = append(alternatives, alternative.Name)
		}
		dataType := core.DataType{ID: semanticID(program.Module, "type", declaration.Name), Name: declaration.Name, Alternatives: alternatives, Span: declaration.Span}
		types[declaration.Name] = dataType
		result.Program.DataTypes = append(result.Program.DataTypes, dataType)
	}

	for _, function := range program.Funcs {
		functionID := semanticID(program.Module, "fn", function.Name)
		if !function.Body.HasClosedVariant() {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("core.invalid_body", function.Span, "function must have exactly one body variant"))
			continue
		}
		if function.Body.Linear != nil {
			checked, diagnostics := checkLinear(program.Module, functionID, function)
			result.Diagnostics = append(result.Diagnostics, diagnostics...)
			if len(diagnostics) == 0 {
				result.Program.Schema = core.Schema1
				result.Program.Functions = append(result.Program.Functions, checked)
			}
			continue
		}
		matchID := semanticID(program.Module, "match", function.Name)
		dataType, ok := types[function.Parameter.Type.Constructor]
		if !ok {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("type.unknown", function.Parameter.Span, "unknown parameter type"))
			continue
		}
		if !sameType(function.ReturnType, function.Parameter.Type) {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("type.return_mismatch", function.Span, "S1 match result must have the parameter type"))
			continue
		}
		if function.Body.Scrutinee != function.Parameter.Name {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("name.unknown_scrutinee", function.Body.Span, "match scrutinee is not the function parameter"))
			continue
		}
		seen := make(map[string]bool)
		arms := make([]core.MatchArm, 0, len(function.Body.Arms))
		for index, arm := range function.Body.Arms {
			if seen[arm.Pattern] {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("match.subsumed", arm.Span, "alternative is already matched"))
				continue
			}
			if !contains(dataType.Alternatives, arm.Pattern) {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("match.unreachable", arm.Span, "pattern is not an alternative of the scrutinee type"))
				continue
			}
			if !contains(dataType.Alternatives, arm.Value) {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("type.invalid_variant", arm.Span, "match result is not an alternative of the return type"))
				continue
			}
			seen[arm.Pattern] = true
			arms = append(arms, core.MatchArm{
				ID: fmt.Sprintf("%s:arm:%d", functionID, index), EdgeID: fmt.Sprintf("%s:edge:%s", matchID, arm.Pattern),
				Pattern: arm.Pattern, Value: arm.Value,
			})
		}
		missing := make([]string, 0)
		for _, alternative := range dataType.Alternatives {
			if !seen[alternative] {
				missing = append(missing, alternative)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			causes := make([]diagnostic.Cause, 0, len(missing))
			for _, name := range missing {
				causes = append(causes, diagnostic.Cause{Kind: "missing_alternative", Detail: name})
			}
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("match.non_exhaustive", function.Body.Span, "match does not cover every alternative", causes...))
		}
		result.Program.Functions = append(result.Program.Functions, core.Function{
			ID: functionID, Name: function.Name,
			EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
			Parameter:  core.Parameter{ID: semanticID(program.Module, "parameter", function.Name+"."+function.Parameter.Name), Name: function.Parameter.Name, Type: function.Parameter.Type.Constructor},
			ReturnType: function.ReturnType.Constructor,
			Match:      &core.Match{ID: matchID, PointID: functionID + ":point:match", Scrutinee: function.Body.Scrutinee, Arms: arms},
			Span:       function.Span,
		})
	}
	return result
}

func checkLinear(module, functionID string, function ast.FuncDecl) (core.Function, []diagnostic.Diagnostic) {
	parameterType := coreType(function.Parameter.Type)
	if !sameType(function.ReturnType, function.Parameter.Type) {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.return_mismatch", function.Span, "linear result must have the parameter type")}
	}
	derived, err := ability.Derive(parameterType)
	if err != nil {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.unknown", function.Parameter.Span, err.Error())}
	}
	typeID := functionID + ":type:0"
	parameterID := functionID + ":place:0"
	linear := &core.LinearBody{
		ID:         functionID + ":linear",
		Types:      []core.TypeFact{{ID: typeID, Shape: parameterType, Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses}},
		Places:     []core.Place{{ID: parameterID, Name: function.Parameter.Name, TypeID: typeID}},
		Operations: []core.LinearOperation{},
	}
	type placeState struct {
		place        core.Place
		declared     diagnostic.Span
		initialized  bool
		movedAt      *diagnostic.Span
		moveTargetID string
		loan         *loanState
	}
	typeFacts := linear.Types[0]
	loanUses := discoverLoanLastUses(function.Body.Linear)
	places := map[string]*placeState{
		function.Parameter.Name: {place: linear.Places[0], declared: function.Parameter.Span, initialized: true},
	}
	activeLoans := make(map[string]map[string]*loanState)
	endLoans := func(index int) {
		for ownerID, loans := range activeLoans {
			for loanID, loan := range loans {
				if loan.lastUse <= index {
					delete(loans, loanID)
				}
			}
			if len(loans) == 0 {
				delete(activeLoans, ownerID)
			}
		}
	}
	useAfterMove := func(span diagnostic.Span, state *placeState) diagnostic.Diagnostic {
		causes := []diagnostic.Cause{
			{Kind: "declared_here", Span: spanPointer(state.declared)},
			{Kind: "moved_here", Span: state.movedAt},
			{Kind: "place", Detail: state.place.ID},
			{Kind: "transfer_target", Detail: state.moveTargetID},
			{Kind: "type", Detail: state.place.TypeID},
		}
		return diagnostic.ErrorWithRepairs(
			"ownership.use_after_move", span, "value was used after ownership transferred", causes,
			diagnostic.Repair{Kind: "use_transfer_target", Detail: state.moveTargetID},
			diagnostic.Repair{Kind: "move_use_before_transfer"},
		)
	}
	for index, binding := range function.Body.Linear.Bindings {
		source, ok := places[binding.RHS.Source]
		if !ok {
			return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("name.unknown", binding.RHS.Span, "binding source is unknown")}
		}
		if !source.initialized {
			return core.Function{}, []diagnostic.Diagnostic{useAfterMove(binding.RHS.Span, source)}
		}
		target := core.Place{ID: fmt.Sprintf("%s:place:%d", functionID, index+1), Name: binding.Name, TypeID: source.place.TypeID}
		kind := core.OpCopy
		var loan *loanState
		switch binding.RHS.Kind {
		case "take":
			if loans := activeLoans[source.place.ID]; len(loans) > 0 {
				loanIDs := make([]string, 0, len(loans))
				for loanID := range loans {
					loanIDs = append(loanIDs, loanID)
				}
				sort.Strings(loanIDs)
				blocking := loans[loanIDs[0]]
				causes := []diagnostic.Cause{
					{Kind: "borrow_created_here", Span: spanPointer(blocking.borrowedAt)},
					{Kind: "borrow_used_later", Span: spanPointer(blocking.lastUseSpan)},
					{Kind: "loan", Detail: blocking.id},
					{Kind: "owner", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return core.Function{}, []diagnostic.Diagnostic{diagnostic.ErrorWithRepairs(
					"ownership.move_while_borrowed", binding.RHS.Span, "cannot transfer ownership while a future-used shared loan is live", causes,
					diagnostic.Repair{Kind: "move_after_last_borrow_use"},
				)}
			}
			kind = core.OpMove
			source.initialized = false
			source.movedAt = spanPointer(binding.RHS.Span)
			source.moveTargetID = target.ID
		case "borrow":
			kind = core.OpBorrowShared
			use := loanUses[binding.Name]
			loan = &loanState{
				id: fmt.Sprintf("%s:loan:%d", functionID, index), ownerID: source.place.ID,
				borrowedAt: binding.RHS.Span, lastUse: use.index, lastUseSpan: use.span,
			}
			if activeLoans[source.place.ID] == nil {
				activeLoans[source.place.ID] = make(map[string]*loanState)
			}
			activeLoans[source.place.ID][loan.id] = loan
		default:
			if !ability.Has(derived, core.AbilityCopy) {
				causes := []diagnostic.Cause{
					{Kind: "declared_here", Span: spanPointer(source.declared)},
					{Kind: "missing_ability", Detail: missingAbilityDetail(typeFacts, core.AbilityCopy)},
					{Kind: "place", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return core.Function{}, []diagnostic.Diagnostic{diagnostic.ErrorWithRepairs(
					"ownership.transfer_requires_take", binding.RHS.Span, "noncopyable binding requires explicit take", causes,
					diagnostic.Repair{Kind: "insert_take"},
				)}
			}
		}
		linear.Places = append(linear.Places, target)
		places[binding.Name] = &placeState{place: target, declared: binding.Span, initialized: true, loan: loan}
		linear.Operations = append(linear.Operations, core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", functionID, index), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, index),
			Kind: kind, SourceID: source.place.ID, TargetID: target.ID, LoanID: loanID(loan), TypeID: source.place.TypeID,
		})
		endLoans(index)
	}
	returned, ok := places[function.Body.Linear.Result]
	if !ok {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("name.unknown", function.Body.Linear.Span, "linear result is unknown")}
	}
	if !returned.initialized {
		return core.Function{}, []diagnostic.Diagnostic{useAfterMove(function.Body.Linear.Span, returned)}
	}
	ordinal := len(linear.Operations)
	linear.Operations = append(linear.Operations, core.LinearOperation{
		ID: fmt.Sprintf("%s:op:%d", functionID, ordinal), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, ordinal),
		Kind: core.OpReturn, SourceID: returned.place.ID, TypeID: returned.place.TypeID,
	})
	return core.Function{
		ID: functionID, Name: function.Name, EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter: core.Parameter{ID: parameterID, Name: function.Parameter.Name, Type: parameterType.Constructor}, ReturnType: function.ReturnType.Constructor,
		Linear: linear, Span: function.Span,
	}, nil
}

type loanUse struct {
	index int
	span  diagnostic.Span
}

type loanState struct {
	id          string
	ownerID     string
	borrowedAt  diagnostic.Span
	lastUse     int
	lastUseSpan diagnostic.Span
}

func discoverLoanLastUses(body *ast.LinearBody) map[string]loanUse {
	uses := make(map[string]loanUse)
	for index, binding := range body.Bindings {
		if binding.RHS.Kind == "borrow" {
			uses[binding.Name] = loanUse{index: index, span: binding.RHS.Span}
		}
		if use, ok := uses[binding.RHS.Source]; ok {
			use.index = index
			use.span = binding.RHS.Span
			uses[binding.RHS.Source] = use
		}
	}
	if use, ok := uses[body.Result]; ok {
		use.index = len(body.Bindings)
		use.span = body.Span
		uses[body.Result] = use
	}
	return uses
}

func missingAbilityDetail(fact core.TypeFact, requested core.Ability) string {
	for _, witness := range fact.NegativeWitnesses {
		if witness.Ability == requested {
			return string(requested) + ":" + strings.Join(witness.Path, ".")
		}
	}
	return string(requested)
}

func loanID(loan *loanState) string {
	if loan == nil {
		return ""
	}
	return loan.id
}

func spanPointer(span diagnostic.Span) *diagnostic.Span {
	copy := span
	return &copy
}

func coreType(value ast.TypeRef) core.TypeRef {
	result := core.TypeRef{Constructor: value.Constructor, Arguments: make([]core.TypeRef, 0, len(value.Arguments))}
	for _, argument := range value.Arguments {
		result.Arguments = append(result.Arguments, coreType(argument))
	}
	return result
}

func sameType(left, right ast.TypeRef) bool {
	return typeKey(coreType(left)) == typeKey(coreType(right))
}

func typeKey(value core.TypeRef) string {
	if len(value.Arguments) == 0 {
		return value.Constructor
	}
	parts := make([]string, 0, len(value.Arguments))
	for _, argument := range value.Arguments {
		parts = append(parts, typeKey(argument))
	}
	return value.Constructor + "<" + strings.Join(parts, ",") + ">"
}

func semanticID(module, kind, name string) string { return "s1:" + module + ":" + kind + ":" + name }

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
