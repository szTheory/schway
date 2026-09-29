package debugmap_test

import (
	"context"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/debugmap"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// branchSource is the branch fixture this package's tests join against: it
// carries a match arm body with a shared borrow, a move, and a return, all
// in one arm — the "at minimum one match arm, one move, one borrow, and one
// return" the plan requires.
const branchSource = `module debugmap.sample

export {
  type Switch
  fn choose
}

data Switch =
  | On
  | Off

fn choose(flag: Switch) -> Switch {
  match flag {
    On => {
      let view = borrow flag
      let noted = view
      let moved = take flag
      moved
    }
    Off => {
      let moved = take flag
      moved
    }
  }
}
`

func checkedProgram(t testing.TB, source string) (ast.Program, core.Program) {
	t.Helper()
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("unexpected parse diagnostics: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected check diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("expected valid core, got %+v", validated.Problems)
	}
	return parsed.Program, validated.Program()
}

// TestDebugMapJoinsSourceCoreAndOperation is the plan's central falsifier:
// a side table maps each source span to its core ID and to the operation,
// point, or edge identity it produced, covering at minimum one match arm,
// one move, one borrow, and one return.
func TestDebugMapJoinsSourceCoreAndOperation(t *testing.T) {
	astProgram, coreProgram := checkedProgram(t, branchSource)
	built, work, err := debugmap.Build(context.Background(), astProgram, coreProgram)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if work == 0 || work != len(built.Entries) {
		t.Fatalf("expected work to count entries 1:1, work=%d entries=%d", work, len(built.Entries))
	}
	if built.Schema != debugmap.Schema {
		t.Fatalf("schema=%q want=%q", built.Schema, debugmap.Schema)
	}

	kinds := make(map[string]bool)
	for _, entry := range built.Entries {
		if entry.Availability != debugmap.Available {
			t.Fatalf("entry %+v: expected available, got %q", entry, entry.Availability)
		}
		if entry.OperationID == "" || entry.PointID == "" || entry.CoreID == "" {
			t.Fatalf("entry %+v: joined identities must not be empty", entry)
		}
		if entry.SourceSpan == (diagnostic.Span{}) {
			t.Fatalf("entry %+v: expected a real source span", entry)
		}
		kinds[entry.Kind] = true
	}
	for _, wantKind := range []string{"borrow_shared", "move", "return"} {
		if !kinds[wantKind] {
			t.Fatalf("missing kind %q in joined entries: %+v", wantKind, built.Entries)
		}
	}
}

// TestDebugMapReportsHonestAbsence is D-04's falsifier: a deliberately
// absent value must report not_captured, never available and never a
// guessed span.
func TestDebugMapReportsHonestAbsence(t *testing.T) {
	astProgram, coreProgram := checkedProgram(t, branchSource)
	built, _, err := debugmap.Build(context.Background(), astProgram, coreProgram)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	absent := debugmap.Resolve(built, "this-operation-id-was-never-produced")
	if absent.Availability != debugmap.NotCaptured {
		t.Fatalf("expected not_captured for an absent operation ID, got %+v", absent)
	}
	if absent.SourceSpan != (diagnostic.Span{}) || absent.CoreID != "" || absent.PointID != "" {
		t.Fatalf("not_captured entry must not carry a fabricated span or identity: %+v", absent)
	}

	present := built.Entries[0]
	found := debugmap.Resolve(built, present.OperationID)
	if found.Availability != debugmap.Available || found != present {
		t.Fatalf("expected the real entry back for a real operation ID: got %+v want %+v", found, present)
	}
}

// TestDebugMapCapsFailClosed is D-04's cap falsifier: exceeding the declared
// entry cap fails closed with a stable code rather than truncating.
func TestDebugMapCapsFailClosed(t *testing.T) {
	functionID := "s1:debugmap.cap_test:fn:overflow"
	bindings := make([]ast.Binding, 0, debugmap.MaxEntries+2)
	operations := make([]core.LinearOperation, 0, debugmap.MaxEntries+2)
	for index := 0; index < debugmap.MaxEntries+2; index++ {
		bindings = append(bindings, ast.Binding{Name: "binding", Span: diagnostic.Span{Start: index, End: index + 1}})
		operations = append(operations, core.LinearOperation{
			ID: functionID + ":op:overflow", PointID: functionID + ":point:overflow", Kind: core.OpCopy,
		})
	}
	astProgram := ast.Program{Funcs: []ast.FuncDecl{{
		Name: "overflow",
		Body: ast.Body{Linear: &ast.LinearBody{Bindings: bindings, Result: "binding"}},
	}}}
	coreProgram := core.Program{Functions: []core.Function{{
		ID: functionID, Name: "overflow",
		Linear: &core.LinearBody{Operations: operations},
	}}}
	_, _, err := debugmap.Build(context.Background(), astProgram, coreProgram)
	if err != debugmap.ErrEntryCapExceeded {
		t.Fatalf("expected ErrEntryCapExceeded, got %v", err)
	}
}

// TestDebugMapIdentityUsesOrdinals proves every entry's own identity is a
// function-local semantic ordinal, never a source byte offset — the source
// span is carried as data on the entry, but must never leak into the ID
// itself.
func TestDebugMapIdentityUsesOrdinals(t *testing.T) {
	astProgram, coreProgram := checkedProgram(t, branchSource)
	built, _, err := debugmap.Build(context.Background(), astProgram, coreProgram)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(built.Entries) == 0 {
		t.Fatal("expected at least one entry")
	}
	for index, entry := range built.Entries {
		want := entry.CoreID + ":debug:" + itoa(index)
		if entry.ID != want {
			t.Fatalf("entry %d: ID=%q want=%q (function-local ordinal)", index, entry.ID, want)
		}
	}
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := ""
	for value > 0 {
		digits = string(rune('0'+value%10)) + digits
		value /= 10
	}
	return digits
}
