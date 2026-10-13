package protocol

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
)

const (
	diagnosticTabStop        = 4
	diagnosticSnippetColumns = 120
)

// CapturedSource carries the bytes and user-facing label already associated
// with one source unit in the current invocation. Unit keys are dotted module
// names; the empty key is the entry source.
type CapturedSource struct {
	Label string
	Bytes []byte
}

// PresentationContext is an invocation-local sidecar. It is omitted from
// protocol JSON and command/diagnostic identity.
type PresentationContext struct {
	Sources map[string]CapturedSource
}

// WithPresentation returns a result with a private copy of captured source
// bytes so later caller mutation cannot change its presentation snapshot.
func WithPresentation(result Result, context PresentationContext) Result {
	copied := PresentationContext{Sources: make(map[string]CapturedSource, len(context.Sources))}
	for unit, source := range context.Sources {
		captured := source
		if source.Bytes != nil {
			captured.Bytes = append([]byte(nil), source.Bytes...)
		}
		copied.Sources[unit] = captured
	}
	result.Presentation = &copied
	return result
}

func enrichDiagnostics(result Result) Result {
	if result.Presentation == nil || result.Diagnostics == nil {
		return result
	}
	result.Diagnostics = append([]diagnostic.Diagnostic(nil), result.Diagnostics...)
	for index := range result.Diagnostics {
		problem := &result.Diagnostics[index]
		if source, ok := result.Presentation.Sources[problem.Unit]; ok {
			problem.Primary = locatedSpan(problem.Primary, source.Bytes, problem.SourcePosition)
		} else {
			problem.Primary.Line, problem.Primary.Col = 0, 0
		}
		if problem.Causes != nil {
			problem.Causes = append([]diagnostic.Cause(nil), problem.Causes...)
			for causeIndex := range problem.Causes {
				cause := &problem.Causes[causeIndex]
				if cause.Span == nil {
					continue
				}
				span := *cause.Span
				if source, ok := result.Presentation.Sources[cause.EffectiveUnit(problem.Unit)]; ok {
					span = locatedSpan(span, source.Bytes, true)
				} else {
					span.Line, span.Col = 0, 0
				}
				cause.Span = &span
			}
		}
		if problem.Repairs != nil {
			problem.Repairs = append([]diagnostic.Repair(nil), problem.Repairs...)
			for repairIndex := range problem.Repairs {
				repair := &problem.Repairs[repairIndex]
				if repair.Span == nil {
					continue
				}
				span := *repair.Span
				if source, ok := result.Presentation.Sources[repair.EffectiveUnit(problem.Unit)]; ok {
					span = locatedSpan(span, source.Bytes, true)
				} else {
					span.Line, span.Col = 0, 0
				}
				repair.Span = &span
			}
		}
	}
	return result
}

func locatedSpan(span diagnostic.Span, source []byte, sourcePosition bool) diagnostic.Span {
	span.Line, span.Col = 0, 0
	if span.Start < 0 || span.End < span.Start || span.End > len(source) {
		return span
	}
	if span.Start == 0 && span.End == 0 && !sourcePosition {
		return span
	}
	line, column, ok := positionAt(source, span.Start)
	if !ok {
		return span
	}
	span.Line, span.Col = line, column
	return span
}

// positionAt maps byte offsets to 1-based lines and Unicode-scalar columns.
// Offsets inside a valid UTF-8 rune map to its first byte. CRLF is one newline;
// both bytes of its terminator map to the preceding line's end position.
func positionAt(source []byte, offset int) (int, int, bool) {
	if offset < 0 || offset > len(source) {
		return 0, 0, false
	}
	line, column := 1, 1
	for index := 0; index < offset; {
		if source[index] == '\r' && index+1 < len(source) && source[index+1] == '\n' {
			if offset <= index+1 {
				return line, column, true
			}
			index += 2
			line++
			column = 1
			continue
		}
		if source[index] == '\n' {
			index++
			line++
			column = 1
			continue
		}
		_, size := utf8.DecodeRune(source[index:])
		if size == 0 {
			return 0, 0, false
		}
		if offset < index+size {
			return line, column, true
		}
		index += size
		column++
	}
	return line, column, true
}

