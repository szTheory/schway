package ast

import "github.com/codename-lang/lang/internal/compiler/diagnostic"

type Program struct {
	Module  string
	Exports []Export
	Data    []DataDecl
	Funcs   []FuncDecl
}

type Export struct {
	Kind string
	Name string
	Span diagnostic.Span
}

type DataDecl struct {
	Name         string
	Alternatives []Alternative
	Span         diagnostic.Span
}

type Alternative struct {
	Name string
	Span diagnostic.Span
}

type FuncDecl struct {
	Name         string
	Parameter    Parameter
	ReturnOrigin *BorrowOrigin
	ReturnType   TypeRef
	Body         Body
	Span         diagnostic.Span
}

// BorrowOrigin is the Phase 3 return-type annotation `borrow(path)` /
// `borrow mut(path)` preceding an ordinary return TypeRef. Its presence is
// the syntactic discriminant for the borrowed-view return case (OWN-04):
// exactly one path is supported this phase (the function's own parameter
// name — the language has a single parameter and no field-path-bearing
// executable shape), and Access names the declared access mode.
type BorrowOrigin struct {
	Path   string
	Access string // "shared" | "exclusive"
	Span   diagnostic.Span
}

type Parameter struct {
	Name string
	Type TypeRef
	Span diagnostic.Span
}

type TypeRef struct {
	Constructor string
	Arguments   []TypeRef
}

type Body struct {
	MatchExpr
	Linear *LinearBody
}

func (b Body) HasMatch() bool { return b.MatchExpr.Scrutinee != "" }

func (b Body) HasClosedVariant() bool { return b.HasMatch() != (b.Linear != nil) }

type MatchExpr struct {
	Scrutinee string
	Arms      []MatchArm
	Span      diagnostic.Span
}

type MatchArm struct {
	Pattern string
	Value   string
	// Body is the Phase 3 extension: an arm's value position may hold a full
	// linear body instead of a bare alternative name. Exactly one of Value
	// and Body is populated — see HasClosedVariant, the arm-level analog of
	// Body.HasClosedVariant (03-PATTERNS inconsistency I-7).
	Body *LinearBody
	Span diagnostic.Span
}

// HasClosedVariant reports whether exactly one of the arm's two value forms
// (a bare alternative name, or a full linear body) is populated.
func (a MatchArm) HasClosedVariant() bool { return (a.Value != "") != (a.Body != nil) }

type LinearBody struct {
	Bindings []Binding
	Result   string
	Span     diagnostic.Span
}

type Binding struct {
	Name string
	RHS  RHS
	Span diagnostic.Span
}

type RHS struct {
	Kind   string
	Source string
	Span   diagnostic.Span
}
