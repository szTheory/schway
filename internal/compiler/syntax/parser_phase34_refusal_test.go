package syntax_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func TestPhase34UnsupportedSurfaceRefusals(t *testing.T) {
	const source = "module phase34.logical_and\nexport { fn main }\nfn main(n: U64) -> U64 { if n < 2 && n < 3 { n } else { n } }\n"
	const hint = "Logical && is unsupported; use explicit if/else control flow."

	parsed := syntax.Parse([]byte(source))
	if !bytes.Equal(parsed.Tree.Bytes(), []byte(source)) {
		t.Fatal("refused logical-and source did not remain lossless")
	}
	if len(parsed.Diagnostics) != 1 {
		t.Fatalf("logical-and diagnostics = %+v, want one dedicated refusal", parsed.Diagnostics)
	}
	problem := parsed.Diagnostics[0]
	start := strings.Index(source, "&&")
	if problem.Code != "surface.unsupported_logical_and" || problem.Message != hint || problem.Primary.Start != start || problem.Primary.End != start+2 {
		t.Fatalf("logical-and diagnostic = %+v, want code surface.unsupported_logical_and at [%d:%d] with hint %q", problem, start, start+2, hint)
	}

	checked := session.Check([]byte(source))
	if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "surface.unsupported_logical_and" {
		t.Fatalf("session.Check diagnostics = %+v, want only the logical-and refusal", checked.Diagnostics)
	}
	if !reflect.DeepEqual(checked.Program, core.Program{}) {
		t.Fatalf("session.Check admitted rejected syntax: %+v", checked.Program)
	}
}
