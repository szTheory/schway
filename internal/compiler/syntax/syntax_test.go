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

func TestTokenBudgetStopsWithOneStableDiagnostic(t *testing.T) {
	result := syntax.Parse([]byte(strings.Repeat("x ", syntax.MaxTokens+1)))
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "syntax.input_limit" {
		t.Fatalf("token limit diagnostics=%+v", result.Diagnostics)
	}
	if len(result.Tree.Tokens) != syntax.MaxTokens+1 { // bounded tokens plus EOF
		t.Fatalf("token count=%d want=%d", len(result.Tree.Tokens), syntax.MaxTokens+1)
	}
}

func TestPhase19NumericToken(t *testing.T) {
	for _, spelling := range []string{"0", "00", "08", "42", "4_2", "0x2a", "0b101010"} {
		tokens, diagnostics := syntax.Lex([]byte(spelling))
		if len(diagnostics) != 0 || len(tokens) != 2 || tokens[0].Kind != syntax.TokenNumber || tokens[0].Text != spelling || tokens[0].Span.Start != 0 || tokens[0].Span.End != len(spelling) {
			t.Errorf("Lex(%q) = tokens %+v, diagnostics %+v", spelling, tokens, diagnostics)
		}
	}
}

func TestPhase19NumericMalformed(t *testing.T) {
	for _, spelling := range []string{"0x", "0b", "1_", "1__2", "0x_F", "0b2", "0o77", "42u64", "42abc"} {
		tokens, diagnostics := syntax.Lex([]byte(spelling))
		if len(diagnostics) != 1 || len(tokens) != 2 || tokens[0].Kind != syntax.TokenUnknown || tokens[0].Text != spelling || tokens[0].Span.Start != 0 || tokens[0].Span.End != len(spelling) || diagnostics[0].Primary.Start != 0 || diagnostics[0].Primary.End != len(spelling) {
			t.Errorf("Lex(%q) = tokens %+v, diagnostics %+v; want one full-span malformed token", spelling, tokens, diagnostics)
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
		if function.Body.Linear != nil {
			function.Body.Linear.Span = noSpan
			for bindingIndex := range function.Body.Linear.Bindings {
				function.Body.Linear.Bindings[bindingIndex].Span = noSpan
				function.Body.Linear.Bindings[bindingIndex].RHS.Span = noSpan
			}
			if match := function.Body.Linear.TerminalMatch; match != nil {
				match.Span = noSpan
				for armIndex := range match.Arms {
					match.Arms[armIndex].Span = noSpan
					if body := match.Arms[armIndex].Body; body != nil {
						body.Span = noSpan
						for bindingIndex := range body.Bindings {
							body.Bindings[bindingIndex].Span = noSpan
							body.Bindings[bindingIndex].RHS.Span = noSpan
						}
					}
				}
			}
		}
		for armIndex := range function.Body.Arms {
			function.Body.Arms[armIndex].Span = noSpan
			if body := function.Body.Arms[armIndex].Body; body != nil {
				body.Span = noSpan
				for bindingIndex := range body.Bindings {
					body.Bindings[bindingIndex].Span = noSpan
					body.Bindings[bindingIndex].RHS.Span = noSpan
				}
			}
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

func TestOwnershipRoundTrip(t *testing.T) {
	for _, name := range []string{"owned_transfer.lang", "implicit_copy.lang", "ability_shapes.lang"} {
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase2", name))
		if err != nil {
			t.Fatal(err)
		}
		parsed := syntax.Parse(source)
		if !bytes.Equal(parsed.Tree.Bytes(), source) {
			t.Fatalf("%s: ownership CST lost bytes", name)
		}
		if len(parsed.Diagnostics) != 0 {
			t.Fatalf("%s: unexpected diagnostics: %v", name, diagnosticIdentity(parsed.Diagnostics))
		}
		canonical := syntax.Format(parsed.Tree)
		if !bytes.Equal(canonical, source) {
			t.Fatalf("%s is not canonical:\n%s", name, canonical)
		}
		reparsed := syntax.Parse(canonical)
		if len(reparsed.Diagnostics) != 0 || !bytes.Equal(canonical, syntax.Format(reparsed.Tree)) {
			t.Fatalf("%s did not reach a formatting fixed point: %v", name, diagnosticIdentity(reparsed.Diagnostics))
		}
		if !reflect.DeepEqual(withoutSpans(parsed.Program), withoutSpans(reparsed.Program)) {
			t.Fatalf("%s changed the closed-body semantic projection", name)
		}
		if !reflect.DeepEqual(comments(parsed.Tree), comments(reparsed.Tree)) {
			t.Fatalf("%s changed comments", name)
		}
	}
}

// TestArmBodyRoundTrips is 03-01-02's syntax falsifier: a match arm's value
// position may hold a full linear body, parses losslessly, formats to a
// fixed point, and reparses to the same semantic token projection — and the
// bare-name arm form (exercised by the Phase 1 fixtures TestOwnershipRoundTrip
// already covers) remains legal and unchanged alongside it.
func TestArmBodyRoundTrips(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", "branch_view.lang"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if !bytes.Equal(parsed.Tree.Bytes(), source) {
		t.Fatal("branch CST lost bytes")
	}
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnosticIdentity(parsed.Diagnostics))
	}
	if len(parsed.Program.Funcs) != 1 || len(parsed.Program.Funcs[0].Body.Arms) != 2 {
		t.Fatalf("expected one function with two arms: %+v", parsed.Program.Funcs)
	}
	for _, arm := range parsed.Program.Funcs[0].Body.Arms {
		if !arm.HasClosedVariant() {
			t.Fatalf("arm %+v does not have exactly one value form", arm)
		}
		if arm.Body == nil {
			t.Fatalf("arm %+v expected a linear body", arm)
		}
	}
	canonical := syntax.Format(parsed.Tree)
	if !bytes.Equal(canonical, source) {
		t.Fatalf("branch fixture is not canonical:\n%s", canonical)
	}
	reparsed := syntax.Parse(canonical)
	if len(reparsed.Diagnostics) != 0 || !bytes.Equal(canonical, syntax.Format(reparsed.Tree)) {
		t.Fatalf("branch fixture did not reach a formatting fixed point: %v", diagnosticIdentity(reparsed.Diagnostics))
	}
	if !reflect.DeepEqual(withoutSpans(parsed.Program), withoutSpans(reparsed.Program)) {
		t.Fatal("branch fixture changed the closed-body semantic projection across reparse")
	}
	if !reflect.DeepEqual(comments(parsed.Tree), comments(reparsed.Tree)) {
		t.Fatal("branch fixture changed comments across reparse")
	}
}

// TestExclusiveBorrowRoundTrips is 03-02-01's syntax falsifier: `borrow mut`
// parses losslessly, formats to a fixed point, and preserves token identity
// across canonicalisation — the D-12a slice's frontend half. Plain `borrow`
// stays legal and unchanged alongside it (TestOwnershipRoundTrip already
// covers that).
func TestExclusiveBorrowRoundTrips(t *testing.T) {
	source := []byte("module owned.exclusive_borrow\n\nexport {\n  fn relay\n}\n\nfn relay(buffer: Buffer) -> Buffer {\n  let view = borrow mut buffer\n  let delivered = take buffer\n  delivered\n}\n")
	parsed := syntax.Parse(source)
	if !bytes.Equal(parsed.Tree.Bytes(), source) {
		t.Fatal("exclusive borrow CST lost bytes")
	}
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnosticIdentity(parsed.Diagnostics))
	}
	if len(parsed.Program.Funcs) != 1 || parsed.Program.Funcs[0].Body.Linear == nil || len(parsed.Program.Funcs[0].Body.Linear.Bindings) != 2 {
		t.Fatalf("expected one function with two bindings: %+v", parsed.Program.Funcs)
	}
	if got := parsed.Program.Funcs[0].Body.Linear.Bindings[0].RHS.Kind; got != "borrow_mut" {
		t.Fatalf("expected the first binding's RHS kind to be borrow_mut, got %q: %+v", got, parsed.Program.Funcs[0].Body.Linear.Bindings[0])
	}
	canonical := syntax.Format(parsed.Tree)
	if !bytes.Equal(canonical, source) {
		t.Fatalf("exclusive borrow fixture is not canonical:\n%s", canonical)
	}
	reparsed := syntax.Parse(canonical)
	if len(reparsed.Diagnostics) != 0 || !bytes.Equal(canonical, syntax.Format(reparsed.Tree)) {
		t.Fatalf("exclusive borrow fixture did not reach a formatting fixed point: %v", diagnosticIdentity(reparsed.Diagnostics))
	}
	if !reflect.DeepEqual(withoutSpans(parsed.Program), withoutSpans(reparsed.Program)) {
		t.Fatal("exclusive borrow fixture changed the closed-body semantic projection across reparse")
	}
	if !reflect.DeepEqual(semanticTokens(parsed.Tree), semanticTokens(reparsed.Tree)) {
		t.Fatal("exclusive borrow fixture changed its semantic token projection across reparse")
	}
}

