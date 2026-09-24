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
	"defect":  TokenDefect,
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
		if source[offset] >= '0' && source[offset] <= '9' {
			// Consume one bounded candidate before validating it so suffixes,
			// invalid radix digits, and separator errors cannot be accepted as
			// a valid numeric prefix followed by another token.
			offset++
			for offset < len(source) {
				next, width := utf8.DecodeRune(source[offset:])
				if !(isNumericCandidateByte(source[offset]) || unicode.IsLetter(next) || unicode.IsDigit(next)) {
					break
				}
				offset += width
			}
			text := string(source[start:offset])
			span := diagnostic.Span{Start: start, End: offset}
			if validNumericLiteral(text) {
				tokens = append(tokens, Token{Kind: TokenNumber, Text: text, Span: span})
			} else {
				tokens = append(tokens, Token{Kind: TokenUnknown, Text: text, Span: span})
				problem(diagnostic.Error("syntax.malformed_numeric_literal", span, "malformed numeric literal"))
			}
			continue
		}
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

func isNumericCandidateByte(value byte) bool {
	return value >= '0' && value <= '9' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value == '_'
}

func validNumericLiteral(text string) bool {
	base, digits := 10, text
	if len(text) >= 2 && text[0] == '0' {
		switch text[1] {
		case 'x', 'X':
			base, digits = 16, text[2:]
		case 'b', 'B':
			base, digits = 2, text[2:]
		case 'o', 'O':
			return false
		}
	}
	if digits == "" {
		return false
	}
	previousDigit := false
	for index := 0; index < len(digits); index++ {
		value := digits[index]
		if value == '_' {
			if !previousDigit || index+1 == len(digits) || numericDigit(digits[index+1]) < 0 || numericDigit(digits[index+1]) >= base {
				return false
			}
			previousDigit = false
			continue
		}
		digit := numericDigit(value)
		if digit < 0 || digit >= base {
			return false
		}
		previousDigit = true
	}
	return previousDigit
}

func numericDigit(value byte) int {
	switch {
	case value >= '0' && value <= '9':
		return int(value - '0')
	case value >= 'a' && value <= 'f':
		return int(value-'a') + 10
	case value >= 'A' && value <= 'F':
		return int(value-'A') + 10
	default:
		return -1
	}
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
