package check

import (
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
)

// phase31PeerHeader is the inline single-file counterpart of
// examples/checksum/checksum.schway. check cannot import session, so the
// foreign block, FileError and main shape are written out here. The ByteView
// loop is named phase31sum and carries no modulus literal.
const phase31PeerHeader = `module phase31.peer

export {
  fn main
}

foreign C {
  fn schway_file_bytes_acquire(path: %s) -> FileBytes {
    unwind: forbidden
    nonlocal_exit: forbidden
    mode: "acquire"
    abi: schway_file_bytes_acquire_fn
    allocator: "libc_malloc"
    release: "schway_file_bytes_release"
    fails: %s
  }

  fn schway_file_bytes_view(owner: FileBytes) -> ByteView {
    unwind: forbidden
    nonlocal_exit: forbidden
    mode: "borrow"
    abi: schway_file_bytes_view_fn
  }

  fn schway_file_bytes_release(owner: FileBytes) -> Unit {
    unwind: forbidden
    nonlocal_exit: forbidden
    mode: "consume"
    abi: schway_file_bytes_release_fn
    allocator: "libc_malloc"
  }
}

data FileError =
  | FileTooLarge
  | FileUnavailable
  | AllocFailed

data FileFailure =
  | Broken

fn phase31sum(bytes: ByteView) -> U64 {
  var a = 1
  var i = 0
  var n = byte_len(bytes)
  while i < n {
    a = a + byte_at(bytes, i)
    i = i + 1
  }
  a
}

fn low16(sum: U64) -> U64 {
  var r = sum
  r = r + r
  r
}

fn sink(sum: U64) -> Unit {
  var r = sum
  r = r + r
}

fn main(path: PathToken) -> U64 {
  let owner = try schway_file_bytes_acquire(path)
  let view = schway_file_bytes_view(owner)
  let sum = phase31sum(view)
`

func phase31PeerSource(path, fails, tail string) []byte {
	return []byte(strings.Replace(strings.Replace(phase31PeerHeader, "%s", path, 1), "%s", fails, 1) + tail)
}

const (
	phase31PeerTail5 = "  sum\n}\n"
	phase31PeerTail6 = "  let low = low16(sum)\n  low\n}\n"
)

// phase31PeerCheckedMain checks the inline source and returns its main.
func phase31PeerCheckedMain(t *testing.T, source []byte) (core.Function, []string) {
	t.Helper()
	program, codes := phase29CheckSource(t, source)
	if len(codes) != 0 {
		return core.Function{}, codes
	}
	for _, function := range program.Functions {
		if function.Name == "main" {
			return function, nil
		}
	}
	t.Fatal("checked program has no main")
	return core.Function{}, nil
}

// phase31PeerForge returns a deep-enough copy of the checked 6-operation main
// (operations, places and types are copied) after applying mutate.
func phase31PeerForge(t *testing.T, tail string, mutate func(f *core.Function)) core.Function {
	t.Helper()
	main, codes := phase31PeerCheckedMain(t, phase31PeerSource("PathToken", "FileError", tail))
	if len(codes) != 0 {
		t.Fatalf("diagnostics: %v", codes)
	}
	linear := *main.Linear
	linear.Operations = append([]core.LinearOperation{}, linear.Operations...)
	linear.Places = append([]core.Place{}, linear.Places...)
	linear.Types = append([]core.TypeFact{}, linear.Types...)
	main.Linear = &linear
	mutate(&main)
	return main
}

func phase31PeerFirst(codes []string) string {
	if len(codes) == 0 {
		return ""
	}
	return codes[0]
}

