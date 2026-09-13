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
// guard is defensive -- constructed here via a directly-injected zero-
// argument RHS that bypasses the arity gate by construction (Arguments is
// empty, so len(binding.RHS.Arguments) != 1 never triggers the earlier
// check.call_arity_unsupported return because this test targets the
// emission site's OWN guard, not the arity gate). Since the arity gate at
// the top of resolveCallBinding always fires first on a zero-argument
// call, this test instead demonstrates the guard's other fail-closed leg:
// an empty callee name.
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

// ---------------------------------------------------------------------
// 13-05 Task 2: use_matching_argument and the D-13-10 uniqueness gate
// ---------------------------------------------------------------------

// TestUseMatchingArgumentUniquenessGate is 13-05 Task 2's Test 1
// (D-13-10): all three partitions of the uniqueness gate, asserted by
// exact repair count. Zero and two-or-more matches are POSITIVE tests of
// correct refusal, not workarounds -- the resulting unrepairable driver
// outcome on those two partitions is exactly what D-13-10 requires.
func TestUseMatchingArgumentUniquenessGate(t *testing.T) {
	rhsSpan := diagnostic.Span{Start: 10, End: 30}
	binding := ast.Binding{
		Name: "result",
		RHS:  ast.RHS{Kind: "call", Callee: "identity", Arguments: []string{"value"}, Span: rhsSpan},
	}
	contracts := map[string]calleeContract{
		"identity": {ID: "s1:m:fn:identity", ParameterType: "Buffer", ReturnType: "Buffer"},
	}

	t.Run("exactly one match fires the repair", func(t *testing.T) {
		typeFact := core.TypeFact{ID: "s1:m:fn:main:type:0", Shape: core.TypeRef{Constructor: "Byte"}}
		places := map[string]*placeState{
			"value": {place: core.Place{ID: "s1:m:fn:main:place:0", Name: "value", TypeID: typeFact.ID}, initialized: true},
		}
		_, _, diag := resolveCallBinding("s1:m:fn:main", 0, binding, places, contracts, typeFact, nil)
		if diag == nil || diag.Code != checkCallArgumentTypeMismatch {
			t.Fatalf("expected %s, got %+v", checkCallArgumentTypeMismatch, diag)
		}
		if len(diag.Repairs) != 1 {
			t.Fatalf("expected exactly one repair on the unique-match partition, got %d: %+v", len(diag.Repairs), diag.Repairs)
		}
		repair := diag.Repairs[0]
		if repair.Kind != "use_matching_argument" {
			t.Fatalf("expected kind use_matching_argument, got %q", repair.Kind)
		}
		if repair.Applicability != diagnostic.ApplicabilityMachineApplicable {
			t.Fatalf("expected MachineApplicable, got %q", repair.Applicability)
		}
		if repair.Applicability == diagnostic.ApplicabilityRequiresConfirmation {
			t.Fatal("use_matching_argument must never be RequiresConfirmation -- the gate is a precondition, not a downgrade")
		}
		if !diagnostic.DriverEligible(repair) {
			t.Fatalf("expected DriverEligible to return true for %+v", repair)
		}
	})

	t.Run("zero matches emits no repair", func(t *testing.T) {
		// Synthetic: a places map where the argument's own entry carries a
		// DIFFERENT TypeID than typeFact.ID. Not reachable from real
		// parsed source (every place in a function shares its own single
		// type fact, D-07-09), but a direct unit call to resolveCallBinding
		// can construct it to exercise the zero-match partition, which the
		// diagnostic's own trigger condition otherwise makes unreachable
		// through real source (the argument itself always shares
		// typeFact.ID in production).
		typeFact := core.TypeFact{ID: "s1:m:fn:main:type:0", Shape: core.TypeRef{Constructor: "Byte"}}
		places := map[string]*placeState{
			"value": {place: core.Place{ID: "s1:m:fn:main:place:0", Name: "value", TypeID: "s1:m:fn:main:type:other"}, initialized: true},
		}
		_, _, diag := resolveCallBinding("s1:m:fn:main", 0, binding, places, contracts, typeFact, nil)
		if diag == nil || diag.Code != checkCallArgumentTypeMismatch {
			t.Fatalf("expected %s, got %+v", checkCallArgumentTypeMismatch, diag)
		}
		if len(diag.Repairs) != 0 {
			t.Fatalf("expected zero repairs on the zero-match partition, got %d: %+v", len(diag.Repairs), diag.Repairs)
		}
	})

	t.Run("two or more matches emits no repair", func(t *testing.T) {
		typeFact := core.TypeFact{ID: "s1:m:fn:main:type:0", Shape: core.TypeRef{Constructor: "Byte"}}
		places := map[string]*placeState{
			"value": {place: core.Place{ID: "s1:m:fn:main:place:0", Name: "value", TypeID: typeFact.ID}, initialized: true},
			"spare": {place: core.Place{ID: "s1:m:fn:main:place:1", Name: "spare", TypeID: typeFact.ID}, initialized: true},
		}
		_, _, diag := resolveCallBinding("s1:m:fn:main", 0, binding, places, contracts, typeFact, nil)
		if diag == nil || diag.Code != checkCallArgumentTypeMismatch {
			t.Fatalf("expected %s, got %+v", checkCallArgumentTypeMismatch, diag)
		}
		if len(diag.Repairs) != 0 {
			t.Fatalf("expected zero repairs on the two-or-more-match partition, got %d: %+v", len(diag.Repairs), diag.Repairs)
		}
	})
}

// TestUseMatchingArgumentUninitializedPlaceIsNotAMatch is 13-05 Task 2's
// Test 2: a place that is in scope but not initialized (moved away) never
// counts as a match, even when its TypeID would otherwise qualify.
func TestUseMatchingArgumentUninitializedPlaceIsNotAMatch(t *testing.T) {
	rhsSpan := diagnostic.Span{Start: 10, End: 30}
	binding := ast.Binding{
		Name: "result",
		RHS:  ast.RHS{Kind: "call", Callee: "identity", Arguments: []string{"value"}, Span: rhsSpan},
	}
	contracts := map[string]calleeContract{
		"identity": {ID: "s1:m:fn:identity", ParameterType: "Buffer", ReturnType: "Buffer"},
	}
	typeFact := core.TypeFact{ID: "s1:m:fn:main:type:0", Shape: core.TypeRef{Constructor: "Byte"}}
	moved := diagnostic.Span{Start: 1, End: 2}
	places := map[string]*placeState{
		"value": {place: core.Place{ID: "s1:m:fn:main:place:0", Name: "value", TypeID: typeFact.ID}, initialized: true},
		// "spare" shares the same TypeID (would otherwise be a second
		// match) but is NOT initialized -- moved away before this call.
		"spare": {place: core.Place{ID: "s1:m:fn:main:place:1", Name: "spare", TypeID: typeFact.ID}, initialized: false, movedAt: &moved},
	}
	_, _, diag := resolveCallBinding("s1:m:fn:main", 0, binding, places, contracts, typeFact, nil)
	if diag == nil || diag.Code != checkCallArgumentTypeMismatch {
		t.Fatalf("expected %s, got %+v", checkCallArgumentTypeMismatch, diag)
	}
	if len(diag.Repairs) != 1 {
		t.Fatalf("expected exactly one repair (the uninitialized place must not count), got %d: %+v", len(diag.Repairs), diag.Repairs)
	}
	if diag.Repairs[0].Replacement != "identity(value)" {
		t.Fatalf("expected the replacement to name the sole initialized match (value), got %q", diag.Repairs[0].Replacement)
	}
}
