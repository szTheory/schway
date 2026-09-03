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
	Name       string
	Parameter  Parameter
	ReturnType TypeRef
	Body       Body
	Span       diagnostic.Span
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
	Span    diagnostic.Span
}

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
