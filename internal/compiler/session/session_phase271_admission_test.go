package session_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// requirePhase271ControlReached proves the origin-unpublishable control is
// actually reached by every layer before the admission tests assert refusal
// (D-27.1-12): checker clean, corevalidate valid, a linear U64->U64 entry, and
// exactly one core.origin_omitted problem from the published-origin peer.
func requirePhase271ControlReached(t *testing.T) []byte {
	t.Helper()
	source := mustPhase25Fixture(t, "shared_copy_accept.schway")
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("phase27.1 control not reached: checker diagnostics %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("phase27.1 control not reached: corevalidate refused: %+v", validated.Problems)
	}
	program := validated.Program()
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		t.Fatalf("phase27.1 control not reached: entry function: %v", err)
	}
	if entry.Linear == nil || entry.Match != nil || entry.Parameter.Type != "U64" || entry.ReturnType != "U64" {
		t.Fatalf("phase27.1 control not reached: entry is not a linear U64->U64 function: %+v", entry)
	}
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_omitted" {
		t.Fatalf("phase27.1 control not reached: ValidatePublished = %+v, want exactly [core.origin_omitted]", problems)
	}
	return source
}

// phase271MissingClangRunner returns a runner whose clang does not exist, so
// any route that proceeds past admission surfaces native.tool_missing.
func phase271MissingClangRunner(t *testing.T) native.Runner {
	t.Helper()
	return native.Runner{ClangPath: filepath.Join(t.TempDir(), "missing-clang")}
}

func TestPhase271BuildApplicationRefusesOriginUnpublishable(t *testing.T) {
	source := requirePhase271ControlReached(t)
	outDir := t.TempDir()
	artifact := filepath.Join(outDir, "app")
	receipt, diagnostics, err := session.BuildApplication(context.Background(), source, artifact, phase271MissingClangRunner(t))
	if err != nil {
		_, isTool := err.(*native.ToolError)
		t.Fatalf("BuildApplication returned err=%v (*native.ToolError=%v); want diagnostic refusal with nil error", err, isTool)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != "core.origin_omitted" {
		t.Fatalf("BuildApplication diagnostics = %+v, want exactly [core.origin_omitted]", diagnostics)
	}
	if !reflect.DeepEqual(receipt, native.BuildReceipt{}) {
		t.Fatalf("BuildApplication receipt = %+v, want zero value", receipt)
	}
	entries, readErr := os.ReadDir(outDir)
	if readErr != nil || len(entries) != 0 {
		t.Fatalf("output directory not empty after refusal: entries=%v err=%v", entries, readErr)
	}
	for _, path := range []string{artifact, artifact + ".schway-build.json"} {
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("%s exists after refusal (stat err=%v)", path, statErr)
		}
	}
	checkResult, checkErr := session.CheckCommandFile(testsupport.ProjectPath("testdata", "phase25", "shared_copy_accept.schway"))
	if checkErr != nil || len(checkResult.Diagnostics) == 0 {
		t.Fatalf("CheckCommandFile result=%+v err=%v", checkResult, checkErr)
	}
	if !reflect.DeepEqual(diagnostics[0], checkResult.Diagnostics[0]) {
		t.Fatalf("build diagnostic differs from check:\n build=%+v\n check=%+v", diagnostics[0], checkResult.Diagnostics[0])
	}
}
