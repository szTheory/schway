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

// linearGeneratedTypes and linearGeneratedKinds span the Phase 02 linear
// surface: every declared type shape the phase admits (leaf, generic, nested
// generic, aggregate) crossed with every binding kind.
var (
	linearGeneratedTypes = []string{"Byte", "Buffer", "Box<Byte>", "Box<Box<Byte>>", "Pair<Byte, Buffer>"}
	linearGeneratedKinds = []string{"", "take ", "borrow "}
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
