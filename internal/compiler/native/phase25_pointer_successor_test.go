package native_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase25SharedNativeShape(t *testing.T) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal("Phase 25 native pointer witness requires installed clang")
	}
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase25", "shared_copy_accept.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("shared-copy fixture failed to check: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitProgramNativeForTest(checked.Program)
	if err != nil {
		t.Fatalf("shared pointer helper was refused: %v", err)
	}
	path := filepath.Join(t.TempDir(), "shared_copy.c")
	if err := os.WriteFile(path, []byte(generated), 0o600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(clang, "-std=c17", "-Werror", "-fsyntax-only", path).CombinedOutput(); err != nil {
		t.Fatalf("clang rejected emitted shared-pointer C: %v\n%s", err, output)
	}
}

func phase25ExclusiveNativeProgram(t *testing.T) (core.Program, string) {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase25", "exclusive_copy_accept.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("exclusive-copy fixture failed to check: %+v", checked.Diagnostics)
	}
	program := checked.Program
	helper := program.Functions[0]
	typeID, callerID := "phase25:native-main:type:u64", "phase25:native-main"
	parameterID, resultID := callerID+":place:parameter", callerID+":place:result"
	program.Functions = append(append([]core.Function(nil), program.Functions...), core.Function{
		ID: callerID, Name: "main", EntryPointID: callerID + ":point:entry", ReturnPointID: callerID + ":point:return",
		Parameter: core.Parameter{ID: parameterID, Name: "input", Type: "U64"}, ReturnType: "U64",
		Linear: &core.LinearBody{
			ID: callerID + ":linear", Types: []core.TypeFact{{ID: typeID, Shape: core.TypeRef{Constructor: "U64"}}},
			Places: []core.Place{{ID: parameterID, Name: "input", TypeID: typeID}, {ID: resultID, Name: "result", TypeID: typeID}},
			Operations: []core.LinearOperation{
				{ID: callerID + ":op:0", PointID: callerID + ":point:linear:0", Kind: core.OpCall, SourceID: parameterID, TargetID: resultID, TypeID: typeID, CalleeID: helper.ID},
				{ID: callerID + ":op:1", PointID: callerID + ":point:linear:1", Kind: core.OpReturn, SourceID: resultID, TypeID: typeID},
			},
		},
	})
	return program, program.Functions[0].Linear.Operations[1].ID
}

func TestPhase25ExclusiveWrongResult(t *testing.T) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal("Phase 25 native pointer witness requires installed clang")
	}
	program, copyID := phase25ExclusiveNativeProgram(t)
	generated, err := cgen.EmitProgramNativeForTest(program)
	if err != nil {
		t.Fatalf("exclusive pointer helper was refused: %v", err)
	}
	mutated := ""
	for _, line := range strings.Split(generated, "\n") {
		if strings.Contains(line, "/* copy: "+copyID+" */") {
			comment := strings.Index(line, " /* copy:")
			assign := strings.Index(line[:comment], " = ")
			mutated += line[:assign+3] + "UINT64_C(99); /* reached wrong-result control */\n"
		} else {
			mutated += line + "\n"
		}
	}
	if mutated == generated {
		t.Fatalf("exclusive copy operation %q was absent from generated C", copyID)
	}
	for _, test := range []struct{ name, source, want string }{
		{"baseline", generated, `"value":"65"`}, {"mutated helper", mutated, `"value":"99"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			cPath, binaryPath := filepath.Join(dir, "exclusive.c"), filepath.Join(dir, "exclusive")
			if err := os.WriteFile(cPath, []byte(test.source), 0o600); err != nil {
				t.Fatal(err)
			}
			if output, err := exec.Command(clang, "-std=c17", "-Werror", cPath, "-o", binaryPath).CombinedOutput(); err != nil {
				t.Fatalf("clang failed: %v\n%s", err, output)
			}
			output, err := exec.Command(binaryPath, "65").CombinedOutput()
			if err != nil || !strings.Contains(string(output), test.want) {
				t.Fatalf("native result=%q err=%v, want payload containing %q", output, err, test.want)
			}
		})
	}
}
