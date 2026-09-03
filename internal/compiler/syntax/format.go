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
	out      strings.Builder
	indent   int
	lineOpen bool
	previous Kind
	contexts []string
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
		f.newline()
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
	case TokenType:
		f.ensureLine()
		f.write("type ")
	case TokenMatch:
		f.ensureLine()
		f.write("match ")
	case TokenIdentifier:
		f.ensureLine()
		f.out.WriteString(token.Text)
		f.lineOpen = true
		if f.context() == "export" && (f.previous == TokenType || f.previous == TokenFn) {
			f.newline()
		} else if f.previous == TokenPipe || f.previous == TokenFatArrow {
			f.newline()
		}
	case TokenDot:
		f.out.WriteByte('.')
		f.lineOpen = true
	case TokenLBrace:
		context := "block"
		switch f.previous {
		case TokenExport:
			context = "export"
		case TokenIdentifier:
			if strings.Contains(f.currentLine(), "match ") {
				context = "match"
			} else {
				context = "function"
			}
		}
		f.trimSpace()
		f.out.WriteString(" {\n")
		f.lineOpen = false
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
		f.out.WriteString(" =")
		f.newline()
		f.indent++
		f.contexts = append(f.contexts, "data")
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
