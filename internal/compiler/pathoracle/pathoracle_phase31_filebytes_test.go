package pathoracle_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase31PeerChecked returns the checked checksum program. With trailing set
// it checks a tempdir copy of examples/checksum whose main ends with one
// trailing U64 -> U64 call to a function defined in the entry module.
func phase31PeerChecked(t *testing.T, trailing bool) core.Program {
	t.Helper()
	source := testsupport.ProjectPath("examples", "checksum", "checksum.schway")
	if trailing {
		dir := t.TempDir()
		for _, name := range []string{"checksum.schway", "adler32.schway"} {
			data, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", name))
			if err != nil {
				t.Fatal(err)
			}
			if name == "checksum.schway" {
				text := string(data)
				const body = "  let sum = adler32(view)\n  sum\n}\n"
				if strings.Count(text, body) != 1 {
					t.Fatal("checksum.schway no longer has the expected main tail")
				}
				text = strings.Replace(text, body, "  let sum = adler32(view)\n  let low = low16(sum)\n  low\n}\n\nfn low16(sum: U64) -> U64 {\n  var r = sum\n  r = r + r\n  r\n}\n", 1)
				data = []byte(text)
			}
			if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		source = filepath.Join(dir, "checksum.schway")
	}
	checked, err := session.CheckFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

// phase31PeerForge deep-copies the checked program through its canonical JSON
// form and hands the copy to mutate together with the entry main.
func phase31PeerForge(t *testing.T, trailing bool, mutate func(p *core.Program, f *core.Function)) core.Program {
	t.Helper()
	raw, err := json.Marshal(phase31PeerChecked(t, trailing))
	if err != nil {
		t.Fatal(err)
	}
	var program core.Program
	if err := json.Unmarshal(raw, &program); err != nil {
		t.Fatal(err)
	}
	for index := range program.Functions {
		if program.Functions[index].Name == "main" && program.Functions[index].Parameter.Type == "PathToken" {
			mutate(&program, &program.Functions[index])
		}
	}
	return program
}

func phase31PeerRenumber(f *core.Function) {
	for index := range f.Linear.Operations {
		f.Linear.Operations[index].ID = fmt.Sprintf("%s:op:%d", f.ID, index)
		f.Linear.Operations[index].PointID = fmt.Sprintf("%s:point:linear:%d", f.ID, index)
	}
}

type phase31PeerCase struct {
	name   string
	mutate func(p *core.Program, f *core.Function)
}

// phase31PeerCases are the nine seeded faults, all forged on the 6-operation
// program. Every peer carries every row.
func phase31PeerCases() []phase31PeerCase {
	return []phase31PeerCase{
		{"dropped-release", func(_ *core.Program, f *core.Function) {
			ops := f.Linear.Operations
			f.Linear.Operations = append(append([]core.LinearOperation{}, ops[:3]...), ops[4:]...)
			phase31PeerRenumber(f)
		}},
		{"release-before-view-call", func(_ *core.Program, f *core.Function) {
			ops := f.Linear.Operations
			ops[2], ops[3] = ops[3], ops[2]
			phase31PeerRenumber(f)
		}},
		{"second-trailing-call", func(_ *core.Program, f *core.Function) {
			ops := f.Linear.Operations
			second := ops[4]
			second.SourceID = ops[4].TargetID
			second.TargetID = f.ID + ":place:6"
			f.Linear.Places = append(f.Linear.Places, core.Place{ID: second.TargetID, Name: "again", TypeID: ops[4].TypeID})
			f.Linear.Operations = append(append(append([]core.LinearOperation{}, ops[:5]...), second), ops[5:]...)
			phase31PeerRenumber(f)
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
			phase31PeerRenumber(f)
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
}

func TestPhase31FileBytesSeededFaults(t *testing.T) {
	for _, trailing := range []bool{false, true} {
		name := "positive-5op"
		if trailing {
			name = "positive-6op"
		}
		t.Run(name, func(t *testing.T) {
			if err := pathoracle.ValidateFileBytesOwnerPaths(phase31PeerForge(t, trailing, func(*core.Program, *core.Function) {})); err != nil {
				t.Fatalf("unforged program refused: %v", err)
			}
		})
	}
	for _, tc := range phase31PeerCases() {
		t.Run(tc.name, func(t *testing.T) {
			err := pathoracle.ValidateFileBytesOwnerPaths(phase31PeerForge(t, true, tc.mutate))
			if err == nil {
				t.Fatal("forged program was accepted")
			}
			if !strings.HasPrefix(err.Error(), "pathoracle.file_bytes_owner_") && !strings.HasPrefix(err.Error(), "pathoracle.release_while_view_live") {
				t.Fatalf("error = %v, want a pathoracle.file_bytes_owner_ code", err)
			}
		})
	}
}
