package syntax

import (
	"unicode"
	"unicode/utf8"

	"github.com/codename-lang/lang/internal/compiler/diagnostic"
)

var keywords = map[string]Kind{
	"module":  TokenModule,
	"export":  TokenExport,
	"type":    TokenType,
	"data":    TokenData,
	"fn":      TokenFn,
	"match":   TokenMatch,
	"foreign": TokenForeign,
	"try":     TokenTry,
	"discard": TokenDiscard,
	"because": TokenBecause,
}

func Lex(source []byte) ([]Token, []diagnostic.Diagnostic) {
	tokens, diagnostics, truncated := lex(source)
	if truncated {
		diagnostics = append(diagnostics, tooManyErrors(len(source)))
	}
	return tokens, diagnostics
}

func lex(source []byte) ([]Token, []diagnostic.Diagnostic, bool) {
	tokens := make([]Token, 0, len(source)/3)
	diagnostics := make([]diagnostic.Diagnostic, 0)
	truncated := false
	problem := func(value diagnostic.Diagnostic) {
		if len(diagnostics) < maxPrimaryDiagnostics {
			diagnostics = append(diagnostics, value)
		} else {
			truncated = true
		}
	}
	for offset := 0; offset < len(source); {
		if len(tokens) >= MaxTokens {
			tokens = append(tokens, Token{Kind: TokenEOF, Span: diagnostic.Span{Start: offset, End: offset}})
			return tokens, []diagnostic.Diagnostic{inputLimit(offset)}, false
		}
		start := offset
		if source[offset] == '/' && offset+1 < len(source) && source[offset+1] == '/' {
			offset += 2
			for offset < len(source) && source[offset] != '\n' {
				offset++
			}
			tokens = append(tokens, Token{Kind: TokenComment, Text: string(source[start:offset]), Span: diagnostic.Span{Start: start, End: offset}})
			continue
		}

		if source[offset] == '"' {
			// Phase 4's only string-literal use is the allocator identity
			// (`allocator: "libc_malloc"`); no escape sequences are
			// supported this phase -- an unterminated or newline-crossing
			// literal is a lexer-level rejection, not a parser recovery.
			end := offset + 1
			closed := false
			for end < len(source) && source[end] != '\n' {
				if source[end] == '"' {
					end++
					closed = true
					break
				}
				end++
			}
			if !closed {
				tokens = append(tokens, Token{Kind: TokenUnknown, Text: string(source[offset:end]), Span: diagnostic.Span{Start: offset, End: end}})
				problem(diagnostic.Error("syntax.unterminated_string", diagnostic.Span{Start: offset, End: end}, "string literal is not terminated"))
				offset = end
				continue
			}
			tokens = append(tokens, Token{Kind: TokenString, Text: string(source[offset:end]), Span: diagnostic.Span{Start: offset, End: end}})
			offset = end
			continue
		}

		r, size := utf8.DecodeRune(source[offset:])
		if r == utf8.RuneError && size == 1 {
			offset++
			tokens = append(tokens, Token{Kind: TokenUnknown, Text: string(source[start:offset]), Span: diagnostic.Span{Start: start, End: offset}})
			problem(diagnostic.Error("syntax.invalid_utf8", diagnostic.Span{Start: start, End: offset}, "source is not valid UTF-8"))
			continue
		}
		if unicode.IsSpace(r) {
			offset += size
			for offset < len(source) {
				next, nextSize := utf8.DecodeRune(source[offset:])
				if next == utf8.RuneError && nextSize == 1 || !unicode.IsSpace(next) {
					break
				}
				offset += nextSize
			}
			tokens = append(tokens, Token{Kind: TokenWhitespace, Text: string(source[start:offset]), Span: diagnostic.Span{Start: start, End: offset}})
			continue
		}
		if unicode.IsLetter(r) || r == '_' {
			offset += size
			for offset < len(source) {
				next, nextSize := utf8.DecodeRune(source[offset:])
				if !(unicode.IsLetter(next) || unicode.IsDigit(next) || next == '_') {
					break
				}
				offset += nextSize
			}
			text := string(source[start:offset])
			kind := TokenIdentifier
			if keyword, ok := keywords[text]; ok {
				kind = keyword
			}
			tokens = append(tokens, Token{Kind: kind, Text: text, Span: diagnostic.Span{Start: start, End: offset}})
			continue
		}

		kind, width := punctuation(source[offset:])
		if width == 0 {
			width = size
			kind = TokenUnknown
			problem(diagnostic.Error("syntax.unexpected_byte", diagnostic.Span{Start: start, End: start + width}, "unexpected source character"))
		}
		offset += width
		tokens = append(tokens, Token{Kind: kind, Text: string(source[start:offset]), Span: diagnostic.Span{Start: start, End: offset}})
	}
	tokens = append(tokens, Token{Kind: TokenEOF, Span: diagnostic.Span{Start: len(source), End: len(source)}})
	return tokens, diagnostics, truncated
}

func punctuation(source []byte) (Kind, int) {
	if len(source) >= 2 {
		switch string(source[:2]) {
		case "->":
			return TokenArrow, 2
		case "=>":
			return TokenFatArrow, 2
		}
	}
	switch source[0] {
	case '{':
		return TokenLBrace, 1
	case '}':
		return TokenRBrace, 1
	case '(':
		return TokenLParen, 1
	case ')':
		return TokenRParen, 1
	case ':':
		return TokenColon, 1
	case '.':
		return TokenDot, 1
	case '|':
		return TokenPipe, 1
	case '=':
		return TokenEqual, 1
	default:
		return TokenUnknown, 0
	}
}
