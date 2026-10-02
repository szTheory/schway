package syntax

import (
	"bytes"
	"strings"
)

// Format owns the single Phase 1 source projection. It is intentionally a
// token/CST projection: comments remain data and semantic tokens never move.
func Format(tree Tree) []byte {
	f := formatter{contexts: make([]string, 0, 3)}
	tokens := make([]Token, 0, len(tree.Tokens))
	for _, token := range tree.Tokens {
		if token.Kind == TokenEOF || token.Kind == TokenWhitespace {
			continue
		}
		tokens = append(tokens, token)
	}
	for index, token := range tokens {
		next := TokenEOF
		if index+1 < len(tokens) {
			next = tokens[index+1].Kind
		}
		if token.Kind == TokenIdentifier && next == TokenEqual && index > 0 && f.lineOpen && (f.context() == "function" || f.context() == "arm") {
			previous := tokens[index-1]
			if previous.Span.End <= token.Span.Start && bytes.Contains(tree.Source[previous.Span.End:token.Span.Start], []byte("\n")) {
				f.newline()
			}
		}
		f.token(token, next)
	}
	return append(bytes.TrimRight([]byte(f.out.String()), " \t\r\n"), '\n')
}

type formatter struct {
	out           strings.Builder
	indent        int
	lineOpen      bool
	previous      Kind
	contexts      []string
	linearBinding bool
	// tryCall is true between a `try` keyword and its call's closing paren,
	// so that closing paren (and only that one -- never a function's own
	// parameter-list paren) breaks the line the same way a plain binding's
	// source identifier does.
	tryCall bool
	// callBinding is D-07-01/D-07-40's bare-call analogue of tryCall: true
	// between a bare call's callee identifier (`let r = g(x)`, no `try`) and
	// its closing paren, so that closing paren breaks the line the same way
	// tryCall's does -- without it, the callee identifier would end its own
	// line prematurely (the same rule a plain binding's source identifier
	// follows), leaving the argument list stranded on the next line.
	callBinding bool
	// header records the declaration keyword that opened the line currently
	// being written, so an opening brace is classified by the construct it
	// belongs to rather than by whichever token happens to precede it. The
	// preceding token is not a sound classifier: a generic return type ends
	// the header with `>` (TokenRAngle), not an identifier.
	header string
	// discardRationale is true between a `discard ... because` statement's
	// `because` keyword and its rationale string, so that string's closing
	// quote breaks the line the same way a try-call's closing paren does --
	// the next statement (another `discard`, a `let`, or the linear body's
	// final bare Result identifier) starts fresh.
	discardRationale bool
	// pendingDefectReason is true between a `defect` keyword and its reason
	// string, mirroring discardRationale's shape (no intervening keyword here,
	// so this flag is set directly by the TokenDefect case).
	pendingDefectReason bool
}

