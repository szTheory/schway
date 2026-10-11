package check

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
)

func phase29FixtureSource(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile("../../../examples/phase29/adler32_fixture.schway")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func phase29FixtureMain(t *testing.T, program core.Program) core.Function {
	t.Helper()
	for _, function := range program.Functions {
		if function.Name == "main" {
			return function
		}
	}
	t.Fatal("fixture program has no main")
	return core.Function{}
}

// TestPhase29FixtureMainCoreShape pins the D-15 hand-built core: exactly five
// operations in order, the acquire sourced from the U64 entry parameter, the
// borrow producing a ByteView, the OpCall consuming that view, the release
// minted from the acquire's own release policy, and the return of the call
// result.
func TestPhase29FixtureMainCoreShape(t *testing.T) {
	program, codes := phase29CheckSource(t, []byte(phase29FixtureSource(t)))
	if len(codes) != 0 {
		t.Fatalf("fixture diagnostics: %v", codes)
	}
	main := phase29FixtureMain(t, program)
	if main.Linear == nil || len(main.Linear.Blocks) != 0 || len(main.Linear.Edges) != 0 {
		t.Fatalf("main linear body = %+v, want straight-line", main.Linear)
	}
	wantKinds := []core.OperationKind{core.OpForeignCall, core.OpForeignCall, core.OpCall, core.OpRelease, core.OpReturn}
	operations := main.Linear.Operations
	if len(operations) != len(wantKinds) {
		t.Fatalf("main has %d operations, want %d", len(operations), len(wantKinds))
	}
	for index, want := range wantKinds {
		if operations[index].Kind != want {
			t.Fatalf("operation %d kind = %q, want %q", index, operations[index].Kind, want)
		}
	}
	shapeOf := func(placeID string) string {
		for _, place := range main.Linear.Places {
			if place.ID != placeID {
				continue
			}
			for _, fact := range main.Linear.Types {
				if fact.ID == place.TypeID {
					return fact.Shape.Constructor
				}
			}
		}
		return ""
	}
	if operations[0].SourceID != main.Parameter.ID || shapeOf(operations[0].TargetID) != "FixtureBytes" {
		t.Fatalf("acquire = %+v", operations[0])
	}
	if operations[1].SourceID != operations[0].TargetID || shapeOf(operations[1].TargetID) != "ByteView" || operations[1].Foreign.Mode != "borrow" {
		t.Fatalf("view = %+v", operations[1])
	}
	if operations[2].SourceID != operations[1].TargetID || shapeOf(operations[2].TargetID) != "U64" || operations[2].CalleeID == "" {
		t.Fatalf("call = %+v", operations[2])
	}
	if operations[3].ReleasesOperationID != operations[0].ID || operations[3].SourceID != operations[0].TargetID ||
		operations[3].Foreign == nil || operations[3].Foreign.Symbol != operations[0].Foreign.Release || operations[3].Allocator != operations[0].Allocator {
		t.Fatalf("release = %+v, want minted from the acquire policy %+v", operations[3], operations[0].Foreign)
	}
	if operations[4].SourceID != operations[2].TargetID {
		t.Fatalf("return = %+v, want the call result", operations[4])
	}
}

// TestPhase29FixtureCalleeNotNamePinned renames the library function to
// checksum (declaration and call site) and requires the same core modulo the
// callee ID: the OpCall is resolved through the callee-contract table, not by
// a pinned name.
func TestPhase29FixtureCalleeNotNamePinned(t *testing.T) {
	original := phase29FixtureSource(t)
	renamed := strings.Replace(strings.Replace(original, "fn adler32(", "fn checksum(", 1), "= adler32(view)", "= checksum(view)", 1)
	if renamed == original || !strings.Contains(renamed, "fn checksum(bytes: ByteView) -> U64") {
		t.Fatal("rename did not apply")
	}
	want, codes := phase29CheckSource(t, []byte(original))
	if len(codes) != 0 {
		t.Fatalf("original diagnostics: %v", codes)
	}
	got, codes := phase29CheckSource(t, []byte(renamed))
	if len(codes) != 0 {
		t.Fatalf("renamed diagnostics: %v", codes)
	}
	wantMain, gotMain := phase29FixtureMain(t, want), phase29FixtureMain(t, got)
	wantCall, gotCall := &wantMain.Linear.Operations[2], &gotMain.Linear.Operations[2]
	if !strings.HasSuffix(wantCall.CalleeID, ":fn:adler32") || !strings.HasSuffix(gotCall.CalleeID, ":fn:checksum") {
		t.Fatalf("callee IDs = %q, %q", wantCall.CalleeID, gotCall.CalleeID)
	}
	gotCall.CalleeID = wantCall.CalleeID
	if !reflect.DeepEqual(wantMain.Linear, gotMain.Linear) {
		t.Fatalf("renamed main core differs beyond the callee ID:\nwant %+v\ngot  %+v", wantMain.Linear, gotMain.Linear)
	}
}

// TestPhase29FixtureShapeFaults pins the first diagnostic code for each seeded
// caller-shape fault: a ByteView-returning main, a FixtureBytes-taking callee,
// a second call on the moved view, a wrong abi name and a missing release
// policy.
func TestPhase29FixtureShapeFaults(t *testing.T) {
	base := phase29FixtureSource(t)
	callTail := "  let sum = adler32(view)\n  sum\n}\n"
	if !strings.Contains(base, callTail) {
		t.Fatal("fixture main tail changed")
	}
	cases := []struct {
		name   string
		mutate func(string) string
		code   string
	}{
		{
			name: "view returned from main",
			mutate: func(s string) string {
				s = strings.Replace(s, "fn main(case: U64) -> U64 {", "fn main(case: U64) -> ByteView {", 1)
				return strings.Replace(s, callTail, "  view\n}\n", 1)
			},
			code: "check.fixture_owner_shape_unsupported",
		},
		{
			name: "callee takes the owner",
			mutate: func(s string) string {
				s = strings.Replace(s, "let sum = adler32(view)", "let sum = takes_owner(view)", 1)
				return s + "\nfn takes_owner(owner: FixtureBytes) -> U64 {\n  var x = 1\n  x\n}\n"
			},
			code: "check.call_argument_type_mismatch",
		},
		{
			name: "second call on the moved view",
			mutate: func(s string) string {
				return strings.Replace(s, callTail, "  let sum = adler32(view)\n  let again = adler32(view)\n  again\n}\n", 1)
			},
			code: "check.fixture_owner_shape_unsupported",
		},
		{
			name: "wrong abi name",
			mutate: func(s string) string {
				return strings.Replace(s, "abi: schway_fixture_bytes_view_fn", "abi: schway_fixture_bytes_view_other", 1)
			},
			code: "check.fixture_owner_contract_invalid",
		},
		{
			name: "missing release policy",
			mutate: func(s string) string {
				return strings.Replace(s, "    release: \"schway_fixture_bytes_release\"\n", "", 1)
			},
			code: "check.fixture_owner_contract_invalid",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mutated := tc.mutate(base)
			if mutated == base {
				t.Fatal("mutation did not apply")
			}
			_, codes := phase29CheckSource(t, []byte(mutated))
			if len(codes) == 0 {
				t.Fatal("mutated fixture was accepted")
			}
			if codes[0] != tc.code {
				t.Fatalf("first code = %q (all %v), want %q", codes[0], codes, tc.code)
			}
		})
	}
	if _, codes := phase29CheckSource(t, []byte(base)); len(codes) != 0 {
		t.Fatalf("unmutated fixture diagnostics: %v", codes)
	}
}
