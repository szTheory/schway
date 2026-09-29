package diagnostic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

const (
	Schema  = "schway.diagnostic/0"
	Schema1 = "schway.diagnostic/1"
)

type Span struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type Cause struct {
	Kind   string `json:"kind"`
	Detail string `json:"detail,omitempty"`
	Span   *Span  `json:"span,omitempty"`
}

type Repair struct {
	Kind   string `json:"kind"`
	Detail string `json:"detail,omitempty"`
	// Span, Replacement, and Applicability grow a Repair from a mere
	// classification into an applicable edit (D-06-24). All three are
	// deliberately NON-identity-bearing: see the repairKinds comment in
	// ErrorWithRepairs for why a coordinate shift must never move a
	// diagnostic's ID.
	Span          *Span  `json:"span,omitempty"`
	Replacement   string `json:"replacement,omitempty"`
	Applicability string `json:"applicability,omitempty"`
}

// The closed Applicability vocabulary (D-06-24), modeled on rustc's
// Applicability enum. Only ApplicabilityMachineApplicable is driver-eligible;
// everything else routes to the recorded, non-gating agent exercise
// (D-06-23, D-06-30).
const (
	ApplicabilityMachineApplicable    = "MachineApplicable"
	ApplicabilityRequiresConfirmation = "RequiresConfirmation"
	ApplicabilityUnspecified          = "Unspecified"
)

// Applicabilities returns the closed set of valid Repair.Applicability
// values.
func Applicabilities() []string {
	return []string{ApplicabilityMachineApplicable, ApplicabilityRequiresConfirmation, ApplicabilityUnspecified}
}

// ValidateRepair refuses any Applicability value outside the closed
// vocabulary. An empty Applicability is treated as unset and is valid here;
// callers that want the normalized value should use NormalizeApplicability.
func ValidateRepair(r Repair) error {
	if r.Applicability == "" {
		return nil
	}
	for _, valid := range Applicabilities() {
		if r.Applicability == valid {
			return nil
		}
	}
	return fmt.Errorf("diagnostic: invalid repair applicability %q", r.Applicability)
}

// NormalizeApplicability maps an empty Applicability to
// ApplicabilityUnspecified. It never defaults to ApplicabilityMachineApplicable:
// defaulting toward driver-eligible is exactly the fail-open shape this
// project refuses.
func NormalizeApplicability(applicability string) string {
	if applicability == "" {
		return ApplicabilityUnspecified
	}
	return applicability
}

// DriverEligible reports whether a repair may be applied mechanically by the
// repair driver (06-13). A repair is eligible only when it declares
// MachineApplicable, carries both a span and replacement text to act on, AND
// carries a non-empty Kind — a repair that claims machine-applicability
// without the material to apply it, or without even naming what it is, is
// not eligible; treating either as eligible would be the exact fail-open
// shape this project refuses (D-06-27.2's structured-vocabulary-removal
// guard depends on Kind mattering here, not merely as reporting metadata).
// Everything not eligible (RequiresConfirmation, Unspecified, empty, or
// incomplete MachineApplicable) routes to the recorded, non-gating agent
// exercise (D-06-23, D-06-30).
func DriverEligible(r Repair) bool {
	return r.Applicability == ApplicabilityMachineApplicable && r.Kind != "" && r.Span != nil && r.Replacement != ""
}

type Diagnostic struct {
	Schema   string   `json:"schema"`
	ID       string   `json:"id"`
	Code     string   `json:"code"`
	Severity string   `json:"severity"`
	Primary  Span     `json:"primary_span"`
	Message  string   `json:"message"`
	Causes   []Cause  `json:"causes,omitempty"`
	Repairs  []Repair `json:"repairs,omitempty"`
}

func Error(code string, span Span, message string, causes ...Cause) Diagnostic {
	identity := struct {
		Schema string
		Code   string
		Span   Span
		Causes []Cause
	}{Schema: Schema, Code: code, Span: span, Causes: causes}
	encoded, _ := json.Marshal(identity)
	sum := sha256.Sum256(encoded)
	return Diagnostic{Schema: Schema, ID: "diagnostic:" + hex.EncodeToString(sum[:12]), Code: code, Severity: "error", Primary: span, Message: message, Causes: causes}
}

// ErrorWithRepairs constructs the feature-specific diagnostic/1 record used by
// ownership checking. Repair order is canonical, and only repair kinds—not
// display detail or message prose—participate in semantic identity.
func ErrorWithRepairs(code string, span Span, message string, causes []Cause, repairs ...Repair) Diagnostic {
	repairs = append([]Repair(nil), repairs...)
	sort.Slice(repairs, func(left, right int) bool {
		if repairs[left].Kind == repairs[right].Kind {
			return repairs[left].Detail < repairs[right].Detail
		}
		return repairs[left].Kind < repairs[right].Kind
	})
	// Span, Replacement, and Applicability are deliberately excluded from
	// repairKinds: a coordinate shift (an edit to a line above this
	// diagnostic) must never change a diagnostic's published ID (D-06-24).
	// Only Kind participates in semantic identity.
	repairKinds := make([]string, len(repairs))
	for index, repair := range repairs {
		repairKinds[index] = repair.Kind
	}
	identity := struct {
		Schema      string
		Code        string
		Span        Span
		Causes      []Cause
		RepairKinds []string
	}{Schema: Schema1, Code: code, Span: span, Causes: causes, RepairKinds: repairKinds}
	encoded, _ := json.Marshal(identity)
	sum := sha256.Sum256(encoded)
	return Diagnostic{
		Schema: Schema1, ID: "diagnostic:" + hex.EncodeToString(sum[:12]), Code: code,
		Severity: "error", Primary: span, Message: message, Causes: causes, Repairs: repairs,
	}
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("%s [%d:%d]: %s", d.Code, d.Primary.Start, d.Primary.End, d.Message)
}
