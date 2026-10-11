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
