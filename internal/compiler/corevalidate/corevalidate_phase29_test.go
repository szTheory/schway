package corevalidate_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase29AdlerPrefix = "s1:testdata.phase29.adler32_library:fn:adler32:"

// phase29Forge deep-copies the checked adler32 library through its canonical
// JSON form and hands the copy's adler32 function to mutate. The checked
// program is never edited in place and no production code is touched.
func phase29Forge(t *testing.T, mutate func(f *core.Function)) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase29", "adler32_library.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("adler32 library diagnostics: %+v", checked.Diagnostics)
	}
	raw, err := json.Marshal(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	var program core.Program
	if err := json.Unmarshal(raw, &program); err != nil {
		t.Fatal(err)
	}
	for i := range program.Functions {
		if program.Functions[i].Name == "adler32" {
			mutate(&program.Functions[i])
			return program
		}
	}
	t.Fatal("adler32 not found")
	return program
}

func phase29Op(f *core.Function, suffix string) *core.LinearOperation {
	for i := range f.Linear.Operations {
		if f.Linear.Operations[i].ID == phase29AdlerPrefix+suffix {
			return &f.Linear.Operations[i]
		}
	}
	panic("operation not found: " + suffix)
}

// phase29InsertFirst inserts op as the first operation of the named block and
// renumbers every operation ID / point ID so corevalidate's operation-order
// law still holds; the forged operation is the only semantic change.
func phase29InsertFirst(f *core.Function, block string, op core.LinearOperation) {
	for i := range f.Linear.Blocks {
		if f.Linear.Blocks[i].ID != phase29AdlerPrefix+block {
			continue
		}
		first := f.Linear.Blocks[i].OperationIDs[0]
		f.Linear.Blocks[i].OperationIDs = append([]string{op.ID}, f.Linear.Blocks[i].OperationIDs...)
		for j := range f.Linear.Operations {
			if f.Linear.Operations[j].ID != first {
				continue
			}
			ops := append([]core.LinearOperation(nil), f.Linear.Operations[:j]...)
			ops = append(ops, op)
			f.Linear.Operations = append(ops, f.Linear.Operations[j:]...)
			renumbered := map[string]string{}
			for k := range f.Linear.Operations {
				id := fmt.Sprintf("%s:op:%d", f.ID, k)
				renumbered[f.Linear.Operations[k].ID] = id
				f.Linear.Operations[k].ID = id
				f.Linear.Operations[k].PointID = fmt.Sprintf("%s:point:linear:%d", f.ID, k)
			}
			for b := range f.Linear.Blocks {
				for k, id := range f.Linear.Blocks[b].OperationIDs {
					f.Linear.Blocks[b].OperationIDs[k] = renumbered[id]
				}
			}
			return
		}
	}
	panic("block not found: " + block)
}

func phase29AddViewPlace(f *core.Function) {
	f.Linear.Places = append(f.Linear.Places, core.Place{ID: phase29AdlerPrefix + "place:34", Name: "view", TypeID: phase29AdlerPrefix + "type:0", Mutable: true})
}

// phase29FirstCode is the first problem code, or "" for an accepted program.
func phase29FirstCode(result corevalidate.Result) string {
	if result.Valid {
		return ""
	}
	return result.Problems[0].Code
}

