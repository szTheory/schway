package syntax_test

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestCSTRoundTrip(t *testing.T) {
	for _, name := range []string{"toggle.lang", "comments.lang", "non_exhaustive.lang"} {
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", name))
		if err != nil {
			t.Fatal(err)
		}
		parsed := syntax.Parse(source)
		if !bytes.Equal(parsed.Tree.Bytes(), source) {
			t.Fatalf("%s: CST lost bytes", name)
		}
	}
}

func TestFormatIdempotent(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", "toggle.lang"))
	if err != nil {
		t.Fatal(err)
	}
	first := session.Format(source)
	second := session.Format(first.Canonical)
	if len(first.Diagnostics) != 0 || len(second.Diagnostics) != 0 || !bytes.Equal(first.Canonical, second.Canonical) {
		t.Fatalf("formatter is not a fixed point:\nfirst=%q\nsecond=%q\nfirst diagnostics=%+v\nsecond diagnostics=%+v", first.Canonical, second.Canonical, first.Diagnostics, second.Diagnostics)
	}
	before := withoutSpans(syntax.Parse(source).Program)
	after := withoutSpans(syntax.Parse(first.Canonical).Program)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("semantic AST changed:\nbefore=%+v\nafter=%+v", before, after)
	}
}

func withoutSpans(program ast.Program) ast.Program {
	noSpan := diagnostic.Span{}
	for index := range program.Exports {
		program.Exports[index].Span = noSpan
	}
	for dataIndex := range program.Data {
		program.Data[dataIndex].Span = noSpan
		for alternativeIndex := range program.Data[dataIndex].Alternatives {
			program.Data[dataIndex].Alternatives[alternativeIndex].Span = noSpan
		}
	}
	for functionIndex := range program.Funcs {
		function := &program.Funcs[functionIndex]
		function.Span = noSpan
		function.Parameter.Span = noSpan
		function.Body.Span = noSpan
		for armIndex := range function.Body.Arms {
			function.Body.Arms[armIndex].Span = noSpan
		}
	}
	return program
}

func TestCommentPreservation(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", "comments.lang"))
	if err != nil {
		t.Fatal(err)
	}
	formatted := session.Format(source)
	if len(formatted.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", formatted.Diagnostics)
	}
	before := comments(syntax.Parse(source).Tree)
	after := comments(syntax.Parse(formatted.Canonical).Tree)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("comments changed:\nbefore=%q\nafter=%q", before, after)
	}
}

func TestFormatCheck(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	result, err := session.FormatFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(result.Source, result.Canonical) {
		t.Fatalf("committed toggle fixture is not canonical:\n%s", result.Canonical)
	}
}

func comments(tree syntax.Tree) []string {
	var result []string
	for _, token := range tree.Tokens {
		if token.Kind == syntax.TokenComment {
			result = append(result, token.Text)
		}
	}
	return result
}
