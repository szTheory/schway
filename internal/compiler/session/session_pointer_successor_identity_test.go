package session

import (
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase25EscapeDiagnosticFunctionIdentity(t *testing.T) {
	source := []byte("module phase25.escape_identity\nexport { fn relay }\nfn owned(value: Buffer) -> Buffer {\n  let result = take value\n  result\n}\nfn relay(value: Buffer) -> Buffer {\n  let result = borrow value\n  result\n}\n")
	checked := Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("identity fixture did not typecheck: %+v", checked.Diagnostics)
	}
	problems := originvalidate.ValidatePublished(checked.Program)
	if len(problems) != 1 || problems[0].Code != "core.origin_omitted" || !strings.Contains(problems[0].FunctionID, ":fn:relay") || problems[0].ReturnOperationID == "" {
		t.Fatalf("origin peer did not publish a structured offending-return identity: %+v", problems)
	}
	first := originPeerRefusalDiagnostic(problems[0], checked.Program, checked.Tree)
	problems[0].Detail = "misleading peer prose names the harmless function owned"
	second := originPeerRefusalDiagnostic(problems[0], checked.Program, checked.Tree)
	if first.Primary != second.Primary || first.ID != second.ID || len(first.Causes) == 0 || len(second.Causes) == 0 || first.Causes[0].Span == nil || second.Causes[0].Span == nil || *first.Causes[0].Span != *second.Causes[0].Span {
		t.Fatalf("changing peer prose redirected structured source selection: first=%+v second=%+v", first, second)
	}
	if got := string(source[first.Primary.Start:first.Primary.End]); got != "result" || first.Primary.Start != strings.LastIndex(string(source), "result") {
		t.Fatalf("structured function identity selected the wrong similar return %q at %d", got, first.Primary.Start)
	}
	span := *first.Causes[0].Span
	if got := string(source[span.Start:span.End]); got != "borrow value" {
		t.Fatalf("structured origin identity selected the wrong borrow %q", got)
	}
	if first.Schema != "lang.diagnostic/0" || first.Code != "core.origin_omitted" {
		t.Fatalf("source projection changed the diagnostic wire contract: %+v", first)
	}
}

func TestPhase25CheckCommandPeerObservation(t *testing.T) {
	previous := checkCommandPeerObservedForTest
	var observed []string
	checkCommandPeerObservedForTest = func(name string) { observed = append(observed, name) }
	t.Cleanup(func() { checkCommandPeerObservedForTest = previous })
	result, err := CheckCommandFile(testsupport.ProjectPath("testdata", "phase3", "sequential_shared_then_exclusive_accept.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "pass" {
		t.Fatalf("accepted source did not reach all command peers: %+v", result)
	}
	want := []string{"corevalidate", "originvalidate", "pathoracle"}
	if len(observed) != len(want) {
		t.Fatalf("command consulted peers %v, want %v", observed, want)
	}
	for i := range want {
		if observed[i] != want[i] {
			t.Fatalf("command peer order is %v, want %v", observed, want)
		}
	}
}
