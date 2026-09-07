package diagnostic

import (
	"encoding/json"
	"reflect"
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