func (f *formatter) token(token Token, next Kind) {
	switch token.Kind {
	case TokenComment:
		if f.lineOpen {
			f.out.WriteByte(' ')
		} else {
			f.writeIndent()
		}
		f.out.WriteString(token.Text)
		// A comment is the one token that breaks a line without ending the
		// construct that opened it, so a header interrupted by a trailing
		// comment continues on the next line and must keep its classification.
		header := f.header
		f.newline()
		f.header = header
	case TokenModule:
		f.blankBeforeTopLevel()
		f.write("module ")
	case TokenExport:
		f.blankBeforeTopLevel()
		f.write("export ")
	case TokenData:
		f.blankBeforeTopLevel()
		f.write("data ")
	case TokenFn:
		if f.context() == "export" {
			f.ensureLine()
		} else {
			f.blankBeforeTopLevel()
		}
		f.write("fn ")
		if f.context() == "foreign" {
			f.header = "foreign_fn"
		} else {
			f.header = "function"
		}
	case TokenForeign:
		f.blankBeforeTopLevel()
		f.write("foreign ")
		f.header = "foreign"
	case TokenTry:
		f.write("try ")
		f.tryCall = true
	case TokenType:
		f.ensureLine()
		f.write("type ")
	case TokenMatch:
		f.ensureLine()
		f.write("match ")
		f.header = "match"
	case TokenLet:
		if f.lineOpen {
			f.newline()
		}
		f.write("let ")
		f.linearBinding = true
	case TokenVar:
		if f.lineOpen {
			f.newline()
		}
		f.write("var ")
	case TokenIf, TokenWhile:
		if f.lineOpen {
			f.newline()
		}
		f.write(token.Text + " ")
		f.header = token.Text
	case TokenElse:
		if f.lineOpen {
			f.newline()
		}
		f.write("else ")
		f.header = "else"
	case TokenDiscard:
		if f.lineOpen {
			f.newline()
		}
		f.write("discard ")
	case TokenBecause:
		f.trimSpace()
		f.out.WriteString(" because ")
		f.lineOpen = true
		f.discardRationale = true
	case TokenDefect:
		if f.lineOpen {
			f.newline()
		}
		f.write("defect ")
		f.pendingDefectReason = true
	case TokenTake:
		f.write("take ")
	case TokenBorrow:
		f.write("borrow ")
	case TokenMut:
		f.write("mut ")
	case TokenIdentifier:
		wasOpen := f.lineOpen
		f.ensureLine()
		// A return type identifier immediately follows the borrow-origin
		// annotation's closing paren (`borrow(path) Type`). The parameter-list
		// RParen is always followed by TokenArrow, never an identifier
		// directly, so this is an unambiguous signal for the annotation case.
		// Guarded on wasOpen: a try-call's closing paren already broke the
		// line (see TokenRParen above), so the next identifier -- the linear
		// body's own Result -- starts a fresh, unindented-by-a-space line
		// rather than continuing this one.
		if f.previous == TokenRParen && wasOpen {
			f.out.WriteByte(' ')
		}
		f.out.WriteString(token.Text)
		f.lineOpen = true
		if f.context() == "export" && (f.previous == TokenType || f.previous == TokenFn) {
			f.newline()
		} else if f.previous == TokenPipe || f.previous == TokenFatArrow {
			f.newline()
		} else if (f.context() == "function" || f.context() == "arm") && f.linearBinding && f.previous == TokenEqual && next == TokenLParen {
			// D-07-01/D-07-40: this identifier is a bare call's callee, not
			// a plain binding source -- its own argument list still follows
			// on this line, so (unlike the branch below) this must NOT end
			// the line here. See callBinding's doc comment.
			f.callBinding = true
		} else if (f.context() == "function" || f.context() == "arm") && f.linearBinding && (f.previous == TokenEqual || f.previous == TokenTake || f.previous == TokenBorrow || f.previous == TokenMut) {
			f.newline()
			f.linearBinding = false
		} else if f.context() == "foreign_fn" && f.previous == TokenColon {
			// A policy value (`unwind: forbidden`) ends its own line, the
			// same "write value, then break" shape TokenPipe's alternative
			// identifiers use above.
			f.newline()
		}
	case TokenString:
		f.ensureLine()
		f.out.WriteString(token.Text)
		f.lineOpen = true
		if f.context() == "foreign_fn" && f.previous == TokenColon {
			f.newline()
		} else if f.discardRationale && (f.context() == "function" || f.context() == "arm") {
			f.discardRationale = false
			f.newline()
		} else if f.pendingDefectReason && (f.context() == "function" || f.context() == "arm") {
			f.pendingDefectReason = false
			f.newline()
		}
	case TokenNumber:
		f.ensureLine()
		f.out.WriteString(token.Text)
		f.lineOpen = true
		if (f.context() == "function" || f.context() == "arm") && f.linearBinding && f.previous == TokenEqual {
			f.newline()
			f.linearBinding = false
		}
	case TokenDot:
		f.out.WriteByte('.')
		f.lineOpen = true
	case TokenLBrace:
		context := "block"
		if f.previous == TokenExport {
			context = "export"
		} else if f.header != "" {
			context = f.header
		}
		if f.lineOpen {
			f.trimSpace()
			f.out.WriteString(" {\n")
		} else {
			// The header was broken by a trailing comment, so the brace opens
			// its own line and must not inherit a separator space.
			f.writeIndent()
			f.out.WriteString("{\n")
		}
		f.lineOpen = false
		f.header = ""
		f.indent++
		f.contexts = append(f.contexts, context)
	case TokenRBrace:
		if f.lineOpen {
			f.newline()
		}
		if f.indent > 0 {
			f.indent--
		}
		closed := f.context()
		if len(f.contexts) > 0 {
			f.contexts = f.contexts[:len(f.contexts)-1]
		}
		f.writeIndent()
		f.out.WriteByte('}')
		f.lineOpen = true
		f.newline()
		if closed == "export" {
			f.newline()
		}
	case TokenLParen:
		// A borrow-origin annotation's opening paren immediately follows
		// `borrow`/`mut` with no space (`borrow(path)`, `borrow mut(path)`).
		// The parameter-list `(` always follows an identifier (the function
		// name), never TokenBorrow/TokenMut, so this is unambiguous.
		if f.previous == TokenBorrow || f.previous == TokenMut {
			f.trimSpace()
		}
		f.out.WriteByte('(')
		f.lineOpen = true
	case TokenRParen:
		f.trimSpace()
		f.out.WriteByte(')')
		f.lineOpen = true
		if f.tryCall {
			f.tryCall = false
			if f.context() == "function" || f.context() == "arm" {
				f.newline()
			}
		} else if f.callBinding {
			f.callBinding = false
			if f.context() == "function" || f.context() == "arm" {
				f.newline()
			}
		}
	case TokenColon:
		f.trimSpace()
		f.out.WriteString(": ")
		f.lineOpen = true
	case TokenEqual:
		f.trimSpace()
		if f.context() == "function" || f.context() == "arm" {
			f.out.WriteString(" = ")
			f.lineOpen = true
		} else {
			f.out.WriteString(" =")
			f.newline()
			f.indent++
			f.contexts = append(f.contexts, "data")
		}
	case TokenPlus:
		f.trimSpace()
		f.out.WriteString(" + ")
		f.lineOpen = true
	case TokenPipe:
		f.ensureLine()
		f.write("| ")
	case TokenArrow:
		f.trimSpace()
		f.out.WriteString(" -> ")
		f.lineOpen = true
	case TokenFatArrow:
		f.trimSpace()
		f.out.WriteString(" => ")
		f.lineOpen = true
		// A brace directly following => opens an arm body, not a bare
		// generic block: classify it the same way a function or match
		// header does, so its bindings format on their own lines. A bare
		// arm value (an identifier) clears this via the FatArrow branch
		// in the TokenIdentifier case above before any brace is seen.
		f.header = "arm"
	case TokenLAngle:
		f.trimSpace()
		if f.header == "if" || f.header == "while" {
			f.out.WriteString(" < ")
		} else {
			f.out.WriteByte('<')
		}
		f.lineOpen = true
	case TokenRAngle:
		f.trimSpace()
		f.out.WriteByte('>')
		f.lineOpen = true
	case TokenComma:
		f.trimSpace()
		f.out.WriteString(", ")
		f.lineOpen = true
	default:
		f.ensureLine()
		f.out.WriteString(token.Text)
		f.lineOpen = true
	}
	if token.Kind != TokenWhitespace && token.Kind != TokenComment {
		f.previous = token.Kind
	}
}

