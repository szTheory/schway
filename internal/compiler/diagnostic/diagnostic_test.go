package diagnostic

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

// --- Task 1: Repair grows additively -------------------------------------

func TestRepairCarriesSpanReplacementApplicability(t *testing.T) {
	span := Span{Start: 1, End: 5}
	repair := Repair{Kind: "insert_take", Span: &span, Replacement: "take x", Applicability: ApplicabilityMachineApplicable}
	encoded, err := json.Marshal(repair)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"span", "replacement", "applicability"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("expected key %q present in %s", key, encoded)
		}
	}
}

func TestApplicabilityVocabularyIsClosed(t *testing.T) {
	got := Applicabilities()
	want := []string{ApplicabilityMachineApplicable, ApplicabilityRequiresConfirmation, ApplicabilityUnspecified}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Applicabilities() = %v, want %v", got, want)
	}

	for _, v := range got {
		if err := ValidateRepair(Repair{Kind: "k", Applicability: v}); err != nil {
			t.Errorf("ValidateRepair(%q) = %v, want nil", v, err)
		}
	}
	for _, bad := range []string{"machineapplicable", "MACHINE_APPLICABLE", "Bogus", "true"} {
		if err := ValidateRepair(Repair{Kind: "k", Applicability: bad}); err == nil {
			t.Errorf("ValidateRepair(%q) = nil, want error", bad)
		}
	}
}

func TestEmptyApplicabilityDefaultsToUnspecified(t *testing.T) {
	if got := NormalizeApplicability(""); got != ApplicabilityUnspecified {
		t.Fatalf("NormalizeApplicability(\"\") = %q, want %q", got, ApplicabilityUnspecified)
	}
	if got := NormalizeApplicability(ApplicabilityMachineApplicable); got != ApplicabilityMachineApplicable {
		t.Fatalf("NormalizeApplicability(%q) = %q, want unchanged", ApplicabilityMachineApplicable, got)
	}

	// An empty Repair (zero value) must not marshal any of the three new
	// keys, and its Applicability must never be treated as MachineApplicable.
	var zero Repair
	if zero.Applicability != "" {
		t.Fatalf("zero-value Repair.Applicability = %q, want empty", zero.Applicability)
	}
	if NormalizeApplicability(zero.Applicability) == ApplicabilityMachineApplicable {
		t.Fatalf("empty applicability must never normalize to MachineApplicable")
	}
	encoded, err := json.Marshal(zero)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"span", "replacement", "applicability"} {
		if _, ok := decoded[key]; ok {
			t.Fatalf("zero-value Repair must omit key %q, got %s", key, encoded)
		}
	}
}

// --- Task 2: the identity split is executable and self-invalidating ------

func diagnosticWithRepair(repair Repair) Diagnostic {
	return ErrorWithRepairs("test.code", Span{Start: 0, End: 1}, "message", nil, repair)
}

func TestRepairIdentityUsesKindOnly(t *testing.T) {
	base := Repair{Kind: "insert_take", Detail: "d"}

	// Negative: span does not move the ID.
	shiftedSpan := Span{Start: 100, End: 200}
	withSpan := base
	withSpan.Span = &shiftedSpan
	if diagnosticWithRepair(base).ID != diagnosticWithRepair(withSpan).ID {
		t.Errorf("Span must not change diagnostic ID")
	}

	// Negative: replacement does not move the ID.
	withReplacement := base
	withReplacement.Replacement = "take x"
	if diagnosticWithRepair(base).ID != diagnosticWithRepair(withReplacement).ID {
		t.Errorf("Replacement must not change diagnostic ID")
	}

	// Negative: applicability does not move the ID.
	withApplicability := base
	withApplicability.Applicability = ApplicabilityMachineApplicable
	if diagnosticWithRepair(base).ID != diagnosticWithRepair(withApplicability).ID {
		t.Errorf("Applicability must not change diagnostic ID")
	}

	// Positive control: Kind DOES change the ID -- proves the test is not inert.
	differentKind := base
	differentKind.Kind = "different_kind"
	if diagnosticWithRepair(base).ID == diagnosticWithRepair(differentKind).ID {
		t.Errorf("Kind must change diagnostic ID (positive control failed)")
	}

	// Pre-existing property, re-asserted: the top-level diagnostic message
	// does not change the ID either.
	same := ErrorWithRepairs("test.code", Span{Start: 0, End: 1}, "message one", nil, base)
	other := ErrorWithRepairs("test.code", Span{Start: 0, End: 1}, "message two", nil, base)
	if same.ID != other.ID {
		t.Errorf("top-level message change must not change diagnostic ID")
	}
}

func TestRepairSpanShiftDoesNotChangeDiagnosticID(t *testing.T) {
	// Build a diagnostic from a real-shaped fixture: a repair carrying a
	// concrete span into source text, as check.go's insert_take repair does.
	original := Span{Start: 109, End: 115}
	repair := Repair{
		Kind:          "insert_take",
		Span:          &original,
		Replacement:   "take buffer",
		Applicability: ApplicabilityMachineApplicable,
	}
	before := ErrorWithRepairs(
		"ownership.transfer_requires_take",
		Span{Start: 109, End: 115},
		"noncopyable binding requires explicit take",
		[]Cause{{Kind: "declared_here", Span: &Span{Start: 63, End: 69}}},
		repair,
	)

	// Simulate an edit to a line above this diagnostic: every span in the
	// repair shifts by a fixed offset. This is the exact failure D-06-24
	// exists to prevent.
	const offset = 40
	shifted := Span{Start: original.Start + offset, End: original.End + offset}
	shiftedRepair := repair
	shiftedRepair.Span = &shifted
	after := ErrorWithRepairs(
		"ownership.transfer_requires_take",
		Span{Start: 109, End: 115},
		"noncopyable binding requires explicit take",
		[]Cause{{Kind: "declared_here", Span: &Span{Start: 63, End: 69}}},
		shiftedRepair,
	)

	if before.ID != after.ID {
		t.Fatalf("diagnostic ID moved after a repair span shift: before=%q after=%q", before.ID, after.ID)
	}
}

// TestRepairFieldEnumerationIsExhaustive reflects over Repair and compares
// the sorted field-name slice against a literal expected slice. If you are
// reading this because the test failed: you added a field to Repair without
// deciding whether it is identity-bearing. Decide explicitly, then either
// (a) leave it out of ErrorWithRepairs's identity struct and repairKinds
// loop (non-identity-bearing, like Span/Replacement/Applicability), or
// (b) fold it into the identity struct deliberately, understanding that
// doing so changes every diagnostic ID that carries a repair with this field
// set. Only then update the expected field list below.
func TestRepairFieldEnumerationIsExhaustive(t *testing.T) {
	expected := []string{"Applicability", "Detail", "Kind", "Replacement", "Span"}

	typ := reflect.TypeOf(Repair{})
	got := make([]string, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		got[i] = typ.Field(i).Name
	}
	sort.Strings(got)

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Repair field set changed: got %v, want %v -- see test doc comment: classify the new field as identity-bearing or not before updating this list", got, expected)
	}
}
