package syntax_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

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

func TestRecovery(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", "malformed.lang"))
	if err != nil {
		t.Fatal(err)
	}

	first := syntax.Parse(source)
	second := syntax.Parse(source)
	if !bytes.Equal(first.Tree.Bytes(), source) {
		t.Fatal("malformed CST did not preserve every input byte")
	}
	if got := diagnosticIdentity(first.Diagnostics); !reflect.DeepEqual(got, diagnosticIdentity(second.Diagnostics)) {
		t.Fatalf("recovery diagnostics are not deterministic:\nfirst=%v\nsecond=%v", got, diagnosticIdentity(second.Diagnostics))
	}
	if len(first.Diagnostics) == 0 || len(first.Diagnostics) > 21 {
		t.Fatalf("expected 1..21 diagnostics, got %d", len(first.Diagnostics))
	}
	if len(first.Program.Funcs) != 2 || first.Program.Funcs[1].Name != "preserved" {
		t.Fatalf("recovery crossed the broken declaration: funcs=%+v", first.Program.Funcs)
	}

	overflow := syntax.Parse(bytes.Repeat([]byte{'@'}, 100))
	if len(overflow.Diagnostics) != 21 || overflow.Diagnostics[20].Code != "syntax.too_many_errors" {
		t.Fatalf("diagnostic cap is not 20 + terminal truncation fact: %+v", diagnosticIdentity(overflow.Diagnostics))
	}
}

func TestInvalidUTF8(t *testing.T) {
	source := append([]byte("module invalid\nexport {}\n"), 0xff, 0xfe)
	parsed := syntax.Parse(source)
	if !bytes.Equal(parsed.Tree.Bytes(), source) {
		t.Fatal("invalid UTF-8 was not preserved losslessly")
	}
	if len(parsed.Diagnostics) < 2 || parsed.Diagnostics[0].Code != "syntax.invalid_utf8" || parsed.Diagnostics[1].Code != "syntax.invalid_utf8" {
		t.Fatalf("invalid UTF-8 diagnostics are missing or unstable: %+v", diagnosticIdentity(parsed.Diagnostics))
	}
}

func FuzzParseFormat(f *testing.F) {
	for _, name := range []string{"toggle.lang", "comments.lang", "malformed.lang"} {
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(source)
	}
	f.Add([]byte{0xff, 0xfe, '{', '}'})

	f.Fuzz(func(t *testing.T, source []byte) {
		parsed := syntax.Parse(source)
		if !bytes.Equal(parsed.Tree.Bytes(), source) {
			t.Fatal("CST round trip changed fuzz input")
		}
		if len(parsed.Diagnostics) > 21 {
			t.Fatalf("unbounded diagnostics: %d", len(parsed.Diagnostics))
		}
		if len(parsed.Diagnostics) != 0 {
			return
		}
		first := syntax.Format(parsed.Tree)
		formatted := syntax.Parse(first)
		if len(formatted.Diagnostics) != 0 {
			t.Fatalf("formatter produced invalid source: %q diagnostics=%v", first, diagnosticIdentity(formatted.Diagnostics))
		}
		second := syntax.Format(formatted.Tree)
		if !bytes.Equal(first, second) {
			t.Fatalf("formatter is not idempotent:\nfirst=%q\nsecond=%q", first, second)
		}
	})
}

