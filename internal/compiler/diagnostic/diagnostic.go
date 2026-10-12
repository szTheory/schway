package diagnostic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

const (
	Schema  = "lang.diagnostic/0"
	Schema1 = "lang.diagnostic/1"
)

type Span struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type Cause struct {
	Kind   string `json:"kind"`
	Detail string `json:"detail,omitempty"`
	Span   *Span  `json:"span,omitempty"`
	// Unit names the module (dotted name, never a filesystem path) the cause
	// belongs to. Empty means the diagnostic's own unit (D-11).
	Unit string `json:"unit,omitempty"`
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
	// Unit names the module the repair's Span is relative to (D-11). Empty
	// means the diagnostic's own unit. It is NON-identity-bearing on the
	// repair: the owning diagnostic's Unit carries identity.
	Unit string `json:"unit,omitempty"`
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
	// Unit is the dotted module name the diagnostic's spans are file-local to.
	// Empty means the entry module (D-11); Span stays file-local.
	Unit string `json:"unit,omitempty"`
}

// unitOutOfIdentityForTest is a negative-control seam: when set, the unit is
// left out of the identity hash so a test can observe the collision the real
// identity prevents. Never set outside tests.
var unitOutOfIdentityForTest bool

func identityUnit(unit string) string {
	if unitOutOfIdentityForTest {
		return ""
	}
	return unit
}

func errorID(code string, span Span, causes []Cause, unit string) string {
	identity := struct {
		Schema string
		Code   string
		Span   Span
		Causes []Cause
		Unit   string `json:",omitempty"`
	}{Schema: Schema, Code: code, Span: span, Causes: causes, Unit: identityUnit(unit)}
	encoded, _ := json.Marshal(identity)
	sum := sha256.Sum256(encoded)
	return "diagnostic:" + hex.EncodeToString(sum[:12])
}

func errorWithRepairsID(code string, span Span, causes []Cause, repairKinds []string, unit string) string {
	identity := struct {
		Schema      string
		Code        string
		Span        Span
		Causes      []Cause
		RepairKinds []string
		Unit        string `json:",omitempty"`
	}{Schema: Schema1, Code: code, Span: span, Causes: causes, RepairKinds: repairKinds, Unit: identityUnit(unit)}
	encoded, _ := json.Marshal(identity)
	sum := sha256.Sum256(encoded)
	return "diagnostic:" + hex.EncodeToString(sum[:12])
}

func Error(code string, span Span, message string, causes ...Cause) Diagnostic {
	return Diagnostic{Schema: Schema, ID: errorID(code, span, causes, ""), Code: code, Severity: "error", Primary: span, Message: message, Causes: causes}
}

// InUnit attributes the diagnostic to a module and recomputes its ID from the
// unit-bearing identity, so the same code and span in two units never share an
// ID. An empty unit returns the receiver unchanged (the entry module keeps its
// pre-phase ID).
func (d Diagnostic) InUnit(unit string) Diagnostic {
	if unit == "" {
		return d
	}
	out := d
	out.Unit = unit
	if d.Schema == Schema1 {
		kinds := make([]string, len(d.Repairs))
		for index, repair := range d.Repairs {
			kinds[index] = repair.Kind
		}
		out.ID = errorWithRepairsID(d.Code, d.Primary, d.Causes, kinds, unit)
	} else {
		out.ID = errorID(d.Code, d.Primary, d.Causes, unit)
	}
	return out
}

// EffectiveUnit returns the cause's own unit, else the diagnostic's (D-11).
func (c Cause) EffectiveUnit(diagnosticUnit string) string {
	if c.Unit != "" {
		return c.Unit
	}
	return diagnosticUnit
}

// EffectiveUnit returns the repair's own unit, else the diagnostic's (D-11).
func (r Repair) EffectiveUnit(diagnosticUnit string) string {
	if r.Unit != "" {
		return r.Unit
	}
	return diagnosticUnit
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
	return Diagnostic{
		Schema: Schema1, ID: errorWithRepairsID(code, span, causes, repairKinds, ""), Code: code,
		Severity: "error", Primary: span, Message: message, Causes: causes, Repairs: repairs,
	}
}

func (d Diagnostic) String() string {
	if d.Unit != "" {
		return fmt.Sprintf("%s unit=%s [%d:%d]: %s", d.Code, d.Unit, d.Primary.Start, d.Primary.End, d.Message)
	}
	return fmt.Sprintf("%s [%d:%d]: %s", d.Code, d.Primary.Start, d.Primary.End, d.Message)
}
