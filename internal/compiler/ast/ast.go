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
	ReturnType string
	Body       MatchExpr
	Span       diagnostic.Span
}

type Parameter struct {
	Name string
	Type string
	Span diagnostic.Span
}

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
