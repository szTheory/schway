package pathoracle_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
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
// renumbers every operation ID / point ID so pathoracle's operation-order
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

func phase29Endpoints(program core.Program, f core.Function) ([]core.LoanEndpoint, error) {
	endpoints, _, err := pathoracle.RecomputeEndpoints(f, pathoracle.BuildCalleeLookup(program))
	return endpoints, err
}

// TestPhase29PathoracleByteReadsCreateNoLoans: bytes_len / bytes_at read the
// view into a fresh U64 and never create a loan, so an acyclic ByteView body
// has zero endpoints; a forged unknown source place is still refused by the
// oracle's own place index.
func TestPhase29PathoracleByteReadsCreateNoLoans(t *testing.T) {
	const source = `module phase29.pathoracle.straight
export { fn f }

fn f(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  var x = byte_at(bytes, 0)
  n = n + x
  n
}
`
	checked := session.Check([]byte(source))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}
	var function core.Function
	for _, candidate := range checked.Program.Functions {
		if candidate.Name == "f" {
			function = candidate
		}
	}
	reads := 0
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpBytesLen || operation.Kind == core.OpBytesAt {
			reads++
		}
	}
	if reads != 2 || len(function.Linear.Blocks) == 0 {
		t.Fatalf("fixture shape: %d byte reads, %d blocks", reads, len(function.Linear.Blocks))
	}
	endpoints, err := phase29Endpoints(checked.Program, function)
	if err != nil || len(endpoints) != 0 {
		t.Fatalf("endpoints = %v, err = %v; want none", endpoints, err)
	}

	raw, err := json.Marshal(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	var forged core.Program
	if err := json.Unmarshal(raw, &forged); err != nil {
		t.Fatal(err)
	}
	for fi := range forged.Functions {
		if forged.Functions[fi].Name != "f" {
			continue
		}
		for oi := range forged.Functions[fi].Linear.Operations {
			if forged.Functions[fi].Linear.Operations[oi].Kind == core.OpBytesAt {
				forged.Functions[fi].Linear.Operations[oi].SourceID = "no-such-place"
			}
		}
		_, err := phase29Endpoints(forged, forged.Functions[fi])
		if err == nil || !strings.HasPrefix(err.Error(), "pathoracle.unknown_place") {
			t.Fatalf("forged unknown source: err = %v, want pathoracle.unknown_place", err)
		}
	}
}

// TestPhase29NeighborSeededFaults pins the four D-08 neighbor rows against the
// path oracle. The oracle has no matching law for any of them: it checks only
// cycles, place existence and loan liveness, never operand types, so it
// refuses every cyclic CFG identically (pathoracle.cfg_back_edge), whether or
// not the forged operation is present. Each row is pinned as that unchanged
// outcome with its noLaw reason; refusal of the content itself belongs to the
// corevalidate and originvalidate peers.
func TestPhase29NeighborSeededFaults(t *testing.T) {
	p := phase29AdlerPrefix
	copyView := func(block string) func(f *core.Function) {
		return func(f *core.Function) {
			phase29AddViewPlace(f)
			phase29InsertFirst(f, block, core.LinearOperation{ID: p + "op:57", Kind: core.OpCopy, SourceID: p + "place:0", TargetID: p + "place:34", TypeID: p + "type:0"})
		}
	}
	const noTypeLaw = "pathoracle checks only cycles, place existence and loan liveness, not operand or result types"
	rows := []struct {
		name  string
		forge func(f *core.Function)
		noLaw string
	}{
		{"n1 copy of the ByteView parameter", copyView("block:entry"), noTypeLaw},
		{"n2 store of the ByteView into a non-parameter place", func(f *core.Function) {
			phase29AddViewPlace(f)
			phase29InsertFirst(f, "block:entry", core.LinearOperation{ID: p + "op:57", Kind: core.OpScalarStore, SourceID: p + "place:0", StoreTargetID: p + "place:34", TypeID: p + "type:0"})
		}, noTypeLaw},
		{"n3 return carrying the ByteView", func(f *core.Function) {
			phase29Op(f, "op:56").SourceID = p + "place:0"
			phase29Op(f, "op:56").TypeID = p + "type:0"
		}, noTypeLaw},
		{"n4 ByteView place carried across the back edge", copyView("block:while-body:2"), noTypeLaw},
	}
	controlProgram := phase29Forge(t, func(*core.Function) {})
	var control string
	for _, f := range controlProgram.Functions {
		if f.Name == "adler32" {
			_, err := phase29Endpoints(controlProgram, f)
			if err == nil {
				t.Fatal("unforged adler32 loop unexpectedly accepted by the path oracle")
			}
			control = err.Error()
		}
	}
	if !strings.HasPrefix(control, "pathoracle.cfg_back_edge") {
		t.Fatalf("control = %q, want pathoracle.cfg_back_edge", control)
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			if row.noLaw == "" {
				t.Fatal("every pathoracle row must carry a noLaw reason")
			}
			program := phase29Forge(t, row.forge)
			for _, f := range program.Functions {
				if f.Name != "adler32" {
					continue
				}
				_, err := phase29Endpoints(program, f)
				if err == nil || !strings.HasPrefix(err.Error(), "pathoracle.cfg_back_edge") {
					t.Fatalf("outcome = %v, want the unchanged cyclic refusal %q", err, control)
				}
			}
		})
	}
}