func TestPhase31FileBytesSeededFaults(t *testing.T) {
	t.Run("positive-5op", func(t *testing.T) {
		main := phase31PeerForge(t, phase31PeerTail5, func(*core.Function) {})
		if len(main.Linear.Operations) != 5 || len(checkFileBytesOwnerPostAssembly(main)) != 0 {
			t.Fatalf("5-operation program refused: %d operations", len(main.Linear.Operations))
		}
	})
	t.Run("positive-6op", func(t *testing.T) {
		main := phase31PeerForge(t, phase31PeerTail6, func(*core.Function) {})
		if len(main.Linear.Operations) != 6 || len(checkFileBytesOwnerPostAssembly(main)) != 0 {
			t.Fatalf("6-operation program refused: %d operations", len(main.Linear.Operations))
		}
	})

	// Core-level rows go through checkFileBytesOwnerPostAssembly.
	coreRow := func(name, code string, mutate func(f *core.Function)) {
		t.Run(name, func(t *testing.T) {
			main := phase31PeerForge(t, phase31PeerTail6, mutate)
			diagnostics := checkFileBytesOwnerPostAssembly(main)
			if len(diagnostics) == 0 {
				t.Fatal("forged core was accepted")
			}
			if !strings.HasPrefix(diagnostics[0].Code, code) {
				t.Fatalf("first code = %q, want prefix %q", diagnostics[0].Code, code)
			}
		})
	}
	coreRow("dropped-release", "check.file_bytes_owner_", func(f *core.Function) {
		ops := f.Linear.Operations
		f.Linear.Operations = append(append([]core.LinearOperation{}, ops[:3]...), ops[4:]...)
	})
	coreRow("release-before-view-call", "check.release_while_view_live", func(f *core.Function) {
		ops := f.Linear.Operations
		ops[2], ops[3] = ops[3], ops[2]
	})
	coreRow("trailing-call-before-release", "check.file_bytes_owner_", func(f *core.Function) {
		ops := f.Linear.Operations
		ops[3], ops[4] = ops[4], ops[3]
	})

	// Source-level rows go through check.Program on a mutated inline source.
	sourceRow := func(name, code, path, fails, tail string) {
		t.Run(name, func(t *testing.T) {
			_, codes := phase31PeerCheckedMain(t, phase31PeerSource(path, fails, tail))
			if len(codes) == 0 {
				t.Fatal("forged source was accepted")
			}
			if !strings.HasPrefix(codes[0], code) {
				t.Fatalf("first code = %q, want prefix %q", codes[0], code)
			}
		})
	}
	sourceRow("second-trailing-call", "check.file_bytes_owner_", "PathToken", "FileError", "  let low = low16(sum)\n  let again = low16(low)\n  again\n}\n")
	sourceRow("trailing-call-takes-view", "check.file_bytes_owner_", "PathToken", "FileError", "  let low = low16(view)\n  low\n}\n")
	sourceRow("trailing-call-takes-owner", "check.file_bytes_owner_", "PathToken", "FileError", "  let low = low16(owner)\n  low\n}\n")
	sourceRow("trailing-callee-not-u64", "check.file_bytes_owner_", "PathToken", "FileError", "  let low = sink(sum)\n  low\n}\n")
	sourceRow("acquire-parameter-not-pathtoken", "check.file_bytes_owner_", "U64", "FileError", phase31PeerTail6)
	sourceRow("acquire-fails-not-fileerror", "check.file_bytes_owner_", "PathToken", "FileFailure", phase31PeerTail6)
}

// TestPhase31FileBytesNeighborRefusals names the source-level neighbours of
// the third owner shape (D-16): each is refused with the shape code.
func TestPhase31FileBytesNeighborRefusals(t *testing.T) {
	for _, tail := range []string{phase31PeerTail5, phase31PeerTail6} {
		if _, codes := phase31PeerCheckedMain(t, phase31PeerSource("PathToken", "FileError", tail)); len(codes) != 0 {
			t.Fatalf("positive control refused: %v", codes)
		}
	}
	cases := []struct {
		name string
		tail string
	}{
		{"second trailing call", "  let low = low16(sum)\n  let again = low16(low)\n  again\n}\n"},
		{"trailing call takes the view", "  let low = low16(view)\n  low\n}\n"},
		{"trailing callee returns Unit", "  let low = sink(sum)\n  low\n}\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, codes := phase31PeerCheckedMain(t, phase31PeerSource("PathToken", "FileError", tc.tail))
			if phase31PeerFirst(codes) != "check.file_bytes_owner_shape_unsupported" {
				t.Fatalf("codes = %v, want first check.file_bytes_owner_shape_unsupported", codes)
			}
		})
	}
}

// TestPhase31FileErrorDeclarationIsPinned arms WR-02: the adapter statuses map
// onto FileError by position, so check refuses every declaration other than
// FileTooLarge, FileUnavailable, AllocFailed in that order. The faithful row is
// the control; each mutant is the same program with only the declaration
// changed.
func TestPhase31FileErrorDeclarationIsPinned(t *testing.T) {
	declared := "data FileError =\n  | FileTooLarge\n  | FileUnavailable\n  | AllocFailed\n"
	source := func(declaration string) []byte {
		text := string(phase31PeerSource("PathToken", "FileError", phase31PeerTail5))
		if strings.Count(text, declared) != 1 {
			t.Fatalf("FileError declaration anchor occurs %d times", strings.Count(text, declared))
		}
		return []byte(strings.Replace(text, declared, declaration, 1))
	}
	if _, codes := phase31PeerCheckedMain(t, source(declared)); len(codes) != 0 {
		t.Fatalf("faithful FileError refused: %v", codes)
	}
	for name, declaration := range map[string]string{
		"reversed":  "data FileError =\n  | AllocFailed\n  | FileUnavailable\n  | FileTooLarge\n",
		"swapped":   "data FileError =\n  | FileUnavailable\n  | FileTooLarge\n  | AllocFailed\n",
		"too-few":   "data FileError =\n  | FileTooLarge\n  | FileUnavailable\n",
		"too-many":  "data FileError =\n  | FileTooLarge\n  | FileUnavailable\n  | AllocFailed\n  | Extra\n",
		"renamed":   "data FileError =\n  | FileTooLarge\n  | FileUnavailable\n  | OutOfMemory\n",
		"duplicate": "data FileError =\n  | FileTooLarge\n  | FileTooLarge\n  | AllocFailed\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, codes := phase31PeerCheckedMain(t, source(declaration))
			if phase31PeerFirst(codes) != "check.file_bytes_error_alternatives" {
				t.Fatalf("codes = %v, want check.file_bytes_error_alternatives first", codes)
			}
		})
	}
}
