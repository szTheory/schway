package check

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
)

// ---------------------------------------------------------------------
// 13-05 Task 1: wrap_call_in_try on syntax.fallible_call_not_consumed
// ---------------------------------------------------------------------

// TestWrapCallInTryRepairIsDriverEligible is 13-05 Task 1's Test 1
// (D-13-09.2): a binding whose RHS is a call to a declared foreign symbol,
// not consumed by `try`, yields syntax.fallible_call_not_consumed carrying
// exactly one wrap_call_in_try repair whose Span equals binding.RHS.Span,
// whose Replacement re-emits the call as the operand of `try`, and whose
// Applicability is machine-applicable (diagnostic.DriverEligible true).
func TestWrapCallInTryRepairIsDriverEligible(t *testing.T) {
	rhsSpan := diagnostic.Span{Start: 42, End: 68}
	binding := ast.Binding{
		Name: "handle",
		RHS: ast.RHS{
			Kind: "call", Callee: "lang_res_open", Arguments: []string{"request"}, Span: rhsSpan,
		},
		Span: diagnostic.Span{Start: 30, End: 68},
	}
	typeFact := core.TypeFact{ID: "s1:m:fn:main:type:0", Shape: core.TypeRef{Constructor: "Byte"}}
	places := map[string]*placeState{
		"request": {place: core.Place{ID: "s1:m:fn:main:place:0", Name: "request", TypeID: typeFact.ID}, initialized: true},
	}
	foreignSymbols := map[string]foreignSymbolInfo{
		"lang_res_open": {Name: "lang_res_open"},
	}

	_, _, diag := resolveCallBinding("s1:m:fn:main", 0, binding, places, map[string]calleeContract{}, typeFact, foreignSymbols)
	if diag == nil {
		t.Fatal("expected syntax.fallible_call_not_consumed, got nil")
	}
	if diag.Code != "syntax.fallible_call_not_consumed" {
		t.Fatalf("expected syntax.fallible_call_not_consumed, got %q", diag.Code)
	}
	if len(diag.Repairs) != 1 {
		t.Fatalf("expected exactly one repair, got %d: %+v", len(diag.Repairs), diag.Repairs)
	}
	repair := diag.Repairs[0]
	if repair.Kind != "wrap_call_in_try" {
		t.Fatalf("expected kind wrap_call_in_try, got %q", repair.Kind)
	}
	if repair.Span == nil || *repair.Span != rhsSpan {
		t.Fatalf("expected repair span to equal binding.RHS.Span (%+v), got %+v", rhsSpan, repair.Span)
	}
	wantReplacement := "try lang_res_open(request)"
	if repair.Replacement != wantReplacement {
		t.Fatalf("expected replacement %q, got %q", wantReplacement, repair.Replacement)
	}
	if repair.Applicability != diagnostic.ApplicabilityMachineApplicable {
		t.Fatalf("expected MachineApplicable, got %q", repair.Applicability)
	}
	if !diagnostic.DriverEligible(repair) {
		t.Fatalf("expected DriverEligible to return true for %+v", repair)
	}
}

// TestWrapCallInTryEmitsNoRepairOnMalformedCall is 13-05 Task 1's Test 2:
// the fail-closed fallback. Calls are arity-1 in this language and the
// arity check at the top of resolveCallBinding already refuses anything
// else before this emission site is ever reached, but the repair's own
// guard is defensive -- demonstrated here via the guard's other
// fail-closed leg: an empty callee name.
func TestWrapCallInTryEmitsNoRepairOnMalformedCall(t *testing.T) {
	rhsSpan := diagnostic.Span{Start: 42, End: 68}
	binding := ast.Binding{
		Name: "handle",
		RHS: ast.RHS{
			Kind: "call", Callee: "", Arguments: []string{"request"}, Span: rhsSpan,
		},
	}
	typeFact := core.TypeFact{ID: "s1:m:fn:main:type:0", Shape: core.TypeRef{Constructor: "Byte"}}
	places := map[string]*placeState{
		"request": {place: core.Place{ID: "s1:m:fn:main:place:0", Name: "request", TypeID: typeFact.ID}, initialized: true},
	}
	foreignSymbols := map[string]foreignSymbolInfo{
		"": {Name: ""},
	}

	_, _, diag := resolveCallBinding("s1:m:fn:main", 0, binding, places, map[string]calleeContract{}, typeFact, foreignSymbols)
	if diag == nil {
		t.Fatal("expected syntax.fallible_call_not_consumed, got nil")
	}
	if len(diag.Repairs) != 0 {
		t.Fatalf("expected zero repairs on an empty callee name, got %d: %+v", len(diag.Repairs), diag.Repairs)
	}
}
