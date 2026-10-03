package session_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func TestPhase27OperatorSyntax(t *testing.T) {
	source := []byte("module phase27_operators\nexport { fn main }\nfn main(n: U64) -> Unit {\n  // remainder binds more tightly than equality\n  if n % 5 == 0 {\n    write_u64(n)\n  } else {\n    write_newline()\n  }\n}\n")
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("operator source diagnostics: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) != 1 || checked.Program.Functions[0].ReturnType != "Unit" {
		t.Fatalf("checked program functions: %+v", checked.Program.Functions)
	}
	function := checked.Program.Functions[0]
	places := make(map[string]core.Place, len(function.Linear.Places))
	types := make(map[string]string, len(function.Linear.Types))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	for _, fact := range function.Linear.Types {
		types[fact.ID] = fact.Shape.Constructor
	}
	var equal, remainder *core.LinearOperation
	for index := range function.Linear.Operations {
		op := &function.Linear.Operations[index]
		switch op.Kind {
		case core.OpEqualU64:
			equal = op
		case core.OpRemainderU64:
			remainder = op
		}
	}
	if equal == nil || remainder == nil {
		t.Fatalf("typed equality/remainder operations missing: %+v", function.Linear.Operations)
	}
	if types[places[equal.TargetID].TypeID] != "Bool" || types[places[remainder.TargetID].TypeID] != "U64" || equal.SourceID != remainder.TargetID {
		t.Fatalf("operator result types or precedence are wrong: equal=%+v remainder=%+v", equal, remainder)
	}

	first := session.Format(source)
	second := session.Format(first.Canonical)
	if len(first.Diagnostics) != 0 || len(second.Diagnostics) != 0 || !bytes.Equal(first.Canonical, second.Canonical) || !bytes.Contains(first.Canonical, []byte("// remainder binds more tightly than equality")) || !bytes.Contains(first.Canonical, []byte("n % 5 == 0")) {
		t.Fatalf("operator formatting changed or lost source: first=%q second=%q diagnostics=%+v/%+v", first.Canonical, second.Canonical, first.Diagnostics, second.Diagnostics)
	}
	for _, token := range syntax.Parse(first.Canonical).Tree.Tokens {
		if token.Kind != syntax.TokenPercent && token.Kind != syntax.TokenEqualEqual {
			continue
		}
		if token.Span.End <= token.Span.Start || string(first.Canonical[token.Span.Start:token.Span.End]) != token.Text {
			t.Errorf("operator %q lost its source span: %+v", token.Text, token.Span)
		}
	}

	wrongType := []byte("module phase27_wrong_operator_type\nexport { fn main }\nfn main(n: U64) -> U64 {\n  var flag = n < 3\n  var bad = flag % 2\n  n\n}\n")
	if wrong := session.Check(wrongType); len(wrong.Diagnostics) == 0 || wrong.Diagnostics[0].Code != "type.scalar_operand" {
		t.Fatalf("Bool remainder operand was not rejected: %+v", wrong.Diagnostics)
	}
	malformed := []byte(strings.Replace(string(source), "n % 5 == 0", "n % )", 1))
	parsed := syntax.Parse(malformed)
	if len(parsed.Diagnostics) == 0 || parsed.Diagnostics[0].Primary.End <= parsed.Diagnostics[0].Primary.Start {
		t.Fatalf("malformed operator expression did not recover with a span: %+v", parsed.Diagnostics)
	}
}

func TestPhase27StaticZero(t *testing.T) {
	source := []byte("module phase27_static_zero\nexport { fn main }\nfn main(n: U64) -> U64 {\n  let rem = n % 0\n  n\n}\n")
	checked := session.Check(source)
	if len(checked.Diagnostics) == 0 || checked.Diagnostics[0].Code != "type.scalar_remainder_zero" {
		t.Fatalf("literal zero remainder divisor was not rejected: %+v", checked.Diagnostics)
	}
	zeroOffset := strings.Index(string(source), "n % 0") + len("n % ")
	if checked.Diagnostics[0].Primary.Start != zeroOffset || checked.Diagnostics[0].Primary.End != zeroOffset+1 {
		t.Fatalf("literal zero diagnostic span=%+v, want divisor byte [%d,%d)", checked.Diagnostics[0].Primary, zeroOffset, zeroOffset+1)
	}
}