func renderPresentedDiagnostic(problem diagnostic.Diagnostic, context *PresentationContext) (string, bool) {
	if context == nil || problem.Primary.Line < 1 || problem.Primary.Col < 1 {
		return "", false
	}
	source, ok := context.Sources[problem.Unit]
	if !ok {
		return "", false
	}
	snippet, marker, ok := renderSourceLine(source.Bytes, problem.Primary)
	if !ok {
		return "", false
	}
	file := cropEscapedASCII(source.Label)
	message := cropEscapedASCII(problem.Message)
	unit := cropEscapedASCII(problem.Unit)
	var output strings.Builder
	fmt.Fprintf(&output, "%s %s", problem.ID, problem.Code)
	if problem.Unit != "" {
		fmt.Fprintf(&output, " unit=%s", unit)
	}
	fmt.Fprintf(&output, " %s:%d:%d [%d:%d]: %s\n%s\n%s\n", file, problem.Primary.Line, problem.Primary.Col, problem.Primary.Start, problem.Primary.End, message, snippet, marker)
	return output.String(), true
}

func renderSourceLine(source []byte, span diagnostic.Span) (string, string, bool) {
	if span.Start < 0 || span.End < span.Start || span.End > len(source) {
		return "", "", false
	}
	lineStart, lineEnd, ok := sourceLineBounds(source, span.Start)
	if !ok {
		return "", "", false
	}
	line := source[lineStart:lineEnd]
	startOffset := span.Start - lineStart
	if startOffset > len(line) {
		startOffset = len(line)
	}
	endOffset := span.End - lineStart
	if endOffset < 0 {
		endOffset = 0
	}
	continues := span.End > lineEnd
	if endOffset > len(line) {
		endOffset = len(line)
	}
	rendered := renderSourceASCII(line)
	startColumn := renderColumnAt(line, startOffset)
	endColumn := renderColumnAt(line, endOffset)
	if endColumn < startColumn {
		endColumn = startColumn
	}
	snippet, visibleStart, prefixWidth := cropRenderedLine(rendered, startColumn)
	caretColumn := prefixWidth + startColumn - visibleStart
	caretWidth := endColumn - startColumn
	if caretWidth < 1 {
		caretWidth = 1
	}
	visibleEnd := visibleStart + len(snippet) - prefixWidth
	if len(snippet) >= 3 && strings.HasSuffix(snippet, "...") {
		visibleEnd -= 3
	}
	if caretColumn >= len(snippet) {
		caretWidth = 1
	} else if available := len(snippet) - caretColumn; caretWidth > available {
		caretWidth = available
	}
	if visibleEnd < startColumn {
		caretWidth = 1
	}
	marker := strings.Repeat(" ", caretColumn) + "^" + strings.Repeat("~", caretWidth-1)
	if continues {
		marker += " [continues]"
	}
	return snippet, marker, true
}

func sourceLineBounds(source []byte, offset int) (int, int, bool) {
	if offset < 0 || offset > len(source) {
		return 0, 0, false
	}
	lineStart := 0
	for index := 0; index < offset; {
		if source[index] == '\r' && index+1 < len(source) && source[index+1] == '\n' {
			if offset <= index+1 {
				return lineStart, index, true
			}
			lineStart = index + 2
			index += 2
			continue
		}
		if source[index] == '\n' {
			if offset == index {
				return lineStart, index, true
			}
			lineStart = index + 1
			index++
			continue
		}
		_, size := utf8.DecodeRune(source[index:])
		if size == 0 {
			return 0, 0, false
		}
		index += size
	}
	for lineEnd := lineStart; lineEnd < len(source); {
		if source[lineEnd] == '\n' || (source[lineEnd] == '\r' && lineEnd+1 < len(source) && source[lineEnd+1] == '\n') {
			return lineStart, lineEnd, true
		}
		_, size := utf8.DecodeRune(source[lineEnd:])
		if size == 0 {
			return 0, 0, false
		}
		lineEnd += size
	}
	return lineStart, len(source), true
}

func renderSourceASCII(source []byte) string {
	var rendered strings.Builder
	column := 0
	for index := 0; index < len(source); {
		value := source[index]
		switch {
		case value == '\t':
			spaces := diagnosticTabStop - column%diagnosticTabStop
			rendered.WriteString(strings.Repeat(" ", spaces))
			column += spaces
			index++
		case value < 0x20 || value == 0x7f:
			fmt.Fprintf(&rendered, "\\x%02X", value)
			column += 4
			index++
		case value == '\\':
			rendered.WriteString("\\\\")
			column += 2
			index++
		case value < utf8.RuneSelf:
			rendered.WriteByte(value)
			column++
			index++
		default:
			runeValue, size := utf8.DecodeRune(source[index:])
			if runeValue == utf8.RuneError && size == 1 {
				fmt.Fprintf(&rendered, "\\x%02X", value)
				column += 4
				index++
			} else {
				fmt.Fprintf(&rendered, "\\u{%X}", runeValue)
				column += len(fmt.Sprintf("\\u{%X}", runeValue))
				index += size
			}
		}
	}
	return rendered.String()
}

