package corevalidate

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/syntax"
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

// phase31PeerForge checks the inline source through syntax.Parse and
// check.Program, deep-copies the program through its JSON form and hands the
// copy to mutate. It returns the program and the index of main.
func phase31PeerForge(t *testing.T, tail string, mutate func(p *core.Program, f *core.Function)) (core.Program, int) {
	t.Helper()
	parsed := syntax.Parse(phase31PeerSource("PathToken", "FileError", tail))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check: %+v", checked.Diagnostics)
	}
	raw, err := json.Marshal(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	var program core.Program
	if err := json.Unmarshal(raw, &program); err != nil {
		t.Fatal(err)
	}
	for index := range program.Functions {
		if program.Functions[index].Name == "main" {
			mutate(&program, &program.Functions[index])
			return program, index
		}
	}
	t.Fatal("no main")
	return core.Program{}, 0
}

func phase31PeerCodes(result Result) []string {
	codes := make([]string, 0, len(result.Problems))
	for _, problem := range result.Problems {
		codes = append(codes, problem.Code)
	}
	return codes
}

func TestPhase31FileBytesSeededFaults(t *testing.T) {
	for name, tail := range map[string]string{"positive-5op": phase31PeerTail5, "positive-6op": phase31PeerTail6} {
		t.Run(name, func(t *testing.T) {
			program, index := phase31PeerForge(t, tail, func(*core.Program, *core.Function) {})
			if result := Validate(program); !result.Valid {
				t.Fatalf("unforged program refused: %v", phase31PeerCodes(result))
			}
			v := &validator{program: program}
			if !v.fileBytesOwnerLifecycleValid(&program.Functions[index]) {
				t.Fatal("unforged lifecycle refused")
			}
		})
	}
	renumber := func(f *core.Function) {
		for index := range f.Linear.Operations {
			f.Linear.Operations[index].ID = fmt.Sprintf("%s:op:%d", f.ID, index)
			f.Linear.Operations[index].PointID = fmt.Sprintf("%s:point:linear:%d", f.ID, index)
		}
	}
	cases := []struct {
		name   string
		mutate func(p *core.Program, f *core.Function)
	}{
		{"dropped-release", func(_ *core.Program, f *core.Function) {
			ops := f.Linear.Operations
			f.Linear.Operations = append(append([]core.LinearOperation{}, ops[:3]...), ops[4:]...)
			renumber(f)
		}},
		{"release-before-view-call", func(_ *core.Program, f *core.Function) {
			ops := f.Linear.Operations
			ops[2], ops[3] = ops[3], ops[2]
			renumber(f)
		}},
		{"second-trailing-call", func(_ *core.Program, f *core.Function) {
			ops := f.Linear.Operations
			second := ops[4]
			second.SourceID = ops[4].TargetID
			second.TargetID = f.ID + ":place:6"
			f.Linear.Places = append(f.Linear.Places, core.Place{ID: second.TargetID, Name: "again", TypeID: ops[4].TypeID})
			f.Linear.Operations = append(append(append([]core.LinearOperation{}, ops[:5]...), second), ops[5:]...)
			renumber(f)
			f.Linear.Operations[6].SourceID = second.TargetID
		}},
		{"trailing-call-takes-view", func(_ *core.Program, f *core.Function) {
			f.Linear.Operations[4].SourceID = f.Linear.Operations[1].TargetID
		}},
		{"trailing-call-takes-owner", func(_ *core.Program, f *core.Function) {
			f.Linear.Operations[4].SourceID = f.Linear.Operations[0].TargetID
		}},
		{"trailing-call-before-release", func(_ *core.Program, f *core.Function) {
			ops := f.Linear.Operations
			ops[3], ops[4] = ops[4], ops[3]
			renumber(f)
		}},
		{"trailing-callee-not-u64", func(p *core.Program, f *core.Function) {
			for index := range p.Functions {
				if p.Functions[index].ID == f.Linear.Operations[4].CalleeID {
					p.Functions[index].ReturnType = "Unit"
				}
			}
		}},
		{"acquire-parameter-not-pathtoken", func(_ *core.Program, f *core.Function) {
			f.Linear.Operations[0].Foreign.ParameterType = "U64"
		}},
		{"acquire-fails-not-fileerror", func(_ *core.Program, f *core.Function) {
			f.Linear.Operations[0].Foreign.Fails = "FileFailure"
		}},
	}
	contract := map[string]bool{"acquire-parameter-not-pathtoken": true, "acquire-fails-not-fileerror": true}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			program, index := phase31PeerForge(t, phase31PeerTail6, tc.mutate)
			if result := Validate(program); result.Valid {
				t.Fatal("forged program was accepted by Validate")
			}
			// Generic replay laws may refuse first (first-problem-only), so the
			// file-bytes matchers are exercised directly: each must refuse alone.
			v := &validator{program: program}
			function := &program.Functions[index]
			if contract[tc.name] {
				types := map[string]core.TypeFact{}
				for _, fact := range function.Linear.Types {
					types[fact.ID] = fact
				}
				places := map[string]core.Place{}
				for _, place := range function.Linear.Places {
					places[place.ID] = place
				}
				operation := function.Linear.Operations[0]
				if v.validateFileBytesForeignOperation(operation, types, places) {
					t.Fatal("file-bytes foreign matcher accepted the forged acquire contract")
				}
				if len(v.problems) == 0 || v.problems[0].Code != "foreign.operation_contract_invalid" {
					t.Fatalf("problems = %+v, want foreign.operation_contract_invalid", v.problems)
				}
				return
			}
			if v.fileBytesOwnerLifecycleValid(function) {
				t.Fatal("file-bytes lifecycle matcher accepted the forged shape")
			}
		})
	}
}
