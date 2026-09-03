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
	TokenUnknown    Kind = "unknown"
)

type Token struct {
	Kind Kind
	Text string
	Span diagnostic.Span
}

func (t Token) Trivia() bool { return t.Kind == TokenWhitespace || t.Kind == TokenComment }