// TestPublicViewRoundTrips is 03-06-01's syntax falsifier for OWN-04: the
// borrow(path)/borrow mut(path) return-type annotation parses losslessly,
// formats to a fixed point, and preserves token identity across reparse.
func TestPublicViewRoundTrips(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", "public_view.lang"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if !bytes.Equal(parsed.Tree.Bytes(), source) {
		t.Fatal("public view fixture CST lost bytes")
	}
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnosticIdentity(parsed.Diagnostics))
	}
	if len(parsed.Program.Funcs) != 1 {
		t.Fatalf("expected one function: %+v", parsed.Program.Funcs)
	}
	origin := parsed.Program.Funcs[0].ReturnOrigin
	if origin == nil || origin.Path != "buffer" || origin.Access != "shared" {
		t.Fatalf("expected a shared origin naming buffer, got %+v", origin)
	}
	canonical := syntax.Format(parsed.Tree)
	if !bytes.Equal(canonical, source) {
		t.Fatalf("public view fixture is not canonical:\n%s", canonical)
	}
	reparsed := syntax.Parse(canonical)
	if len(reparsed.Diagnostics) != 0 || !bytes.Equal(canonical, syntax.Format(reparsed.Tree)) {
		t.Fatalf("public view fixture did not reach a formatting fixed point: %v", diagnosticIdentity(reparsed.Diagnostics))
	}
	if !reflect.DeepEqual(withoutSpans(parsed.Program), withoutSpans(reparsed.Program)) {
		t.Fatal("public view fixture changed the closed-body semantic projection across reparse")
	}
	if !reflect.DeepEqual(semanticTokens(parsed.Tree), semanticTokens(reparsed.Tree)) {
		t.Fatal("public view fixture changed its semantic token projection across reparse")
	}

	exclusive := []byte("module owned.public_view_exclusive\n\nexport {\n  fn view\n}\n\nfn view(buffer: Buffer) -> borrow mut(buffer) Buffer {\n  let observed = borrow mut buffer\n  observed\n}\n")
	exclusiveParsed := syntax.Parse(exclusive)
	if len(exclusiveParsed.Diagnostics) != 0 {
		t.Fatalf("unexpected exclusive-origin diagnostics: %v", diagnosticIdentity(exclusiveParsed.Diagnostics))
	}
	exclusiveOrigin := exclusiveParsed.Program.Funcs[0].ReturnOrigin
	if exclusiveOrigin == nil || exclusiveOrigin.Access != "exclusive" {
		t.Fatalf("expected an exclusive origin, got %+v", exclusiveOrigin)
	}
	if canonicalExclusive := syntax.Format(exclusiveParsed.Tree); !bytes.Equal(canonicalExclusive, exclusive) {
		t.Fatalf("exclusive public view fixture is not canonical:\n%s", canonicalExclusive)
	}
}

