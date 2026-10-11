package check

import (
	"encoding/json"
	"os"
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
	if len(codes) == 0 || codes[0] != "type.scalar_operand" {
		t.Fatalf("byte_at over a U64 operand: codes = %v, want first type.scalar_operand", codes)
	}
}
