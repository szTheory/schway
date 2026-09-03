package diagnostic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

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
	ID       string  `json:"id"`
	Code     string  `json:"code"`
	Severity string  `json:"severity"`
	Primary  Span    `json:"primary_span"`
	Message  string  `json:"message"`
	Causes   []Cause `json:"causes,omitempty"`
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

func (d Diagnostic) String() string {
	return fmt.Sprintf("%s [%d:%d]: %s", d.Code, d.Primary.Start, d.Primary.End, d.Message)
}
