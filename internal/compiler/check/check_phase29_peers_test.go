package check

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

const phase29AdlerPrefix = "s1:testdata.phase29.adler32_library:fn:adler32:"

func phase29CheckSource(t *testing.T, source []byte) (core.Program, []string) {
	t.Helper()
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		codes := make([]string, 0, len(parsed.Diagnostics))
		for _, diagnostic := range parsed.Diagnostics {
			codes = append(codes, diagnostic.Code)
		}
		return core.Program{}, codes
	}
	checked := Program(parsed.Program)
	codes := make([]string, 0, len(checked.Diagnostics))
	for _, diagnostic := range checked.Diagnostics {
		codes = append(codes, diagnostic.Code)
	}
	return checked.Program, codes
}

func phase29CheckedLibrary(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile("../../../testdata/phase29/adler32_library.schway")
	if err != nil {
		t.Fatal(err)
	}
	program, codes := phase29CheckSource(t, source)
	if len(codes) != 0 {
		t.Fatalf("adler32 library diagnostics: %v", codes)
	}
	return program
}

// phase29ForgedLibrary deep-copies the checked adler32 core through its
// canonical JSON form and lets mutate edit the copy's adler32 function.
func phase29ForgedLibrary(t *testing.T, mutate func(f *core.Function)) (core.Program, core.Function) {
	t.Helper()
	raw, err := json.Marshal(phase29CheckedLibrary(t))
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
			return program, program.Functions[i]
		}
	}
	t.Fatal("adler32 not found")
	return program, core.Function{}
}

// TestPhase29PeerIndependence: the checked adler32 core is accepted by check's
// own fixed point and by both independent validators; a forged bytes_at whose
// source is a U64 place is refused by each validator with its own code. The
// checker never reads a forged core -- its law is the builder's source-level
// typing -- so the source-level counterpart (byte_at over a U64 operand) is
// pinned beside it.
func TestPhase29PeerIndependence(t *testing.T) {
	genuine, function := phase29ForgedLibrary(t, func(*core.Function) {})
	if _, code := checkScalarFixedPoint(function); code != "" {
		t.Fatalf("check fixed point refused the genuine library: %q", code)
	}
	if result := corevalidate.Validate(genuine); !result.Valid {
		t.Fatalf("corevalidate refused the genuine library: %+v", result.Problems)
	}
	if problems := originvalidate.ValidatePublished(genuine); len(problems) != 0 {
		t.Fatalf("originvalidate refused the genuine library: %+v", problems)
	}

	forged, forgedFunction := phase29ForgedLibrary(t, func(f *core.Function) {
		for i := range f.Linear.Operations {
			if f.Linear.Operations[i].Kind == core.OpBytesAt {
				f.Linear.Operations[i].SourceID = phase29AdlerPrefix + "place:1"
			}
		}
	})
	if result := corevalidate.Validate(forged); result.Valid || result.Problems[0].Code != "core.scalar_operation_invalid" {
		t.Fatalf("corevalidate on forged bytes_at: %+v", result)
	}
	if problems := originvalidate.ValidatePublished(forged); len(problems) == 0 || problems[0].Code != "core.scalar_operation_invalid" {
		t.Fatalf("originvalidate on forged bytes_at: %+v", problems)
	}
	// check's fixed point is an initialization/authority analysis over cores
	// it built itself: it has no operand-type law for byte reads, so a
	// type-forged bytes_at passes it. That gap is pinned here (not hidden) and
	// is exactly why the two validators each carry their own law.
	if _, code := checkScalarFixedPoint(forgedFunction); code != "" {
		t.Fatalf("check fixed point on the forged core = %q; the pinned gap closed, move the refusal into the assertions above", code)
	}
	// Source-level counterpart: the checker refuses a byte read whose operand
	// is not a ByteView by itself.
	_, codes := phase29CheckSource(t, []byte("module phase29.peer.u64\nexport { fn f }\n\nfn f(n: U64) -> U64 {\n  var x = byte_at(n, 0)\n  x\n}\n"))
	if len(codes) == 0 || !strings.HasPrefix(codes[0], "type.") {
		t.Fatalf("byte_at over a U64 operand: codes = %v, want a type.* refusal first", codes)
	}
}

func phase29LoopSource(params, result, body string) []byte {
	return []byte("module phase29.neighbor\nexport { fn f }\n\nfn f(" + params + ") -> " + result + " {\n" + body + "}\n")
}

