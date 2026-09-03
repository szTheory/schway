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
}

type Repair struct {
	Kind   string `json:"kind"`
	Detail string `json:"detail,omitempty"`
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
