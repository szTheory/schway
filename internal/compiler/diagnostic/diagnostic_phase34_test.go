package diagnostic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type phase34IdentityFixture struct {
	IdentityCases []struct {
		Name         string `json:"name"`
		DiagnosticID string `json:"diagnostic_id"`
		Marshaled    string `json:"marshaled_json"`
	} `json:"identity_cases"`
}

func TestPhase34DiagnosticIdentityProjection(t *testing.T) {
	baselineBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "phase34", "identity-baseline.json"))
	if err != nil {
		t.Fatal(err)
	}
	var baseline phase34IdentityFixture
	if err := json.Unmarshal(baselineBytes, &baseline); err != nil {
		t.Fatal(err)
	}

	span := Span{Start: 40, End: 45}
	cause := Cause{Kind: "declared_here", Detail: "binding origin", Span: &Span{Start: 3, End: 8}}
	insert := Repair{Kind: "insert_take", Detail: "ownership", Span: &Span{Start: 12, End: 18}, Replacement: "take false", Applicability: ApplicabilityMachineApplicable}
	borrow := Repair{Kind: "replace_borrow", Detail: "borrow", Span: &Span{Start: 28, End: 34}, Replacement: "borrow false", Applicability: ApplicabilityRequiresConfirmation}
	legacyRepair := func(repair Repair) Diagnostic {
		return ErrorWithRepairs("compat.identity", span, "legacy repair diagnostic", nil, repair, borrow)
	}
	fixtures := map[string]Diagnostic{
		"error-nil-causes":                       Error("compat.legacy", Span{Start: 1, End: 2}, "legacy error without causes"),
		"error-nested-cause-span":                Error("compat.legacy", Span{Start: 9, End: 10}, "legacy error with cause", Cause{Kind: "nested_origin", Detail: "declared", Span: &Span{Start: 3, End: 8}}),
		"error-with-repairs-nil-causes":          legacyRepair(insert),
		"error-with-repairs-nested-causes":       ErrorWithRepairs("compat.identity", span, "legacy repair diagnostic", []Cause{cause}, insert, borrow),
		"error-with-repairs-shifted-coordinates": ErrorWithRepairs("compat.identity", span, "legacy repair diagnostic", []Cause{cause}, Repair{Kind: "insert_take", Detail: "ownership", Span: &Span{Start: 112, End: 118}, Replacement: "take false", Applicability: ApplicabilityMachineApplicable}, Repair{Kind: "replace_borrow", Detail: "borrow", Span: &Span{Start: 128, End: 134}, Replacement: "borrow false", Applicability: ApplicabilityRequiresConfirmation}),
		"error-with-repairs-different-kind":      ErrorWithRepairs("compat.identity", span, "legacy repair diagnostic", []Cause{cause}, insert, Repair{Kind: "different_repair_kind", Detail: "borrow", Span: &Span{Start: 28, End: 34}, Replacement: "borrow false", Applicability: ApplicabilityRequiresConfirmation}),
		"in-unit-source":                         ErrorWithRepairs("compat.identity", span, "legacy repair diagnostic", []Cause{cause}, insert, borrow).InUnit("draft.identity"),
		"in-unit-other":                          ErrorWithRepairs("compat.identity", span, "legacy repair diagnostic", []Cause{cause}, insert, borrow).InUnit("other.identity"),
	}
	if len(baseline.IdentityCases) != len(fixtures) {
		t.Fatalf("identity fixture count = %d, want %d", len(baseline.IdentityCases), len(fixtures))
	}
	for _, expected := range baseline.IdentityCases {
		t.Run(expected.Name, func(t *testing.T) {
			got, ok := fixtures[expected.Name]
			if !ok {
				t.Fatalf("no diagnostic constructor for frozen case %q", expected.Name)
			}
			if got.ID != expected.DiagnosticID {
				t.Fatalf("ID = %q, want frozen %q", got.ID, expected.DiagnosticID)
			}
			encoded, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if string(encoded) != expected.Marshaled {
				t.Fatalf("legacy diagnostic JSON changed:\n got: %s\nwant: %s", encoded, expected.Marshaled)
			}
			marked := got.AtSource()
			marked.Primary.Line, marked.Primary.Col = 1, 1
			if marked.ID != expected.DiagnosticID {
				t.Fatalf("source provenance or derived coordinates changed frozen ID: %q", marked.ID)
			}
		})
	}

	// Test-only mutation control: this models the tempting regression of
	// hashing the public Span, including its derived coordinates. The frozen
	// ID must detect that the control is materially different.
	base := fixtures["error-nil-causes"]
	mutated := base
	mutated.Primary.Line, mutated.Primary.Col = 1, 1
	type coordinateBearingIdentity struct {
		Schema string  `json:"Schema"`
		Code   string  `json:"Code"`
		Span   Span    `json:"Span"`
		Causes []Cause `json:"Causes"`
	}
	encodedMutation, err := json.Marshal(coordinateBearingIdentity{Schema: base.Schema, Code: base.Code, Span: mutated.Primary, Causes: base.Causes})
	if err != nil {
		t.Fatal(err)
	}
	mutationHash := sha256.Sum256(encodedMutation)
	mutationID := "diagnostic:" + hex.EncodeToString(mutationHash[:12])
	if mutationID == base.ID {
		t.Fatalf("coordinates-in-identity negative control was inert: %s", mutationID)
	}
}
