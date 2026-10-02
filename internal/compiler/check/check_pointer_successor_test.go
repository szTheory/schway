package check_test

import (
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase25SharedSource(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase25", "shared_copy_accept.schway"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("shared-copy fixture failed to parse: %+v", parsed.Diagnostics)
	}
	result := check.Program(parsed.Program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("ordinary shared read/copy helper was refused: %+v", result.Diagnostics)
	}
	if len(result.Program.Functions) != 1 {
		t.Fatalf("checked function count=%d, want 1", len(result.Program.Functions))
	}
	function := result.Program.Functions[0]
	if function.Name != "shared_copy" || function.Parameter.Type != "U64" || function.ReturnType != "U64" {
		t.Fatalf("unexpected helper contract: %+v", function)
	}
	if function.Linear == nil || len(function.Linear.Operations) != 2 {
		t.Fatalf("helper must contain one shared borrow and one return: %+v", function.Linear)
	}
	borrow, ret := function.Linear.Operations[0], function.Linear.Operations[1]
	if borrow.Kind != core.OpBorrowShared || borrow.SourceID != function.Parameter.ID || borrow.LoanID == "" {
		t.Fatalf("checked helper lost its explicit shared-family fact: %+v", borrow)
	}
	if ret.Kind != core.OpReturn || ret.SourceID != borrow.TargetID {
		t.Fatalf("helper must return the copied shared value: %+v", ret)
	}
}

func TestPhase25ExclusiveSource(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase25", "exclusive_copy_accept.schway"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("exclusive-copy fixture failed to parse: %+v", parsed.Diagnostics)
	}
	result := check.Program(parsed.Program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("ordinary exclusive read/copy helper was refused: %+v", result.Diagnostics)
	}
	if len(result.Program.Functions) != 1 {
		t.Fatalf("checked function count=%d, want 1", len(result.Program.Functions))
	}
	function := result.Program.Functions[0]
	if function.Name != "exclusive_copy" || function.Parameter.Type != "U64" || function.ReturnType != "U64" {
		t.Fatalf("unexpected helper contract: %+v", function)
	}
	if function.Linear == nil || len(function.Linear.Operations) != 3 {
		t.Fatalf("helper must contain one exclusive borrow, one copy, and one return: %+v", function.Linear)
	}
	borrow, copyValue, ret := function.Linear.Operations[0], function.Linear.Operations[1], function.Linear.Operations[2]
	if borrow.Kind != core.OpBorrowExclusive || borrow.SourceID != function.Parameter.ID || borrow.LoanID == "" {
		t.Fatalf("checked helper lost its explicit exclusive-family fact: %+v", borrow)
	}
	if copyValue.Kind != core.OpCopy || copyValue.SourceID != borrow.TargetID || ret.Kind != core.OpReturn || ret.SourceID != copyValue.TargetID {
		t.Fatalf("helper must copy the exclusive value before returning it: copy=%+v return=%+v", copyValue, ret)
	}
}

func TestPhase25TransferCallerComposition(t *testing.T) {
	transfer, err := os.ReadFile(testsupport.ProjectPath("examples", "phase24", "transfer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	source := strings.Replace(string(transfer), "  let value = try schway_file_byte_use(owner)\n  value", "  let value = try schway_file_byte_use(owner)\n  let shared = shared_copy(value)\n  let exclusive = exclusive_copy(shared)\n  exclusive", 1)
	source = strings.Replace(source, "export {\n  fn main\n}", "export {\n  fn main\n  fn shared_copy\n  fn exclusive_copy\n}", 1)
	source += `
fn shared_copy(value: U64) -> U64 {
  let borrowed = borrow value
  let copied = borrowed
  copied
}

fn exclusive_copy(value: U64) -> U64 {
  let borrowed = borrow mut value
  let copied = borrowed
  copied
}
`
	if source == string(transfer) {
		t.Fatal("caller composition replacement did not match the Phase 24 source")
	}
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("composed transfer source failed to parse: %+v", parsed.Diagnostics)
	}
	result := check.Program(parsed.Program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("bounded transfer caller composition was refused: %+v", result.Diagnostics)
	}
	var main *core.Function
	for index := range result.Program.Functions {
		if result.Program.Functions[index].Name == "main" {
			main = &result.Program.Functions[index]
			break
		}
	}
	if main == nil || main.Linear == nil {
		t.Fatalf("checked composition has no main linear body: %+v", result.Program.Functions)
	}
	want := []core.OperationKind{core.OpCall, core.OpForeignCall, core.OpCall, core.OpCall, core.OpRelease, core.OpReturn}
	if len(main.Linear.Operations) != len(want) {
		t.Fatalf("main operations=%+v, want sequence %v", main.Linear.Operations, want)
	}
	for index, kind := range want {
		if got := main.Linear.Operations[index].Kind; got != kind {
			t.Fatalf("main operation %d kind=%s, want %s; operations=%+v", index, got, kind, main.Linear.Operations)
		}
	}
	if main.Linear.Operations[1].Foreign == nil || main.Linear.Operations[1].Foreign.Fails != "UseError" {
		t.Fatalf("typed owner-use failure is not retained before helpers: %+v", main.Linear.Operations[1])
	}
	helperIDs := make(map[string]string)
	for _, function := range result.Program.Functions {
		if function.Name == "shared_copy" || function.Name == "exclusive_copy" {
			helperIDs[function.Name] = function.ID
		}
	}
	if main.Linear.Operations[2].CalleeID != helperIDs["shared_copy"] {
		t.Fatalf("shared helper result did not enter composition: %+v", main.Linear.Operations[2])
	}
	if main.Linear.Operations[3].CalleeID != helperIDs["exclusive_copy"] {
		t.Fatalf("exclusive helper result did not enter composition: %+v", main.Linear.Operations[3])
	}
	if main.Linear.Operations[3].SourceID != main.Linear.Operations[2].TargetID {
		t.Fatalf("shared result does not feed exclusive helper: shared=%+v exclusive=%+v", main.Linear.Operations[2], main.Linear.Operations[3])
	}
	if main.Linear.Operations[5].SourceID != main.Linear.Operations[3].TargetID {
		t.Fatalf("exclusive result does not feed return: exclusive=%+v return=%+v", main.Linear.Operations[3], main.Linear.Operations[5])
	}
}