func renderColumnAt(source []byte, offset int) int {
	if offset < 0 {
		return 0
	}
	if offset > len(source) {
		offset = len(source)
	}
	column := 0
	for index := 0; index < offset; {
		value := source[index]
		if value == '\t' {
			column += diagnosticTabStop - column%diagnosticTabStop
			index++
			continue
		}
		if value < 0x20 || value == 0x7f {
			column += 4
			index++
			continue
		}
		if value == '\\' {
			column += 2
			index++
			continue
		}
		if value < utf8.RuneSelf {
			column++
			index++
			continue
		}
		runeValue, size := utf8.DecodeRune(source[index:])
		if runeValue == utf8.RuneError && size == 1 {
			column += 4
			index++
			continue
		}
		column += len(fmt.Sprintf("\\u{%X}", runeValue))
		if offset < index+size {
			return column - len(fmt.Sprintf("\\u{%X}", runeValue))
		}
		index += size
	}
	return column
}

func cropRenderedLine(rendered string, startColumn int) (snippet string, visibleStart int, prefixWidth int) {
	if len(rendered) <= diagnosticSnippetColumns {
		return rendered, 0, 0
	}
	if startColumn < 0 {
		startColumn = 0
	}
	if startColumn > len(rendered) {
		startColumn = len(rendered)
	}
	wantedStart := startColumn - 40
	if wantedStart < 0 {
		wantedStart = 0
	}
	leftIndex, leftColumn := renderedBoundary(rendered, wantedStart, false)
	prefixWidth = 0
	if leftIndex > 0 {
		prefixWidth = 3
	}
	contentBudget := diagnosticSnippetColumns - prefixWidth
	wantedEnd := leftColumn + contentBudget
	if wantedEnd < len(rendered) {
		contentBudget -= 3
		wantedEnd = leftColumn + contentBudget
	}
	rightIndex, rightColumn := renderedBoundary(rendered, wantedEnd, false)
	suffix := rightIndex < len(rendered)
	for prefixWidth+(rightIndex-leftIndex)+boolWidth(suffix) > diagnosticSnippetColumns {
		rightIndex, rightColumn = renderedBoundary(rendered, rightColumn-1, false)
		suffix = rightIndex < len(rendered)
	}
	var output strings.Builder
	if prefixWidth > 0 {
		output.WriteString("...")
	}
	output.WriteString(rendered[leftIndex:rightIndex])
	if suffix {
		output.WriteString("...")
	}
	return output.String(), leftColumn, prefixWidth
}

func boolWidth(value bool) int {
	if value {
		return 3
	}
	return 0
}

func renderedBoundary(rendered string, target int, roundUp bool) (int, int) {
	if target <= 0 {
		return 0, 0
	}
	if target >= len(rendered) {
		return len(rendered), len(rendered)
	}
	for index := 0; index < len(rendered); {
		next := index + 1
		if rendered[index] == '\\' && index+1 < len(rendered) {
			switch rendered[index+1] {
			case '\\':
				next = index + 2
			case 'x':
				if index+4 <= len(rendered) {
					next = index + 4
				}
			case 'u':
				if close := strings.IndexByte(rendered[index+2:], '}'); close >= 0 {
					next = index + 2 + close + 1
				}
			}
		}
		if target < next {
			if roundUp {
				return next, next
			}
			return index, index
		}
		index = next
		if target == index {
			return index, index
		}
	}
	return len(rendered), len(rendered)
}

func escapeTextASCII(value string) string {
	var escaped strings.Builder
	for index := 0; index < len(value); {
		character := value[index]
		switch {
		case character < 0x20 || character == 0x7f:
			fmt.Fprintf(&escaped, "\\x%02X", character)
			index++
		case character == '\\':
			escaped.WriteString("\\\\")
			index++
		case character < utf8.RuneSelf:
			escaped.WriteByte(character)
			index++
		default:
			runeValue, size := utf8.DecodeRuneInString(value[index:])
			if runeValue == utf8.RuneError && size == 1 {
				fmt.Fprintf(&escaped, "\\x%02X", character)
				index++
			} else {
				fmt.Fprintf(&escaped, "\\u{%X}", runeValue)
				index += size
			}
		}
	}
	return escaped.String()
}

func cropEscapedASCII(value string) string {
	escaped := escapeTextASCII(value)
	if len(escaped) <= diagnosticSnippetColumns {
		return escaped
	}
	cut, _ := renderedBoundary(escaped, diagnosticSnippetColumns-3, false)
	return escaped[:cut] + "..."
}
