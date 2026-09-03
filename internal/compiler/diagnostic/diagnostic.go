package diagnostic

import "fmt"

const Schema = "lang.diagnostic/0"

type Span struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type Cause struct {
	Kind   string `json:"kind"`
	Detail string `json:"detail,omitempty"`
	Span   *Span  `json:"span,omitempty"`
}

type Diagnostic struct {
	Schema   string  `json:"schema"`
	Code     string  `json:"code"`
	Severity string  `json:"severity"`
	Primary  Span    `json:"primary_span"`
	Message  string  `json:"message"`
	Causes   []Cause `json:"causes,omitempty"`
}

func Error(code string, span Span, message string, causes ...Cause) Diagnostic {
	return Diagnostic{Schema: Schema, Code: code, Severity: "error", Primary: span, Message: message, Causes: causes}
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("%s [%d:%d]: %s", d.Code, d.Primary.Start, d.Primary.End, d.Message)
}
