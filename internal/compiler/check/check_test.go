package check

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
)

func TestOwnershipSequenceExhaustive(t *testing.T) {
	for length := 0; length <= 4; length++ {
		cases := 1
		for index := 0; index < length; index++ {
			cases *= 4
		}
		for encoded := 0; encoded < cases; encoded++ {
			body := generatedOwnershipBody(encoded, length)
			got := analyzeStraightLine("test:fn", "owner", diagnostic.Span{Start: 1, End: 6}, byteTypeFact(), &body)
			want := oracleStraightLine("test:fn", "owner", byteTypeFact(), &body)
			assertSupportEqual(t, fmt.Sprintf("length=%d case=%d", length, encoded), got, want)
		}
	}

	witness := ast.LinearBody{
		Bindings: []ast.Binding{
			binding("view", "borrow", "owner", 10),
			binding("moved", "take", "owner", 20),
			binding("observed", "read", "view", 30),
		},
		Result: "moved",
		Span:   diagnostic.Span{Start: 10, End: 38},
	}
	got := analyzeStraightLine("test:witness", "owner", diagnostic.Span{Start: 1, End: 6}, bufferTypeFact(), &witness)
	if got.DiagnosticCode != "ownership.move_while_borrowed" || len(got.Operations) != 1 {
		t.Fatalf("four-step witness changed: %+v", got)
	}
	if len(got.LoanFinalUses) != 1 || got.LoanFinalUses[0].OperationIndex != 2 {
		t.Fatalf("later loan read is not load-bearing: %+v", got.LoanFinalUses)
	}
	withoutLaterRead := witness
	withoutLaterRead.Bindings = withoutLaterRead.Bindings[:2]
	withoutLaterRead.Result = "moved"
	withoutLaterRead.Span.End = 28
	legal := analyzeStraightLine("test:unused", "owner", diagnostic.Span{Start: 1, End: 6}, bufferTypeFact(), &withoutLaterRead)
	if legal.DiagnosticCode != "" {
		t.Fatalf("unused loan did not end before move: %+v", legal)
	}
}

func TestOwnershipOracleTracksLoansPerOwner(t *testing.T) {
	body := ast.LinearBody{
		Bindings: []ast.Binding{
			binding("view", "borrow", "owner", 0),
			binding("other", "read", "owner", 4),
			binding("moved", "take", "other", 8),
			binding("observed", "read", "view", 12),
		},
		Result: "moved",
		Span:   diagnostic.Span{End: 20},
	}
	got := analyzeStraightLine("test:multi", "owner", diagnostic.Span{}, byteTypeFact(), &body)
	want := oracleStraightLine("test:multi", "owner", byteTypeFact(), &body)
	assertSupportEqual(t, "borrow one owner while moving another", got, want)
	if got.DiagnosticCode != "" {
		t.Fatalf("unrelated owner move was blocked: %+v", got)
	}

	shadowed := ast.LinearBody{
		Bindings: []ast.Binding{binding("value", "read", "owner", 0), binding("value", "read", "owner", 4)},
		Result:   "value",
		Span:     diagnostic.Span{End: 12},
	}
	assertSupportEqual(t, "shadowed binding", analyzeStraightLine("test:shadow", "owner", diagnostic.Span{}, byteTypeFact(), &shadowed), oracleStraightLine("test:shadow", "owner", byteTypeFact(), &shadowed))
}

func TestOwnershipWorkSeries(t *testing.T) {
	for _, operations := range []int{10, 100, 1_000, 10_000} {
		bindings := make([]ast.Binding, operations)
		for index := range bindings {
			bindings[index] = binding(fmt.Sprintf("copy%d", index), "read", "owner", index*2)
		}
		body := ast.LinearBody{Bindings: bindings, Result: "owner", Span: diagnostic.Span{End: operations*2 + 1}}
		got := analyzeStraightLine("test:scale", "owner", diagnostic.Span{}, byteTypeFact(), &body)
		wantWork := 1 + 2*(operations+1)
		if got.DiagnosticCode != "" || len(got.Operations) != operations+1 || got.Work != wantWork {
			t.Fatalf("operations=%d result=%+v want work=%d", operations, got, wantWork)
		}
		if got.Work > 2*len(got.Operations)+1 {
			t.Fatalf("operations=%d exceeded linear bound: work=%d core_ops=%d", operations, got.Work, len(got.Operations))
		}
	}
}

