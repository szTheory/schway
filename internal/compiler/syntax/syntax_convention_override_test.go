// This file proves OWN-05's SOURCE-LAYER claim only (D-09-34, D-09-35): a
// call-site override of a callee's declared ownership convention is not
// expressible in Lang source, proven as STRUCTURAL ABSENCE rather than as a
// refusal that merely happens to fire today. No grammar production ever
// admits a convention-shaped keyword (`take`, `borrow`, `borrow mut`) inside
// a call's argument list, in any of the three call-parse routes the grammar
// has, and `ast.RHS` -- the ONLY struct a call's arguments could possibly
// land in -- declares no field that could carry such an annotation even if
// a future parser change accidentally admitted the token sequence.
//
// "Not expressible in source" is NOT the same claim as "not expressible in
// core": corevalidate's whole role is validating a core artifact it did not
// produce, and a hostile or corrupted producer never goes through this
// parser at all. That second, separate, core-layer claim is proven in
// internal/compiler/core/core_convention_absence_test.go
// (TestConventionOverrideNotExpressibleInCore /
// TestParameterModeDecodeIsClosedSet) -- do not read this file's refusal as
// covering that trust boundary too.
package syntax_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// readParserSource reads parser.go's own source text so
// TestOwnershipShapedKeywordVocabularyIsClosed can scan it, mirroring the
// corevalidate package's own source-text-scan import-independence guards
// (never assuming the vocabulary rather than reading it fresh).
func readParserSource(t *testing.T) (string, error) {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "syntax", "parser.go"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ownershipShapedKeywords is the closed set of source keywords that can ever
// annotate a binding's ownership convention (parser.go's own
// normalizeOwnershipTokens retext-classifies exactly these identifier-typed
// tokens into TokenLet/TokenTake/TokenBorrow/TokenMut). "let" is excluded
// below because it introduces a BINDING, never a call argument's own
// convention -- the three genuinely convention-shaped keywords a call
// argument could ever plausibly carry are take, borrow, and borrow's mut
// modifier (which only ever appears immediately after borrow). This test
// attempts all three, standalone and combined as "borrow mut", against
// every call-argument-parsing route the grammar has.
var ownershipShapedKeywords = []string{"take", "borrow", "borrow mut"}

// callParseRouteSource builds a minimal, otherwise well-formed Lang program
// whose body attempts to annotate a call ARGUMENT with keyword, in one of
// the three call-parse routes the grammar admits: a bare call, a `try`
// call, and a `discard ... because` call. The callee name is irrelevant to
// parsing (callee resolution is a check-time fact, D-07-01) so a fixed
// unresolved name is used throughout.
func callParseRouteSource(route, keyword string) []byte {
	var body string
	switch route {
	case "bare_call":
		body = fmt.Sprintf("let result = f(%s x)\n  result", keyword)
	case "try_call":
		body = fmt.Sprintf("let result = try f(%s x)\n  result", keyword)
	case "discard_call":
		body = fmt.Sprintf("discard f(%s x) because \"rationale text\"\n  x", keyword)
	default:
		panic("unknown route " + route)
	}
	source := fmt.Sprintf(`module phase09.convention_override

export {
  fn main
}

fn main(x: Byte) -> Byte {
  %s
}
`, body)
	return []byte(source)
}

// TestConventionOverrideNotExpressibleInSource is D-09-34's structural-
// absence proof at the source layer: for every convention-shaped keyword,
// attempted as a call-argument annotation in every call-parse route the
// grammar has, syntax.Parse must refuse with syntax.expected_call_argument
// -- proving the grammar never admits the shape, not merely that today's
// fixtures happen not to exercise it.
func TestConventionOverrideNotExpressibleInSource(t *testing.T) {
	routes := []string{"bare_call", "try_call", "discard_call"}
	for _, route := range routes {
		for _, keyword := range ownershipShapedKeywords {
			t.Run(route+"/"+keyword, func(t *testing.T) {
				source := callParseRouteSource(route, keyword)
				parsed := syntax.Parse(source)
				found := false
				for _, problem := range parsed.Diagnostics {
					if problem.Code == "syntax.expected_call_argument" {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected syntax.expected_call_argument refusing %q as a call-argument annotation in route %q, got diagnostics %+v", keyword, route, parsed.Diagnostics)
				}
			})
		}
	}
}

// TestOwnershipShapedKeywordVocabularyIsClosed is this test's own tripwire
// against silently going stale: it statically re-derives the set of
// identifier-typed keywords parser.go's normalizeOwnershipTokens retexts
// (TokenLet/TokenTake/TokenBorrow/TokenMut's own source keywords) and
// asserts it equals EXACTLY {"let", "take", "borrow", "mut"}. If a future
// plan adds a new ownership-shaped keyword to that switch, this test's own
// explicit expected set stops matching -- forcing a human to extend
// ownershipShapedKeywords above (and therefore
// TestConventionOverrideNotExpressibleInSource's coverage) rather than
// letting a new hole slip past unexercised. This is a source-text scan
// (like the corevalidate package's own import-independence guards), not a
// runtime behavior assertion, because the keyword set is compile-time fixed
// vocabulary, not a runtime value.
func TestOwnershipShapedKeywordVocabularyIsClosed(t *testing.T) {
	source, err := readParserSource(t)
	if err != nil {
		t.Fatalf("read parser.go: %v", err)
	}
	start := strings.Index(source, "func normalizeOwnershipTokens")
	if start < 0 {
		t.Fatal("normalizeOwnershipTokens not found in parser.go -- has it been renamed or removed?")
	}
	// Bound the scan to this one function's body by finding its closing
	// brace at column 0 (the next top-level declaration).
	rest := source[start:]
	end := strings.Index(rest, "\n}\n")
	if end < 0 {
		t.Fatal("could not find the end of normalizeOwnershipTokens")
	}
	body := rest[:end]

	found := map[string]bool{}
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "case \"") {
			continue
		}
		value := strings.TrimPrefix(trimmed, "case \"")
		value = strings.TrimSuffix(value, "\":")
		// Punctuation retext cases ("<", ">") are unrelated to ownership
		// keywords; only alphabetic case labels are ownership-shaped.
		if value == "" || !isAlpha(value) {
			continue
		}
		found[value] = true
	}

	want := map[string]bool{"let": true, "take": true, "borrow": true, "mut": true}
	if len(found) != len(want) {
		t.Fatalf("ownership-shaped keyword vocabulary changed: got %v, want %v -- extend ownershipShapedKeywords and this test's expected set together", found, want)
	}
	for keyword := range want {
		if !found[keyword] {
			t.Fatalf("expected normalizeOwnershipTokens to retext keyword %q, got %v", keyword, found)
		}
	}
}