// TestArmBodyLimits is 03-01-03's syntax falsifier for T-03-01: a match
// exceeding the declared arm limit is rejected fail-closed with a bounded,
// stable diagnostic rather than accepted or run to exhaustion.
func TestArmBodyLimits(t *testing.T) {
	var arms strings.Builder
	const overCap = 100
	for i := 0; i < overCap; i++ {
		fmt.Fprintf(&arms, "  Alt%d => { let held = take flag\n  held\n  }\n", i)
	}
	var alternatives strings.Builder
	for i := 0; i < overCap; i++ {
		fmt.Fprintf(&alternatives, "  | Alt%d\n", i)
	}
	source := []byte("module owned.arm_limit\n\nexport {\n  type Wide\n  fn pick\n}\n\ndata Wide =\n" +
		alternatives.String() + "\nfn pick(flag: Wide) -> Wide {\n  match flag {\n" + arms.String() + "  }\n}\n")
	parsed := syntax.Parse(source)
	found := false
	for _, d := range parsed.Diagnostics {
		if d.Code == "syntax.arm_limit" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected syntax.arm_limit among diagnostics, got %v", diagnosticIdentity(parsed.Diagnostics))
	}
	if len(parsed.Program.Funcs) != 1 || len(parsed.Program.Funcs[0].Body.Arms) > overCap {
		t.Fatalf("arm cap did not bound the parsed arm count: got %d", len(parsed.Program.Funcs[0].Body.Arms))
	}
}

func TestOwnershipRecovery(t *testing.T) {
	source := []byte(`module owned.recovery
export { fn broken fn preserved }

fn broken(value: Pair<Byte, Box<Buffer) -> Byte {
  let copy = value
  copy
}

fn preserved(value: Byte) -> Byte {
  value
}
`)
	first := syntax.Parse(source)
	second := syntax.Parse(source)
	if !bytes.Equal(first.Tree.Bytes(), source) {
		t.Fatal("malformed ownership CST did not preserve every byte")
	}
	if len(first.Diagnostics) == 0 || len(first.Diagnostics) > 21 {
		t.Fatalf("expected bounded recovery diagnostics, got %d", len(first.Diagnostics))
	}
	if !reflect.DeepEqual(diagnosticIdentity(first.Diagnostics), diagnosticIdentity(second.Diagnostics)) {
		t.Fatalf("ownership recovery is nondeterministic: first=%v second=%v", diagnosticIdentity(first.Diagnostics), diagnosticIdentity(second.Diagnostics))
	}
	if len(first.Program.Funcs) != 2 || first.Program.Funcs[1].Name != "preserved" {
		t.Fatalf("generic recovery crossed its enclosing declaration: %+v", first.Program.Funcs)
	}
}