func TestPhase29CorevalidateByteOpLaws(t *testing.T) {
	p := phase29AdlerPrefix
	if result := corevalidate.Validate(phase29Forge(t, func(*core.Function) {})); !result.Valid {
		t.Fatalf("positive control (unforged adler32) rejected: %+v", result.Problems)
	}
	rows := []struct {
		name    string
		forge   func(f *core.Function)
		want    string
		wantOff string
		// armed is true when the disabled law changes the observed code. The
		// two rows that stay refused with the same code are also refused by an
		// independent block-replay arm that re-derives the operand types.
		armed bool
	}{
		{"bytes_at source is a U64 place", func(f *core.Function) { phase29Op(f, "op:7").SourceID = p + "place:1" }, "core.scalar_operation_invalid", "core.scalar_operation_invalid", false},
		{"bytes_at missing RightID", func(f *core.Function) { phase29Op(f, "op:7").RightID = "" }, "core.scalar_operation_invalid", "core.place_uninitialized", true},
		{"bytes_at non-U64 RightID", func(f *core.Function) { phase29Op(f, "op:7").RightID = p + "place:6" }, "core.scalar_operation_invalid", "core.scalar_operation_invalid", false},
		{"bytes_at non-U64 target", func(f *core.Function) { phase29Op(f, "op:7").TargetID = p + "place:6" }, "core.scalar_operation_invalid", "core.place_uninitialized", true},
		{"bytes_len target is the parameter place", func(f *core.Function) { phase29Op(f, "op:3").TargetID = p + "place:0" }, "core.scalar_operation_invalid", "core.place_uninitialized", true},
		{"bytes_at stray StoreTargetID", func(f *core.Function) { phase29Op(f, "op:7").StoreTargetID = p + "place:1" }, "core.scalar_operation_invalid", "core.scalar_field_kind_exclusive", true},
		{"bytes_at stray edge field", func(f *core.Function) { phase29Op(f, "op:7").TrueEdgeID = p + "edge:1" }, "core.scalar_operation_invalid", "core.scalar_field_kind_exclusive", true},
		{"bytes_len carries a RightID", func(f *core.Function) { phase29Op(f, "op:3").RightID = p + "place:3" }, "core.scalar_operation_invalid", "core.scalar_field_kind_exclusive", true},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			program := phase29Forge(t, row.forge)
			if got := phase29FirstCode(corevalidate.Validate(program)); got != row.want {
				t.Fatalf("law on: first code = %q, want %q", got, row.want)
			}
			restore := corevalidate.SetDisableByteOpLawForTest(true)
			got := phase29FirstCode(corevalidate.Validate(program))
			restore()
			if got != row.wantOff {
				t.Fatalf("law off (fault armed): first code = %q, want %q", got, row.wantOff)
			}
			if row.armed && got == row.want {
				t.Fatalf("disabling the byte-op law did not change the outcome (%q); the law is not load-bearing for this row", got)
			}
		})
	}
	// The law restores: after every toggle a forged program is refused again.
	if got := phase29FirstCode(corevalidate.Validate(phase29Forge(t, rows[1].forge))); got != "core.scalar_operation_invalid" {
		t.Fatalf("toggle leaked: first code = %q", got)
	}
}

// TestPhase29NeighborSeededFaults forges the four D-08 neighbor rows over the
// checked adler32 core and arms, per row, the existing corevalidate law that
// refuses it. With the law's toggle set the observed result differs (another
// existing code or acceptance), so each law is load-bearing for its row.
func TestPhase29NeighborSeededFaults(t *testing.T) {
	p := phase29AdlerPrefix
	copyView := func(block string) func(f *core.Function) {
		return func(f *core.Function) {
			phase29AddViewPlace(f)
			phase29InsertFirst(f, block, core.LinearOperation{ID: p + "op:57", Kind: core.OpCopy, SourceID: p + "place:0", TargetID: p + "place:34", TypeID: p + "type:0"})
		}
	}
	rows := []struct {
		name    string
		forge   func(f *core.Function)
		want    string
		toggle  func(bool) func()
		wantOff string
	}{
		{"n1 copy of the ByteView parameter", copyView("block:entry"), "core.scalar_backedge_authority", corevalidate.SetDisableBackEdgeAuthorityLawForTest, "core.ability.copy_denied"},
		{"n2 store of the ByteView into a non-parameter place", func(f *core.Function) {
			phase29AddViewPlace(f)
			phase29InsertFirst(f, "block:entry", core.LinearOperation{ID: p + "op:57", Kind: core.OpScalarStore, SourceID: p + "place:0", StoreTargetID: p + "place:34", TypeID: p + "type:0"})
		}, "core.scalar_backedge_authority", corevalidate.SetDisableBackEdgeAuthorityLawForTest, ""},
		{"n3 return carrying the ByteView", func(f *core.Function) {
			phase29Op(f, "op:56").SourceID = p + "place:0"
			phase29Op(f, "op:56").TypeID = p + "type:0"
		}, "core.return_type_mismatch", corevalidate.SetDisableReturnValueLawForTest, ""},
		{"n4 ByteView place carried across the back edge", copyView("block:while-body:2"), "core.scalar_backedge_authority", corevalidate.SetDisableBackEdgeAuthorityLawForTest, "core.ability.copy_denied"},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			program := phase29Forge(t, row.forge)
			if got := phase29FirstCode(corevalidate.Validate(program)); got != row.want {
				t.Fatalf("law on: first code = %q, want %q", got, row.want)
			}
			restore := row.toggle(true)
			got := phase29FirstCode(corevalidate.Validate(program))
			restore()
			if got != row.wantOff || got == row.want {
				t.Fatalf("law off (fault armed): first code = %q, want %q (must differ from %q)", got, row.wantOff, row.want)
			}
		})
	}
}