func TestGeneratedRoundTrips(t *testing.T) {
	const generationSeed int64 = 0x51a6e
	count := 0
	lastFailure := ""
	property := func(caseID uint64) bool {
		count++
		source := generatedProgram(caseID, true)
		parsed := syntax.Parse(source)
		if !bytes.Equal(parsed.Tree.Bytes(), source) {
			lastFailure = fmt.Sprintf("case=%d lost CST bytes\nsource=%q", caseID, source)
			return false
		}
		if len(parsed.Diagnostics) != 0 {
			lastFailure = fmt.Sprintf("case=%d did not parse\nsource=%q\ndiagnostics=%v", caseID, source, diagnosticIdentity(parsed.Diagnostics))
			return false
		}

		canonical := syntax.Format(parsed.Tree)
		reparsed := syntax.Parse(canonical)
		if len(reparsed.Diagnostics) != 0 || !bytes.Equal(canonical, syntax.Format(reparsed.Tree)) {
			lastFailure = fmt.Sprintf("case=%d did not reach a fixed point\nsource=%q\ncanonical=%q\ndiagnostics=%v", caseID, source, canonical, diagnosticIdentity(reparsed.Diagnostics))
			return false
		}
		if !reflect.DeepEqual(withoutSpans(parsed.Program), withoutSpans(reparsed.Program)) {
			lastFailure = fmt.Sprintf("case=%d changed semantic order\nsource=%q\ncanonical=%q", caseID, source, canonical)
			return false
		}
		if !reflect.DeepEqual(comments(parsed.Tree), comments(reparsed.Tree)) {
			lastFailure = fmt.Sprintf("case=%d changed comments\nsource=%q\ncanonical=%q", caseID, source, canonical)
			return false
		}

		invalid := generatedProgram(caseID, false)
		first := session.Check(invalid).Diagnostics
		second := session.Check(invalid).Diagnostics
		if !reflect.DeepEqual(diagnosticIdentity(first), diagnosticIdentity(second)) || !containsDiagnostic(first, "match.non_exhaustive") {
			lastFailure = fmt.Sprintf("case=%d invalid mutation was not deterministic\nsource=%q\nfirst=%v\nsecond=%v", caseID, invalid, diagnosticIdentity(first), diagnosticIdentity(second))
			return false
		}
		return true
	}

	config := &quick.Config{MaxCount: 1000, Rand: rand.New(rand.NewSource(generationSeed))}
	if err := quick.Check(property, config); err != nil {
		t.Fatalf("generated property failed (seed=%d): %v\n%s", generationSeed, err, lastFailure)
	}
	if count != config.MaxCount {
		t.Fatalf("generated property ran %d cases, want %d (seed=%d)", count, config.MaxCount, generationSeed)
	}
}

func generatedProgram(caseID uint64, exhaustive bool) []byte {
	separator := " "
	if caseID%2 == 1 {
		separator = "\t"
	}
	typeName := fmt.Sprintf("Mode%d", caseID%997)
	functionName := fmt.Sprintf("rotate%d", caseID%991)
	alternativeCount := int(caseID%4) + 2
	alternatives := make([]string, alternativeCount)
	for index := range alternatives {
		alternatives[index] = fmt.Sprintf("A%d_%d", caseID%983, index)
	}

	var source strings.Builder
	if caseID%3 == 0 {
		fmt.Fprintf(&source, "// generated case %d\n", caseID)
	}
	fmt.Fprintf(&source, "module%sgenerated.case%d\n\n", separator, caseID%977)
	fmt.Fprintf(&source, "export%s{%s type%s%s%s fn%s%s%s}\n\n", separator, separator, separator, typeName, separator, separator, functionName, separator)
	fmt.Fprintf(&source, "data%s%s%s=%s\n", separator, typeName, separator, separator)
	for _, alternative := range alternatives {
		fmt.Fprintf(&source, "%s|%s%s\n", separator, separator, alternative)
	}
	fmt.Fprintf(&source, "\nfn%s%s(state:%s%s)%s->%s%s%s{\n", separator, functionName, separator, typeName, separator, separator, typeName, separator)
	fmt.Fprintf(&source, "%smatch%sstate%s{\n", separator, separator, separator)
	limit := alternativeCount
	if !exhaustive {
		limit--
	}
	offset := int(caseID % uint64(alternativeCount))
	for index := 0; index < limit; index++ {
		position := (index + offset) % alternativeCount
		next := (position + 1) % alternativeCount
		if caseID%5 == 0 && index == 0 {
			source.WriteString("// generated arm\n")
		}
		fmt.Fprintf(&source, "%s%s%s=>%s%s\n", separator, alternatives[position], separator, separator, alternatives[next])
	}
	source.WriteString("}\n}\n")
	return []byte(source.String())
}

func containsDiagnostic(diagnostics []diagnostic.Diagnostic, code string) bool {
	for _, problem := range diagnostics {
		if problem.Code == code {
			return true
		}
	}
	return false
}

func diagnosticIdentity(diagnostics []diagnostic.Diagnostic) []string {
	result := make([]string, len(diagnostics))
	for index, problem := range diagnostics {
		result[index] = fmt.Sprintf("%s:%d:%d", problem.Code, problem.Primary.Start, problem.Primary.End)
	}
	return result
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
