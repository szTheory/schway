package execution

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// InvocationSegment names one caller-owned call-site occurrence below an
// entry activation. Ordinal is deliberately part of the public grammar even
// though the current acyclic language always emits zero.
type InvocationSegment struct {
	OpCallID string
	Ordinal  int
}

// Invocation is the decoded, structured form of a canonical /2 invocation.
type Invocation struct {
	EntryID  string
	Segments []InvocationSegment
}

// FormatInvocation renders the sole /2 occurrence-identity spelling.
func FormatInvocation(entryID string, segments []InvocationSegment) (string, error) {
	if entryID == "" {
		return "", fmt.Errorf("invocation entry ID is empty")
	}
	if !utf8.ValidString(entryID) {
		return "", fmt.Errorf("invocation entry ID is not UTF-8")
	}
	var value strings.Builder
	value.WriteString("inv:entry:")
	value.WriteString(escapeInvocationComponent(entryID))
	for _, segment := range segments {
		if segment.OpCallID == "" {
			return "", fmt.Errorf("invocation call ID is empty")
		}
		if !utf8.ValidString(segment.OpCallID) {
			return "", fmt.Errorf("invocation call ID is not UTF-8")
		}
		if segment.Ordinal < 0 {
			return "", fmt.Errorf("invocation ordinal is negative")
		}
		value.WriteByte('/')
		value.WriteString(escapeInvocationComponent(segment.OpCallID))
		value.WriteByte('#')
		value.WriteString(strconv.Itoa(segment.Ordinal))
	}
	return value.String(), nil
}

// ParseInvocation accepts only the canonical grammar. In particular, it
// checks escaping while decoding rather than normalizing an alternate input.
func ParseInvocation(value string) (Invocation, error) {
	const prefix = "inv:entry:"
	if !strings.HasPrefix(value, prefix) {
		return Invocation{}, fmt.Errorf("invocation has no entry prefix")
	}
	rest := strings.TrimPrefix(value, prefix)
	parts := strings.Split(rest, "/")
	if len(parts) == 0 {
		return Invocation{}, fmt.Errorf("invocation entry is missing")
	}
	entryID, err := unescapeInvocationComponent(parts[0])
	if err != nil {
		return Invocation{}, fmt.Errorf("invocation entry: %w", err)
	}
	parsed := Invocation{EntryID: entryID}
	for _, part := range parts[1:] {
		if strings.Count(part, "#") != 1 {
			return Invocation{}, fmt.Errorf("invocation segment has invalid ordinal delimiter")
		}
		callEncoded, ordinalText, _ := strings.Cut(part, "#")
		callID, err := unescapeInvocationComponent(callEncoded)
		if err != nil {
			return Invocation{}, fmt.Errorf("invocation call ID: %w", err)
		}
		if ordinalText == "" || (len(ordinalText) > 1 && ordinalText[0] == '0') {
			return Invocation{}, fmt.Errorf("invocation ordinal is not canonical")
		}
		for _, b := range []byte(ordinalText) {
			if b < '0' || b > '9' {
				return Invocation{}, fmt.Errorf("invocation ordinal is not decimal")
			}
		}
		ordinal, err := strconv.Atoi(ordinalText)
		if err != nil {
			return Invocation{}, fmt.Errorf("invocation ordinal: %w", err)
		}
		parsed.Segments = append(parsed.Segments, InvocationSegment{OpCallID: callID, Ordinal: ordinal})
	}
	canonical, err := FormatInvocation(parsed.EntryID, parsed.Segments)
	if err != nil || canonical != value {
		return Invocation{}, fmt.Errorf("invocation is not canonical")
	}
	return parsed, nil
}

func escapeInvocationComponent(value string) string {
	const hex = "0123456789ABCDEF"
	var escaped strings.Builder
	for _, b := range []byte(value) {
		if invocationLiteral(b) {
			escaped.WriteByte(b)
			continue
		}
		escaped.WriteByte('%')
		escaped.WriteByte(hex[b>>4])
		escaped.WriteByte(hex[b&0x0f])
	}
	return escaped.String()
}

func unescapeInvocationComponent(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("component is empty")
	}
	decoded := make([]byte, 0, len(value))
	for i := 0; i < len(value); i++ {
		b := value[i]
		if invocationLiteral(b) {
			decoded = append(decoded, b)
			continue
		}
		if b != '%' || i+2 >= len(value) || !uppercaseHex(value[i+1]) || !uppercaseHex(value[i+2]) {
			return "", fmt.Errorf("component has non-canonical escape")
		}
		decoded = append(decoded, hexByte(value[i+1])<<4|hexByte(value[i+2]))
		i += 2
	}
	if !utf8.Valid(decoded) {
		return "", fmt.Errorf("component is not UTF-8")
	}
	return string(decoded), nil
}

func invocationLiteral(b byte) bool {
	return b == ':' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || b == '-' || b == '.' || b == '_' || b == '~'
}

func uppercaseHex(b byte) bool { return b >= '0' && b <= '9' || b >= 'A' && b <= 'F' }

func hexByte(b byte) byte {
	if b >= '0' && b <= '9' {
		return b - '0'
	}
	return b - 'A' + 10
}
