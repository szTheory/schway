package session_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
)

// Phase 30 session route tests (plan 30-08). Fixtures are generated under
// t.TempDir() only and never committed (D-27).

// writeModuleFixture writes files (slash-separated relative name to content)
// under a fresh TempDir and returns the path of main.schway.
func writeModuleFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, "main.schway")
}

const (
	phase30TwiceEntry = "module app.main\n\nimport lib {\n  fn twice\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = twice(n)\n  r\n}\n"
	phase30TwiceLib   = "module lib\n\nexport {\n  fn twice\n}\n\nfn twice(n: U64) -> U64 {\n  let d = n + n\n  d\n}\n"
)

func phase30TwiceFixture(t *testing.T) string {
	t.Helper()
	return writeModuleFixture(t, map[string]string{"main.schway": phase30TwiceEntry, "lib.schway": phase30TwiceLib})
}

func phase30ResultJSON(t *testing.T, result protocol.Result) string {
	t.Helper()
	encoded, err := protocol.JSON(result)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func phase30ResultHuman(t *testing.T, result protocol.Result) string {
	t.Helper()
	human, err := protocol.Human(result)
	if err != nil {
		t.Fatal(err)
	}
	return human
}

// phase30InterpretU64 runs a checked merged program under the interpreter the
// way the application lane does: resolve the entry, prune to what it reaches,
// then run it on one canonical U64 input.
func phase30InterpretU64(t *testing.T, program core.Program, input string) string {
	t.Helper()
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		t.Fatal(err)
	}
	pruned, err := callgraph.PruneToReachable(program, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	result, err := interp.RunApplication(pruned, entry.Name, input)
	if err != nil {
		t.Fatal(err)
	}
	return result.Execution.Outcome.Value
}

// TestPhase30CheckFileThroughLoader is the tracer: check and the interpreter
// route load an entry that imports a library, through modload and the merged
// program (D-22).
func TestPhase30CheckFileThroughLoader(t *testing.T) {
	entry := phase30TwiceFixture(t)

	result, err := session.CheckCommandFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != protocol.StatusPass {
		t.Fatalf("CheckCommandFile status = %s: %s", result.Status, phase30ResultJSON(t, result))
	}

	checked, err := session.CheckFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("CheckFile diagnostics: %+v", checked.Diagnostics)
	}
	if len(checked.UnitTrees) != 2 {
		t.Fatalf("UnitTrees = %d trees, want entry and lib", len(checked.UnitTrees))
	}
	if got := phase30InterpretU64(t, checked.Program, "21"); got != "42" {
		t.Fatalf("interpreter result = %q, want 42", got)
	}

	t.Run("interpreter-route", func(t *testing.T) {
		byteEntry := writeModuleFixture(t, map[string]string{
			"main.schway": "module app.main\n\nimport lib {\n  fn same\n}\n\nexport {\n  fn main\n}\n\nfn main(value: Byte) -> Byte {\n  let r = same(value)\n  r\n}\n",
			"lib.schway":  "module lib\n\nexport {\n  fn same\n}\n\nfn same(value: Byte) -> Byte {\n  value\n}\n",
		})
		executions, diagnostics, err := session.RunInterpreterFile(byteEntry)
		if err != nil || len(diagnostics) != 0 || len(executions) != 1 {
			t.Fatalf("RunInterpreterFile = %d executions, %+v, %v", len(executions), diagnostics, err)
		}
		if executions[0].Outcome.Value != "7" {
			t.Fatalf("outcome = %q, want 7", executions[0].Outcome.Value)
		}
		command, err := session.RunInterpreterCommandFile(byteEntry)
		if err != nil || command.Status != protocol.StatusPass || len(command.Executions) != 1 {
			t.Fatalf("RunInterpreterCommandFile = %s, %v", phase30ResultJSON(t, command), err)
		}
	})

	t.Run("no-import-bytes-identical", func(t *testing.T) {
		source, err := os.ReadFile(filepath.Join("..", "..", "..", "examples", "phase22", "identity.schway"))
		if err != nil {
			t.Fatal(err)
		}
		path := writeModuleFixture(t, map[string]string{"main.schway": string(source)})
		fromFile, err := session.CheckFile(path)
		if err != nil {
			t.Fatal(err)
		}
		fromBytes := session.Check(source)
		fileJSON, _ := json.Marshal(fromFile)
		byteJSON, _ := json.Marshal(fromBytes)
		if !bytes.Equal(fileJSON, byteJSON) {
			t.Fatalf("CheckFile differs from Check(bytes):\n%s\n%s", fileJSON, byteJSON)
		}
		if fromFile.UnitTrees != nil {
			t.Fatal("a no-import program must carry no UnitTrees")
		}
	})
}