func (f *formatter) context() string {
	if len(f.contexts) == 0 {
		return ""
	}
	return f.contexts[len(f.contexts)-1]
}

func (f *formatter) currentLine() string {
	value := f.out.String()
	if index := strings.LastIndexByte(value, '\n'); index >= 0 {
		return value[index+1:]
	}
	return value
}

func (f *formatter) blankBeforeTopLevel() {
	if f.context() == "data" {
		f.indent--
		f.contexts = f.contexts[:len(f.contexts)-1]
	}
	if f.lineOpen {
		f.newline()
	}
	value := f.out.String()
	if value != "" && !strings.HasSuffix(value, "\n\n") {
		f.out.WriteByte('\n')
		f.lineOpen = false
	}
	f.ensureLine()
}

func (f *formatter) ensureLine() {
	if !f.lineOpen {
		f.writeIndent()
	}
}

func (f *formatter) writeIndent() {
	if f.lineOpen {
		return
	}
	f.out.WriteString(strings.Repeat("  ", f.indent))
	f.lineOpen = true
}

func (f *formatter) write(value string) {
	f.ensureLine()
	f.out.WriteString(value)
	f.lineOpen = true
}

func (f *formatter) newline() {
	// Ending a line ordinarily ends the declaration header it carried. The one
	// exception is a trailing comment, which breaks the line mid-header and
	// restores the classification itself.
	f.header = ""
	value := f.out.String()
	if strings.HasSuffix(value, "\n") {
		f.lineOpen = false
		return
	}
	f.trimSpace()
	f.out.WriteByte('\n')
	f.lineOpen = false
}

func (f *formatter) trimSpace() {
	value := f.out.String()
	trimmed := strings.TrimRight(value, " \t")
	if len(trimmed) == len(value) {
		return
	}
	f.out.Reset()
	f.out.WriteString(trimmed)
}