func FuzzOwnershipLinear(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{2, 1, 3})
	f.Add([]byte{1, 0})
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 32 {
			input = input[:32]
		}
		body := generatedOwnershipBodyBytes(input)
		got := analyzeStraightLine("test:fuzz", "owner", diagnostic.Span{Start: 1, End: 6}, byteTypeFact(), &body)
		want := oracleStraightLine("test:fuzz", "owner", byteTypeFact(), &body)
		assertSupportEqual(t, fmt.Sprintf("input=%v", input), got, want)
	})
}

func generatedOwnershipBody(encoded, length int) ast.LinearBody {
	input := make([]byte, length)
	for index := range input {
		input[index] = byte(encoded % 4)
		encoded /= 4
	}
	return generatedOwnershipBodyBytes(input)
}

func generatedOwnershipBodyBytes(input []byte) ast.LinearBody {
	body := ast.LinearBody{Result: "owner"}
	latestLoan := "missing"
	for index, raw := range input {
		name := fmt.Sprintf("value%d", index)
		kind := "read"
		source := "owner"
		switch raw % 4 {
		case 1:
			kind = "take"
		case 2:
			kind = "borrow"
			latestLoan = name
		case 3:
			source = latestLoan
		}
		body.Bindings = append(body.Bindings, binding(name, kind, source, index*4))
	}
	body.Span = diagnostic.Span{End: len(input)*4 + 1}
	return body
}

func binding(name, kind, source string, start int) ast.Binding {
	span := diagnostic.Span{Start: start, End: start + 3}
	return ast.Binding{Name: name, RHS: ast.RHS{Kind: kind, Source: source, Span: span}, Span: span}
}

func byteTypeFact() core.TypeFact {
	return core.TypeFact{
		ID: "test:fn:type:0", Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
		Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
		NegativeWitnesses: []core.AbilityWitness{},
	}
}

func bufferTypeFact() core.TypeFact {
	return core.TypeFact{
		ID: "test:fn:type:0", Shape: core.TypeRef{Constructor: "Buffer", Arguments: []core.TypeRef{}},
		Abilities: []core.Ability{core.AbilityDrop, core.AbilitySend, core.AbilityEscape},
		NegativeWitnesses: []core.AbilityWitness{
			{Ability: core.AbilityCopy, Path: []string{"Buffer"}},
			{Ability: core.AbilityShare, Path: []string{"Buffer"}},
		},
	}
}

// oracleStraightLine deliberately duplicates the tiny semantic law. It does
// not call production last-use, transition, identity, normalization, or state
// snapshot helpers.
type testOraclePlace struct {
	id          string
	initialized bool
}

type testOracleLoan struct {
	ownerID string
	lastUse int
}