// TestPhase29NeighborRefusals pins, with ByteView landed, the first
// diagnostic of every loop-function neighbor (D-08). Each source is var-first
// so it reaches the scalar parser without the let-first widening.
func TestPhase29NeighborRefusals(t *testing.T) {
	const loopTail = "  var i = 0\n  while i < n {\n    i = i + 1\n  }\n  i\n"
	cases := []struct {
		name   string
		source []byte
		code   string
	}{
		{"borrow inside the loop", phase29LoopSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n  var i = 0\n  while i < n {\n    let b = borrow n\n    i = i + 1\n  }\n  i\n"), "syntax.expected_scalar_operand"},
		{"take inside the loop", phase29LoopSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n  var i = 0\n  while i < n {\n    let t = take n\n    i = i + 1\n  }\n  i\n"), "syntax.expected_scalar_operand"},
		{"Buffer parameter", phase29LoopSource("b: Buffer", "U64", "  var n = 1\n"+loopTail), "type.scalar_signature"},
		{"Buffer result", phase29LoopSource("bytes: ByteView", "Buffer", "  var n = byte_len(bytes)\n"+loopTail), "type.scalar_signature"},
		{"ByteView result", phase29LoopSource("bytes: ByteView", "ByteView", "  var n = byte_len(bytes)\n"+loopTail), "type.scalar_signature"},
		{"copying the view", phase29LoopSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n  let c = bytes\n"+loopTail), "type.scalar_binding"},
		{"storing the view in a var", phase29LoopSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n  var s = bytes\n"+loopTail), "type.scalar_binding"},
		{"assigning the view", phase29LoopSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n  var s = 0\n  s = bytes\n"+loopTail), "type.scalar_assignment"},
		{"returning the view", phase29LoopSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n  var i = 0\n  while i < n {\n    i = i + 1\n  }\n  bytes\n"), "type.scalar_result"},
	}
	// Positive control: the admitted adler32 shape checks clean, and so does a
	// minimal var-first loop of the same form as every neighbor above.
	control := phase29LoopSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n"+loopTail)
	if _, codes := phase29CheckSource(t, control); len(codes) != 0 {
		t.Fatalf("positive control diagnostics: %v", codes)
	}
	phase29CheckedLibrary(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, codes := phase29CheckSource(t, tc.source)
			if len(codes) == 0 || codes[0] != tc.code {
				t.Fatalf("diagnostics = %v, want first %s", codes, tc.code)
			}
		})
	}
	// A forged loan carried across the back edge keeps check.cfg_back_edge with
	// the loan category, and a ByteView place carried outside the parameter
	// place is classified with the existing view/loan category.
	t.Run("forged loan parameter", func(t *testing.T) {
		_, function := phase29ForgedLibrary(t, func(f *core.Function) {
			phase29PrependOp(f, "block:while-body:2", core.LinearOperation{ID: phase29AdlerPrefix + "op:57", Kind: core.OpBorrowShared, SourceID: phase29AdlerPrefix + "place:0", TargetID: phase29AdlerPrefix + "place:34", LoanID: "loan:forged", TypeID: phase29AdlerPrefix + "type:0"})
		})
		if _, code, category := checkScalarFixedPointDetailed(function); code != "check.cfg_back_edge" || category != "loan" {
			t.Fatalf("code = %q category = %q, want check.cfg_back_edge / loan", code, category)
		}
	})
	t.Run("ByteView carried across the back edge", func(t *testing.T) {
		_, function := phase29ForgedLibrary(t, func(f *core.Function) {
			phase29PrependOp(f, "block:while-body:2", core.LinearOperation{ID: phase29AdlerPrefix + "op:57", Kind: core.OpCopy, SourceID: phase29AdlerPrefix + "place:0", TargetID: phase29AdlerPrefix + "place:34", TypeID: phase29AdlerPrefix + "type:0"})
		})
		if _, code, category := checkScalarFixedPointDetailed(function); code != "check.cfg_back_edge" || category != "loan_derived_provenance" {
			t.Fatalf("code = %q category = %q, want check.cfg_back_edge / loan_derived_provenance", code, category)
		}
	})
}

// phase29PrependOp adds a ByteView place and puts op at the head of the block.
// check's fixed point resolves operations by ID, so no renumbering is needed.
func phase29PrependOp(f *core.Function, block string, op core.LinearOperation) {
	f.Linear.Places = append(f.Linear.Places, core.Place{ID: phase29AdlerPrefix + "place:34", Name: "view", TypeID: phase29AdlerPrefix + "type:0", Mutable: true})
	f.Linear.Operations = append(f.Linear.Operations, op)
	for i := range f.Linear.Blocks {
		if f.Linear.Blocks[i].ID == phase29AdlerPrefix+block {
			f.Linear.Blocks[i].OperationIDs = append([]string{op.ID}, f.Linear.Blocks[i].OperationIDs...)
		}
	}
}

// phase29AdlerMutantSource is the adler32 library body with its per-byte read
// bound through an in-loop `let step`. The function stays var-first (its first
// statement is `var a = 1`), so the in-loop `let` is parsed by scalarStatements
// and the mutant needs no let-first widening.
const phase29AdlerMutantSource = `module phase29.events.mutant

export {
  fn adler32
}

fn adler32(bytes: ByteView) -> U64 {
  var a = 1
  var b = 0
  var i = 0
  var n = byte_len(bytes)
  while i < n {
    let step = byte_at(bytes, i)
    a = a + step
    a = a % 65521
    b = b + a
    b = b % 65521
    i = i + 1
  }
  var r = b
  r = r + r
  r = r + a
  r
}
`

// phase29LoopOperations returns every operation in a block that lies on a
// cycle through a back edge: reachable from the back edge's target (the loop
// header) and able to reach the back edge's source.
func phase29LoopOperations(f core.Function) []core.LinearOperation {
	succ := map[string][]string{}
	pred := map[string][]string{}
	for _, edge := range f.Linear.Edges {
		succ[edge.FromBlockID] = append(succ[edge.FromBlockID], edge.ToBlockID)
		pred[edge.ToBlockID] = append(pred[edge.ToBlockID], edge.FromBlockID)
	}
	reach := func(start string, next map[string][]string) map[string]bool {
		seen := map[string]bool{start: true}
		queue := []string{start}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			for _, n := range next[current] {
				if !seen[n] {
					seen[n] = true
					queue = append(queue, n)
				}
			}
		}
		return seen
	}
	inLoop := map[string]bool{}
	for _, edge := range f.Linear.Edges {
		if edge.Pattern != "back" {
			continue
		}
		forward, backward := reach(edge.ToBlockID, succ), reach(edge.FromBlockID, pred)
		for id := range forward {
			if backward[id] {
				inLoop[id] = true
			}
		}
	}
	operations := map[string]core.LinearOperation{}
	for _, operation := range f.Linear.Operations {
		operations[operation.ID] = operation
	}
	var out []core.LinearOperation
	for _, block := range f.Linear.Blocks {
		if !inLoop[block.ID] {
			continue
		}
		for _, id := range block.OperationIDs {
			out = append(out, operations[id])
		}
	}
	return out
}

func phase29CountKind(operations []core.LinearOperation, kind core.OperationKind) int {
	count := 0
	for _, operation := range operations {
		if operation.Kind == kind {
			count++
		}
	}
	return count
}

// TestPhase29NoLoopEventsStructure: the admitted adler32 loop contains no
// OpCopy (an OpCopy records a value.copied event per execution), and the
// let-in-loop mutant is observed to contain one, so the walk is not inert.
func TestPhase29NoLoopEventsStructure(t *testing.T) {
	adler := func(program core.Program) core.Function {
		for _, f := range program.Functions {
			if f.Name == "adler32" {
				return f
			}
		}
		t.Fatal("adler32 not found")
		return core.Function{}
	}
	library := adler(phase29CheckedLibrary(t))
	loop := phase29LoopOperations(library)
	if len(loop) == 0 || phase29CountKind(loop, core.OpBytesAt) != 1 {
		t.Fatalf("loop walk found %d operations and %d bytes_at; the walk is mis-aimed", len(loop), phase29CountKind(loop, core.OpBytesAt))
	}
	if copies := phase29CountKind(loop, core.OpCopy); copies != 0 {
		t.Fatalf("adler32 loop blocks contain %d OpCopy operations, want 0", copies)
	}

	// Armed control: the mutant must parse cleanly, check, and carry a copy.
	parsed := syntax.Parse([]byte(phase29AdlerMutantSource))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("mutant parse diagnostics: %+v", parsed.Diagnostics)
	}
	checked := Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("mutant check diagnostics: %+v", checked.Diagnostics)
	}
	mutantLoop := phase29LoopOperations(adler(checked.Program))
	if copies := phase29CountKind(mutantLoop, core.OpCopy); copies == 0 {
		t.Fatal("the let-in-loop mutant has no OpCopy in its loop; the structural walk is inert")
	}
}