func isAlpha(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}

// TestCallArgumentGrammarAdmitsBareIdentifiersOnly is D-09-34's stronger,
// structural half: a parse refusal alone proves only that today's parser
// happens to decline a convention-annotated argument; this test proves
// there is nowhere for the fact to live even if a future parser change
// admitted the token sequence. callArguments() returns a bare []string
// (never a richer per-argument struct), and ast.RHS -- the only struct a
// call's Arguments field lands in -- declares no field whose name contains
// "mode", "convention", or "ownership" (case-insensitive). A future
// additive field that WOULD create the override slot fails this test
// loudly, in its own diff, rather than silently creating a hole.
func TestCallArgumentGrammarAdmitsBareIdentifiersOnly(t *testing.T) {
	rhsType := reflect.TypeOf(ast.RHS{})
	forbidden := []string{"mode", "convention", "ownership"}
	for i := 0; i < rhsType.NumField(); i++ {
		field := rhsType.Field(i)
		lowered := strings.ToLower(field.Name)
		for _, term := range forbidden {
			if strings.Contains(lowered, term) {
				t.Fatalf("ast.RHS.%s's name contains %q -- a per-argument convention override slot now exists; D-09-34's structural-absence claim is broken", field.Name, term)
			}
		}
	}

	argumentsField, ok := rhsType.FieldByName("Arguments")
	if !ok {
		t.Fatal("ast.RHS no longer declares an Arguments field")
	}
	if argumentsField.Type.Kind() != reflect.Slice || argumentsField.Type.Elem().Kind() != reflect.String {
		t.Fatalf("ast.RHS.Arguments changed shape: got %s, want []string -- a widened element type is exactly how a per-argument override slot would be created", argumentsField.Type)
	}
}