func oracleStraightLine(functionID, parameterName string, typeFact core.TypeFact, body *ast.LinearBody) ownershipSupport {
	lastUses := make(map[string]int)
	loanOrder := make([]string, 0)
	for index, candidate := range body.Bindings {
		if candidate.RHS.Kind != "borrow" {
			continue
		}
		last := index
		for later := index + 1; later < len(body.Bindings); later++ {
			if body.Bindings[later].RHS.Source == candidate.Name {
				last = later
			}
		}
		if body.Result == candidate.Name {
			last = len(body.Bindings)
		}
		lastUses[candidate.Name] = last
		loanOrder = append(loanOrder, candidate.Name)
	}

	result := ownershipSupport{Operations: []core.LinearOperation{}, LoanFinalUses: []loanFinalUseFact{}, States: []ownershipStateFact{}, Work: typeNodeCountOracle(typeFact.Shape) + len(body.Bindings) + 1}
	for _, name := range loanOrder {
		index := -1
		for candidateIndex, candidate := range body.Bindings {
			if candidate.Name == name {
				index = candidateIndex
				break
			}
		}
		result.LoanFinalUses = append(result.LoanFinalUses, loanFinalUseFact{LoanID: fmt.Sprintf("%s:loan:%d", functionID, index), Binding: name, OperationIndex: lastUses[name]})
	}

	parameterID := functionID + ":place:0"
	places := map[string]*testOraclePlace{parameterName: {id: parameterID, initialized: true}}
	activeLoans := make(map[string]testOracleLoan)
	for index, candidate := range body.Bindings {
		result.Work++
		source, ok := places[candidate.RHS.Source]
		if !ok {
			result.DiagnosticCode = "name.unknown"
			return result
		}
		if !source.initialized {
			result.DiagnosticCode = "ownership.use_after_move"
			return result
		}
		if candidate.RHS.Kind == "take" && hasFutureLoanOracle(activeLoans, source.id, index) {
			result.DiagnosticCode = "ownership.move_while_borrowed"
			return result
		}
		kind := core.OpCopy
		loanID := ""
		switch candidate.RHS.Kind {
		case "take":
			kind = core.OpMove
			source.initialized = false
		case "borrow":
			kind = core.OpBorrowShared
			loanID = fmt.Sprintf("%s:loan:%d", functionID, index)
			activeLoans[loanID] = testOracleLoan{ownerID: source.id, lastUse: lastUses[candidate.Name]}
		default:
			if !oracleHasAbility(typeFact.Abilities, core.AbilityCopy) {
				result.DiagnosticCode = "ownership.transfer_requires_take"
				return result
			}
		}
		targetID := fmt.Sprintf("%s:place:%d", functionID, index+1)
		places[candidate.Name] = &testOraclePlace{id: targetID, initialized: true}
		result.Operations = append(result.Operations, core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", functionID, index), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, index),
			Kind: kind, SourceID: source.id, TargetID: targetID, LoanID: loanID, TypeID: typeFact.ID,
		})
		for id, loan := range activeLoans {
			if loan.lastUse <= index {
				delete(activeLoans, id)
			}
		}
		result.States = append(result.States, oracleSnapshot(index, places, activeLoans))
	}
	result.Work++
	returned, ok := places[body.Result]
	if !ok {
		result.DiagnosticCode = "name.unknown"
		return result
	}
	if !returned.initialized {
		result.DiagnosticCode = "ownership.use_after_move"
		return result
	}
	ordinal := len(body.Bindings)
	result.Operations = append(result.Operations, core.LinearOperation{
		ID: fmt.Sprintf("%s:op:%d", functionID, ordinal), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, ordinal),
		Kind: core.OpReturn, SourceID: returned.id, TypeID: typeFact.ID,
	})
	result.States = append(result.States, oracleSnapshot(ordinal, places, activeLoans))
	return result
}

func oracleSnapshot(index int, places map[string]*testOraclePlace, activeLoans map[string]testOracleLoan) ownershipStateFact {
	initialized := make([]string, 0)
	for _, place := range places {
		if place.initialized {
			initialized = append(initialized, place.id)
		}
	}
	sort.Strings(initialized)
	loans := make([]string, 0, len(activeLoans))
	for id := range activeLoans {
		loans = append(loans, id)
	}
	sort.Strings(loans)
	return ownershipStateFact{OperationIndex: index, InitializedPlaces: initialized, ActiveLoans: loans}
}

func hasFutureLoanOracle(active map[string]testOracleLoan, ownerID string, index int) bool {
	for _, loan := range active {
		if loan.ownerID == ownerID && loan.lastUse > index {
			return true
		}
	}
	return false
}

func oracleHasAbility(abilities []core.Ability, wanted core.Ability) bool {
	for _, candidate := range abilities {
		if candidate == wanted {
			return true
		}
	}
	return false
}

func typeNodeCountOracle(value core.TypeRef) int {
	count := 1
	for _, argument := range value.Arguments {
		count += typeNodeCountOracle(argument)
	}
	return count
}

func assertSupportEqual(t *testing.T, name string, got, want ownershipSupport) {
	t.Helper()
	if !reflect.DeepEqual(got.Operations, want.Operations) ||
		!reflect.DeepEqual(got.LoanFinalUses, want.LoanFinalUses) ||
		!reflect.DeepEqual(got.States, want.States) ||
		got.Work != want.Work || got.DiagnosticCode != want.DiagnosticCode {
		t.Fatalf("%s support mismatch:\ngot  %+v\nwant %+v", name, got, want)
	}
}
