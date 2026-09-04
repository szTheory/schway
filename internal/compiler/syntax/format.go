package syntax

import (
	"bytes"
	"strings"
)

// Format owns the single Phase 1 source projection. It is intentionally a
// token/CST projection: comments remain data and semantic tokens never move.
func Format(tree Tree) []byte {
	f := formatter{contexts: make([]string, 0, 3)}
	for _, token := range tree.Tokens {
		if token.Kind == TokenEOF || token.Kind == TokenWhitespace {
			continue
		}
		f.token(token)
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
	// header records the declaration keyword that opened the line currently
	// being written, so an opening brace is classified by the construct it
	// belongs to rather than by whichever token happens to precede it. The
	// preceding token is not a sound classifier: a generic return type ends
	// the header with `>` (TokenRAngle), not an identifier.
	header string
}

func (f *formatter) token(token Token) {
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
		f.header = "function"
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
	case TokenTake:
		f.write("take ")
	case TokenBorrow:
		f.write("borrow ")
	case TokenMut:
		f.write("mut ")
	case TokenIdentifier:
		f.ensureLine()
		// A return type identifier immediately follows the borrow-origin
		// annotation's closing paren (`borrow(path) Type`). The parameter-list
		// RParen is always followed by TokenArrow, never an identifier
		// directly, so this is an unambiguous signal for the annotation case.
		if f.previous == TokenRParen {
			f.out.WriteByte(' ')
		}
		f.out.WriteString(token.Text)
		f.lineOpen = true
		if f.context() == "export" && (f.previous == TokenType || f.previous == TokenFn) {
			f.newline()
		} else if f.previous == TokenPipe || f.previous == TokenFatArrow {
			f.newline()
		} else if (f.context() == "function" || f.context() == "arm") && f.linearBinding && (f.previous == TokenEqual || f.previous == TokenTake || f.previous == TokenBorrow || f.previous == TokenMut) {
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
		f.out.WriteByte('<')
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
