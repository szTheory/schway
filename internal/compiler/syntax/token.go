package syntax

import "github.com/codename-lang/lang/internal/compiler/diagnostic"

type Kind string

const (
	TokenEOF        Kind = "eof"
	TokenWhitespace Kind = "whitespace"
	TokenComment    Kind = "comment"
	TokenIdentifier Kind = "identifier"
	TokenModule     Kind = "module"
	TokenExport     Kind = "export"
	TokenType       Kind = "type"
	TokenData       Kind = "data"
	TokenFn         Kind = "fn"
	TokenMatch      Kind = "match"
	TokenLet        Kind = "let"
	TokenTake       Kind = "take"
	TokenBorrow     Kind = "borrow"
	TokenMut        Kind = "mut"
	TokenForeign    Kind = "foreign"
	TokenTry        Kind = "try"
	TokenDiscard    Kind = "discard"
	TokenBecause    Kind = "because"
	// TokenDefect is Phase 4's terminal defect keyword (D-04-15): admissible
	// only in a linear body's terminal position, as an alternative to a bare
	// Result identifier.
	TokenDefect Kind = "defect"
	TokenString     Kind = "string"
	TokenLBrace     Kind = "{"
	TokenRBrace     Kind = "}"
	TokenLParen     Kind = "("
	TokenRParen     Kind = ")"
	TokenColon      Kind = ":"
	TokenDot        Kind = "."
	TokenPipe       Kind = "|"
	TokenEqual      Kind = "="
	TokenArrow      Kind = "->"
	TokenFatArrow   Kind = "=>"
	TokenLAngle     Kind = "<"
	TokenRAngle     Kind = ">"
	TokenComma      Kind = ","
	TokenUnknown    Kind = "unknown"
)

type Token struct {
	Kind Kind
	Text string
	Span diagnostic.Span
}

func (t Token) Trivia() bool { return t.Kind == TokenWhitespace || t.Kind == TokenComment }