func TestTypeApplicationLimits(t *testing.T) {
	nested := func(boxes int) string {
		return strings.Repeat("Box<", boxes) + "Byte" + strings.Repeat(">", boxes)
	}
	program := func(parameterType string) []byte {
		return []byte("module owned.limits\nexport { fn bounded }\nfn bounded(value: " + parameterType + ") -> Byte { value }\n")
	}
	if result := syntax.Parse(program(nested(63))); len(result.Diagnostics) != 0 {
		t.Fatalf("depth 64 was rejected: %v", diagnosticIdentity(result.Diagnostics))
	}
	tooDeep := syntax.Parse(program(nested(64)))
	if !containsDiagnostic(tooDeep.Diagnostics, "syntax.type_limit") || len(tooDeep.Diagnostics) > 21 {
		t.Fatalf("depth 65 did not terminate at the stable cap: %v", diagnosticIdentity(tooDeep.Diagnostics))
	}

	wide := func(arguments int) string {
		values := make([]string, arguments)
		for index := range values {
			values[index] = "Byte"
		}
		return "Pair<" + strings.Join(values, ",") + ">"
	}
	if result := syntax.Parse(program(wide(4094))); len(result.Diagnostics) != 0 {
		t.Fatalf("4,096 type nodes were rejected: %v", diagnosticIdentity(result.Diagnostics))
	}
	first := syntax.Parse(program(wide(4095)))
	second := syntax.Parse(program(wide(4095)))
	if !containsDiagnostic(first.Diagnostics, "syntax.type_limit") || len(first.Diagnostics) > 21 {
		t.Fatalf("4,097 type nodes did not terminate at the stable cap: %v", diagnosticIdentity(first.Diagnostics))
	}
	if !reflect.DeepEqual(diagnosticIdentity(first.Diagnostics), diagnosticIdentity(second.Diagnostics)) {
		t.Fatalf("type-node limit diagnostics are nondeterministic: first=%v second=%v", diagnosticIdentity(first.Diagnostics), diagnosticIdentity(second.Diagnostics))
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
	for _, name := range []string{"owned_transfer.lang", "implicit_copy.lang", "ability_shapes.lang"} {
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase2", name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(source)
	}
	// A generic type combined with a binding: the combination the shipped
	// fixtures never reach (ability_shapes.lang is generic but has no `let`,
	// owned_transfer.lang binds but is not generic).
	f.Add([]byte("module generic.binding\n\nexport {\n  fn relay\n}\n\nfn relay(subject: Box<Byte>) -> Box<Byte> {\n  let held = take subject\n  held\n}\n"))
	f.Add([]byte("module generic.pair\n\nexport {\n  fn relay\n}\n\nfn relay(subject: Pair<Byte, Buffer>) -> Pair<Byte, Buffer> {\n  let held = borrow subject\n  held\n}\n"))
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
		if !reflect.DeepEqual(semanticTokens(parsed.Tree), semanticTokens(reparsed.Tree)) {
			lastFailure = fmt.Sprintf("case=%d changed token identity\nsource=%q\ncanonical=%q", caseID, source, canonical)
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

// TestGeneratedLinearRoundTrips is the linear-surface counterpart of
// TestGeneratedRoundTrips. A linear body cannot share a module with a match
// body (core.mixed_body_versions), so the linear generator is a separate
// module shape rather than an extra function inside generatedProgram.
func TestGeneratedLinearRoundTrips(t *testing.T) {
	const generationSeed int64 = 0x11ea40
	count := 0
	lastFailure := ""
	property := func(caseID uint64) bool {
		count++
		source := generatedLinearProgram(caseID)
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
		if len(reparsed.Diagnostics) != 0 {
			lastFailure = fmt.Sprintf("case=%d canonical form does not reparse\nsource=%q\ncanonical=%q\ndiagnostics=%v", caseID, source, canonical, diagnosticIdentity(reparsed.Diagnostics))
			return false
		}
		if !bytes.Equal(canonical, syntax.Format(reparsed.Tree)) {
			lastFailure = fmt.Sprintf("case=%d is not a formatting fixed point\nsource=%q\ncanonical=%q\nsecond=%q", caseID, source, canonical, syntax.Format(reparsed.Tree))
			return false
		}
		if !reflect.DeepEqual(semanticTokens(parsed.Tree), semanticTokens(reparsed.Tree)) {
			lastFailure = fmt.Sprintf("case=%d changed token identity\nsource=%q\ncanonical=%q", caseID, source, canonical)
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

		// Spans legitimately move when source is reindented, so the invariant
		// canonicalisation must preserve is the diagnostic code sequence.
		first := diagnosticCodes(session.Check(source).Diagnostics)
		second := diagnosticCodes(session.Check(canonical).Diagnostics)
		if !reflect.DeepEqual(first, second) {
			lastFailure = fmt.Sprintf("case=%d canonicalisation changed checking\nsource=%q\ncanonical=%q\nfirst=%v\nsecond=%v", caseID, source, canonical, first, second)
			return false
		}
		return true
	}

	config := &quick.Config{MaxCount: 1000, Rand: rand.New(rand.NewSource(generationSeed))}
	if err := quick.Check(property, config); err != nil {
		t.Fatalf("generated linear property failed (seed=%d): %v\n%s", generationSeed, err, lastFailure)
	}
	if count != config.MaxCount {
		t.Fatalf("generated linear property ran %d cases, want %d (seed=%d)", count, config.MaxCount, generationSeed)
	}
}

// TestGeneratedKindsIncludeExclusiveBorrow proves the generator's reachable
// operation alphabet actually includes the exclusive-borrow spelling
// (03-05-03/D-10's must_haves.truths), not merely that linearGeneratedKinds
// lists the string: a case selected to land on "borrow mut " must parse,
// round-trip to a canonical fixed point, and the resulting AST binding must
// carry RHS.Kind == "borrow_mut".
func TestGeneratedKindsIncludeExclusiveBorrow(t *testing.T) {
	found := false
	for caseID := uint64(0); caseID < 200; caseID++ {
		kind := linearGeneratedKinds[(caseID/5)%uint64(len(linearGeneratedKinds))]
		if kind != "borrow mut " {
			continue
		}
		bindingCount := int((caseID / 15) % 4)
		if bindingCount == 0 {
			continue // no binding to inspect
		}
		found = true
		source := generatedLinearProgram(caseID)
		parsed := syntax.Parse(source)
		if len(parsed.Diagnostics) != 0 {
			t.Fatalf("case=%d exclusive-borrow generation did not parse: %v\n%s", caseID, diagnosticIdentity(parsed.Diagnostics), source)
		}
		canonical := syntax.Format(parsed.Tree)
		reparsed := syntax.Parse(canonical)
		if len(reparsed.Diagnostics) != 0 || !bytes.Equal(canonical, syntax.Format(reparsed.Tree)) {
			t.Fatalf("case=%d exclusive-borrow generation is not a formatting fixed point\nsource=%q\ncanonical=%q", caseID, source, canonical)
		}
		bindings := parsed.Program.Funcs[0].Body.Linear.Bindings
		if len(bindings) == 0 || bindings[0].RHS.Kind != "borrow_mut" {
			t.Fatalf("case=%d did not produce an RHS.Kind==borrow_mut binding: %+v", caseID, bindings)
		}
	}
	if !found {
		t.Fatalf("no generated case in the probed range selected the exclusive-borrow kind — the alphabet or its selection changed")
	}
}

// linearBranchKinds spans the arm-body binding-kind alphabet: implicit
// copy, take, shared borrow, and (03-05-03/D-10) the exclusive spelling.
var linearBranchKinds = []string{"", "take ", "borrow ", "borrow mut "}

// generatedBranchProgram emits a 2-arm match function whose arms hold full
// linear bodies (03-01's arm-body extension), parameterised over
// {arm-0 binding kind} x {0, 1, or 2 bindings in arm 0} x {header-comment
// placement on the arm's own opening line} x {separator/leading-comment
// variation generatedProgram already uses}.
//
// Reachable shapes, stated explicitly per D-10 (the standing rule adopted
// after three Phase 02 gate failures were each a green test whose reachable
// input space omitted the hard case):
//   - match bodies whose arms carry full linear bodies (branch bodies), not
//     bare alternative names;
//   - an arm with ZERO bindings (Result is the arm's own aliased parameter,
//     referenced directly);
//   - an arm with a REBORROW whose original is the arm's own leading
//     binding: `let view = borrow subject` (or `borrow mut`) followed by
//     `let reviewed = borrow view`. This is read as "before the branch" per
//     D-10's own required interpretation, since 03-01's per-arm aliasing
//     gives every arm its own fresh copy of the scrutinee and there is no
//     region literally preceding the match in this grammar at all — the
//     earliest a loan can exist within an arm's own control-flow path is
//     its leading binding, exactly analogous to how 03-03's own fixture
//     pair reads "the borrowing arm's own edge" (see 03-03-SUMMARY.md,
//     key-decisions).
//   - the exclusive borrow spelling ("borrow mut ") alongside implicit
//     copy, take, and shared borrow, inside an arm body;
//   - a comment trailing an arm's own opening header line
//     (`Pattern => {  // arm tail N`), the CR-02 brace-classifier
//     placement (02-DEBT.md "Process debt") relocated inside a branch arm
//     rather than only a function declaration header.
//
// Explicitly UNREACHABLE by this generator, and by this language:
//   - any loop or back edge — the language has neither a loop construct nor
//     recursion (OWN-03 is acyclic-scoped this phase; see ROADMAP §Phase 3
//     "Scope note");
//   - a mixed bare/body match — 03-01 requires every arm to carry a body
//     once any arm does (checkBranch's core.mixed_arm_forms gate);
//   - "a generic type combined with a binding inside an arm body" — a match
//     scrutinee must be a declared nominal sum type by grammar
//     (checkBranch's DataType.Alternatives lookup), which can never be a
//     generic application like Box<Byte>, so this shape cannot exist in an
//     arm body at all. The historically-missed generic-type/header-comment
//     interaction (02-DEBT.md's CR-02 precedent) is instead pinned on the
//     STRAIGHT-LINE surface, where it is reachable: generatedLinearProgram's
//     own linearGeneratedTypes already crosses every generic shape with the
//     identical header-tail-comment placement (see its caseID%11==0 branch),
//     and TestGeneratedKindsIncludeExclusiveBorrow now additionally proves
//     that surface reaches the exclusive-borrow spelling too.
func generatedBranchProgram(caseID uint64) []byte {
	separator := " "
	if caseID%2 == 1 {
		separator = "\t"
	}
	typeName := fmt.Sprintf("Branch%d", caseID%991)
	functionName := fmt.Sprintf("choose%d", caseID%983)
	first := fmt.Sprintf("A%d", caseID%977)
	second := fmt.Sprintf("B%d", caseID%977)
	bindingKind := linearBranchKinds[(caseID/5)%uint64(len(linearBranchKinds))]
	bindingCount := int((caseID / 15) % 3) // 0, 1, or 2 bindings in arm 0

	var source strings.Builder
	if caseID%3 == 0 {
		fmt.Fprintf(&source, "// generated branch case %d\n", caseID)
	}
	fmt.Fprintf(&source, "module%sbranch.case%d\n\n", separator, caseID%971)
	fmt.Fprintf(&source, "export%s{%s type%s%s%s fn%s%s%s}\n\n", separator, separator, separator, typeName, separator, separator, functionName, separator)
	fmt.Fprintf(&source, "data%s%s%s=%s\n", separator, typeName, separator, separator)
	fmt.Fprintf(&source, "%s|%s%s\n", separator, separator, first)
	fmt.Fprintf(&source, "%s|%s%s\n\n", separator, separator, second)
	if caseID%11 == 0 {
		fmt.Fprintf(&source, "fn%s%s(subject:%s%s)%s->%s%s// header tail %d\n{\n", separator, functionName, separator, typeName, separator, separator, typeName, caseID)
	} else {
		fmt.Fprintf(&source, "fn%s%s(subject:%s%s)%s->%s%s%s{\n", separator, functionName, separator, typeName, separator, separator, typeName, separator)
	}
	matchScrutinee := "subject"
	if caseID%17 == 0 {
		matchScrutinee = "computed"
		fmt.Fprintf(&source, "%slet%scomputed%s=%ssubject\n", separator, separator, separator, separator)
	}
	fmt.Fprintf(&source, "%smatch%s%s%s{\n", separator, separator, matchScrutinee, separator)

	// Arm 0: the varied arm -- 0, 1, or 2 bindings, an optional
	// header-trailing comment, and the full kind alphabet including the
	// exclusive spelling. A second binding is always a reborrow of the
	// first (the only way to legally extend a loan's lifetime without an
	// implicit copy, since a borrowed place shares its owner's TypeID and a
	// noncopyable owner denies copy — see 03-02-SUMMARY.md's "Issues
	// Encountered").
	if caseID%13 == 0 {
		fmt.Fprintf(&source, "%s%s%s=>%s{%s// arm tail %d\n", separator, first, separator, separator, separator, caseID)
	} else {
		fmt.Fprintf(&source, "%s%s%s=>%s{\n", separator, first, separator, separator)
	}
	result := matchScrutinee
	for index := 0; index < bindingCount; index++ {
		name := fmt.Sprintf("hold%d", index)
		kind := bindingKind
		if index > 0 {
			kind = "borrow "
		}
		if caseID%7 == 0 && index == 0 {
			fmt.Fprintf(&source, "%s%s// generated arm binding\n", separator, separator)
		}
		fmt.Fprintf(&source, "%s%slet%s%s%s=%s%s%s\n", separator, separator, separator, name, separator, separator, kind, result)
		result = name
	}
	fmt.Fprintf(&source, "%s%s%s\n", separator, separator, result)
	fmt.Fprintf(&source, "%s}\n", separator)

	// Arm 1: a zero-binding arm — the aliased parameter is the arm's own
	// Result directly, with no bindings at all.
	fmt.Fprintf(&source, "%s%s%s=>%s{\n", separator, second, separator, separator)
	fmt.Fprintf(&source, "%s%s%s\n", separator, separator, matchScrutinee)
	fmt.Fprintf(&source, "%s}\n", separator)

	source.WriteString("}\n}\n")
	return []byte(source.String())
}

// TestGeneratedBranchBodiesRoundTrip proves the arm-body generator's
// canonicalization is lossless and its checking is stable across
// canonicalization, mirroring TestGeneratedLinearRoundTrips exactly but for
// the branch surface (03-05-03).
func TestGeneratedBranchBodiesRoundTrip(t *testing.T) {
	const generationSeed int64 = 0x8ea3011
	count := 0
	lastFailure := ""
	property := func(caseID uint64) bool {
		count++
		source := generatedBranchProgram(caseID)
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
		if len(reparsed.Diagnostics) != 0 {
			lastFailure = fmt.Sprintf("case=%d canonical form does not reparse\nsource=%q\ncanonical=%q\ndiagnostics=%v", caseID, source, canonical, diagnosticIdentity(reparsed.Diagnostics))
			return false
		}
		if !bytes.Equal(canonical, syntax.Format(reparsed.Tree)) {
			lastFailure = fmt.Sprintf("case=%d is not a formatting fixed point\nsource=%q\ncanonical=%q\nsecond=%q", caseID, source, canonical, syntax.Format(reparsed.Tree))
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

		// Spans legitimately move when source is reindented; the invariant
		// canonicalisation must preserve is the diagnostic code sequence
		// through the FULL checker (not just the parser), exactly as
		// TestGeneratedLinearRoundTrips already requires for the
		// straight-line surface.
		first := diagnosticCodes(session.Check(source).Diagnostics)
		second := diagnosticCodes(session.Check(canonical).Diagnostics)
		if !reflect.DeepEqual(first, second) {
			lastFailure = fmt.Sprintf("case=%d canonicalisation changed checking\nsource=%q\ncanonical=%q\nfirst=%v\nsecond=%v", caseID, source, canonical, first, second)
			return false
		}
		return true
	}

	config := &quick.Config{MaxCount: 500, Rand: rand.New(rand.NewSource(generationSeed))}
	if err := quick.Check(property, config); err != nil {
		t.Fatalf("generated branch property failed (seed=%d): %v\n%s", generationSeed, err, lastFailure)
	}
	if count != config.MaxCount {
		t.Fatalf("generated branch property ran %d cases, want %d (seed=%d)", count, config.MaxCount, generationSeed)
	}
}

// linearGeneratedTypes and linearGeneratedKinds span the Phase 02 linear
// surface: every declared type shape the phase admits (leaf, generic, nested
// generic, aggregate) crossed with every binding kind.
var (
	linearGeneratedTypes = []string{"Byte", "Buffer", "Box<Byte>", "Box<Box<Byte>>", "Pair<Byte, Buffer>"}
	// linearGeneratedKinds spans the linear binding-kind alphabet. 03-05-03
	// (D-10) extends the original three-kind set (implicit copy, take,
	// shared borrow) with the exclusive spelling ("borrow mut ") so the
	// generator's reachable operation alphabet matches every kind the
	// checker actually admits, not just the pre-03-02 subset.
	linearGeneratedKinds = []string{"", "take ", "borrow ", "borrow mut "}
)

// generatedLinearProgram emits a single-function linear module parameterised
// over {declared type} x {binding kind} x {0,1,2,3 bindings}, with the same
// separator and comment variation generatedProgram uses.
func generatedLinearProgram(caseID uint64) []byte {
	separator := " "
	if caseID%2 == 1 {
		separator = "\t"
	}
	declaredType := linearGeneratedTypes[caseID%uint64(len(linearGeneratedTypes))]
	bindingKind := linearGeneratedKinds[(caseID/5)%uint64(len(linearGeneratedKinds))]
	bindingCount := int((caseID / 15) % 4)
	functionName := fmt.Sprintf("carry%d", caseID%967)

	var source strings.Builder
	if caseID%3 == 0 {
		fmt.Fprintf(&source, "// generated linear case %d\n", caseID)
	}
	fmt.Fprintf(&source, "module%slinear.case%d\n\n", separator, caseID%971)
	fmt.Fprintf(&source, "export%s{%s fn%s%s%s}\n\n", separator, separator, separator, functionName, separator)
	// A comment trailing the declaration header breaks the line without ending
	// the header, which is the one placement that can defeat brace
	// classification; cover it alongside the uninterrupted spelling.
	if caseID%11 == 0 {
		fmt.Fprintf(&source, "fn%s%s(subject:%s%s)%s->%s%s// header tail %d\n{\n", separator, functionName, separator, declaredType, separator, separator, declaredType, caseID)
	} else {
		fmt.Fprintf(&source, "fn%s%s(subject:%s%s)%s->%s%s%s{\n", separator, functionName, separator, declaredType, separator, separator, declaredType, separator)
	}
	result := "subject"
	for index := 0; index < bindingCount; index++ {
		name := fmt.Sprintf("hold%d", index)
		if caseID%7 == 0 && index == 0 {
			source.WriteString("// generated binding\n")
		}
		fmt.Fprintf(&source, "%slet%s%s%s=%s%s%s\n", separator, separator, name, separator, separator, bindingKind, result)
		result = name
	}
	fmt.Fprintf(&source, "%s%s\n}\n", separator, result)
	return []byte(source.String())
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
	if caseID%11 == 0 {
		fmt.Fprintf(&source, "\nfn%s%s(state:%s%s)%s->%s%s// header tail %d\n{\n", separator, functionName, separator, typeName, separator, separator, typeName, caseID)
	} else {
		fmt.Fprintf(&source, "\nfn%s%s(state:%s%s)%s->%s%s%s{\n", separator, functionName, separator, typeName, separator, separator, typeName, separator)
	}
	if caseID%13 == 0 {
		fmt.Fprintf(&source, "%smatch%sstate%s// match tail %d\n%s{\n", separator, separator, separator, caseID, separator)
	} else {
		fmt.Fprintf(&source, "%smatch%sstate%s{\n", separator, separator, separator)
	}
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

func diagnosticCodes(diagnostics []diagnostic.Diagnostic) []string {
	result := make([]string, len(diagnostics))
	for index, problem := range diagnostics {
		result[index] = problem.Code
	}
	return result
}

// semanticTokens projects the token identity the formatter must preserve:
// every non-whitespace, non-EOF token in order. Fusing two identifiers into
// one, or dropping one, changes this projection even when the result still
// parses.
func semanticTokens(tree syntax.Tree) []string {
	var result []string
	for _, token := range tree.Tokens {
		if token.Kind == syntax.TokenWhitespace || token.Kind == syntax.TokenEOF {
			continue
		}
		result = append(result, fmt.Sprintf("%v:%s", token.Kind, token.Text))
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

// TestForeignCallRoundTrips proves the `foreign C { }` declaration block and
// the `try <callee>(<args>)` fallible-call form parse losslessly, format to
// a fixed point, and reparse to the same semantic token projection
// (D-04-01/D-04-04/D-04-05/D-04-10's surface syntax).
func TestForeignCallRoundTrips(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.lang"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %v", parsed.Diagnostics)
	}
	if len(parsed.Program.Foreign) != 1 || len(parsed.Program.Foreign[0].Symbols) != 1 {
		t.Fatalf("foreign declaration did not parse: %+v", parsed.Program.Foreign)
	}
	symbol := parsed.Program.Foreign[0].Symbols[0]
	if symbol.Name != "lang_res_open" {
		t.Fatalf("symbol name = %q, want lang_res_open", symbol.Name)
	}
	policies := make(map[string]string, len(symbol.Policies))
	for _, policy := range symbol.Policies {
		policies[policy.Key] = policy.Value
	}
	if policies["unwind"] != "forbidden" || policies["nonlocal_exit"] != "forbidden" || policies["allocator"] != "libc_malloc" || policies["fails"] != "AcquireError" {
		t.Fatalf("policies = %+v", policies)
	}

	if len(parsed.Program.Funcs) != 1 || len(parsed.Program.Funcs[0].Body.Linear.Bindings) != 1 {
		t.Fatalf("unexpected function shape: %+v", parsed.Program.Funcs)
	}
	rhs := parsed.Program.Funcs[0].Body.Linear.Bindings[0].RHS
	if rhs.Kind != "try_call" || rhs.Callee != "lang_res_open" || len(rhs.Arguments) != 1 || rhs.Arguments[0] != "request" {
		t.Fatalf("try-call RHS = %+v", rhs)
	}

	formatted := syntax.Format(parsed.Tree)
	reparsed := syntax.Parse(formatted)
	if len(reparsed.Diagnostics) > 0 {
		t.Fatalf("reparse diagnostics: %v", reparsed.Diagnostics)
	}
	if got, want := semanticTokens(reparsed.Tree), semanticTokens(parsed.Tree); !reflect.DeepEqual(got, want) {
		t.Fatalf("semantic token projection changed:\ngot:  %v\nwant: %v", got, want)
	}
	reformatted := syntax.Format(reparsed.Tree)
	if !bytes.Equal(formatted, reformatted) {
		t.Fatalf("format is not a fixed point:\nfirst:  %s\nsecond: %s", formatted, reformatted)
	}
}

// TestFallibleCallUnconsumedRejected pins D-04-06: a bare fallible call in a
// binding right-hand side (no `try`) must never reach a checked core
// function.
//
// D-07-40 (deliberate edit): this refusal's ENFORCEMENT LAYER relocated from
// parse time to check time in Phase 07 (D-07-01) -- a bare call's callee
// identity (Lang function, foreign symbol, or unresolved) is a check-time
// fact, since the parser now accepts the "call" shape unconditionally so a
// bare call to a declared Lang function can be admitted. The PUBLISHED CODE
// (syntax.fallible_call_not_consumed) did not change; only where it is
// asserted did. This test now asserts syntax.Parse itself produces NO error
// diagnostic for this source (the parser no longer refuses this shape), and
// that the refusal appears from session.Check instead, with the identical
// code.
func TestFallibleCallUnconsumedRejected(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "fallible_call_unconsumed.lang"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	for _, problem := range parsed.Diagnostics {
		if problem.Severity == "error" {
			t.Fatalf("D-07-40: syntax.Parse must produce no error diagnostic for a bare call shape now that check owns the refusal, got %v", problem)
		}
	}
	checked := session.Check(source)
	if len(checked.Program.Functions) != 0 {
		t.Fatalf("a bare fallible call must never reach a checked core function, got %d", len(checked.Program.Functions))
	}
	found := false
	for _, problem := range checked.Diagnostics {
		if problem.Code == "syntax.fallible_call_not_consumed" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected syntax.fallible_call_not_consumed from check, got %v", checked.Diagnostics)
	}
}

// TestDefectTerminatorRoundTrips proves D-04-15's `defect "<reason>"`
// terminator parses into ast.LinearBody.DefectReason and formats to a fixed
// point, mirroring TestForeignCallRoundTrips' shape.
func TestDefectTerminatorRoundTrips(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "defect_terminal.lang"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %v", parsed.Diagnostics)
	}
	if len(parsed.Program.Funcs) != 1 {
		t.Fatalf("unexpected function shape: %+v", parsed.Program.Funcs)
	}
	var defectArm *ast.MatchArm
	for index, arm := range parsed.Program.Funcs[0].Body.Arms {
		if arm.Pattern == "Halt" {
			defectArm = &parsed.Program.Funcs[0].Body.Arms[index]
		}
	}
	if defectArm == nil || defectArm.Body == nil || defectArm.Body.DefectReason != "halt requested" {
		t.Fatalf("Halt arm did not parse a defect terminator: %+v", defectArm)
	}
	if defectArm.Body.Result != "" {
		t.Fatalf("a defect body must not also carry a Result: %q", defectArm.Body.Result)
	}

	formatted := syntax.Format(parsed.Tree)
	reparsed := syntax.Parse(formatted)
	if len(reparsed.Diagnostics) > 0 {
		t.Fatalf("reparse diagnostics: %v", reparsed.Diagnostics)
	}
	if got, want := semanticTokens(reparsed.Tree), semanticTokens(parsed.Tree); !reflect.DeepEqual(got, want) {
		t.Fatalf("semantic token projection changed:\ngot:  %v\nwant: %v", got, want)
	}
	reformatted := syntax.Format(reparsed.Tree)
	if !bytes.Equal(formatted, reformatted) {
		t.Fatalf("format is not a fixed point:\nfirst:  %s\nsecond: %s", formatted, reformatted)
	}
}
